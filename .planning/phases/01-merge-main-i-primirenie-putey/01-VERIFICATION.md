---
phase: 01-merge-main-i-primirenie-putey
verified: 2026-11-21T12:04:56Z
status: passed
score: 18/18 must-haves verified
covered_files:
  - ".planning/phases/01-merge-main-i-primirenie-putey/01-01-PLAN.md"
  - ".planning/phases/01-merge-main-i-primirenie-putey/01-01-SUMMARY.md"
  - ".planning/phases/01-merge-main-i-primirenie-putey/01-02-PLAN.md"
  - ".planning/phases/01-merge-main-i-primirenie-putey/01-02-SUMMARY.md"
  - "go.mod"
  - "go.sum"
  - "internal/api/breaking_v1.go"
  - "internal/core/breaking_profiles.go"
  - "internal/core/core.go"
  - "internal/sourceview/local.go"
  - "internal/sourceview/root_tree.go"
  - "internal/sourceview/session.go"
  - "internal/sourceview/session_test.go"
covered_digest: "v4:sha256:3963a1f8f5b17bb8266699e0a1d809d4fcda957c8ef7ceb38c3b3ec55527536e"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 1: Merge main и примирение путей Verification Report

**Phase Goal:** Ветка содержит актуальный `main` (#253 «Use compiler for breaking check»): breaking-путь и lint-путь оба потребляют `easyprotoc/protoc`, конфликты `internal/core` разрешены семантически, поведение lint-пути не изменилось
**Verified:** 2026-11-21T12:04:56Z
**Status:** passed
**Re-verification:** No — initial verification

**Mode note.** ROADMAP помечает фазу как `Mode: mvp`, но её Goal не является user story (`As a … I want … so that …`). Обе PLAN-файлы явно фиксируют это (MVP framing note; ratified v1.0 precedent) и несут контракт через goal-backward truths. Фаза инфраструктурная (merge/интеграция), пользовательского потока для User Flow Coverage нет; проверка выполнена по четырём Success Criteria ROADMAP и `must_haves` обоих планов. Отдельных human-flow пунктов это не создаёт.

**Goal wording vs D-06.** Формулировка Goal «breaking-путь и lint-путь оба потребляют `easyprotoc/protoc`» прочитана через CONTEXT D-03/D-06: Phase 1 намеренно оставляет lint-путь на байт-замороженном вендор-пути `internal/easypc/protoc` (dormant до Phase 2, IMPORT-CONS-01). Deliverable Phase 1: breaking использует `easyprotoc/protoc`, lint-поведение не изменилось, конфликты разрешены семантически, ветка не отстаёт. Gap за «lint ещё не импортирует easyprotoc/protoc» не ставится — это требование Phase 2 по дизайну.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `origin/main` (`5d36266`, #253) — предок `HEAD`; `HEAD..origin/main` = 0, ветка не отстаёт | ✓ VERIFIED | `git merge-base --is-ancestor origin/main HEAD` rc 0; `git rev-list --count HEAD..origin/main` = 0; `git ls-remote origin refs/heads/main` = `5d36266d71e85f8f0fc3fa41f787a03ee5c68a79` |
| 2 | Мерж-коммит имеет ровно двух родителей в порядке (pre-merge base, затем `5d36266`); first-parent changeset ровно 9 файлов; lint-контур в нём отсутствует | ✓ VERIFIED | `994c503` parents = `79e353a… 5d36266…`; `git diff --name-only 994c503^1 994c503` = ровно 9 ожидаемых путей, lint-путей нет |
| 3 | В дереве нет конфликтных маркеров | ✓ VERIFIED | `git grep -E '^(<<<<<<< |>>>>>>> )'` по всему tracked-дереву — 0 совпадений; `*.go` — 0 |
| 4 | `internal/core/breaking_profiles.go` байт-идентичен blob'у `origin/main` `6e45a055…`, импортирует `easyprotoc/protoc`, legacy-путь отсутствует | ✓ VERIFIED | blob `HEAD:…` = `6e45a055c1e12efb25d14d210fdfa5ee0c52ba90` = `origin/main:…`; импорт `github.com/easyp-tech/easyprotoc/protoc` (строка 14), `protoc.Compile` (122); `protocompile` count = 0 |
| 5 | `go.mod` содержит точные пины `easyprotoc v0.0.0-20261104120806-f378f3f3a05f` и `google.golang.org/protobuf v1.37.0-rc.1`; `go mod tidy -diff` пуст; нет `replace`/`vendor` | ✓ VERIFIED | `go.mod:10`, `go.mod:24`; `go list -m` вернул обе точные версии; tidy output 0 байт; `grep -c '^replace'` = 0, `'^vendor'` = 0 |
| 6 | `internal/core/dom.go` байт-идентичен pre-merge blob'у ветки `52bfbd7…` и объявляет `type Position = protoast.Position` | ✓ VERIFIED | blob `994c503:…` = `994c503^1:…` = `52bfbd7d645a4c2d587ea84d08c9bbf27e7e0f67`; alias на строке 19, импорт `internal/protoast` — 1 |
| 7 | `internal/core/core.go` — union-блоб `34987d1…`; импортирует `internal/protoast` и `internal/sourceview`; несёт и `sourceSession`, и `importCache`/`parsedImports` | ✓ VERIFIED | blob `HEAD:…` = `34987d1824df1c4c14337df4b04c2625761e7b65`; импорты строки 14–15; `sourceSession         *sourceview.Session` строка 32; `importCache *parsedImports` строки 47–54; `Options.SourceSession` строка 84 |
| 8 | Lint-контур имеет нулевой diff против pre-merge base; мерж не меняет ни одного файла контура | ✓ VERIFIED | `git diff --stat 79e353a 994c503 -- internal/rules internal/easypc internal/protoast 'internal/core/lint_*' internal/core/link.go internal/core/proto_info_read.go internal/core/dom.go` — пусто |
| 9 | Breaking/sourceview/API код байт-идентичен `origin/main` на мерж-коммите | ✓ VERIFIED | `git diff origin/main 994c503 -- internal/core/breaking_profiles.go internal/sourceview internal/api/breaking_v1.go` — пусто |
| 10 | `go build ./...` и `go vet ./...` завершаются с кодом 0 | ✓ VERIFIED | Независимый прогон: build rc 0; vet rc 0 |
| 11 | Полная `-race` батарея зелёная: exit 0, 0 FAIL, ≥41 ok, 0 race-warnings, 0 skips | ✓ VERIFIED | `go test -race -count=1 -timeout 20m ./...` rc 0; 41 `ok`, 0 `FAIL`, 0 `WARNING: DATA RACE`, 0 `--- SKIP`; лог `/tmp/01-verify-full-battery.log` |
| 12 | `TestAcceptanceGates` проходит: Gates A–D ×2 против ratified baseline `1e9a596`, presence assert'ится и не skip'ается | ✓ VERIFIED | `--- PASS: TestAcceptanceGates (9.64s)`, exact RUN = 1, FAIL = 0; baseline `git cat-file -e 1e9a596^{commit}` rc 0; never-skip проверка в `acceptance_gates_test.go:144-151`; Gates A–D ×2 — `:252-390` |
| 13 | `TestLintGoldenFullRegistry` проходит с пином 94 findings; константа не изменена; ожидания не редактировались | ✓ VERIFIED | `--- PASS: TestLintGoldenFullRegistry (0.02s)`, exact RUN = 1; `goldenFindingCount = 94` — ровно 1 вхождение; expectation-freeze `79e353a..HEAD` — 0 файлов |
| 14 | D-07 position-гейты проходят | ✓ VERIFIED | `TestPositionRuneColumns` — 19 PASS, exact RUN = 1; `TestProtoastPositionsMatchLegacyContract` — 11 PASS, exact RUN = 1; оба rc 0 |
| 15 | Core/CLI parity-матрица проходит по точным именам тестов | ✓ VERIFIED | `TestLintMatrix`, `TestLintErrorPathFidelity`, `TestLintSuppressionParity` — по 1 RUN/PASS; `TestLintStdoutGolden`, `TestLintLinkMatrix`, `TestLintLinkMatrixExitCodes`, `TestPrintIssuesCompileGolden` — по 1 RUN/PASS; 0 FAIL |
| 16 | Тестируемое дерево — дерево мержа: `MERGE_SHA` — предок `HEAD`, вне `.planning` diff пуст, tracked-tree чистое до/после | ✓ VERIFIED | `git merge-base --is-ancestor 994c503 HEAD` rc 0; `git diff 994c503 HEAD -- ':(exclude).planning'` пуст (`non_planning_after_merge=0`); `git status --porcelain --untracked-files=no` пуст |
| 17 | Edge/ordering-контракт: до мержа ветка ровно на 1 коммит позади; порядок родителей стабилен; набор изменённых файлов отсортирован; empty-edge guard не сработал | ✓ VERIFIED | `git rev-list --count 79e353a..5d36266` = 1 (pre-merge behind = 1, значит empty-edge не применим); parents зафиксированы в порядке `79e353a 5d36266`; changeset получен сортированным; lint-порядок пинится golden 94 (truth 13) |
| 18 | Отложенные пункты зафиксированы честно: errcheck — Phase 3; push вне scope; `task`/`docker` отсутствуют | ✓ VERIFIED | `01-02-SUMMARY.md` «Issues Encountered» и Next Phase Readiness; `01-REVIEW-DISPOSITION.md` (open: 0, total: 10, все deferred); ROADMAP Phase 3 — владелец lint-гейта; README/SUMMARY фиксируют push out of scope |

**Score:** 18/18 truths verified (0 present, behavior-unverified)

### Prohibitions (must-NOT)

| Statement | Status | Evidence |
|-----------|--------|----------|
| MUST NOT resolve `breaking_profiles.go` со стороны ветки или вернуть legacy bufbuild compiler | ✓ VERIFIED (did not happen) | blob = `origin/main`; `easyprotoc/protoc` присутствует; `protocompile` = 0 |
| MUST NOT rebase/amend/rewrite опубликованную историю и MUST NOT push | ✓ VERIFIED (did not happen) | remote tip `origin/feat/lint-easypc-infra` — предок pre-merge base (`is-ancestor` rc 0); локально ahead 19; ровно один мерж-коммит; push не выполнялся |
| MUST NOT редактировать lint-контур, golden/parity-ожидания или sourceview-файлы внутри мерж-коммита | ✓ VERIFIED (did not happen) | lint-контур zero-diff; main-side файлы byte-identical `origin/main`; freeze ожиданий пуст |
| MUST NOT стейджить untracked planning/runtime артефакты; только явные pathspecs и ровно 9 файлов | ✓ VERIFIED (did not happen) | first-parent changeset = 9 ожидаемых путей; `.planning` в коммите отсутствует |

### Deferred Items

Пункты, сознательно не входящие в Phase 1 и покрытые последующими фазами (не gap'ы).

| # | Item | Addressed In | Evidence |
|---|------|-------------|----------|
| 1 | Lint-потребители всё ещё импортируют `internal/easypc/protoc`; переключение на `easyprotoc/protoc` не выполнено | Phase 2 (IMPORT-CONS-01) | CONTEXT D-03/D-06; ROADMAP Phase 2 SC1 перечисляет `internal/core/link.go`, `internal/protoast/*`; Phase 1 обязана сохранять вендор-путь байт-замороженным |
| 2 | Main's pre-existing golangci-lint errcheck finding в `internal/sourceview/session_test.go:29` | Phase 3 | `01-REVIEW-DISPOSITION.md` WR-01 deferred; D-09 исключает `lint:go` из Phase 1; ROADMAP Phase 3 SC1 |
| 3 | Push ветки в origin | Out of scope фазы (milestone decision) | `git status -sb`: ahead 19; remote ref не двигался; SUMMARY 01-01/01-02 фиксируют push out of scope |

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/core/breaking_profiles.go` | main's #253 breaking path на `easyprotoc/protoc` | ✓ VERIFIED | exists; blob = `origin/main`; `easyprotoc/protoc` присутствует; gsd-tools artifacts: passed |
| `internal/core/core.go` | union lint-state и sourceview-session state | ✓ VERIFIED | exists; union-блоб `34987d1`; `sourceSession`/`importCache`/`Options.SourceSession` на месте. gsd-tools пометил literal-паттерн `sourceSession *sourceview.Session` отсутствующим из-за gofmt-выравнивания (несколько пробелов); ручная проверка строки 32 подтверждает поле |
| `internal/core/dom.go` | `type Position = protoast.Position` | ✓ VERIFIED | exists; blob = pre-merge base; alias строка 19 |
| `go.mod` | MVS-пины из `origin/main` | ✓ VERIFIED | обе точные версии; 0 replace/vendor |
| `internal/sourceview/session.go` | per-command root cache из #253 | ✓ VERIFIED | exists; `NewSession`, `OpenLocal`, `Close`; byte-identical `origin/main` |
| `internal/sourceview/root_tree.go` | per-directory root handles из #253 | ✓ VERIFIED | exists; byte-identical `origin/main` |
| `internal/api/breaking_v1.go` | session lifecycle wiring из #253 | ✓ VERIFIED | exists; `sourceview.NewSession()`, `defer`-замыкание с `session.Close()`, `SourceSession` в `core.Options`; byte-identical `origin/main` |
| `internal/api/acceptance_gates_test.go` | Gates A–D ×2, baseline `1e9a596` | ✓ VERIFIED | exists; константа ровно 1 вхождение; тест PASS |
| `internal/rules/golden_full_registry_test.go` | 94-finding golden с пином | ✓ VERIFIED | exists; `goldenFindingCount = 94` ровно 1 вхождение; тест PASS |
| `internal/protoast/position_test.go` | D-07a rune-column gate | ✓ VERIFIED | exists; PASS |
| `internal/core/position_parity_test.go` | D-07b legacy position-contract gate | ✓ VERIFIED | exists; PASS |
| `internal/api/lint_link_matrix_test.go` | CLI link matrix, exit codes, COMPILE golden | ✓ VERIFIED | exists; все 4 именованных теста PASS |
| `.planning/phases/01-merge-main-i-primirenie-putey/01-02-SUMMARY.md` | gate evidence table, SHAs, deferred items | ✓ VERIFIED | exists; identifiers/gates/deferred совпадают с независимо перепроверенными значениями |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/core/breaking_profiles.go` | `go.mod` | `protoc.Compile` + `breakingImportResolver` требуют пин `easyprotoc/protoc` | ✓ WIRED | импорт строка 14, `protoc.Compile` строка 122, resolver строка 161; пин в `go.mod:10`. gsd-tools вернул false по literal-пути `go.mod` — ложный negative |
| `internal/core/core.go` | `internal/sourceview/session.go` | `sourceSession *sourceview.Session`, `Options.SourceSession`, `c.sourceSession.OpenLocal` | ✓ WIRED | поле строка 32, присваивание строка 114, вызов `OpenLocal` в `breakingSourceAccessor` (`breaking_profiles.go:221`); `session.go` определяет `OpenLocal` строка 61 |
| `internal/core/dom.go` | `internal/protoast/protoast.go` | `type Position = protoast.Position` | ✓ WIRED | alias строка 19, импорт `internal/protoast` строка 12 |
| `internal/api/breaking_v1.go` | `internal/sourceview/session.go` | `sourceview.NewSession()` + deferred `session.Close()` → `core.Options.SourceSession` | ✓ WIRED | строки 74/76 и 306/308 (NewSession/Close), `SourceSession: session` строки 155 и 479 |
| `internal/api/acceptance_gates_test.go` | `internal/api` | `acceptanceBaselineCommit = "1e9a596"`; `buildEasypBinary` строит working tree | ✓ WIRED | константа строка 29; тест PASS; baseline commit присутствует |
| `internal/rules/golden_full_registry_test.go` | `internal/rules` | `goldenFindingCount = 94` до golden-сравнения | ✓ WIRED | константа строка 60; `require.Len(..., goldenFindingCount)` до compare; тест PASS |
| `internal/protoast/position_test.go` | `internal/protoast` | D-07a rune-column контракт | ✓ WIRED | тест PASS (19 subtest PASS); gsd-tools false по literal-пути, ручная проверка подтверждает |
| `internal/core/position_parity_test.go` | `internal/core/dom.go` | `fromMetaPos` + `Position = protoast.Position` | ✓ WIRED | тест PASS (11 subtest PASS); alias подтверждён в `dom.go:19` |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| `internal/core/breaking_profiles.go` | `protoc.Compile(...)` | `github.com/easyp-tech/easyprotoc/protoc` из `go.mod` | Да — реальный compile/resolver путь, не статический ответ | ✓ FLOWING |
| `internal/api/breaking_v1.go` | `session` → `core.Options.SourceSession` | `sourceview.NewSession()` | Да — lifecycle до `OpenLocal` в `breakingSourceAccessor` | ✓ FLOWING |
| `internal/core/core.go` | `sourceSession`, `importCache` | `Options.SourceSession`; lint-парсинг импортов | Да — оба состояния реально используются | ✓ FLOWING |
| Lint-контур (`internal/rules`, `internal/easypc`, `internal/protoast`, `internal/core/link.go`, `lint_*`) | lint findings / golden | Вендор-путь `internal/easypc/protoc`, byte-frozen | Да — 94 findings golden и Gates A–D ×2 подтверждены | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Merge ancestry + aliveness | `git merge-base --is-ancestor origin/main HEAD`; `git rev-list --count HEAD..origin/main` | rc 0; 0 | ✓ PASS |
| Remote main unchanged | `git ls-remote origin refs/heads/main` | `5d36266…` | ✓ PASS |
| Merge shape | `git log -1 --format=%P 994c503` | `79e353a… 5d36266…` | ✓ PASS |
| First-parent changeset | `git diff --name-only 994c503^1 994c503` | 9 ровно ожидаемых файлов | ✓ PASS |
| Conflict markers | `git grep -E '^(<<<<<<< |>>>>>>> )'` | 0 | ✓ PASS |
| Lint-contour freeze | `git diff --stat 79e353a 994c503 -- <contour>` | пусто | ✓ PASS |
| Main-side equality | `git diff origin/main 994c503 -- breaking/sourceview/api` | пусто | ✓ PASS |
| Build | `go build ./...` | rc 0 | ✓ PASS |
| Vet | `go vet ./...` | rc 0 | ✓ PASS |
| Tidy | `go mod tidy -diff` | 0 байт | ✓ PASS |
| Full battery | `go test -race -count=1 -timeout 20m ./...` | rc 0; 41 ok / 0 FAIL / 0 race / 0 skip | ✓ PASS |
| Acceptance A–D ×2 | `go test -race -count=1 ./internal/api -run '^TestAcceptanceGates$' -v` | PASS 9.64s; exact RUN 1 | ✓ PASS |
| Golden 94 | `go test -race -count=1 ./internal/rules -run '^TestLintGoldenFullRegistry$' -v` | PASS 0.02s; exact RUN 1 | ✓ PASS |
| D-07 positions | `go test -race -count=1 ./internal/protoast -run '^TestPositionRuneColumns$' -v`; `./internal/core -run '^TestProtoastPositionsMatchLegacyContract$' -v` | 19 PASS / 11 PASS | ✓ PASS |
| Core matrix | `go test -race ./internal/core -run 'TestLintMatrix\|TestLintErrorPathFidelity\|TestLintSuppressionParity' -v` | 3/3 PASS, 0 FAIL | ✓ PASS |
| CLI matrix | `go test -race ./internal/api -run 'TestLintStdoutGolden\|TestLintLinkMatrix$\|TestLintLinkMatrixExitCodes\|TestPrintIssuesCompileGolden' -v` | 4/4 PASS, 0 FAIL | ✓ PASS |
| Expectation freeze | `git diff --name-only 79e353a HEAD -- internal/rules internal/protoast <core/api expectation files>` | 0 файлов | ✓ PASS |
| Tree attribution | `git diff 994c503 HEAD -- ':(exclude).planning'` | пусто | ✓ PASS |

Прогоны выполнены независимо верификатором на текущем `HEAD=77c93ec` (docs-only потомок мержа); battery запущена один раз с `-count=1`.

### Probe Execution

| Probe | Command | Result | Status |
|-------|---------|--------|--------|
| — | — | Probe-скрипты фазой не объявлены (`grep` по PLAN/SUMMARY пуст); conventional `scripts/*/tests/probe-*.sh` отсутствуют (0) | N/A |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| IMPORT-REC-01 | `01-01`, `01-02` | Merge `origin/main` (#253) с семантически разрешёнными конфликтами `internal/core`; breaking на `easyprotoc/protoc`, protobuf `v1.37.0-rc.1`; ветка не отстаёт; полная батарея зелёная | ✓ SATISFIED | Truths 1–18: ancestry/behind 0, blob-гейты, пины, зелёная battery, parity без правки ожиданий |

**Orphaned requirements:** нет. В `REQUIREMENTS.md` к Phase 1 привязан только `IMPORT-REC-01`; он объявлен в frontmatter обоих планов. Остальные ID (`IMPORT-DEP-01`, `IMPORT-CONS-01`, `IMPORT-VENDRM-01`, `IMPORT-SEAL-01`, `IMPORT-PARITY-01`) привязаны к Phase 2–4.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| `internal/api/acceptance_gates_test.go` | IMPORT-REC-01 | Да | 0 | Нет — оракул внешний (бинарник baseline `1e9a596`), provenance VALID | Value/Behavioral (multiset identity, NEW3 delta, zero-COMPILE) | PASS |
| `internal/rules/golden_full_registry_test.go` | IMPORT-REC-01 | Да | 0 | Нет — literals заморожены pre-flip; пин 94 до compare | Value (byte-level golden + count) | PASS |
| `internal/protoast/position_test.go` | IMPORT-REC-01 | Да | 0 | Нет | Value/Behavioral (rune columns) | PASS |
| `internal/core/position_parity_test.go` | IMPORT-REC-01 | Да | 0 | Нет | Value/Behavioral (legacy contract parity) | PASS |
| `internal/api/lint_link_matrix_test.go` / `lint_stdout_golden_test.go` | IMPORT-REC-01 | Да | 0 | Нет | Value (stdout/link/exit-code golden) | PASS |
| `internal/core/lint_linked_matrix_test.go` / `lint_suppressions_parity_test.go` | IMPORT-REC-01 | Да | 0 | Нет | Value (matrix/suppression parity) | PASS |

**Disabled tests on requirements:** 0 → не блокер. **Circular patterns detected:** 0. **Insufficient assertions:** 0.

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts. Gate: `honored 9/9`, `not_honored: []`, `blocking: false` (D-01…D-09).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | `TBD`/`FIXME`/`XXX`/`TODO`/`HACK`/`PLACEHOLDER` в файлах, изменённых фазой | ℹ️ Info | Не найдено. Единственное совпадение пустого-impl скана — `return []string{owner}, nil` в `internal/api/breaking_v1.go:226`, легитимная ветка выбора profile, не stub |

Code review findings WR-01…WR-03 и IN-01…IN-07 из `01-REVIEW.md` — предсуществующие/upstream, все `deferred` (`open: 0`), не debt-маркеры и не блокеры фазы.

### Human Verification Required

N/A — фаза инфраструктурная (merge/integration), пользовательских элементов нет; все acceptance-критерии проверены программно. Единственный Manual-Only пункт `01-VALIDATION.md` («merge commit message quality») проверен по содержанию: body содержит `#253`, все три разрешённых файла и правило main-wins/vendor-dormant. Стилистическая оценка прозы не является must-have и не переводит фазу в `human_needed`.

### Gaps Summary

Gap'ов нет. Все 18 truths и 4 prohibitions подтверждены фактическим состоянием git-истории и независимыми прогонами тестов; все 13 артефактов существуют, содержательны и подключены; все 8 key links подтверждены в исходниках. Ветка содержит актуальный `origin/main` (#253), `internal/core` разрешён семантически (main wins для breaking/sourceview, union для `core.go`, ветка сохраняет `protoast.Position`), lint-контур байт-заморожен, полная `-race` батарея и паритет-оракулы зелёные без правки ожиданий.

---

_Verified: 2026-11-21T12:04:56Z_
_Verifier: Claude (gsd-verifier)_
