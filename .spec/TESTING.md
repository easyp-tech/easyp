<!-- generated: 2026-09-30, template: development.md -->
# Testing

## Overview

EasyP tests are Go <code>*_test.go</code> files distributed beside the packages they
exercise. The test task runs every package with coverage, the race detector,
and test caching disabled:

~~~bash
task test
~~~

The command expands to:

~~~bash
bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...
~~~

Tests use the standard <code>testing</code> package and <code>github.com/stretchr/testify</code>.
Testify's <code>require</code> package is the most common assertion style in the lint
rule tests, while some smaller package tests use <code>t.Fatalf</code> or <code>t.Errorf</code>.

## Test Package Naming

Both external and internal test packages are present; select the package style
that matches the code being tested and neighboring tests.

Lint rule tests use an external package:

~~~go
package rules_test

import (
    "testing"

    "github.com/easyp-tech/easyp/internal/rules"
)
~~~

For example, <code>internal/rules/file_lower_snake_case_test.go</code> imports
<code>internal/rules</code> explicitly and tests its exported rule type. Tests in
<code>internal/core</code> commonly use the internal package instead:

~~~go
package core
~~~

<code>internal/core/generate_path_test.go</code> calls the unexported <code>stripPrefix</code>
function directly, which requires the internal <code>core</code> package. Private behavior is tested from same-package tests rather than exposed
through test-only export shims.

## Test File Structure

Rule validation tests commonly use a table keyed by scenario name, make
parallelism explicit, create test data through a shared helper, and use
Testify assertions:

~~~go
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
~~~

The actual test additionally verifies rule issue fields. The abbreviated
example shows the recurring structure without duplicating all fixture data.

## Table-Driven Tests

The repository uses both map- and slice-based table tests.

Rule tests typically use <code>map[string]struct</code> where each map key is the
subtest name. They rebind the range values before the closure:

~~~go
for name, tc := range tests {
    name, tc := name, tc
    t.Run(name, func(t *testing.T) {
        // use name and tc
    })
}
~~~

This is a retained convention in older tests. Go 1.26.6 uses per-iteration
variables for this loop form, so rebinding is no longer necessary for capture
safety; it does not isolate shared fixtures or mocks.

Other tests use a slice with an explicit <code>name</code> field. For example,
<code>internal/adapters/plugin/options_test.go</code> defines:

~~~go
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
~~~

Prefer descriptive scenario names. The project testing skill calls for
<code>name string</code> as the first table field, input fields next, setup fields after
that, and expected values last. Existing legacy tests do not all follow every
new convention, so use nearby tests and the testing skill when adding code.

## Parallel Execution

Many rule tests call <code>t.Parallel()</code> at the top level and in every subtest:

~~~go
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
~~~

The project testing skill treats this as the standard for table-driven tests.
Keep each parallel subtest isolated: create mutable dependencies inside the
subtest and use <code>t.TempDir()</code> for filesystem state. Do not remove
parallelism merely to work around coupled test state.

Not every existing test is parallel. For example,
<code>internal/core/generate_path_test.go</code> uses sequential subtests, and
<code>internal/config/v1/environment_test.go</code> uses <code>t.Setenv</code> for process
environment variables. When modifying a test, preserve its safety requirements rather
than adding parallelism to code with shared global state.

## Helpers, Fixtures, and Cleanup

<code>internal/rules/init_test.go</code> supplies the shared <code>start(t testing.TB)</code>
helper for rule tests. It parses <code>.proto</code> files from <code>testdata/</code> and returns
both <code>*require.Assertions</code> and a map of <code>core.ProtoInfo</code>.

The helper marks itself as a helper and registers file cleanup:

~~~go
func parseFile(t testing.TB, assert *require.Assertions, path string) core.ProtoInfo {
    t.Helper()

    f, err := os.Open(path)
    assert.NoError(err)
    t.Cleanup(func() { assert.NoError(f.Close()) })

    got, err := protoparser.Parse(f)
    assert.NoError(err)
    // interpret the parsed proto and its imports
}
~~~

