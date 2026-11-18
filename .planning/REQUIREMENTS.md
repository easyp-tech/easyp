# Requirements: EasyP — lint-easypc-infra

**Defined:** 2026-11-17
**Core Value:** EasyP остаётся drop-in buf-совместимым инструментом: формат вывода, коды выхода и конфигурация не ломаются между версиями — миграции внутренностей не должны быть заметны пользователю.

Источник: `.spec/features/lint-easypc-infra/requirements.md` (Status: In Review, interview 2026-11-17). Требования перенесены без переосмысления; нумерация REQ-x.y исходного документа сохранена в тексте для трассируемости.

## v1 Requirements

Requirements for milestone «Миграция линтера на собственный компилятор». Topological order: VEND → ENG → PARITY → NEWRULES → COMPAT/PERF.

### Vendoring

- [x] **LINT-VEND-01** (REQ-1.1): WHEN исходники фичи собраны, the system SHALL содержать `easypc` (компилятор целиком, включая `cmd/` и `testdata/`) и `protoast`, вкопированные из `compiler-lab/easyp-cli` с импортами на `github.com/easyp-tech/easyp/internal/...`
- [x] **LINT-VEND-02** (REQ-1.2): WHEN выполнен `go mod tidy`, the system SHALL не содержать новых внешних зависимостей для компилятора (без внешнего модуля и без `replace`)

### Engine

- [ ] **LINT-ENG-01** (REQ-2.1): WHEN `easyp lint` запущен на компилирующемся корпусе, the system SHALL построить модель каждого файла из слинкованных синтаксических деревьев `easypc` и выполнить все применимые правила без изменения порядка обхода файлов
- [ ] **LINT-ENG-02** (REQ-2.2): WHEN хотя бы один файл корпуса не проходит линковку (ошибки компиляции, не связанные с отсутствующим импортом), the system SHALL выдать для этих файлов замечания с именем правила `COMPILE` (позиция и сообщение компилятора) и не запускать правила на корпусе
- [ ] **LINT-ENG-03** (REQ-2.3): WHEN линковка невозможна из-за отсутствующего импорта, the system SHALL вернуться к разбору каждого файла по отдельности и репортить ошибку чтения/синтаксиса как до этого изменения (без COMPILE issues)
- [ ] **LINT-ENG-04** (REQ-2.4): WHEN файл содержит ошибки чтения или синтаксиса, the system SHALL вернуть ошибку выполнения команды для этого файла как до изменения
- [ ] **LINT-ENG-05** (REQ-2.5): WHEN правило включено и файл попадает под `ignore`/`ignore_only` или подавление комментариями (`allowCommentIgnores`), the system SHALL применять те же правила подавления, что и текущий main
- [ ] **LINT-ENG-06** (REQ-2.6): WHEN конфигурация lint содержит неизвестное имя правила или группы в `lint.use`, the system SHALL вернуть ошибку валидации как до изменения

### Parity

- [ ] **LINT-PARITY-01** (REQ-3.1): WHEN `easyp lint` (новая реализация) и текущий main запущены на корпусе `api-common` и `googleapis-nocompute` (все файлы компилируются), the system SHALL выдать идентичное множество замечаний для всех 43 текущих правил (правило, файл, строка, колонка)
- [ ] **LINT-PARITY-02** (REQ-3.2): WHEN правило текущего main имеет поведение, изменённое в main после снятия лабораторной копии (например, подавления через комментарии), the system SHALL сохранить семантику текущего main, а не лабораторной копии

### New Rules

- [ ] **LINT-NEWRULES-01** (REQ-4.1): WHEN lint выполняется с группой `MINIMAL`, the system SHALL включать правило `PACKAGE_NO_IMPORT_CYCLE` (размещение как в лабораторном CLI)
- [ ] **LINT-NEWRULES-02** (REQ-4.2): WHEN lint выполняется с группой `BASIC`, the system SHALL включать правила `SYNTAX_SPECIFIED` и `FIELD_NOT_REQUIRED`
- [ ] **LINT-NEWRULES-03** (REQ-4.3): WHEN новые правила запущены на корпусах, the system SHALL выдавать замечания, идентичные выводу лабораторного CLI на тех же корпусах
- [ ] **LINT-NEWRULES-04** (REQ-4.4): WHEN генерируются JSON-схемы конфигурации, the system SHALL включать имена новых правил и групп в enum значений `lint.use` и отражать обновлённый набор в MCP-описаниях `mcp/easypconfig` (`task schema:generate`)

### Compatibility

- [ ] **LINT-COMPAT-01** (REQ-5.1): WHEN `easyp lint` выдаёт замечания, the system SHALL сохранять текущий формат вывода issues, пути и коды выхода CLI без изменений (включая появление новых `COMPILE` в существующем формате)
- [ ] **LINT-COMPAT-02** (REQ-5.2): WHEN выполнены команды breaking-проверки, generate, `mod tidy` и `migrate`, the system SHALL работать как до изменения (старая модель go-protoparser для breaking сохраняется; `go-protoparser` и `protocompile` остаются в `go.mod`)

