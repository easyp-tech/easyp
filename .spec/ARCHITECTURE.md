<!-- generated: 2026-09-30, template: core.md -->
# EasyP Architecture

EasyP v1 separates CLI composition, module operations, generation preparation, and execution. The CLI entry point is <code>cmd/easyp</code>; configuration contracts live in <code>internal/config/v1</code>.

## Responsibilities and dependencies

| Component | Responsibility and owned state | Dependencies |
|-----------|--------------------------------|--------------|
| <code>internal/api</code> | Flags, process paths/environment, adapter construction, policy command orchestration, output and exit status | Module/generation operations, configuration, rules, core, concrete adapters |
| <code>internal/modules</code> | Dependency selection, lock validation, source roots and ownership, manifest edits, coordinated project-file updates | V1 models, <code>Source</code>/<code>Cache</code> contracts, metadata reader, filesystem |
| <code>internal/adapters/gitmodules</code> | Git candidates/revisions, checkout lifetime, persistent object cache and locking, tracked-file hashes and installation | System Git, <code>module_config</code>, module contracts, filesystem |
| <code>internal/adapters/module_config</code> | Adapt repository metadata into a named module and import roots | Native <code>protobuf.mod</code> parser, legacy EasyP and Buf readers |
| <code>internal/migration</code> | Preview plans, legacy conversion, integrity verification gates, backups and rollback | V1 models, module resolution, explicit migration repository, filesystem |
| <code>internal/workspace</code> | Repository boundary and ancestor/module/config discovery | Filesystem |
| <code>internal/schemagen</code> | Write versioned JSON Schemas for YAML configs and latest aliases | V1 schema builder, filesystem |
| <code>internal/generation</code> | Discover generator configs, inherit options, select modules, translate to engine options, run generation | V1 models, module source operations, core, logger |
| <code>internal/core</code> | Proto parsing/compilation, managed descriptors, plugin execution, lint and breaking engines | Explicit roots/options, rules, plugin executors and Git tree walker |
| <code>internal/fs/fs</code> | Directory walking, exclusive regular-file copy, temporary-file replacement | OS filesystem |

<code>modules</code>, <code>generation</code>, and the infrastructure adapters do not import the CLI package. Neither module operations nor generation require <code>*cli.Context</code>. The CLI resolves <code>EASYPPATH</code> and supplies a cache instance; only the Git adapter knows its on-disk layout.

## Dependency operations

<code>modules.Resolve(ctx, module, source, pins)</code> owns version selection and traversal. <code>Source.Fetch</code> returns module metadata and its reproducible lock entry. The resolver does not run Git, manage checkout paths, or install files. Repeated source/version requests reuse fetched metadata; versionless requirements retain existing pins during tidy. Output lock entries are sorted by source.

The contracts reflect actual consumers:

- <code>Source</code>: fetch revision metadata for graph resolution.
- <code>Cache</code>: install/verify a lock and read installed module metadata.
- <code>Repository</code>: source plus cache for <code>Get</code> and <code>Tidy</code>.
- <code>VersionedRepository</code>: also list versions for <code>Update</code>.

<code>Download</code> and <code>Vendor</code> need only <code>Cache</code>. The Git adapter implements these contracts. Unit tests use explicit fake answers/errors without Git; adapter and integration tests retain local repository fixtures.

~~~text
get / mod tidy / mod update
  -> read manifest and relevant pins
  -> Resolve through Source
  -> Cache.Install and cached metadata
  -> roots, collisions and unresolved-import checks
  -> compute manifest edits and lock
  -> persist project files

mod download
  -> read manifest and lock
  -> validate requirements before installation
  -> install and verify transitive cached metadata

mod vendor
  -> validate/ensure locked sources and collisions
  -> copy into temporary directory
  -> replace easyp_vendor, restoring the old directory on replacement failure
~~~

<code>protobuf.mod</code> uses native directives parsed by <code>v1.ParseModule</code>; <code>protobuf.lock</code> uses strict v1 YAML. Dependencies come from the manifest, not generation inputs. Manifest editing preserves comments and formatting. <code>writeV1ResolvedFiles</code> owns the order of manifest/lock updates and restores the original manifest when lock replacement fails. Restoration errors are returned together with the original error. Each file is replaced through a temporary file; the pair is **not** a filesystem transaction or a crash-atomic update.

## Source roots and ownership

<code>modules.ModuleSources</code> resolves roots relative to the module directory and validates their directories. <code>SourceRoots</code> carries both physical paths and module identities. Local replacements precede locked sources in <code>EnsureSources</code>; source order is preserved.

<code>EnsureLockedSources</code> may install dependencies. <code>ReadManifest</code>, <code>ReadLock</code>, <code>LocalSources</code>, and <code>CachedSources</code> do not download or rewrite project files. <code>CachedSources</code> assumes installation/hash verification has already occurred.

Generation, policy commands, and module operations share root/collision checks. <code>WalkProtoFiles</code> owns dependency-source traversal, including exclusion of hidden directories, vendored sources and nested module boundaries. Generator discovery and policy ancestor traversal remain separate: their inclusion and inheritance rules differ.

