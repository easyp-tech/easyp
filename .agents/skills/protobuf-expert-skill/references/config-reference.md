# EasyP Configuration Reference (easyp.yaml)

Complete reference for all configuration sections and options.

## Minimal Example

```yaml
lint:
  use:
    - DEFAULT
```

## Full Structure

```yaml
# ─── Lint ───────────────────────────────────────────────
lint:
  use:                              # Required. Rule names or group names.
    - DEFAULT                       # Groups: MINIMAL, BASIC, DEFAULT, COMMENTS, UNARY_RPC
    - COMMENTS
    - PACKAGE_NO_IMPORT_CYCLE       # Individual rule names also accepted

  except:                           # Rules to exclude (even if included by group)
    - SERVICE_SUFFIX
    - PACKAGE_VERSION_SUFFIX

  enum_zero_value_suffix: _UNSPECIFIED  # Suffix for ENUM_ZERO_VALUE_SUFFIX rule (default: _UNSPECIFIED)
  service_suffix: Service               # Suffix for SERVICE_SUFFIX rule (default: Service)

  allow_comment_ignores: true       # Allow // easyp:off / // easyp:on in proto files

  ignore:                           # Directories to skip entirely
    - proto/vendor
    - proto/third_party

  ignore_only:                      # Per-rule path ignores
    FIELD_LOWER_SNAKE_CASE:
      - proto/legacy
    COMMENT_FIELD:
      - proto/internal

# ─── Dependencies ──────────────────────────────────────
deps:
  - github.com/googleapis/googleapis              # Latest default branch
  - github.com/grpc-ecosystem/grpc-gateway@v2.0.0 # Specific tag/version
  - github.com/user/repo@abc123def                # Specific commit hash

# ─── Code Generation ──────────────────────────────────
generate:
  inputs:                           # Proto file sources
    - directory: proto              # Local directory (shorthand)
    - directory:                    # Full form with all fields
        path: proto                 #   Path to proto files
        root: .                     #   Root directory (default: ".")

    - git_repo:                     # Remote git repository
        url: github.com/user/repo@v1.0.0
        sub_directory: proto        # Optional: subdirectory within repo
        root: .                     # Optional: root within subdirectory

  plugins:
    - name: go                      # Plugin source (one of: name, remote, path, command)
      out: gen/go                   # Output directory
      opts:                         # Plugin-specific options (key-value)
        paths: source_relative
      with_imports: false           # Generate code for imported files too

    - remote: buf.build/grpc/go     # Remote plugin from registry
      out: gen/go
      opts:
        paths: source_relative

    - path: /usr/local/bin/protoc-gen-custom  # Local binary path
      out: gen/custom

    - command:                      # Command array to execute
        - docker
        - run
        - --rm
        - protoc-gen-custom
      out: gen/custom

  managed:                          # Managed mode — auto-set file/field options
    enabled: true

    disable:                        # Disable managed mode for specific targets
      - module: google.protobuf     # By module name
      - package: com.example        # By proto package
      - path: proto/internal        # By file path
      - file_option: go_package     # By file option name
      - field_option: "(custom)"    # By field option name
      - field: pkg.Message.field    # By fully qualified field name

    override:                       # Override specific options
      - file_option: go_package     # Override a file option
        value: github.com/myorg/pkg
        module: myapp               # Optional: limit to module
        package: com.example        # Optional: limit to proto package
        path: proto/api             # Optional: limit to path

      - field_option: "(validate.rules)"  # Override a field option
        value: true
        field: pkg.Message.field    # Optional: limit to specific field

      - file_option: java_package
        value: com.myorg.proto

# ─── Breaking Change Detection ────────────────────────
breaking:
  ignore:                           # Paths to exclude from breaking checks
    - proto/internal
    - proto/experimental

  against_git_ref: main             # Default git ref to compare against
```

## Plugin Source Priority

Each plugin must specify exactly ONE source:

| Source | Description | Example |
|--------|-------------|---------|
| `name` | Built-in plugin name | `go`, `go-grpc`, `grpc-gateway`, `openapiv2`, `validate-go` |
| `remote` | Registry URL | `buf.build/grpc/go` |
| `path` | Local binary path | `/usr/local/bin/protoc-gen-foo` |
| `command` | Command array | `["docker", "run", "gen-image"]` |

Plugin `opts` values can be scalars or arrays:

```yaml
opts:
  paths: source_relative                    # Scalar value
  require_unimplemented_servers: false       # Boolean value
```

## Managed Mode

When `managed.enabled: true`, EasyP automatically sets file options (like `go_package`, `java_package`) based on the proto file's package and path, reducing boilerplate.

Use `managed.disable` to exclude specific modules, packages, or paths from managed mode.
Use `managed.override` to set specific option values for specific modules.

## Validation

Always validate config after editing:

```sh
easyp validate-config
easyp validate-config --format json
```
