---
phase: 02-linker-i-marshrutizatsiya-oshibok
plan: 03
subsystem: api
tags: [lint, exit-codes, action-path, subprocess, golden, compile-diagnostics, error-routing]

# Dependency graph
requires:
  - phase: 02-linker-i-marshrutizatsiya-oshibok (plans 02-01/02-02)
    provides: (*Core).linkSources three-outcome classifier, CompileRuleName "COMPILE", D-03 issue shape/sort pinned at core level, LineCol byte-column convention
  - phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli
    provides: openImportFile boundary (OpenImportFileError carrier), existing exit route in internal/api/lint.go (ErrHasLintIssue→1, ErrorAs *OpenImportFileError→errExit 2, default→log.Fatal 1), printIssues text/JSON printers
provides:
  - internal/api/lint_link_matrix_test.go: TestLintLinkMatrix — the D-05 three-class matrix proven in process through the unchanged actionV1 route (missing import → *core.OpenImportFileError with FileName; syntax → plain command error; lint findings → ErrHasLintIssue; clean → nil)
  - TestLintLinkMatrixExitCodes — actual CLI exit codes 0/1/2 through the built binary (subprocess, per-child Dir/Env), including the "Cannot import file" stderr message with the missing file name
  - TestPrintIssuesCompileGolden — COMPILE remarks pinned in the existing output format: byte-exact text line with the empty-SourceName artifact, JSON with zero Position.Filename/Offset, D-03 order stability through printIssues
  - The compatibility contract Phase 5's flip depends on, verified while the linker is still dormant (zero production changes)
affects: [Phase 3 rule engine, Phase 5 lint-path flip and matrix re-run]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 3037
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: [] # zero new dependencies — go mod tidy leaves go.mod/go.sum byte-unchanged (T-02-SC gate held)
  patterns: [two-level D-05 matrix (in-process error-type assertions + subprocess numeric exit codes), sync.OnceValues shared binary build for subprocess cells, per-child exec.Cmd.Dir/Env isolation without parent-process mutation, printIssues output golden pinning]

key-files:
  created:
    - internal/api/lint_link_matrix_test.go
    - .planning/phases/02-linker-i-marshrutizatsiya-oshibok/deferred-items.md
  modified: []

key-decisions:
  - "Fixture config uses the v1 spelling `linters: default: MINIMAL enable: [FILE_LOWER_SNAKE_CASE]` — the plan's literal `lint: use:` shape is the legacy v0 format that v1.ParsePolicy rejects (KnownFields strict decode, pinned by TestLintV1RejectsLegacyConfig); deviation Rule 3, documented below"
  - "Subprocess cells stay sequential (no t.Parallel) per the go-testing process-state exception; the binary builds once per run via sync.OnceValues into an OS temp dir and is never committed"
  - "The COMPILE golden pins the shared shape contract from package api (calling unexported linkSources across packages is impossible); the code comment marks the real classifier-through-printIssues run as part of the Phase 5 matrix re-run (research A1)"

patterns-established:
  - "Two-level D-05 verification: error CLASSES asserted in process through actionV1, actual exit CODES asserted only in subprocesses against the built binary — internal-function tests cannot establish process exit codes (.spec/ERRORS.md)"
  - "Clean-corpus layout for MINIMAL-group lint: file in a directory matching its package (PackageDirectoryMatch), lower_snake_case file name (FILE_LOWER_SNAKE_CASE), proto3 message with snake_case field"

