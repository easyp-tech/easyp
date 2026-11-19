---
phase: "04"
slug: "perenos-43-pravil-s-batch-paritetom"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-11-19"
---

# Phase 04 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (testify assertions) |
| **Config file** | none — existing colocated `_test.go` layout |
| **Quick run command** | `go test -count=1 ./internal/rules/... ./internal/core/...` |
| **Full suite command** | `go test -race -count=1 ./... && bin/golangci-lint run ./...` |
| **Estimated runtime** | ~240 seconds (api package dominates; full suite minus plugin/generation carve-out) |

---

## Sampling Rate

- **After every task commit:** Run `go test -count=1 ./internal/rules/... ./internal/core/...`
- **After every plan wave:** Run `go test -race -count=1 ./internal/rules/... ./internal/core/...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 240 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| (заполняется планами; каждая batch-ячейка — parity-тест против live-старого движка) | 0X | X | LINT-PARITY-02 | T-04-* | Идентичные замечания старое-vs-новое (rule, path, line, col) на общих фиксатурах | parity/table | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity'` | ❌ W0 | ⬜ pending |
| (Wave 0: disk-instrumentation cell) | 04-01 | 0 | LINT-PARITY-02 | T-04-* | readSource ровно один раз на файл корпуса | unit | `go test -race -count=1 ./internal/core/ -run 'TestReadSourceOnce'` | ❌ W0 | ⬜ pending |
| (Wave 0: position probe) | 04-01 | 0 | LINT-PARITY-02 | T-04-* | Decl.Pos таблица по 6 kinds закреплена | unit/probe | `go test -race -count=1 ./internal/core/ -run 'TestDeclPositionProbe'` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Parity harness (live old engine + dormant new engine in-process, sorted multiset compare) — Wave 0 план
- [ ] Позиционный probe (Decl.Pos таблица для Import/Field/Message/Package/RPC/Option) — Wave 0 план
- [ ] Disk-instrumentation cell (standing constraint 999.1) — Wave 0 план
- [ ] Пилотное правило (PackageDirectoryMatch) как smoke батча — Wave 0 план
- [ ] `Existing infrastructure covers test framework — go test готов`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| *(none — parity автоматизирован per-batch; корпусный lint-diff — ворота Phase 7)* | LINT-PARITY-01 | — | — |

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 240s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
