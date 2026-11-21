---
phase: "07"
slug: "priemka-paritet-korpusov-i-proizvoditelnost"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-11-21"
---

# Phase 07 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (testify) + subprocess binary goldens + gates-runner |
| **Config file** | none |
| **Quick run command** | `go test -count=1 ./internal/core/... ./internal/rules/... ./internal/adapters/plugin/...` |
| **Full suite command** | `go test -race -count=1 ./... && bin/golangci-lint run ./... && go run ./cmd/easyp schema-gen --out-dir $(mktemp) && diff -r schemas $(mktemp)` |
| **Estimated runtime** | ~300 seconds (plugin package + gates) |

---

## Sampling Rate

- **After every task commit:** Run `go test -count=1 ./internal/core/... ./internal/rules/...`
- **After every plan wave:** Run `go test -race -count=1 ./...`
- **Before `/gsd:verify-work`:** Full suite must be green (incl. the wazero race fix)
- **Max feedback latency:** 300 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| (заполняется планами) | 0X | X | LINT-PARITY-01 | T-07-* | Gates A–D: main-vs-main determinism, zero-COMPILE precondition, 4-tuple multiset identity (×2 runs), 3-rule delta accounting | gates-runner | `go test -race -count=1 ./internal/core/ -run 'TestAcceptanceGates'` | ❌ W0 | ⬜ pending |
| 07-01 T1 (pin) + T2 (fix) | 01 | 1 | (SC3) | T-07-01 | wazero race fixed — plugin package green under -race | unit/race | `go test -race -count=1 ./internal/adapters/plugin/` | ✅ | ✅ green |
| (заполняется планами) | 0X | X | (SC3) | T-07-* | Milestone gates: full -race suite + golangci + schema drift | battery | `go test -race -count=1 ./...` | ✅ | ⬜ pending |
| (заполняется планами) | 0X | X | LINT-PARITY-01, LINT-PERF-01 | T-07-* | Corpora/perf legs — READY-TO-FIRE scripts + human-check (unblock path 06-USER-SETUP.md) | checkpoint | `checkpoint:human-verify` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Baseline binary build from `1e9a596` (pre-milestone main)
- [ ] Gates-runner harness (Gates A–D over the golden corpus)
- [ ] `Existing infrastructure covers test framework`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Corpora parity (api-common 92, googleapis-nocompute 7414) | LINT-PARITY-01 | Private corpora unreachable on this machine (spike-verified) | Unblock per 06-USER-SETUP.md, fire the ready scripts, compare vs main binary |
| Perf measurement on googleapis-nocompute | LINT-PERF-01 | Same reachability constraint | Same scripts; timing recorded per protocol |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 300s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
