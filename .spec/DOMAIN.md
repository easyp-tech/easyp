<!-- generated: 2026-09-30, template: core.md -->
# EasyP Domain Model

EasyP operates on Protocol Buffers source trees, lint findings, generation targets and plugins, and Git-backed dependency modules. Engine types live in <code>internal/core</code>; native configuration and dependency records live in <code>internal/config/v1</code>; <code>internal/modules</code>, <code>internal/generation</code>, and <code>internal/migration</code> own their application workflows.

## Core Proto Entities

### <code>ProtoInfo</code>

Defined in <code>internal/core/dom.go</code>. It is the parsed view of one protobuf source file and the imported files available to it.

~~~go
type ProtoInfo struct {
	Path                 string
	ImportPath           string
	Info                 *unordered.Proto
	ProtoFilesFromImport map[ImportPath]*unordered.Proto
}
~~~

- <code>Path</code> is the source file path.
- <code>ImportPath</code> is relative to a declared source root, independently of the scan root.
- <code>Info</code> is the parsed protobuf syntax tree.
- <code>ProtoFilesFromImport</code> maps import paths to parsed imported files.

### <code>Issue</code> and <code>IssueInfo</code>

Defined in <code>internal/core/dom.go</code>. Linting and breaking checks emit these values.

~~~go
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
~~~

- <code>Position</code> identifies the protobuf source location.
- <code>SourceName</code> records the checked source.
- <code>Message</code> is supplied by the rule or checker.
- <code>RuleName</code> is the upper-snake-case rule identity.
- <code>Path</code> is added when the issue is associated with a walked file.

### <code>ProtoData</code> and <code>Collection</code>

Breaking checks collect parsed files into a package-keyed representation.

~~~go
type Collection struct {
	Imports  map[ImportPath]Import
	Services map[string]Service
	Messages map[string]Message
	OneOfs   map[string]OneOf
	Enums    map[string]Enum
}

type ProtoData map[PackageName]*Collection
~~~

Each collection groups proto declarations for one <code>PackageName</code>. Nested messages, oneofs, and enums retain a dotted parent path such as <code>MainMessage.NestedMessage</code>.

### Declaration Wrappers

<code>internal/core/dom.go</code> wraps parser nodes with their source context:

| Type | Context retained |
|------|------------------|
| <code>Import</code> | Proto file path, package name, and parsed import. |
| <code>Service</code> | Proto file path, package name, and parsed service. |
| <code>Message</code> | Message path, proto file path, package name, and parsed message. |
| <code>OneOf</code> | Oneof path, proto file path, package name, and parsed oneof. |
| <code>Enum</code> | Enum path, proto file path, package name, and parsed enum. |

## Rule Model

### <code>Rule</code>

Every lint rule implements the core interface:

~~~go
type Rule interface {
	Message() string
	Validate(ProtoInfo) ([]Issue, error)
}
~~~

<code>internal/rules/builder.go</code> creates engine rules from <code>config.LintConfig</code> (<code>Use</code>, <code>Except</code>, <code>IgnoreOnly</code>). Producer policy uses <code>linters.default</code>, <code>linters.enable</code>, <code>linters.disable</code>, <code>linters-settings</code>, and <code>issues.exclude-rules</code> in <code>easyp.yaml</code>; <code>v1.Policy.LintConfig</code> performs the conversion. Native presets are <code>MINIMAL</code>, <code>BASIC</code>, <code>STANDARD</code> (the default), and <code>COMMENTS</code>. <code>STANDARD</code> expands the internal <code>MINIMAL</code>, <code>BASIC</code>, and <code>DEFAULT</code> groups. The engine also defines the <code>UNARY_RPC</code> group. Native enable/disable and exclusion selectors are validated through <code>rules.ValidateNames</code>, which accepts registered rules/groups, including <code>UNARY_RPC</code>; it is not a <code>linters.default</code> preset.

### Ignore Model

The engine receives <code>Ignore</code>, <code>IgnoreOnly</code>, <code>AllowCommentIgnores</code>, and <code>KnownLintRules</code> through <code>core.Options</code>. <code>AllowCommentIgnores</code> is per engine, defaults to true when converting native policy, and is controlled by <code>linters.allow_comment_ignores</code>.

