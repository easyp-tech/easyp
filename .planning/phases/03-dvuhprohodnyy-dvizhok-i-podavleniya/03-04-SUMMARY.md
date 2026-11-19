---
phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya
plan: 04
subsystem: rules
tags: [validation-parity, lint-use, err-invalid-rule, parity-guard, phase-closure, audit, test-only]

# Dependency graph
requires:
  - phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya (plan 03-01)
    provides: Options.LinkedRules/Core.linkedRules consumption side (core.New takes LintRule values directly, no error surface) — the pre-validated-list contract the guard pins
  - phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya (plans 03-02/03-03)
    provides: green sibling suites the closure gates re-verify; no engine fix mandates carried
provides:
  - Validation parity guard TestLintUseValidationParity (7 green -race named cells): unknown rule/group/except/ignore_only names in lint.use fail with core.ErrInvalidRule identity (errors.Is) and main's exact wrapped texts (validateConfigRuleNames: use:/except:/ignore_only: prefixes); PACKAGE_NO_IMPORT_CYCLE special case pinned until Phase 6; valid config constructs both engines with no second name validation; registry sanity cell
  - Phase-closure audit recorded: old lint path + internal/easypc byte-unchanged (dormancy), go.mod/go.sum tidy-zero, schemas drift-free, golangci-lint 0 issues, full -race suite green except the documented pre-existing wazero race in internal/adapters/plugin (Phase-1 deferred item, out of phase scope)
  - Phase-3 requirement surface complete: LINT-ENG-01, LINT-ENG-05, LINT-ENG-06 all plan-proven; ROADMAP criterion 3 is explicit test evidence
affects: [Phase 4 rule porting (validation unchanged), Phase 5 CLI flip (PACKAGE_NO_IMPORT_CYCLE removal owns the special case, schemas consume the same registry), Phase 6/7 acceptance gates]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 1094
  tasks: 2
  commits: 1

# Tech tracking
tech-stack:
  added: [] # zero new dependencies — go mod tidy leaves go.mod/go.sum byte-unchanged (T-03-SC gate held)
  patterns: [keep-untouched-plus-prove parity guard (the passing guard IS the proof; any failure is drift evidence), exact-text error pinning via require.ErrorIs sentinel identity + require.ErrorContains on the exact wrapped prefixes, consumption-side construction pin (core.New with LinkedRules doubles, no error surface to assert), machine-checked phase-closure gate chain]

key-files:
  created:
    - internal/rules/validation_parity_test.go
  modified:
    - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-VALIDATION.md

key-decisions:
  - "Validation parity = keep-untouched + prove, not reimplement: rules.New remains THE validator; the guard exercises it through the public constructor and asserts ErrInvalidRule identity plus exact per-selection wrapped texts so any rewording/relocation/weakening fails the guard"
  - "No second engine-side name validation pinned as a contract: the valid_config cell constructs the linked engine via core.New(Options{LinkedRules: ...}) and asserts the absence of any error surface — one name surface (the registry) feeds config validation, directive validation (knownLintRules) and future schemas"
  - "PACKAGE_NO_IMPORT_CYCLE special case pinned by the guard as main's deliberate exception; removal is Phase 6's flip-plan decision, never a silent change here"
  - "Pre-existing wazero -race failure in internal/adapters/plugin recorded against its Phase-1 deferred item, never fixed in-phase (T-03-SC forbids dependency changes; sealed/3rd-party scope) and never silently skipped"

patterns-established:
  - "Parity-guard cell shape: table cells carry {name, cfg, wantPrefix, wantName}; error cells assert sentinel identity + selection prefix + offending name — three independent failure modes per cell"
  - "Phase-closure audit chain: full -race suite (with documented out-of-scope exceptions), dormancy diff-stat, easypc porcelain, tidy zero-diff, schema-gen diff against schemas/, golangci-lint — each gate's result recorded in the plan summary"

