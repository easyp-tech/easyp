---
phase: 02-linker-i-marshrutizatsiya-oshibok
plan: 04
subsystem: core
tags: [protobuf, linker, protoc, diagnostics, warn-drop, mutex, race, tdd]

# Dependency graph
requires:
  - phase: 02-linker-i-marshrutizatsiya-oshibok (plan 02-02)
    provides: internal/core/link.go classifier (inline WithErrorHandler closure, three-outcome classification, hasPerFileSyntaxError precedence) and the 16-cell TestLinkSources matrix — the regression proof for zero behavior change
  - phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli
    provides: sealed internal/easypc compiler — protoc.Diagnostic{Path, Line, Col, Message, IsWarn} value shape and protoc.WithErrorHandler option (read-only)
provides:
  - internal/core/link.go: diagCollector(mu *sync.Mutex, diags *[]protoc.Diagnostic) func(protoc.Diagnostic) — the inline closure extracted verbatim into a named unexported constructor; linkSources registers the identical function value (protoc.WithErrorHandler(diagCollector(&diagsMu, &diags))); godoc pins the etalon warn-drop policy (research A3)
  - internal/core/link_test.go: TestLinkSourcesWarnDrop with named cells warn_dropped, non_warn_collected, warn_mix_order, handler_race_mixed — the IsWarn drop branch exercised by synthetic protoc.Diagnostic values through the production handler (gap G-02-3, LINT-ENG-02)
  - .planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-04-task1.json — machine-validated RED evidence (RED_EVIDENCE_OK)
affects: [Phase 3 rule engine (consumes linkSources as-is), Phase 5 lint-path flip, 02-UAT G-02-3 closure]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 1468
  tasks: 2
  commits: 3

# Tech tracking
tech-stack:
  added: [] # zero new dependencies — go mod tidy leaves go.mod/go.sum byte-unchanged (T-02-SC gate held)
  patterns: [behavior-preserving closure extraction into a named unexported constructor driven directly from a same-package test (the test exercises the real drop branch, not a copy), synthetic protoc.Diagnostic fixtures with neutral messages (no diagnostic-text matching, D-01), WaitGroup fan-out contention pin asserting exact counts and per-identity Line tallies under -race]

key-files:
  created:
    - .planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-04-task1.json
  modified:
    - internal/core/link.go
    - internal/core/link_test.go

key-decisions:
  - "Seam shape: behavior-preserving extraction of the inline WithErrorHandler closure into diagCollector — linkSources keeps registering the identical function value, so the synthetic-warn test drives the real production handler, not a parallel copy (the plan's rejected alternative was a test-only driver; no such path exists without editing sealed internal/easypc)"
  - "RED is a swallow-all stub: the test must compile before GREEN, so the stub keeps diagCollector's final signature while dropping nothing; non_warn_collected and warn_mix_order fail on their planned require assertions while warn_dropped passes trivially — the exact planned RED shape"
  - "Contention pin asserts exactness, not bounds: require.Len equals goroutines*pairsPerGoroutine error posts and per-Line identity counts are exact; order is deliberately unasserted (interleaving is legal, Pitfall 7 pins the mutex discipline only)"

patterns-established:
  - "Handler-seam testing: construct the collector via its production constructor, post synthetic Diagnostic literals, assert the target slice — the sanctioned drive for branches unreachable through natural compiler fixtures"
  - "Race-cell convention: one shared handler, >=4 goroutines, sync.WaitGroup, all require/assert calls after wg.Wait() in the test goroutine, exact-count assertions derived from the posted sequence"

