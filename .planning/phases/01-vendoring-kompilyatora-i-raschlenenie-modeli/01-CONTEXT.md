# Phase 1: Вендоринг компилятора и расчленение модели - Context

**Gathered:** 2026-11-17
**Status:** Ready for planning

<domain>
## Phase Boundary

`internal/easypc` (компилятор целиком: 13 пакетов, включая `cmd/` и `testdata/`) и `internal/protoast` вкопированы из compiler-lab (ветка `easyp-cli`, зафиксированный SHA) как чистое добавление с нулевым диффом `go.mod`/`go.sum`. Lint-путь полностью расщепляет модель чтения с breaking: breaking становится единственным владельцем legacy `protoInfoRead`/`ProtoInfo`, lint получает собственные типы. Контракт позиций (1-based строка, 1-based колонка в рунах) доказан двухуровневым тестом ДО любого переноса правил. Поведение всех подсистем для пользователей не меняется.

Не входит: линкер/движок (Phase 2–3), перенос правил (Phase 4), переключение lint-пути (Phase 5).

</domain>

<decisions>
## Implementation Decisions

### Вендоринг хелперов
- **D-01:** `lint_helpers.go` и `casing.go` переносятся в Phase 1 ЦЕЛИКОМ, как в лаборатории (не подмножество, не отложено в Phase 4). Обоснование: будущий синк с компилятором тривиален, дифф при переносе правил в Phase 4 минимален. Допустимый временный «мёртвый» код до Phase 4.
- **D-02:** Исключения в golangci-lint для хелперов НЕ настраиваются. Экспортируемые функции не должны триггерить линтер; если `task lint` всё же упадёт — исполнитель решает по факту (минимальное точечное исключение).

### Гигиена вендор-дерева
- **D-03:** Провенанс фиксируется в `internal/easypc/VENDORED.md`: upstream репозиторий (compiler-lab), ветка (`easyp-cli`), SHA, дата вкопирования, процедура обновления (синка). — **Reversibility:** reversible — один файл, удаляется/обновляется тривиально.
- **D-04:** Правки внутри `internal/easypc` ЗАПРЕЩЕНЫ, кроме единственного исключения: переписывание импорт-путей на `github.com/easyp-tech/easyp/internal/...` один раз при вкопировании. Любые фиксы компилятора идут вверх по потоку (compiler-lab) и приходят синком. Запрет отражается в AGENTS.md и VENDORED.md. — **Reversibility:** costly — правило предотвращает drift вендоренной копии; отмена ломает трассируемость синков.
- **D-05:** Собственные тесты easypc ВКЛЮЧАЮТСЯ в `task test` как обычные пакеты; `testdata/` остаётся в репозитории (нужна этим тестам).

### Механика расщепления модели
- **D-06:** Полное расщепление сразу: lint получает собственные типы модели чтения, breaking забирает legacy `protoInfoRead`/`ProtoInfo` в единоличное владение. Без общего типа/alias — чистая граница, Phase 3/5 не тянут наследие. — **Reversibility:** costly — откат затрагивает все места конструирования в `internal/core`.
- **D-07:** Контракт позиций доказывается ДВУМЯ уровнями тестов: (а) юнит-тест на конверсию спанов компилятора → `protoast.Position` (1-based строка, 1-based колонка в рунах); (б) golden-тест сквозного пути: proto-файл с многобайтовым UTF-8 (кириллица/эмодзи в комментариях и строках) → замечания текущего линтера vs позиции protoast. Ловит off-by-one и byte-vs-rune до переноса правил.

### Эталонная ревизия
- **D-08:** Вендоринг выполняется с ЗАФИКСИРОВАННОГО SHA ветки `easyp-cli` (определяется при планировании/исполнении, записывается в VENDORED.md до вкопирования). Новые синки — только осознанно, с новым SHA.
- **D-09:** Доступ executor'а к compiler-lab — через `gh repo clone` с этой машины (gh-аутентификация уже работает, explore-исследование клонировало так же). Не риск, стандартный путь.

### Claude's Discretion
- Точная механика переименования/расщепления типов в `internal/core` (какие имена, где границы файлов) — planner решает по фактическому коду.
- Выбор конкретных UTF-8 фикстур и их placement (testdata-структура).
- Формат VENDORED.md (минимально достаточный: repo/branch/SHA/date/procedure).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Спецификация фичи
- `.spec/features/lint-easypc-infra/requirements.md` — исходные требования milestone (REQ-1.1, REQ-1.2 — эта фаза), глоссарий (easypc, protoast, паритет), топологический порядок
- `.spec/features/lint-easypc-infra/explore.md` — исследование эталона: состав compiler-lab (48 .go-файлов, 13 пакетов), отличия lint.go лаборатории, список новых правил

### Правила репозитория
- `AGENTS.md` — конвенции: colocated тесты, `task`-таргеты, запрет параллельных agent-файлов; сюда добавится запрет правок в `internal/easypc` (D-04)
- `.spec/agent-rules.md` — обязательные стандарты кодирования
- `Taskfile.yml` — таргеты test/lint/build для verification-команд

### Research
- `.planning/research/STACK.md` — процедура вендоринга (copy-in), почему альтернативы (module+replace, git subtree, go mod vendor) отвергнуты
- `.planning/research/ARCHITECTURE.md` — направление зависимостей (`api → core → protoast → easypc`), паттерн расщепления, прецедент generate.go
- `.planning/research/PITFALLS.md` — P1 (позиционная точность: 1-based руны vs 0-based спаны), P9 (vendor drift), P5 (двойная модель)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/parallel` — bounded atomic-index scheduling; будет переиспользован движком (Phase 3), в этой фазе не трогаем
- `internal/core/lint.go` (124 строки) — текущий последовательный движок; его модель чтения подлежит расщеплению
- `internal/core/dom.go:48` — `core.ProtoInfo.Info *unordered.Proto` (go-protoparser), точка расщепления типов
- `internal/core/comment_suppressions.go` — reflection-walk по go-protoparser Meta; остаётся на месте (принадлежит старому пути)

### Established Patterns
- Colocated тесты: каждый `<file>.go` + `<file>_test.go` (testify) — фикстуры позиций следуют этому паттерну
- `core.New` получает полные опции, без post-construction сеттеров — новые типы должны это сохранять
- Схемы генерируются только через `task schema:generate` — в этой фазе не меняются (новые правила = Phase 6)

### Integration Points
- `internal/api/runtime.go` (`buildCore`) — шов, где lint/breaking движки конструируются; расщепление типов должно пройти через него без изменения поведения
- `internal/core/breaking_check.go` — потребитель `protoInfoRead`/`Issue.Position`; после расщепления — единственный владелец legacy-пути

</code_context>

<specifics>
## Specific Ideas

- Эталон вендоринга — приватный репозиторий `easyp-tech/compiler-lab`, ветка `easyp-cli`; клонирование через `gh repo clone` уже проверено explore-исследованием (2026-11-16)
- Объём вендоринга: ~3.4 МБ (48 .go-файлов, 33 пакета без тестов, `pkg/` + `protoc/` + `cmd/easyp-protoc` + testdata)
- Импорт-пути в лабораторной копии УЖЕ переписаны на `github.com/easyp-tech/easyp/internal/...` — переписывание минимально

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 1-Вендоринг компилятора и расчленение модели*
*Context gathered: 2026-11-17*
