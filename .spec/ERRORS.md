<!-- generated: 2026-07-27, template: errors.md -->
# Error Handling

EasyP is a command-line application. It does not expose an HTTP or gRPC
server API, so it has no HTTP status mapping, gRPC status mapping, or
server-side JSON error envelope. In this document, the API layer means the
command handlers in `internal/api`.

## Error Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│ CLI command layer: internal/api                              │
│ Prints lint, breaking, or config-validation results.         │
│ Maps selected errors to process exit code 1 or 2.             │
├─────────────────────────────────────────────────────────────┤
│ Application layer: internal/core                             │
│ Returns sentinel and typed domain errors.                    │
│ Adds operation context using fmt.Errorf("operation: %w").    │
├─────────────────────────────────────────────────────────────┤
│ Adapter layer: internal/adapters                             │
│ Converts selected filesystem and Git conditions to sentinels.│
│ Wraps other infrastructure errors with operation context.    │
├─────────────────────────────────────────────────────────────┤
│ Infrastructure                                               │
│ Filesystem, Git, subprocess, plugin, parser, and YAML errors│
└─────────────────────────────────────────────────────────────┘
```

Errors normally propagate upward as Go `error` values. Core and adapters use
`errors.Is` to identify sentinels without discarding wrapped context; CLI
handlers use `errors.Is` and `errors.As` where they provide specific behavior.
Unhandled errors are returned to `main`, where `app.Run` failures are passed to
`log.Fatal` (`cmd/easyp/main.go`), terminating the process with exit code 1.

`errExit` in `internal/api/temporaly_helper.go` is the explicit user-error
path: it logs at error level and calls `os.Exit` with the supplied code. It is
currently used by the `lint` and `breaking` command handlers.

## Core and Model Sentinel Catalog

The following sentinels are defined by `internal/core` and
`internal/core/models`. They have no machine-readable error codes beyond their
Go identifiers and message text.

| Name | Defined in | Meaning | CLI-specific mapping |
|---|---|---|---|
| `ErrInvalidRule` | `internal/core/core.go` | A configured lint rule is not recognized. | No dedicated mapping; returned error ultimately exits 1. |
| `ErrRepositoryDoesNotExist` | `internal/core/core.go` | A required Git repository cannot be opened. | `breaking` logs “Repository does not exist in current directory” and exits 2. |
| `ErrEmptyInputFiles` | `internal/core/core.go` | Generation resolved no input `.proto` files. | `generate` logs a warning and returns success (0). |
| `ErrRootOutsideProject` | `internal/core/breaking_check.go` | The breaking-check working root lies outside its Git repository. | No dedicated mapping; returned error ultimately exits 1. |
| `ErrVersionNotFound` | `internal/core/models/errors.go` | A requested dependency version or revision cannot be found. | `mod download`, `mod update`, and `mod vendor` exit 1 directly. |
| `ErrFileNotFound` | `internal/core/models/errors.go` | A Git-backed file cannot be read. | Used internally as an expected absence by module-config reading; otherwise propagates. |
| `ErrModuleNotInstalled` | `internal/core/models/errors.go` | An installed module directory is absent while calculating its hash. | No dedicated mapping; returned error ultimately exits 1. |
| `ErrModuleInfoFileNotFound` | `internal/core/models/errors.go` | Cached installed-module metadata is absent. | `Core.Get` treats it as a cache miss and installs the module. |
| `ErrHashDependencyMismatch` | `internal/core/models/errors.go` | Cached module contents do not match the lock-file hash. | No dedicated mapping; returned error ultimately exits 1. |
| `ErrRequestedVersionNotGenerated` | `internal/core/models/module.go` | A version is not in EasyP’s generated-version form. | Returned by `RequestedVersion.GetParts`; no CLI-specific mapping. |
| `ErrModuleNotFoundInLockFile` | `internal/core/models/lock_file_info.go` | A module has no lock-file entry. | `Core.Get` accepts it when checking an installed module; `Core.Download` also handles it as expected absence. |

## CLI Outcome Sentinels and Typed Errors

`internal/api` also defines command-control sentinels. They represent expected
command outcomes rather than reusable core-domain failures.

| Name or type | Defined in | Exit code / user-facing behavior |
|---|---|---|
| `ErrHasLintIssue` | `internal/api/lint.go` | Lint issues are printed to stdout, then `lint` exits 1. |
| `ErrHasValidateIssue` | `internal/api/lint.go` | `validate-config` returns it after outputting an invalid result; `main` logs it and exits 1. |
| `ErrBreakingCheckIssue` | `internal/api/breaking_check.go` | Breaking issues are printed to stdout, then `breaking` exits 1. |
| `ErrPathNotAbsolute` | `internal/api/temporaly_helper.go` | Returned if the EasyP cache path cannot be made absolute; no dedicated output mapping. |
| `*core.OpenImportFileError` | `internal/core/dom.go` | `lint` and `breaking` log “Cannot import file” with `file name`, then exit 2. |
| `*core.GitRefNotFoundError` | `internal/core/dom.go` | `breaking` logs “Cannot find git ref” with `ref`, then exits 2. |

The typed errors carry the field needed for user-facing logging:
`OpenImportFileError.FileName` and `GitRefNotFoundError.GitRef`. They are
matched with `errors.As`, not a sentinel identity check.

## Exit Code Behavior

| Exit code | When it is used | Observable behavior |
|---|---|---|
| `0` | Successful command execution; also `generate` with `ErrEmptyInputFiles`. | Normal output, or a warning for empty generation input. |
| `1` | Lint findings, breaking-change findings, dependency version not found, and unhandled returned errors. | Findings are printed for lint/breaking; other errors follow their handler or `log.Fatal` path. |
| `2` | Selected invalid runtime conditions in `lint` and `breaking`. | `errExit` writes a structured `slog` error to stderr with an actionable message and attributes. |

There is no central exit-code mapper. Each command handler decides whether to
return an error, call `os.Exit(1)`, or call `errExit(..., 2, ...)`. Therefore,
new commands should document any explicit exit behavior they introduce.

## Wrapping Convention

The project convention is to add the failing operation as a prefix and retain
the original error with `%w`. `Core.Get` provides the representative pattern:

```go
installedModuleInfo, err = c.storage.ReadInstalledModuleInfo(cacheDownloadPaths)
if err != nil {
    if !errors.Is(err, models.ErrModuleInfoFileNotFound) {
        return fmt.Errorf("c.storage.ReadInstalledModuleInfo: %w", err)
    }
    needToInstall = true
}
```

This pattern is in `internal/core/get.go`. It first identifies an expected
sentinel, then wraps every other adapter failure with the core operation that
failed. The `breaking` core implementation applies the same rule at method
boundaries, for example `fmt.Errorf("c.Download: %w", err)` in
`internal/core/breaking_check.go`.

At the CLI boundary, commands add their own context before returning errors,
such as `fmt.Errorf("config.New: %w", err)` and
`fmt.Errorf("generator.Generate: %w", err)`. Commands must preserve `%w` when
the caller needs to inspect a sentinel or typed error using `errors.Is` or
`errors.As`.

## User-Facing Output Formats

EasyP does not serialize a general error response object. Output is
command-specific:

- `lint` and `breaking` print each `core.IssueInfo` to stdout. Text output is
  `path:line:column:source message (rule)`; JSON output encodes one issue per
  line in `internal/api/lint.go`.
- `validate-config` prints a result, not an error envelope. JSON output has
  `valid`, optional `errors`, and optional `warnings`; every validation issue
  contains `code`, `message`, optional `line`, `column`, and `severity`.
- In text mode, `validate-config` prints `VALID: true|false`, followed by
  `ERRORS` and/or `WARNINGS` rows when present.
- Explicit `errExit` cases are logged to stderr through the configured
  text-format `slog` handler.

`config.ValidationIssue` is a typed diagnostic record, not an `error`
implementation. `ValidateRaw` assigns `envsubst_error` when environment
variable expansion fails and `yaml_validation` for schema-collector findings.
Warnings alone leave `valid` true; only `SeverityError` makes validation fail.

## Sentinel Errors Versus Error Types

Use a sentinel when identity alone controls behavior. The core model errors
are compared with `errors.Is`, including `ErrVersionNotFound` in `mod` commands
and `ErrModuleInfoFileNotFound` in the cache-miss path.

Use a typed error when the caller needs structured context to present an
actionable diagnostic. `OpenImportFileError` and `GitRefNotFoundError` are the
current examples and are inspected with `errors.As`.

No project-defined error interface was found. The code relies on Go’s standard
`error` interface plus `errors.Is` and `errors.As`.

## Retry Policy

No generic retry, backoff, or retryability classification is implemented in
`internal/`. Git, filesystem, plugin, and remote-executor failures propagate to
their callers. Callers should not assume retry behavior from a sentinel unless
they add and document it explicitly.

## Error Logging

The application initializes a text `slog` logger that writes to stderr in
`cmd/easyp/main.go`; debug logging is enabled only with the debug flag. Core
operations log useful operational context, for example dependency `module` and
`version` in `Core.Get`.

Log a failure at the boundary that terminates or meaningfully handles it.
`errExit` is the current terminal error logger. Core and adapter layers should
generally wrap errors with context rather than log the same failure again,
which avoids duplicate user-visible error records.