<code>internal/core/comment_suppressions.go</code> parses actual parser comments. It supports <code>easyp:disable</code> / <code>easyp:enable</code> regions and declaration-scoped suppressions, plus the legacy <code>nolint:</code> and <code>buf:lint:ignore</code> forms. Malformed directives, unknown rules, and unmatched enables produce errors. <code>issues.exclude-rules</code> is the native path/rule suppression model; engine fields are not interchangeable with native YAML keys.

## Module and Dependency Entities

### Native Manifest

Defined in <code>internal/config/v1/module.go</code>:

~~~go
type Module struct {
    Name     string
    Roots    []string
    Requires []Requirement
    Replaces []Replacement
}

type Requirement struct {
    Module  string
    Version string
}

type Replacement struct {
    Module string
    Target string
}
~~~

<code>Module.Name</code> is the module identity, and <code>Roots</code> are import roots relative to its directory. <code>ParseModule</code> reads native <code>protobuf.mod</code> directives (<code>module</code>, <code>roots</code>, <code>require</code>, <code>replace</code>); this file is not YAML and has no JSON Schema. Roots default to <code>.</code>. The parser handles token-boundary <code>#</code> / <code>//</code> comments, optional first-line BOM, and multiline blocks; it rejects duplicate requirements/replacements, unknown directives, inline blocks, and roots that lexically leave the module. Filesystem checks occur later in <code>modules.ModuleSources</code>.

<code>Requirement.Version</code> is an optional token. Resolution accepts semantic versions and full Git commit IDs; an omitted version is a weak requirement. <code>modules.Resolve</code> chooses the highest semantic minimum, reconciles exact commit constraints, and uses a supplied existing pin for versionless dependencies when available. It rejects incompatible tag/commit constraints. An omitted version has no separate enum or pseudo-version model.

<code>Replacement.Target</code> names a local directory. Relative targets are resolved from the module directory; dependency roots come from the replacement's native manifest. Main-module replacements apply throughout the effective graph; replacements declared by dependencies are ignored. <code>EffectiveGraph</code> keeps local metadata separate from real remote revisions. Local operations preserve the shared lock; only explicit get/update requests can change root manifest requirements. Vendor can copy this effective graph without changing the lock. Frozen mode explicitly verifies existing published graphs; automatic unknown-import discovery is intentionally excluded.

### Native Lock

Defined in <code>internal/config/v1/lock.go</code>:

| Type / field | Go type | YAML key |
|--------------|---------|----------|
| <code>Lock.Version</code> | <code>int</code> | <code>version</code> |
| <code>Lock.Modules</code> | <code>[]LockedModule</code> | <code>modules</code> |
| <code>LockedModule.Source</code> | <code>string</code> | <code>source</code> |
| <code>LockedModule.Version</code> | <code>string</code> | <code>version</code> |
| <code>LockedModule.Commit</code> | <code>string</code> | <code>commit</code> |
| <code>LockedModule.Hash</code> | <code>string</code> | <code>hash</code> |

<code>protobuf.lock</code> is one strict YAML document with numeric <code>version: 1</code>. Every locked source is unique. <code>Version</code> must be a semantic version or a full Git commit; <code>Commit</code> is a 40- or 64-character hexadecimal ID. A commit-valued version must equal <code>Commit</code> ignoring case. <code>Hash</code> is <code>h1:</code> plus a base64-encoded 32-byte digest. Parsing validates the document; installation verifies the actual Git-backed contents against it. A legacy <code>easyp.lock</code> is a different format and must be migrated with integrity verification, not renamed.

### Resolved Metadata and Source Ownership

Defined in <code>internal/modules/resolve.go</code> and <code>internal/modules/sources.go</code>:

~~~go
type Fetched struct {
    Module v1.Module
    Lock   v1.LockedModule
}

type SourceRoot struct {
    Path   string
    Module string
}

type SourceRoots []SourceRoot
~~~

