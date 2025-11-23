# GitHub Actions

Полное руководство по интеграции EasyP в конвейеры GitHub Actions для линтинга, проверки несовместимых изменений, генерации кода и автоматизации релизов.

## Зачем использовать EasyP в CI

Встраивание EasyP в GitHub Actions даёт:
- Единообразие правил линтинга и генерации между разработчиками.
- Автоматическое выявление несовместимых изменений (breaking changes) до мержа.
- Гарантию, что сгенерированный код не «устарел» относительно .proto.
- Возможность централизованной проверки зависимостей и версии конфигурации.

## Предварительные требования

1. Репозиторий содержит файл конфигурации `easyp.yaml` (или находится в корне, если запускается без `-cfg`).
2. Установлены / определены зависимости в секции `deps` (если нужны внешние импорты).
3. Для приватных Git-репозиториев настроен доступ (SSH key / deploy key / токен).
4. В workflow есть шаги по установке Go (если EasyP поставляется через `go install`) или заранее подготовленный бинарь.

## Базовый пример (быстрый старт)

Минимальный workflow для проверки линта и несовместимых изменений при открытия Pull Request:

```yaml
name: easyp-ci

on:
  pull_request:
    paths:
      - '**/*.proto'
      - 'easyp.yaml'

jobs:
  lint-and-breaking:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install EasyP
        run: go install github.com/easyp-tech/easyp/cmd/easyp@latest

      - name: Verify EasyP version
        run: |
          which easyp
          easyp version

      - name: Lint
        run: easyp lint

      - name: Breaking changes check (against main)
        run: easyp breaking --against git-ref=origin/main
```

## Установка EasyP

Варианты:
1. Через `go install` (как в примере выше).
2. Использование заранее собранного бинаря (артефакт релиза).

```yaml
- name: Download EasyP binary
  run: |
    curl -sL https://github.com/easyp-tech/easyp/releases/download/v1.0.0/easyp_linux_amd64 -o /usr/local/bin/easyp
    chmod +x /usr/local/bin/easyp
```

## Кэширование

Для ускорения повторных сборок можно кэшировать:
- Go модульный кеш (если используете генерацию с плагинами Go).
- Внешние зависимые Git-клонирования (папки `vendor` или содержимое `deps` при использовании `easyp mod download` / `easyp mod update`).
- Сгенерированный код (если он тяжёлый, но чаще генерируется заново).

Пример кэширования Go:

```yaml
- name: Cache Go build
  uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: go-${{ runner.os }}-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      go-${{ runner.os }}-
```

Пример кэширования каталога зависимостей EasyP (если вы используете локальный vendor-подход):

```yaml
- name: Cache proto deps
  uses: actions/cache@v4
  with:
    path: .easyp/deps
    key: deps-${{ runner.os }}-${{ hashFiles('easyp.yaml') }}
    restore-keys: |
      deps-${{ runner.os }}-
```

(Создайте `.easyp/deps` как целевой каталог, если ваша стратегия скачивания такова.)

## Линтинг

Команда:

```bash
easyp lint
```

Рекомендуется:
- В PR: всегда запускать.
- На main: можно запускать для верификации перед релизом.

Расширенный пример с выводом подробностей и игнорированием комментариев (если включено в конфиг):

```yaml
- name: Lint proto files
  run: |
    set -e
    easyp lint -v
```

## Проверка несовместимых изменений (Breaking Changes)

EasyP может сравнивать текущее состояние против ветки / коммита / тега.

Примеры:
```bash
# Сравнить против origin/main
easyp breaking --against git-ref=origin/main

# Сравнить против тега релиза
easyp breaking --against git-ref=v1.5.2

# Игнорировать определённые правила (если у вас настроено в easyp.yaml)
easyp breaking
```

Workflow шаг:

```yaml
- name: Breaking changes check
  run: easyp breaking --against git-ref=origin/main
```

Если нужно «мягко» пропустить (например, выводить предупреждения, но не падать билд):

```yaml
- name: Breaking changes (soft)
  run: |
    if ! easyp breaking --against git-ref=origin/main; then
      echo "::warning title=Breaking Changes Detected::Обнаружены потенциально несовместимые изменения"
    fi
```

## Генерация кода

Обычно для генерации:
```bash
easyp generate
```

Если нужно явно указать конфиг:
```bash
easyp -cfg easyp.yaml generate
```

Пример шага:

```yaml
- name: Generate code
  run: |
    easyp generate
    git status --short
```

Можно добавить проверку «не осталось ли не закоммиченных изменений» (гарантия, что сгенерированный код актуален):

```yaml
- name: Ensure generated code is up-to-date
  run: |
    easyp generate
    if [ -n "$(git status --porcelain)" ]; then
      echo "Generated code is outdated. Please commit changes."
      git diff --name-only
      exit 1
    fi
```

## Матрица (Matrix strategy)

Если у вас несколько вариантов конфигураций или разных языков генерации (через плагины), можно использовать matrix:

