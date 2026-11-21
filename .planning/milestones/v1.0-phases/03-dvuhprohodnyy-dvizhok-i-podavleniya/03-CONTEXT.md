# Phase 3: Двухпроходный движок и подавления - Context

**Gathered:** 2026-11-19
**Status:** Ready for planning

<domain>
## Phase Boundary

Перезаписанный lint-движок на слинкованных деревьях `easypc`/`protoast`, подготовленный в Phase 2 (`(*Core).linkSources`): параллельные окна по файлам, `ModuleRule` (Collect/Finish) для кросс-файловых правил, детерминированный порядок вывода, центральный фильтр подавлений (ignore/ignore_only конфига + комментарии) и валидация имён правил/групп в `lint.use` с паритетом main. Движок готов и проверен на уровне core/rules, но НЕ врезан в `easyp lint` (переключение — Phase 5; dormancy D-04 продолжается).

Не входит: перенос 43 правил (Phase 4), переключение lint-пути (Phase 5), линкер (готов в Phase 2).

</domain>

<decisions>
## Implementation Decisions

### Детерминизм вывода
- **D-01:** Полный детерминированный порядок выдачи замечаний — тотальная сортировка по (path, rule, line, col) на выходе движка. Main недетерминирован (map iteration); новый движок стабилен по определению — это усиление, не паритет (порядок main не является контрактом; проверено research Phase 1). — **Reversibility:** reversible — одна точка сортировки на выходе.

### Параллелизм
- **D-02:** Параллельные окна через `internal/parallel` (bounded atomic-index scheduling), размер окна и параллелизм — как в эталоне лаборатории (256 файлов на окно, NumCPU воркеров); per-file правила параллельны, `ModuleRule.Finish` — последовательно после сбора всех окон. Порог «маленький корпус — без горутин» из эталона сохраняется.

### ModuleRule
- **D-03:** Контракт `ModuleRule` (Collect/Finish) переносится из эталона в `internal/core` для кросс-файловых правил (PACKAGE_* семейство); обычные правила остаются per-file. Список ModuleRule-правил фиксируется на этапе движка (из 43 текущих правил — те, что в лаборатории ModuleRule); сами правила переносятся в Phase 4.

### Подавления
- **D-04:** Центральный фильтр подавлений в одной точке движка: ignore/ignore_only конфига + комментарийные подавления (`allowCommentIgnores`) с семантикой текущего main (REQ-3.2: main выигрывает у лаборатории). Комментарийный walk переписывается на lint-модель чтения (protoast-деревья из слинкованных файлов) вместо reflection по go-protoparser Meta; паритет доказывается таблицей тестов main-vs-new на фикстурах с комментариями.
- **D-05:** `shouldIgnore` остаётся методом Core с проверкой ignore/ignore_only как в main; комментарийные подавления собираются per-file в первом проходе и применяются во втором — двухпроходная схема эталона.

### Валидация конфигурации
- **D-06:** Неизвестное имя правила или группы в `lint.use` → ошибка валидации с паритетом main (текст/место ошибки — как в main); валидация живёт на уровне `internal/core` при конструировании движка, множество имён — из реестра правил builder'а.

### Модель чтения движка
- **D-07:** Движок строит per-file модель из `Linked.Syntax` (protoast-деревья, подготовленные Phase 2) один раз на файл; lint-типы чтения из расщепления Phase 1 (D-06 Phase 1) наполняются из деревьев — без повторного чтения диска для файлов корпуса; `parseOnTheirOwn`-путь эталона для файлов, откатившихся на пофайловый разбор.

### Claude's Discretion
- Точные имена/границы файлов движка (lint engine rewrite рядом с текущим lint.go, старый остаётся до Phase 5).
- Внутренняя структура checkedFile/результатов окон.
- Формат таблицы паритет-тестов подавлений.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Спецификация фичи
- `.spec/features/lint-easypc-infra/requirements.md` — REQ-2.1 (модель из слинкованных деревьев, порядок обхода), REQ-2.5 (подавления как в main), REQ-2.6 (валидация lint.use)
- `.planning/REQUIREMENTS.md` — LINT-ENG-01, LINT-ENG-05, LINT-ENG-06 (все три — Phase 3) + Open Design Question «порядок эмиссии замечаний» (решается здесь, D-01)

### Правила репозитория
- `AGENTS.md` — конвенции; запрет правок в `internal/easypc`; colocated тесты
- `.spec/agent-rules.md`; `Taskfile.yml`

### Roadmap
- `.planning/ROADMAP.md` § Phase 3 — цель, success criteria, research flag

### Prior phases
- `.planning/phases/01-vendoring-kompilyatora-i-raschlenenie-modeli/01-CONTEXT.md` — D-06 (расщепление моделей чтения), D-07 (позиции 1-based руны)
- `.planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-CONTEXT.md` — D-01..D-05 (классификатор, COMPILE-формат, dormancy), `(*Core).linkSources` контракт

### Research
- `.planning/research/PITFALLS.md`, `.planning/research/ARCHITECTURE.md` — направление зависимостей, паттерны

### Эталон
- Эталонный движок: `compiler-lab/easyp-cli/internal/core/lint.go` (клон исследователя `/tmp/gsd-research/compiler-lab`, fallback `gh repo clone easyp-tech/compiler-lab` @ `08cef10`, директория `easyp-cli`) + эталонные `comment_suppressions` и ModuleRule-правила того же дерева

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `(*Core).linkSources` (Phase 2) — трёхисходный вход движка: деревья / COMPILE issues / откат
- `internal/parallel` — bounded atomic-index scheduling (59 строк), переиспользуется как есть
- `internal/core/comment_suppressions.go` (231 строка) — текущий комментарийный walk по go-protoparser Meta; семантика подавлений = источник паритета
- `internal/core/lint.go` (124 строки) — текущий последовательный движок; остаётся владельцем пользовательского пути до Phase 5
- `internal/rules/builder.go` — реестр правил/групп; источник имён для D-06
- Эталонные ModuleRule-примеры: `compiler-lab/.../internal/rules/package_same_java_package.go`, `package_same_go_package.go` (Collect/Finish)

### Established Patterns
- `core.New` получает полные опции; colocated testify тесты; `-race` обязателен
- Фикстуры testdata-дерева (прецедент Phase 2 `testdata/link/`)

### Integration Points
- `internal/api/runtime.go` (`buildCore`) — движок конструируется рядом с существующим, без переключения
- `internal/core/lint.go` — старый путь не меняется до Phase 5; новый движок dormant, как линкер Phase 2

</code_context>

<specifics>
## Specific Ideas

- Эталон лаборатории: `lint.go` 236 строк — окна по 256 файлов, `parseOnTheirOwn` для откатившихся файлов, `shouldIgnore` + комментарийные подавления, `close`-гигиена
- main выигрывает у лаборатории на подавлениях (REQ-3.2/2.5): поведение подавлений фиксируется по main, не по эталону
- MVP-режим фазы — прецедент фаз 1–2 ратифицирован (goal-backward верификация, без User Story)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 3-Двухпроходный движок и подавления*
*Context gathered: 2026-11-19*
