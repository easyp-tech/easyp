# AGENTS.md

Instructions for AI coding agents working in this repository.

## Project overview

EasyP is a Protocol Buffers CLI toolkit (`github.com/easyp-tech/easyp`, Go 1.24) that provides:

- linting (buf-compatible rules)
- breaking change detection
- code generation (local and remote plugins)
- Git-based dependency / package management (`easyp mod`)

Human-facing docs: [README.md](README.md) and https://easyp.tech.

## Layout

| Path | Role |
|------|------|
| `cmd/easyp` | CLI entrypoint |
| `internal/api` | CLI command wiring |
| `internal/core` | Business logic (lint, generate, mod, breaking) |
| `internal/rules` | Lint rules (each rule + colocated `_test.go`) |
| `internal/config` | `easyp.yaml` / `protobuf.mod` parsing and validation |
| `internal/adapters` | Git, storage, lockfile, plugins |
| `mcp/easypconfig` | MCP tool + config schema metadata (source of truth) |
| `schemas/` | Generated JSON Schema artifacts |
| `docs/` | Documentation site (Vite) |

## What is .spec/
.spec/ is a project documentation directory optimized for AI agent (LLM) consumption. It contains:

* Structured descriptions of architecture, packages, and domain model
* Code conventions and testing patterns
* Infrastructure and tooling descriptions
* Agent rules (agent-rules.md) with mandatory coding standards
* Purpose: give the agent full project context without reading the entire source code.


## Build, test, lint

Prefer Task targets from [`Taskfile.yml`](Taskfile.yml):

```sh
task init              # install local tools (golangci-lint, gotestsum, mockery)
task build             # go build -o easyp ./cmd/easyp
task test              # gotestsum with -race and coverage
task lint              # golangci-lint (+ hadolint where applicable)
task quality           # test + lint
task schema:generate   # regenerate schemas/*.json
task schema:check      # generate and fail if schemas drift
task mocks             # regenerate mockery mocks for core/storage interfaces
```

Equivalents without Task:

```sh
go build -o easyp ./cmd/easyp
go test -race -count=1 ./...
go run ./cmd/easyp schema-gen
```

After behavior changes, run the relevant tests (at least the packages you touched). For config schema changes, run `task schema:check` before finishing.

## Conventions

- **New lint rules**: add `internal/rules/<rule>.go` + `_test.go`, register via the existing rule builder; mirror patterns in neighboring rules.
- **`easyp.yaml` schema**: change metadata/validation in `mcp/easypconfig` and/or `internal/config`, then regenerate `schemas/` with `task schema:generate`. Do not hand-edit generated schema JSON.
- **Tests**: use `testify`; regenerate mocks with `task mocks` when core/storage interfaces change.
- Prefer pointing agents/humans at README and docs over copying long usage text into this file.

## Do not

- Commit secrets, credentials, or unrelated local artifacts.
- Hand-edit `schemas/easyp-config-v1.schema.json` or `schemas/easyp-config.schema.json` without regenerating.
- Commit or “fix” noise under `docs/node_modules`, `coverage.out`, or the local `easyp` binary in the repo root.
- Add parallel agent instruction files (`.cursor/rules/`, `CLAUDE.md`) unless explicitly requested; keep this file the source of truth.
