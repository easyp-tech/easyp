# Генератор

[[toc]]

EasyP включает мощный генератор, который упрощает процесс получения кода из proto‑файлов. Благодаря YAML‑конфигурациям работа становится гораздо более удобной и интуитивной по сравнению с прямым использованием команды protoc.

## Ключевые особенности генератора

1. **Упрощённая генерация кода**:
    - Генерация кода из proto через декларативную `YAML`‑конфигурацию.
    - Избавляет от необходимости писать длинные и сложные команды protoc.

2. **Обёртка над protoc**:
    - EasyP выступает как удобный слой над protoc, предоставляя декларативный API.
    - Поддерживает все опции и плагины, доступные protoc.

3. **Гибкость и кастомизация**:
    - Используются те же параметры, что и у плагинов protoc, прямо в конфиге.
    - Поддерживается множество плагинов и их параметры в одной конфигурации.

4. **Генерация из нескольких источников**:
    - Локальные директории и удалённые Git‑репозитории одновременно.
    - Лёгкая интеграция с существующими проектами и репозиториями.

5. **Удалённая генерация**:
    - Генерация из удалённых Git репозиториев без локального checkout.

6. **Интеграция с менеджером пакетов**:
    - Прозрачная работа с зависимостями через пакетный менеджер EasyP.
    - Автоматическое разрешение и подключение proto‑зависимостей.

## Configuration Reference

### Complete Configuration Example

```yaml
# Package manager dependencies
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  - github.com/bufbuild/protoc-gen-validate@v0.10.1

# Code generation configuration
generate:
  inputs:
    # Local directory input
    - directory: 
        path: "proto"
        root: "."
    
    # Remote Git repository input
    - git_repo:
        url: "github.com/acme/weather@v1.2.3"
        sub_directory: "proto/api"
        out: "external"
    
    # Another remote repository
    - git_repo:
        url: "https://github.com/company/internal-protos.git"
        sub_directory: "definitions"
        out: "internal"

  plugins:
    # Local plugin execution
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/mycompany/myproject
      with_imports: true
    
    # Local plugin with custom options
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: source_relative
        require_unimplemented_servers: false

  # Managed mode - автоматическая установка file и field опций
  managed:
    enabled: true
    disable:
      - module: github.com/googleapis/googleapis  # Отключить для конкретного модуля
    override:
      - file_option: go_package_prefix
        value: github.com/mycompany/myproject/gen/go
      - file_option: java_package_prefix
        value: com.mycompany
      - file_option: csharp_namespace_prefix
        value: MyCompany
      - field_option: jstype
        value: JS_STRING
        path: api/v1/  # Применить к конкретному пути

```

### Input Sources

#### Локальный ввод (Local Directory Input)

Локальная директория — самый распространённый и простой способ указать откуда брать proto для генерации. Используйте этот вариант, если файлы уже находятся в репозитории проекта.

**Когда применять:**
- Proto файлы находятся в кодовой базе проекта
- Нужен полный контроль над структурой каталогов
- Проект — один сервис или приложение
- Файлы меняются не слишком часто

```yaml
inputs:
  - directory: "proto"                    # Простой строковый формат
  
  # Или подробный формат
  - directory:
      path: "proto"                       # Путь к proto файлам
      root: "."                           # Корень для разрешения import
```

Параметр `root` особенно полезен в монорепозиториях: он позволяет управлять тем, откуда будут резолвиться пути импортов. Если указать родительскую директорию, импорты будут оцениваться относительно неё, а не текущего рабочего каталога.

**Параметры:**

| Параметр | Тип | Обязателен | Значение по умолчанию | Описание |
|----------|-----|------------|-----------------------|----------|
| `path` | string | ✅ | - | Каталог, содержащий proto файлы |
| `root` | string | ❌ | `"."` | Корневая директория для разрешения import путей |

**Примеры:**

Ниже показана разница между базовым указанием директории и расширенной конфигурацией с пользовательским корнем импортов:

```yaml
# Basic usage - Simple path specification
inputs:
  - directory: "api/proto"

# Advanced usage with custom root - Controls import path resolution
inputs:
  - directory:
      path: "services/auth/proto" 
      root: "services/auth"        # Imports will be relative to this path
```

#### Удалённый Git репозиторий (Remote Git Repository Input)

Этот тип входных данных позволяет генерировать код из proto‑файлов, находящихся в удалённых репозиториях, без локального checkout. Особенно полезно при использовании API других команд или внешних поставщиков.

**Когда использовать:**
- Потребление proto определений из других сервисов/команд
- Интеграция с внешними вендорскими API
- Работа с общими библиотеками proto в нескольких проектах
- Необходимо гарантировать корректную (зафиксированную) версию внешних API

**Рекомендации:**
- В продакшене всегда фиксируйте версии (`@v1.0.0`, а не latest)
- Используйте семантические версии при наличии — проще сопровождать
- Предпочитайте публичные теги вместо хеша коммита (лучше отслеживаемость)

```yaml
inputs:
  - git_repo:
      url: "github.com/company/protos@v1.0.0"    # Обязательное: репозиторий + версия
      sub_directory: "api"                       # Опционально: поддиректория внутри репо
      out: "external"                            # Опционально: локальная директория для извлечения
```

Параметр `out` задаёт куда локально будут извлечены proto‑файлы. Полезно для организации нескольких удалённых источников и избежания конфликтов имен.

**Параметры:**

| Параметр | Тип | Обязателен | По умолчанию | Описание |
|----------|-----|------------|--------------|----------|
| `url` | string | ✅ | - | URL Git репозитория с необязательной версией / тегом / коммитом |
| `sub_directory` | string | ❌ | `""` | Поддиректория внутри репозитория, где лежат proto |
| `out` | string | ❌ | `""` | Локальная директория для извлечённых proto |

