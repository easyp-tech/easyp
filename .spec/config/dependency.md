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
- Each dependency is a separate <code>require</code>. A version may be omitted, a semantic version, or a full Git commit, and is always a separate token: <code>require module v1.2.3</code>. Forms such as <code>require module@v1.2.3</code> or <code>replace module@v1.2.3 =&gt; path</code> are rejected instead of treating <code>@version</code> as part of module identity. An <code>@</code> inside a URL authority, such as <code>ssh://git@host/repo</code>, remains part of the transport URL.
- Versionless requirements resolve repository HEAD on first use. Tidy preserves an existing locked commit; update refreshes it. Tags are not required for this path or explicit commits.
- Semantic versions resolve actual tags. For a nested module, that means a directory-prefixed tag such as <code>common/v1.2.3</code>; an untagged module should use an omitted version or commit.
- <code>easyp get repo@common-protos-1_3_1</code> also accepts named Git tags. It resolves the tag and writes the full commit into the requirement and lock; subsequent frozen operations do not depend on that tag. Branch names are not named tags.
- <code>replace &lt;module&gt; =&gt; &lt;local-path&gt;</code> supplies local source roots. Relative paths resolve from the consuming manifest directory; absolute replacement paths are used directly. Roots within a module cannot escape its directory.
- Normal project commands use v1 manifests. Legacy metadata is read by dependency compatibility adapters and the explicit <code>easyp migrate</code> preview/apply workflow. <code>easyp migrate --interactive</code> starts the migration wizard; migration never consists of simply renaming a legacy lock. See [CLI.md](../CLI.md).

Nested modules get separate lock entries even when their commits match. The Git adapter tries repository candidates and the metadata reader confirms the requested module name at the candidate directory. A root manifest does not hide named nested modules.

## Major versions and Git mapping

EasyP uses Go-style semantic import versioning for module identities. Native <code>module</code>, <code>require</code>, <code>replace</code> source identities, fetched metadata and lock entries are checked; changing only the consumer manifest cannot bypass the check.

