<!-- generated: 2026-07-27, template: core.md -->
# EasyP Architecture

## Overview

EasyP is a Go command-line application that uses handlers in `internal/api` to assemble a `core.Core` service from configuration, rules, and infrastructure adapters.

```text
CLI transport
cmd/easyp/main.go + urfave/cli
          |
          v
Application wiring
internal/api handlers and buildCore
          |
          v
Domain operations
internal/core + internal/rules
          |
          v
Adapters and infrastructure
internal/adapters + filesystem, Git, local cache, plugins
```

The primary executable is `cmd/easyp`. It registers command handlers for linting, modules, completion, initialization, generation, schema generation, file listing, configuration validation, and breaking-change checks.

## CLI Transport

Package: `cmd/easyp` and `internal/api`.

| File | Responsibility |
|------|----------------|
| `cmd/easyp/main.go` | Creates the `urfave/cli/v2` app, global flags, logger, and handler command list. |
| `internal/api/interface.go` | Defines the `Handler` interface used to supply `*cli.Command` values. |
| `internal/api/lint.go` | Resolves paths, loads config, invokes `Core.Lint`, and prints text or JSON issues. |
| `internal/api/generate.go` | Resolves generation roots and invokes `Core.Generate`. |
| `internal/api/mod.go` | Defines `mod download`, `mod update`, and `mod vendor`. |
| `internal/api/breaking_check.go` | Configures the comparison Git ref and invokes `Core.BreakingCheck`. |
| `internal/api/temporaly_helper.go` | Builds `core.Core` and converts config types into core types. |

Handlers own CLI flags, command-specific exit handling, configuration loading, and output presentation. They use `core.Core` for operations rather than implementing linting, generation, or dependency resolution directly.

## Application Wiring

`buildCore` in `internal/api/temporaly_helper.go` is the composition root for the main commands. It:

1. Builds selected lint rules with `rules.New`.
2. Opens the project lock file through `adapters/lock_file`.
3. Resolves `EASYPPATH`, defaulting to `$HOME/.easyp`.
4. Instantiates storage, module-config, console, and Git-walker adapters.
5. Combines `protobuf.mod` with Git repository generation inputs.
6. Converts configured plugins, inputs, managed-mode rules, and breaking-check settings.
7. Constructs `core.Core`.

The hard-coded vendor directory passed to the core is `easyp_vendor`. Package-manager behavior is documented in depth in [config/dependency.md](./config/dependency.md).

## Domain Operations

Package: `internal/core`.

| File | Responsibility |
|------|----------------|
| `core.go` | Defines `Core`, its injected dependencies, and shared sentinels. |
| `lint.go` | Downloads dependencies, walks `.proto` files, applies configured rules, and returns issues. |
| `generate.go` | Resolves imports and inputs, compiles descriptors, applies managed mode, executes plugins, and writes output. |
| `breaking_check.go` | Reads current and Git-ref proto states, collects entities by package, and compares them. |
| `download.go`, `update.go`, `get.go`, `vendor.go` | Implement module acquisition, lock updates, cached installation, and vendoring. |
| `dom.go` | Defines proto-domain representations such as `ProtoInfo`, `Issue`, and `ProtoData`. |
| `managed_mode.go` | Applies configured file and field option changes to descriptors. |

`Core` depends on consumer-facing interfaces such as `Rule`, `Storage`, `LockFile`, `ModuleConfig`, `CurrentProjectGitWalker`, and `DirWalker`. Concrete adapters are injected by the API layer.

## Adapters and Infrastructure

Package: `internal/adapters`.

| Package | Responsibility |
|---------|----------------|
| `adapters/storage` | Manages the cache and installed module trees under `EASYPPATH`. |
| `adapters/lock_file` | Reads, writes, iterates, and checks `protobuf.lock`. |
| `adapters/modfile` | Parses and writes `protobuf.mod` dependency declarations. |
| `adapters/repository/git` | Resolves revisions, fetches Git objects, reads repository files, and archives proto sources. |
| `adapters/go_git` | Provides directory walkers for project Git references. |
| `adapters/module_config` | Reads supported module layouts from a repository. |
| `adapters/plugin` | Supplies local, remote, built-in, and command plugin executors. |
| `adapters/console` | Runs local commands through platform-specific console implementations. |
| `adapters/prompter` | Provides interactive prompting. |

The `mcp/easypconfig` package is separate from command execution: it registers an MCP tool that describes the `easyp.yaml` schema and generates the schema metadata used by `schema-gen`.

## Directory Structure

```text
easyp/
├── cmd/easyp/                 # CLI executable
├── internal/
│   ├── api/                   # urfave/cli command handlers and composition root
│   ├── core/                  # lint, breaking, generation, and module workflows
│   │   ├── models/            # module, revision, lock, and error value types
│   │   └── path_helpers/      # path predicates used by core
│   ├── adapters/              # Git, storage, lockfile, plugin, and console adapters
│   ├── config/                # easyp.yaml parsing and validation
│   ├── rules/                 # concrete protobuf lint rules and registry
│   ├── fs/                    # filesystem walker implementation
│   ├── logger/                # logger abstraction and implementations
│   └── flags/                 # shared CLI flags
├── mcp/easypconfig/           # MCP config-description tool and schema model
├── schemas/                   # generated JSON Schema artifacts
├── docs/                      # Vite documentation site
└── Taskfile.yml               # build, test, lint, schema, and mock tasks
```

## Key Design Decisions

1. **CLI handlers are thin adapters.**
   - Each command implements `api.Handler` and returns a `*cli.Command`.
   - Handlers load configuration and map known errors to process behavior.

2. **Core logic uses injected interfaces.**
   - `Core.New` accepts storage, lock-file, module-config, console, Git-walker, and rule dependencies.
   - This supports mocks under `internal/core/mocks` in tests.

3. **Dependency resolution is Git-based.**
   - `Core.Download` is used before linting, generation, and breaking checks.
   - Resolved versions and hashes are persisted in `protobuf.lock`; installed sources are cached outside the project.

4. **Generation uses protobuf descriptors and plugin executors.**
   - `Core.Generate` compiles files with `protocompile`.
   - Plugin choice is selected from command, remote, built-in, or local sources.

5. **Schema metadata has a code source of truth.**
   - `mcp/easypconfig` reflects a schema model and exposes it through MCP.
   - Generated files in `schemas/` are refreshed through `task schema:generate`.

## Primary Data Flows

### Lint

```text
easyp lint
  -> api.Lint.Action
  -> config.New(easyp.yaml)
  -> api.buildCore
  -> Core.Lint
  -> Core.Download
  -> DirWalker.WalkDir(.proto files)
  -> Core.protoInfoRead + configured Rule.Validate
  -> []core.IssueInfo
  -> text or JSON output
```

### Generation

```text
easyp generate
  -> api.Generate.Action
  -> config.New + api.buildCore
  -> Core.Generate
  -> dependency and input resolution
  -> protocompile.Compiler.Compile
  -> optional ApplyManagedMode
  -> plugin.Executor.Execute
  -> GenerateBucket.DumpToFs
```

### Breaking Change Check

```text
easyp breaking --against <ref>
  -> api.BreakingCheck.Action
  -> Core.BreakingCheck
  -> current filesystem walker + Git-ref directory walker
  -> readProtoFiles and collect
  -> BreakingChecker.Check
  -> []core.IssueInfo
```
