# Roadmap: EasyP

## Overview

Milestone «Миграция линтера на собственный компилятор» (`lint-easypc-infra`): перевод `easyp lint` с go-protoparser на вендорированный компилятор `easypc` и модель `protoast` (эталон — проверенная реализация в приватном `compiler-lab`, ветка `easyp-cli`). Путь: чистое добавление (вендоринг без новых зависимостей, расчленение lint/breaking модели) → линкер с трёхсторонней маршрутизацией ошибок → двухпроходный движок с паритетом подавлений → перенос всех 43 правил батчами с верификацией паритета на уровне батча → одно переключение lint-пути с воротами совместимости → три новых правила с атомарной регенерацией схем и MCP-описаний → приёмка на корпусах (паритет с main + производительность не хуже main). Неизменные инварианты всего пути: drop-in buf-совместимость (формат вывода issues, пути, коды выхода CLI), неприкосновенность breaking/generate/`mod tidy`/`migrate` (LINT-COMPAT-02) и батч-уровневая верификация паритета при переносе правил.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Вендоринг компилятора и расчленение модели** - easypc + protoast вкопированы без новых зависимостей; lint-путь отделён от модели breaking (completed 2026-11-19)
- [x] **Phase 2: Линкер и маршрутизация ошибок** - линковка корпусов с трёхсторонней классификацией: COMPILE / откат при отсутствующем импорте / ошибки чтения-синтаксиса (completed 2026-11-19)
- [x] **Phase 3: Двухпроходный движок и подавления** - параллельные окна, ModuleRule, детерминированный порядок, паритет ignore/комментариев/валидации конфига (completed 2026-11-19)
- [ ] **Phase 4: Перенос 43 правил с батч-паритетом** - все правила на protoast, семантика main побеждает лабораторию, паритет верифицируется на уровне батча
- [ ] **Phase 5: Переключение lint-пути и ворота совместимости** - easyp lint на компиляторе, старый контур удалён, формат/коды выхода и остальные подсистемы неизменны
- [ ] **Phase 6: Новые правила и регенерация схем/MCP** - PACKAGE_NO_IMPORT_CYCLE, SYNTAX_SPECIFIED, FIELD_NOT_REQUIRED + атомарное обновление схем и MCP-описаний
- [ ] **Phase 7: Приёмка — паритет корпусов и производительность** - идентичный main вывод на api-common и googleapis-nocompute, не медленнее main

## Phase Details

### Phase 1: Вендоринг компилятора и расчленение модели
**Goal**: Компилятор `easypc` и модель `protoast` живут в репозитории как чистое добавление без единой новой зависимости; lint-путь больше не делит модель чтения с breaking, и это видно как пустой дифф поведения
**Mode:** mvp
**Depends on**: Nothing (first phase)
**Requirements**: LINT-VEND-01, LINT-VEND-02
**Success Criteria** (what must be TRUE):
  1. В репозитории присутствуют `internal/easypc` (компилятор целиком, включая `cmd/` и `testdata/`) и `internal/protoast`, вкопированные из `compiler-lab/easyp-cli` с импортами на `github.com/easyp-tech/easyp/internal/...`; `go build ./...` собирается, собственные тесты вендоренных пакетов проходят
  2. `go mod tidy` не изменяет `go.mod` и `go.sum` — ни одного нового внешнего модуля, ни `replace`
  3. `task test` и `task lint` зелёные после вендоринга и расчленения общей модели (`Issue.Position` на `protoast.Position`, `protoInfoRead` в собственности breaking) — поведение breaking/generate/`mod tidy`/`migrate` для пользователей не меняется
  4. Фиксатура с мультибайтовым UTF-8 содержимым доказывает, что позиции protoast воспроизводят контракт вывода main (1-based строка, 1-based колонка в рунах) — до начала любого переноса правил
**Plans**: 4/4 plans complete
**Design note** (решить в /gsd-discuss-phase): полнота вендоринга `lint_helpers.go`/`casing.go` — целиком или только используемое линтером подмножество (LINT-VEND-01)
**Plans**: 4 plans
Plans:
**Wave 1**
- [x] 01-01-PLAN.md — Vendor drop (tracer): easypc+protoast copy-in from compiler-lab main@08cef10, zero-diff deps, VENDORED.md + lint exclusion + AGENTS.md seal, one atomic commit

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 01-02-PLAN.md — Position contract (TDD, D-07): rune columns in protoast build.go pos(), D-07(a) unit + D-07(b) golden parity tests on UTF-8+tab fixture, QF1002 fixed
- [x] 01-03-PLAN.md — Model split (D-06): Issue.Position → protoast.Position via single conversion boundary; lint-owned lintProtoInfoRead vs breaking-owned BreakingProtoInfo

