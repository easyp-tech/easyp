<!-- generated: 2026-07-27, template: core.md -->
# EasyP Package Reference

This reference groups source packages by their role. Generated mocks are omitted except where they establish a test seam.

## Executable

### `cmd/easyp`
**CLI entry point** — creates the EasyP `urfave/cli/v2` application and installs its commands.

| File | Description |
|------|-------------|
| `main.go` | Initializes `slog`-backed logging, sets global flags, registers API handlers, and runs the app. |

The command list is created from values implementing `internal/api.Handler`.

## API Layer

### `internal/api`
**Command wiring** — maps CLI commands and flags to configuration parsing, `core.Core` construction, execution, and CLI output.

| File | Description |
|------|-------------|
| `interface.go` | Declares the `Handler` command-provider interface. |
| `lint.go` | Implements `lint`, issue output formats, and lint exit behavior. |
| `generate.go` | Implements `generate` and root resolution. |
| `breaking_check.go` | Implements `breaking` and its Git reference handling. |
| `mod.go` | Implements the `mod` command and module subcommands. |
| `init.go` | Implements project initialization. |
| `validate.go` | Implements configuration validation. |
| `ls_files.go` | Implements listing installed dependency files. |
| `schema_gen.go` | Implements config schema generation. |
| `completion.go` | Implements shell completion support. |
| `temporaly_helper.go` | Provides logger access, `buildCore`, managed-mode conversion, and cache-path resolution. |

The package instantiates adapters and translates `config.Config` into the core's dependency types.

## Application and Domain Layer

### `internal/core`
**Business workflows** — performs linting, protobuf compilation and generation, dependency management, and breaking-change analysis.

| File group | Description |
|------------|-------------|
| `core.go`, `dom.go` | Defines `Core`, its ports, proto representations, rules, issues, and plugin/input values. |
| `lint.go`, `check_lint_ignore.go` | Walks local proto files, reads imports, and applies rules and ignore behavior. |
| `breaking_check.go`, `breaking_checker.go` | Loads current and historical proto models and checks compatibility. |
| `generate.go`, `generate_bucket.go`, `generate_insertion_point.go` | Builds descriptors, invokes plugins, stages generated files, and supports insertion points. |
| `managed_mode.go` | Applies configured descriptor option changes. |
| `download.go`, `update.go`, `get.go`, `vendor.go` | Resolve, install, lock, update, and vendor modules. |
| `module_path.go`, `proto_info_read.go`, `fs.go` | Resolve module paths, parse proto imports, and define filesystem ports. |
| `init.go`, `init_template.go` | Provide project initialization behavior. |
| `instruction_parser.go` | Parses protobuf option/instruction names. |

The package owns consumer-side interfaces including `Storage`, `LockFile`, `ModuleConfig`, `Rule`, `Repo`, and `CurrentProjectGitWalker`.

### `internal/core/models`
**Module value objects and sentinels** — represents dependencies, resolved revisions, cached-install metadata, lock entries, and module errors.

| File | Description |
|------|-------------|
| `module.go` | Defines `Module`, `RequestedVersion`, `ModuleHash`, install metadata, and pseudo-version parsing. |
| `revision.go` | Defines resolved Git commit and version data. |
| `lock_file_info.go` | Defines a lock-file entry. |
| `module_config.go` | Defines module directories and transitive dependencies. |
| `cache_download_paths.go` | Defines cache archive and info paths. |
| `errors.go` | Defines dependency-management sentinel errors. |

### `internal/core/path_helpers`
**Path predicates** — contains helpers used when walking selected or ignored paths.

| File | Description |
|------|-------------|
| `is_target_path.go` | Tests whether a path matches the targeted path set. |

## Rule Layer

### `internal/rules`
**Concrete protobuf lint rules** — builds the selected rule set from lint configuration and implements buf-compatible naming, package, import, comment, enum, service, and RPC checks.

| File group | Description |
|------------|-------------|
| `builder.go` | Defines named rule groups and creates `[]core.Rule` from configuration. |
| `package_*.go`, `directory_*.go` | Validate package declarations, directories, and language package options. |
| `message_*.go`, `oneof_*.go`, `enum_*.go` | Validate message, oneof, and enum names and values. |
| `service_*.go`, `rpc_*.go` | Validate service and RPC naming, request/response forms, and streaming policy. |
| `import_*.go` | Validate import use and forbid weak or public imports. |
| `comment_*.go` | Validate comments for protobuf declarations. |
| `protovalidate.go` | Supports protovalidate-aware rule behavior. |

Each rule follows the `core.Rule` interface: it provides a message and validates a `core.ProtoInfo`.

## Configuration and Schema

