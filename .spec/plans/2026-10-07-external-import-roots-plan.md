# External Import Roots Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Resolve and reproduce metadata-free dependency roots and migrate legacy Git root/sub_directory without changing SDK namespaces or generation scope.

**Architecture:** Authoritative metadata supplies ordinary Module.Roots; no-config inference and explicit hints supply verified fallback roots in LockedModule.Roots. Cache returns a per-lock effective view over immutable source bytes. Migration and CLI reuse this resolution contract.

**Tech Stack:** Existing Go 1.26.6, testify, yaml.v3, protocompile, Git, Task, and existing schemas; no new dependencies.

Spec: [2026-10-07-external-import-roots-design.md](2026-10-07-external-import-roots-design.md).

## Task 1: Root metadata, schema and immutable cached views

Ownership: internal/config/v1/{lock,module,schema} Go files and tests, internal/adapters/module_config, internal/adapters/gitmodules. Generated schemas remain controller-owned until integration.

- [ ] Write failing lock/parser tests accepting `Roots []string` with yaml roots,omitempty and rejecting noncanonical/escaping/duplicate roots. Add metadata tests distinguishing authoritative default `.` from no root metadata.
- [ ] Run `GOTOOLCHAIN=go1.26.6 go test -mod=readonly -race -count=1 ./internal/config/v1 ./internal/adapters/module_config` and confirm expected failures.
- [ ] Add `Module.RootsFromMetadata bool` as non-manifest runtime provenance; native parser marks it true, Buf/legacy root readers mark it only when establishing roots. Add `LockedModule.Roots []string` and reflected lock schema descriptions/constraints.
- [ ] Add focused bounded root validation/application helper and roots-aware Cache.Fetch/FetchMigration methods. Keep existing methods delegating with nil roots. Authoritative roots may be matched but never overridden. Pass roots through snapshot preparation and pinned cold installation before strict selected alias validation. Return recorded fallback roots in the fetched lock; keep hashes source based. Include canonical root selections in installed-cache identity so a warm scope cannot bypass alias checks for another scope.
- [ ] Add an optional roots-resolution inspection capability to Fetched/Source: pinned metadata, logical proto bytes and identities, and path-resolution problems. Mark no-config provisional inspection explicitly with RootInspection.Provisional and return an empty hash, making accidental lock installation invalid. Authoritative paths keep complete hashes. Metadata stays strict, while no-config alias faults wait for root selection. Own inspection data after checkout cleanup. Final install re-stages the pinned tree with selected roots and strict alias checks. Existing Source fakes may use the ordinary cache-backed path, while real Git resolution uses this early capability.
- [ ] Test Cached(entry) applying roots only to an independent effective module, cold installation, same bytes/different consumer roots, metadata conflict/malformed metadata, canonical roots, root aliases, needed escapes/cycles and invalid aliases outside an explicit root.
- [ ] Run touched packages with race; self-review and commit only owned files.

Concrete public data additions:

```go
// LockedModule existing fields stay unchanged.
Roots []string `yaml:"roots,omitempty"`
// Module runtime metadata, not a protobuf.mod directive.
RootsFromMetadata bool
```

Concrete cache methods:

```go
func (c *Cache) FetchWithRoots(ctx context.Context, source, version string, roots []string) (modules.Fetched, error)
func (c *Cache) FetchMigrationWithRoots(ctx context.Context, source, version, legacyHash string, roots []string) (modules.Fetched, error)
func (c *Cache) FetchForRootResolution(ctx context.Context, source, version string) (modules.Fetched, error)
```

## Task 2: Shared no-config root resolution and module operations

Ownership: new internal/modules/import_roots*.go, modules/get.go, operations.go, update.go and relevant tests; coordinate a final cache intrinsic-inference hook with Task 1 owner after helper is ready.

- [ ] Write failing real-source tests for dependency file api/service/v1/a.proto importing service/v1/b.proto within that same dependency; assert lock roots and exact import names. Include intrinsic short imports, no-evidence default, ambiguity, multiple disjoint roots, external-consumer mismatch preserving roots and files, current already-resolved imports, duplicates/overlap, package-vs-filename nonheuristic behavior and unrelated malformed sources.
- [ ] Run the focused tests and confirm root resolution fails before implementation.
- [ ] Implement a small bounded candidate index and namespace verifier. Candidates strip complete import-path suffixes from bounded logical paths. Fixed metadata/hints/locked choices cannot be altered; use only import/file evidence from the same module and validate external incoming/reachable imports afterward. Use one intrinsic-only component independently for each metadata-free module; never infer a provider root from consumer/other-module code. Accept exactly one layout, reject ambiguous/invalid layouts with source/import/candidate context.
- [ ] Integrate with revision resolution through a scoped source wrapper that requests inspections and retains selected source/commit data. Select per-module intrinsic roots before the first final installation, collision/import validation and transaction writes. Re-fetch/re-stage final selected roots through the roots-aware interface when required. Preserve immediate immutable commit/tag checks; defer only hash checks of explicitly marked provisional inspections and enforce the original locked hash against the complete final fetch before install/write. Resolve only existing graph modules; retain MVS. Carry valid root choices for unchanged pins and revalidate changed revisions.
- [ ] Add `GetWithRoots(ctx, root, requirement, repository, roots)` with Get delegating using nil roots. Use a scoped roots-aware source wrapper where supported, preserving early snapshot hint handling. Reuse cached effective metadata and locked roots rather than creating manifests in cache. Handle local overlay roots in memory and preserve the shared lock.
- [ ] Add failures proving manifest/lock bytes unchanged on ambiguous roots, unknown imports, conflicts and invalid hints. Add the real-Git intrinsic fixture: api/service/v1/a.proto imports service/v1/b.proto and an unused escaping alias lies outside api; infer api before strict selected-alias validation. Add a negative consumer-only fixture with no intrinsic imports that fails unchanged without a hint and succeeds with explicit api. Test the same alias becoming selected under an explicit different root, including warm-cache attempts. Add same-revision GetWithRoots replacing a previous fallback choice. Check frozen stays a read-only replay operation and no-config inference cannot introduce remote module identities.
- [ ] Add root-authority transition diagnostics for update: compare old fallback/new authoritative import mappings, allow physical move preserving namespace, stop implicit namespace shifts with revisions/roots/renamed imports and actionable explicit-adoption or old-pin guidance. Explicit matching new get hints acknowledge adoption only after caller/import checks. Verify no writes on implicit conflict and successful checked adoption.
- [ ] Run modules race tests; review/commit. Have Task 1 owner connect intrinsic inference to completed snapshot preparation, then re-run adapter and module race tests.

