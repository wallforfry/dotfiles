# Fix a Skill

## Scope

`/skill-manager fix <slug>` applies one justified change set to one skill. It accepts exactly two
inputs:

1. reproducible findings from a current `doctor` or `cross-check` report;
2. an explicitly requested behaviour change, including to a skill that passes.

With neither, refuse to edit and ask what behaviour should change. A passing skill is not a defect.

## Procedure

1. Read `conventions.md` completely.
2. Identify the slug and the input type: findings, or requested evolution.
3. Record the baseline: `doctor <slug>` plus the Skills section of `bash scripts/verify.sh`, unless
   a current cross-check report is the only input.
4. For an evolution, write the requested contract as one testable sentence.
5. Present the findings or the contract before writing anything.
6. Apply only what the input authorises, in the order below.
7. Rerun `doctor <slug>`, `verify.sh`, and the eval scenarios when activation changed.
8. Compare every baseline PASS and the recorded contract against the new state.
9. Run `sync-index` and require a byte-identical second run.

## Correction order

1. **Frontmatter.** Remove any field outside the six standard ones. Move a top-level `category`
   under `metadata`. Fix types and the `name` constraint before touching prose.
2. **Description and activation.** Distinguishing case first, concrete triggers, under the
   400-character target. Do not add a router because none exists.
3. **Body.** Add or repair `Overview`, `Usage`, `Steps` or `Workflow`, `Gotchas`, `Constraints`, in
   order, with three concrete entries in each of the last two. Do not invent domain behaviour to
   fill a section: ask when the missing rule is not derivable.
4. **Templated shell.** Move executable shell with positional placeholders into `scripts/` and call
   it from the body. Escape a literal token in prose once.
5. **Progressive disclosure.** Move detail out of a `SKILL.md` approaching 500 lines, split a
   reference only when behaviour genuinely differs, and route the siblings from `Steps`.
6. **Evals and index.** Repair a present eval file to its schema; never create one just because it
   is optional. Regenerate the README only once the skill itself passes.

## Regression handling

Every baseline PASS stays PASS. When a check regresses, revert only the edit that caused it,
investigate, and rerun `doctor`. Never use `git reset` or `git checkout <path>` to tidy up: this
repository is worked on in worktrees whose sibling changes are not yours to discard, and a
`git checkout` on a file you have just written destroys it.

An evolution succeeds only when its recorded contract is demonstrated and unrelated behaviour is
unchanged. A formatting-only diff is not evidence of anything.

## Checklist

- The input is a current finding or an explicit behaviour change.
- The baseline was recorded before the first write.
- `conventions.md` was read.
- Only the authorised files changed.
- `doctor` and `verify.sh` pass afterwards.
- No baseline PASS regressed.
- `sync-index` is byte-identical on its second run.
- `chezmoi diff` was read when a file was added, renamed or deleted.

## Constraints

- Never edit a passing skill without an explicit request for a behaviour change.
- Never modify more than one skill in one `fix` operation.
- Never invent a finding or a domain rule to justify a write.
- Never apply a correction inside the read-only `cross-check` operation.
- Never declare the fix done while `scripts/verify.sh` is red.