requirements-completed: [LINT-ENG-02]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "G-02-3 closure: the IsWarn drop branch is exercised by a synthetic warn diagnostic through the exact handler linkSources registers with protoc.WithErrorHandler — a warn leaves the collector slice empty and structurally cannot become a COMPILE issue (the slice is diagnostics' only channel into the D-03 sort and COMPILE construction)"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSourcesWarnDrop/warn_dropped"
        status: pass
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSourcesWarnDrop/non_warn_collected"
        status: pass
    human_judgment: false
  - id: D2
    description: "The drop is selective, not a swallow-all: a mixed warn/error sequence through one handler instance yields exactly the error diagnostics in arrival order — the COMPILE input stream is unshrunk and order-preserving"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSourcesWarnDrop/warn_mix_order"
        status: pass
    human_judgment: false
  - id: D3
    description: "Mutex discipline holds under contention: 4 goroutines posting fixed alternating warn/error sequences through one shared diagCollector collect exactly the posted error count with exact per-Line identity tallies, clean under -race (Pitfall 7 carried to the extracted seam)"
    requirement: LINT-ENG-02
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSourcesWarnDrop/handler_race_mixed"
        status: pass
    human_judgment: false
  - id: D4
    description: "Zero classifier behavior change: the extraction is behavior-preserving — all 16 pre-existing TestLinkSources cells green unchanged, full 20-cell matrix green across two consecutive -race -count=2 runs, whole internal/core package green"
    verification:
      - kind: unit
        ref: "tests/internal/core/link_test.go#TestLinkSources (16 cells)"
        status: pass
      - kind: other
        ref: "command: go test -race -count=2 ./internal/core/ -run 'TestLinkSources' (40 cell-passes, 0 failures)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Dormancy, the internal/easypc seal, and the zero-dependency invariant hold at plan close: git diff --stat empty on internal/core/lint.go, internal/core/proto_info_read.go, internal/api/lint.go, internal/core/dom.go; git status --porcelain internal/easypc/ empty; go mod tidy leaves go.mod/go.sum zero-diff; golangci-lint 0 issues"
    verification:
      - kind: other
        ref: "command: git diff --stat internal/core/lint.go internal/core/proto_info_read.go internal/api/lint.go internal/core/dom.go (empty) + git status --porcelain internal/easypc/ (empty)"
        status: pass
      - kind: other
        ref: "command: go mod tidy && git diff --exit-code go.mod go.sum (exit 0)"
        status: pass
      - kind: other
        ref: "command: ./bin/golangci-lint run ./internal/core/... (0 issues)"
        status: pass
    human_judgment: false

# Metrics
duration: 5 min
completed: 2026-11-19
status: complete
commits: 3
plan_head_before: 8fc84eb75db0822b75ac752e17ffede3ffcff4f2
plan_head_after: dc49ec17c6b2ff71b2fcc44733bcaf4ab3c6f408
---

# Phase 02 Plan 04: Synthetic Warn Fixture (G-02-3 Gap Closure) Summary

**diagCollector extracted from the inline WithErrorHandler closure with TestLinkSourcesWarnDrop proving via synthetic protoc.Diagnostic values that warns are dropped at the handler and structurally cannot become COMPILE issues — 20 green -race cells across two consecutive runs, zero classifier behavior change, easypc seal and go.mod/go.sum untouched**

## Performance

- **Duration:** 5 min
- **Started:** 2026-11-19T11:11:26Z
- **Completed:** 2026-11-19T11:16:33Z
- **Tasks:** 2
- **Files modified:** 2 production/test files + 1 planning evidence file

## TDD Cycles

