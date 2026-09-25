---
name: writing-style
description: >
  Write or edit prose for human readers without machine-written tics, in French or English. Use when
  drafting or reviewing a README, docs, a PR description, an article, a post or a message sent on
  the user's behalf. Make sure to use it whenever text will be published or read by someone else,
  even if style is never mentioned.
metadata:
  category: dev
---

# Writing Style

## Overview

Rules for prose that a person reads: documentation, PR descriptions, articles, posts, messages. They
remove the patterns that make text read as machine-written: rejected-frame contrasts, inflated
vocabulary, decorative analogies, fake importance. Code, commit subjects and terminal reports keep
their own conventions from `harness/SOUL.md`; this skill applies on top of them, never against them.

Priority when rules collide: accurate, then clear, then specific, then natural. A style rule never
justifies an awkward or less exact sentence.

## Usage

Load it before writing a deliverable in prose, or when asked to edit, tighten or "de-AI" a text.
Examples: a French `docs/` page, an English README, a LinkedIn post, a PR body.

The rules cover both languages. Detect the language of the target, not of the conversation, and
apply the matching column of [references/rules.md](references/rules.md).

## Steps

1. Identify the mode: technical (docs, README, PR), published (article, post), editing, or message.
   Technical writing favours clarity over personality; see the mode table in the reference.
2. Read [references/rules.md](references/rules.md) completely.
3. Write the text normally first, starting with the useful answer.
4. Run the final pass, silently, on the draft:
   1. Cut a throat-clearing first sentence.
   2. Replace vague claims with a number, name, date or concrete example.
   3. Remove inflated importance.
   4. Search for rejected-frame contrasts, across sentence boundaries, and delete the rejected half.
   5. Replace bloated verbs with `is`, `has`, `est`, `a`.
   6. Remove banned vocabulary, dead openings and dead transitions.
   7. Delete analogies that fail the permission test, and metaphor verbs used for abstract work.
   8. Vary sentence and paragraph length.
   9. Cut an ending that only repeats the point.
5. Reread the result once aloud in your head: if it sounds forced, simplify it.

## Gotchas

- **Banning an exact technical word** - `harness`, `align`, `optimize`, `transparent`, `scalable` or
  `robust` can carry a precise technical meaning. Replacing it with a vaguer word loses accuracy;
  keep it when it is the exact term and no cleaner one exists.
- **Chat-style voice in documentation** - fragments, one-sentence paragraphs and `I`/`you` suit a
  post, not a French `docs/` page that records reasoning. Apply them only in published and message
  modes.
- **Over-correcting** - a text that visibly avoids every banned item reads as a checklist, which is
  its own tic. Write normally first, then remove what sounds machine-made.
- **Treating every contrast as banned** - correcting a real fact, date, number, scope or legal or
  technical distinction is allowed ("mardi, pas jeudi"). Only contrast used for effect is banned.

## Constraints

- Never write an em dash or a middle dot, in either language, per `harness/SOUL.md`.
- Never trade accuracy or a precise term for style.
- Never use a banned item except when quoting or naming it.
- Never add an analogy to an answer under 800 words.
- Write the text in the language of its destination, per the repository's language rule.

## References

- [references/rules.md](references/rules.md) - modes, bans in English and French, contrast rule,
  analogy rule, specificity examples.
