---
last_mapped_commit: 5c9185fd17b72af60e0b7c50d6d713d466722555
last_mapped_at: 2026-11-17
---
# Codebase Structure

**Analysis Date:** 2026-11-17

## Directory Layout

```
easyp/
├── cmd/
│   ├── easyp/            # Main CLI entrypoint (main.go)
│   └── easyp-mcp/        # Separate MCP server entrypoint
├── internal/
│   ├── api/              # CLI command wiring (urfave/cli Handler structs)
│   ├── modules/          # Dependency resolution, manifest/lock ops, import roots, vendor
│   ├── generation/       # Consumer discovery, inheritance, selection, descriptor export
│   ├── migration/        # Legacy→v1 migration plans and guarded apply
│   ├── core/             # Engines: parse/compile, lint, breaking, generate, managed mode
│   │   └── path_helpers/ # Target path matching, v1 source exclusions
│   ├── rules/            # Lint rules, one file per rule + colocated tests
│   ├── config/           # Shared engine/legacy config types
│   │   └── v1/           # v1 parsers (protobuf.mod, YAML), validation, JSON Schema
│   ├── adapters/         # Infrastructure: gitmodules, gitcommand, gitsnapshot, go_git,
│   │                     # module_config, modfile, plugin(+wasm), bsr, console, prompter
│   ├── policy/           # Section-scoped `extends` resolver
│   ├── workspace/        # Repository boundary and ancestor discovery
│   ├── parallel/         # Bounded atomic-index scheduling
│   ├── protosource/      # Proto package abstractions
│   ├── sourceview/       # Bounded logical fs view over io/fs
│   ├── schemagen/        # Writes versioned JSON Schema artifacts
│   ├── fs/               # fs/fs (walkers, atomic file replace), fs/go_git (Git tree FS)
│   ├── flags/            # Shared CLI flags
│   ├── logger/           # Logger contract
│   └── version/          # Build metadata
├── mcp/easypconfig/      # MCP config-description tool (schema exposure)
├── schemas/              # GENERATED JSON Schemas (do not hand-edit)
├── examples/proto/       # Sample proto files
├── .spec/                # AI-consumable project docs (ARCHITECTURE, PACKAGES, CLI, ...)
└── Taskfile.yml          # Task targets: init/build/test/lint/quality/schema:*
```

## Directory Purposes

**`internal/api`:**
- Purpose: One file per CLI command; flags → operations → output/exit
- Key files: `interface.go` (Handler), `runtime.go` (`buildCore`), `lint_v1.go`, `breaking_v1.go`, `generate.go`, `mod_v1.go`, `get_v1.go`, `migrate.go`, `migrate_interactive.go`, `ls_files_v1.go`, `init_v1.go`, `schema_gen.go`, `validate.go`, `policy_v1.go`

**`internal/modules`:**
- Purpose: All dependency management logic; largest package
- Key files: `resolve.go`, `repository.go`, `operations.go`, `get.go`, `update.go`, `vendor.go`, `tidy_operation.go`, `manifest_edit.go`, `project_files.go`, `import_roots_*.go` (resolution/result/overlay/tidy), `import_rewrites*.go`, `resolved_file_observations.go`, `resolved_files_transaction.go`, `bsr.go`, `effective_graph.go`, `frozen.go`

**`internal/generation`:**
- Key files: `generate.go`, `config.go` (v1→engine translation), `discovery.go`, `inherit.go`, `selection.go`, `module.go`, `descriptor_set.go`, `descriptor_output.go`, `frozen.go`, `workspace_index.go`

**`internal/core`:**
- Key files: `core.go`, `dom.go`, `lint.go`, `generate.go`, `generate_bucket.go`, `generate_plugins.go`, `generate_inputs.go`, `managed_mode.go`, `breaking_checker.go`, `breaking_check.go`, `breaking_profiles.go`, `comment_suppressions.go`

**`internal/rules`:**
- Purpose: 88 files; lint rules named after the buf-compatible rule (snake_case)
- Key files: `builder.go` (registration), e.g. `comment_message.go` + `comment_message_test.go`

