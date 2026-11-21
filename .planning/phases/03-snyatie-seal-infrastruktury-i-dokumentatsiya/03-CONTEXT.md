# Phase 3: Снятие seal-инфраструктуры и документация - Context

**Gathered:** 2026-11-21
**Status:** Ready for planning

<domain>
## Phase Boundary

Seal-инфраструктура вендоренной копии упраздняется, действующие конфиги/документация описывают компилятор как закреплённую модульную зависимость: `.golangci.yml`-исключение `internal/easypc` и его комментарий удаляются; параграф «Do not» в AGENTS.md заменяется правилом про модуль (фиксы — upstream, приходят bump'ом версии; копировать код модуля в репозиторий нельзя); `.spec` фиксирует потребление модуля и локальный шов `internal/protoast`. Заодно закрываются три doc-citation INFO-находки Phase 2 и устраняется pre-existing errcheck в `internal/sourceview/session_test.go:29`, из-за которого `bin/golangci-lint run ./...` красный. `bin/golangci-lint run ./...` становится зелёным (0 issues).

Граница фазы: НЕ меняем поведение кода (кроме тестового errcheck-фикса и комментариев); полная батарея/паритет — Phase 4.

</domain>

<decisions>
## Implementation Decisions

### Конфиг линтера
- **D-01:** `.golangci.yml`: удалить `exclusions.paths: [internal/easypc]` и комментарий про sealed vendor/compiler-lab/VENDORED.md; блок `exclusions.rules` (ST1005) сохранить без изменений.

### AGENTS.md
- **D-02:** Заменить bullet «Edit anything inside internal/easypc…» на правило: компилятор потребляется как закреплённый модуль `github.com/easyp-tech/easyprotoc`; код модуля не копируется/не вендорится в репозиторий; фиксы уходят upstream и приходят bump'ом версии в `go.mod`. Добавить в Layout-таблицу строку `internal/protoast` (локальный шов адаптации, дивергенция rune-колонок D-07).
- **D-03:** `.claude/CLAUDE.md` — генерируемое зеркало; seal-упоминаний не содержит → не трогаем (никакой регенерации в этой фазе).

### .spec
- **D-04:** `.spec/ARCHITECTURE.md`: добавить точное предложение в раздел про `internal/core`, что пакеты компилятора (`protoc`, `protoc/syntax`, `parser`) приходят из закреплённого модуля `github.com/easyp-tech/easyprotoc`, а `internal/protoast` — локальный шов с дивергенцией rune-колонок (D-07). `.spec/PACKAGES.md`: добавить строку `internal/protoast` (шов) и обновить строку `internal/core`, если она описывает удалённые файлы (`lint.go` и т.п. — проверить по факту и поправить только фактически неверное). Исторические `.spec/plans/*` и `.spec/features/lint-easypc-infra/*` — не трогать (история).

### Doc-citation фиксы Phase 2
- **D-05:** `internal/core/link_test.go:288`: квалифицировать путь — `github.com/easyp-tech/easyprotoc/internal/compiler/driver.go:319-323` (модуль в пине), без изменения смысла комментария.
- **D-06:** `internal/protoast/build.go` re-sync комментарий: явно назвать три поверхности дивергенции (`build.go` `pos()`, `protoast.go` package-doc, `protoast_test.go` tab-ожидания) и оба D-07 гейта (`TestPositionRuneColumns`, `TestProtoastPositionsMatchLegacyContract`).
- **D-07:** `internal/rules/package_no_import_cycle_test.go:28`: «The easypc linker» → «The easyprotoc linker».

### Errcheck
- **D-08:** `internal/sourceview/session_test.go:29`: `defer session.Close()` → `t.Cleanup(func(){ if err := session.Close(); err != nil { t.Errorf("close session: %v", err) } })` (тест-онли; обёртка по конвенции `"<callee>: %w"`).

### Ворота
- **D-09:** `go build ./...`; `bin/golangci-lint run ./...` → 0 issues (это и есть IMPORT-SEAL-01 SC1); `go test -race -count=1 ./internal/sourceview ./internal/core ./internal/protoast ./internal/rules`; grep «easypc|compiler-lab|VENDORED» по действующим конфигам/докам (`.golangci.yml`, `AGENTS.md`, `.spec/*.md` вне plans/features) пуст; `go mod tidy -diff` пуст. Полная батарея — Phase 4.
- **D-10:** Коммиты атомарные: (1) config/docs+doc-citations, (2) errcheck-фикс (тестовая правка) — каждый зелёный на lint/targeted-тестах; без push.

### Claude's Discretion
Точные формулировки правил в AGENTS.md/.spec, порядок правок, разбиение на коммиты в рамках D-10.

</decisions>

<canonical_refs>
## Canonical References

- `.planning/REQUIREMENTS.md` §Seal — IMPORT-SEAL-01
- `.planning/ROADMAP.md` §Phase 3 — success criteria
- `.planning/phases/02-.../02-REVIEW.md` — INFO-находки IN-01..IN-03 (disposition deferred → эта фаза)
- `.planning/phases/01-.../01-RESEARCH.md` §Common Pitfalls 4-5 — pre-existing errcheck, upstream fidelity
- `AGENTS.md`, `.golangci.yml`, `.spec/ARCHITECTURE.md`, `.spec/PACKAGES.md`
- `internal/protoast/protoast.go` (doc), `internal/protoast/build.go` (`pos()`), `internal/protoast/position_test.go`, `internal/core/position_parity_test.go` (D-07 гейты)
- `internal/sourceview/session_test.go` (errcheck), `internal/core/link_test.go:288`, `internal/rules/package_no_import_cycle_test.go:28`

</canonical_refs>

<code_context>
## Existing Code Insights

- Все seal-упоминания живут ровно в двух действующих файлах: `.golangci.yml` (строки 7,9) и `AGENTS.md` (строка 92). `.spec` вендоринг не упоминает — там нужно добавить точное описание модульной зависимости.
- `bin/golangci-lint run ./...` сейчас ровно 1 issue (errcheck в session_test.go:29).
- `.claude/CLAUDE.md` — сгенерированное зеркало без seal-текста.

</code_context>

<deferred>
## Deferred Ideas

- Полная паритет-батарея и schema-гейты — Phase 4.
- Bump easyprotoc / `feat/easyp-integration` — v2.
</deferred>
