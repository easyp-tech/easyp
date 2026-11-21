# Phase 4: Приёмка — полный паритет - Context

**Gathered:** 2026-11-21
**Status:** Ready for planning

<domain>
## Phase Boundary

Финальная приёмка милестоуна v1.1 на текущем HEAD: полный набор ворот доказывает, что модульная зависимость `github.com/easyp-tech/easyprotoc` не изменила поведение. Фаза — evidence-only: исходники/доки/конфиги не меняются, создаётся только SUMMARY (и, при необходимости, evidence-файлы в `.planning/`). Прогоняются: `go build ./...`; одна полная `go test -race -count=1 -timeout 20m -v ./...`; именованные гейты (Gates A–D ×2 против baseline `1e9a596`, golden 94, оба D-07, core/CLI матрица) по точным `=== RUN` + `--- PASS`; `bin/golangci-lint run ./...` → ровно `0 issues.`; schema-drift (`go run ./cmd/easyp schema-gen` → нулевой diff); `go mod tidy -diff` пуст; seal grep пуст; пины точные.

Ожидаемое следствие удаления вендоренного дерева: батарея v1.0 (41 тест-пакет, 3718 тестов) сокращается до 34 тест-пакетов — это дизайн (7 пакетов вендора удалены вместе с их тестами), а не регрессия. Паритет доказывают golden/Gates/D-07, а не число тестов.

</domain>

<decisions>
## Implementation Decisions

### Режим фазы
- **D-01:** Evidence-only: никаких изменений исходников/тестов/ожиданий/доков. Разрешённые новые файлы — SUMMARY плана (и опционально evidence JSON/MD в каталоге фазы).
- **D-02:** Одна полная батарея, один раз, с `-v` (skip-строки наблюдаемы); ожидаемо 34 `ok` / 0 `FAIL` / 0 race / 0 непредусмотренных skip (известный: `TestGenerateGoHeaderUsesCompilerVersion`, protoc-gen-go отсутствует) / 3 `[no test files]`.
- **D-03:** Именованные гейты: `TestAcceptanceGates`, `TestLintGoldenFullRegistry`, `TestPositionRuneColumns`, `TestProtoastPositionsMatchLegacyContract`, `TestLintMatrix`, `TestLintErrorPathFidelity`, `TestLintSuppressionParity`, `TestLintStdoutGolden`, `TestLintLinkMatrix`, `TestLintLinkMatrixExitCodes`, `TestPrintIssuesCompileGolden` — каждый по exact RUN + PASS.
- **D-04:** Атрибуция: зафиксировать HEAD SHA, гейты прогоняются на чистом дереве; tree HEAD = дерево после финальных коммитов фазы 3; никаких `.planning`-правок между прогонами, кроме метаданных SUMMARY.
- **D-05:** Schema drift: `go run ./cmd/easyp schema-gen` (эквивалент `task schema:generate`) → `git diff --exit-code schemas/` пуст.
- **D-06:** Сравнение с v1.0: зафиксировать в SUMMARY дельту 41→34 тест-пакета и 3718→N тестов как дизайн (удалены вендорные тесты), с указанием, что паритет поведенческий (golden/Gates/D-07), а не количественный.
- **D-07:** Падение любого гейта — дефект милестоуна: не «чинить» ожидание, а записать gap и остановиться.
- **D-08:** Push/PR — вне фазы; решение о публикации принимается на milestone-close (UAT-гейт с пользователем).
- **D-09:** Без `task`/Docker; использовать `go`, `bin/golangci-lint`.

### Claude's Discretion
Порядок прогонов, набор evidence-файлов, формулировки SUMMARY.

</decisions>

<canonical_refs>
- `.planning/REQUIREMENTS.md` — IMPORT-PARITY-01
- `.planning/ROADMAP.md` §Phase 4 — success criteria
- `.planning/phases/01-.../01-RESEARCH.md` §Validation Architecture — exact gate names and logs
- `.planning/phases/01-.../01-02-SUMMARY.md`, `02-.../02-02-SUMMARY.md` — prior gate evidence
- `.planning/milestones/v1.0-ROADMAP.md`/`v1.0-REQUIREMENTS.md` — v1.0 acceptance baseline (Gates A–D, golden 94, 3718 tests)
- `internal/api/acceptance_gates_test.go`, `internal/rules/golden_full_registry_test.go`, `internal/protoast/position_test.go`, `internal/core/position_parity_test.go`
</canonical_refs>

<code_context>
- HEAD после фаз 1–3: merge #253, публичный модуль, вендор-дерево удалено, lint 0 issues.
- Полная батарея: 34 тест-пакета (было 41); известный skip один.
- Schema-gen — через `go run ./cmd/easyp schema-gen`; schemas/ не должны дрейфовать (v1.1 не меняла схемы).

</code_context>

<deferred>
## Deferred Ideas

- Push/PR и milestone-close — после фазы, UAT-гейт.
- Corpora letters / bump easyprotoc — backlog/v2.
</deferred>