**Варианты формата URL:**

EasyP поддерживает несколько форматов для ссылки на удалённые репозитории. Каждый подходит под разные требования по стабильности:

- **Tagged versions** — для продакшена, стабильные неизменяемые ссылки
- **Semantic versions** — читаемость и управление зависимостями
- **Commit hashes** — доступ к конкретному коммиту, если нет тега
- **Latest** — только для разработки, непредсказуемо
- **Full HTTPS URLs** — приватные репозитории или нестандартный хостинг

```yaml
# Тег — стабильно для продакшена
url: "github.com/googleapis/googleapis@common-protos-1_3_1"

# Семантическая версия
url: "github.com/grpc-ecosystem/grpc-gateway@v2.19.1"  

# Хеш коммита — точечный фикс
url: "github.com/company/protos@abc123def456"

# Latest — только dev, НЕ для продакшена
url: "github.com/company/protos"

# Полный HTTPS URL — приватный или кастомный Git
url: "https://github.com/company/private-protos.git"
```

**Примеры:**

Ниже паттерны использования удалённых источников в разных сценариях:

```yaml
# Публичный репозиторий с фиксированной версией — типично для внешних API
inputs:
  - git_repo:
      url: "github.com/googleapis/googleapis@common-protos-1_3_1"
      sub_directory: "google"
      out: "googleapis"

# Приватный репозиторий с аутентификацией — внутренние API
inputs:
  - git_repo:
      url: "github.com/mycompany/internal-protos@v2.1.0"
      sub_directory: "api/definitions"
      out: "internal"

# Несколько удалённых источников — часто в микросервисной архитектуре
inputs:
  - git_repo:
      url: "github.com/grpc-ecosystem/grpc-gateway@v2.19.1"
      sub_directory: "protoc-gen-openapiv2/options"
      out: "gateway"
  - git_repo:
      url: "github.com/bufbuild/protoc-gen-validate@v0.10.1"  
      sub_directory: "validate"
      out: "validate"
```

### Конфигурация плагинов

Конфигурация плагинов определяет какие генераторы кода будут запущены и как они себя ведут. EasyP поддерживает любой плагин protoc, что делает его крайне гибким для разных языковых экосистем и сценариев.

На верхнем уровне есть **четыре способа указать, как именно запускать плагин**:

- **`name`** – запуск плагина по имени из `PATH` или использование встроенного плагина.
- **`path`** – запуск плагина по абсолютному/относительному пути к исполняемому файлу.
- **`remote`** – удалённый плагин по URL (через remote‑executor EasyP).
- **`command`** – запуск плагина через произвольную команду (например, `go run ...`).

Для каждого плагина должен быть указан **ровно один** из параметров `name`, `path`, `remote` или `command`.

#### Плагин по имени (`name`)

Локальный режим по имени — стандартный: плагины установлены в системе и запускаются напрямую через EasyP.

**Когда использовать `name`:**
- Стандартные языки (Go, Python, TypeScript и т.д.)
- Есть контроль над окружением сборки
- Критична производительность (нет сетевых задержек)
- Хочется опираться на `PATH` или встроенные плагины

**Требования к установке:**
- Плагины установлены и доступны в `PATH` (если не используются встроенные).
- Имена следуют шаблону `protoc-gen-{name}`.
- Установка через менеджеры пакетов (go install, npm install, pip install и т.п.).

```yaml
plugins:
  - name: go                              # Имя плагина (локальный или встроенный)
    out: ./generated                      # Директория для вывода
    opts:                                 # Опции конкретного плагина
      paths: source_relative
      module: github.com/mycompany/project
    with_imports: true                    # Включить импорт зависимостей
```

Параметр `with_imports` критичен при использовании зависимостей из пакетного менеджера: установите `true`, чтобы прототипы из секции `deps` попали в генерацию.

#### Плагин по пути (`path`)

Иногда нужно запускать плагин из конкретного бинарника, не добавляя его в `PATH` (например, бинарь в репозитории или в build‑директории). В этом случае можно указать явный путь:

```yaml
plugins:
  - path: ./bin/protoc-gen-my-custom
    out: ./gen/custom
    opts:
      foo: bar
```

**Когда использовать `path`:**
- Бинарники плагинов хранятся в репозитории (для воспроизводимых сборок).
- Используются несколько версий одного и того же плагина параллельно.
- Не хочется «загрязнять» глобальный `PATH`.

#### Удалённый плагин (`remote`)

Удалённые плагины выполняются по URL. EasyP отправляет на удалённую сторону `CodeGeneratorRequest` и получает обратно `CodeGeneratorResponse`.

Ниже приведён реальный пример использования EasyP API Service как удалённого исполнителя плагинов (тот же формат используется в разделе про API Service):

```yaml
generate:
  plugins:
    # Удалённое выполнение плагинов через EasyP API Service
    - remote: api.easyp.tech/protobuf/go:v1.36.10
      out: .
      opts:
        paths: source_relative

    - remote: api.easyp.tech/grpc/go:v1.5.1
      out: .
      opts:
        paths: source_relative
```

**Типичные сценарии для `remote`:**
- Централизованный сервис плагинов внутри компании (например, EasyP API Service).
- Запуск тяжёлых плагинов на выделенных машинах вместо CI‑агентов.
- Совместное использование одной реализации плагина разными командами.

#### Выполнение плагина через команду (`command`)

Вы можете указать плагин как массив команд для выполнения. Это полезно для запуска плагинов через `go run` или любые другие инструменты без предварительной установки бинарника плагина:

```yaml
plugins:
  - command: ["go", "run", "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.25.1"]
    out: ./gen/go
    opts:
      paths: source_relative
```

