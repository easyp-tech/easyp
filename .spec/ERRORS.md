<!-- generated: 2026-05-14, template: errors.md -->
# EasyP Error Handling

## 1. Error Architecture

```
┌──────────────────────────────────────────────────────┐
│  API Layer (internal/api/)                           │
│  Maps sentinel/typed errors → exit codes (0, 1, 2)   │
│  Logs fatal errors, calls os.Exit()                  │
├──────────────────────────────────────────────────────┤
│  Core Layer (internal/core/)                         │
│  Returns sentinel errors (ErrEmptyInputFiles, etc.)  │
│  Returns typed errors (OpenImportFileError, etc.)    │
│  Wraps lower-layer errors with context               │
├──────────────────────────────────────────────────────┤
│  Adapter Layer (internal/adapters/)                   │
│  Wraps infrastructure errors (git, fs, network)      │
│  Returns models.Err* sentinels where appropriate     │
├──────────────────────────────────────────────────────┤
│  Config Layer (internal/config/)                     │
│  Returns validation errors (inline + sentinel)       │
│  Private sentinels (errFileNotFound)                 │
└──────────────────────────────────────────────────────┘
```

**Propagation rules:**
- Adapters **create** infrastructure errors, wrap with `fmt.Errorf("funcName: %w", err)`
- Core **wraps** adapter errors, returns domain sentinels/typed errors
- API **maps** errors to exit codes via `errors.Is()` / `errors.As()`
- Errors are **logged** only at the API layer (avoid double-logging)

## 2. Business Error Catalog

### Core Errors (`internal/core/core.go`)

| Error | Exit Code | Description |
|-------|-----------|-------------|
| `ErrInvalidRule` | — | Invalid lint rule name in config |
| `ErrRepositoryDoesNotExist` | 2 | Git repository not found in working dir |
| `ErrEmptyInputFiles` | 0 (warning) | No input files for generation |

### Core Typed Errors (`internal/core/dom.go`)

| Error Type | Exit Code | Description |
|-----------|-----------|-------------|
| `OpenImportFileError{FileName}` | 2 | Cannot import a proto file |
| `GitRefNotFoundError{GitRef}` | 2 | Git reference not found |

### Model Errors (`internal/core/models/errors.go`)

| Error | Exit Code | Description |
|-------|-----------|-------------|
| `ErrVersionNotFound` | 1 | Dependency version tag doesn't exist |
| `ErrFileNotFound` | — | File not found in repository |
| `ErrModuleNotInstalled` | — | Module not downloaded to cache |
| `ErrModuleInfoFileNotFound` | — | Module metadata missing |
| `ErrHashDependencyMismatch` | — | Lock file hash doesn't match |

### Model Errors (other files)

| Error | Location | Description |
|-------|----------|-------------|
| `ErrRequestedVersionNotGenerated` | `models/module.go` | Requested version not generated |
| `ErrModuleNotFoundInLockFile` | `models/lock_file_info.go` | Module missing from `easyp.lock` |

### API Errors (`internal/api/`)

| Error | Exit Code | Description |
|-------|-----------|-------------|
| `ErrHasLintIssue` | 1 | Lint issues found |
| `ErrHasValidateIssue` | 1 | Config validation errors |
| `ErrBreakingCheckIssue` | 1 | Breaking changes found |
| `ErrPathNotAbsolute` | — | Internal: path is not absolute |

### Config Errors (`internal/config/config.go`)

| Error | Description |
|-------|-------------|
| `errFileNotFound` (private) | Config file not found |
| Inline validation errors | disable/override rule validation |

## 3. Error Wrapping Convention

### Adapter → Core

```go
// internal/adapters/storage/storage.go
func (s *Storage) Download(ctx context.Context, ...) error {
    if err := ...; err != nil {
        return fmt.Errorf("storage.Download: %w", err)
    }
}
```

### Core → API

```go
// internal/core/download.go
func (c *Core) Download(ctx context.Context) error {
    if err := ...; err != nil {
        return fmt.Errorf("core.Download: %w", err)
    }
}
```

### API → Exit Code

```go
// internal/api/lint.go
func (l Lint) Action(ctx *cli.Context) error {
    err := l.action(ctx, log)
    if err != nil {
        var e *core.OpenImportFileError
        switch {
        case errors.Is(err, ErrHasLintIssue):
            os.Exit(1)
        case errors.As(err, &e):
            errExit(log, 2, "Cannot import file", slog.String("file name", e.FileName))
        default:
            return err
        }
    }
    return nil
}
```

## 4. Sentinel Errors vs Typed Errors

### Sentinel Errors

Used when only the **identity** matters (no extra context):

```go
var ErrVersionNotFound = errors.New("version not found")
var ErrEmptyInputFiles = errors.New("empty input files")

// Checked with:
if errors.Is(err, models.ErrVersionNotFound) { ... }
```

### Typed Errors

Used when errors carry **extra context** (file name, git ref):

```go
type OpenImportFileError struct {
    FileName string
}
func (e *OpenImportFileError) Error() string {
    return fmt.Sprintf("open import file `%s`", e.FileName)
}

// Checked with:
var e *core.OpenImportFileError
if errors.As(err, &e) {
    log.Error("Cannot import file", slog.String("file name", e.FileName))
}
```

## 5. Error Logging

| Layer | Level | What to Log |
|-------|-------|-------------|
| API | `Error` | Fatal errors before `os.Exit(2)` |
| API | `Warn` | Non-fatal conditions (empty input files) |
| Core | — | Errors are wrapped and returned, not logged |
| Adapters | `Debug` | Infrastructure details (storage paths) |

**Rule:** Log at the top (API layer), wrap at the bottom (adapters/core).

```go
// API layer — logs and exits
func errExit(log logger.Logger, code int, msg string, attrs ...slog.Attr) {
    log.Error(context.Background(), msg, attrs...)
    os.Exit(code)
}
```

## 6. Exit Code Mapping Summary

| Exit Code | Errors | Commands |
|-----------|--------|----------|
| `0` | No errors / `ErrEmptyInputFiles` (warning) | All |
| `1` | `ErrHasLintIssue`, `ErrBreakingCheckIssue`, `ErrHasValidateIssue`, `ErrVersionNotFound` | lint, breaking, validate, mod |
| `2` | `OpenImportFileError`, `GitRefNotFoundError`, `ErrRepositoryDoesNotExist` | lint, breaking |
