# EasyP v1 Dependency Management

Source of truth: `internal/modules`, `internal/config/v1`, `internal/adapters/gitmodules` and `internal/adapters/module_config`. The architectural boundaries are documented in [ARCHITECTURE.md](../ARCHITECTURE.md).

## Declaration

Each independently required module has a `protobuf.mod`:

```text
module github.com/acme/contracts/orders
roots proto
require (
    github.com/acme/contracts/common
    github.com/acme/types v1.2.3
)
```

- `module` identifies the module, including its directory when nested in a repository.
- `roots` are relative to that manifest's directory and default to `.`.
- Each dependency is a separate `require`. A version may be omitted, a semantic version, or a full Git commit.
- Versionless requirements resolve repository HEAD on first use. Tidy preserves an existing locked commit; update refreshes it. Tags are not required for this path or explicit commits.
- Semantic versions resolve actual tags. For a nested module, that means a directory-prefixed tag such as `common/v1.2.3`; an untagged module should use an omitted version or commit.
- `replace <module> => <local-path>` supplies local source roots. The replacement path is relative to the consuming module.
- Commands use v1 manifests. Legacy formats are supported only when adapting dependencies.

Nested modules get separate lock entries even when their commits match. The Git adapter tries repository candidates and the metadata reader confirms the requested module name at the candidate directory. A root manifest does not hide named nested modules.

## Lock and cache

`protobuf.lock` records sorted module entries with source, selected version, exact commit and `h1:` content hash. The hash covers tracked regular-file content using Go's directory-hash algorithm. Symlinks/non-regular tracked entries are rejected by installation.

The CLI resolves `EASYPPATH` once for a command that needs the cache (default `$HOME/.easyp`). `gitmodules.Cache` owns `<EASYPPATH>/v1/git`, source keys, temporary checkout names and installation paths. Application callers request cached module metadata and its physical directory; they do not assemble cache paths.

`Cache.Install` verifies locked contents. `Cache.Cached` reads installed metadata without downloading or rewriting files and requires prior verification. Git credentials and transport configuration remain the responsibility of system Git.

## Operations

| Command | Application operation | Behavior |
|---------|-----------------------|----------|
| `get <module>[@version\|@commit]` | `modules.Get` | Add/promote a direct requirement, resolve the graph and add transitive requirements |
| `mod tidy` | `modules.Tidy` | Resolve requirements, preserve versionless pins, validate imports and write manifest/lock |
| `mod download` | `modules.Download` | Validate the lock against the manifest before installing exact locked contents |
| `mod update` | `modules.Update` | Refresh HEAD requirements and tagged requirements within the existing major version; retain explicit commit pins |
| `mod vendor` | `modules.Vendor` | Verify locked sources and copy their import paths into `easyp_vendor` |

Tidy/get/update keep existing manifest comments. Tidy/update classify imported transitive dependencies as direct; other transitive entries carry `// indirect`. Get adds the requested dependency as direct and its transitive dependencies as indirect. Lock-writing operations reject local replacements because they cannot produce a reproducible remote lock from them. Generation and policy imports can use local replacements.

The resolver accepts `Source.Fetch`, independent of Git/cache. It selects the highest required semantic version, rejects conflicting commit requirements, visits each revision once and sorts resulting entries. `Update` additionally needs version enumeration; download/vendor only need the cache contract.

## Metadata and imports

`module_config` handles dependency metadata in these forms:

- V1 `protobuf.mod`, including named nested manifests.
- Legacy EasyP `protobuf.mod` requirements and `easyp.yaml` inputs.
- Buf v1 workspace/module configs and Buf v2 module roots.
- No config: the repository directory is the default root.

Root config detection checks existence; parsing belongs to format-specific readers. Missing optional files are normal. Invalid files and filesystem errors are returned. Buf roots take precedence over legacy EasyP roots when both formats occur; legacy requirements are still extracted. Buf registry dependencies are not automatically converted to Git identities.

A nested module root `proto` becomes `<checkout>/<module-directory>/proto`. A file below that root is imported without either physical prefix. `modules.SourceRoots` preserves module identity for managed selectors. Common import paths in different physical roots are rejected rather than silently selecting one.

Generation, lint, breaking and module operations share root/collision rules. Source traversal excludes hidden directories, vendored output and nested-module boundaries. Config discovery and policy inheritance use their own traversal rules.

## Persistence guarantees

Manifest/lock updates are coordinated by `writeV1ResolvedFiles`. If lock replacement fails after a manifest change, the original manifest is restored; any restoration failure is returned with the original error. Each individual file uses a temporary file and rename. There is no atomic transaction or crash-recovery guarantee for the pair.

Vendor output is built in a temporary directory before replacing the current output; replacement failure attempts to restore the previous vendor directory. Temporary checkout and staging directories are cleaned up by their owners.

## Tests

Pure resolver tests cover semver selection, commits, versionless pins, repeated visits, cancellation and source errors. Filesystem tests cover manifest preservation, stale-lock rejection before installation, and vendor staging. Local-Git integration tests retain nested-module, hash, tag/HEAD and command coverage. Generation tests use explicit directories/cache values instead of process-wide cwd/env changes.