В этом режиме EasyP:
- **запускает указанную команду** как дочерний процесс;
- **пишет `CodeGeneratorRequest` в stdin** процесса;
- **читает `CodeGeneratorResponse` из stdout**, как и для обычных плагинов protoc.

**Приоритет источников плагина:**
1. `command` — выполнение через указанную команду (наивысший приоритет)
2. `remote` — удалённый плагин через URL
3. `name` — локальный плагин из PATH или встроенный плагин
4. `path` — путь к исполняемому файлу плагина

**Параметры (источники и общие опции плагина):**

| Параметр | Тип | Обязателен | По умолчанию | Описание |
|----------|-----|------------|--------------|----------|
| `name` | string | ❌ | - | Имя / идентификатор плагина (например, `go`, `go-grpc`, `grpc-gateway`) |
| `command` | []string | ❌ | - | Команда для выполнения плагина (например, `["go", "run", "package"]`) |
| `remote` | string | ❌ | - | URL удалённого плагина |
| `path` | string | ❌ | - | Путь к исполняемому файлу плагина |
| `out` | string | ✅ | - | Директория для сгенерированных файлов |
| `opts` | map[string](string \| []string) | ❌ | `{}` | Специфичные опции плагина; значения-списки передаются как повторяющиеся `key=value` |
| `with_imports` | bool | ❌ | `false` | Включать proto из зависимостей |

**Примечание:** Для каждого плагина должен быть указан ровно один источник (`name`, `command`, `remote` или `path`).
Если `opts.outputServices` задан как `["grpc-js", "generic-definitions"]`, EasyP передаст `outputServices=grpc-js,outputServices=generic-definitions`.

**Примеры использования источника `command`:**

```yaml
generate:
  inputs:
    - directory: "proto"

  plugins:
    # 1) gRPC-Gateway через go run (без предварительной установки бинарника)
    - command: ["go", "run", "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.25.1"]
      out: ./gen/go
      opts:
        paths: source_relative
        generate_unbound_methods: true

    # 2) protoc-gen-validate через go run
    - command: ["go", "run", "github.com/bufbuild/protoc-gen-validate@v0.10.1"]
      out: ./gen/go
      opts:
        paths: source_relative
        lang: go

    # 3) Произвольный кастомный скрипт-обёртка
    - command: ["bash", "./scripts/custom-protoc-plugin.sh"]
      out: ./gen/custom
      opts:
        foo: bar
```

#### Встроенные плагины (Builtin Plugins)

EasyP поддерживает встроенные плагины для базовых языков protobuf и gRPC. Эти плагины встроены в бинарник как WASM-модули и не требуют установки внешних зависимостей.

**Преимущества встроенных плагинов:**
- **Портативность**: Один бинарник со всеми необходимыми плагинами
- **Удобство**: Не требуется установка внешних зависимостей
- **Стабильность**: Версии плагинов зафиксированы в бинарнике
- **Изоляция**: Не зависит от системных установок плагинов

**Поддерживаемые встроенные плагины:**

#### Protobuf базовые плагины

Следующие плагины встроены для генерации базового protobuf кода:

| Имя плагина | Описание | Соответствующий protoc плагин |
|-------------|----------|-------------------------------|
| `cpp` | Генерация C++ кода из proto файлов | `protoc-gen-cpp` |
| `csharp` | Генерация C# кода из proto файлов | `protoc-gen-csharp` |
| `java` | Генерация Java кода из proto файлов | `protoc-gen-java` |
| `kotlin` | Генерация Kotlin кода из proto файлов | `protoc-gen-kotlin` |
| `objc` | Генерация Objective-C кода из proto файлов | `protoc-gen-objc` |
| `php` | Генерация PHP кода из proto файлов | `protoc-gen-php` |
| `python` | Генерация Python кода из proto файлов | `protoc-gen-python` |
| `ruby` | Генерация Ruby кода из proto файлов | `protoc-gen-ruby` |

#### gRPC плагины

Следующие плагины встроены для генерации gRPC кода:

| Имя плагина | Описание | Соответствующий protoc плагин |
|-------------|----------|-------------------------------|
| `grpc_cpp` | Генерация gRPC кода для C++ | `grpc_cpp_plugin` |
| `grpc_csharp` | Генерация gRPC кода для C# | `grpc_csharp_plugin` |
| `grpc_java` | Генерация gRPC кода для Java | `grpc_java_plugin` |
| `grpc_node` | Генерация gRPC кода для Node.js | `grpc_node_plugin` |
| `grpc_objc` | Генерация gRPC кода для Objective-C | `grpc_objective_c_plugin` |
| `grpc_php` | Генерация gRPC кода для PHP | `grpc_php_plugin` |
| `grpc_python` | Генерация gRPC кода для Python | `grpc_python_plugin` |
| `grpc_ruby` | Генерация gRPC кода для Ruby | `grpc_ruby_plugin` |

**Логика выбора плагина:**

EasyP использует следующий приоритет при выборе executor'а для плагина:

1. **Удалённый плагин** (если указан `url`) — всегда имеет наивысший приоритет
2. **Встроенный плагин** (если плагин встроен и не найден в PATH) — используется автоматически
3. **Локальный плагин** (из PATH) — используется по умолчанию для обратной совместимости

**Пример использования:**

```yaml
generate:
  inputs:
    - directory: "proto"
  plugins:
    # Встроенный плагин Python (автоматически используется, если protoc-gen-python не найден в PATH)
    - name: python
      out: ./gen/python
      opts:
        pyi_out: ./gen/python
    
    # Встроенный gRPC плагин для Python
    - name: grpc_python
      out: ./gen/python
    
    # Встроенный плагин C++ (автоматически используется, если protoc-gen-cpp не найден в PATH)
    - name: cpp
      out: ./gen/cpp
      opts:
        dllexport_decl: EXPORT
    
    # Встроенный gRPC плагин для C++
    - name: grpc_cpp
      out: ./gen/cpp
```