requirements-completed: [LINT-ENG-02, LINT-ENG-03, LINT-ENG-04]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "D-05 in-process matrix: missing import routes to *core.OpenImportFileError carrying 'missing.proto' (exit-2 class), syntax error routes to a plain command error and NOT the carrier (exit-1 class), FILE_LOWER_SNAKE_CASE findings route to ErrHasLintIssue, clean compliant corpus routes nil — all through the unchanged actionV1"
    requirement: LINT-ENG-03
    verification:
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrix/missing_import_exit2_class"
        status: pass
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrix/syntax_error_command_class"
        status: pass
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrix/lint_issue_exit1_class"
        status: pass
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrix/compiling_exit0_class"
        status: pass
    human_judgment: false
  - id: D2
    description: "Actual CLI exit codes through the built easyp binary: 2 on missing import with the 'Cannot import file' + file-name stderr message, 1 on syntax error, 0 on a clean corpus with no issue output"
    requirement: LINT-ENG-04
    verification:
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrixExitCodes/missing_import_code_2"
        status: pass
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrixExitCodes/syntax_error_code_1"
        status: pass
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrixExitCodes/compiling_clean_code_0"
        status: pass
    human_judgment: false
  - id: D3
    description: "COMPILE remarks render in the existing printIssues format: text line 'test.proto:4:3: Required fields are not allowed in proto3. (COMPILE)' pinned byte-for-byte (empty-SourceName artifact), JSON with zero Position.Filename/Offset and rule COMPILE after unmarshal, D-03-sorted input order rendered unchanged"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/api/lint_link_matrix_test.go#TestPrintIssuesCompileGolden/text_golden"
        status: pass
      - kind: unit
        ref: "tests/internal/api/lint_link_matrix_test.go#TestPrintIssuesCompileGolden/json_golden"
        status: pass
      - kind: unit
        ref: "tests/internal/api/lint_link_matrix_test.go#TestPrintIssuesCompileGolden/multi_issue_order"
        status: pass
    human_judgment: false
  - id: D4
    description: "Dormancy and zero-install invariants: internal/api/lint.go, internal/api/lint_v1.go, internal/core/lint.go byte-unchanged; go.mod/go.sum diff-free after go mod tidy; no binary artifacts under cmd/ or internal/; matrix green twice consecutively"
    verification:
      - kind: other
        ref: "command: git diff --stat internal/api/lint.go internal/api/lint_v1.go internal/core/lint.go (empty)"
        status: pass
      - kind: other
        ref: "command: go mod tidy && git diff --exit-code go.mod go.sum (exit 0)"
        status: pass
      - kind: other
        ref: "command: go test -race -count=1 ./internal/api/ -run 'TestLintLinkMatrix|TestLintLinkMatrixExitCodes|TestPrintIssuesCompileGolden' run twice, both ok"
        status: pass
    human_judgment: false
  - id: D5
    description: "Flagged assumption A1 (surfaced for ratification at phase verification): the class-1 (COMPILE) through-Action exit-code run is physically unreachable before the Phase 5 flip (Lint() must not change, D-04); per research interpretation A1, class 1 is verified at classifier level (plans 02-01/02-02) plus this plan's printIssues golden, and the identical matrix re-runs through Action in Phase 5"
    requirement: LINT-ENG-02
    verification: []
    human_judgment: true
    rationale: "Whether the A1 two-level verification of D-05 satisfies LINT-ENG-02 before the Phase 5 flip is the phase's main verification assumption — a user must ratify it; automation cannot decide policy interpretation."

# Metrics
duration: 25 min
completed: 2026-11-19
status: complete
commits: 3
plan_head_before: a77a13d596fe8dd9624c5be182a196138804efc5
plan_head_after: 35fc26b0eb1a3ec0ecd4383dfdf610a7b0cd6d3b
---

# Phase 02 Plan 03: Action-Path Matrix and Exit Codes Summary

**D-05's three-class error matrix proven through the real routing surface — in-process actionV1 error-type assertions plus built-binary exit codes 2/1/0 — and COMPILE remarks pinned in the existing printIssues format, with zero production-code changes (linker dormant)**

## Performance

- **Duration:** 25 min
- **Started:** 2026-11-19T04:40:40Z
- **Completed:** 2026-11-19T05:06:12Z
- **Tasks:** 3
- **Files modified:** 1 created (test file), 1 planning file (deferred-items.md)

## Accomplishments

