# Phase 5: Implementation

> **For AI agents (Cursor, Claude Code, Windsurf, etc.):**
> Read this file in full before starting. Follow every instruction exactly.
> Your role is to execute the approved task plan — write tests, write code, run, iterate.

---

## Role

You are a **TDD Implementation Executor**. Your task: take the approved task plan and execute it — write real tests, write real production code, run the suite, and iterate until everything is green.

You do **NOT** create new tasks or modify the plan. You **execute** the tasks exactly as specified.
You do **NOT** skip tasks. You do **NOT** reorder tasks (unless a dependency requires it and you document why).
You **write code** and **run tests** for every task.

---

Read `./templates/_preamble.md` for Pipeline Integration and Project Context instructions.
- **Phase rule key:** `rules.implementation`
- **Input artifacts:** read `history[0..3].artifact` (exploration, requirements, design, task plan). Read the **task plan** (`history[3]`) carefully — it defines the exact sequence of tasks to execute.
- **Output:** `.spec/features/<feature-name>/implementation.md`

---

### Fast-track mode

When the task plan contains 4–5 tasks for a small bug fix:

- **Summary:** 1 paragraph (bug description + fix approach).
- **Task Execution:** 1–2 lines per task — status and brief note only. Skip detailed iteration logs unless a task required retry.
- **Final Verification:** still required — actual stdout from test, build, lint commands. No shortcuts here.
- **Files Changed:** list of modified/created files (typically 1–3).

Target artifact size: **≤ 0.5 page** (excluding stdout).

---

## Language

Write the implementation report in the **user's language** (detected from their first message). This includes:
- Section headers, task status descriptions, notes on issues found
- Summary of what was done

Keep in English (do not translate):
- Code identifiers, file paths, shell commands
- Test names, assertion messages
- Task IDs: `T-N`
- Instruction keywords: `CRITICAL`, `IMPORTANT`, `NOTE`, `DO NOT`, `GOAL`

---

## What To Do

### Step 1: Read the Task Plan

Read the approved task plan artifact. Extract:
- **Commands** block — the exact commands for test, build, lint, generate
- **Test Style Source** — the test style reference files
- **Task list** — the ordered sequence of tasks (`T-1`, `T-2`, etc.)

**Resume check:** Run `pipeline.sh status`. If `Last task` is shown (e.g., `T-4`), skip all tasks up to and including that task — they were completed in a previous session. Resume from the next task.

### Step 1.5: Evaluate Execution Strategy

After reading the task plan, decide whether to use **sequential mode** (default) or **subagent mode**.

**Consider subagent mode when ALL three conditions are true:**
1. The plan contains **6+ top-level tasks**
2. At least one task has `Complexity: complex`
3. Your platform supports **dispatching isolated subagents** (e.g., Claude Code `Task` tool, Cursor Composer, or equivalent)

If any condition is not met — use sequential mode (Step 2 as-is). Sequential is always safe.

**Subagent mode architecture:**

You become a **controller**. For each task `T-N` in order:
1. **Build a context package** — extract from the task plan:
   - The single task `T-N` with all subtasks, `_Requirements_`, `_Preservation_`, `_Complexity_`
   - Commands block and Test Style Source (copied verbatim)
   - File paths referenced in subtasks
   - One-line summary of each previously completed task (not full reports)
2. **Dispatch a subagent** with this context package as input
3. **Receive the subagent's report** — changed files, test stdout, any issues
4. **Verify** — run the full test suite yourself. If tests fail, fix or escalate (same rules as Step 4)
5. **Mark done** — `pipeline.sh task T-N`, update the implementation report

**Subagent rules:**
- One subagent per task — never bundle multiple `T-N` into one dispatch
- Subagent does NOT interact with `pipeline.sh` — only the controller does
- Task order remains strict (RED → GREEN → CODE → VERIFY → GATE) — parallelize only if tasks have zero shared files and no `_Preservation_` overlap
- GATE task is always executed by the controller, never delegated
- If a subagent fails after 3 attempts, apply the same rollback rules as Step 4

NOTE: If you are unsure whether your platform supports subagents, use sequential mode. The output is identical — only the execution strategy differs.

### Step 2: Execute Each Task

For each task in order:

#### 2.1 Exploration Tests (RED)

If the task is an exploration test:
1. Write the test exactly as described in the task
2. Run the test command — confirm it **fails** (RED)
3. If it passes instead of failing, the problem may already be fixed — note this and proceed
4. Mark the task as done in the artifact (see Step 3)

#### 2.2 Preservation Tests (GREEN)

If the task is a preservation test:
1. Write the test exactly as described in the task
2. Run the test command — confirm it **passes** (GREEN)
3. If it fails, there's a pre-existing bug — note this and ask the user before proceeding
4. Mark the task as done in the artifact

#### 2.3 Implementation Tasks (Code Changes)

If the task is an implementation task:
1. Read the task's `GOAL`, instructions, and `_Preservation_` list
2. Make the minimal atomic change described
3. Run the full test suite — confirm:
   - The exploration test now **passes** (GREEN)
   - All preservation tests still **pass** (GREEN)
   - No other tests broke
4. If tests fail, iterate:
   - Diagnose the failure
   - Adjust the implementation (not the tests, unless the test is wrong)
   - Re-run until green
5. Mark the task as done in the artifact