### Performance

- [ ] **LINT-PERF-01** (REQ-6.1): WHEN `easyp lint` запущен на корпусе `googleapis-nocompute`, the system SHALL завершаться не медленнее текущего main (потенциальный выигрыш — как в лаборатории — не является требованием, только ориентиром)

## v2 Requirements

Deferred to future milestones (серия продолжится после стабилизации lint-пути):

### Compiler Expansion

- **FUTURE-COMPILER-01**: Перевод breaking-check на `easypc` (с удалением go-protoparser из go.mod после обеих миграций)
- **FUTURE-COMPILER-02**: Перевод generate на `easypc`
- **FUTURE-COMPILER-03**: Перевод `mod tidy` и `migrate` на `easypc`
- **FUTURE-COMPILER-04**: Поддержка editions (2023) в приёмочных корпусах — статус easypc не верифицирован
- **FUTURE-HARDENING-01**: Go native fuzzing (`testing.F`) лексера/парсера easypc — пост-паритетное укрепление

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Перевод breaking/generate/`mod tidy`/migrate на компилятор | Отдельные будущие фичи; в этом milestone subsystems не затрагиваются (REQ-5.2) |
| Удаление `go-protoparser`/`protocompile` из `go.mod` | Нужны остальным подсистемам в этом milestone (REQ-5.2) |
| Оптимизация скорости сверх «не медленнее main» | Выигрыш лаборатории (20 мин → 0.08 с) — ориентир, не требование (REQ-6.1) |
| Изменение формата вывода issues и кодов выхода CLI | Совместимость первична (REQ-5.1); «buf-совместимость» означает семантику диагностики, не значения exit codes (у buf 100) |
| Auto-fix, плагины правил, персистентный кэш lint | Anti-features по research (FEATURES.md) |
| Новые выходные форматы lint | Не требуется совместимостью; отдельная будущая фича |

## Open Design Questions

Перенесены из исходного документа — решаются в `/gsd-discuss-phase` (затрагивают структуру фаз, не состав требований):

| Вопрос | Затронутые требования |
|--------|----------------------|
| Переключать lint-путь одним изменением (модель+движок+правила) или через временный мост старой/новой модели | LINT-ENG-01, LINT-PARITY-01 |
| Паритетные тесты на каждый батч правил против старой реализации или только корпусный lint-diff на финише | LINT-PARITY-01, LINT-NEWRULES-03 |
| В каких батчах и в каком порядке переносить 43 правила | LINT-PARITY-01 |
| Переносить ли лабораторные `lint_helpers.go`/`casing.go` целиком или только используемое линтером подмножество | LINT-VEND-01 |
| Порядок эмиссии замечаний внутри правила в новом движке (main недетерминирован из-за map iteration) | LINT-ENG-01 |

## Verification Commands

| Действие | Команда | Источник |
|----------|---------|----------|
| Test | `task test` | Taskfile.yml |
| Test (точечно) | `go test -race -count=1 ./internal/...` | AGENTS.md |
| Build | `task build` | Taskfile.yml |
| Lint | `task lint` | Taskfile.yml |
| Schema check | `task schema:check` | Taskfile.yml |
| Паритет корпусов | прогон lint-diff стенда `compiler-lab/bench/lab.py` (api-common, googleapis-nocompute) против main и лабораторного CLI | compiler-lab/bench |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| LINT-VEND-01 | Phase 1 | Complete |
| LINT-VEND-02 | Phase 1 | Complete |
| LINT-ENG-01 | Phase 3 | Pending |
| LINT-ENG-02 | Phase 2 | Pending |
| LINT-ENG-03 | Phase 2 | Pending |
| LINT-ENG-04 | Phase 2 | Pending |
| LINT-ENG-05 | Phase 3 | Pending |
| LINT-ENG-06 | Phase 3 | Pending |
| LINT-PARITY-01 | Phase 7 | Pending |
| LINT-PARITY-02 | Phase 4 | Pending |
| LINT-NEWRULES-01 | Phase 6 | Pending |
| LINT-NEWRULES-02 | Phase 6 | Pending |
| LINT-NEWRULES-03 | Phase 6 | Pending |
| LINT-NEWRULES-04 | Phase 6 | Pending |
| LINT-COMPAT-01 | Phase 5 | Pending |
| LINT-COMPAT-02 | Phase 5 | Pending |
| LINT-PERF-01 | Phase 7 | Pending |

**Coverage:**
- v1 requirements: 17 total
- Mapped to phases: 17
- Unmapped: 0 ✓

---
*Requirements defined: 2026-11-17*
*Last updated: 2026-11-17 after initial definition (перенос из .spec/features/lint-easypc-infra/requirements.md)*
