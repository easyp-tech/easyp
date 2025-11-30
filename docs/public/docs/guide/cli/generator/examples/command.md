# Executing Plugin via Command

EasyP supports executing plugins via custom commands. This is especially useful for running plugins through `go run` without prior installation.

## Benefits

- **No installation required**: Plugins are executed directly via command
- **Versioning**: You can specify a specific plugin version via `@version`
- **Flexibility**: Support for any commands, not just `go run`

## Example: gRPC Gateway via go run

Below is an example of using `protoc-gen-grpc-gateway` via `go run`:

```yaml
version: v1alpha

deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.25.1

generate:
  inputs:
    - directory: "proto"
  plugins:
    # Go plugin (local)
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/mycompany/api
    
    # gRPC plugin (local)
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: source_relative
    
    # gRPC Gateway via go run
    - command: ["go", "run", "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.25.1"]
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true
```

## Example: Validate via go run

Using `protoc-gen-validate` via command:

```yaml
version: v1alpha

deps:
  - github.com/bufbuild/protoc-gen-validate@v0.10.1

generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
    
    # Validate via go run
    - command: ["go", "run", "github.com/bufbuild/protoc-gen-validate/cmd/protoc-gen-validate-go@v0.10.1"]
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true
```

## Example: Custom Command

You can use any command, not just `go run`:

```yaml
generate:
  plugins:
    # Run via node
    - command: ["node", "/path/to/protoc-gen-custom.js"]
      out: ./gen/custom
    
    # Run via python
    - command: ["python3", "-m", "protoc_gen_tool"]
      out: ./gen/python
    
    # Run executable file
    - command: ["./tools/protoc-gen-custom"]
      out: ./gen/custom
```

## Plugin Source Priority

EasyP uses the following priority when selecting a plugin source:

1. **`command`** — execute via specified command (highest priority)
2. **`remote`** — remote plugin via URL
3. **`name`** — local plugin from PATH or builtin plugin
4. **`path`** — path to plugin executable file

## Important Notes

- **Only one source**: Only one plugin source (`name`, `command`, `remote`, or `path`) should be specified
- **Versioning**: When using `go run` with a package from GitHub, always specify a version via `@version` for reproducibility
- **Performance**: Execution via `go run` is slower than using installed plugins, as compilation happens each time

## Code Generation

After configuring, run:

```bash
easyp generate
```

EasyP will automatically execute the specified commands to generate code.

