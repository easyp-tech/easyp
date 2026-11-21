# Phase 2: Переключение потребителей и удаление вендор-копии - Context

**Gathered:** 2026-11-21
**Status:** Ready for planning

<domain>
## Phase Boundary

Импорты компилятора у 7 потребителей переключаются с вендор-путей `github.com/easyp-tech/easyp/internal/easypc/...` на публичные пакеты модуля `github.com/easyp-tech/easyprotoc`; дерево `internal/easypc` (48+ файлов, `VENDORED.md`, `cmd/easyp-protoc`, `cmd/lexbench`) удаляется целиком в том же атомарном коммите, `go mod tidy` идемпотентен и сохраняет easyprotoc прямой зависимостью. `internal/protoast` остаётся локальным швом — меняются только его import-строки. Поведение не меняется: полная `-race` батарея, golden 94, Gates A–D ×2 и D-07 позиции зелёные без правки ожиданий.

Граница фазы: НЕ трогаем `.golangci.yml`, `AGENTS.md` и `.spec` (seal-инфраструктура — Phase 3). НЕ меняем логику protoast, правила, движок, API.

</domain>

<decisions>
## Implementation Decisions

### Маппинг импортов
- **D-01:** Ровно три замены, без алиасов и промежуточных пакетов-шимов:
  - `github.com/easyp-tech/easyp/internal/easypc/protoc` → `github.com/easyp-tech/easyprotoc/protoc`
  - `github.com/easyp-tech/easyp/internal/easypc/protoc/syntax` → `github.com/easyp-tech/easyprotoc/protoc/syntax`
  - `github.com/easyp-tech/easyp/internal/easypc/pkg/parser` → `github.com/easyp-tech/easyprotoc/parser`
- **D-02:** Потребители — ровно 7 файлов (recon 2026-11-21): `internal/core/link.go`, `internal/core/link_test.go`, `internal/core/lint_linked_matrix_test.go`, `internal/protoast/build.go`, `internal/protoast/outline.go`, `internal/protoast/protoast.go` (комментарий, ссылающийся на `internal/easypc/VENDORED.md`), `internal/protoast/protoast_test.go`. Других потребителей нет.

### Удаление вендор-копии
- **D-03:** `internal/easypc` удаляется целиком в том же атомарном коммите, что и переключение импортов: сборка обязана быть зелёной на этом коммите. Отдельного «промежуточного» коммита с полу-удалённым деревом нет.
- **D-04:** `cmd/easyp-protoc` и `cmd/lexbench` удаляются вместе с деревом; внешних потребителей у них нет (recon: ссылки только внутри вендор-дерева и его собственных тестов). Замена локального бинаря не нужна — upstream предоставляет `cmd/easyprotoc` как модульный бинарь (зафиксировать в доках Phase 3).

### Protoast
- **D-05:** `internal/protoast` остаётся локальным швом: меняются только import-строки (и комментарий с путём к VENDORED.md). Логика `build.go` (включая rune-колонки `pos()`), `outline.go`, `protoast.go` и все тесты, кроме импортов, — без изменений. D-07a/D-07b обязаны пройти.

### Зависимости и tidy
- **D-06:** `easyprotoc` в `go.mod` уже прямой (после merge main #253). После удаления вендор-дерева `go mod tidy` не должен удалить его и не должен давать diff: `go mod tidy -diff` пуст. Никаких `replace`/`vendor`.

### Seal-инфраструктура
- **D-07:** `.golangci.yml`-исключение `internal/easypc`, комментарий про compiler-lab/VENDORED.md и seal-параграф AGENTS.md в Phase 2 **не трогаются** (владелец — Phase 3). Устаревшее исключение несуществующего пути безвредно на одну фазу.

### Ветка, коммиты, ворота
- **D-08:** Работа в текущей ветке `feat/lint-easypc-infra`; коммиты атомарные, без push.
- **D-09:** Ворота: `go build ./...`; полная `go test -race -count=1 ./...`; именованные гейты — `TestAcceptanceGates` (Gates A–D ×2), `TestLintGoldenFullRegistry` (94, байт-точно), `TestPositionRuneColumns` + `TestProtoastPositionsMatchLegacyContract` (D-07), core/CLI матрица; `go mod tidy -diff` пуст; grep `internal/easypc` по `*.go` пуст (кроме истории в `.planning`); `git ls-files internal/easypc` пуст. Ожидания не правятся: падение = ошибка замены.

### Claude's Discretion
Порядок правок внутри атомарного коммита, формулировка комментария-замены в `protoast.go`, точные сообщения коммитов.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone / phase
- `.planning/PROJECT.md` — цель v1.1, инварианты
- `.planning/REQUIREMENTS.md` §Consumers/Vendor removal/Dependency — IMPORT-DEP-01, IMPORT-CONS-01, IMPORT-VENDRM-01
- `.planning/ROADMAP.md` §Phase 2 — success criteria
- `.planning/phases/01-merge-main-i-primirenie-putey/01-RESEARCH.md` — layout mapping, upstream packages, MVS pins
- `.planning/phases/01-merge-main-i-primirenie-putey/01-SECURITY.md` — supply-chain register

### Код
- `internal/easypc/VENDORED.md` — провенанс удаляемой копии (удаляется этой фазой)
- `internal/protoast/protoast.go` — шов, дивергенция rune-колонок (комментарий + ссылка на VENDORED.md к обновлению)
- `internal/core/link.go` — линкер lint-пути (потребитель `protoc`)
- `go.mod` / `go.sum` — пины easyprotoc/protobuf

### Upstream
- `github.com/easyp-tech/easyprotoc` @ `f378f3f3a05febd98f54aff2be9a570de239a59a` — публичные пакеты `protoc`, `protoc/syntax`, `parser`

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- Вендор-копия побайтово равна публичному модулю (diff по protoc/compile.go, parser/ast.go, protoc/syntax/syntax.go, protoc/resolver.go — только import-строки/комментарии) → замена импортов механическая.
- `internal/protoast/position_test.go` + `internal/core/position_parity_test.go` — D-07 гейты.

### Established Patterns
- Атомарная замена с зелёными воротами (прецедент v1.0 Phase 5 flip).
- Ожидания тестов не правятся при миграциях (main/эталон wins).

### Integration Points
- 7 файлов-потребителей; `go.mod`; удаляемое дерево `internal/easypc`.

</code_context>

<specifics>
## Specific Ideas

- Один feat-коммит: «feat(lint): consume public easyprotoc module; drop vendored easypc tree» с перечислением маппинга в теле.
- Проверка отсутствия вендор-путей: `grep -rn "internal/easypc/" --include='*.go' .` → пусто.

</specifics>

<deferred>
## Deferred Ideas

- Seal-инфраструктура и доки (`.golangci.yml`, AGENTS.md, `.spec`) — Phase 3.
- Bump easyprotoc при появлении тегов / мерже `feat/easyp-integration` — v2.
</deferred>
