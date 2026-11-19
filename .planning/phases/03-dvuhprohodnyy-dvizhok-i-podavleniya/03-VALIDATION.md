---
phase: "03"
slug: "dvuhprohodnyy-dvizhok-i-podavleniya"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-11-19"
---

# Phase 03 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (testify assertions) |
| **Config file** | none — existing colocated `_test.go` layout |
| **Quick run command** | `go test -count=1 ./internal/core/... ./internal/rules/...` |
| **Full suite command** | `go test -race -count=1 ./... && bin/golangci-lint run ./...` |
| **Estimated runtime** | ~240 seconds (api package dominates) |

*Note: `task`/gotestsum отсутствуют на PATH — прямые go-команды (research A7, Phase 2 precedent).*

---

## Sampling Rate

- **After every task commit:** Run `go test -count=1 ./internal/core/... ./internal/rules/...`
- **After every plan wave:** Run `go test -race -count=1 ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 240 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| (заполняется планами) | 0X | X | LINT-ENG-01 | T-03-* | Модель из слинкованных деревьев, порядок обхода сохранён | unit/table | `go test -race -count=1 ./internal/core/ -run 'TestLintEngine'` | ❌ W0 | ⬜ pending |
| (заполняется планами) | 02 | 2 | LINT-ENG-05 | T-03-03 | Подавления с паритетом main (main-vs-new таблица): 9-cell EOF matrix, byte-exact error texts, policy independence, 3 dialects, ignore/ignore_only (walk, per-file, Finish re-filter), tab position pin | table/golden | `go test -race -count=1 ./internal/core/ -run 'TestLintLinkedSuppressionParity'` (21 named rows: TestLintLinkedSuppressionParity/{eof_with_newline, eof_without_newline, block_comment_at_eof, orphan_enable, unknown_trailing_rule, unattached_disable, unpaired_attached_disable, marker_in_string, nested_closing_brace, malformed_directive_exact_text, unknown_name_exact_text, policy_allow_suppresses, policy_disallow_keeps, nolint_dialect_suppresses, buf_dialect_suppresses, nolint_unknown_name_error, config_ignore_walk_filter, ignore_only_dir_prefix, ignore_only_finish_refilter, finish_comment_refilter, tab_directive_position_pin}) | ✅ | ✅ green |
| (заполняется планами) | 0X | X | LINT-ENG-06 | T-03-* | Валидация lint.use с паритетом main | unit | `go test -count=1 ./internal/core/ -run 'TestLintUseValidation'` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Фикстуры подавлений (комментарии-директивы, ignore/ignore_only конфиги) — testdata-дерево, размещение — discretion planner'а
- [ ] Тестовые дубли правил (representative per-file + ModuleRule) для main-vs-new parity — из существующих правил internal/rules
- [ ] `Existing infrastructure covers test framework — go test готов`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| *(none — движок dormant, всё проверяется автоматикой; переключение — Phase 5)* | — | — | — |

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
