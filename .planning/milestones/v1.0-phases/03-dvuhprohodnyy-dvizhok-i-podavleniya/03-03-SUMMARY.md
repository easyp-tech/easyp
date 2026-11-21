---
phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya
plan: 03
subsystem: core
tags: [lint-engine, determinism, race-audit, compile-class, suppression-bypass, matrix, tdd, test-only]

# Dependency graph
requires:
  - phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya (plan 03-01)
    provides: LintLinked engine (indexed walk-order drain, serial Collect, Finish re-filter), engine test doubles, newLinkedEngine/writeCorpus conventions
  - phase: 02-linker-i-marshrutizatsiya-oshibok
    provides: linkSources three-outcome classifier, CompileRuleName, Phase 2 matrix cells mirrored one level up (multi_file_corpus, dep_file_semantic_error, deterministic_order_twice)
provides:
  - Parallel-correctness matrix TestLintLinkedMatrix: 9 green -race named cells proving ROADMAP criteria 1 and 4 — repeated-run byte determinism, shared-double statelessness audit with exact per-file counts and line identity, serial walk-order Collect (race detector as the serialization proof), ctx cancellation propagation, walk-order error precedence
  - COMPILE-class matrix through the engine: mixed-corpus isolation with rules skipped corpus-wide, dep-file diagnostics on the dep's import path, suppression-bypass pin (A6/IN-02 documented in-code for the Phase 5 flip plan), multi-diagnostic compile determinism (D-03+D-01 sort composed)
  - Zero engine defects: every cell passed against the plan-03-01 engine on the first run — determinism, serialization, and error precedence hold as designed
affects: [03-04, Phase 4 rule porting (statefulness tripwire), Phase 5 CLI flip (IN-02 decision inherited)]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 3581
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: [] # zero new dependencies — go mod tidy leaves go.mod/go.sum byte-unchanged (T-03-SC gate held)
  patterns: [repeated-run byte-identity via require.Equal over full issue slices with shared double instances across runs, identity tallies (per-file rule counts + per-file declaration-line counts with hand-computed expected lines), deliberately unsynchronized Collect recording slice as the -race serialization proof, COMPILE-shape assertions per the Phase 2 etalon numbers with bare CompileRuleName, armed-everything suppression-bypass pin]

key-files:
  created:
    - internal/core/lint_linked_matrix_test.go
  modified:
    - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-VALIDATION.md

key-decisions:
  - "No engine fix needed: all 9 matrix cells passed against the 03-01 engine on the first -race run (and stayed green across 5+3 stability runs) — repeated-run determinism, indexed walk-order drain, serial Collect placement, and ctx propagation hold as designed; the engine-fix mandate from the waves stays unused, and no fix(03-03) commit exists"
  - "serialCollectModuleRule appends to an unsynchronized slice on purpose: Collect runs in the serial drain, so the append is lock-free safe — a concurrent Collect would trip the race detector on that exact slice, making the detector itself the serialization proof (doubleModuleRule in lint_linked_test.go stays mutex-protected and untouched)"
  - "Identity beyond counts: race_shared_doubles_identity tallies findings per (file, rule) AND per (file, declaration line) with hand-computed expected lines (m0 at 5/9, imports shift the rest to 7/11), so a cross-file bleed or shifted position cannot pass the exact maps"
  - "compile_bypasses_suppression is an intentional behavior pin with the A6/IN-02 deferral documented in-code (grep 'IN-02' >= 1 per acceptance): the Phase 5 flip plan inherits the COMPILE-suppressibility decision consciously, not accidentally"
  - "COMPILE shape asserted per the Phase 2 etalon numbers with the bare CompileRuleName constant (package core cannot reference core.CompileRuleName); zero Filename/Offset/SourceName and verbatim messages pinned on the isolation cell"

patterns-established:
  - "Matrix-cell style for parallel-correctness proofs: named subtests, t.Parallel + per-cell t.TempDir, exact counts never range bounds, all assertions after the wait point"
  - "Suppression-bypass pin shape: arm every channel (AllowCommentIgnores, directive fixture above the offending declaration, IgnoreOnly naming the erroneous path, double eligible on the clean files) and assert the COMPILE issues return unchanged"

