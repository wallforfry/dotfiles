---
name: agent-instructions
description: >
  Maintain coding-agent instructions and their discovery paths. Use when editing AGENTS.md,
  CLAUDE.md, a harness file, an agent rule or an instruction projection. Make sure to use it
  whenever agent guidance or its deployment changes, even if the request names only one agent or
  looks like a wording fix.
metadata:
  category: dev
---

# Agent Instructions

## Overview

Instructions are paid for on every task: an always-loaded file enters the context whether or not it
is relevant. So the question before any edit is not "is this rule true?" but "does every task need
it?" - what is conditional belongs to a skill, which loads on demand.

The second failure mode is duplication. One canonical source, projections around it: an edit to a
copy diverges silently, and the agent reading the stale copy is the one that will not know.

## Usage

Use this skill for any change to global or project instructions, to an agent rule file, or to the
configuration that exposes them. For example: "add a rule about X to AGENTS.md", "why is CLAUDE.md
not picking this up", "move this section into a skill".

In this repository the canonical sources are `harness/AGENTS.md` (technical rules),
`harness/SOUL.md` (voice) and `harness/USER.md` (user preferences); everything else is a
projection. Read [references/maintenance.md](references/maintenance.md) for the full map and the
policy.

## Steps

1. Read `references/maintenance.md` completely.
2. Decide where the rule belongs: always-loaded only if every task needs it, otherwise a skill.
3. Locate the canonical source and edit it there, never a projection or a deployed copy.
4. Enumerate every consumer of that source - importer, template, installer, index, verification
   barrier - and update each in the same commit.
5. Verify the deployed effect: `chezmoi diff` for the destination, `chezmoi execute-template` for a
   template read in isolation, `bash scripts/verify.sh` for the barrier.
6. State which profiles you exercised. A rule reached through `.profile` has two renderings, and the
   machine you are on only proves one.

## Gotchas

- **Adding a conditional procedure to an always-loaded file** - every future task pays its context
  cost, including the ones it cannot help. Put it in a skill and let the description route to it.
- **Editing the projection instead of the source** - `dot_claude/AGENTS.md.tmpl` is a one-line
  `include` of `harness/AGENTS.md`. Content written there is content that exists twice.
- **Editing the deployed file** - `~/.claude/AGENTS.md` is a chezmoi destination; the next
  `chezmoi apply` reverts it. The checkout is the only place to write.
- **Assuming two hosts read the same file the same way** - a Claude Markdown import, a Codex command
  rule and a Cursor rule have different discovery semantics. Verify the consumer's contract instead
  of trusting a familiar file name.
- **Updating one consumer only** - the others keep the old behaviour and nothing announces it. The
  installer, the index and the barrier are consumers too.
- **Growing a file rather than reworking it** - a rule added beside a stale one leaves both in
  context, and the agent picks whichever it reads first. Refine or delete instead of accumulating.

## Constraints

- Keep one canonical source per piece of shared guidance; projections carry no content.
- Never put a conditional procedure in an always-loaded instruction file.
- Never claim a projection healthy without checking its real deployed destination.
- Never add guidance that does not change agent behaviour; delete stale guidance rather than
  qualifying it.
- Never edit a chezmoi destination under `~`; edit the source in the checkout.

## References

- [references/maintenance.md](references/maintenance.md) - the instruction map of this repository,
  the placement policy and the consumer checklist.