requirements-completed: [LINT-ENG-06]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Validation parity guard (LINT-ENG-06/D-06): unknown rule/group names in lint.use, unknown except and ignore_only names all fail with errors.Is core.ErrInvalidRule and main's exact wrapped texts; PACKAGE_NO_IMPORT_CYCLE stays 'is not implemented'; a valid config returns non-nil rules/ignoreOnly and constructs the linked engine with no second name validation; the registry covers every group key and rule name — 7 green -race cells"
    requirement: LINT-ENG-06
    verification:
      - kind: unit
        ref: "tests/internal/rules/validation_parity_test.go#TestLintUseValidationParity/unknown_rule_in_use"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/validation_parity_test.go#TestLintUseValidationParity/unknown_group_in_use"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/validation_parity_test.go#TestLintUseValidationParity/unknown_name_in_except"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/validation_parity_test.go#TestLintUseValidationParity/unknown_name_in_ignore_only"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/validation_parity_test.go#TestLintUseValidationParity/package_no_import_cycle_special_case"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/validation_parity_test.go#TestLintUseValidationParity/valid_config_still_constructs"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/validation_parity_test.go#TestLintUseValidationParity/registry_covers_all_names"
        status: pass
    human_judgment: false
  - id: D2
    description: "The existing main validation test (the parity oracle's own suite) stays green untouched — guard-mutation criterion: weakening the validation path turns it red"
    verification:
      - kind: unit
        ref: "tests/internal/rules/selection_test.go#TestInvalidLintNamesAreRejectedBeforeFiltering"
        status: pass
    human_judgment: false
  - id: D3
    description: "Phase-closure invariants machine-checked: dormancy (lint.go/api lint.go/runtime.go diff-stat empty, internal/easypc porcelain empty), dependency tidiness (go mod tidy zero-diff), schema drift-freedom (schema-gen diff against schemas/), golangci-lint 0 issues"
    verification:
      - kind: other
        ref: "command: git diff --stat internal/core/lint.go internal/api/lint.go internal/api/runtime.go (empty) && git status --porcelain internal/easypc/ (empty)"
        status: pass
      - kind: other
        ref: "command: go mod tidy && git diff --exit-code go.mod go.sum"
        status: pass
      - kind: other
        ref: "command: tmp=$(mktemp -d) && go run ./cmd/easyp schema-gen --out-dir \"$tmp\" && diff -r \"$tmp\" schemas/ (exit 0)"
        status: pass
      - kind: other
        ref: "command: ./bin/golangci-lint run ./... (0 issues, exit 0)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Full -race suite gate: every phase-scope package green (internal/rules, internal/core, internal/api, ...); the single failing package is internal/adapters/plugin with the pre-existing wazero GetWazeroVersion data race — classified out of phase scope per the plan's own carve-out and recorded against its Phase-1 deferred item"
    verification:
      - kind: other
        ref: "command: go build ./... && go test -race -count=1 ./internal/... (all packages ok except internal/adapters/plugin — documented exception)"
        status: pass
    human_judgment: true
    rationale: "The suite-level pass/fail depends on classifying the internal/adapters/plugin wazero race as pre-existing and out of phase scope (Phase-1 deferred item 01-vendoring.../deferred-items.md:3-5). The commands ran and their output is recorded here, but ratifying that classification — rather than treating it as a phase-03 regression — is a verifier judgment."

# Metrics
duration: 12 min
completed: 2026-11-19
status: complete
commits: 1
plan_head_before: 19c2b81155c1d4ecf502d6d8d4fe4a552189f4dc
plan_head_after: 0718751f411f16129049fa19497fddd59544f9dc
---

# Phase 03 Plan 04: Validation Parity Guard and Phase Closure Summary

**lint.use validation parity pinned as a 7-cell guard through the untouched rules.New chain (ErrInvalidRule identity + exact wrapped texts, PACKAGE_NO_IMPORT_CYCLE special case included), plus the phase-closure audit: dormancy/tidy/schema/lint all clean, full -race suite green except the documented pre-existing wazero race**

## Performance

- **Duration:** 12 min
- **Started:** 2026-11-19T14:32:12Z
- **Completed:** 2026-11-19T14:44:00Z
- **Tasks:** 2
- **Files modified:** 2 (1 created test file, 1 planning file VALIDATION.md)

## TDD Cycles

- **Task 1 (`0718751`, test):** `TestLintUseValidationParity` created with the 7 planned named cells and run before commit. The MVP+TDD gate predicate `gsd_run query task.is-behavior-adding` returned `false` (test-only `<files>`: only `validation_parity_test.go`), so the RED gate is exempt (03-01 Task 2 / 03-02 / 03-03 precedent; plan frontmatter is `type: execute`, not `type: tdd`). The plan's own RED design anticipated the shape: "cells are expected GREEN against the untouched path (the guard IS the proof); any failure means main's validation path drifted, which is RED evidence to capture and resolve" — all 7 cells passed on the first -race run, so no `tdd-red-evidence-03-04-task1.json` was required (evidence persistence is scoped to the failure case) and no `fix(03-04)` arose.
- **Task 2 (no commit — audit-only):** the five phase-closure gates were run and their results recorded below; the plan expects no repo changes from this task and none occurred (`git status` shows only pre-existing working-tree noise unrelated to this plan).
- **REFACTOR:** none — test-only plan; zero production files changed.

