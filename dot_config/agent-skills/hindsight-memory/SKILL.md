---
name: hindsight-memory
description: >
  Use Hindsight for durable project memory. Use when prior work, a lasting preference, a project
  convention, or a reusable procedure could affect the task. Make sure to use it whenever cross-session
  continuity matters, even if the user does not name Hindsight.
compatibility: Requires the Hindsight CLI and a repository registered in ~/.hindsight/coding-agent.json.
metadata:
  category: ops
---

# Hindsight Memory

## Overview

Hindsight is the shared, long-term memory complement to the current conversation and repository
files. It stores durable project knowledge in the bank mapped to the current repository. The
repository-to-bank mapping is local configuration, never skill content.

## Usage

Use this skill before a non-trivial task when a prior decision, convention, failure, or procedure
may matter. The coding-agent integration retains the current session automatically; explicit
ingestion is only for an external source that future work on the same bank needs.

## Steps

1. Resolve the repository root with `git rev-parse --show-toplevel`.
2. Read only the `mapPathToBank` entry for that root from `~/.hindsight/coding-agent.json`; if no
   bank is mapped, continue without Hindsight and do not invent one.
3. Recall relevant context through the configured Hindsight MCP tools before deciding or changing
   code. Treat recalled memory as context to verify, not as authority over the repository or user.
4. Do not retain the current conversation manually: the coding-agent integration ingests it at the
   end of the session. Retain external documents or an out-of-band decision explicitly only when it
   is not already represented by that transcript.
5. Use reflect only for a genuinely complex,
   memory-dependent decision. State when an answer is inferred from memory.

## Gotchas

- **Using the current directory instead of the Git root** - subdirectories then miss their mapping;
  resolve the root before looking up the bank.
- **Treating recalled text as a source of truth** - stale or incorrect memories can override current
  code; verify it against the repository, documentation, or the user before acting.
- **Retaining credentials or personal data** - the bank is shared durable storage; redact secrets and
  avoid personal information before retaining anything.
- **Retaining the session a second time** - duplicated memories distort later recall; rely on the
  coding-agent integration for the current conversation.

## Constraints

- Never read, print, retain, or expose the API token from `~/.hindsight`.
- Never create a bank or mapping implicitly; configuration remains an explicit user action.
- Never manually retain the current conversation or duplicate a fact already ingested from it.
- Never retain a claim that has not been observed or clearly label it as an inference.
- Never use Hindsight as a replacement for current repository files, tests, or user instructions.
