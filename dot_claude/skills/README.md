# Skills

Source of truth: `dot_claude/skills/` in the dotfiles repository, deployed by chezmoi to
`~/.claude/skills` and shared with `~/.claude_pro/skills` through a symlink. Edit the source, never
the deployed copy.

## Conventions

- One skill per directory, each with a `SKILL.md`.
- Optional subdirectories: `references/`, `assets/`, `scripts/`, `evals/`.
- Skill content is written in English (see the root `AGENTS.md` language rule).
- Frontmatter carries `metadata.category`, either `dev` or `ops`.
- The tables below are derived from frontmatter. Never edit a row by hand: run
  `/skill-manager sync-index`.

## Dev

| Skill | Description |
| --- | --- |
| `adr` | Write, amend or supersede an architecture decision record. |
| `agent-instructions` | Maintain coding-agent instructions and their discovery paths. |
| `merge-verdict` | Deliver a merge verdict on an open pull request, yours or another author's. |

## Ops

| Skill | Description |
| --- | --- |
| `do-nothing-script` | Turn a repeated manual procedure into a do-nothing script (Slimmon): one function per step, printed then awaiting the operator, automated one step at a time. |
| `handoff` | Hand the current work to a fresh session instead of letting the context compact. |
| `neovim` | Maintain the dotfiles repository's Neovim and LazyVim configuration. |
| `scripts` | Create and maintain portable shell scripts in the dotfiles repository. |
| `skill-manager` | Manage the skills of the dotfiles repository: create, doctor, fix, cross-check, and rebuild their README index. |

## Origin

`handoff`, `do-nothing-script`, `merge-verdict`, `skill-manager` and the `agent-handoff` Stop hook
are adapted from <https://github.com/SebastienElet/dotfiles> (BSD 2-Clause, Copyright (c) 2014
Sébastien ELET), whose `LICENSE` requires that this notice be retained. `scripts`, `neovim` and
`adr` are written for this repository and keep only the shape of their originals.
