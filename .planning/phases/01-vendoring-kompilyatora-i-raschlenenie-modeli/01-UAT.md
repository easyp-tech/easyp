---
status: testing
phase: 01-Вендоринг компилятора и расчленение модели
source: [01-VERIFICATION.md]
started: 2026-11-19T00:00:00Z
updated: 2026-11-19T00:00:00Z
---

## Current Test

number: 1
name: Ratify the QF1001 finding-scoped exclusion on casing.go (executor-requested)
expected: |
  A ratification decision on `linters.exclusions.rules: {staticcheck QF1001, paths: internal/rules/casing\.go}`
  (.golangci.yml lines 27-29) as the resolution of the constraint triad: D-01 byte-identity +
  "casing files are never excluded" + green lint gate, given upstream casing.go:49 carries QF1001
  and the lab's own config tolerates it. Alternatives: accept a red lint gate for verbatim files,
  or fix upstream and re-sync. Current scope silences only QF1001 for that one file; all other
  linters and staticcheck codes stay active on casing.go.
awaiting: user response

## Tests

### 1. Ratify the QF1001 finding-scoped exclusion on casing.go (executor-requested)
expected: |
  Ratification (or alternative) of the minimal point-exclusion `{staticcheck QF1001, paths:
  internal/rules/casing\.go}` in .golangci.yml — the only way to satisfy byte-identity + green
  lint simultaneously; plan 01-04 coverage row D4 is `human_judgment: true`.
result: [pending]

### 2. Acknowledge the VENDORED.md edit under the internal/easypc seal (01-01 P1 letter)
expected: |
  Acknowledgment that editing internal/easypc/VENDORED.md (commit 31885b4, planned explicitly by
  01-02) does not violate the "no file under internal/easypc may be modified" prohibition —
  provenance maintenance is the intended exception (D-04 targets vendored upstream code; vendored
  code files have exactly one touching commit: 690a9cc).
result: [pending]

### 3. Acknowledge the go-billy indirect→direct marker move against the byte-identical tidy gate (01-01 T3 letter, REQ-1.2)
expected: |
  Acknowledgment that commit 8f969d8 (go-billy v5.6.2 promoted indirect→direct, dedicated
  pre-vendor commit) satisfies LINT-VEND-02's substance: no new module, no replace, no new
  require entry, go.sum untouched. The byte-identical tidy expectation was falsified by a
  pre-existing stale go.mod condition (reproduced on pristine HEAD), not by vendoring.
result: [pending]

## Summary

total: 3
passed: 0
issues: 0
pending: 3
skipped: 0
blocked: 0

## Gaps

None — all 23 must-haves verified with codebase and command evidence; status is human_needed
solely for the three flagged judgment-tier prohibitions above.
