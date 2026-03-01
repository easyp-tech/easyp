# easypconfig

`mcp/easypconfig` is the source-of-truth package for `easyp.yaml` schema metadata and the MCP tool `easyp_config_describe`.

## Go usage

```go
import (
    "github.com/easyp-tech/easyp/mcp/easypconfig"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func register(server *mcp.Server) {
    easypconfig.RegisterTool(server)
}
```

Programmatic access:

- `easypconfig.Describe(...)`
- `easypconfig.SchemaByPath()`
- `easypconfig.MarshalConfigJSONSchema()`

## Non-Go usage

Versioned and latest JSON Schema artifacts are generated into:

- `schemas/easyp-config-v1.schema.json`
- `schemas/easyp-config.schema.json`

This allows external tooling (for example Kotlin/JetBrains plugins) to consume the same schema without linking Go code.

Generate/update artifacts:

```sh
go run ./cmd/easyp-schema-gen
# or
task schema:generate
```
