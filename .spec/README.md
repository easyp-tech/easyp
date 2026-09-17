<!-- generated: 2026-05-14, template: bootstrap.md -->
# EasyP Documentation

This folder contains documentation to help LLMs and developers quickly understand the EasyP project context.

## Documentation Index

### Core
- [ARCHITECTURE.md](./ARCHITECTURE.md) — Layered architecture, component diagram, data flows
- [PACKAGES.md](./PACKAGES.md) — Go package inventory with responsibilities
- [DOMAIN.md](./DOMAIN.md) — Core types, interfaces, domain rules
- [CODE_STYLE.md](./CODE_STYLE.md) — Naming conventions, error handling, patterns

### Development
- [TOOLS.md](./TOOLS.md) — Task runner, linters, mockery, GoReleaser
- [TESTING.md](./TESTING.md) — Test framework, fixtures, mocks, coverage
- [ERRORS.md](./ERRORS.md) — Error types, wrapping, exit codes

### Infrastructure
- [DEPLOYMENT.md](./DEPLOYMENT.md) — Docker, CI/CD, release process, Homebrew
- [CLI.md](./CLI.md) — Full CLI command reference

### Meta
- [AGENTS.md](./AGENTS.md) — How AI agents should use this directory
- [agent-rules.md](./agent-rules.md) — Mandatory rules for AI agents

## Quick Facts

| Aspect | Technology |
|--------|------------|
| **Language** | Go 1.24+ |
| **Module** | `github.com/easyp-tech/easyp` |
| **Architecture** | Clean/Hexagonal (cmd → api → core ← adapters) |
| **CLI Framework** | `github.com/urfave/cli/v2` |
| **Proto Parsing** | `github.com/yoheimuta/go-protoparser/v4` |
| **Proto Compiling** | `github.com/bufbuild/protocompile` |
| **Git Operations** | `github.com/go-git/go-git/v5` |
| **WASM Runtime** | `github.com/tetratelabs/wazero` |
| **MCP SDK** | `github.com/modelcontextprotocol/go-sdk` |
| **Testing** | `github.com/stretchr/testify` + `gotestsum` |
| **Mocking** | `github.com/vektra/mockery` v2.41.0 |
| **Linting** | `golangci-lint` v2.1.6 (staticcheck) |
| **Task Runner** | [Task](https://taskfile.dev) v3 |
| **Config** | YAML (`easyp.yaml`) with envsubst |
| **Release** | GoReleaser v2 |
| **CI** | GitHub Actions |

## Project Structure

```
easyp/
├── cmd/easyp/              # CLI entrypoint (main.go)
├── internal/
│   ├── api/                # CLI command handlers (Handler interface)
│   ├── config/             # easyp.yaml parsing, validation, types
│   ├── core/               # Business logic: lint, generate, download, breaking
│   │   ├── models/         # Domain value objects
│   │   ├── mocks/          # Generated mocks for core interfaces
│   │   ├── path_helpers/   # Path utility functions
│   │   └── templates/      # Init template files
│   ├── rules/              # Lint rule implementations (one file per rule)
│   ├── adapters/
│   │   ├── console/        # Console adapter (stdin/stdout)
│   │   ├── go_git/         # Git walker adapter
│   │   ├── lock_file/      # Lock file adapter
│   │   ├── module_config/  # Module config adapter
│   │   ├── plugin/         # Plugin executors (local, remote, builtin, command)
│   │   ├── prompter/       # Interactive prompts (bubbletea)
│   │   ├── repository/     # Git repository adapter
│   │   └── storage/        # Dependency cache/storage
│   ├── flags/              # Global CLI flag definitions
│   ├── fs/                 # Filesystem abstractions (DirWalker, FS)
│   ├── logger/             # Structured logger wrapper (slog)
│   ├── schemagen/          # JSON Schema generation logic
│   └── version/            # Build version info
├── mcp/easypconfig/        # MCP tool: easyp_config_describe
├── schemas/                # Generated JSON Schema artifacts
├── wellknownimports/       # Embedded well-known proto imports
├── testdata/               # Test fixtures (proto files)
├── docs/                   # Documentation site (Node.js)
├── .agents/skills/sdd/     # Spec-Driven Development skill
├── Taskfile.yml            # Task runner configuration
├── Dockerfile              # Multi-stage Docker build
├── .goreleaser.yaml        # Release configuration
├── easyp.yaml              # EasyP's own config (dogfooding)
└── easyp.lock              # Lock file for proto dependencies
```

## Running

```sh
task init               # Install dev tools (linter, gotestsum, mockery)
task build              # Build binary
task test               # Run tests (-race, -count=1)
task lint               # Run golangci-lint + hadolint
task quality            # Run both test and lint
task coverage           # Open coverage report in browser
task mocks              # Regenerate all mocks
task schema:generate    # Regenerate JSON Schema artifacts
task install            # Install binary to GOPATH
```

## Key Interfaces / Entry Points

| Interface | Location | Purpose |
|-----------|----------|---------|
| `Handler` | `internal/api/interface.go` | CLI command handler: `Command() *cli.Command` |
| `Rule` | `internal/core/dom.go` | Lint rule: `Message()` + `Validate(ProtoInfo)` |
| `DirWalker` | `internal/core/fs.go` | Filesystem walking + read/write |
| `FS` | `internal/core/fs.go` | File operations: Open, Create, Exists, Remove |
| `Storage` | `internal/core/` | Dependency cache management |
| `Executor` | `internal/adapters/plugin/interface.go` | Plugin execution |
| `Repo` | `internal/core/dom.go` | Git repository operations |

## Adding New Features

### Adding a new lint rule
1. Create `internal/rules/<rule_name>.go` with struct implementing `core.Rule`
2. Create `internal/rules/<rule_name>_test.go` with test cases
3. Add test fixtures to `testdata/` if needed
4. Register in `internal/rules/builder.go` (add to the appropriate group function)
5. Rule name is auto-derived: `PascalCase` → `UPPER_SNAKE_CASE`

### Adding a new CLI command
1. Create `internal/api/<command>.go` with struct implementing `Handler`
2. Add `var _ Handler = (*CommandName)(nil)` for compile-time check
3. Register in `cmd/easyp/main.go` via `buildCommand(...)` call
