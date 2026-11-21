---
phase: "04"
slug: "priemka-polnyy-paritet"
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-11-21"
---

# Phase 04 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` + testify v1.11.1; golangci-lint v2.14.0 |
| **Config file** | `.golangci.yml` |
| **Quick run command** | `go build ./... && bin/golangci-lint run ./...` |
| **Full suite command** | `go test -race -count=1 -timeout 20m -v ./...` (one invocation) |
| **Estimated runtime** | ~300 seconds |

---

## Sampling Rate

- **After every task:** n/a — evidence-only phase; one battery, then named gates
- **Phase gate:** full battery + 11 named gates (RUN+PASS) + lint/schema/tidy/seal/pin predicates
- **Max feedback latency:** 420 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 01 | 1 | IMPORT-PARITY-01 | T-04-24, T-04-28 | Battery on the committed tree; one run; observable skip | integration | `go test -race -count=1 -timeout 20m -v ./...` (34 ok / 0 FAIL / 1 known skip / 3 no-test) | ✅ | ✅ green |
| 04-01-02 | 01 | 1 | IMPORT-PARITY-01 | T-04-22, T-04-23 | Non-vacuous gates; unmoved oracles | e2e/unit | 11 gates by exact `=== RUN` + `--- PASS`; golden 94; Gates A–D ×2 (1e9a596); D-07 both | ✅ | ✅ green |
| 04-01-03 | 01 | 1 | IMPORT-PARITY-01 | T-04-25..27, T-04-30 | No suppression/schema/pin/tidy drift; honest delta | integration | `bin/golangci-lint run ./...` = `0 issues.`; schema-gen zero diff; tidy empty; pins exact; seal empty; delta recorded | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements — no new tests.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Code review applicability | — (process) | Evidence-only phase: zero source delta (`git diff FINAL_CODE_SHA HEAD -- ':(exclude).planning'` empty), so there is no code changeset to review | Confirm the zero-diff predicate in the SUMMARY; code review recorded as not-applicable for this phase |

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
