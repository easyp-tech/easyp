# EasyP

## What This Is

EasyP — CLI-тулкит для Protocol Buffers на Go: линтинг (buf-совместимые правила), детекция breaking changes, генерация кода (локальные/remote/WASM-плагины) и Git-центричное управление зависимостями (`easyp mod`). Пользователи — разработчики proto-API и команды, держащие большие корпуса proto-файлов в CI. Этот GSD-проект — рамка для развития EasyP в целом; milestones добавляются по мере подготовки фич в `.spec/features/`.

## Core Value

EasyP остаётся drop-in buf-совместимым инструментом: формат вывода, коды выхода и конфигурация не ломаются между версиями — миграции внутренностей не должны быть заметны пользователю.

## Current Milestone

**lint-easypc-infra** — перевод `easyp lint` с go-protoparser на собственный компилятор `easyp-protoc` (вендорится как `internal/easypc`) и модель `protoast`. Эталон — проверенная реализация в приватном `easyp-tech/compiler-lab` (ветка `easyp-cli`). Цели: производительность не хуже main (в лаборатории 20 мин → 0.08 с на `compute.proto`), buf-совместимые `COMPILE`-замечания на некомпилирующихся файлах, паритет всех 43 правил, три новых правила. Остальные подсистемы (breaking, generate, `mod tidy`, `migrate`) не затрагиваются. Требования перенесены в REQUIREMENTS.md из `.spec/features/lint-easypc-infra/requirements.md`.

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

### Active

Milestone «Миграция линтера на собственный компилятор» — полный список в REQUIREMENTS.md (LINT-*):

- [ ] Вендоринг компилятора `easypc` и модели `protoast` из compiler-lab (без внешних зависимостей)
- [ ] Новый lint-движок на слинкованных деревьях `easypc` с сохранением подавлений/`ignore`/валидации конфига
- [ ] Паритет всех 43 текущих правил на корпусах `api-common` и `googleapis-nocompute`
- [ ] Новые правила `PACKAGE_NO_IMPORT_CYCLE`, `SYNTAX_SPECIFIED`, `FIELD_NOT_REQUIRED` + обновление схем и MCP-описаний
- [ ] Совместимость вывода/кодов выхода и неизменность breaking/generate/mod tidy/migrate
- [ ] Производительность lint не хуже main на `googleapis-nocompute`

### Out of Scope

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

- **Tech stack**: Go 1.26.6, urfave/cli v2, testify; новые зависимости для компилятора запрещены (REQ-1.2) — вендоринг в `internal/`
- **Compatibility**: формат вывода issues, пути и коды выхода CLI неизменны; buf-совместимое поведение — контракт
- **Scope**: breaking/generate/`mod tidy`/`migrate` не меняются в этом milestone (REQ-5.2)
- **Verification**: `task test` / `task lint` / `task schema:check`; паритет корпусов через `compiler-lab/bench/lab.py` — требует доступа к приватному репозиторию compiler-lab
- **Process**: репозиторий живёт по AGENTS.md — colocated тесты правил, регенерация схем через `task schema:generate`, без ручного редактирования `schemas/`

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Эталон — compiler-lab, а не написание с нуля | Реализация уже проверена: паритет замечаний и 0.08 с на `compute.proto` в лаборатории | — Pending |
| Компилятор вендорится как `internal/easypc`, без внешнего модуля | Нет новых внешних зависимостей и `replace` (REQ-1.2) | — Pending |
| Требования перенесены из `.spec/features/lint-easypc-infra/requirements.md` без переосмысления | Документ прошёл interview и In Review | — Pending |
| Остальные подсистемы остаются на go-protoparser | Минимизация риска milestone; перевод — будущие фичи | — Pending |

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
*Last updated: 2026-11-17 after initialization*
