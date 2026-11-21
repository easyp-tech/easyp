# Roadmap: EasyP

## Shipped Milestones

- [x] **v1.0: lint-easypc-infra** — `easyp lint` migrated to the vendored `easypc` compiler (7 phases, 32 plans, shipped 2026-11-21, PR #258) — archived: `.planning/milestones/v1.0-ROADMAP.md`

## Milestone v1.1: import-easyprotoc

**Overview:** Репозиторий компилятора `easyp-tech/easyprotoc` стал публичным, поэтому вендор-копия `internal/easypc` заменяется обычной модульной зависимостью. На `main` коммит #253 уже потребляет `easyprotoc/protoc` в breaking-пути с псевдоверсией `v0.0.0-20261104120806-f378f3f3a05f`; этот майлстоун приводит lint-путь к той же зависимости: сперва ветка примиряется с `main` (#253), затем потребители переключаются на публичные пакеты модуля, вендор-дерево и seal-инфраструктура удаляются, и полная паритет-батарея v1.0 подтверждает отсутствие регрессий. Инвариант всего пути: drop-in buf-совместимость (формат вывода issues, пути, коды выхода) и неприкосновенность `internal/protoast`-шва (дивергенция rune-колонок D-07 сохраняется).

**Status:** 🚧 In progress
**Phases:** 1-4
**Milestone Requirements:** IMPORT-REC-01, IMPORT-DEP-01, IMPORT-CONS-01, IMPORT-VENDRM-01, IMPORT-SEAL-01, IMPORT-PARITY-01

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Merge main и примирение путей** - ветка содержит #253: breaking и lint оба на `easyprotoc`, конфликты `internal/core` разрешены, полная батарея зелёная (completed 2026-11-21)
- [x] **Phase 2: Переключение потребителей и удаление вендор-копии** - импорты на публичные пакеты модуля, `internal/easypc` удалён целиком, сборка и тесты зелёные (completed 2026-11-21)
- [x] **Phase 3: Снятие seal-инфраструктуры и документация** - `.golangci.yml`-исключение и seal-параграф AGENTS.md упразднены, `.spec` описывает импорт модуля (completed 2026-11-21)
- [ ] **Phase 4: Приёмка — полный паритет** - полная `-race`-батарея, Gates A–D ×2, golden 94 байт-точно, D-07 позиции, lint/schema-гейты зелёные

## Phase Details

### Phase 1: Merge main и примирение путей
**Goal**: Ветка содержит актуальный `main` (#253 «Use compiler for breaking check»): breaking-путь и lint-путь оба потребляют `easyprotoc/protoc`, конфликты `internal/core` разрешены семантически, поведение lint-пути не изменилось
**Mode:** mvp
**Depends on**: Nothing (first phase)
**Requirements**: IMPORT-REC-01
**Success Criteria** (what must be TRUE):
  1. `git merge-base --is-ancestor origin/main HEAD` истинно (merge-коммит `origin/main` применён), конфликтных маркеров в дереве нет
  2. `go.mod` содержит `github.com/easyp-tech/easyprotoc v0.0.0-20261104120806-f378f3f3a05f` и `google.golang.org/protobuf v1.37.0-rc.1` (как на main #253); `internal/core/breaking_profiles.go` использует `easyprotoc/protoc`
  3. Полная батарея после слияния зелёная: `go build ./...`, `go test -race -count=1 ./...`
  4. Паритет lint-пути после слияния не сдвинулся: корпусные parity-тесты и golden 94 findings проходят байт-точно без правки ожиданий
**Plans:** 2/2 plans complete
- [x] 01-01-PLAN.md — Merge origin/main (#253) with semantic internal/core resolution
- [x] 01-02-PLAN.md — Post-merge full battery and parity-gate evidence

### Phase 2: Переключение потребителей и удаление вендор-копии
**Goal**: Все потребители компилятора импортируют публичные пакеты `github.com/easyp-tech/easyprotoc`, вендор-дерево `internal/easypc` удалено, сборка и полная тестовая батарея зелёные при идемпотентном `go mod tidy`
**Mode:** mvp
**Depends on**: Phase 1
**Requirements**: IMPORT-DEP-01, IMPORT-CONS-01, IMPORT-VENDRM-01
**Success Criteria** (what must be TRUE):
  1. Импорты `github.com/easyp-tech/easyp/internal/easypc/...` отсутствуют в кодовой базе; потребители (`internal/core/link.go`, `link_test.go`, `lint_linked_matrix_test.go`, `internal/protoast/{build,outline,protoast,protoast_test}.go`) импортируют `easyprotoc/protoc`, `easyprotoc/protoc/syntax`, `easyprotoc/parser`
  2. Дерево `internal/easypc` отсутствует целиком: `git ls-files internal/easypc` пуст (включая `VENDORED.md`, `cmd/easyp-protoc`, `cmd/lexbench`)
  3. `go build ./...` и `go test -race -count=1 ./...` зелёные; повторный `go mod tidy` не меняет `go.mod`/`go.sum`; `replace`/`vendor`-директив нет
  4. Дивергенция protoast сохранена без правок шва: `internal/protoast/position_test.go` (D-07a) и `internal/core/position_parity_test.go` (D-07b) зелёные
**Plans:** 2/2 plans complete
- [x] 02-01-PLAN.md — Atomic consumer swap to public easyprotoc imports plus full `internal/easypc` deletion
- [x] 02-02-PLAN.md — Full `-race` battery and named parity-gate evidence on the committed swap tree

### Phase 3: Снятие seal-инфраструктуры и документация
**Goal**: Действующие конфиги и документация больше не описывают вендоринг: исключение линтера и seal-параграф сняты, `.spec` фиксирует потребление публичного модуля и порядок обновления версии
**Mode:** mvp
**Depends on**: Phase 2
**Requirements**: IMPORT-SEAL-01
**Success Criteria** (what must be TRUE):
  1. `.golangci.yml` не содержит исключения `internal/easypc` и комментария про compiler-lab/`VENDORED.md`; `task lint:go` зелёный
  2. AGENTS.md: параграф «Do not» про sealed vendor copy заменён описанием потребления модуля (версия пинится в `go.mod`, фиксы/ре-синк — через bump версии, а не копирование); layout-таблицы не вводят в заблуждение
  3. `.spec` (актуальные документы, где упоминается lint-путь/архитектура) и README описывают компилятор как зависимость; ссылок на `internal/easypc/VENDORED.md` в действующих доках нет
  4. Grep по действующим конфигам/докам не находит вендор-упоминаний; исторические записи (`.planning/milestones/`, `.spec/features/lint-easypc-infra/`) сохранены как история
**Plans:** 2/2 plans complete
- [x] 03-01-PLAN.md — Remove the seal infrastructure and document the pinned easyprotoc module
- [x] 03-02-PLAN.md — Fix the pre-existing session close errcheck and prove the phase gates

### Phase 4: Приёмка — полный паритет
**Goal**: Полный набор приёмки v1.0 проходит на модульной зависимости без единого изменения ожиданий: паритет замечаний, позиции, детерминизм и гейты качества подтверждены
**Mode:** mvp
**Depends on**: Phase 3
**Requirements**: IMPORT-PARITY-01
**Success Criteria** (what must be TRUE):
  1. `go test -race -count=1 ./...` — вся батарея зелёная (41+ пакет, ≥3718 тестов, 0 FAIL), включая корпусный lint-паритет Gates A–D ×2 и golden 94 findings байт-точно
  2. D-07 позиции: `internal/protoast/position_test.go` и `internal/core/position_parity_test.go` зелёные; multibyte-UTF-8 фиксатура воспроизводит контракт main (1-based строка, 1-based колонка в рунах)
  3. `task lint:go`, `task schema:check` и `task build` зелёные; `schemas/` без дрейфа
  4. Итоговый diff приёмки не содержит изменений поведения: `easyp lint` на эталонном корпусе даёт byte-identical вывод против зафиксированного golden
**Plans:** 1/1 plans executed
- [x] 04-01-PLAN.md — Acceptance evidence: one full -race battery, all named parity gates, lint/schema/tidy predicates, and the honest v1.0 delta

**Acceptance note (v1.1, CONTEXT D-06):** после удаления вендорного дерева батарея — 34 тест-пакета вместо 41; 7 пакетов удалены вместе с вендорными тестами (дизайн, не регрессия). Поведенческий паритет доказывают golden 94 findings, Gates A–D ×2 и оба D-07 position-гейта, а не число тестов.

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Merge main и примирение путей | 2/2 | Complete    | 2026-11-21 |
| 2. Переключение потребителей и удаление вендор-копии | 2/2 | Complete    | 2026-11-21 |
| 3. Снятие seal-инфраструктуры и документация | 2/2 | Complete    | 2026-11-21 |
| 4. Приёмка — полный паритет | 1/1 | In Progress | - |

## Backlog

### Phase 999.1: Follow-up — Phase 03 deferred UAT follow-up: Test 1 (BACKLOG)

**Goal:** Resolve the UAT checkpoint deferred during Phase 03 verification
**Source phase:** 03
**Follow-ups:**
- [ ] Test 1: Add a disk-instrumentation cell in Phase 4 asserting readSource is invoked exactly once per corpus file on a compiling corpus (behavioral proof for the no-disk-reread guarantee, currently structural) (deferred 2026-11-19)

### Phase 999.2: Follow-up — Phase 03 deferred UAT follow-up: Test 2 (BACKLOG)

**Goal:** Resolve the UAT checkpoint deferred during Phase 03 verification
**Source phase:** 03
**Follow-ups:**
- [ ] Test 2: Standing constraint for Phase 4: at each rule-port batch confirm ported rules never consult comment directives or ignoreOnly themselves; keep central-filter cells green (deferred 2026-11-19)

### Corpora letters (LINT-PARITY-01 / LINT-PERF-01, from Phase 7)

- [ ] Unblock private corpora (GOOGLEAPIS fresh clone per Pitfall 7 + CORPUS_ARGS per `07-USER-SETUP.md`), fire `/tmp/p7/run-parity.sh parity` + perf protocol; verdict `median(new) <= median(main)` (deferred 2026-11-21, pending_user_unblock)

### Hadolint

- [ ] Run `task lint:docker` once Docker is available (Dockerfile untouched by v1.0) (deferred 2026-11-21)
