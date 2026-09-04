---
name: harness-audit
description: >
  Measure harness cost, deployment lag, activation, adherence, and mutation detection. Use when
  auditing whether a rule, skill or subagent earns its place, or before adding or removing one.
  Make sure to use it whenever a harness claim needs a count, even if measurement is never named.
compatibility: >
  The dotfiles checkout, `python3`, and Claude or Codex JSONL transcripts. Override their roots with
  `CLAUDE_PROJECTS` and `CODEX_SESSIONS`.
metadata:
  category: ops
---

# Harness Audit

## Overview

The harness states rules; this skill measures whether they act. `scripts/harness-audit.sh` reports
fixed and amortized context cost, deployment lag, host-normalized activation and adherence, then a
promise-to-control matrix for `scripts/verify.sh`. Frequency of use is not usefulness: a rule
activated once may guard an irreversible loss, so read counts against what each component protects.

## Usage

```bash
bash ~/dotfiles/scripts/harness-audit.sh
```

The script measures the checkout it ships in, whatever the working directory: it resolves its own
location and never the current repository, so it is run unchanged from a session opened in any
project. Name the working checkout, not `chezmoi source-path`: that path is the deployment clone.

Environment: `CLAUDE_PROJECTS` and `CODEX_SESSIONS` select transcript roots;
`HARNESS_RULES_SINCE` selects the rule-introduction date. Aggregates are cached under
`~/.cache/harness-audit`; neither paths nor transcript content are stored.

Exit 1 means a mutation slipped past the barrier, or a measurement could not be made. A measurement
that did not run is reported as not done, never as green.

## Steps

1. Run the script and read the five measurements. It copies the current tracked and untracked state
   into a clone under `$HOME/.cache` and never mutates the working tree.
2. Compare the always-loaded total to the previous run. A section that grew must be justified by a
   rule that changes behaviour on tasks unrelated to its subject; otherwise it belongs in a skill,
   and `agent-instructions` owns that move.
3. Read the deployment lag as normal, never as a defect: the chezmoi source is a distinct clone of
   the working checkout by design (ADR-001), and a merged commit takes effect at the next
   `chezmoi update`. The count is measured against the last fetch, without one.
4. Read every zero-activation skill against its denominator. A skill scoped to this repository is
   measured on the sessions in this repository, and a skill created days ago is measured on days.
5. Read the adherence rates as correlational only. The model and the host prompt changed over the
   same period, so the split proves an association, never a cause.
6. Treat a mutant accepted or an anti-mutant rejected as a barrier regression: fix `verify.sh`, then
   re-run.
7. Map every promise to a rejecting mutant, an accepting anti-mutant, or an explicit observation.
   A promise absent from the matrix is unmeasured.

## Gotchas

- **Reading an activation count as a verdict** - `scripts` has zero activations and guards the
  bootstrap of a host with no package manager. Frequency and criticality are different axes.
- **Printing a transcript path** - project directory names carry client and employer names, and this
  repository is public (ADR-016). The script aggregates and never prints a path; keep it that way.
- **Comparing rates across a rule change and a model change** - both happened in August 2026. The
  before/after split is the only natural experiment available here, and it is confounded.
- **Adding a mutation that the barrier was never meant to catch** - the barrier's scope is its
  claim. A mutation outside it belongs to a new check in `verify.sh`, added first.
- **Naming the script by a relative path** - a session runs in some other project far more often
  than in this checkout, and `bash scripts/harness-audit.sh` then resolves to nothing or to that
  project's own script. Name the checkout by an absolute path, as `Usage` does.
- **Running the copy inside the deployment clone** - `chezmoi source-path` is a distinct clone
  (ADR-001), so the script measures that tree instead: the deployment lag collapses to "single
  clone", and every count describes whatever commit the clone last pulled.
- **Dropping a dirty tree from the capture** - the audit validates the current tracked and untracked
  state. Any failed patch or copy makes the measurement unavailable instead of falling back to HEAD.
- **Treating a missing host signal as zero** - Codex transcripts expose no explicit skill event.
  Report activation as unknown for that host instead of inventing zero activations.

## Constraints

- Never print or paste a transcript or project path into a file, a log or a commit message.
- Never mutate the working tree; every mutation happens in the clone under `$HOME/.cache`.
- Never report a measurement that did not run as passing.
- Never store raw transcript content or paths in the cache; persist aggregate counters and hashes.
- Never claim a rule works from its wording; cite the count, or say the rule is unmeasured.
- Never remove a component on an activation count alone, without naming what it protects.
