# Requirements: EasyP — import-easyprotoc

**Defined:** 2026-11-21
**Core Value:** EasyP остаётся drop-in buf-совместимым инструментом: формат вывода, коды выхода и конфигурация не ломаются между версиями — миграции внутренностей не должны быть заметны пользователю.

Контекст: репозиторий `easyp-tech/easyprotoc` (бывший приватный compiler-lab) стал публичным (2026-11-20, последняя правка main `f378f3f` от 2026-11-04). Вендор-копия `internal/easypc` побайтово совпадает с публичным `main` (проверено diff: `protoc/compile.go`, `parser/ast.go`, `protoc/syntax/syntax.go`, `protoc/resolver.go` — только import-пути и комментарии). На `main` EasyP коммит #253 уже потребляет `easyprotoc/protoc` в breaking-пути с той же псевдоверсией. Milestone отменяет решение v1.0 о вендоринге (REQ-1.2) и переводит lint-путь на модульную зависимость.

## v1 Requirements

Requirements for milestone «import-easyprotoc». Topological order: REC → DEP → CONS → VENDRM → SEAL → PARITY.

### Reconciliation

- [x] **IMPORT-REC-01**: WHEN ветка v1.1 готова к переключению импортов, the system SHALL содержать merge `origin/main` (#253 «Use compiler for breaking check») с семантически разрешёнными конфликтами `internal/core` (breaking-путь на `easyprotoc/protoc`, protobuf `v1.37.0-rc.1`); после слияния ветка не отстаёт от `origin/main` и полная батарея тестов зелёная

### Dependency

- [x] **IMPORT-DEP-01**: WHEN резолвятся зависимости модуля, the system SHALL требовать `github.com/easyp-tech/easyprotoc` в псевдоверсии main @ `f378f3f` (идентично main #253) без `replace`/`vendor`-директив; повторный `go mod tidy` не производит дрейфа в `go.mod`/`go.sum`

### Consumers

- [x] **IMPORT-CONS-01**: WHEN компилятор используется в `internal/core` (`link.go`, `link_test.go`, `lint_linked_matrix_test.go`) и `internal/protoast` (`build.go`, `outline.go`, `protoast.go`, `protoast_test.go`), the system SHALL импортировать публичные пакеты `github.com/easyp-tech/easyprotoc/protoc`, `.../protoc/syntax`, `.../parser`; пути `github.com/easyp-tech/easyp/internal/easypc/...` отсутствуют в кодовой базе (grep пуст)

### Vendor removal

- [x] **IMPORT-VENDRM-01**: WHEN репозиторий проверяется, the system SHALL не содержать дерево `internal/easypc` целиком (48+ файлов, включая `VENDORED.md`, `cmd/easyp-protoc`, `cmd/lexbench`); функция «protoc-совместимый бинарь» остаётся доступной через `go run github.com/easyp-tech/easyprotoc/cmd/easyprotoc` без локальной копии

### Seal / docs

- [x] **IMPORT-SEAL-01**: WHEN действующие конфиги и документация читаются, the system SHALL не содержать seal-инфраструктуры: `.golangci.yml`-исключение `internal/easypc` снято, параграф «Do not» в AGENTS.md заменён на описание потребления модуля, `.spec` (ARCHITECTURE/PACKAGES/CODE_STYLE/README и связанные) описывают импорт вместо вендоринга; исторические записи v1.0 в `.planning/milestones/` и `.spec/features/lint-easypc-infra/` сохраняются как история

### Parity

- [ ] **IMPORT-PARITY-01**: WHEN выполняется приёмка, the system SHALL пройти полный паритет без регрессий: `go build ./...`; `go test -race -count=1 ./...` зелёный (вся батарея v1.0, включая корпусный lint-паритет Gates A–D ×2, golden 94 findings байт-точно, D-07 position-тесты `internal/protoast/position_test.go` и `internal/core/position_parity_test.go`); `task lint:go` зелёный; `task schema:check` без дрейфа

## v2 Requirements

Deferred to future milestones:
- **Bump** псевдоверсии при появлении тегов у upstream или мерже ветки `feat/easyp-integration` (generation sources/descriptor metadata fix — меняет семантику ошибок `Link`, нужен codegen-серии FUTURE-COMPILER-01..03)
- **Hardening** (FUTURE-HARDENING-01): fuzzing лексера/парсера
- Corpora letters acceptance (обязательство v1.0, pending_user_unblock)

## Out of Scope

- Замена/переработка шва `internal/protoast` (дивергенция rune-колонок D-07 сохраняется как есть)
- Изменения API компилятора (в т.ч. взятие `feat/easyp-integration` до мержа upstream)
- Перевод generate/`mod tidy`/`migrate` на компилятор (breaking уже переведён на main #253; остальное — будущие фичи)
- Вкладка upstream (PR в easyprotoc) — отдельная будущая работа

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| IMPORT-REC-01 | Phase 1 — Merge main и примирение путей | Complete |
| IMPORT-DEP-01 | Phase 2 — Переключение потребителей и удаление вендор-копии | Complete |
| IMPORT-CONS-01 | Phase 2 — Переключение потребителей и удаление вендор-копии | Complete |
| IMPORT-VENDRM-01 | Phase 2 — Переключение потребителей и удаление вендор-копии | Complete |
| IMPORT-SEAL-01 | Phase 3 — Снятие seal-инфраструктуры и документация | Complete |
| IMPORT-PARITY-01 | Phase 4 — Приёмка — полный паритет | Pending |

**Coverage:** 6/6 v1 requirements mapped ✓
