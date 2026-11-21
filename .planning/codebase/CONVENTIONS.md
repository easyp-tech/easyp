---
last_mapped_commit: 5c9185fd17b72af60e0b7c50d6d713d466722555
last_mapped_at: 2026-11-17
---
# Coding Conventions

**Analysis Date:** 2026-11-17

Authoritative sources: `AGENTS.md`, `.spec/agent-rules.md`, `.spec/CODE_STYLE.md`, project skills `.agents/skills/go-code-style/SKILL.md` and `.agents/skills/go-testing/SKILL.md`. These project rules override generic Go guides when they conflict.

## Naming Patterns

**Files:**
- Lowercase, underscore-separated Go file names (`rpc_pascal_case.go`, `package_defined.go`)
- Tests colocated as `<file>_test.go` next to the implementation
- Lint rules: `internal/rules/<rule>.go` plus `<rule>_test.go`

**Directories:**
- Short lower-case package names; adapter subpackages named by capability (`internal/adapters/gitmodules`, `internal/adapters/plugin`)
- No parallel agent instruction files (`.cursor/rules/`, `CLAUDE.md`); root `AGENTS.md` + `.spec/` are the source of truth

**Functions/Types:**
- Interfaces: descriptive names, no `I` prefix; prefer defining them in the consumer package (`internal/core` owns `Rule`, `FS`, `DirWalker`; `internal/modules` owns `Source`, `Cache`, `Repository`; `internal/adapters/plugin/interface.go` owns `Executor`)
- Enums: reserve zero value with `_ = iota` so unset values are never silently valid
- Rule names derived via `core.GetRuleName` (concrete rule type name → upper snake case)
- Keep config structs unexported when local; export domain entities and adapter implementations

**Struct tags:**
- Native YAML structs use `yaml` tags; shared legacy types may also carry `json` tags
- `v1.Module` has no YAML tags (`protobuf.mod` uses a directive parser)
- Preserve established public spellings: snake_case (`with_imports`) and hyphenated keys (`linters-settings`, `exclude-rules`); never rename established keys

## Code Style

**Formatting:**
- `gofmt` via project tooling (`task lint`)
- Exported symbols need godoc comments in English beginning with the symbol name

**Linting:**
- golangci-lint (v2 config) via `task lint:go` — `.golangci.yml` enables `staticcheck` only; it does NOT configure `gci`, so import grouping is convention, not tool-enforced
- `task lint:docker` runs Hadolint on root `Dockerfile` (requires Docker)
- ST1005 exclusion: identifier-style error wrappers `fmt.Errorf("Open: %w", err)` are exempt; capitalized prose errors are not

**Key style rules (mandatory per `.spec/agent-rules.md`):**
- No assignment inside `if` conditions except block-scoped `:=`; assign then check on separate lines
- No inline comments on `if`/`for`/`return` lines; comments go above
- Prefer `any` over `interface{}` (staticcheck default)

## Import Organization

**Order:**
1. Standard library
2. Third-party packages
3. Project packages (`github.com/easyp-tech/easyp/...`)

Groups separated by blank lines. Run `gofmt`.

**Path Aliases:** None; full module paths from `github.com/easyp-tech/easyp`.

## Error Handling

**Patterns (strictly enforced):**
- Wrap every propagated failure as `fmt.Errorf("<called_function>: %w", err)` — callee function/method name ONLY, no package or receiver prefix (`"Open: %w"`, `"Fetch: %w"`, `"lintProtoInfoRead: %w"`)
- Use `errors.Is` for sentinels, `errors.As` for typed errors; preserve `%w` chains
- Reuse existing errors in their owning packages — do not invent duplicate sentinels:
  - `internal/core/core.go`: `ErrInvalidRule`, `ErrRepositoryDoesNotExist`, `ErrEmptyInputFiles`
  - `internal/core/dom.go`: `OpenImportFileError`, `GitRefNotFoundError` (typed)
  - `internal/modules/immutable_versions.go`: `ErrLockedVersionChanged`
  - `internal/config/v1/legacy_detection.go`: `ErrLegacyConfiguration`
- No central domain-to-gRPC-status/exit-code mapper; read the actual handler and `cmd/easyp/main.go` for exit behavior
- Log failures where handled/terminal; lower layers wrap and return

## Resource Closing

Never bare `defer x.Close()`:
```go
defer func() {
    _ = file.Close()
}()
```
When close failure matters operationally, log path+error via the injected logger (see `internal/core` `Core.close`).

## Logging

**Framework:** `log/slog` text handler, stored in `cli.App.Metadata`; handlers retrieve via `getLogger`; core and adapters receive the logger as a dependency.

**Patterns:**
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

**Exports:** Exported types need godoc; unexported when scope is local.

**Barrel Files:** Not used.

**Generated artifacts:** Never hand-edit `schemas/*.schema.json` (6 files); change `internal/config/v1` then `task schema:generate` (verify with `task schema:check`). Keep MCP field descriptions in `mcp/easypconfig` aligned.

**Dependencies (domain):** vendor dir is `easyp_vendor`, not `vendor/`; dependencies come from `protobuf.mod` `require` directives, locked via `protobuf.lock`, cached under `EASYPPATH` (default `~/.easyp`); no remotes/mirrors/auth in `easyp.yaml` (auth via system git).

---

*Convention analysis: 2026-11-17*
