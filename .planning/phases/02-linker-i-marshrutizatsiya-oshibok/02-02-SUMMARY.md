---
phase: 02-linker-i-marshrutizatsiya-oshibok
plan: 02
subsystem: core
tags: [protobuf, linker, protoc, compile-diagnostics, syntax-precedence, linecol, tdd]

# Dependency graph
requires:
  - phase: 02-linker-i-marshrutizatsiya-oshibok (plan 02-01)
    provides: (*Core).linkSources three-outcome classifier, CompileRuleName, lintResolver typed missing-import seam, TestLinkSources 6-cell base matrix
  - phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli
    provides: vendored internal/easypc compiler (protoast.Parse, driver cycle diagnostics, LineCol convention), protoast.Position, openImportFile boundary, Phase 1 D-07 rune-column contract (diverged from here)
provides:
  - internal/core/link.go: hasPerFileSyntaxError — behavioral syntax/semantic precedence (per-file protoast.Parse after failed Link); syntax error in any corpus file beats COMPILE and falls back to the per-file path (LINT-ENG-04)
  - TestLinkSources extended to 16 named cells: full missing-import matrix (transitive depth-2, blocked-by-importFileAllowed, WKT, non-local COMPILE), LineCol position pinning (tab/UTF-8), import-cycle 2:1 pin, cross-file D-03 ordering on a testdata corpus (LINT-ENG-02/03/04)
  - internal/core/testdata/link/ fixture tree (tab_utf8/broken.proto, multi_file/{a,b}.proto) for the cells whose coordinate conventions and ordering must be provable
affects: [02-03 Action-path matrix, Phase 3 rule engine, Phase 5 lint-path flip]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 4274
  tasks: 3
  commits: 4

# Tech tracking
tech-stack:
  added: [] # zero new dependencies — go mod tidy leaves go.mod/go.sum byte-unchanged (T-02-SC gate held)
  patterns: [behavioral syntax/semantic precedence via per-file protoast.Parse (no Semantic field at the protoc.Diagnostic seam), blocked-import flow through importFileAllowed with empty importRoots construction, LineCol byte-column pinning against the D-07 rune contract, cross-file D-03 sort proof on natural dep-first input order]

key-files:
  created:
    - internal/core/testdata/link/tab_utf8/broken.proto
    - internal/core/testdata/link/multi_file/a.proto
    - internal/core/testdata/link/multi_file/b.proto
    - .planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-02-task1.json
  modified:
    - internal/core/link.go
    - internal/core/link_test.go

key-decisions:
  - "Precedence is behavioral, not field-based: protoc.Diagnostic has no Semantic field (research A5), so hasPerFileSyntaxError parses each corpus file with protoast.Parse after a failed Link and outranks the COMPILE class on the first syntax error"
  - "The !ErrCompilationFailed fallback branch stays etalon-verbatim with a not-unit-triggerable comment — it is the driver's top-level resolve-failure path, unreachable with a complete linted map (research Pitfall 1)"
  - "transitive_missing_depth2 pins the exit-2 carrier via the DIRECT importer (b.proto): readFilesFromImport is deliberately one-level (Phase 1 D-06), so processing b.proto is where missing.proto surfaces — mirroring the real walk, which lints every corpus file"
  - "blocked_import constructs its Core with EMPTY importRoots plus the importFileAllowed denial — the importRoots loop resolves without consulting the denial, so empty roots are mandatory for the ErrNotExist fall-through (plan fixture constraint honored)"
  - "COMPILE positions pinned at the compiler LineCol convention (tab stop 8, byte columns): fixture hand-computed to column 37 vs D-07 rune column 19 before the first run; the actual diagnostic confirmed the hand computation (Phase 1 D-07(a) discipline)"

patterns-established:
  - "Syntax-over-COMPILE precedence: any corpus file failing protoast.Parse routes the whole corpus to the per-file path where that file becomes a command error, exactly as before the linker existed"
  - "Missing-import classification proof pattern: assert the classifier outcome (nil, nil) AND the per-file fallback target via lintProtoInfoRead errors.As *OpenImportFileError — pins the exit-2 carrier end to end"
  - "Fixture-loading pattern for classifier cells: os.ReadFile relative to the package dir (mustReadFixture helper), display paths kept clean (broken.proto, a.proto)"

