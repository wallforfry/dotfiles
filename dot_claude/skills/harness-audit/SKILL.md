---
name: harness-audit
description: >
  Re-measure the harness: context cost, deployment lag, real skill and subagent activation,
  adherence to the two observable rules, and what the verification barrier actually detects.
  Use when asked whether a rule, a skill or a subagent earns its place, before adding or removing
  one, or when an audit needs refreshing. Make sure to use it whenever a claim about the harness
  needs a count rather than an opinion, even if measurement is never named.
compatibility: >
  The dotfiles repository checkout, `python3`, and the agent transcripts under
  `~/.claude/projects` (override with `CLAUDE_PROJECTS`).
metadata:
  category: ops
---

# Harness Audit

## Overview

The harness states rules; this skill measures whether they act. `scripts/harness-audit.sh` produces
five counts, and the counts are the finding: an always-loaded byte total, how far the chezmoi source
clone lags `origin/main`, an activation count per skill, a violation rate before and after a rule's
introduction date, and a mutation-detection score for `scripts/verify.sh`. Frequency of use is not usefulness: a rule activated once may guard an
irreversible loss, so read the counts against what each component protects, never alone.

## Usage

```bash
bash scripts/harness-audit.sh
```

Environment: `CLAUDE_PROJECTS` for another transcript directory, `HARNESS_RULES_SINCE` for another
rule-introduction date (default `2026-08-17`, the day the typography and comment rules landed).

Exit 1 means a mutation slipped past the barrier, or a measurement could not be made. A measurement
that did not run is reported as not done, never as green.

## Steps

1. Run the script and read the five sections. It clones the repository into `$HOME/.cache` and never
   mutates the working tree.
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
6. Treat a missed mutation as a barrier regression: fix `scripts/verify.sh` first, then re-run.
7. To measure a rule the script does not cover, add a mutation to the `MUT` block rather than
   asserting the rule works. A rule with no mutation is an unmeasured rule.

## Gotchas

- **Reading an activation count as a verdict** - `scripts` has zero activations and guards the
  bootstrap of a host with no package manager. Fréquence and criticality are different axes.
- **Printing a transcript path** - project directory names carry client and employer names, and this
  repository is public (ADR-016). The script aggregates and never prints a path; keep it that way.
- **Comparing rates across a rule change and a model change** - both happened in August 2026. The
  before/after split is the only natural experiment available here, and it is confounded.
- **Adding a mutation that the barrier was never meant to catch** - the barrier's scope is its
  claim. A mutation outside it belongs to a new check in `verify.sh`, added first.
- **Running it on a dirty tree** - the clone follows the current branch, so uncommitted work is
  absent from the mutation section and the counts describe the last commit.

## Constraints

- Never print or paste a transcript or project path into a file, a log or a commit message.
- Never mutate the working tree; every mutation happens in the clone under `$HOME/.cache`.
- Never report a measurement that did not run as passing.
- Never claim a rule works from its wording; cite the count, or say the rule is unmeasured.
- Never remove a component on an activation count alone, without naming what it protects.
