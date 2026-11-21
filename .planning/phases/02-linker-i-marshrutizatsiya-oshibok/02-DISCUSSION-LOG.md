# Phase 2: Линкер и маршрутизация ошибок - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-11-19
**Phase:** 2-Линкер и маршрутизация ошибок
**Areas discussed:** Детекция отсутствующего импорта; Гранулярность отката; Форма COMPILE-замечаний; Размещение классификатора; Верификация через путь Action
**Mode:** `--auto` (manager-сессия; пользователь дал директиву «продолжай автоматически, если от пользователя не требуется решений»)

---

## Детекция отсутствующего импорта

| Option | Description | Selected |
|--------|-------------|----------|
| Typed resolver seam | Собственный Resolver в protoc.Link: «не найдено» → типизированная ошибка с путём импорта; текст диагностик не парсится | ✓ |
| String matching | Разбор текста диагностики `Import "..." was not found or had errors.` | |
| Pre-resolve imports | Предварительный обход импортов перед линковкой | |

**User's choice:** [auto] typed resolver seam (recommended default)
**Notes:** Research flag ROADMAP предписывает resolver через importRoots + importFileAllowed + wellknownimports — typed seam этому соответствует; `protoc.ErrFileNotFound` уже существует как типизированный sentinel.

---

## Гранулярность отката

| Option | Description | Selected |
|--------|-------------|----------|
| Corpus-level fallback | Один отсутствующий импорт → весь корпус разбирается пофайлово, без COMPILE | ✓ |
| Per-file fallback | Откат только для файлов с отсутствующим импортом | |

**User's choice:** [auto] corpus-level fallback (recommended default)
**Notes:** Буква LINT-ENG-03 («вернуться к разбору каждого файла по отдельности») и лабораторное поведение; `OpenImportFileError` сохраняет маршрут exit-кода 2.

---

## Форма COMPILE-замечаний

| Option | Description | Selected |
|--------|-------------|----------|
| One issue per diagnostic | rule=COMPILE, путь/строка/колонка/текст из Diagnostic | ✓ |
| Aggregated per file | Одна issue на файл со склеенным текстом | |

**User's choice:** [auto] one issue per diagnostic (recommended default)
**Notes:** Формат вывода фиксирован LINT-COMPAT-01 (Phase 5); позиции компилятора 1-based — контракт Phase 1 D-07; порядок выдачи детерминированный (сортировка в классификаторе), паритет порядка — на фиксатурах.

---

## Размещение классификатора

| Option | Description | Selected |
|--------|-------------|----------|
| internal/core | Рядом с lint-путём; edits в internal/easypc запрещены (D-04 фазы 1) | ✓ |
| internal/easypc | Запрещено seal'ом фазы 1 | |

**User's choice:** [auto] internal/core (recommended default)
**Notes:** До Phase 5 линкер dormant — текущий `Lint()` не меняется.

---

## Верификация через путь Action

| Option | Description | Selected |
|--------|-------------|----------|
| Action-path tests | Фиксатурная матрица трёх классов через код маршрутизации internal/api/lint.go | ✓ |
| Unit-only | Только юнит-тесты классификатора без exit-кодов | |

**User's choice:** [auto] Action-path tests (recommended default)
**Notes:** Success criterion 4 ROADMAP: коды выхода CLI проверены через путь Action до врезки движка.

---

## Claude's Discretion

- Имена файлов/типов классификатора, сигнатура API
- Структура фиксатур (testdata трёх классов ошибок)
- Связка «резолвер → import roots»
- Выделение diagnostic→issue в отдельную утилиту или встраивание

## Deferred Ideas

None — discussion stayed within phase scope