- v0 and v1 share the unsuffixed identity. From v2 onward, the final path component is <code>vN</code>, where N is a decimal integer of at least 2 without leading zeros, and the SemVer major must equal N.
- <code>/v0</code>, <code>/v1</code>, <code>/v01</code> and numeric dotted suffixes such as <code>/v2.0</code> are invalid module suffixes. Ordinary components such as <code>/2024</code> or <code>/v2beta</code> are not major suffixes. The check concerns module identities, not directories or protobuf import paths inside a module.
- <code>golang.org/x/mod/module.SplitPathVersion</code> and <code>CheckPathMajor</code> implement the suffix rules (including Go's <code>gopkg.in/name.vN</code> spelling). Absolute local Git fixture paths and explicit transport URLs remain supported; the URL host/path is checked without treating credentials or a port as a module suffix. Transport spelling remains part of identity and is not silently normalized.
- Omitting a version or using a full commit does not waive exact native manifest identity checks. These refs provide no semantic major to infer; a valid declared identity is still required.

| Requested identity | Allowed physical manifest directory | Release tag |
|---|---|---|
| <code>repo</code> at <code>v1.2.0</code> | repository root | <code>v1.2.0</code> |
| <code>repo/v2</code> at <code>v2.0.0</code> | root or <code>v2/</code> | <code>v2.0.0</code> |
| <code>repo/sub/v2</code> at <code>v2.0.0</code> | <code>sub/</code> or <code>sub/v2/</code> | <code>sub/v2.0.0</code> |

The final slash major suffix is logical: it is excluded from the repository candidate and the tag prefix. A tag alone does not establish module identity; the manifest at an allowed location must declare the exact requested source. Lightweight and annotated tags resolve to commits. A <code>v2/v2.0.0</code> tag does not release <code>repo/v2</code>.

<code>repo</code> and <code>repo/v2</code> are separate MVS, lock, cache and replacement identities. They may coexist when their protobuf import paths are distinct. EasyP never rewrites protobuf packages or imports; duplicate import paths still fail, even for identical content. Same-identity maximum-minimum selection, exact SHA/tag agreement, immutable tags, local overlays and explicit frozen checks remain in force.

The Go pre-module <code>+incompatible</code> exception is supported for an unsuffixed v2+ requirement whose exact Git revision has no native root <code>protobuf.mod</code> and no matching nested native manifest. <code>easyp get repo@v2.0.0</code> adds the marker after verifying this boundary; a manually written manifest must include it explicitly. For example, <code>repo v2.0.0+incompatible</code> resolves the root tag <code>v2.0.0</code>, never a tag literally ending in <code>+incompatible</code>. Fetch, migration, cold installation and warm cache reads verify this metadata boundary. Update preserves the marker for eligible legacy releases and rejects a selected revision that has become native. Markers on v0/v1 or on suffixed identities are rejected. An unmarked unsuffixed v2+ requirement remains invalid.

Migration does not invent a namespace, add a compatibility marker, or change a source/version. For an old unsuffixed native v2+ dependency, select a published <code>/vN</code> identity and matching version, or explicitly choose an unsuffixed v0/v1 release, then retry. A published native manifest can never qualify for the legacy exception. As in Go, an unversioned local replacement may contain a native manifest with the same unsuffixed identity even when replacing an explicit legacy <code>+incompatible</code> requirement. This overlay does not certify a published revision and cannot change the shared lock; frozen mode rejects replacements. Requirements read from older metadata are checked again when used at the actual source boundary.

See [Go major version suffixes](https://go.dev/ref/mod#major-version-suffixes) and [mapping versions to commits](https://go.dev/ref/mod#vcs-version).

## Lock and cache

<code>protobuf.lock</code> is YAML with integer <code>version: 1</code> and a <code>modules</code> sequence. Each entry has <code>source</code>, <code>version</code>, <code>commit</code> and <code>hash</code>; the resolver sorts by source. Commits must be full 40- or 64-character hexadecimal hashes; <code>hash</code> is <code>h1:</code> plus a base64 SHA-256 digest. A commit-valued version must equal the commit. Duplicate sources, unknown fields and multiple YAML documents are rejected. The hash covers the installed regular-file snapshot using Go's directory-hash algorithm. Internal file, directory/root and metadata aliases resolve from the pinned Git tree and are materialized under logical names; Git pointer targets must be relative and stay within that tree. Selected unsafe, dangling, cyclic or submodule-crossing inputs fail. Raw descendant submodules remain opaque skipped boundaries. Invalid unused auxiliary links are omitted. Buf filters select logical proto paths before hashing and installation; regular non-proto files and safe non-proto aliases remain included. The sole v1 hash policy is h1/Hash1 over this snapshot. Cache identity includes source, commit and hash. Historical v0 archive reconstruction is used only by explicit migrate input verification.

The CLI resolves <code>EASYPPATH</code> once for a command that needs the cache (default <code>$HOME/.easyp</code>). <code>gitmodules.Cache</code> owns <code>&lt;EASYPPATH&gt;/v1/git</code>, source keys, temporary checkout names and installation paths. Installed snapshots live under its <code>modules/&lt;source-key&gt;/&lt;snapshot-key&gt;</code> layout, where the snapshot key hashes commit plus content hash; reusable bare object stores live under <code>objects/&lt;remote-key&gt;</code>, with OS locks for concurrent fetches. Application callers request cached module metadata and its physical directory; they do not assemble cache paths.

<code>Cache.Install</code> verifies locked contents, including cache hits. A newly installed snapshot is checked with the full <code>h1:</code> directory hash. On Darwin/Linux, later unchanged hits use a sidecar verification stamp keyed by the expected lock hash plus an inode/ctime metadata fingerprint; any metadata change invalidates the stamp and forces the full directory hash again. Content changes therefore remain detectable even when size and mtime are restored. Platforms without a strong change token keep the full-hash path. The stamp is an optimization of the per-user local cache, not an additional source of dependency identity. Cached objects support shallow pinned-commit fetches with an advertised-history fallback; neither path substitutes HEAD for an unavailable pin. <code>Cache.Cached</code> reads installed metadata without downloading or rewriting module contents and requires prior verification. Git credentials and transport configuration remain the responsibility of system Git.

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

A pure-local graph works unpublished and offline without a lock. New remote requirements can require network access; matching published revisions and versionless pins use verified cache when available. Remote plugins retain their own network requirements. These overlay rules apply outside explicit frozen mode; <code>CI=true</code> alone does not change them.

Import checking parses root sources and their reachable dependency/embedded
imports, validates portable relative import paths, and reports the importing file.
Unresolved imports fail resolution with <code>cannot resolve imports</code>;
tidy does not infer a new Git module identity from an unknown proto import.
<code>modules.ErrLockedVersionChanged</code> protects an already locked semantic version
from silently changing commit/hash; explicit versionless updates still refresh
HEAD. See <code>internal/modules/immutable_versions.go</code> and [ERRORS.md](../ERRORS.md)
for the distinction between such errors and CLI exit classification.

The resolver accepts <code>Source.Fetch</code>, independent of Git/cache. It selects the highest required semantic version, treats versionless requirements as weak constraints, and rejects incompatible exact commits (including tag/commit disagreement). It caches revision fetches within a resolution, rebuilds provisional HEAD edges when stronger requirements appear, and sorts resulting entries. <code>Update</code> additionally needs version enumeration; vendor and published download only need the cache contract; local-overlay download additionally uses Source when a new remote requirement needs resolution.

## Frozen graph validation

Explicit <code>--frozen</code> requires <code>protobuf.mod</code> and <code>protobuf.lock</code> for each selected native dependency graph, even for a module with no dependencies. An empty graph must have an empty lock. A plain source tree without a manifest fails with a missing-manifest diagnostic. Frozen mode is never inferred from CI environment variables. See [CLI flag placement](../CLI.md#frozen-dependency-mode).

All root replacement directives are forbidden, even when unused. Validation rejects them before opening replacement directories or accessing dependency caches/remotes. Dependency manifests still supply transitive requirements; dependency-local replacements remain ignored. The locked closure must satisfy direct and transitive requirements with no missing or stale entries. Malformed locks, incompatible versions/commits and unexpected graph entries fail instead of triggering resolution.

The frozen path does not call the version resolver, query HEAD or tags, or create/rewrite manifests and locks. Exact locked commits may be downloaded when absent from cache; installed sources are lock-hash verified on cold paths, while unchanged Darwin/Linux warm hits reuse the verified metadata stamp and fall back to a full <code>h1:</code> check after any detected filesystem change. This permits reproducible dependency selection with a cold cache and does not imply offline execution. Existing Git credentials, transport configuration, and plugin network requirements still apply.

- <code>mod download</code> validates and installs the locked graph; <code>mod vendor</code> validates it before replacing vendor output.
- <code>mod tidy</code>, <code>get</code>, <code>mod update</code>, and <code>init</code> refuse frozen operation. <code>migrate</code>, including preview and interactive mode, is also rejected.
- Generation preflights every selected graph before any plugin runs. Explicit workspace-relative module paths use the selected module's manifest and lock; a generator-only directory needs no separate pair. Dependency identity selectors require the consumer manifest and lock, without requiring lock files inside downloaded dependencies. Unrelated projects are ignored and existing explicit/recursive selection rules remain unchanged.
- Lint retains effective policy/module grouping. Breaking uses the historical manifest and lock for its baseline, independently from the current module. Local-only <code>ls-files</code> output still requires frozen graph validation.

Run normal <code>mod tidy</code> after removing local replacements to prepare the shared lock, then commit the manifest and lock. Frozen success never certifies unpublished local overlays, and normal replacement/vendor semantics above remain unchanged.

## Metadata and imports

<code>module_config</code> handles dependency metadata in these forms:

- V1 <code>protobuf.mod</code>, including named nested manifests.
- Legacy EasyP <code>protobuf.mod</code> requirements and <code>easyp.yaml</code> inputs.
- Buf v1 workspace/module configs and Buf v2 module roots.
- No config: the repository directory is the default root.

Root config detection checks existence; parsing belongs to format-specific readers. Missing optional files are normal. Invalid files and filesystem errors are returned. Buf roots take precedence over legacy EasyP roots when both formats occur; legacy requirements are still extracted. Buf v2 <code>modules[].includes/excludes</code> and v1/v1beta1 <code>build.excludes</code> filter proto files without changing their import paths. All paths are relative to the declaring <code>buf.yaml</code>; nested v1 workspace configs are rebased to the repository. A file selected by any of the Buf modules remains available.

Buf <code>deps</code> in Git dependencies are passed to <code>modules.BSRResolver</code>. The current <code>bsr.StaticResolver</code> uses an explicit table of fixed Git compatibility snapshots. Its output is the existing <code>v1.Requirement</code>; the dependency graph, Git acquisition, import checks and indirect requirement editing keep their existing paths. Native <code>protobuf.mod</code> takes precedence over Buf metadata. Direct BSR requirements in a consumer's <code>protobuf.mod</code> are not supported.

### BSR compatibility snapshots

The Buf adapter reads v1/v1beta1/v2 <code>buf.yaml</code> and optional <code>buf.lock</code>, including configs in v1 workspaces. It preserves the BSR identity, requested reference, locked BSR commit/digest and declaring config path. All locked transitive BSR dependencies participate. Identical declarations are deduplicated; conflicting references, invalid pins and declarations missing from an existing Buf lock fail before manifest/lock updates. Declared dependencies are resolved even when their imports are unused. Unknown BSR modules produce <code>unsupported BSR module: ...; no mapping in the BSR resolver</code>; no repository is guessed.

| BSR module | Git module | Fixed Git revision |
|---|---|---|
| <code>buf.build/googleapis/googleapis</code> | <code>github.com/googleapis/googleapis</code> | <code>03a91044136a014466d4293eb1fe91f2b02075d2</code> |
| <code>buf.build/grpc-ecosystem/grpc-gateway</code> | <code>github.com/grpc-ecosystem/grpc-gateway</code> | <code>v2.31.0+incompatible</code> |
| <code>buf.build/mercari/grpc-federation</code> | <code>github.com/mercari/grpc-federation</code> | <code>v1.27.0</code> |
| <code>buf.build/envoyproxy/protoc-gen-validate</code> | <code>github.com/envoyproxy/protoc-gen-validate</code> | <code>v1.3.3</code> |
| <code>buf.build/bufbuild/protovalidate</code> | <code>github.com/bufbuild/protovalidate</code> | <code>v1.2.2</code> |

grpc-gateway v2.26.1/v2.31.0 and grpc-federation v1.24.0/v1.27.0 declare only the Googleapis BSR dependency. Their BSR commits differ, and this backend deliberately selects the same tested Googleapis snapshot. It does **not** prove equivalence with a BSR reference or verify the BSR digest. The CLI reports this limitation when resolving; the parent Git entry records it as <code>resolution: compatibility_snapshot</code>:

~~~yaml
# Inside the requiring Git module's protobuf.lock entry:
bsr:
  - dependency:
      module: buf.build/googleapis/googleapis
      commit: 62f35d8aed1149c291d606d958a7ce32
      config: buf.yaml
    git:
      module: github.com/googleapis/googleapis
      version: 03a91044136a014466d4293eb1fe91f2b02075d2
    resolution: compatibility_snapshot
~~~

The usual Git lock entry and verified <code>h1:</code> hash still pin the installed source. Frozen and cached operations verify the original Buf metadata against the recorded bindings and reuse their Git targets without calling the BSR backend. Locks created before this feature must be regenerated with <code>easyp mod tidy</code> if a dependency contains BSR declarations. Git tags remain subject to the existing immutable-version guard.

The backend is wired in [internal/api/mod_v1_cache.go](../../internal/api/mod_v1_cache.go) through <code>gitmodules.NewWithBSRResolver</code>. A future EasyP Service implementation replaces <code>bsr.StaticResolver</code> there and implements the same <code>modules.BSRResolver</code>; it need not change the dependency graph. The temporary backend has no BSR API requests, Buf executable dependency, HTTP client, implicit latest or universal BSR-to-Git inference. Other BSR revisions can require features absent from the compatibility snapshot. Separate manifests that require incompatible exact Git commits retain the existing conflict diagnostic.

A nested module root <code>proto</code> becomes <code>&lt;checkout&gt;/&lt;module-directory&gt;/proto</code>. A file below that root is imported without either physical prefix. <code>modules.SourceRoots</code> preserves module identity for managed selectors. Common import paths in different physical roots are rejected rather than silently selecting one.

Generation, lint, breaking and module operations share root/collision rules. Source traversal excludes hidden directories, vendored output and nested-module boundaries via <code>internal/core/path_helpers/v1_source.go</code>. Config discovery and policy inheritance use their own traversal rules.

## Persistence guarantees

Manifest/lock updates are coordinated by <code>writeV1ResolvedFiles</code>. If lock replacement fails after a manifest change, the original manifest is restored; any restoration failure is returned with the original error. Each individual file uses a temporary file and rename. There is no atomic transaction or crash-recovery guarantee for the pair.

Vendor output is built in a temporary directory before replacing the current output; replacement failure attempts to restore the previous vendor directory. Temporary checkout and staging directories are cleaned up by their owners.

## Tests

Pure resolver tests cover semver selection, commits, versionless pins, repeated visits, cancellation and source errors. Filesystem tests cover manifest preservation, stale-lock rejection before installation, and vendor staging. Local-Git integration tests retain nested-module, hash, tag/HEAD and command coverage. Generation tests use explicit directories/cache values instead of process-wide cwd/env changes.

## Producer-policy dependencies

Shared policies use the same declared identities and exact locked snapshots. <code>extends</code> never infers a Git repository or duplicates a dependency version. See [policy-extends](policy-extends.md). Validation only inspects already cached, verified contents; execution can install an existing locked commit.
