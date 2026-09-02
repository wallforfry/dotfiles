# Skill Conventions

Source of truth for every skill under `dot_claude/skills/`. It separates the Agent Skills standard
from this repository's stricter local rules.

## 1. Progressive disclosure

A skill exposes three levels of context:

1. frontmatter `name` and `description`, always loaded for discovery;
2. `SKILL.md`, loaded on activation and kept under 500 lines;
3. `references/`, `scripts/`, `assets/`, `evals/`, loaded or executed only when needed.

Keep the procedure and the routing in `SKILL.md`. Move detailed variants, long tables, reusable
commands and fixtures into a resource directory. The same rule governs the repository's instruction
files: what is conditional belongs to a skill, not to an always-loaded file.

## 2. Frontmatter

The standard allows exactly six top-level fields:

```yaml
---
name: example-skill
description: >
  Handle a specific reusable workflow. Use when its concrete trigger occurs. Make sure to use it
  whenever the implicit case appears, even if the workflow is never named.
license: MIT
compatibility: Requires an authenticated `gh`.
allowed-tools: Read Grep
metadata:
  category: ops
---
```

Only `name` and `description` are required by the standard. This repository additionally requires
`metadata.category`.

| Field | Type | Constraint |
| --- | --- | --- |
| `name` | string | 1-64 lowercase letters, digits and hyphens; no leading, trailing or double hyphen; equals the directory name |
| `description` | string | non-empty, at most 1024 characters, local target under 400 |
| `license` | string | optional |
| `compatibility` | string | optional, non-empty when present, at most 500 characters; names the tools the skill assumes |
| `allowed-tools` | string | optional, space-separated; experimental, support varies by host |
| `metadata` | map | string keys to string values; `category` is `dev` or `ops` |

`disable-model-invocation`, `user-invocable`, `argument-hint`, `model`, `paths`, `hooks` and a
top-level `category` are host-specific or non-standard. Do not add one: this collection is written
for Claude Code but kept portable, and `scripts/verify.sh` reads `metadata.category` and nothing
else. Do not move an unknown field under `metadata` on your own either; ask, since it changes what
the field means.

### Categories

| Category | Scope |
| --- | --- |
| `dev` | writing, reviewing and deciding on code: review, architecture, decision records |
| `ops` | the machine and the agent itself: tooling, configuration, procedures, skill and instruction management |

Two categories, because the repository has two. Add a third only when a skill genuinely fits
neither, and update `scripts/verify.sh` and `sync-index` in the same commit - the category list
lives in three places and a fourth would be a fourth thing to forget.

### Description format

The description is the router: it is the only part of a skill that is always in context. Put the
distinguishing case first and follow this pattern:

```text
<What it handles>. Use when <concrete conditions>. Make sure to use it whenever <implicit cases>,
even if <the domain is never named>.
```

Stay under 400 characters: host skill lists truncate long descriptions well before the standard's
1024-character ceiling. Avoid generic openings, passive labels and appended keyword dumps.

## 3. Layout

```text
dot_claude/skills/<slug>/
  SKILL.md
  references/   optional detailed guidance
  scripts/      optional executable logic, `executable_` prefixed
  assets/       optional output templates
  evals/        optional activation scenarios
```

Create only the directories the skill actually needs. Never add a chezmoi attribute to `SKILL.md` or
to a reference: attributes rename the destination. `scripts/` entries that must be executable at
their destination carry `executable_`, per the repository's `AGENTS.md` layout table.

There is no second collection and no adapter to keep in sync: `dot_claude_pro/symlink_skills` points
the `pro` profile at the same deployed directory. A skill is therefore never registered anywhere,
and never duplicated for another profile.

## 4. Reference segmentation

Split a reference by scope only when behaviour genuinely differs, and name the siblings
`<topic>-<scope>.md`. Identical rules stay in one shared file. When scoped siblings exist, route
them explicitly from `SKILL.md`:

```markdown
1. Identify the target.
2. For a `run_onchange_` script, read `references/hooks.md`.
3. For a helper under `scripts/`, read `references/helpers.md`.
```

A two-word filename with no same-topic sibling is an ordinary reference, not a scoped one.

## 5. Body

Required sections, in this order:

1. frontmatter
2. one H1 title
3. `## Overview`
4. `## Usage`
5. `## Steps` or `## Workflow`
6. `## Gotchas`
7. `## Constraints`

`## References` and examples are optional. `Gotchas` and `Constraints` each carry at least three
concrete entries. Each gotcha names a cause, its consequence and its correction; generic advice a
capable agent already knows is noise. This is a local quality bar, not a portability claim.

## 6. Writing principles

- Record the non-obvious reason inside the skill, where it loads on demand.
- Give one default, and say when the alternative applies.
- Write numbered actions, not abstract declarations.
- One job per skill.
- One strong example beats three variants.
- Front-load the search terms in the description and the overview.
- Put repeated or deterministic shell in `scripts/`.
- Prose is English; typography follows `harness/SOUL.md` - no em dash, no middle dot.

## 7. Positional shell placeholders

Some hosts template a `SKILL.md` body as a slash command before an agent reads it, so `$0` to `$9`,
`$@` and `$ARGUMENTS` can be substituted by invocation arguments. The rewritten command may still
run, and silently do the wrong thing.

- Executable shell with a positional placeholder goes in `scripts/`.
- `SKILL.md` calls the script instead of reproducing its internals.
- A literal token in prose is escaped with one backslash before the dollar sign, never two.
- `references/` and `scripts/` are read as files and are not templated.

## 8. Activation routing

The description routes by default, and an absent router rule is never a finding. Add one only after
identical realistic prompts, run at least three times without it, reproduce a missed or wrong
activation. Keep the prompts and the results, add the smallest rule that distinguishes the sibling
skills, then rerun the same prompts.

## 9. README index

`dot_claude/skills/README.md` is derived from `name`, the folded `description` and
`metadata.category`. Never edit a row by hand. See [sync-index.md](sync-index.md) for the generation
rules, and [evals.md](evals.md) for the optional scenario files.

## 10. What the barrier already checks

`scripts/verify.sh`, Skills section, is the mechanical floor:

- `SKILL.md` exists and its `name` equals the directory;
- a `description` exists and contains `Use when`;
- `metadata.category` is `dev` or `ops`;
- the skill is listed in the README under the section matching its category;
- the row count matches the directory count, and no row lacks a directory.

Everything else in this file is judgement, and belongs to `doctor`. A red barrier is a FAIL that no
judgement overrides; a green barrier is not a clean audit.
