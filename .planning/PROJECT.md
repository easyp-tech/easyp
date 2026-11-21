# EasyP

## What This Is

EasyP — CLI-тулкит для Protocol Buffers на Go: линтинг (buf-совместимые правила), детекция breaking changes, генерация кода (локальные/remote/WASM-плагины) и Git-центричное управление зависимостями (`easyp mod`). Пользователи — разработчики proto-API и команды, держащие большие корпуса proto-файлов в CI. Этот GSD-проект — рамка для развития EasyP в целом; milestones добавляются по мере подготовки фич в `.spec/features/`.

## Core Value

EasyP остаётся drop-in buf-совместимым инструментом: формат вывода, коды выхода и конфигурация не ломаются между версиями — миграции внутренностей не должны быть заметны пользователю.

## Current Milestone: v1.1 import-easyprotoc

**Goal:** lint-путь EasyP потребляет публичный модуль `github.com/easyp-tech/easyprotoc` вместо вендор-копии `internal/easypc`; вендор-дерево и seal-инфраструктура удаляются.

**Target features:**
- `go.mod`: `require github.com/easyp-tech/easyprotoc` (псевдоверсия main @ `f378f3f` — та же, что breaking-путь на main с #253)
- Переписывание импортов потребителей (`internal/core/link.go`, `internal/protoast`) на публичные пакеты модуля: `protoc`, `protoc/syntax`, `parser`
- Удаление `internal/easypc` (включая `VENDORED.md`, `cmd/easyp-protoc`, `cmd/lexbench`), снятие `.golangci.yml`-исключения и seal-параграфа AGENTS.md
- Полный паритет: сборка, корпус lint (Gates A–D), D-07 позиции, `-race ./...`; `.spec` отражает импорт вместо вендоринга

## Current State

**v1.0 «lint-easypc-infra» SHIPPED 2026-11-21** (PR #258). `easyp lint` работает на собственном вендоренном компиляторе `easypc` + модели `protoast`: паритет замечаний с main (автоматизированное ядро — Gates A–D ×2, golden 94 findings байт-точно), COMPILE-замечания на некомпилирующихся корпусах, детерминированный порядок вывода, 44 правила (3 новых: PACKAGE_NO_IMPORT_CYCLE, SYNTAX_SPECIFIED, FIELD_NOT_REQUIRED), схемы/MCP атомарно обновлены. Полный `-race ./...` зелёный: 41 пакет, 3718 тестов, 0 FAIL; wazero-гонка устранена. Живые обязательства в backlog: corpora letters (ready-to-fire скрипты + unblock path), hadolint при появлении Docker.

## Next Milestone Goals

*(черновик — финализировать в `/gsd:new-milestone`)*
- Продолжение серии: перевод generate/`mod tidy`/`migrate` на модуль `easyprotoc` (FUTURE-COMPILER-01..03; breaking уже на easyprotoc с main #253)
- Bump псевдоверсии easyprotoc при появлении тегов у upstream / мержа ветки `feat/easyp-integration` (generation sources + descriptor metadata fix — нужен для codegen-серии, меняет семантику ошибок `Link`)
- Corpora acceptance: разблокировать приватные корпуса, прогнать lint-diff паритет + perf-гейт на полном масштабе (letters v1.0)
- Hardening: fuzzing лексера/парсера (FUTURE-HARDENING-01), editions support (FUTURE-COMPILER-04)
- Lab improvements deferred behind parity: resolution-based ImportUsed, cross-file RPCRequestResponseUnique, swift_prefix fix

## Business Context

- **Customer**: разработчики/команды, использующие `easyp` в CI для proto-корпусов (внутренние и внешние пользователи easyp.tech)
- **Revenue model**: open-source инструмент, продвигающий экосистему easyp-tech
- **Success metric**: паритет lint-замечаний с main на приёмочных корпусах при не худшей производительности
- **Strategy notes**: `.spec/features/lint-easypc-infra/` (explore.md, requirements.md)

## Requirements

### Validated

Существующие возможности кодовой базы (см. `.planning/codebase/`):

- ✓ Buf-совместимый линтер, 43 правила с colocated-тестами — existing
- ✓ Breaking-change detection с профилями FILE/PACKAGE/WIRE_JSON/WIRE — existing
- ✓ Генерация кода: локальные, remote (gRPC) и WASM-плагины — existing
- ✓ Git-центричные зависимости: `protobuf.mod` / `protobuf.lock`, verified checkouts, vendoring — existing
- ✓ v1-конфигурация: `easyp.yaml`, `easyp.gen.yaml`, JSON-схемы (`task schema:generate`), MCP-сервер конфигурации — existing
- ✓ Миграция legacy → v1 (`easyp migrate`, интерактивный мастер) — existing
- ✓ Вендоринг компилятора `easypc` и модели `protoast` из compiler-lab — Phase 1 (байт-verbatim, ноль новых зависимостей, пустой дифф поведения)
- ✓ Lint на слинкованных деревьях `easypc` + `protoast`: два прохода, центральные подавления, детерминированный вывод — v1.0 (PR #258)
- ✓ Паритет 43 правил + 3 новых правила, схемы/MCP согласованно обновлены, Gates A–D ×2, golden 94 findings байт-точно — v1.0
- ✓ Полный `-race ./...` зелёный (41 пакет, 3718 тестов), wazero-гонка устранена — v1.0 Phase 7

### Active

Milestone v1.1 «import-easyprotoc» — полный список в REQUIREMENTS.md (IMPORT-*):

- [ ] `go.mod` требует публичный модуль `github.com/easyp-tech/easyprotoc` (псевдоверсия main @ `f378f3f`)
- [ ] Потребители (`internal/core`, `internal/protoast`) импортируют публичные пакеты `protoc`/`protoc/syntax`/`parser` — вендор-пути исчезли
- [ ] Дерево `internal/easypc` удалено с `VENDORED.md` и cmd-бинарниками; seal-инфраструктура снята (`.golangci.yml`-исключение, параграф AGENTS.md)
- [ ] Полный паритет без регрессий: сборка, корпус lint (Gates A–D), D-07 position-тесты, `-race ./...`

### Out of Scope

- Взятие upstream-ветки `feat/easyp-integration` (generation sources/descriptor metadata fix) — unmerged, меняет семантику ошибок `Link`, нужен только codegen-серии; v1.1 пинится на main @ `f378f3f`
- Перевод breaking/generate/`mod tidy`/migrate на компилятор `easypc` — отдельные будущие фичи (серия продолжится после стабилизации lint-пути)
- Удаление `go-protoparser`/`protocompile` из `go.mod` — они нужны остальным подсистемам (REQ-5.2)
- Оптимизация скорости сверх «не медленнее main» — выигрыш лаборатории (20 мин → 0.08 с) — ориентир, не требование
- Изменение формата вывода issues и кодов выхода CLI — совместимость первична

## Context

- Эталонная реализация уже написана и проверена в приватном `easyp-tech/compiler-lab` (ветка `easyp-cli`): `internal/easypc` (13 пакетов), `internal/protoast` (~1000 строк), переработанный `internal/core/lint.go` (236 строк, параллельные окна по 256 файлов), 43 переписанных правила + 4 новых (`casing`, `field_not_required`, `package_no_import_cycle`, `syntax_specified`, `lint_helpers.go`).
- Ключевые отличия лаборатории от main: два прохода lint, `protoc.Link` с `WithSyntaxTrees(true)`, `ModuleRule` (`Collect`/`Finish`) для кросс-файловых правил, `core.Position = protoast.Position`, `ProtoInfo.File *protoast.File` + `ImportedFiles`.
- Приёмочный корпус: `api-common` (92 файла), `googleapis-nocompute` (7414 файлов) в `compiler-lab/bench/projects`; стенд lint-diff — `compiler-lab/bench/lab.py`.
- Открытые design-вопросы (решить в `/gsd-discuss-phase`): мост vs одно переключение lint-пути; паритетные тесты по батчам vs только корпусный lint-diff; батчи и порядок переноса 43 правил; полнота вендоринга `lint_helpers.go`/`casing.go`.
- Main-репозиторий на `5c9185f` на момент маппинга; поведение main, изменённое после снятия лабораторной копии, имеет приоритет над лабораторией (REQ-3.2).
- Технический долг вне milestone: см. `.planning/codebase/CONCERNS.md` (streaming-breaking, insertion point CRLF, `RunError.Unwrap`).

## Constraints

- **Tech stack**: Go 1.26.6, urfave/cli v2, testify; компилятор потребляется как внешний модуль `github.com/easyp-tech/easyprotoc` (v1.1; запрет новых зависимостей REQ-1.2 действовал только на вендоринг v1.0 и отменён — репозиторий компилятора публичен)
- **Compatibility**: формат вывода issues, пути и коды выхода CLI неизменны; buf-совместимое поведение — контракт
- **Scope**: breaking/generate/`mod tidy`/`migrate` не меняются в этом milestone (REQ-5.2)
- **Verification**: `task test` / `task lint` / `task schema:check`; паритет корпусов через `compiler-lab/bench/lab.py` — требует доступа к приватному репозиторию compiler-lab
- **Process**: репозиторий живёт по AGENTS.md — colocated тесты правил, регенерация схем через `task schema:generate`, без ручного редактирования `schemas/`

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Эталон — compiler-lab, а не написание с нуля | Реализация уже проверена: паритет замечаний и 0.08 с на `compute.proto` в лаборатории | — Pending |
| Компилятор вендорится как `internal/easypc`, без внешнего модуля | Нет новых внешних зависимостей и `replace` (REQ-1.2) | ✓ Phase 1 — ноль новых require/replace, go.sum нетронут |
| Требования перенесены из `.spec/features/lint-easypc-infra/requirements.md` без переосмысления | Документ прошёл interview и In Review | — Pending |
| Остальные подсистемы остаются на go-protoparser | Минимизация риска milestone; перевод — будущие фичи | — Pending |
| Точечные lint-исключения для дословно вендоренного кода (QF1001/casing.go, U1000/lint_helpers.go) вместо правки апстрима | D-01 байт-идентичность + зелёный lint-гейт несовместимы иначе; исправления уходят апстримом и приходят re-sync | ✓ Phase 1 — ратифицировано на UAT (2026-11-19) |
| Классификация отказов линковки — типизированный шов резолвера (не парсинг текста диагностик) + корпусный откат при отсутствующем импорте | `protoc.ErrFileNotFound`/`OpenImportFileError` — типизированный канал надёжнее строкового матчинга; откат соответствует букве REQ-2.3 | ✓ Phase 2 — dormant-классификатор, exit-коды 0/1/2 доказаны через Action-путь |
| Новые идентификаторы вместо моста: LintLinked/LinkedFileInfo/LintRule/linkedCommentSuppressions в новых файлах core | 10 коллизий имён со старым контрактом в пакете core; bridge — антипаттерн (research); консолидация имён — Phase 5 | ✓ Phase 3 — dormant-движок, 49+15 ячеек под -race |
| Детерминизм вывода — усиление, не паритет: тотальная сортировка (path, rule, line, col) на выходе движка | Порядок main не контракт (map iteration); паритет порядка ломает детерминизм без пользы | ✓ Phase 3 — repeated_runs_byte_identical (5 прогонов) |
| Coexistence Option A: suffixed <Rule>Linked типы + LintRuleName() override; старый путь байт-заморожен до Phase 5 | Один контракт не может реализовать оба Validate-сигнатуры; bridge — антипаттерн | ✓ Phase 4 — 41/41 портов, реестр parity-proven, 53 parity-ячейки |
| Переключение через identity-консолидацию (не правку шва): buildCore/lint_v1 текстуально неизменны, binding'и меняются через rename map | Откат = revert одного коммита (roadmap SC); минимальный-флип отвергнут как лишний шов | ✓ Phase 5 — flip 4238fbe атомарен, golden байт-идентичен, консолидация 263f067 |
| Новые правила = verbatim lab-порты + один атомарный registry/schema/MCP коммит | Правила, схемы и MCP меняются согласованно (SC-3); derivation loop закрыт (registry → AllLintUseValues → enums → MCP) | ✓ Phase 6 — реестр 44, teeth-parity 2/2/1/1/1/1, corpora leg → Phase 7 |
| Milestone acceptance = automated core (teeth Gates A-D ×2, golden 94, SC3 full -race battery) + corpora letters pending_user_unblock | Приватные корпуса недоступны (spike дважды); ready-to-fire скрипты + one-command path держат обязательство живым | ✓ Phase 7 — 12/12 (2 override), wazero race fixed (first full green -race: 41 pkgs/3718 tests) |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-11-21 — milestone v1.1 import-easyprotoc started*
