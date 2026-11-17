<!-- GSD:project-start source:PROJECT.md -->

## Project

**EasyP**

EasyP — CLI-тулкит для Protocol Buffers на Go: линтинг (buf-совместимые правила), детекция breaking changes, генерация кода (локальные/remote/WASM-плагины) и Git-центричное управление зависимостями (`easyp mod`). Пользователи — разработчики proto-API и команды, держащие большие корпуса proto-файлов в CI. Этот GSD-проект — рамка для развития EasyP в целом; milestones добавляются по мере подготовки фич в `.spec/features/`.

**Core Value:** EasyP остаётся drop-in buf-совместимым инструментом: формат вывода, коды выхода и конфигурация не ломаются между версиями — миграции внутренностей не должны быть заметны пользователю.

### Constraints

- **Tech stack**: Go 1.26.6, urfave/cli v2, testify; новые зависимости для компилятора запрещены (REQ-1.2) — вендоринг в `internal/`
- **Compatibility**: формат вывода issues, пути и коды выхода CLI неизменны; buf-совместимое поведение — контракт
- **Scope**: breaking/generate/`mod tidy`/`migrate` не меняются в этом milestone (REQ-5.2)
- **Verification**: `task test` / `task lint` / `task schema:check`; паритет корпусов через `compiler-lab/bench/lab.py` — требует доступа к приватному репозиторию compiler-lab
- **Process**: репозиторий живёт по AGENTS.md — colocated тесты правил, регенерация схем через `task schema:generate`, без ручного редактирования `schemas/`

<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->

## Technology Stack

## Languages

- Go 1.26.6 (`go.mod` declares `go 1.26.6`) - entire CLI toolkit, all `cmd/` and `internal/` packages
- Protocol Buffers - embedded proto files and generated plugin descriptors (e.g. `easyp/generator/v1` API types from `github.com/easyp-tech/service/api`)
- YAML - user-facing config: `easyp.yaml` (lint/breaking policy), `easyp.gen.yaml` (generation), `protobuf.mod` / `protobuf.lock`
- JSON Schema - generated config schemas in `schemas/` (do not hand-edit)
- Dockerfile - root `Dockerfile` for the CLI image
- Shell - `Taskfile.yml` tasks and `scripts/` helpers (e.g. `scripts/check-proto-example.sh`, `scripts/validate-release-tag.sh`)

## Runtime

