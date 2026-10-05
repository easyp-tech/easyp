<!-- generated: 2026-09-30, template: core.md -->
# EasyP Code Style

This document records repository conventions from <code>AGENTS.md</code>, <code>.spec/agent-rules.md</code>, and existing source. Project rules take precedence over generic Go guidance.

## Layer Structure

~~~text
cmd/easyp
  -> internal/api              CLI handlers and composition
    -> internal/modules        dependency operations, consumer-owned source/cache ports
    -> internal/generation     consumer selection, engine preparation, descriptor export
    -> internal/migration      preview, conversion, verification and file application
    -> internal/core           lint, breaking, compilation and plugin engines

internal/adapters/gitmodules   Git object/install cache and revision verification
internal/adapters/module_config dependency metadata adaptation
internal/adapters/plugin       local, remote, command and bundled execution
internal/config/v1             native manifest, lock, policy, generator and schemas
internal/config                shared/legacy engine configuration types
internal/workspace             command context and discovery boundaries
internal/rules                 core.Rule implementations
internal/schemagen             generated schema artifacts
mcp/easypconfig                 MCP config descriptions backed by v1 schemas
~~~

| Layer | Convention |
|-------|------------|
| CLI/API | Implement <code>api.Handler</code>, build a <code>*cli.Command</code>, parse flags, construct adapters and map known CLI outcomes. |
| Modules | Keep source/cache contracts in <code>internal/modules</code>; receive context and explicit roots instead of CLI context or environment lookups. |
| Generation | Convert v1 consumer config to complete <code>core.Options</code>, prepare all selected targets, then execute plugins/output. |
| Migration | Separate immutable preview candidates from explicit dependency access and application; preserve backup and input-recheck guarantees. |
| Core | Operate on explicit engine inputs and consumer-side interfaces; keep dependency resolution in the module layer. Core currently constructs concrete plugin/console executors. |
| Adapters | Implement Git, cache, filesystem, plugin, console, metadata and prompt mechanics. |
| Rules | Implement <code>core.Rule</code> in a dedicated source file with a colocated test file. |
| Config/schema | Keep native configuration parsing/validation and schema definitions in <code>internal/config/v1</code>; generate artifacts through <code>internal/schemagen</code>. |

## Public Types and Tags

Exported types have Go doc comments that begin with the type name. Use the tags required by each representation. Native YAML structs use <code>yaml</code> tags; shared legacy types may also have <code>json</code> tags. Native <code>v1.Module</code> has no YAML tags because <code>protobuf.mod</code> uses a directive parser. For example, <code>internal/config/v1/plugin.go</code> defines:

| <code>Plugin</code> field | Go type | YAML key |
|-------------------------|---------|----------|
| <code>Name</code> | <code>string</code> | <code>name</code> |
| <code>Path</code> | <code>string</code> | <code>path</code> |
| <code>Command</code> | <code>[]string</code> | <code>command</code> |
| <code>Remote</code> | <code>string</code> | <code>remote</code> |
| <code>Version</code> | <code>string</code> | <code>version</code> |
| <code>Out</code> | <code>string</code> | <code>out</code> |
| <code>Opts</code> | <code>PluginOptions</code> | <code>opts</code> |
| <code>WithImports</code> | <code>bool</code> | <code>with_imports</code> |

Preserve native field spellings: YAML includes snake_case names such as <code>with_imports</code> and hyphenated keys such as <code>linters-settings</code> and <code>exclude-rules</code>. Do not rename established keys to satisfy a generic style preference. Keep types unexported when their scope is local; exported domain/config entities and adapter implementations use exported names.

## Conversion Boundaries

Conversion belongs to the workflow that owns it. <code>buildCore</code> in <code>internal/api/runtime.go</code> converts producer lint/breaking settings into engine options. <code>prepareV1GeneratorConfig</code> in <code>internal/generation/config.go</code> converts native plugins, selected module roots, output locations and managed rules into <code>core.Options</code>:

~~~go
cfg.Plugins = append(cfg.Plugins, core.Plugin{
    Source: core.PluginSource{Name: plugin.Name, Path: plugin.Path, Command: plugin.Command, Remote: remote},
    Out:    outRel, Options: plugin.Opts, WithImports: plugin.WithImports,
})
~~~

Remote source/version are joined for the engine, and plugin output remains relative to the generator config. Module roots become <code>core.InputFilesDir</code> values. Native dependencies come from <code>protobuf.mod</code>; the engine does not receive a Git-generation-input collection. Core engine structs do not carry YAML tags for these representations. Pass complete options to <code>core.New</code>; it copies import roots, module ownership and known lint rule names.