<code>Source.Fetch</code> returns one revision's metadata and lock record. <code>Cache.Install</code> verifies/installs locked contents, while <code>Cache.Cached</code> returns an installed directory and <code>v1.Module</code> metadata without re-verifying content. <code>SourceRoots.Paths</code> preserves order; <code>FileModules</code> maps root-relative proto import names to identities for managed selectors after callers check collisions.

<code>internal/adapters/module_config</code> adapts native manifests and supported legacy EasyP/Buf dependency metadata to <code>v1.Module</code>. <code>internal/adapters/gitmodules</code> owns checkout, object-cache and installation details; the module application layer does not expose a legacy installed-module model. See [config/dependency.md](./config/dependency.md) for resolution, integrity and installation details.

## Generation Model

### Engine Plugins and Inputs

Core generation receives already-converted values:

~~~go
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
~~~

Plugin execution source selection is command first, then remote, then a built-in plugin absent from <code>PATH</code>, and otherwise local execution.

~~~go
type InputFilesDir struct {
    Path string
    Root string
}

type Inputs struct {
    InputFilesDir []InputFilesDir
}
~~~

These are engine inputs assembled from selected modules and their declared roots by <code>internal/generation/config.go</code>; there is no native <code>generate.inputs</code> field or engine Git-input collection. Dependencies are declared in <code>protobuf.mod</code>, and generation selection uses <code>generate.modules</code> in <code>easyp.gen.yaml</code>.

<code>core.GenerationPlan</code> holds a compiled graph after managed options are applied. <code>Core.PrepareGeneration</code> creates it without running plugins; <code>DescriptorSet</code> returns a copied descriptor set; <code>ExecuteInto</code> stages plugin output in a shared <code>GenerateBucket</code>. <code>internal/generation/descriptor_set.go</code> prepares selected targets before plugin execution and validates descriptor export plans; <code>GenerationPlan</code> replaces the old <code>Query</code> description.

### Managed Mode

The configuration model contains:

~~~go
type ManagedMode struct {
	Enabled bool
	Disable []ManagedDisableRule
	Override []ManagedOverrideRule
}
~~~

Disable rules can match a module, protobuf package, path, file option, field option, or field. Override rules set a file or field option value and can be scoped by module, package, path, or field.

The configuration validator requires valid option combinations: a rule cannot use both file and field options, and a field scope requires a field option.

## Configuration Model

Native v1 separates producer policy, consumer generation, module requirements and the lock:

| File | Type / parser | Role |
|------|---------------|------|
| <code>easyp.yaml</code> | <code>v1.Policy</code> / <code>ParsePolicy</code> | Linter selection/settings, issue exclusions, breaking policy. |
| <code>easyp.gen.yaml</code> | <code>v1.Generate</code> / <code>ParseGenerate</code> | Module selection, plugins and generation options. |
| <code>protobuf.mod</code> | <code>v1.Module</code> / <code>ParseModule</code> | Module identity, roots, native requirements and replacements. |
| <code>protobuf.lock</code> | <code>v1.Lock</code> / <code>ParseLock</code> | Reproducible dependency versions, commits and content hashes. |

<code>v1.Policy</code> has <code>Version</code>, <code>Linters</code>, <code>LinterSettings</code>, <code>Issues</code>, and <code>Breaking</code>. <code>internal/config/v1/policy.go</code> defines their concrete structures and translates policy to engine configuration. <code>Config</code> in <code>internal/config/config.go</code> remains an internal policy-engine carrier and contains shared/legacy configuration types; it is not the native user-facing file model.

The generation structures in <code>internal/config/v1/generate.go</code> are:

