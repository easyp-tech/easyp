---
phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya
plan: 01
subsystem: core
tags: [lint-engine, protoast, easypc, parallel, module-rule, suppressions, determinism, tdd]

# Dependency graph
requires:
  - phase: 02-linker-i-marshrutizatsiya-oshibok
    provides: (*Core).linkSources three-outcome classifier (linked trees / COMPILE / (nil,nil) fallback), CompileRuleName, hasPerFileSyntaxError precedence
  - phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli
    provides: vendored internal/easypc (protoc.Linked with syntax trees), protoast (Parse/Read/Build/Outline/OutlineOf/Comments/Owner), openImportFile boundary
provides:
  - Dormant two-pass linked lint engine in internal/core: LintLinked (parallel read pass -> one corpus linkSources -> windows of 256 via internal/parallel -> indexed serial drain in walk order -> ModuleRule.Finish -> D-01 total sort by (path, rule, line, col))
  - New rule contracts in internal/core/lint_model.go: LinkedFileInfo, LintRule, ModuleRule (Collect/Finish), Resolution, LinkedRuleName, plus the binding etalon->new rename-map header for Phase 5 consolidation
  - Fallback read path (lint_linked_read.go): readSource, protoInfoFromSource, protoInfoFromLinked (tree dropped after build), linkedFilesFromImport/linkedFileFromImport over the openImportFile security boundary, importCacheKey/cacheLintedOutline, parseOnTheirOwn
  - protoast-based comment suppression walk (lint_suppressions.go): directiveComments + linkedCommentSuppressions carrying main's pairing machine and byte-identical error texts
  - link.go additive adapter: linkedFiles/linkedOutline/outline (mutex + sync.Once), resolution, usesThrough
  - Core/Options additive fields: linkedRules []LintRule + Options.LinkedRules, importCache *parsedImports
  - TestLintLinked: 12 green -race named cells (7 engine + 5 fallback matrix) with test-double rules
affects: [03-02 suppression parity table, 03-03 lint.use validation parity, 03-04, Phase 4 rule porting, Phase 5 CLI flip]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 11619
  tasks: 2
  commits: 3

# Tech tracking
tech-stack:
  added: [] # zero new dependencies — go mod tidy leaves go.mod/go.sum byte-unchanged (T-03-SC gate held)
  patterns: [two-pass windowed engine (parallel read -> one Link -> windowed model build -> indexed walk-order drain), stateless per-file rules + serial ModuleRule Collect/Finish lifecycle, central suppression filter (per-file ranges applied to findings and re-applied to Finish findings), mutex + sync.Once outline serving under parallel windows, main-wins error-wrapper parity pin (fs.WalkDir: c.lintProtoInfoRead:), D-01 total sort as the single deliberate delta vs the etalon]

key-files:
  created:
    - internal/core/lint_model.go
    - internal/core/lint_linked.go
    - internal/core/lint_linked_read.go
    - internal/core/lint_suppressions.go
    - internal/core/lint_linked_test.go
    - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/tdd-red-evidence-03-01-task1.json
  modified:
    - internal/core/core.go
    - internal/core/link.go

key-decisions:
  - "Rule storage split (research A1 consequence, resolved deliberately): the old core.Rule and the new core.LintRule both require Validate but with different parameter types, so one type can never implement both — the new engine reads a NEW additive Core field linkedRules []LintRule with additive Options.LinkedRules; the old engine keeps reading c.rules and never sees linkedRules, so dormancy holds"
  - "Main-wins error wrapper pinned at the drain: the engine wraps per-file model errors with the exact literal fs.WalkDir: c.lintProtoInfoRead: %w (documented in-code as a deliberate parity pin, not the actual callee name), proven by three exact-string test cells"
  - "shouldIgnore mirrored, not shared: shouldIgnoreLinked(rule LintRule, path) carries byte-for-byte main semantics; the parameter type pins the original to the old contract"
  - "Etalon error texts carried verbatim after identifier renames (readFilesFromImport:/readFileFromImport: wraps in lint_linked_read.go) — texts are the user-visible REQ-2.4 contract, identifiers are internal"
  - "Engine tests construct the Core via New(Options{LinkedRules: ...}) with literal rule-name lists instead of rules.AllRuleNames(): package core cannot import internal/rules (the rules package imports core — an import cycle); the construction shape from the plan is preserved otherwise"

patterns-established:
  - "Linked-engine test doubles: stateless per-file rule doubles share one instance corpus-wide; a ModuleRule double records walk-order Collect paths and counts Validate invocations (engine must never call it); recordingRule observes processed files and imported outlines for fallback cells"
  - "Exact-count race pinning: per-file rule-count maps asserted exactly (never range bounds) under -race with a shared sibling-import outline surface"
  - "COMPILE-vs-fallback routing proof pattern: rules-skip on COMPILE (compile_error_rules_skipped) vs rules-run-on-fell-back-corpus (fallback_rules_still_run) as the two sides of the classifier contract"

