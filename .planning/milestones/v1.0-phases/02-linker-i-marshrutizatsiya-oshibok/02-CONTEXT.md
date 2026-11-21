# Phase 2: Линкер и маршрутизация ошибок - Context

**Gathered:** 2026-11-19
**Status:** Ready for planning

<domain>
## Phase Boundary

Линковка корпусов через `easypc` (`protoc.Link` + `WithSyntaxTrees`) с корректной трёхсторонней классификацией отказов, работающей на уровне `internal/core`: (1) ошибки компиляции → замечания с именем правила `COMPILE` (позиция и сообщение компилятора) в существующем формате issues, правила на корпусе не запускаются; (2) отсутствующий импорт → корпусный откат на пофайловый разбор по существующему пути, типизированная ошибка `OpenImportFileError` доходит до существующего exit-кода 2; (3) ошибки чтения/синтаксиса → ошибка выполнения команды для файла как в main. Классификация готова и проверена фиксатурной матрицей с кодами выхода через путь Action, но НЕ врезана в пользовательский `easyp lint` (переключение — Phase 5).

Не входит: движок правил на деревьях, подавления, детерминизм вывода (Phase 3); перенос правил (Phase 4); переключение lint-пути (Phase 5).

</domain>

<decisions>
## Implementation Decisions

### Классификация «отсутствующий импорт»
- **D-01:** Детекция отсутствующего импорта — через типизированный шов резолвера, а не по тексту диагностик. Резолвер линкера строится из import roots команды (как в main, `internal/api/roots.go`) + well-known imports (`internal/easypc/protoc/wkt`); «файл не найден» резолвером (`protoc.ErrFileNotFound`) преобразуется в собственную типизированную ошибку линкера, несущую путь импорта. Диагностический текст компилятора (`Import "..." was not found or had errors.`) не парсится. [auto] Selected: «typed resolver seam» (recommended default) — надёжнее строкового матчинга и проверяется юнит-тестами.

### Гранулярность отката
- **D-02:** Откат КОРПУСНЫЙ: при отсутствии хотя бы одного импорта весь корпус пере-разбирается пофайлово существующим путём (`lintProtoInfoRead`), без `COMPILE`-замечаний; `OpenImportFileError` сохраняет свой маршрут до exit-кода 2 в `internal/api/lint.go`. Соответствует букве LINT-ENG-03 («вернуться к разбору каждого файла по отдельности») и лабораторному поведению.

### Форма COMPILE-замечаний
- **D-03:** По одной `IssueInfo` на диагностику компилятора: имя правила `COMPILE`, путь — display-path файла (как показывает компилятор), строка/колонка из `Diagnostic` (позиции компилятора уже 1-based — контракт позиций доказан в Phase 1, D-07), текст — сообщение компилятора без префиксов. Порядок выдачи — детерминированный (сортировка по файлу/позиции на стороне классификатора; main на этом классе ошибок выдаёт замечания компилятора в его порядке — паритет порядка проверяется на фиксатурах).

