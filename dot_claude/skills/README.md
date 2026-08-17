# Skills

Source of truth: `dot_claude/skills/` in the dotfiles repository, deployed by chezmoi to
`~/.claude/skills` and shared with `~/.claude_septeo/skills` through a symlink. Edit the source, never
the deployed copy.

One skill per directory, each with a `SKILL.md`. Optional `references/`, `assets/` and `scripts/`
subdirectories. Skill content is written in English (see the root `AGENTS.md` language rule).

| Skill | Description |
| --- | --- |
| `adr` | Write, amend or supersede an architecture decision record. |
| `do-nothing-script` | Turn a repeated manual procedure into a do-nothing script: one function per step, printed then awaiting the operator, automated one step at a time. |
| `handoff` | Hand the current work to a fresh session instead of letting the context compact. |
| `merge-verdict` | Deliver a merge verdict on an open pull request, yours or another author's. |
| `neovim` | Maintain this repository's Neovim and LazyVim configuration. |
| `scripts` | Create and maintain portable shell scripts in this repository. |

## Origin

`handoff`, `do-nothing-script`, `merge-verdict` and the `agent-handoff` Stop hook are adapted from
<https://github.com/SebastienElet/dotfiles> (BSD 2-Clause, Copyright (c) 2014 Sébastien ELET), whose
`LICENSE` requires that this notice be retained. `scripts`, `neovim` and `adr` are written for this
repository and keep only the shape of their originals.
