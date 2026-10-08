# Kotlin MCP Server Example

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

Kotlin requires a dual-generation flow: Java protobuf output + Kotlin protobuf output + `lang=kotlin` MCP sidecar.

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
  - name: java
    with_imports: true
    out: src/generated/main/java
  - name: kotlin
    out: src/generated/main/kotlin
  - command: ["go", "run", "github.com/easyp-tech/protoc-gen-mcp/cmd/protoc-gen-mcp@v0.7.1"]
    out: src/generated/main/kotlin
    opts:
      paths: source_relative
      lang: kotlin
```

Run `easyp mod tidy` once to write `protobuf.lock`, and commit it with the other files.

## Generated Files

| File | Description |
|---|---|
| `src/generated/main/java/.../*.java` | Protobuf Java types (required dependency for Kotlin) |
| `src/generated/main/kotlin/.../*.kt` | Protobuf Kotlin types |
| `src/generated/main/kotlin/.../*_mcp.kt` | MCP sidecar with `<Service>ToolHandler` interface and `register<Service>Tools()` |

## build.gradle.kts

```kotlin
plugins {
    kotlin("jvm") version "2.1.0"
    application
}

repositories {
    mavenCentral()
}

dependencies {
    implementation("io.modelcontextprotocol.sdk:kotlin-sdk-server:0.4.0")
    implementation("com.google.protobuf:protobuf-java:4.31.1")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.9.0")
}

application {
    mainClass = "com.example.MyServiceServerKt"
}
```

## Handler Implementation

```kotlin
package com.example

import io.modelcontextprotocol.sdk.Server
import io.modelcontextprotocol.sdk.StdioServerTransport
import myapi.v1.registerMyServiceAPITools
import myapi.v1.MyServiceAPIToolHandler
import myapi.v1.ToolRequestContext
import myapi.v1.CreateItemRequest
import myapi.v1.CreateItemResponse
import myapi.v1.HealthRequest
import myapi.v1.HealthResponse

class MyServiceAPI : MyServiceAPIToolHandler {
  override fun createItem(ctx: ToolRequestContext, req: CreateItemRequest): CreateItemResponse {
    return CreateItemResponse.newBuilder().setId("item-1").build()
  }

  override fun health(ctx: ToolRequestContext, req: HealthRequest): HealthResponse {
    return HealthResponse.newBuilder().setStatus("ok").build()
  }
}

suspend fun main() {
  val server = Server(
    name = "myapi-mcp",
    version = "0.1.0",
    capabilities = Server.Capabilities(tools = Server.ToolsCapability())
  )
  registerMyServiceAPITools(server, MyServiceAPI(), "myapi")
  server.connect(StdioServerTransport())
}
```

## Key Notes

- Requires Java protobuf output + Kotlin protobuf output + `lang=kotlin` MCP sidecar — all three are part of the working build graph
- Generated handlers implement `<Service>ToolHandler`, registered through `register<Service>Tools(server, impl, namespace?)`
- User-authored protos do not need a Go `go_package` option

## Build and Run

```bash
gradle installDist
./build/install/myserver/bin/myserver
```

Generated tool names: `myapi_CreateItem`, `myapi_Health`.