**Требования:**

Встроенные плагины включены в бинарник EasyP:

```bash
# Сборка
go build ./cmd/easyp

# Установка
go install github.com/easyp-tech/easyp/cmd/easyp@latest
```

**Обратная совместимость:**

Встроенные плагины полностью совместимы с существующими конфигурациями. Если плагин найден в PATH, он будет использован вместо встроенного. Это гарантирует, что:

- Существующие конфигурации продолжают работать без изменений
- Можно переопределить встроенный плагин, установив его в системе
- Приоритет отдаётся локальным установкам для гибкости

### Справочник опций плагинов

Ниже перечислены наиболее часто используемые плагины и их настройки. Каждый плагин имеет параметры, влияющие на результат генерации — понимание этих опций важно для получения нужного вывода.

#### Go Plugins

Плагины для Go — самые зрелые и распространённые. Опция `paths` управляет тем, как формируются пути импортов; остальные опции дают тонкую настройку вывода.

```yaml
plugins:
  # protoc-gen-go — генерирует структуры и базовые protobuf функции
  - name: go
    out: ./gen/go
    opts:
      paths: source_relative              # source_relative | import
      module: github.com/company/project  # Go module path для импорта
      
  # protoc-gen-go-grpc — генерирует gRPC серверные/клиентские заглушки
  - name: go-grpc
    out: ./gen/go
    opts:
      paths: source_relative
      require_unimplemented_servers: false  # Включение UnimplementedServer
```

#### gRPC-Gateway Plugins

Позволяют экспонировать gRPC сервисы как REST API и генерировать OpenAPI. Критично для HTTP/JSON поверх gRPC.

```yaml
plugins:
  # protoc-gen-grpc-gateway — reverse proxy REST→gRPC
  - name: grpc-gateway
    out: ./gen/go
    opts:
      paths: source_relative
      generate_unbound_methods: true      # Включать методы без HTTP привязок
      
  # protoc-gen-openapiv2 — генерирует OpenAPI/Swagger
  - name: openapiv2  
    out: ./gen/openapi
    opts:
      simple_operation_ids: true          # Простые operationId
      generate_unbound_methods: false     # Исключить методы без HTTP аннотаций
      json_names_for_fields: true         # Использовать JSON имена
```

#### Validation Plugins

Генерируют код валидации на основе правил в proto, устраняя ручную проверку.

```yaml
plugins:
  # protoc-gen-validate — генерация проверок полей
  - name: validate-go
    out: ./gen/go
    opts:
      paths: source_relative
      lang: go                           # Целевой язык
```

#### TypeScript/JavaScript Plugins

TypeScript плагины нужны фронтенду: типобезопасные интерфейсы и клиенты.

```yaml
plugins:
  # protoc-gen-ts — типы и сериализация
  - name: ts
    out: ./gen/typescript  
    opts:
      declaration: true                   # Генерация .d.ts
      target: es2017                      # Целевой ECMAScript
      
  # protoc-gen-grpc-web — gRPC-Web клиенты для браузера
  - name: grpc-web
    out: ./gen/web
    opts:
      import_style: typescript           # Стиль импорта
      mode: grpcweb                      # Режим транспорта
```

## Managed Mode

Managed mode автоматически устанавливает file и field опции в protobuf дескрипторах во время генерации кода без изменения исходных `.proto` файлов. Эта функция совместима с managed mode в `buf` и обеспечивает единообразный способ управления языково-специфичными опциями в вашей кодовой базе.

**Ключевые преимущества:**
- **Без изменений proto файлов**: Опции применяются во время генерации, proto файлы остаются чистыми
- **Согласованные значения по умолчанию**: Автоматическое применение языковых соглашений об именовании
- **Централизованная конфигурация**: Управление всеми опциями в одном месте (`easyp.yaml`)
- **Правила для модулей**: Применение разных опций к разным модулям или путям
- **Совместимость с buf**: Работает так же, как managed mode в `buf`

### Как это работает

Когда managed mode включён, EasyP автоматически применяет file и field опции к protobuf дескрипторам перед генерацией кода. Это происходит в памяти, поэтому исходные `.proto` файлы остаются неизменными.

**Значения по умолчанию** применяются для определённых опций на основе языковых соглашений:
- Java: `java_package_prefix` по умолчанию `"com"`, `java_multiple_files` по умолчанию `true`
- C#: `csharp_namespace` по умолчанию PascalCase имени пакета
- Ruby: `ruby_package` по умолчанию PascalCase с разделителем `::`
- PHP: `php_namespace` по умолчанию PascalCase с разделителем `\`
- Objective-C: `objc_class_prefix` по умолчанию первые буквы частей пакета
- C++: `cc_enable_arenas` по умолчанию `true`

**Overrides** позволяют установить конкретные значения для опций с поддержкой фильтрации по модулю, пути или полю.

**Disables** позволяют предотвратить изменение managed mode определённых опций или файлов.

### Конфигурация

```yaml
generate:
  managed:
    enabled: true
    disable:
      # Отключить managed mode для конкретного модуля
      - module: github.com/googleapis/googleapis
      
      # Отключить конкретную опцию глобально
      - file_option: java_package_prefix
      
      # Отключить для конкретного пути
      - path: legacy/
        file_option: go_package
      
      # Отключить field опцию для конкретного поля
      - field_option: jstype
        field: com.example.User.id
    
    override:
      # Переопределить go_package_prefix для всех файлов
      - file_option: go_package_prefix
        value: github.com/mycompany/myproject/gen/go
      
      # Переопределить для конкретного модуля
      - file_option: java_package_prefix
        value: com.mycompany
        module: github.com/mycompany/internal-protos
      
      # Переопределить для конкретного пути
      - file_option: csharp_namespace_prefix
        value: MyCompany
        path: api/v1/
      
      # Переопределить для нескольких файлов с одинаковым значением используя префиксный путь
      # Это совпадет и с internal/cms/bmi.proto, и с internal/cms/bmi_service.proto
      - file_option: go_package
        value: spec/cms/bmi
        path: "internal/cms/bmi"
      
      # Переопределить field опцию для конкретного пути
      - field_option: jstype
        value: JS_STRING
        path: api/v1/
      
      # Переопределить для конкретного поля
      - field_option: jstype
        value: JS_NUMBER
        field: com.example.User.big_id
