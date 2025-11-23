# Миграция с protoc на EasyP

> Исходная англоязычная версия файла содержала только заголовок и пометку “Work in progress”. Ниже — полноценный русскоязычный черновик раздела миграции. Уточни, если требуется сократить или изменить уровень детализации.

## 1. Краткое сравнение

| Аспект | `protoc` | EasyP |
|--------|----------|-------|
| Базовая роль | Базовый компилятор `.proto` файлов | Инструмент “всё в одном”: линтинг, проверка несовместимых изменений, менеджер пакетов, генерация |
| Конфигурация | Чаще shell‑скрипты / Makefile / вручную передаваемые флаги | Единый файл `easyp.yaml` (YAML или JSON) |
| Управление зависимостями | Ручной `git clone`, `submodule`, vendoring | Секция `deps` с Git‑репозиториями |
| Набор правил линтера | Нет встроенного | Встроенный линтер с правилами (можно включать/исключать) |
| Проверка breaking changes | Отсутствует из коробки | Команда `easyp breaking` |
| Расширенные плагины | Установка и вызов через флаги `--plugin` | Плагины через конфиг; поддержка удалённого выполнения (API Service) и стандартного локального |
| Единообразие в CI | Нужно собирать вручную | Консистентные команды: `lint`, `breaking`, `generate` |

## 2. Когда имеет смысл мигрировать

Миграция на EasyP особенно полезна если:
- Есть потребность в автоматической проверке качества (стиль, именование) `.proto` файлов.
- Нужна защита от нелегитимных несовместимых изменений при эволюции схем.
- В проекте используется несколько языков генерации и растёт число скриптов вокруг `protoc`.
- Вы хотите единый конфиг вместо множества shell/Makefile комбинаций.
- Планируется централизованное исполнение плагинов (через API Service) или унификация окружения в CI.

## 3. Типичный сценарий “до” (только protoc)

Пример: локальные скрипты / Makefile:

```makefile
protoc \
  -I . \
  -I third_party \
  --go_out=./gen/go --go_opt=paths=source_relative \
  --go-grpc_out=./gen/go --go-grpc_opt=paths=source_relative \
  api/echo/v1/echo.proto
```

Дополнительные зависимости (googleapis, validate, grpc-gateway) скачиваются вручную или через submodules.

## 4. Типичный сценарий “после” (EasyP)

Единый файл `easyp.yaml`:

```yaml
version: v1alpha

deps:
  - github.com/googleapis/googleapis
  - github.com/envoyproxy/protoc-gen-validate

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
    - name: validate
      out: gen/go
      opts:
        paths: source_relative
        lang: "go"
```

Команды:
```bash
easyp mod update      # Скачивание/обновление deps
easyp lint            # Линтинг .proto
easyp breaking --against git-ref=origin/main
easyp generate        # Генерация кода
```

## 5. Пошаговая миграция

1. Инвентаризация:
   - Соберите список всех вызовов `protoc` (скрипты, Makefile, CI).
   - Зафиксируйте используемые плагины и их версии.
2. Создание `easyp.yaml`:
   - Добавьте `version: v1alpha`.
   - Перенесите плагины в секцию `generate.plugins`.
   - Укажите опции, которые ранее передавались в командной строке.
3. Зависимости:
   - Все внешние каталоги (googleapis, validate, grpc-gateway и т.п.) занесите в `deps`.
   - Удалите лишние submodules при необходимости.
4. Линтер:
   - Запустите `easyp lint` и посмотрите на предупреждения.
   - При необходимости добавьте `lint.use`, `lint.ignore`, `lint.except`.
5. Проверка breaking changes:
   - Выберите референс (обычно `origin/main` или предыдущий тег).
   - Запустите `easyp breaking --against git-ref=origin/main`.
   - Настройте игнорируемые изменения (`breaking.ignore`), если нужно.
6. Обновление CI:
   - Замените прямые вызовы `protoc` на `easyp generate`.
   - Добавьте отдельные шаги `lint` и `breaking`.
7. Очистка:
   - Удалите устаревшие генерационные скрипты, если они дублируют функционал EasyP.
8. Документация:
   - Обновите README / CONTRIBUTING (новый способ генерации).
9. Верификация:
   - Сравните результат генерации до и после (git diff).
   - Убедитесь, что имена файлов и пути совпадают или скорректируйте структуру.

## 6. Замена в CI (пример GitHub Actions)

```yaml
- name: Install EasyP
  run: go install github.com/easyp-tech/easyp/cmd/easyp@latest

- name: Lint
  run: easyp lint

- name: Breaking
  run: easyp breaking --against git-ref=origin/main

- name: Generate
  run: easyp generate
```

## 7. Работа с плагинами

