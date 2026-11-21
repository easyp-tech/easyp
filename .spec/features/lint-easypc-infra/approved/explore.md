# Exploration: lint-easypc-infra

Дата: 2026-11-16. Ветка: `feature/lint-easypc-infra`.

## Intent

Перевести `easyp lint` с go-protoparser на собственный компилятор `easyp-protoc`
(парсер + модель `protoast`), опираясь на уже сделанную и проверенную в приватном
репозитории `easyp-tech/compiler-lab` работу. Цели: скорость (в лаборатории
`compute.proto` 4 МБ — 20 мин → 0.08 с при тех же замечаниях), единый парсер с
будущим переходом остальных подсистем (generate/breaking/tidy/migrate) на
компилятор, уход от зависимости `yoheimuta/go-protoparser` в lint-пути.

Это первая фича из серии; компиляция generate/breaking — отдельные будущие фичи.

## Investigation

Источники: main-репозиторий (5c9185f) и `compiler-lab/easyp-cli` (лабораторная
копия CLI, клон через gh; компилятор уже вкопирован туда как `internal/easypc`).

### Текущее состояние main

- Движок: `internal/core/lint.go` (124 строки) — последовательный `WalkDir`,
  для каждого файла `protoInfoRead` → `commentSuppressions` → все правила по
  очереди. Модель: `core.ProtoInfo.Info *unordered.Proto` (go-protoparser),
  см. `internal/core/dom.go:48`.
- Правила: 43 файла в `internal/rules` + colocated-тесты; каждый обходить AST
  go-protoparser напрямую (например `protoInfo.Info.ProtoBody.Messages`).
- `Issue.Position` — `go-protoparser/parser/meta.Position`.

### Что сделано в compiler-lab (проверено клоном)

- `internal/easypc/` — компилятор целиком: 48 .go-файлов (33 без тестов),
  13 пакетов (`pkg/`: lexer, parser, linker, sourceinfo, serialize…;
  `protoc/`: API + WKT; `cmd/easyp-protoc`; testdata). Импорты уже переписаны
  на `github.com/easyp-tech/easyp/internal/easypc/...`, внешнего модуля нет.
- `internal/protoast/` — модель для линтера (~1000 строк: `build.go` 462,
  `protoast.go` 220, `walk.go` 82 с Visitor-паттерном, `comments.go`, `outline.go`).
  Зависит только от `easypc/pkg/parser` и `easypc/protoc/syntax`.
- `internal/core/lint.go` (236 строк) — переработан: два прохода, параллельные
  окна по 256 файлов (`internal/parallel`), линковка через `protoc.Link` с
  `WithSyntaxTrees(true)`, при ошибке компиляции — issues `COMPILE` как у buf
  (`internal/core/link.go:25`), кэш импортов, новый интерфейс `ModuleRule`
  (`Collect`/`Finish`) для кросс-файловых правил.
- `core.Position = protoast.Position`; `ProtoInfo.File *protoast.File` +
  `ImportedFiles map[ImportPath]*protoast.File`.
- Все 43 правила main существуют в лаборатории в переписанном под protoast
  виде, плюс 4 новых: `casing`, `field_not_required`, `package_no_import_cycle`,
  `syntax_specified` (+ `lint_helpers.go`).

### Ключевая находка — связка с breaking

`core.ProtoInfo` (go-protoparser модель) используется не только линтером:
`internal/core/breaking_check.go:104 readProtoFiles` → `collect` (`dom.go:134`)
строит из `unordered.Proto` коллекции для breaking. В лаборатории эта проблема
решена тем, что breaking заодно переведён на компилятор (`compilerbackend`) —
но это вне scope нашей фичи. Значит, при lint-only миграции старая модель
должна остаться живой для breaking: lint строит новую модель, breaking
продолжает читать go-protoparser. `go-protoparser` из `go.mod` в этой фиче
не удаляется.

### Расхождение наборов правил

Лабораторные правила переписаны с более старой базы main. Main с октября
эволюционировал (например `check_lint_ignore_test.go`, подавления через
комментарии `allowCommentIgnores` и т.п.). Поэтому файлы правил из лаборатории
нельзя копировать 1:1 — каждое правило переносится на модель protoast с
проверкой семантики против текущего main.

### Тесты и приёмка

- Правила тестируются testify-даблами в colocated `<file>_test.go`; тесты тоже
  переписываются под protoast.
- В лаборатории есть приёмочный стенд: `bench/lab.py` (`lint`/`lint-diff` —
  снимки линтера и сравнение), корпуса `api-common` (92 файла), `googleapis`
  (7416), пересборка `bench/setup.sh` из googleapis. Критерий в лаборатории —
  те же замечания на корпусе.

## Build Tooling

- **Orchestrator:** task (Taskfile.yml)
- **Test:** `task test` (gotestsum, -race, coverage); точечно `go test -race -count=1 ./internal/...`
- **Build:** `task build` (`go build -o easyp ./cmd/easyp`)
- **Lint:** `task lint` (golangci-lint + hadolint)
- **Generate:** `task schema:generate` (не затрагивается); сheck: `task schema:check`
- **Source:** Taskfile.yml, AGENTS.md, .spec/TOOLS.md

