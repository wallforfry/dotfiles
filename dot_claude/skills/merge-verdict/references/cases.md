# Verdict cases

Three behavioural cases. Each names what is reviewed, the verdict it must reach, and the criteria
that decide pass or fail. Cases A and B target different forges so both command sets in
`references/forges.md` get exercised.

Run A and B against real pull requests: a case that never touched a forge proves nothing about a
skill whose first phase is anchoring. A scratch repository is fine - the domain does not matter, the
shape of the diff does. Case C deliberately isolates the phase 5 ledger and makes no claim about
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

## Execution record

**None of these cases has been run in this repository.** The cases state what the skill is supposed
to do; nothing here is evidence that it does. Anyone claiming otherwise is doing exactly what phase 4
exists to prevent - reading an intention as a measurement.

Append a run here when one happens: date, forge, which phases were reached, the verdict obtained
against the verdict expected, and what came back into the skill. Keep the record free of anything
belonging to the reviewed repository - no PR number, SHA, branch name, build count or defect detail.
This file is committed to a public repository; the work it was exercised on is not.

## Declared gaps

Until a run says otherwise, nothing validates: the idempotent update of the marker, the
duplicate-verdict guard, publication on either forge, `gh pr review --request-changes` as a native
blocking state, or a flat _approved_ verdict.
