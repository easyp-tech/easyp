---
phase: 07-priemka-paritet-korpusov-i-proizvoditelnost
plan: 04
subsystem: testing
tags: [wazero, sync-mutex, synctest, race-detector, evidence-integrity, test-comments, gap-closure]

# Dependency graph
requires:
  - phase: 07-priemka-paritet-korpusov-i-proizvoditelnost (07-01)
    provides: the production serialization fix (runtimeCreationMu + newRuntimeSerialized) and the TestMain priming whose narrative was falsified
  - phase: 07-priemka-paritet-korpusov-i-proizvoditelnost (07-REVIEW)
    provides: the falsification experiments (design source) and the measured registry-vs-enabled counts (35/44 HEAD, 32/41 baseline)
provides:
  - Corrected TestMain narrative: priming REMOVES the race signal; the production serialization's regression protection is pinned structurally (WR-01 narrative closed)
  - Deterministic held-mutex serialization probe in TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes — mechanical regression detection for runtimeCreationMu (RED when the Lock/Unlock pair is removed, compile-RED when the var is deleted), with the peak==1 overlap counter retained as secondary detector (WR-01 protection closed)
  - Corrected acceptanceCorpusConfig comment stating the measured truth: STANDARD enables 35 of the 44-name HEAD registry (32 of 41 baseline); COMMENTS+UNARY_RPC never enabled; Gate D unaffected (WR-02 closed)
  - Shipped behavior unchanged: production code byte-identical to HEAD; both touched packages green under -race; lint 0 issues
affects: [phase-07 verification gates (verify-work, UAT), any future evidence-time disposition reasoning from the test comments]

# Actuals (#2632) — pairs with the plan's estimate to calibrate future estimates.
actuals:
  tokens: 2385    # chars/4 over the realized diff (9542 chars, +138/-21 across 3 files)
  tasks: 2
  commits: 2      # MEASURED: git rev-list --count gsd-plan-head-before-07-04..HEAD
plan_head_before: 9e8ce794ea7acb4e28838669202192af1bd7fd3e
plan_head_after: f1baa65

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Held-mutex ordering probe: to pin a package-global serialization from a test seam, hold the mutex and prove a concurrent construction cannot complete until release — green is structural (guaranteed by the mutex under test), never timing-dependent"
    - "synctest.Wait quiescence excludes mutex-blocked goroutines (not 'durably blocked') — a probe that holds a mutex across synctest.Wait deadlocks; such phases run in real time outside the bubble"

key-files:
  created:
    - .planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-04-SUMMARY.md
  modified:
    - internal/adapters/plugin/main_test.go
    - internal/adapters/plugin/builtin_run_test.go
    - internal/api/acceptance_gates_test.go

key-decisions:
  - "[Rule 1] Replaced the REVIEW snippet's wrapper-window peak probe with a held-mutex ordering probe: the window conflates mutex-wait with construction overlap (measured false RED under -race with the CORRECT mutex — peak int64(2); measured false-green-prone without it — 1 of 2 runs), and a sync.Mutex block is not durably blocked for synctest.Wait (deadlock proven by goroutine dump), so the probe phase runs in real time"
  - "The probe pins BOTH production creation sites — the deadlock diagnosis surfaced that runBuiltinModule constructs a runtime PER INVOCATION at builtin_impl.go:143 in addition to the lazy compiler site at builtin_run.go:159, so one probe invocation constructs twice through the production factory"
  - "Phase-2 constructions are gate-staggered (scope B parks until scope A's whole invocation loop finishes) so the retained peak==1 assertion is deterministic — six wrapper-windowed constructions across two concurrent scopes false-positive under -race timing"
  - "The counted comment phrases were reflowed onto one line each so the grep gates match ('35 of the 44-name HEAD registry', '32 of the 41-name baseline registry')"

patterns-established:
  - "Held-mutex ordering probe for lazy-init serialization pins: hold the production mutex, release a parked construction request, prove non-completion while held via a buffered completion channel + one-sided dwell"
  - "Mutation-matrix coverage for a referenced guard: pair-only removal gives a behavioral RED, var+pair removal gives a compile RED — both deterministic, unlike window-based overlap detection"

requirements-completed: [LINT-PARITY-01]

