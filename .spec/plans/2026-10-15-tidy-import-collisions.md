# Reachable duplicate imports during mod tidy

> For agentic workers: execute inline using the test-driven-development and
> verification-before-completion skills. Review the plan and final changes.

**Goal:** `mod tidy` ignores unused dependency import-path duplicates, resolves
byte-identical reachable copies as one logical proto, and rejects differing
reachable contracts before changing consumer sources, manifest or lock.

**Architecture:** Keep the strict whole-source collision check for other module
operations. Tidy owns a bounded, observed resolver for its consumer targets and
reachable import closure. Enumerate every allowed candidate for a used name,
read its verified/proposed bytes through the existing `tidySourceView`, compare
exact bytes, and retain a deterministic canonical source. Every compared copy
must participate in transaction input rechecks. Preserve producer filters,
module boundaries, aliases, exact pins and root-inference guarantees.

The plan review additionally requires used-name membership rechecks, including
candidate additions after an absent lookup; alias identity rechecks; an explicit
read-only overlay checkpoint with remote pin verification; the same canonical
root/owner in validation and compilation; and historical rename tests alongside
unused duplicate inventory. Root inference already indexes one producer at a
time, so its strict namespace checker needs no relaxation.

**Tech stack:** Go 1.26.6, existing protocompile/sourceview/transaction machinery,
standard testing and testify. No dependency, toolchain or schema changes.

## Contract and scope

- A direct and transitive requirement on the same module remains one selection.
- Unused duplicates with equal or different contents do not invalidate tidy.
- Reachable equal bytes are deduplicated by import name, including transitive,
  public/weak imports and copies reached through separate module roots.
- Reachable differing bytes remain an error identifying the import and paths.
- Consumer targets are used sources; preserve checks for conflicting target
  import names rather than silently selecting a different consumer file.
- Local replacements use the same tidy import checks but retain their existing
  validation-only behavior: no source/manifest/lock writes and no invented pins.
- Root inference and verified historical import renames retain ownership checks;
  equality never justifies guessing roots or rebinding a renamed contract.
- Failed comparison/read/parse/cancellation/recheck leaves project files intact.
- Download, get/update, vendor, generation and frozen checks retain their current
  collision contract. This task does not add consumer excludes.

## Files and responsibilities

- `internal/modules/imports.go`: extract bounded per-root candidate lookup so
  tidy can enumerate matches without changing ordinary first-match resolution.
- `internal/modules/tidy_imports.go` (new): tidy-only used-name comparison and
  deterministic logical deduplication, using observed verified source reads.
- `internal/modules/import_rewrites_validation.go`: route reachable validation
  through the tidy resolver, preserve import owners and proposed-byte compile.
- `internal/modules/tidy_operation.go`: remove the eager whole-dependency
  collision gate and route local overlays through tidy-specific validation.
- `internal/modules/tidy_overlay.go` (new if needed): capture consumer selection
  and effective replacement roots without changing the shared lock.
- Root-rename namespace helpers: change only if a regression demonstrates that
  unused duplicate inventory blocks an otherwise uniquely proven rename.
- `internal/modules/tidy_collisions_test.go`: real pinned Git repositories,
  local replacements, preservation/idempotence and content/conflict behavior.
- `.spec/config/dependency.md`, `V1_RELEASE_NOTES.md`: document the tidy contract
  and its exact-byte rule, distinguish it from other operations.

## Execution

- [x] Add real consumer contracts importing a dependency that imports another;
  create a direct requirement on that common provider as well.
- [x] Add unused equal/differing duplicates, direct/transitive equal copies,
  differing reachable contents and failed-transaction preservation assertions.
- [x] Cover pinned Git and local replace modes, multiple roots, aliases, producer
  exclusions, missing/invalid imports and a second tidy run without changes.
- [x] Run the new tests with Go 1.26.6 and `-race`; record expected failures.
- [x] Implement observed candidate enumeration and exact-byte comparison.
- [x] Integrate it into both tidy paths and retain all equal-copy rechecks.
- [x] Add a mutation regression for a compared noncanonical copy: changing it
  before commit must fail without project writes.
- [x] Run new and related module/root-rename/sourceview suites with `-race`.
- [x] Build fresh CLI/MCP and exercise normal CLI tidy on the real fixtures.
- [x] Run the full source race suite and Go linter with repository-pinned tools.
- [x] Review the diff against the fetched main, fix material findings and report
  the checks actually run. Preserve unrelated and untracked files.

## Decisions

Exact raw-byte equality is intentional: comments, options or declarations that
differ do not qualify for automatic deduplication. No filesystem file is deleted
or rewritten to deduplicate imports. A logical import is parsed once after all
its allowed copies pass equality and source-ownership checks.

## Review-driven checkpoints

- Historical equal copies use previous lock order and declared root order;
  an unchanged dependency or an owned consumer source retains its binding.
- Local overlay metadata is parsed from observed bytes before requirements or
  remote acquisition. Compatibility adaptation also retains metadata absence
  and inventories, including newly introduced nested metadata directories.
- Metadata-free local root inference retains its complete intrinsic proto
  evidence and inventory; fixed metadata roots keep reachable-body validation.
- Optional absent metadata below a directory alias retains its link trace.
  Missing import candidates below valid directory aliases are distinct from
  dangling aliases. Authorized nested-repository absence produces an unresolved
  import instead of dereferencing a missing file's nil metadata.
