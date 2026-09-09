# Verdict cases

Four behavioural cases. Each names what is reviewed, the verdict it must reach, and the criteria
that decide pass or fail. Cases A and B target different forges so both command sets in
`references/forges.md` get exercised.

Run A and B against real pull requests: a case that never touched a forge proves nothing about a
skill whose first phase is anchoring. A scratch repository is fine - the domain does not matter, the
shape of the diff does. Cases C and D isolate the contract and ledger rules and make no claim about
forge behaviour.

## Case A - changes required (Bitbucket)

**Diff.** An endpoint that reads a state, decides from it, and writes. The read sits outside the
transaction that performs the write, the retry loop wraps only the write, and no unique index backs
the uniqueness the service checks in code. Unit tests exist and pass, sequentially.

**Expected verdict:** _changes required_.

**Pass criteria**

- Anchored on the head SHA, with `<!-- merge-verdict:<pr>:<head-sha-12> -->` as the first line.
- At least one blocker stated as an ordered sequence ending in a broken invariant: the
  out-of-transaction read, the stale retry, or the missing constraint. Failure classes 1, 2 and 3.
- The barrier paragraph gives counts *and* states that the passing tests are sequential, and
  therefore say nothing about the interleaving that motivates the blockers.
- A fix ticket is linked, and no ticket exists solely to request or record a re-review.
- The closing sentence forbids the merge. On Bitbucket that sentence is the entire enforcement; its
  absence fails the case even when everything else is right.

**Fail signals**

- "The tests pass, so the concurrency looks fine."
- A blocker phrased as a risk - "this could be racy" - with no sequence.
- Naming or structure remarks inside the blocking paragraph.
- Two comments published, or a second comment added when one already carries the same `<pr>:<sha>`.

## Case B - approved with reservations (GitHub)

**Diff.** A small, correct bug fix: a null-handling defect repaired at its cause, one regression test
covering the reported input, and a controlled faulty variant reproducing its test-first RED. A
second, unchanged input path reaches the same function and is not covered; the pre-existing error
code is undocumented.

**Expected verdict:** _approved with reservations_.

**Pass criteria**

- The ledger records the repaired input, its passing test on the exact head, and the reproduced RED
  as its negative witness.
- The reservation is named as a bounded consequence - the uncovered second path, the undocumented
  code - with what would lift each.
- The barrier paragraph gives counts and states that coverage stops at the reported input path.
- No blocker. Neither an uncovered path nor an undocumented code is a mechanism that loses data;
  promoting either to a block fails the case.
- The closing sentence states the merge criterion instead of forbidding the merge.

**Fail signals**

- Blocking to be safe, on coverage or on documentation.
- Approving flatly, with both reservations dropped or buried in prose.
- "Everything is green", with no counts.
- A barrier paragraph that never says what the single test does not cover.

## Case C - changed behaviour without negative witnesses

**Evidence package.** A change claims four observable contracts: historical cursors keep their
original ordering, a replayed create returns the original resource without a second allocation, a
schema migration publishes atomically, and the public contract declares the conflict returned for an
identical in-flight request. The exact head passes a large aggregate barrier, and the record holds no
test-first RED and no faulty variant for any of the four.

**Expected verdict:** _changes required_.

**Pass criteria**

- The ledger keeps all four contracts as four separate rows.
- Aggregate barrier results are never substituted for behaviour-level evidence.
- Every missing negative witness is recorded as `absent`.
- The verdict blocks, and the four witnesses are the lift criteria.

**Fail signals**

- Either approval verdict, because the aggregate barrier is green.
- A summary paragraph that drops one or more rows.
- A passing regression test described as a negative witness with no observed failing counterpart.

## Case D - linked requirement missing from the diff

**Evidence package.** A linked issue promises AC-1: submitting a valid booking preserves its data,
and AC-2: submitting an unauthorized booking is refused without changing its state. The PR implements
only AC-1, with reproduced positive evidence and a negative witness on the exact head. Its aggregate
barrier passes. AC-2 has no implementation or behaviour-level evidence and is absent from the diff;
the ticket nevertheless has both criteria checked and a completed status.

**Expected verdict:** _changes required_.

**Pass criteria**

- The ledger contains separate AC-1 and AC-2 rows with their issue source links.
- AC-1 retains its implementation and reproduced positive and negative evidence.
- AC-2 records implementation and both evidence cells as `absent`, despite the completed ticket.
- The blocker names the missing refusal and state preservation; implementing AC-2 and reproducing
  its positive evidence and negative witness on the reviewed head are the lift criteria.

**Fail signals**

- Either approval verdict because all changed behaviour is proven or the barrier passes.
- Omitting AC-2 because the diff contains no related code.
- Using the ticket, its checked criteria or its status as proof of implementation or acceptance.
- Moving the promised refusal to a non-blocking reservation or merely documenting its absence.

**Discriminating control.** Remove AC-2 from the promised scope through an explicit authoritative
scope decision linked from the issue, rather than a reviewer assumption. With only AC-1 promised,
its complete evidence and no other blocker, the missing AC-2 alone must no longer block approval;
the ledger retains AC-2 with result `excluded`, its identifier and that decision's source link, and
`not required - excluded` in implementation and evidence cells. AC-1 still requires its full proof;
a documented deferral without an authoritative exclusion must keep AC-2 `absent` and block.

## Execution record

**Evidence status:** observation only, not reproducible evidence for either approval verdict.

Cases C and D were observed on 2026-09-09 with Codex in separate fresh, read-only subagents. Case C
returned _changes required_, kept four ledger rows, recorded all four negative witnesses as
`absent`, and made those witnesses the lift criteria. Case D returned _changes required_ while AC-2
was promised but absent; its discriminating control retained AC-2 as `excluded`, filled its three
evidence cells with `not required - excluded`, and returned _approved_. All observed verdicts
matched the cases.

These observations exercised only the contract inventory, ledger and verdict logic. They had no PR, head
SHA, forge, authenticated barrier or publication step, so they are not evidence for anchoring,
barrier execution or forge behaviour. Cases A and B remain unrun.

One run has happened that is not a case. On four self-authored pull requests of this repository,
phases 1 to 5 only, publication not reached: three verdicts of _approved_ or _approved with
reservations_ and one reservation that came straight back into the skill. The failure-class sweep
could not be delegated - the session's own rules forbade spawning a subagent - which is why phase 3
now carries a fallback instead of a step that can only be violated. The repository's barrier was
executed on each of the four head SHAs. Nothing about publication, the marker or either forge was
exercised.

Append a run here when one happens: date, forge, which phases were reached, the verdict obtained
against the verdict expected, and what came back into the skill. Keep the record free of anything
belonging to the reviewed repository - no PR number, SHA, branch name, build count or defect detail.
This file is committed to a public repository; the work it was exercised on is not.

## Declared gaps

Nothing yet validates the idempotent update of the marker, the duplicate-verdict guard, publication
on either forge, `gh pr review --request-changes` as a native blocking state, or a flat _approved_
verdict on a real pull request.
