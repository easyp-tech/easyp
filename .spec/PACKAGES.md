<!-- generated: 2026-09-30, template: core.md -->
# EasyP Package Reference

## Application packages

| Package | Main files | Responsibility |
|---------|------------|----------------|
| <code>cmd/easyp</code> | <code>main.go</code> | Registered CLI commands, logger and global flags |
| <code>cmd/easyp-mcp</code> | <code>main.go</code> | Separate MCP server entry point |
| <code>internal/api</code> | <code>get_v1.go</code>, <code>mod_v1*.go</code>, <code>generate.go</code> | Parse command inputs and call module/generation operations |
| <code>internal/api</code> | <code>lint_v1.go</code>, <code>breaking_v1.go</code>, <code>policy_v1.go</code>, <code>policy_imports.go</code>, <code>runtime.go</code> | Policy selection, import-root preparation, lint/breaking engine composition and output |
| <code>internal/api</code> | <code>validate.go</code>, <code>schema_gen.go</code>, <code>ls_files_v1.go</code>, <code>init_v1.go</code> | Validation presentation, schema generation, file listing and project initialization |
| <code>internal/api</code> | <code>migrate.go</code>, <code>migrate_interactive.go</code>, <code>migrate_prompt.go</code> | Flag-only migration and terminal wizard, preview and authorization prompts |
| <code>internal/modules</code> | <code>resolve.go</code> | Select revisions through the small <code>Source</code> contract |
| <code>internal/modules</code> | <code>operations.go</code>, <code>get.go</code>, <code>update.go</code>, <code>vendor.go</code>, <code>repository.go</code> | Independent application operations and their required source/cache contracts |
| <code>internal/modules</code> | <code>sources.go</code>, <code>local_sources.go</code>, <code>locked_sources.go</code>, <code>collisions.go</code>, <code>walk.go</code>, <code>imports.go</code> | Roots, ownership, replacements, lock validation, source walking and imports |
| <code>internal/modules</code> | <code>manifest_edit.go</code>, <code>manifest_requirements.go</code>, <code>project_files.go</code> | Preserve manifest text, classify direct/indirect requirements and coordinate persistence |
| <code>internal/generation</code> | <code>generate.go</code>, <code>discovery.go</code>, <code>inherit.go</code>, <code>selection.go</code> | Discover configs, inherit options and select modules |
| <code>internal/generation</code> | <code>config.go</code>, <code>module.go</code> | Translate config and construct a fully configured generation engine |
| <code>internal/generation</code> | <code>descriptor_set.go</code>, <code>descriptor_output.go</code>, <code>descriptor_conflict.go</code> | Prepare target graphs, validate export destinations and descriptor compatibility, execute shared output bucket |
| <code>internal/migration</code> | <code>migration.go</code>, <code>legacy.go</code>, <code>convert.go</code>, <code>manifest.go</code>, <code>lock.go</code>, <code>roots.go</code>, <code>apply.go</code>, <code>validate.go</code> | Build validated migration candidates, verify historical dependency pins, preserve backups, apply with rechecks/rollback |
| <code>internal/workspace</code> | <code>discovery.go</code> | Repository boundary, ancestor config/module selection, recursive discovery exclusions |
| <code>internal/core</code> | <code>core.go</code>, <code>dom.go</code>, <code>proto_info_read.go</code>, <code>lint.go</code> | Engine options, proto representations and lint execution |
| <code>internal/core</code> | <code>generate.go</code>, <code>managed_mode.go</code>, <code>generate_bucket.go</code>, <code>generate_insertion_point.go</code> | Descriptor compilation, managed mode, plugins and output |
| <code>internal/core</code> | <code>comment_suppressions.go</code>, <code>instruction_parser.go</code>, <code>unstable_package_matcher.go</code> | Per-engine lint directives, protobuf instruction names and unstable package matching |
| <code>internal/core/path_helpers</code> | <code>is_target_path.go</code>, <code>v1_source.go</code> | Target path matching and v1 source exclusions |
| <code>internal/core</code> | <code>breaking_check.go</code>, <code>breaking_checker.go</code> | Read and compare current/historical proto models |
| <code>internal/rules</code> | <code>builder.go</code> and individual rules | Named lint groups and implementations of <code>core.Rule</code> |

<code>modules</code> and <code>generation</code> are callable with <code>context.Context</code> and explicit inputs. Core no longer owns dependency downloads or lock updates. CLI code does not know the physical Git cache layout.

## Configuration and metadata