## Naming Conventions

| Subject | Convention |
|---------|------------|
| Go source | Use lowercase, underscore-separated file names. |
| Tests | Place tests beside the implementation as <code>&lt;file&gt;_test.go</code>. |
| Lint rules | Use <code>internal/rules/&lt;rule&gt;.go</code> plus <code>&lt;rule&gt;_test.go</code>. |
| Packages | Use short lower-case package names; adapter subpackages describe a capability. |
| Interfaces | Use descriptive names without an <code>I</code> prefix. |
| Enums | Reserve the zero value with <code>_ = iota</code> when defining enums. |
| Rule names | <code>core.GetRuleName</code> converts concrete rule type names to upper snake case. |
| Struct tags | Preserve the actual public format; most fields use snake_case, with established hyphenated policy keys. |

Examples of rule-file naming include <code>rpc_pascal_case.go</code>, <code>package_defined.go</code>, and <code>enum_first_value_zero.go</code>.

## Interface Conventions

Interfaces are generally defined by the package that consumes them. <code>internal/core</code> owns <code>Rule</code>, <code>FS</code>, <code>DirWalker</code>, and <code>CurrentProjectGitWalker</code>. <code>internal/modules</code> owns <code>Source</code>, <code>Cache</code>, <code>Repository</code>, and <code>VersionedRepository</code>; <code>internal/migration</code> extends the source contract for historical verification. Plugin executors implement the adapter-level <code>Executor</code> contract in <code>internal/adapters/plugin/interface.go</code>.

~~~go
type Rule interface {
	Message() string
	Validate(ProtoInfo) ([]Issue, error)
}

type Executor interface {
	Execute(ctx context.Context, plugin Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error)
	GetName() string
}
~~~

Rules:

- Put <code>context.Context</code> first in interface and function signatures.
- Put <code>error</code> last in results.
- Do not prefix interface names with <code>I</code>.
- Prefer passing a small consumer-owned interface over a concrete adapter.
- Update colocated fakes when changing the contracts they implement. <code>Taskfile.yml</code> retains <code>task mock</code> / <code>task mocks</code>; the aggregate target still names a removed <code>core.Console</code> interface and requires an out-of-scope tooling correction before it is a reliable regeneration command.

## Error Propagation

Wrap returned errors at each failing call site using the called function or method name only:

~~~go
original, module, err := ReadManifest(root)
if err != nil {
    return fmt.Errorf("ReadManifest: %w", err)
}
~~~

~~~go
if err := modules.CheckImportCollisions(moduleDir, module.Roots, importRoots.Paths()); err != nil {
    return nil, fmt.Errorf("CheckImportCollisions: %w", err)
}
~~~

These call sites are in <code>internal/modules/get.go</code> and <code>internal/generation/module.go</code>. Preserve <code>%w</code> chains so handlers can use <code>errors.Is</code> for sentinels and <code>errors.As</code> for typed errors. Current examples include <code>core.ErrEmptyInputFiles</code>, <code>modules.ErrLockedVersionChanged</code>, <code>v1.ErrLegacyConfiguration</code>, <code>*core.OpenImportFileError</code>, and <code>*core.GitRefNotFoundError</code>.

Place errors in the package that owns their meaning; current ownership spans core, modules, v1 config, and CLI finding sentinels. Do not recreate the removed dependency-model error package or invent a central gRPC status mapper for this CLI. Command exit mappings must be read from handlers and the entry point; contextual import errors are not automatically equivalent to a typed missing-import error. See [ERRORS.md](./ERRORS.md).

Keep reassignment and error checks on separate lines; a short <code>:=</code> declaration scoped to an <code>if</code> remains allowed. Preserve call context without adding an unrelated package label to every error.

## Resource Closing

Do not use a bare <code>defer x.Close()</code>. Wrap close calls and explicitly handle or intentionally discard the returned error:

~~~go
defer func() {
	_ = file.Close()
}()
~~~

When the close failure is operationally useful, <code>Core.close</code> logs the path and error with the repository logger.

## Imports

Order imports as standard library, third-party packages, then project packages. Separate groups with blank lines:

~~~go
import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
)
~~~

Preserve this repository grouping and run <code>gofmt</code>. The current <code>.golangci.yml</code> explicitly enables <code>staticcheck</code>; it does not configure <code>gci</code>, so do not claim import grouping is enforced by that tool.

## Comments and Documentation

- Write Go comments and Godoc in English.
- Begin exported-symbol documentation with that symbol's name.
- Put explanatory comments above <code>if</code>, <code>for</code>, and <code>return</code> lines rather than inline.
- Keep generated schema JSON out of manual edits.

