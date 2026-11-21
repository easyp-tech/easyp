---
phase: 07-priemka-paritet-korpusov-i-proizvoditelnost
plan: 01
subsystem: plugins
tags: [wazero, data-race, sync-mutex, race-detector, tdd, wasm, synctest]

# Dependency graph
requires:
  - phase: 01-vendor-laba-i-fundament
    provides: BuiltinRun runtime-factory injection seam (builtin_run.go:59) and the deferred wazero race item (STATE.md blocker)
provides:
  - Default wazero runtime construction serialized at ONE point (runtimeCreationMu + newRuntimeSerialized covering builtin_run.go:144 and builtin_impl.go:143 through the shared field)
  - Two-scope concurrency pin TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes exercising the production default factory under -race
  - Machine-validated RED evidence (RED_EVIDENCE_OK) at tdd-red-evidence-07-01-task1.json
  - SC3 full-suite gate unblocked: go test -race -count=1 ./... green for the first time since Phase 1 (41 packages ok, 0 FAIL)
affects: [07-02 (perf baseline must be measured on the fixed state), 07-03 (SC3 battery + parity evidence chain depends on this fix; commit 1f82b61 recorded for its evidence)]

# Actuals (#2632) — pairs with the plan's estimate to calibrate future estimates.
actuals:
  tokens: 2971    # chars/4 over the realized diff (11883 chars, +108/-1 across 4 files)
  tasks: 2
  commits: 2      # MEASURED: git rev-list --count gsd-plan-head-before-07-01..HEAD at SUMMARY write
plan_head_before: a06fcdda6dee8bd121570c525eb7a4c65973d91b
plan_head_after: 1f82b61

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "One-point serialization of a package-global lazy-init seam: wrap only the default factory VALUE, never the field mechanism, so the injection seam stays assignable"
    - "TestMain priming of a third-party lazy global (wazero internal/version) to make a test binary deterministic under -race when frozen test doubles call the raw constructor"

key-files:
  created:
    - internal/adapters/plugin/main_test.go
    - .planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/tdd-red-evidence-07-01-task1.json
  modified:
    - internal/adapters/plugin/builtin_run.go
    - internal/adapters/plugin/builtin_run_test.go

key-decisions:
  - "Serialize ONLY the default runtimeFactory value: package-level runtimeCreationMu + newRuntimeSerialized assigned at builtin_run.go:74; both creation sites stay untouched and route through the shared field — builtin_impl.go zero-diff, frozen seam overrides byte-untouched"
  - "wazero upgrade (v1.10-v1.12) rejected per flagged assumption A4 resolved-by-decision — easyp-side localization per RESEARCH recommendation; reversible, a later ratified re-sync can revisit"
  - "[Rule 2/3 deviation] plugin package gains main_test.go TestMain that primes wazero's internal/version lazy global before m.Run forks test goroutines: the frozen raw seam doubles cannot be ordered by any production-side mutex, and the one-point fix alone left the package red 2/2 post-fix runs"

patterns-established:
  - "One-point race fix pattern: package-level mutex wrapping the default factory constructor; field stays a plain assignable value so test seams keep overriding it directly"
  - "Lazy-global priming in TestMain for -race determinism with frozen test doubles (happens-before via goroutine fork, zero production impact)"

requirements-completed: [LINT-PARITY-01]  # enabler only — the parity implementation lives in 07-02/07-03; shared-ID gate keeps the REQUIREMENTS checkbox pending until those summaries exist

coverage:
  - id: D1
    description: "Default wazero runtime construction is serialized at exactly one point (default factory value wrapped by runtimeCreationMu/newRuntimeSerialized); both creation call sites and the three frozen seam overrides unchanged"
    requirement: LINT-PARITY-01
    verification:
      - kind: unit
        ref: "tests/internal/adapters/plugin/builtin_run_test.go#TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes"
        status: pass
      - kind: other
        ref: "grep gates: newRuntimeSerialized>=2, runtimeCreationMu>=3, runtimeFactory: wazero==0, run.runtimeFactory = func>=3, rt := run.runtimeFactory>=1, git diff --exit-code builtin_impl.go"
        status: pass
    human_judgment: false
  - id: D2
    description: "Both previously poisoned packages (internal/adapters/plugin, internal/generation) green under -race, and the full -race ./... suite green for the first time since Phase 1 (SC3 unblocked)"
    requirement: LINT-PARITY-01
    verification:
      - kind: unit
        ref: "command: go test -race -count=1 ./internal/adapters/plugin/ ./internal/generation/ (green 4/4 consecutive runs)"
        status: pass
      - kind: integration
        ref: "command: go test -race -count=1 ./... (exit 0, 41 packages ok, 0 FAIL)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Canonical TDD RED machine-validated before GREEN: package-level -race failure with DATA RACE frames inside wazero internal/version.GetWazeroVersion, captured at the canonical evidence path and accepted by the checker"
    requirement: LINT-PARITY-01
    verification:
      - kind: other
        ref: "command: gsd-tools check tdd-red-evidence tdd-red-evidence-07-01-task1.json --raw -> RED_EVIDENCE_OK (target_test_failed)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Two-scope concurrency pin exercises the PRODUCTION default factory (no seam override) so any future regression in runtime-creation serialization re-trips the detector at package scope"
    requirement: LINT-PARITY-01
    verification:
      - kind: unit
        ref: "tests/internal/adapters/plugin/builtin_run_test.go#TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes"
        status: pass
    human_judgment: false

