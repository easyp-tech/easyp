---
phase: 02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii
verified: 2026-11-21T13:44:09Z
status: passed
score: 18/18 must-haves verified
covered_files:
  - ".planning/phases/02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii/02-01-PLAN.md"
  - ".planning/phases/02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii/02-01-SUMMARY.md"
  - ".planning/phases/02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii/02-02-PLAN.md"
  - ".planning/phases/02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii/02-02-SUMMARY.md"
  - "go.mod"
  - "go.sum"
  - "internal/api/acceptance_gates_test.go"
  - "internal/api/lint_link_matrix_test.go"
  - "internal/api/lint_stdout_golden_test.go"
  - "internal/core/link.go"
  - "internal/core/link_test.go"
  - "internal/core/lint_linked_matrix_test.go"
  - "internal/core/position_parity_test.go"
  - "internal/protoast/build.go"
  - "internal/protoast/outline.go"
  - "internal/protoast/position_test.go"
  - "internal/protoast/protoast.go"
  - "internal/protoast/protoast_test.go"
  - "internal/rules/golden_full_registry_test.go"
covered_digest: "v4:sha256:cc31c4fb73e0d54e8234d0be01ec5bb2abece79d3f0bbad8ccb6d89b5c67dd6a"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 2: Переключение потребителей и удаление вендор-копии Verification Report

**Phase Goal:** Все потребители компилятора импортируют публичные пакеты `github.com/easyp-tech/easyprotoc`, вендор-дерево `internal/easypc` удалено, сборка и полная тестовая батарея зелёные при идемпотентном `go mod tidy`
**Verified:** 2026-11-21T13:44:09Z
**Status:** passed
**Re-verification:** No — initial verification

**Mode note (MVP guard).** ROADMAP помечает фазу как `Mode: mvp`, но её Goal не является user story: `user-story.validate` вернул `valid: false` с четырьмя ошибками формата (`As a … I want … so that …`). Фаза инфраструктурная (import-swap + удаление дерева + тестовая батарея), пользовательского потока для User Flow Coverage нет; обе PLAN-файлы явно фиксируют этот разрыв (MVP framing note, ratified Phase 1 precedent) и несут контракт через goal-backward truths. Проверка выполнена по четырём Success Criteria ROADMAP и `must_haves` обоих планов; User Flow Coverage не фабриковался. Если для строгого MVP-контракта нужна нормализация Goal, владельцу следует выполнить `/gsd:mvp-phase 02`; это метаданные фазы, а не дефект её результата, и отдельного human-test пункта не создаёт.

