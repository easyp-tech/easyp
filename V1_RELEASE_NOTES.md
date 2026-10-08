# EasyP v1.0 — pre-release contract notes

EasyP v1.0 has no stable release. It is prepared for explicitly selected nightly
versions; the default installation remains the stable v0.17.0 release. The
implementation and the current `.spec` documents are the source of truth for
the v1 contract.

## Distribution and Go module compatibility

After the v1 branch is merged into `main`, v1 is published only with
`v1.0.0-nightly.YYYYMMDD.N` tags. The workflow marks releases as prereleases,
keeps GitHub Latest unchanged, and publishes only the exact Docker tag. It does
not publish a Homebrew formula. See [installation](README.md#v1-nightly-explicit-opt-in)
and [release operations](.spec/DEPLOYMENT.md).

The Go module path remains `github.com/easyp-tech/easyp`: v0 and v1 share a path,
and `/vN` is required only for v2 and later. Pinned v0 requirements are unchanged
by a branch merge. While v1 has only prerelease tags, `go install ...@latest` and
`go get -u` for v0 consumers select the existing stable release. Explicit nightly
requirements opt into the v1 API and configuration contract, which may break
v0 callers. Do not create a stable `v1.0.0` Git tag during the nightly period:
Go reads Git tags independently of GitHub Release flags, and that tag would
become eligible for ordinary updates. See [Go's version rules](https://go.dev/doc/modules/version-numbers)
and [version queries](https://go.dev/ref/mod#version-queries).

Go uses one version per module path in each build. A dependency that requires a
higher v1 nightly can also raise the selected EasyP version through
[minimal version selection](https://go.dev/ref/mod#minimal-version-selection).
The v0/v1 path therefore provides no API isolation between them.

An earlier pilot design RFC was used to bootstrap the v1 branch. Some of its examples and proposed defaults were intentionally superseded during implementation and review. They must not be used as compatibility requirements when they conflict with the contracts below.

## Superseded pilot RFC decisions

| Pilot RFC area | Current v1.0 contract |
|---|---|
| Module identity could be implicit | Every native `protobuf.mod` has exactly one `module <identity>` directive. Nested modules include their repository subdirectory in that identity. |
| A v2+ semantic version could use an unsuffixed module identity | Module identities follow Go major-version semantics: v0/v1 are unsuffixed and native v2+ modules use `/vN`. Verified pre-native repositories can use `+incompatible`; a native manifest cannot bypass the suffix rule. |
| Lock examples used short commits and no content hash | `protobuf.lock` stores a full 40- or 64-character Git commit and an `h1:` content hash for each locked Git module. Both identity and installed contents are verified. |
| `generate` without selectors recursively ran every generator below the working directory | Automatic generation selects the nearest ancestor `easyp.gen.yaml`. Recursive execution is explicit with `easyp generate --all`; individual projects use repeatable `--project`. |
| Plugin `version` was required for both local/bundled and remote plugins | `remote` plugins require a separate pinned semantic `version`. `name`, `path`, and `command` plugins must not declare `version`, because EasyP cannot verify that field against the selected local executable. |
| Go package layout followed the pilot source-layout examples | `options.go.package_prefix` derives Go import/output paths from the protobuf package hierarchy. This keeps independently selected modules such as `user.v1` and `order.v1` in distinct generated Go directories even when both source modules contain `v1/*.proto`. Explicit managed `go_package_prefix` overrides keep their existing buf-compatible file-path semantics. |
| Empty `breaking.categories` selected the historical checker | A v1 policy with omitted or empty `breaking.categories` defaults to the descriptor-backed `FILE` profile. Explicit `FILE`, `PACKAGE`, `WIRE_JSON`, and `WIRE` profiles remain available. |

## Dependency and lock compatibility

The native dependency format is documented in [`.spec/config/dependency.md`](.spec/config/dependency.md). In particular:

- `require` versions are separate tokens, not part of module identity.
- `replace` is a main-module local development overlay. Local replacement state is not written into the shared `protobuf.lock`.
- `--frozen` is explicit; it validates an existing manifest/lock graph and never infers frozen behavior from `CI=true`.
- `mod tidy` does not discover arbitrary Git repositories from proto import strings.
- Metadata-free declared dependencies can resolve missing import roots from an unambiguous pinned source/import map inside that same module. Consumer code only validates the resolved namespace and cannot select roots. Checked `get --import-root` and legacy Git input roots supply explicit hints; verified fallback roots are stored in the dependency's `protobuf.lock` entry and replayed in frozen/cold-cache operation. Existing native/Buf/legacy root metadata remains authoritative.
- After a declared dependency revision changes its namespace, `mod tidy` repairs uniquely verified consumer import bindings, validates the proposed source and commits source/manifest/lock together. It preserves non-import bytes and reports applied renames; ambiguous or deleted imported bindings stop before writes. Tidy uses manifest versions without upgrading them; generation selectors and SDK output paths remain separate.
- Known BSR dependencies encountered inside Git dependency metadata can be mapped through the current compatibility-snapshot resolver. Unknown BSR modules fail explicitly; no Git repository is guessed.

## Generation compatibility

The current command behavior is documented in [`.spec/CLI.md`](.spec/CLI.md):

- generation discovery is nearest-project by default and recursive only with `--all`;
- `generate.modules` selects module identities or workspace module paths; object entries add per-module paths/packages, intersected with global filters;
- `generate.packages` selects exact protobuf package names;
- `generate.paths` selects literal module-directory-relative files or directory subtrees, intersecting packages; it remains available without a module list;
- dependencies are available for imports but are not automatically generation targets;
- `with_imports` is per-plugin and does not change descriptor-export `--include_imports` semantics.
- source-relative Go output follows the effective descriptor `go_package`, including disable rules, overrides, and path markers; plugin-controlled import layouts stay intact. See [generation details](.spec/config/review-generation-and-baselines.md).

## Migration

`easyp migrate` is the supported v0-to-v1 transition path. It accepts the documented legacy compatibility metadata, preserves legacy inputs until apply succeeds, and does not silently reinterpret a pilot-RFC file as a native v1 manifest.

Legacy Git inputs become requirements and generation module selections. Custom
`root` is a checked resolution hint; `sub_directory` becomes module-local paths
after pinned source/import-name verification. Local and Git input filters keep
their own module scope. Corrections appear in the verified preview, and failures
do not write consumer manifest/lock files.

The wizard preserves legacy directory selections through literal
<code>generate.paths</code>, retaining root <code>.</code>, source locations and
import names. Paths select files or component-bounded directory subtrees and
intersect optional package selectors. Same-package build copies outside the
selected path stay out of generation. Complete packages remain a mixed-root
fallback. Unknown selectors fail before plugins, including no-plugin parents
selected with <code>--all</code>. Historical lock entries with annotated-tag spelling such as
`v0.4.0^{}` retain their original hash verification and unchanged backup bytes.
Released v0 proto archive hashes are verified before the native materialized-snapshot
hash is calculated. Archive attributes that omit selected/reachable proto sources
or change verified bytes block migration. Unused export-ignored sources outside
a verified filtered Git selection are permitted; whole-input target widening
still fails.

Internal file, directory, import-root and metadata symlinks are supported.
Logical paths keep their protobuf import names. Git targets resolve only from
the pinned tree; installed snapshots contain regular resolved bytes and work
with <code>core.symlinks=false</code>. Selected inputs cannot escape their owner,
cross undeclared nested repositories/submodules or form cycles. Invalid unused
auxiliary links are omitted. Migration verifies historical archive contents
separately; it never substitutes the new snapshot hash for a v0 input hash.

When reviewing v1.0, treat differences listed in this document as deliberate pre-release contract changes. A failing literal example from the earlier pilot RFC is not by itself a product defect if the current behavior matches this document and the linked `.spec` contract.
