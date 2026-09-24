# EasyP configuration MCP

`easyp_config_describe` describes the v1 files developers edit: `easyp.yaml` and `easyp.gen.yaml`. It returns field descriptions, valid examples, and JSON Schema fragments. The schema comes from the same `internal/config/v1.SchemaJSON` function used by `schema-gen` and CLI validation.

Run the stdio server:

```sh
go run ./cmd/easyp-mcp
```

The MCP client must start this command from the EasyP repository, or run a built `easyp-mcp` binary. The server writes MCP messages to stdout and diagnostics to stderr.

The tool accepts `file` (`easyp.yaml` by default) and `path` (`$` by default). For example, `{"file":"easyp.gen.yaml","path":"plugins[0].out"}` describes `plugins[].out`. Set `include_schema`, `include_fields`, `include_examples`, or `include_children` to `false` to omit those sections; `examples_limit` limits examples to 1–50.

Go programs embedding an MCP server can register the same tool:

```go
server := mcp.NewServer(&mcp.Implementation{Name: "my-server", Version: "v1"}, nil)
easypconfig.RegisterTool(server)
```

`protobuf.mod` and `protobuf.lock` are outside this MCP tool. The CLI checks them with `easyp validate-config`; `schema-gen` still writes a JSON Schema for `protobuf.lock`.
