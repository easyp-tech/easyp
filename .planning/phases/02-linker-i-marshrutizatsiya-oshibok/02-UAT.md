---
status: complete
phase: 02-Линкер и маршрутизация ошибок
source: [02-VERIFICATION.md]
started: 2026-11-19T06:00:00Z
updated: 2026-11-19T10:35:38.643Z
---

## Current Test

[testing complete]

## Tests

### 1. Ratify flagged assumption A1 — two-level class-1 verification before the Phase 5 flip
expected: |
  Explicit acceptance (or rejection and a gap-closure round) of the two-level interpretation of
  ROADMAP SC-4: class-1 (COMPILE) verification before the Phase 5 flip is satisfied by the
  classifier cells (02-01/02-02) + printIssues golden (02-03), with the identical matrix re-run
  through Action in Phase 5. Class-1 through-Action is physically unreachable while the linker is
  dormant (D-04).
result: pass

### 2. Resolve the MVP-mode discrepancy (Mode: mvp, goal is not a User Story)
expected: |
  Either ratify the Phase 1 precedent (goal-backward verification against the four ROADMAP success
  criteria for this internal compiler-migration milestone; inventing a story framing would be
  fabrication) or run /gsd:mvp-phase 02 to set a proper User Story goal and re-verify.
result: pass

### 3. Re-pin the warn-drop policy when a natural warn fixture appears (Phase 5 matrix re-run)
expected: |
  A test cell exercising the IsWarn drop branch (link.go:45-47) with a real warn diagnostic — no
  natural warn fixture exists in this phase; branch is unexercised code (flagged assumption A3
  from plan 02-02; Phase 5 re-pin).
result: issue
reported: "Add a synthetic warn fixture now — the user chose a synthetic warn diagnostic test over the Phase 5 deferral"
severity: minor

### 4. Confirm retention of the etalon-verbatim defensive fallback branch
expected: |
  The defensive fallback branch (!errors.Is(err, protoc.ErrCompilationFailed) -> (nil, nil)) stays
  with its not-unit-triggerable comment (plan 02-02 documents it as deliberate, never deleted).
result: pass

## Summary

total: 4
passed: 3
issues: 1
pending: 0
skipped: 0
blocked: 0

## Gaps

- gap_id: G-02-3
  truth: "The IsWarn drop branch (internal/core/link.go:45-47) is exercised by a test with a warn diagnostic"
  status: failed
  reason: "User reported: Add a synthetic warn fixture now — the user chose a synthetic warn diagnostic test over the Phase 5 deferral"
  severity: minor
  test: 3
  root_cause: "No warn-producing proto fixture exists in this phase; the compiler driver rarely emits warn diagnostics, so the errorHandler closure's IsWarn drop never fires in any of the 16 natural cells"
  artifacts:
    - path: "internal/core/link.go"
      issue: "IsWarn drop branch at lines 45-47 unexercised by TestLinkSources"
  missing:
    - "A synthetic warn diagnostic driven through linkSources' diagnostic path (test seam or injected handler) asserting warns are dropped and never become COMPILE issues"
  debug_session: "Diagnosis pre-completed: research A3 + 02-VERIFICATION.md behavior_unverified item 10 document the root cause; no debug agent dispatched"

Deferred to Phase 5 (informational, matching Phase 5 goal text): WR-01, WR-02, full class-1 matrix
re-run. Known non-gap: pre-existing wazero -race failure in internal/adapters/plugin (Phase 1
acknowledged gap; dedicated fix needed before the Phase 7 final gate).
