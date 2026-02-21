# gRPC-Gateway

## Установка плагинов

В дополнение к плагинам для работы с gRPC, необходимо установить следующие плагины для gRPC-Gateway:

    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

Эти команды установят плагины `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-openapiv2` и
`protoc-gen-grpc-gateway` для использования с EasyP.

## Пример Proto сервиса

Ниже — исходный proto‑файл для сервиса Echo:

    syntax = "proto3";

    package api.echo.v1;

    option go_package = "github.com/easyp-tech/example/api/echo/v1;pb";

    service EchoAPI {
      rpc Echo(EchoRequest) returns (EchoResponse);
    }

    message EchoRequest {
      string payload = 1;
    }

    message EchoResponse {
      string payload = 2;
    }

Чтобы использовать gRPC-Gateway, обновите proto‑файл, добавив HTTP‑опции:

    syntax = "proto3";
    import "google/api/annotations.proto";

    package api.echo.v1;

    option go_package = "github.com/easyp-tech/example/api/echo/v1;pb";

    service EchoAPI {
      rpc Echo(EchoRequest) returns (EchoResponse) {
        option (google.api.http) = { // [!code ++]
          post: "/api/v1/echo"       // [!code ++]
          body: "*"                  // [!code ++]
        };                           // [!code ++]
      }
    }

    message EchoRequest {
      string payload = 1;
    }

    message EchoResponse {
      string payload = 2;
    }

## Настройка конфигурации

Обновите файл `easyp.yaml`, добавив необходимые зависимости и плагины:
    deps:  # [!code ++]
      - github.com/googleapis/googleapis  # [!code ++]

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
        - name: grpc-gateway  # [!code ++]
          out: .  # [!code ++]
          opts:  # [!code ++]
            paths: source_relative  # [!code ++]
        - name: openapiv2        # [!code ++]
          out: .  # [!code ++]
          opts: # [!code ++]
            simple_operation_ids: false  # [!code ++]
            generate_unbound_methods: false  # [!code ++]

Секция `deps` перечисляет зависимости, необходимые для импортов proto‑файлов.
В данном случае добавляем `github.com/googleapis/googleapis`,
поскольку там находится файл `annotations.proto`, используемый в определении сервиса.

### Обновление зависимостей

После обновления конфигурации выполните команду для загрузки указанных зависимостей:

    easyp mod update

Больше деталей по управлению зависимостями см. в разделе [Package Manager](../../package-manager/package-manager.md).

## Генерация кода

Для генерации кода используйте:

    easyp -cfg easyp.yaml generate

Если флаг `-cfg` не указан, по умолчанию используется файл `easyp.yaml` из текущей директории:

    easyp generate

Теперь у вас сгенерированы Go‑код и код gRPC-Gateway, с которыми можно работать напрямую.