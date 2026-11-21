---
phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti
reviewed: 2026-11-20T21:24:46Z
depth: deep
files_reviewed: 14
files_reviewed_list:
  - internal/core/link.go
  - internal/api/runtime.go
  - internal/api/lint_v1.go
  - internal/core/lint_linked.go
  - internal/core/lint_model.go
  - internal/core/lint_suppressions.go
  - internal/core/core.go
  - internal/core/dom.go
  - internal/core/lint_linked_read.go
  - internal/core/comment_suppressions.go
  - internal/rules/builder.go
  - internal/rules/lint_helpers.go
  - .golangci.yml
  - internal/api/lint_link_matrix_test.go
findings:
  critical: 0
  warning: 1
  info: 2
  total: 3
  necessity: 1
status: issues_found
---

# Phase 05: Code Review Report

**Reviewed:** 2026-11-20T21:24:46Z
**Depth:** deep
**Files Reviewed:** 14
**Status:** issues_found

## Summary

Reviewed the Phase 5 flip diff (3065960..HEAD: atomic flip 4238fbe, consolidation 263f067, 05-02 error-path fixes) at deep depth, with the full diff scope (184 files, 44 production deletions) swept for deletion completeness. `go build ./...`, `go vet ./...` (compiles all test files), and the core/rules/api/migration/config test suites — including `-race` on core, the 73-finding golden, the CLI stdout golden, and the exit-code matrix — all pass.

Verified clean on the surfaces the phase plan called out:

- **The seam is textually minimal.** `internal/api/runtime.go` differs from 3065960 by the 3-line comment only; `internal/api/lint_v1.go` is byte-identical to f27d2d1 and has zero diff inside the range. `buildCore` feeds the linked engine correctly: `rules.New` builds the 41-rule registry, `Options.Rules` carries the linked `core.Rule` contract, `KnownLintRules` comes from `rules.AllRuleNames()`.
- **Consolidation renames are complete.** `LinkedFileInfo`→`ProtoInfo`, `LintRule`→`Rule`, `LintLinked`→`Lint`, `shouldIgnoreLinked`→`shouldIgnore`, `linkedCommentSuppressions`→`commentSuppressions` — no dangling references survive anywhere (`go vet` compiles every test file; grep sweeps for `lintProtoInfoRead`, `isLintDirective`, `fileOption`, `leadingCommentDocuments`, `LintRule`, `LinkedFileInfo`, `CheckIsIgnored`, old rule names found only deliberate error-chain pins and two stale comments, filed as IN-02).
- **Registry trim feeds downstream correctly.** The derived registry (41 rule names via reflection over the canonical type names + 5 group keys) matches `schemas/easyp-v1.schema.json` enums exactly (no stale names such as COMMENT_FIELD_MESSAGE), feeds `migration.expandRules` via `rules.AllGroups()`, `ValidateNames` via `AllLintUseValues()`, and api per-path exclusion via `AllGroups()`.
- **Breaking path is untouched** — zero diff in breaking_check.go / breaking_checker.go / breaking_profiles.go / proto_comments.go across the whole range; the trimmed `comment_suppressions.go` retains the shared primitives (`suppressionRange(s)`, `directiveComment`, `parseDirectivePrefix`, `compactDirectiveNames`, `collectDirectiveComments`) that the breaking read path consumes; the lint pairing-machine logic moved to `lint_suppressions.go` verbatim. The freeze-gate byte-equality concern therefore does not hide any breaking-path change in this range.
- **The suppression walk stays central** (`checkFile` → `commentSuppressions`, ModuleRule Finish filtered through the same `suppressed` map), and the IN-02 COMPILE-bypass decision is pinned by `compile_bypasses_suppression` and `ignored_broken_dependency_not_suppressed`.
- **The read path is intact** — `parseOnTheirOwn` fallback registers linted-file outlines and records per-file syntax errors; the COMPILE branch never outranks a per-file syntax error (`hasPerFileSyntaxError`).
- **.golangci.yml** — both removed exclusions match their deleted/trimmed targets: the `internal/rules/lint_helpers.go` unused-helper exclusion died with the helpers (the file now holds only used `groupedValues` helpers, referenced by all nine group-rule ports), and the `casing.go` QF1001 exclusion died with the file. No orphan entries remain.
- **The class-1 COMPILE cells genuinely drive the CLI.** `TestLintLinkMatrix/compile_issue_exit1_class` invokes the real `actionV1` (not a unit seam), and `TestLintLinkMatrixExitCodes/compile_error_code_1` asserts the built binary's process exit code 1 with `(COMPILE)` on stdout.

One real defect survived this sweep: the 05-02 broken-import discrimination is narrower than the production open funnels it guards (WR-01 below). Two minor items (one necessity-classified dead guard, one stale-comment finding) round out the report.

## Warnings

### WR-01: Broken-import discrimination misses real open failures that carry no `*os.PathError` — they degrade into misleading COMPILE remarks

**File:** `internal/core/link.go:174-177` (discriminator), with `internal/core/proto_info_read.go:103-147` and `internal/sourceview/local.go:42,56,176,209` (producing funnels)
**Issue:** The 05-02 WR-01 fix classifies "broken import" as exactly `errors.As(err, &pathErr)` for `*os.PathError`, on the stated premise that "a real open failure carries an `*os.PathError` — the OS cause" and the plain non-local rejection wraps no PathError. That premise is false for the production open funnels the resolver actually calls:

