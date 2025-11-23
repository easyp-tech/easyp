# Менеджер пакетов

[[toc]]

EasyP предоставляет мощный менеджер пакетов для protobuf‑зависимостей, упрощающий их управление за счёт децентрализованного подхода на базе Git. В отличие от централизованных решений, EasyP работает напрямую с Git‑репозиториями, давая вам полный контроль.

## Обзор

Менеджер пакетов EasyP следует философии **Go modules** — любой Git‑репозиторий может быть источником. Это даёт преимущества:

- **Децентрализация** — нет единой точки отказа или контроля
- **Безопасность** — прямой доступ к исходникам
- **Гибкость** — публичные, приватные и корпоративные репозитории
- **Воспроизводимость** — lock‑файлы обеспечивают одинаковые сборки
- **Производительность** — локальный кеш сокращает сетевые запросы

### Основные возможности

| Возможность | Описание |
|-------------|----------|
| **Git-Native** | Любой Git репозиторий, не нужен спец‑сервер |
| **Множество форматов версий** | Теги, коммиты, псевдо‑версии, latest |
| **Lock Files** | Воспроизводимые сборки через `easyp.lock` |
| **Локальный кеш** | Архитектура как у Go modules |
| **Vendoring** | Локальное копирование для оффлайн сборок |
| **YAML конфигурация** | Простые и читаемые декларации |

## Архитектура

EasyP использует двухуровневый кеш, вдохновлённый Go modules:

```
~/.easyp/
├── cache/
│   ├── download/              # Архивы + контрольные суммы
│   │   └── github.com/
│   │       └── googleapis/
│   │           └── googleapis/
│   │               ├── v1.2.3.zip       # Архив
│   │               ├── v1.2.3.ziphash   # Checksum
│   │               └── v1.2.3.info      # Метаданные
│   └── {git-hash}/            # Bare репозитории Git (внутренне)
└── mod/                       # Распакованные готовые модули
    └── github.com/
        └── googleapis/
            └── googleapis/
                ├── v1.2.3/           # Тегированная версия
                │   ├── google/
                │   │   ├── api/
                │   │   └── rpc/
                │   └── ...
                └── v0.0.0-20250101123456-abc123def/  # Псевдо‑версия
                    ├── google/
                    └── ...
```

### Расположение кеша

| Окружение | Путь | Как задать |
|-----------|------|-----------|
| **По умолчанию** | `$HOME/.easyp` | Автоматически |
| **Свое** | Любая директория | Переменная `EASYPPATH` |
| **CI/CD** | В каталоге проекта | `export EASYPPATH=$CI_PROJECT_DIR/.easyp` |

## Конфигурация

### Базовая конфигурация

Зависимости описываются в `easyp.yaml`:

```yaml
version: v1alpha

deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  - github.com/bufbuild/protoc-gen-validate
```

### Расширенные примеры

#### Multi-Environment Setup
```yaml
# development.easyp.yaml
deps:
  - github.com/googleapis/googleapis              # Последнее для разработки
  - github.com/mycompany/internal-protos          # Внутренние изменения
  - github.com/bufbuild/protoc-gen-validate       # Новые фичи

# production.easyp.yaml
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1       # Зафиксировано
  - github.com/mycompany/internal-protos@v2.1.0                # Стабильно
  - github.com/bufbuild/protoc-gen-validate@v0.10.1            # Проверено
```

#### Приватные репозитории
```yaml
deps:
  # Публичное
  - github.com/googleapis/googleapis@common-protos-1_3_1

  # Частные
  - github.com/mycompany/auth-protos@v1.5.0
  - github.com/mycompany/common-types@v2.0.1

  # Внутренний GitLab
  - gitlab.company.com/platform/messaging-protos@v0.3.0
```

## Стратегии версионирования

Поддерживаются разные способы фиксации версий:

### 1. Семантические теги (для продакшена)

```yaml
deps:
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  - github.com/googleapis/googleapis@common-protos-1_3_1
```

**Используйте когда:**
- Продакшен
- Стабильное потребление API
- Нужна воспроизводимость

### 2. Последний тег (development)

