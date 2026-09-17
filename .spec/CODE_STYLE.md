<!-- generated: 2026-05-14, template: core.md -->
# EasyP Code Style

## 1. Layer Structure

```
cmd/easyp/          → Entrypoint (no business logic)
  │
  ▼
internal/api/       → Transport layer (CLI flags → Core calls → output formatting)
  │
  ▼
internal/core/      → Business logic (pure domain, depends on interfaces only)
  │
  ▼
internal/adapters/  → Infrastructure (git, storage, plugins, filesystem)
```

**Rules:**
- `cmd/` may only import `internal/api/` and `internal/flags/`
- `internal/api/` may import `internal/core/`, `internal/config/`, `internal/flags/`, `internal/adapters/`
- `internal/core/` may **NOT** import `internal/adapters/` (depends on interfaces only)
- `internal/adapters/` may import `internal/core/` (to implement its interfaces)

## 2. Naming Conventions

### Files

| Layer | Pattern | Example |
|-------|---------|---------|
| Commands | `<command>.go` | `lint.go`, `generate.go` |
| Rules | `<rule_name>.go` | `enum_pascal_case.go` |
| Tests | `*_test.go` alongside source | `enum_pascal_case_test.go` |
| Mocks | `mocks/` subdirectory | `core/mocks/mock_Rule.go` |
| Adapters | `<purpose>.go` | `storage.go`, `walker.go` |

### Structs

| Layer | Visibility | Example |
|-------|-----------|---------|
| API handlers | Public | `type Lint struct{}` |
| Core interfaces | Public | `type Rule interface{}` |
| Core types | Public | `type ProtoInfo struct{}` |
| Adapter types | Public (implements interface) | `type FSWalker struct{}` |
| Internal helpers | Private | `type configLoader struct{}` |

### Rules

| Convention | Example |
|-----------|---------|
| Struct name → Rule name | `EnumPascalCase` → `ENUM_PASCAL_CASE` |
| Rule groups | `MINIMAL`, `BASIC`, `DEFAULT`, `COMMENTS`, `UNARY_RPC` |

## 3. Error Propagation

### Wrapping Pattern

Always wrap errors with the calling function name:

```go
// ✅ Correct
if err != nil {
    return fmt.Errorf("config.New: %w", err)
}

// ❌ Wrong
if err != nil {
    return err  // no context
}
if err != nil {
    return fmt.Errorf("failed: %w", err)  // vague context
}
```

### Error Types by Layer

| Layer | Error Type | Example |
|-------|-----------|---------|
| Core | Sentinel errors | `var ErrEmptyInputFiles = errors.New(...)` |
| Core | Typed errors | `type OpenImportFileError struct { FileName string }` |
| API | Exit code mapping | `errors.Is(err, ErrHasLintIssue) → os.Exit(1)` |
| Adapters | Wrapped stdlib errors | `fmt.Errorf("os.Open: %w", err)` |

### Error Chain Example

```go
// Adapter layer
func (s *Storage) Download(...) error {
    return fmt.Errorf("storage.Download: %w", err)
}

// Core layer
func (c *Core) Download(ctx context.Context) error {
    return fmt.Errorf("core.Download: %w", err)
}

// API layer
func (m Mod) Download(ctx *cli.Context) error {
    if errors.Is(err, models.ErrVersionNotFound) {
        os.Exit(1)
    }
    return fmt.Errorf("cmd.Download: %w", err)
}
```

## 4. Interface Conventions

### Argument Order

1. `context.Context` first (always)
2. Required parameters
3. Optional parameters / options structs

### Interface Compliance

```go
var _ Handler = (*Lint)(nil)           // compile-time check
var _ core.Rule = (*EnumPascalCase)(nil)
```

### Interface Design

- Keep interfaces small (1-3 methods)
- Define interfaces in the package that uses them (core), not the package that implements them (adapters)
- Use descriptive method names: `Validate`, `Download`, `GetFiles`

## 5. Import Ordering

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "os"

    // 2. Third-party
    "github.com/urfave/cli/v2"

    // 3. Internal packages
    "github.com/easyp-tech/easyp/internal/config"
    "github.com/easyp-tech/easyp/internal/core"
)
```

Groups separated by blank lines, sorted alphabetically within each group.

## 6. Logging Conventions

| Layer | Log Level | What to Log |
|-------|-----------|-------------|
| API | Warn | Non-fatal conditions (empty input files) |
| API | Error | Fatal errors before `os.Exit()` |
| Core | Debug | Internal state, decisions |
| Adapters | Debug | External calls (storage paths, git operations) |

```go
log := getLogger(ctx)
log.Debug(ctx.Context, "Use storage", slog.String("path", easypPath))
log.Warn(ctx.Context, "empty input files!")
log.Error(ctx.Context, "Cannot import file", slog.String("file name", e.FileName))
```

**Never log:** secrets, credentials, full file contents.

## 7. Test File Organization

| Pattern | Purpose |
|---------|---------|
| `*_test.go` alongside source | Unit tests |
| `testdata/` | Proto file fixtures |
| `mocks/` next to interface | Generated mocks |
| `_test` package suffix | Black-box tests |

## 8. Quick Reference

| Aspect | API Layer | Core Layer | Adapters Layer |
|--------|-----------|------------|----------------|
| **Structs** | Public | Public | Public |
| **Errors** | Exit codes via `os.Exit()` | Sentinel + typed errors | Wrapped stdlib |
| **Logging** | Warn/Error | Debug | Debug |
| **Dependencies** | config, core, adapters | Interfaces only | Core interfaces, stdlib |
| **Testing** | Integration | Unit + mocks | Unit |