requirements-completed: [LINT-ENG-01]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Two-pass linked engine end-to-end (LINT-ENG-01/D-07/D-02): compiling multi-file corpus (one imports a sibling) lints through linked trees — 9 findings across 4 messages + module Finish, byte-identical repeated runs under -race"
    requirement: LINT-ENG-01
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/linked_corpus_deterministic"
        status: pass
    human_judgment: false
  - id: D2
    description: "D-01 deterministic total order: output sorted by (path, rule, line, col) with ALPHA before BETA despite anti-alphabetical rule registration"
    requirement: LINT-ENG-01
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/engine_sort_order"
        status: pass
    human_judgment: false
  - id: D3
    description: "COMPILE routing through the new engine honors the Phase 2 classifier contract: in-proto3 required field yields exactly one etalon-shaped COMPILE issue (4:3, verbatim message), doubles silent, no suppression machinery on COMPILE (A6)"
    requirement: LINT-ENG-01
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/compile_error_rules_skipped"
        status: pass
    human_judgment: false
  - id: D4
    description: "Central comment suppression filter (D-04): easyp:disable above a declaration filters that finding under AllowCommentIgnores=true and keeps both findings under false — both sides of the flag through the protoast owner walk"
    requirement: LINT-ENG-01
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/comment_suppression_central"
        status: pass
    human_judgment: false
  - id: D5
    description: "ModuleRule lifecycle (D-03/D-05): Validate never called, Collect sees every corpus file exactly once in walk order, Finish finding passes shouldIgnore/suppression re-filter and lands in output; ignore_only skips per file"
    requirement: LINT-ENG-01
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/module_rule_lifecycle"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/ignore_only_skips_per_file"
        status: pass
    human_judgment: false
  - id: D6
    description: "Engine race/statelessness matrix (ROADMAP criterion 4, T-03-02): 8-file corpus with shared rule instances and a shared sibling-import outline surface returns exactly 33 findings with exact per-file rule counts, green under -race"
    requirement: LINT-ENG-01
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/race_multi_file"
        status: pass
    human_judgment: false
  - id: D7
    description: "Fallback read path (D-07, T-03-01): missing import surfaces *OpenImportFileError under the exact main-wins prefix fs.WalkDir: c.lintProtoInfoRead:; syntax and read errors carry the same prefix; rules still run on the parseable files of a fell-back corpus; sibling imports serve non-nil outlines through linkedFileFromImport/cacheLintedOutline"
    requirement: LINT-ENG-01
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/fallback_missing_import_command_error"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/fallback_syntax_error_wrapper"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/fallback_read_error"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/fallback_rules_still_run"
        status: pass
      - kind: unit
        ref: "tests/internal/core/lint_linked_test.go#TestLintLinked/fallback_outline_cache"
        status: pass
    human_judgment: false

# Metrics
duration: 20 min
completed: 2026-11-19
status: complete
commits: 3
plan_head_before: e19599c4807d08cfe8775d6bb4f2d10da64dcf78
plan_head_after: c5f7f4b9629166fb6869b7542865257ed883f4e5
---

# Phase 03 Plan 01: Two-Pass Linked Engine Summary

**Dormant two-pass linked lint engine ported from the compiler-lab etalon with renamed identifiers: parallel read -> one linkSources -> windowed (256) model build from linked trees, central comment suppressions via the protoast owner walk, serial ModuleRule lifecycle, and the D-01 total sort — 12 green -race cells, old lint path byte-unchanged**

## Performance

- **Duration:** 20 min
- **Started:** 2026-11-19T13:23:31Z
- **Completed:** 2026-11-19T13:44:05Z
- **Tasks:** 2
- **Files modified:** 8 (5 created, 2 modified, 1 planning evidence file)

## TDD Cycles

