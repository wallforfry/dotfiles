---
name: adr
description: >
  Write, amend or supersede an architecture decision record. Use when a structural choice needs
  recording, when asked whether something deserves an ADR, when a change would contradict a recorded
  decision, or when an ADR index needs rebuilding. Make sure to use it whenever a decision's rationale
  is about to be written down, even if the term ADR is never used.
---

# ADR

## Overview

An ADR answers one question: why is it this way and not another way? What the code does is readable
in the code; what was rejected, and why, is readable nowhere else.

This skill is the procedure. The **convention is owned by the repository**, not by this file: each
repository's ADR directory carries its own README with the decision test, the section format, the
scope rule and the index. Read it first and follow it - where it disagrees with anything here, it
wins.

## Steps

1. **Locate the convention.** Look for `docs/adr/README.md`, `adr/README.md`, or a decision index
   named in the repository's `AGENTS.md`. Read it completely before writing anything.
   - If the repository has ADRs but no README describing the convention, follow the shape of the
     existing records and say in your report that the convention is undocumented.
   - If the repository has no ADR directory at all, do not invent one silently. Say so, and offer
     `~/.local/share/chezmoi/docs/adr/README.md` as a model to bootstrap from.

2. **Apply the repository's test before writing.** Most changes do not warrant an ADR. If the test
   fails, say which condition fails and stop - a rejected ADR is a normal outcome, and saying so is
   more useful than producing a record that dilutes the index.

3. **Ground every claim.** The `Contexte` section is evidence, not narration: the error message, the
   measurement, the commit that introduced or reverted the thing. Quote a commit body when you use
   it, and mark reconstruction as reconstruction. Never invent a motivation that sounds plausible -
   an ADR is read years later as fact.

4. **Get the rejected alternatives right.** This is the section that gives an ADR its value, and the
   one most often padded. Each alternative needs a real reason for rejection, and an alternative that
   was actually tried and reverted is the strongest entry there is - name its commits.

5. **State the cost.** `Conséquences` lists what the decision costs, not only what it buys. An ADR
   with no cost section is advertising, and a reader will not trust the rest of it.

6. **Update the index**, then run the repository's index-verification command and report its output.

7. **To change a decision**, follow the repository's scope rule. Where only in-force decisions are
   recorded - as in the dotfiles repository - that means writing a new ADR, carrying the old one's
   alternatives into its `Alternatives écartées`, and deleting the old file. Never edit an in-force
   ADR to reflect a change of mind: that rewrites history rather than recording it.

## Gotchas

- **Writing the ADR instead of proposing it** - recording a decision commits the author to it in
  writing. Propose the ADR and its content, and get agreement before adding it to the index, unless
  the request was explicitly to write one.
- **An ADR for a bugfix** - the commit body is the right medium. It becomes an ADR only when the fix
  reveals a durable platform constraint that will be re-litigated.
- **Renumbering or reusing a number** - the index, the other records and the git history all point at
  numbers. Take max + 1, always.
- **Contradicting an ADR silently while doing something else** - the failure this whole directory
  exists to prevent. When a task's obvious implementation conflicts with a record, say so, name the
  ADR that would need superseding, and deliver what was asked.
- **A decision phrased as an intention** - "we should prefer X" is not a decision. Write what is done,
  in the present indicative, so a reader can check the code against it.
- **Padding the alternatives** - three plausible-sounding rejections are worth less than one real
  one. Drop what you cannot justify.
- **Translating an existing record** - the language rule applies to new records; retranslating an old
  one loses the wording its conclusions were reached in, for no reader.

## Constraints

- Always read the repository's ADR convention before writing; never assume this file's shape.
- Never create an ADR without agreement, unless explicitly asked to write one.
- Never state a reconstructed motivation as a measured one.
- Never publish an ADR with an empty or padded `Alternatives écartées`.
- Never edit an in-force ADR to change its decision; supersede it.
- Never reuse or renumber an ADR number.
- Always update the index and run its verification in the same change.