- Fresh CLI tests in easyp-test cover Git and replace modes, shared requirements,
  preservation, idempotence and unchanged strict generation collision checks.
  Existing v1 scenario arguments were updated to main's command-local --config;
  assertions and v0 scenarios were not changed.

Verification uses Go 1.26.6 and golangci-lint 2.14.0. The final source race run
limits package/test parallelism to 2/8 after earlier runs under simultaneous
builds encountered transient Git pipe/tree-read failures. No dependencies or
toolchain directives were changed.

## Integration into dependency snapshot performance PR (2026-11-01)

The existing design above is approved for integration into
`fix/dependency-snapshot-performance` (PR #247). Keep the command-scoped
snapshot verification, immutable file inventories, source ordering and shared
generation ownership index. Only tidy receives the reachable byte-equality
collision policy; generation and other module operations stay strict.

- [x] Restore the collision regressions and verify the current branch rejects
  the newly allowed unused and byte-identical cases (Go 1.26.6, `-race`).
- [x] Restore tidy's observed candidate resolver, canonical historical binding,
  local replacement metadata checks and optional alias resolution.
- [x] Merge `effective_graph.go` explicitly: retain `cachedModuleSources` for
  remote snapshots, inject the observed local readers and retain remote lock
  evidence for independent tidy verification. Do not replace the PR's index
  path with the stash's earlier filesystem traversal.
- [x] Reproduce additions of equal and differing copies to an indexed remote
  snapshot after overlay verification. Probe live candidate membership in tidy
  without modifying the graph/generation indexes, then pass all nine mutation
  cases with `-race`.
- [x] Verify the combined collision/overlay/sourceview regressions and the PR's
  snapshot inventory, immutable walk and generation regressions.
- [x] Run the complete source race suite with a timeout sufficient for builtin
  WASM compilation, pinned lint, a fresh CLI build and `git diff --check`.
- [x] Review the final integration for correctness and verify the original
  stash remains available.

Use the repository's existing Taskfile/tool versions and Go 1.26.6 with
read-only module resolution. Do not change dependencies or toolchains. The
companion uncommitted `easyp-test` work remains a separate consumer fixture;
validate it with a fresh CLI if its harness is compatible with this branch.

Integration verification completed with Go 1.26.6 and read-only modules:

- The restored closure regression first reproduced the 12 missing-behavior
  failures on PR #247. Combined module/sourceview/metadata package race tests
  then passed.
- Review found that an indexed remote overlay could hide a new import copy
  added immediately after verification. The added equal/differing-copy cases
  failed before the live membership fix; all nine mutation cases passed after it.
- Final `go test -mod=readonly -race -count=1 -p=2 -parallel=8 -timeout=15m ./...`
  passed all 3,352 test/subtest actions, including snapshot/index/generation and
  API coverage.
- A fresh CLI passed all six `TestTidyReachableImportCollisions` e2e cases in
  `easyp-test`, including unchanged strict generation rejection after tidy.
- Pinned golangci-lint 2.14.0 reported zero issues; formatting and
  `git diff --check` passed. Independent follow-up review found no remaining
  actionable issues. The original stash was retained unchanged.

## Shared dependency version selection clarification (2026-11-01)

The motivating report can describe a consumer and a transitive producer
requiring different versions of the same module. This is a version-selection
case before file collisions: select one highest semantic minimum and use only
that revision's proto contents. The existing resolver already implements this
policy; differing source bytes between its unselected and selected revisions
are allowed.

- [x] Verify eight real-Git cases in
  `internal/generation/mvs_integration_test.go`: direct/transitive minima in
  both directions, both requirement orders, fresh and lower previous locks.
  Confirm that one v1.5.0 is selected over v1.2.0, unrequired v1.9.0 is not
  selected, and differing proto contents compile in warm/cold frozen generation.
- [x] Clarify MVS, final-graph file collisions and exact commit constraints in
  the dependency specification without changing the established selection policy.
- [x] Review and rerun the affected generation/resolver tests and pinned lint.
  Both package checks passed with Go 1.26.6 and `-race`; golangci-lint 2.14.0
  reported zero issues. Independent review approved the tests and documentation.

## Historical collision false positive (2026-11-01)

Code-path audit found a reproducible error in `tidyImportBindings`: an old lock
can legitimately contain unused differing copies, but after removing a provider
and importing the remaining current file, historical comparison rejects the
project. The exact current manifest/source succeeds without the old lock.

- [x] Reproduce six failures across cross-module and same-module/root old
  conflicts, both previous lock orders, and equal current copies. Confirm that
  missing current names and conflicting current copies remain rejected.
- [x] Drop only ambiguous historical bindings when all selected revisions have
  complete verified inspections and all candidates for the current name have
  identical bytes. Keep the logical import unchanged. Preserve previous
  canonical bindings for equal historical copies and reject guessed renames.
- [x] Pass the new regressions and existing canonical-owner, unverified identity
  and deleted-old-export tests with Go 1.26.6 and `-race`.
- [x] Review the change, run the complete source race suite and pinned lint,
  build a fresh CLI and exercise the current-versus-old-lock reproduction.
  Final source race run passed 3,371 test/subtest actions; golangci-lint 2.14.0
  reported zero issues. The previous CLI reproduced the historical error, while
  the fresh CLI passed tidy, idempotence and descriptor generation, producing
  the same lock as the fresh current project. Independent review approved the
  fix after its current-name proof was cached once per planning call; an
  isolated 80-name allocation diagnostic fell from 174.4 MiB to 2.20 MiB.
