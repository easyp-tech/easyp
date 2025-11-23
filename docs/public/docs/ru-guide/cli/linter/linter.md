# Линтер

[[toc]]

## Зачем нужен линтер для proto файлов?

Линтеры играют ключевую роль в современной разработке, особенно когда речь о `.proto` файлах. Они обеспечивают единый стиль и структуру, помогают находить потенциальные ошибки на ранних этапах и поддерживают читаемость кода. Это даёт выгоды:

- **Снижение затрат на разработку:** Ошибки ловятся раньше, меньше времени на отладку.
- **Улучшение командного взаимодействия:** Единообразный код легче понимать и сопровождать.
- **Бизнес-эффективность:** Более качественный код → меньше инцидентов в проде → ниже стоимость поддержки.

## Справочник по конфигурации

Конфигурация линтера настраивается в секции `lint` файла `easyp.yaml`.

### Полный пример конфигурации

```yaml
version: v1alpha

lint:
  # Используемые категории и правила
  use:
    - MINIMAL
    - BASIC
    - COMMENT_SERVICE
    - COMMENT_RPC
  
  # Кастомные суффиксы
  enum_zero_value_suffix: "UNSPECIFIED"
  service_suffix: "Service"
  
  # Исключения директорий и файлов
  ignore:
    - "vendor/"
    - "third_party/"
    - "legacy/old_protos/"
  
  except:
    - COMMENT_FIELD
    - COMMENT_MESSAGE
  
  # Разрешить игнорирование правил комментариями
  allow_comment_ignores: true
  
  # Игнор отдельных правил только для указанных путей
  ignore_only:
    COMMENT_SERVICE: ["legacy/", "vendor/"]
    SERVICE_SUFFIX: ["proto/external/"]
```

### Параметры конфигурации

#### `use` ([]string)
Определяет набор правил или категорий для включения. Можно смешивать категории и конкретные правила.

**Категории:**
- **MINIMAL**: Базовая согласованность пакетов
- **BASIC**: Соглашения по именованию и структуре  
- **DEFAULT**: Рекомендованные правила для большинства проектов
- **COMMENTS**: Требования к комментариям
- **UNARY_RPC**: Ограничение на streaming RPC

**Индивидуальные правила:** Любое имя правила (например `ENUM_PASCAL_CASE`, `FIELD_LOWER_SNAKE_CASE`)

**Примеры:**
```yaml
# Только категории
use: [MINIMAL, BASIC, DEFAULT]

# Смешивание
use:
  - MINIMAL
  - COMMENT_SERVICE
  - COMMENT_RPC
  - ENUM_PASCAL_CASE

# Точный набор
use:
  - PACKAGE_DEFINED
  - SERVICE_PASCAL_CASE
  - FIELD_LOWER_SNAKE_CASE
```

#### `enum_zero_value_suffix` (string)
Обязывает нулевое значение enum иметь заданный суффикс.

**Обычно:** `"UNSPECIFIED"`, `"UNKNOWN"`, `"DEFAULT"`
```yaml
enum_zero_value_suffix: "UNSPECIFIED"
```

```proto
enum Status {
  STATUS_UNSPECIFIED = 0;  // Требуемый суффикс
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}
```

#### `service_suffix` (string)
Требуемый суффикс для имён сервисов (унификация стиля).

```yaml
service_suffix: "Service"
```

```proto
service UserService { // Обязательный суффикс
  rpc GetUser(...) returns (...);
}
```

#### `ignore` ([]string)
Полное исключение путей из линтинга.

**Use cases:**
- vendor / third_party
- сгенерированные файлы
- легаси подпакеты
- тестовые фикстуры

```yaml
ignore:
  - "vendor/"
  - "third_party/"
  - "testdata/"
  - "proto/legacy/"
  - "**/*_test.proto"
```

#### `except` ([]string)
Глобально отключает правила.

```yaml
except:
  - COMMENT_FIELD
  - COMMENT_MESSAGE
  - SERVICE_SUFFIX
  - ENUM_ZERO_VALUE_SUFFIX
```

#### `allow_comment_ignores` (bool)
Разрешает инлайн‑игнор правил комментариями.

```yaml
allow_comment_ignores: true
```

```proto
// buf:lint:ignore COMMENT_SERVICE
service LegacyUserAPI {
  // nolint:COMMENT_RPC
  rpc GetUser(...) returns (...);
}
```

#### `ignore_only` (map[string][]string)
Таргетированное игнорирование конкретных правил для указанных путей.

```yaml
ignore_only:
  COMMENT_SERVICE:
    - "proto/legacy/"
    - "vendor/"
  SERVICE_SUFFIX:
    - "proto/external/"
    - "third_party/"
  FIELD_LOWER_SNAKE_CASE:
    - "proto/legacy/old_messages.proto"
    - "vendor/external_api.proto"
```

## Игнор правил комментариями

Если включён `allow_comment_ignores`, можно точечно отключать правила.

### Форматы комментариев

#### Совместимый с Buf
```proto
// buf:lint:ignore RULE_NAME
```

#### Нативный EasyP
```proto
// nolint:RULE_NAME
```