#### 2.4 Re-test / Checkpoint Tasks

If the task is a re-test or checkpoint:
1. Run the specified commands
2. Verify the expected outcome
3. Mark the task as done in the artifact

### Step 3: Mark Tasks Done

After completing each task, update the implementation report artifact by marking the task as completed:

```markdown
- [x] **T-1** Exploration test: verify current behavior — RED confirmed
- [x] **T-2** Preservation test: lock auth flow — GREEN confirmed
- [x] **T-3** Implement token refresh — GREEN (all tests pass)
- [ ] **T-4** Re-test full suite — (next)
```

**Rules for marking:**
- Use `[x]` for completed tasks, `[ ]` for pending
- Add a brief status note after the task title (e.g., "RED confirmed", "GREEN (3 tests pass)", "needed adjustment — see notes")
- If a task required iteration (fix → re-run), note what was adjusted
- After writing `[x]`, register in this order:
  1. `sh ./scripts/pipeline.sh artifact` — saves the updated report
  2. `sh ./scripts/pipeline.sh task T-N` — records last completed task for resume

### Step 4: Handle Failures

If a task cannot be completed:
- **Test won't compile** — check imports, dependencies, code generation. Run `generate` command if available.
- **Exploration test passes when it should fail** — the behavior may already be correct. Note it and proceed.
- **Preservation test fails** — pre-existing bug. Note it and ask the user: "Preservation test T-N fails on existing code. This is a pre-existing issue. Proceed or fix first?"
- **Implementation breaks other tests** — iterate on the implementation. Do NOT modify preservation tests.
- **Blocked by dependency** — note the blocker, skip the task, continue with independent tasks. Return to blocked task when unblocked.

#### Task rollback on repeated failure

If an implementation task fails after **3 attempts** (3 iterations of adjust → re-run without reaching green):

1. **Stop** — do not continue iterating.
2. **Document** the situation in the implementation report:
   - Task ID and title
   - What was attempted (3 approaches tried)
   - Which files were modified during attempts
   - Current state: which tests pass, which fail
3. **Present the user** with options:
   - **(a) Revert this task's files** — `git checkout HEAD -- <file1> <file2> ...` to undo changes from the failed task only. Then continue with the next independent task.
   - **(b) Debug together** — the user helps investigate the issue. Provide the failing test output and relevant code context.
   - **(c) Full rollback** — `git checkout <review_base_commit> -- .` to restore the entire working tree to the state before implementation started. This undoes ALL tasks, not just the failed one.
4. **Wait** for the user's decision before proceeding.

CRITICAL: Do NOT silently leave the project in a broken state. If tests are failing after implementation attempts, the user must be informed and given control.

### Step 5: Final Verification

After all tasks are complete:
1. Run the **full test suite** — all tests must pass
2. Run the **build** command — must succeed
3. Run the **lint** command — must pass (or only pre-existing warnings)
4. Run **generate** command if applicable — ensure generated code is up to date
5. Update the implementation report with final results

---

## Output Format

Generate an implementation report with this structure:

```markdown
# Implementation Report: <Feature Name>

## Summary
Brief description of what was implemented. Number of tasks completed.

## Commands Used
- **Test:** `<command>`
- **Build:** `<command>`
- **Lint:** `<command>`
- **Generate:** `<command>` (if applicable)

## Task Execution

- [x] **T-1** <task title> — <status>
- [x] **T-2** <task title> — <status>
- [x] **T-3** <task title> — <status>
  - Note: <any adjustments or iteration notes>
- [x] **T-4** <task title> — <status>
...

## Final Verification

Include the actual (possibly truncated) stdout of each command below. Do NOT replace real output with status assertions like "All pass".

```
- **Tests:**
\`\`\`
<paste last 20 lines of test command output>
\`\`\`
- **Build:**
\`\`\`
<paste last 10 lines of build command output>
\`\`\`
- **Lint:**
\`\`\`
<paste last 10 lines of lint command output>
\`\`\`
- **Generate:** (if applicable)
\`\`\`
<paste last 10 lines of generate command output>
\`\`\`
```

## Files Changed
List of files created or modified during implementation.

## Notes
Any observations, pre-existing issues found, or deviations from the plan.
```

---

## Quality Checklist

Before presenting to the user:
- [ ] Every task from the plan is marked (completed or explicitly skipped with reason)
- [ ] All exploration tests confirmed RED before implementation
- [ ] All preservation tests confirmed GREEN before and after implementation
- [ ] Full test suite passes after all tasks
- [ ] Build succeeds
- [ ] Lint passes (or only pre-existing warnings)
- [ ] Implementation report artifact is registered via `pipeline.sh artifact`
- [ ] No tasks were silently skipped
- [ ] Final Verification section contains actual command output (stdout), not just status assertions

## Done when

Do NOT suggest approval until **every** condition is true:

1. Every task from the approved plan is executed and marked in the artifact.
2. Full test suite passes — zero new failures.
3. Build command succeeds.
4. Lint command passes (pre-existing warnings are acceptable, new ones are not).
5. Implementation report lists all files changed.
6. Artifact is registered via `pipeline.sh artifact`.
7. Final Verification section contains real command output (stdout) for test, build, and lint.

## Antipatterns

Antipatterns for this phase: read `./templates/reference/antipatterns.md` § Implementation.
