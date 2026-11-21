---
phase: 02-linker-i-marshrutizatsiya-oshibok
plan: 01
subsystem: core
tags: [protobuf, linker, protoc, compile-diagnostics, resolver, tdd]

# Dependency graph
requires:
  - phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli
    provides: vendored internal/easypc compiler (protoc.Link, Resolver/ErrFileNotFound, WithSyntaxTrees/WithResolver/WithErrorHandler), protoast Position model, IssueInfo/OpenImportFileError in internal/core, openImportFile boundary (importRoots + importFileAllowed + wellknownimports)
provides:
  - internal/core/link.go: CompileRuleName ("COMPILE") and (*Core).linkSources — the three-outcome corpus classifier (linked trees / COMPILE issues / (nil, nil) fallback) that Phase 3's engine and Phase 5's flip consume
  - lintResolver — typed missing-import seam (D-01): linted map for corpus files, exclusive delegation to c.openImportFile for imports, OpenImportFileError/os.ErrNotExist -> atomic missing flag -> protoc.ErrFileNotFound
  - internal/core/link_test.go: TestLinkSources with 6 named cells pinning the etalon COMPILE shape, WKT resolution, missing-import fallback, dep-file diagnostic paths, D-03 sort determinism, multi-file race path
affects: [02-02 contract matrix, 02-03 Action-path matrix, Phase 3 rule engine, Phase 5 lint-path flip]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 2125
  tasks: 2
  commits: 3

# Tech tracking
tech-stack:
  added: [] # zero new dependencies — go.mod/go.sum byte-unchanged (REQ-1.2 invariant held)
  patterns: [typed resolver seam over openImportFile, three-outcome classification (linked/compileIssues/fallback), atomic missing flag + mutex-guarded diagnostic collection, slices.SortStableFunc D-03 deterministic ordering]

key-files:
  created:
    - internal/core/link.go
    - internal/core/link_test.go
    - .planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-01-task1.json
  modified:
    - .planning/config.json

key-decisions:
  - "RED commit carries CompileRuleName const alongside the zero-value stub — minimal name-binding so the test compiles and fails on its planned assertions (a build-failure run is INVALID_RED per tdd.md)"
  - "REFACTOR step skipped — GREEN output already matched go-code-style conventions (import grouping, godoc, callee-only error wraps); nothing to commit"
  - "git.allow_default_branch_commits:true added to .planning/config.json — the documented #3819 override that makes the project's branching_strategy:none (plan work commits on the current branch, Phase 1 precedent) explicit to the commit guard"
  - "dep_file_semantic_error asserts the exact single diagnostic (b.proto, 6:3, COMPILE, '\"Ghost\" is not defined.') probed from the real compiler instead of a vacuous loop over message text"

patterns-established:
  - "Classifier contract for Phase 3/5: (linked, nil) = corpus linked, rules consume trees; (nil, issues) = COMPILE class, rules must not run; (nil, nil) = fallback to the existing per-file path (missing import / non-compilation failure / empty diagnostics)"
  - "Missing-import detection ONLY via lintResolver.missing (atomic.Bool) mapped from typed errors — diagnostic message text is never parsed"
  - "TDD RED evidence record persisted at .planning/phases/<phase>/tdd-red-evidence-<plan>-task<N>.json and machine-validated via gsd_run check tdd-red-evidence before GREEN"

requirements-completed: [LINT-ENG-02, LINT-ENG-03]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Compile-error corpus classifies to exactly one COMPILE issue with the etalon shape (Path test.proto, Line 4, Column 3, verbatim 'Required fields are not allowed in proto3.', zero Filename/Offset/SourceName) and nil linked trees"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/compile_error"
        status: pass
    human_judgment: false
  - id: D2
    description: "Compiling corpus importing google/protobuf/timestamp.proto returns non-nil *protoc.Linked and zero issues — WKT resolves through openImportFile's wellknownimports fallback; protoc.WithStandardImports deliberately not used (single WKT source)"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/compiling_with_wkt"
        status: pass
    human_judgment: false
  - id: D3
    description: "Corpus with a missing import classifies to the (nil, nil) fallback via the typed resolver seam (OpenImportFileError/os.ErrNotExist -> missing atomic flag -> protoc.ErrFileNotFound), with zero COMPILE issues — never by parsing diagnostic text"
    requirement: LINT-ENG-03
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/missing_import"
        status: pass
    human_judgment: false
  - id: D4
    description: "Dep-file semantic error diagnostic carries the dep's import path, not the importer's (b.proto '\"Ghost\" is not defined.', COMPILE)"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/dep_file_semantic_error"
        status: pass
    human_judgment: false
  - id: D5
    description: "D-03 deterministic ordering: two consecutive linkSources runs on the same multi-diagnostic corpus produce equal issue slices sorted by path, then line, then column"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/deterministic_order_twice"
        status: pass
    human_judgment: false
  - id: D6
    description: "Multi-file corpus (>=2 linted-map files, one clean, one with a compile error) under -race: only the erroneous file yields COMPILE issues — exercises the concurrent resolver/preload path"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources/multi_file_corpus"
        status: pass
    human_judgment: false

