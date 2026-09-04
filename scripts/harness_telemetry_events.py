import dataclasses
import json
import re
from typing import Optional

from harness_telemetry_schema import recognized_record


CODE_EXTENSIONS = (
    ".c", ".cc", ".cpp", ".cs", ".css", ".cts", ".go", ".h", ".hpp", ".html", ".java",
    ".js", ".jsx", ".kt", ".kts", ".lua", ".mjs", ".mts", ".php", ".py", ".rb", ".rs",
    ".scss", ".sh", ".sql", ".svelte", ".swift", ".toml", ".ts", ".tsx", ".vue", ".yaml",
    ".yml",
)
ERAS = ("avant", "depuis", "inconnu")
UNKNOWN = "<inconnu>"
NONE = "<aucun>"
EXPLICIT_WRITERS = ("write", "edit", "multiedit", "notebookedit", "apply_patch", "filechange")
SHELL_WRITERS = ("bash", "exec_command", "commandexecution")
SHELL_WRITE = re.compile(
    r"(?:>{1,2}\s*(?!/dev/null(?:\s|$))[^&\s]|\b(?:tee|touch|install|cp|mv)\s|\bsed\s+[^\n]*\s-i(?:\s|$)|"
    r"\bgit\s+(?:commit|apply|am|cherry-pick|merge|rebase)\b|\.write_(?:text|bytes)\s*\(|"
    r"\bopen\s*\([^\n]*[\"'][wax][bt+]?[\"'])",
    re.IGNORECASE,
)


@dataclasses.dataclass(frozen=True)
class Event:
    source: str
    kind: str
    timestamp: Optional[str]
    role: Optional[str] = None
    text: Optional[str] = None
    name: Optional[str] = None
    payload: object = None


def decode_json(value):
    if isinstance(value, dict):
        return value
    if not isinstance(value, str):
        return value
    try:
        return json.loads(value)
    except (TypeError, ValueError):
        return value


def claude_events(record):
    timestamp = record.get("timestamp")
    message = record.get("message")
    if not isinstance(message, dict):
        return
    role = message.get("role")
    content = message.get("content")
    if not isinstance(content, list):
        return
    for block in content:
        if not isinstance(block, dict):
            continue
        if block.get("type") == "text":
            yield Event("claude", "text", timestamp, role=role, text=block.get("text"))
        elif block.get("type") == "tool_use":
            yield Event(
                "claude", "tool", timestamp, name=block.get("name"), payload=block.get("input")
            )


def codex_events(record):
    timestamp = record.get("timestamp")
    payload = record.get("payload")
    if not isinstance(payload, dict):
        return
    kind = payload.get("type")
    if kind == "item_completed":
        item = payload.get("item")
        if not isinstance(item, dict):
            return
        item_type = item.get("type")
        if item_type == "FileChange":
            yield Event("codex", "tool", timestamp, name="filechange", payload=item)
        elif item_type == "CommandExecution":
            yield Event("codex", "tool", timestamp, name="commandexecution", payload=item)
        return
    if kind == "message":
        role = payload.get("role")
        for block in payload.get("content") or ():
            if isinstance(block, dict) and block.get("type") in ("input_text", "output_text"):
                yield Event("codex", "text", timestamp, role=role, text=block.get("text"))
    elif kind == "function_call":
        yield Event(
            "codex",
            "tool",
            timestamp,
            name=payload.get("name"),
            payload=decode_json(payload.get("arguments")),
        )
    elif kind == "custom_tool_call":
        source = payload.get("input")
        if not isinstance(source, str):
            yield Event("codex", "tool", timestamp, name=payload.get("name"), payload=source)
            return
        nested = re.findall(r"\btools\.([A-Za-z][A-Za-z0-9_]*)\s*\(", source)
        if nested:
            for name in nested:
                yield Event("codex", "tool", timestamp, name=name, payload=source)
        else:
            yield Event("codex", "tool", timestamp, name=payload.get("name"), payload=source)


def normalized_events(source, record):
    if source == "claude":
        yield from claude_events(record)
    else:
        yield from codex_events(record)


def labelled(payload, *keys):
    if not isinstance(payload, dict):
        return UNKNOWN
    for key in keys:
        if key in payload:
            value = payload[key]
            return NONE if value is None else str(value)
    return UNKNOWN


def era(timestamp, since):
    if not isinstance(timestamp, str) or not re.match(r"^\d{4}-\d{2}-\d{2}", timestamp):
        return "inconnu"
    return "avant" if timestamp[:10] < since else "depuis"


def command_text(value):
    if isinstance(value, str):
        return value
    if isinstance(value, list):
        parts = [command_text(item) for item in value]
        return " ".join(part for part in parts if part)
    return ""


def is_potential_write(event):
    name = (event.name or "").lower()
    if name == "write_stdin":
        return True
    if name not in SHELL_WRITERS:
        return False
    payload = event.payload
    if isinstance(payload, dict):
        payload = payload.get("command") or payload.get("cmd")
    return SHELL_WRITE.search(command_text(payload)) is not None


def filechange_lines(payload):
    changes = payload.get("changes")
    if not isinstance(changes, dict):
        return None
    lines = []
    for path, change in changes.items():
        if not str(path).endswith(CODE_EXTENSIONS) or not isinstance(change, dict):
            continue
        content = change.get("unified_diff") or change.get("content")
        if not isinstance(content, str):
            return None
        if change.get("type") == "add":
            lines.extend(content.splitlines())
        elif change.get("type") == "update":
            lines.extend(
                line[1:] for line in content.splitlines()
                if line.startswith("+") and not line.startswith("+++")
            )
    return lines


def patch_lines(payload):
    text = payload.replace("\\n", "\n")
    selected = False
    lines = []
    for line in text.splitlines():
        if line.startswith(("*** Add File: ", "*** Update File: ")):
            selected = line.split(": ", 1)[1].endswith(CODE_EXTENSIONS)
        elif line.startswith(("*** Delete File: ", "*** End Patch")):
            selected = False
        elif line.startswith("+++ "):
            path = line[4:]
            selected = (path[2:] if path.startswith("b/") else path).endswith(CODE_EXTENSIONS)
        elif selected and line.startswith("+") and not line.startswith("+++"):
            lines.append(line[1:])
    return lines


def content_lines(event):
    name = (event.name or "").lower()
    payload = event.payload
    if name == "write" and isinstance(payload, dict):
        path = payload.get("file_path") or payload.get("path")
        return str(payload.get("content") or "").splitlines() if str(path).endswith(CODE_EXTENSIONS) else []
    if name == "edit" and isinstance(payload, dict):
        path = payload.get("file_path") or payload.get("path")
        return str(payload.get("new_string") or "").splitlines() if str(path).endswith(CODE_EXTENSIONS) else []
    if name == "multiedit" and isinstance(payload, dict):
        path = payload.get("file_path") or payload.get("path")
        edits = payload.get("edits")
        if not str(path).endswith(CODE_EXTENSIONS) or not isinstance(edits, list):
            return []
        return [line for edit in edits if isinstance(edit, dict) for line in str(edit.get("new_string") or "").splitlines()]
    if name == "notebookedit" and isinstance(payload, dict):
        path = payload.get("notebook_path") or payload.get("file_path")
        return str(payload.get("new_source") or "").splitlines() if str(path).endswith(".ipynb") else []
    if name == "filechange" and isinstance(payload, dict):
        return filechange_lines(payload)
    if name != "apply_patch" or not isinstance(payload, str):
        return None
    return patch_lines(payload)
