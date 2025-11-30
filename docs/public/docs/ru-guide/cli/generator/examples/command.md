# Выполнение плагина через команду

EasyP поддерживает выполнение плагинов через произвольные команды. Это особенно полезно для запуска плагинов через `go run` без предварительной установки.

## Преимущества

- **Не требует установки**: Плагины запускаются напрямую через команду
- **Версионирование**: Можно указать конкретную версию плагина через `@version`
- **Гибкость**: Поддержка любых команд, не только `go run`

## Пример: gRPC Gateway через go run

Ниже пример использования `protoc-gen-grpc-gateway` через `go run`:

```yaml
version: v1alpha

deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.25.1

generate:
  inputs:
    - directory: "proto"
  plugins:
    # Go плагин (локальный)
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/mycompany/api
    
    # gRPC плагин (локальный)
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: source_relative
    
    # gRPC Gateway через go run
    - command: ["go", "run", "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.25.1"]
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true
```

## Пример: Validate через go run

Использование `protoc-gen-validate` через команду:

```yaml
version: v1alpha

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
    
    # Validate через go run
    - command: ["go", "run", "github.com/bufbuild/protoc-gen-validate/cmd/protoc-gen-validate-go@v0.10.1"]
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true
```

## Пример: Кастомная команда

Вы можете использовать любую команду, не только `go run`:

```yaml
generate:
  plugins:
    # Запуск через node
    - command: ["node", "/path/to/protoc-gen-custom.js"]
      out: ./gen/custom
    
    # Запуск через python
    - command: ["python3", "-m", "protoc_gen_tool"]
      out: ./gen/python
    
    # Запуск исполняемого файла
    - command: ["./tools/protoc-gen-custom"]
      out: ./gen/custom
```

## Приоритет источников плагина

EasyP использует следующий приоритет при выборе источника плагина:

1. **`command`** — выполнение через указанную команду (наивысший приоритет)
2. **`remote`** — удалённый плагин через URL
3. **`name`** — локальный плагин из PATH или встроенный плагин
4. **`path`** — путь к исполняемому файлу плагина

## Важные замечания

- **Только один источник**: Должен быть указан только один источник плагина (`name`, `command`, `remote` или `path`)
- **Версионирование**: При использовании `go run` с пакетом из GitHub, всегда указывайте версию через `@version` для воспроизводимости
- **Производительность**: Выполнение через `go run` медленнее, чем использование установленных плагинов, так как каждый раз происходит компиляция

## Генерация кода

После настройки конфигурации выполните:

```bash
easyp generate
```

EasyP автоматически выполнит указанные команды для генерации кода.