- Go toolchain 1.26.6 (CI uses `go-version-file: go.mod`)
- Docker required for `task lint:docker` (Hadolint) and `task docker:build`
- Git required for module operations (EasyP's dependency model is Git-based)
- Go modules (`go.mod` / `go.sum`); no lockfile tooling beyond go.sum

## Frameworks

- CLI framework: `github.com/urfave/cli/v2` v2.27.7 - command wiring in `cmd/easyp` / `internal/api`
- gRPC: `google.golang.org/grpc` v1.83.2 - remote plugin execution (`internal/adapters/plugin/remote.go`)
- Protobuf runtime: `google.golang.org/protobuf` v1.36.11
- Proto compilation: `github.com/bufbuild/protocompile` v0.14.1 (`internal/core`)
- MCP server SDK: `github.com/modelcontextprotocol/go-sdk` v1.6.1 (`cmd/easyp-mcp`, `mcp/easypconfig`)
- `github.com/stretchr/testify` v1.11.1 - all `*_test.go` files
- `gotestsum` v1.13.0 (pinned in `bin/`) - test runner with `-race -count=1` and coverage
- `mockery` v2.53.7 (pinned, optional) - mocks for `core.Rule`, `core.CurrentProjectGitWalker`, `console.Console`
- Task (Taskfile.yml, v3 in CI) - all dev workflows; prefer `task` targets over raw commands
- `golangci-lint` v2.14.0 (pinned in `bin/`, config in `.golangci.yml`)
- Hadolint v2.12.1-beta (Docker image, not local install)
- Docker Buildx + QEMU for multi-arch release builds (`.github/workflows/release.yml`)

## Key Dependencies

- `github.com/bufbuild/protocompile` v0.14.1 - proto parsing/compilation, the heart of `internal/core`
- `github.com/yoheimuta/go-protoparser/v4` v4.14.2 - proto source parsing for lint rules and comment suppressions
- `github.com/go-git/go-git/v5` v5.16.3 - in-process Git operations for module cache/snapshots (`internal/adapters/go_git`, `internal/adapters/gitmodules`)
- `github.com/tetratelabs/wazero` v1.9.0 (+ `github.com/wasilibs/wazero-helpers`) - WASM runtime for builtin plugins (`internal/adapters/plugin/builtin_impl.go`, `builtin_run.go`)
- `github.com/easyp-tech/service/api` v1.0.0 - remote generator gRPC API (`easyp/generator/v1`)
- `github.com/urfave/cli/v2` v2.27.7 - CLI framework
- `github.com/a8m/envsubst` v1.4.3 - env substitution in config files
- `github.com/Yakwilik/go-yamlvalidator` v0.3.1 + `gopkg.in/yaml.v3` - YAML config parsing/validation
- `github.com/samber/lo` v1.52.0 - generic utilities
- `github.com/charmbracelet/x/term` v0.2.1 - terminal interaction (interactive migrate wizard, prompts)
- `golang.org/x/mod` v0.38.0 - semver handling
- `github.com/santhosh-tekuri/jsonschema/v6` (indirect, via schema tooling) and `github.com/google/jsonschema-go` v0.4.3 - JSON Schema generation in `internal/schemagen`
- Embedded plugin protos (indirect): `github.com/easyp-tech/protoc-gen-easydoc` v0.4.0, `github.com/easyp-tech/protoc-gen-mcp` v0.5.0

## Configuration

- `EASYP_PROXY` - proxy host for Git operations (validated host-only, see `internal/config/proxy.go`)
- `EASYP_GIT_TIMEOUT` - per-operation Git timeout duration (see `internal/adapters/gitcommand/command.go`)
- `GIT_TERMINAL_PROMPT=0` is forced by the tool itself for non-interactive Git
- No `.env` file mechanism; configuration comes from CLI flags, `easyp.yaml`, `easyp.gen.yaml`, `protobuf.mod` and `protobuf.lock`
- `Taskfile.yml` - all task definitions, pinned tool versions
- `.golangci.yml` - golangci-lint v2 config (staticcheck enabled; ST1005 exemption for identifier-style error wrappers per `.spec/agent-rules.md`)
- `Dockerfile` - root CLI image
- `schemas/*.json` - generated via `go run ./cmd/easyp schema-gen --out-dir schemas`; never hand-edit

## Platform Requirements

- Go 1.26.6, Task 3.x, Docker (for lint/docker tasks), Git
- `task init` installs pinned helpers into repo-local `bin/`
- Distributed as a Go binary (`easyp`) plus Docker image; released via GitHub Actions on `v1.0.0-nightly.*` tags (multi-arch), separate MCP binary `cmd/easyp-mcp`
- Docs site lives in the separate `easyp-tech/docs` repository

<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->

## Conventions

## Naming Patterns

- Lowercase, underscore-separated Go file names (`rpc_pascal_case.go`, `package_defined.go`)
- Tests colocated as `<file>_test.go` next to the implementation
- Lint rules: `internal/rules/<rule>.go` plus `<rule>_test.go`
- Short lower-case package names; adapter subpackages named by capability (`internal/adapters/gitmodules`, `internal/adapters/plugin`)
- No parallel agent instruction files (`.cursor/rules/`, `CLAUDE.md`); root `AGENTS.md` + `.spec/` are the source of truth
- Interfaces: descriptive names, no `I` prefix; prefer defining them in the consumer package (`internal/core` owns `Rule`, `FS`, `DirWalker`; `internal/modules` owns `Source`, `Cache`, `Repository`; `internal/adapters/plugin/interface.go` owns `Executor`)
- Enums: reserve zero value with `_ = iota` so unset values are never silently valid
- Rule names derived via `core.GetRuleName` (concrete rule type name → upper snake case)
- Keep config structs unexported when local; export domain entities and adapter implementations
- Native YAML structs use `yaml` tags; shared legacy types may also carry `json` tags
- `v1.Module` has no YAML tags (`protobuf.mod` uses a directive parser)
- Preserve established public spellings: snake_case (`with_imports`) and hyphenated keys (`linters-settings`, `exclude-rules`); never rename established keys

## Code Style

- `gofmt` via project tooling (`task lint`)
- Exported symbols need godoc comments in English beginning with the symbol name
- golangci-lint (v2 config) via `task lint:go` — `.golangci.yml` enables `staticcheck` only; it does NOT configure `gci`, so import grouping is convention, not tool-enforced
- `task lint:docker` runs Hadolint on root `Dockerfile` (requires Docker)
- ST1005 exclusion: identifier-style error wrappers `fmt.Errorf("Open: %w", err)` are exempt; capitalized prose errors are not
- No assignment inside `if` conditions except block-scoped `:=`; assign then check on separate lines
- No inline comments on `if`/`for`/`return` lines; comments go above
- Prefer `any` over `interface{}` (staticcheck default)

## Import Organization

## Error Handling

- Wrap every propagated failure as `fmt.Errorf("<called_function>: %w", err)` — callee function/method name ONLY, no package or receiver prefix (`"Open: %w"`, `"Fetch: %w"`, `"protoInfoRead: %w"`)
- Use `errors.Is` for sentinels, `errors.As` for typed errors; preserve `%w` chains
- Reuse existing errors in their owning packages — do not invent duplicate sentinels:
- No central domain-to-gRPC-status/exit-code mapper; read the actual handler and `cmd/easyp/main.go` for exit behavior
- Log failures where handled/terminal; lower layers wrap and return

## Resource Closing

## Logging

- Structured attributes: `slog.String`, `slog.Int`, `slog.Any`
- CLI/API: log terminal failures before `os.Exit`
- Core: operation start/completion, debug resolution data, recoverable warnings
- No project-wide correlation-ID convention

## Comments

- English only; godoc begins with the symbol name
- Explanatory comments above `if`/`for`/`return`, never inline

## Function/Interface Design

- `context.Context` first in every signature; `error` last
- Methods check context cancellation during dependency resolution, filesystem walks, compilation, executor calls
- Prefer small consumer-owned interfaces over concrete adapters
- Conversion boundaries live in the owning workflow: `buildCore` in `internal/api/runtime.go` (producer lint/breaking → engine options); `prepareV1GeneratorConfig` in `internal/generation/config.go` (native plugins/roots → `core.Options`)
- Core engine structs carry no YAML tags; pass complete options to `core.New`

## Module Design

<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->

## Architecture

## System Overview

```text

```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| CLI app | Global flags, logger, command registration | `cmd/easyp/main.go` |
| Command wiring | Each `api.Handler` struct parses flags and orchestrates one command | `internal/api/interface.go` + per-command files |
| Module ops | Dependency selection (`Resolve`), lock validation, roots/ownership, manifest edits, vendor, tidy | `internal/modules/resolve.go`, `internal/modules/operations.go` |
| Generation | Consumer discovery, option inheritance, module selection, translation to `core.Options`, run | `internal/generation/generate.go`, `internal/generation/config.go` |
| Migration | Legacy conversion, integrity verification, backups, rollback | `internal/migration/migration.go`, `internal/migration/apply.go` |
| Core engine | Proto compile, lint engine, breaking engine, plugin execution, managed mode, descriptor output | `internal/core/core.go`, `internal/core/generate.go`, `internal/core/lint.go` |
| Lint rules | 88 files: each rule implements `core.Rule`, registered via builder | `internal/rules/builder.go` + `<rule>.go` |
| Git module cache | Verified checkouts, object cache with OS locks, snapshot hashing, install | `internal/adapters/gitmodules/` |
| Config v1 | `protobuf.mod` parser, lock/policy/generator YAML, JSON Schema | `internal/config/v1/module.go`, `lock.go`, `policy.go`, `generate.go`, `schema.go` |
| Schema generation | Writes versioned/latest JSON Schema artifacts | `internal/schemagen/schemagen.go` |
| MCP reference | Config-description tool exposing v1 schemas | `mcp/easypconfig/` |
| Workspace | Repository boundary, ancestor config/module discovery | `internal/workspace/discovery.go` |

## Pattern Overview

- `internal/modules` and `internal/generation` accept `context.Context` plus explicit inputs — no `*cli.Context` below the API layer.
- Small contracts (`Source`, `Cache`, `Repository`, `VersionedRepository` in `internal/modules/repository.go`) decouple operations from Git specifics.
- CLI supplies `EASYPPATH` and a cache instance; only the Git adapter knows the on-disk cache layout (`v1/git` under the EasyP dir).
- Concurrency is bounded: 8 parallel lock-entry installs, 4 parallel plugin invocations per target, one shared WASM compilation scope per generation run (`internal/parallel/parallel.go`).
- Unit tests use handwritten fakes; adapter/integration tests use local Git fixtures.

## Layers

- Purpose: flag parsing, adapter construction, command-local orchestration, output and exit status
- Location: `internal/api/`
- Depends on: modules, generation, migration, config/v1, rules, core, adapters
- Key files: `runtime.go` (`buildCore` constructs lint/breaking engines), `lint_v1.go`, `breaking_v1.go`, `generate.go`, `mod_v1.go`, `migrate.go`
- `internal/modules`: `Resolve(ctx, module, source, pins)` owns version selection/traversal; `Cache.Install` handles installation; `TidyWithReport` sequences tidy phases; `writeV1ResolvedFiles` orders manifest/lock persistence (per-file temp replacement, NOT crash-atomic)
- `internal/generation`: `generation.Run(...)` is the single entry; `config.go` owns v1→engine translation; `discovery.go`/`inherit.go`/`selection.go` handle consumers
- `internal/migration`: `migration.Build` produces candidates; `Plan.Apply` rechecks state, stages files+backups, rolls back on ordinary failures
- Purpose: proto parsing/compilation, managed descriptors, lint/breaking engines, plugin execution
- `core.New` receives complete options (import roots, file-module mappings) — no post-construction setters
- Depends on: `internal/rules`, plugin executors, Git tree walker (`internal/adapters/go_git`)
- Breaking profiles FILE/PACKAGE/WIRE_JSON/WIRE implemented in `breaking_profiles.go`/`breaking_check.go`; omitted/empty categories default to FILE
- Each adapter is bounded: `gitcommand` runs Git with deadlines/cancellation; `gitsnapshot` is an immutable tree/blob FS; `module_config` adapts native/Buf/legacy metadata into module contracts; `plugin` provides local/remote/WASM/command executors

## Data Flow

### Generation (primary path)

### Module dependency operation (get/tidy/update)

### Lint/Breaking

## Key Abstractions

- Purpose: lint rule contract
- Examples: `internal/rules/comment_message.go`, all `internal/rules/*.go`
- Pattern: builder registration in `internal/rules/builder.go`; each rule colocated with `_test.go`
- Purpose: version fetch, install/verify, Get/Tidy, Update contracts
- Location: `internal/modules/repository.go`
- Pattern: implemented by `internal/adapters/gitmodules`; fakes in tests
- Purpose: one CLI command per struct implementing `Command() *cli.Command`
- Location: `internal/api/interface.go`
- Single source of truth in `internal/config/v1`; JSON Schema via `v1.SchemaJSON`, consumed by both `internal/schemagen` and `mcp/easypconfig`

## Entry Points

- urfave/cli app; registers `api.Lint`, `api.Mod`, `api.Get`, `api.Completion`, `api.Init`, `api.Migrate`, `api.Generate`, `api.LsFiles`, `api.SchemaGen`, `api.Validate`, `api.Breaking`, etc.; slog-based logger to stderr
- Separate MCP server entry point for the config reference
- MCP tool `easyp_config_describe`; read-only, never reads project files

## Architectural Constraints

- **Dependency direction:** `internal/api` may import everything; `modules`/`generation`/adapters never import the CLI package. Core does not own dependency downloads or lock updates.
- **Threading:** Bounded worker pools (8 installs, 4 plugin invocations per target); `internal/parallel` provides bounded atomic-index scheduling; WASM scope closing cancels and joins active work.
- **Persistence guarantees:** Project files are replaced via temp files individually; manifest/lock pair is NOT a filesystem transaction or crash-atomic. Migration apply is similarly guarded-but-not-atomic.
- **Frozen mode:** Explicit frozen operation verifies only the published graph; unknown-import-to-module discovery intentionally excluded (see `.spec/ERRORS.md`).
- **Local replacements:** Ephemeral overlay graph, never persisted to the shared lock.
- **No circular imports** between layers; rule direction enforced by convention (API→operations→core→adapters).

## Anti-Patterns

### Calling `core.New` setters after construction

### Building lint rules for generation runs

### Hand-editing generated schema JSON

## Error Handling

- Bounded Git phase diagnostics from `internal/adapters/gitcommand`
- Structured YAML validation issues from `internal/config/v1`
- Rollback on ordinary write failures in migration apply and resolved-files transaction (`internal/modules/resolved_files_transaction.go`)

## Cross-Cutting Concerns

<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->

## Project Skills

| Skill | Description | Path |
|-------|-------------|------|
| epctl-commands | "Use when adding or changing EasyP CLI commands, handlers in internal/api, or registration in cmd/easyp/main.go. epctl-commands is the retained legacy skill name." | `.agents/skills/epctl-commands/SKILL.md` |
| go-code-style | "Use when writing, reviewing, or refactoring Go code or creating packages in the github.com/easyp-tech/easyp CLI repository." | `.agents/skills/go-code-style/SKILL.md` |
| go-testing | "Use when writing, reviewing, or debugging Go tests in the github.com/easyp-tech/easyp CLI repository, including table tests, test doubles, and CLI state isolation." | `.agents/skills/go-testing/SKILL.md` |
| protobuf-expert-skill | "Protocol Buffers expert with deep EasyP CLI v1 knowledge, including migration from EasyP v0. Use when: writing or reviewing .proto files; configuring easyp.yaml, easyp.gen.yaml, protobuf.mod or protobuf.lock; choosing lint rules or breaking profiles; setting up code generation plugins; managing proto dependencies (easyp get, mod tidy); detecting breaking API changes; debugging easyp errors; upgrading or migrating an EasyP v0 project (easyp.lock, deps:, lint.use, generate.inputs) to v1 with easyp migrate, or when easyp reports 'legacy EasyP configuration detected'; migrating from buf; protobuf style and API design." | `.agents/skills/protobuf-expert-skill/SKILL.md` |
| protoc-gen-mcp-skill | "Build MCP servers from protobuf definitions using protoc-gen-mcp and easyp. Use when: creating an MCP server, generating MCP tools from proto files, building a proto-first MCP server in Go, Python, Java, Kotlin, TypeScript, or JavaScript, configuring easyp (v1: protobuf.mod, easyp.yaml, easyp.gen.yaml) for MCP generation, migrating an existing easyp v0 setup, adding MCP tool annotations to protobuf services, implementing MCP tool handlers, setting up ProtoJSON-based MCP tools, or any task involving protobuf-to-MCP code generation. Also use when the user mentions protoc-gen-mcp, mcp proto, proto mcp server, easyp mcp, or wants type-safe MCP bindings from .proto files." | `.agents/skills/protoc-gen-mcp-skill/SKILL.md` |
| sdd | > Spec-driven development pipeline with 6 phases: Explore, Requirements, Design, Task Plan, Implementation, Review. Enforces human approval gates between phases. Also provides a standalone documentation workflow for generating or updating project docs without starting a feature pipeline. Use when user wants structured feature development, spec-first approach, or says "I want to add feature X", "new feature", "implement", "build", "generate documentation", "update docs", "actualize the documentation". Keywords: spec, requirements, design document, TDD plan, task plan, implementation, code review, pipeline, approval gates, WHEN/SHALL, generate docs, update docs, documentation queue. | `.agents/skills/sdd/SKILL.md` |
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->

## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-fast` for a trivial task inline, with no subagents and no PLAN.md
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->

## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
