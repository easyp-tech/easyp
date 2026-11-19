---
phase: "02"
slug: "linker-i-marshrutizatsiya-oshibok"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-11-19"
---

# Phase 02 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (testify assertions) |
| **Config file** | none — existing colocated `_test.go` layout |
| **Quick run command** | `go test -count=1 ./internal/core/...` |
| **Full suite command** | `go test -race -count=1 ./... && bin/golangci-lint run ./...` |
| **Estimated runtime** | ~120 seconds (full race suite) |

*Note: `task`/gotestsum отсутствуют на PATH этой машины (research A7) — фоллбек на прямые go-команды.*

---

## Sampling Rate

- **After every task commit:** Run `go test -count=1 ./internal/core/...`
- **After every plan wave:** Run `go test -race -count=1 ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| (заполняется планами) | 0X | X | LINT-ENG-02 | T-02-* | COMPILE-issues только при провале линковки с compile-ошибками | unit/golden | `go test -count=1 ./internal/core/ -run TestLink` | ❌ W0 | ⬜ pending |
| (заполняется планами) | 0X | X | LINT-ENG-03 | T-02-* | Откат пофайлового разбора без COMPILE | unit | `go test -count=1 ./internal/core/ -run TestMissingImport` | ❌ W0 | ⬜ pending |
| (заполняется планами) | 0X | X | LINT-ENG-04 | T-02-* | Ошибка чтения/синтаксиса как в main | unit/action | `go test -count=1 ./internal/api/...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Фиксатурная матрица трёх классов ошибок (компиляция / отсутствующий импорт / чтение-синтаксис) — testdata-дерево в `internal/core/testdata/` (размещение — discretion planner'а)
- [ ] `Existing infrastructure covers test framework — go test готов, новой установки не требуется`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| *(none — все поведения фазы имеют автоматическую верификацию; полный Action-прогон класса 1 пере-запускается в Phase 5 per research A1)* | LINT-ENG-02 | — | — |

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