### Размещение и форма классификатора
- **D-04:** Классификатор живёт в `internal/core` (правки в `internal/easypc` запрещены — D-04 фазы 1). Новый файл рядом с lint-путём (имя — discretion planner'а), публичная функция/тип, которую следующий за линкером код (движок Phase 3, переключение Phase 5) будет вызывать вместо текущего пофайлового чтения. Текущий `Lint()` не меняется — до Phase 5 линкер dormant для пользователей.

### Верификация через путь Action
- **D-05:** Фиксатурная матрица трёх классов ошибок (компиляция / отсутствующий импорт / чтение-синтаксис) проверяет коды выхода CLI через путь Action — тесты вызывают тот же код маршрутизации ошибок, который вызывается в `internal/api/lint.go`, без изменения видимого пользователю поведения `easyp lint`. Матрица зелёна ДО врезки движка.

### Claude's Discretion
- Имена файлов/типов классификатора и точная сигнатура API — по фактическому коду.
- Структура и размещение фиксатур (testdata-дерево трёх классов ошибок).
- Механика связки «резолвер → import roots» (переиспользование подготовленных roots vs собственная сборка).
- Нужна ли отдельная утилита сопоставления diagnostic→issue или она встраивается в классификатор.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Спецификация фичи
- `.spec/features/lint-easypc-infra/requirements.md` — REQ-2.2 (COMPILE-замечания), REQ-2.3 (откат при отсутствующем импорте), REQ-2.4 (ошибки чтения/синтаксиса как в main); глоссарий
- `.planning/REQUIREMENTS.md` — LINT-ENG-02/03/04 + трассировка (все три — Phase 2)

### Правила репозитория
- `AGENTS.md` — конвенции; запрет правок в `internal/easypc`
- `.spec/agent-rules.md` — обязательные стандарты кодирования
- `Taskfile.yml` — таргеты test/lint/build

### Roadmap
- `.planning/ROADMAP.md` § Phase 2 — цель, success criteria, research flag: «классификация ошибок линковки — поверхность высочайшего риска (semantic cliff); резолвер идёт через `importRoots` + `importFileAllowed` + `wellknownimports`, без обхода»

### Prior phases
- `.planning/phases/01-vendoring-kompilyatora-i-raschlenenie-modeli/01-CONTEXT.md` — D-04 (seal `internal/easypc`), D-06 (расщепление модели: у lint собственные типы), D-07 (контракт позиций 1-based доказан)

### Research
- `.planning/research/PITFALLS.md` — семантический обрыв классификации ошибок линковки; позиционная точность
- `.planning/research/ARCHITECTURE.md` — направление зависимостей, шов `buildCore`

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/easypc/protoc/link.go` — `protoc.Link(ctx, files, opts...)` с `WithSyntaxTrees`, `WithErrorHandler`; отказ — типизированный `ErrCompilationFailed` (link.go:100)
- `internal/easypc/protoc/resolver.go` — `Resolver`-контракт, `ErrFileNotFound`, `DirResolver`/`MemResolver`/`MultiResolver` — шов для типизированной детекции отсутствующего импорта
- `internal/easypc/protoc/diagnostic.go` — `fromInternalDiag` → `Diagnostic{Path, Line, Col, Message, IsWarn, Semantic}` (1-based позиции)
- `internal/easypc/protoc/wkt` — well-known imports для резолвера
- `internal/core/proto_info_read.go` + `OpenImportFileError` (`internal/core/dom.go`) — существующий пофайловый путь и типизированная ошибка отсутствующего импорта — цель отката
- `internal/api/lint.go` — существующая обработка ошибок и exit-кодов lint-команды (маршрут exit-кода 2)

### Established Patterns
- Компиляторные диагностики форматируются `path:line:col: message` (`Diagnostic.String()`)
- `core.New` получает полные опции без post-construction сеттеров; import roots готовятся в `internal/api/roots.go`
- Colocated тесты `<file>_test.go` (testify); тестовые дубли hand-written

### Integration Points
- `internal/core/lint.go` — текущий walker + `lintProtoInfoRead`; классификатор встаёт рядом, не меняя его до Phase 5
- `internal/api/lint.go` — маршрут exit-кодов: `OpenImportFileError` → 2; матрица D-05 тестирует этот путь
- `internal/api/runtime.go` (`buildCore`) — точка конструирования движка (Phase 3+)

</code_context>

<specifics>
## Specific Ideas

- Research flag из ROADMAP: поверхность высочайшего риска — semantic cliff классификации отказов линковки; резолвер через `importRoots` + `importFileAllowed` + `wellknownimports`, без обхода файловой системы
- Поведение main выигрывает у лаборатории (дух REQ-3.2): ошибки чтения/синтаксиса и маршрут exit-кодов — точно как в main
- MVP-режим фазы (ROADMAP `Mode: mvp`) — цель не User Story; планам разрешено не изобретать story-фрейминг (прецедент Phase 1)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 2-Линкер и маршрутизация ошибок*
*Context gathered: 2026-11-19*
