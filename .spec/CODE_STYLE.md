<!-- generated: 2026-07-27, template: core.md -->
# EasyP Code Style

This document records repository conventions from `AGENTS.md`, `.spec/agent-rules.md`, and existing source. Project rules take precedence over generic Go guidance.

## Layer Structure

```text
cmd/easyp
  -> internal/api       CLI handlers and composition
    -> internal/core    domain workflows and ports
      -> internal/adapters
           Git, storage, lock file, plugins, console

internal/config         external easyp.yaml model and validation
internal/rules          core.Rule implementations
mcp/easypconfig         config schema metadata and MCP exposure
```

| Layer | Convention |
|-------|------------|
| CLI | Implement `api.Handler`, build a `*cli.Command`, parse flags, call core, and map known CLI outcomes. |
| API | Construct concrete adapters and convert `config` values into core values in `buildCore`. |
| Core | Define consumer-side interfaces and operate through them; avoid importing concrete adapter packages where a port exists. |
| Adapters | Implement filesystem, Git, storage, lockfile, plugin, console, and prompt mechanics. |
| Rules | Implement `core.Rule` in a dedicated source file with a colocated test file. |
| Config/schema | Keep user-facing configuration validation in `internal/config` and schema metadata in `mcp/easypconfig`. |

## Public Types and Tags

Exported types have Go doc comments that begin with the type name. Configuration structs use `json` and `yaml` tags:

```go
type Plugin struct {
	Name    string   `json:"name,omitempty" yaml:"name,omitempty"`
	Remote  string   `json:"remote,omitempty" yaml:"remote,omitempty"`
	Path    string   `json:"path,omitempty" yaml:"path,omitempty"`
	Command []string `json:"command,omitempty" yaml:"command,omitempty"`

	Out         string     `json:"out" yaml:"out"`
	Opts        PluginOpts `json:"opts,omitempty" yaml:"opts,omitempty"`
	WithImports bool       `json:"with_imports,omitempty" yaml:"with_imports,omitempty"`
}
```

Use snake_case YAML and JSON field names where tags are required. Keep configuration types unexported when their scope is local; exported domain entities and adapter implementations use exported names.

## Conversion Boundaries

The API layer is the explicit conversion boundary from configuration to core:

```go
core.Plugin{
	Source: core.PluginSource{
		Name:    p.Name,
		Remote:  p.Remote,
		Path:    p.Path,
		Command: p.Command,
	},
	Out:         p.Out,
	Options:     p.Opts,
	WithImports: p.WithImports,
}
```

`buildCore` also converts managed-mode rules and divides generation inputs into local directory and Git-repository collections. Core types therefore do not use YAML or JSON tags for the configuration representation.

## Naming Conventions

| Subject | Convention |
|---------|------------|
| Go source | Use lowercase, underscore-separated file names. |
| Tests | Place tests beside the implementation as `<file>_test.go`. |
| Lint rules | Use `internal/rules/<rule>.go` plus `<rule>_test.go`. |
| Packages | Use short lower-case package names; adapter subpackages describe a capability. |
| Interfaces | Use descriptive names without an `I` prefix. |
| Enums | Reserve the zero value with `_ = iota` when defining enums. |
| Rule names | `core.GetRuleName` converts concrete rule type names to upper snake case. |
| Struct tags | Use snake_case field names for YAML and environment-facing data. |

Examples of rule-file naming include `rpc_pascal_case.go`, `package_defined.go`, and `enum_first_value_zero.go`.

## Interface Conventions

Interfaces are generally defined by the package that consumes them. `internal/core` owns ports such as `Rule`, `Storage`, `LockFile`, `ModuleConfig`, `CurrentProjectGitWalker`, and `Repo`.

```go
type Rule interface {
	Message() string
	Validate(ProtoInfo) ([]Issue, error)
}

type Executor interface {
	Execute(ctx context.Context, plugin Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error)
	GetName() string
}
```

Rules:

- Put `context.Context` first in interface and function signatures.
- Put `error` last in results.
- Do not prefix interface names with `I`.
- Prefer passing a small consumer-owned interface over a concrete adapter.
- Regenerate mocks with `task mocks` when core or storage interfaces change.

## Error Propagation

Wrap returned errors at each failing call site using the called function or method name only:

