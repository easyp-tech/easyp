# Go

## Установка плагинов

Сначала установите необходимые плагины для работы с gRPC:

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Эти команды установят плагины `protoc-gen-go` и `protoc-gen-go-grpc` для использования с EasyP.

## Пример proto‑сервиса

Ниже пример proto-файла для сервиса Echo:

```proto
syntax = "proto3";

package api.echo.v1;

option go_package = "github.com/easyp-tech/example/api/echo/v1;pb";

service EchoAPI {
  rpc Echo(EchoRequest) returns (EchoResponse);
  rpc EchoStream(EchoStreamRequest) returns (EchoResponse);
}

message EchoRequest {
  string payload = 1;
}

message EchoResponse {
  string payload = 2;
}

message EchoStreamRequest {
  string payload = 1;
}

message EchoStreamResponse {
  string payload = 2;
}
```

## Настройка конфигурации

Создайте и настройте файл конфигурации `easyp.yaml`:

```yaml
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
```

Этот файл указывает использование двух плагинов: `go` для генерации Go-кода и `go-grpc` для генерации gRPC-кода, вместе с их опциями.

## Генерация кода

Чтобы сгенерировать код, выполните команду:

```shell
easyp --cfg easyp.yaml generate
```

Если флаг `--cfg` не указан, по умолчанию используется файл `easyp.yaml` в текущей директории:

```shell
easyp generate
```

Теперь у вас есть сгенерированный Go‑код, с которым можно напрямую работать.