# Metrics
duration: 11 min
completed: 2026-11-19
status: complete
commits: 3
plan_head_before: b5c26dff10db87e20058ebff5b05390c3f4a9d60
plan_head_after: 8b9e412f5f006e5ba391a777d242638fe4ea2411
---

# Phase 02 Plan 01: Corpus Linker Classification (Tracer) Summary

**protoc.Link corpus classifier in internal/core with the three-outcome contract (linked trees / COMPILE issues / silent fallback) behind a typed resolver seam, proven failing-test-first and pinned by 6 green -race cells**

## Performance

- **Duration:** 11 min
- **Started:** 2026-11-19T03:51:56Z
- **Completed:** 2026-11-19T04:03:26Z
- **Tasks:** 2
- **Files modified:** 2 created (plus 2 planning files)

## TDD Cycle (Task 1, tracer)

- **RED (`082c41d`, test):** TestLinkSources with cells `compile_error` / `compiling_with_wkt` / `missing_import` plus a zero-value `linkSources` stub. First RED run was a build failure (`undefined: CompileRuleName`) — invalid RED; fixed by declaring the const beside the stub so the run compiled and failed on the planned assertions: `compile_error` got 0 issues (expected 1), `compiling_with_wkt` got nil linked (expected non-nil). Evidence persisted to `tdd-red-evidence-02-01-task1.json` and machine-validated: `gsd_run check tdd-red-evidence` → **RED_EVIDENCE_OK** (reason `target_test_failed`, go-test format, 3 tests discovered). Semantic assessment: the target test executed and failed at its planned `require.Len`/`require.NotNil` assertions; `missing_import` passed intentionally (the zero-value stub coincides with the fallback outcome).
- **GREEN (`a081bcb`, feat):** `linkSources` ported from the etalon (02-PATTERNS.md excerpts) with exactly the planned adaptations: returns `*protoc.Linked` directly (no lab `linkedFiles` wrapper); D-03 sort via `slices.SortStableFunc` + `cmp.Or` (path, line, col); diagnostics collected under a `sync.Mutex` with `IsWarn` dropped; `missing` kept as `sync/atomic.Bool`; resolver serves linted paths from the map and delegates imports exclusively to `c.openImportFile` (errors.As `*OpenImportFileError` / errors.Is `os.ErrNotExist` → missing + `protoc.ErrFileNotFound`, other errors pass through, success path `defer r.core.close` + `io.ReadAll` with `fmt.Errorf("ReadAll: %w", err)`); Link options exactly `WithResolver`/`WithSyntaxTrees(true)`/`WithErrorHandler`; fallback branch logs the etalon Debug message. Test green under `-race`.
- **REFACTOR:** none — GREEN output already satisfied import-grouping/godoc conventions; no commit made.
- **Hardening (Task 2, `8b9e412`, test-only):** three named cells added — `dep_file_semantic_error` (pins the dep diagnostic carrying the dep's import path: `b.proto`, 6:3, COMPILE, `"Ghost" is not defined.`), `deterministic_order_twice` (two runs, `require.Equal` on full slices; natural unsorted input order `[z.proto, a.proto]` comes out sorted `[a.proto, z.proto]`), `multi_file_corpus` (≥2 linted-map files under `-race`; only `broken.proto` yields issues). Six cells total, all green; no classifier defects exposed, production code unchanged.

## TDD Gate Compliance

RED→GREEN sequence verified: `feat(02-01)` a081bcb descends from `test(02-01)` 082c41d (`git merge-base --is-ancestor` OK, #5275 ancestry rule). Task 2 is test-only (runtime gate exempt) and committed under `test(02-01)`. No REFACTOR commit required. The runtime gate check `task.is-behavior-adding` returned true for the plan and was honored — RED evidence was validated before any implementation commit.

## Accomplishments

- Three-outcome classification contract implemented and test-proven: linked trees for compiling corpora (incl. WKT), COMPILE issues in the D-03 shape/sort for compile-error corpora, silent (nil, nil) fallback for missing imports (LINT-ENG-02 + D-01/D-03 of LINT-ENG-03)
- Typed resolver seam (D-01) is the only missing-import channel: zero filesystem access in the resolver, zero diagnostic-text parsing, no `protoc.WithStandardImports` (single WKT source remains `wellknownimports`)
- Dormancy invariant held: `internal/core/lint.go`, `internal/api/lint.go` and everything under `internal/easypc` byte-unchanged; go.mod/go.sum diff-free

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: failing classification test** - `082c41d` (test)
2. **Task 1 GREEN: corpus linker classification with typed resolver seam** - `a081bcb` (feat)
3. **Task 2: contract hardening cells** - `8b9e412` (test)

**Plan metadata:** docs commit (this SUMMARY + state/roadmap updates).

_TDD plan: RED → GREEN (→ REFACTOR skipped, nothing to clean) → test-only hardening._

## Files Created/Modified

- `internal/core/link.go` - CompileRuleName const, (*Core).linkSources three-outcome classifier, lintResolver typed missing-import seam
- `internal/core/link_test.go` - TestLinkSources, 6 named cells (compile_error, compiling_with_wkt, missing_import, dep_file_semantic_error, deterministic_order_twice, multi_file_corpus)
- `.planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-01-task1.json` - persisted machine-validated RED evidence (classifier input record)
- `.planning/config.json` - git.allow_default_branch_commits:true (see Deviations)

## Decisions Made

- RED commit carries the `CompileRuleName` const with the stub: a name binding, not behavior — required so the RED run fails on assertions instead of compilation (invalid RED)
- `dep_file_semantic_error` pins the probed exact diagnostic (single issue, b.proto 6:3, `"Ghost" is not defined.`) rather than a message-text loop that could pass vacuously
- `git.allow_default_branch_commits:true` set per the documented #3819 override — aligns the commit guard with the project's `branching_strategy: none` and the Phase 1 precedent of committing plan work on the current branch

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Set git.allow_default_branch_commits:true in .planning/config.json**
- **Found during:** Task 1 (first commit attempt)
- **Issue:** The Step-0 commit guard (#2924/#3819) classifies `main` as protected and would refuse every commit of this plan; this project explicitly configures `branching_strategy: none` and all four Phase 1 plans committed plan work to the current branch
- **Fix:** Added the guard's own documented override `git.allow_default_branch_commits: true` to `.planning/config.json` (planning config only, zero code impact)
- **Files modified:** .planning/config.json
- **Verification:** `gsd_run query git.base-branch --is-protected main` → false; all subsequent commits ran normally with hooks
- **Committed in:** plan metadata commit

---

**Total deviations:** 1 auto-fixed (1 blocking issue, planning-config scope)
**Impact on plan:** No scope creep; no production code touched outside the two planned files.

## Issues Encountered

- First RED run failed to compile (`undefined: CompileRuleName`) — resolved inside the RED step by declaring the const beside the stub, per tdd.md's invalid-RED rule; second run failed on planned assertions and passed the machine classifier.

## User Setup Required

None - no external service configuration required.

## Verification Log

- `go build ./... && go test -race -count=1 ./internal/core/ -run 'TestLinkSources'` → ok, 7/7 PASS (parent + 6 named cells)
- `go test -race -count=1 ./internal/core/` (whole package) → ok, no regressions
- `./bin/golangci-lint run ./internal/core/...` → 0 issues (no exclusion needed; dormant linkSources is test-exercised)
- `go mod tidy && git diff --exit-code go.mod go.sum` → zero diff
- `git diff --stat internal/core/lint.go internal/api/lint.go` → empty (D-04 dormancy)
- Source assertions: `CompileRuleName = "COMPILE"`, `lintResolver` with `atomic.Bool`, `linkSources` signature `(*protoc.Linked, []IssueInfo)` present; no `WithStandardImports`/`os.Stat`/`filepath.Walk`; no `was not found or had errors` literal in link.go/link_test.go
- Tracer feedback gate (auto mode): verify re-run end-to-end after Task 1 → green, expansion authorized

## Next Phase Readiness

- Ready for 02-02 (syntax-precedence `hasPerFileSyntaxError`, blocked/non-local import cells, `testdata/link/` fixtures)
- Phase 3's engine can consume `(*Core).linkSources`' three-outcome contract as-is; Phase 5's flip calls it from the lint path unchanged

---
*Phase: 02-linker-i-marshrutizatsiya-oshibok*
*Completed: 2026-11-19*

## Self-Check: PASSED

- internal/core/link.go exists on disk
- internal/core/link_test.go exists on disk
- Commits 082c41d, a081bcb, 8b9e412 all ancestors of HEAD
- `gsd_run check evaluation-scope --plan 02-01 --commits-only` resolved with 3 commits (source: plan-subjects)
- All acceptance criteria re-run after all tasks: green (see Verification Log)