- **Task 1 RED (`0353377`, test):** `TestLintLinked` created with the 7 planned named cells plus inert contract bindings (lint_model.go contracts + Options.LinkedRules/Core.linkedRules bindings + a zero-value LintLinked stub). The run compiled and discovered all cells; every cell failed exactly on its planned `require.Len` assertion because the stub returns (nil, nil) — no build/setup/discovery faults. Evidence persisted to `tdd-red-evidence-03-01-task1.json`; `gsd_run check tdd-red-evidence` → **RED_EVIDENCE_OK** (reason `target_test_failed`, go-test format, 10 failing tests incl. both comment-suppression sub-cases). Semantic assessment: target test executed and failed on the planned behavior gap (engine body absent), not a syntax or fixture fault.
- **Task 1 GREEN (`3ac8abf`, feat):** the etalon engine ported into new files under the binding rename map — LintLinked body (walk -> parallel read -> link-or-fallback -> windows of 256 -> indexed serial drain -> Finish loop -> D-01 `slices.SortStableFunc`), lint_linked_read.go (read path + cached import outlines + parseOnTheirOwn + protoInfoFromLinked with tree drop), lint_suppressions.go (protoast Comments/Owner walk + main-verbatim pairing machine), link.go additive adapter (linkedFiles/outline with mutex+sync.Once, resolution, usesThrough). Two of my own new test fixtures were wrong on first GREEN run (see Deviations 1) — fixtures fixed, engine untouched. All 7 cells green under -race; tracer verify re-run end-to-end per the auto-mode feedback gate and passed.
- **REFACTOR:** none — GREEN output already matched go-code-style conventions (import grouping, godoc, wrap style); no behavior-preserving cleanup needed.
- **Task 2 (`c5f7f4b`, test-only):** five fallback-matrix cells added — `fallback_missing_import_command_error`, `fallback_syntax_error_wrapper`, `fallback_read_error` (chmod 0o000 with restore cleanup), `fallback_rules_still_run`, `fallback_outline_cache`. All passed on the first run — no engine defect exposed, so no `fix(03-01)` commit was needed. 12/12 TestLintLinked cells green.

## TDD Gate Compliance

RED→GREEN verified for Task 1: `feat(03-01)` 3ac8abf descends from `test(03-01)` 0353377 (linear history on the working branch; `git merge-base --is-ancestor` holds). RED evidence machine-validated before any implementation commit (`gsd_run check tdd-red-evidence` → RED_EVIDENCE_OK). Task 2 is test-only (no source files in scope; not behavior-adding) → exempt from the RED gate per the MVP+TDD gate definition; the plan scopes its commit to `test(03-01)` with fixes only if a cell exposed a defect — none did. No REFACTOR commit required.

## Accomplishments

- LINT-ENG-01 real and test-proven on the linked path: each file's model built from the linked easypc syntax trees exactly once per file (protoast.Build over Linked.Syntax, tree dropped after build), windows of 256 through internal/parallel.Run, findings reassembled only in the indexed serial drain and totally sorted at engine exit
- ModuleRule contract ported with the lifecycle guarantee: per-file rules run concurrently (stateless doubles proven under -race), Collect runs serially in walk order, Finish runs once with shouldIgnore + suppression re-filter
- Comment suppressions computed per file from protoast Comments/Owner and applied centrally on both sides of AllowCommentIgnores; the skeleton (parseDirectivePrefix, compactDirectiveNames, pairing machine, four error texts) carried byte-identical from main
- Fallback read path proven end-to-end: missing import → *OpenImportFileError exit-2 carrier, syntax/read errors → command errors, all under the exact main-wins wrapper chain; cached import outlines serve sibling imports through the openImportFile security boundary
- Dormancy held: internal/core/lint.go, internal/api/lint.go, internal/api/runtime.go, and everything under internal/easypc byte-unchanged; go.mod/go.sum diff-free (T-03-SC); reflection-free new files (grep 0)

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: failing two-pass linked engine test** - `0353377` (test)
2. **Task 1 GREEN: implement two-pass linked-file lint engine** - `3ac8abf` (feat)
3. **Task 2: fallback matrix and wrapper pins** - `c5f7f4b` (test)

_Plan metadata: docs commit (this SUMMARY + state/roadmap updates + RED evidence record)._

_TDD plan: RED → GREEN → REFACTOR skipped (nothing to clean) → one test-only matrix task._

## Files Created/Modified

- `internal/core/lint_model.go` - LinkedFileInfo/LintRule/ModuleRule/Resolution contracts, LinkedRuleName, binding etalon->new rename-map header, D-03 ModuleRule list godoc
- `internal/core/lint_linked.go` - LintLinked engine body, lintWindow=256, checkedFile/checkFile/collectFile, shouldIgnoreLinked mirror, D-01 total sort
- `internal/core/lint_linked_read.go` - readSource, protoInfoFromSource, linkedFilesFromImport/linkedFileFromImport (cached), importCacheKey, cacheLintedOutline, parseOnTheirOwn, protoInfoFromLinked (tree drop)
- `internal/core/lint_suppressions.go` - linkedDirectiveComment/linkedParsedDirective, directiveComments protoast walk, linkedCommentSuppressions
- `internal/core/link.go` - additive: linkedFiles/linkedOutline/outline, resolution, usesThrough (linkSources and friends untouched)
- `internal/core/core.go` - additive: linkedRules + importCache fields, parsedImports type, Options.LinkedRules + New assignment
- `internal/core/lint_linked_test.go` - TestLintLinked, 12 named cells + engine test doubles
- `.planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/tdd-red-evidence-03-01-task1.json` - machine-validated RED evidence record

## Decisions Made