Nested protobuf module roots are rebased by <code>module_config</code> from the nested manifest directory into the checkout directory. Import names remain relative to those roots; repository/module prefixes do not become part of an import.

The Git cache lives beneath the CLI-supplied EasyP directory in <code>v1/git</code>; <code>internal/adapters/gitmodules/object_cache.go</code> maintains reusable bare object repositories with OS locks. <code>Cache.Cached</code> reads installed metadata; <code>Install</code> supplies content verification before cached metadata is trusted.

The X20 decisions for local replacements, frozen operation, and unknown-import-to-module mapping remain unresolved. Current lock-writing and vendoring paths reject local replacements, and unresolved imports are reported; do not infer an implemented frozen mode or automatic import discovery.

## Generation

~~~text
api.Generate.Action: flags + cwd + cache
  -> generation.Run(context, logger, cache, Request)
  -> discover configs / inherit generation options
  -> select sibling, workspace, replacement or locked module
  -> ensure sources and reject import collisions
  -> translate v1 config to core.Options
  -> core.New(complete options)
  -> Core.PrepareGeneration for each selected target
  -> compiled graph / managed mode / optional descriptor-export validation
  -> GenerationPlan.ExecuteInto shared GenerateBucket
  -> generated output / optional descriptor output
~~~

<code>internal/generation/config.go</code> owns translation to engine types, including output locations and managed-option priority. Managed rule slices are independently constructed for each module. <code>core.New</code> receives and copies import roots and file-module mappings; there are no follow-up <code>SetImportRoots</code>/<code>SetFileModules</code> calls.

<code>Request.WorkDir</code> controls discovery, relative descriptor output and local plugin execution. Plugin output remains relative to the generator config. Default discovery chooses the nearest ancestor generator within the workspace; repeatable <code>--project</code> selects consumers explicitly and <code>--all</code> enables recursive discovery. Generation builds no unrelated lint rules. Only generation options inherit; plugins and module selections remain local to the consumer.

## Policy and validation commands

Lint/breaking CLI adapters select producer policy and prepare shared module import roots before constructing the engine. <code>buildCore</code> in <code>internal/api/runtime.go</code> constructs only lint/breaking engines. <code>core.Options</code> carries <code>AllowCommentIgnores</code> and <code>KnownLintRules</code> per engine; comment directives are parsed and applied to findings by <code>internal/core/comment_suppressions.go</code>. Breaking comparisons use historical Git trees and the baseline-specific roots prepared by the API layer.

<code>validate-config</code> calls <code>config/v1.ValidatePath</code>; recursive file discovery and structured YAML validation belong to that package. <code>internal/schemagen</code> writes schemas from <code>v1.SchemaJSON</code>; <code>mcp/easypconfig</code> exposes the same model to MCP clients. The generated pairs are <code>schemas/easyp-v1.schema.json</code> / <code>schemas/easyp.schema.json</code>, <code>schemas/easyp.gen-v1.schema.json</code> / <code>schemas/easyp.gen.schema.json</code>, and <code>schemas/protobuf.lock-v1.schema.json</code> / <code>schemas/protobuf.lock.schema.json</code>. The native manifest is not a YAML schema target.

<code>linters.extends</code> and <code>breaking.extends</code> are reserved without policy-loading semantics, and <code>generate.packages</code> does not select packages. Breaking supports <code>FILE</code>; other categories are unsupported. Policy parsing and both policy conversions share semantic checks in <code>internal/config/v1/policy_semantics.go</code>; native policy/generator parsers also validate expanded YAML against the schema. Structured validation returns diagnostics through its own entry point.

## Migration

<code>internal/api/migrate.go</code> selects flag-only preview or the terminal wizard in <code>internal/api/migrate_interactive.go</code>. The wizard asks for directory and module identity, displays a preview, obtains separate consent for dependency/cache access when needed, then confirms applying the displayed plan. <code>--module</code> selects a noninteractive preview unless interactive mode is explicitly requested.

<code>migration.Build</code> owns legacy conversion into policy, generator, manifest and lock candidates. Dependency verification uses an explicitly supplied migration repository and is gated by <code>ResolveLock</code>; conversion never executes plugins. <code>Plan.Apply</code> rechecks observed state, stages files and byte-identical backups, and rolls back ordinary write failures. Legacy <code>easyp.lock</code> stays unchanged; native output conflicts require manual reconciliation. This is not a crash-atomic multi-file transaction.

## Verification boundaries

- Resolver, version selection, config conversion, manifest editing and collision rules: parallel table-driven unit tests.
- Git/checkouts/hash verification and filesystem replacement: temporary local repositories/directories.
- CLI: retained integration tests for get/tidy/download/update/vendor, nested modules, policies and error propagation. Cases changing environment/cwd run sequentially.
- Generation: explicit working directories and cache dependencies allow parallel execution without CLI/environment setup.

See [dependency management](config/dependency.md) for the v1 module contract and [package reference](PACKAGES.md) for file ownership.
