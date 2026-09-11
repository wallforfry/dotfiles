---
name: merge-verdict
description: >
  Deliver a merge verdict on an open pull request, yours or another author's. Use when asked to
  review a PR, whether it is safe to merge or approve, for a blocking review, or a re-review after
  fixes. Make sure to use it whenever a merge decision is at stake, even if the request only says
  "look at this PR".
compatibility: >
  Authenticated `gh` (GitHub) or `bkt` (Bitbucket), plus an issue tracker CLI or MCP when a blocking
  defect needs a fix ticket.
metadata:
  category: dev
---

# Merge Verdict

## Overview

The output is a verdict that commits to a merge decision, not a list of remarks. Three properties
separate it from a default review: every blocking finding names a failure mechanism that breaks an
invariant the code claims to hold, every changed observable behaviour carries behaviour-level
evidence measured on the exact head SHA under review, and the limits of that evidence are declared
inside the verdict. A green barrier whose gaps stay implicit
is the failure this skill exists to prevent - a verdict is only as strong as what it admits it did
not test.

Not for re-reading your own diff before committing: that is `superpowers:requesting-code-review` and
`superpowers:verification-before-completion`. This skill judges an open pull request and engages a
decision the team will act on. That the PR is your own changes nothing about the verdict, only about
how it is enforced - see phase 6.

## Usage

`/merge-verdict <pr-number|pr-url>` - the forge is detected from `git remote get-url origin`.

Typical cases: "is #1042 safe to merge?" (phase 1 finds the branch stacked and the shown diff twice
its real size), "review this PR before I approve it" (phase 3 turns a vague unease into a named
mechanism, or drops it), "they pushed the fixes, re-review" (phase 6 finds the previous marker, sees
a new SHA, and publishes a second verdict rather than editing the first).

Run the six phases in order; phase 4 precedes phase 5 because a verdict without an executed barrier
is an opinion. This skill and its references are English; the published verdict follows the language
of the PR. Publishing in phase 6 is outward-facing and visible to the team - ask for confirmation
first, unless the request explicitly says to post directly.

## Steps

1. **Anchor.** Resolve the PR number, head SHA, real destination branch and checks state with
   `references/forges.md`, then check out that exact head. A repository that has its own forge skill
   or wrapper wins over those raw commands. If the PR is stacked on another unmerged PR, say so and
   make retargeting after the parent merges part of the verdict: the diff you are reading is not the
   diff that will land. A review not anchored on a named SHA is invalid.

2. **Understand before judging.** Read the PR description, the design documents and ADRs it cites, and
   the whole diff. Read the linked issue and its authoritative parent requirements, including acceptance
   criteria and approved scope decisions; preserve their identifiers and source links. The ticket is
   contractual input, never evidence that the head implements or satisfies it. If a linked contract
   cannot be read or its scope is unresolved, record that gap and block approval until resolved.
   The attachments are part of the description: a screenshot or an uploaded log is
   evidence the author supplied, and the forge returns it as raw markup that a text pass slides over.
   Open every one, in the comments too. Restate the invariant the code claims to hold, in one
   sentence. Producing a blocking finding before the flow is traced end to end is forbidden: the
   mechanism is what makes a finding blocking, and you cannot name a mechanism you have not followed.

   Then inventory every requirement promised by the linked issue, including those absent from the
   diff, plus every externally observable behaviour the diff adds, removes or changes. Open the
   ledger: one row per requirement or additional behaviour, with its contract source, implementation
   on the exact head, positive evidence, negative witness, and result. Map overlapping diff behaviour
   to the requirement row without dropping any criterion. Do not infer delivery from ticket status,
   checkboxes or prose, or treat an omission from the diff as an exclusion from the promised scope.
   A negative witness is an executed test-first RED, or a controlled faulty
   variant derived from the exact head and shown to fail when the behaviour is broken; restore and
   verify the head before running its positive barrier. A passing regression test with no observed
   failing counterpart is positive evidence only, and a green aggregate barrier is evidence for no
   individual row. Missing evidence is recorded as `absent`, never inferred.
   An explicit authoritative scope decision may exclude a requirement: retain its row and identifier,
   link that decision, and set its result to `excluded`. Only that row is exempt from implementation
   and evidence; mark those cells `not required - excluded`. A stated deferral or reviewer assumption
   is not an exclusion, and exclusion never waives defects in behaviour the head actually changes.

