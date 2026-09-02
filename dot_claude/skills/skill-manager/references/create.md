# Create a Skill

## Inputs

Gather before writing:

- the purpose, in one sentence;
- the concrete explicit and implicit triggers;
- inputs, outputs and hard constraints;
- the category, `dev` or `ops`;
- which of `references/`, `scripts/`, `assets/`, `evals/` the skill actually needs.

Ask one question at a time, and only when the answer changes the skill's behaviour. Do not invent a
domain rule to fill a section.

## Procedure

1. Read `conventions.md` completely.
2. Normalise the slug and check it against the `name` constraint.
3. Confirm `dot_claude/skills/<slug>/` does not exist. Never overwrite it.
4. Create the directory and only the resource subdirectories the inputs established.
5. Write `SKILL.md` from the template below.
6. Add `license`, `compatibility` or `allowed-tools` only when there is a real value to put in them.
7. Put executable shell with positional placeholders in `scripts/`, prefixed `executable_`.
8. Route scoped sibling references from `## Steps` when behaviour differs.
9. Run `sync-index`, then `bash scripts/verify.sh`, then `doctor <slug>`.
10. Read `chezmoi diff` for the new files: it is the only proof of where they land.

## Minimal template

```markdown
---
name: <slug>
description: >
  <Distinguishing case first>. Use when <concrete conditions>. Make sure to use it whenever
  <implicit cases>, even if <the domain is never named>.
metadata:
  category: dev | ops
---

# <Title>

## Overview

<Purpose, boundary and core principle, in two to four sentences.>

## Usage

<Invocation, arguments, and two or three typical cases.>

## Steps

1. <First complete action.>
2. <Second complete action.>
3. <Verification action.>

## Gotchas

- **<Specific cause>** - <consequence and correction>.
- **<Specific cause>** - <consequence and correction>.
- **<Specific cause>** - <consequence and correction>.

## Constraints

- <Hard must or must-not rule.>
- <Hard must or must-not rule.>
- <Hard must or must-not rule.>
```

## Completion

Creation is complete when `doctor <slug>` passes, `scripts/verify.sh` is green, the skill is indexed
under its category, a second `sync-index` changes no byte, and `chezmoi diff` shows the expected
destination.

## Constraints

- Never overwrite an existing skill.
- Never create an unused resource directory.
- Never emit an empty optional frontmatter field.
- Never add a chezmoi attribute to `SKILL.md` or a reference.
- Always run `doctor`, `sync-index` twice and `verify.sh` before declaring the skill done.
