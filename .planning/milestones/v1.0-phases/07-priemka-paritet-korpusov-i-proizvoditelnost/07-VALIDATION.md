---
phase: "07"
slug: "priemka-paritet-korpusov-i-proizvoditelnost"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: true) (#2117)
status: validated
nyquist_compliant: false
wave_0_complete: true
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
| 07-02 T1 (tracer) | 02 | 1 | LINT-PARITY-01 | T-07-02 | Gates A–D: main-vs-main determinism, zero-COMPILE precondition, 4-tuple multiset identity (×2 runs), 3-rule delta accounting | gates-runner | `go test -race -count=1 ./internal/api/ -run TestAcceptanceGates` | ✅ | ⬜ pending |
| 07-01 T1 (pin) + T2 (fix) | 01 | 1 | (SC3) | T-07-01 | wazero race fixed — plugin package green under -race | unit/race | `go test -race -count=1 ./internal/adapters/plugin/` | ✅ | ✅ green |
| 07-03 T1 (tracer battery) | 03 | 2 | (SC3) | T-07-06 | Milestone gates on the final state: full -race suite (3718 tests, incl. plugin/generation, TestAcceptanceGates, golden corpus, standing-constraint cells), golangci zero issues, schema drift empty via one tmp dir; hadolint = honest manual item (Docker absent) | battery | `bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./... && bin/golangci-lint run ./... && tmpd=$(mktemp -d) && go run ./cmd/easyp schema-gen --out-dir "$tmpd" && diff -r schemas "$tmpd"` | ✅ | ✅ green |
| 07-03 T2+T3 (evidence + disposal) | 03 | 2 | LINT-PARITY-01, LINT-PERF-01 | T-07-04 | Corpora/perf legs: parity-evidence-07.json (baseline 1e9a596 / 41, head / 44, Gates A-D teeth, corpora_leg pending_user_unblock with fresh clone_attempts 2026-11-21T04:15:09Z) + perf protocol committed + end-of-phase human check (unblock path 07-USER-SETUP.md) | evidence + checkpoint | `python3 -c "import json;d=json.load(open('.planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/parity-evidence-07.json'));assert d['corpora_leg']['clone_attempts'] and d['corpora_leg']['status']=='pending_user_unblock'"` (automated disposition gate) | ✅ | ⬜ pending (MANUAL-ONLY until corpora unblock) |

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
| Corpora parity (api-common 92, googleapis-nocompute 7414) | LINT-PARITY-01 | Private corpora unreachable on this machine (research spike 2026-11-21 + fresh attempts at 07-03 execution — gh exit 1, SSH exit 128, env vars unset; records in parity-evidence-07.json corpora_leg.clone_attempts) | Unblock per 07-USER-SETUP.md (GOOGLEAPIS must be a fresh public googleapis clone — Pitfall 7 revision-compatibility), then the executor fires /tmp/p7/run-parity.sh parity unchanged; compare vs main binary via Gates A-D |
| Perf measurement on googleapis-nocompute (gate: median(new) <= median(main)) | LINT-PERF-01 | Same reachability constraint; protocol committed in perf-evidence-07.md before any measurement | Same unblock; perf protocol (warmup 1 + median-of-5 interleaved, no-race) fires unchanged; api-common recorded, not gated |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 300s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-11-21 (nyquist audit: rows real; corpora legs honestly deferred pending_user_unblock per ratified UAT)
