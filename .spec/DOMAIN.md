<!-- generated: 2026-07-27, template: core.md -->
# EasyP Domain Model

EasyP operates on Protocol Buffers source trees, lint findings, generation inputs and plugins, and Git-backed dependency modules. The core domain types are defined primarily in `internal/core`, `internal/core/models`, and `internal/config`.

## Core Proto Entities

### `ProtoInfo`

Defined in `internal/core/dom.go`. It is the parsed view of one protobuf source file and the imported files available to it.

```go
type ProtoInfo struct {
	Path                 string
	Info                 *unordered.Proto
	ProtoFilesFromImport map[ImportPath]*unordered.Proto
}
```

- `Path` is the source file path.
- `Info` is the parsed protobuf syntax tree.
- `ProtoFilesFromImport` maps import paths to parsed imported files.

### `Issue` and `IssueInfo`

Defined in `internal/core/dom.go`. Linting and breaking checks emit these values.

```go
type Issue struct {
	Position   meta.Position
	SourceName string
	Message    string
	RuleName   string
}

type IssueInfo struct {
	Issue
	Path string
}
```

- `Position` identifies the protobuf source location.
- `SourceName` records the checked source.
- `Message` is supplied by the rule or checker.
- `RuleName` is the upper-snake-case rule identity.
- `Path` is added when the issue is associated with a walked file.

### `ProtoData` and `Collection`

Breaking checks collect parsed files into a package-keyed representation.

```go
type Collection struct {
	Imports  map[ImportPath]Import
	Services map[string]Service
	Messages map[string]Message
	OneOfs   map[string]OneOf
	Enums    map[string]Enum
}

type ProtoData map[PackageName]*Collection
```

Each collection groups proto declarations for one `PackageName`. Nested messages, oneofs, and enums retain a dotted parent path such as `MainMessage.NestedMessage`.

### Declaration Wrappers

`internal/core/dom.go` wraps parser nodes with their source context:

| Type | Context retained |
|------|------------------|
| `Import` | Proto file path, package name, and parsed import. |
| `Service` | Proto file path, package name, and parsed service. |
| `Message` | Message path, proto file path, package name, and parsed message. |
| `OneOf` | Oneof path, proto file path, package name, and parsed oneof. |
| `Enum` | Enum path, proto file path, package name, and parsed enum. |

## Rule Model

### `Rule`

Every lint rule implements the core interface:

```go
type Rule interface {
	Message() string
	Validate(ProtoInfo) ([]Issue, error)
}
```

`internal/rules/builder.go` creates selected rules using the configured `lint.use`, `lint.except`, and `lint.ignore_only` values. The named groups are `MINIMAL`, `BASIC`, `DEFAULT`, `COMMENTS`, and `UNARY_RPC`.

### Ignore Model

The core accepts:

- `ignore []string` for paths skipped while walking.
- `ignoreOnly map[string][]string` for paths skipped for a specific rule.
- Comment directives when `allow_comment_ignores` is enabled.

`Core.CheckIsIgnored` recognizes both `nolint:<RULE_NAME>` and the backward-compatible `buf:lint:ignore <RULE_NAME>` form.

## Module and Dependency Entities

### `Module`

Defined in `internal/core/models/module.go`.

```go
type Module struct {
	Name    string
	Version RequestedVersion
}
```

`Name` is a remote repository path. `Version` comes from a `repo[@version]` dependency declaration. When omitted, it is `RequestedVersion("")`, exposed as the `Omitted` constant.

### Version Values

```go
type ModuleHash string
type RequestedVersion string

type GeneratedVersionParts struct {
	Datetime   string
	CommitHash string
}
```

`RequestedVersion` can identify an explicit request or parse a generated version in the `v0.0.0-<datetime>-<commit>` form. Its methods are `GetParts`, `IsGenerated`, and `IsOmitted`.

### Resolved Module State