3. **Sweep the failure classes.** Put all ten questions in `references/failure-classes.md` to the diff.
   Record, per class, one of: not applicable, holds because `<evidence>`, or broken by `<mechanism>`.
   Only the third form produces a blocker from this sweep; the contract ledger takes precedence
   independently. When the head under review was written in this session,
   delegate the sweep to a fresh read-only context scoped to the diff: the context that produced the
   diff shares the blind spot that produced the defect, and records `holds` for the class it has just
   broken. When no fresh context is available - a host that spawns no subagent, or a session whose
   rules forbid it - run the sweep here and write its provenance into the barrier paragraph as a
   declared limit of the evidence. Skipping the sweep is not the alternative, and neither is claiming
   an independence you did not have. The sweep can add findings; it can neither replace a ledger row nor turn `absent`
   behaviour-level evidence into `holds`.

4. **Run the barrier, then declare its holes.** Run lint, typecheck and tests the way the project runs
   them - including inside a container when the project requires it, since numbers from the wrong
   runner are not evidence for this head. Read the CI configuration (`bitbucket-pipelines.yml`,
   `.github/workflows/`) to find the gate that actually blocks the merge before running anything: in a
   turborepo it is often not the package script of the same name. Report counts: builds, errors,
   warnings against the project's threshold, tests passed over tests run. Then enumerate what the
   barrier does not reach: sequential tests say nothing about a race, jsdom nothing about a browser,
   an in-memory database nothing about PostgreSQL, one platform nothing about the others. If nothing
   exercises the changed code, that absence is the review's first finding, not a reason to announce
   green.

   Then attribute every piece of evidence: measured here, supplied by the author and not reproduced
   here, or absent. The three are not interchangeable. "Not observed" written over evidence sitting
   in the description is a false statement about the author's work, and a lift criterion asking for a
   run the PR already shows asks them to repeat themselves - what is missing there is a control in
   the repository, so name the control, not the re-run. Complete the ledger from what this review
   actually reproduced; the author's evidence stays attributed and does not fill an approval row
   until it is reproduced.

5. **Return a verdict.** Exactly one of _changes required_, _approved with reservations_, _approved_.
   Write it in the order defined by `assets/verdict-template.md`. Lead with the decision, one next
   action and the current status before the detail. Keep ledger rows in visible groups of five when
   more are needed, without omitting any row or evidence. State every blocker by location, cause,
   failure mechanism and lift criterion. A non-excluded promised requirement without implementation
   or evidence blocks; name the missing outcome and the implementation and evidence needed to lift it. Both approval verdicts
   require every non-excluded row to hold an implementation, reproduced positive evidence on the exact head and a
   reproduced negative witness, with no contradictory result; otherwise the verdict is _changes
   required_ and the missing evidence is the lift criterion.
   Each other blocking finding carries its named mechanism and its lift criterion - what must become true
   for the block to go away. Reservations are for mechanisms whose consequence is bounded; a mechanism
   that can lose or corrupt data blocks even when the author disagrees. A style, naming or structure
   preference never blocks: label it non-blocking, or drop it.

6. **Trace and publish.** Use `business-issue` to open or reuse a fix ticket for blocking defects,
   passing the reviewed PR, the initial verdict when one exists, and the exact relation intent. The
   ticket publication establishes a native PR-issue development link when the available tracker and
   forge integration support it; it uses closing semantics only when that PR is expected to deliver
   the ticket. Never open a ticket solely to request or record a re-review: the new head-specific
   verdict comment is that record. Write one general comment from `assets/verdict-template.md`,
   prefixed with the idempotency marker `<!-- merge-verdict:<pr>:<head-sha-12> -->` - not a rain of
   inline comments. Search the existing comments for that marker first: a verdict carrying the same
   `<pr>:<sha>` is updated in place, never duplicated. Re-read before publishing - past about thirty
   lines, non-blocking remarks are posing as blockers, but never shorten the ledger to fit that
   count. On GitHub, _changes required_ is published with
   `gh pr review --request-changes`, the native state, unless you authored the PR: GitHub refuses it
   there, and Bitbucket has no reliable equivalent at all. Whenever that native state is unavailable,
   the comment _is_ the verdict and its closing sentence carries the whole enforcement - say so, so
   the reader knows nothing mechanical is holding the merge button.

## Gotchas

- **The base the forge reports is not the base to review** - and it errs in both directions. A PR that
  targets its parent branch hides what the parent still owes; a PR that targets the integration branch
  while sitting on an unmerged parent swells with the parent's work. Recompute the base from the
  parent's head and put retargeting in the lift criterion.
