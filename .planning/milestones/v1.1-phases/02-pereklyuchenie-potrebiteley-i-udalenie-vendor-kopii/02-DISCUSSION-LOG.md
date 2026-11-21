# Phase 2: Переключение потребителей и удаление вендор-копии - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-11-21
**Phase:** 2 — Переключение потребителей и удаление вендор-копии
**Areas discussed:** Маппинг импортов, Удаление вендор-копии, Protoast, Зависимости и tidy, Seal-инфраструктура, Ворота

---

## Маппинг импортов

| Option | Description | Selected |
|--------|-------------|----------|
| Прямые замены путей | 3 замены, без алиасов/шимов | ✓ |
| Алиасы `easypc = easyprotoc` | Скрывает миграцию, лишний слой | |
| Промежуточные пакеты-шимы | Bridge — антипаттерн (прецедент v1.0) | |

**User's choice:** [auto] прямые замены (recommended default)

---

## Удаление вендор-копии

| Option | Description | Selected |
|--------|-------------|----------|
| Удалить всё дерево в том же атомарном коммите | Build зелёный на коммите; rollback = revert | ✓ |
| Удалять по частям | Промежуточные состояния с полу-деревом | |

**User's choice:** [auto] атомарное удаление (recommended default)

---

## Protoast

| Option | Description | Selected |
|--------|-------------|----------|
| Менять только импорты | Логика и rune-колонки без изменений; D-07 гейты | ✓ |
| Переписать protoast на модуль | Выходит за границу фазы, ломает D-07 | |

**User's choice:** [auto] только импорты (recommended default)

---

## Зависимости и tidy

| Option | Description | Selected |
|--------|-------------|----------|
| Сохранить easyprotoc прямым, tidy -diff пуст | Пины уже на месте с merge | ✓ |
| go get / bump | Не нужен, риск дрейфа | |

**User's choice:** [auto] сохранить как есть (recommended default)

---

## Seal-инфраструктура

| Option | Description | Selected |
|--------|-------------|----------|
| Не трогать в Phase 2 | Владелец — Phase 3; устаревшее исключение безвредно | ✓ |
| Удалить сейчас | Смешивает фазы, срывает атомарность | |

**User's choice:** [auto] не трогать (recommended default)

---

## Ворота

| Option | Description | Selected |
|--------|-------------|----------|
| Полная батарея + named parity + tidy + grep | Как D-09 ROADMAP/Phase 2 SC | ✓ |
| Только build | Не доказало бы отсутствие регрессий | |

**User's choice:** [auto] полная батарея (recommended default)

---

## Claude's Discretion

Порядок правок внутри коммита, формулировки сообщений, комментарий в `protoast.go`.

## Deferred Ideas

- Seal/docs — Phase 3.
- Bump easyprotoc — v2.
