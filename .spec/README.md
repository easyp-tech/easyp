<!-- generated: 2026-07-27, template: bootstrap.md -->
# EasyP Documentation

This folder contains documentation to help LLMs and developers quickly understand the project context.

## Documentation Index

### Core

- [ARCHITECTURE.md](./ARCHITECTURE.md) — Layered architecture (cmd → api → core → adapters)
- [PACKAGES.md](./PACKAGES.md) — Package reference for `internal/*`, `mcp/`, `cmd/`
- [DOMAIN.md](./DOMAIN.md) — Domain model (modules, lockfile, lint rules, plugins)
- [CODE_STYLE.md](./CODE_STYLE.md) — Go conventions expanded from agent-rules

### Development

- [TOOLS.md](./TOOLS.md) — Taskfile, golangci-lint, mockery, gotestsum
- [TESTING.md](./TESTING.md) — testify patterns, mocks, race tests

### Config & Dependencies

- [config/dependency.md](./config/dependency.md) — Git-native package manager (`easyp mod`, lockfile, cache)
- [agent-rules.md](./agent-rules.md) — Mandatory rules for AI agents

### Domain (CLI toolkit)

- [CLI.md](./CLI.md) — Commands, flags, exit codes
- [ERRORS.md](./ERRORS.md) — Domain / models sentinel errors and CLI mapping
- [DEPLOYMENT.md](./DEPLOYMENT.md) — Docker, CI/CD, releases

## Quick Facts

| Aspect | Technology |
|--------|------------|
| **Language** | Go 1.24 (`github.com/easyp-tech/easyp`) |
| **Product** | Protocol Buffers CLI toolkit |
| **CLI** | urfave/cli (`cmd/easyp`) |
| **Capabilities** | lint, breaking, generate, `easyp mod` |
| **Config** | `easyp.yaml` (+ envsubst); schema via `mcp/easypconfig` → `schemas/` |
| **Deps** | Git repositories; declare in `protobuf.mod`; lockfile `protobuf.lock`; cache `EASYPPATH` |
| **Build** | Task (`Taskfile.yml`) |
| **Lint** | golangci-lint (`.golangci.yml`) |
| **Tests** | gotestsum, `-race`, testify |
| **Human docs** | `docs/` (Vite site), https://easyp.tech |

## Project Structure

```
easyp/
├── cmd/easyp/              # CLI entrypoint
├── internal/
│   ├── api/                # CLI command wiring
│   ├── core/               # Business logic (lint, generate, mod, breaking)
│   ├── rules/              # Lint rules (+ colocated tests)
│   ├── config/             # easyp.yaml parse/validate
│   └── adapters/           # Git, storage, lockfile, plugins, console
├── mcp/easypconfig/        # Config schema metadata (source of truth)
├── schemas/                # Generated JSON Schema artifacts
├── docs/                   # Documentation site
├── .spec/                  # Agent-oriented project docs (this tree)
└── Taskfile.yml
```

## Running

```sh
task init              # install local tools (golangci-lint, gotestsum, mockery)
task build             # go build -o easyp ./cmd/easyp
task test              # gotestsum with -race and coverage
task lint              # golangci-lint (+ hadolint where applicable)
task quality           # test + lint
task schema:generate   # regenerate schemas/*.json
task schema:check      # fail if schemas drift
task mocks             # regenerate mockery mocks
```

Without Task:

```sh
go build -o easyp ./cmd/easyp
go test -race -count=1 ./...
go run ./cmd/easyp schema-gen
```

## Ports

N/A — EasyP is a CLI tool, not a long-running server. Remote plugin execution uses gRPC to an external EasyP plugin API when configured; there is no default listen port in-repo.

## Key Interfaces / Entry Points

| Entry | Role |
|-------|------|
| `cmd/easyp` | Process entry; registers CLI handlers |
| `internal/api` | Commands: lint, generate, breaking, mod, init, ls-files, schema-gen |
| `internal/core.Core` | Business logic facade |
| `internal/core` Storage / LockFile / ModuleConfig | Package-manager ports |
| `easyp.yaml` + `protobuf.mod` + `protobuf.lock` | Config, declared deps, and locked proto dependencies |
| `mcp/easypconfig` | Config schema / MCP metadata |

## Adding New Features

1. **Lint rule** — add `internal/rules/<rule>.go` + `_test.go`; register via existing rule builder; mirror neighboring rules.
2. **`easyp.yaml` schema** — change `mcp/easypconfig` and/or `internal/config`, then `task schema:generate` (never hand-edit `schemas/*.json`).
3. **CLI command** — wire in `internal/api`, register from `cmd/easyp`; follow existing handler patterns.
4. **Core/storage interfaces** — update interfaces, `task mocks`, add/adjust tests with `-race`.
5. Before finishing behavior changes: run tests for touched packages; for schema changes run `task schema:check`.
