<!-- generated: 2026-05-14, template: development.md -->
# EasyP Testing

## 1. Test Package Naming

Tests use the **external test package** convention (`_test` suffix):

```go
package rules_test  // external package — tests internal/rules from outside
```

This enforces testing only through the public API of the package.

## 2. Test File Structure

Example from `internal/rules/enum_pascal_case_test.go`:

```go
package rules_test

import (
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/yoheimuta/go-protoparser/v4/parser/meta"

    "github.com/easyp-tech/easyp/internal/core"
    "github.com/easyp-tech/easyp/internal/rules"
)

func TestEnumPascalCase_Validate(t *testing.T) {
    t.Parallel()

    tests := map[string]struct {
        fileName   string
        wantIssues *core.Issue
        wantErr    error
    }{
        "invalid": {
            fileName: invalidAuthProto,
            wantIssues: &core.Issue{
                Position: meta.Position{
                    Line:   49,
                    Column: 1,
                },
                SourceName: "social_network",
                Message:    "enum name must be in PascalCase",
                RuleName:   "ENUM_PASCAL_CASE",
            },
        },
        "valid": {
            fileName: validAuthProto,
            wantErr:  nil,
        },
    }

    for name, tc := range tests {
        name, tc := name, tc
        t.Run(name, func(t *testing.T) {
            t.Parallel()

            r, protos := start(t)

            rule := rules.EnumPascalCase{}
            issues, err := rule.Validate(protos[tc.fileName])
            r.ErrorIs(err, tc.wantErr)
            // ... assert issues
        })
    }
}
```

## 3. Key Patterns

### Table-Driven Tests (map-based)

All rule tests use `map[string]struct{}` for test cases:

```go
tests := map[string]struct {
    fileName   string
    wantIssues *core.Issue
    wantErr    error
}{
    "valid":   { ... },
    "invalid": { ... },
}
```

**Why map?** Test names are the keys — readable and self-documenting.

### Parallel Execution

All tests run in parallel:

```go
func TestFoo(t *testing.T) {
    t.Parallel()
    // ...
    t.Run(name, func(t *testing.T) {
        t.Parallel()
        // ...
    })
}
```

### Variable Capture in Range Loops

Loop variables are captured to avoid closure issues:

```go
for name, tc := range tests {
    name, tc := name, tc  // capture
    t.Run(name, func(t *testing.T) {
        // safe to use name, tc in parallel subtests
    })
}
```

### Shared Test Helper: `start()`

Rule tests use a `start(t)` helper that:
- Creates a `require` asserter
- Parses all test proto files from `testdata/`
- Returns `(asserter, protoInfoMap)`

### Test Fixtures

Proto files for testing live in `testdata/` organized by scenario:
- `testdata/valid/auth.proto` — valid proto file
- `testdata/invalid/auth.proto` — proto file with known issues

## 4. Mock Generation

| Setting | Value |
|---------|-------|
| **Tool** | `mockery` v2.41.0 |
| **Command** | `task mocks` |
| **Output** | `mocks/` subdirectory next to interface |
| **Config** | `with-expecter: true`, `inpackage: false`, `disable-version-string: true` |

Generated mocks implement core interfaces (`Rule`, `DirWalker`, `Storage`, etc.) using `testify/mock`.

**Usage in tests:**

```go
mockRule := mocks.NewMockRule(t)
mockRule.EXPECT().Validate(mock.Anything).Return([]core.Issue{}, nil)
```

## 5. Integration Tests

- No separate integration test suite — the project is a CLI tool
- All tests are unit tests with mocked dependencies
- Proto parsing is tested against real `.proto` fixtures in `testdata/`
- Build tags are not used for test separation

## 6. Commands

```bash
# Unit tests (with race detector)
task test

# Coverage report (opens in browser)
task coverage

# Regenerate mocks before testing
task mocks

# Full quality check (test + lint)
task quality
```

### Running specific tests

```bash
# Single test
go test -v -run TestEnumPascalCase ./internal/rules/

# Single package
go test -race -count=1 ./internal/core/

# All tests with verbose output
go test -v -race -count=1 ./...
```