Use <code>t.Helper()</code> for reusable helpers so failure locations point to the test
call site. Use <code>t.Cleanup()</code> for resources opened by the test. Local Git snapshot tests live in
<code>internal/adapters/go_git/snapshot_test.go</code>; generation/descriptor fixtures
live in <code>internal/generation</code>. These are ordinary Go tests, without a separate
snapshot-testing framework.

## Assertions and Errors

Testify is a direct module dependency. Rule tests frequently construct
<code>require.New(t)</code> and use the returned assertion object:

~~~go
assert := require.New(t)
assert.Equal(expMessage, message)
assert.NoError(err)
~~~

For new tests, use <code>require</code> for a condition that must hold before later
assertions can be meaningful, such as an error-free operation or a non-nil
value. Use <code>assert</code> for independent value checks when continued reporting is
useful. For expected sentinel errors, the project testing skill specifies
<code>require.ErrorIs(t, err, wantErr)</code>.

Some existing compact tests use direct fatal failures:

~~~go
if result != tt.expected {
    t.Fatalf("flattenOptions() result = %q, want %q", result, tt.expected)
}
~~~

Follow the surrounding test's established assertion style when making a
small extension, while using the project testing conventions for new suites.

## Mock Generation

Mockery is the generated-mock tool. <code>Taskfile.yml</code> pins Mockery v2.53.7 and
installs it into <code>bin/</code> through <code>task init</code>.

~~~bash
task mocks
~~~

The optional Taskfile helper generates <code>Rule</code> and <code>CurrentProjectGitWalker</code>
from <code>internal/core</code> into <code>internal/core/mocks</code>, and <code>Console</code> from
<code>internal/adapters/console</code> into <code>internal/adapters/console/mocks</code>.
There are no checked-in Mockery outputs or <code>go:generate</code> directives in the
current source. Neither build nor test depends on generation. Do not commit
unused outputs or recreate removed interfaces to satisfy an old command.
Generated mocks import <code>testify/mock</code>, whose <code>objx</code> dependency is not
needed by the handwritten doubles. To check optional outputs without changing
the root <code>go.mod</code>/<code>go.sum</code>, resolve those dependencies in a temporary modfile:

~~~sh
(
    set -eu
    mock_check_dir=$(mktemp -d)
    trap 'rm -rf "$mock_check_dir"' EXIT
    cp go.mod go.sum "$mock_check_dir/"
    go test -mod=mod -modfile="$mock_check_dir/go.mod" ./internal/core/mocks ./internal/adapters/console/mocks
)
~~~

Keep Mockery's tool dependencies separate from the root module. If a tool
build or package-loading error mentions an older <code>golang.org/x/tools</code>,
record that error rather than changing application dependencies to fix the tool.

<code>internal/modules/resolve_test.go</code> and <code>internal/modules/operations_test.go</code>
use small hand-written implementations of the <code>Source</code>/repository contracts.
Keep mutable fake state isolated per test. Update the matching test double
when an interface changes, and do not hand-edit generated outputs if future
Mockery generation is introduced.

## Integration Tests

There is no separate integration-test directory, test build tag, <code>TestMain</code>,
or test-container configuration in the current test tree. <code>task test</code> runs
all Go tests in <code>./...</code>; it is therefore the single repository test command
rather than a unit/integration split.

Tests that parse Protocol Buffers use checked-in files under <code>testdata/</code>.
Local Git integration tests in <code>internal/adapters/gitmodules</code> and
<code>internal/api</code> initialize temporary repositories and run system Git; they do
not need a public remote. Module tests cover resolver ordering, retained
comments, lock/hash checks, replacements and vendor staging. They do not
start an external database or Docker Compose stack. EasyP's Git-backed
module cache and <code>easyp_vendor</code> output belong to
the dependency-management flow; see
[config/dependency.md](config/dependency.md), not a file-upload test suite.

## Commands

~~~bash
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

# Optional Mockery targets for core and console
task mocks

# Offline Taskfile dispatch and failure checks (requires only Task and shell tools)
task dev-tools:check

# Validate the native v1 example in a temporary workspace
task proto:check
~~~

Run at least the package tests you touched after a behavior change. Run
<code>task test</code> before handing off broader changes; it is also the command used
by the GitHub Actions test workflow after <code>task init</code>.
