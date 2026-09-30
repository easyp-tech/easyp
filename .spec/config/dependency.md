# EasyP v1 Dependency Management

Source of truth: <code>internal/modules</code>, <code>internal/config/v1</code>, <code>internal/adapters/gitmodules</code> and <code>internal/adapters/module_config</code>. The architectural boundaries are documented in [ARCHITECTURE.md](../ARCHITECTURE.md).

## Declaration

Each independently required module has a native text <code>protobuf.mod</code>, parsed by
<code>ParseModule</code> in <code>internal/config/v1/module.go</code>. This is the dependency source;
<code>easyp.gen.yaml</code> selects modules for generation and does not declare dependencies
through the removed <code>generate.inputs</code> format. Example manifest:

~~~text
module github.com/acme/contracts/orders
roots proto
require (
    github.com/acme/contracts/common
    github.com/acme/types v1.2.3
)
~~~

- Exactly one <code>module</code> directive is required; it identifies the module, including its directory when nested in a repository. <code>roots</code>, <code>require</code> and <code>replace</code> accept multiline blocks or individual directives. Block entries cannot share the opening/closing line. Duplicate require/replace sources are errors.
- <code>roots</code> are relative to that manifest's directory and default to <code>.</code>.
- Each dependency is a separate <code>require</code>. A version may be omitted, a semantic version, or a full Git commit.
- Versionless requirements resolve repository HEAD on first use. Tidy preserves an existing locked commit; update refreshes it. Tags are not required for this path or explicit commits.
- Semantic versions resolve actual tags. For a nested module, that means a directory-prefixed tag such as <code>common/v1.2.3</code>; an untagged module should use an omitted version or commit.
- <code>replace &lt;module&gt; =&gt; &lt;local-path&gt;</code> supplies local source roots. Relative paths resolve from the consuming manifest directory; absolute replacement paths are used directly. Roots within a module cannot escape its directory.
- Normal project commands use v1 manifests. Legacy metadata is read by dependency compatibility adapters and the explicit <code>easyp migrate</code> preview/apply workflow. <code>easyp migrate --interactive</code> starts the migration wizard; migration never consists of simply renaming a legacy lock. See [CLI.md](../CLI.md).

Nested modules get separate lock entries even when their commits match. The Git adapter tries repository candidates and the metadata reader confirms the requested module name at the candidate directory. A root manifest does not hide named nested modules.

## Lock and cache

<code>protobuf.lock</code> is YAML with integer <code>version: 1</code> and a <code>modules</code> sequence. Each entry has <code>source</code>, <code>version</code>, <code>commit</code> and <code>hash</code>; the resolver sorts by source. Commits must be full 40- or 64-character hexadecimal hashes; <code>hash</code> is <code>h1:</code> plus a base64 SHA-256 digest. A commit-valued version must equal the commit. Duplicate sources, unknown fields and multiple YAML documents are rejected. The hash covers tracked regular-file content using Go's directory-hash algorithm. Symlinks/non-regular tracked entries are rejected by installation.

The CLI resolves <code>EASYPPATH</code> once for a command that needs the cache (default <code>$HOME/.easyp</code>). <code>gitmodules.Cache</code> owns <code>&lt;EASYPPATH&gt;/v1/git</code>, source keys, temporary checkout names and installation paths. Installed snapshots live under its <code>modules/&lt;source-key&gt;/&lt;commit&gt;</code> layout; reusable bare object stores live under <code>objects/&lt;remote-key&gt;</code>, with OS locks for concurrent fetches. Application callers request cached module metadata and its physical directory; they do not assemble cache paths.

<code>Cache.Install</code> verifies locked contents, including cache hits. Cached objects support shallow pinned-commit fetches with an advertised-history fallback; neither path substitutes HEAD for an unavailable pin. <code>Cache.Cached</code> reads installed metadata without downloading or rewriting files and requires prior verification. Git credentials and transport configuration remain the responsibility of system Git.

## Operations

| Command | Application operation | Behavior |
|---------|-----------------------|----------|
| <code>get &lt;module&gt;[@version\|@commit]</code> | <code>modules.Get</code> | Add/promote a direct requirement, resolve the graph and add transitive requirements |
| <code>mod tidy</code> | <code>modules.Tidy</code> | Resolve requirements, preserve versionless pins, validate imports and write manifest/lock |
| <code>mod download</code> | <code>modules.Download</code> | Validate the lock against the manifest before installing exact locked contents |
| <code>mod update</code> | <code>modules.Update</code> | Refresh HEAD requirements and tagged requirements within the existing major version; retain explicit commit pins |
| <code>mod vendor</code> | <code>modules.Vendor</code> | Verify locked sources and copy their import paths into <code>easyp_vendor</code> |

Tidy/get/update preserve existing manifest comments while editing requirements.
<code>Get</code> adds/promotes the requested requirement as direct and records other derived
requirements as <code>// indirect</code>. <code>Update</code> resolves from explicit direct entries,
then classifies transitive requirements imported by root proto files as direct.
<code>Tidy</code> also classifies newly added transitive entries and promotes an existing
<code>// indirect</code> entry when a root proto directly imports it. The current
<code>augmentV1ManifestRequirements</code> checks existing derived entries instead of
skipping them, preserves their version spelling, and removes only the indirect
marker. Its regression case is <code>TestTidyPromotesIndirectWithoutChangingLock</code>
in <code>internal/modules/operations_test.go</code>.
Sources: <code>internal/modules/get.go</code>, <code>internal/modules/update.go</code>,
<code>internal/modules/operations.go</code>, <code>internal/modules/manifest_requirements.go</code>.

### Local effective graph

