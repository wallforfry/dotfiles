# Global AI Instructions

Deployed to every agent's user scope. Voice and values live in `@SOUL.md`, user preferences in
`@USER.md`; this file holds only technical rules that apply to every project.

@SOUL.md
@USER.md

## Critical Analysis (ALWAYS)

Before implementing any request:

- Challenge assumptions and unclear requirements.
- Question the "why" behind a request, not just how to satisfy it.
- Verify alignment with the project's existing patterns before introducing a new one.
- Point out when a request conflicts with existing code, architecture, or a recorded decision.
- Never record a workaround for a defect in code we own: fix it, or open a ticket and reference it
  from a `TODO`. A memorised dance around our own bug guarantees the bug survives.

Raising a concern never blocks delivery: state it, then proceed as described in `USER.md`.

## Context Management

- For open-ended exploration, research, or multi-file searches, delegate to a subagent instead of
  reading files directly in the main thread.
- For independent multi-step tasks — especially two or more unrelated ones — dispatch parallel
  subagents rather than doing them serially inline.
- Keep the main thread for orchestration and decisions; push bulk reading, grepping, and exploration
  into subagents.

## Verification Claims

- **Check that the barrier covers what changed.** Before saying "green", confirm that a linter and a
  test actually run on what you touched. If nothing covers it, that gap is the first thing to fix —
  not a reason to claim green.
- **Name the environment.** Every piece of evidence states where it was produced and is valid only
  there. Green in one container, one shell, or one OS says nothing about the others the project
  supports: list the supported targets, say which you exercised.
- **Report counts, not adjectives.** "18/18 builds, 0 errors, 7/7 tests" is evidence; "all good" is
  not.

## Code Style

- **Write no comment.** One is admissible only when it records a fact living outside the file —
  upstream defect, protocol quirk, deliberate deviation — and names that fact. Doc comments the
  project's tooling requires are out of scope.
- Match the surrounding code's naming and idiom. The neighbourhood sets the idiom, never the quality
  bar: do not reproduce a nearby defect.

## Instruction and Skill Maintenance

- Keep always-loaded files limited to stable guidance that applies to every task; put conditional
  procedures in skills so they load only when relevant.
- Maintain one canonical source for shared guidance. Agent-specific paths (`CLAUDE.md`,
  `.claude/skills`) contain adapters or projections, never a second copy to keep in sync.
- Prefer `AGENTS.md` for shared instructions and keep `CLAUDE.md` as a thin Claude adapter when both
  exist.
- Update every importer, generated file, installer, and index that consumes a source you changed.
- Add context only when it changes agent behaviour; refine or remove stale guidance instead of
  accumulating it.
