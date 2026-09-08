# Issue contract

Adapt the headings to the recipient's language and existing conventions. Omit empty sections;
the outline is not a requirement to manufacture detail. The description is the canonical contract,
with links to shared source material rather than copied parent descriptions.

## Suggested shape

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
