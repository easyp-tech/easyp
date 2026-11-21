# Phase 1: Вендоринг компилятора и расчленение модели - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-11-17
**Phase:** 1-Вендоринг компилятора и расчленение модели
**Areas discussed:** Вендоринг хелперов, Гигиена вендор-дерева, Механика расщепления модели, Эталонная ревизия

---

## Вендоринг хелперов

**Q1: Переносить lint_helpers.go/casing.go в Phase 1 целиком или подмножеством?**

| Option | Description | Selected |
|--------|-------------|----------|
| Целиком (рекоменд.) | Копируем как есть; синк тривиален, diff в Phase 4 минимален; временный мёртвый код | ✓ |
| Только подмножество | Только используемое линтером; нет мёртвого кода, но Phase 4 доносит недостающее | |
| Перенести в Phase 4 | В Phase 1 только easypc + protoast | |

**User's choice:** Целиком
**Notes:** —

**Q2: Как обращаться с потенциально мёртвым кодом хелперов до Phase 4?**

| Option | Description | Selected |
|--------|-------------|----------|
| Исключить из линта | Точечные исключения в .golangci.yml до Phase 4 | |
| Не трогать (рекоменд.) | Экспортируемые функции не триггерят golangci-lint | ✓ |
| На усмотрение Claude | Planner решит по факту | |

**User's choice:** Не трогать
**Notes:** —

## Гигиена вендор-дерева

**Q3: Как фиксировать провенанс вендоренного компилятора?**

| Option | Description | Selected |
|--------|-------------|----------|
| VENDORED.md (рекоменд.) | Файл в internal/easypc: repo, ветка, SHA, дата, процедура обновления | ✓ |
| Только в коммите | SHA в коммит-сообщении и PROJECT.md | |
| Не фиксировать | История в git | |

**User's choice:** VENDORED.md
**Notes:** —

**Q4: Запретить ли правки внутри internal/easypc?**

| Option | Description | Selected |
|--------|-------------|----------|
| Запрет + синк (рекоменд.) | Правки только синком из compiler-lab; фиксы — вверх по потоку; исключение — импорт-пути | ✓ |
| Разрешить с пометкой | Точечные правки с пометкой локальных отклонений в VENDORED.md | |
| Без правил | Решать по ситуации | |

**User's choice:** Запрет + синк
**Notes:** Отражить в AGENTS.md и VENDORED.md

**Q5: Запускать ли собственные тесты easypc в task test?**

| Option | Description | Selected |
|--------|-------------|----------|
| Включить (рекоменд.) | Тесты easypc как обычные пакеты; testdata остаётся | ✓ |
| Исключить из test | Быстрее CI, но регрессии ловятся позже | |
| Тесты да, линт нет | Включить тесты, исключить из golangci-lint | |

**User's choice:** Включить
**Notes:** —

## Механика расщепления модели

**Q6: Как расщепить модель чтения между lint и breaking в Phase 1?**

| Option | Description | Selected |
|--------|-------------|----------|
| Общий alias | Issue.Position общий, alias на protoast.Position; меньше диффа, но связанность остаётся | |
| Полное расщепление (рекоменд.) | Lint получает собственные типы; breaking забирает legacy себе; чистая граница сразу | ✓ |
| На усмотрение Claude | Planner взвесит дифф по коду | |

**User's choice:** Полное расщепление
**Notes:** —

**Q7: Каким тестом доказывать контракт позиций?**

| Option | Description | Selected |
|--------|-------------|----------|
| Golden-паритет (рекоменд.) | Многобайтовый UTF-8 → старый линтер vs protoast, сквозной путь | |
| Юнит на конверсию | Только protoast.Position: спаны → 1-based/руны | |
| Оба уровня | Юнит + golden | ✓ |

**User's choice:** Оба уровня
**Notes:** —

## Эталонная ревизия

**Q8: Какую ревизию compiler-lab брать как эталон?**

| Option | Description | Selected |
|--------|-------------|----------|
| Фикс. SHA (рекоменд.) | Конкретный SHA easyp-cli в VENDORED.md до вкопирования; синки — осознанно | ✓ |
| HEAD на момент копии | Текущий HEAD, SHA пост-фактум | |
| На усмотрение | Planner решит | |

**User's choice:** Фикс. SHA
**Notes:** —

**Q9: Как executor получит доступ к compiler-lab?**

| Option | Description | Selected |
|--------|-------------|----------|
| gh с этой машины | gh repo clone, аутентификация работает | ✓ |
| Локальная копия | Копия на диске, указать путь | |
| Отметить риском | Доступа нет / нужны шаги | |

**User's choice:** gh с этой машины
**Notes:** Проверено explore-исследованием (2026-11-16)

---

## Claude's Discretion

- Механика переименования/расщепления типов в internal/core (имена, границы файлов)
- Выбор UTF-8 фикстур и их placement
- Формат VENDORED.md (минимально достаточный)

## Deferred Ideas

None — discussion stayed within phase scope