### Примеры

#### Игнор правил сервиса
```proto
// buf:lint:ignore COMMENT_SERVICE
// buf:lint:ignore SERVICE_SUFFIX
service UserAPI {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

#### Игнор RPC
```proto
service UserService {
  // nolint:COMMENT_RPC
  // nolint:RPC_REQUEST_STANDARD_NAME
  rpc GetUserInfo(UserInfoReq) returns (UserInfoResp);
}
```

#### Игнор Message и Field
```proto
// buf:lint:ignore COMMENT_MESSAGE
message UserData {
  // nolint:COMMENT_FIELD
  // nolint:FIELD_LOWER_SNAKE_CASE
  string userName = 1;
}
```

#### Игнор Enum
```proto
// nolint:COMMENT_ENUM
// nolint:ENUM_ZERO_VALUE_SUFFIX
enum UserType {
  UNKNOWN = 0; // Обычно требовался бы суффикс
  ADMIN = 1;
  USER = 2;
}
```

#### Несколько правил в одной строке
```proto
// buf:lint:ignore COMMENT_SERVICE,SERVICE_SUFFIX
service LegacyAPI {
  // nolint:COMMENT_RPC,RPC_REQUEST_STANDARD_NAME
  rpc getData(DataReq) returns (DataResp);
}
```

### Рекомендации

#### Используйте редко
Частое игнорирование = сигнал пересмотреть конфигурацию.
```proto
// Легаси сервис - buf:lint:ignore SERVICE_SUFFIX
service UserAPI { ... }
```

#### Плохой пример
```proto
// nolint:COMMENT_SERVICE,SERVICE_SUFFIX,COMMENT_RPC
service UserAPI { ... }
```

#### Добавляйте пояснения
```proto
// Совместимость с внешним API
// buf:lint:ignore SERVICE_SUFFIX
service ExternalUserAPI {
  // Легаси имя метода
  // nolint:RPC_REQUEST_STANDARD_NAME
  rpc getUserData(UserReq) returns (UserResp);
}
```

#### Предпочитайте конфигурацию
Если нужно игнорировать правило в группе файлов — используйте `ignore_only`.

## Категории линтера

Категории помогают быстро выбрать уровень строгости.

**Когда использовать:**
- **MINIMAL**: Минимальная базовая целостность
- **BASIC**: Частые соглашения имен
- **DEFAULT**: Дополнительные проверки качества
- **COMMENTS**: Обязательные комментарии (API, команды)
- **UNARY_RPC**: Ограничить streaming RPC

### Группировки правил

#### MINIMAL
- `DIRECTORY_SAME_PACKAGE`
- `PACKAGE_DEFINED`
- `PACKAGE_DIRECTORY_MATCH`
- `PACKAGE_SAME_DIRECTORY`

#### BASIC
- `ENUM_FIRST_VALUE_ZERO`
- `ENUM_NO_ALLOW_ALIAS`
- `ENUM_PASCAL_CASE`
- `ENUM_VALUE_UPPER_SNAKE_CASE`
- `FIELD_LOWER_SNAKE_CASE`
- `IMPORT_NO_PUBLIC`
- `IMPORT_NO_WEAK`
- `IMPORT_USED`
- `MESSAGE_PASCAL_CASE`
- `ONEOF_LOWER_SNAKE_CASE`
- `PACKAGE_LOWER_SNAKE_CASE`
- `PACKAGE_SAME_CSHARP_NAMESPACE`
- `PACKAGE_SAME_GO_PACKAGE`
- `PACKAGE_SAME_JAVA_MULTIPLE_FILES`
- `PACKAGE_SAME_JAVA_PACKAGE`
- `PACKAGE_SAME_PHP_NAMESPACE`
- `PACKAGE_SAME_RUBY_PACKAGE`
- `PACKAGE_SAME_SWIFT_PREFIX`
- `RPC_PASCAL_CASE`
- `SERVICE_PASCAL_CASE`

#### DEFAULT
- `ENUM_VALUE_PREFIX`
- `ENUM_ZERO_VALUE_SUFFIX`
- `FILE_LOWER_SNAKE_CASE`
- `RPC_REQUEST_RESPONSE_UNIQUE`
- `RPC_REQUEST_STANDARD_NAME`
- `RPC_RESPONSE_STANDARD_NAME`
- `PACKAGE_VERSION_SUFFIX`
- `SERVICE_SUFFIX`

#### COMMENTS
- `COMMENT_ENUM`
- `COMMENT_ENUM_VALUE`
- `COMMENT_FIELD`
- `COMMENT_MESSAGE`
- `COMMENT_ONEOF`
- `COMMENT_RPC`
- `COMMENT_SERVICE`

#### UNARY_RPC
- `RPC_NO_CLIENT_STREAMING`
- `RPC_NO_SERVER_STREAMING`

## Заключение

Внедрение EasyP линдера для ваших proto файлов заметно улучшит workflow, качество и поддерживаемость кода. Полная совместимость с Buf позволяет командам мигрировать без боли, используя гибкую конфигурацию и мощные возможности из коробки.