| Package | Responsibility |
|---------|----------------|
| <code>internal/config/v1</code> | Native directive parser in <code>module.go</code>; lock/policy/generator YAML in <code>lock.go</code>, <code>policy.go</code>, <code>generate.go</code>; <code>schema.go</code>, <code>policy_semantics.go</code>, <code>semantic_validation.go</code>, recursive validation and YAML issues |
| <code>internal/config</code> | Shared engine/legacy lint, breaking, plugin and managed types; legacy dependency readers live in <code>internal/adapters/module_config</code> |
| <code>internal/adapters/module_config</code> | Select supported dependency metadata modes and adapt nested module, Buf and legacy EasyP roots/requirements |
| <code>internal/adapters/modfile</code> | Legacy <code>direct</code>/<code>replace</code> manifest parsing for dependency compatibility |
| <code>internal/schemagen</code> | <code>schemagen.go</code> writes versioned schema artifacts and latest aliases from <code>v1.SchemaJSON</code> |
| <code>mcp/easypconfig</code> | MCP config-description tool and schema exposure backed by <code>internal/config/v1</code> |

Generated JSON Schemas are <code>schemas/easyp-v1.schema.json</code>, <code>schemas/easyp.schema.json</code>, <code>schemas/easyp.gen-v1.schema.json</code>, <code>schemas/easyp.gen.schema.json</code>, <code>schemas/protobuf.lock-v1.schema.json</code>, and <code>schemas/protobuf.lock.schema.json</code>. Regenerate with <code>task schema:generate</code>, then run <code>task schema:check</code>. Native <code>protobuf.mod</code> is parsed directly and has no JSON Schema.

Section-scoped <code>extends</code> is implemented in <code>internal/policy</code>, backed by the verified consumer graph. Generation <code>packages</code> is an exact-name selector; FILE/PACKAGE/WIRE_JSON/WIRE breaking profiles are implemented in the core descriptor comparator. These struct/schema fields are not proof of runtime support. See [domain model](DOMAIN.md) and [dependency management](config/dependency.md) for current limits, including local overlays, explicit frozen validation and intentionally excluded unknown-import discovery.

## Infrastructure

| Package | Main files / contract |
|---------|-----------------------|
| <code>internal/sourceview</code> | Bounded logical resolve/open/walk over standard io/fs; local os.Root reads and alias topology checks |
| <code>internal/adapters/gitsnapshot</code> | Immutable Git tree/blob filesystem, SHA-1/SHA-256 repository support and host path collision checks |
| <code>internal/adapters/gitmodules</code> | <code>cache.go</code>, <code>git.go</code>: cache layout and Git execution; <code>object_cache.go</code>, <code>object_lock_unix.go</code>, <code>object_lock_windows.go</code>: reusable Git object repositories and OS locks; <code>checkout.go</code>, <code>git_source.go</code>: revision/candidate selection; <code>download.go</code>, <code>files.go</code>: installation and materialized snapshot hashing; <code>identity.go</code>: optional Git origin identity; <code>migration.go</code>, <code>migration_config.go</code>, <code>migration_selection.go</code>: historical revision/hash verification |
| <code>internal/adapters/plugin</code> | Local, remote, built-in WASM and command executors; <code>Info</code> carries the explicit local execution directory |
| <code>internal/adapters/go_git</code> | Historical project-tree walkers for breaking checks |
| <code>internal/adapters/console</code> | Platform command execution |
| <code>internal/adapters/prompter</code> | Interactive prompting |
| <code>internal/fs/go_git</code> | Read-only filesystem and directory walkers over historical Git trees |
| <code>internal/fs/fs</code> | Core filesystem walker, exclusive regular-file copying and atomic replacement of an individual file |
| <code>internal/logger</code> | Logger contract and implementations |
| <code>internal/flags</code> | Shared CLI flags |
| <code>internal/version</code> | Build/compiler version metadata |

Normal Git dependency reads use <code>module_config</code>. The Git adapter also has a dedicated historical YAML reader in <code>internal/adapters/gitmodules/migration_config.go</code> to reproduce legacy roots for integrity verification. Filesystem helpers do not interpret manifests, locks or module identities.

See [architecture](ARCHITECTURE.md) for direction of dependencies, ownership, and persistence guarantees.

## MCP configuration reference

<code>easyp_config_describe</code> covers all four v1 formats. <code>protobuf.mod</code> returns text grammar and examples; YAML files, including <code>protobuf.lock</code>, return their actual JSON Schemas. The reference never reads project files or resolves dependencies. See [MCP reference contract](config/mcp-module-reference.md).