# Metrics
duration: 27min
completed: 2026-11-21
status: complete
---

# Phase 07 Plan 01: wazero race fix (TDD) Summary

**Default wazero runtime construction serialized at one point (package mutex on the factory value); the plugin package green under `-race` for the first time since Phase 1 and the full `-race ./...` suite green (41 packages, 0 FAIL) — SC3 unblocked**

## Performance

- **Duration:** 27 min
- **Started:** 2026-11-21T02:57:46Z
- **Completed:** 2026-11-21T03:25:23Z
- **Tasks:** 2
- **Files modified:** 4 (2 production/test code, 1 new test file, 1 evidence record)

## Accomplishments
- One-point race fix: package-level `runtimeCreationMu sync.Mutex` + `newRuntimeSerialized` wrapping ONLY the default `runtimeFactory` value at the `EnsureBuiltinRun` construction (the raw `wazero.NewRuntimeWithConfig` default is gone); both creation sites (`builtin_run.go:144`, `builtin_impl.go:143`) are covered through the same field — `builtin_impl.go` has a ZERO diff and the three frozen seam overrides are byte-untouched
- Two-scope concurrency pin `TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes`: two independent `EnsureBuiltinRun` scopes created concurrently, each invoking `runBuiltinModule` on the echo WASM through the PRODUCTION default factory (no seam override), synchronized via `synctest.Wait` + channel join, zero sleeps
- Machine-validated RED: package-level `go test -race -count=1 ./internal/adapters/plugin/` exit 1 with `WARNING: DATA RACE` inside wazero v1.9.0 `internal/version.GetWazeroVersion()` (version.go:23 read vs :46 write through concurrent `NewRuntimeWithConfig`); `gsd-tools check tdd-red-evidence` printed `RED_EVIDENCE_OK` (verdict `target_test_failed`, GREEN authorized)
- GREEN: both previously poisoned packages green under `-race` 4/4 consecutive runs, and the FULL `go test -race -count=1 ./...` suite green (exit 0, 41 packages ok) — the SC3 milestone gate is unblocked; `bin/golangci-lint run ./internal/adapters/plugin/` 0 issues
- SC3 battery prerequisites now green on the fixed state; 07-03's evidence chain records the fix commit `1f82b61`

## Task Commits

Each task was committed atomically (canonical TDD, RED precedes GREEN, ancestry verified):

1. **Task 1 (RED): two-scope concurrency pin + machine-validated RED evidence** - `4f20f59` (test)
2. **Task 2 (GREEN — tracer): one-point serialization of the default runtime factory** - `1f82b61` (feat)

**Plan metadata:** see docs commit (SUMMARY + STATE/ROADMAP + VALIDATION row)

_Note: no REFACTOR commit — the minimal fix needed no cleanup._

## TDD Gate Compliance

| Gate | Commit | Evidence |
|------|--------|----------|
| RED | `test(07-01)` `4f20f59` | RED_EVIDENCE_OK (machine checker, verdict `target_test_failed`); semantic assessment: the package run failed on the planned race class (DATA RACE inside wazero internals), not build/fixture/unrelated failures |
| GREEN | `feat(07-01)` `1f82b61` | `go test -race -count=1 ./internal/adapters/plugin/ ./internal/generation/` green, 4/4 consecutive runs |
| REFACTOR | — | not needed (no cleanup beyond the minimal fix) |

## Files Created/Modified
- `internal/adapters/plugin/builtin_run.go` - `runtimeCreationMu` + `newRuntimeSerialized` (serialized default factory) replacing the raw default value at the `EnsureBuiltinRun` construction
- `internal/adapters/plugin/builtin_run_test.go` - new `TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes` pin (Task 1; seam overrides untouched)
- `internal/adapters/plugin/main_test.go` - NEW: `TestMain` priming wazero's internal/version lazy global before test goroutines fork (deviation — see below)
- `.planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/tdd-red-evidence-07-01-task1.json` - canonical RED evidence record (RED_EVIDENCE_OK)

