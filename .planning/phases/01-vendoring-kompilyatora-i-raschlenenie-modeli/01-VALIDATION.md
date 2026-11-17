---
phase: "1"
slug: "vendoring-kompilyatora-i-raschlenenie-modeli"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: true
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
| 01-01-T1 (tracer: vendor drop) | 01-01 | 1 | LINT-VEND-01, LINT-VEND-02 | T-01-01, T-01-SC | pinned-SHA assert before copy; tidy zero-diff | build + unit + gate | `go build ./... && go test -count=1 ./internal/easypc/... ./internal/protoast/... && go mod tidy && git diff --exit-code go.mod go.sum` | ❌ until 01-01-T1 (trees are the deliverable) | ⬜ |
| 01-01-T2 (hygiene + atomic commit) | 01-01 | 1 | LINT-VEND-01 | T-01-02, T-01-SC | seal + scoped lint exclusion | gate | `out=$(bin/golangci-lint run ./... 2>&1); echo "$out" \| grep -qE '\.go:[0-9]+:[0-9]+' && test "$(echo "$out" \| grep -E '\.go:[0-9]+:[0-9]+' \| grep -vc 'internal/protoast/build.go')" -eq 0` | ❌ until 01-01-T2 (VENDORED.md new) | ⬜ |
| 01-02-T1 (RED: D-07 tests) | 01-02 | 2 | LINT-VEND-01 | T-01-03 | — | unit (failing-by-design) | `go test -count=1 ./internal/protoast/ -run 'TestPosition' …; a=$?; go test -count=1 ./internal/core/ -run 'TestProtoastPositions' …; b=$?; test $a -ne 0 && test $b -ne 0` (full command in PLAN) | ❌ Wave 2 — new position_test.go + position_parity_test.go + fixture | ⬜ |
| 01-02-T2 (GREEN: rune columns) | 01-02 | 2 | LINT-VEND-01 | T-01-03 | contract enforced at the seam | unit | `go test -count=1 ./internal/protoast/... && go test -count=1 ./internal/core/ -run 'TestProtoastPositions'` | ✅ after 01-01 (vendored files edited) | ⬜ |
| 01-02-T3 (gates + divergence record) | 01-02 | 2 | LINT-VEND-01, LINT-VEND-02 | T-01-SC | tidy zero-diff re-run | full suite + lint + gate | `go test -race -count=1 ./... && bin/golangci-lint run ./... && go mod tidy && git diff --exit-code go.mod go.sum` | ✅ existing suites | ⬜ |
| 01-03-T1 (Position swap) | 01-03 | 2 | LINT-VEND-01 | T-01-04 | values-unchanged diff gate | regression | `go build ./... && go test -race -count=1 ./internal/core/... ./internal/rules/...` + diff-of-diffs values gate (in PLAN) | ✅ existing suites (42 test files edited) | ⬜ |
| 01-03-T2 (read-model split) | 01-03 | 2 | LINT-VEND-01 | T-01-04, T-01-SC | tidy zero-diff re-run | regression + new twin test | `go build ./... && go test -race -count=1 ./... && bin/golangci-lint run ./... && go mod tidy && git diff --exit-code go.mod go.sum` | ✅ existing + ❌ twin test new in Wave 2 | ⬜ |
| 01-04-T1 (casing verbatim) | 01-04 | 3 | LINT-VEND-01 | T-01-05 | pinned-SHA assert + byte-identical diff | unit + diff gate | `diff <clone>/easyp-cli/internal/rules/casing.go internal/rules/casing.go && diff <clone>/easyp-cli/internal/rules/casing_test.go internal/rules/casing_test.go && go test -count=1 ./internal/rules/` | ❌ until 01-04-T1 | ⬜ |
| 01-04-T2 (lint_helpers adapted) | 01-04 | 3 | LINT-VEND-01 | T-01-05, T-01-SC | bounded delta + tidy zero-diff | build + unit + lint | `go build ./... && go test -race -count=1 ./internal/rules/... && bin/golangci-lint run ./... && go mod tidy && git diff --exit-code go.mod go.sum` | ❌ until 01-04-T2 | ⬜ |

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