coverage:
  - id: D1
    description: "WR-01 narrative closed: the falsified '-race then proves the production serialization' claim is gone; the TestMain comment states priming removes the race signal and the regression protection is pinned structurally"
    requirement: LINT-PARITY-01
    verification:
      - kind: other
        ref: "grep gates: 'then proves the production serialization' == 0, 'removes the race signal' >= 1, 'pinned structurally' >= 1 (all pass)"
        status: pass
    human_judgment: false
  - id: D2
    description: "WR-01 protection closed: TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes mechanically detects loss of the serialization — held-mutex ordering probe (deterministic) + retained peak==1 overlap counter, delegating to the production constructor through both creation sites"
    requirement: LINT-PARITY-01
    verification:
      - kind: unit
        ref: "tests/internal/adapters/plugin/builtin_run_test.go#TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes (green under -race, 5/5 stability runs + verify-chain run)"
        status: pass
      - kind: other
        ref: "mutation checks: Lock/Unlock pair removed -> RED 3/3 with 'runtime construction must never overlap across scopes'; var+pair removed (plan-literal) -> build failure (undefined: runtimeCreationMu); guard restored git diff --exit-code"
        status: pass
    human_judgment: false
  - id: D3
    description: "WR-02 closed: acceptanceCorpusConfig comment states the measured registry-vs-enabled truth (35/44 HEAD, 32/41 baseline, COMMENTS+UNARY_RPC never enabled, Gate D unaffected via NEW3-in-MINIMAL/BASIC); false coverage claim gone"
    requirement: LINT-PARITY-01
    verification:
      - kind: other
        ref: "grep gates: 'every rule of' == 0, '35 of the 44-name HEAD registry' >= 1, '32 of the 41-name baseline registry' >= 1, 'UNARY_RPC' >= 1 (all pass)"
        status: pass
      - kind: unit
        ref: "tests/internal/api/acceptance_gates_test.go#TestAcceptanceGates (green under -race, 10.7s, baseline oracle 1e9a596 rebuilt)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Shipped behavior unchanged: production code byte-identical to HEAD, both touched packages green under -race, lint 0 issues, no forbidden artifacts"
    requirement: LINT-PARITY-01
    verification:
      - kind: other
        ref: "command: go test -race -count=1 ./internal/adapters/plugin/ ./internal/api/ (both ok); git diff --exit-code internal/adapters/plugin/builtin_run.go; bin/golangci-lint run both packages = 0 issues; git status shows only pre-existing unrelated files"
        status: pass
    human_judgment: false

# Metrics
duration: 34min
completed: 2026-11-21
status: complete
---

# Phase 07 Plan 04: gap closure — evidence narrative integrity Summary

**WR-01/WR-02 closed test-only: the serialization's regression protection restored by a deterministic held-mutex probe (mutation-check proven), the TestMain priming narrative corrected to the falsification experiments, and the corpus-config comment corrected to the measured 35/44 (32/41) registry-vs-enabled truth — production code byte-identical to HEAD**

## Performance

- **Duration:** 34 min
- **Started:** 2026-11-21T04:55:46Z
- **Completed:** 2026-11-21T05:30:37Z
- **Tasks:** 2
- **Files modified:** 3 (all test files; +138/−21)

## Accomplishments
- WR-01 narrative closed: `main_test.go` no longer claims the -race detector proves the production serialization — the comment now states the experiment-verified truth (priming removes the race signal for the rest of the process; the regression protection is pinned structurally), keeping the accurate wazero v1.9.0 lazy-global, frozen-seam, and fork-level happens-before facts
- WR-01 protection closed: `TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes` now mechanically detects loss of the serialization — a held-mutex ordering probe (a construction requested while `runtimeCreationMu` is held must not complete until release; green is structural, guaranteed by the mutex under test) plus the REVIEW's peak==1 overlap counter retained as secondary detector; constructions delegate to the production constructor through BOTH creation sites (`builtin_run.go:159` lazy compiler and `builtin_impl.go:143` per-invocation runtime)
- WR-02 closed: `acceptanceCorpusConfig` comment states the measured registry-vs-enabled truth — STANDARD expands to MINIMAL+BASIC+DEFAULT (policy.go:109) enabling 35 of the 44-name HEAD registry and 32 of the 41-name baseline registry; the COMMENTS (7) and UNARY_RPC (2) registry rules are never enabled; all three NEW3 rules sit inside MINIMAL/BASIC so Gate D is unaffected; the rule-count discipline sentence is preserved
- Mutation checks recorded: Lock/Unlock pair removed → pin RED 3/3 with the planned "runtime construction must never overlap across scopes" message; var+pair removed (plan-literal variant) → build failure (`undefined: runtimeCreationMu`); guard restored and proven byte-identical (`git diff --exit-code`)
- Plan-level verification green: both packages under `go test -race -count=1` ok, `bin/golangci-lint run` 0 issues on both, production code byte-identical, no schemas/coverage/binaries touched