```

### Поддерживаемые File Options

| Опция | Описание | Есть значение по умолчанию? |
|-------|----------|----------------------------|
| `go_package` | Go import path | ❌ |
| `go_package_prefix` | Префикс для Go import paths | ❌ |
| `java_package` | Имя Java пакета | ❌ |
| `java_package_prefix` | Префикс для Java пакетов | ✅ (`"com"`) |
| `java_package_suffix` | Суффикс для Java пакетов | ❌ |
| `java_multiple_files` | Генерировать несколько Java файлов | ✅ (`true`) |
| `java_outer_classname` | Имя внешнего класса | ✅ (PascalCase + "Proto") |
| `java_string_check_utf8` | UTF-8 валидация | ❌ |
| `csharp_namespace` | C# namespace | ✅ (PascalCase) |
| `csharp_namespace_prefix` | Префикс для C# namespaces | ❌ |
| `ruby_package` | Имя Ruby модуля | ✅ (PascalCase с `::`) |
| `ruby_package_suffix` | Суффикс для Ruby пакетов | ❌ |
| `php_namespace` | PHP namespace | ✅ (PascalCase с `\`) |
| `php_metadata_namespace` | PHP metadata namespace | ❌ |
| `php_metadata_namespace_suffix` | Суффикс для PHP metadata | ❌ |
| `objc_class_prefix` | Objective-C префикс класса | ✅ (Первые буквы) |
| `swift_prefix` | Swift префикс | ❌ |
| `optimize_for` | Оптимизация генерации кода | ❌ |
| `cc_enable_arenas` | C++ arena аллокация | ✅ (`true`) |

### Поддерживаемые Field Options

| Опция | Описание | Применяется к |
|-------|----------|---------------|
| `jstype` | JavaScript тип для 64-битных целых чисел | `int64`, `uint64`, `sint64`, `fixed64`, `sfixed64` |

### Примеры

#### Базовая настройка со значениями по умолчанию

Включите managed mode для получения автоматических значений по умолчанию для всех поддерживаемых языков:

```yaml
generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: ./gen/go
    - name: java
      out: ./gen/java
    - name: csharp
      out: ./gen/csharp
  managed:
    enabled: true
```

Это автоматически установит:
- `java_package` в `com.<package>` для всех файлов
- `java_multiple_files` в `true`
- `csharp_namespace` в PascalCase имени пакета
- `ruby_package` в PascalCase с разделителем `::`
- И многое другое...

#### Кастомный префикс Go пакета

Переопределите префикс Go пакета для вашего проекта:

```yaml
generate:
  managed:
    enabled: true
    override:
      - file_option: go_package_prefix
        value: github.com/mycompany/myproject/gen/go
```

Это установит `go_package` в `github.com/mycompany/myproject/gen/go/<package>` для всех файлов.

#### Динамические пути Go пакетов с маркерами

Для более сложной генерации путей можно использовать маркеры в значениях `go_package_prefix` или `go_package`:

```yaml
generate:
  managed:
    enabled: true
    override:
      # Использовать путь файла напрямую (без расширения .proto)
      - file_option: go_package_prefix
        value: github.com/mycompany/{{file_path}}
      
      # Использовать только путь директории
      - file_option: go_package_prefix
        value: github.com/mycompany/{{file_dir}}
      
      # Удалить префикс из пути директории
      - file_option: go_package_prefix
        value: github.com/mycompany/{{file_dir_without:internal/}}
      
      # Удалить префикс из полного пути файла
      - file_option: go_package_prefix
        value: github.com/mycompany/{{file_path_without:internal/}}
```

**Доступные маркеры:**
- `{{file_path}}` - Полный путь файла без расширения `.proto`
  - Пример: `internal/cms/as.proto` → `internal/cms/as`
- `{{file_dir}}` - Только путь директории, без имени файла
  - Пример: `internal/cms/as.proto` → `internal/cms`
- `{{file_dir_without:prefix/}}` - Путь директории с удалением префикса и базового имени файла без суффиксов `_service`/`_grpc`
  - Пример: `{{file_dir_without:internal/}}` для `internal/cms/as_service.proto` → `cms/as`
- `{{file_path_without:prefix/}}` - Полный путь файла с удалением префикса
  - Пример: `{{file_path_without:internal/}}` для `internal/cms/as.proto` → `cms/as`

#### Переопределения для конкретных модулей

Примените разные опции к разным модулям:

```yaml
generate:
  managed:
    enabled: true
    override:
      # Значение по умолчанию для всех файлов
      - file_option: go_package_prefix
        value: github.com/mycompany/myproject/gen/go
      
      # Конкретное переопределение для внутреннего модуля
      - file_option: go_package_prefix
        value: github.com/mycompany/internal/gen/go
        module: github.com/mycompany/internal-protos
```

#### Отключение для внешних зависимостей

Отключите managed mode для внешних зависимостей, у которых уже установлены опции:

```yaml
generate:
  managed:
    enabled: true
    disable:
      - module: github.com/googleapis/googleapis
      - module: github.com/grpc-ecosystem/grpc-gateway