requirements-completed: [] # LINT-ENG-02/03/04 shared with plan 02-03 — blocked by the #2388 shared-ID gate until 02-03 produces its SUMMARY

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Syntax-error precedence (LINT-ENG-04): a corpus with any syntax-broken file falls back to (nil, nil) with zero COMPILE issues — both when the rest is clean and when another file carries a semantic error (syntax beats COMPILE); the semantic-only control still emits COMPILE"
    requirement: LINT-ENG-04
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/syntax_error_falls_back"
        status: pass
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/syntax_beats_compile"
        status: pass
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/semantic_only_emits_compile"
        status: pass
    human_judgment: false
  - id: D2
    description: "Missing-import class matrix (LINT-ENG-03): transitive missing at dep depth 2 and blocked-by-importFileAllowed imports classify to the (nil, nil) fallback and their per-file path surfaces *OpenImportFileError carrying the import path (exit-2 carrier); WKT resolves as the compiling class; non-local \"../x.proto\" imports classify COMPILE per etalon (research A4)"
    requirement: LINT-ENG-03
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/transitive_missing_depth2"
        status: pass
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/blocked_import"
        status: pass
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/wkt_resolves"
        status: pass
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/non_local_import_compile"
        status: pass
    human_judgment: false
  - id: D3
    description: "COMPILE positions follow the compiler LineCol convention (1-based line, 1-based column, tab stop 8, byte columns): the tab/UTF-8 fixture pins line 6 / column 37, provably divergent from the Phase 1 D-07 rune-column value 19 (comment records both)"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/tab_utf8_position"
        status: pass
    human_judgment: false
  - id: D4
    description: "Import-cycle diagnostics are pinned at the driver's hard-coded Line 2 / Column 1 with the 'recursively imports itself' message chain"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/import_cycle"
        status: pass
    human_judgment: false
  - id: D5
    description: "D-03 deterministic ordering holds across files: the multi_file testdata corpus with natural dep-first input order emits issues sorted a.proto then b.proto (path, then line, then col)"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/multi_file_matrix_order"
        status: pass
    human_judgment: false

# Metrics
duration: 15 min
completed: 2026-11-19
status: complete
commits: 4
plan_head_before: c896eee848152b19736970d8343e9f480346be57
plan_head_after: c0994dc4d41cf83cf51ff8ba4d9f57d7c2961dff
---

# Phase 02 Plan 02: Classification Semantics Matrix Summary

**Syntax-over-COMPILE precedence via per-file protoast.Parse, the full missing-import matrix (transitive/blocked/WKT/non-local) with *OpenImportFileError exit-2 carriers, and LineCol position pinning on a provably divergent tab/UTF-8 fixture — 16 green -race cells, zero production surface outside internal/core/link.go**

## Performance

- **Duration:** 15 min
- **Started:** 2026-11-19T04:17:54Z
- **Completed:** 2026-11-19T04:32:58Z
- **Tasks:** 3
- **Files modified:** 5 (2 modified, 3 fixtures created, plus 1 planning evidence file)

## TDD Cycles