## Task Commits

Each task was committed atomically:

1. **Task 1: WR-01 — TestMain narrative + serialization probe** - `bc3f853` (test)
2. **Task 2: WR-02 — corpus-config comment truth** - `f1baa65` (test)

**Plan metadata:** see docs commit (SUMMARY + STATE/ROADMAP + REQUIREMENTS)

_Note: both tasks test-only; plan type is execute (not tdd) — RED/GREEN gate exempt per the canonical predicate (no non-test source files touched; `tdd_mode` active but the tasks carry no behavior block)._

## TDD Gate Compliance

Not applicable — plan type is `execute` with test-only tasks (no `<behavior>` block, no non-test source files in scope). The canonical behavior-adding predicate is false for both tasks; `tdd_mode` RED/GREEN enforcement does not fire.

## Files Created/Modified
- `internal/adapters/plugin/main_test.go` - corrected TestMain narrative (priming removes the race signal; structural pinning)
- `internal/adapters/plugin/builtin_run_test.go` - held-mutex ordering probe + retained peak==1 overlap counter + gate-staggered two-scope echo coverage
- `internal/api/acceptance_gates_test.go` - corrected acceptanceCorpusConfig coverage comment (comment-only; every assertion byte-identical)

## Decisions Made
- [Rule 1] The REVIEW WR-01 snippet's wrapper-window probe was replaced by a held-mutex ordering probe. Empirical proof from the planned verification chain itself: with the CORRECT mutex the verbatim probe went RED under -race (`expected 1, actual int64(2)` — the in-flight window includes mutex WAIT time, so two concurrent construction requests overlap in the window even with perfectly serialized critical sections), while without the mutex it stayed green 1 of 2 runs (scheduling luck). A goroutine blocked on `sync.Mutex` is additionally NOT "durably blocked" for `synctest.Wait` (goroutine dump: `[sync.Mutex.Lock, synctest bubble 1]` without the durable marker) — a holder across `synctest.Wait` deadlocks the bubble. The corrected probe: phase 1 (real time, outside the bubble) holds the mutex, releases a parked construction request, and asserts non-completion while held via a buffered completion channel and a one-sided 100ms dwell (the dwell only bounds violation latency — the mutex itself guarantees the green outcome); phase 2 (synctest bubble) keeps the two-scope echo coverage with constructions gate-staggered so the retained `peak == 1` assertion is deterministic
- The probe deliberately exercises BOTH production creation sites: the deadlock diagnosis revealed `runBuiltinModule` constructs a per-invocation runtime at `builtin_impl.go:143` (the documented per-invocation-runtime design) in addition to the lazy compiler at `builtin_run.go:159` — one probe invocation constructs twice, both through the production factory
- Mutation matrix adapted to the corrected mechanism: the plan-literal variant (var + pair deleted) is a compile-time RED because the probe holds the var (strictly stronger than the snippet's timing-based signal); the pair-only variant is the behavioral RED with the planned message. Both recorded in place of the plan's single-variant expectation
- Task 2's counted phrases were reflowed onto single lines so the grep gates match literally (the first wording split "35 of / the 44-name…" across lines and failed its own gate)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] REVIEW WR-01 probe snippet unimplementable as specified — replaced with a deterministic held-mutex ordering probe**
- **Found during:** Task 1 verify chain (the plan's own `-race` gate)
- **Issue:** The verbatim snippet's `inFlight`/`peak` window brackets the factory CALL, which includes the time spent WAITING on the production mutex — so with the CORRECT mutex the pin went RED under -race (measured: `expected int(1), actual int64(2)`; the -race run of the verify chain failed), and without the mutex it stayed GREEN 1 of 2 plain runs (the plan's own mutation check could not produce a reliable RED either). A second, independent defect: holding the mutex across `synctest.Wait` deadlocks the bubble, because a mutex-blocked goroutine is not "durably blocked" (proven by the 20s-timeout goroutine dump during fix attempt 2). The diagnosis also surfaced that `runBuiltinModule` constructs a runtime per invocation (`builtin_impl.go:143`), doubling the constructions the probe sees
- **Fix:** Two-phase pin: phase 1 (real time) = held-mutex ordering probe — holder goroutine locks `runtimeCreationMu`, the probe scope's wrapper parks on `proceedConstruction` before delegating, main releases the park, dwells 100ms on a buffered `constructed` channel (non-completion while held is guaranteed by the mutex itself — the green outcome is structural, never timing-dependent), then releases and joins; phase 2 (synctest bubble) = two-scope echo coverage with constructions gate-staggered (B parks until A's whole invocation loop completes) so the retained verbatim `require.EqualValues(t, 1, peak.Load(), "runtime construction must never overlap across scopes")` assertion is stable. The wrapper still delegates to `newRuntimeSerialized` through the lazy creation site — every construction flows through the production constructor
- **Files modified:** internal/adapters/plugin/builtin_run_test.go (test-only; production builtin_run.go restored byte-identical after each mutation)
- **Verification:** pin green under -race 5/5 stability runs + verify-chain run; full package green under -race (38-39s); mutation pair-removed RED 3/3 with the planned message; mutation var+pair-removed build failure; `git diff --exit-code` after each mutation
- **Committed in:** `bc3f853`

