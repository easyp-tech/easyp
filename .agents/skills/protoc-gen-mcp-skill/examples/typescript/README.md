# TypeScript MCP Server Example

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

TypeScript requires Protobuf-ES `_pb.ts` files first, then `lang=typescript` MCP sidecar.

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
  - command: ["./node_modules/.bin/protoc-gen-es"]
    with_imports: true
    out: src/generated
    opts:
      target: ts
      import_extension: js
  - command: ["go", "run", "github.com/easyp-tech/protoc-gen-mcp/cmd/protoc-gen-mcp@v0.7.1"]
    out: src/generated
    opts:
      paths: source_relative
      lang: typescript
```

Run `easyp mod tidy` once to write `protobuf.lock`, and commit it with the other files.

## Generated Files

| File | Description |
|---|---|
| `src/generated/proto/myapi_pb.ts` | Protobuf-ES message types and schemas |
| `src/generated/proto/myapi_mcp.ts` | MCP sidecar with `<Service>ToolHandler` interface and `register<Service>Tools()` |

## package.json

```json
{
  "name": "myapi-mcp-server",
  "type": "module",
  "scripts": {
    "build": "tsc",
    "start": "node dist/server.js"
  },
  "dependencies": {
    "@modelcontextprotocol/sdk": "1.29.0",
    "@bufbuild/protobuf": "2.12.0",
    "ajv": "8.20.0"
  },
  "devDependencies": {
    "typescript": "5.7.0",
    "@bufbuild/protoc-gen-es": "2.12.0"
  }
}
```

## tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "outDir": "dist",
    "rootDir": "src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  },
  "include": ["src"]
}
```

## Handler Implementation

```typescript
import { create } from "@bufbuild/protobuf";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";

import {
  CreateItemResponseSchema,
  HealthResponseSchema,
} from "./generated/proto/myapi_pb.js";
import {
  type MyServiceAPIToolHandler,
  type ToolRequestContext,
  registerMyServiceAPITools,
} from "./generated/proto/myapi_mcp.js";

class MyServiceAPI implements MyServiceAPIToolHandler {
  createItem(_ctx: ToolRequestContext, req: CreateItemRequest): CreateItemResponse {
    return create(CreateItemResponseSchema, { id: "item-1" });
  }

  health(_ctx: ToolRequestContext, _req: HealthRequest): HealthResponse {
    return create(HealthResponseSchema, { status: "ok" });
  }
}

const server = new Server(
  { name: "myapi-mcp", version: "0.1.0" },
  { capabilities: { tools: {} } },
);
registerMyServiceAPITools(server, new MyServiceAPI(), "myapi");
await server.connect(new StdioServerTransport());
```

## Key Notes

- Uses Protobuf-ES (`_pb.ts` files) with NodeNext-compatible `.js` import specifiers
- Generated MCP sidecar imports the official `@modelcontextprotocol/sdk` low-level `Server`
- Preserves raw JSON Schemas, validates through Ajv, maps payloads through Protobuf-ES ProtoJSON
- Use `import_extension: js` so emitted imports work under Node ESM and `moduleResolution: NodeNext`

## Build and Run

```bash
npm install
npm run build
npm start
```

Generated tool names: `myapi_CreateItem`, `myapi_Health`.
