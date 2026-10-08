# EasyP v1 Code Cohesion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Implementers are sequential in the shared isolated worktree; each task has an independent spec gate and then quality gate. Steps use checkbox syntax.

**Goal:** Reduce data coupling and make roots, tidy and migration flow readable without changing v1 behavior.
**Architecture:** Existing packages retain their ports. A checked resolution result replaces resolver-state access; cohesive tidy phases use narrow transaction operations; migration groups local/Git evidence and uses one candidate write operation.
**Tech Stack:** Go 1.26.6, existing sourceview, protocompile, testify, Git and Task. No dependency or format changes.

Design: [2026-10-08-code-cohesion-design.md](2026-10-08-code-cohesion-design.md).
Baseline: 898bb664e273dc8a91fbb85eb7c48dc5dfd25d95.
Branch: codex/readability-cohesion-20261008, separate from feature PR #236.

## File and ownership map

- Task1 owns graph coordination and revision evidence in internal/modules: operations.go, import_roots_resolution.go, import_roots_identity.go, import_roots_transition.go, effective_graph.go, import_roots_overlay.go, new import_roots_result.go, and minimal call-site adaptation in import_rewrites.go / import_rewrites_validation.go. Existing inference/search algorithms remain intact.
- Task2 owns tidy orchestration, proposed source validation and module transaction access: import_rewrites*.go, resolved_files_transaction.go, project_files.go; introduce tidy_operation.go, tidy_inputs.go, tidy_input_verification.go and resolved_file_observations.go as needed to keep one responsibility per file.
- Task3 owns internal/migration: migration.go, apply.go and adjacent transaction helpers, git_selection.go, git_imports.go; introduce local_selection_proof.go, build_candidates.go, migration_namespaces.go and migration_bindings.go for the distinct existing responsibilities.
- Task4 owns Git migration request interpretation: internal/adapters/gitmodules/migration.go, migration_mapping.go, migration_integrity.go and new migration_request.go.
- Controller owns .spec design/plan/architecture/package docs, temporary evidence, fresh binaries, easyp-test invocation and publication. No implementation worker edits these files or another task's production files except the minimal listed Task1 call-site adapters.

Names describe the intended responsibility; consolidate an anticipated small file with its natural owner when it would only contain forwarding glue. Do not manufacture a file split or interface for every helper.

## Task 1: Checked root resolution result

- [ ] Run the existing roots/transition/overlay/tidy behavior regressions before changing code:
  GOTOOLCHAIN=go1.26.6 go test -mod=readonly -race -count=1 -timeout=5m ./internal/modules
  Expected: PASS on the baseline.
- [ ] Introduce a named revision key and one explicit graph request value instead of positional preserveHeads/hints/tidy booleans. A named transition policy must represent checked get/update versus tidy repair planning; reserve enum zero and make a missing policy safe rather than implicitly dropping guards.
- [ ] Have resolveV1Graph return a concrete result with the final lock and owned per-revision evidence. Accessors must copy mutable nested bytes/roots/maps; no consumer receives *importRootSource. Keep incomplete/provisional observations inside resolution only. Repositories without inspections remain supported; distinguish absence of evidence from an incomplete evidence result.
- [ ] Move tidy binding planning off importRootSource into a cohesive planner with current checked evidence, previous lock and the existing pinned-source capability. Reuse old-scope proofs without widening or weakening the original hash check. Retain the generic repository old-namespace/absence fallback.
- [ ] Keep old/new namespace guards explicit in get/update graph orchestration and all effective overlay finalization paths. Audit finalize and finalizeSelections callers so moving a guard cannot silently bypass overlays, root hints or historical baselines.
- [ ] Adapt retainPinnedSources and tidy callers to the checked result, and preserve current behavior before the Task2 orchestration rewrite. No direct fetched/locked map reads outside resolver/evidence implementation.
- [ ] Add a behavioral red/green regression if evidence ownership is not covered: mutations of a returned/current source view or provider buffers must not corrupt a separately retained checked view or its pin/hash proof. Refactoring under existing passing behavioral regressions is the refactor phase; do not add tests for private method names/struct shapes.
- [ ] Run modules/API/generation races and focused Git adapter roots regressions. Inspect actual call graph, self-review and commit only Task1-owned files.
- [ ] Independent spec compliance review; fix/re-review. Then independent quality review; fix/re-review before Task2 starts.