---

**Total deviations:** 1 auto-fixed (Rule 1 bug — planned snippet replaced by a mechanism implementing the plan's stated intent: mechanical, deterministic regression detection)
**Impact on plan:** The plan's verbatim-snippet requirement was in conflict with its own verify chain (empirically proven, not assumed); the fix keeps every hard constraint — test name unchanged, peak assertion verbatim and retained, delegation to the production constructor, production byte-identical, honest mutation-record — and strictly dominates the snippet on every mutation variant (deterministic RED where the snippet was flaky, compile RED where the snippet was blind).

### Mutation-check outcome (plan output requirement)

- Attempt 1 (verbatim snippet, guard removed, plain go test): GREEN — no overlap materialized in that run. Recorded honestly per the plan's contingency: the window probe's firing is scheduling-dependent (1 of 2 runs).
- Attempt 2 re-run (guard still removed): RED (`actual int64(2)`) in both the isolated pin and the package-wide run — confirming the REVIEW's "fires whenever constructions actually overlap" claim, but also proving the probe fires with the CORRECT mutex under -race (false positive) — the trigger for the Rule 1 fix above.
- Final corrected mechanism: guard pair removed → RED 3/3, deterministic, with the planned message; guard var+pair removed (plan-literal) → build failure (`undefined: runtimeCreationMu`); guard restored → green 5/5 under -race + verify chain.

## Issues Encountered
- None beyond the documented deviation. The WR-02 comment edit needed one reflow (the counted phrases initially straddled line breaks and failed their own grep gates) — fixed within the task before commit.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- 07-REVIEW-DISPOSITION parity achieved: 0 open-after-plan — WR-01 and WR-02 closed, 0 deferred, 0 rejected; the production fix stands untouched
- The phase's evidence narrative now matches the falsification experiments; future evidence-time dispositions can reason from the corrected comments
- Phase 07 plans 4/4 have summaries — phase verification gates (verify-work, UAT, VALIDATION.md final rows) are the remaining milestone steps

---
*Phase: 07-priemka-paritet-korpusov-i-proizvoditelnost*
*Completed: 2026-11-21*

## Self-Check: PASSED

- Files created/modified verified on disk: `main_test.go`, `builtin_run_test.go`, `acceptance_gates_test.go`, `07-04-SUMMARY.md` — all FOUND
- Commits verified in HEAD ancestry: `bc3f853` (Task 1), `f1baa65` (Task 2), `40b14c9` (docs) — all FOUND
- Measured from ledger `gsd-plan-head-before-07-04` (`9e8ce79`): 2 production commits, diff +138/−21 across 3 files (9542 chars → 2385 tokens at chars/4)
- Plan-level verification re-verified at SUMMARY time: both packages green under -race, lint 0 issues, production byte-identical