- **Task 1 RED (`7be8b07`, test):** TestLinkSourcesWarnDrop added with three cells — warn_dropped, non_warn_collected, warn_mix_order — driving synthetic `protoc.Diagnostic` literals (neutral messages, no text matching — D-01) through `diagCollector`, the exact constructor linkSources registers with `protoc.WithErrorHandler`. A swallow-all stub with the final signature (`func diagCollector(mu *sync.Mutex, diags *[]protoc.Diagnostic) func(protoc.Diagnostic)`) kept link.go compiling. The run compiled (`go build ./...` clean) and discovered all 3 cells; non_warn_collected failed on its planned `require.Equal` (expected the one posted error, got nil — nothing appended) and warn_mix_order on its planned `require.Len` (`"[]" should have 2 item(s), but has 0`); warn_dropped passed trivially as planned. Evidence persisted to `tdd-red-evidence-02-04-task1.json`; `gsd_run check tdd-red-evidence` → **RED_EVIDENCE_OK** (reason `target_test_failed`, go-test format, target `TestLinkSourcesWarnDrop/non_warn_collected`). Semantic assessment: target executed and failed on the planned assertion because the stub ignores every diagnostic; no build/setup/discovery faults.
- **Task 1 GREEN (`2066ef5`, feat):** stub body replaced by the inline closure body extracted verbatim — early return on `d.IsWarn` (the drop branch at the former link.go:45-47), otherwise `mu.Lock()`/`defer mu.Unlock()` and `*diags = append(*diags, d)`; linkSources registers `protoc.WithErrorHandler(diagCollector(&diagsMu, &diags))` — identical function value, identical closure contract. Godoc states the etalon warn-drop policy (research A3) so a future compiler-lab re-sync does not silently undo it. The three-outcome classification, fallback branches, hasPerFileSyntaxError precedence, and the D-03 sort stayed byte-identical. All 3 new cells green; all 16 pre-existing TestLinkSources cells green — that IS the zero-behavior-change proof. Ancestry: feat descends from test (`git merge-base --is-ancestor` OK).
- **REFACTOR:** none — GREEN output already matched go-code-style conventions (godoc on the new constructor, import grouping unchanged, no cleanup needed).
- **Task 2 (`dc49ec1`, test-only):** handler_race_mixed added as the fourth subtest — 4 goroutines × 25 warn/error pairs through one shared handler with sync.WaitGroup; after Wait, `require.Len(t, diags, 100)` exact total and exact per-Line identity tallies, no warn leakage, `-race` clean. Full matrix green twice consecutively (`-count=2`, 20 named cells × 2 = 40 cell-passes, 0 failures). No handler defect exposed → no feat commit (per plan, fix-the-code-never-the-expectation applies only to genuine defects).

## TDD Gate Compliance

RED→GREEN verified for Task 1: `feat(02-04)` 2066ef5 descends from `test(02-04)` 7be8b07 (ancestry rule, 02-02 precedent). RED evidence machine-validated **before** the feat commit (`gsd_run check tdd-red-evidence` → RED_EVIDENCE_OK; GREEN authorized). RED failed intentionally on the planned assertions (require.Equal / require.Len on the empty collector), not on build/discovery faults. No REFACTOR commit required. Task 2 is test-only (adds one subtest; no `<files>` production source change → not behavior-adding); the plan scopes its commit to `test(02-04)` with a feat only if the cell exposed a genuine defect — it did not.

## Accomplishments

- G-02-3 closed: the IsWarn drop branch is exercised by an executed assertion, not a code read — a synthetic warn posted through the production handler never enters the collector slice, and the slice is diagnostics' only channel into the D-03 sort and COMPILE issue construction, so a warn structurally cannot surface as COMPILE (research A3 re-pinned now per user decision)
- Selectivity proven: non_warn_collected and warn_mix_order pin that the drop is IsWarn-selective and the COMPILE input stream is neither shrunk nor corrupted by warn interleaving
- Pitfall 7 discipline carried to the extracted seam: handler_race_mixed pins the mutex under 4-goroutine contention with exact counts under -race
- Determinism pinned: full 20-cell matrix (16 pre-existing + 4 new) green across two consecutive `-race -count=2` runs
- Dormancy, the internal/easypc seal, and the go.mod/go.sum zero-diff invariant re-audited green; golangci-lint 0 issues

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: synthetic warn-driven handler drop cells** - `7be8b07` (test)
2. **Task 1 GREEN: extract diagnostic handler seam; synthetic warn exercises the drop** - `2066ef5` (feat)
3. **Task 2: handler contention pin and matrix determinism** - `dc49ec1` (test)

