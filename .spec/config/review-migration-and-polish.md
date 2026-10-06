# Migration and remaining review corrections

## Native repository configuration (X-7)

The CLI repository uses native v1 policy, generation, manifest and lock files.
The previous api generation directory did not exist; its obsolete dependencies
and legacy root lock were removed from this repository, not from user projects.
A documented hello/v1 example uses the bundled Python generator and a well-known
Timestamp import, without external protobuf dependencies. The proto:check task
copies only real config and example inputs to a temporary directory, validates,
lints, verifies the lock and compares two generations. CI runs the same task.
Intentional negative fixtures under testdata remain unchanged.

## Explicit v0 migration (X-1, X-6, X-11)

The migrate command in flag-only mode previews real output content by default. Writing requires
--write; optional dependency resolution/cache access requires --resolve-lock.
The module identity is explicit. Migrating does not execute plugins or save
expanded environment secrets. Effective lint selection, legacy comment opt-in,
settings, literal/prefix exclusions, plugin argv/options/with_imports, supported
managed rules and representable input roots are transferred. The selected
physical files and protobuf import names must remain equivalent.

Legacy direct/indirect directives become require. Old full-SHA/pseudo-version
pins retain their commit. A tag-only pin is accepted only after verifying the
legacy installed-tree hash. The new materialized snapshot hash is calculated
independently. Missing historical pins, retags and hash mismatches fail. The old
easyp.lock remains byte-identical. Empty projects receive a native empty lock,
so the retained old lock cannot block later normal module commands.

Internal file, directory, import-root and metadata symlinks are supported.
Logical paths keep their protobuf import names. Git targets resolve only from
the pinned tree; installed snapshots contain regular resolved bytes and work
with <code>core.symlinks=false</code>. Selected inputs cannot escape their owner,
cross undeclared nested repositories/submodules or form cycles. Invalid unused
auxiliary links are omitted. Migration verifies historical archive contents
separately; it never substitutes the new snapshot hash for a v0 input hash.

Git index framing, stage and local-path validation live in the pure
internal/adapters/gitindex parser. Snapshot selection and metadata preflight
apply their own mode policies. Migration reads one index, records the selected
regular files and omitted auxiliary links, then verifies historical contents
in migration_integrity.go. FetchMigration keeps source equivalence and native
hash calculation as separate steps; no-link whole-tree proof and archive-only
proof after omitted links remain distinct.

The application transaction stages all results and .v0.bak recovery copies,
checks observed contents/permissions/source scope again, and rolls back ordinary
failures. Conflicting existing files and unsafe symlink destinations are refused.
A validated native project is an idempotent no-op. Several filesystem renames
are not a crash-atomic transaction; uncooperative concurrent writers remain
outside the guarantee. Recovery material is retained if rollback itself fails.

Conservative refusal is part of the contract, not silent success: unsupported
external/sliced inputs, custom Git-input roots, ambiguous managed selectors,
missing historical identity, full lock migration with local replacements and
Buf-config conversion may need manual migration. Native v1 execution remains
separate; detection of old configs points to migrate instead of accepting v0.

## Supported policy features and explicit reserved fields (X-5)

issues.exclude-rules.path supports relative glob segments *, ?, character classes
and whole-segment **. Literal directories cover descendants, literal proto files
match exactly. The effective issues policy's directory is the base. Whole-file
exclusions are applied before import preparation; named rules/groups only affect
matching files. A pathless empty linter list continues to explicitly suppress
all checks in its policy scope; examples document this rather than changing it
silently.

breaking.ignore_unstable filters only unstable final package-version components
on both sides. Stable declarations, imports and references remain checked.
FILE remains the only additional category. Shared extends and generate.packages
are reserved: schemas reject nonempty values rather than advertising execution.
Public English/Russian references and guides reflect supported code, explicit
project selection and plugin-specific with_imports.

## Independently confirmed smaller defects

- protobuf.mod accepts initial BOM and token-start # comments, rejects duplicate
  require/replace sources and malformed/inline blocks with source line numbers.
  Manifest editing preserves these comments, BOM, URL text and indirect markers.
- validate-config checks missing/non-directory local replacement targets.
- YAML policy and generation files require exactly one document.
- PACKAGE_DIRECTORY_MATCH uses the declared module import path, independently
  of --root, while diagnostics retain the selected user-facing path.
- breaking.ignore is bound to its policy file, not to the Git repository root.
  Baselines support branches, lightweight/annotated tags and Git commit refs;
  an existing bare branch name retains precedence over a same-named tag.
- command plugins pass executable and argv directly, without implicit shell
  string evaluation. A shell script remains possible via explicit sh -c argv.
- an explicitly selected global output format is not hidden by a local default;
  a local explicit format still wins. Duplicate filename prefixes are removed
  from configuration errors, and parent paths are not mistaken for ignored dirs.

## Still requires a separate contract

X-20 local-replacement lock semantics, an explicit frozen/reproducible mode and
unknown import-to-Git-module discovery have not been invented here. Likewise,
remote policy distribution through extends, package selectors and additional
breaking categories remain outside the supported subset. Different major
versions and registry-specific Buf dependency identity need explicit decisions.

## Verification entry points

~~~bash
go test -race -count=1 ./...
go vet ./...
task schema:check
task proto:check
~~~

The external v1 easyp-test suite has CLI regression tests for migration,
historical hash/commit preservation, preview/no-write failures and policy/path
corrections. EASYP_BIN and EASYP_SOURCE must both point at the tested build.

## Interactive migration

A terminal invocation without --module starts the line-based wizard. Explicit
--interactive requests it even with supplied flags; --interactive=false disables
it. Both application input and output must be terminals. No /dev/tty fallback
is used to reinterpret piped data as consent. Existing --module invocations
remain noninteractive unless the wizard is explicitly requested.

Missing directory and identity values are prompted and validated. Native module
metadata or local Git origin may supply a credential-free identity suggestion;
it is never accepted without user input. The initial preview requires no
repository access. Deferred dependency verification has its own default-no
confirmation, followed by a fresh complete preview and a separate default-no
write confirmation. Prefilled --write/--resolve-lock do not bypass these gates.
EOF, malformed input and unavailable output never imply consent. Already-v1
projects terminate without resolution or write questions. The reviewed plan is
applied without rebuilding it after consent; its input snapshots are rechecked.
Before and after authorized resolution, the original input state is also checked.
Cancellation after authorized resolution may retain cache entries, not project
changes. Existing application crash/concurrent-writer limits remain unchanged.
