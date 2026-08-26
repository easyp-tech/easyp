---
name: go-testing
description: "Go testing conventions enforcer. Use when: writing tests, reviewing test code, adding test cases, creating test files. Covers table-driven tests, t.Parallel, mocks, assertions, naming conventions."
argument-hint: "Describe the test you are writing or reviewing"
---

# Go Testing — EasyP Service

You enforce project-specific Go testing conventions for the EasyP Service codebase. These rules are mandatory and take precedence over general Go testing guides.

## When to Use

- Writing any Go test in this project
- Reviewing or refactoring existing tests
- Adding test cases to existing table-driven tests
- Creating new test files
- Diagnosing test failures related to shared state or race conditions

Every Go test change MUST comply with these rules. Apply them automatically — do not ask the user for permission.

## Anti-Patterns (NEVER DO)

These are the most common mistakes. Memorize and never repeat them.

### 1. Test case without a name

```go
// ❌ WRONG — anonymous struct without name field
tests := []struct {
    input   string
    wantErr bool
}{
    {"foo", false},
    {"", true},
}

// ✅ CORRECT — every case has a descriptive name
tests := []struct {
    name    string
    input   string
    wantErr bool
}{
    {
        name:  "valid_input",
        input: "foo",
    },
    {
        name:    "empty_input",
        input:   "",
        wantErr: true,
    },
}
```

### 2. Test without t.Parallel()

```go
// ❌ WRONG — missing t.Parallel() at both levels
func TestCreatePlugin(t *testing.T) {
    tests := []struct{ ... }{...}

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test body
        })
    }
}

// ✅ CORRECT — t.Parallel() at top-level AND inside each sub-test
func TestCreatePlugin(t *testing.T) {
    t.Parallel()

    tests := []struct{ ... }{...}

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            // test body
        })
    }
}
```

### 3. Shared mutable state between test cases

```go
// ❌ WRONG — cases share and mutate the same mock
func TestListPlugins(t *testing.T) {
    t.Parallel()

    reg := &mockRegistry{} // shared across cases!

    tests := []struct {
        name  string
        setup func()
    }{
        {
            name:  "empty",
            setup: func() { reg.listResult = nil },
        },
        {
            name:  "with_results",
            setup: func() { reg.listResult = []PluginInfo{{}} }, // mutates shared reg!
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            tt.setup()       // RACE: parallel sub-tests mutate shared state
            // ...
        })
    }
}

// ✅ CORRECT — each case creates its own dependencies
func TestListPlugins(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name  string
        setup func(*mockRegistry)
        // ...
    }{
        {
            name:  "empty",
            setup: func(r *mockRegistry) { r.listResult = nil },
        },
        {
            name:  "with_results",
            setup: func(r *mockRegistry) { r.listResult = []PluginInfo{{}} },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            reg := &mockRegistry{} // fresh per sub-test
            tt.setup(reg)
            // ...
        })
    }
}
```

> **Red flag:** if one test case can affect another, it means shared mutable state exists. This is a bug in the test (or the code under test). Fix the coupling — do not disable `t.Parallel()`.

### 4. Using assert for fatal preconditions

```go
// ❌ WRONG — assert does not stop the test; nil dereference follows
result, err := c.GetPlugin(ctx, req)
assert.NoError(t, err)
assert.Equal(t, "grpc", result.Group) // panic if result is nil

// ✅ CORRECT — require stops the test on failure
result, err := c.GetPlugin(ctx, req)
require.NoError(t, err)
assert.Equal(t, "grpc", result.Group)
```

### 5. Ignoring the error case

```go
// ❌ WRONG — only checks happy path
require.NoError(t, err)

// ✅ CORRECT — check specific error with ErrorIs
if tt.wantErr != nil {
    require.ErrorIs(t, err, tt.wantErr)
    return
}
require.NoError(t, err)
```

## Table-Driven Test Template

This is the canonical test structure for the project:

```go
func TestMethodName(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string          // REQUIRED, always first field
        // inputs
        req     SomeRequest
        // setup / dependencies
        setup   func(*mockDep)
        // expected outputs
        want    *SomeResult
        wantErr error
    }{
        {
            name: "success",
            req:  SomeRequest{Field: "value"},
            setup: func(m *mockDep) {
                m.result = &SomeResult{Field: "value"}
            },
            want: &SomeResult{Field: "value"},
        },
        {
            name: "not_found",
            req:  SomeRequest{Field: "missing"},
            setup: func(m *mockDep) {
                m.err = ErrNotFound
            },
            wantErr: ErrNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            // Create fresh dependencies per sub-test
            dep := &mockDep{}
            if tt.setup != nil {
                tt.setup(dep)
            }

            // Create the system under test
            sut := New(dep)

            // Execute
            got, err := sut.Method(context.Background(), tt.req)

            // Assert
            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Struct field order

Always follow this order in the test case struct:

1. `name string` — descriptive test case name (always first)
2. Input fields — request, arguments, parameters
3. Setup fields — `setup func(...)` for configuring mocks
4. Expected output fields — `want*`, `wantErr`

## t.Parallel() Rules

### Every table-driven test uses t.Parallel() at two levels

1. **Top-level** — `t.Parallel()` right after `func TestXxx(t *testing.T) {`
2. **Sub-test** — `t.Parallel()` right after `t.Run(tt.name, func(t *testing.T) {`

### Why two levels?

- **Top-level** `t.Parallel()`: allows this test function to run in parallel with other `Test*` functions in the same package.
- **Sub-test** `t.Parallel()`: allows table-driven sub-tests to run in parallel with each other within the same `Test*` function.

### When a test cannot be parallel

If a test case cannot run in parallel, it indicates one of these problems:

| Symptom | Root Cause | Fix |
|---------|-----------|-----|
| Test modifies package-level variable | Global mutable state | Inject dependency via constructor |
| Test reads file written by another case | Shared filesystem side effect | Use `t.TempDir()` per case |
| Test binds to a fixed port | Port conflict | Use port `0` (OS-assigned) |
| Tests share a mock instance | Shared mutable state | Create mock inside `t.Run` |

**Rule:** never disable `t.Parallel()` to work around coupling. Refactor the code or test instead.

## Mock Patterns

### Inline mocks — no code generation

All mocks are hand-written structs defined in `_test.go` files. No `mockgen`, `moq`, or similar tools.

```go
type mockRegistry struct {
    getResult    Plugin
    getErr       error
    listResult   []PluginInfo
    createResult *PluginInfo
    createErr    error
}

func (m *mockRegistry) Get(ctx context.Context, group, name, version string) (Plugin, error) {
    return m.getResult, m.getErr
}

func (m *mockRegistry) Create(ctx context.Context, info PluginInfo) (*PluginInfo, error) {
    return m.createResult, m.createErr
}

func (m *mockRegistry) List(ctx context.Context, filter PluginFilter) ([]PluginInfo, error) {
    return m.listResult, nil
}
```

### Mock struct naming

- Prefix: `mock` (lowercase — unexported)
- Name: the interface being mocked
- Example: `mockRegistry` implements `Registry`, `mockFeatureGate` implements `FeatureGate`

### Mock location

Same `_test.go` file that uses them. If multiple test files in a package need the same mock, put it in a shared `helpers_test.go`.

## Assertion Patterns

Uses `testify` (`github.com/stretchr/testify`):

| Function | Package | When to Use |
|----------|---------|-------------|
| `require.NoError(t, err)` | `require` | Error must be nil to continue |
| `require.ErrorIs(t, err, target)` | `require` | Error must match sentinel |
| `require.ErrorContains(t, err, substr)` | `require` | Error message must contain text |
| `require.NotNil(t, result)` | `require` | Result must exist before accessing fields |
| `assert.Equal(t, expected, actual)` | `assert` | Value comparison (test continues on failure) |
| `assert.Len(t, slice, n)` | `assert` | Collection length check |
| `assert.Empty(t, slice)` | `assert` | Collection must be empty |
| `assert.True(t, cond)` | `assert` | Boolean condition |

**Rule of thumb:**
- `require` — for preconditions: if this fails, the rest of the test is meaningless (stops test)
- `assert` — for assertions: check values, the test can continue to report multiple failures

## Naming Conventions

### Test files

| Source File | Test File |
|-------------|-----------|
| `core.go` | `core_test.go` |
| `crud.go` | `crud_test.go` |
| `registry.go` | `registry_test.go` |

### Test functions

| Target | Pattern | Example |
|--------|---------|---------|
| Exported method | `Test<Method>` | `TestCreatePlugin` |
| Unexported function | `Test_<function>` | `Test_validateName` |
| Multiple scenarios | `Test<Method>_<scenario>` | `TestCreatePlugin_WithTags` |

### Test case names

- Lowercase, descriptive, use underscores for readability
- Describe the scenario, not the assertion

```go
// ✅ Good names
"success"
"not_found"
"feature_denied"
"invalid_plugin_name"
"empty_version_defaults_to_latest"
"duplicate_returns_conflict"

// ❌ Bad names
"test1"
"case_2"
"should_work"
"error"
"TestCreatePlugin"  // don't repeat the function name
```

### Test package

Tests use the **same package** (not `_test` suffix). This allows access to unexported symbols:

```go
// internal/core/crud_test.go
package core  // same package — NOT core_test
```

## Quick Checklist

Before submitting any Go test code, verify:

- [ ] Every test case struct has `name string` as the first field
- [ ] `t.Parallel()` is called at top-level AND inside each `t.Run`
- [ ] No shared mutable state between test cases
- [ ] Mocks are created fresh inside each `t.Run` (not shared)
- [ ] `require` is used for preconditions, `assert` for value checks
- [ ] Error cases use `require.ErrorIs(t, err, expectedErr)`
- [ ] Test file uses the same package (no `_test` suffix)
- [ ] Test case names are lowercase and descriptive
- [ ] `setup` function (if used) receives mock pointers, does not capture outer scope