**Plan metadata:** docs commit (this SUMMARY + state/roadmap updates).

_TDD plan: RED → GREEN → REFACTOR skipped (nothing to clean) → one test-only contention task._

## Files Created/Modified

- `internal/core/link.go` - inline WithErrorHandler closure extracted verbatim into `diagCollector` (unexported, godoc pins the A3 warn-drop policy); linkSources registers the identical function value; no other classifier line changed
- `internal/core/link_test.go` - TestLinkSourcesWarnDrop with 4 named cells (warn_dropped, non_warn_collected, warn_mix_order, handler_race_mixed) joining the existing 16-cell matrix
- `.planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-04-task1.json` - machine-validated RED evidence record (RED_EVIDENCE_OK)

## Decisions Made

- Extraction is the seam (per plan): the test drives the real production handler value — `protoc.Link` receives exactly what the test exercises — because no natural warn-producing fixture exists in the vendored tree (no construction site sets IsWarn=true; verified in the plan) and producing one via the driver would require editing sealed internal/easypc (forbidden, D-04)
- Synthetic diagnostics carry neutral messages; the test constructs values and never parses or matches diagnostic text (D-01)
- handler_race_mixed asserts per-identity counts by Line, not order — interleaving is legal; exactness (never range/GreaterThan bounds) is the contract per acceptance criteria

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Verification Log

- `go build ./... && go test -race -count=1 ./internal/core/ -run 'TestLinkSourcesWarnDrop'` → ok (warn_dropped, non_warn_collected, warn_mix_order + handler_race_mixed), exit 0
- `go test -race -count=2 -v ./internal/core/ -run 'TestLinkSources'` → ok, 20 named cells × 2 runs = 40 cell-passes, 0 failures (determinism gate; run with -v so the cell count is observable)
- `go test -race -count=1 ./internal/core/` (whole package) → ok, no regressions
- `./bin/golangci-lint run ./internal/core/...` → 0 issues
- `go mod tidy && git diff --exit-code go.mod go.sum` → zero diff
- `git diff --stat internal/core/lint.go internal/core/proto_info_read.go internal/api/lint.go internal/core/dom.go` → empty; `git status --porcelain internal/easypc/` → empty (dormancy + Phase 1 D-04 seal)
- Source assertions: exactly one `func diagCollector` definition in link.go; `protoc.WithErrorHandler(diagCollector(&diagsMu, &diags))` is the sole registration; no duplicated inline closure remains inside linkSources; require.Len in handler_race_mixed is an exact count derived from the posted sequence
- TDD gate: RED evidence RED_EVIDENCE_OK validated before the feat commit; feat 2066ef5 descends from test 7be8b07

## Next Phase Readiness

- The G-02-3 gap is closed; 02-UAT Test 3's synthetic-warn decision is satisfied. The natural-warn re-pin remains a Phase 5 matrix re-run informational item (02-UAT deferred section) — unchanged
- LINT-ENG-02 marked complete (all declaring plans 02-01/02-02/02-03/02-04 now have summaries — shared-ID gate satisfied)
- Phase 3's engine and Phase 5's flip consume `linkSources` unchanged; the classifier contract is now proven including the warn-drop policy under contention

---
*Phase: 02-linker-i-marshrutizatsiya-oshibok*
*Completed: 2026-11-19*

## Self-Check: PASSED

- internal/core/link.go, internal/core/link_test.go, tdd-red-evidence-02-04-task1.json, and this SUMMARY exist on disk
- Commits 7be8b07, 2066ef5, dc49ec1 are ancestors of HEAD (measured via git merge-base --is-ancestor)
- `gsd_run check evaluation-scope --plan 02-04 --commits-only` resolved (plan-subject commits on this branch)
- All task acceptance criteria re-run after all tasks: green (see Verification Log)
- Plan-level `<verification>` suite re-run: all green (5-min duration; 3 plan commits measured from ledger 8fc84eb..HEAD)

