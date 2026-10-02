# semctx evidence

How a semctx report enters phases 1 to 4. semctx is a static proof surface: it maps a diff to the
symbols, exported contracts, authored invariants and tests it reaches, and returns a Plane A
verdict. It never builds or runs the code, so it selects what to examine and what to re-test; it
proves no behaviour.

This file says where that evidence goes in a verdict and nothing else. The tool contracts, verdict
namespaces and CLI ladder belong to the plugin: read `semctx:semctx-verify` for Plane A and
`semctx:semctx-control` for index health, control freshness and the read-only audit lane, and
follow them when this file is silent.

## Guard

Both conditions, tested on the capability itself, in this order:

1. The checkout holding the exact head of phase 1 has a `.semctx/` directory at its root:
   `test -d "<checkout>/.semctx"`. Test that checkout, and never substitute another one for it.
2. The semctx MCP tools answer: `semctx_control_status` with the absolute `repositoryRoot` returns a
   report. A tool the host does not list, or a call that errors, fails the guard.

The first condition is checked without any semctx call. When either fails, this file has no effect:
no further call, no field, and no mention of semctx anywhere in the verdict. Neither the host, the
profile nor the forge decides; only these two tests do. That silence hides no limit: semctx fills
no ledger cell, so its absence removes no evidence the verdict relies on.

`.semctx/` is mostly ignored state: a worktree created for the review carries only what the
repository versions, such as `.semctx/config.json` and `.semctx/semantic/`, and no index. The guard
can then hold on a checkout that was never indexed, which phase 1 reports as `unavailable`.

## Phase 1 - anchor the index

1. Call `semctx_control_status` and `semctx_index_health`. Record, verbatim and in full, the indexed
   head commit of the freshness seal (`indexedHeadCommit`), the seal hash, the control-freshness
   verdict and its reasons, and the index health binding, freshness and coverage with theirs.
2. The evidence is usable only when all of these hold:
   - `indexedHeadCommit` equals the full head SHA under review;
   - control freshness is `FRESH`;
   - index binding is `valid`;
   - coverage is `complete` or `partial`, never `insufficient`.
3. When `indexedHeadCommit` is null, as on a checkout never indexed (`UNSEALED`), the evidence is
   `unavailable` and no reindex is offered: a first index is setup work, not a review's. When it
   names another commit, control freshness reads `STALE` with `HEAD_MISMATCH`, possibly beside the
   other mismatches a head change entails. The evidence is then `unavailable` by default, with the
   reported reasons. A reindex of this head replaces it only when every condition below holds:
   - the user confirms it for this review, after being told that it rebuilds the index of that
     checkout and that the index then describes the reviewed head;
   - the checkout's `HEAD` is the reviewed SHA and `git status --porcelain` is empty;
   - no handoff state sits under `.semctx/working/`, since a capsule awaiting resume would go stale;
   - the repository's instructions do not restrict execution to a container; silent instructions
     mean ask, never assume.
   Run `index --json --root "<checkout>"`, with the absolute path of the checkout under review, from
   the CLI rung `semctx:semctx-control` selects, under a wall-clock limit set before starting. Never
   omit `--root`: the CLI otherwise indexes its working directory, and an agent shell may have
   returned to the primary checkout, which none of the conditions above examined. Then repeat steps
   1 and 2, and record the new seal hash and the words "reindexed during this review".
4. Any other failed condition, an overrun limit, a failed index, or a step 2 still failing after the
   reindex makes the evidence `unavailable`. Never retry on another commit.

`unavailable` concerns this optional surface only: phases 2 to 4 then make no semctx call, the
barrier paragraph states it with its reason, and it never turns a ledger row `absent`.

The plugin forbids reindexing merely to make a stale state green, and limits audits to read-only
surfaces. This reindex binds the index to the head under review before any verdict is read, never
after one, and it is the review's only semctx write; it needs the user's confirmation for that
reason. Never run `semctx_setup`, in any form, from a review: it creates repository state and can
rewrite a versioned `.gitignore`. Never enable guarded mode or install a semctx hook.

## Phase 2 - candidate ledger rows

1. Finish the inventory of phase 2 from the issue and the diff first, so semctx's contribution can
   be counted against it.
2. Build the diff from the base recomputed in phase 1: `git diff <merge-base> <head>`. Pass it to
   `semctx_verify_change` as `gitDiff`. Omitting `gitDiff` analyses `git diff HEAD`, which on the
   clean checkout of phase 1 is empty.
3. Each impacted exported contract and each impacted invariant is a candidate "changed behaviour"
   row. A candidate is a lead, never proof: confirm it by reading the code on the head, then add it
   as an ordinary row whose source cites the semctx finding, or merge it into the row that already
   covers the same behaviour, or drop it with a one-line reason.
4. Count three numbers for the barrier paragraph: candidates proposed, candidates retained, and
   retained candidates that the inventory of step 1 did not already hold.

Plane B runs only when `.semctx/semantic/` holds authored, non-empty invariants; a model with no
node, as left by setup, gets no call. Then call `semctx_semantic_check`, and `semctx_semantic_slice`
on each impacted symbol: every authored invariant of a slice is one more candidate under step 3,
counted with the others. Stay in the read-only lane: call no change-contract tool, since a review
holds no change contract of its own. Plane C never runs in a review.

## Phase 3 - BLOCK as a question

Each Plane A `BLOCK` or `WARN` finding becomes a question put to the diff under the failure class it
touches: which sequence of steps breaks that invariant or contract on this head? Record the answer
in the sweep as for any class. It blocks only as "broken by `<mechanism>`"; a `BLOCK` with no
mechanism the sweep can name blocks nothing. `semctx:semctx-verify` requires resolving every `BLOCK`
before work is declared finished: that rule binds the author completing a change, while this review
judges a head it does not finish. The missing test it reports remains a ledger matter: a retained
row without a negative witness is `absent`, and blocks through the ledger, not through semctx.

## Phase 4 - barrier and limits

1. Run every test in `recommendedTests` with the project's runner, in a container when the project
   requires one. These runs sit outside the CI gate: report their counts in the semctx clause, never
   merged into the gate counts, and name every recommended test the gate does not run or that could
   not be run.
2. A Plane A `PASS` fills no ledger cell, neither positive evidence nor negative witness.
3. Before writing the clause, call `semctx_control_status` again. A seal hash or indexed commit
   that differs from the one recorded in phase 1 means the index moved under the review, perhaps by
   another session in the same checkout: the evidence becomes `unavailable` and every candidate it
   proposed is dropped or re-confirmed by reading the code.
4. Write one semctx clause in the barrier paragraph, fields kept separate and never merged into one
   health claim: index binding, index freshness and coverage from `semctx_index_health`, control
   freshness with its seal hash and indexed commit, and the Plane A verdict. Then give the three
   candidate counts and the recommended tests run beyond the gate.
5. Coverage `partial` forbids every negative claim: no "nothing else is impacted", "no other
   contract is touched", "the blast radius stops here". `complete` permits "semctx reports no
   further impact", attributed, never as a fact about the system.

Example clause, with invented values shortened here for width; a real clause writes the seal hash
and the indexed commit in full:

```text
semctx: binding valid; index freshness FRESH; coverage partial; control freshness FRESH, seal
sha256:4f1c...e09a on a1b2c3d4e5f6..., reindexed during this review; Plane A WARN. 3 candidates,
2 retained, 1 absent from the initial inventory; 4 recommended tests run outside the gate, 4/4
passed, 2 not run by CI. Partial coverage: no claim is made about code semctx did not analyse.
```

When the evidence is `unavailable`, the clause says so with its reason and nothing else.
