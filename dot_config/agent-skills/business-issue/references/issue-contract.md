# Issue contract

Adapt the headings to the recipient's language and existing conventions. Omit empty sections;
the outline is not a requirement to manufacture detail. The description is the canonical contract,
with links to shared source material rather than copied parent descriptions.

## Suggested shape

Start the draft with two short, visible lines before its detailed contract:

```text
Outcome: the business result now established or the current decision state
Next: one action or decision needed to advance the issue
```

Use `Decision:` instead of `Outcome:` when no result is established yet. The recipient may translate
these labels, but preserve their order and make both values concrete. This opening is an entry point,
not a replacement for any required section below.

For a multi-step handoff, state `Progress` as the current step and next action. When an error matters
to the next action, state its `Location`, `Cause` and `Correction`. Keep a visible numbered or bullet
group to five items or fewer; split a longer group by purpose. The regression fixture exercises every
one of these presentation rules alongside acceptance scenarios and completion evidence.

```text
Title: the business outcome or decision

Summary
Who faces what situation, what changes, and the principal boundary.

Outcome and scope
Included result, preserved behaviours and explicit exclusions.

Rules and evidence
Rule | Source location and date/version | Status
Distinguish observations, approved decisions, proposals and unknowns.

Open decisions and prerequisites
Question | Options and consequences | Decision owner | What it blocks
Link actual dependencies; omit speculative chains.

Relations to establish
Subject | Relation | Target | Reason and expected side effect
Record parent/child direction, blocking direction, canonical duplicate and related work. Distinguish
a pull request that implements the issue from one that is only supporting context, and state whether
merge is intended to close the issue.

Acceptance scenarios
AC-1: Given a situation, when an action occurs, then an observable result follows.
AC-2: A relevant failure or refusal preserves the required state.
AC-3: A previously supported scenario still produces its intended result.

Completion evidence
For each criterion: expected observation, how it can be checked, and required environment.
Identify who may accept a business decision when human judgment is required.
```

Keep identifiers stable when refining criteria; mark a superseded criterion explicitly rather than
reusing its identifier for a different requirement. An outcome without a known verification method
needs investigation, not an arbitrary passing test count.

For a research issue, completion is an answered question or an explicit remaining uncertainty with
evidence, alternatives and a decision owner. It is not a promise to implement whichever option wins.
For implementation work, link the chosen decision and describe behaviour without prematurely fixing
the implementation when several designs satisfy it.

## Example of a consequential unknown

A dispatcher persona asks to release only "approved" bookings. The source is a proposed interview
summary; current code distinguishes draft and submitted bookings, but has no approval event.

Do not rename submitted to approved or invent a new approver. Determine whether approval is a new
capability, a synonym ratified by the business owner, or an unsupported assumption. Until resolved,
frame that decision with its source and affected scenario. Existing code informs feasibility; it
does not decide what the product should do.