**`internal/adapters/gitmodules`:**
- Key files: `cache.go`, `git.go`, `object_cache.go` (+ `object_lock_unix.go`/`object_lock_windows.go`), `checkout.go`, `git_source.go`, `download.go`, `files.go`, `migration.go`

**`internal/config/v1`:**
- Key files: `module.go` (`protobuf.mod` parser), `lock.go`, `policy.go`, `generate.go`, `schema.go` (`SchemaJSON`), `policy_semantics.go`, `semantic_validation.go`

## Key File Locations

**Entry Points:**
- `cmd/easyp/main.go`: main CLI app and command registration
- `cmd/easyp-mcp/main.go`: MCP server

**Configuration:**
- `Taskfile.yml`: build/test/lint/schema targets
- `.golangci.yml` (lint config, if present), `go.mod`

**Core Logic:**
- `internal/modules/resolve.go`: dependency graph resolution
- `internal/generation/generate.go`: generation entrypoint
- `internal/core/core.go`: engine construction

**Testing:**
- Colocated `<file>_test.go` next to every source file (testify)
- Fixtures: `internal/modules/fixtures_test.go`, `internal/generation/fixtures_test.go`, `internal/api/v1_fixtures_test.go`

**Generated artifacts:**
- `schemas/easyp-v1.schema.json`, `schemas/easyp.schema.json`, `schemas/easyp.gen-v1.schema.json`, `schemas/easyp.gen.schema.json`, `schemas/protobuf.lock-v1.schema.json`, `schemas/protobuf.lock.schema.json`

## Naming Conventions

**Files:**
- `snake_case.go` matching the single concern they own (e.g., `import_roots_resolution.go`, `breaking_checker.go`)
- Tests: `<file>_test.go` colocated; test-only helpers suffixed `fixtures_test.go`
- v1-related commands/files carry a `_v1` suffix in `internal/api` (e.g., `mod_v1.go`, `lint_v1.go`)

**Directories:**
- `internal/<domain>` singular concept; adapters grouped under `internal/adapters/<infra>`

**Packages:**
- `internal/api`, `internal.modules`, etc. — Go package name equals last path segment; imports use `github.com/easyp-tech/easyp/internal/...`

## Where to Add New Code

**New lint rule:**
- Implementation: `internal/rules/<rule>.go` + `internal/rules/<rule>_test.go`
- Register in the existing rule builder (`internal/rules/builder.go`); mirror a neighboring rule

**New CLI command:**
- Create an `api.Handler` struct in `internal/api/<command>_v1.go` (with `Command()` and `Action`)
- Register it in `cmd/easyp/main.go` `buildCommand(...)`
- Put reusable logic in `internal/modules` or `internal/generation`, not in the API layer

**New config field / schema change:**
- Types + rules: `internal/config/v1/`
- Regenerate: `task schema:generate`, verify with `task schema:check`
- Update MCP field descriptions in `mcp/easypconfig` when behavior changes

**New module operation:**
- `internal/modules/<operation>.go`, exposed via a contract in `repository.go` if it needs Source/Cache

**New plugin executor:**
- `internal/adapters/plugin/` (WASM variants in `internal/adapters/plugin/wasm/`)

**Utilities:**
- Filesystem helpers: `internal/fs/fs`
- Bounded concurrency: `internal/parallel`

**Tests:**
- Always colocated `<file>_test.go` in the same package, using testify

## Special Directories

**`schemas/`:**
- Purpose: Versioned and latest JSON Schema artifacts
- Generated: Yes (`task schema:generate` / `easyp schema-gen`)
- Committed: Yes — never hand-edit; `task schema:check` enforces no drift

**`.spec/`:**
- Purpose: AI-optimized project docs (ARCHITECTURE.md, PACKAGES.md, CLI.md, ERRORS.md, config guides)
- Generated: partially (templates dated), maintained deliberately
- Committed: Yes — start at `.spec/README.md`

**`bin/` (via `task init`):**
- Purpose: Pinned local tools (golangci-lint, gotestsum, mockery)
- Generated: Yes — not committed

**`.planning/codebase/`:**
- Purpose: This codebase map consumed by GSD planning commands

---

*Structure analysis: 2026-11-17*
