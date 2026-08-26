<!-- generated: 2026-07-27, template: development.md -->
# Testing

## Overview

EasyP tests are Go `*_test.go` files distributed beside the packages they
exercise. The test task runs every package with coverage, the race detector,
and test caching disabled:

```bash
task test
```

The command expands to:

```bash
bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...
```

Tests use the standard `testing` package and `github.com/stretchr/testify`.
Testify's `require` package is the most common assertion style in the lint
rule tests, while some smaller package tests use `t.Fatalf` or `t.Errorf`.

## Test Package Naming

Both external and internal test packages are present; select the package style
that matches the code being tested and neighboring tests.

Lint rule tests use an external package:

```go
package rules_test

import (
    "testing"

    "github.com/easyp-tech/easyp/internal/rules"
)
```

For example, `internal/rules/file_lower_snake_case_test.go` imports
`internal/rules` explicitly and tests its exported rule type. Tests in
`internal/core` commonly use the internal package instead:

```go
package core
```

`internal/core/generate_path_test.go` calls the unexported `stripPrefix`
function directly, which requires the internal `core` package. There are no
`export_test.go` files in the repository, so private behavior is tested from
same-package tests rather than exposed through test-only exports.

## Test File Structure

Rule validation tests commonly use a table keyed by scenario name, make
parallelism explicit, create test data through a shared helper, and use
Testify assertions:

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

The actual test additionally verifies rule issue fields. The abbreviated
example shows the recurring structure without duplicating all fixture data.

## Table-Driven Tests

The repository uses both map- and slice-based table tests.

Rule tests typically use `map[string]struct` where each map key is the
subtest name. They rebind the range values before the closure:

```go
for name, tc := range tests {
    name, tc := name, tc
    t.Run(name, func(t *testing.T) {
        // use name and tc
    })
}
```

This protects closure code from range-variable capture and is especially
important when subtests run in parallel.

Other tests use a slice with an explicit `name` field. For example,
`internal/adapters/plugin/options_test.go` defines:

```go
tests := []struct {
    name      string
    options   map[string][]string
    expected  string
    hasResult bool
}{
    {
        name:      "single scalar value",
        options:   map[string][]string{"env": {"node"}},
        expected:  "env=node",
        hasResult: true,
    },
}
```

Prefer descriptive scenario names. The project testing skill calls for
`name string` as the first table field, input fields next, setup fields after
that, and expected values last. Existing legacy tests do not all follow every
new convention, so use nearby tests and the testing skill when adding code.

## Parallel Execution

Many rule tests call `t.Parallel()` at the top level and in every subtest:

```go
func TestRule_Validate(t *testing.T) {
    t.Parallel()

    for name, tc := range tests {
        name, tc := name, tc
        t.Run(name, func(t *testing.T) {
            t.Parallel()
            // isolated test body
        })
    }
}
```

The project testing skill treats this as the standard for table-driven tests.
Keep each parallel subtest isolated: create mutable dependencies inside the
subtest and use `t.TempDir()` for filesystem state. Do not remove
parallelism merely to work around coupled test state.

Not every existing test is parallel. For example,
`internal/core/generate_path_test.go` uses sequential subtests, and
`internal/config/config_test.go` sets and unsets process environment
variables. When modifying a test, preserve its safety requirements rather
than adding parallelism to code with shared global state.

## Helpers, Fixtures, and Cleanup

`internal/rules/init_test.go` supplies the shared `start(t testing.TB)`
helper for rule tests. It parses `.proto` files from `testdata/` and returns
both `*require.Assertions` and a map of `core.ProtoInfo`.

The helper marks itself as a helper and registers file cleanup:

```go
func parseFile(t testing.TB, assert *require.Assertions, path string) core.ProtoInfo {
    t.Helper()

    f, err := os.Open(path)
    assert.NoError(err)
    t.Cleanup(func() { assert.NoError(f.Close()) })

    got, err := protoparser.Parse(f)
    assert.NoError(err)
    // interpret the parsed proto and its imports
}
```

Use `t.Helper()` for reusable helpers so failure locations point to the test
call site. Use `t.Cleanup()` for resources opened by the test. The repository
does not contain golden-file or snapshot-test infrastructure.

## Assertions and Errors

Testify is a direct module dependency. Rule tests frequently construct
`require.New(t)` and use the returned assertion object:

```go
assert := require.New(t)
assert.Equal(expMessage, message)
assert.NoError(err)
```

For new tests, use `require` for a condition that must hold before later
assertions can be meaningful, such as an error-free operation or a non-nil
value. Use `assert` for independent value checks when continued reporting is
useful. For expected sentinel errors, the project testing skill specifies
`require.ErrorIs(t, err, wantErr)`.

Some existing compact tests use direct fatal failures:

```go
if result != tt.expected {
    t.Fatalf("flattenOptions() result = %q, want %q", result, tt.expected)
}
```

Follow the surrounding test's established assertion style when making a
small extension, while using the project testing conventions for new suites.

## Mock Generation

Mockery is the generated-mock tool. `Taskfile.yml` pins Mockery v2.41.0 and
installs it into `bin/` through `task init`.

```bash
task mocks
```

The task generates mocks for storage and core interfaces. Generated files
include `internal/adapters/storage/mocks/LockFile.go` and types under
`internal/core/mocks/`. They begin with `Code generated by mockery. DO NOT
EDIT.` and use `testify/mock`.

Two core test-only mocks are generated in place by directives in
`internal/core/mod.go`:

```go
//go:generate mockery --name Storage --dir . --output . --outpkg core --filename storage_mock_test.go --structname StorageMock --inpackage --testonly
//go:generate mockery --name LockFile --dir . --output . --outpkg core --filename lockfile_mock_test.go --structname LockFileMock --inpackage --testonly
```

Run `task mocks` after changing core or storage interfaces covered by the
Taskfile. Run `go generate ./internal/core` when the in-package test-only
directives must be refreshed. Do not hand-edit generated mock files.

## Integration Tests

There is no separate integration-test directory, test build tag, `TestMain`,
or test-container configuration in the current test tree. `task test` runs
all Go tests in `./...`; it is therefore the single repository test command
rather than a unit/integration split.

Tests that parse Protocol Buffers use checked-in files under `testdata/`.
They do not start an external database, Docker Compose stack, or other
service. EasyP's Git-backed module cache and `easyp_vendor` output belong to
the dependency-management flow; see
[config/dependency.md](config/dependency.md), not a file-upload test suite.

## Commands

```bash
# Full suite: coverage, race detector, and no cache
task test

# Direct Go equivalent without gotestsum's package formatter
go test -race -count=1 ./...

# Run one package
go test -race -count=1 ./internal/rules

# Run one test or subtest pattern
go test -race -count=1 ./internal/rules -run 'TestFileLowerSnakeCase_Validate'

# Open the coverage profile produced by task test
task coverage

# Regenerate Mockery outputs after interface changes
task mocks
```

Run at least the package tests you touched after a behavior change. Run
`task test` before handing off broader changes; it is also the command used
by the GitHub Actions test workflow after `task init`.
