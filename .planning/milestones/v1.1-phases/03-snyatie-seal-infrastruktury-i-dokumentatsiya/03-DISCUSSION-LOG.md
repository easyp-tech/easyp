# Phase 3: Снятие seal-инфраструктуры и документация - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-11-21
**Phase:** 3 — Снятие seal-инфраструктуры и документация
**Areas discussed:** .golangci.yml, AGENTS.md, .spec, Doc-citations Phase 2, Errcheck, Ворота

---

## .golangci.yml

| Option | Description | Selected |
|--------|-------------|----------|
| Удалить исключение internal/easypc + комментарий | Путь не существует; исключение мертво | ✓ |
| Оставить | Мёртвая конфигурация, вводит в заблуждение | |

**User's choice:** [auto] удалить (recommended default)

---

## AGENTS.md

| Option | Description | Selected |
|--------|-------------|----------|
| Заменить seal-параграф правилом про модуль + строка internal/protoast в Layout | Актуальная модель: пин, upstream-фиксы, без копирования | ✓ |
| Удалить bullet без замены | Потерялось бы правило «не копировать код модуля» | |

**User's choice:** [auto] заменить с правилом (recommended default)

---

## .spec

| Option | Description | Selected |
|--------|-------------|----------|
| Добавить описание модульной зависимости и шва protoast; чинить только фактически неверное | Минимальные точные правки | ✓ |
| Полный рерайт .spec | Вне границы фазы, риск дрейфа | |

**User's choice:** [auto] минимальные точные правки (recommended default)

---

## Doc-citations Phase 2 (IN-01..IN-03)

| Option | Description | Selected |
|--------|-------------|----------|
| Закрыть в docs-пасе этой фазы | Естественная принадлежность | ✓ |
| Оставить open | Дрейф документации | |

**User's choice:** [auto] закрыть (recommended default)

---

## Errcheck

| Option | Description | Selected |
|--------|-------------|----------|
| Исправить session_test.go:29 через t.Cleanup | Делает lint:go зелёным (SC1) | ✓ |
| Игнорировать | SC1 недостижим | |

**User's choice:** [auto] исправить (recommended default)

---

## Ворота

| Option | Description | Selected |
|--------|-------------|----------|
| build + golangci 0 issues + targeted -race + grep-empty + tidy | Полная батарея — Phase 4 | ✓ |
| Полная батарея здесь | Дублирование Phase 4 | |

**User's choice:** [auto] targeted-ворота (recommended default)

---

## Claude's Discretion

Формулировки, порядок, разбиение коммитов в рамках D-10.

## Deferred Ideas

- Полная батарея/паритет — Phase 4.
- Bump easyprotoc — v2.
