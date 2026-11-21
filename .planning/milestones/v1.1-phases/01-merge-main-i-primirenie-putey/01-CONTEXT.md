# Phase 1: Merge main и примирение путей - Context

**Gathered:** 2026-11-21
**Status:** Ready for planning

<domain>
## Phase Boundary

Ветка `feat/lint-easypc-infra` (носитель PR #258) получает актуальный `origin/main` вместе с коммитом #253 «Use compiler for breaking check»: breaking-путь начинает потреблять `github.com/easyp-tech/easyprotoc/protoc`, `go.mod` получает пин `v0.0.0-20261104120806-f378f3f3a05f` и protobuf `v1.37.0-rc.1`. Конфликты `internal/core` разрешаются семантически: поведение main выигрывает там, где #253 меняет breaking/sourceview, а переработанный lint-контур ветки остаётся байт-неизменным (вендор-путь живёт до Phase 2). По завершении фазы ветка не отстаёт от `origin/main`, полная батарея зелёная, паритет lint-пути не сдвинут.

Граница фазы: НЕ переключаем импорты lint-потребителей и НЕ удаляем `internal/easypc` — это Phase 2. НЕ трогаем `internal/protoast` — его импорты меняются в Phase 2.

</domain>

<decisions>
## Implementation Decisions

### Стиль слияния
- **D-01:** Слияние `origin/main` выполняется **merge-коммитом** (не rebase). PR #258 с 40+ коммитами открыт; rebase переписал бы историю и сломал review. Merge сохраняет SHA и делает последующий мерж PR #258 в main тривиальным. — **Reversibility:** costly — откат переписывает публичную историю PR-ветки.
- **D-02:** Конфликты `internal/core` правятся **внутри merge-коммита** (один коммит `merge origin/main`), не отдельными follow-up коммитами: PR-мерж не должен требовать ручного разрешения конфликтов.

### Семантика разрешения конфликтов
- **D-03:** Правило приоритета: **main wins** для поведения breaking/sourceview (#253 — `breaking_profiles.go`, `core.go`, `dom.go`, `internal/sourceview/*`); **ветка wins** для переработанного lint-контура (новые файлы `lint_linked.go`, `link.go`, `proto_info_read.go` и т.д. остаются как есть). Пересечение (общие структуры `Core`, `dom.go`, `core.go`) — semantic merge: обе стороны сосуществуют, breaking использует `easyprotoc/protoc`, lint остаётся на вендор-пути `internal/easypc/protoc`.
- **D-04:** Если после merge падает тест из-за изменения поведения main (#253), разбор точечный: поведение main — истина для breaking/sourceview (прецедент REQ-3.2 v1.0: «main wins over lab copy»); lint-ожидания (golden, parity) не правятся — падение означает ошибку разрешения конфликта, а не новую норму.

### Зависимости
- **D-05:** `go.mod` принимает main-пин `github.com/easyp-tech/easyprotoc v0.0.0-20261104120806-f378f3f3a05f` и `google.golang.org/protobuf v1.37.0-rc.1` **как есть** (MVS, без понижения, без `replace`). Тот же пин переиспользуется в Phase 2 — никаких повторных go get.

### Дормация вендор-пути
- **D-06:** `internal/easypc` и все импорты `internal/easypc/...` в Phase 1 **не трогаются** (byte-frozen): фаза проверяет, что merge не сломал существующий lint-путь. Удаление/переключение — строго Phase 2.
- **D-07:** `internal/protoast` в Phase 1 не меняется (D-07a/D-07b position-тесты — часть гейта).

### Ветка и коммиты
- **D-08:** Работа идёт в текущей ветке `feat/lint-easypc-infra` (GSD `branching_strategy: none`); коммиты атомарные, дефолтная ветка не затрагивается.

### Ворота фазы
- **D-09:** Гейт после merge: `go build ./...` + `go test -race -count=1 ./...` (полная батарея), плюс корпусный lint-паритет Gates A–D ×2 и golden 94 findings байт-точно без правки ожиданий; `git merge-base --is-ancestor origin/main HEAD` истинно.

### Claude's Discretion
Конкретные hunks при разрешении конфликтов, порядок прогонов, формулировки комментариев в merge-коммите — на усмотрение исполнителя в рамках D-03/D-04.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone / phase
- `.planning/PROJECT.md` — цель v1.1, инварианты, Key Decisions v1.0
- `.planning/REQUIREMENTS.md` §Reconciliation — IMPORT-REC-01 (полный текст требования)
- `.planning/ROADMAP.md` §Phase 1 — success criteria фазы
- `.planning/STATE.md` — Accumulated Context, блокеры v1.1

### Код и конвенции
- `AGENTS.md` — проектные конвенции (build/test/lint, layout, запреты)
- `.spec/ARCHITECTURE.md` — архитектура lint/breaking путей (контекст конфликтов `internal/core`)
- `internal/easypc/VENDORED.md` — провенанс вендор-копии (удаляется в Phase 2, в Phase 1 — источник истины о пине `08cef10`)
- `go.mod` / `go.sum` — текущие пины зависимостей

### Upstream
- `github.com/easyp-tech/easyprotoc` @ `f378f3f3a05febd98f54aff2be9a570de239a59a` — публичный main компилятора (вендор-копия ему соответствует)
- Коммит EasyP #253 (5d36266) — «Use compiler for breaking check»: эталон разрешения для breaking/sourceview

### v1.0 прецеденты (архив)
- `.planning/milestones/v1.0-ROADMAP.md` — инварианты v1.0 (main wins, D-04 seal, dormancy)
- `.planning/milestones/v1.0-REQUIREMENTS.md` — REQ-3.2 (main behavior wins over lab copy)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/core/link.go` — линкер lint-пути на вендор-компиляторе (не трогать в Phase 1)
- `internal/core/breaking_profiles.go` — после merge получает easyprotoc/protoc от main (#253)
- `internal/sourceview/session.go`, `root_tree.go` — новые файлы main #253 (session cache, root tree)
- `internal/protoast/position_test.go` + `internal/core/position_parity_test.go` — D-07 позиционные гейты

### Established Patterns
- Дормация: старый и новый контуры сосуществуют, поведение доказывается байт-идентичными golden
- Parity-гейты: Gates A–D ×2 на teeth-корпусе, golden 94 findings байт-точно
- Merge-конвенция: атомарные коммиты в текущей ветке (GSD branching_strategy: none)
- `-race` полная батарея: 41+ пакет, ≥3718 тестов (эталон v1.0 Phase 7)

### Integration Points
- `internal/core/core.go`, `dom.go`, `breaking_profiles.go` — конфликтные точки merge (main #253 vs переработанный core ветки)
- `internal/sourceview/*` — main #253 добавляет/меняет; ветка могла не трогать
- `go.mod` / `go.sum` — MVS-слияние пинов
- `.planning/phases/01-merge-main-i-primirenie-putey/` — артефакты фазы

</code_context>

<specifics>
## Specific Ideas

- Merge-коммит сообщением вида `merge: origin/main (#253 Use compiler for breaking check)` с перечислением разрешённых конфликтных файлов в теле.
- Проверка «ветка не отстаёт»: `git rev-list --count HEAD..origin/main` == 0 после merge.

</specifics>

<deferred>
## Deferred Ideas

- Взятие upstream-ветки `feat/easyp-integration` (generation metadata fix) — не в v1.1, ждёт мержа upstream / codegen-серии.
- Bump easyprotoc при появлении тегов — v2 requirements.

</deferred>