- **The comment body passed inline** - apostrophes, backticks and accented text break shell quoting and
  silently truncate the comment. Always pass the body through a file (see `references/forges.md`).
- **A stale marker updated instead of superseded** - the SHA in the marker going stale on the next push
  is the point. Same `<pr>:<sha>` → update that comment; different SHA → publish a new verdict and
  leave the old one as the record of what was judged.
- **A re-review ticket opened for traceability** - it duplicates the head-specific verdict and adds
  tracker noise without tracking a defect.
- **A fix ticket merely mentions the reviewed PR** - readers can follow the URL, but native
  development views and automations cannot follow the relationship; pass the relation intent to
  `business-issue` and verify the resulting edge.
- **"Approved with reservations" used to avoid a disagreement** - that state is a claim that the
  consequence is bounded. If you cannot state the bound, the verdict is _changes required_.
- **A ticket requirement absent from the diff disappears from review** - a correct partial change
  can still miss the promised outcome. Inventory the linked contract before the diff and block every
  unimplemented or unproven promise; documenting the omission does not fulfill it.
- **The author's evidence read as absent** - a screenshot of the output comes back from the forge as
  an `<img>` tag inside the body, and a text pass slides over it. The verdict then announces that
  nothing was observed and makes the merge conditional on steps the author has already run and
  attached, which reads as not having read the PR. Open every attachment in phase 2, and label
  supplied evidence as theirs.
- **A green aggregate barrier read as behavioural evidence** - a suite that passes proves the suite
  passes. A row of the ledger is filled by a test that was seen to fail when the behaviour is broken,
  not by a total.
- **A concise opening hides the evidence** - a reader sees a decision but not its basis; keep the
  decision, next action and status short, then retain the complete ledger and the barrier limits.
- **Numbers copied from the PR's own pipeline** - a green pipeline is context for phase 1, never the
  barrier of phase 4. The barrier is what you ran, authenticated, on the head you checked out.
- **The package script mistaken for the CI gate** - the repository's `lint` script may walk the whole
  tree while the pipeline lints only changed files. Its count then measures a backlog that predates
  the head under review. Take the command from the CI configuration, and name which one you ran.

## Constraints

- Never approve without having executed the barrier on the exact head under review.
- Never write "everything is green": report counts, or report that nothing ran.
- Never report a count from a command the pipeline does not run; name the gate you executed.
- Never publish a blocking finding without a named failure mechanism and a lift criterion.
- Never block on style, naming or structure preference; label it non-blocking.
- Never open a review that is not anchored on a head SHA.
- Never issue either approval verdict unless every non-excluded promised issue requirement, including those
  absent from the diff, and every additional changed observable behaviour has an implementation,
  reproduced positive evidence on the exact head and a reproduced negative witness.
- Never treat the linked ticket, its status or its acceptance checkboxes as implementation evidence.
- Never infer behaviour-level evidence from an aggregate green barrier.
- Never call a passing test a negative witness without an observed failure when the behaviour is
  broken.
- Never report as absent, or demand in a lift criterion, evidence the PR already supplies.
- Never omit a ledger row, evidence limit or blocker mechanism to make a verdict shorter.
- Never sweep the failure classes on a head you wrote in this session from the context that wrote it
  without saying so in the verdict; independence is either obtained, or declared absent.
- Never leave a limit of the evidence implicit; the barrier's gaps belong in the verdict text.
- Never publish two verdicts for the same `<pr>:<sha>`; update the existing comment instead.
- Never create a ticket solely to request, schedule or record a re-review.
- Never create or update a fix ticket outside `business-issue`; pass it the PR relation and intended
  closing effect instead of maintaining a second publication procedure here.
- Never publish without confirmation, unless the request explicitly says to post directly.

## References

- [references/forges.md](references/forges.md) - forge detection and the GitHub/Bitbucket command
  parity table. Read in phase 1, before the first CLI call.
- [references/failure-classes.md](references/failure-classes.md) - the ten failure classes as questions
  to put to the diff. Read in phase 3.
- [assets/verdict-template.md](assets/verdict-template.md) - the verdict skeleton with its required
  slots, ledger included. Filled in phase 5, published in phase 6.
- [references/cases.md](references/cases.md) - four behavioural cases with their expected verdicts,
  and the record of what they have never validated.
