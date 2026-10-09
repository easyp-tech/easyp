# Selective Git snapshots implementation plan

**Goal:** Read, copy and hash dependency proto files and relevant configuration,
without processing unrelated source, documentation or build artifacts.

**Approved scope:** The user approved narrowing the snapshot, removing the
whole-tree digest when migration has no historical lock, and preserving arbitrary
policy-reference names and bounded aliases. Existing v1 has no stable release;
keep one hash policy and do not introduce compatibility namespaces or lock formats.

**Architecture:** Keep immutable Git path/tree metadata separate from file-body
access. Native snapshots contain module/Buf/EasyP metadata and proto files; an
optional pinned filesystem source supplies explicitly referenced policy files
through the existing policy graph. Mutable local replacements continue using
their existing filesystem and bounds. Historical whole-tree verification remains
isolated to an actual legacy digest; initial migration validates proto archive
selection and path collisions without hashing unrelated bodies.

**Tech stack:** Go 1.26.6, existing Git/go-git adapters, io/fs, sourceview, testify.
No dependency, Go version or generated-schema changes.

- [x] Add failing regressions: README/Go/build files absent from the native
      snapshot and unable to change its h1 hash; lazy directory listings do not
      decode their blobs; unused unsafe auxiliary links remain omitted.
- [x] Make Git directory entries lazy (including SHA-256 size metadata) and add
      a bounded candidate walk that skips non-selected file bodies before resolve.
      Preserve mode, gitlink, invalid-path and alias-cycle behavior.
- [x] Narrow metadata/proto materialization and native cache hashing, installation
      and inspection. Reuse the same selection rules throughout these paths.
- [x] Preserve producer policies with arbitrary names via a checked immutable
      filesystem supplied by the policy graph. Validate pinned object content,
      directory/module bounds and aliases; do not allow discovery of new modules
      or revisions. Policy validation remains read-only.
- [x] Remove whole-tree hashing without a historical digest; preserve proto
      archive/path equivalence and real historical-hash checks at the pinned SHA.
- [x] Update tests that deliberately asserted unrelated files were retained;
      add real CLI migration/generation and extends cold/warm-cache regressions.
- [x] Run relevant and full source race suites, linter, standard v1 e2e with fresh
      CLI/MCP and EASYP_SOURCE, and paired large-mirror checks. Update docs and
      source contract notes. Commit/push to the existing regression branch.

Review focus: policy fragments not reachable from producer easyp.yaml, arbitrary
extensions, logical module prefixes/aliases, SHA-256, object integrity, and
read-only validation when sources are cached. No default CLI installation or
foreign/untracked resources are changed.


## Reviewed integration contracts

- `modules.PolicyModule.Files` is an optional lazy `PolicyFiles.Read(ctx, relative)`
  capability. It returns logical `Path`, pinned canonical identity and bytes.
  The graph attaches it only after installing/verifying the reached pin and
  finding the actual manifest prefix. Local replacements keep the OS path.
- `policy.Resolver` registers sources by module boundary and carries them through
  recursive local references. It uses virtual logical paths, pinned identities
  for cycles/cache keys, and literal parsing for producer policy content.
- A source-binding sidecar maps source/commit to a cache-owned bare object store.
  Read-only requests never fetch or create files. Missing objects report explicit
  `mod download` guidance; explicit download repairs retained objects even with a warm native
  snapshot. Object IDs/parent-tree bindings are verified before trusting bodies.
- Read-only cached verification does not create verification stamps. Every stamp
  hit first checks that the snapshot contains only the current selection policy;
  old stamps cannot authorize retained unrelated files.
- Native hashing remains h1 over the selected snapshot, with no hash-version
  compatibility namespace or new lock fields. Genuine legacy whole-tree hashes
  come from raw pinned Git bytes, not the selective snapshot. Missing legacy
  hashes never trigger whole-tree or redundant archive digests.
- Generic Git directory listings are lazy; selected walks exclude regular junk
  before resolution. Classifying a link to an unrelated large regular file only
  reads pointer/header metadata, not its body. SHA-256 has the same guarantee.


## Verification notes

Ordinary Install is lazy about unused policy objects. Only explicit Download calls
RepairPolicySources. Repair validates pinned Git connectivity without hashing
unrelated blob bodies, preserves damaged stores by quarantine, then restores all
locked commits additively so a shared store retains earlier healthy pins.
Resolver pinned/load caches are scoped by consumer graph as well as module
boundary; a later local replacement cannot inherit an earlier producer reader.

Current Git archives alias.proto but not data.proto/target.txt for a '*.proto'
pathspec; this real no-lock case must remain rejected rather than pretending the
v0 source was available. A separate synthetic ZIP regression verifies selective
loading of non-proto target bodies when those entries actually exist in an old
archive. Unused archive bodies remain unread.


## Verified outcome

Go 1.26.6 and dependencies unchanged. Full source race suite passed (889 top-level
tests, 3191 test actions, no test skips/failures), followed by affected repair/lock
race regressions after the final localized locking fix. Standard independent v1
race suite passed (209 top-level tests, 969 test actions, no skips/failures) with
fresh CLI/MCP and EASYP_SOURCE. Linter zero issues.

A real consumer contract using its pinned dependency passed tidy, arbitrary-name
policy inheritance (base.rules -> strict.conf), validation, and Python generation.
File bytes and mtimes across the cache were identical before/after validation.
Changing only README changes the commit but leaves the selected h1 unchanged.

Isolated paired migration with 4000 retained unrelated files: previous full
snapshot 1.628s, selective snapshot 0.236s. The selective hash is identical at
100, 1000 and 4000 irrelevant-file counts. Old nightly whole-file hashes correctly
fail immutable-hash checks without changing the lock; preserve the old lock as a
backup and run tidy to create the current unreleased v1 snapshot policy.

Private user mirror/hour-long timing and optional live_dependencies/python_grpc
suites were not exercised for this change. No installed CLI or foreign/untracked
workspace files were modified.