```go
cfg, err := config.New(ctx.Context, configPath)
if err != nil {
	return fmt.Errorf("config.New: %w", err)
}
```

```go
if err := c.Download(ctx); err != nil {
	return nil, fmt.Errorf("c.Download: %w", err)
}
```

Use `errors.Is` for sentinels and `errors.As` for typed errors. Examples in API handlers include `core.ErrEmptyInputFiles`, `models.ErrVersionNotFound`, `*core.OpenImportFileError`, and `*core.GitRefNotFoundError`.

Domain and package-manager sentinels belong in `internal/core` or `internal/core/models`; do not add duplicate sentinels in adapters or API handlers. Detailed error catalog documentation belongs in [ERRORS.md](./ERRORS.md).

## Resource Closing

Do not use a bare `defer x.Close()`. Wrap close calls and explicitly handle or intentionally discard the returned error:

```go
defer func() {
	_ = cfgFile.Close()
}()
```

When the close failure is operationally useful, `Core.close` logs the path and error with the repository logger.

## Imports

Order imports as standard library, third-party packages, then project packages. Separate groups with blank lines:

```go
import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
)
```

Use project formatting and `gci`/`golangci-lint` rather than manually preserving an invalid import order.

## Comments and Documentation

- Write Go comments and Godoc in English.
- Begin exported-symbol documentation with that symbol's name.
- Put explanatory comments above `if`, `for`, and `return` lines rather than inline.
- Keep generated schema JSON out of manual edits.

## Testing Organization

| Test target | Location and pattern |
|-------------|----------------------|
| A core behavior | `internal/core/<file>_test.go` |
| A lint rule | `internal/rules/<rule>_test.go` |
| An adapter | Its adapter package with `<file>_test.go` |
| A mocked core port | `internal/core/mocks/` |
| A mocked storage port | `internal/adapters/storage/mocks/` |

Tests use `testify`. Use `require` for fatal preconditions and `assert` for non-fatal assertions. Table-driven tests require a `name` field. Call `t.Parallel()` at the top level and inside each `t.Run`, and create dependencies per case instead of sharing mutable mocks.

Run relevant tests after behavior changes. The repository's comprehensive task uses race detection:

```sh
task test
```

## Logging

The CLI initializes an `slog` text handler and stores the project logger in `cli.App.Metadata`. API handlers retrieve it with `getLogger`; core and adapters receive it as a dependency.

| Layer | Typical use |
|-------|-------------|
| CLI/API | Log terminal failures before calling `os.Exit` for recognized conditions. |
| Core | Log operation start/completion, debug resolution data, and warnings for recoverable conditions. |
| Adapters | Log infrastructure-relevant actions through the injected logger. |

Use structured attributes such as `slog.String`, `slog.Int`, and `slog.Any`. Existing generation and breaking flows pass `context.Context` to logging methods. No project-wide correlation-ID convention is defined in the inspected code.

## Concurrency

No project-wide goroutine, worker-pool, channel, or synchronization pattern is established in the inspected core workflows. Methods accept and check `context.Context` during filesystem walks, compilation, and executor calls; preserve that cancellation propagation when extending these paths.

## Quick Reference

| Aspect | API | Core | Adapters |
|--------|-----|------|----------|
| Main responsibility | CLI translation and composition | Domain workflows and ports | Concrete infrastructure |
| Config representation | `internal/config` values | Converted core values | Adapter-specific values |
| Error handling | Map known outcomes; wrap unknown failures | Wrap port failures; expose core sentinels | Wrap OS/Git/plugin failures |
| Dependency direction | Imports core and adapters | Depends on interfaces | Implements core-facing behavior |
| Tests | Handler-focused | Workflow and model tests | Adapter behavior tests |

## Tooling Rules

- Format Go code with `gofmt` and run `task lint` when appropriate.
- Use `task build`, `task test`, `task quality`, and `task schema:check` as defined in `Taskfile.yml`.
- Update `mcp/easypconfig` and/or `internal/config` before regenerating schema files with `task schema:generate`.
- Do not hand-edit `schemas/easyp-config-v1.schema.json` or `schemas/easyp-config.schema.json`.
- Do not commit generated local binaries, `coverage.out`, `docs/node_modules`, secrets, or credentials.
- Dependency and vendoring conventions are documented in [config/dependency.md](./config/dependency.md); the vendor directory is `easyp_vendor`.
