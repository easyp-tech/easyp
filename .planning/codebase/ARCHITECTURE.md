---
last_mapped_commit: 5c9185fd17b72af60e0b7c50d6d713d466722555
last_mapped_at: 2026-11-17
---
<!-- refreshed: 2026-11-17 -->
# Architecture

**Analysis Date:** 2026-11-17

## System Overview

EasyP is a Go CLI toolkit for Protocol Buffers: linting, breaking-change detection, code generation (local/remote/WASM plugins), and Git-based dependency management. Layered CLI architecture: a thin command layer (`internal/api`) composes pure-ish operation packages (`internal/modules`, `internal/generation`, `internal/migration`) on top of an engine layer (`internal/core`) and infrastructure adapters.

```text
┌───────────────────────────────────────────────────────────────────┐
│                         CLI entrypoints                           │
│  `cmd/easyp` (main CLI)          `cmd/easyp-mcp` (MCP server)     │
├───────────────────────────────────────────────────────────────────┤
│                       Command / wiring layer                      │
│  `internal/api` (flags, command structs, engine composition)      │
│  `internal/flags`, `internal/version`                             │
├──────────────────────┬─────────────────────┬──────────────────────┤
│  Module operations   │  Generation          │  Migration           │
│  `internal/modules`  │  `internal/generation`│ `internal/migration`│
│  (resolve, tidy,     │  (discovery,         │  (legacy→v1 plans,  │
│   get, update,       │   inheritance,       │   guarded apply)    │
│   vendor, locks)     │   selection, run)    │                      │
├──────────────────────┴─────────────────────┴──────────────────────┤
│                         Engine layer                               │
│  `internal/core` (parse/compile, lint, breaking, plugins,          │
│   managed mode, generate bucket)  +  `internal/rules` (lint rules) │
├───────────────────────────────────────────────────────────────────┤
│                     Config & shared models                         │
│  `internal/config`, `internal/config/v1` (parsers, JSON Schema),   │
│  `internal/policy` (extends resolver), `internal/workspace`,       │
│  `internal/protosource`, `internal/sourceview`, `internal/parallel`│
├───────────────────────────────────────────────────────────────────┤
│                        Infrastructure adapters                     │
│  `internal/adapters/*` (gitmodules, gitcommand, gitsnapshot,       │
│  go_git, module_config, modfile, plugin(+wasm), bsr, console,      │
│  prompter), `internal/fs/*`, `internal/logger`                     │
└───────────────────────────────────────────────────────────────────┘
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

**Overall:** Layered clean-ish CLI architecture with explicit dependency injection and contract interfaces.

**Key Characteristics:**
- `internal/modules` and `internal/generation` accept `context.Context` plus explicit inputs — no `*cli.Context` below the API layer.
- Small contracts (`Source`, `Cache`, `Repository`, `VersionedRepository` in `internal/modules/repository.go`) decouple operations from Git specifics.
- CLI supplies `EASYPPATH` and a cache instance; only the Git adapter knows the on-disk cache layout (`v1/git` under the EasyP dir).
- Concurrency is bounded: 8 parallel lock-entry installs, 4 parallel plugin invocations per target, one shared WASM compilation scope per generation run (`internal/parallel/parallel.go`).
- Unit tests use handwritten fakes; adapter/integration tests use local Git fixtures.

## Layers

**CLI layer (`internal/api`):**
- Purpose: flag parsing, adapter construction, command-local orchestration, output and exit status
- Location: `internal/api/`
- Depends on: modules, generation, migration, config/v1, rules, core, adapters
- Key files: `runtime.go` (`buildCore` constructs lint/breaking engines), `lint_v1.go`, `breaking_v1.go`, `generate.go`, `mod_v1.go`, `migrate.go`

**Operations layer:**
- `internal/modules`: `Resolve(ctx, module, source, pins)` owns version selection/traversal; `Cache.Install` handles installation; `TidyWithReport` sequences tidy phases; `writeV1ResolvedFiles` orders manifest/lock persistence (per-file temp replacement, NOT crash-atomic)
- `internal/generation`: `generation.Run(...)` is the single entry; `config.go` owns v1→engine translation; `discovery.go`/`inherit.go`/`selection.go` handle consumers
- `internal/migration`: `migration.Build` produces candidates; `Plan.Apply` rechecks state, stages files+backups, rolls back on ordinary failures

**Engine layer (`internal/core`):**
- Purpose: proto parsing/compilation, managed descriptors, lint/breaking engines, plugin execution
- `core.New` receives complete options (import roots, file-module mappings) — no post-construction setters
- Depends on: `internal/rules`, plugin executors, Git tree walker (`internal/adapters/go_git`)
- Breaking profiles FILE/PACKAGE/WIRE_JSON/WIRE implemented in `breaking_profiles.go`/`breaking_check.go`; omitted/empty categories default to FILE

**Adapter layer (`internal/adapters`):**
- Each adapter is bounded: `gitcommand` runs Git with deadlines/cancellation; `gitsnapshot` is an immutable tree/blob FS; `module_config` adapts native/Buf/legacy metadata into module contracts; `plugin` provides local/remote/WASM/command executors

## Data Flow

### Generation (primary path)

1. `cmd/easyp` dispatches `Generate` command (`cmd/easyp/main.go`)
2. `api.Generate.Action` collects flags, cwd, cache (`internal/api/generate.go`)
3. `generation.Run(ctx, logger, cache, Request)` discovers configs and inherits generation options (`internal/generation/generate.go`, `discovery.go`)
4. Source selection: sibling, workspace, replacement, or locked module (`internal/generation/selection.go`)
5. Roots/collision checks shared with policy commands (`internal/modules/walk.go`)
6. Translation to `core.Options` (`internal/generation/config.go`)
7. `core.New` + `PrepareGeneration` per target → compiled graph, managed mode, optional descriptor export (`internal/core/generate.go`)
8. `GenerationPlan.ExecuteInto` shared `GenerateBucket` → output written sorted (`internal/core/generate_bucket.go`)

### Module dependency operation (get/tidy/update)

1. Read manifest + relevant pins (`internal/modules/manifest_edit.go`)
2. `Resolve` through `Source` contract, MVS selection (`internal/modules/resolve.go`)
3. `Cache.Install` + cached metadata; roots/collision/unresolved-import checks
4. Compute manifest edits and lock; persist via temp-file replacement (`internal/modules/project_files.go`)

### Lint/Breaking

1. `api` selects producer policy, prepares shared import roots (`internal/api/roots.go`)
2. `buildCore` constructs only the needed engine (`internal/api/runtime.go`)
3. Comment directives applied by `internal/core/comment_suppressions.go`
4. Breaking comparisons use historical Git trees + baseline-specific roots (`internal/core/breaking_check.go`, `internal/adapters/go_git`)

**State Management:** No persistent in-process state; all state lives in project files (`easyp.yaml`, `easyp.gen.yaml`, `protobuf.mod`, `protobuf.lock`), the Git cache on disk, and per-operation in-memory structures (e.g., `graphResolveResult` returns copied slices/maps).

## Key Abstractions

**`core.Rule`:**
- Purpose: lint rule contract
- Examples: `internal/rules/comment_message.go`, all `internal/rules/*.go`
- Pattern: builder registration in `internal/rules/builder.go`; each rule colocated with `_test.go`

**`modules.Source` / `Cache` / `Repository` / `VersionedRepository`:**
- Purpose: version fetch, install/verify, Get/Tidy, Update contracts
- Location: `internal/modules/repository.go`
- Pattern: implemented by `internal/adapters/gitmodules`; fakes in tests

**`api.Handler`:**
- Purpose: one CLI command per struct implementing `Command() *cli.Command`
- Location: `internal/api/interface.go`

**v1 config model + schema:**
- Single source of truth in `internal/config/v1`; JSON Schema via `v1.SchemaJSON`, consumed by both `internal/schemagen` and `mcp/easypconfig`

## Entry Points

**`cmd/easyp` (`cmd/easyp/main.go`):**
- urfave/cli app; registers `api.Lint`, `api.Mod`, `api.Get`, `api.Completion`, `api.Init`, `api.Migrate`, `api.Generate`, `api.LsFiles`, `api.SchemaGen`, `api.Validate`, `api.Breaking`, etc.; slog-based logger to stderr

**`cmd/easyp-mcp`:**
- Separate MCP server entry point for the config reference

**`mcp/easypconfig`:**
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

**What happens:** Setting import roots or file-module mappings post-construction.
**Why it's wrong:** `core.New` copies roots/mappings at construction; follow-up mutation is not supported.
**Do this instead:** Pass complete options to `core.New` (see `internal/generation/config.go`).

### Building lint rules for generation runs

**What happens:** Constructing unrelated lint engines during generation.
**Why it's wrong:** Wasteful; generation inherits only generation options.
**Do this instead:** Only lint/breaking commands use `buildCore` (`internal/api/runtime.go`).

### Hand-editing generated schema JSON

**What happens:** Editing `schemas/*.json` directly.
**Why it's wrong:** Drifts from `internal/config/v1` source of truth; `task schema:check` fails.
**Do this instead:** Change types in `internal/config/v1`, then `task schema:generate`.

## Error Handling

**Strategy:** Idiomatic Go error returns with wrapping; multi-file writers return restoration errors together with the original error (`internal/modules/project_files.go`).

**Patterns:**
- Bounded Git phase diagnostics from `internal/adapters/gitcommand`
- Structured YAML validation issues from `internal/config/v1`
- Rollback on ordinary write failures in migration apply and resolved-files transaction (`internal/modules/resolved_files_transaction.go`)

## Cross-Cutting Concerns

**Logging:** `internal/logger` contract over slog; text handler to stderr; debug via global flag.
**Validation:** `config/v1.ValidatePath` for recursive structured YAML validation; semantic checks shared in `internal/config/v1/policy_semantics.go`.
**Authentication:** None (Git over HTTPS/local; remote plugins dialed through a passthrough).

---

*Architecture analysis: 2026-11-17*
