# Phase 1: Merge main и примирение путей - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-11-21
**Phase:** 1 — Merge main и примирение путей
**Areas discussed:** Стиль слияния, Семантика разрешения конфликтов, Зависимости, Дормация вендор-пути, Ветка и коммиты, Ворота фазы

---

## Стиль слияния

| Option | Description | Selected |
|--------|-------------|----------|
| Merge-коммит | Сохраняет SHA и историю открытого PR #258; последующий мерж тривиален | ✓ |
| Rebase на origin/main | Линейная история, но переписывает 40+ коммитов открытого PR | |
| Squash | Неприменимо для поддержки ветки | |

**User's choice:** [auto] Merge-коммит (recommended default)
**Notes:** PR #258 открыт и не вмержён; rebase сломал бы review. Один merge-коммит, конфликты — внутри него.

---

## Семантика разрешения конфликтов

| Option | Description | Selected |
|--------|-------------|----------|
| main wins для breaking/sourceview, ветка wins для lint-контура | Семантический merge: оба пути сосуществуют, lint не трогается до Phase 2 | ✓ |
| Ветка wins везде | Отбросило бы #253 (breaking на easyprotoc) — противоречит цели милстоуна | |
| Отложить merge до Phase 2 | Смешало бы два независимых риска в одной фазе | |

**User's choice:** [auto] main wins для breaking/sourceview, ветка wins для lint (recommended default)
**Notes:** Прецедент REQ-3.2 v1.0: поведение main — истина. Падение golden/parity после merge = ошибка разрешения, не новая норма.

---

## Зависимости

| Option | Description | Selected |
|--------|-------------|----------|
| Принять main-пин как есть | easyprotoc @ f378f3f + protobuf v1.37.0-rc.1 (MVS) | ✓ |
| Понизить protobuf до v1.36.11 | Конфликт с требованием easyprotoc; лишний риск | |
| replace-директива | Запрещён конвенцией; маскирует реальный MVS | |

**User's choice:** [auto] Принять main-пин как есть (recommended default)
**Notes:** Тот же пин переиспользуется в Phase 2.

---

## Дормация вендор-пути

| Option | Description | Selected |
|--------|-------------|----------|
| internal/easypc byte-frozen в Phase 1 | Merge не трогает вендор-дерево и его импорты; переключение — Phase 2 | ✓ |
| Совместить merge и переключение | Два риска в одной фазе; конфликты стало бы невозможно атрибутировать | |

**User's choice:** [auto] byte-frozen (recommended default)
**Notes:** Phase 1 доказывает, что merge не сломал текущий lint-путь.

---

## Ветка и коммиты

| Option | Description | Selected |
|--------|-------------|----------|
| feat/lint-easypc-infra (текущая) | GSD branching_strategy: none; PR #258 расширяется | ✓ |
| Новая стековая ветка | Откладывает связывание с PR #258 | |

**User's choice:** [auto] текущая ветка (recommended default)
**Notes:** Зафиксировано в STATE.md как блокер v1.1.

---

## Ворота фазы

| Option | Description | Selected |
|--------|-------------|----------|
| Полная батарея + parity/golden | go build, go test -race ./..., Gates A–D ×2, golden 94 байт-точно | ✓ |
| Только go build | Не доказало бы отсутствие регрессий | |

**User's choice:** [auto] полная батарея (recommended default)

---

## Claude's Discretion

Конкретные hunks разрешения конфликтов, порядок прогонов, формулировка сообщения merge-коммита.

## Deferred Ideas

- `feat/easyp-integration` (generation metadata fix) — не в v1.1.
- Bump easyprotoc при появлении тегов — v2.