- LINT-ENG-03's exit-2 class proven end to end: missing import reaches the existing route both in process (`actionV1` returns an error matching `*core.OpenImportFileError` with `FileName == "missing.proto"`) and through the built binary (exit code 2, stderr carrying "Cannot import file" and the file name)
- LINT-ENG-04's command-error class proven the same way: syntax error routes to a plain error (explicitly NOT the typed carrier) and the binary exits 1 via main's log.Fatal, exactly as before any change
- LINT-ENG-02's output contract pinned pre-flip: COMPILE remarks render through the unchanged `printIssues` — byte-exact text line with the empty-SourceName artifact, JSON serializing zero `Position.Filename`/`Offset`, D-03 order preserved
- Compatibility matrix is green with `internal/api/lint.go`, `internal/api/lint_v1.go`, `internal/core/lint.go` byte-unchanged — the dormancy invariant Phase 5's flip relies on

## Task Commits

Each task was committed atomically:

1. **Task 1: in-process Action matrix** - `e8a9292` (test)
2. **Task 2: subprocess exit-code matrix** - `6e1bb45` (test)
3. **Task 3: COMPILE output golden** - `35fc26b` (test)

**Plan metadata:** docs commit (this SUMMARY + state/roadmap/requirements updates).

_All three tasks are test-only (`task.is-behavior-adding` = false; plan type: execute), so each produced a single `test(02-03)` commit — no RED/GATE cycle applies._

## Files Created/Modified

- `internal/api/lint_link_matrix_test.go` - `TestLintLinkMatrix` (4 parallel in-process cells), `TestLintLinkMatrixExitCodes` (3 sequential subprocess cells over a sync.OnceValues-built binary), `TestPrintIssuesCompileGolden` (3 parallel golden cells), shared fixture/action helpers
- `.planning/phases/02-linker-i-marshrutizatsiya-oshibok/deferred-items.md` - out-of-scope discovery record (pre-existing plugin-package race)

## Decisions Made

- Fixture easyp.yaml uses the v1 `linters:` spelling (`default: MINIMAL` + `enable: [FILE_LOWER_SNAKE_CASE]`) instead of the plan's literal legacy `lint: use:` shape — see Deviations
- Subprocess cells share one binary build (`sync.OnceValues`) but each gets an isolated `t.TempDir()` corpus and its own `exec.Cmd.Dir`/inherited `Env`; parent process state is never mutated, and the test stays sequential per the go-testing process-state exception
- The lint-issue and clean cells place the proto in a directory matching its package so `FILE_LOWER_SNAKE_CASE` is the only rule finding in the violation cell (MINIMAL group adds only the package/directory rules)
- The COMPILE golden documents in code that the composition seam proven here is the shared shape contract; the real classifier-output-through-printIssues run belongs to the Phase 5 matrix re-run (research A1)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixture config spelling: legacy `lint:` shape replaced with the v1 `linters:` selection**
- **Found during:** Task 1 (pre-implementation source check)
- **Issue:** The plan's literal fixture `version: v1\nlint:\n  use: [FILE_LOWER_SNAKE_CASE]\n` cannot produce any planned behavior: `v1.ParsePolicy` decodes with `KnownFields(true)` and `v1.Policy` has no `lint` field (only `linters`), and unversioned `lint:` documents are rejected as legacy (`ErrLegacyConfiguration`) — exactly what the existing `TestLintV1RejectsLegacyConfig` pins. Every matrix cell would have failed at config parse instead of reaching the error classes.
- **Fix:** Used `version: v1\nlinters:\n  default: MINIMAL\n  enable:\n    - FILE_LOWER_SNAKE_CASE\n` — the equivalent v1 selection (FILE_LOWER_SNAKE_CASE lives in the DEFAULT group per `internal/rules/builder.go`; MINIMAL adds only the four package/directory rules, which the clean corpus satisfies by placing files in a package-matching directory). The plan itself hedged toward this alternative in 02-PATTERNS.md ("or `linters:\n  default: MINIMAL`"). Behavior of all four cells matches the plan's behavior bullets verbatim.
- **Files modified:** internal/api/lint_link_matrix_test.go (fixture constant only)
- **Verification:** all 10 matrix cells green under `-race`, twice consecutively
- **Committed in:** e8a9292 (part of Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking fixture correction, test-scope only)
**Impact on plan:** No scope creep; zero production files touched anywhere in the plan.

