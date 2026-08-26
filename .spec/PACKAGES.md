<!-- generated: 2026-05-14, template: core.md -->
# EasyP Packages

## API Layer

### `internal/api`
**CLI command handlers** — Each struct implements `Handler` interface (`Command() *cli.Command`).

| File | Description |
|------|-------------|
| `lint.go` | Lint command handler + text/JSON output printers |
| `generate.go` | Generate command handler + `resolveRoots()` helper |
| `breaking_check.go` | Breaking change detection handler |
| `init.go` | Interactive config initialization handler |
| `mod.go` | Package manager handler (download/update/vendor) |
| `validate.go` | Config validation handler |
| `ls_files.go` | Proto file listing handler |
| `schema_gen.go` | JSON Schema generation handler |
| `completion.go` | Shell completion script generator |
| `temporaly_helper.go` | Shared: `buildCore()`, `getLogger()`, `getEasypPath()`, `convertManagedModeConfig()` |

### `internal/flags`
**Global CLI flag definitions** — Shared flags (`--cfg`, `--debug`, `--format`) and format helpers.

| File | Description |
|------|-------------|
| `flags.go` | `Config`, `DebugMode`, `Format` flag definitions |
| `format.go` | `GetFormat()` helper, format constants (`text`, `json`) |
| `enum_value.go` | `EnumValue` type for `--format` flag validation |

## Business Logic Layer

### `internal/core`
**Central business logic** — `Core` struct orchestrates all operations.

| File | Description |
|------|-------------|
| `core.go` | `Core` struct, `New()` constructor, dependency injection |
| `dom.go` | Domain types: `Rule`, `Issue`, `ProtoInfo`, `Plugin`, `Repo` |
| `lint.go` | `Core.Lint()` — walks files, applies rules |
| `generate.go` | `Core.Generate()` — compiles protos, invokes plugins |
| `breaking.go` | `Core.BreakingCheck()` — compares proto versions via git |
| `download.go` | `Core.Download()` — fetches dependencies |
| `initialize.go` | `Core.Initialize()` — creates easyp.yaml via templates |
| `ls_files.go` | `Core.ListFiles()` — lists proto files with imports |
| `vendor.go` | `Core.Vendor()` — copies deps to vendor dir |
| `update.go` | `Core.Update()` — updates dependency versions |
| `nolint.go` | `nolint:` / `buf:lint:ignore` comment parser |
| `managed_mode.go` | Managed mode for proto file options |

### `internal/core/models`
**Domain value objects** — `Revision`, `CacheDownloadPaths`, `RequestedVersion`.

### `internal/core/path_helpers`
**Path utility functions** — Helpers for path resolution in proto files.

### `internal/core/templates`
**Init templates** — Embedded template files for `easyp init`.

### `internal/core/mocks`
**Generated mocks** — Mockery-generated mocks for core interfaces.

### `internal/rules`
**Lint rule implementations** — One file per rule.

| File | Description |
|------|-------------|
| `builder.go` | Group expansion, rule construction from config |
| `enum_pascal_case.go` | `ENUM_PASCAL_CASE` rule |
| `field_lower_snake_case.go` | `FIELD_LOWER_SNAKE_CASE` rule |
| `service_suffix.go` | `SERVICE_SUFFIX` rule |
| `package_defined.go` | `PACKAGE_DEFINED` rule |
| ... | ~40+ rule files (one per rule) |

Rule groups: `MINIMAL`, `BASIC`, `DEFAULT`, `COMMENTS`, `UNARY_RPC`.

## Adapters Layer

### `internal/adapters/storage`
**Dependency cache management** — Stores downloaded proto dependencies in `$EASYPPATH` (`~/.easyp`).

### `internal/adapters/go_git`
**Git walker adapter** — Implements `CurrentProjectGitWalker` for reading proto files from git refs (used by breaking check).

### `internal/adapters/lock_file`
**Lock file adapter** — Reads/writes `easyp.lock` with resolved dependency revisions.

### `internal/adapters/module_config`
**Module config adapter** — Reads module configuration from downloaded dependencies.

### `internal/adapters/console`
**Console adapter** — stdin/stdout interaction abstraction.

### `internal/adapters/prompter`
**Interactive prompter** — TUI prompts using `charmbracelet/bubbletea` for `easyp init`.

### `internal/adapters/repository`
**Git repository adapter** — Implements `Repo` interface. Sub-package `git/` handles low-level git operations.

### `internal/adapters/plugin`
**Plugin executors** — Implements `Executor` interface with 4 strategies.

| Sub-package | Executor Type | Description |
|-------------|--------------|-------------|
| (root) | `local` | Invokes `protoc-gen-<name>` from PATH |
| (root) | `remote` | Calls EasyP API service |
| (root) | `path` | Invokes plugin at specific filesystem path |
| (root) | `command` | Executes arbitrary command |
| `wasm/` | `builtin` | Runs WASM plugins via `wazero` |

## Shared / Internal

### `internal/config`
**Configuration parsing** — Reads and validates `easyp.yaml`. Types: `Config`, `LintConfig`, `Generate`, `BreakingCheck`, `Plugin`.

### `internal/fs`
**Filesystem abstractions** — `DirWalker` and `FS` interfaces.

| Sub-package | Description |
|-------------|-------------|
| `fs/` | OS filesystem walker (`FSWalker`) |
| `go_git/` | Git-based filesystem walker |

### `internal/logger`
**Structured logger** — Wrapper around `slog`. Provides `Logger` interface with `Debug`, `Info`, `Warn`, `Error` methods.

### `internal/schemagen`
**JSON Schema generation** — Generates schema files from Go types for `easyp.yaml` config validation.

### `internal/version`
**Build version info** — Version, commit, and build date injected via `-ldflags`.

## Top-Level Packages

### `mcp/easypconfig`
**MCP tool** — `easyp_config_describe` tool for Model Context Protocol. Source-of-truth for JSON Schema.

### `wellknownimports`
**Embedded well-known protos** — `google/protobuf/*` proto files embedded in binary via `//go:embed`.
