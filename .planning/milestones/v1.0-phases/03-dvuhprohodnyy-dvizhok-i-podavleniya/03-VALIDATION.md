---
phase: "03"
slug: "dvuhprohodnyy-dvizhok-i-podavleniya"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: validated
nyquist_compliant: true
wave_0_complete: true
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
| (заполняется планами) | 01/03 | 1/2 | LINT-ENG-01 | T-03-01, T-03-02, T-03-04 | Модель из слинкованных деревьев, порядок обхода сохранён; параллельная корректность движка: детерминизм повторных прогонов, аудит statelessness с shared-дублями и точными count, serial Collect в порядке обхода, ctx-cancel, walk-order приоритет ошибок, COMPILE-матрица через движок (изоляция, dep-пути, bypass подавлений, детерминизм) | unit/table | `go test -race -count=1 ./internal/core/ -run 'TestLintLinked'` (12 named cells, план 03-01) и `go test -race -count=1 ./internal/core/ -run 'TestLintLinkedMatrix'` (9 named cells: TestLintLinkedMatrix/{repeated_runs_byte_identical, race_shared_doubles_identity, collect_serial_walk_order, ctx_cancel_surfaces, walk_order_error_precedence, mixed_corpus_compile_isolation, dep_file_diagnostic_through_engine, compile_bypasses_suppression, compile_deterministic_twice}, план 03-03) | ✅ | ✅ green |
| (заполняется планами) | 02 | 2 | LINT-ENG-05 | T-03-03 | Подавления с паритетом main (main-vs-new таблица): 9-cell EOF matrix, byte-exact error texts, policy independence, 3 dialects, ignore/ignore_only (walk, per-file, Finish re-filter), tab position pin | table/golden | `go test -race -count=1 ./internal/core/ -run 'TestLintLinkedSuppressionParity'` (21 named rows: TestLintLinkedSuppressionParity/{eof_with_newline, eof_without_newline, block_comment_at_eof, orphan_enable, unknown_trailing_rule, unattached_disable, unpaired_attached_disable, marker_in_string, nested_closing_brace, malformed_directive_exact_text, unknown_name_exact_text, policy_allow_suppresses, policy_disallow_keeps, nolint_dialect_suppresses, buf_dialect_suppresses, nolint_unknown_name_error, config_ignore_walk_filter, ignore_only_dir_prefix, ignore_only_finish_refilter, finish_comment_refilter, tab_directive_position_pin}) | ✅ | ✅ green |
| (заполняется планами) | 04 | 2 | LINT-ENG-06 | T-03-05, T-03-SC | Валидация lint.use с паритетом main: неизвестные имена правила/группы в use, except и ignore_only дают `errors.Is` `core.ErrInvalidRule` с точными префиксами `validateConfigRuleNames: use:/except:/ignore_only:`; special case `PACKAGE_NO_IMPORT_CYCLE` ("is not implemented") закреплён до Phase 6; валидный конфиг конструирует старый путь и linked-движок без второй валидации имён; реестр покрывает все group-ключи и имена правил | unit/table | `go test -race -count=1 ./internal/rules/ -run 'TestLintUseValidationParity'` (7 named cells: TestLintUseValidationParity/{unknown_rule_in_use, unknown_group_in_use, unknown_name_in_except, unknown_name_in_ignore_only, package_no_import_cycle_special_case, valid_config_still_constructs, registry_covers_all_names}, план 03-04) | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] Фикстуры подавлений (комментарии-директивы, ignore/ignore_only конфиги) — corpus-фикстуры lint_linked_test.go / lint_suppressions_parity_test.go (планы 03-01/03-02)
- [x] Тестовые дубли правил (representative per-file + ModuleRule) для main-vs-new parity — движковые дубли 03-01 + parity-дубли 03-02
- [x] `Existing infrastructure covers test framework — go test готов`

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
