# JavaScript MCP Server Example

There is no `lang=javascript`. JavaScript projects consume compiled TypeScript output — the `.js` files and `.d.ts` declarations produced by building the TypeScript target.

This example shows how to import compiled generated output from a TypeScript project using `// @ts-check` and JSDoc type annotations.

## Prerequisites

A TypeScript project must be built first to produce the `.js` and `.d.ts` files. See [examples/typescript](../typescript/README.md) for the TypeScript setup.

## Proto Definition

Same proto as the TypeScript example:

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
    };
  }

  rpc Health(google.protobuf.Empty) returns (HealthResponse) {
    option (mcp.options.v1.method) = {
      title: "Health check"
      description: "Verify the server is alive."
    };
  }
}

message CreateItemRequest {
  string name = 1 [(mcp.options.v1.field) = {
    description: "Item name."
    min_length: 1
    max_length: 200
  }];
  int32 count = 2;
  repeated string tags = 3;
  optional string note = 4;
}

message CreateItemResponse {
  string id = 1;
}

message HealthResponse {
  string status = 1;
}
```

## TypeScript Build (prerequisite)

```bash
# Build the TypeScript project first
cd ../typescript
npm install
npm run build
```

This produces:
- `dist/generated/proto/myapi_pb.js` + `.d.ts`
- `dist/generated/proto/myapi_mcp.js` + `.d.ts`

## JavaScript Server Implementation

```js
// @ts-check

import { create } from "@bufbuild/protobuf";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";

// Import compiled TypeScript output
import {
  CreateItemResponseSchema,
  HealthResponseSchema,
} from "../typescript/dist/generated/proto/myapi_pb.js";
import { registerMyServiceAPITools } from "../typescript/dist/generated/proto/myapi_mcp.js";

/**
 * @typedef {import("../typescript/dist/generated/proto/notebook_mcp.js").MyServiceAPIToolHandler} MyServiceAPIToolHandler
 * @typedef {import("../typescript/dist/generated/proto/notebook_mcp.js").ToolRequestContext} ToolRequestContext
 * @typedef {import("../typescript/dist/generated/proto/notebook_pb.js").CreateItemRequest} CreateItemRequest
 * @typedef {import("../typescript/dist/generated/proto/notebook_pb.js").CreateItemResponse} CreateItemResponse
 * @typedef {import("../typescript/dist/generated/proto/notebook_pb.js").HealthRequest} HealthRequest
 * @typedef {import("../typescript/dist/generated/proto/notebook_pb.js").HealthResponse} HealthResponse
 */

/** @implements {MyServiceAPIToolHandler} */
class MyServiceAPI {
  /**
   * @param {ToolRequestContext} _ctx
   * @param {CreateItemRequest} req
   * @returns {CreateItemResponse}
   */
  createItem(_ctx, req) {
    return create(CreateItemResponseSchema, { id: "item-1" });
  }

  /**
   * @param {ToolRequestContext} _ctx
   * @param {HealthRequest} _req
   * @returns {HealthResponse}
   */
  health(_ctx, _req) {
    return create(HealthResponseSchema, { status: "ok" });
  }
}

const server = new Server(
  { name: "myapi-mcp-js", version: "0.1.0" },
  { capabilities: { tools: {} } },
);
const handler = new MyServiceAPI();
registerMyServiceAPITools(server, handler, "myapi");
await server.connect(new StdioServerTransport());
```

## package.json

```json
{
  "name": "myapi-mcp-js",
  "type": "module",
  "scripts": {
    "start": "node src/server.js"
  },
  "dependencies": {
    "@modelcontextprotocol/sdk": "1.29.0",
    "@bufbuild/protobuf": "2.12.0"
  }
}
```

## Key Notes

- No separate `lang=javascript` generation exists
- JavaScript imports the compiled `.js` output and `.d.ts` metadata from the TypeScript build
- Use `// @ts-check` and JSDoc `@typedef` / `@implements` for editor type verification
- The generated `.d.ts` files provide full editor autocomplete and type checking for plain JavaScript

## Run

```bash
npm install
npm start
```

Generated tool names: `myapi_CreateItem`, `myapi_Health`.
