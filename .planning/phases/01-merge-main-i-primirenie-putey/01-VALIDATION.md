---
phase: "01"
slug: "merge-main-i-primirenie-putey"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-11-21"
---

# Phase 01 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` + testify v1.11.1 |
| **Config file** | none — package-level `go test`; race flags from Taskfile convention |
| **Quick run command** | `go build ./... && go test -count=1 ./internal/core ./internal/rules ./internal/protoast` |
| **Full suite command** | `go test -race -count=1 ./...` (≈7 min; `internal/api` ≈309 s cold) |
| **Estimated runtime** | ~420 seconds |

---

## Sampling Rate

- **After every task commit:** Run quick run command
- **After every plan wave:** Run full suite command
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 420 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | IMPORT-REC-01 | T-01-01 / T-01-04 | Merge commit resolves conflicts deterministically (blob gates), no silent `--ours` drift | integration (shell) | `git merge-base --is-ancestor origin/main HEAD && test "$(git rev-list --count HEAD..origin/main)" -eq 0`; blob gates in plan verify | ✅ | ⬜ pending |
| 01-01-02 | 01 | 1 | IMPORT-REC-01 | T-01-02 | Breaking path uses `easyprotoc/protoc`, pins present, no `replace` | unit (module) | `go list -m github.com/easyp-tech/easyprotoc google.golang.org/protobuf`; `grep -q easyprotoc/protoc internal/core/breaking_profiles.go`; `go mod tidy -diff` | ✅ | ⬜ pending |
| 01-01-03 | 01 | 1 | IMPORT-REC-01 | T-01-03 | Lint contour byte-identical; parity not shifted | e2e/unit | `git diff --stat $PRE_MERGE_BASE HEAD -- internal/rules internal/easypc internal/protoast internal/core/lint_* internal/core/link.go internal/core/proto_info_read.go internal/core/dom.go` (empty); `go test -race -count=1 ./internal/api -run '^TestAcceptanceGates$'`; `go test -race -count=1 ./internal/rules -run '^TestLintGoldenFullRegistry$'` | ✅ | ⬜ pending |
| 01-01-04 | 01 | 2 | IMPORT-REC-01 | T-01-05 | Full battery green on committed merge | integration | `go build ./... && go test -race -count=1 ./...`; D-07 position tests | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements — no new Go tests are needed. The shell-level blob/ancestry gates live in the plan task verify blocks.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Merge commit message quality | IMPORT-REC-01 | Commit prose is not machine-asserted by the repo's gates | Read `git log -1` body: lists 3 resolved files, names #253, states main-wins/vendor-dormant rule |

*All other phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 420s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