## Decisions Made
- Wrapped ONLY the default factory value — the `runtimeFactory` field stays a plain assignable value so the three frozen seam overrides (`builtin_run_test.go`) keep overriding it directly; neither creation call site nor Close/error-joining semantics touched
- wazero upgrade rejected (A4 resolved-by-decision): easyp-side serialization is small, local, and reversible; an upgrade would touch the REQ-5.2 generate compatibility surface with unverified upstream fix status
- [Rule 2/3] Added `main_test.go` `TestMain` priming: diagnosis showed the one-point fix is necessary but not sufficient for the TEST binary — the frozen raw seam doubles call `wazero.NewRuntimeWithConfig` directly, so no production-side mutex can order them within the lazy-init window (post-fix runs were red 2/2 before priming, with race stacks exclusively through the raw seam calls and the serialized default); priming on the TestMain goroutine performs the single lazy write with a happens-before edge (goroutine fork) to every test body, after which all accesses are reads and never race — zero production impact, frozen seams untouched

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] + [Rule 3 - Blocking] `main_test.go` TestMain priming of wazero's lazy version global**
- **Found during:** Task 2 (GREEN) — first post-fix `-race` run of the package still failed
- **Issue:** The plan's one-point fix serializes only the DEFAULT factory path. The three frozen seam doubles (`run.runtimeFactory = func(...)` closures) call the raw `wazero.NewRuntimeWithConfig` constructor directly, and wazero v1.9.0's `internal/version.GetWazeroVersion()` has an unsynchronized lazy global (read-empty at version.go:23 → slow `debug.ReadBuildInfo` → write at version.go:46) — concurrent first-window calls interleave unsynchronized. Post-fix package runs were RED 2/2 with race stacks through the raw seam calls (`builtin_run_test.go:291/:337`) and the serialized default. No production-side change can order those test-closure calls; touching the frozen seams is prohibited by the plan (freeze guard: `run.runtimeFactory = func` count >= 3, byte-untouched overrides).
- **Fix:** New `internal/adapters/plugin/main_test.go` with a `TestMain` that constructs and closes one throwaway runtime via `newRuntimeSerialized` BEFORE `m.Run()` forks test goroutines — the single lazy write happens with a happens-before edge to every test body; all later constructions (default AND raw seam) observe a fully initialized global, so every remaining access is a read and reads never race. Production binary unaffected (TestMain exists only in the test binary).
- **Files modified:** internal/adapters/plugin/main_test.go (new, 31 lines)
- **Verification:** package + generation green under `-race` 4/4 consecutive runs; full `-race ./...` suite exit 0 (41 packages); all freeze-guard grep gates green; `builtin_impl.go` zero-diff
- **Committed in:** `1f82b61` (Task 2 GREEN commit)

---

**Total deviations:** 1 auto-fixed (1 missing-critical/blocking combined)
**Impact on plan:** The deviation is the minimal seam-preserving completion of the plan's own acceptance criterion (package green under `-race`); it adds one 31-line test-only file and leaves every prohibition intact (no seam edit, no call-site edit, no builtin_impl.go diff, no runtime singleton/shared-runtime rewrite).

## Issues Encountered
- None beyond the deviation above. The pre-fix race reproduced reliably (2/2 baseline probes + research), and post-fix-without-priming also reproduced reliably (2/2), making the diagnosis deterministic rather than flaky.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- SC3 full-suite gate (`go test -race -count=1 ./...`) is GREEN on the committed final state — 07-03's milestone battery can run without the wazero carve-out for the first time
- 07-02 (parity/perf scripts) and 07-03 (SC3 battery + corpora disposal) proceed; 07-03's evidence chain records fix commit `1f82b61`
- LINT-PARITY-01 stays pending in REQUIREMENTS.md by the shared-ID gate until 07-02/07-03 summaries exist (this plan is the enabler, not the implementation)

---
*Phase: 07-priemka-paritet-korpusov-i-proizvoditelnost*
*Completed: 2026-11-21*

## Self-Check: PASSED

- Files created/modified verified on disk: `builtin_run.go`, `builtin_run_test.go`, `main_test.go`, `tdd-red-evidence-07-01-task1.json`, `07-01-SUMMARY.md` — all FOUND
- Commits verified in HEAD ancestry: `4f20f59` (RED), `1f82b61` (GREEN), `733510b` (docs) — all FOUND
- Measured from ledger `gsd-plan-head-before-07-01` (`a06fcdd`): 2 production commits, diff +108/−1 across 4 files
- Full-suite acceptance signal re-verified: `go test -race -count=1 ./...` exit 0 (41 packages ok, 0 FAIL)
