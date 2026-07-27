---
name: protoc-gen-mcp-skill
description: "Build MCP servers from protobuf definitions using protoc-gen-mcp and easyp. Use when: creating an MCP server, generating MCP tools from proto files, building a proto-first MCP server in Go, configuring easyp for MCP generation, adding MCP tool annotations to protobuf services, implementing MCP tool handlers, setting up ProtoJSON-based MCP tools, or any task involving protobuf-to-MCP code generation. Also use when the user mentions protoc-gen-mcp, mcp proto, proto mcp server, easyp mcp, or wants type-safe MCP bindings from .proto files."
---

# protoc-gen-mcp — Proto-First MCP Server Generator

Generate type-safe Go MCP tool bindings from annotated protobuf services.
Protobuf is the source of truth: define your service once in `.proto`, generate
both `*.pb.go` and `*.mcp.go`, implement the handler interface, and serve.

## When to Use

- Building a new MCP server and want type-safe, schema-validated tools
- Already have protobuf services and want to expose them as MCP tools
- Need JSON Schema validation on MCP tool inputs derived from proto definitions
- Want ProtoJSON as the wire format for MCP tool requests and responses

## Prerequisites

Install [easyp](https://easyp.tech) — the recommended way to lint and generate:

```bash
brew install easyp-tech/tap/easyp
```

Or install from source:

```bash
go install github.com/easyp-tech/easyp/cmd/easyp@latest
```

See https://easyp.tech/docs for full documentation.

## Step-by-Step Workflow

### Step 1: Define Your Proto Service

Create a `.proto` file with service, methods, and MCP annotations:

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

  // CreateItem creates a new item.
  rpc CreateItem(CreateItemRequest) returns (CreateItemResponse) {
    option (mcp.options.v1.method) = {
      title: "Create item"
      description: "Create a new item with validation."
      annotations: { read_only_hint: false }
    };
  }

  // Health returns server status.
  rpc Health(google.protobuf.Empty) returns (HealthResponse) {
    option (mcp.options.v1.method) = {
      title: "Health check"
      description: "Verify the server is alive."
      annotations: { read_only_hint: true }
    };
  }
}

message CreateItemRequest {
  // name is required (singular, non-optional in proto3).
  string name = 1 [(mcp.options.v1.field) = {
    description: "Item name."
    examples: [{ string_value: "Widget" }]
    min_length: 1
    max_length: 200
  }];

  // count has a default and numeric bounds.
  int32 count = 2 [(mcp.options.v1.field) = {
    default_value: { number_value: 1 }
    minimum: 1
    maximum: 1000
  }];

  // tags is optional because it is repeated.
  repeated string tags = 3 [(mcp.options.v1.field) = {
    max_items: 20
    unique_items: true
  }];

  // note is optional because of the `optional` keyword.
  optional string note = 4;
}

message CreateItemResponse {
  string id = 1;
}

message HealthResponse {
  string status = 1;
}
```

### Step 2: Configure easyp

Create `easyp.yaml` in your project root. This single file drives both
`protoc-gen-go` (standard Go protobuf) and `protoc-gen-mcp` (MCP bindings):

```yaml
deps:
  - github.com/easyp-tech/protoc-gen-mcp@v0.3.1

lint:
  use:
    - PACKAGE_DEFINED
    - PACKAGE_VERSION_SUFFIX
    - RPC_NO_CLIENT_STREAMING
    - RPC_NO_SERVER_STREAMING

generate:
  inputs:
    - directory:
        path: proto          # directory containing your .proto files
        root: "."
  plugins:
    - name: go
      out: .
      opts:
        paths: source_relative
    - command: ["go", "run", "github.com/easyp-tech/protoc-gen-mcp/cmd/protoc-gen-mcp@latest"]
      out: .
      opts:
        paths: source_relative
```

For reproducible builds, pin a specific version tag instead of `@latest`:

```yaml
    - command: ["go", "run", "github.com/easyp-tech/protoc-gen-mcp/cmd/protoc-gen-mcp@v0.3.1"]
```

Why easyp over raw protoc:
- Single `easyp.yaml` config manages all plugins, lint rules, and dependencies
- Both `*.pb.go` and `*.mcp.go` are generated in one command
- Built-in linting catches streaming RPCs and other unsupported patterns early
- Git-native dependency management with lock files for reproducibility
- No need to install `protoc` or manage plugin binaries manually

### Step 3: Generate Code

```bash
# Validate config
easyp validate-config

# Download dependencies
easyp mod download

# Lint proto files
easyp lint -p proto -r .

