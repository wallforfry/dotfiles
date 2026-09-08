# Verdict template

Fill this into `verdict.md`, then publish with the commands in `references/forges.md`. A slot marked
REQUIRED that cannot be filled means the phase it belongs to was not done - go back to it instead of
publishing an incomplete verdict. This file is English; the verdict itself follows the language of the PR.

## Skeleton

```text
<!-- merge-verdict:<pr>:<head-sha-12> -->
## Independent verdict - <changes required | approved with reservations | approved>

<Anchor sentence - REQUIRED. Whose work, on which head SHA, against which base: the real one, naming
the parent PR when the branch is stacked. Then the CI state, open tasks and conflicts, or "none
observed".>

<Contract and behaviour ledger - REQUIRED when a linked issue promises requirements or the diff
changes an observable behaviour. One row per issue requirement or additional changed behaviour,
including promises absent from the diff, in the order inventoried in phase 2. Never a prose summary,
and never shortened to keep the verdict small.

| Requirement / behaviour | Contract source | Implementation on this head | Positive evidence on this head | Negative witness | Result |
| --- | --- | --- | --- | --- | --- |
| <criterion ID and observable outcome> | <issue criterion link, or PR/diff source> | <verified location or absent> | <the test or run reproduced here> | <the observed RED, or the faulty variant that failed> | held / absent / excluded |
>

<Blocking paragraph - one clause per blocker: the mechanism, then the invariant it breaks. Close with
one sentence stating what must become true to lift them. Omit this paragraph entirely when the verdict
is "approved".>

<Barrier paragraph - REQUIRED. Open with "Authenticated local validation on this exact head:" and give
counts, never adjectives. Then, in the same paragraph, REQUIRED: what this evidence does not cover. The
verdict is invalid without that second half.>

<Non-blocking remarks - three lines at most, one per remark, each prefixed "Non-blocking:". Keep the
ones that would change a reviewer's decision, drop the rest: past three, the section is a second review
competing with the verdict for attention. Drop it entirely rather than pad it.>

Fix: <url>                REQUIRED when the verdict blocks
Initial review: <url>     REQUIRED when a previous verdict exists on this PR

<Closing sentence - REQUIRED, one line, executable:
  changes required           → "Do not approve or merge this head."
  approved with reservations  → "Mergeable once <criterion>."
  approved                    → "Approved on this head."
When the verdict blocks and the forge carries no native blocking state - Bitbucket, or your own PR on
GitHub - the same line says that this comment is the only thing holding the merge.>
```

A non-excluded row is complete only when its implementation was verified on this head and both evidence columns
were reproduced during this review. The source column identifies an obligation, never proof of
fulfillment: ticket prose, status and checked criteria cannot fill implementation or evidence cells.
Evidence the author supplied but that was not reproduced is written as theirs, in the barrier paragraph,
and leaves the row `absent`. One `absent` row forbids both approval verdicts.

For a requirement explicitly excluded by an authoritative scope decision, retain its identifier and
row, cite that decision in the source column, use result `excluded`, and write
`not required - excluded` in its implementation and evidence cells. Only this row is exempt; exclusion
is a scope decision, never proof of fulfillment or permission to ignore changed behaviour.

## Filled example

A _changes required_ verdict. Every identifier and figure below is invented - a committed skill carries
nothing from the repositories it was exercised on. The example is English because these files are; a real
verdict is written in the language of its PR. Note what the barrier paragraph does: it gives numbers, then
immediately spends a sentence dismantling its own green.

```text
<!-- merge-verdict:1042:a1b2c3d4e5f6 -->
## Independent verdict - changes required

Review of PR #1042 on a1b2c3d4e5f6, stacked base feat/ledger-read-side@9f8e7d6c5b4a. Pipeline #318
green, no task and no conflict observed.

| Requirement / behaviour | Contract source | Implementation on this head | Positive evidence on this head | Negative witness | Result |
| --- | --- | --- | --- | --- | --- |
| Closing a ledger creates exactly one successor | ISSUE-158/AC-1 | closing transaction | 7/7 close unit tests green | absent - no faulty variant run | absent |
| A concurrent write is carried into the successor | ISSUE-158/AC-2 | absent | absent | absent | absent |

Blockers: the snapshot and the controls both run before the closing transaction, so a concurrent write
can vanish from the successor; two simultaneous closes can create two successors, because the retry never
re-reads the winning result and the uniqueness constraints that would refuse the second one do not exist.
Lift: implement atomic closure and concurrent-write preservation, reproduce their positive tests and
negative witnesses against PostgreSQL, then rebase once the parent PR merges.

Authenticated local validation on this exact head: lint green (18/18 builds, 0 errors, the 145-warning
threshold respected), typecheck green on both touched packages, 7/7 close unit tests green. Those tests
remain sequential: they cover none of the concurrent interleaving that motivates the blockers above.

Non-blocking: the documented 409/412 codes no longer match the real behaviour.

Fix: https://tracker.example/ISSUE-158
Initial review: https://forge.example/pull-requests/1042/comments/155

Do not approve or merge this head.
```

## Self-check before publishing

- The marker is the first line, and its SHA is the head you actually checked out.
- The ledger carries every linked issue criterion, even absent from the diff, and every additional
  changed behaviour inventoried in phase 2; each has a source, none are merged away.
- No non-excluded promised requirement lacks implementation or reproduced evidence in an approval verdict.
- Every `excluded` row retains its identifier and authoritative decision link; only that row is exempt.
- Ticket content and status appear only as contractual input, never as fulfillment evidence.
- No approval verdict ships with an `absent` cell.
- Evidence the author supplied is attributed to them, never counted as reproduced.
- Every clause in the blocking paragraph names a sequence of steps, not a quality judgement.
- The barrier paragraph contains digits, and a sentence saying what those digits do not prove.
- The barrier paragraph names the command it ran, and that command is the one CI runs.
- The closing sentence tells the reader what to do, not how the reviewer feels.
- At most three non-blocking lines; a fourth means the section is competing with the verdict.
- No re-review ticket and no `Re-review:` slot; the head-specific verdict is the re-review record.
- Total under about thirty lines. Past that, preferences have leaked into the blocking section.