requirements-completed: [LINT-ENG-01]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Parallel-correctness audit (ROADMAP criteria 1 and 4): 5 repeated in-process runs byte-identical over an 8-file shared-import corpus with shared double instances (full-slice require.Equal), exact per-file rule and declaration-line tallies under -race, Collect recorded in exact walk order twice through an unsynchronized slice, pre-cancelled ctx surfaces as context.Canceled, and the first walk-order syntax error (a_bad.proto) beats the later one"
    requirement: LINT-ENG-01
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedMatrix/repeated_runs_byte_identical"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedMatrix/race_shared_doubles_identity"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedMatrix/collect_serial_walk_order"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedMatrix/ctx_cancel_surfaces"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedMatrix/walk_order_error_precedence"
        status: pass
    human_judgment: false
  - id: D2
    description: "COMPILE-class matrix through the engine (Phase 2 classifier contract one level up): mixed 3-file corpus isolates COMPILE to the erroneous file in etalon shape with rules skipped corpus-wide, dep semantic errors carry the dep's import path with the verbatim compiler message, COMPILE provably bypasses ignoreOnly/comment-suppression channels (A6/IN-02 pinned for Phase 5), and multi-diagnostic compile output is byte-identical across runs in D-03+D-01 order"
    requirement: LINT-ENG-01
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedMatrix/mixed_corpus_compile_isolation"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedMatrix/dep_file_diagnostic_through_engine"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedMatrix/compile_bypasses_suppression"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintLinkedMatrix/compile_deterministic_twice"
        status: pass
    human_judgment: false

# Metrics
duration: 10 min
completed: 2026-11-19
status: complete
commits: 2
plan_head_before: 4b129f6d3930a0838dd8befe9ae34e95b1197705
plan_head_after: 8647bf1631f9599a1046f648a8bcbe8ce2626ac9
---

# Phase 03 Plan 03: Parallel Correctness and COMPILE Matrix Summary

**Parallel-correctness and COMPILE matrix proven through the linked engine: 9 green -race cells — repeated-run byte determinism, shared-double statelessness audit with exact counts, serial walk-order Collect, ctx-cancel propagation, walk-order error precedence, and the Phase 2 COMPILE contract (isolation, dep paths, suppression bypass A6/IN-02) — zero engine defects exposed**

## Performance

- **Duration:** 10 min
- **Started:** 2026-11-19T14:16:15Z
- **Completed:** 2026-11-19T14:26:28Z
- **Tasks:** 2
- **Files modified:** 2 (1 created test file, 1 planning file)

## TDD Cycles

- **Task 1 (`fdfff03`, test):** `TestLintLinkedMatrix` created with the 5 planned parallel-correctness cells over an 8-file shared-import corpus. The MVP+TDD gate predicate `gsd_run query task.is-behavior-adding` returned `false` (test-only `<files>`), so the RED gate is exempt (03-01 Task 2 / 03-02 precedent). The plan's RED design anticipated this shape: the repeated-runs and walk-order cells "fail on assertions against any nondeterministic or completion-order implementation — against the plan-03-01 engine they are expected to pass, which IS the proof". All 5 cells passed on the first -race run and stayed green across 5 stability runs — the passing table is the determinism proof, no RED evidence JSON was required (the plan scopes evidence persistence to the failure case), and no `fix(03-03)` engine commit arose.
- **Task 2 (`8647bf1`, test-only):** four COMPILE-class cells added, mirroring the Phase 2 classifier cells one level up. All 4 passed on the first run (+3 more stability runs); no engine defect exposed, `lint_linked.go` untouched, `linkSources` untouched. 9/9 matrix cells green under -race.
- **REFACTOR:** none — test-only plan; no production code changed.

## TDD Gate Compliance

Both tasks are test-only (`<files>` lists only `lint_linked_matrix_test.go`); the MVP+TDD gate verb returned `is_behavior_adding=false`, exempting them from the RED gate. Plan frontmatter is `type: execute` (not `type: tdd`), so the plan-level RED/GREEN commit-pattern gate does not apply. Commit sequence: `test(03-03)` fdfff03 → `test(03-03)` 8647bf1. No `feat(03-03)` exists because no cell exposed an engine defect — the plan scopes engine fixes to `fix(03-03)` only when the matrix exposes one, which it did not. No REFACTOR commit required.

## Accomplishments

- ROADMAP criterion 1 is test-proven: 5 repeated `LintLinked` runs over an 8-file corpus return byte-identical issue slices (full-slice `require.Equal`, shared double instances across runs, cache-warm included), exact 33-finding totals, green under -race with zero DATA RACE reports
- ROADMAP criterion 4 is test-proven: the statelessness audit runs shared rule-double instances across the whole corpus with exact per-file rule counts and per-file declaration-line tallies (identity, no cross-file bleed) under -race; a stateful rule port in Phase 4 will trip `race_shared_doubles_identity`
- ModuleRule.Collect serialization (D-03) is proven by the detector itself: `serialCollectModuleRule` records paths into an unsynchronized slice and the recorded sequence equals the walk order exactly, twice per run
- Cancellation and error precedence pinned: a pre-cancelled context surfaces as `context.Canceled` through the engine's `%w` wraps, and with two syntax-error files the indexed drain always reports the first in walk order (`a_bad.proto`), never the later one
- The Phase 2 classifier contract holds from the engine's seat: mixed corpora isolate COMPILE to the erroneous file in exact etalon shape with rules skipped corpus-wide, dep diagnostics carry `b.proto` with the verbatim `Ghost` message, compile output is byte-identical across runs (D-03 sort composed with D-01), and COMPILE provably bypasses every suppression channel with the A6/IN-02 deferral documented in-code for the Phase 5 flip plan
- Dormancy held: `internal/core/link.go`, `internal/core/lint.go`, `internal/api/lint.go`, and `internal/easypc/` byte-unchanged (unstaged and committed); go.mod/go.sum diff-free (T-03-SC); 03-02's handoff finding (none) confirmed — no engine fix mandate from either wave-2 sibling

