# Tracker publication

Use this only for external creation or updates. Discover the actual tool interface and supported
fields through its documentation or help instead of retaining vendor commands in this skill.
Repository guidance may own transport and naming conventions; keep this skill's investigation and
contract as the common layer rather than running two complete issue-writing workflows.

## Resolve the destination

Honor the requested tracker, workspace, repository or team, and project. Verify their identities
live; the CLI's default scope may belong to another team. If ambiguity could publish to the wrong
audience, ask for the destination while completing the local draft. Use an authenticated connector
or existing CLI session; unavailable access must not prevent delivery of the draft.

| Destination | What to establish before publication |
| --- | --- |
| Linear | Workspace, team and project are distinct; inspect the team's actual workflow, issue relations, hierarchy and source-control integration |
| GitHub Issues / Projects | Identify whether the requested artifact is a repository issue or a project draft item; inspect native hierarchy, dependencies and development links because an issue and its project membership are separate operations |
| Other tracker | Discover its work-item type, description support, ownership, states and native relations rather than assuming either model above |

Probe the live connector, CLI help or API schema for the intended operations. Product documentation
describes a vendor, not necessarily the installed connector, repository configuration, plan or the
caller's permissions. Honor draft-item requests; do not silently turn them into repository issues.

## Plan the relation graph

Represent the requested semantics before translating them into vendor fields. Use one row per edge
with an explicit subject and target:

| Relation | Direction to record |
| --- | --- |
| `parent` | the subject issue belongs under the target issue |
| `child` | the subject issue contains the target issue |
| `blockedBy` | the subject issue cannot proceed until the target issue allows it |
| `blocking` | the subject issue must allow the target issue to proceed |
| `duplicate` | the subject issue duplicates the canonical target issue |
| `related` | the subject and target issues share useful context without hierarchy or dependency |
| `implementedBy` | the subject issue is expected to be delivered by the target pull request |
| `mentionedBy` | the subject issue receives context from the target pull request, which does not deliver it |

The names above describe intent, not fields to send blindly. Translate each edge to the tracker's
native direction and relation type. Use native hierarchy, dependency, duplicate and development
links when the live interface supports them. A URL in a description is not a native relation, and
a parent relation is not a substitute for a dependency: decomposition and execution order answer
different questions.

For an issue and pull request, decide the merge effect before linking them. Use a delivery or closing
link only when that pull request is expected to satisfy the issue and closing it on merge is intended.
Use a non-closing development link or a plain contextual reference for partial, investigative or
supporting work. Closing keywords, default-branch restrictions and source-control integrations differ;
inspect the destination's current behaviour rather than assuming that a textual reference links or
closes anything.

When the destination or available tool cannot represent an intended edge natively, retain it in the
issue's `Relations to establish` section as an explicit link, name the missing native capability and
report the fallback. Never claim that fallback text created native metadata.

## Mutate within the authorized scope

Before writing, re-read the target and search for equivalent work. Creation or update requests
authorize the stated work, not extra projects, reassignment of historical issues or migration of
an entire backlog. Respect already-granted authorization; do not add a confirmation step solely
because the artifact will be external. Ask only when the scope or a consequential choice remains
undecided, or when a binding approval requirement applies.

Preserve unrelated owners, labels, states, dates and relationships. Check whether collection fields
replace or merge before updating them. Reuse observed conventions rather than importing a fixed
taxonomy. Split a batch by dependency, keep the identifiers returned by successful operations, and
never infer success from an attempted request.

For a batch, create or resolve every issue and pull request first, retain their returned identifiers,
then add the planned edges. Before each relation mutation, read the existing edges: an identical edge
is already satisfied, while an inverse or conflicting edge must be resolved from the contract rather
than layered beside it. Apply each missing edge once. Do not reverse `blockedBy` into `blocking`, turn
`related` into a dependency, or attach a pull request with closing semantics merely because its URL
appears in the issue.

If the outcome is uncertain, read the destination before retrying. If an issue exists but adding it
to a project fails, keep that issue, report the partial result and retry only the missing operation.
Do not create a second issue to repair missing membership. If access is revoked, stop external
writes and return the draft or known identifiers with the unresolved operation.

## Verify the result

Read back the description and acceptance criteria, target scope, project membership and fields that
were meant to remain unchanged. For every planned edge, verify its type and direction from the issue
that owns the relation; where the destination exposes the inverse, verify that view too. For a pull
request, verify both the issue's development state and the PR's issue link, including whether merge
will close the issue. Report counts of native relations created, already present and replaced by
fallback links, plus every unchecked edge. A created or linked item is not evidence of product
delivery.
