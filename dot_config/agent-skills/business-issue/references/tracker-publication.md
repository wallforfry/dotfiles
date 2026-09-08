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
| Linear | Workspace, team and project are distinct; inspect the team's actual workflow and existing issue relations |
| GitHub Issues / Projects | Identify whether the requested artifact is a repository issue or a project draft item; an issue and its membership in a project are separate operations |
| Other tracker | Discover its work-item type, description support, ownership, states and native relations rather than assuming either model above |

Use native hierarchy and dependencies where supported. When the destination cannot represent one,
use an explicit link and describe the limitation instead of claiming a native relation was created.
Honor draft-item requests; do not silently turn them into repository issues.

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

If the outcome is uncertain, read the destination before retrying. If an issue exists but adding it
to a project fails, keep that issue, report the partial result and retry only the missing operation.
Do not create a second issue to repair missing membership. If access is revoked, stop external
writes and return the draft or known identifiers with the unresolved operation.

## Verify the result

Read back the description and acceptance criteria, target scope, parent/dependencies, project
membership and fields that were meant to remain unchanged. Report what exists, what is still
blocked and what was not checked. A created item is not evidence of product delivery.
