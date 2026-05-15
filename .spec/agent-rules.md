<!-- generated: 2026-05-14, template: bootstrap.md -->
# Agent Rules — EasyP

Mandatory rules for AI agents working on the EasyP codebase.

## Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) conventions.
- `golangci-lint` v2.1.6 with `staticcheck` enabled — all code must pass.
- One type/concern per file. File name = `snake_case.go`.
- Avoid global state. Dependencies are injected via constructors.
- Use `var _ Interface = (*Struct)(nil)` to verify interface compliance at compile time.

## Naming Conventions

- **Files:** `snake_case.go` (e.g., `enum_pascal_case.go`)
- **Test files:** `*_test.go` alongside the source file
- **Packages:** short, lowercase, no underscores
- **Rule structs:** PascalCase, auto-converted to `UPPER_SNAKE_CASE` for rule names
- **Mock dirs:** `mocks/` subdirectory next to the interface definition
- **Config fields:** `snake_case` in YAML, matching Go struct tags

## Error Handling

- **Always wrap errors with context:** `fmt.Errorf("funcName: %w", err)`
- Never return bare errors — the call chain must be traceable.
- Use sentinel errors (`var ErrX = errors.New(...)`) for expected conditions.
- Use typed errors (structs implementing `error`) for domain-specific context (e.g., `OpenImportFileError`, `GitRefNotFoundError`).
- Exit codes: `0` = success, `1` = issues found, `2` = infrastructure error.

## Testing

- Framework: `github.com/stretchr/testify` (`assert`, `require`).
- Runner: `gotestsum` with `--format pkgname`.
- Flags: `-race -count=1` (always).
- Fixtures: proto files in `testdata/` organized by scenario.
- Mocks: generated with `mockery` v2.41.0 into `mocks/` subdirectories.
- Mockery config: `with-expecter: true`, `inpackage: false`, `disable-version-string: true`.
- Each lint rule has a dedicated `*_test.go` file.

## Dependencies

- Add dependencies via `go get`. Update `go.mod` and `go.sum`.
- Regenerate mocks after interface changes: `task mocks`.
- Regenerate schemas after config type changes: `task schema:generate`.
- Never vendor Go dependencies (Go modules only).

## Formatting

- Use `gofmt` / `goimports` (enforced by `golangci-lint`).
- Comments: preserve all existing comments unrelated to your changes.
- Logging: use the structured `logger.Logger` wrapper around `slog`. Log to stderr.
- Output: use `fmt.Fprintf(w, ...)` for command output to stdout.
