# Phase 3: Двухпроходный движок и подавления - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.

**Date:** 2026-11-19
**Phase:** 3-Двухпроходный движок и подавления
**Areas discussed:** Детерминизм вывода; Параллелизм; ModuleRule; Подавления; Валидация конфигурации; Модель чтения движка
**Mode:** `--auto` (manager-цепочка; директива пользователя «продолжай автоматически»)

---

## Детерминизм вывода

| Option | Description | Selected |
|--------|-------------|----------|
| Полный детерминизм (path, rule, line, col) | Тотальная сортировка на выходе движка | ✓ |
| Паритет недетерминизма main | Воспроизводить map-iteration порядок | |

**User's choice:** [auto] полный детерминизм (recommended default)
**Notes:** Усиление, не ломающее контракт — порядок main не является контрактом (research Phase 1); Open Design Question из REQUIREMENTS.md закрыт.

## Параллелизм

| Option | Description | Selected |
|--------|-------------|----------|
| internal/parallel + окна эталона | 256 файлов/окно, NumCPU воркеров, порог малого корпуса | ✓ |
| Последовательный движок | Без параллелизма | |

**User's choice:** [auto] эталонная схема (recommended default)

## ModuleRule

| Option | Description | Selected |
|--------|-------------|----------|
| Порт Collect/Finish контракта | Для PACKAGE_* кросс-файловых правил | ✓ |
| Всё per-file | Кросс-файловые проверки иначе | |

**User's choice:** [auto] порт контракта (recommended default)

## Подавления

| Option | Description | Selected |
|--------|-------------|----------|
| Центральный фильтр, семантика main | ignore/ignore_only + комментарии; walk на protoast-модели | ✓ |
| Подавления как в лаборатории | Семантика эталона | |

**User's choice:** [auto] семантика main (recommended default — REQ-3.2)
**Notes:** Двухпроходная схема: сбор подавлений per-file в первом проходе, применение во втором.

## Валидация конфигурации

| Option | Description | Selected |
|--------|-------------|----------|
| В internal/core при конструировании | Имена из реестра builder'а, паритет текста ошибки | ✓ |
| В config/v1 | На уровне парсинга конфига | |

**User's choice:** [auto] core при конструировании (recommended default)

## Модель чтения движка

| Option | Description | Selected |
|--------|-------------|----------|
| Из Linked.Syntax один раз на файл | Без повторного чтения диска; parseOnTheirOwn для откатившихся | ✓ |
| Повторное чтение диска | Простота vs производительность | |

**User's choice:** [auto] из деревьев (recommended default)

## Claude's Discretion

- Имена/границы файлов движка; структура checkedFile; формат таблицы паритет-тестов

## Deferred Ideas

None — discussion stayed within phase scope
