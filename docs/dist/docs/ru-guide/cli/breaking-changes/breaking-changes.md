# Обнаружение несовместимых изменений (Breaking Changes Detection)

[[toc]]

Механизм EasyP для обнаружения несовместимых (breaking) изменений помогает поддерживать обратную совместимость ваших protobuf API, автоматически выявляя изменения, способные сломать существующих клиентов. Это критично для стабильных API в продакшене.

## Обзор

Проверка несовместимых изменений сравнивает текущий набор protobuf‑файлов с предыдущей версией (обычно из другой Git ветки) и фиксирует модификации, которые могут вызвать проблемы совместимости для уже работающих клиентов.

### Ключевые возможности

- **Сравнение через Git**: Сопоставление с любой Git‑ссылкой (ветка, тег, commit)
- **Комплексный анализ**: Проверка сервисов, сообщений, enum'ов, полей и import'ов
- **Выборочное игнорирование**: Пропуск указанных директорий из анализа
- **Детализированные отчёты**: Понятные ошибки с именами файлов, строками и позициями

## Как работает

Детектор breaking changes выполняет следующие шаги:

1. **Получение ветки сравнения**: Забирает proto-файлы из указанной Git‑ссылки
2. **Парсинг обеих версий**: Анализирует структуру текущих и предыдущих файлов
3. **Сравнение сущностей**: Последовательно проверяет все элементы protobuf на breaking изменения
4. **Формирование отчёта**: Генерирует подробные записи об ошибках с локациями и описаниями

## Уровень проверки

EasyP реализует уровень проверки **WIRE+**:
- ✅ **Полная совместимость по wire‑формату** — старые и новые версии могут обмениваться данными
- ✅ **Обнаружение удалений элементов** — сервисы, сообщения, поля и т.п.
- ✅ **Безопасность типов** — выявляет несовместимые изменения типов
- ❌ **Переименования полей/методов** — пока не детектируются (в планах)
- ❌ **Изменения на уровне файлов** — переносы пакетов, file options ещё не проверяются

Это даёт сильные гарантии совместимости и меньше ограничений, чем инструменты, требующие совместимости сгенерированного кода.

## Конфигурация

Настройка в `easyp.yaml`:

```yaml
breaking:
  # Git reference для сравнения (branch, tag или commit hash)
  against_git_ref: "main"
  
  # Директории, игнорируемые в анализе breaking changes
  ignore:
    - "experimental"
    - "internal/proto"
    - "vendor"
```

### Параметры конфигурации

| Option | Description | Default | Required |
|--------|-------------|---------|----------|
| `against_git_ref` | Git‑ссылка для сравнения | `"master"` | No |
| `ignore` | Список директорий для исключения из анализа | `[]` | No |

## Использование

### Базовый пример

Сравнение текущих изменений с веткой main:

```bash
easyp breaking --against main
```

### Использование файла конфигурации

```bash
easyp --cfg my-config.yaml breaking
```

### Переопределение Git‑ссылки

```bash
easyp breaking --against feature/new-api
```

## Уровень проверки

EasyP сейчас работает на уровне **WIRE+**, обеспечивая совместимость wire‑формата плюс защиту от удалений и изменения критичных типов.

### Сравнение с категориями Buf

| Check Type | Buf WIRE | Buf WIRE_JSON | Buf FILE | EasyP Current |
|------------|----------|---------------|----------|---------------|
| **Element Deletions** |
| Service deletion | ❌ | ❌ | ✅ | ✅ |
| RPC method deletion | ❌ | ❌ | ✅ | ✅ |
| Message deletion | ❌ | ❌ | ✅ | ✅ |
| Field deletion (by number) | ✅ | ✅ | ✅ | ✅ |
| Enum deletion | ❌ | ❌ | ✅ | ✅ |
| Enum value deletion | ✅ | ✅ | ✅ | ✅ |
| OneOf deletion | ❌ | ❌ | ✅ | ✅ |
| Import deletion | ❌ | ❌ | ✅ | ✅ |
| **Type Changes** |
| Field type change | ✅ | ✅ | ✅ | ✅ |
| RPC request/response type | ✅ | ✅ | ✅ | ✅ |
| Optional/required changes | ✅ | ✅ | ✅ | ✅ |
| **Naming (Generated Code)** |
| Field rename (same number) | ❌ | ✅ | ✅ | ❌ |
| Enum value rename | ❌ | ✅ | ✅ | ✅ |
| **File Structure** |
| Package change | ✅ | ✅ | ✅ | ❌ |
| File options (go_package, etc) | ❌ | ❌ | ✅ | ❌ |
| Moving types between files | ❌ | ❌ | ✅ | ❌ |

### Что это означает

**✅ EasyP обнаружит:**
- Все breaking изменения wire‑формата
- Удаления сервисов, методов, сообщений, полей
- Изменения типов, ломающие сериализацию
- Переименования enum‑значений (при сохранении номера)

**❌ EasyP НЕ обнаружит:**
- Переименование полей (с тем же номером)
- Изменение package
- File options (go_package, java_package, и т.п.)
- Перемещение типов между файлами в одном package

