# EasyP Configuration Reference

EasyP can be configured through CLI flags, environment variables, and configuration files. This guide covers all configuration options available in EasyP.

## CLI Flags

### Global Flags

Available for all commands:

| Flag | Short | Environment | Description | Default |
|------|-------|-------------|-------------|---------|
| `--cfg` | `-c` | `EASYP_CFG` | Configuration file path | `easyp.yaml` |
| `--config` | | `EASYP_CFG` | Alias for `--cfg` | `easyp.yaml` |
| `--debug` | `-d` | `EASYP_DEBUG` | Enable debug mode | `false` |
| `--format` | `-f` | `EASYP_FORMAT` | Формат вывода для команд с поддержкой нескольких форматов (`text`/`json`) | значение по умолчанию зависит от команды |

**Examples:**
```bash
# Use custom config file
easyp --cfg production.easyp.yaml lint

# Enable debug logging
easyp --debug lint

# Short form
easyp -c custom.yaml -d lint
```

### Command-Specific Flags

**Lint command:**
```bash
easyp lint [flags]
```

| Flag | Short | Environment | Description | Default |
|------|-------|-------------|-------------|---------|
| `--path` | `-p` | | Directory path to lint | `.` |
| `--root` | `-r` | | Базовая директория для поиска файлов | Текущая рабочая директория |
| `--format` | `-f` | `EASYP_FORMAT` | Использует глобальный флаг формата (`text`/`json`) | Использует глобальное значение по умолчанию |

**Examples:**
```bash
# Lint specific directory
easyp lint --path proto/

# Lint из поддиректории с правильным разрешением импортов
easyp lint --root src/IPC/Contracts --path .

# JSON output format
easyp lint --format json   # глобальный флаг

# Combined flags
easyp lint -p proto/ -f json
```

**Generate command:**
```bash
easyp generate [flags]
```

| Flag | Short | Environment | Description | Default |
|------|-------|-------------|-------------|---------|
| `--path` | `-p` | `EASYP_ROOT_GENERATE_PATH` | Root path for generation | `.` |
| `--root` | `-r` | | Базовая директория для поиска файлов | Текущая рабочая директория |

**Examples:**
```bash
# Generate from specific path
easyp generate --path api/

# Generate из поддиректории с правильным разрешением импортов
easyp generate --root src/IPC/Contracts --path .

# Using environment variable
EASYP_ROOT_GENERATE_PATH=proto/ easyp generate
```

**Breaking command:**
```bash
easyp breaking [flags]
```

| Flag | Short | Environment | Description | Default |
|------|-------|-------------|-------------|---------|
| `--against` | | | Git ref to compare against | (required) |
| `--path` | `-p` | | Directory path to check | `.` |
| `--format` | `-f` | `EASYP_FORMAT` | Использует глобальный флаг формата (`text`/`json`) | Использует глобальное значение по умолчанию |

**Examples:**
```bash
# Check against main branch
easyp breaking --against main

# Check specific directory against develop branch
easyp breaking --against develop --path proto/

# JSON output
easyp breaking --against main --format json   # глобальный флаг
```

**Init command:**
```bash
easyp init [flags]
```

| Flag | Short | Environment | Description | Default |
|------|-------|-------------|-------------|---------|
| `--dir` | `-d` | `EASYP_INIT_DIR` | Directory to initialize | `.` |

**Examples:**
```bash
# Initialize current directory
easyp init

# Initialize specific directory
easyp init --dir proto-project/
```

`easyp init` работает интерактивно:
- Если в корне целевой директории есть `buf.yml`/`buf.yaml`, команда предложит миграцию из Buf.
- Если `easyp.yaml` уже существует, команда запросит подтверждение на перезапись.

**Validate-config command:**
```bash
easyp validate-config [flags]
```

