# Generation, baseline and reproducibility review fixes

## X-14: Go package prefix is not full managed mode

`options.go.package_prefix` applies Go option overrides without enabling
defaults for Java, C#, PHP, Ruby, Objective-C or C++. Ordinary prefixes derive
multi-segment Go import paths from the protobuf package hierarchy. This keeps
modules such as `user.v1` and `order.v1` in distinct generated directories even
when both contain `v1/*.proto` and use `paths=source_relative`. Single-segment
packages and path markers retain their existing file path semantics.

Source-relative Go output follows the final descriptor `go_package`, relative
to the configured prefix (the stable parent for marker templates). Disabled
options and explicit overrides therefore keep imports and output directories
consistent. Packages outside that prefix keep their plugin output paths.
`paths=import` and the plugin's `module` option retain plugin-controlled layout.

The internal GoPackageOnly runtime mode is not a new YAML setting. Empty
options are not materialized in otherwise unchanged descriptors. Other
languages and field-option overrides require an explicit
`generate.managed.enabled: true`.

The existing ordering remains: inherited prefixes precede explicit managed
Go overrides; a directly specified options.go.package_prefix follows them.
Go disable selectors, including go_package itself, still suppress the prefix.
Explicit full managed mode preserves its language defaults and overrides.

## X-16: revision-local module contexts for breaking

Each selected module is compared separately. Both revisions discover their own
module roots and dependencies. Baseline protobuf sources and manifests/locks
are read from an isolated Git snapshot, never from the current working tree.
The caller's checkout, manifest and lock are not modified; the verified
module cache may be populated. Snapshot directories are removed after use.

The selected paths are relative to the Git repository, so changing --root does
not change the identity of an otherwise identical checked file. Explicit target
paths take precedence over import aliases when building the checker input.
Baseline-only files and deleted modules remain checkable, including via --path.

Policy ownership is still determined from current configuration. Each effective
policy uses its own baseline; dependencies are checked with the policy of the
selected files, not an unrelated ancestor at the module directory. Repeated
identical findings are emitted once. FILE checks and whole-section policy
inheritance remain enabled. This change does not add new Git ref kinds.

Local replacements within the repository use their baseline contents. Absolute
paths are mapped back into the snapshot, including macOS filesystem aliases.
External local replacement directories have no recoverable historical identity;
these are explicitly rejected for baseline comparison instead of silently using
current contents. A locked dependency provides a reproducible baseline.

## R-20: vendor files are imports, not lint targets

Target discovery skips easyp_vendor and hidden storage/Git directories before
reading policies or constructing module contexts. A legacy or invalid policy
under vendor therefore does not break a root lint invocation. A vendor target
explicitly selected through --path is excluded too. Import resolution for user
sources continues through their module roots and verified dependencies.

Local replacement sources are also import-only during consumer lint and
breaking checks, including legacy/Buf metadata and nested modules. Their paths
are excluded before reading their policies. Selecting a replacement explicitly
through `--path` still validates its policy. Absolute in-repository targets are
rebased into the baseline snapshot during source selection as well as import
resolution. Unused replacements pointing to files cannot hide checked sources.

## X-24: a recorded version cannot silently change identity

Tidy, get and update verify every fetched previously locked semantic version
against the recorded commit and content hash. A mismatch returns
ErrLockedVersionChanged with both identities before installing the new graph or
writing protobuf.mod/protobuf.lock. This covers direct and transitive
requirements and works independently of a warm installed-module cache.

Update still refreshes versionless requirements and can select a new semantic
version; the versionless HEAD pins are not accidentally frozen by the guard.
Download remains lock-first and can retrieve the old exact commit after a tag
moves, provided that commit remains available from the source.

## X-26: documentation publication belongs to its repository

The obsolete CLI .github/workflows/docs.yml has been removed. It attempted to
build a deleted docs directory and push generated files to the default branch
from a release tag. Documentation builds and image publication are independently
owned by easyp-tech/docs and its .github/workflows/docker-publish.yml. No new
cross-repository credentials, dispatch or duplicate publishing step is added.

## Regression

Local tests cover Go-only transformation, explicit managed defaults, snapshot
isolation, module roots, import aliases, filesystem aliases, immutable version
checks and the absent obsolete workflow. The v1 easyp-test suite additionally
runs CLI scenarios for all five findings, cold/warm caches, changed baseline
dependencies, nested policies, FILE moves, deleted modules and transitive retags.
The source-only workflow guard needs EASYP_SOURCE; executable scenarios always
require an absolute EASYP_BIN. All newly observed defect cases are reproduced on
the previous CLI before validation of the fix.
