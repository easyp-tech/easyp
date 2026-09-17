<!-- generated: 2026-05-14, template: core.md -->
# EasyP Domain Model

## 1. Core Entities

### Rule

The fundamental interface for all lint rules.

```go
// internal/core/dom.go
type Rule interface {
    Message() string
    Validate(ProtoInfo) ([]Issue, error)
}
```

### ProtoInfo

Central data structure passed to every lint rule. Contains parsed proto AST and imports.

```go
// internal/core/dom.go
type ProtoInfo struct {
    Path                 string
    Info                 *unordered.Proto
    ProtoFilesFromImport map[ImportPath]*unordered.Proto
}
```

### Issue / IssueInfo

Represents a lint or breaking check finding.

```go
// internal/core/dom.go
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

### Plugin / PluginSource

Describes a code generation plugin and its source type.

```go
// internal/core/dom.go
type PluginSource struct {
    Name    string   // local: protoc-gen-<name>
    Remote  string   // remote: API service
    Path    string   // path: specific filesystem path
    Command []string // command: arbitrary command
}

type Plugin struct {
    Source      PluginSource
    Out         string
    Options     map[string][]string
    WithImports bool
}
```

### Inputs

Source configuration for code generation.

```go
// internal/core/dom.go
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

### Repo

Git repository operations interface.

```go
// internal/core/dom.go
type Repo interface {
    GetFiles(ctx context.Context, revision models.Revision, dirs ...string) ([]string, error)
    ReadFile(ctx context.Context, revision models.Revision, fileName string) (string, error)
    Archive(ctx context.Context, revision models.Revision, paths models.CacheDownloadPaths) error
    ReadRevision(ctx context.Context, version models.RequestedVersion) (models.Revision, error)
    Fetch(ctx context.Context, revision models.Revision) error
}
```

## 2. Value Objects

### Type Aliases

```go
type ImportPath string   // path in proto import statement
type PackageName string  // proto package name
```

### models.Revision

```go
// internal/core/models/revision.go
type Revision struct { /* commit hash and metadata */ }
```

### models.CacheDownloadPaths

```go
// internal/core/models/cache_download_paths.go
type CacheDownloadPaths struct { /* paths for cached downloads */ }
```

### Collection

Aggregated proto data per package.

```go
// internal/core/dom.go
type Collection struct {
    Imports  map[ImportPath]Import
    Services map[string]Service
    Messages map[string]Message   // key: message path (supports nesting)
    Enums    map[string]Enum
    OneOfs   map[string]OneOf
}

type ProtoData map[PackageName]*Collection
```

## 3. Error Types

### Typed Errors

```go
type OpenImportFileError struct { FileName string }
type GitRefNotFoundError struct { GitRef string }
```

### Sentinel Errors

```go
var ErrEmptyInputFiles          // no input files for generation
var ErrRepositoryDoesNotExist   // git repo not found
```

For the full error catalog see [ERRORS.md](./ERRORS.md).

## 4. Key Functions

### AppendIssue

Checks `nolint:` / `buf:lint:ignore` comments before appending an issue.

```go
func AppendIssue(issues []Issue, rule Rule, pos meta.Position, name string, comments []*parser.Comment) []Issue
```

### GetRuleName

Auto-derives rule name from struct name: `PascalCase → UPPER_SNAKE_CASE`.

```go
func GetRuleName(rule Rule) string {
    return toUpperSnakeCase(reflect.TypeOf(rule).Elem().Name())
}
// EnumPascalCase → ENUM_PASCAL_CASE
```

### GetPackageName

Extracts package name from parsed proto file.

```go
func GetPackageName(protoFile *unordered.Proto) PackageName
```