```

#### Типобезопасность JavaScript

Установите `jstype` в `JS_STRING` для всех 64-битных целочисленных полей, чтобы предотвратить потерю точности в JavaScript:

```yaml
generate:
  managed:
    enabled: true
    override:
      - field_option: jstype
        value: JS_STRING
        path: api/v1/  # Применить к конкретному пути
```

### Совпадение путей

Совпадение путей в managed mode использует префиксное совпадение (как в `buf`):

- **Путь директории** (заканчивается на `/`): Совпадает со всеми файлами в этой директории и поддиректориях
  - Пример: `path: "internal/cms/"` совпадает с `internal/cms/as.proto`, `internal/cms/node.proto`, `internal/cms/v1/service.proto`
- **Точный путь файла** (заканчивается на `.proto`): Совпадает только с этим конкретным файлом
  - Пример: `path: "internal/cms/as.proto"` совпадает только с `internal/cms/as.proto`
- **Префиксный путь** (без завершающего `/` или `.proto`): Использует префиксное совпадение (не директориально-осознанное)
  - Пример: `path: "internal/cms"` совпадает с `internal/cms/as.proto`, но также с `internal/cmsv2/file.proto`
  - **Совет**: Используйте префиксные пути для группировки нескольких файлов с одинаковым значением. Например, `path: "internal/cms/bmi"` совпадает и с `internal/cms/bmi.proto`, и с `internal/cms/bmi_service.proto`, что позволяет установить одно и то же значение `go_package` для обоих файлов одним правилом.

### Приоритет правил

Когда несколько правил соответствуют одному файлу или полю, применяется следующий приоритет:

1. **Disable правила** имеют приоритет — если опция отключена, она не будет применена
2. **Override правила** применяются по порядку — последнее совпадающее правило побеждает
3. **Значения по умолчанию** применяются только если нет совпадающего override и опция не отключена

### Совместимость с buf

Managed mode в EasyP совместим с managed mode в `buf`. Тот же формат конфигурации и поведение применяются, что упрощает миграцию между инструментами или использование обоих в одном workflow.

## Генерация Descriptor Set

**https://protobuf.dev/programming-guides/techniques/#self-description**

EasyP поддерживает генерацию бинарных FileDescriptorSet файлов с помощью флага `--descriptor_set_out`. Это позволяет создавать самоописывающиеся protobuf сообщения, которые включают информацию о схеме вместе с данными.

**Флаги CLI:**

- `--descriptor_set_out <путь>` - Путь для вывода бинарного FileDescriptorSet
- `--include_imports` - Включить все транзитивные зависимости в FileDescriptorSet

**Пример:**

```bash
# Генерация descriptor set только с целевыми файлами
easyp generate --descriptor_set_out=./schema.pb

# Генерация descriptor set со всеми зависимостями
easyp generate --descriptor_set_out=./schema.pb --include_imports
```

Самоописывающиеся сообщения полезны для динамического парсинга сообщений, валидации схемы во время выполнения, реестров схем и создания универсальных gRPC клиентов. Для получения дополнительной информации см. [документацию Protocol Buffers о самоописании](https://protobuf.dev/programming-guides/techniques/#self-description).

## Интеграция с менеджером пакетов

Одной из самых мощных возможностей EasyP является бесшовная интеграция генератора кода с менеджером пакетов. Это устраняет проблему ручного управления proto‑зависимостями и гарантирует, что при генерации используются корректные (зафиксированные) версии импортируемых файлов.

**Ключевые преимущества:**
- **Автоматическое разрешение зависимостей**: не нужно вручную прописывать пути импортов
- **Согласованность версий**: фиксация в `easyp.lock` гарантирует воспроизводимость
- **Транзитивные зависимости**: вложенные цепочки подтягиваются автоматически
- **Производительность**: локальный кеш — скачивание один раз и повторное использование

### Автоматическое разрешение зависимостей

Когда вы указываете зависимости в секции `deps`, генератор автоматически добавляет их в путь импортов. Ваши proto‑файлы могут делать стандартные `import` из этих зависимостей без дополнительной настройки.

**Как это работает:**
1. EasyP скачивает и кеширует зависимости из `deps`
2. Во время генерации кешированные файлы автоматически добавляются в import‑path protoc
3. Ваши proto‑файлы используют обычные import строки для обращения к зависимостям
4. При `with_imports: true` в вывод попадают и локальные файлы, и файлы зависимостей

Ниже простой пример — зависимости указываются один раз в `deps`, далее они доступны при генерации автоматически:

```yaml
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1

generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true    # This automatically includes googleapis and grpc-gateway protos
```

### Примеры использования зависимостей

Ниже показаны распространённые способы интеграции внешних proto‑зависимостей в процесс генерации.

#### Использование Google APIs

Google APIs — одна из самых популярных коллекций зависимостей: стандартные типы, аннотации для REST, проверки и общие структуры данных.

**Когда использовать Google APIs:**
- Построение REST поверх gRPC (gRPC-Gateway)
- Нужны стандартные типы (`Timestamp`, `Duration`, `Any`)
- Требуются аннотации поведения полей (field behavior)
- Интеграция с сервисами Google Cloud

Минимальная конфигурация для подключения Google APIs:

```yaml
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1

generate:
  inputs:
    - directory: "api/proto"
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true
```

**Важно:** Всегда фиксируйте версию (например `common-protos-1_3_1`) — не используйте latest в продакшене.

После настройки ваши proto могут импортировать определения Google API. Ниже пример сервиса с HTTP аннотацией:

```proto
// api/proto/service.proto
syntax = "proto3";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";

