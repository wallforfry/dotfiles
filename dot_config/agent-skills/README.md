# Skills

Source of truth: `dot_config/agent-skills/` in the dotfiles repository, deployed by chezmoi to
`~/.config/agent-skills` and reached by every host - `~/.claude/skills`, `~/.claude_pro/skills`,
`~/.codex/skills` - through a symlink. Edit the source, never
the deployed copy.

## Conventions

- One skill per directory, each with a `SKILL.md`.
- Optional subdirectories: `references/`, `assets/` and `scripts/`.
- Skill content is written in English (see the root `AGENTS.md` language rule).
- Frontmatter carries `metadata.category`, either `dev` or `ops`.
- The tables below are derived from frontmatter. Never edit a row by hand: run
  `/skill-manager sync-index`.

## Dev

| Skill | Description |
| --- | --- |
| `adr` | Write, amend or supersede an architecture decision record. |
| `agent-instructions` | Maintain the always-loaded agent instruction files and their discovery paths. |
| `business-issue` | Investigate business needs and turn them into executable issues across trackers. |
| `merge-verdict` | Deliver a merge verdict on an open pull request, yours or another author's. |
| `obsolescence` | Decide whether obsolete code or data is deleted or migrated. |

## Ops

| Skill | Description |
| --- | --- |
| `containerized-mcp` | Run Docker-delivered MCP servers through reusable named containers. |
| `do-nothing-script` | Turn a repeated manual procedure into a do-nothing script (Slimmon): one function per step, printed then awaiting the operator, automated one step at a time. |
| `handoff` | Hand the current work to a fresh session instead of letting the context compact. |
| `harness-audit` | Measure harness cost, deployment lag, activation, adherence, and mutation detection. |
| `harness-reflection` | Turn a repeated agent failure into one evidence-backed harness change. |
| `i-have-adhd` | Shape agent output for a reader with ADHD. |
| `neovim` | Maintain the dotfiles repository's Neovim and LazyVim configuration. |
| `scripts` | Create and maintain portable bootstrap shell in the dotfiles repository. |
| `skill-manager` | Manage the skills of the dotfiles repository: create, doctor, fix, cross-check, and rebuild their README index. |
| `web-fetching` | Retrieve web pages or search results safely through escalating fetch tiers. |

## Origin

`handoff`, `do-nothing-script`, `merge-verdict`, `skill-manager`, `harness-reflection` and the
`agent-handoff` Stop hook are adapted from <https://github.com/SebastienElet/dotfiles>
(BSD 2-Clause, Copyright (c) 2014 Sébastien ELET), whose `LICENSE` requires that this notice be retained. `scripts`, `neovim` and
`adr` are written for this repository and keep only the shape of their originals.

`i-have-adhd` is adapted from <https://github.com/ayghri/i-have-adhd> (MIT, Copyright (c) 2026
Ayoub Ghriss). Its license requires that this notice be retained.
