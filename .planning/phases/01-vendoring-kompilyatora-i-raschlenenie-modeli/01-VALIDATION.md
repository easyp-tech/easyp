---
phase: "1"
slug: "vendoring-kompilyatora-i-raschlenenie-modeli"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-11-17"
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Derived from `01-RESEARCH.md` § Validation Architecture.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify v1.11.1; runner gotestsum v1.13.0 |
| **Config file** | `Taskfile.yml` (targets `test`, `lint`, `lint:go`); `.golangci.yml` (golangci-lint v2) |
| **Quick run command** | `go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/protoast/...` |
| **Full suite command** | `go test -race -count=1 ./...` |
| **Estimated runtime** | ~60–180 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/protoast/...` + `go build ./...`
- **After every plan wave:** Run `go test -race -count=1 ./...` + `golangci-lint run ./...` (with `internal/easypc` path exclusion)
- **Before `/gsd-verify-work`:** Full suite green, lint green, tidy zero-diff, both position tests green
- **Max feedback latency:** 180 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| (filled by planner — one row per PLAN task) | | | | | | | | | |

Placeholder rows the planner must cover (from RESEARCH.md):

| Behavior | Requirement | Test Type | Automated Command | File Exists |
|----------|-------------|-----------|-------------------|-------------|
| Vendored trees present & compiling; own tests pass | LINT-VEND-01 | build + unit | `go build ./... && go test -count=1 ./internal/easypc/... ./internal/protoast/...` | ❌ until vendor commit |
| `go mod tidy` zero-diff | LINT-VEND-02 | gate | `go mod tidy && git diff --exit-code go.mod go.sum` | ❌ Wave 0 |
| breaking/generate/mod/migrate unchanged (SC-3) | LINT-COMPAT-02 (gate) | regression | `go test -race -count=1 ./...` | ✅ existing suites |
| span→Position conversion, UTF-8+tabs (SC-4a) | LINT-VEND-01 | unit | `go test -count=1 ./internal/protoast/ -run TestPosition` | ❌ Wave 0 — new `position_test.go` |
| golden legacy-vs-protoast anchor parity (SC-4b) | LINT-VEND-01 | integration | `go test -count=1 ./internal/core/ -run TestProtoastPositions` | ❌ Wave 0 — new `position_parity_test.go` + fixture |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/easypc/**` + `internal/protoast/**` — vendor drop (REQ-1.1 carrier), from compiler-lab main @ `08cef10`, directory `easyp-cli/`
- [ ] `internal/protoast/position_test.go` — D-07(a), UTF-8+tab fixture
- [ ] `internal/core/position_parity_test.go` + `testdata/position_contract_utf8.proto` — D-07(b)
- [ ] `.golangci.yml` exclusion for `internal/easypc` + `internal/easypc/VENDORED.md` + AGENTS.md D-04 note
- [ ] Environment provisioning if Taskfile gates used: `task` binary + `task init`

*Existing test infrastructure (Go testing, testify, Taskfile) covers the framework requirement — no new framework install needed.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| VENDORED.md provenance correctness (SHA `08cef10909c0722f7042ba17b5c11951f9c6b6f6` matches actual cloned revision) | LINT-VEND-01 | SHA provenance is a human-trust record | `gh repo view easyp-tech/compiler-lab --json ...` or clone and `git rev-parse` to confirm |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 180s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
