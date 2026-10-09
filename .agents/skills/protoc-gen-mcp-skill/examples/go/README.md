# Go MCP Server Example

## Proto Definition

```proto
syntax = "proto3";

package myapi.v1;

option go_package = "github.com/you/myproject/myapi/v1;myapiv1";

import "mcp/options/v1/options.proto";
import "google/protobuf/empty.proto";

service MyServiceAPI {
  option (mcp.options.v1.service) = {
    namespace: "myapi"
    description: "My tools exposed as MCP tools."
  };

  rpc CreateItem(CreateItemRequest) returns (CreateItemResponse) {
    option (mcp.options.v1.method) = {
      title: "Create item"
      description: "Create a new item with validation."
      annotations: { read_only_hint: false }
    };
  }

  rpc Health(google.protobuf.Empty) returns (HealthResponse) {
    option (mcp.options.v1.method) = {
      title: "Health check"
      description: "Verify the server is alive."
      annotations: { read_only_hint: true }
    };
  }
}

message CreateItemRequest {
  string name = 1 [(mcp.options.v1.field) = {
    description: "Item name."
    examples: [{ string_value: "Widget" }]
    min_length: 1
    max_length: 200
  }];

  int32 count = 2 [(mcp.options.v1.field) = {
    default_value: { integer_value: 1 }
    minimum: 1
    maximum: 1000
  }];

  repeated string tags = 3 [(mcp.options.v1.field) = {
    max_items: 20
    unique_items: true
  }];

  optional string note = 4;
}

message CreateItemResponse {
  string id = 1;
}

message HealthResponse {
  string status = 1;
}
```

## EasyP configuration (v1)

EasyP v1 splits the configuration into three files. Coming from a v0 `easyp.yaml` with `deps:` / `generate.inputs`? Run `easyp migrate`. See the protobuf-expert-skill migration guide.

`protobuf.mod`, with module identity and dependencies:

```text
module github.com/you/myproject

require github.com/easyp-tech/protoc-gen-mcp v0.7.1
```

`easyp.yaml`, with the lint policy:

```yaml
version: v1

linters:
  default: MINIMAL
  enable: [PACKAGE_VERSION_SUFFIX, UNARY_RPC]
  disable: [DIRECTORY_SAME_PACKAGE, PACKAGE_DIRECTORY_MATCH, PACKAGE_SAME_DIRECTORY]
```

`easyp.gen.yaml`, with code generation:

```yaml
version: v1

generate:
  paths: [proto]

plugins:
  - name: go
    out: .
    opts:
      paths: source_relative
  - command: ["go", "run", "github.com/easyp-tech/protoc-gen-mcp/cmd/protoc-gen-mcp@v0.7.1"]
    out: .
    opts:
      paths: source_relative
```

Run `easyp mod tidy` once to write `protobuf.lock`, and commit it with the other files.

## Generated Files

| File | Description |
|---|---|
| `myapi.pb.go` | Standard protobuf Go types |
| `myapi.mcp.go` | MCP handler interface + registration helper |

## Handler Implementation

```go
package main

import (
	"context"
	"log"

	myapiv1 "github.com/you/myproject/myapi/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type handler struct{}

func (handler) CreateItem(
	_ context.Context,
	req *myapiv1.CreateItemRequest,
) (*myapiv1.CreateItemResponse, error) {
	return &myapiv1.CreateItemResponse{Id: "item-1"}, nil
}

func (handler) Health(
	_ context.Context,
	_ *emptypb.Empty,
) (*myapiv1.HealthResponse, error) {
	return &myapiv1.HealthResponse{Status: "ok"}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "myapi-mcp",
		Version: "v0.1.0",
	}, nil)

	if err := myapiv1.RegisterMyServiceAPITools(server, handler{}); err != nil {
		log.Fatal(err)
	}

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
```

## Namespace Override at Registration

```go
myapiv1.RegisterMyServiceAPITools(server, handler{},
	mcpruntime.WithNamespace("custom_prefix"),
)
```

## Run

```bash
go run ./cmd/myserver
```

Generated tool names: `myapi_CreateItem`, `myapi_Health`.
