---
name: scripts
description: >
  Create and maintain portable bootstrap shell in the dotfiles repository. Use when adding or
  editing a chezmoi run_ script, a shebang, bootstrap error handling, tool installation, or the Go
  build hook. Make sure to use it whenever code runs before the repository's Go CLI is available.
metadata:
  category: ops
---

# Scripts

## Overview

The remaining shell in the dotfiles repository is the bootstrap boundary. It runs on three very
different hosts: macOS with Homebrew, Linux with a package manager, and a Synology NAS with neither.
A script that assumes one of them breaks the bootstrap of the others, exactly when the Go CLI is not
available to repair it. Portability here is the feature.

Three scripts live at this boundary:

- **`run_before_unlock-age-key.sh.tmpl`** unlocks encrypted state before chezmoi needs it.
- **`run_onchange_before_install-tools.sh.tmpl`** installs Go and the other required tools.
- **`run_onchange_after_build-dotfiles.sh.tmpl`** builds the Go CLI only after the compiler exists.

Post-bootstrap orchestration belongs in `cmd/dotfiles` or a cohesive package under `internal/`, not
in a new shell helper.

## Steps

1. Confirm that the code must run before the Go CLI exists. If not, implement it in Go and keep the
   shell boundary unchanged.
2. Read the neighbouring script before changing names, arguments, or exit codes, and check every
   caller. `run_onchange_before_install-tools.sh.tmpl` is the reference for host degradation.
3. Use `#!/bin/sh` with `set -eu` and POSIX syntax for bootstrap: the NAS has no bash beyond
   Entware, and macOS still ships bash 3.2.
4. Detect capability, never platform, with `command -v <tool> >/dev/null`. Branch on `.chezmoi.os`
   in the template only when the difference is genuinely per-OS, such as Homebrew.
5. Degrade instead of failing. A missing optional tool prints one warning to stderr, names the
   consequence, and lets the script continue - `⚠️ starship non installé, prompt zsh par défaut`. Only
   a missing prerequisite of a later step is fatal.
6. Never write to `/tmp`: it is mounted `noexec` on DSM, so an extracted binary is unusable there.
   Use `mktemp -d "$HOME/.cache/<name>.XXXXXX"` with a `trap 'rm -rf "$tmp"' EXIT`.
7. Pin every downloaded version once, in an authoritative variable. Add a comment only for an
   external or non-deducible constraint, such as an upstream incompatibility; the variable already
   changes the rendered content and makes chezmoi re-run the script.
8. Detect architecture from `uname -m`, map `x86_64`→`amd64` and `aarch64|arm64`→`arm64`, and exit
   cleanly with a warning on anything else rather than downloading a wrong binary.
9. Verify with `go run ./cmd/dotfiles verify`, render the template in isolation when needed, and
   inspect `chezmoi diff`. Say which OS and profiles you exercised.

## Gotchas

- **`&&` where `set -e` bites** - `command -v zsh >/dev/null && echo installed` fails the whole script
  when zsh is absent. Use an `if` block for any test whose false branch is acceptable.
- **`VAR=value sudo …`** - sudo purges the environment, so the assignment is lost. Use
  `sudo env VAR=value …`.
- **Assuming GNU flags** - `sed -i`, `readlink -f`, `date -d` and `grep -P` differ or are absent on
  macOS and DSM. Prefer portable syntax; branch explicitly when there is no portable form.
- **Moving bootstrap into Go** - the compiler and repository CLI do not exist on a new machine. Keep
  installation and compilation in POSIX shell, then invoke Go after the build succeeds.
- **A `run_onchange_` script that is not idempotent** - it re-runs on every content change, including
  a comment fix. Guard every effect with a `command -v` or an existence test.
- **`exit 1` in a `run_` script for an optional tool** - it aborts `chezmoi apply` and leaves the home
  directory half-configured. Warn and `exit 0`.
- **A secret in a script** - instruct the operator to fetch it, or read it from the environment.
  Secrets belong in `encrypted_private_dot_secrets.age`.

## Constraints

- Only bootstrap scripts use shell; post-bootstrap orchestration is implemented in Go.
- Bootstrap scripts use `#!/bin/sh` and POSIX syntax.
- Always `set -eu`.
- Never write executables under `/tmp`; use `$HOME/.cache` with a cleanup trap.
- Never ignore a command failure that affects correctness; never make an optional tool fatal.
- Never assume a package manager, Homebrew, sudo, or GNU coreutils are present.
- User-facing lines go to stdout for success and stderr for warnings, in French, matching the
  existing emoji-prefixed style.
- Comments record only external or non-deducible facts. Shebangs, tool directives and documentation
  required by tooling are outside this rule.
- Do not rename a script or change its interface without checking all callers.