**Wave 3** *(blocked on Wave 2 completion)*
- [x] 01-04-PLAN.md — Helper vendoring (D-01): casing.go/casing_test.go byte-verbatim, lint_helpers.go with 3 seam adaptations, D-02 fallback only on actual lint failure

### Phase 2: Линкер и маршрутизация ошибок
**Goal**: Линковка корпусов через `easypc` с корректной трёхсторонней классификацией отказов: COMPILE-замечания, откат на пофайловый разбор при отсутствующем импорте, ошибки чтения/синтаксиса как в main
**Mode:** mvp
**Depends on**: Phase 1
**Requirements**: LINT-ENG-02, LINT-ENG-03, LINT-ENG-04
**Success Criteria** (what must be TRUE):
  1. Корпус, где хотя бы один файл не проходит линковку из-за ошибок компиляции (не отсутствующий импорт), даёт по этим файлам замечания с именем правила `COMPILE` (позиция и сообщение компилятора) в существующем формате issues; правила на корпусе не запускаются
  2. Корпус с отсутствующим импортом откатывается на разбор каждого файла по отдельности и репортит ошибки чтения/синтаксиса как до изменения, без `COMPILE`-замечаний; типизированная ошибка отсутствующего импорта (`OpenImportFileError`) доходит до существующего пути exit-кода 2
  3. Файл с ошибкой чтения или синтаксиса возвращает ошибку выполнения команды для этого файла в точности как до изменения
  4. Фиксатурная матрица всех трёх классов ошибок (компиляция / отсутствующий импорт / чтение-синтаксис) проходит с кодами выхода CLI, проверенными через путь Action — до врезки движка в `easyp lint`
**Plans**: 4/4 plans complete
Plans:
- [x] 02-04-PLAN.md

**Wave 1**
- [x] 02-01-PLAN.md — Corpus linker tracer: typed resolver seam (D-01) + three-outcome classification (linked / COMPILE issues / fallback) in internal/core/link.go, COMPILE shape pinned at etalon numbers (LINT-ENG-02/03)

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 02-02-PLAN.md — Classification matrix completion: syntax-over-COMPILE precedence (LINT-ENG-04), missing-import variants (transitive/blocked/WKT/non-local), tab-UTF-8 LineCol pinning, testdata tree
- [x] 02-03-PLAN.md — Action-path matrix (D-05): in-process actionV1 error classes, subprocess exit codes 0/1/2, COMPILE printIssues golden; lint.go/api untouched (dormancy)
**Research flag**: классификация ошибок линковки — поверхность высочайшего риска (semantic cliff); резолвер идёт через `importRoots` + `importFileAllowed` + `wellknownimports`, без обхода

### Phase 3: Двухпроходный движок и подавления
**Goal**: Перезаписанный lint-движок на слинкованных деревьях — параллельные окна, ModuleRule для кросс-файловых правил, детерминированный порядок вывода, центральный фильтр подавлений и валидация конфигурации с паритетом main; движок готов и проверен, но ещё не врезан в CLI
**Mode:** mvp
**Depends on**: Phase 2
**Requirements**: LINT-ENG-01, LINT-ENG-05, LINT-ENG-06
**Success Criteria** (what must be TRUE):
  1. Движок строит модель каждого файла из слинкованных синтаксических деревьев `easypc` и выполняет применимые правила без изменения порядка обхода файлов: сборка результатов по индексу, вывод детерминирован между прогонами при параллельных окнах (прогон под `-race` зелёный)
  2. Подавления идентичны текущему main: `ignore`/`ignore_only` и `allowCommentIgnores` со всеми диалектами комментариев (`easyp:disable`/`nolint`/`buf:lint:ignore`) — перенесённый набор suppression-фиксатур проходит целиком
  3. Неизвестное имя правила или группы в `lint.use` возвращает ту же ошибку валидации конфигурации, что и текущий main
  4. Кросс-файловые правила через `ModuleRule` (Collect/Finish) корректно завершаются в параллельном движке; приёмники правил не пишут в свои поля между прогонами (аудит statelessness в составе race-теста движка)
