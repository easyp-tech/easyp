# Task Type Templates

Templates for each task type in the TDD implementation plan. Referenced from `task-plan.md`.

---

## RED: Exploration Test (for bug fixes)

> GOAL: Demonstrate the defect exists before any fix is applied.

```
### Task: Write exploration test for <defect description>

*_Requirements: X.Y_*
*_Bug_Condition: <condition that triggers the bug>_*
*_Expected_Behavior: <what should happen instead>_*

CRITICAL: This test MUST FAIL when run against the unmodified codebase.
IMPORTANT: Do not fix anything yet. The test exists only to confirm the bug is reproducible.
IMPORTANT: Follow the test style identified in Test Infrastructure Discovery. Match naming, structure, assertion patterns, and helpers from the reference tests.
DO NOT: Write more than one test in this task. One defect = one exploration test.

Instructions:
1. Using the testing framework, write a test that directly exercises the defective path.
2. Run: `<test command>`
3. Confirm the test fails with the expected failure message.
4. Commit the failing test as evidence of the defect.
```

---

## GREEN: Preservation Test

> GOAL: Lock correct behavior in areas adjacent to the change so it cannot be accidentally broken.

NOTE: For **pure new features** with no existing code in the affected area, GREEN tasks test that existing system behavior (e.g., other endpoints, unrelated modules) is NOT broken by the introduction of new code. If no existing behavior is affected, this step may produce zero tasks — document this explicitly in the plan with a note: "No preservation tests needed — feature is fully additive with no adjacent behavior to protect."

```
### Task: Write preservation tests for <component or behavior>

*_Requirements: X.Y_*

IMPORTANT: These tests must pass BEFORE any implementation changes are made.
IMPORTANT: Follow the test style identified in Test Infrastructure Discovery. Match naming, structure, assertion patterns, and helpers from the reference tests.
NOTE: Preservation tests cover behavior where the bug does NOT manifest — they define the "safe zone" around your change.
DO NOT: Modify production code during this task.

Instructions:
1. Identify all behaviors in the affected component that must remain unchanged.
2. For each behavior, write a test using the testing framework.
3. Run: `<test command>`
4. All preservation tests must pass (GREEN). If any fail, stop and investigate.
5. Commit the preservation tests before touching production code.
```

---

## CODE: Implementation Task

> GOAL: Make the minimal atomic change that satisfies the failing exploration test without breaking preservation tests.

```
### Task: Implement <specific change>

*_Requirements: X.Y_*
*_Preservation: <list of correctness properties that must hold>_*

CRITICAL: Change only the file specified in this subtask. One subtask = one file.
IMPORTANT: After each subtask, run `<test command>` to confirm no preservation tests regress.
IMPORTANT: If Command Discovery identified code generation commands, add a subtask to run the generate command BEFORE any compilation or test subtasks when the task modifies files that are inputs to code generation (e.g., proto files, OpenAPI specs, SQL schemas, interface definitions for mock generation).
DO NOT: Refactor unrelated code. DO NOT introduce new abstractions not in the design document.

Subtasks:
- [ ] 1. <Action in file A> — `<test command>`
- [ ] 2. <Action in file B> — `<test command>`
- [ ] 3. <Action in file C> — `<test command>`

After all subtasks: Run `<build command>` and `<lint command>` to confirm no compilation or style errors.
```

---

## VERIFY: Re-test

> GOAL: Confirm the exploration test now passes after the implementation.

```
### Task: Re-run exploration test for <defect description>

*_Requirements: X.Y_*

CRITICAL: This is the SAME test written in RED. Do not modify it.
GOAL: The test must now pass (GREEN). If it still fails, the implementation is incomplete.

Instructions:
1. Run: `<test command> <exploration-test-name>`
2. Confirm the test passes.
3. Run the full suite: `<test command>`
4. Confirm all preservation tests still pass.
```

---

## GATE: Checkpoint

> GOAL: Verify the entire implementation is complete, all tests pass, and all requirements are covered.

```
### Task: Checkpoint — verify full coverage

*_Requirements: ALL_*

CRITICAL: This task must be the LAST task in the plan. Do not add it before all other tasks are complete.

Instructions:
1. Run the full test suite: `<test command>`
2. Confirm 100% of tests pass (GREEN).
3. Run: `<build command>` — confirm no errors.
4. Run: `<lint command>` — confirm no violations.
5. Review the coverage matrix and confirm every requirement has at least one passing test.
6. Confirm no orphan tasks remain (every task traceable to a requirement).
7. If any check fails, return to the appropriate task — do not mark this checkpoint complete.
```
