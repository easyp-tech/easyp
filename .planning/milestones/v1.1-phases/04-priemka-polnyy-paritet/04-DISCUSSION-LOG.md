# Phase 4: Приёмка — полный паритет - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-11-21
**Phase:** 4 — Приёмка — полный паритет
**Areas discussed:** Режим фазы, Батарея, Гейты, Schema drift, Сравнение с v1.0, Публикация

---

## Режим фазы

| Option | Description | Selected |
|--------|-------------|----------|
| Evidence-only | Никаких изменений кода; только SUMMARY/evidence | ✓ |
| Разрешить фиксы по ходу | Маскирует регрессии, ломает атрибуцию | |

**User's choice:** [auto] evidence-only (recommended default)

## Батарея

| Option | Description | Selected |
|--------|-------------|----------|
| Одна `-race -v` батарея | 34 ok / 0 FAIL / известный skip | ✓ |
| Повторные прогоны «до зелёного» | Маскирует флаки/регрессии | |

**User's choice:** [auto] одна батарея (recommended default)

## Гейты

| Option | Description | Selected |
|--------|-------------|----------|
| RUN+PASS по 11 именам | Gates A–D, golden, D-07, матрицы | ✓ |
| Только exit code | Вакуумные прогоны не ловятся | |

**User's choice:** [auto] exact RUN+PASS (recommended default)

## Schema drift

| Option | Description | Selected |
|--------|-------------|----------|
| schema-gen + zero diff | Схемы не менялись в v1.1 | ✓ |
| Пропустить | Дрейф остался бы незамеченным | |

**User's choice:** [auto] schema-gen (recommended default)

## Сравнение с v1.0

| Option | Description | Selected |
|--------|-------------|----------|
| Зафиксировать 41→34 как дизайн | Удалены вендорные тесты; паритет поведенческий | ✓ |
| Считать регрессией | Неверно: тесты удалены вместе с деревом | |

**User's choice:** [auto] зафиксировать дельту (recommended default)

## Публикация

| Option | Description | Selected |
|--------|-------------|----------|
| Push/PR — после фазы, UAT-гейт | Внешнее действие требует решения пользователя | ✓ |
| Push в фазе | Нарушает границу и требует согласия | |

**User's choice:** [auto] после фазы (recommended default)

---

## Claude's Discretion

Порядок прогонов, evidence-файлы, формулировки.

## Deferred Ideas

- Push/PR, milestone-close — UAT-гейт.
- Corpora letters, bump easyprotoc — backlog/v2.
