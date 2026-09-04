---
name: obsolescence
description: >
  Decide whether an obsolete path is deleted or migrated, and refuse speculative compatibility.
  Use when removing code, a schema, an endpoint, a flag or a migration, when a compatibility shim
  or a legacy alias is about to be added, or when a rename must keep the old name working. Make
  sure to use it whenever a change would leave two ways of doing one thing, even if compatibility
  is never named.
metadata:
  category: dev
---

# Obsolescence

## Overview

Every path kept alive is paid for by every later change. Remove rather than layer, unless something
outside the repository depends on the old shape. What the project owes the outside world is the
first question, not the last: a repository with production users trades removal for migration, one
that has not shipped owes nothing.

## Usage

Read this skill before deleting an obsolete path, and before adding anything whose purpose is to
keep an old shape working: a shim, an alias, a dual-read or dual-write path, a backfill, a
deprecated-but-kept signature.

## Steps

1. **Establish what the project owes the outside world.** Where the answer is recorded - `USER.md`,
   an ADR, the README - follow it. Where it is not, establish it once and record it. Every rule
   below applies with that answer in hand, never by default.
2. **Aim for the smallest coherent design that represents the product today.** Obsolete code,
   schemas, endpoints, configuration, aliases and transitional paths are deleted, not deprecated.
3. **Add no compatibility shim, legacy alias, dual-read or dual-write path, or data-preserving
   backfill** unless the user asks for it or a published contract requires it. Speculative
   compatibility is dead code with a plausible name.
4. **Change an internal interface and its callers in the same commit**, rather than keeping the old
   signature beside the new. Internal interfaces are not public contracts.
5. **Treat development and test data as disposable.** Recreate the database instead of complicating
   the product to preserve a local state.
6. **Keep the correctness properties.** Database invariants, transactional safety, migration
   idempotence and deterministic setup survive every cleanup; they are not compatibility
   concessions.
7. **Treat migration history as a replaceable baseline, and keep the chain coherent.** Never rewrite
   an applied migration without resetting the development and test databases it touched, and
   consolidate a baseline as its own coordinated change, never as incidental work in a feature
   branch.

## Gotchas

- **Deprecating instead of deleting** - a deprecated path is a live path with a comment. It still
  compiles, still gets called, and still has to be updated by every later change.
- **Assuming the old shape has users** - the assumption is cheap to state and expensive to keep.
  Check the callers, the published contract and the recorded answer from step 1 before paying for a
  migration nobody needs.
- **A rename that keeps the old name working** - two names for one thing means the next reader picks
  whichever they find first, and both have to be maintained until someone proves the old one is dead.
- **Consolidating a migration baseline inside a feature branch** - the reset it requires lands on
  everyone else's development database as a surprise. It is its own coordinated change.
- **Trading a correctness property for a smaller diff** - dropping a constraint or a transaction to
  make a cleanup simpler turns a maintenance task into a data defect.

## Constraints

- Never add a compatibility shim, legacy alias, dual path or backfill without an explicit request or
  a published contract that requires it.
- Never keep an old internal signature beside its replacement.
- Never rewrite an applied migration without resetting the databases it touched.
- Never drop a database invariant, a transaction boundary or a migration's idempotence during a
  cleanup.
- Always state what the project owes the outside world before choosing removal over migration.
