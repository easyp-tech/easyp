<!-- generated: 2026-07-27, template: bootstrap.md -->
# Agent Rules — EasyP

Mandatory rules for AI agents. Prefer these over generic Go guides when they conflict. Expanded skills: `.agents/skills/go-code-style`, `.agents/skills/go-testing`. Dependency details: [config/dependency.md](./config/dependency.md).

## Code Style

- Wrap errors as `fmt.Errorf("<called_function>: %w", err)` — function/method name only, **no package prefix**.
- Do not assign inside `if` conditions except `:=` for block-scoped vars; assign then check on separate lines.
- Never bare `defer x.Close()` — always wrap Close and log/handle the error.
- No inline comments on `if` / `for` / `return` lines; comments go above if needed.
- Imports: stdlib → third-party → project (`gci`); comments and godoc in English; exported symbols need godoc starting with the name.
- Interfaces: context first, error last, no `I` prefix; prefer defining interfaces in the consumer package.
- Enums: reserve zero with `_ = iota` so unset values are never silently valid.

## Naming Conventions

- Files: colocated tests as `<file>_test.go`; lint rules as `internal/rules/<rule>.go`.
- Types: exported domain/entities and adapter implementations; keep config structs unexported when local.
- Struct tags: `snake_case` (`yaml`, env) — follow `tagliatelle` / existing config patterns.
- Do not invent parallel agent instruction files (`.cursor/rules/`, `CLAUDE.md`); root `AGENTS.md` + `.spec/` are the source of truth.

## Error Handling

- Wrap at every call site with `%w`; use `errors.Is` / `errors.As` for sentinels.
- Domain/package-manager sentinels live under `internal/core` / `internal/core/models` (e.g. `ErrVersionNotFound`); do not invent duplicate sentinels elsewhere.
- CLI: map known failures to documented exit codes (e.g. `ErrVersionNotFound` → exit 1 for `mod`).

## Testing

- Use `testify` (`require` for fatal preconditions, `assert` for non-fatal checks).
- Table-driven tests with a required `name` field; call `t.Parallel()` at top level and inside each `t.Run`.
- No shared mutable mocks/state across parallel sub-tests — construct deps per case.
- Prefer `require.ErrorIs` for expected errors; run with `-race` (`task test`).
- When core/storage interfaces change, regenerate mocks with `task mocks`.

## Dependencies

- Proto deps: declare in `protobuf.mod` (and/or `generate.inputs[].git_repo.url`); lock with `protobuf.lock`; cache under `EASYPPATH` (default `~/.easyp`).
- Vendor directory is `easyp_vendor`, not `vendor/`.
- Use `easyp mod download` (lock-first) vs `easyp mod update` (refresh from `protobuf.mod`); commit `protobuf.lock` for CI reproducibility.
- No remotes/mirrors/auth fields in `easyp.yaml` — auth via system git.
- Do not hand-edit `schemas/easyp-config*.schema.json`; regenerate via `task schema:generate`.

## Formatting

- Format with `gofmt` / project tooling; lint with `task lint` (golangci-lint + hadolint where applicable).
- Do not commit secrets, `docs/node_modules`, `coverage.out`, or the local `easyp` binary in the repo root.
- After behavior changes, run tests for touched packages; after schema changes, run `task schema:check`.

## Quick Checklist

- [ ] Error wrap = callee name only; no bare `defer Close()`
- [ ] Tests: named cases, `t.Parallel`, no shared mutable mocks
- [ ] Schema JSON regenerated, not hand-edited; vendor = `easyp_vendor`