| Тип | В protoc | В EasyP |
|-----|----------|---------|
| Стандартные (`protoc-gen-go`) | Указание через `--go_out` | `plugins: - name: go ...` |
| gRPC | `--go-grpc_out` | Второй плагин `go-grpc` |
| Validate | Доп. установка + `--validate_out` | `- name: validate` с opts |
| gRPC-Gateway / OpenAPI | `--grpc-gateway_out` / `--openapiv2_out` | Плагины `grpc-gateway`, `openapiv2` |
| Кастомные | Путь к бинарю / `--plugin=protoc-gen-xxx` | `- name: xxx` или удалённый `remote:` (если используется API Service) |

Если вы переходите на удалённое выполнение (API Service), записи в конфиге могут выглядеть так:

```yaml
generate:
  plugins:
    - remote: api.easyp.tech/protobuf/go:v1.36.10
      out: gen/go
      opts:
        paths: source_relative
```

## 8. Проверка корректности после миграции

1. `git diff` после `easyp generate` — не должно быть неожиданных изменений.
2. Сравнить размер и содержимое каталога сгенерированных файлов.
3. Запустить ваш потребляющий код (сборка, тесты).
4. Проверить, что плагины (validate, grpc-gateway) продолжают вставлять необходимые аннотации / опции.
5. Линт не должен выдавать критичных ошибок — скорректируйте стиль .proto, если требуется.

## 9. Частые проблемы

| Проблема | Причина | Решение |
|----------|---------|---------|
| Плагин не найден | Не установлен или не указан в PATH | Выполнить `go install ...` или добавить в конфиг как remote |
| Ошибка импорта `.proto` | Отсутствует репозиторий в `deps` | Добавить адрес в секцию `deps` и `easyp mod update` |
| Несовместимые изменения на каждом PR | Отсутствует игнор/настройка сравнения | Уточнить ветку/тег в `--against` или настроить `breaking.ignore` |
| Локально работает, в CI падает | В CI нет зависимостей (не вызван update) | Добавить шаг `easyp mod update` перед `generate` |
| Дубликаты сгенерированных файлов | Изменена структура путей или `out` | Синхронизировать `out` и `paths: source_relative` опции |

## 10. Минимальный пример конфигурации (только Go)

```yaml
version: v1alpha

generate:
  plugins:
    - name: go
      out: gen/go
      opts:
        paths: source_relative
```

Команды:
```bash
easyp generate
```

## 11. Расширенный пример (Go + gRPC + Gateway + Validate + Breaking и Lint)

```yaml
version: v1alpha

lint:
  use:
    - ENUM_FIRST_VALUE_ZERO
    - SERVICE_SUFFIX
  allow_comment_ignores: false
  enum_zero_value_suffix: _NONE
  service_suffix: Service

deps:
  - github.com/googleapis/googleapis
  - github.com/envoyproxy/protoc-gen-validate

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
    - name: validate
      out: gen/go
      opts:
        paths: source_relative
        lang: "go"
    - name: grpc-gateway
      out: gen/go
      opts:
        paths: source_relative
    - name: openapiv2
      out: gen/openapi
      opts:
        simple_operation_ids: true

breaking:
  against_git_ref: main
  ignore:
    - ENUM_VALUE_SAME_NAME
```

## 12. Checklist миграции

- [ ] Собраны все текущие вызовы `protoc`.
- [ ] Создан `easyp.yaml` с секциями `generate` и (опционально) `lint`, `breaking`.
- [ ] Добавлены все внешние зависимости в `deps`.
- [ ] Успешно выполнен `easyp mod update`.
- [ ] `easyp generate` даёт идентичный или ожидаемо обновлённый результат.
- [ ] `easyp lint` проходит без критичных ошибок (или ошибки обработаны).
- [ ] `easyp breaking --against git-ref=origin/main` не выдаёт неожиданных блокирующих нарушений.
- [ ] CI обновлён (заменены прямые вызовы `protoc`).
- [ ] Документация проекта обновлена (README: как генерировать).

## 13. Дополнительные рекомендации

- Сначала мигрируйте один сервис / пакет — снизит риск массовых конфликтов.
- Закоммитьте `easyp.yaml` отдельно, затем выполните генерацию и линт — проще ревью.
- Если нужны разные наборы плагинов для разных микросервисов — используйте несколько конфигов (например: `easyp.microservice-a.yaml` + `easyp.microservice-b.yaml`).
- Внедрите проверку актуальности генерации: шаг CI, который валидирует отсутствие diff после `easyp generate`.

## 14. Статус

Этот файл — русская расширенная версия, заменяющая placeholder “Work in progress”. При появлении новых возможностей EasyP (дополнительные поля в конфиге, новые плагины, расширения breaking‑проверок) обновите соответствующие разделы.

---

Если нужен упрощённый вариант (только перевод без расширения), сообщи — могу сократить до минимальных двух блоков.