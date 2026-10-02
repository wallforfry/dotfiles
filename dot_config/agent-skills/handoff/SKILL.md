---
name: handoff
description: >
  Hand the current work to a fresh session instead of letting the context compact. Use when the
  window is nearly full, a Stop hook reports the handoff threshold, or a resume prompt is asked
  for. Make sure to use it whenever a session must continue elsewhere, even if the context limit
  is never named.
metadata:
  category: ops
---

# Handoff

## Overview

Produce a short, copy-pasteable prompt that lets a fresh session resume the current work, and then
stop. Compaction summarises a transcript and keeps its noise; a handoff restates only the live
state, so the next session starts small and accurate.

`~/.claude/hooks/agent-handoff` triggers this skill automatically once context usage passes the
threshold, but it is also useful on demand. That trigger exists on Claude only: Codex has a `Stop`
event, but a handler goes live only once an operator approves its content hash by hand, so no
deployment can arm it (ADR-021 of the dotfiles checkout). Elsewhere, invoke the
skill yourself when the window gets tight.

## Usage

```text
/handoff
```

No options. The output is a single fenced block for the user to paste into a new session.

## Steps

1. Stop the current work - do not start a new edit, search, or tool loop beyond steps 2 and 3.
2. Finish making the work durable: save unsaved files and, if a change is complete and the user asked
   for it, commit it. A resume prompt pointing at lost edits is worthless.
3. Only when this session obtained a semctx planning bundle from `semctx_control_plan_change`, in a
   repository holding `.semctx/`, for a write task the user authorized: make one
   `semctx_control_handoff` call with that bundle and the current progress, without loading any
   other skill or tool. Keep the returned capsule hash; when the call is refused or errors, keep its
   reason instead and do not retry. In any other session, including a read-only one, whose lane
   forbids a handoff write, or one resumed from a capsule that did not plan again, make no semctx
   call and skip this step.
4. Write the resume prompt as one fenced block, addressed to the next agent, in the language of the
   conversation, covering exactly:
   - **Goal** - the task in one or two sentences, including the user's own constraints.
   - **Done** - what is already done and verified, with file paths.
   - **Next step** - the single next concrete action.
   - **Files** - the paths the next session needs to read first.
   - **semctx** - only after step 3: the capsule hash and the repository root it was captured in,
     with the instruction to run `semctx_semantic_check`, then `semctx_control_resume` on that hash,
     before any edit, as `semctx:semctx-control` requires of a fresh context; or the refusal reason.
     Omit the line otherwise.
5. Keep it under ~200 words. Name files instead of quoting them; the next session can read them.
6. When the work happens in a git worktree, name the worktree path and branch under **Files** - the
   next session starts in the primary working directory otherwise.
7. End your turn immediately after the block. Do not add follow-up work or offer to continue.

## Gotchas

- **Summarising the conversation instead of the state** - a handoff is not a transcript summary. Drop
  abandoned approaches, tool noise, and anything the next agent can read from a file; keep only what
  it needs to act.
- **Emitting the prompt and then continuing to work** - the point is to end the session before
  compaction. Any further tool call adds context and defeats it, so stop after the block.
- **Leaving work uncommitted** - the next session inherits the working tree, not the reasoning. State
  explicitly under **Done** whether changes are committed, staged, or only on disk.
- **Being triggered mid-task by the hook** - the threshold fires at the end of a turn regardless of
  where the work stands. Say plainly under **Next step** that a step is half-done, rather than
  implying it is complete.
- **Expecting the hook to fire with no known window** - it needs `CLAUDE_CODE_AUTO_COMPACT_WINDOW` or
  `HANDOFF_TOKEN_THRESHOLD`. Invoke the skill manually when neither is set.
- **Waiting for the hook on another host** - no host but Claude arms it. Waiting there means
  compacting instead of handing off.
- **Calling `semctx_control_handoff` without a planning bundle** - the request requires the bundle
  that `semctx_control_plan_change` returned; one assembled for the occasion binds the capsule to
  no reviewed plan and resumes as a false continuity. Skip the step instead.
- **Passing the capsule hash without its repository root** - the capsule lives in the ignored
  working state of the checkout that captured it, so a resume against another worktree finds
  nothing. Name that root next to the hash.
- **Reading a resumed capsule as completed work** - its progress is a requested boundary, not an
  execution history, and a stale state resumes as a null capsule. The next session revalidates
  the capsule before trusting its progress.
- **Writing the block in English out of habit** - the section labels above are English because this
  file is; the block itself follows the conversation's language.

## Constraints

- Never keep working after emitting the handoff block - end the turn there.
- Never invent progress: only claim what was actually run and verified in this session.
- Keep the block self-contained - the next session sees no part of this conversation.
- Do not write the resume prompt to a file unless the user asks; the deliverable is text to paste.
  The semctx capsule of step 3 is semctx working state, not the resume prompt, and needs no request.
- Do not attempt to disable or block compaction from the skill - that is the hook's job.
- Never call a semctx tool unless this session holds a semctx planning bundle, and never invent one
  to obtain a capsule hash.
