---
phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti
plan: 02
subsystem: lint-engine
tags: [linker, error-classification, importFileAllowed, ctx-cancellation, COMPILE-diagnostics, tdd, pre-flip]

# Dependency graph
requires:
  - phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti (plan 01)
    provides: the committed full-registry golden and CLI stdout golden — the byte-level flip net these error-path fixes must not shift
  - phase: 02-dvuhprohodnyy-dvizhok-i-podavleniya
    provides: the linkSources three-outcome classifier, the lintResolver seam, and the WR-01/WR-02/IN-01/IN-02 deferred findings this plan fixes
provides:
  - WR-01 fixed pre-flip: an import that fails to open/read for a non-notFound OS reason forces the classifier fallback — the run degrades exactly like the old per-file engine (command error carrying the OS cause through the same openImportFile funnels), never a misleading "not found"-shaped COMPILE remark
  - WR-02 fixed pre-flip: the linker's corpus branch consults importFileAllowed on the rooted full path built exactly as openDiskSource builds it (ASVS V4, T-05-01) — a policy-excluded file can no longer be served from memory; the per-file fallback yields the exact os.ErrNotExist carrier
  - IN-01 fixed pre-flip: the sibling-resolver ctx.Err() guard in lintResolver.FindFileByPath restores cancellation promptness
  - IN-02 keep-decision recorded in the linkSources doc comment and pinned by an engine-level cell — COMPILE diagnostics stay outside all suppression machinery through the flip
  - four engine-level cells in TestLintLinkedErrorPathFidelity joining the standing gate set for 05-03's flip battery
affects: [Phase 5 flip (05-03), compat gates, Phase 6/7 corpus gates]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 3747        # 14991 diff chars over internal/core (link.go + lint_linked_matrix_test.go) at chars/4 (estimate was 32000; one production file + one test file)
  tasks: 3
  commits: 2          # MEASURED: git rev-list --count 2328a1e6cf5fab370bff7c6f3a44f0bd15649208..HEAD (task commits; excludes the metadata docs commit)
  plan_head_before: 2328a1e6cf5fab370bff7c6f3a44f0bd15649208
  plan_head_after: 94860c09732719d4c3d54c41182d240db35a1558

# Tech tracking
tech-stack:
  added: []
  patterns: [typed error-path discrimination (errors.As on *os.PathError sets broken; notFound stays missing; no-PathError rejections stay raw), policy-mirror funnel input (rooted full path via RootPath + FromSlash shared by openDiskSource and the linker corpus branch), sibling-resolver ctx-guard shape]

key-files:
  created: []
  modified:
    - internal/core/link.go
    - internal/core/lint_linked_matrix_test.go
    - .planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-VALIDATION.md

key-decisions:
  - "WR-01 cell re-derived against live source (Rule 3): the plan's directory fixture cannot drive the resolver error path — FSWalker.Open refuses non-regular files with ErrUnsupported before any read, and a directory named dep.proto is itself walked as a corpus file, so pass-1 readErrs skip linking before the resolver ever runs; permission bits are not root-safe in CI. The cell injects an *os.PathError (os.ErrPermission) through Options.OpenSourceFile — deterministic, root-safe, identical funnel — and the post-fix expectation is re-derived from the plan's 'returns no error' to a command error carrying the OS cause: the plan's own must-have truth ('behaves like the old per-file engine') and threat model T-05-02 ('the OS cause surfaces as a command error') both require it"
  - "WR-02 cell re-derived to the resolver boundary (Rule 3): an engine-level RED is impossible because pass-1 readSource consults the same openDiskSource policy funnel — with the plan's own fixture the run already errors with the exact os.ErrNotExist carrier pre-fix (probe-verified), so 'run succeeds and lints b.proto from memory' does not exist; the cell drives the lintResolver corpus branch directly (pre-fix it serves denied content from memory — the actual bypass) and additionally pins openDiskSource's exact carrier and the engine-level end-to-end denial"
  - "Typed broken discrimination (the plan's flagged-assumption 2 refinement, implemented as specified): broken is set only for error chains carrying *os.PathError — the plain non-local rejection wraps no PathError and stays raw, preserving the frozen etalon A4 non-local COMPILE classification; the linkSources fallback condition gains the broken term"