Only replacements in the main (consuming) module apply, including to required transitive modules. Dependency manifests contribute their requirements, but their replacements are ignored. A replacement alone never introduces a dependency; unused replacements are not read or downloaded. All relative replacement paths are based on the main manifest directory, even at transitive depths. Absolute paths work directly. A native replacement manifest must declare the replaced identity; missing directories and identity mismatches are contextual errors. Legacy metadata uses the existing roots/requirements adapter without executing dependency plugins or loading dependency generation policy.

<code>EnsureEffectiveGraph</code> reuses the existing version resolver with local metadata lookup. Local modules have no synthetic version, commit, hash or lock entry. Ordinary dependency cycles are deduplicated; contradictory local directory identities fail. Needed unreplaced modules reuse matching published revisions (or versionless pins) after <code>Cache.Install</code> verification. New fork requirements resolve through the real <code>Source.Fetch</code> interface and immutable-version guard, then install verified snapshots. Selection and roots remain in memory.

While the main manifest contains any replacements, every operation preserves existing <code>protobuf.lock</code> bytes and does not create a missing lock. No local source information or fork-only transitive graph is persisted there. This is an explicit EasyP lock invariant: <code>protobuf.lock</code> is not Go's <code>go.sum</code>, and this does not claim that Go never changes <code>go.sum</code> with replacements.

- <code>mod tidy</code> validates the effective graph/imports and installs needed remote snapshots without changing the manifest or lock.
- <code>get</code> can add/promote/change the explicitly requested requirement; <code>mod update</code> refreshes unreplaced direct requirements with the existing version rules. Replaced requirements need no remote fetch or version lookup. Neither command promotes fork-only transitives or writes a lock.
- <code>mod download</code> resolves and verifies only needed unreplaced snapshots, including new fork dependencies, without downloading replaced roots or unused published-lock entries.
- Generate (including selected dependencies and both descriptor modes), lint, ls-files and breaking use the same overlay. In-repository baseline replacements are mapped into the Git snapshot; external baseline replacements remain errors. A versionless remote dependency introduced by a historical fork requires a historical lock pin; current HEAD cannot stand in for the baseline.
- <code>mod vendor</code> snapshots the effective local and remote dependency sources into <code>easyp_vendor</code> without writing the shared lock. It is a local artifact, not proof of a reproducible published graph. Remove replacements and run <code>mod tidy</code> before publishing the manifest/lock.

A pure-local graph works unpublished and offline without a lock. New remote requirements can require network access; matching published revisions and versionless pins use verified cache when available. Remote plugins retain their own network requirements. There is no new frozen flag or CI-only replacement rule in this change.

Import checking parses root sources and their reachable dependency/embedded
imports, validates portable relative import paths, and reports the importing file.
Unresolved imports fail resolution with <code>cannot resolve imports</code>;
tidy does not infer a new Git module identity from an unknown proto import.
<code>modules.ErrLockedVersionChanged</code> protects an already locked semantic version
from silently changing commit/hash; explicit versionless updates still refresh
HEAD. See <code>internal/modules/immutable_versions.go</code> and [ERRORS.md](../ERRORS.md)
for the distinction between such errors and CLI exit classification.

The resolver accepts <code>Source.Fetch</code>, independent of Git/cache. It selects the highest required semantic version, treats versionless requirements as weak constraints, and rejects incompatible exact commits (including tag/commit disagreement). It caches revision fetches within a resolution, rebuilds provisional HEAD edges when stronger requirements appear, and sorts resulting entries. <code>Update</code> additionally needs version enumeration; vendor and published download only need the cache contract; local-overlay download additionally uses Source when a new remote requirement needs resolution.

## Metadata and imports

<code>module_config</code> handles dependency metadata in these forms:

- V1 <code>protobuf.mod</code>, including named nested manifests.
- Legacy EasyP <code>protobuf.mod</code> requirements and <code>easyp.yaml</code> inputs.
- Buf v1 workspace/module configs and Buf v2 module roots.
- No config: the repository directory is the default root.

Root config detection checks existence; parsing belongs to format-specific readers. Missing optional files are normal. Invalid files and filesystem errors are returned. Buf roots take precedence over legacy EasyP roots when both formats occur; legacy requirements are still extracted. Buf registry dependencies are not automatically converted to Git identities.

A nested module root <code>proto</code> becomes <code>&lt;checkout&gt;/&lt;module-directory&gt;/proto</code>. A file below that root is imported without either physical prefix. <code>modules.SourceRoots</code> preserves module identity for managed selectors. Common import paths in different physical roots are rejected rather than silently selecting one.

Generation, lint, breaking and module operations share root/collision rules. Source traversal excludes hidden directories, vendored output and nested-module boundaries via <code>internal/core/path_helpers/v1_source.go</code>. Config discovery and policy inheritance use their own traversal rules.

## Persistence guarantees

Manifest/lock updates are coordinated by <code>writeV1ResolvedFiles</code>. If lock replacement fails after a manifest change, the original manifest is restored; any restoration failure is returned with the original error. Each individual file uses a temporary file and rename. There is no atomic transaction or crash-recovery guarantee for the pair.

Vendor output is built in a temporary directory before replacing the current output; replacement failure attempts to restore the previous vendor directory. Temporary checkout and staging directories are cleaned up by their owners.

## Tests

Pure resolver tests cover semver selection, commits, versionless pins, repeated visits, cancellation and source errors. Filesystem tests cover manifest preservation, stale-lock rejection before installation, and vendor staging. Local-Git integration tests retain nested-module, hash, tag/HEAD and command coverage. Generation tests use explicit directories/cache values instead of process-wide cwd/env changes.
