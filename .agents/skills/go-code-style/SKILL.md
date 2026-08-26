---
name: go-code-style
description: "Project-specific Go code style enforcer. Use when: writing Go code, reviewing Go code, refactoring, creating new packages. Covers error wrapping, defer patterns, variable assignment, naming, imports, and other project conventions."
argument-hint: "Describe the Go code you are writing or reviewing"
---

# Go Code Style — EasyP Service

You enforce project-specific Go coding conventions for the EasyP Service codebase. These rules are mandatory and take precedence over general Go style guides.

## When to Use

- Writing any Go code in this project
- Reviewing or refactoring existing Go code
- Creating new packages, types, or functions
- Generating code from templates

Every Go code change MUST comply with these rules. Apply them automatically — do not ask the user for permission.

## Anti-Patterns (NEVER DO)

These are the most common mistakes. Memorize and never repeat them.

### 1. Error wrapping with package prefix

```go
// ❌ WRONG — never add package name prefix
return fmt.Errorf("goosemigrate: sql.Open: %w", err)
return fmt.Errorf("registry: c.db.Query: %w", err)

// ✅ CORRECT — only the called function/method name
return fmt.Errorf("sql.Open: %w", err)
return fmt.Errorf("c.db.Query: %w", err)
```

### 2. Assignment inside `if` condition

```go
// ❌ WRONG — never assign and check in the same line
if err = db.PingContext(ctx); err != nil {
    return fmt.Errorf("db.PingContext: %w", err)
}

if _, err = provider.Up(ctx); err != nil {
    return fmt.Errorf("provider.Up: %w", err)
}

// ✅ CORRECT — assignment and check on separate lines
err = db.PingContext(ctx)
if err != nil {
    return fmt.Errorf("db.PingContext: %w", err)
}

_, err = provider.Up(ctx)
if err != nil {
    return fmt.Errorf("provider.Up: %w", err)
}
```

**Exception:** short variable declaration with `:=` IS allowed in `if` when the variable is scoped to the block:

```go
// ✅ OK — new variable scoped to the if block
if info, err := c.registry.Get(ctx, name); err != nil {
    ...
}
```

### 3. Bare `defer Close()`

```go
// ❌ WRONG — silently ignores close errors
defer db.Close()
defer file.Close()
defer rows.Close()

// ✅ CORRECT — always log close errors
defer func() {
    if err := db.Close(); err != nil {
        slog.Error("db.Close", "error", err)
    }
}()
```

This applies to ALL resources that return an error from Close: `*sql.DB`, `*os.File`, `*sql.Rows`, `io.Closer`, etc.

### 4. Inline comments on control flow

```go
// ❌ WRONG — no inline comments on if/for/return
if err != nil { // check error
    return err // propagate
}

// ✅ CORRECT — comment above if needed, or omit entirely
// Validate connection before proceeding.
if err != nil {
    return err
}
```

## Error Handling Rules

1. **Wrap at every call site** with `fmt.Errorf("<called_function>: %w", err)`
2. **Format:** only the function/method name, no package prefix
3. **Domain errors** are sentinel vars in `core/domain.go`: `core.ErrNotFound`, `core.ErrInvalidPluginName`, etc.
4. **Never create domain errors** outside `internal/core/domain.go`
5. **API error mapping:** `ErrorToStatus()` maps domain errors → gRPC status codes

```go
// Error chain example:
// adapter → core → api
return fmt.Errorf("c.registry.Create: %w", err)        // in core
return fmt.Errorf("c.db.QueryRowContext: %w", err)      // in adapter
// api layer uses ErrorToStatus() to map, does not wrap
```

## Import Ordering

Three groups, separated by blank lines. Enforced by `gci`:

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "log/slog"

    // 2. Third-party
    "github.com/pressly/goose/v3"
    "google.golang.org/grpc"

    // 3. Project packages
    "github.com/easyp-tech/service/internal/core"
)
```

## Naming Conventions

### Files

| Layer | Pattern | Example |
|-------|---------|---------|
| Domain | `domain.go`, `core.go` | `internal/core/domain.go` |
| Adapters | `<name>.go` | `internal/adapters/registry/registry.go` |
| API | `api.go`, `mcp.go` | `internal/api/api.go` |
| Tracing | `tracing_<target>.go` | `internal/telemetry/tracing_core.go` |
| Tests | `<file>_test.go` | `internal/core/crud_test.go` |

### Types

| Category | Exported? | Example |
|----------|-----------|---------|
| Domain entities | Yes | `PluginInfo`, `AuditEntry` |
| Domain interfaces | Yes | `Registry`, `Plugin`, `FeatureGate` |
| Config structs | No | `config`, `server`, `ports` |
| Adapter implementations | Yes | `Store`, `Registry`, `Worker` |

### Enums

```go
type Feature int

const (
    _ Feature = iota
    FeatureCodeGeneration
    FeaturePluginListing
)
```

**Rule:** zero value (`0`) is always reserved with `_` so that uninitialized enum variables are never silently valid.

## Interface Rules

1. **Context first:** `func (r *Registry) Get(ctx context.Context, ...) (Plugin, error)`
2. **Error always last**
3. **No `I` prefix:** `Registry`, not `IRegistry`
4. **Single-method interfaces preferred**
5. **Interface in consumer package**

## Comments

- **Language:** English only
- **Godoc:** every exported symbol has a comment starting with its name
- **No inline comments** on `if`, `for`, `return` lines unless genuinely non-obvious
- **Package comment:** `// Package <name> provides ...` in one file per package

## Struct Tags

| Layer | Tags | Case |
|-------|------|------|
| Domain (`core/`) | None | — |
| Config (`cmd/`) | `env:""` + `yaml:""` | `snake_case` |
| Proto (generated) | Don't modify | — |

Tag case: always `snake_case` — enforced by `tagliatelle` linter.

## Concurrency Patterns

- `errgroup` for parallel server startup
- `WorkerPool` for bounded plugin execution
- `signal.NotifyContext` + `forceShutdown` goroutine
- Audit worker: single goroutine, buffered channel

## Quick Checklist

Before submitting any Go code, verify:

- [ ] Error wrapping uses only function name, no package prefix
- [ ] No assignment inside `if` conditions (except `:=` for new scoped vars)
- [ ] All `defer Close()` wrapped with error logging
- [ ] Imports ordered: stdlib → third-party → project
- [ ] All exported symbols have godoc comments in English
- [ ] No inline comments on control flow lines
- [ ] Domain errors defined only in `core/domain.go`
- [ ] Enum types use `_ = iota` to reserve zero value