| Flag | Short | Environment | Description | Default |
|------|-------|-------------|-------------|---------|
| `--config` | `-c` | `EASYP_CFG` | Путь до файла конфигурации | `easyp.yaml` |
| `--format` (global) | `-f` | `EASYP_FORMAT` | Глобальный формат вывода для команд с несколькими форматами (`json` или `text`) | зависит от команды (`json` для `validate-config`) |

**Examples:**
```bash
# Проверить конфиг по умолчанию с JSON выводом (статус 0 если ошибок нет)
easyp validate-config

# Проверить другой файл и вывести в текстовом формате (глобальный --format)
easyp --format text validate-config --config example.easyp.yaml
```

**Package management commands:**

#### `easyp mod download`
Downloads dependencies based on lock file priority:

1. **If `easyp.lock` exists** - downloads exact versions from lock file
2. **If `easyp.lock` is missing** - downloads versions from `easyp.yaml` and creates `easyp.lock`

```bash
# Download exact versions (recommended for production)
easyp mod download
```

#### `easyp mod update`  
Always downloads dependencies from `easyp.yaml`, ignoring existing lock file:

1. **Ignores `easyp.lock`** completely
2. **Downloads versions from `easyp.yaml`**
3. **Creates/updates `easyp.lock`** with new versions

```bash
# Update dependencies and lock file
easyp mod update
```

#### `easyp mod vendor`
Copies proto files from dependencies to local `vendor/` directory (similar to `go mod vendor`).

```bash
# Create local vendor directory with dependencies
easyp mod vendor
```

No additional flags. Uses global `--cfg` flag for configuration.

## Environment Variables

EasyP supports environment variables for configuration:

| Variable | Description | Default |
|----------|-------------|---------|
| `EASYP_CFG` | Path to configuration file | `easyp.yaml` |
| `EASYP_DEBUG` | Enable debug logging | `false` |
| `EASYPPATH` | Cache and modules storage directory | `$HOME/.easyp` |
| `EASYP_FORMAT` | Формат вывода для поддерживаемых команд (`text`/`json`). Если не указан, каждая команда использует своё значение по умолчанию. | значение по умолчанию зависит от команды |
| `EASYP_ROOT_GENERATE_PATH` | Root path for generate command | `.` |
| `EASYP_INIT_DIR` | Directory for init command | `.` |

**Examples:**
```bash
# Custom cache directory
export EASYPPATH=/tmp/easyp-cache
easyp mod download

# Debug mode via environment
export EASYP_DEBUG=true
easyp lint

# Custom config file
export EASYP_CFG=config/easyp.yaml
easyp generate
```

## Configuration File

The `easyp.yaml` file is the main configuration file for EasyP, defining how your proto files are linted, generated, and managed. This file is typically placed at the root of your project alongside your proto files.

### File Structure Overview

```
.
├── easyp.yaml
├── easyp.lock
├── proto/
│   ├── user/
│   │   └── user.proto
│   └── order/
│       └── order.proto
└── vendor/
```

### Configuration Format

EasyP supports both YAML and JSON configuration formats:

#### YAML Format (Recommended)
```yaml
lint:
  use:
    - BASIC
    - COMMENT_SERVICE
deps:
  - github.com/googleapis/googleapis@v1.0.0
generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: .
      opts:
        paths: source_relative
breaking:
  ignore:
    - proto/experimental/
  against_git_ref: main
```

#### JSON Format
```json
{
  "lint": {
    "use": ["BASIC", "COMMENT_SERVICE"]
  },
  "deps": [
    "github.com/googleapis/googleapis@v1.0.0"
  ],
  "generate": {
    "inputs": [
      {"directory": "proto"}
    ],
    "plugins": [
      {
        "name": "go",
        "out": ".",
        "opts": {
          "paths": "source_relative"
        }
      }
    ]
  },
  "breaking": {
    "ignore": ["proto/experimental/"],
    "against_git_ref": "main"
  }
}
```

### Переменные окружения в конфигурации

