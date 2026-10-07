# Checked Tidy Import Renames Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans. Implement sequentially with a read-only spec gate and then quality gate.

**Goal:** Update uniquely verified consumer import names during mod tidy and commit source/manifest/lock together without changing dependency authority or version-selection semantics.

**Architecture:** Root resolution yields verified old/new namespaces. A small import-edit planner creates bounded byte edits for owned consumer files; an in-memory resolver validates proposed source against the new graph. Extend the existing module metadata transaction to include those edits. A structured tidy result feeds CLI messages after commit.

**Tech Stack:** Go 1.26.6, existing protocompile AST/parser/compiler, sourceview, testify, Git and current CLI/logger. No new dependencies or Go version changes.

Design: [2026-10-07-tidy-import-renames-design.md](2026-10-07-tidy-import-renames-design.md).

## Task 2b: Import rewrite workflow

Ownership for one worker after Task2 resolver reviews finish:

- New `internal/modules/import_rewrites*.go` and colocated tests: verified mappings, literal-span edits, proposed source validation and structured report, separated by responsibility.
- Modify `internal/modules/operations.go`, `internal/modules/import_roots_transition.go` and related resolver hooks for tidy-specific rename planning without weakening get/update guards.
- Modify `internal/modules/manifest_requirements.go` and its colocated tests so direct/indirect augmentation uses proposed import bytes or verified imported ownership, including transitive-module promotion and comment preservation.
- Modify `internal/modules/project_files.go` (`writeV1ResolvedFiles` / `writeV1Lock`) and add a focused `internal/modules/resolved_files_transaction.go` helper with colocated tests. The current pair writer uses `internal/fs/fs.WriteAtomicFile` plus manifest rollback; extend that same module write path to staged source+manifest+lock edits with state rechecks/rollback, instead of creating an unrelated writer.
- Modify `internal/api/mod_v1.go` (`Mod.Tidy`) and `internal/api/mod.go` for deterministic post-commit messages/help, and relevant API tests. No CLI GET changes are required by this extension.
- Small coordinated adapter addition: `internal/adapters/gitmodules/cache.go` exposes owned source-cache directories through the optional consumer port; add a focused cache ownership regression. This excludes current and unreferenced snapshots, even with EASYPPATH inside a consumer root, without reading environment variables in modules.
- Controller owns documentation, schemas/MCP if needed, standalone easyp-test and final builds/pushes. Do not touch others' edits.

- [x] Write a real-Git failing test: old no-metadata api/service/v1/svc.proto with checked old root; new native roots api at a later pin; manifest now requires new version, consumer still imports svc.proto. Tidy must rewrite only the import to service/v1/svc.proto, update lock, report mapping and leave provider trees intact. Observe current namespace guard RED.
- [x] Write an omitted-default case and old-pin/hash/tag failure controls. Verify fresh no-lock inference, unchanged roots and namespace-preserving physical moves retain current behavior.
- [x] Add parsed-literal edit tests for regular/public/weak imports, escaped/concatenated literals, comments, non-import strings, CRLF and modes. Observe RED before implementation.
- [x] Extract/reuse verified pinned namespace information at a tidy-specific resolution boundary before unresolved-import validation. Preserve old source binding when another source/dependency now exports the same old name; use the unique proven rename or report ambiguity. Keep root inference independent of consumer code; do not choose arbitrary names or use basename/package heuristics. No old verified mapping means no automatic repair.
- [x] Implement bounded owning-module source selection and edit only verified local physical targets once. Exclude dependency/cache/nested/workspace/vendor sources according to established boundaries; test that owned build copies participate and outside-owner copies stay untouched without adding build/gitignore heuristics. Preserve internal aliases.
- [x] Validate proposed imports with the new dependency view. Compile each changed source and its reachable closure directly through existing protocompile, preserving unchanged runs' current validation behavior and avoiding full-module duplicate-symbol checks for independent targets/build copies. Feed manifest augmentation the same proposed import view. Abort on needed deleted contracts, ambiguity, collisions, integrity and compile/symbol errors before project writes.
- [x] Extend existing staged module file transaction to source edits plus manifest/lock. Add source-state/race and injected rollback tests asserting every source and metadata byte/mode. Preserve causes with callee `%w` labels.
- [x] Keep Tidy error-only API delegating to a structured reporting variant. CLI logs actual applied changes after commit and SDK/selector guidance. Frozen rejection remains early. Update/get diagnostics point to a manifest-change/tidy flow without rewriting source themselves.
- [x] Run `GOTOOLCHAIN=go1.26.6 go test -mod=readonly -race -count=1 ./internal/modules ./internal/api ./internal/adapters/gitmodules`, focused real-source regressions, vet and pinned lint. Self-review, commit only owned files.
- [x] Independent spec review, fix/re-review; independent code quality review, fix/re-review before migration Task3 starts.

## Controller integration

- [x] Add standalone CLI RED->GREEN regressions for source rewrite, old/new root diagnostics, failure byte preservation, separate breaking baselines and SDK outputs.
- [x] Update `.spec/CLI.md`, dependency docs, V1 release notes and EN/RU public docs. Keep generated schemas unchanged unless model changes justify regeneration.
- [ ] Rebuild final CLI/MCP and run full source race, standard v1 standalone suite and opt-in live/Python checks after all implementation tasks.
- [ ] Commit/push the authorized source/tests/docs branches and attach PRs. No merge or release tag.
