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
| `private_dot_gnupg/` | `gpg-agent.conf` only, deployed to `~/.gnupg` with restricted permissions |
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
  `references/`, `assets/`, `scripts/`, `evals/` subdirectories. Frontmatter carries
  `metadata.category` (`dev` or `ops`), from which `dot_claude/skills/README.md` is derived.
- `dot_claude_pro/` contains only `symlink_` entries pointing into `~/.claude`, so the work
  profile shares one source instead of a second copy. It is ignored outside the `pro` profile.
- **Never add the `exact_` attribute to `dot_claude/` or `private_dot_gnupg/`.** `~/.claude` holds
  live state - sessions, projects, plugins - and `~/.gnupg` holds the keyring and agent sockets.
  chezmoi would delete whatever the source does not carry.

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

Verify locally before delivering: CI repeats the same barrier, it does not replace it, and it only
reports once the change is pushed. **`bash scripts/verify.sh` is the
mechanical barrier**: shell syntax and template rendering across three profile combinations, skill
frontmatter and its README table, the ADR index, sensitive names in the tree and in unpushed commit
messages, and encryption of every `.age` fragment. It reports counts and exits non-zero on the first
failing check. Run it before every commit.

A green barrier is not correctness. Before committing or pushing a change whose deployed effect
matters, delegate the reading of `chezmoi diff` to the `dotfiles-reviewer` subagent: it judges what
the script cannot decide.

What it cannot decide, verify by hand:

- `chezmoi diff` - the deployed effect of the change. Read it in full; the script says nothing about
  whether the change is correct, only that it is well-formed.
- `chezmoi execute-template < <file>` - render a template in isolation when the diff is unclear.
- `bash -n <script>` - syntax of any shell script touched.
- Templates using `.profile` must be checked against both values, not just the current machine's.
- Say which OS you exercised. This repository targets macOS, Linux and Synology DSM; DSM notably
  mounts `/tmp` with `noexec`, which is why `scriptTempDir` is set in `.chezmoi.toml.tmpl`.

`bash scripts/harness-audit.sh` is the measurement counterpart, run by hand and never in CI: the
always-loaded byte total, the real activation count of every skill and subagent, the violation rate
of the two observable rules before and after their introduction, and the mutation-detection score of
the barrier itself. It exits non-zero when an injected defect slips past `verify.sh` or when a
measurement could not be made. The `harness-audit` skill carries how to read its counts.

`.github/workflows/verify.yml` runs the barrier on Linux and then a real `chezmoi apply` on
`ubuntu-latest` and `macos-latest` for both profiles, on push to `main`, on pull requests and on
demand ([ADR-020](docs/adr/020-verification-en-ci.md)). It needs the `AGE_KEY` secret and never
covers DSM, which has no runner: a DSM-specific change stays a manual check.

**No command added there may write a target's rendered content.** `apply --verbose` and
`chezmoi diff` emit a unified diff of what they write, which on a public runner publishes the
cleartext of every `age` fragment. `verify.sh` refuses both in `.github/`; read the deployed effect
with `chezmoi status`, which prints paths only.

## Secrets

**This repository is public.** Every file in it, and every commit message, is world-readable
([ADR-016](docs/adr/016-depot-public-sensible-chiffre.md)).

- Nothing sensitive goes in, in clear or in prose: no hostname, IP address, login, employer, client
  or internal project name, in a template, a script, a log, a commit message, or documentation.
- **Nor the machine's third-party stack.** The list of installed tools - PaaS, CI, secret-manager
  CLIs - maps the infrastructure its owner works on. Name only the tools this repository depends on,
  because a deployed file requires them ([ADR-007](docs/adr/007-outillage-run-onchange.md)).
- What is sensitive and still needed at deploy time is encrypted with `age`: `~/.secrets`,
  `~/.ssh/config.d/nas.conf`, `~/.config/zsh/pro.zsh`, `~/.config/zsh/pro.zprofile`,
  `~/.config/git/pro.gitconfig`, `~/.claude/CONTEXT.md`, `~/.config/dotfiles/sensible.txt` - the
  list of names `scripts/verify.sh` forbids, which is itself the data to protect. Add one with
  `chezmoi add --encrypt <path>`, which the operator runs: it needs the passphrase, so an agent
  cannot.
- Public files load those fragments without naming what is in them - a `[ -f … ] && source …`, a git
  `[include]`, an `@CONTEXT.md` guarded by the `pro` profile.
- Fragments specific to the `pro` profile are listed in `.chezmoiignore` under
  `{{ if ne .profile "pro" }}`, so other machines never prompt for a passphrase they do not need.
- An encrypted file is still downloadable by anyone. The passphrase is the only barrier, and it is
  attackable offline: never weaken it, never write it down here.