- **Task 1 RED (`7064ec1`, test):** three cells added — `syntax_error_falls_back`, `syntax_beats_compile`, `semantic_only_emits_compile`. The run compiled and discovered all 9 cells; the two syntax cells failed exactly on their planned `require.Nil(issues)` assertions: the classifier emitted COMPILE issues (`expected '}'` at truncated.proto/a.proto 7:1) for syntax-error corpora — the precise LINT-ENG-04 violation. Control cell passed as intended. Evidence persisted to `tdd-red-evidence-02-02-task1.json`; `gsd_run check tdd-red-evidence` → **RED_EVIDENCE_OK** (reason `target_test_failed`, go-test format). Semantic assessment: target test executed and failed on the planned behavior gap, no build/setup/discovery faults.
- **Task 1 GREEN (`4093e9f`, feat):** `hasPerFileSyntaxError(paths, sources)` added to link.go — iterates paths calling `protoast.Parse`, true on first error; invoked inside `linkSources` on the failed-Link branch after the missing/non-ErrCompilationFailed/empty-diags check and before the D-03 sort, logging the etalon Debug line on fallback. The `!errors.Is(err, protoc.ErrCompilationFailed)` branch kept etalon-verbatim with the documented not-unit-triggerable comment. All 9 cells green under `-race`; golangci-lint clean.
- **REFACTOR:** none — GREEN output already matched go-code-style conventions (import grouping, godoc, no behavior-preserving cleanup needed).
- **Task 2 (`f37f319`, test-only):** four matrix cells added — `transitive_missing_depth2`, `blocked_import`, `wkt_resolves`, `non_local_import_compile`. No classifier gap exposed: every cell proved inherited `openImportFile`/resolver semantics, so per the plan no feat commit was made. Table restructured with `setup`/`newCore` hooks; asserts receive engine+walker for the per-file `lintProtoInfoRead` ErrorAs assertions (exit-2 carrier pins).
- **Task 3 (`c0994dc`, test-only):** testdata/link/ tree created; three pinning cells added — `tab_utf8_position` (hand-computed Line 6 / Column 37 confirmed by the actual diagnostic on the first run; D-07 rune column would be 19 — divergence recorded in the test comment), `import_cycle` (driver's hard-coded 2:1 with the recursive-imports message), `multi_file_matrix_order` (natural dep-first input order emits a.proto-then-b.proto).

## TDD Gate Compliance

RED→GREEN verified for Task 1: `feat(02-02)` 4093e9f descends from `test(02-02)` 7064ec1 (`git merge-base --is-ancestor` OK, #5275 ancestry rule). RED evidence machine-validated before any implementation commit (`gsd_run check tdd-red-evidence` → RED_EVIDENCE_OK). Tasks 2 and 3 are test-only (no `<files>` source changes → not behavior-adding); the plan explicitly scopes their feat commits to "only if a fix was needed" — no fix was needed, so only `test(02-02)` commits exist for them. No REFACTOR commit required.

## Accomplishments

- LINT-ENG-04 proven: syntax error in any corpus file beats COMPILE through behavioral per-file parsing — no diagnostic-text scanning, no reliance on the nonexistent `Diagnostic.Semantic` field
- LINT-ENG-03 matrix complete: transitive (dep depth 2) and blocked imports fall back with the typed `*OpenImportFileError` carrier on the per-file path (exit-2 route); WKT compiles; non-local imports classify COMPILE per etalon
- LINT-ENG-02 position/shape pinned: LineCol byte-column convention pinned by a fixture that provably diverges from the D-07 rune contract (37 vs 19); cycle diagnostics pinned at 2:1; D-03 sorting proven across files on a testdata corpus
- Dormancy invariants held: `internal/core/lint.go`, `internal/core/proto_info_read.go`, `internal/api/lint.go`, `internal/core/dom.go`, everything under `internal/easypc` byte-unchanged; go.mod/go.sum diff-free (T-02-SC gate)

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: syntax-precedence cells** - `7064ec1` (test)
2. **Task 1 GREEN: per-file syntax precedence beats COMPILE** - `4093e9f` (feat)
3. **Task 2: missing-import class matrix** - `f37f319` (test)
4. **Task 3: position and shape pinning cells** - `c0994dc` (test)

**Plan metadata:** docs commit (this SUMMARY + state/roadmap updates + RED evidence record).

_TDD plan: RED → GREEN → REFACTOR skipped (nothing to clean) → two test-only matrix/pinning tasks._

## Files Created/Modified

- `internal/core/link.go` - `hasPerFileSyntaxError` (per-file protoast.Parse) wired into the failed-Link branch before issue construction
- `internal/core/link_test.go` - TestLinkSources extended from 6 to 16 named cells; setup/newCore table hooks; mustReadFixture helper
- `internal/core/testdata/link/tab_utf8/broken.proto` - fixture: leading tab + Cyrillic comment + emoji before the error token on one line
- `internal/core/testdata/link/multi_file/a.proto` - clean-shape corpus file with a proto3 `required` error, imports b.proto
- `internal/core/testdata/link/multi_file/b.proto` - dep file with a semantic (`Ghost`) error
- `.planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-02-task1.json` - machine-validated RED evidence record

## Decisions Made

- Precedence lives inside the classifier (etalon lint.go:60-77 mechanics): `hasPerFileSyntaxError` runs on the failed-Link path only — the happy path never pays for extra parses
- `transitive_missing_depth2` asserts the exit-2 carrier through the direct importer (b.proto), not the top-level file: readFilesFromImport is one-level by design (D-06), and the real walk lints every corpus file, so b.proto is where the typed error surfaces
- `blocked_import` Core carries empty importRoots + the denial: the importRoots loop would resolve the dep without consulting importFileAllowed, defeating the cell
- Warn-drop policy (research A3) re-flagged, not silently assumed: no natural warn-producing fixture exists in this phase; the drop stays code-asserted in the port and is re-pinned when a natural warn fixture appears (Phase 5 matrix re-run)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] transitive_missing_depth2 per-file assertion retargeted from a.proto to b.proto**
- **Found during:** Task 2 (first test run, before any commit)
- **Issue:** The cell initially called `lintProtoInfoRead(t.Context(), walker, "a.proto")` expecting an error; it returned nil because `readFilesFromImport` is deliberately one-level (Phase 1 D-06) — a.proto's own read never opens b.proto's imports. The plan's prose ("on the same corpus") did not pin which file surfaces the typed error
- **Fix:** Assert on the direct importer: `lintProtoInfoRead(..., "b.proto")` errors with `*OpenImportFileError{FileName: "missing.proto"}` — mirroring the real walk, which processes every corpus file. Comment added to the cell explaining the one-level semantics. The classifier half ((nil, nil) fallback at dep depth) passed unchanged
- **Files modified:** internal/core/link_test.go
- **Verification:** `go test -race -count=1 ./internal/core/ -run 'TestLinkSources'` green; no production code touched
- **Committed in:** f37f319 (corrected form; the failed variant never entered git history)

---

**Total deviations:** 1 auto-fixed (1 test-expectation bug, Rule 1)
**Impact on plan:** No scope creep; the fix tightened the cell to the semantically correct per-file target. No production code changed outside internal/core/link.go.

## Issues Encountered

- A transient duplicate-cell splice during the Task 3 edit left link_test.go briefly uncompilable (`missing ',' before newline`); diagnosed from the composite-literal error, removed the orphan fragment, and the first compile-clean run passed all three cells with the hand-computed literals intact. Self-corrected editing noise, zero plan impact.

## User Setup Required

None - no external service configuration required.

## Verification Log

- `go build ./... && go test -race -count=1 ./internal/core/ -run 'TestLinkSources'` → ok, 16/16 named cells pass, exit 0
- `go test -race -count=1 ./internal/core/` (whole package) → ok, no regressions
- `./bin/golangci-lint run ./internal/core/...` → 0 issues
- `go mod tidy && git diff --exit-code go.mod go.sum` → zero diff
- `git diff --stat internal/core/lint.go internal/core/proto_info_read.go internal/api/lint.go internal/core/dom.go internal/easypc` → empty (dormancy + boundary untouched)
- `git ls-files` lists all three testdata fixtures (tracked, not gitignored)
- TDD gate: RED 7064ec1 → GREEN 4093e9f with verified ancestry; evidence RED_EVIDENCE_OK
- Source assertions: `hasPerFileSyntaxError` calls `protoast.Parse` and is invoked only on the failed-Link path before issue construction; no diagnostic-message keyword matching in link_test.go; no new error type in link.go; tab_utf8 expected Column 37 != rune-count Column 19 (comment records both)

## Next Phase Readiness

- Ready for 02-03 (Action-path matrix, exit codes, printIssues golden) — the classifier contract is now proven across the whole error matrix, not just the demo paths
- Phase 3's engine and Phase 5's flip can consume `linkSources` as-is; LINT-ENG-02/03/04 stay formally open until 02-03 lands (shared-ID gate #2388), then mark complete for all three plans
- Open follow-ups carried deliberately: warn-drop re-pin when a natural warn fixture exists (Phase 5), recover boundary for vendored-compiler panics (Phase 3/5 decision, FUTURE-HARDENING-01)

---
*Phase: 02-linker-i-marshrutizatsiya-oshibok*
*Completed: 2026-11-19*

## Self-Check: PASSED

- internal/core/link.go, internal/core/link_test.go, and all three testdata/link fixtures exist on disk
- Commits 7064ec1, 4093e9f, f37f319, c0994dc are ancestors of HEAD (measured via git merge-base --is-ancestor)
- `gsd_run check evaluation-scope --plan 02-02 --commits-only` resolved (plan-subject commits on this branch)
- All task acceptance criteria re-run after all tasks: green (see Verification Log)
- Plan-level <verification> suite re-run: all green