| Type / field | Go type | YAML key |
|--------------|---------|----------|
| <code>Generate.Version</code> | <code>string</code> | <code>version</code> |
| <code>Generate.InheritedGoPackagePrefix</code> | <code>bool</code> | <code>-</code> (excluded from YAML) |
| <code>Generate.Generate</code> | <code>GenerateTargets</code> | <code>generate</code> |
| <code>Generate.Plugins</code> | <code>[]Plugin</code> | <code>plugins</code> |
| <code>Generate.Options</code> | <code>GenerateOptions</code> | <code>options</code> |
| <code>GenerateTargets.Modules</code> | <code>[]string</code> | <code>modules</code> |
| <code>GenerateTargets.Packages</code> | <code>[]string</code> | <code>packages</code> |
| <code>GenerateTargets.Managed</code> | <code>config.ManagedMode</code> | <code>managed</code> |
| <code>GenerateOptions.Go</code> | <code>GoOptions</code> | <code>go</code> |
| <code>GoOptions.PackagePrefix</code> | <code>*string</code> | <code>package_prefix</code> |

A nil <code>PackagePrefix</code> allows nearest-ancestor inheritance; an explicit empty string stops inheritance. Plugins and generation targets remain consumer-owned. The native <code>v1.Plugin</code> has <code>Name</code>, <code>Path</code>, <code>Command</code>, <code>Remote</code>, <code>Version</code>, <code>Out</code>, <code>Opts</code>, and <code>WithImports</code> fields. Exactly one source must be set; remote plugins require a pinned semantic version in the separate <code>version</code> field. Native plugin options use <code>v1.PluginOptions</code> and are converted to the engine's map of string slices.

Native policy/generation YAML expands environment values through <code>internal/config/v1/environment.go</code>; the native manifest and lock do not share that expansion path. Runtime parsing and structured <code>ValidatePath</code> diagnostics are distinct entry points: do not assume that accepting a field in a schema means its runtime feature is implemented.

Current configuration capabilities are explicit: <code>generate.packages</code> selects exact protobuf package names; <code>linters.extends</code> and <code>breaking.extends</code> load bounded local or verified locked-module policies, with presence-aware local overrides. Breaking categories FILE/PACKAGE/WIRE_JSON/WIRE select distinct compiled-descriptor compatibility profiles; omitted categories retain legacy checks. Check the current parser, conversion methods and validation code before claiming where a reserved field fails.

<code>internal/config/v1/schema.go</code> is the schema source. <code>internal/schemagen</code> writes the six artifacts: <code>schemas/easyp-v1.schema.json</code>, <code>schemas/easyp.schema.json</code>, <code>schemas/easyp.gen-v1.schema.json</code>, <code>schemas/easyp.gen.schema.json</code>, <code>schemas/protobuf.lock-v1.schema.json</code>, and <code>schemas/protobuf.lock.schema.json</code>. <code>mcp/easypconfig</code> consumes the same v1 schema. Regenerate artifacts instead of editing JSON directly.

## Migration Model

<code>internal/migration/migration.go</code> defines <code>Options</code> (<code>Dir</code>, <code>Module</code>, <code>ResolveLock</code>, <code>Repository</code>), candidate <code>Output</code> (<code>Name</code>, <code>Content</code>, <code>Mode</code>, <code>Unchanged</code>), and <code>Plan</code>. <code>Build</code> prepares concrete candidates, <code>Outputs</code> returns copies for review, <code>CheckUnchanged</code> rechecks inputs, and <code>Apply</code> applies the prepared plan. It does not run generation plugins or expand variable placeholders.

The CLI migration wizard in <code>internal/api/migrate_interactive.go</code> first builds a preview without dependency access, requests separate dependency/cache authorization when needed, then confirms applying the displayed files and backups. Flag-only mode uses an explicit module identity and preview by default. Legacy backups and rollback support do not make multiple file replacements crash-atomic.

## Business Errors

See [ERRORS.md](./ERRORS.md) for error ownership, wrapping, command mappings and retry guidance. Current examples include <code>core.ErrInvalidRule</code>, <code>core.ErrRepositoryDoesNotExist</code>, <code>core.ErrEmptyInputFiles</code>, <code>modules.ErrLockedVersionChanged</code>, and <code>v1.ErrLegacyConfiguration</code>; typed core errors include <code>OpenImportFileError</code> and <code>GitRefNotFoundError</code>. Dependency failures also use contextual wrapped errors from module, manifest and cache operations. The removed dependency-model package is not an error source.