## Options Considered

### Option A: Перенос лабораторного lint-стека целиком (рекомендуется)
Вкопировать `internal/easypc` + `internal/protoast` как в лаборатории, перенести
движок (`lint.go`, `link.go`, окна, ModuleRule), затем переносить правила по
батчам с reconciling против текущей семантики main.
- Плюсы: максимальное переиспользование проверенного кода; buf-совместимое
  поведение (COMPILE issues); производительность лаборатории; путь к полной
  миграции на компилятор уже проложен.
- Минусы: большой дифф; правила нельзя копировать 1:1 (расхождение с main);
  требуется вкопировать компилятор, пока он нужен lint'у лишь частично.
- Сложность: высокая, но делится на батчи.

### Option B: Минимальный сабсет — только парсер, без линковки
Взять из `easypc` только `pkg/parser` + `protoc/syntax`, добавить protoast,
движок не трогать (последовательный), правила перевести.
- Плюсы: меньше дифф; нет зависимости от linker/protoc API.
- Минусы: выбрасывает валидированную в лаборатории линковку и buf-поведение
  COMPILE; часть выигрыша в скорости теряется на импортах; всё равно придётся
  потом делать работу Option A при миграции остальных подсистем.
- Сложность: средняя.

### Option C: Заодно перевести breaking/generate (полная миграция лаборатории)
- Плюсы: сразу решает связку ProtoInfo↔breaking, убирает дублирование моделей.
- Минусы: scope уходит далеко за «перевести линтер»; дифф неприемлемо велик
  для одного ревью.
- Сложность: очень высокая. Отложено в отдельные фичи серии.

## Constraints & Risks

- **Совместимость**: пользовательский вывод lint должен совпадать с текущим
  на корпусах (позиции строк/колонок, сообщения). Новое поведение — issues
  `COMPILE` на некомпилирующихся файлах (как buf) — видимое изменение, нужно
  явное подтверждение.
- **Связка с breaking**: старая go-protoparser модель остаётся для breaking;
  временное дублирование моделей неизбежно до миграции breaking.
- **Расхождение правил main vs лаборатория**: риск тихой потери семантики при
  копировании; нужен паритетный прогон (lint-diff) как приёмочный критерий.
- **Вендоринг**: `easypc` — 3.4 МБ вместе с testdata; соответствует подходу
  лаборатории (без внешнего модуля и replace), но объём коммита нужно учесть.
- **Зависимости**: `go-protoparser` остаётся в go.mod (breaking и др.);
  `protocompile` остаётся (mod tidy импорты). В go.mod ничего не добавляется
  извне.

## Recommended Direction

Option A: перенос лабораторного lint-стека, разбитый внутри фичи на этапы —
(1) вендоринг `easypc` + `protoast`, (2) движок lint без смены правил
(мост старой модели?) — нет: движок и модель меняются вместе, поэтому
(2) движок + модель + правила единым переключением lint-пути, но перенос
правил по батчам с паритетными тестами против старой реализации до
финального переключения.

## Scope Boundaries

- **Must-have (v1):**
  - `internal/easypc` + `internal/protoast` вкопированы из лаборатории
    (компилятор целиком, включая `cmd/` и `testdata/`, как в лаборатории —
    решено пользователем 2026-11-17);
  - lint-путь (движок + все 43 правила main) работает на protoast;
  - 4 новых правила лаборатории включены: `casing`, `field_not_required`,
    `package_no_import_cycle`, `syntax_specified` (решено пользователем);
  - buf-совместимое поведение COMPILE принято: файл, не прошедший линковку,
    репортится как issues `COMPILE`, правила на нём не запускаются (решено
    пользователем);
  - паритет замечаний с текущим main на корпусах лаборатории (api-common,
    googleapis-nocompute) при отсутствии ошибок компиляции;
  - `task test`, `task lint` зелёные.
- **Deferred (v2):**
  - удаление `go-protoparser` (зависит от миграции breaking);
  - миграция breaking/generate/tidy/migrate на компилятор.
- **Needs spike:**
  - приёмочный прогон lint-diff: воспроизвести стенд `bench/lab.py` локально
    (корпуса пересобираются из googleapis; 2.9 ГБ parity-выводов в repo нет).

## Assumptions & Open Questions

Допущения:
- [ASSUMPTION: `compiler-lab/easyp-cli` — эталонный источник; доступ через gh
  сохранится на время работы]
- [ASSUMPTION: компилятор вкопится в `internal/easypc` как в лаборатории —
  без внешнего модуля и без `replace` в go.mod]
- [ASSUMPTION: `internal/parallel` в main уже совместим с лабораторным
  использованием (пакет есть в обоих)] — проверить в Design.

Открытые вопросы — закрыты пользователем 2026-11-17:
1. 4 новых правила лаборатории — включить в фичу.
2. COMPILE-поведение — принять.
3. Вендоринг — `easypc` целиком.
4. Приёмочный корпус — api-common + googleapis-nocompute.

Остался один вопрос для Requirements: включение 4 новых правил расширяет
паритетный критерий — их замечания сравниваются не с текущим main (там их
нет), а с выводом лабораторной версии CLI на тех же корпусах.