## Testing Organization

| Test target | Location and pattern |
|-------------|----------------------|
| A core behavior | <code>internal/core/&lt;file&gt;_test.go</code> |
| A lint rule | <code>internal/rules/&lt;rule&gt;_test.go</code> |
| An adapter | Its adapter package with <code>&lt;file&gt;_test.go</code> |
| Module source/cache fakes | <code>internal/modules/fixtures_test.go</code> |
| Generation fixtures/fakes | <code>internal/generation/fixtures_test.go</code> and neighboring tests |
| CLI integration fixtures | <code>internal/api/v1_fixtures_test.go</code> |
| Migration behavior | <code>internal/migration/apply_test.go</code>, <code>internal/migration/lock_test.go</code> and neighboring tests |

Tests use <code>testify</code>. Use <code>require</code> for fatal preconditions and <code>assert</code> for non-fatal assertions. Table-driven tests require a descriptive <code>name</code> field. Use <code>t.Parallel()</code> at the top level and inside each <code>t.Run</code> when the case is isolated, and create dependencies per case instead of sharing mutable fakes. Tests that change process environment or working directory run sequentially; prefer explicit work directories for application tests. Use <code>t.TempDir()</code> for independent filesystem state and <code>ErrorIs</code> / <code>ErrorAs</code> for error contracts.

Run relevant tests after behavior changes. The repository's comprehensive task uses race detection:

~~~sh
task test
~~~

## Logging

The CLI initializes a <code>slog</code> text handler and stores the project logger in <code>cli.App.Metadata</code>. API handlers retrieve it with <code>getLogger</code>; core and adapters receive it as a dependency.

| Layer | Typical use |
|-------|-------------|
| CLI/API | Log terminal failures before calling <code>os.Exit</code> for recognized conditions. |
| Core | Log operation start/completion, debug resolution data, and warnings for recoverable conditions. |
| Adapters | Log infrastructure-relevant actions through the injected logger. |

Use structured attributes such as <code>slog.String</code>, <code>slog.Int</code>, and <code>slog.Any</code>. Existing generation and breaking flows pass <code>context.Context</code> to logging methods. No project-wide correlation-ID convention is defined in the inspected code.

## Concurrency

Methods accept and check <code>context.Context</code> during dependency resolution, filesystem walks, compilation, and executor calls; preserve cancellation propagation. The Git object cache uses per-repository OS locks in <code>internal/adapters/gitmodules/object_cache.go</code> with platform-specific lock implementations; acquisition waits observe context cancellation. Migration <code>Plan</code> is not safe for concurrent use, and file replacement does not exclude noncooperating writers. No common server worker-pool pattern is established by these CLI workflows.

## Quick Reference

| Aspect | API | Core | Adapters |
|--------|-----|------|----------|
| Main responsibility | CLI translation and composition | Lint/breaking/compilation engines | Concrete infrastructure |
| Config representation | Native <code>internal/config/v1</code> and converted policy values | Complete <code>core.Options</code> | Adapter-specific values and v1 metadata |
| Error handling | Map known outcomes; wrap unknown failures | Wrap port failures; expose core sentinels | Wrap OS/Git/plugin failures |
| Dependency direction | Imports application packages and adapters | Uses filesystem/Git ports and plugin executors | Implements source/cache/engine-facing behavior |
| Tests | Handler-focused | Workflow and model tests | Adapter behavior tests |

## Tooling Rules

- Format Go code with <code>gofmt</code> and run <code>task lint</code> when appropriate.
- Use <code>task build</code>, <code>task test</code>, <code>task quality</code>, and <code>task schema:check</code> as defined in <code>Taskfile.yml</code>.
- Update <code>internal/config/v1</code> models/schema rules before <code>task schema:generate</code>; keep MCP field descriptions in <code>mcp/easypconfig</code> aligned.
- Do not hand-edit <code>schemas/easyp-v1.schema.json</code>, <code>schemas/easyp.schema.json</code>, <code>schemas/easyp.gen-v1.schema.json</code>, <code>schemas/easyp.gen.schema.json</code>, <code>schemas/protobuf.lock-v1.schema.json</code>, or <code>schemas/protobuf.lock.schema.json</code>.
- Do not commit generated local binaries, <code>coverage.out</code>, dependency-installation artifacts, secrets, or credentials.
- Dependency and vendoring conventions are documented in [config/dependency.md](./config/dependency.md); the vendor directory is <code>easyp_vendor</code>.
