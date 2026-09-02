# Activation Scenarios

## What the file is

`evals/trigger-queries.json` records realistic prompts that should, and should not, activate a
skill. It is a scenario file, not a result: this repository runs no eval harness, and nothing
executes these prompts on its own.

Write one when the activation behaviour matters, or when a description is about to be rewritten for
routing reasons. Absence is never a finding.

## Schema

```json
{
  "skill": "example-skill",
  "version": "1.0",
  "queries": [
    {
      "query": "Create an example artefact",
      "should_activate": true,
      "reason": "Direct request for the skill's workflow"
    },
    {
      "query": "Review an unrelated pull request",
      "should_activate": false,
      "reason": "merge-verdict owns pull request review"
    }
  ]
}
```

Doctor requires valid JSON, a `skill` equal to the directory name, a string `version`, a non-empty
`queries` array, a string `query`, a boolean `should_activate` and a string `reason` in every entry,
and at least one positive and one negative case.

## Writing scenarios

- Use phrasing a user would actually type, in the language they would type it.
- Cover the direct trigger and the implicit case that does not name the domain.
- Add negatives from the adjacent skill most likely to false-positive.
- Explain the expected boundary in `reason`, not the implementation.
- Keep the scenarios stable while comparing two descriptions: changing both at once measures
  nothing.

## Running them

Use a fresh session per prompt, so an earlier conclusion does not leak into the next. Give the
session the prompt and the repository, never the expected answer. Record whether the skill activated
and whether its procedure actually governed the response - a skill that loads and is then ignored
counts as a failure.

Report the result as counts and name the environment: "9 of 11 prompts routed as expected, Claude
Code on macOS". An unexecuted scenario file proves nothing and must never be reported as a rate.