**Plans**: 4/4 plans complete
Plans:
**Wave 1**
- [x] 03-01-PLAN.md — Two-pass linked-file engine tracer: etalon port with renamed identifiers (no bridge), model from Linked.Syntax, windows of 256, central suppressions, ModuleRule Collect/Finish, D-01 total sort, fallback read path + main-wins wrapper (LINT-ENG-01; D-01/D-02/D-03/D-05/D-07)

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 03-02-PLAN.md — Suppression parity table: main's fixture matrix (EOF, policy, dialects) re-driven through the new engine, config ignore/ignore_only, tab position pin (LINT-ENG-05; D-04/D-05)
- [x] 03-03-PLAN.md — Parallel-correctness matrix: statelessness/race audit, repeated-run determinism, walk-order error precedence, COMPILE-class routing through the engine (LINT-ENG-01)
- [x] 03-04-PLAN.md — lint.use validation parity guard through rules.New + phase dormancy/dependency/schema closure audit (LINT-ENG-06; D-06)

**Design notes** (решить в /gsd-discuss-phase): порядок эмиссии замечаний внутри правила (main недетерминирован из-за map iteration — детерминированная сортировка видима пользователю); решение «мост vs одно переключение» принимается совместно с Phase 5. Рекомендован `/gsd-plan-phase --research-phase`

### Phase 4: Перенос 43 правил с батч-паритетом
**Goal**: Все 43 текущих правила переписаны на protoast с семантикой текущего main (main побеждает лабораторную копию); каждый батч переноса верифицирован паритетом старое-vs-новое до перехода к следующему
**Mode:** mvp
**Depends on**: Phase 3
**Requirements**: LINT-PARITY-02
**Success Criteria** (what must be TRUE):
  1. Все 43 правила дают на общих фиксатурах идентичные старой реализации замечания (правило, файл, строка, колонка); перенос выполняется батчами, и каждый батч верифицирован по паритету до начала следующего — корпусный lint-diff на финише остаётся воротами, а не сессией отладки
  2. Правила, чьё поведение изменилось в main после снятия лабораторной копии (например, подавления через комментарии), следуют семантике текущего main, а не лаборатории; colocated-тесты переносятся из main
  3. У каждого перенесённого правила есть colocated `<rule>_test.go`; `task test` и `task lint` зелёные на всех промежуточных состояниях фазы
**Plans**: TBD (ожидаемо по плану на батч правил)
**Design notes** (решить в /gsd-discuss-phase): батчи и порядок переноса 43 правил (естественные кандидаты — группы конфига MINIMAL → BASIC → DEFAULT → COMMENTS → UNARY_RPC); механизм батч-верификации — per-batch паритетные тесты против старой реализации или корпусный lint-diff

### Phase 5: Переключение lint-пути и ворота совместимости
**Goal**: `easyp lint` работает на компиляторе; старый lint-контур удалён; формат вывода, пути, коды выхода и все остальные подсистемы не изменились — единственная точка пользовательски видимого изменения, откат = revert
**Mode:** mvp
**Depends on**: Phase 4
**Requirements**: LINT-COMPAT-01, LINT-COMPAT-02
**Success Criteria** (what must be TRUE):
  1. `easyp lint` выдаёт замечания в неизменном формате: пути, формат issues и коды выхода CLI совпадают с main до переключения на фиксатурном корпусе, включая новые `COMPILE`-замечания в существующем формате
  2. Старый lint-контур и старые реализации правил удалены из кодовой базы; `easyp lint` обслуживается только движком на компиляторе (врезка — одно изменение в шве `buildCore`)
  3. breaking, generate, `mod tidy` и `migrate` проходят свои тестовые наборы в том же прогоне CI — поведение как до изменения; `go-protoparser` и `protocompile` остаются в `go.mod`
**Plans**: TBD
**Design note**: здесь фиксируется решение «мост vs одно переключение», принятое совместно с Phase 3 (LINT-ENG-01, LINT-PARITY-01)