```yaml
deps:
  - github.com/googleapis/googleapis
  - github.com/bufbuild/protoc-gen-validate
```

**Подходит для:**
- Активной разработки
- Получения свежих возможностей
- Тестирования совместимости

### 3. Хеш коммита (Bleeding Edge)

```yaml
deps:
  - github.com/bufbuild/protoc-gen-validate@abc123def456789abcdef123456789abcdef1234
```

**Когда нужно:**
- Нерелизные фичи
- Тест конкретного фикса
- Вклад в upstream

### 4. Псевдо‑версии (автоматически)

Если нет подходящего тега — создаётся псевдо‑версия:

```
Формат: v0.0.0-{timestamp}-{short-commit-hash}
Пример: v0.0.0-20250908104020-660ec2d64e07f2fa8947527443af058b3d7169df
```

## Команды

### `easyp mod download`

Скачивает и устанавливает все объявленные зависимости.

**Что делает:**
1. Разрешает версии (теги → коммиты)
2. Скачивает архивы в `cache/download`
3. Проверяет контрольные суммы
4. Распаковывает в `cache/mod`
5. Обновляет `easyp.lock`

**Пример:**
```bash
easyp mod download
easyp -cfg production.easyp.yaml mod download
EASYPPATH=/tmp/easyp-cache easyp mod download
```

### `easyp mod vendor`

Копирует все proto зависимости в локальный `easyp_vendor/` для оффлайна.

**Сценарии:**
- Docker сборки
- Air-gapped окружения
- Воспроизводимость
- Ускорение

```bash
easyp mod vendor
```

Структура:
```
easyp_vendor/
├── github.com/
│   ├── googleapis/
│   │   └── googleapis/
│   │       ├── google/
│   │       │   ├── api/
│   │       │   │   ├── annotations.proto
│   │       │   │   └── http.proto
│   │       │   └── rpc/
│   │       │       └── status.proto
│   │       └── ...
│   └── grpc-ecosystem/
│       └── grpc-gateway/
│           └── protoc-gen-openapiv2/
│               └── options/
│                   └── annotations.proto
```

### `easyp mod update`

Обновляет версии согласно `easyp.yaml`, пересоздаёт lock.

```bash
easyp mod update
```

## Lock файл

`easyp.lock` фиксирует точные версии и хеш содержимого:

```
github.com/bufbuild/protoc-gen-validate v0.0.0-20250908104020-660ec2d64e07f2fa8947527443af058b3d7169df h1:ZZ5JyUkmrj9OBHM+gOCzeL5L/pAKVbsUl051yhhJTjU=
github.com/googleapis/googleapis v0.0.0-20250909114430-8727b5ba7f23fbbfddda58239e8bc6b547e05878 h1:eI+XYpPio3fxl9H5/VjW2PxlxM/7yqPjEq3oQ6jUkj4=
github.com/grpc-ecosystem/grpc-gateway v2.19.1 h1:01NNlCezvwUQ07ZvblXH0kelWq8hNl2qb44bOMcaSTQ=
```

**Формат строки:**
- Путь модуля
- Версия (тег или псевдо)
- Хеш содержимого (`h1:`)

**Практики:**
✅ Коммитить `easyp.lock`  
✅ Осознанно обновлять `mod update`  
✅ Ревью изменений версий  
❌ Не редактировать вручную  

## Аутентификация

### Публичные репозитории

Работают без настроек:

```yaml
deps:
  - github.com/googleapis/googleapis
  - github.com/bufbuild/protoc-gen-validate
```

### Приватные репозитории

#### SSH ключи (рекомендуется)

```bash
git config --global url."git@github.com:".insteadOf "https://github.com/"
git config --global url."git@gitlab.com:".insteadOf "https://gitlab.com/"
git config --global url."git@gitlab.company.com:".insteadOf "https://gitlab.company.com/"
```

Конфиг остаётся с HTTPS URL:

```yaml
deps:
  - github.com/mycompany/private-protos@v1.0.0
  - gitlab.company.com/platform/shared-types@v2.1.0
```

#### Personal Access Tokens

```bash
git config --global credential.helper store
echo "https://username:token@github.com" >> ~/.git-credentials

git config --global url."https://username:token@github.com/mycompany".insteadOf "https://github.com/mycompany"
```

