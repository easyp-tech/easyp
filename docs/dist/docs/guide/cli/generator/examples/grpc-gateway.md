# gRPC-Gateway

## Installing Plugins

In addition to the plugins for working with gRPC, you also need to install the following plugins for gRPC-Gateway:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

These commands will install the `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-openapiv2`, and
`protoc-gen-grpc-gateway` plugins for use with EasyP`.

## Example Proto Service

Here is the initial proto file for an Echo service:

```proto
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
```

To use gRPC-Gateway, update the proto file to include HTTP options:

```proto
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
```

## Configuration Setup

Update your `easyp.yaml` configuration file to include the necessary dependencies and plugins:

```yaml
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
```

The `deps` section lists dependencies required for proto file imports.
In this case, we add `github.com/googleapis/googleapis`
because it contains the `annotations.proto` file used in the proto service definition.

### Updating Dependencies

After updating your configuration file, run the following command to download the specified dependencies:

```bash
easyp mod update
```

For more details on managing dependencies, refer to the [Package Manager](../../package-manager/package-manager.md) section.

## Generating Code

To generate the code, use the following command:

```bash
easyp -cfg easyp.yaml generate
```

If the `-cfg` flag is not specified, the `easyp.yaml` file in the current directory will be used by default:

```bash
easyp generate
```

Now you have the generated Go and gRPC-Gateway code, which you can interact with directly. 