### Phase 6: Новые правила и регенерация схем/MCP
**Goal**: Три новых правила работают на готовом движке — `PACKAGE_NO_IMPORT_CYCLE` в MINIMAL, `SYNTAX_SPECIFIED` и `FIELD_NOT_REQUIRED` в BASIC — с выводом, идентичным лабораторному CLI, и атомарно обновлёнными схемами и MCP-описаниями
**Mode:** mvp
**Depends on**: Phase 5
**Requirements**: LINT-NEWRULES-01, LINT-NEWRULES-02, LINT-NEWRULES-03, LINT-NEWRULES-04
**Success Criteria** (what must be TRUE):
  1. Группа `MINIMAL` включает `PACKAGE_NO_IMPORT_CYCLE` (кросс-файловое, на ModuleRule); группа `BASIC` включает `SYNTAX_SPECIFIED` и `FIELD_NOT_REQUIRED` (размещение как в лабораторном CLI); спец-отклонение `PACKAGE_NO_IMPORT_CYCLE` в валидации имён правил убрано
  2. Замечания новых правил на приёмочных корпусах идентичны выводу лабораторного CLI на тех же корпусах
  3. Сгенерированные JSON-схемы включают имена новых правил и групп в enum значений `lint.use`; `task schema:check` без дрейфа; MCP-описания `mcp/easypconfig` отражают обновлённый набор — правила, схемы и MCP-описания меняются согласованно, одним изменением
**Plans**: TBD
**Design note** (решить в /gsd-discuss-phase): механизм проверки паритета новых правил против лабораторного CLI (LINT-NEWRULES-03)

### Phase 7: Приёмка — паритет корпусов и производительность
**Goal**: Финальные ворота milestone: идентичность замечаний текущему main на приёмочных корпусах и производительность не хуже main на большом корпусе
**Mode:** mvp
**Depends on**: Phase 6
**Requirements**: LINT-PARITY-01, LINT-PERF-01
**Success Criteria** (what must be TRUE):
  1. lint-diff стенд показывает идентичные множества замечаний с текущим main (правило, файл, строка, колонка) для всех 43 правил на корпусах `api-common` (92 файла) и `googleapis-nocompute` (7414 файлов); прогон повторён дважды для исключения недетерминизма; база main-vs-main снята
  2. `easyp lint` на `googleapis-nocompute` завершается не медленнее текущего main при флагах, эквивалентных CI; результат измерения зафиксирован
  3. Ворота milestone зелёные на финальном состоянии: `task test`, `task lint`, `task schema:check`
**Plans**: TBD
**Research flag**: воспроизведение стенда `lab.py` (доступ к приватному compiler-lab, пересборка корпусов) — валидировать спайком до фазы или в её начале

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Вендоринг компилятора и расчленение модели | 4/4 | Complete    | 2026-11-19 |
| 2. Линкер и маршрутизация ошибок | 4/4 | Complete    | 2026-11-19 |
| 3. Двухпроходный движок и подавления | 4/4 | Complete    | 2026-11-19 |
| 4. Перенос 43 правил с батч-паритетом | 0/TBD | Not started | - |
| 5. Переключение lint-пути и ворота совместимости | 0/TBD | Not started | - |
| 6. Новые правила и регенерация схем/MCP | 0/TBD | Not started | - |
| 7. Приёмка — паритет корпусов и производительность | 0/TBD | Not started | - |

---
*Roadmap created: 2026-11-17 (milestone lint-easypc-infra, 17/17 requirements mapped)*

## Backlog

### Phase 999.1: Follow-up — Phase 03 deferred UAT follow-up: Test 1 (BACKLOG)

**Goal:** Resolve the UAT checkpoint deferred during Phase 03 verification
**Source phase:** 03
**Deferred at:** 2026-11-19 during /gsd-verify-work 03 session completion
**Follow-ups:**
- [ ] Test 1: Add a disk-instrumentation cell in Phase 4 asserting readSource is invoked exactly once per corpus file on a compiling corpus (behavioral proof for the no-disk-reread guarantee, currently structural) (deferred 2026-11-19)

### Phase 999.2: Follow-up — Phase 03 deferred UAT follow-up: Test 2 (BACKLOG)

**Goal:** Resolve the UAT checkpoint deferred during Phase 03 verification
**Source phase:** 03
**Deferred at:** 2026-11-19 during /gsd-verify-work 03 session completion
**Follow-ups:**
- [ ] Test 2: Standing constraint for Phase 4: at each rule-port batch confirm ported rules never consult comment directives or ignoreOnly themselves; keep central-filter cells (comment_suppression_central, finish_comment_refilter, compile_bypasses_suppression) green (deferred 2026-11-19)