- `SourceRoots.OpenSourceFile` (wired as `Options.OpenSourceFile` in `buildCore` for every module-scoped lint) delegates to `sourceview.OpenLocalSelected`, whose real failures return `ErrUnsupported` (non-regular import target, dangling source alias: `"source alias %q is dangling: ...: %w" wrapping ErrUnsupported`), `ErrChanged`, `ErrNestedRepository` (`errors.Join(...)` shapes), and `localView`/`Resolve` failures — none of these is or wraps an `*os.PathError`.
- `FSWalker.Open` (the consumer-only path, `internal/fs/fs/adapter.go:19` → `sourceview.OpenLocal`) returns the same `ErrUnsupported`/`ErrChanged` shapes for the same inputs.

Consequence for `import "dep.proto"` where the resolved target is, e.g., a non-regular file or a dangling source alias under a module root:

1. `lintResolver.FindFileByPath` (`link.go:161-178`): the error is neither `*OpenImportFileError`/`os.ErrNotExist` (so `missing` stays false) nor a `*os.PathError` (so `broken` stays false); the raw error is returned to the compiler.
2. The vendored driver converts **any** import-load failure into the diagnostic `Import %q was not found or had errors.` plus `ErrCompilationFailed` (`internal/easypc/pkg/compiler/driver.go:344-360`).
3. `linkSources` (`link.go:58`): `missing=false`, `broken=false`, `errors.Is(err, ErrCompilationFailed)=true`, `len(diags)>0`, no per-file syntax error → the corpus is classified **COMPILE**.

The run then prints a misleading "not found"-shaped COMPILE remark, skips every lint rule corpus-wide (research A6), and exits 1 via `ErrHasLintIssue` — exactly the masking behavior 05-02's WR-01 was landed to prevent. The old per-file engine produced a command error carrying the OS cause (`openImportFile` → `Open: %w` → exit 1), and the new engine still does so for the PathError subset (proved by `unreadable_import_falls_back`), so the fidelity fix holds only for that subset. The WR-01 test injects a `*os.PathError` carrier, which is why the gap is invisible to the suite.

**Fix:** Classify by contract, not by error shape. Either (a) mark `broken` for every resolver error that is not one of the three plain rejections — wrap `openImportFile`'s `"invalid import path %q"` rejection (and the `*OpenImportFileError`/`ErrFileNotFound` cases) in a typed sentinel so the resolver can distinguish "deliberate plain rejection" (stay raw → etalon A4 COMPILE) from "anything else" (`broken.Store(true)` → fallback to the per-file path, which re-derives the exact carrier); or (b) wrap real funnel failures at `openImportFile`/`openDiskSource` so every OS-level failure carries a typed cause the resolver can detect. Add a matrix cell driving `OpenSourceFile` that returns `ErrUnsupported` (the funnels' own sentinel) and asserts the command-error fallback, not a COMPILE remark.

## Info

### IN-01: Already-defended-elsewhere guard — the post-fallback `readErrs` re-check can never fire

**File:** `internal/core/lint_linked.go:73`
**Category:** necessity
**Issue:** The `else if len(compileIssues) > 0 && !slices.ContainsFunc(readErrs, ...)` condition defends against per-file parse errors recorded by `parseOnTheirOwn` outranking COMPILE issues. That state is provably impossible on the only path where the condition evaluates: reaching it requires `linkSources` to have returned `(nil, compileIssues>0)`, which per `internal/core/link.go:58-69` requires `hasPerFileSyntaxError(paths, sources) == false` — the identical `protoast.Parse` call over the identical (unmutated until line 102) `sources` that `parseOnTheirOwn` then re-parses. So `ContainsFunc` is always false there and only `len(compileIssues) > 0` decides. The concern is already fully owned by `hasPerFileSyntaxError` inside `linkSources` (and by the line 67 guard for read errors). Harmless at runtime; dead condition text.
**Fix:** Drop the `!slices.ContainsFunc(readErrs, ...)` term and keep `len(compileIssues) > 0`, or replace it with a comment citing the `hasPerFileSyntaxError` ownership so a future editor does not read the redundant term as load-bearing.

### IN-02: Stale rename-map comments reference symbols/files deleted by this phase

**File:** `internal/rules/import_used_linked.go:186-188`; `internal/core/proto_info_read.go:18`
**Issue:** The consolidation and flip left two comments asserting facts the diff just invalidated: (1) `existInProto`'s doc says "the name is taken in this package by main's byte-frozen import_used.go, so the port carries the Linked suffix (rename map)" — `import_used.go` was deleted in 4238fbe and the function now holds the plain unsuffixed name `existInProto`, so the sentence describes a state that no longer exists; (2) `breakingProtoInfoRead`'s doc calls itself "the breaking path's own twin of lintProtoInfoRead" — `lintProtoInfoRead` was deleted in the same range. Neither affects behavior; both will mislead the next reader tracing the rename map.
**Fix:** Reword `existInProto`'s comment to state the post-consolidation fact (the port reclaimed the etalon name once the byte-frozen file was deleted), and update the `breakingProtoInfoRead` doc to reference `protoInfoFromSource`/the deleted-twin history without naming a live symbol that does not exist.

---

_Generated by gsd-code-reviewer (deep)_
