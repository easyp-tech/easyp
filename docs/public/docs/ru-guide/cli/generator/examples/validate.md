# Validate

## Установка плагинов

Сначала установите необходимые плагины для работы с gRPC и проверок (validate):

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/envoyproxy/protoc-gen-validate@latest
```

Эти команды установят плагины `protoc-gen-go`, `protoc-gen-go-grpc` и `protoc-gen-validate` для использования вместе с EasyP.

## Пример Proto сервиса

Ниже пример proto‑файла для сервиса Echo:

```proto
syntax = "proto3";

package api.echo.v1;

import "validate/validate.proto";

option go_package = "github.com/easyp-tech/example/api/echo/v1;pb";

service EchoAPI {
  rpc Echo(EchoRequest) returns (EchoResponse);
  rpc EchoStream(EchoStreamRequest) returns (EchoResponse);
}

message EchoRequest {
  string payload = 1 [(validate.rules).string = {max_len: 200}];
}

message EchoResponse {
  string payload = 2;
}

message EchoStreamRequest {
  string payload = 1 [(validate.rules).string = {max_len: 200}];
}

message EchoStreamResponse {
  string payload = 2;
}
```

## Настройка конфигурации

Создайте и настройте файл конфигурации `easyp.yaml`:

```yaml
version: v1alpha

deps: # [!code ++]
  - github.com/bufbuild/protoc-gen-validate  # [!code ++]

generate:
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
    - name: validate
      out: .
      opts:
        paths: source_relative
        lang: "go"
```

## Генерация кода

Чтобы сгенерировать код, выполните команду:

```shell
easyp -cfg easyp.yaml generate
```

Если флаг `-cfg` не указан, по умолчанию используется файл `easyp.yaml` из текущего каталога:

```shell
easyp generate
```

Теперь у вас есть сгенерированный Go‑код с поддержкой валидации, и вы можете напрямую использовать его в проекте.