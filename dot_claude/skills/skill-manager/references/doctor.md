# Doctor Skills

## Scope

- `/skill-manager doctor` audits every directory under `dot_claude/skills/`.
- `/skill-manager doctor <slug>` audits one skill.
- `README.md` is not a skill directory.

Doctor is read-only: it reports findings for a later `fix`.

## Status model

- **FAIL** - a standard rule or a mandatory local convention fails, or `scripts/verify.sh` is red on
  this skill.
- **WARN** - a qualitative, non-blocking weakness.
- **PASS** - no FAIL and no WARN.
- A missing optional tool is an environment limitation, reported once, never a status downgrade.

## Procedure

1. Read `conventions.md` completely.
2. Enumerate the requested directories.
3. Run `bash scripts/verify.sh` once for the whole run and attribute its Skills findings per skill.
4. Apply the checks below to each skill.
5. Produce one report per skill, then a summary table.
6. Propose the exact correction for every finding, and modify nothing.

## Frontmatter

- `name` matches `[a-z0-9]+(-[a-z0-9]+)*`, is 1-64 characters and equals the directory;
- `description` is a non-empty string of at most 1024 characters;
- `compatibility`, if present, is non-empty and at most 500 characters;
- `allowed-tools`, if present, is a space-separated string, and is reported as experimental;
- `metadata.category` is `dev` or `ops`;
- no top-level field outside `name`, `description`, `license`, `compatibility`, `allowed-tools`,
  `metadata`;
- the description leads with the distinguishing case, carries concrete `Use when` conditions and
  stays under the 400-character target. A weak but valid description is WARN.

## Body

- one H1 after the frontmatter;
- `Overview`, `Usage`, `Steps` or `Workflow`, `Gotchas`, `Constraints`, in that order;
- at least three gotchas, each naming a cause, a consequence and a correction;
- at least three hard constraints;
- under 500 lines, or the detail is disclosed through `references/` and routed from the body;
- no em dash and no middle dot, per `harness/SOUL.md`.

## Resources and routing

- only needed `references/`, `scripts/` and `assets/` directories exist, and all are tracked;
- same-topic scoped siblings have explicit conditional routing in `SKILL.md`;
- identical content is not duplicated across two references;
- no chezmoi attribute on `SKILL.md` or on a reference; `scripts/` executables carry `executable_`;
- the absence of an activation router is never a finding; a router rule is WARN unless its repeated
  behavioural evidence is identified.

## Templated shell

Search the `SKILL.md` body for unescaped `$0` to `$9`, including `${1}`, plus `$@` and `$ARGUMENTS`.
Executable shell containing a match is FAIL and moves to `scripts/`. A literal token in prose is
escaped once. Do not scan `references/` or `scripts/`: they are not templated.

## Report format

```text
### <slug> [PASS | WARN | FAIL]

Barrier (verify.sh): PASS | FAIL: <finding>
Frontmatter: PASS | <exact finding>
Body: PASS | <exact finding>
Resources: PASS | <exact finding>
Templated shell: PASS | <exact finding>
Evals: PASS | absent (optional) | <exact finding>
Index: PASS | <exact finding>
Action needed: none | <one exact corrective action per finding>
```

Then one summary table:

```text
| Skill | Status | Barrier | Local | Action |
| --- | --- | --- | --- | --- |
| <slug> | PASS | PASS | PASS | none |
```

## Common findings

| Finding | Exact correction |
| --- | --- |
| Non-standard top-level field | Remove it, or ask before mapping it under `metadata` |
| Category outside `dev`/`ops` | Choose one, or extend the three places that list them |
| Description above the local target | Keep the first specific clause and the triggers; move the detail into the body |
| Missing `Gotchas` | Add three repository-specific cause, consequence and correction entries |
| Missing `Constraints` | Add three hard must or must-not rules |
| Positional placeholder in the body | Move the shell to `scripts/`; escape literal prose once |
| Missing conditional routing | Route each scoped sibling from `Steps` |
| Missing index row | Run `sync-index` once the skill itself passes |
| Em dash or middle dot | Replace with a hyphen or an epicene wording |

## Constraints

- Never modify a file during doctor.
- Never report a red `verify.sh` as PASS.
- Never fail a skill for lacking an activation router or an eval file.
- Never turn a qualitative judgement into a mandatory finding without a rule in `conventions.md`.
- Always give a reproducible finding and an exact correction for every FAIL and WARN.
