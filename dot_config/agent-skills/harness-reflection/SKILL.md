---
name: harness-reflection
description: >
  Turn a repeated agent failure into one evidence-backed harness change. Use when the same material
  outcome fails twice, a recovered approach fails again, or the user asks what the harness should
  learn. Make sure to use it whenever retrying stops producing new information, even if learning or
  reflection is never mentioned.
metadata:
  category: ops
---

# Harness Reflection

## Overview

Two identical failures are a signal to investigate, never a proof that a permanent rule is correct.
This skill stops the retry loop and converts it into exactly one falsifiable candidate, which stays
session-local until the user approves a trial.

The failure mode it exists to prevent is the instruction file that grows a rule for every incident:
each one is paid for on every later task, and none of them was ever tested. A candidate that cannot
be disproved is not a learning, it is a superstition with a commit hash.

## Usage

Use it after a second materially equivalent failure, after a recovery that reproduces an earlier
failure, or when reviewing what the harness should learn from a finished task. Do not use it for one
ordinary command error whose message already carries its correction.

Typical cases: two sessions independently miss the same verification step although the rule is
discoverable; a template renders on this machine and fails on the other profile for the third time;
a subagent keeps returning a shape the caller cannot use.

## Steps

1. **Name the outcome.** State the repeated observable outcome, the intended one, and why the
   attempts are materially equivalent - same outcome, same cause, same recovery. Stop repeating the
   unchanged approach.
2. **Keep the smallest evidence.** The failing command and its output, the environment, the
   repository, and the recovery that worked or did not. No transcript, no secret, no personal data.
3. **Classify the cause** as `task-specific`, `owned-defect`, `external-transient`,
   `missing-capability` or `harness-gap`. Choose `harness-gap` only when a reusable instruction,
   skill, tool preference, sequence or routing decision could plausibly have changed the outcome.
4. **Check what already exists** - the instruction files, the skills, the ADRs, and the dependency's
   own current documentation - before proposing anything. An owned defect gets fixed, not
   documented: a memorised dance around our own bug guarantees the bug survives.
5. **Return one decision.** Either `skip`, with the reason and the next diagnostic action, or
   `propose`, with one candidate whose type is `verification-step`, `tool-preference`,
   `sequence-recipe`, `avoid-strategy` or `subagent-routing`.
6. **Make the candidate falsifiable.** It carries its trigger, the desired behaviour, its scope, the
   evidence, a counterexample, the falsifier, an expiry condition, and the cheapest behavioural
   trial that could disprove it.
7. **Wait for approval, then place it.** A skill change goes through `skill-manager`; an instruction
   or deployment change through `agent-instructions`; a mechanical check through
   `scripts/verify.sh`, the only place a rule becomes enforceable rather than advisory.
8. **Promote on evidence.** Three independent sessions where the trial changed the target behaviour,
   with no contradictory result. Roll back on two failed trials, one safety regression, or a veto.

## Gotchas

- **Counting unrelated failures** - a flaky network call and a wrong assumption about the repository
  do not form a pattern. Compare outcome, cause and recovery before calling two failures the same.
- **Learning from recurrence alone** - repeated mistakes often share one wrong premise, and the rule
  then encodes the premise. Require a trial that could fail.
- **Encoding a defect we own** - the workaround becomes permanent policy and the defect becomes
  invisible. Fix the code, or open a ticket and reference it from a `TODO`.
- **Writing the broad instruction first** - prefer the narrowest carrier the trial proves: a
  barrier check before a prose rule, a skill before an always-loaded rule. The placement policy
  itself belongs to `agent-instructions`; do not restate it here.
- **Proposing a rule that only an agent can enforce** - prose is advisory and the next session may
  read past it. When the check is mechanical, its home is `scripts/verify.sh`.
- **Reflecting inside the failing session** - the context that produced the failure shares the
  premise that caused it. When the candidate concerns judgement rather than mechanics, have a fresh
  session state the cause independently.

## Constraints

- Never edit an instruction, a skill, a hook or the barrier without explicit approval of the trial.
- Never let a learned rule change its own evaluator, its evidence threshold or its promotion policy.
- Never include a secret, a credential, a private prompt or a raw transcript in a candidate.
- Never return more than one candidate; a list is a way of avoiding the choice.
- Never present recurrence, or an activation scenario, as proof of improvement: name the behavioural
  oracle and the environment it ran in.
