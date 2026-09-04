---
name: skill-manager
description: >
  Manage the skills of the dotfiles repository: create, doctor, fix, cross-check, and rebuild their
  README index. Use when creating or changing any skill, including a requested behaviour change to a
  passing one. Make sure to use it whenever a SKILL.md, a frontmatter field or the skills index is
  edited, even if the request names only a typo.
compatibility: >
  The dotfiles repository checkout, since every path is relative to `dot_claude/skills/`, and
  `bash scripts/verify.sh` for the mechanical part of the audit.
metadata:
  category: ops
---

# Skill Manager

## Overview

`dot_claude/skills/` is the only skill collection of this repository; chezmoi deploys it to
`~/.claude/skills` and the `pro` profile shares it through `dot_claude_pro/symlink_skills`. A skill
therefore needs no registration step, and editing a deployed copy is always a mistake.

Five operations: scaffold a skill, audit one or all of them, apply a justified change, report
inter-skill inconsistencies, and rebuild the derived README index. `scripts/verify.sh` already
decides the mechanical part - frontmatter `name`, category, index membership. This skill covers what
a script cannot decide: whether a description routes, whether a body is complete, whether two skills
collide.

## Usage

```text
/skill-manager create <slug>     - scaffold a new skill
/skill-manager doctor [slug]     - audit one skill, or the whole collection
/skill-manager fix <slug>        - apply findings or a requested evolution
/skill-manager cross-check       - report inter-skill inconsistencies, write nothing
/skill-manager sync-index        - rebuild the deterministic README index
```

`<slug>` is kebab-case and equals the directory name. `doctor` without a slug audits every
directory. `fix` without a slug asks which skill should change rather than guessing.

The boundary with `agent-instructions`: that skill owns the always-loaded files - `harness/`, the
projections, the adapters - and decides whether a rule belongs in one. This skill owns the inside of
a skill and the derived index. A section moved out of an always-loaded file arrives here.

## Steps

1. Identify the operation: `create`, `doctor`, `fix`, `cross-check` or `sync-index`.
2. Read [references/conventions.md](references/conventions.md) completely, before any write.
3. Read the operation's own reference below and follow it exactly.
4. Run `bash scripts/verify.sh` and treat its Skills section as the mechanical floor: a red barrier
   is a FAIL no judgement overrides.
5. For `cross-check`, present the report and stop. Every write goes through a later `fix`.
6. After `create`, `fix`, a rename or a deletion, run `sync-index` and require that a second run
   changes no byte.

## Gotchas

- **Editing the deployed copy** - `~/.claude/skills/<slug>/SKILL.md` is a chezmoi destination, and
  the next `chezmoi apply` silently reverts it. Edit `dot_claude/skills/<slug>/` in the checkout.
- **Adding a chezmoi attribute to a skill file** - `executable_`, `private_` and their siblings are
  interpreted by chezmoi and rename the destination. Skill files carry no attribute; only
  `scripts/` entries may need `executable_`.
- **Trusting a green barrier as a clean audit** - `verify.sh` checks name, category and index
  membership. It says nothing about a description that routes to the wrong skill, a missing
  `Constraints` section, or two skills that overlap.
- **Reformatting the README by hand** - the table is derived. A manual edit survives until the next
  `sync-index` and then looks like a regression; change the frontmatter instead.
- **Applying a cross-check finding inline** - cross-check is read-only, so its report can be trusted
  as an observation. Record it, then run `fix <slug>` as a separate operation.
- **Adding an activation router by reflex** - the description is the router. An absent router rule
  is never a finding, and a rule that exists to stop a skill activating means the skill is wrong.

## Constraints

- Always read `references/conventions.md` before writing any skill file.
- Never overwrite an existing `SKILL.md` without an explicit create, fix, rename or delete request.
- Never modify a file during `doctor` or `cross-check`.
- Never edit a README table row by hand; regenerate it with `sync-index`.
- Never modify more than one skill in one `fix` operation.
- Never claim an audit green while `scripts/verify.sh` is red.
- Never add an activation router without repeated behavioural evidence.
- Write every skill and reference in English, per the repository's language rule.

## References

- [references/conventions.md](references/conventions.md) - the frontmatter, body and index
  conventions. Read first, for every operation.
- [references/create.md](references/create.md) - scaffolding procedure and minimal template.
- [references/doctor.md](references/doctor.md) - the audit, its status model and report format.
- [references/fix.md](references/fix.md) - correction order and regression handling.
- [references/cross-check.md](references/cross-check.md) - the read-only inter-skill detectors.
- [references/sync-index.md](references/sync-index.md) - deterministic README generation.