## Task 2c: Preserve a working default namespace in mixed-layout repositories

Ownership: a fresh sequential worker in internal/modules/import_roots selection/constraint/index helpers and tests. Keep metadata, explicit hints, immutable scope proofs and source boundaries unchanged.

- [ ] Reproduce the real-Git public/service.proto -> public/types.proto default edge beside an unrelated tools/aux/service.proto -> aux_types.proto short edge. Prove the new resolver rejects it while the pre-resolver implementation accepts default names and cold/frozen replay.
- [ ] On zero consistent intrinsic inferred layouts only, retain a valid default namespace with existing intrinsic bindings when missing locally matchable declarations originate outside default-bound owner/target identities. Implement the decision explicitly from intrinsic physical identities, without error-string matching or consumer/package/directory heuristics.
- [ ] Preserve connected contradiction errors (direct and chained), multi-solution ambiguity, candidate/search limits, cancellation, selected alias failures, collisions and final pin/hash verification. Reaching an unresolved short-layout source must still fail unchanged.
- [ ] Add permanent regressions and run modules/Git adapter races, vet/lint and the pinned live grpc v1.84.0 get/tidy/generate/cold/frozen/vendor scenario using a fresh binary. Keep Go1.26.6 and dependencies unchanged.
- [ ] Complete independent spec then quality gates before Task3 starts.

## Task 3: Verified migration of Git root/sub_directory

Ownership: internal/migration, migration-specific cache verification hooks in coordination with Task 1; no CLI source modifications yet.

- [ ] Write failing real-Git tests for default/root-only/sub_directory-only/both inputs, duplicate Git inputs, mixed local selectors and Git targets, authoritative dependency roots and historical archive root rewriting. Assert require, module selections, root lock metadata, exact source/import maps, and unchanged inputs on failure.
- [ ] Observe the current custom-root guard and whole-Git/local-selector guard fail these tests.
- [ ] Replace unconditional guards with literal bounded input validation and per-module selection planning. Convert local paths/packages into the local module object when Git modules are also selected. Maintain compact inference for a sole unfiltered local module and root-free Git shorthands.
- [ ] Add an optional roots-aware migration repository interface; call it for explicit legacy hints. Verify the historical hash and pin before resolving roots. Use cache-backed pinned sources to compare legacy archive-normalized selection with v1 logical filenames; translate sub_directory into physical module-relative selectors.
- [ ] Keep unresolved previews blocked for apply until dependency verification is authorized. No plugins execute. Surface actual namespace/scope conflicts and proposed corrections. Recheck selected sources and root choices before apply.
- [ ] Run migration/API race tests, including PTY wizard cancellation/confirmation cases and transaction rollback. Commit owned changes after spec and code review.

## Task 4: CLI, schemas, documentation and end-to-end behavior

Ownership: controller internal/api GET wiring/flags/tests, mcp/easypconfig, schemas via generator, .spec/CLI.md/dependency docs/V1 release notes; easyp-test/docs use their own clean owned branches and applicable rules.

- [ ] Add a failing CLI test for repeatable get --import-root, frozen rejection and invalid hints rejected before fetch. Inspect actual Get CLI wiring; add the flag alongside existing GET flags and use GetWithRoots.
- [ ] Generate schemas through `GOTOOLCHAIN=go1.26.6 task schema:generate`; update MCP root-field descriptions and mod/get reference. Run schema tests/check; do not hand edit generated JSON.
- [ ] Document roots as resolution metadata rather than generation selectors; add no-config inference/ambiguity, explicit hints, locked roots and frozen semantics, and updated migration examples.
- [ ] Add standalone v1 live local-Git regressions in easyp-test for migration with roots, SDK filename equivalence, source_relative/import/default Go plugin modes, cold cache/frozen replay, independent consumers, metadata precedence, no write on failures and breaking imported-contract changes.
- [ ] Build fresh CLI and MCP into a temporary directory using `GOTOOLCHAIN=go1.26.6 go build -mod=readonly` and run the full standard v1 suite with EASYP_BIN, EASYP_MCP_BIN and EASYP_SOURCE. Keep public plugin/cache downloads separate from private/customer sources.
- [ ] Run `go test -mod=readonly -race -count=1 ./...`, lint, schema check and relevant docs validation. Compile generated SDKs. Review exact diff and evidence; fix confirmed issues without unrelated refactoring.
- [ ] Commit/push authorized implementation/test/docs branches, create reviewable PRs and attach each to this task. Do not merge or tag a release.

## Review checkpoints

- [ ] Independent spec/plan review before code.
- [ ] Each implementation task: red test evidence, green race tests, spec compliance review, then quality review. Keep implementers sequential in the shared worktree.
- [ ] Final integration/code review and full test evidence before reporting completion.
