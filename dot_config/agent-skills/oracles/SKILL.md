---
name: oracles
description: >
  Write a TypeScript module's spec as executable laws checked against an independent model, have the
  developer approve their plain-language reading, then verify the implementation against them.
  Use when a stateful or rule-heavy module needs its intent pinned before agent-written code ships.
  Make sure to use it whenever someone asks to prove code matches intent, even if oracles are never
  named.
metadata:
  category: dev
---

# Oracles

## Overview

Moves human review from the diff to the intent. The spec is written as code: a naive reference model
plus laws, run with `fast-check` and Vitest over thousands of generated inputs and call sequences.
The agent translates the laws into plain language for the developer to approve; once approved, the
implementation is judged by the laws, not by reading it. This is oracle and property-based checking,
not a mathematical proof: say so, because a wrong law gives a green light on the wrong behaviour.

## Usage

```text
/oracles <module or area>   - pin the promises of an existing module, then check it
/oracles new <feature>      - write and approve the laws before any implementation
```

Typical targets: a store, a cache, a queue, a state machine, permissions, parsing and serialisation,
or business rules such as shares that must always sum to the whole. Not pure UI, integration glue,
or behaviour nobody has decided yet; steps 1 and 4 exist to force that decision.

## Steps

1. **Frame the promises.** Read the module, its callers, its tests and its docs. List each promise
   in one sentence, as a caller observes it, never as internal mechanics. Probe the real code with
   small scripts to confirm each behaviour before writing it as a law. Record every disagreement
   between docs, code and tests as a design question.
2. **Check the tooling.** Use the repository's package manager and test runner. Add `fast-check` as
   a dev dependency only if absent. Offer `@stryker-mutator/core` with
   `@stryker-mutator/vitest-runner` for step 7. `@tanstack/oracles` is unpublished and still moving:
   use it only when the developer asks for it, and then follow the skills it ships.
3. **Write the spec in code.** Create `<module>.oracle.test.ts` next to the module, in three parts:
   - the **model**: the obvious, slow implementation of the promise (a `Map`, an array, a
     filter and sort), importing nothing from production;
   - the **generators**: `fc.Arbitrary` inputs and, for stateful code, `fc.commands` over the public
     operations, covering empty, duplicates, extremes, Unicode, date boundaries, zero and negative
     amounts, and unusual call orders;
   - the **laws**: one `it()` per promise, named by the promise in plain words. Use equivalence to
     the model after every command, invariants, round trips, idempotence or monotonicity where the
     promise implies them, and the inputs that must fail and how.
   Set `numRuns` explicitly, at least 500 for critical laws. See the skeleton below.
4. **Translate and get approval.** Give the developer, in their language: each law as one plain
   sentence, what the model treats as correct, what the generators cover and do not cover, and the
   open design questions with a recommended option each. Stop until the laws are approved. When the
   developer corrects a sentence, fix the law in code first, then translate again: the code is the
   spec, the prose is its reading.
5. **Run the laws on the implementation.** For each failure, keep fast-check's shrunk counterexample
   and seed, and decide with the developer whether the code or the law is wrong.
6. **Fix oracle first.** A confirmed bug is fixed while its law fails, until it passes. Pin the
   shrunk counterexample so it survives later runs: as a fixed case (`examples` in `fc.assert`) for
   a plain property; for an `fc.commands` sequence, as a dedicated `it()` that writes the shrunk
   commands out by hand and runs them with `fc.modelRun`. A `seed` with `replayPath` replays the
   same case only while every generator stays unchanged, then silently tests something else. For
   `/oracles new`, the implementation starts only after step 4.
7. **Measure what no law judges.** Run mutation testing on the module with the oracle file as the
   only suite. Report each survivor as a missing law, dead code, or a speed-only shortcut, and
   propose the law that would kill it. If Stryker is not installed, say so and offer it.
8. **Deliver.** Put the approved laws and their plain-language reading in the pull request
   description: that is what a reviewer reads. Wire the oracle file into the existing CI; if it is
   slow, lower `numRuns` on pull requests and keep a long scheduled run rather than dropping laws.
   End with the approved laws, the bugs found, the design questions settled, and the gaps left.

### Skeleton for stateful code

```ts
import fc from 'fast-check'
import { describe, expect, it } from 'vitest'
import { createStore } from './store'

type Model = { items: Map<string, string> }
type Real = ReturnType<typeof createStore>

class Put implements fc.Command<Model, Real> {
  constructor(readonly key: string, readonly value: string) {}
  check = () => true
  run(model: Model, real: Real) {
    model.items.set(this.key, this.value)
    real.put(this.key, this.value)
    expect(real.get(this.key)).toEqual(model.items.get(this.key))
  }
  toString = () => `put(${this.key}, ${this.value})`
}

const commands = [
  fc.tuple(fc.string(), fc.string()).map(([k, v]) => new Put(k, v)),
  // Get, Delete, Clear...
]

describe('store laws', () => {
  it('behaves like a Map for any sequence of put, get and delete', () => {
    fc.assert(
      fc.property(fc.commands(commands, { maxCommands: 50 }), (cmds) => {
        fc.modelRun(() => ({ model: { items: new Map() }, real: createStore() }), cmds)
      }),
      { numRuns: 1000 },
    )
  })
})
```

## Gotchas

- **A model as clever as production** - it reproduces the same bugs and the laws pass for the wrong
  reason. Keep it naive; when it grows, simplify it or split the promise.
- **Importing production logic into the model** - the oracle stops being independent and checks the
  code against itself. Import only the public API under test, and only on the real side.
- **Weakening a law to turn it green** - the failure was the finding. Decide with the developer
  whether the law or the code is wrong; never loosen a generator or an assertion silently.
- **Clock, randomness or network inside a law** - failures stop replaying from their seed. Inject
  fake timers, a seeded random source and fakes for I/O.
- **Treating proposed laws as approved** - the laws you wrote are a proposal until the developer
  confirms their plain-language reading; skipping step 4 turns the check into self-grading.

## Constraints

- Never present the result as a proof or an absolute guarantee: state what is covered, with how many
  runs, and what is not.
- Never implement or fix production code before the developer has approved the laws.
- Never generate the model from, or share code with, the implementation under test.
- Never drop a law or a pinned counterexample to make CI faster; lower `numRuns` instead.
- Always keep the shrunk counterexample and seed of every failure in the report.