EasyP поддерживает расширение переменных окружения непосредственно в файле конфигурации `easyp.yaml`. Это позволяет использовать переменные окружения для динамических значений конфигурации.

**Пример со всеми поддерживаемыми возможностями:**
```yaml
deps:
  # Базовое расширение: ${VAR} - расширяется в значение переменной VAR
  - ${GOOGLEAPIS_REPO}@${GOOGLEAPIS_VERSION}
  
  # Значение по умолчанию: ${VAR:-default} - использует default, если VAR не установлена или пустая
  - ${GNOSTIC_REPO:-github.com/google/gnostic}@${GNOSTIC_VERSION:-v0.7.0}

generate:
  inputs:
    # Значение по умолчанию, если INPUT_DIR не установлена
    - directory: ${INPUT_DIR:-proto}
  plugins:
    - name: go
      # Базовое расширение
      out: ${OUTPUT_DIR}
      opts:
        # Значение по умолчанию
        module: ${MODULE_NAME:-github.com/example/project}
        timeout: ${TIMEOUT:-30}
        
        # Экранирование: $$ становится буквальным $, $${VAR} становится буквальным ${VAR}
        path: "${BASE_DIR}/$${TEMP}/file"         # Результат: "/tmp/${TEMP}/file" (если BASE_DIR=/tmp)
        literal: "$$"                             # Результат: "$"
```

**Поддерживаемый синтаксис:**
- `${VAR}` - расширяется в значение переменной `VAR`
- `${VAR:-default}` - использует `default`, если `VAR` не установлена или пустая
- `${VAR:=default}` - использует `default`, если `VAR` не установлена или пустая
- `${VAR-default}` - использует `default`, если `VAR` не установлена (пустая строка сохраняется)
- `$${VAR}` или `$$VAR` - экранирует в буквальную строку `${VAR}` или `$VAR`
- `$$` - экранирует в буквальный символ `$`

**Примечание:** Переменные окружения расширяются до парсинга YAML, поэтому любой паттерн `${STRING}` будет обработан. Используйте `$$` для экранирования символов доллара, когда нужны буквальные значения.

## Configuration Fields

### `version`

**Опционально (legacy-совместимость).** Поле поддерживается для обратной совместимости и может быть опущено в новых конфигурациях.

**Type:** `string`
**Default:** не задано
**Рекомендация:** если сохраняете поле, используйте `v1alpha`

```yaml
# Опциональное поле для совместимости (можно опустить)
version: v1alpha
```

На поведение рантайма это поле не влияет.

### `lint`

**Optional.** Configures proto file linting rules and behavior.

**Type:** `object`
**Default:** Empty (no linting rules applied)

```yaml
lint:
  use:
    - BASIC
    - COMMENT_SERVICE
  enum_zero_value_suffix: "UNSPECIFIED"
  service_suffix: "Service"
  ignore:
    - vendor/
    - proto/legacy/
  except:
    - COMMENT_FIELD
  allow_comment_ignores: true
  ignore_only:
    COMMENT_SERVICE:
      - proto/experimental/
```

#### `lint.use`

**Optional.** Specifies which linter rules or rule categories to apply.

**Type:** `[]string`
**Default:** `[]` (no rules)

**Available categories:**
- `MINIMAL` - Essential package consistency checks
- `BASIC` - Naming conventions and common patterns
- `DEFAULT` - Additional recommended rules
- `COMMENTS` - Comment requirements
- `UNARY_RPC` - Streaming RPC restrictions

**Individual rules:** Any specific rule name (e.g., `ENUM_PASCAL_CASE`, `FIELD_LOWER_SNAKE_CASE`)

```yaml
lint:
  use:
    - MINIMAL           # Use all minimal rules
    - BASIC             # Use all basic rules
    - COMMENT_SERVICE   # Require service comments
    - ENUM_PASCAL_CASE  # Specific rule
```

**Rule Categories:**