service MyService {
  rpc GetData(GetDataRequest) returns (GetDataResponse) {
    option (google.api.http) = {
      get: "/v1/data"
    };
  }
}
```

#### Использование правил валидации

`protoc-gen-validate` даёт мощную проверку полей прямо в определениях proto, устраняя необходимость писать отдельную логику валидации в приложении.

**Когда применять валидацию:**
- Проверка входных данных API
- Ограничения моделей БД
- Проверка конфигурационных файлов
- Любая ситуация, где критична целостность данных

**Преимущества:**
- Правила валидации — часть proto (единый источник истины)
- Генерация создаёт функции проверки автоматически
- Единообразие проверок между языками
- Производительнее чем проверка через runtime reflection

```yaml
deps:
  - github.com/bufbuild/protoc-gen-validate@v0.10.1

generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true
    - name: validate-go
      out: ./gen/go  
      opts:
        paths: source_relative
```

**Примечание:** Нужны одновременно зависимость (для импортов proto) и плагин (для генерации кода), чтобы получить полноценную поддержку валидации.

Так выглядят правила валидации в proto — сгенерированный код будет автоматически их применять:

```proto
// proto/user.proto
syntax = "proto3";

import "validate/validate.proto";

message User {
  string email = 1 [(validate.rules).string.email = true];
  int32 age = 2 [(validate.rules).int32.gte = 0];
}
```

#### Комплексная конфигурация с несколькими зависимостями

Пример ниже показывает продакшен‑конфигурацию, объединяющую несколько зависимостей и плагины для полного цикла разработки API:

```yaml
deps:
  # Core Google APIs - Standard types and HTTP annotations
  - github.com/googleapis/googleapis@common-protos-1_3_1
  
  # gRPC Gateway for REST APIs - Enables HTTP/JSON interfaces
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  
  # Validation rules - Field-level validation constraints
  - github.com/bufbuild/protoc-gen-validate@v0.10.1
  
  # Company internal shared types - Common business objects
  - github.com/mycompany/shared-protos@v1.5.0

generate:
  inputs:
    - directory: "api/proto"
  plugins:
    # Go code generation - Core protobuf structures
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/mycompany/myservice
      with_imports: true
      
    # gRPC service stubs - Server and client interfaces
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: source_relative
        require_unimplemented_servers: false
        
    # REST Gateway - HTTP-to-gRPC proxy code
    - name: grpc-gateway
      out: ./gen/go
      opts:
        paths: source_relative
        
    # OpenAPI documentation - API specification
    - name: openapiv2
      out: ./gen/openapi
      opts:
        simple_operation_ids: true
        
    # Validation code - Input validation functions
    - name: validate-go
      out: ./gen/go
      opts:
        paths: source_relative
```

### Интеграция с кешем зависимостей

Генератор использует модульный кеш EasyP для ускорения сборок:

```bash
# Download dependencies once
easyp mod download

# Generate code (uses cached dependencies)
easyp generate

# Dependencies are cached in ~/.easyp/mod/
ls ~/.easyp/mod/github.com/googleapis/googleapis/
```

## Удалённая генерация

Удалённая генерация — мощная возможность, позволяющая генерировать код из proto‑файлов в удалённых Git‑репозиториях без локального checkout. Это поддерживает микросервисную архитектуру, где команды используют API друг друга без жёсткой связки.

**Преимущества:**
- **Развязка разработки**: Команды развиваются независимо, используя версии API друг друга
- **Контроль версий**: Фиксация внешних API гарантирует стабильность
- **Меньше размер репозитория**: Нет необходимости в vendoring/submodule внешних proto
- **Автоматические обновления**: Легко перейти на новую версию по готовности

**Рекомендации:**
- В продакшене используйте только тегированные версии
- В разработке можно тестировать latest, но в проде фиксируйте
- При наличии используйте семантическое версионирование
- Учитывайте сетевые ограничения в CI/CD

### Источники удалённых proto

Генерация напрямую из удалённых репозиториев — полезно в микросервисной архитектуре, где разные команды владеют разными proto.

**Типовой workflow:**
1. Команда A публикует proto с версиями в Git
2. Команда B добавляет их в свой `easyp.yaml`
3. При генерации EasyP скачивает и подключает удалённые proto
4. Сгенерированный код содержит клиентские библиотеки сервисов команды A

Практический пример объединения локальных и удалённых источников в одной конфигурации генерации:

```yaml
generate:
  inputs:
    # Local protos - Your service's own API definitions
    - directory: "proto"
    
    # Remote public repository - External vendor API
    - git_repo:
        url: "github.com/acme/weather-api@v2.1.0"
        sub_directory: "proto/weather/v1"
        out: "external/weather"
    
    # Remote private repository - Internal company API
    - git_repo:
        url: "github.com/mycompany/internal-apis@main"
        sub_directory: "user-service/proto"
        out: "internal/user"
        
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true
```



### Сценарии использования удалённой генерации

Примеры ниже показывают реальные ситуации, где удалённая генерация даёт ощутимую выгоду.

#### Много команд (Multi-Team Development)

Мультикомандная разработка особенно выигрывает: вместо координации общих репозиториев и сложного управления зависимостями — независимая эволюция API с контролируемыми обновлениями через версии.

Паттерн ценен в крупных организациях, где:
- Команды имеют разные циклы релизов и скорости разработки
- Владение API определено, потребление — массовое  
- Нужно избежать overhead от координации общих proto репозиториев
- Команды используют разные стеки, но должны взаимодействовать

```yaml
# Team A (Order Service) generates from Team B's proto definitions
generate:
  inputs:
    # Local service definitions - APIs owned by this team
    - directory: "proto/orders"
    
    # User service protos from another team - Stable, versioned API
    - git_repo:
        url: "github.com/company/user-service@v1.8.0"  
        sub_directory: "api/proto"
        out: "external/users"
        
    # Payment service protos - Different team, different version
    - git_repo:
        url: "github.com/company/payment-service@v2.3.1"
        sub_directory: "proto/payment/v2"  
        out: "external/payments"
        
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/company/order-service
      with_imports: true