## Правила Breaking Changes

EasyP обнаруживает следующие типы breaking изменений:

### Сравнение уровней

| Detection Level | Description | EasyP Support |
|----------------|-------------|---------------|
| **WIRE** | Совместимость только по wire‑формату | ✅ **Full support** |
| **WIRE+** | Wire + обнаружение удалений | ✅ **Current level** |
| **FILE** | Совместимость сгенерированного кода | ❌ Partial (planned) |

## Категории правил

Каждая категория имеет свою документацию с примерами:

### 🚨 Изменения Service и RPC

| Rule | Description | Status |
|------|-------------|---------|
| [SERVICE_NO_DELETE](./rules/service-no-delete.md) | Services cannot be deleted | ✅ Implemented |
| [RPC_NO_DELETE](./rules/rpc-no-delete.md) | RPC methods cannot be deleted | ✅ Implemented |
| [RPC_SAME_REQUEST_TYPE](./rules/rpc-same-request-type.md) | RPC request types cannot be changed | ✅ Implemented |
| [RPC_SAME_RESPONSE_TYPE](./rules/rpc-same-response-type.md) | RPC response types cannot be changed | ✅ Implemented |

### 📦 Изменения Message и Field

| Rule | Description | Status |
|------|-------------|---------|
| [MESSAGE_NO_DELETE](./rules/message-no-delete.md) | Messages cannot be deleted | ✅ Implemented |
| [FIELD_NO_DELETE](./rules/field-no-delete.md) | Fields cannot be deleted | ✅ Implemented |
| [FIELD_SAME_TYPE](./rules/field-same-type.md) | Field types cannot be changed | ✅ Implemented |
| [FIELD_SAME_CARDINALITY](./rules/field-same-cardinality.md) | Field optionality (optional/required) cannot be changed | ✅ Implemented |

### 🔢 Изменения Enum

| Rule | Description | Status |
|------|-------------|---------|
| [ENUM_NO_DELETE](./rules/enum-no-delete.md) | Enums cannot be deleted | ✅ Implemented |
| [ENUM_VALUE_NO_DELETE](./rules/enum-value-no-delete.md) | Enum values cannot be deleted | ✅ Implemented |
| [ENUM_VALUE_SAME_NAME](./rules/enum-value-same-name.md) | Enum value names cannot be changed | ✅ Implemented |

### 🔗 Изменения OneOf

| Rule | Description | Status |
|------|-------------|---------|
| [ONEOF_NO_DELETE](./rules/oneof-no-delete.md) | OneOf fields cannot be deleted | ✅ Implemented |
| [ONEOF_FIELD_NO_DELETE](./rules/oneof-field-no-delete.md) | Fields within oneofs cannot be deleted | ✅ Implemented |
| [ONEOF_FIELD_SAME_TYPE](./rules/oneof-field-same-type.md) | OneOf field types cannot be changed | ✅ Implemented |

### 📥 Изменения Import

| Rule | Description | Status |
|------|-------------|---------|
| [IMPORT_NO_DELETE](./rules/import-no-delete.md) | Import statements cannot be removed | ✅ Implemented |

## Не обнаруживается сейчас

Изменения ниже **НЕ детектируются** EasyP (могут ломать сгенерированный код):

| Change Type | Example | Impact |
|-------------|---------|--------|
| Field renaming | `string name = 1` → `string full_name = 1` | Ломает код |
| Package changes | `package v1` → `package v2` | Меняются пути импорта |
| File options | `option go_package = "old"` → `option go_package = "new"` | Меняется расположение кода |
| Moving between files | Message перемещено в другой .proto | Зависимости import'ов |

## Документация правил

См. подробные файлы правил для примеров и стратегий миграции:

- Service: [SERVICE_NO_DELETE](./rules/service-no-delete.md), [RPC_NO_DELETE](./rules/rpc-no-delete.md), [RPC_SAME_REQUEST_TYPE](./rules/rpc-same-request-type.md), [RPC_SAME_RESPONSE_TYPE](./rules/rpc-same-response-type.md)
- Message: [MESSAGE_NO_DELETE](./rules/message-no-delete.md), [FIELD_NO_DELETE](./rules/field-no-delete.md), [FIELD_SAME_TYPE](./rules/field-same-type.md), [FIELD_SAME_CARDINALITY](./rules/field-same-cardinality.md)
- Enum: [ENUM_NO_DELETE](./rules/enum-no-delete.md), [ENUM_VALUE_NO_DELETE](./rules/enum-value-no-delete.md), [ENUM_VALUE_SAME_NAME](./rules/enum-value-same-name.md)
- OneOf: [ONEOF_NO_DELETE](./rules/oneof-no-delete.md), [ONEOF_FIELD_NO_DELETE](./rules/oneof-field-no-delete.md), [ONEOF_FIELD_SAME_TYPE](./rules/oneof-field-same-type.md)
- Import: [IMPORT_NO_DELETE](./rules/import-no-delete.md)

Каждое правило содержит:
- ❌ Плохие примеры (breaking)
- ✅ Безопасные альтернативы
- 🔧 Стратегии миграции
- 📋 Реальные сообщения ошибок из EasyP