**Provenance.** Evidence ниже получен независимо: состояние git пересчитано в текущем дереве (`HEAD c044ee1`), полная батарея запущена один раз самим верификатором, upstream-бинарь собран и запущен на временном `.proto`. Утверждения SUMMARY перепроверены, а не приняты на веру.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | (ROADMAP SC1) Пути `github.com/easyp-tech/easyp/internal/easypc/...` отсутствуют в tracked Go-коде | ✓ VERIFIED | `git grep -n 'internal/easypc' -- '*.go'` → rc 1, 0 строк; `git grep` по не-`.go` даёт только `.golangci.yml`/`AGENTS.md` (принято D-07) и prose без пути |
| 2 | (ROADMAP SC1) Потребители импортируют публичные пакеты: ровно 9 строк `github.com/easyp-tech/easyprotoc/`, из них 8 в шести swap-файлах, per-file 1/1/1/2/2/0/1 | ✓ VERIFIED | `git grep -h -E '^[[:space:]]*"github\.com/easyp-tech/easyprotoc/' -- '*.go' \| wc -l` = 9; список: `breaking_profiles.go:14`, `link.go:17`, `link_test.go:10`, `lint_linked_matrix_test.go:15`, `build.go:10-11`, `outline.go:4-5`, `protoast_test.go:6`; импортируются все три пакета `protoc`, `protoc/syntax`, `parser` |
| 3 | (ROADMAP SC1) Импортированные символы реально используются, а не существуют формально | ✓ VERIFIED | `link.go`: `protoc.Link/Diagnostic/Resolver/ErrCompilationFailed/ErrFileNotFound`; `build.go`: `syntax.Parse`, `parser.Decl*`; `outline.go`: `parser.*`/`syntax.File`; `protoast_test.go`: `syntax.Parse`; `go build ./...` rc 0 |
| 4 | (02-01) Правка семи файлов доказуемо не behavioral: 14 удалённых строк содержат `internal/easypc`, 14 добавленных непустых — `easyprotoc`; strip-gate молчит | ✓ VERIFIED | awk-гейт по `git diff -U0 df23d64 5342455`: `removed=14 added=14 bad=0`; per-file strip-diff (комментарии, пустые строки, import-пути) — `STRIP-OK` для всех 7 файлов |
| 5 | (ROADMAP SC2) Дерево `internal/easypc` отсутствует целиком, включая `VENDORED.md`, `cmd/easyp-protoc`, `cmd/lexbench` | ✓ VERIFIED | `git ls-files internal/easypc` = 0; `git ls-tree -r 5342455 -- internal/easypc` = 0; `git cat-file -e 5342455:internal/easypc` → ABSENT; каталога нет на диске; все три именованных пути ABSENT; `cmd/` содержит только `easyp`, `easyp-mcp` |
| 6 | (02-01) Удаление атомарно и точно: один коммит, 648 D + 7 M + 0 A, удаления только внутри `internal/easypc` | ✓ VERIFIED | `git rev-list --count df23d64..5342455` = 1; `git diff --name-status` = 655 путей (`D=648 M=7 A=0`); удалений вне `internal/easypc/` = 0; разбивка 555 `.proto` + 48 `.go` + 30 `.tokens` + 14 `.golden` + 1 `VENDORED.md` |
| 7 | (02-01) Атомарный коммит собирается сам: build/vet/gofmt чисты, быстрые consumer-пакеты зелёные | ✓ VERIFIED | `go build ./...` rc 0; `go vet ./...` rc 0, stderr 0 байт; `gofmt -l` по семи файлам пуст; `internal/core`, `internal/protoast`, `internal/rules` — `ok` в полной батарее |
| 8 | (ROADMAP SC3) Полная `-race` батарея зелёная: rc 0, 34 `ok`, 0 FAIL, 0 `--- FAIL`, 0 race, только известный skip, 3 no-test | ✓ VERIFIED | Независимый прогон `go test -race -count=1 -timeout 20m -v ./...` rc 0 (`/tmp/verify-02-battery.log`, 13891 строка): `ok=34`, `FAIL=0`, `--- FAIL=0`, `WARNING: DATA RACE=0`, `--- SKIP=1` (`TestGenerateGoHeaderUsesCompilerVersion`, protoc-gen-go отсутствует в PATH), `?`=3 (`cmd/easyp-mcp`, `internal/logger`, `wellknownimports`); tracked tree чист до и после |
| 9 | (ROADMAP SC3) Пины точные и прямые, совпадают с `origin/main` #253; `replace`/`vendor` отсутствуют; `go.mod`/`go.sum` не изменены | ✓ VERIFIED | `go list -m`: `easyprotoc v0.0.0-20261104120806-f378f3f3a05f indirect=false`, `protobuf v1.37.0-rc.1 indirect=false`; те же строки в `origin/main:go.mod`; `grep -c '^replace'`=0, `'^vendor'`=0; `git diff` base..swap и swap..HEAD по `go.mod go.sum` пуст; `go mod verify` — all modules verified |
| 10 | (ROADMAP SC3) Повторный `go mod tidy` не даёт дрейфа | ✓ VERIFIED | `go mod tidy -diff` → 0 байт, rc 0 |
| 11 | (02-01 / IMPORT-VENDRM-01) protoc-совместимый бинарь доступен через модуль без локальной копии | ✓ VERIFIED | `go list -f '{{.Name}}' …/cmd/easyprotoc` = `main`; `go build -o /dev/null …/cmd/easyprotoc` rc 0; поведенческий прогон `go run …/cmd/easyprotoc -I /tmp/easyprotoc-verify -o …/out.bin …/check.proto` rc 0, получен descriptor-set (64 байта) |
| 12 | (02-01) Пост-состояние графа пакетов соответствует измеренному: 37 пакетов, 34 с тестами, 3 без | ✓ VERIFIED | `go list ./... \| wc -l` = 37; test-package фильтр = 34; no-test ровно 3 |
| 13 | (02-02) `TestAcceptanceGates` проходит: Gates A–D ×2 против baseline `1e9a596`, baseline присутствует, константа не сдвинута | ✓ VERIFIED | В батарее: `=== RUN   TestAcceptanceGates` ×1, `--- PASS: TestAcceptanceGates (7.05s)`; `git cat-file -e 1e9a596^{commit}` rc 0; `grep -c 'acceptanceBaselineCommit = "1e9a596"'` = 1; тест строит baseline-бинарь из worktree `1e9a596` и сравнивает 4 прогона (main-r1/r2, new-r1/r2), Gates A–D |
| 14 | (02-02) `TestLintGoldenFullRegistry` проходит байт-точно с 94 findings; константа единственна; ожидания не пересняты | ✓ VERIFIED | В батарее: `=== RUN   TestLintGoldenFullRegistry` ×1, `--- PASS: … (0.08s)`; `grep -c 'goldenFindingCount = 94'` = 1; сравнение `require.Equal(t, goldenIssues, got)` с purity/total-order guard; `internal/rules` zero-diff к base |
| 15 | (ROADMAP SC4) Дивергенция protoast сохранена без правок шва; оба D-07 гейта зелёные | ✓ VERIFIED | `TestPositionRuneColumns` 19 PASS-строк, `TestProtoastPositionsMatchLegacyContract` 11 PASS-строк, у каждого exact RUN=1/PASS=1; strip-gate подтверждает отсутствие изменений `pos()` и фикстур; `position_test.go:13-21` и `position_parity_test.go:16-21` фиксируют контракт |
| 16 | (02-02) Core/CLI parity-матрица проходит по точным именам тестов | ✓ VERIFIED | 7 тестов, у каждого `RUN=1 PASS=1`, `--- FAIL=0`: `TestLintMatrix`, `TestLintErrorPathFidelity`, `TestLintSuppressionParity` (43 PASS-строк суммарно), `TestLintStdoutGolden`, `TestLintLinkMatrix`, `TestLintLinkMatrixExitCodes`, `TestPrintIssuesCompileGolden` (16 PASS-строк суммарно) |
| 17 | (02-02) Ноль сдвинутых ожиданий: modified-set ровно семь D-02 файлов, `internal/rules` и `internal/api` byte-identical к base, константы неизменны | ✓ VERIFIED | `git diff --diff-filter=M df23d64 5342455` — ровно 7 путей; `git diff df23d64 5342455 -- internal/rules internal/api` пуст; обе константы по одному вхождению |
| 18 | (02-02) Evidence атрибутируемо: HEAD — docs-only потомок swap-коммита, tracked tree чист | ✓ VERIFIED | `git merge-base --is-ancestor 5342455 HEAD` true; `git diff --name-only 5342455 HEAD -- ':(exclude).planning'` пуст; после swap только 5 docs-коммитов, все `.planning`; `git status --porcelain --untracked-files=no` пуст до/после прогонов |