patterns-established:
  - "Typed error-path discrimination at the resolver seam: notFound (OpenImportFileError/os.ErrNotExist) → missing; *os.PathError carriers → broken; everything else stays raw for the compiler to classify (A4)"
  - "Policy-funnel input identity: the linker corpus branch computes filepath.Join(rooted.RootPath(), filepath.FromSlash(path)) — the byte-identical construction openDiskSource uses — so one policy function sees one input shape on every file-access funnel"

requirements-completed: [LINT-COMPAT-01]  # copied verbatim from the plan frontmatter; the REQUIREMENTS.md checkbox stays unmarked by the shared-ID gate (05-03/05-04 also declare it)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "WR-01: an import whose open fails for a non-notFound OS reason forces the classifier fallback — LintLinked degrades exactly like the old per-file engine (command error carrying the OS cause through openImportFile), never a misleading COMPILE remark; pre-fix RED showed the misleading 'Import ... was not found or had errors.' COMPILE remark"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedErrorPathFidelity/unreadable_import_falls_back"
        status: pass
    human_judgment: false
  - id: D2
    description: "WR-02: the linker corpus branch honors importFileAllowed on the rooted full path exactly as openDiskSource does — denial returns protoc.ErrFileNotFound with missing stored; openDiskSource's exact os.ErrNotExist carrier and the engine-level end-to-end denial are pinned in the same cell"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedErrorPathFidelity/blocked_corpus_file_errors"
        status: pass
    human_judgment: false
  - id: D3
    description: "IN-01: FindFileByPath honors ctx cancellation with the sibling-resolver guard shape, from the corpus branch and the disk branch alike"
    requirement: LINT-COMPAT-01
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedErrorPathFidelity/find_by_path_honors_ctx"
        status: pass
    human_judgment: false
  - id: D4
    description: "IN-02 keep-decision recorded in the linkSources doc comment and pinned: a broken dependency excluded from the corpus by Ignore still yields its COMPILE diagnostics; the A4 non-local COMPILE classification and every pre-existing classification cell stayed green"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedErrorPathFidelity/ignored_broken_dependency_not_suppressed"
        status: pass
    human_judgment: false
  - id: D5
    description: "Dormant-path gate: full internal/core + internal/rules suites green under -race, go build clean, go mod tidy diff-empty, breaking-owned files (proto_info_read.go, proto_comments.go) zero-diff, bin/golangci-lint 0 issues on internal/core"
    requirement: LINT-COMPAT-01
    verification:
      - kind: other
        ref: "command: go test -race -count=1 ./internal/core/... ./internal/rules/... && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && git diff --exit-code -- internal/core/proto_info_read.go internal/core/proto_comments.go && bin/golangci-lint run ./internal/core/..."
        status: pass
    human_judgment: false

# Metrics
duration: 22 min
started: 2026-11-20T19:30:29Z
completed: 2026-11-20T19:52:37Z
status: complete
---

# Phase 05 Plan 02: Linker Error-Path Fidelity Summary

**WR-01/WR-02/IN-01 fixed on the dormant linked path with typed error discrimination preserving the frozen A4 boundary, and the IN-02 keep-decision recorded in code — RED machine-validated before GREEN, golden net untouched**

## Performance

- **Duration:** 22 min
- **Started:** 2026-11-20T19:30:29Z
- **Completed:** 2026-11-20T19:52:37Z
- **Tasks:** 3
- **Files modified:** 2 production/test files (+1 planning artifact: 05-VALIDATION.md rows; +RED evidence record; +this SUMMARY)

## Accomplishments

