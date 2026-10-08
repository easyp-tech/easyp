---
name: protoc-gen-mcp-skill
description: "Build MCP servers from protobuf definitions using protoc-gen-mcp and easyp. Use when: creating an MCP server, generating MCP tools from proto files, building a proto-first MCP server in Go, Python, Java, Kotlin, TypeScript, or JavaScript, configuring easyp (v1: protobuf.mod, easyp.yaml, easyp.gen.yaml) for MCP generation, migrating an existing easyp v0 setup, adding MCP tool annotations to protobuf services, implementing MCP tool handlers, setting up ProtoJSON-based MCP tools, or any task involving protobuf-to-MCP code generation. Also use when the user mentions protoc-gen-mcp, mcp proto, proto mcp server, easyp mcp, or wants type-safe MCP bindings from .proto files."
---

# protoc-gen-mcp — Proto-First MCP Server Generator

Generate type-safe MCP tool bindings from annotated protobuf services for **Go, Python, Java, Kotlin, and TypeScript**. JavaScript projects consume compiled TypeScript output. Protobuf is the source of truth: define your service once in `.proto`, generate language-specific protobuf types and MCP sidecars, implement the handler interface, and serve.

## When to Use

- Building a new MCP server and want type-safe, schema-validated tools
- Already have protobuf services and want to expose them as MCP tools
- Need JSON Schema validation on MCP tool inputs derived from proto definitions
- Want ProtoJSON as the wire format for MCP tool requests and responses
- Targeting Go, Python, Java, Kotlin, TypeScript, or JavaScript runtimes

## Prerequisites

Install [easyp](https://easyp.tech) **v1**, the recommended way to lint and generate. v1 is published as nightly tags (`v1.0.0-nightly.YYYYMMDD.N`); `brew` and `go install …@latest` still install v0.17.0:

```bash
EASYP_VERSION=v1.0.0-nightly.20261008.1   # pick the newest from github.com/easyp-tech/easyp/releases
go install "github.com/easyp-tech/easyp/cmd/easyp@$EASYP_VERSION"
easyp --version
```

EasyP v1 uses three config files plus a lock: `protobuf.mod` (module and dependencies), `easyp.yaml` (lint policy), `easyp.gen.yaml` (plugins) and `protobuf.lock`.

**Existing project with a v0 `easyp.yaml`** (top-level `deps:`, `lint.use`, `generate.inputs`), or easyp failing with `legacy EasyP configuration detected`? Migrate first with `easyp migrate`. Follow the step-by-step procedure in the [protobuf-expert-skill](https://github.com/easyp-tech/protobuf-expert-skill) (`references/migration-v0-to-v1.md`).

See https://easyp.tech/docs for full documentation.

## Step-by-Step Workflow

### Step 1: Define Your Proto Service

Create a `.proto` file with service, methods, and MCP annotations using `mcp/options/v1/options.proto`:

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

### Step 2: Choose Your Language

Each language has its own EasyP configuration (`protobuf.mod`, `easyp.yaml`, `easyp.gen.yaml`) and handler pattern. See the examples:

| Language | Example | Handler Style |
|---|---|---|
| Go | [examples/go](examples/go/README.md) | `<Service>ToolHandler` interface |
| Python (dataclass) | [examples/python/dataclass](examples/python/dataclass/README.md) | Dataclasses from `*_mcp.py` |
| Python (protobuf) | [examples/python/protobuf](examples/python/protobuf/README.md) | Raw `*_pb2` classes |
| Python (dataclass+protobuf) | [examples/python/dataclass-protobuf](examples/python/dataclass-protobuf/README.md) | Both surfaces side by side |
| Java | [examples/java](examples/java/README.md) | Nested `<Service>ToolHandler` in `*Mcp.java` |
| Kotlin | [examples/kotlin](examples/kotlin/README.md) | `<Service>ToolHandler` interface |
| TypeScript | [examples/typescript](examples/typescript/README.md) | `<Service>ToolHandler` interface |
| JavaScript | [examples/javascript](examples/javascript/README.md) | Compiled TS output + JSDoc types |

### Step 3: Generate Code

```bash
easyp mod tidy           # first run / after editing protobuf.mod: resolve protoc-gen-mcp and write protobuf.lock
easyp validate-config
easyp lint
easyp generate           # targets come from easyp.gen.yaml (generate.paths); v1 has no -p/-r flags
```

In CI use the committed lock exactly: `easyp --frozen mod download && easyp --frozen generate && git diff --exit-code`.

The `mcp/options/v1/options.proto` import comes from the `require github.com/easyp-tech/protoc-gen-mcp vX.Y.Z` line in `protobuf.mod`. Keep that version equal to the `protoc-gen-mcp@vX.Y.Z` used in the plugin `command`.

### Step 4: Implement the Handler and Serve

Follow the language-specific example linked above.

## Key Concepts

### Tool Naming

Generated tool names follow the pattern `{namespace}_{MethodName}`. Dots in the namespace are normalized to underscores. Override the method segment with `mcp.options.v1.method.name`.

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
  default_value: { integer_value: 42 }
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

// Enum: title and description
option (mcp.options.v1.enum) = { title: "Status" };

// Enum value: hide sentinel zero-value
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

## Testing With MCP Inspector

```bash
# Go
npx -y @modelcontextprotocol/inspector go run ./cmd/myserver

# Python
npx -y @modelcontextprotocol/inspector python server.py

# TypeScript (build first)
npx -y @modelcontextprotocol/inspector node dist/server.js

# Java/Kotlin (after installDist)
npx -y @modelcontextprotocol/inspector ./build/install/myserver/bin/myserver
```

## Reference Files

For detailed lookup tables, read these files from this skill directory:

- `references/options-reference.md` — full MCP proto options with all fields and examples
- `references/schema-mapping.md` — proto type → JSON Schema mapping, well-known types, nullability rules