**Score:** 18/18 truths verified (0 present, behavior-unverified).

### Prohibitions (must-NOT)

| Statement | Status | Evidence |
|-----------|--------|----------|
| D-01: не вводить import alias, shim-пакет или `replace`/`vendor` | ✓ VERIFIED (did not happen) | alias-импортов `easyprotoc` = 0; `go.work`/`vendor` отсутствуют; `replace`/`vendor` в `go.mod` = 0; добавленных файлов 0 |
| D-03: не разделять swap и удаление на коммиты, не оставлять полу-удалённое дерево | ✓ VERIFIED (did not happen) | Ровно один feat-коммит `5342455`; 648 удалений в нём же; build на коммите зелёный |
| D-09: не править ожидания, golden, parity-фикстуры, правила, движок, API, veto-файлы | ✓ VERIFIED (did not happen) | modified-set = семь consumer-файлов; `internal/rules`/`internal/api` zero-diff; константы single-occurrence; никаких re-capture |
| D-05: не менять логику `internal/protoast`, включая rune-колонки `pos()` и tab-фикстуры | ✓ VERIFIED (did not happen) | Strip-gate молчит для всех семи файлов; менялись только import-строки и комментарии |
| D-07: не трогать `.golangci.yml`, `AGENTS.md`, `.spec/**` в этой фазе | ✓ VERIFIED (did not happen) | `git diff --name-only df23d64 HEAD -- .golangci.yml AGENTS.md .spec` пуст — за всю фазу, не только в swap-коммите |
| D-04: не добавлять локальный replacement-бинарь и не портировать `cmd/easyp-protoc`/`cmd/lexbench` | ✓ VERIFIED (did not happen) | 0 добавленных файлов; `cmd/` = `easyp`, `easyp-mcp`; upstream `cmd/easyprotoc` строится и запускается |
| D-08: не делать push в этой фазе | ✓ VERIFIED (did not happen) | `git branch -r --contains 5342455` = 0; `git ls-remote origin` не содержит SHA; ветка локально ahead 30 |
| 02-02: evidence-only, без изменений исходников | ✓ VERIFIED (did not happen) | Коммиты после swap — только `.planning`; non-planning diff от `5342455` до HEAD пуст |
| 02-02: не перезапускать батарею «до зелёного», не ослаблять `-race`/`-timeout` | ✓ VERIFIED (evidence consistent) | SUMMARY фиксирует один прогон с ratified-флагами; верификатор независимо выполнил ровно один прогон; признаков повторных прогонов/ослабления нет |
| 02-02: не считать `task lint:go`/`task schema:check` гейтами Phase 2 | ✓ VERIFIED (did not happen) | Эти команды не заявлены в evidence; владение — Phase 3/4 (D-09) |