**MINIMAL:**
- `DIRECTORY_SAME_PACKAGE`
- `PACKAGE_DEFINED`
- `PACKAGE_DIRECTORY_MATCH`
- `PACKAGE_SAME_DIRECTORY`

**BASIC:**
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

**DEFAULT:**
- `ENUM_VALUE_PREFIX`
- `ENUM_ZERO_VALUE_SUFFIX`
- `FILE_LOWER_SNAKE_CASE`
- `RPC_REQUEST_RESPONSE_UNIQUE`
- `RPC_REQUEST_STANDARD_NAME`
- `RPC_RESPONSE_STANDARD_NAME`
- `PACKAGE_VERSION_SUFFIX`
- `SERVICE_SUFFIX`

**COMMENTS:**
- `COMMENT_ENUM`
- `COMMENT_ENUM_VALUE`
- `COMMENT_FIELD`
- `COMMENT_MESSAGE`
- `COMMENT_ONEOF`
- `COMMENT_RPC`
- `COMMENT_SERVICE`

**UNARY_RPC:**
- `RPC_NO_CLIENT_STREAMING`
- `RPC_NO_SERVER_STREAMING`

#### `lint.enum_zero_value_suffix`

**Optional.** Specifies the required suffix for enum zero values.

**Type:** `string`
**Default:** `""` (no suffix required)
**Common values:** `"UNSPECIFIED"`, `"UNKNOWN"`, `"DEFAULT"`

```yaml
lint:
  enum_zero_value_suffix: "UNSPECIFIED"
```

This enforces enum zero values like:
```proto
enum Status {
  STATUS_UNSPECIFIED = 0;  // Required suffix
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}
```

#### `lint.service_suffix`

**Optional.** Specifies the required suffix for service names.

**Type:** `string`
**Default:** `""` (no suffix required)
**Common values:** `"Service"`, `"API"`, `"Svc"`

```yaml
lint:
  service_suffix: "Service"
```

This enforces service names like:
```proto
service UserService {    // Required "Service" suffix
  rpc GetUser(...) returns (...);
}
```

#### `lint.ignore`

**Optional.** Directories or files to exclude from all linting rules.

**Type:** `[]string`
**Default:** `[]`

```yaml
lint:
  ignore:
    - vendor/
    - proto/legacy/
    - testdata/
    - "**/*_test.proto"
```

Paths are relative to the `easyp.yaml` file location. Supports glob patterns.

#### `lint.except`

**Optional.** Disables specific rules globally across the entire project.

**Type:** `[]string`
**Default:** `[]`

```yaml
lint:
  except:
    - COMMENT_FIELD
    - COMMENT_MESSAGE
    - SERVICE_SUFFIX
```

#### `lint.allow_comment_ignores`

**Optional.** Enables inline comment-based rule ignoring within proto files.

**Type:** `boolean`
**Default:** `false`

```yaml
lint:
  allow_comment_ignores: true
```

When enabled, allows comments like:
```proto
// buf:lint:ignore COMMENT_SERVICE
service LegacyAPI {
  // nolint:COMMENT_RPC
  rpc GetData(...) returns (...);
}
```

#### `lint.ignore_only`

**Optional.** Disables specific rules only for certain files or directories.

**Type:** `map[string][]string`
**Default:** `{}`

```yaml
lint:
  ignore_only:
    COMMENT_SERVICE:
      - proto/legacy/
      - vendor/
    SERVICE_SUFFIX:
      - proto/external/
```

**Key:** Rule name or category
**Value:** Array of file paths or directories

### `deps`

**Optional.** Lists external proto dependencies to download and manage.

**Type:** `[]string`  
**Default:** `[]`

#### Dependency Format

Dependencies follow the format: `$GIT_LINK@$VERSION`

**Components:**
- `$GIT_LINK` - Git repository URL (GitHub, GitLab, etc.)
- `$VERSION` - Git tag or full commit hash (optional)

