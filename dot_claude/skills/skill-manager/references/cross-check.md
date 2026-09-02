# Cross-Check Skills

`/skill-manager cross-check` analyses `dot_claude/skills/` as a whole and reports what only shows up
between skills. It is **read-only**: it produces a report and stops. Run `doctor` first - a skill
with a malformed description or no `Constraints` section poisons D1 and D4 with noise.

## Procedure

1. List the skill directories, excluding `README.md`, and read each `SKILL.md`.
2. Extract, per skill: `name`, `description`, the trigger phrases after `Use when` and
   `Make sure to use it whenever`, the `## Constraints` entries, the other skills it mentions, its
   functional domain, and the file names under `references/` without reading them yet.
3. Run the five detectors below, reading reference bodies only where a detector demands it.
4. Produce the report, present it, and stop.

## D1 - Trigger overlap

Two descriptions competing for the same activation. Tokenise each description, dropping
`use`, `when`, `make`, `sure`, `it`, `this`, `skill`, `even`, `if`, `the`, `a`, `an`, `for`, `with`,
`to`, `and`, `or`, `in`, `on`, `is`, `are`. Compute `|A ∩ B| / |A ∪ B|` over every pair.

- 40-64 percent: WARN.
- 65 percent and above: CRITICAL.

```text
[D1] Trigger overlap: <skill-A> / <skill-B> - <X>%
  Shared tokens: <list>
  Risk: <skill-A> may activate where <skill-B> is intended
```

Heuristic, not proof: two complementary skills can legitimately share vocabulary.

## D2 - Content duplication

The same procedure written twice, which will diverge. Two passes, to bound the reading:

1. compare only the `## Steps` and `## Constraints` sections of the `SKILL.md` files;
2. for each pair above roughly 30 percent similarity, read their `references/` and compare those.

Flag two or more consecutive steps that are structurally similar: same verbs, same tools, same
order.

```text
[D2] Content duplication: <skill-A>[/<file>] § Steps / <skill-B>[/<file>] § Steps
  Similar steps: "<excerpt>"
  Recommendation: extract into a shared reference, or into one skill
```

## D3 - Dead reference

A skill naming a skill that does not exist. Search for slug-shaped mentions only: backticked slugs
without a `/`, `[text](<slug>)`, `See <slug>`. Exclude anything containing a `/` (a path such as
`docs/adr/README.md`), any URL, any namespaced identifier such as `superpowers:brainstorming`, and
the subcommands of this skill (`/skill-manager doctor`). Check each extracted slug against the
directory listing.

```text
[D3] Dead reference: <skill-A> mentions "<slug>" - no such directory
  Line: <number or excerpt>
  Recommendation: fix the name, or create the skill
```

## D4 - Rule contradiction

Two skills whose `## Constraints` oppose each other on the same subject: "always X" against "never
X", "prefer X" against "avoid X". Include the normative sections of reference files. Some
contradictions are legitimate when the contexts differ; say which.

```text
[D4] Rule contradiction: <skill-A>[/<file>] vs <skill-B>[/<file>]
  Rule A: "<full text>"
  Rule B: "<full text>"
  Subject: <X>
  Recommendation: align the rules, or state the context each applies to
```

## D5 - Slug ambiguity

Two slugs close enough to confuse the router: same root with a different suffix
(`git-commit` / `git-commits`), Levenshtein distance of 2 or less, or the same domain with a
synonymous verb (`pr-create` / `pr-open`). Informational.

## D6 - Scoped reference conflict

Only for `references/<topic>-<scope>.md` files that **have a same-topic sibling** in the same
directory. A two-word file name with no sibling - `cross-check.md`, `sync-index.md` - is an ordinary
reference and is skipped entirely.

- Same topic, same skill, two scopes, near-identical content: WARN, the split may be pointless.
- Same topic, two skills, substantially similar: WARN, extract a shared reference.
- Scoped siblings with no conditional routing in `## Steps`: CRITICAL, the agent cannot know which
  to load.

## Report

```text
# Cross-Check Report - dot_claude/skills/

Analysed: <N> skills | Detectors: D1 D2 D3 D4 D5 D6

## Critical
...

## Warnings
...

## Info
...

| # | Detector | Severity | Skills | Suggested action |
| --- | --- | --- | --- | --- |

No file was modified.
```

Close by asking which findings should be fixed, and by which `fix` operation.

## Constraints

- Never modify a file during cross-check; refuse an inline fix and point at `fix <slug>`.
- Never read outside `dot_claude/skills/` for a detector; if an outside read seems necessary, finish
  the report first and ask.
- Never report the absence of an activation router as a finding.
- Never present D1, D2 or D6 similarity scores as proof; they are indicators, and the files decide.
- Never run cross-check as a substitute for `doctor`; it complements it, and comes after it.
