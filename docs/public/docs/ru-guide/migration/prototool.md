# Prototool

## Состояние
Этот раздел был помечен как “Work in progress” в оригинальной английской версии. Ниже — переведённое и дополненное практическое руководство по миграции с Prototool на EasyP. Имя инструмента Prototool не переводится.

---

## Цель миграции
Перейти от конфигурации и рабочих процессов Prototool к эквиваленту на EasyP с минимальными изменениями, сохранив:
- Линтинг .proto файлов.
- Генерацию кода через protoc-плагины.
- Управление зависимостями.
- Проверку несовместимых изменений (в Prototool это приходилось реализовывать внешними инструментами; в EasyP встроено).

---

## Кратко о различиях

| Область | Prototool | EasyP | Комментарий |
|---------|-----------|-------|-------------|
| Конфигурация | `prototool.yaml` | `easyp.yaml` | Форматы YAML/JSON поддерживаются; структура отличается. |
| Линтинг | Встроенные наборы + кастомные правила | Наборы через `lint.use`, тонкая настройка ignore/except | EasyP даёт больше гибкости группировки. |
| Генерация | Раздел `generate` + плагины через Docker/локально | `generate.plugins` (локальные/WASM/удалённые) | EasyP может выполнять плагины удалённо (API Service). |
| Зависимости | Обычно через include-пути / внешние git clone вручную | Секция `deps` (Git репозитории) + `mod update` | Более декларативно. |
| Breaking Changes | Нет встроенного механизма | Секция `breaking` + команда `easyp breaking` | Из коробки. |
| Lock-in | Зависимость от модели Prototool | Минималистичный формат, «любой Git — пакет» | Облегчает выход / интеграцию с другими экосистемами. |

---

## Типовая структура Prototool (упрощённо)

```yaml
# prototool.yaml (пример)
protoc:
  include_paths:
    - api
    - third_party
lint:
  group: basic
  rules:
    add:
      - ENUM_VALUE_PREFIX
generate:
  plugins:
    - name: go
      out: gen/go
    - name: go-grpc
      out: gen/go
```

Особенности:
- `include_paths` задаёт пути поиска импортов.
- `lint.group` + `rules.add/remove` регулируют активные проверки.
- `generate.plugins` описывает плагины для protoc.

---

## Эквивалент на EasyP

```yaml
# easyp.yaml (пример)
version: v1alpha

deps:
  - github.com/org/common-protos         # вместо manual include
  - github.com/org/third-party-protos

lint:
  use:
    - MINIMAL
    - BASIC
  enum_zero_value_suffix: "_NONE"
  service_suffix: "Service"
  ignore:
    - COMMENT_FIELD
  except:
    - ENUM_VALUE_PREFIX
  allow_comment_ignores: true

generate:
  plugins:
    - name: go
      out: gen/go
      opts:
        paths: source_relative
    - name: go-grpc
      out: gen/go
      opts:
        paths: source_relative
        require_unimplemented_servers: false

breaking:
  against_git_ref: main
  ignore:
    - FIELD_SAME_TYPE
```

Комментарии:
- `deps` заменяет необходимость вручную прописывать include paths: EasyP подтянет их через git.
- `lint.use` активирует наборы правил (MINIMAL/BASIC/DEFAULT и т.д.).
- Дополнительные настройки (`enum_zero_value_suffix`, `service_suffix`) задают политики нейминга.
- `breaking` позволяет сразу контролировать несовместимые изменения.

---

## Шаги миграции

1. Инвентаризация:
   - Соберите текущий `prototool.yaml`, списки include paths и плагины.
   - Зафиксируйте используемые кастомные правила, если есть.

2. Зависимости:
   - Любые каталоги, которые вы вручную добавляли в `include_paths`, преобразуйте в git-репозитории (если ещё нет).
   - Добавьте их в `deps` (формат: `github.com/owner/repo` или путь к вашей приватной системе Git).

