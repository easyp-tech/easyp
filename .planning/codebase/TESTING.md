---
last_mapped_commit: 5c9185fd17b72af60e0b7c50d6d713d466722555
last_mapped_at: 2026-11-17
---
# Testing Patterns

**Analysis Date:** 2026-11-17

## Test Framework

**Runner:**
- Go `testing` package via gotestsum (pinned, installed to `bin/` by `task init`)
- Config: `Taskfile.yml` `test` target

**Assertion Library:**
- `github.com/stretchr/testify` — `require` for fatal preconditions, `assert` for independent value checks

**Run Commands:**
```bash
task test               # Full suite: gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...
go test -race -count=1 ./...          # Direct equivalent
go test -race -count=1 ./internal/rules   # One package
go test -race -count=1 ./internal/rules -run 'TestFileLowerSnakeCase_Validate'  # One test
task coverage           # go tool cover -html=coverage.out
task quality            # test + lint
```

~324 `*_test.go` files across the repo; 45 rule test files in `internal/rules/`.

## Test File Organization

**Location:** Colocated — `<file>_test.go` beside the implementation.

**Package style:** Both external (`package rules_test` for `internal/rules`, e.g. `internal/rules/file_lower_snake_case_test.go`) and internal (`package core`, e.g. `internal/core/generate_path_test.go` testing unexported `stripPrefix`). Match the package style of neighboring tests; private behavior is tested same-package, never via test-only export shims.

**Fixture/double locations:**
- `internal/rules/init_test.go` — shared `start(t testing.TB)` helper + `testdata/` protos
- `internal/modules/fixtures_test.go` — module source/cache fakes
- `internal/generation/fixtures_test.go` — generation/descriptor fixtures
- `internal/api/v1_fixtures_test.go` — CLI integration fixtures
- `internal/migration/apply_test.go`, `internal/migration/lock_test.go` — migration behavior

## Test Structure

**Rule test pattern (map-keyed table, from `internal/rules/*_test.go`):**
```go
func TestFileLowerSnakeCase_Validate(t *testing.T) {
    t.Parallel()

    tests := map[string]struct {
        fileName   string
        wantIssues *core.Issue
        wantErr    error
    }{
        "invalid": {fileName: invalidAuthProto3},
        "valid":   {fileName: validAuthProto},
    }

    for name, tc := range tests {
        name, tc := name, tc
        t.Run(name, func(t *testing.T) {
            t.Parallel()
            r, protos := start(t)
            rule := rules.FileLowerSnakeCase{}
            issues, err := rule.Validate(protos[tc.fileName])
            r.ErrorIs(err, tc.wantErr)
            if tc.wantIssues != nil {
                r.Contains(issues, *tc.wantIssues)
            }
        })
    }
}
```

**Slice-based table pattern (new tests must have `name string` as first field):**
```go
tests := []struct {
    name      string
    options   map[string][]string
    expected  string
    hasResult bool
}{ /* ... */ }
```
Order: name, inputs, setup, expected values last. Lowercase underscore scenario names (`missing_manifest`, `canceled_context`).

**Helpers:**
- `t.Helper()` in reusable helpers so failures point to call sites
- `t.Cleanup()` for resources opened by the test (e.g. `parseFile` in `internal/rules/init_test.go` closes files in cleanup)

## Parallelism

- `t.Parallel()` at top level AND inside each isolated subtest — the standard for table tests
- Construct mutable dependencies inside each subtest; no shared mutable mocks/fakes across parallel subtests
- Exceptions that must stay sequential: tests changing process cwd, environment (`t.Setenv`, e.g. `internal/config/v1/environment_test.go`), or other global process state (e.g. `internal/core/generate_path_test.go`)
- Do not remove parallelism to work around coupled state; fix the coupling

## Mocking

**Framework:** Mockery v2.53.7 pinned via `task init`; optional `task mocks` generates mocks for `core.Rule`, `core.CurrentProjectGitWalker` (into `internal/core/mocks`) and `console.Console` (into `internal/adapters/console/mocks`). Not a build/test prerequisite; no checked-in Mockery outputs or `go:generate` directives.

**Patterns:** Handwritten doubles dominate — small in-test implementations of `Source`/repository contracts in `internal/modules/resolve_test.go` and `internal/modules/operations_test.go`. Keep mutable fake state per-test. Update doubles when consumer interfaces change; compile generated mocks before use; never commit unused mocks (Mockery's `testify/mock` pulls `objx` — keep it out of the root `go.mod`).

**What to Mock:** source/cache/repository ports at the `internal/modules` boundary; console output.
**What NOT to Mock:** filesystem (`t.TempDir()`), system Git (integration tests init temp repos and run real git), `testdata/` proto files.

## Fixtures and Factories

**Test Data:** Checked-in `.proto` files under `testdata/` directories (parsed via `protoparser` in `internal/rules/init_test.go`); `t.TempDir()` for all independent filesystem state; temporary Git repos initialized on the fly for `internal/adapters/gitmodules` and `internal/api` integration tests.

## Coverage

**Requirements:** None enforced (no threshold). Profile written to `coverage.out` by `task test`; view with `task coverage`. Do not commit `coverage.out`.

## Test Types

**Unit Tests:** The bulk — rules, modules, core, config, adapters, all colocated.

**Integration Tests:** No separate directory, build tag, `TestMain`, or containers. Git integration tests run system Git against temporary local repos (no public remote, no Docker Compose, no database). `task test` runs everything under `./...` — single repository test command.

**E2E Tests:** Not used.

## Common Patterns

**Error assertions:**
- `require.ErrorIs` for sentinels, `require.ErrorAs` for typed details, `require.ErrorContains` for diagnostics without sentinels; `require.NoError` on success — do not invent sentinels to simplify tests

**CLI handler tests** (see `internal/api/validate_test.go`, `.agents/skills/go-testing/SKILL.md`): each case builds its own path, flag set, `cli.Context`, and writer — never shares CLI parser setup.

**Legacy style:** Some compact tests use `t.Fatalf`/`t.Errorf` directly and map-keyed tables with `name, tc := name, tc` rebinding (unnecessary since Go 1.22 per-iteration variables but retained). When extending an existing test, follow its local style; use the project conventions above for new suites.

---

*Testing analysis: 2026-11-17*
