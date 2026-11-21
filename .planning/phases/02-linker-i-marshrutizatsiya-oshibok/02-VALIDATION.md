---
phase: "02"
slug: "linker-i-marshrutizatsiya-oshibok"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-11-19"
---

# Phase 02 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Derived from `02-RESEARCH.md` § Validation Architecture.

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
| 02-01-T1 (tracer: classifier + typed seam) | 02-01 | 1 | LINT-ENG-02, LINT-ENG-03 | T-02-01, T-02-02, T-02-03, T-02-SC | COMPILE etalon shape (test.proto 4:3, verbatim message, zero Filename/Offset/SourceName); linked trees incl. WKT; missing-import fallback via typed resolver seam (no diagnostic-text parsing) | unit | `go build ./... && go test -race -count=1 ./internal/core/ -run 'TestLinkSources'` — fails when: non-zero exit or cells compile_error / compiling_with_wkt / missing_import absent from the run | ✅ | ✅ green (executed 2026-11-19) |
| 02-01-T2 (contract hardening) | 02-01 | 1 | LINT-ENG-02 | T-02-02, T-02-SC | dep diagnostics carry the dep's import path (b.proto); twice-run deterministic order; ≥2-file corpus under -race | unit | same command — fails when: fewer than 6 named cells pass (dep_file_semantic_error / deterministic_order_twice / multi_file_corpus absent) | ✅ | ✅ green (executed 2026-11-19) |
| 02-02-T1 (syntax precedence, LINT-ENG-04) | 02-02 | 2 | LINT-ENG-04 | T-02-02, T-02-05, T-02-SC | syntax error beats COMPILE: broken corpus → fallback (nil, nil); syntax+semantic mix → fallback; semantic-only control still emits COMPILE (precedence does not over-trigger) | unit | same command — fails when: syntax_error_falls_back / syntax_beats_compile / semantic_only_emits_compile absent from the run | ✅ | ✅ green (executed 2026-11-19) |
| 02-02-T2 (missing-import matrix, LINT-ENG-03) | 02-02 | 2 | LINT-ENG-03 | T-02-01, T-02-04, T-02-06, T-02-SC | transitive (depth 2) and importFileAllowed-blocked imports → fallback + per-file `*OpenImportFileError` (exit-2 carrier); WKT resolves as compiling; non-local path ("../") classifies COMPILE per etalon | unit | same command — fails when: transitive_missing_depth2 / blocked_import / wkt_resolves / non_local_import_compile absent from the run | ✅ | ✅ green (executed 2026-11-19) |
| 02-02-T3 (position/shape pinning) | 02-02 | 2 | LINT-ENG-02 | T-02-05, T-02-06, T-02-SC | LineCol convention (tab-stop 8, byte columns) pinned by a fixture provably diverging from the D-07 rune column; cycle diagnostic at 2:1 with chain message; D-03 sort across files on the testdata corpus | unit + fixture | same command — fails when: tab_utf8_position / import_cycle / multi_file_matrix_order absent; `git ls-files -- internal/core/testdata/link/` lists tab_utf8/broken.proto + multi_file/{a,b}.proto | ✅ | ✅ green (executed 2026-11-19) |
| 02-03-T1 (in-process action matrix) | 02-03 | 2 | LINT-ENG-03, LINT-ENG-04 | T-02-07, T-02-SC | missing import → `*core.OpenImportFileError` (FileName == "missing.proto", exit-2 class); syntax → plain command error, NOT OpenImportFileError (exit-1 class); lint violation → ErrHasLintIssue; clean → nil (exit-0 class) | integration (actionV1) | `go build ./... && go test -race -count=1 ./internal/api/ -run 'TestLintLinkMatrix$'` — fails when: non-zero exit or cells missing_import_exit2_class / syntax_error_command_class / lint_issue_exit1_class / compiling_exit0_class absent | ✅ | ✅ green (executed 2026-11-19) |
| 02-03-T2 (subprocess exit codes) | 02-03 | 2 | LINT-ENG-03, LINT-ENG-04 | T-02-07, T-02-08, T-02-09, T-02-SC | built binary exits 2 with "Cannot import file" + missing file name on missing import; exits 1 on syntax error; exits 0 printing nothing on clean corpus (numeric codes from `*exec.ExitError`, children via exec.Cmd.Dir) | smoke (subprocess) | `go build ./... && go test -race -count=1 ./internal/api/ -run 'TestLintLinkMatrixExitCodes'` — fails when: non-zero exit or any of missing_import_code_2 / syntax_error_code_1 / compiling_clean_code_0 fails or is absent | ✅ | ✅ green (executed 2026-11-19) |
| 02-03-T3 (COMPILE output golden) | 02-03 | 2 | LINT-ENG-02 | T-02-SC | text `test.proto:4:3: Required fields are not allowed in proto3. (COMPILE)` byte-exact incl. the empty-SourceName artifact; JSON with zero Filename/Offset and rule COMPILE; multi-issue render order stable | golden | `go build ./... && go test -race -count=1 ./internal/api/ -run 'TestPrintIssuesCompileGolden'` — fails when: non-zero exit or text_golden / json_golden / multi_issue_order fails or is absent | ✅ | ✅ green (executed 2026-11-19) |
| 02-04-T1 (synthetic warn drop, G-02-3) | 02-04 | 3 | LINT-ENG-02 | T-02-10, T-02-SC | IsWarn drop branch exercised through the exact `diagCollector` handler protoc.Link receives: warn never enters the diagnostics slice; non-warn collected exactly; mixed warn/error sequence preserved in arrival order; RED evidence `tdd-red-evidence-02-04-task1.json` | unit (seam) | `go build ./... && go test -race -count=1 ./internal/core/ -run 'TestLinkSourcesWarnDrop'` — fails when: non-zero exit or warn_dropped / non_warn_collected / warn_mix_order absent | ✅ | ✅ green (executed 2026-11-19) |
| 02-04-T2 (contention + determinism) | 02-04 | 3 | LINT-ENG-02 | T-02-10, T-02-SC | exact-count collection (require.Len) under 4-goroutine contention under -race, no warn leakage, no lost updates; full 20-cell matrix green across two consecutive runs | unit + determinism gate | `go build ./... && go test -race -count=2 ./internal/core/ -run 'TestLinkSources'` — fails when: non-zero exit on either consecutive run, handler_race_mixed absent, or fewer than 20 named cells | ✅ | ✅ green (executed 2026-11-19) |