Suggested private value shape, with concrete implementation names chosen by the owner:

~~~go
type rootRevisionKey struct { source, commit string }
type graphResolveRequest struct {
    preserveHeads bool
    hints map[string][]string
    transitions namespaceTransitionPolicy
}
~~~

Normalization and immutable tag checks retain their existing order. Current evidence is selected by exact MVS pin, never a second HEAD lookup.

## Task 2: Cohesive tidy phases and transaction observations

- [ ] Inspect current TidyWithReport checkpoints and run existing tidy/transaction cases before changes:
  GOTOOLCHAIN=go1.26.6 go test -mod=readonly -race -count=1 -timeout=5m ./internal/modules ./internal/api
  Expected: PASS.
- [ ] Move public report types and the top-level tidy operation into a focused owner. Make the operation visibly perform capture -> resolve -> plan -> validate -> commit, with early error handling. Supporting concrete units own consumer observations, import binding/edit planning and proposed-source compilation; the transaction owns file-state storage and staged changes.
- [ ] Preserve distinct ownership checkpoints: initial cache exclusion; manifest/lock capture; early consumer capture for complete cache ownership; late capture for generic repositories after new directories are known; previous-pin boundaries after old-source proof; final source selection/integrity rechecks.
- [ ] Add narrow transaction operations for captured bytes/physical identities and child read-only observations. Replace all external reads of expected/changes and writes to inputs. Return owned/copy-safe observations; bounded SourceRoots validation remains in the observation owner. Preserve nil observation scope semantics for metadata.
- [ ] Keep old namespaces, dependency metadata and absence observations alive through commit. Preserve verifyCapturedInputs -> cache verification -> verifyCapturedInputs, compilation of every changed source/closure, and the post-staging recheck. Retain existing test injection points and rollback/recovery behavior.
- [ ] Keep token-only imports, physical alias deduplication, source bytes/CRLF/modes, indirect augmentation and committed-only deterministic reports unchanged. Preserve read-only local replacement behavior and early frozen rejection at the CLI boundary.
- [ ] Run existing literal/binding/cache-ownership/metadata-mutation/rollback regressions with -race; add a red/green behavioral case only for a proven uncovered boundary.
- [ ] Run modules/API/Git adapter races, self-review and commit owned files.
- [ ] Independent spec gate, then quality gate; fix/re-review before Task3.

Do not replace filesystem/root proof with a general callback or storage abstraction. The existing private transaction stays the writer; observations are a narrow view of that implementation.

## Task 3: Migration build stages, local proof and candidate updates

- [ ] Run current real-Git native/Buf/default/scoped archive, local-selection and transaction regressions:
  GOTOOLCHAIN=go1.26.6 go test -mod=readonly -race -count=1 -timeout=5m ./internal/migration
  Expected: PASS.
- [ ] Group Plan's inputs/roots/packages/paths/source inventory into a named local selection proof with construction and rechecking. Keep exact selected imports, bytes and alias topology; empty local inputs must retain their current semantics. Local namespace observation takes the local identity directly, not p.git.localName.
- [ ] Extract ordered Build stages for capture/parse, local/Git selection, known-candidate rendering/preflight, optional dependency verification, namespace/source proof and final outputs. Capture conflicting output metadata and known candidate errors before explicitly permitted dependency access.
- [ ] Separate filesystem namespace observation from pure selection translation and binding/compilation proof. Preserve reversed legacy local root precedence, producer FIRST-prefix archive rewrites, physical generation filters, full available-namespace collision checks, selected/reachable witness bytes and deferred compile mode for Apply rechecks.
- [ ] Add one migration candidate operation that atomically updates the in-memory preview and staged write proposal. Replace manual p.tx.changes edits in replaceCandidate and backup mode handling; retain byte/mode checks of existing backups and unmodified easyp.lock.
- [ ] Preserve read-only no-resolve previews, native no-op/conflicts, explicit resolution authorization, historical pin/hash verification, default and scoped export-ignore semantics, and beforeApply after staging.
- [ ] Run migration/API races plus existing failure/cancellation/confirmation/rollback cases. If extraction exposes an uncovered behavior boundary, demonstrate it red before fixing; do not modify behavior assertions to accept changed output.
- [ ] Self-review/commit; independent spec gate then quality gate before Task4.