3. Линтинг:
   - Определите какой group использовался (`basic`, `default`, …) → сопоставьте с наборами EasyP (`MINIMAL`, `BASIC`, `DEFAULT`).
   - Индивидуальные добавления (`rules.add`) перенесите в `lint.use` или `except`/`ignore`:
     - Если правило нужно активировать — включите через набор или явно.
     - Если нужно отключить — добавьте в `ignore`.

4. Генерация:
   - Сохраните список плагинов. В большинстве случаев имена совпадут (`go`, `go-grpc`, др.).
   - Если раньше использовали Docker-окружение Prototool — проверьте поддержку удалённых плагинов через API Service EasyP (опционально).
   - Перепишите в секцию `generate.plugins`.

5. Дополнительные опции:
   - Настройки нейминга enum/service перенести в `enum_zero_value_suffix`, `service_suffix`.
   - Если были внешние скрипты для проверки breaking changes — удалите их и включите `breaking` с `against_git_ref`.

6. Верификация:
   - Запустите: `easyp mod update` (скачивание deps).
   - Затем: `easyp lint`.
   - Далее: `easyp generate`.
   - И, если нужно: `easyp breaking --against git-ref=origin/main`.

7. CI/CD:
   - Обновите workflow (GitHub Actions / GitLab) — замените вызовы Prototool на EasyP команды.
   - Добавьте шаг проверки несовместимых изменений.

---

## Сопоставление правил (общая идея)

| Prototool концепт | EasyP |
|-------------------|------|
| lint.group        | lint.use (наборы) |
| rules.add/remove  | except / ignore |
| include_paths     | deps (git источники) |
| plugins           | generate.plugins |
| (внешние скрипты для breaking) | breaking секция + команда |

---

## Приватные репозитории

Если ранее Prototool просто видел локальные каталоги, теперь:
- Добавьте SSH ключи/токены в CI для доступа к приватным Git.
- Укажите их в `deps` как обычные git-URL (формат зависит от реализации EasyP; для публичных — стандартный).

---

## Типовые проблемы при миграции

| Проблема | Причина | Решение |
|----------|---------|---------|
| Правила линтера не совпадают | Неверное сопоставление групп | Проверить документацию EasyP по наборам правил |
| Импорты не находятся | Забыли добавить репозиторий в deps | Добавить в `deps` и выполнить `easyp mod update` |
| Генерация падает на опциях | Плагин требует другие flags | Проверить `opts` и документацию плагина |
| breaking сообщает слишком много | Слишком строгий базовый набор | Добавить конкретные правила в `breaking.ignore` |
| Долгая генерация | Нет кеша/много deps | Включить кэширование в CI (vendor/.easyp/deps) |

---

## Минимальная конфигурация после миграции

```yaml
version: v1alpha
lint:
  use:
    - MINIMAL
generate:
  plugins:
    - name: go
      out: gen/go
      opts:
        paths: source_relative
breaking:
  against_git_ref: main
```

---

## Рекомендации по упрощению

- Начните с минимального `lint.use` (MINIMAL), затем постепенно включайте дополнительные наборы.
- Не переносите слепо все ignore — часть правил может уже быть улучшена в EasyP.
- Сначала настроить deps и генерацию, затем усложнять линт/нейминг.
- Фиксируйте миграцию по шагам (коммиты: deps → lint → generate → breaking → CI).

---

## Проверка результата

Скрипт локальной быстрой проверки (пример):

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "[1] deps update"
easyp mod update

echo "[2] lint"
easyp lint

echo "[3] breaking vs main"
easyp breaking --against git-ref=origin/main || echo "Breaking changes detected (review needed)"

echo "[4] generate"
easyp generate

echo "Done."
```

---

## Заключение

Миграция с Prototool на EasyP сводится к:
1. Переносу путей и внешних каталогов в декларативные `deps`.
2. Переформатированию линта под `lint.use` + `ignore/except`.
3. Переносу генерации в `generate.plugins`.
4. Добавлению встроенного механизма `breaking`.
5. Адаптации CI.

После этого конфигурация становится более прозрачной, а проверка несовместимых изменений — из коробки.

---

*Готово: базовый перевод и расширенное руководство по миграции с Prototool на EasyP.*