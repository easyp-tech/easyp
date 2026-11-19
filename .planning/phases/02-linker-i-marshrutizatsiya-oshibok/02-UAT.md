---
status: testing
phase: 02-Линкер и маршрутизация ошибок
source: [02-VERIFICATION.md]
started: 2026-11-19T06:00:00Z
updated: 2026-11-19T06:00:00Z
---

## Current Test

number: 1
name: Ratify flagged assumption A1 — two-level class-1 verification before the Phase 5 flip
expected: |
  Explicit acceptance (or rejection and a gap-closure round) of the two-level interpretation of
  ROADMAP SC-4: class-1 (COMPILE) verification before the Phase 5 flip is satisfied by the
  classifier cells (02-01/02-02) + printIssues golden (02-03), with the identical matrix re-run
  through Action in Phase 5. Class-1 through-Action is physically unreachable while the linker is
  dormant (D-04 byte-unchanged lint.go/api/easypc).
awaiting: user response

## Tests

### 1. Ratify flagged assumption A1 — two-level class-1 verification before the Phase 5 flip
expected: |
  Explicit acceptance (or rejection and a gap-closure round) of the two-level interpretation of
  ROADMAP SC-4: class-1 (COMPILE) verification before the Phase 5 flip is satisfied by the
  classifier cells (02-01/02-02) + printIssues golden (02-03), with the identical matrix re-run
  through Action in Phase 5. Class-1 through-Action is physically unreachable while the linker is
  dormant (D-04).
result: [pending]

### 2. Resolve the MVP-mode discrepancy (Mode: mvp, goal is not a User Story)
expected: |
  Either ratify the Phase 1 precedent (goal-backward verification against the four ROADMAP success
  criteria for this internal compiler-migration milestone; inventing a story framing would be
  fabrication) or run /gsd:mvp-phase 02 to set a proper User Story goal and re-verify.
result: [pending]

### 3. Re-pin the warn-drop policy when a natural warn fixture appears (Phase 5 matrix re-run)
expected: |
  A test cell exercising the IsWarn drop branch (link.go:45-47) with a real warn diagnostic — no
  natural warn fixture exists in this phase; branch is unexercised code (flagged assumption A3
  from plan 02-02; Phase 5 re-pin).
result: [pending]

### 4. Confirm retention of the etalon-verbatim defensive fallback branch
expected: |
  The defensive fallback branch (!errors.Is(err, protoc.ErrCompilationFailed) -> (nil, nil)) stays
  with its not-unit-triggerable comment (plan 02-02 documents it as deliberate, never deleted).
result: [pending]

## Summary

total: 4
passed: 0
issues: 0
pending: 4
skipped: 0
blocked: 0

## Gaps

None — 8/10 must-haves verified with re-derived evidence; all four roadmap SCs hold at the
reachable level; status is human_needed solely for the four judgment items above. Deferred to
Phase 5 (informational, matching Phase 5 goal text): WR-01, WR-02, full class-1 matrix re-run.
Known non-gap: pre-existing wazero -race failure in internal/adapters/plugin (Phase 1
acknowledged gap; dedicated fix needed before the Phase 7 final gate).