```go
type Revision struct {
	CommitHash string
	Version    string
}

type LockFileInfo struct {
	Name    string
	Version string
	Hash    ModuleHash
}

type InstalledModuleInfo struct {
	ModuleName      string
	Hash            ModuleHash
	RevisionVersion string
}
```

- `Revision` represents the commit and resulting version selected from Git.
- `LockFileInfo` is the persisted dependency lock entry.
- `InstalledModuleInfo` describes a cached installed module and its directory hash.

### Module Configuration

```go
type ModuleConfig struct {
	Directories  []string
	Dependencies []Module
}
```

It describes directories containing protobuf files and dependencies discovered in a remote module configuration. See [config/dependency.md](./config/dependency.md) for the full dependency resolution and installation flow.

## Generation Model

### Plugin and Input

Core generation receives already-converted values:

```go
type PluginSource struct {
	Name    string
	Remote  string
	Path    string
	Command []string
}

type Plugin struct {
	Source      PluginSource
	Out         string
	Options     map[string][]string
	WithImports bool
}
```

Plugin execution source selection is command first, then remote, then a built-in plugin absent from `PATH`, and otherwise local execution.

```go
type InputGitRepo struct {
	URL          string
	SubDirectory string
	Root         string
}

type InputFilesDir struct {
	Path string
	Root string
}

type Inputs struct {
	InputFilesDir []InputFilesDir
	InputGitRepos []InputGitRepo
}
```

`Query` holds resolved import directories, plugins, and protobuf files used to create a `CodeGeneratorRequest`.

### Managed Mode

The configuration model contains:

```go
type ManagedMode struct {
	Enabled bool
	Disable []ManagedDisableRule
	Override []ManagedOverrideRule
}
```

Disable rules can match a module, protobuf package, path, file option, field option, or field. Override rules set a file or field option value and can be scoped by module, package, path, or field.

The configuration validator requires valid option combinations: a rule cannot use both file and field options, and a field scope requires a field option.

## Configuration Model

The top-level `internal/config.Config` has four sections:

```go
type Config struct {
	Lint          LintConfig
	Deps          []string
	Generate      Generate
	BreakingCheck BreakingCheck
}
```

Configuration parsing expands environment variables before YAML unmarshalling and calls `Config.Validate`. A generation plugin must have exactly one source among `name`, `remote`, `path`, and `command`.

The MCP schema model and generated schema files describe the same user-facing configuration domain. Schema changes originate in `mcp/easypconfig` and/or `internal/config`, not directly in `schemas/`.

## Business Errors

For the full business error catalog (codes, mappings, and retry guidance) see [ERRORS.md](./ERRORS.md). That document is generated separately.

Key sentinels currently include:

| Sentinel | Defined in | Meaning |
|----------|------------|---------|
| `core.ErrInvalidRule` | `internal/core` | Configured lint rule is not known. |
| `core.ErrRepositoryDoesNotExist` | `internal/core` | A required project repository is absent. |
| `core.ErrEmptyInputFiles` | `internal/core` | Generation selected no protobuf input files. |
| `models.ErrVersionNotFound` | `internal/core/models` | A requested module version cannot be resolved. |
| `models.ErrFileNotFound` | `internal/core/models` | A module configuration file was not found. |
| `models.ErrModuleNotInstalled` | `internal/core/models` | An installed module directory is absent. |
| `models.ErrModuleInfoFileNotFound` | `internal/core/models` | Cached install metadata is absent. |
| `models.ErrHashDependencyMismatch` | `internal/core/models` | Cached module contents do not match the expected hash. |
| `models.ErrModuleNotFoundInLockFile` | `internal/core/models` | A module has no lock-file entry. |
| `models.ErrRequestedVersionNotGenerated` | `internal/core/models` | A version cannot be parsed as a generated version. |

The CLI additionally uses command-level sentinels such as `api.ErrHasLintIssue` and `api.ErrBreakingCheckIssue` to produce exit code 1 for findings.
