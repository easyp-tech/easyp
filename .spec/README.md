<!-- generated: 2026-09-30, template: bootstrap.md -->
# EasyP Documentation

This folder contains documentation to help LLMs and developers quickly understand the project context.

## Documentation Index

### Core

- [ARCHITECTURE.md](./ARCHITECTURE.md) — CLI wiring, modules, generation, migration, engines and adapters
- [PACKAGES.md](./PACKAGES.md) — Package reference for <code>internal/*</code>, <code>mcp/</code>, <code>cmd/</code>
- [DOMAIN.md](./DOMAIN.md) — Domain model (modules, lockfile, lint rules, plugins)
- [CODE_STYLE.md](./CODE_STYLE.md) — Go conventions expanded from agent-rules

### Development

- [TOOLS.md](./TOOLS.md) — Taskfile, golangci-lint, mockery, gotestsum
- [TESTING.md](./TESTING.md) — testify patterns, mocks, race tests

### Config & Dependencies

- [config/dependency.md](./config/dependency.md) — Git-native package manager (<code>easyp mod</code>, lockfile, cache)
- [agent-rules.md](./agent-rules.md) — Mandatory rules for AI agents

### Domain (CLI toolkit)

- [CLI.md](./CLI.md) — Commands, flags, exit codes
- [ERRORS.md](./ERRORS.md) — Current error types, CLI flow and unresolved mappings
- [DEPLOYMENT.md](./DEPLOYMENT.md) — Docker, CI/CD, releases

## Quick Facts

| Aspect | Technology |
|--------|------------|
| **Language** | Go 1.26.6 (<code>github.com/easyp-tech/easyp</code>) |
| **Product** | Protocol Buffers CLI toolkit |
| **CLI** | urfave/cli v2 (<code>cmd/easyp</code>) |
| **Capabilities** | lint, breaking, generate, get, mod, init, migrate, ls-files, validate-config, schema-gen, completion |
| **Config** | v1 <code>easyp.yaml</code> policy + <code>easyp.gen.yaml</code> generation; <code>internal/config/v1</code> models/schema, <code>internal/schemagen</code> writer |
| **Deps** | Git repositories; declare in <code>protobuf.mod</code>; lockfile <code>protobuf.lock</code>; cache <code>EASYPPATH</code> |
| **Build** | Task (<code>Taskfile.yml</code>) |
| **Lint** | golangci-lint (<code>.golangci.yml</code>) |
| **Tests** | gotestsum, <code>-race</code>, testify |
| **Human docs** | Root [README.md](../README.md); site maintained in separate <code>easyp-tech/docs</code> repository, https://easyp.tech |

## Project Structure

~~~
easyp/
├── cmd/easyp/              # CLI entrypoint
├── internal/
│   ├── api/                # CLI command wiring
│   ├── core/               # Lint/breaking engines, parsing, plugin execution
│   ├── modules/            # Resolver, native manifests, locks, imports, vendor
│   ├── generation/         # Consumer discovery, module selection, descriptor export
│   ├── migration/          # Legacy conversion plans and guarded apply
│   ├── rules/              # Lint rules (+ colocated tests)
│   ├── config/v1/          # Native v1 models, parsing, validation and schemas
│   ├── schemagen/          # Writes generated schemas
│   └── adapters/           # Git/cache, dependency metadata, plugins, prompts
├── mcp/easypconfig/        # MCP descriptions consuming the v1 schema
├── schemas/               # Six generated JSON Schema artifacts
├── .spec/                  # Agent-oriented project docs (this tree)
└── Taskfile.yml
~~~

## Running

~~~sh
task init              # install local tools (golangci-lint, gotestsum, mockery)
task build             # go build -o easyp ./cmd/easyp
task test              # gotestsum with -race and coverage
task lint              # golangci-lint (+ hadolint where applicable)
task quality           # test + lint
task schema:generate   # regenerate schemas/*.json
task schema:check      # fail if schemas drift
task mocks             # optional mocks for actual core and console interfaces
task dev-tools:check   # isolated offline Taskfile regression
~~~

Without Task:

~~~sh
go build -o easyp ./cmd/easyp
go test -race -count=1 ./...
go run ./cmd/easyp schema-gen
~~~

<code>task lint</code> runs the local Go linter and Hadolint against root <code>Dockerfile</code>. See [TOOLS.md](./TOOLS.md) for pinned tool prerequisites and <code>task dev-tools:check</code>. Docker build helpers build locally and never publish.

## Ports

N/A — EasyP is a CLI tool, not a long-running server. Remote plugin execution uses gRPC to an external EasyP plugin API when configured; there is no default listen port in-repo.

## Key Interfaces / Entry Points

| Entry | Role |
|-------|------|
| <code>cmd/easyp</code> | Process entry; registers CLI handlers |
| <code>internal/api</code> | All registered command handlers; see [CLI.md](./CLI.md) for flags and aliases |
| <code>Core</code> in <code>internal/core/core.go</code> | Per-engine lint, breaking and low-level generation state |
| <code>Source</code>, <code>Cache</code>, <code>Repository</code>, <code>VersionedRepository</code> in <code>internal/modules</code> | Dependency resolution and verified installation contracts |
| <code>internal/migration</code> + <code>internal/api/migrate_interactive.go</code> | Preview/apply conversion and interactive migration wizard |
| <code>easyp.yaml</code> + <code>easyp.gen.yaml</code> + <code>protobuf.mod</code> + <code>protobuf.lock</code> | Policy, generation, declared dependencies and locked revisions |
| <code>mcp/easypconfig</code> | Config schema / MCP metadata |

## Adding New Features

1. **Lint rule** — add <code>internal/rules/&lt;rule&gt;.go</code> + <code>&lt;file&gt;_test.go</code>; register via existing rule builder; mirror neighboring rules.
2. **v1 schema** — change <code>internal/config/v1</code>, then <code>task schema:generate</code>; update MCP descriptions when semantics change. Never hand-edit <code>schemas/*.json</code>.
3. **CLI command** — wire in <code>internal/api</code>, register from <code>cmd/easyp</code>; follow existing handler patterns.
4. **Interfaces** — update consumer interfaces and their test doubles; consult [TESTING.md](./TESTING.md) before generating optional Mockery test doubles. Add/adjust tests with <code>-race</code>.
5. Before finishing behavior changes: run tests for touched packages; for schema changes run <code>task schema:check</code>.
6. Reserved extensions/package selectors/non-<code>FILE</code> breaking categories and unresolved X20 mappings are recorded in [CLI.md](./CLI.md) and [ERRORS.md](./ERRORS.md); do not infer support from a declared field alone.