```yaml
jobs:
  generate:
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        lang: [go, python]
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go (only for go)
        if: matrix.lang == 'go'
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install EasyP
        run: go install github.com/easyp-tech/easyp/cmd/easyp@latest

      - name: Generate
        run: |
          if [ "${{ matrix.lang }}" = "go" ]; then
            easyp generate
          elif [ "${{ matrix.lang }}" = "python" ]; then
            easyp generate --plugin python
          fi
```

## Работа с приватными репозиториями (deps)

Если вы указываете приватные Git-пакеты в секции `deps`:

1. Настройте deploy key или используйте `GIT_SSH_COMMAND` с токеном.
2. Добавьте known_hosts (либо установите `StrictHostKeyChecking=no` для внутренних систем, но осознанно).

Пример:

```yaml
- name: Prepare SSH
  run: |
    mkdir -p ~/.ssh
    echo "${{ secrets.DEPLOY_KEY }}" > ~/.ssh/id_rsa
    chmod 600 ~/.ssh/id_rsa
    ssh-keyscan github.com >> ~/.ssh/known_hosts

- name: Update deps
  run: easyp mod update
```

## Оптимизация времени сборки

- Используйте кэш для Go / deps.
- Запускайте линт/брейкинг только если изменились `.proto` или `easyp.yaml` (paths фильтр).
- Разделяйте workflow: быстрый линт в PR, тяжелая генерация в push на main.
- Выносите генерацию OpenAPI / gRPC-Gateway в отдельный job (параллелизм).

## Продвинутый workflow (многостадийный)

```yaml
name: easyp-pipeline

on:
  pull_request:
    paths:
      - '**/*.proto'
      - 'easyp.yaml'
  push:
    branches:
      - main

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install EasyP
        run: go install github.com/easyp-tech/easyp/cmd/easyp@latest

      - name: Lint
        run: easyp lint

      - name: Breaking check (against main)
        if: github.event_name == 'pull_request'
        run: easyp breaking --against git-ref=origin/main

  generate:
    needs: validate
    if: github.event_name == 'push'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Install EasyP
        run: go install github.com/easyp-tech/easyp/cmd/easyp@latest
      - name: Generate Code
        run: easyp generate
      - name: Check diffs
        run: |
          if [ -n "$(git status --porcelain)" ]; then
            echo "Generated files changed after generate step."
            git diff --name-only
            exit 1
          fi
```

## Автоматизация релиза (пример)

Можно добавить workflow, который по тегу создаёт релиз и валидирует proto:

```yaml
name: release-verify

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  release-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install EasyP
        run: go install github.com/easyp-tech/easyp/cmd/easyp@latest

      - name: Lint
        run: easyp lint

      - name: Breaking vs previous tag
        run: |
          PREV_TAG=$(git tag --sort=-creatordate | grep -v "${GITHUB_REF_NAME}" | head -n1 || true)
          if [ -n "$PREV_TAG" ]; then
            echo "Comparing against $PREV_TAG"
            easyp breaking --against git-ref=$PREV_TAG
          else
            echo "No previous tag, skipping breaking check"
          fi

      - name: Generate to ensure consistency
        run: easyp generate
```

## Отладка

Если нужно больше логов:
```bash
easyp lint -v
easyp breaking -v --against git-ref=origin/main
easyp generate -v
```

Добавьте вывод окружения для диагностики:
```yaml
- name: Debug environment
  run: |
    env | sort
    easyp version
```

## Переменные окружения (примерно)

| Переменная | Назначение |
|------------|-----------|
| `EASYP_SERVICE_HOST` | Хост удалённого API Service (если используете удалённые плагины) |
| `EASYP_SERVICE_GRPC_PORT` | Порт удалённого сервиса |
| `GIT_SSH_COMMAND` | Кастомная команда SSH (частные репозитории) |
| `GOFLAGS` | Можно выставить `-mod=readonly` для Go генерации |

(Используйте только те, что действительно нужны вашему проекту.)

## Частые ошибки

| Симптом | Причина | Решение |
|---------|---------|---------|
| Линт падает на неизвестных правилах | Правила указаны в `easyp.yaml`, но версия EasyP их не поддерживает | Обновить EasyP (`go install ...@latest`) |
| `breaking` сравнение ничего не находит | Неверный `--against` (нет fetch удалённой ветки) | Добавить шаг `git fetch --all --tags` |
| Сгенерированный код в PR отличается от main | Отсутствует проверка генерации в CI | Добавить шаг «Ensure generated code is up-to-date» |
| Приватные deps не скачиваются | Нет SSH ключа или токена | Настроить secrets + шаг подготовки SSH |
| Долгое выполнение | Отсутствует кеш | Добавить кэширование модулей и deps |

## Рекомендации по поддержке

- Заготовить шаблон workflow и копировать между сервисами (DRY).
- Включить `paths` фильтры, чтобы не тратить минуты на нерелевантные изменения.
- Проверять `easyp.yaml` при изменениях отдельно — можно запускать «конфигурационный тест» (например, `easyp lint` без файлов, чтобы убедиться, что конфиг валиден).
- При появлении новых правил линтера обновляйте конфиг и фиксируйте ошибки сразу в одной ветке.

---

Готово: вы можете расширять разделы генерации, добавлять OpenAPI / grpc-gateway / validate плагины, матрицы по языкам и интегрировать тесты поверх сгенерированного кода, если требуется.