## Issues Encountered

- The full-suite phase gate (`go test -race -count=1 ./...`) fails in `internal/adapters/plugin` (~10 failing tests plus a data-race warning inside the third-party `tetratelabs/wazero@v1.9.0` version cache). Verified pre-existing: the package fails identically at this plan's base commit `a77a13d` (checked in a throwaway worktree before any 02-03 commit). Out of scope per the deviation rules' scope boundary — recorded in `deferred-items.md` for the phase verifier rather than fixed here. All targeted suites for this plan (`./internal/api/`, `./internal/core/`) are green under `-race`.

## User Setup Required

None - no external service configuration required.

## Verification Log

- `go build ./... && go test -race -count=1 ./internal/api/ -run 'TestLintLinkMatrix$'` → ok, 4/4 cells
- `go build ./... && go test -race -count=1 ./internal/api/ -run 'TestLintLinkMatrixExitCodes'` → ok, 3/3 cells
- `go build ./... && go test -race -count=1 ./internal/api/ -run 'TestPrintIssuesCompileGolden'` → ok, 3/3 cells
- Combined matrix run twice consecutively → ok both times (RESEARCH determinism requirement)
- `./bin/golangci-lint run ./...` → 0 issues
- `go mod tidy && git diff --exit-code go.mod go.sum` → zero diff (T-02-SC)
- `git diff --stat internal/api/lint.go internal/api/lint_v1.go internal/core/lint.go` → empty (dormancy, D-04)
- `git status --porcelain cmd/ internal/` after the subprocess run → no binary artifacts (binary lives in an OS temp dir)
- Source assertions: missing-import cell uses `require.ErrorAs` into `*core.OpenImportFileError` with `FileName == "missing.proto"`; syntax cell uses `require.Error` + `require.NotErrorAs`; no `os.Exit`/`os.Chdir`/`os.Setenv` calls anywhere in the test file; exit-code cells assert concrete numeric codes from `exec.ExitError.ExitCode()`
- TDD gate: plan is `type: execute` with test-only tasks; `task.is-behavior-adding` returned false — no RED-commit requirement; three `test(02-03)` commits verified in git log

## TDD Gate Compliance

Not applicable as a RED/GREEN plan: frontmatter `type: execute`, all tasks test-only (`task.is-behavior-adding` = false, no `<files>` outside `*_test.go`). All three commits carry the `test(02-03)` scope; no `feat`/`refactor` commits were required and none exist. The orchestrator's RED-commit gate does not fire for non-behavior-adding plans.

## Next Phase Readiness

- Phase 2 complete (3/3 plans): the classifier contract (02-01/02-02) and the Action-path compatibility matrix (this plan) are both green — Phase 3's rule engine can consume `(*Core).linkSources` and Phase 5's flip inherits a pinned output contract plus this matrix to re-run through Action
- LINT-ENG-02/03/04 become markable complete once this SUMMARY exists (shared-ID gate #2388 satisfied: 02-03 was the last declaring plan)
- Carried for the phase verifier: flagged assumption A1 (class-1 through-Action deferred to Phase 5) needs ratification; pre-existing plugin-package `-race` failure documented in `deferred-items.md` blocks a clean full-suite gate reading until fixed in its own subsystem

---
*Phase: 02-linker-i-marshrutizatsiya-oshibok*
*Completed: 2026-11-19*

## Self-Check: PASSED

- internal/api/lint_link_matrix_test.go exists on disk
- Commits e8a9292, 6e1bb45, 35fc26b are ancestors of HEAD (git merge-base --is-ancestor)
- `gsd_run check evaluation-scope --plan 02-03 --commits-only` resolved with 3 commits on this branch (source: plan-subjects)
- All task acceptance criteria re-run after all tasks: green (see Verification Log)
- Plan-level `<verification>` suite re-run: all green except the documented pre-existing plugin-package failure (out of scope, deferred)
