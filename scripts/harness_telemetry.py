#!/usr/bin/env python3

import argparse
import collections
import hashlib
import json
import os
import re
import stat
import sys
import tempfile
import time

from harness_telemetry_events import (
    ERAS,
    EXPLICIT_WRITERS,
    NONE,
    UNKNOWN,
    content_lines,
    era,
    is_potential_write,
    labelled,
    normalized_events,
    recognized_record,
)
from harness_telemetry_report import activated_body_cost, print_report, rate


CACHE_VERSION = 9


def empty_summary(source):
    return {
        "source": source,
        "days": [],
        "skills": {},
        "agents": {},
        "blocks": {},
        "dash": {},
        "middle_dot": {},
        "lines": {},
        "comments": {},
        "write_tools": {},
        "uninspectable_writes": {},
        "records": {},
        "unknown_records": {},
        "invalid_records": {},
    }


def increment(summary, field, key, amount=1):
    summary[field][key] = summary[field].get(key, 0) + amount


def summarize_file(source, path, since):
    summary = empty_summary(source)
    with open(path, encoding="utf-8", errors="replace") as handle:
        for line in handle:
            try:
                record = json.loads(line)
            except (TypeError, ValueError):
                increment(summary, "invalid_records", source)
                continue
            if not isinstance(record, dict):
                increment(summary, "invalid_records", source)
                continue
            increment(summary, "records", source)
            if not recognized_record(source, record):
                increment(summary, "unknown_records", source)
                continue
            timestamp = record.get("timestamp")
            if isinstance(timestamp, str) and re.match(r"^\d{4}-\d{2}-\d{2}", timestamp):
                summary["days"].append(timestamp[:10])
            for event in normalized_events(source, record):
                period = era(event.timestamp, since)
                if event.kind == "text" and event.role == "assistant" and isinstance(event.text, str):
                    increment(summary, "blocks", period)
                    increment(summary, "dash", period, event.text.count("\u2014"))
                    increment(summary, "middle_dot", period, event.text.count("\u00b7"))
                    continue
                if event.kind != "tool":
                    continue
                name = (event.name or "").lower()
                if name == "skill":
                    increment(summary, "skills", labelled(event.payload, "skill", "name"))
                elif name in ("agent", "task", "spawn_agent"):
                    increment(summary, "agents", labelled(event.payload, "subagent_type", "agent_type"))
                if name not in EXPLICIT_WRITERS:
                    if is_potential_write(event):
                        increment(summary, "write_tools", name)
                        increment(summary, "uninspectable_writes", name)
                    continue
                increment(summary, "write_tools", name)
                written = content_lines(event)
                if written is None:
                    increment(summary, "uninspectable_writes", name)
                    continue
                for written_line in written:
                    stripped = written_line.strip()
                    if not stripped:
                        continue
                    increment(summary, "lines", period)
                    if stripped.startswith(("//", "#", "--", "/*", "*")):
                        increment(summary, "comments", period)
    return summary


def file_digest(path):
    digest = hashlib.sha256()
    with open(path, "rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def transcript_files(root):
    def raise_error(error):
        raise error

    for directory, _, files in os.walk(root, onerror=raise_error):
        for name in files:
            if name.endswith(".jsonl"):
                yield os.path.join(directory, name)


def load_cache(path, since):
    try:
        with open(path, encoding="utf-8") as handle:
            value = json.load(handle)
    except (OSError, TypeError, ValueError):
        return {}
    if value.get("version") != CACHE_VERSION or value.get("since") != since:
        return {}
    return value.get("files") if isinstance(value.get("files"), dict) else {}


def save_cache(path, since, entries):
    directory = os.path.dirname(path)
    os.makedirs(directory, mode=0o700, exist_ok=True)
    descriptor, temporary = tempfile.mkstemp(prefix="telemetry.", dir=directory, text=True)
    try:
        with os.fdopen(descriptor, "w", encoding="utf-8") as handle:
            json.dump({"version": CACHE_VERSION, "since": since, "files": entries}, handle, separators=(",", ":"))
        os.chmod(temporary, stat.S_IRUSR | stat.S_IWUSR)
        os.replace(temporary, path)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


def merge(summaries):
    result = empty_summary("all")
    for summary in summaries:
        result["days"].extend(summary["days"])
        for field in (
            "skills", "agents", "blocks", "dash", "middle_dot", "lines", "comments",
            "write_tools", "uninspectable_writes", "records", "unknown_records", "invalid_records",
        ):
            for key, value in summary[field].items():
                increment(result, field, key, value)
    return result


def collect_source(source, root, since, cache):
    entries = {}
    summaries = []
    hits = 0
    if not root or not os.path.isdir(root):
        return entries, summaries, hits
    for path in transcript_files(root):
        identity = hashlib.sha256(os.path.abspath(path).encode()).hexdigest()
        digest = file_digest(path)
        cached = cache.get(identity)
        if isinstance(cached, dict) and cached.get("digest") == digest and cached.get("source") == source:
            summary = cached.get("summary")
            hits += 1
        else:
            summary = summarize_file(source, path, since)
        if not isinstance(summary, dict):
            summary = summarize_file(source, path, since)
        entries[identity] = {"digest": digest, "source": source, "summary": summary}
        summaries.append(summary)
    return entries, summaries, hits


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--claude-root")
    parser.add_argument("--codex-root")
    parser.add_argument("--since", required=True)
    parser.add_argument("--cache", required=True)
    arguments = parser.parse_args()
    roots = (("claude", arguments.claude_root), ("codex", arguments.codex_root))
    cache = load_cache(arguments.cache, arguments.since)
    next_cache = {}
    summaries = []
    sources = collections.Counter()
    hits = 0
    started = time.monotonic()
    for source, root in roots:
        if not root or not os.path.isdir(root):
            print(f"  {source} : indisponible")
            continue
        try:
            entries, source_summaries, source_hits = collect_source(source, root, arguments.since, cache)
        except OSError:
            print(f"  {source} : lecture interrompue, aucun chemin affiché", file=sys.stderr)
            return 1
        next_cache.update(entries)
        summaries.extend(source_summaries)
        hits += source_hits
        sources[source] += len(source_summaries)
    if not summaries:
        print("  aucun transcript lu", file=sys.stderr)
        return 1
    try:
        save_cache(arguments.cache, arguments.since, next_cache)
    except OSError:
        print("  cache : écriture interrompue", file=sys.stderr)
        return 1
    total = merge(summaries)
    elapsed = round((time.monotonic() - started) * 1000)
    by_source = {source: merge(summary for summary in summaries if summary["source"] == source) for source in sources}
    local_sizes = {
        name: os.path.getsize(f"dot_config/agent-skills/{name}/SKILL.md")
        for name in os.listdir("dot_config/agent-skills")
        if os.path.isfile(f"dot_config/agent-skills/{name}/SKILL.md")
    }
    print_report(sources, total, by_source, hits, elapsed, local_sizes, UNKNOWN, NONE, ERAS)
    malformed = sum(total["invalid_records"].values())
    unknown = sum(total["unknown_records"].values())
    if malformed or unknown:
        print(f"  formats non mesurés : {unknown} inconnus, {malformed} invalides", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