## Task 4: Named Git migration requests

- [ ] Run the existing Git migration archive/roots tests first:
  GOTOOLCHAIN=go1.26.6 go test -mod=readonly -race -count=1 -timeout=5m ./internal/adapters/gitmodules -run 'Test.*Migration'
  Expected: PASS.
- [ ] Convert public roots parameters once into a named internal request that distinguishes whole-namespace proof (nil roots), selection proof retaining producer/intrinsic roots (non-nil empty) and checked explicit roots. Keep public FetchMigration/FetchMigrationWithRoots signatures.
- [ ] Replace native/selection boolean argument lists in legacy layout verification with named request fields/intent. Keep native initial fetch, historical expected-hash, symlink-aware archive variants, selected omission allowances and whole-input completeness checks in the same execution order.
- [ ] Preserve early version/root validation, metadata identity, exact commit/hash, checkout cleanup/causes, snapshot alias materialization, installed cache identity and BSR resolution behavior.
- [ ] Run adapter and migration races, focused export-ignore/default/alias controls and relevant CLI migration cases. Commit owned changes after self-review.
- [ ] Independent spec gate, then quality gate; fix/re-review.

## Task 5: Whole-change verification and publication

- [ ] Update .spec/ARCHITECTURE.md and .spec/PACKAGES.md to describe the actual owners/data flow. Product behavior docs and schemas should have no content drift; no edits in docs/easyp-test repositories unless a demonstrated behavioral coverage gap requires one.
- [ ] Verify boundaries directly: no tidy *importRootSource dependency or fetched-map reads; module transaction storage accessed only by its implementation; migration candidate updates use their owner; no new cyclic package coupling or generic layers.
- [ ] Run GOTOOLCHAIN=go1.26.6 go test -mod=readonly -race -count=1 -timeout=5m ./..., go vet -mod=readonly ./..., pinned golangci-lint 2.14.0, and task schema:check dev-tools:check proto:check. Expected: PASS and unchanged generated schemas/go.mod/go.sum.
- [ ] Build fresh CLI/MCP into owned temporary evidence resources. Run GOTOOLCHAIN=go1.26.6 with EASYP_BIN/EASYP_MCP_BIN/EASYP_SOURCE against the complete easyp-test standard v1 suite, -mod=readonly -race -tags=v1 -count=1 -timeout=15m. Expected: PASS, including Go SDK compilation, source edits/modes, migration, cold/frozen and breaking imported-contract controls.
- [ ] Run affected opt-in live/Python checks when supported; retain exact evidence and explicitly state unavailable reported-plugin coverage.
- [ ] Independent final integration/code review; fix/re-review. Record actual commands, revisions and outcomes.
- [ ] Inspect original checkout/feature branch preservation and clean owned status. Commit, push the authorized refactor branch, create/attach a reviewable PR. If #236 is still open, use its feature branch as the stacked PR base; if merged, base main and incorporate it without force-pushing unrelated branches. No merge or tag.

## Workflow and testing discipline

The human approved the concrete design and autonomous implementation on 2026-10-08. Do not ask again for execution choice or routine code organization. Single implementer at a time; reviewers are read-only; controller does not edit worker-owned production files.

Existing source behavior was developed red/green and passed source/standard v1 race suites plus GitHub CI. This is behavior-preserving refactoring under those regressions. Before/after tests are mandatory; new tests must verify genuine behavior/ownership/checkpoints, not mirror architecture.
