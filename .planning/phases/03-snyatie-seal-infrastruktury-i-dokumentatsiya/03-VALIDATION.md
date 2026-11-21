---
phase: "03"
slug: "snyatie-seal-infrastruktury-i-dokumentatsiya"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-11-21"
---

# Phase 03 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` + testify v1.11.1; golangci-lint v2.14.0 |
| **Config file** | `.golangci.yml` |
| **Quick run command** | `go build ./... && bin/golangci-lint run ./...` |
| **Full suite command** | `go test -race -count=1 ./internal/sourceview ./internal/core ./internal/protoast ./internal/rules` (full battery is Phase 4) |
| **Estimated runtime** | ~60 seconds |

---

## Sampling Rate

- **After every task commit:** `go build ./...` + `bin/golangci-lint run ./...` + targeted `-race` packages
- **After every plan wave:** same scoped suite
- **Before `/gsd:verify-work`:** `0 issues.` lint + scoped `-race` green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | IMPORT-SEAL-01 | T-03-01..05 | Seal text gone from live config/agent docs; module rule present; citations correct; no behavior change | integration | rc-aware seal grep empty; `paths:` absent; `ST1005` intact; comment-only diff gate; 7-file changeset; build + targeted `-race` | ✅ | ✅ green |
| 03-01-02 | 01 | 1 | IMPORT-SEAL-01 | T-03-04, T-03-08 | History preserved; evidence attributable | integration | `.spec/plans`/`.spec/features` zero-diff; commit subject/paths asserted | ✅ | ✅ green |
| 03-02-01 | 02 | 2 | IMPORT-SEAL-01 | T-03-06, T-03-07 | Errcheck fixed at source; no suppression; test semantics preserved | unit | one-file commit; `golangci-lint` exactly `0 issues.`; `-race` sourceview | ✅ | ✅ green |
| 03-02-02 | 02 | 2 | IMPORT-SEAL-01 | T-03-08 | Final battery attributable; history and changeset exact | integration | build; 4 scoped `-race` packages; tidy `-diff` empty; seal grep empty; two-commit range equality | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements — no new tests needed (docs/config + one test-only fix).

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Wording quality of the AGENTS.md module rule and `.spec` seam description | IMPORT-SEAL-01 | Prose accuracy is not machine-asserted beyond keyword presence | Read the AGENTS.md `## Do not` bullet and the ARCHITECTURE/PACKAGES additions; confirm they describe the pinned module, upstream fixes, and the D-07 seam |

*All other phase behaviors have automated verification.*

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 120s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-11-21

## Validation Audit 2026-11-21

| Metric | Count |
|---|---|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