## TDD Gate Compliance

Task 1 is test-only (`<files>` lists only `validation_parity_test.go`); the MVP+TDD gate verb returned `is_behavior_adding=false`, exempting it from the RED gate. Plan frontmatter is `type: execute` (not `type: tdd`), so the plan-level RED/GREEN commit-pattern gate does not apply. Commit sequence: `test(03-04)` 0718751. No `feat(03-04)` exists because no cell exposed a validation-path defect — the plan scopes fixes to the failure case, which did not occur. No REFACTOR commit required.

## Accomplishments

- ROADMAP criterion 3 is guard-proven instead of "we didn't touch it": an unknown rule or group name in `lint.use` returns main's exact validation error through the untouched `rules.New` chain — `errors.Is(err, core.ErrInvalidRule)` plus the exact `validateConfigRuleNames: use: `/`except: `/`ignore_only: ` wrapped texts, asserted per selection so any rewording fails the guard
- The `PACKAGE_NO_IMPORT_CYCLE` special case (`ErrInvalidRule` + "is not implemented") is explicitly pinned as main's deliberate exception; Phase 6 owns its removal and inherits a guard that will fail if it drifts earlier
- The no-second-validator contract is pinned from the consumption side: a valid config yields non-nil `[]core.Rule`/`ignoreOnly` outputs, and `core.New(core.Options{LinkedRules: ...})` constructs a usable engine with no error surface to even assert against — one name surface (the registry) feeding config validation, directive validation (`KnownLintRules` wired at runtime.go:74) and future schemas
- Registry sanity cell proves `AllLintUseValues()` covers every group key (MINIMAL, BASIC, DEFAULT, COMMENTS, UNARY_RPC) and every `AllRuleNames()` entry
- Phase closure machine-checked (see Closure Gates below): the old lint path and `internal/easypc` are byte-unchanged, `go.mod`/`go.sum` are tidy-zero, generated schemas are drift-free, golangci-lint is clean, and the full -race suite is green in every phase-scope package with the single documented out-of-scope exception
- LINT-ENG-06 marked complete — with it, all three Phase-3 requirements (LINT-ENG-01, LINT-ENG-05, LINT-ENG-06) are plan-proven

## Closure Gates (Task 2 results)

| Gate | Command | Result |
|------|---------|--------|
| 1. Full suite | `go build ./... && go test -race -count=1 ./internal/...` | All packages ok — except `internal/adapters/plugin` (pre-existing wazero race, see Deviations/Issues); phase-scope packages incl. rules/core/api all green |
| 2. Dormancy | `git diff --stat internal/core/lint.go internal/api/lint.go internal/api/runtime.go` + `git status --porcelain internal/easypc/` | Both empty — old lint path and sealed tree byte-unchanged (D-04 continuation + Phase 1 seal) |
| 3. Dependencies | `go mod tidy && git diff --exit-code go.mod go.sum` | Zero diff — REQ-1.2/T-03-SC invariant held |
| 4. Schemas | `tmp=$(mktemp -d) && go run ./cmd/easyp schema-gen --out-dir "$tmp" && diff -r "$tmp" schemas/` | Exit 0 — drift-free (no rule names changed in this phase) |
| 5. Lint | `./bin/golangci-lint run ./...` | `0 issues.` — no new findings; the two documented Phase-1 scoped exclusions unchanged (internal/easypc path exclusion; lint_helpers.go unused-check exclusion) |

## Task Commits

Each task was committed atomically:

1. **Task 1: lint.use validation parity guard through rules.New** - `0718751` (test)

_Task 2 is audit-only by design (empty `<files>`, "no repo changes expected") — its results are recorded in this SUMMARY and in the 03-VALIDATION.md map row, committed with the plan metadata._

_Plan metadata: docs commit (this SUMMARY + STATE/ROADMAP/REQUIREMENTS + VALIDATION row, then the self-check pass)._

_Test-only plan: no RED/GREEN/REFACTOR production cycle — see TDD Gate Compliance._

## Files Created/Modified