```

#### Интеграция с API вендоров

Многие вендоры публикуют proto для своих API, позволяя генерировать типобезопасных клиентов вместо ручных HTTP реализаций. Это повышает типовую строгость, автоматизирует сериализацию и часто ускоряет работу.

Преимущества использования proto от вендоров:
- **Type safety**: Проверка типов на этапе компиляции
- **Automatic updates**: Новые возможности приходят с обновлением версии
- **Consistency**: Единые паттерны интерфейсов между интеграциями
- **Performance**: Бинарная сериализация быстрее JSON
- **Documentation**: Proto выступают источником истины для API

```yaml
# Generate clients for external vendor APIs
generate:
  inputs:
    # Vendor's public proto definitions - Financial services
    - git_repo:
        url: "github.com/stripe/stripe-proto@v1.0.0"
        sub_directory: "proto"
        out: "vendor/stripe"
        
    # Communication service APIs - SMS/Voice integration
    - git_repo:  
        url: "github.com/twilio/twilio-protos@v2.1.0"
        sub_directory: "definitions"
        out: "vendor/twilio"
        
  plugins:
    - name: go
      out: ./clients/go
      opts:
        paths: source_relative
        module: github.com/mycompany/integrations
```



## Команды

CLI EasyP предоставляет гибкие варианты запуска генерации с разными конфигурациями и окружениями.

### Базовая генерация

Ниже наиболее часто применяемые варианты команд для разработки и продакшена:

```bash
# Use default easyp.yaml configuration - Most common for development
easyp generate

# Use custom configuration file - Essential for multi-environment setups  
easyp -cfg production.easyp.yaml generate

# Generate with verbose output - Helpful for debugging and CI/CD
easyp -v generate

# Generate with custom cache location - Useful for CI systems or shared environments
EASYPPATH=/tmp/easyp-cache easyp generate
```

### Интеграция с менеджером пакетов

Интеграция с менеджером пакетов позволяет либо явно управлять зависимостями, либо поручить это генератору. Явный путь даёт больше контроля, автоматический — больше удобства:

```bash
# Explicit workflow - Better for CI/CD and when you want to cache dependencies
easyp mod download    # Download and cache dependencies first
easyp generate        # Generate code using cached dependencies

# Automatic workflow - Convenient for development (generate downloads dependencies automatically)
easyp generate
```

### Расширенное использование (Advanced Usage)

Эти варианты полезны для специфических сценариев деплоя, отладки или когда нужен более тонкий контроль процесса генерации:

```bash
# Генерация из конкретной директории (перекрывает настройку в файле конфигурации)
easyp generate --input-dir=./api/proto

# Генерация с vendored зависимостями (оффлайн / Docker контейнеры)
easyp mod vendor
easyp -I easyp_vendor generate

# Использование кастомного protoc (например более новой версии)
PROTOC_PATH=/usr/local/bin/protoc easyp generate
```

## Типовые паттерны (Common Patterns)

Эти схемы отражают реальные сценарии и лучшие практики организации генерации кода в разных структурах проектов.

### Мульти-языковая генерация (Multi-Language Generation)

Мульти-языковая генерация критична для компаний, использующих разные технологии в стеке. EasyP облегчает получение согласованных клиентских библиотек и типов для нескольких языков из одного источника proto.

**Типовые сценарии:**
- **Full‑stack приложения**: backend на Go/Java + фронтенд на TypeScript
- **Платформы данных**: сервисы на Go + аналитика / ML на Python  
- **Микросервисы**: каждый сервис — оптимальный язык под задачу
- **Внешние SDK**: предоставление клиентских библиотек партнёрам
- **Интеграция с легаси**: современные gRPC сервисы + старые системы

**Производительность:**
- Каждый плагин запускается отдельно — время растёт линейно от количества плагинов
- Для больших наборов используйте параллельный запуск (`make -j4`)
- Структурируйте выходные директории иерархически — меньше конфликтов
- Плагины имеют разную скорость — профилируйте сборку для поиска узких мест

**Преимущества сопровождения:**
- Единый источник истины для API снижает риск расхождения схем
- Согласованные типы уменьшают интеграционные баги
- Автоматическая синхронизация при изменении proto исключает ручные правки
- Меньше риска дрейфа API между реализациями на разных языках
- Рефакторинг упрощается — изменения транзитивно отражаются во всех артефактах

Ниже пример типичной мульти-языковой конфигурации для full‑stack приложения: backend, веб‑клиент, аналитика и документация:

```yaml
generate:
  inputs:
    - directory: "proto"
    
  plugins:
    # Go backend services - Primary implementation language
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/company/backend
      with_imports: true
      
    - name: go-grpc  # gRPC server and client stubs
      out: ./gen/go
      opts:
        paths: source_relative
        
    # TypeScript frontend - Web application client code
    - name: ts  
      out: ./gen/typescript
      opts:
        declaration: true       # Generate type definitions
        target: es2020         # Modern JavaScript for browsers
        
    # Python data science - Analytics and ML workflows
    - name: python
      out: ./gen/python
      opts:
        mypy_out: ./gen/python-stubs  # Type checking support
        
    # Documentation - API reference for developers
    - name: doc
      out: ./docs/api
      opts:
        markdown: true         # Generate markdown documentation
```

**Совет по организации:** Используйте отдельные выходные директории для каждого языка, чтобы избежать конфликтов файлов и упростить интеграцию с языковыми сборочными системами.





Генератор EasyP — комплексное решение для генерации кода из Protocol Buffers: от простой локальной разработки до сложных enterprise‑сценариев с удалёнными зависимостями и мульти-языковым выводом.