## Task Commits

Each task was committed atomically:

1. **Task 1: parallel correctness matrix for linked engine** - `fdfff03` (test)
2. **Task 2: COMPILE matrix through linked engine** - `8647bf1` (test)

_Plan metadata: docs commits (this SUMMARY + VALIDATION row, then state/roadmap updates)._

_Test-only plan: no RED/GREEN/REFACTOR production cycle — see TDD Gate Compliance._

## Files Created/Modified

- `internal/core/lint_linked_matrix_test.go` - TestLintLinkedMatrix: 9 named cells + serialCollectModuleRule double + matrixCorpus 8-file shared-import corpus builder
- `.planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-VALIDATION.md` - LINT-ENG-01 row filled with the real -run patterns (TestLintLinked 12 cells, TestLintLinkedMatrix 9 cells)

## Decisions Made

- No engine fix: every cell passed against the 03-01 engine on the first run — the proof the plan asked for; `lint_linked.go` remains untouched across both wave-2 siblings
- Race detector as the serialization proof: the Collect double's unsynchronized append is deliberately lock-free — serial placement makes it safe, concurrency would fail it
- Line-level identity tallies with hand-computed declaration lines strengthen the per-file count audit beyond path-level counts
- `compile_bypasses_suppression` arms all three suppression channels at once (comment ignores + directive fixture + ignoreOnly) and documents the A6/IN-02 deferral in-code so the Phase 5 flip plan inherits the decision explicitly
- COMPILE issue shape asserted per the Phase 2 etalon numbers (4:3, verbatim message, zero Filename/Offset/SourceName) using the bare `CompileRuleName` constant per the link_test.go:36 precedent

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- First `./bin/golangci-lint run ./internal/core/...` after Task 2 flagged staticcheck S1011 on my own new code (loop building `wantOrder` could be `slices.Sorted(maps.Keys(files))` directly). Fixed before the Task 2 commit; the broken variant never entered git history. Lint now reports 0 issues.

## User Setup Required

None - no external service configuration required.

## Verification Log

- `go build ./... && go test -v -race -count=1 ./internal/core/ -run 'TestLintLinkedMatrix'` → ok, 9/9 named cells PASS, exit 0 (stable across 3 additional non-verbose runs; Task 1 cells stable across 5 earlier runs)
- No "DATA RACE" string in any run output (grep 0 across 5 logged Task 1 runs)
- `go test -race -count=1 ./internal/core/` → ok (whole package, no regressions, run after each task)
- `./bin/golangci-lint run ./internal/core/...` → 0 issues (after the S1011 fix)
- `go mod tidy && git diff --exit-code go.mod go.sum` → zero diff (REQ-1.2/T-03-SC)
- Dormancy: `git diff --stat internal/core/link.go internal/core/lint.go internal/api/lint.go` empty; `git status --porcelain internal/easypc/` empty
- Source assertions: `CompileRuleName` count 4 (>=1), `core\.CompileRuleName` count 0, `IN-02` count 2 (>=1); `require.Equal` full-slice comparison present in repeated_runs_byte_identical; exact map-based per-file count assertion present in race_shared_doubles_identity
- Commits measured: `git rev-list --count 4b129f6..HEAD` = 2

## Next Phase Readiness

- Ready for 03-04 (lint.use validation parity): the matrix conventions (matrixCorpus builder, doubles, exact-count style) are reusable in package core; no engine findings to carry
- Phase 4 rule porting inherits the statefulness tripwire: a rule with receiver-field writes fails `race_shared_doubles_identity` under -race
- Phase 5 flip inherits the explicit IN-02 pin: COMPILE bypasses suppression by documented decision (`compile_bypasses_suppression` in-code comment), ready to flip consciously

---
*Phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya*
*Completed: 2026-11-19*

## Self-Check: PASSED

- internal/core/lint_linked_matrix_test.go and this SUMMARY exist on disk
- Commits fdfff03 (Task 1), 8647bf1 (Task 2), df42a7c (docs) are ancestors of HEAD
- `gsd_run check evaluation-scope --plan 03-03 --commits-only` resolved with 3 plan-subject commits on this branch
- All task acceptance criteria re-run after all tasks: green (see Verification Log)
- Plan-level <verification> suite re-run: all green; LINT-ENG-01 marked complete (no sibling plan shares the ID — 03-04 declares only LINT-ENG-06)

