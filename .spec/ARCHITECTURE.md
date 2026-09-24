# EasyP Architecture

EasyP v1 separates CLI composition, module operations, generation preparation, and execution. The CLI entry point is `cmd/easyp`; configuration contracts live in `internal/config/v1`.

## Responsibilities and dependencies

| Component | Responsibility and owned state | Dependencies |
|-----------|--------------------------------|--------------|
| `internal/api` | Flags, process paths/environment, adapter construction, policy command orchestration, output and exit status | Module/generation operations, configuration, rules, core, concrete adapters |
| `internal/modules` | Dependency selection, lock validation, source roots and ownership, manifest edits, coordinated project-file updates | V1 models, `Source`/`Cache` contracts, metadata reader, filesystem |
| `internal/adapters/gitmodules` | Git candidates/revisions, checkout lifetime, cache layout, tracked-file hashes and installation | System Git, `module_config`, module contracts, filesystem |
| `internal/adapters/module_config` | Adapt repository metadata into a named module and import roots | V1 manifest, legacy EasyP and Buf readers |
| `internal/generation` | Discover generator configs, inherit options, select modules, translate to engine options, run generation | V1 models, module source operations, core, logger |
| `internal/core` | Proto parsing/compilation, managed descriptors, plugin execution, lint and breaking engines | Explicit roots/options, rules, plugin executors and Git tree walker |
| `internal/fs/fs` | Directory walking, exclusive regular-file copy, temporary-file replacement | OS filesystem |

`modules`, `generation`, and the infrastructure adapters do not import the CLI package. Neither module operations nor generation require `*cli.Context`. The CLI resolves `EASYPPATH` and supplies a cache instance; only the Git adapter knows its on-disk layout.

## Dependency operations

`modules.Resolve(ctx, module, source, pins)` owns version selection and traversal. `Source.Fetch` returns module metadata and its reproducible lock entry. The resolver does not run Git, manage checkout paths, or install files. Each revision is fetched once; versionless requirements retain existing pins during tidy. Output lock entries are sorted by source.

The contracts reflect actual consumers:

- `Source`: fetch revision metadata for graph resolution.
- `Cache`: install/verify a lock and read installed module metadata.
- `Repository`: source plus cache for `Get` and `Tidy`.
- `VersionedRepository`: also list versions for `Update`.

`Download` and `Vendor` need only `Cache`. The Git adapter implements these contracts. Unit tests use explicit fake answers/errors without Git; adapter and integration tests retain local repository fixtures.

```text
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
```

Manifest editing preserves comments and formatting. `writeV1ResolvedFiles` owns the order of manifest/lock updates and restores the original manifest when lock replacement fails. Restoration errors are returned together with the original error. Each file is replaced through a temporary file; the pair is **not** a filesystem transaction or a crash-atomic update.

## Source roots and ownership

`modules.ModuleSources` resolves roots relative to the module directory and validates their directories. `SourceRoots` carries both physical paths and module identities. Local replacements precede locked sources in `EnsureSources`; source order is preserved.

`EnsureLockedSources` may install dependencies. `ReadManifest`, `ReadLock`, `LocalSources`, and `CachedSources` do not download or rewrite project files. `CachedSources` assumes installation/hash verification has already occurred.

Generation, policy commands, and module operations share root/collision checks. `WalkProtoFiles` owns dependency-source traversal, including exclusion of hidden directories, vendored sources and nested module boundaries. Generator discovery and policy ancestor traversal remain separate: their inclusion and inheritance rules differ.

Git submodule roots are rebased by `module_config` from the nested manifest directory into the checkout directory. Import names remain relative to those roots; repository/module prefixes do not become part of an import.

## Generation

```text
api.Generate.Action: flags + cwd + cache
  -> generation.Run(context, logger, cache, Request)
  -> discover configs / inherit generation options
  -> select sibling, workspace, replacement or locked module
  -> ensure sources and reject import collisions
  -> translate v1 config to core.Options
  -> core.New(complete options)
  -> Core.Generate
  -> compile descriptors / managed mode / plugins / generated output
```

`generation/config.go` owns translation to engine types, including output locations and managed-option priority. Managed rule slices are independently constructed for each module. `Core.New` receives and copies import roots and file-module mappings; there are no follow-up `SetImportRoots`/`SetFileModules` calls.

`Request.WorkDir` controls discovery, relative descriptor output and local plugin execution. Plugin output remains relative to the generator config. Generation no longer builds unrelated lint rules or modifies the global comment-ignore setting.

## Policy and validation commands

Lint/breaking keep their existing policy traversal and comparison semantics. Their CLI adapters prepare shared module import roots before constructing the engine. `runtime.go:buildCore` constructs only lint/breaking engines. The existing comment-ignore setting is still global to those engines; this refactor does not make concurrent policy runs independent.

`validate-config` calls `config/v1.ValidatePath`; recursive file discovery and structured YAML validation belong to that package. Schema generation remains under `mcp/easypconfig` and uses the same v1 models.

## Verification boundaries

- Resolver, version selection, config conversion, manifest editing and collision rules: parallel table-driven unit tests.
- Git/checkouts/hash verification and filesystem replacement: temporary local repositories/directories.
- CLI: retained integration tests for get/tidy/download/update/vendor, nested modules, policies and error propagation. Cases changing environment/cwd run sequentially.
- Generation: explicit working directories and cache dependencies allow parallel execution without CLI/environment setup.

See [dependency management](config/dependency.md) for the v1 module contract and [package reference](PACKAGES.md) for file ownership.
