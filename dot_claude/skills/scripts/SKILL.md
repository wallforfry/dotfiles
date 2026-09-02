---
name: scripts
description: >
  Create and maintain portable shell scripts in the dotfiles repository. Use when adding or editing
  a chezmoi run_ script, a helper under scripts/, a shebang, error handling, or tool installation.
  Make sure to use it whenever a change touches an executable in that repository, even if it is
  called a helper or a hook.
metadata:
  category: ops
---

# Scripts

## Overview

Shell scripts in the dotfiles repository run on three very different hosts: macOS with Homebrew,
Linux with a package manager, and a Synology NAS with neither. A script that assumes one of them
breaks the bootstrap of the others, and the bootstrap is exactly when nothing else is available to
fix it. Portability here is not a style preference - it is the feature.

Two kinds of script live in the repository:

- **`run_onchange_*.sh.tmpl`** - chezmoi hooks, executed during `chezmoi apply`. Rendered as
  templates, re-run when their rendered content changes, and deployed to no destination.
- **`scripts/*.sh`** - one-shot maintenance run by hand. Listed in `.chezmoiignore`, never deployed.

## Steps

1. Read the neighbouring script before changing names, arguments, or exit codes, and check every
   caller. `run_onchange_before_install-tools.sh.tmpl` is the reference for host degradation.
2. Use `#!/bin/sh` with `set -eu` and POSIX syntax for anything that may run during bootstrap: the
   NAS has no bash beyond Entware, and macOS still ships bash 3.2. Use `#!/usr/bin/env bash` with
   `set -euo pipefail` only for a script that never runs before the tooling is in place, and say why.
3. Detect capability, never platform, with `command -v <tool> >/dev/null`. Branch on `.chezmoi.os`
   in the template only when the difference is genuinely per-OS, such as Homebrew.
4. Degrade instead of failing. A missing optional tool prints one warning to stderr, names the
   consequence, and lets the script continue - `⚠️ starship non installé, prompt zsh par défaut`. Only
   a missing prerequisite of a later step is fatal.
5. Never write to `/tmp`: it is mounted `noexec` on DSM, so an extracted binary is unusable there.
   Use `mktemp -d "$HOME/.cache/<name>.XXXXXX"` with a `trap 'rm -rf "$tmp"' EXIT`.
6. Pin every downloaded version in a variable *and* in the header comment block. The header is what
   changes the rendered content, which is what makes chezmoi re-run the script - a version bumped
   only in the variable still changes content, but the header is what makes the diff readable.
7. Detect architecture from `uname -m`, map `x86_64`→`amd64` and `aarch64|arm64`→`arm64`, and exit
   cleanly with a warning on anything else rather than downloading a wrong binary.
8. Verify: `sh -n <file>` (or `bash -n`) for syntax, `chezmoi execute-template < <file>` to inspect a
   rendered `.tmpl`, and `chezmoi diff` for the effect. Say which OS you exercised.

## Gotchas

- **`&&` where `set -e` bites** - `command -v zsh >/dev/null && echo installed` fails the whole script
  when zsh is absent. Use an `if` block for any test whose false branch is acceptable.
- **`VAR=value sudo …`** - sudo purges the environment, so the assignment is lost. Use
  `sudo env VAR=value …`.
- **Assuming GNU flags** - `sed -i`, `readlink -f`, `date -d` and `grep -P` differ or are absent on
  macOS and DSM. Prefer portable syntax; branch explicitly when there is no portable form.
- **Bash 4 features in a bootstrap script** - associative arrays, `mapfile`, `${var,,}` are absent
  from macOS's bash 3.2 and from `sh`.
- **A `run_onchange_` script that is not idempotent** - it re-runs on every content change, including
  a comment fix. Guard every effect with a `command -v` or an existence test.
- **`exit 1` in a `run_` script for an optional tool** - it aborts `chezmoi apply` and leaves the home
  directory half-configured. Warn and `exit 0`.
- **A secret in a script** - instruct the operator to fetch it, or read it from the environment.
  Secrets belong in `encrypted_private_dot_secrets.age`.

## Constraints

- Bootstrap scripts use `#!/bin/sh` and POSIX syntax; `bash` requires a stated reason.
- Always `set -eu` (`set -euo pipefail` in bash).
- Never write executables under `/tmp`; use `$HOME/.cache` with a cleanup trap.
- Never ignore a command failure that affects correctness; never make an optional tool fatal.
- Never assume a package manager, Homebrew, sudo, or GNU coreutils are present.
- User-facing lines go to stdout for success and stderr for warnings, in French, matching the
  existing emoji-prefixed style.
- Do not rename a script or change its interface without checking all callers.