Cross-cutting rows (phase gates beyond single tasks):

| Behavior | Requirement | Threat Ref | Test Type | Automated Command | Status |
|----------|-------------|------------|-----------|-------------------|--------|
| Determinism: identical matrix results twice consecutively | LINT-ENG-02/03/04 (cross) | T-02-10 | integration gate | `go test -race -count=2 ./internal/core/ -run 'TestLinkSources'` | ✅ green (executed 2026-11-19) |
| Supply-chain zero-install gate | all (T-02-SC) | T-02-SC | gate | `go mod tidy && git diff --exit-code go.mod go.sum` | ✅ green (executed 2026-11-19) |
| Dormancy D-04 + Phase 1 easypc seal | all | T-02-01..T-02-10 | gate | `git diff --stat internal/core/lint.go internal/core/proto_info_read.go internal/api/lint.go internal/core/dom.go internal/api/lint_v1.go` empty; `git status --porcelain internal/easypc/` empty | ✅ green (executed 2026-11-19) |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

Cell inventory evidenced green by the audit runs (2026-11-19, `-race`):
- `internal/core/link_test.go` — TestLinkSources (16 cells: compile_error, compiling_with_wkt, missing_import, dep_file_semantic_error, deterministic_order_twice, multi_file_corpus, syntax_error_falls_back, syntax_beats_compile, semantic_only_emits_compile, transitive_missing_depth2, blocked_import, wkt_resolves, non_local_import_compile, tab_utf8_position, import_cycle, multi_file_matrix_order) + TestLinkSourcesWarnDrop (4 cells: warn_dropped, non_warn_collected, warn_mix_order, handler_race_mixed)
- `internal/api/lint_link_matrix_test.go` — TestLintLinkMatrix (4 cells), TestLintLinkMatrixExitCodes (3 cells), TestPrintIssuesCompileGolden (3 cells)
- `internal/core/testdata/link/` — tab_utf8/broken.proto, multi_file/{a,b}.proto

---

## Wave 0 Requirements

- [x] Фиксатурная матрица трёх классов ошибок (компиляция / отсутствующий импорт / чтение-синтаксис) — testdata-дерево в `internal/core/testdata/` (размещение — discretion planner'а) — **present: `internal/core/testdata/link/{tab_utf8,multi_file}/`**
- [x] `Existing infrastructure covers test framework — go test готов, новой установки не требуется`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| *(none — все поведения фазы имеют автоматическую верификацию; полный Action-прогон класса 1 пере-запускается в Phase 5 per research A1, ратифицировано 02-UAT Test 1: pass)* | LINT-ENG-02 | — | — |

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 120s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** validated 2026-11-19 (Nyquist validation audit — 10 task rows + 3 cross-gates executed green; 0 gaps found, 0 escalations; no new tests needed — existing suites cover every requirement edge)

## Validation Audit 2026-11-19

| Metric | Count |
|---|---|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

Audit method: every automated command in the map was executed against the working tree (`-race`); TestLinkSources + TestLinkSourcesWarnDrop returned 20/20 named cells PASS, the api matrix returned 10/10 PASS (including the subprocess exit-code cells 2/1/0), the `-count=2` determinism gate and the supply-chain/dormancy gates were re-run green. Requirement coverage cross-check found no uncovered behavioral edge: LINT-ENG-02 (shape/sort/warn-drop/golden/positions), LINT-ENG-03 (direct/transitive/blocked/WKT + exit-2 route in-process and subprocess), LINT-ENG-04 (precedence cells + command-error class + exit 1) are each pinned by executed tests, so no new tests were generated.