**Format variations:**
- `owner/repo` - Latest commit from default branch
- `owner/repo@v1.0.0` - Specific git tag
- `owner/repo@47b927cbb41c4fdea1292baf` - Full commit hash
- `github.com/owner/repo@version` - Full URL with version
- `gitlab.com/group/repo@tag` - GitLab repository

```yaml
deps:
  # Latest commit from default branch
  - googleapis/googleapis
  
  # Specific tag (recommended for production)
  - googleapis/googleapis@v1.0.0
  
  # Full commit hash (most precise)
  - googleapis/googleapis@47b927cbb41c4fdea1292bafadb8976f
  
  # Different Git hosting
  - gitlab.com/acme/proto@v2.1.0
```

**Note:** If `@$VERSION` is omitted, EasyP downloads the latest commit from the repository's default branch.

### `generate`

**Optional.** Configures code generation from proto files.

**Type:** `object`
**Default:** `{}`

```yaml
generate:
  inputs:
    - directory: "proto"
    - git_repo:
        url: "github.com/acme/common@v1.0.0"
        sub_directory: "proto"
  plugins:
    - name: go
      out: .
      opts:
        paths: source_relative
    - name: go-grpc
      out: .
      opts:
        paths: source_relative
        require_unimplemented_servers: false
```

#### `generate.inputs`

**Optional.** Specifies sources of proto files for generation.

**Type:** `[]object`
**Default:** `[]`

```yaml
generate:
  inputs:
    # Local directory
    - directory: "proto"

    # Local directory with advanced options
    - directory:
        path: "api/proto"
        root: "."

    # Remote git repository
    - git_repo:
        url: "github.com/acme/common@v1.0.0"
        sub_directory: "proto"
        out: "generated"
```

**Directory input fields:**
- `directory` (string or object) - Local directory path
- `directory.path` (string) - Directory path
- `directory.root` (string) - Root path for import resolution (default: ".")

**Git repository input fields:**
- `git_repo.url` (string) - Repository URL with optional version
- `git_repo.sub_directory` (string) - Subdirectory within the repository
- `git_repo.out` (string) - Output directory for generated files

#### `generate.plugins`

**Optional.** Configures protoc plugins for code generation.

**Type:** `[]object`
**Default:** `[]`

```yaml
generate:
  plugins:
    # Local plugin
    - name: go
      out: .
      opts:
        paths: source_relative

    # Remote plugin
    - name: validate-go
      url: "buf.build/bufbuild/protovalidate-go:v0.4.0"
      out: gen/go
      opts:
        paths: source_relative

    # Plugin with import dependencies
    - name: grpc-gateway
      out: .
      with_imports: true
      opts:
        paths: source_relative
```

**Plugin fields:**
- `name` (string, required) - Plugin name (omit `protoc-gen-` prefix)
- `out` (string, required) - Output directory for generated files
- `opts` (map[string](string | []string), optional) - Plugin-specific options; each key can be a single value or an array of values
- `url` (string, optional) - Remote plugin URL for HTTP execution
- `with_imports` (boolean, optional) - Include imported dependencies

**Common plugin options:**
```yaml
# Go plugin options
opts:
  paths: source_relative              # Generate files relative to input
  module: github.com/acme/api        # Go module path

# gRPC Gateway options
opts:
  paths: source_relative
  grpc_api_configuration: api.yaml   # gRPC API configuration

# OpenAPI v2 options
opts:
  simple_operation_ids: true         # Use simple operation IDs
  generate_unbound_methods: false    # Skip unbound methods

# ts-proto options with repeated key values
opts:
  env: node
  outputServices:
    - grpc-js
    - generic-definitions
```

When an `opts` value is a list, EasyP serializes it as repeated plugin params, e.g. `outputServices=grpc-js,outputServices=generic-definitions`.

### `breaking`

**Optional.** Configures backward compatibility checking.

**Type:** `object`
**Default:** `{}`

```yaml
breaking:
  ignore:
    - proto/experimental/
    - proto/internal/
  against_git_ref: main
```