# Generate *.pb.go and *.mcp.go
easyp generate -p proto -r .
```

This produces two files next to your `.proto`:
- `myapi.pb.go` — standard protobuf Go types
- `myapi.mcp.go` — MCP tool handler interface + registration

### Step 4: Implement the Handler

The generated code exposes a `<Service>ToolHandler` interface. Implement it:

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

### Step 5: Run

```bash
go run ./cmd/myserver
```

The server communicates over stdio. Connect any MCP client to it. The generated
tools are `myapi_CreateItem` and `myapi_Health`.

## Key Concepts

### Tool Naming

Generated tool names follow the pattern `{namespace}_{MethodName}`. Dots in
the namespace are normalized to underscores. Override the method segment with
`mcp.options.v1.method.name`.

### Requiredness Policy

Requiredness in generated MCP JSON Schema is determined by proto3 syntax:

| Proto Pattern | Required? |
|---|---|
| `string name = 1` (singular, no `optional`) | YES |
| `optional string name = 1` | NO |
| `repeated string names = 1` | NO |
| `map<string, string> m = 1` | NO |
| `oneof choice { ... }` | NO (unless `mcp.options.v1.oneof.required = true`) |

Fields that are not required accept explicit JSON `null`.

### ProtoJSON Contract

MCP tool I/O uses ProtoJSON encoding. Key differences from plain JSON:

- `int64`/`uint64` are JSON **strings**, not numbers
- `float`/`double` accept `"NaN"`, `"Infinity"`, `"-Infinity"` as strings
- `bytes` use base64 encoding
- Enums use string names (e.g., `"FORECAST_MODE_DAILY"`)
- `Timestamp` → RFC 3339 string, `Duration` → `"3.5s"`, `FieldMask` → `"field1,field2"`

### Supported Protobuf Features

- Scalars, enums, nested messages, repeated, maps, `oneof`, `optional`
- Recursive messages via `$defs`/`$ref`
- Well-known types: `Any`, `Empty`, `Timestamp`, `Duration`, `FieldMask`,
  `Struct`, `Value`, `ListValue`, and all scalar wrapper types

### Fail-Fast Rules

The generator rejects at generation time (not runtime):
- Proto2 syntax
- Streaming RPCs (client, server, or bidirectional)
- Unsupported `google.protobuf.*` types

## Quick Proto Options Reference

```proto
import "mcp/options/v1/options.proto";

// Service: namespace prefix, description, icons
option (mcp.options.v1.service) = {
  namespace: "myapi"
  description: "My API tools."
};

// Method: tool name, title, description, visibility, agent hints
option (mcp.options.v1.method) = {
  name: "CustomName"
  title: "Human Title"
  description: "What this tool does."
  hidden: true
  annotations: {
    read_only_hint: true
    destructive_hint: false
    idempotent_hint: true
  }
};

// Field: description, examples, defaults, validation constraints
[(mcp.options.v1.field) = {
  description: "Field purpose."
  examples: [{ string_value: "example" }]
  default_value: { number_value: 42 }
  pattern: "^[A-Z]"
  min_length: 1
  max_length: 255
  minimum: 0
  maximum: 100
  min_items: 1
  max_items: 50
  unique_items: true
}];

// Oneof: make a oneof group required in the MCP schema
option (mcp.options.v1.oneof) = { required: true };

// Enum: title and description for the enum type
option (mcp.options.v1.enum) = { title: "Status" };

// Enum value: hide sentinel zero-value from the schema
UNSPECIFIED = 0 [(mcp.options.v1.enum_value) = { hidden: true }];
```

For full options details, read `references/options-reference.md` in this skill.

## Common Patterns

### Hide Internal RPCs

```proto
rpc InternalDebug(DebugRequest) returns (DebugResponse) {
  option (mcp.options.v1.method) = { hidden: true };
};
```

### Read-Only vs Destructive Tools

```proto
// Read-only query
option (mcp.options.v1.method) = {
  annotations: { read_only_hint: true }
};

// Destructive mutation
option (mcp.options.v1.method) = {
  annotations: { destructive_hint: true }
};
```

### Namespace Override at Registration

```go
myapiv1.RegisterMyServiceAPITools(server, handler{},
    mcpruntime.WithNamespace("custom_prefix"),
)
```

## Reference Files

For detailed lookup tables, read these files from this skill directory:

- `references/options-reference.md` — full MCP proto options with all fields and examples
- `references/schema-mapping.md` — proto type → JSON Schema mapping, well-known types, nullability rules
