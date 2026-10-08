# Python MCP Server — Protobuf Handler Mode

Set `python_handler: protobuf`. Implement against raw `*_pb2` message classes instead of generated dataclasses.

Use this when you already have server logic written against standard protobuf classes and don't want to introduce a dataclass layer.

## Proto Definition

```proto
syntax = "proto3";

package myapi.v1;

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
  enable: [PACKAGE_LOWER_SNAKE_CASE, PACKAGE_VERSION_SUFFIX, UNARY_RPC]
  disable: [DIRECTORY_SAME_PACKAGE, PACKAGE_DIRECTORY_MATCH, PACKAGE_SAME_DIRECTORY]
```

`easyp.gen.yaml`, with code generation:

```yaml
version: v1

generate:
  paths: [proto]

plugins:
  - name: python
    with_imports: true
    out: .
  - command: ["go", "run", "github.com/easyp-tech/protoc-gen-mcp/cmd/protoc-gen-mcp@v0.7.1"]
    out: .
    opts:
      paths: source_relative
      lang: python
      python_runtime: google.protobuf
      python_handler: protobuf
```

Run `easyp mod tidy` once to write `protobuf.lock`, and commit it with the other files.

## Generated Files

| File | Description |
|---|---|
| `myapi_pb2.py` | Standard protobuf Python types |
| `myapi_mcp.py` | Raw protobuf handler sidecar — `MyServiceAPIToolHandler` typed with `*_pb2` messages |
| `mcp/__init__.py` | Bridge so `mcp.options.*` coexists with the official `mcp` SDK package |

## Handler Implementation

```python
from __future__ import annotations

import anyio
import mcp.server.lowlevel
import mcp.server.stdio

from proto import myapi_mcp, myapi_pb2


class MyServiceAPI(myapi_mcp.MyServiceAPIToolHandler):
    def create_item(
        self,
        _ctx: myapi_mcp.ToolRequestContext,
        req: myapi_pb2.CreateItemRequest,
    ) -> myapi_pb2.CreateItemResponse:
        return myapi_pb2.CreateItemResponse(id="item-1")

    def health(
        self,
        _ctx: myapi_mcp.ToolRequestContext,
        _req: myapi_pb2.HealthRequest,
    ) -> myapi_pb2.HealthResponse:
        return myapi_pb2.HealthResponse(status="ok")


def new_server() -> mcp.server.lowlevel.Server:
    server = mcp.server.lowlevel.Server("myapi-mcp", version="0.1.0")
    myapi_mcp.register_my_service_api_tools(server, MyServiceAPI())
    return server


async def run_stdio_server() -> None:
    server = new_server()
    async with mcp.server.stdio.stdio_server() as (read_stream, write_stream):
        await server.run(
            read_stream,
            write_stream,
            server.create_initialization_options(),
        )


def main() -> None:
    anyio.run(run_stdio_server)


if __name__ == "__main__":
    main()
```

## Key Notes

- In protobuf mode, generated dataclasses, `UNSET`, and dataclass mapper helpers are omitted
- Runtime validation, ProtoJSON parsing, output validation, structured content, tool names, annotations, icons, and execution metadata are still generator-owned
- Handler types are raw `*_pb2` message classes from the standard protobuf Python generator
- For protobuf-only projects, `python_handler: protobuf` writes the raw handler sidecar to the normal `*_mcp.py` path

## Run

```bash
python server.py
```

Generated tool names: `myapi_CreateItem`, `myapi_Health`.