#### `breaking.ignore`

**Optional.** Directories or files to exclude from breaking change detection.

**Type:** `[]string`
**Default:** `[]`

```yaml
breaking:
  ignore:
    - proto/experimental/
    - proto/alpha/
    - testdata/
```

#### `breaking.against_git_ref`

**Optional.** Git reference (branch, tag, or commit) to compare against for breaking changes.

**Type:** `string`
**Default:** `""` (must be specified via CLI flag)

```yaml
breaking:
  against_git_ref: main
```

Can be overridden by the `--against` CLI flag.



## Configuration Examples

### Minimal Configuration

```yaml
lint:
  use:
    - MINIMAL
```

### Development Configuration

```yaml
lint:
  use:
    - BASIC
    - COMMENT_SERVICE
    - COMMENT_RPC
  allow_comment_ignores: true
  ignore:
    - vendor/
    - testdata/
deps:
  - github.com/googleapis/googleapis@v1.0.0
generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: .
      opts:
        paths: source_relative
    - name: go-grpc
      out: .
      opts:
        paths: source_relative
```

### Production Configuration

```yaml
lint:
  use:
    - MINIMAL
    - BASIC
    - DEFAULT
    - COMMENTS
  enum_zero_value_suffix: "UNSPECIFIED"
  service_suffix: "Service"
  ignore:
    - vendor/
  except: []
  allow_comment_ignores: false
deps:
  - github.com/googleapis/googleapis@v1.56.0
  - github.com/grpc-ecosystem/grpc-gateway@v2.18.0
generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: gen/go
      opts:
        paths: source_relative
        module: github.com/acme/api/gen/go
    - name: go-grpc
      out: gen/go
      opts:
        paths: source_relative
        require_unimplemented_servers: false
    - name: grpc-gateway
      out: gen/go
      opts:
        paths: source_relative
    - name: openapiv2
      out: gen/openapi
      opts:
        simple_operation_ids: true
breaking:
  ignore:
    - proto/experimental/
  against_git_ref: main
```

### Multi-Service Configuration

```yaml
lint:
  use:
    - BASIC
    - COMMENT_SERVICE
    - COMMENT_RPC
  service_suffix: "Service"
  ignore_only:
    COMMENT_FIELD:
      - proto/internal/
    SERVICE_SUFFIX:
      - proto/legacy/
deps:
  - github.com/googleapis/googleapis@v1.0.0
  - github.com/acme/common-proto@v2.1.0
generate:
  inputs:
    - directory: "proto/public"
    - directory: "proto/internal"
    - git_repo:
        url: "github.com/acme/shared-proto@v1.0.0"
        sub_directory: "proto"
  plugins:
    - name: go
      out: gen/go
      opts:
        paths: source_relative
    - name: go-grpc
      out: gen/go
      opts:
        paths: source_relative
    - name: grpc-gateway
      out: gen/go
      opts:
        paths: source_relative
    - name: validate-go
      url: "buf.build/bufbuild/protovalidate-go:v0.4.0"
      out: gen/go
      opts:
        paths: source_relative
breaking:
  ignore:
    - proto/internal/
    - proto/experimental/
  against_git_ref: develop
```

## Configuration Validation

EasyP validates configuration files on startup and provides helpful error messages:

```bash
# Invalid rule name
Error: invalid rule: INVALID_RULE_NAME

# Missing required field in generate section
Error: required field "plugins" is missing (path: generate.plugins)

# Invalid dependency format
Error: invalid dependency format: invalid-repo-url
```

Use `easyp --debug` for detailed validation information.

## Migration from Buf

EasyP is fully compatible with Buf configurations. To migrate:

1. Place `buf.yaml` or `buf.yml` in the project root
2. Run `easyp init` and confirm migration when prompted
3. Update `deps` format if using BSR modules
4. Review migrated lint/breaking settings and adjust as needed

Most Buf configurations work without changes in EasyP.