#### Корпоративная среда

```bash
git config --global http.proxy http://proxy.company.com:8080
git config --global https.proxy https://proxy.company.com:8080
git config --global http.sslCAInfo /path/to/certificate.pem
```

## Типовые workflows

### Инициализация проекта

```bash
cat > easyp.yaml << EOF
version: v1alpha
deps:
  - github.com/googleapis/googleapis
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
EOF

easyp mod download
ls ~/.easyp/mod/github.com/googleapis/googleapis/
```

### Добавление зависимости

```bash
echo "  - github.com/bufbuild/protoc-gen-validate@v0.10.1" >> easyp.yaml
easyp mod download
git add easyp.lock
git commit -m "Add protoc-gen-validate dependency"
```

### Обновление

```bash
easyp mod update
git diff easyp.lock
easyp generate
easyp lint
git add easyp.lock
git commit -m "Update dependencies"
```

### Оффлайн режим

```bash
easyp mod vendor
easyp -I easyp_vendor generate
```

## Troubleshooting

### Частые проблемы

#### "Repository not found" / "Authentication failed"

```bash
git ls-remote https://github.com/mycompany/private-repo
git config --list | grep url
```

#### "Version not found"

```bash
git ls-remote --tags https://github.com/googleapis/googleapis
# Проверить корректный тег
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
```

#### "Cache corruption" / "Checksum mismatch"

```bash
rm -rf ~/.easyp
rm -rf ~/.easyp/cache/download
easyp mod download
```

#### Тайм-ауты сети

```bash
git config --global http.lowSpeedLimit 1000
git config --global http.lowSpeedTime 300
git config --global http.proxy http://proxy.company.com:8080
```

### Оптимизация

#### Для больших команд

```bash
export EASYPPATH=/shared/easyp-cache
export EASYPPATH=/team-cache/easyp
```

#### Для CI/CD

```bash
export EASYPPATH=$CI_PROJECT_DIR/.easyp
# Пример кеширования (GitLab CI)
cache:
  key: easyp-$CI_COMMIT_REF_SLUG
  paths:
    - .easyp/
```

#### Управление размером кеша

```bash
du -sh ~/.easyp
du -sh ~/.easyp/cache/download
du -sh ~/.easyp/mod
find ~/.easyp/mod -type d -name "v0.0.0-*" -mtime +30 -exec rm -rf {} \;
```

## Примеры интеграции

### Docker multi-stage

```dockerfile
FROM easyp/easyp:latest AS deps
WORKDIR /workspace
COPY easyp.yaml easyp.lock ./
RUN easyp mod vendor

FROM alpine:latest AS build
WORKDIR /app
COPY --from=deps /workspace/easyp_vendor ./easyp_vendor
COPY . .
RUN easyp -I easyp_vendor generate
```

### Monorepo

```
my-monorepo/
├── services/
│   ├── auth-service/
│   │   └── easyp.yaml
│   └── user-service/
│       └── easyp.yaml
├── shared/
│   └── common-protos/
└── easyp.yaml
```

Каждый `easyp.yaml` может иметь свой набор зависимостей.

## Лучшие практики

### Workflow разработки
- ✅ Latest теги в активной разработке
- ✅ Фиксация версий в продакшене
- ✅ Коммит lock файла
- ✅ Ревью обновлений
- ✅ Тест после обновления

### Безопасность
- ✅ Фиксируйте версии
- ✅ Используйте SSH ключи
- ✅ Оценивайте новые зависимости
- ✅ Следите за уязвимостями
- ❌ Не храните секреты в конфиге

### Производительность
- ✅ Агрессивное кеширование в CI
- ✅ Vendoring для частых сборок
- ✅ Чистка старого кеша
- ✅ Общий кеш для команды

### Командная работа
- ✅ Документация по аутентификации
- ✅ Единые инструменты
- ✅ Автоматизация обновлений с тестами
- ✅ Общий кеш при возможности

Менеджер пакетов EasyP — масштабируемое децентрализованное решение от одиночных проектов до enterprise, сохраняя простоту и надёжность.