- **WR-01 (diagnostic integrity, T-05-02):** a real import open/read failure now forces the classifier fallback. `lintResolver` gained `broken atomic.Bool`, set only when the error chain carries `*os.PathError` (the OS cause) — the `linkSources` fallback condition gained the `broken` term. Before: the compiler's candidate loop swallowed the resolver's raw error into a misleading `Import "dep.proto" was not found or had errors.` COMPILE remark (probe-verified: the engine returned that remark with nil error). After: the corpus falls back per-file and the run errors through the same `openImportFile` funnels the old engine uses — command error carrying the OS cause, never a COMPILE remark.
- **WR-02 (file-selection policy, T-05-01):** the linker's corpus branch (the `linted`-map hit) now consults `r.core.importFileAllowed` on `filepath.Join(rooted.RootPath(), filepath.FromSlash(path))` — the byte-identical rooted-path construction `openDiskSource` uses, so one policy function sees one input shape on every file-access funnel (ASVS V4). Denial stores `missing` and returns `protoc.ErrFileNotFound`; the per-file fallback yields openDiskSource's exact `&os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}` carrier. The frozen file itself is zero-edit (gate-verified).
- **IN-01 (responsiveness, T-05-03):** `FindFileByPath` opens with the sibling-resolver guard — `if ctx != nil && ctx.Err() != nil { return nil, ctx.Err() }` — the shape verbatim from `internal/easypc/protoc/resolver.go`; cancellation is honored from both the corpus branch and the disk branch.
- **IN-02 (keep-decision recorded + pinned):** the `linkSources` doc comment now records that COMPILE diagnostics deliberately bypass `c.ignore`, `ignoreOnly`, and comment suppressions — compiler diagnostics are not rule findings (matches buf/etalon). The new `ignored_broken_dependency_not_suppressed` cell pins it end-to-end: a broken dependency excluded from the corpus by `Ignore` still yields its COMPILE diagnostics, green by construction and through the flip.
- **TDD discipline held:** the RED commit (`bac1dc3`) carries only the four new cells plus the machine-validated evidence record (`RED_EVIDENCE_OK`, target `TestLintLinkedErrorPathFidelity`, exit 1 with exactly the three planned subtest assertion failures); the GREEN commit (`94860c0`) carries only the `link.go` implementation. No REFACTOR commit — the minimal implementation needed no cleanup.
- **The dormant path is flip-ready:** full `internal/core` + `internal/rules` suites green under `-race`, build clean, `go mod tidy` diff-empty, breaking-owned files zero-diff, `bin/golangci-lint run ./internal/core/...` 0 issues. The 05-01 golden net is untouched (the fixes change error-path classification only on inputs the golden corpus does not cover — verified: both golden drives' package suites green in the full run).

## Task Commits

Each task was committed atomically (TDD RED → GREEN):

1. **Task 1 (RED): error-path cells fail on the dormant path** - `bac1dc3` (test; includes the machine-validated RED evidence record)
2. **Task 2 (GREEN): WR-01/WR-02/IN-01 fixes in link.go + IN-02 decision record** - `94860c0` (feat)
3. **Task 3: Dormant-path gate — full core suite + flip-readiness** - verification-only (full -race suites, build, tidy, freeze, golangci — all green; no file changes, no commit)

**Plan metadata:** docs commit (this SUMMARY + STATE/ROADMAP/VALIDATION/WINDOWS updates).

## Files Created/Modified

- `internal/core/link.go` - lintResolver gains `broken atomic.Bool`, the ctx guard, the corpus-branch importFileAllowed policy check, typed broken-flagging on open and read failures; linkSources fallback condition gains the broken term; IN-02 decision paragraph in the doc comment
- `internal/core/lint_linked_matrix_test.go` - NEW `TestLintLinkedErrorPathFidelity` parent with four named cells: `unreadable_import_falls_back`, `blocked_corpus_file_errors`, `find_by_path_honors_ctx`, `ignored_broken_dependency_not_suppressed`
- `.planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/tdd-red-evidence-05-02-task1.json` - machine-validated RED evidence record (RED_EVIDENCE_OK)
- `.planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-VALIDATION.md` - Task rows 1-3 filled with the real -run patterns; WR-01/WR-02 Wave-0 checkbox checked

## TDD Gate Compliance

- RED: `test(05-02)` commit `bac1dc3` — target test `TestLintLinkedErrorPathFidelity` failed on its own planned assertions (exit 1; `--- FAIL` parent line naming the three subtests), evidence record machine-validated `RED_EVIDENCE_OK` via `gsd_run check tdd-red-evidence` before any GREEN.
- GREEN: `feat(05-02)` commit `94860c0` — all four cells green under `-race`; pre-existing classification cells (A4 non_local_import_compile, missing-import matrix, TestLinkSourcesWarnDrop, compile_bypasses_suppression) and standing constraints (TestReadSourceOnce, suppression centrality) stayed green.
- REFACTOR: not needed — the minimal implementation landed clean in GREEN; no cleanup commit.

## Decisions Made

- Typed discrimination exactly per the plan's flagged-assumption 2 refinement: `broken` is set only for `*os.PathError` carriers; the plain non-local rejection wraps no PathError and stays raw, preserving the frozen A4 boundary (the existing `non_local_import_compile` cell stayed green unchanged).
- The corpus-branch policy check keys on `rooted` + `importFileAllowed != nil` (per the plan/research sketch) — deliberately stricter than `openDiskSource`'s check (which also requires `sourceFileOpen != nil`): defense in depth for the linker, and in production `buildCore` sets both fields from the same SourceRoots so the asymmetry never fires there. Documented in the cell comment.
- The WR-01 fixture injects the OS error through `Options.OpenSourceFile` (the production `modules.SourceRoots.OpenSourceFile` func shape) rather than chmod or a directory — deterministic, root-safe in CI, and it exercises the identical `openDiskSource` funnel a real `EACCES` open failure would.
- The IN-02 pin cell uses the file's established broken-file shape (proto3 `required` field — semantic compile diagnostic, same as `mixed_corpus_compile_isolation` and `compile_error_rules_skipped`), giving exactly one pinned COMPILE finding on `dep/broken.proto`; the suppression semantics pinned are identical to the plan's syntax-error wording.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] WR-01 cell fixture and expected outcome re-derived against live source**
- **Found during:** Task 1 (RED probes)
- **Issue:** the plan's fixture (dep.proto as a DIRECTORY) cannot drive the resolver error path: `FSWalker.Open` routes through `sourceview.OpenLocal`, which refuses non-regular files with `ErrUnsupported` before any read happens (no "os.Open succeeds, io.ReadAll fails with EISDIR"), and a directory named `dep.proto` is itself walked as a corpus file, so pass-1 `readErrs` skip linking entirely before the resolver ever runs (probe-verified: the plan's RED mechanism — a COMPILE remark with "File not found." — never fires; the run falls back and errors with `unsupported source target`). Additionally the plan's expected post-fix outcome ("LintLinked returns no error, produces a.proto's own findings") contradicts the plan's own must-have truth ("the run behaves like the old per-file engine") and threat model T-05-02 ("the OS cause surfaces as a command error instead of a false COMPILE remark"): the old engine's `readFilesFromImport` errors through the same `openImportFile` funnels on an unreadable import, so parity means a command error.
- **Fix:** the cell injects an `*os.PathError` (`os.ErrPermission`) through `Options.OpenSourceFile` for the dep.proto path — deterministic, root-safe, identical funnel — and pins the post-fix contract as: `LintLinked` returns a command error carrying the OS cause (`errors.Is(err, os.ErrPermission)`), naming dep.proto, with zero COMPILE findings. Pre-fix RED shows the misleading COMPILE remark (the actual WR-01 misclassification), so the RED mechanism the plan intended is reproduced exactly.
- **Files modified:** internal/core/lint_linked_matrix_test.go
- **Verification:** RED run exit 1 naming the subtest's assertion failure; GREEN under `-race`; recorded in .planning/WINDOWS.md
- **Committed in:** bac1dc3 (RED) / 94860c0 (GREEN)

