# Python MCP Server — Dataclass + Protobuf Handler Mode

Set `python_handler: dataclass+protobuf`. Generates both surfaces in one pass: dataclass API in `*_mcp.py` and raw protobuf API in `*_mcp_pb.py`.

Use this when you want both handler surfaces available — dataclasses for new code and raw `*_pb2` classes for existing logic.

## Proto Definition

```proto
syntax = "proto3";

package tasks.v1;

import "mcp/options/v1/options.proto";
import "google/protobuf/empty.proto";
import "google/protobuf/timestamp.proto";

service TaskAPI {
  option (mcp.options.v1.service) = {
    namespace: "tasks"
    description: "Task management tools."
  };

  rpc CreateTask(CreateTaskRequest) returns (CreateTaskResponse) {
    option (mcp.options.v1.method) = {
      title: "Create task"
      description: "Create a new task."
      annotations: { read_only_hint: false }
    };
  }

  rpc ListTasks(ListTasksRequest) returns (ListTasksResponse) {
    option (mcp.options.v1.method) = {
      title: "List tasks"
      description: "List tasks with optional filters."
      annotations: { read_only_hint: true }
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

message CreateTaskRequest {
  string title = 1 [(mcp.options.v1.field) = {
    description: "Task title."
    examples: [{ string_value: "Buy groceries" }]
    min_length: 1
    max_length: 200
  }];

  repeated string tags = 2 [(mcp.options.v1.field) = {
    max_items: 10
    unique_items: true
  }];
}

message CreateTaskResponse {
  Task task = 1;
}

message ListTasksRequest {
  optional bool include_completed = 1 [(mcp.options.v1.field) = {
    default_value: { bool_value: true }
  }];

  optional int32 limit = 2 [(mcp.options.v1.field) = {
    default_value: { integer_value: 10 }
    minimum: 1
    maximum: 100
  }];

  repeated string tags = 3;
}

message ListTasksResponse {
  repeated Task tasks = 1;
}

message HealthResponse {
  bool ok = 1;
  int32 task_count = 2;
}

message Task {
  string id = 1 [(mcp.options.v1.field) = { read_only: true }];
  string title = 2;
  bool completed = 3;
  repeated string tags = 4;
  google.protobuf.Timestamp created_at = 5;
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
      python_handler: dataclass+protobuf
```

Run `easyp mod tidy` once to write `protobuf.lock`, and commit it with the other files.

## Generated Files

| File | Description |
|---|---|
| `tasks_pb2.py` | Standard protobuf Python types |
| `tasks_mcp.py` | Dataclass sidecar — `CreateTaskRequest`, `Task`, `register_task_api_tools()` |
| `tasks_mcp_pb.py` | Raw protobuf sidecar — `TaskAPIToolHandler` typed with `*_pb2` messages |
| `mcp/__init__.py` | Bridge so `mcp.options.*` coexists with the official `mcp` SDK package |

## Handler Implementation — Using Raw Protobuf Sidecar

```python
from __future__ import annotations

from datetime import datetime, timezone

import anyio
import mcp.server.lowlevel
import mcp.server.stdio
from google.protobuf import timestamp_pb2

from proto import tasks_mcp_pb, tasks_pb2


def _now_timestamp() -> timestamp_pb2.Timestamp:
    timestamp = timestamp_pb2.Timestamp()
    timestamp.FromDatetime(datetime.now(timezone.utc))
    return timestamp


class TaskStore(tasks_mcp_pb.TaskAPIToolHandler):
    def __init__(self) -> None:
        self._next_id = 1
        self._tasks: dict[str, tasks_pb2.Task] = {}

    def create_task(
        self,
        _ctx: tasks_mcp_pb.ToolRequestContext,
        req: tasks_pb2.CreateTaskRequest,
    ) -> tasks_pb2.CreateTaskResponse:
        task_id = f"task-{self._next_id}"
        self._next_id += 1

        task = tasks_pb2.Task(
            id=task_id,
            title=req.title,
            completed=False,
            tags=list(req.tags),
        )
        task.created_at.CopyFrom(_now_timestamp())
        self._tasks[task.id] = task
        return tasks_pb2.CreateTaskResponse(task=task)

    def list_tasks(
        self,
        _ctx: tasks_mcp_pb.ToolRequestContext,
        req: tasks_pb2.ListTasksRequest,
    ) -> tasks_pb2.ListTasksResponse:
        include_completed = req.include_completed if req.HasField("include_completed") else True
        limit = req.limit if req.HasField("limit") else 10
        required_tags = set(req.tags)

        matches: list[tasks_pb2.Task] = []
        for task in self._tasks.values():
            if not include_completed and task.completed:
                continue
            if required_tags and not required_tags.issubset(set(task.tags)):
                continue
            matches.append(task)
            if len(matches) >= limit:
                break

        return tasks_pb2.ListTasksResponse(tasks=matches)

    def health(
        self,
        _ctx: tasks_mcp_pb.ToolRequestContext,
        _req: tasks_pb2.HealthRequest,
    ) -> tasks_pb2.HealthResponse:
        return tasks_pb2.HealthResponse(ok=True, task_count=len(self._tasks))


def new_server() -> mcp.server.lowlevel.Server:
    server = mcp.server.lowlevel.Server("tasks-mcp", version="0.1.0")
    tasks_mcp_pb.register_task_api_tools(server, TaskStore())
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

## Handler Implementation — Using Dataclass Sidecar

The dataclass sidecar in `tasks_mcp.py` is also available for new code:

```python
from proto import tasks_mcp

class DataclassTaskStore:
    def create_task(
        self,
        _ctx: tasks_mcp.ToolRequestContext,
        req: tasks_mcp.CreateTaskRequest,
    ) -> tasks_mcp.CreateTaskResponse:
        return tasks_mcp.CreateTaskResponse(
            task=tasks_mcp.Task(id="task-1", title=req.title)
        )

tasks_mcp.register_task_api_tools(server, DataclassTaskStore())
```

## Key Notes

- `*_mcp.py` remains the dataclass sidecar; `*_mcp_pb.py` exposes the raw protobuf handler sidecar
- Both modules share the same protocol and registration names
- In the protobuf sidecar, generated dataclasses and `UNSET` are omitted
- Use `tasks_mcp_pb` for raw `*_pb2` handlers, `tasks_mcp` for dataclass handlers

## Run

```bash
python server.py
```

Generated tool names: `tasks_CreateTask`, `tasks_ListTasks`, `tasks_Health`.
