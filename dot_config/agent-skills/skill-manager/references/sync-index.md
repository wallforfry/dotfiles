# Sync the README Index

## Purpose

Rebuild `dot_config/agent-skills/README.md` from the skills' frontmatter. The README is a projection,
never a second place where a description lives.

## Procedure

1. List every immediate directory under `dot_config/agent-skills/` containing a `SKILL.md`,
   including one
   not yet tracked by git. Report a directory whose `SKILL.md` is missing.
2. Read `name`, the complete folded `description` and `metadata.category`.
3. Normalise the folded description's whitespace to single ASCII spaces.
4. Take its first sentence: up to the first period followed by whitespace or by the end of the
   string, period included. No arbitrary character limit.
5. Escape any `|` in that sentence.
6. Group by category, `Dev` then `Ops`, and sort slugs bytewise inside each section.
7. Rewrite the file from the template below. Omit an empty section.
8. Report what was added, removed, moved or refused.
9. Run the procedure a second time and require byte-identical output.

A skill whose `metadata.category` is missing or invalid is a `verify.sh` failure. Fix the
frontmatter rather than inventing a section for it: the index must not paper over it.

## Template

```markdown
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
| `<slug>` | <first sentence of the description>. |

## Ops

| Skill | Description |
| --- | --- |

## Origin

<the attribution paragraph, preserved verbatim>
```

## Formatting

There is no formatter in this repository, so the generation defines the bytes: one space on each
side of every `|`, the separator row is exactly `| --- | --- |`, no column padding, no trailing
whitespace, one final newline. Never align columns: alignment depends on the longest row and turns
every addition into a diff of the whole table.

The `## Origin` section is authored prose, not derived. Carry it over unchanged, and update it by
hand when a skill's provenance changes.

## Idempotence check

```bash
shasum -a 256 dot_config/agent-skills/README.md > "${TMPDIR:-/tmp}/skills-index.sha256"
```

Regenerate, then:

```bash
shasum -a 256 -c "${TMPDIR:-/tmp}/skills-index.sha256"
```

The check must pass. Identical frontmatter must always produce identical bytes.

## Constraints

- Never edit a table row by hand.
- Never read the category anywhere but `metadata.category`.
- Never truncate a description at an approximate character count.
- Never keep a row whose directory no longer exists, and never drop a directory from the index.
- Never rewrite the `## Origin` attribution while regenerating.
- Always verify that a second generation changes no byte, then run `bash scripts/verify.sh`.