**2. [Rule 3 - Blocking] WR-02 cell re-derived to the resolver boundary**
- **Found during:** Task 1 (RED probes)
- **Issue:** the plan's engine-level RED claim ("the run succeeds and lints b.proto from memory" pre-fix) does not exist: pass-1 `readSource` reads corpus files through the same `openDiskSource` funnel the fix mirrors, so with the plan's own fixture (ImportFileAllowed + OpenSourceFile set) the run already errors pre-fix with the exact `os.ErrNotExist` carrier — the cell would be green before the fix (unexpected green = cells wrong, per the plan's fails_when). The linker corpus branch can never serve a file pass 1 denied, because both compute the same policy decision on the same input.
- **Fix:** the cell drives the `lintResolver` corpus branch directly over a hand-built state (the actual WR-02 bypass: pre-fix it serves the policy-denied content from memory — probe-verified) and asserts the post-fix contract: `protoc.ErrFileNotFound`, no content, `missing` stored. The same cell additionally pins openDiskSource's exact denial carrier (frozen-file behavior) and the engine-level end-to-end denial (error chain `errors.Is os.ErrNotExist` naming b.proto, no findings) — the complete WR-02 chain the plan described.
- **Files modified:** internal/core/lint_linked_matrix_test.go
- **Verification:** RED run exit 1 naming the subtest's assertion failure; GREEN under `-race`; recorded in .planning/WINDOWS.md
- **Committed in:** bac1dc3 (RED) / 94860c0 (GREEN)

**3. [Rule 3 - Blocking] IN-01 fixture note: plan's line references carried over correctly; ctx guard shape taken from the sibling resolver verbatim**
- **Found during:** Task 2 (implementation)
- **Issue:** none functional — the plan's guard sketch (`if ctx.Err() != nil`) needed the nil-context tolerance the sibling resolvers carry (`if ctx != nil && ctx.Err() != nil`) since protoc may invoke resolvers with a nil context.
- **Fix:** guard implemented in the sibling shape verbatim (`internal/easypc/protoc/resolver.go:27-30`); the cell pins both branches (cached and disk) and the nil-ctx tolerance is preserved by construction.
- **Files modified:** internal/core/link.go
- **Verification:** find_by_path_honors_ctx green under `-race`
- **Committed in:** 94860c0

---

**Total deviations:** 3 auto-fixed (3 Rule 3 blocking: 2 plan-spec re-derivations mandated by the plan's own "if the RED run shows a different failure class, STOP and re-derive the cell" flagged assumption, 1 implementation-shape note)
**Impact on plan:** All three were necessary for the cells to be RED for the right reasons and GREEN to the parity contract the plan itself states; no scope change, zero frozen-file edits, the golden net is untouched.

## Issues Encountered

- The plan's directory-fixture premise was verified false by probes before any cell was pinned (deviation 1): `internal/sourceview` bounds every disk open to regular files, so "os.Open succeeds on a directory" never happens on this tree. The probe evidence (linkSources driven directly, engine-driven variants) is preserved in the RED evidence record's `actual`/`semanticAssessment` fields and the WINDOWS.md entries.
- Pre-existing dirty files in the working tree (`.agents/skills/sdd/scripts/pipeline.sh`, `.planning/phases/03-…/03-VERIFICATION.md`, untracked `.planning/intel/`, `.planning/research/.cache/`, `.planning/state.json`, `.spec/features/`) — out of scope per the deviation scope boundary; left untouched and not committed. (05-01's pending WINDOWS.md ledger entry, left uncommitted by that plan's metadata commit, is included in this plan's metadata commit to close the ledger gap.)

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components. All four cells assert real engine/resolver behavior; the fix is complete on the enumerated error paths.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- **The flip (05-03) inherits a corrected error-path contract.** The three-outcome classifier now behaves like main on missing imports (exit-2 carrier), broken imports (command error with the OS cause), policy-denied files (exact os.ErrNotExist carrier), and non-local imports (COMPILE per etalon A4 — unchanged and pinned). The four new cells join the standing gate set for 05-03's battery.
- **The golden net is untouched:** the fixes change classification only on inputs the 05-01 golden corpus does not cover (all corpus files compile and are policy-allowed there); both golden drives' suites were green in the full-run verification.
- **WR-02 policy-input consistency confirmed:** the corpus branch computes the rooted full path with the byte-identical construction `openDiskSource` uses (`RootPath()` + `filepath.FromSlash`) — same policy function, same input shape, no new file-access funnel.
- No blockers.

---
*Phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti*
*Completed: 2026-11-20*

## Self-Check: PASSED

- All created/modified files exist on disk: internal/core/link.go, internal/core/lint_linked_matrix_test.go, tdd-red-evidence-05-02-task1.json, 05-VALIDATION.md rows, this SUMMARY
- Both task commits are ancestors of HEAD: bac1dc3 (test, RED), 94860c0 (feat, GREEN)
- Commits measured from the persisted ledger: `git rev-list --count 2328a1e6cf5fab370bff7c6f3a44f0bd15649208..HEAD` = 2 (plan_head_before 2328a1e6cf5fab370bff7c6f3a44f0bd15649208, plan_head_after 94860c09732719d4c3d54c41182d240db35a1558)
- All Task-2 acceptance criteria re-run green: new cells + classification matrix under `-race`, breaking-freeze diff empty, `broken` count 9 (>= 4), `ctx.Err()` count 2 (>= 1), TestReadSourceOnce + suppression-centrality cells green
- All Task-3 acceptance criteria green: core+rules suites under `-race`, build, tidy diff-empty, freeze diff-empty, golangci 0 issues on internal/core
- `gsd_run check evaluation-scope --plan 05-02 --commits-only` resolves with 2 commits on this branch
- TDD gate compliance: RED (`test` bac1dc3, RED_EVIDENCE_OK machine-validated) precedes GREEN (`feat` 94860c0); no REFACTOR needed
- WINDOWS.md ledger: two deviation entries appended (WR-01 and WR-02 cell re-derivations); no stubs, no skipped tests, no unrun verifies