## Быстрые примеры

### ✅ Безопасные изменения (разрешены всегда)
```proto
// Добавление новых элементов — безопасно
message User {
  string name = 1;
  string email = 2;
  string phone = 3;  // ✅ Новое поле
}

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc GetUserProfile(GetUserRequest) returns (UserProfile);  // ✅ Новый RPC
}

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_PENDING = 2;  // ✅ Новое enum значение
}
```

### ❌ Несовместимые изменения (обнаруживаются всегда)
```proto
// Удаления и смена типов ломают совместимость
message User {
  string name = 1;
  // ❌ Удалённое поле
}

service UserService {
  // ❌ Удалённый RPC метод
  rpc GetUser(GetUserRequestV2) returns (GetUserResponse);  // ❌ Изменён тип запроса
}
```

### Сценарий: Изменения пока не детектируются

```proto
// 🟡 Ломает сгенерированный код, но проходит проверки EasyP
message User {
  string user_name = 1;    // Переименовано с "name"
  string user_email = 2;   // Переименовано с "email"
}

service UserService {
  rpc GetUserProfile(GetUserRequest) returns (GetUserResponse);  // Переименовано с GetUser
}
```

## Формат вывода

### Текстовый формат (по умолчанию)

```
services.proto:45:1: Previously present RPC "DeleteUser" on service "UserService" was deleted. (BREAKING_CHECK)
messages.proto:15:3: Previously present field "2" with name "email" on message "User" was deleted. (BREAKING_CHECK)
```

### JSON формат

```bash
easyp --format json breaking --against main
```

```json
{
  "path": "services.proto",
  "position": {
    "line": 45,
    "column": 1
  },
  "source_name": "",
  "message": "Previously present RPC \"DeleteUser\" on service \"UserService\" was deleted.",
  "rule_name": "BREAKING_CHECK"
}
```

## Best Practices

### 1. Регулярный запуск
Добавьте проверку несовместимых изменений в CI/CD:

```yaml
- name: Check for breaking changes
  run: easyp breaking --against origin/main
```

### 2. Защита веток
Блокируйте слияния с breaking изменениями:

```yaml
if: github.event_name == 'pull_request'
run: |
  easyp breaking --against origin/main
  if [ $? -eq 1 ]; then
    echo "Breaking changes detected!"
    exit 1
  fi
```

### 3. Стратегия версионирования
При необходимости breaking изменений:
- Создайте новую версию пакета (`myservice.v2`)
- Поддерживайте старую версию в период миграции
- Добавляйте уведомления о деприкации

### 4. Игнорирование путей
Используйте ignore осознанно:

```yaml
breaking:
  ignore:
    - "experimental/**"
    - "internal/**"
    - "**/testing/**"
```

## Частые проблемы

### Проблема: "Repository does not exist"
Решение: Выполняйте команду внутри Git репозитория с нужной веткой.

### Проблема: "Cannot find git ref"
Решение: Проверьте наличие ссылки:
```bash
git branch -a
git tag
```

### Проблема: Ложные срабатывания на сгенерированный код
Решение: Исключите директории генерации:
```yaml
breaking:
  ignore:
    - "generated/**"
    - "**/pb/**"
```

### Проблема: Очень много изменений
Решение:
1. Создайте новую версию API
2. Вносите изменения постепенно
3. Используйте feature flags

## Примеры интеграции

### GitHub Actions

```yaml
name: API Compatibility Check

on:
  pull_request:
    branches: [ main ]

jobs:
  breaking-changes:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      - name: Install EasyP
        run: |
          curl -sSfL https://raw.githubusercontent.com/easyp-tech/easyp/main/install.sh | sh
      - name: Check for breaking changes
        run: |
          ./bin/easyp breaking --against origin/main
```

### GitLab CI

```yaml
breaking-changes:
  stage: test
  image: easyp/lint:latest
  script:
    - git fetch origin main
    - easyp breaking --against origin/main
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
```

### Pre-commit / pre-push hook

```bash
#!/bin/sh
# .git/hooks/pre-push

protected_branch='main'
current_branch=$(git symbolic-ref HEAD | sed -e 's,.*/\(.*\),\1,')

if [ $current_branch = $protected_branch ]; then
    echo "Running breaking changes check..."
    easyp breaking --against HEAD~1
    if [ $? -eq 1 ]; then
        echo "❌ Breaking changes detected. Push rejected."
        exit 1
    fi
    echo "✅ No breaking changes detected."
fi
```

## Troubleshooting

### Режим отладки
Включите debug‑логирование:

```bash
easyp --debug breaking --against main
```

### Ручное сравнение
Для сложных случаев:

```bash
git show main:path/to/file.proto > old_version.proto
easyp lint old_version.proto
easyp lint current_file.proto
```

### Оптимизация производительности
Ограничьте область:

```bash
easyp breaking --against main --path api/
```

Механизм обнаружения несовместимых изменений в EasyP обеспечивает надёжную основу для безопасной эволюции protobuf‑схем при сохранении совместимости API.