### `internal/config`
**`easyp.yaml` parser and validator** — loads YAML after environment substitution and validates configuration objects.

| File | Description |
|------|-------------|
| `config.go` | Defines configuration, generation inputs/plugins, managed mode, parsing, and validation. |
| `lint.go` | Defines lint configuration. |
| `breaking_check.go` | Defines breaking-check configuration. |
| `default.go` | Provides default configuration values. |
| `yaml_validators.go`, `validate_raw.go` | Validate YAML nodes and raw configuration. |
| `plugin_opts.go` | Supports plugin option values expressed as scalars or sequences. |

### `mcp/easypconfig`
**Configuration schema metadata and MCP tool** — defines the config-schema model, generates JSON Schema, indexes schema paths, and serves schema descriptions.

| File | Description |
|------|-------------|
| `tool.go` | Registers the MCP config-description tool. |
| `describe.go` | Describes the complete schema or a requested schema path. |
| `schema.go`, `schema_model.go` | Reflects and models the JSON Schema source. |
| `tool_schemas.go` | Defines MCP tool input and output schemas. |
| `generate.go` | Generates schema content used by the CLI. |
| `spec_docs.go` | Holds schema documentation metadata. |

`schemas/` contains generated JSON Schema artifacts. Do not hand-edit them; regenerate through `task schema:generate`.

## Adapter Layer

### `internal/adapters/storage`
**Dependency cache and installation adapter** — manages archives, installed module directories, metadata, hashes, and lock-aware path lookup under `EASYPPATH`.

| File group | Description |
|------------|-------------|
| `storage.go`, `install.go` | Define storage and install downloaded module archives. |
| `cache_download.go`, `get_cache_download_paths.go` | Compute cached archive and metadata paths. |
| `get_install_dir.go`, `get_installed_module_hash.go` | Locate installed modules and verify their hash. |
| `read_installed_module_info.go`, `write_installed_module_info.go` | Persist install metadata. |
| `create_cache_repository_dir.go`, `sanitize.go` | Create Git-cache paths and sanitize version components. |

### `internal/adapters/lock_file`
**Lock-file adapter** — reads, writes, iterates, and checks the project `easyp.lock`.

| File | Description |
|------|-------------|
| `lock_file.go` | Defines the lock-file adapter. |
| `read.go`, `write.go` | Parse and serialize lock entries. |
| `deps_iter.go`, `is_empty.go` | Iterate entries and detect an empty lock file. |

### `internal/adapters/repository` and `internal/adapters/repository/git`
**Git repository adapter** — supplies repository operations required for dependency resolution.

| File group | Description |
|------------|-------------|
| `repository.go` | Declares the repository port. |
| `git/git.go` | Constructs the Git implementation. |
| `git/read_revision.go` | Resolves a requested version to a revision. |
| `git/fetch.go`, `git/archive.go` | Fetches Git objects and archives proto content. |
| `git/read_file.go`, `git/get_files.go` | Reads repository files and lists paths. |

### `internal/adapters/module_config`
**Remote module-layout reader** — reads module directories and dependencies from supported repository configuration.

| File | Description |
|------|-------------|
| `module_config.go` | Defines the adapter. |
| `read_from_repo.go` | Selects and reads supported module configuration. |
| `read_buf_work.go`, `read_easyp.go` | Read Buf and EasyP layouts. |

### `internal/adapters/plugin`
**Code-generation plugin executors** — executes local, remote, built-in, and command-based plugins.

| File group | Description |
|------------|-------------|
| `interface.go`, `options.go` | Define executor contracts and plugin invocation metadata. |
| `local.go`, `remote.go`, `command.go` | Implement execution sources. |
| `builtin.go`, `builtin_impl.go`, `wasm_embed.go` | Provide built-in plugin support. |

### `internal/adapters/go_git`
**Historical-tree adapter** — returns directory walkers for the current project's Git references.

### `internal/adapters/console`
**Local command adapter** — selects shell behavior and represents console errors.

### `internal/adapters/prompter`
**Interactive input adapter** — provides prompt behavior used by interactive operations.

## Shared Internal Packages

### `internal/fs`
**Filesystem access** — supplies directory walkers used by core operations.

### `internal/logger`
**Logging abstraction** — wraps the logger used by CLI setup, core workflows, and adapters.

### `internal/flags`
**Global command flags** — defines shared config, debug, and output-format flags.

### `internal/version`
**Build and compiler version information** — supplies the CLI and protobuf compiler version values.

## Package-Manager Reference

For dependency declaration, cache layout, lockfile format, Git resolution, transitive dependencies, and `easyp_vendor`, see [config/dependency.md](./config/dependency.md).
