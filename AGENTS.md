# AGENTS.md — EasyP Project Guide

This document is a reference for AI agents working on the EasyP codebase.
It describes the project's purpose, tech stack, architecture, coding patterns,
and development workflows.

## Project Overview

**EasyP** is a modern CLI toolkit for Protocol Buffers development. It provides:

- **Linter** — enforces API design rules (compatible with buf's rule set).
- **Breaking change detector** — source-level API compatibility checks against Git branches.
- **Code generator** — invokes local, remote, builtin, and command-based protoc plugins.
- **Package manager** — Git-based dependency management with a lock file (`easyp.lock`).
- **Config validator** — validates `easyp.yaml` structure and emits JSON/text diagnostics.
- **MCP integration** — `easyp_config_describe` tool for the Model Context Protocol.

Repository: `github.com/easyp-tech/easyp`

## Tech Stack

| Area            | Technology                                                   |
| --------------- | ------------------------------------------------------------ |
| Language        | Go 1.24+                                                     |
| CLI framework   | `github.com/urfave/cli/v2`                                   |
| Proto parsing   | `github.com/yoheimuta/go-protoparser/v4`                     |
| Proto compiling | `github.com/bufbuild/protocompile`                           |
| Git operations  | `github.com/go-git/go-git/v5`                                |
| WASM runtime    | `github.com/tetratelabs/wazero` (builtin plugins)            |
| MCP SDK         | `github.com/modelcontextprotocol/go-sdk`                     |
| JSON Schema     | `github.com/invopop/jsonschema`, `github.com/google/jsonschema-go` |
| TUI             | `github.com/charmbracelet/bubbletea`, `lipgloss`             |
| Testing         | `github.com/stretchr/testify`, `gotestsum`                   |
| Mocking         | `github.com/vektra/mockery` (v2.41.0)                        |
| Linting         | `golangci-lint` (v2.1.6, staticcheck enabled)                |
| Task runner     | [Task](https://taskfile.dev) v3 (`Taskfile.yml`)             |
| Config format   | YAML (`easyp.yaml`), env-var expansion via `envsubst`        |
| Release         | GoReleaser v2 (multi-arch binaries, Docker, Homebrew)        |
| CI              | GitHub Actions (`tests.yml`, `release.yml`, `docs.yml`)      |

## Directory Structure

```
.
├── cmd/easyp/              # CLI entrypoint (main.go)
├── internal/
│   ├── api/                # CLI command handlers (implement Handler interface)
│   ├── config/             # easyp.yaml parsing, validation, types
│   ├── core/               # Business logic: lint, generate, download, breaking, etc.
│   │   ├── models/         # Domain value objects (Revision, CacheDownloadPaths, etc.)
│   │   ├── mocks/          # Generated mocks for core interfaces
│   │   ├── path_helpers/   # Path utility functions
│   │   └── templates/      # Init template files
│   ├── rules/              # Lint rule implementations (one file per rule)
│   ├── adapters/
│   │   ├── console/        # Console adapter (stdin/stdout interaction)
│   │   ├── go_git/         # Git walker adapter (go-git)
│   │   ├── lock_file/      # Lock file adapter
│   │   ├── module_config/  # Module config adapter
│   │   ├── plugin/         # Plugin executors (local, remote, builtin, command)
│   │   │   └── wasm/       # WASM embedded plugins
│   │   ├── prompter/       # Interactive prompts
│   │   ├── repository/     # Git repository adapter
│   │   └── storage/        # Dependency cache/storage
│   ├── flags/              # Global CLI flag definitions
│   ├── fs/                 # Filesystem abstractions (DirWalker, FS interface)
│   ├── logger/             # Structured logger wrapper (slog)
│   ├── schemagen/          # JSON Schema generation logic
│   └── version/            # Build version info
├── mcp/easypconfig/        # MCP tool: easyp_config_describe (source-of-truth for schema)
├── schemas/                # Generated JSON Schema artifacts
├── wellknownimports/       # Embedded well-known proto imports (google/protobuf/*)
├── testdata/               # Test fixtures (proto files for rule tests)
├── docs/                   # Documentation site (Node.js, deployed as Docker)
├── .spec/                  # AI-agent documentation (architecture, CLI, domain, etc.)
├── .github/workflows/      # CI pipelines
├── .agents/skills/sdd/     # Spec-Driven Development skill for AI agents
├── Taskfile.yml            # Task runner configuration
├── Dockerfile              # Multi-stage Docker build
├── .goreleaser.yaml        # Release configuration
├── easyp.yaml              # EasyP's own config (dogfooding)
└── easyp.lock              # Lock file for proto dependencies
```

## Architecture

### Layered Architecture

The project follows a **clean/hexagonal architecture** pattern with three layers:

```
cmd/easyp (main)  →  internal/api (handlers)  →  internal/core (business logic)
                                                        ↓
                                               internal/adapters (infrastructure)
```

1. **`cmd/easyp`** — Entrypoint. Creates `cli.App`, registers handlers, initializes logger.
2. **`internal/api`** — CLI command handlers. Each handler implements the `Handler` interface (`Command() *cli.Command`). Responsible for parsing flags, reading config, building `Core`, and formatting output.
3. **`internal/core`** — All business logic. The `Core` struct is the central orchestrator. It depends on interfaces (ports), not concrete implementations.
4. **`internal/adapters`** — Infrastructure implementations of core interfaces (storage, git, plugins, console).

### Layer Dependency Rules

- `cmd/` may only import `internal/api/` and `internal/flags/`.
- `internal/api/` may import `internal/core/`, `internal/config/`, `internal/flags/`, `internal/adapters/`.
- `internal/core/` may **NOT** import `internal/adapters/` — depends on interfaces only.
- `internal/adapters/` may import `internal/core/` (to implement its interfaces).

### Key Interfaces (Ports)

Defined in `internal/core/`:

| Interface                  | Purpose                                        |
| -------------------------- | ---------------------------------------------- |
| `Rule`                     | Lint rule: `Message() string`, `Validate(ProtoInfo) ([]Issue, error)` |
| `DirWalker`                | Filesystem walking + read/write abstraction    |
| `FS`                       | File operations: `Open`, `Create`, `Exists`, `Remove` |
| `Storage`                  | Dependency cache management                    |
| `ModuleConfig`             | Module configuration reader                    |
| `LockFile`                 | Lock file read/write                           |
| `CurrentProjectGitWalker`  | Git-based directory walker for breaking checks |
| `Repo`                     | Git repository operations (files, archive, revisions) |

Defined in `internal/adapters/plugin/`:

| Interface  | Purpose                                                  |
| ---------- | -------------------------------------------------------- |
| `Executor` | Plugin execution: `Execute(ctx, plugin, request) (response, error)` |

There are four executor implementations: `local`, `remote`, `builtin` (WASM), and `command`.

### Handler Pattern

Each CLI command is a struct in `internal/api/` implementing:

```go
type Handler interface {
    Command() *cli.Command
}
```

Handlers are registered in `cmd/easyp/main.go` via `buildCommand(...)`.

Current handlers: `Lint`, `Mod`, `Completion`, `Init`, `Generate`, `SchemaGen`, `LsFiles`, `Validate`, `BreakingCheck`.

### Lint Rule Pattern

Each rule is a separate file pair in `internal/rules/`:
- `<rule_name>.go` — implementation
- `<rule_name>_test.go` — tests

A rule is a struct implementing `core.Rule`:

```go
var _ core.Rule = (*RuleName)(nil)

type RuleName struct{
    // Optional config fields
}

func (r *RuleName) Message() string {
    return "human-readable error message"
}

func (r *RuleName) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
    var res []core.Issue
    // Iterate over proto elements, check conditions
    // Use core.AppendIssue() to add issues (respects nolint comments)
    return res, nil
}
```

Rules are organized in groups: `MINIMAL`, `BASIC`, `DEFAULT`, `COMMENTS`, `UNARY_RPC`.
The `internal/rules/builder.go` handles group expansion and rule construction.

Rule names are auto-derived from struct names via `PascalCase → UPPER_SNAKE_CASE`
(e.g., `EnumPascalCase` → `ENUM_PASCAL_CASE`).

### Plugin Executors

Code generation supports four plugin types, each with its own `Executor`:

| Type      | Source field | Description                                 |
| --------- | ------------ | ------------------------------------------- |
| `local`   | `name`       | Invokes `protoc-gen-<name>` from PATH       |
| `remote`  | `remote`     | Calls EasyP API service                     |
| `builtin` | (internal)   | Runs WASM plugins via wazero                |
| `command` | `command`    | Executes arbitrary command                  |
| `path`    | `path`       | Invokes plugin at specific filesystem path  |

## Development Workflow

### Prerequisites

```sh
# Install all dev tools (linter, gotestsum, mockery)
task init
```

### Common Commands

```sh
task build              # Build binary
task test               # Run tests with gotestsum (-race, -count=1)
task lint               # Run golangci-lint + hadolint
task quality            # Run both test and lint
task coverage           # Open coverage report in browser
task mocks              # Regenerate all mocks
task schema:generate    # Regenerate JSON Schema artifacts
task schema:check       # Ensure schemas are up to date (CI)
task install            # Install binary to GOPATH
```

### Testing Conventions

- Tests use `testify` (`assert`, `require`).
- Test fixtures live in `testdata/` (proto files organized by scenario).
- Rule tests follow a consistent pattern: each rule has a dedicated `*_test.go`.
- Mocks are generated with `mockery` into `mocks/` subdirectories next to the interfaces.
- Mockery config: `with-expecter: true`, `inpackage: false`, `disable-version-string: true`.
- Tests run with `-race -count=1` by default.
- External test packages (`package rules_test`) — tests access only the public API.
- All tests run in parallel (`t.Parallel()` at top and in subtests).
- Table-driven tests use `map[string]struct{}` (map keys are test names).
- Loop variable capture: `name, tc := name, tc` before `t.Run()`.
- Shared `start(t)` helper in rule tests: parses fixtures, returns asserter + proto map.

**Running specific tests:**
```sh
go test -v -run TestEnumPascalCase ./internal/rules/  # single test
go test -race -count=1 ./internal/core/                # single package
```

### CI Pipeline

| Workflow        | Trigger              | What it does                           |
| --------------- | -------------------- | -------------------------------------- |
| `tests.yml`     | push to main, all PRs | `task init` → `task test`             |
| `release.yml`   | tag push             | GoReleaser: binaries, Docker, Homebrew |
| `docs.yml`      | push to main (docs/) | Build & deploy docs site               |

### Release Process

- Releases are managed by GoReleaser v2 (`.goreleaser.yaml`).
- Multi-arch builds: `darwin`, `linux`, `windows` × `amd64`, `arm64`, `arm`.
- Docker images: `ghcr.io/easyp-tech/easyp` (multi-arch manifest, linux/amd64 + linux/arm64).
- Homebrew tap: `easyp-tech/homebrew-tap`.
- Triggered by pushing a Git tag.

## Coding Conventions

### General

- **Go version**: 1.24+ (specified in `go.mod`).
- **Module path**: `github.com/easyp-tech/easyp`.
- **Error wrapping**: Always wrap errors with context: `fmt.Errorf("funcName: %w", err)`.
- **Logging**: Use the structured `logger.Logger` wrapper around `slog`. Log to stderr.
- **No global state**: Dependencies are injected via constructors (e.g., `core.New(...)`).
- **Interface compliance**: Use `var _ Interface = (*Struct)(nil)` to verify at compile time.

### Naming

- **Files**: `snake_case.go` (one type/concern per file).
- **Rule files**: Named after the rule in `snake_case` (e.g., `enum_pascal_case.go`).
- **Test files**: `*_test.go` alongside the source file.
- **Mock directories**: `mocks/` subdirectory next to the interface definition.
- **Packages**: Short, lowercase, no underscores.

### Import Ordering

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "os"

    // 2. Third-party
    "github.com/urfave/cli/v2"

    // 3. Internal packages
    "github.com/easyp-tech/easyp/internal/config"
    "github.com/easyp-tech/easyp/internal/core"
)
```

Groups separated by blank lines, sorted alphabetically within each group.

### Config & Schema

- `easyp.yaml` is the single configuration file (parsed in `internal/config/`).
- Environment variable expansion is supported via `envsubst` (escape with `$$`).
- JSON Schema artifacts in `schemas/` are auto-generated from Go types (`task schema:generate`).
- The `mcp/easypconfig` package is the source-of-truth for schema + MCP tool metadata.
- **Always regenerate schemas** after modifying config types: `task schema:generate`.

### Docker

- Multi-stage build: `golang:1.25-alpine` builder → `alpine:3.22` runtime.
- Runtime includes: `ca-certificates`, `tzdata`, `git`, `bash`.
- CGO disabled (`CGO_ENABLED=0`).

## Important Domain Concepts

### Proto Info

`core.ProtoInfo` is the central data structure passed to every lint rule. It contains:
- `Path` — file path of the proto file
- `Info` — parsed unordered proto AST (`*unordered.Proto`)
- `ProtoFilesFromImport` — map of imported proto files (for cross-file checks)

### Issue Reporting

Issues are reported via `core.Issue` with position, source name, message, and rule name.
Use `core.AppendIssue()` — it automatically checks for `nolint:` / `buf:lint:ignore` comments.

### Dependency Management

- Dependencies are declared in `easyp.yaml` under `deps:` (Git URLs with optional `@version`).
- Downloaded to a local cache (`~/.cache/easyp/` or similar).
- Lock file (`easyp.lock`) tracks exact revisions.
- Well-known imports (`google/protobuf/*`) are embedded in the binary via `wellknownimports/`.

### Breaking Change Detection

- Compares the current proto files against a Git reference (branch/tag/commit).
- Uses `go-git` to read files from the reference, then compiles both versions
  with `protocompile` and compares descriptors.

## Error Handling

### Exit Codes

| Code | Meaning | Triggering Errors |
|------|---------|-------------------|
| `0` | Success / no issues | — (or `ErrEmptyInputFiles` as warning) |
| `1` | Issues found | `ErrHasLintIssue`, `ErrBreakingCheckIssue`, `ErrHasValidateIssue`, `ErrVersionNotFound` |
| `2` | Infrastructure error | `OpenImportFileError`, `GitRefNotFoundError`, `ErrRepositoryDoesNotExist` |

### Error Types

**Sentinel errors** — identity-only, compared with `errors.Is()`:
```go
var ErrVersionNotFound = errors.New("version not found")         // models/errors.go
var ErrRepositoryDoesNotExist = errors.New("repository does not exist") // core/core.go
```

**Typed errors** — carry extra context, compared with `errors.As()`:
```go
type OpenImportFileError struct{ FileName string }  // core/dom.go
type GitRefNotFoundError struct{ GitRef string }     // core/dom.go
```

### Wrapping Convention

Always wrap with the calling function name:
```go
return fmt.Errorf("storage.Download: %w", err)  // adapter
return fmt.Errorf("core.Download: %w", err)     // core
```

### Logging Rule

**Log at the top (API), wrap at the bottom (adapters/core).** Never log and return the same error.

| Layer | Level | Action |
|-------|-------|--------|
| API | `Error`/`Warn` | Log + `os.Exit()` |
| Core | — | Wrap and return |
| Adapters | `Debug` | Infrastructure details only |

## Documentation

Detailed documentation lives in `.spec/`:

| Document | Purpose |
|----------|---------|
| `README.md` | Documentation index, quick facts, project structure |
| `ARCHITECTURE.md` | Layered architecture, component diagram, data flow |
| `PACKAGES.md` | Go package inventory with responsibilities |
| `DOMAIN.md` | Core types, interfaces, domain rules |
| `CODE_STYLE.md` | Layer rules, naming, error propagation, import ordering |
| `CLI.md` | Full CLI command reference (flags, exit codes, I/O) |
| `TOOLS.md` | Task runner commands, dev setup |
| `TESTING.md` | Test framework, patterns, mocks, fixtures |
| `ERRORS.md` | Complete error catalog, wrapping, exit code mapping |
| `DEPLOYMENT.md` | Docker, CI/CD, GoReleaser, release process |
| `agent-rules.md` | Mandatory rules for AI agents |

Use `.spec/` for deep dives. Use this `AGENTS.md` as the entry point.

## Skills

The project includes the **SDD (Spec-Driven Development)** skill in `.agents/skills/sdd/`.
Use it for structured feature development with the 6-phase pipeline:
`Explore → Requirements → Design → Task Plan → Implementation → Review`.

See `.agents/skills/sdd/SKILL.md` for full documentation.
