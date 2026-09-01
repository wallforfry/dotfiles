# Dotfiles Agent Instructions

Single source of truth for every coding agent working in this repository.

> **Conflict rule:** if an agent-specific adapter disagrees with this document, `AGENTS.md` wins.

The user-scope instructions this repository deploys (`harness/`) also apply here, since they are
installed in `~/.claude`. This file only adds what is specific to the repository.

## Scope

- **Prefer small iterations.** Do only what was asked; avoid expanding to every similar item.
- **Keep changes minimal.** Reuse existing structures and ask when the intended scope is ambiguous.
- **Avoid broad refactors.** Do not migrate or reorganise unrelated configuration by default.

## Layout

chezmoi manages this repository: **a source path mirrors its destination under `$HOME`**. Deduce
where a new file goes from where it must land.

| Path | Contents |
| --- | --- |
| `dot_<name>` | deployed to `~/.<name>` |
| `dot_config/` | deployed to `~/.config/` |
| `dot_local/bin/` | executables deployed to `~/.local/bin`, `executable_` and extensionless |
| `private_dot_ssh/` | deployed to `~/.ssh` with restricted permissions |
| `harness/` | agent instructions, agnostic of any single agent - not deployed as such |
| `scripts/` | one-shot maintenance scripts, not deployed |
| `docs/`, `README.md` | documentation, not deployed |

Attributes carry meaning and are not decoration: `private_` restricts permissions, `encrypted_`
stores the file `age`-encrypted, `symlink_` makes the destination a symlink whose target is the
file's rendered content, `executable_` sets the executable bit, `run_onchange_` re-runs a script
when its rendered content changes.

Everything not deployed is listed in `.chezmoiignore` - which is itself a template, so an entry can
be conditioned on `.profile` or on the OS.

## Agent Configuration

- `harness/AGENTS.md`, `harness/SOUL.md` and `harness/USER.md` are the canonical instruction
  sources: technical rules, agent voice, user preferences respectively. Edit these.
- `dot_claude/{AGENTS,SOUL,USER}.md.tmpl` are one-line projections - `{{ include "harness/…" }}` -
  and `dot_claude/CLAUDE.md` is the Claude adapter that imports them. Never move content into them.
- `dot_claude/skills/<slug>/SKILL.md` holds the skills. One skill per directory; optional
  `references/`, `assets/`, `scripts/` subdirectories.
- `dot_claude_pro/` contains only `symlink_` entries pointing into `~/.claude`, so the work
  profile shares one source instead of a second copy. It is ignored outside the `pro` profile.
- **Never add the `exact_` attribute to `dot_claude/`.** `~/.claude` holds live state - sessions,
  projects, plugins - that chezmoi would delete.

## Architecture Decisions

- `docs/adr/` records the structural decisions of this repository, indexed in `docs/adr/README.md`.
  Only decisions still in force are recorded.
- Never contradict an ADR silently. Either follow it, or state the conflict, then deliver what was
  asked along with the ADR that would need superseding.
- Routine changes - adding a tool to the install script, moving a pinned version, editing a skill -
  need no ADR. `docs/adr/README.md` carries the test, and the `adr` skill the procedure.

## Language

- Documentation and prose (`README.md`, `docs/`) are written in French: they record reasoning for a
  French-speaking author.
- `AGENTS.md`, `harness/AGENTS.md` and every `SKILL.md` are written in English. They are interfaces
  read by agents, kept portable across agents and hosts.
- `harness/SOUL.md` and `harness/USER.md` are French by exception: they describe a voice and a person
  rather than an interface.
- The exception covers prose only. Identifiers, commands, paths, and quoted commit bodies stay
  verbatim.

## Verification

Nothing here runs in CI, so verify locally before delivering:

- `chezmoi diff` - the deployed effect of the change. Read it in full; it is the only barrier.
- `chezmoi execute-template < <file>` - render a template in isolation when the diff is unclear.
- `bash -n <script>` - syntax of any shell script touched.
- Templates using `.profile` must be checked against both values, not just the current machine's.
- Say which OS you exercised. This repository targets macOS, Linux and Synology DSM; DSM notably
  mounts `/tmp` with `noexec`, which is why `scriptTempDir` is set in `.chezmoi.toml.tmpl`.

## Secrets

**This repository is public.** Every file in it, and every commit message, is world-readable
([ADR-016](docs/adr/016-depot-public-sensible-chiffre.md)).

- Nothing sensitive goes in, in clear or in prose: no hostname, IP address, login, employer, client
  or internal project name, in a template, a script, a log, a commit message, or documentation.
- What is sensitive and still needed at deploy time is encrypted with `age`: `~/.secrets`,
  `~/.ssh/config.d/nas.conf`, `~/.config/zsh/pro.zsh`, `~/.config/zsh/pro.zprofile`,
  `~/.config/git/pro.gitconfig`, `~/.claude/CONTEXT.md`. Add one with
  `chezmoi add --encrypt <path>`, which the operator runs: it needs the passphrase, so an agent
  cannot.
- Public files load those fragments without naming what is in them - a `[ -f … ] && source …`, a git
  `[include]`, an `@CONTEXT.md` guarded by the `pro` profile.
- Fragments specific to the `pro` profile are listed in `.chezmoiignore` under
  `{{ if ne .profile "pro" }}`, so other machines never prompt for a passphrase they do not need.
- An encrypted file is still downloadable by anyone. The passphrase is the only barrier, and it is
  attackable offline: never weaken it, never write it down here.
