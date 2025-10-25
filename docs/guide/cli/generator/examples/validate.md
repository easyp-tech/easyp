# Validate

## Installing Plugins

First, install the necessary plugins for working with gRPC:

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/envoyproxy/protoc-gen-validate@latest
```

These commands will install the `protoc-gen-go` and `protoc-gen-go-grpc` and `protoc-gen-validate` plugins for use with EasyP.

## Example Proto Service

Here is an example proto file for an Echo service:

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

## Configuration Setup

Create and configure the easyp.yaml configuration file:

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

## Generating Code

To generate code, use the following command:

```shell
easyp -cfg easyp.yaml generate
```

If the -cfg flag is not specified, the easyp.yaml file in the current directory will be used by default:

```shell
easyp generate
```

Now you have the generated Go code, which you can interact with directly.
