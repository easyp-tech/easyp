---
phase: "02"
slug: "pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-11-21"
---

# Phase 02 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` + testify v1.11.1 |
| **Config file** | none — package-level `go test`; race flags from Taskfile convention |
| **Quick run command** | `go build ./... && go vet ./... && go test -count=1 ./internal/core ./internal/rules ./internal/protoast ./internal/api` |
| **Full suite command** | `go test -race -count=1 -timeout 20m ./...` (≈4–7 min; `internal/api` ≈211–309 s) |
| **Estimated runtime** | ~420 seconds |

---

## Sampling Rate

- **After every task commit:** Run quick run command
- **After every plan wave:** Run full suite command (once, in 02-02)
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 420 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | IMPORT-CONS-01, IMPORT-VENDRM-01 | T-02-01 / T-02-02 | 14 path-line swaps only; vendored tree fully deleted; no stale vendored references | integration | `grep -rn "internal/easypc/" --include='*.go' .` empty; `git ls-files internal/easypc` empty; `go build ./...`; `go vet ./...`; `gofmt -l` empty; quick `-race` tests | ✅ | ✅ green |
| 02-01-02 | 01 | 1 | IMPORT-DEP-01 | T-02-03 | tidy no-op, pins unchanged, no replace/vendor | integration | `TIDY_OUT=$(go mod tidy -diff); test -z "$TIDY_OUT"`; `go list -m` exact pins | ✅ | ✅ green |
| 02-02-01 | 02 | 2 | IMPORT-DEP-01, IMPORT-CONS-01, IMPORT-VENDRM-01 | T-02-04 | Full battery green on the atomic commit (34 ok / 0 FAIL / 0 race) | integration | `go test -race -count=1 -timeout 20m ./...` (log) | ✅ | ✅ green |
| 02-02-02 | 02 | 2 | IMPORT-CONS-01 | T-02-05 | Parity gates green with zero expectation edits | e2e/unit | `TestAcceptanceGates` (Gates A–D ×2), `TestLintGoldenFullRegistry` (94), `TestPositionRuneColumns`, `TestProtoastPositionsMatchLegacyContract`, core/CLI matrix by exact RUN lines; freeze diffstat empty | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements — no new tests are needed (import swap + tree deletion; all behavior gates already exist).

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Commit message quality (atomic swap, mapping list) | IMPORT-CONS-01 | Commit prose is not machine-asserted | Read `git log -1` body: names the module mapping and the deleted tree |

*All other phase behaviors have automated verification.*

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 420s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-11-21

## Validation Audit 2026-11-21

| Metric | Count |
|---|---|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