- Rule storage split on the A1 seam: new engine reads `Core.linkedRules` (additive `Options.LinkedRules`), old engine keeps `Core.rules` — the dual-contract coexistence never crosses; recorded for Phase 5 consolidation
- Main-wins wrapper text `fs.WalkDir: c.lintProtoInfoRead:` pinned now with exact-string cells (research Open Question 3 resolution), documented in-code as a parity pin rather than the real callee
- Etalon wrap texts kept verbatim after renames (`readFilesFromImport:`/`readFileFromImport:` inside lint_linked_read.go) for REQ-2.4 error-chain parity
- Engine tests pass literal rule-name lists instead of `rules.AllRuleNames()`: package core cannot import internal/rules (import cycle); construction otherwise per plan
- `fallback_rules_still_run` observes findings through recorded rule invocations (recordingRule + module double): the missing-import file aborts the drain in walk order (exit-2 semantics), so returned findings cannot exist on that corpus — the plan's "doubles produce their findings" is asserted as ran-on-every-parseable-file, and the error/carrier assertions pin the rest

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Two test-fixture proto sources were invalid on first GREEN run**
- **Found during:** Task 1 (first GREEN test run, before the GREEN commit)
- **Issue:** `linked_corpus_deterministic` used `b.BetaOne` as the qualified type but b.proto declares `package beta` (wrong qualifier → corpus failed linking with a COMPILE issue); `race_multi_file` declared identical message names `FiOne`/`FiTwo` across f1..f7 in the shared package `racepkg` (proto files of one package share one namespace → "already defined" COMPILE issues)
- **Fix:** qualifier corrected to `beta.BetaOne`; race fixtures use per-file unique names (`F<idx>One`/`F<idx>Two`). Fixtures only — the engine needed no change; both corrected cells then passed
- **Files modified:** internal/core/lint_linked_test.go
- **Verification:** full TestLintLinked run green under -race after the fix
- **Committed in:** 3ac8abf (the broken variants never entered git history)

---

**Total deviations:** 1 auto-fixed (1 test-fixture bug, Rule 1)
**Impact on plan:** No scope creep; the fixtures were the only defect and the engine contract is unchanged. No production surface outside internal/core (additive-only in core.go/link.go).

## Issues Encountered

None beyond the fixture fixes above; RED/GREEN/TASK-2 all converged on the first or second run.

## User Setup Required

None - no external service configuration required.

## Verification Log

- `go build ./... && go test -v -race -count=1 ./internal/core/ -run 'TestLintLinked'` → ok, 12/12 named cells PASS, exit 0
- `go test -race -count=1 ./internal/core/` → ok (whole package, no old-path regressions)
- `./bin/golangci-lint run ./internal/core/...` → 0 issues
- `go mod tidy && git diff --exit-code go.mod go.sum` → zero diff (REQ-1.2/T-03-SC)
- Dormancy: `git diff --stat internal/core/lint.go internal/api/lint.go internal/api/runtime.go` empty; `git status --porcelain internal/easypc/` empty
- Source assertions: `c.lintProtoInfoRead: %w` count 1 (main-wins wrapper present), `c.protoInfoRead` count 0 (etalon literal absent), `reflect.` count 0 in all three new/adapter files, `func (c *Core) linkSources` Phase 2 signature intact with additive-only link.go diff
- TDD gate: RED 0353377 → GREEN 3ac8abf with ancestry; evidence RED_EVIDENCE_OK; tracer verify re-run end-to-end (auto-mode feedback gate) → pass
- Commits measured: `git rev-list --count e19599c..HEAD` = 3

## Next Phase Readiness

- Ready for 03-02 (suppression parity table): the protoast walk + pairing machine are in place; the parity oracle is main's byte-identical fixture table re-driven through LintLinked
- Phase 4 rule porting consumes the proven surface: LintRule/ModuleRule contracts, LinkedFileInfo with Resolved/ImportedFiles, the windowed engine with exact-count race precedent
- Phase 5 flip inherits: the rename map in lint_model.go, the main-wins wrapper pins, and the dormancy baseline (old path untouched, byte-verifiable at any commit)

---
*Phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya*
*Completed: 2026-11-19*

## Self-Check: PASSED

- All 5 created engine/test files plus the RED evidence record and this SUMMARY exist on disk
- Commits 0353377 (RED), 3ac8abf (GREEN), c5f7f4b (Task 2), ebfce28 (docs) are ancestors of HEAD
- `gsd_run check evaluation-scope --plan 03-01 --commits-only` resolved (plan-subject commits on this branch)
- All task acceptance criteria re-run after all tasks: green (see Verification Log)
- Plan-level <verification> suite re-run: all green; LINT-ENG-01 marking deferred by the #2388 shared-ID gate until sibling plans 03-02..03-04 produce their summaries