### Deferred Items

Пункты, сознательно не входящие в Phase 2 и покрытые последующими фазами (не gap'ы).

| # | Item | Addressed In | Evidence |
|---|------|-------------|----------|
| 1 | `.golangci.yml`-исключение `internal/easypc`, seal-параграф `AGENTS.md`, `.spec/**` | Phase 3 (IMPORT-SEAL-01) | CONTEXT D-07 и locked boundary фазы; `git diff` base..HEAD по этим путям пуст — сознательное сохранение, не пропуск |
| 2 | `task lint:go`, `task schema:check`, `task build` | Phase 3/4 (IMPORT-PARITY-01) | 02-02-SUMMARY «Deferred Items»; D-09 перечисляет гейты Phase 2 без lint/schema |
| 3 | Push ветки `feat/lint-easypc-infra` | Вне scope фазы (D-08) | `git status -sb`: ahead 30; swap-SHA не содержится ни в одной remote-ветке |
| 4 | Review INFO IN-01..IN-03 (doc/citation polish, stale prose) | Phase 3 docs pass | `02-REVIEW-DISPOSITION.md`: `open: 0`, все три `deferred`; не влияют на behavior |
| 5 | v1.0 corpora letters (LINT-PARITY-01/LINT-PERF-01) | Milestone backlog (`pending_user_unblock`) | 02-02-SUMMARY «Deferred Items»; не являются truth'ами Phase 2 |

### Required Artifacts

Все 16 объявленных артефактов (9 в 02-01 + 7 в 02-02) существуют, содержательны и подключены; `verify.artifacts` вернул `all_passed: true` для обоих планов.

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/core/link.go` | Линкер потребляет `protoc.Link/Linked/LinkedFile/Diagnostic/Resolver/…` из публичного модуля | ✓ EXISTS + SUBSTANTIVE + WIRED | Импорт строка 17; символы используются (44–250); третий-party группа отделена от проекта |
| `internal/core/link_test.go` | Контрактные тесты линкера на публичном `protoc` | ✓ EXISTS + SUBSTANTIVE + WIRED | Импорт строка 10; тесты проходят в батарее |
| `internal/core/lint_linked_matrix_test.go` | Матрица на `protoc.ErrFileNotFound` из модуля | ✓ EXISTS + SUBSTANTIVE + WIRED | Импорт строка 15; `TestLintMatrix`/`TestLintErrorPathFidelity` PASS |
| `internal/protoast/build.go` | `easyprotoc/parser` + `protoc/syntax`; `pos()` не изменён | ✓ EXISTS + SUBSTANTIVE + WIRED | Импорты строки 10–11; `syntax.Parse` строка 26; strip-gate silent |
| `internal/protoast/outline.go` | Те же два пакета; обход деклараций | ✓ EXISTS + SUBSTANTIVE + WIRED | Импорты строки 4–5; `parser.Decl*` используется |
| `internal/protoast/protoast.go` | Package doc ссылается на модульный путь и пин, не на `VENDORED.md` | ✓ EXISTS + SUBSTANTIVE | Импортов нет (по D-05); doc строки 1–11; ссылка на `build.go` + `go.mod` |
| `internal/protoast/protoast_test.go` | Фикстуры позиций/табов на `syntax.Parse` | ✓ EXISTS + SUBSTANTIVE + WIRED | Импорт строка 6; тесты пакета PASS |
| `go.mod` | Прямой точный пин `easyprotoc` и protobuf rc.1 | ✓ EXISTS + SUBSTANTIVE | Строки 10 и 24; `indirect=false`; zero-diff к base |
| `02-01-SUMMARY.md` | SHA, base, счётчики, gate output | ✓ EXISTS + SUBSTANTIVE | Все значения перепроверены независимо |
| `internal/api/acceptance_gates_test.go` | Gates A–D ×2, baseline `1e9a596` | ✓ EXISTS + SUBSTANTIVE + WIRED | Константа строка 29; baseline presence check 144–150; `TestAcceptanceGates` PASS |
| `internal/rules/golden_full_registry_test.go` | Golden 94 с анти-вакуумным пином | ✓ EXISTS + SUBSTANTIVE + WIRED | `goldenFindingCount = 94` строка 60; assert до compare (503); PASS |
| `internal/protoast/position_test.go` | D-07a rune-колонки | ✓ EXISTS + SUBSTANTIVE + WIRED | Hand-computed ожидания; 19 PASS-строк |
| `internal/core/position_parity_test.go` | D-07b legacy-контракт | ✓ EXISTS + SUBSTANTIVE + WIRED | `core.Position = protoast.Position`; 11 PASS-строк |
| `internal/api/lint_link_matrix_test.go` | CLI link matrix, exit codes, COMPILE golden | ✓ EXISTS + SUBSTANTIVE + WIRED | 3 теста PASS по exact RUN |
| `internal/api/lint_stdout_golden_test.go` | Byte-exact stdout/JSON golden | ✓ EXISTS + SUBSTANTIVE + WIRED | PASS в батарее |
| `02-02-SUMMARY.md` | Gate table, SHAs, константы, deferred | ✓ EXISTS + SUBSTANTIVE | Значения сверены с независимым прогоном |

### Key Link Verification

`verify.key-links` — эвристика «to-цель упомянута в from-файле»; для целей вида `go.mod` и sibling-файлов она даёт ложные «Target not referenced in source». Ниже связи проверены вручную по импортам, вызовам и тестам.

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `internal/core/link.go` | `go.mod` | `easyprotoc/protoc` на пине | ✓ WIRED | Импорт строка 17; `protoc.Link` 53, `protoc.WithResolver` 54, `protoc.Link`/`Diagnostic` по всему файлу; build/тесты зелёные |
| `internal/protoast/build.go` | `go.mod` | `easyprotoc/parser` + `protoc/syntax` | ✓ WIRED | Импорты 10–11; `syntax.Parse` 26; `parser.ASTFile` 60; golden/position PASS |
| `internal/protoast/outline.go` | `internal/protoast/build.go` | общие `parser`/`syntax` и `builder` | ✓ WIRED | Один пакет `protoast`; `OutlineOf` использует `syntax.File` и `builder` из build.go; `go build` зелёный |
| `internal/protoast/protoast_test.go` | `internal/protoast/build.go` | `Read`/`Parse` на фикстурах | ✓ WIRED | Тест вызывает `Read(...)`; пакетные тесты PASS |
| `internal/easypc/` | `internal/core/link.go` | удаление безопасно только после swap 8 импортов | ✓ SATISFIED | Дерево удалено тем же коммитом; `git grep` пуст; build/battery зелёные |
| `internal/core/breaking_profiles.go` | `go.mod` | уже импортирует `easyprotoc/protoc` (main #253) | ✓ WIRED | Импорт строка 14; файл не входит в modified-set и не изменён; батарея PASS |
| `internal/api/acceptance_gates_test.go` | `internal/api` | `buildEasypBinary` собирает working tree | ✓ WIRED | Сборка `./cmd/easyp`; attribution-гейт пуст; `TestAcceptanceGates` PASS |
| `internal/rules/golden_full_registry_test.go` | `internal/rules` | `goldenFindingCount` до compare; корпус через swapped `protoast` | ✓ WIRED | Const 60, assert 503; PASS |
| `internal/protoast/position_test.go` | `internal/protoast/build.go` | `Read` над `syntax.Parse`-деревом | ✓ WIRED | Вызов `Read`; PASS |
| `internal/core/link_test.go` | `internal/core/link.go` | тот же пакет, `linkSources` | ✓ WIRED | Same-package тесты; PASS |
| `internal/api/lint_link_matrix_test.go` | `cmd/easyp` | subprocess exit codes 0/1/2 | ✓ WIRED | Build результата; `TestLintLinkMatrixExitCodes` PASS |

**Wiring:** 11/11 связей подтверждены.

### Data-Flow Trace (Level 4)

UI/динамического рендеринга в фазе нет; трассируется источник компилятора — от импорта до реального модуля.

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| `internal/core/link.go` | `*protoc.Linked`, `[]protoc.Diagnostic` | `protoc.Link` из `easyprotoc@v0.0.0-20261104120806-f378f3f3a05f` | Да — модуль резолвится, линк-матрица/parity PASS | ✓ FLOWING |
| `internal/protoast/build.go` | `*syntax.File`, `*parser.ASTFile` | `syntax.Parse`/`parser` из того же модуля | Да — golden 94 и position-гейты PASS | ✓ FLOWING |
| `internal/protoast/outline.go` | `*syntax.File`, `parser.Decl*` | тот же модуль | Да — пакетные тесты PASS | ✓ FLOWING |
| `go.sum`/модульный кэш | integrity-хеши | Go proxy/module cache | Да — `go mod verify`: all modules verified | ✓ FLOWING |

Статических заглушек, hardcoded-фолбэков и mock-источников не обнаружено.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Сборка | `go build ./...` | rc 0 | ✓ PASS |
| Vet | `go vet ./...` | rc 0, stderr 0 байт | ✓ PASS |
| Tidy-идемпотентность | `go mod tidy -diff` | 0 байт | ✓ PASS |
| Пины | `go list -m -f '{{.Path}} {{.Version}} {{.Indirect}}' …` | exact pins, `indirect=false` | ✓ PASS |
| Upstream-бинарь собирается | `go build -o /dev/null …/easyprotoc/cmd/easyprotoc` | rc 0 | ✓ PASS |
| Upstream-бинарь работает | `go run …/easyprotoc/cmd/easyprotoc -I /tmp/easyprotoc-verify -o …/out.bin …/check.proto` | rc 0, descriptor 64 байта | ✓ PASS |
| Граф пакетов | `go list ./...` + test-filter | 37 / 34 / 3 | ✓ PASS |
| Полная батарея | `go test -race -count=1 -timeout 20m -v ./...` | rc 0; 34 ok / 0 FAIL / 0 race / 1 known skip / 3 no-test | ✓ PASS |
| Именованные гейты | grep exact `=== RUN`/`--- PASS` в логе батареи | 11/11 RUN=1 PASS=1; 0 `--- FAIL` | ✓ PASS |
| Expectation freeze | `git diff` base..swap по rules/api, modified-set | пусто; ровно 7 путей | ✓ PASS |
| Изменения коммита | `git diff --name-status` | 648 D + 7 M + 0 A | ✓ PASS |
| Чистота дерева | `git status --porcelain --untracked-files=no` | пусто до/после | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
|-------|---------|--------|--------|
| — | — | Probe-скрипты фазой не объявлены (grep по PLAN/SUMMARY пуст); conventional `scripts/*/tests/probe-*.sh` отсутствуют | N/A |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| IMPORT-CONS-01 | `02-01`, `02-02` | Потребители импортируют `easyprotoc/protoc`, `.../protoc/syntax`, `.../parser`; пути `internal/easypc/...` отсутствуют (grep пуст) | ✓ SATISFIED | Truths 1–4, 15–18: grep rc 1, 9 импортов/8 swap, реальное использование символов, strip-gate, золотые/паритетные гейты без правок ожиданий |
| IMPORT-VENDRM-01 | `02-01`, `02-02` | Дерево `internal/easypc` отсутствует целиком (48+ файлов, `VENDORED.md`, оба cmd); protoc-совместимый бинарь доступен через `go run …/cmd/easyprotoc` без локальной копии | ✓ SATISFIED | Truths 5–6, 11: index/tree/worktree пусты, 648 удалений только в дереве, upstream-бинарь собран и выполнен |
| IMPORT-DEP-01 | `02-01`, `02-02` | `easyprotoc` в псевдоверсии main @ `f378f3f` без `replace`/`vendor`; повторный tidy без дрейфа | ✓ SATISFIED | Truths 9–10: exact direct pin, protobuf rc.1, zero replace/vendor, `tidy -diff` пуст, `go.mod`/`go.sum` zero-diff |

**Orphaned requirements:** нет. В `REQUIREMENTS.md` к Phase 2 привязаны ровно эти три ID, и все три объявлены в frontmatter обоих планов. `IMPORT-SEAL-01` привязан к Phase 3, `IMPORT-PARITY-01` — к Phase 4; `IMPORT-REC-01` закрыт в Phase 1.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| `internal/api/acceptance_gates_test.go` | IMPORT-CONS-01, IMPORT-VENDRM-01 | Да | 0 | Нет — оракул это baseline-бинарь из git-коммита `1e9a596` | Behavioral (4 прогона, Gates A–D) | ✓ PASS |
| `internal/rules/golden_full_registry_test.go` | IMPORT-CONS-01 | Да | 0 | Нет — golden зафиксирован до фазы, `internal/rules` zero-diff; count assert'ится до compare | Value (`require.Equal` с golden + purity/order) | ✓ PASS |
| `internal/protoast/position_test.go` | IMPORT-CONS-01 | Да | 0 | Нет — ожидания hand-computed над UTF-8/tab фикстурой | Value (точные Line/Column/Offset) | ✓ PASS |
| `internal/core/position_parity_test.go` | IMPORT-CONS-01 | Да | 0 | Нет — equality двух независимых моделей | Value (model equality) | ✓ PASS |
| `internal/api/lint_link_matrix_test.go` | IMPORT-CONS-01 | Да | 0 | Нет | Behavioral (exit codes/subprocess) + Value (COMPILE golden) | ✓ PASS |
| `internal/api/lint_stdout_golden_test.go` | IMPORT-CONS-01 | Да | 0 | Нет | Value (byte-exact stdout/JSON) | ✓ PASS |
| `internal/core/link_test.go`, `lint_linked_matrix_test.go` | IMPORT-CONS-01 | Да | 0 | Нет | Behavioral | ✓ PASS |
| `internal/core/generate_compiler_version_test.go` | Не гейт фазы | Да | 1 — environment guard (`protoc-gen-go` отсутствует) | Нет | Value | ✓ PASS (документированный skip) |

**Disabled tests on requirements:** 0. Symlink-тесты `internal/api/generate_config_test.go` имеют Windows-only `t.Skip`, на Linux не срабатывают. **Circular patterns:** 0. **Insufficient assertions:** 0 — все requirement-linked проверки value/behavioral уровня.

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts. Gate: `honored 9/9`, `not_honored: []`, `blocking: false` (D-01…D-09).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | `TBD`/`FIXME`/`XXX`/`TODO`/`HACK`/`PLACEHOLDER`, пустые реализации, placeholder-проза в изменённых файлах | — | Не найдено (grep rc 1 по всем семи файлам) |
| `internal/rules/package_no_import_cycle_test.go` | 28 | Prose «the easypc linker» без пути | ℹ️ Info | Известный residue вне D-02/D-09; reword в Phase 3 (review IN-03) |
| `internal/protoast/protoast.go` | 2 | Старое имя «easyp-protoc» в doc при корректном модульном пути | ℹ️ Info | Doc-drift, не влияет на behavior (review IN-02) |
| `internal/protoast/build.go` | ~458 | Комментарий «the vendored LineCol» (без пути) | ℹ️ Info | Doc-drift после удаления дерева; путь не упомянут |
| `.golangci.yml` / `AGENTS.md` | 7, 9 / 92 | Ссылки на `internal/easypc` | ℹ️ Accepted (out of scope) | Сознательно оставлено D-07; владелец — Phase 3; **не gap** по locked boundary фазы |

**Anti-patterns:** 0 blockers, 0 warnings, 4 info/accepted.

### Human Verification Required

N/A — фаза инфраструктурная (import-swap, удаление дерева, тестовая батарея), пользовательских элементов нет; все acceptance-критерии проверены программно. Единственный Manual-Only пункт `02-VALIDATION.md` («commit message quality») проверен по содержанию: subject `feat(lint): consume public easyprotoc module; drop vendored easypc tree` и тело коммита называют модульный swap и удалённое дерево. Стилистическая оценка прозы не является must-have и не переводит фазу в `human_needed`. MVP-несоответствие формата Goal зафиксировано выше как mode note; оно не является пользовательским тестом.

### Gaps Summary

Gap'ов нет. Все 18 truths и 10 prohibitions подтверждены фактическим состоянием git-истории и независимыми прогонами; все 16 артефактов существуют, содержательны и подключены; все 11 key links подтверждены в исходниках. Потребители компилятора импортируют публичные пакеты `easyprotoc` (9 импорт-строк, 8 из них — swap), вендор-дерево `internal/easypc` отсутствует в индексе, коммите и worktree (648 удалений в одном атомарном feat-коммите), `go build ./...`, `go vet ./...` и полная `-race` батарея зелёные (34 ok / 0 FAIL / 0 race / 1 известный environment skip / 3 no-test), `go mod tidy -diff` пуст при точных прямых пинах и нулевых `replace`/`vendor`. Ожидания не сдвинуты: `internal/rules` и `internal/api` byte-identical к pre-swap base, golden 94 и оба D-07 position-гейта проходят. Ceiling: известные doc-drift INFO-пункты и seal/docs-остатки явно отложены в Phase 3 и не являются gap'ами этой фазы.

---

_Verified: 2026-11-21T13:44:09Z_
_Verifier: Claude (gsd-verifier)_