- `internal/rules/validation_parity_test.go` - TestLintUseValidationParity: 7 named parity cells + parityLintRule double (test-only; zero production files changed)
- `.planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-VALIDATION.md` - LINT-ENG-06 row filled with the real -run pattern and 7 named cells; nyquist-ready state set (`nyquist_compliant: true`, `wave_0_complete: true`, wave-0 checklist ticked)

## Decisions Made

- Parity by keep-untouched + prove: the guard calls the public constructor (`rules.New`) and asserts sentinel identity AND exact text prefixes, so rewording, relocating, or weakening the validation chain fails it — satisfying the plan's GUARD MUTATION criteria without touching production code
- The no-second-validator prohibition is pinned behaviorally (valid config constructs the engine with no error surface) rather than by source-grep, because `core.New` exposes no validation hook to grep for — the absence IS the contract
- Out-of-scope failure handling per the plan's own carve-out: the wazero race is recorded with its deferred-items reference, never "fixed" in-phase (dependency changes are forbidden by T-03-SC; the package is outside the phase's file scope)
- 03-VALIDATION.md set to nyquist-ready: all three Per-Task Verification Map rows are real green -race suites with named cells

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Gate 1 (full suite) fails in `internal/adapters/plugin` only: `TestBuiltinUniversalOutputCompatibility` — "race detected" inside `github.com/tetratelabs/wazero@v1.9.0/internal/version.GetWazeroVersion()`. This is the Phase-1-acknowledged pre-existing issue (`.planning/phases/01-vendoring-kompilyatora-i-raschlenenie-modeli/deferred-items.md` lines 3-5; STATE.md blocker "Pre-existing wazero data race (internal/adapters/plugin, -race)"). No phase-03 commit touches `internal/adapters/` (verified: `git log 19c2b81~3..HEAD -- internal/adapters/` is empty). Recorded here per the plan's instruction; the dedicated subsystem fix remains owned before the Phase 7 full-suite gate.

## User Setup Required

None - no external service configuration required.

## Verification Log

- `go build ./... && go test -race -count=1 -v ./internal/rules/ -run 'TestLintUseValidationParity'` → ok, exit 0, 7/7 named cells PASS
- `go test -race -count=1 ./internal/rules/ -run 'TestInvalidLintNamesAreRejectedBeforeFiltering'` → ok, exit 0 (existing main test untouched and green)
- `git status --porcelain internal/rules/ -- ':!internal/rules/validation_parity_test.go'` → empty (zero production changes in internal/rules)
- Source assertions: `require.ErrorIs(t, err, core.ErrInvalidRule)` x1 (sentinel via errors.Is), `require.ErrorContains(t, err, tt.wantPrefix)` x1 with `validateConfigRuleNames: use: ` x3, `validateConfigRuleNames: except: ` x1, `validateConfigRuleNames: ignore_only: ` x1
- Full suite: all `./internal/...` packages ok under -race except internal/adapters/plugin (documented pre-existing wazero race above)
- Dormancy: `git diff --stat internal/core/lint.go internal/api/lint.go internal/api/runtime.go` empty; `git status --porcelain internal/easypc/` empty
- `go mod tidy && git diff --exit-code go.mod go.sum` → zero diff (REQ-1.2/T-03-SC)
- Schema drift: `go run ./cmd/easyp schema-gen --out-dir "$tmp" && diff -r "$tmp" schemas/` → exit 0
- `./bin/golangci-lint run ./...` → 0 issues
- MVP+TDD gate: `gsd_run query task.is-behavior-adding .planning/phases/.../03-04-PLAN.md --pick is_behavior_adding` → `false` (RED gate exempt)
- Requirements shared-ID gate: `requirements.ready-ids` → LINT-ENG-06 ready (no sibling plan declares it)
- Commits measured: `git rev-list --count 19c2b81..HEAD` = 1 at SUMMARY-write time

## Next Phase Readiness

- Phase 03 is complete (4/4 plans): engine (03-01), suppression parity (03-02), parallel-correctness/COMPILE matrix (03-03), validation parity + closure audit (this plan) — ready for phase verification gates
- Phase 4 rule porting inherits: the proven LintRule/ModuleRule surface, the statefulness tripwire (`race_shared_doubles_identity`), an untouched validation path it must not weaken, and the lint_helpers.go exclusion note (helpers gain callers in Phase 4 — remove the paths exclusion then)
- Phase 5 flip inherits: the registry as the single name surface (schemas can consume it directly), the PACKAGE_NO_IMPORT_CYCLE pin (removal decision documented), and the byte-verifiable dormancy baseline

---
*Phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya*
*Completed: 2026-11-19*
