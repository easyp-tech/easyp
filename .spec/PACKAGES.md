# EasyP Package Reference

## Application packages

| Package | Main files | Responsibility |
|---------|------------|----------------|
| `cmd/easyp` | `main.go` | CLI entry point, logger and global flags |
| `internal/api` | `get_v1.go`, `mod_v1*.go`, `generate.go` | Parse command inputs and call module/generation operations |
| `internal/api` | `lint_v1.go`, `breaking_v1.go`, `policy_v1.go`, `policy_imports.go`, `runtime.go` | Policy selection, import-root preparation, lint/breaking engine composition and output |
| `internal/api` | `validate.go`, `schema_gen.go`, `ls_files_v1.go`, `init_v1.go` | Validation presentation, schema generation, file listing and project initialization |
| `internal/modules` | `resolve.go` | Select revisions through the small `Source` contract |
| `internal/modules` | `operations.go`, `get.go`, `update.go`, `vendor.go`, `repository.go` | Independent application operations and their required source/cache contracts |
| `internal/modules` | `sources.go`, `local_sources.go`, `locked_sources.go`, `collisions.go`, `walk.go`, `imports.go` | Roots, ownership, replacements, lock validation, source walking and imports |
| `internal/modules` | `manifest_edit.go`, `manifest_requirements.go`, `project_files.go` | Preserve manifest text, classify direct/indirect requirements and coordinate persistence |
| `internal/generation` | `generate.go`, `inherit.go`, `selection.go` | Discover configs, inherit options and select modules |
| `internal/generation` | `config.go`, `module.go` | Translate config and construct a fully configured generation engine |
| `internal/core` | `core.go`, `dom.go`, `proto_info_read.go`, `lint.go` | Engine options, proto representations and lint execution |
| `internal/core` | `generate.go`, `managed_mode.go`, `generate_bucket.go`, `generate_insertion_point.go` | Descriptor compilation, managed mode, plugins and output |
| `internal/core` | `breaking_check.go`, `breaking_checker.go` | Read and compare current/historical proto models |
| `internal/rules` | `builder.go` and individual rules | Named lint groups and implementations of `core.Rule` |

`modules` and `generation` are callable with `context.Context` and explicit inputs. Core no longer owns dependency downloads or lock updates. CLI code does not know the physical Git cache layout.

## Configuration and metadata

| Package | Responsibility |
|---------|----------------|
| `internal/config/v1` | `protobuf.mod`, `protobuf.lock`, producer policy, generation config, recursive validation and YAML issues |
| `internal/config` | Shared lint/breaking/managed types and legacy EasyP parsing used for dependency metadata |
| `internal/adapters/module_config` | Select supported dependency metadata modes and adapt nested module, Buf and legacy EasyP roots/requirements |
| `internal/adapters/modfile` | Legacy `direct`/`replace` manifest parsing for dependency compatibility |
| `mcp/easypconfig` | Schema metadata/generation and the MCP config-description tool |

Generated JSON Schemas belong in `schemas/`; regenerate with `task schema:generate`, then run `task schema:check`.

## Infrastructure

| Package | Main files / contract |
|---------|-----------------------|
| `internal/adapters/gitmodules` | `cache.go`, `git.go`: cache layout and Git execution; `checkout.go`, `git_source.go`: revision/candidate selection; `download.go`, `files.go`: installation and tracked-file hashing; `identity.go`: optional Git origin identity |
| `internal/adapters/plugin` | Local, remote, built-in WASM and command executors; `Info` carries the explicit local execution directory |
| `internal/adapters/go_git` | Historical project-tree walkers for breaking checks |
| `internal/adapters/console` | Platform command execution |
| `internal/adapters/prompter` | Interactive prompting |
| `internal/fs/fs` | Core filesystem walker, exclusive regular-file copying and atomic replacement of an individual file |
| `internal/logger` | Logger contract and implementations |
| `internal/flags` | Shared CLI flags |
| `internal/version` | Build/compiler version metadata |

The Git adapter reads module formats through `module_config`; it does not parse YAML. Filesystem helpers do not interpret manifests, locks or module identities.

See [architecture](ARCHITECTURE.md) for direction of dependencies, ownership, and persistence guarantees.
