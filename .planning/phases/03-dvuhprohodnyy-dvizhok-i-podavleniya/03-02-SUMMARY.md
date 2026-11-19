---
phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya
plan: 02
subsystem: core
tags: [lint-engine, suppressions, parity-table, protoast, eof-matrix, ignore-only, tdd, test-only]

# Dependency graph
requires:
  - phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya (plan 03-01)
    provides: LintLinked engine, protoast comment walk (lint_suppressions.go), engine test doubles, LintRule/ModuleRule contracts
  - phase: 02-linker-i-marshrutizatsiya-oshibok
    provides: linkSources feeding the engine's linked path (the parity fixtures drive it end to end)
provides:
  - Fixture-proven suppression parity (LINT-ENG-05/D-04/D-05): main's byte-identical suppression fixtures re-driven through LintLinked — 9-cell EOF matrix, byte-exact directive error texts, policy independence, three dialects, ignore/ignore_only on walk/per-file/Finish surfaces, tab position pin — 21 green -race named rows
  - messagePascalCase + finishIssueModuleRule parity doubles mirroring the oracle rule exactly (top-level non-PascalCase messages; preset Finish findings)
  - Order-stability fix for the pre-existing flaky fallback_rules_still_run assertion (~50% failure rate before)
affects: [03-03 lint.use validation parity, 03-04, Phase 4 rule porting, Phase 5 CLI flip]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 3737
  tasks: 2
  commits: 3

# Tech tracking
tech-stack:
  added: [] # zero new dependencies — go mod tidy leaves go.mod/go.sum byte-unchanged (T-03-SC gate held)
  patterns: [main-vs-new parity table (old-path fixtures as the byte-identical oracle re-driven through LintLinked), oracle-rule-mirroring doubles (same name predicate + top-level walk as rules.MessagePascalCase), preset-Finish ModuleRule double for Finish-side re-filter cells, line-based suppression pin with rune-column literal + documented byte-vs-rune divergence (A4)]

key-files:
  created:
    - internal/core/lint_suppressions_parity_test.go
  modified:
    - internal/core/lint_linked_test.go

key-decisions:
  - "Parity double mirrors the oracle rule, not a generic per-message reporter: messagePascalCase reproduces rules.MessagePascalCase's exact regexp and top-level-only walk so the old-path expectations carry over verbatim (a generic double mis-reports the well-named Outer in nested_closing_brace)"
  - "Config-side cells construct Options directly (Ignore/IgnoreOnly/AllowCommentIgnores literals) so the acceptance source assertions grep real option usage, not helper parameters"
  - "Tab pin asserts a hand-computed rune column literal (taberr.proto:4:2) with the A4 byte-vs-rune divergence documented in-code as a Phase 5 observable — never derived from the engine under test"
  - "No engine fix was needed: zero divergences found in lint_suppressions.go or lint_linked.go across all 21 fixtures; lint_linked.go stays untouched for 03-03 (wave zero-overlap discipline held)"

patterns-established:
  - "Parity-table shape: table rows carry {name, allow, body, wantErr, want} with fixtures verbatim from the oracle tests; error rows assert full path:line:column message texts, count rows additionally assert every finding's RuleName+Path so COMPILE routing cannot pass silently"
  - "Nested-dir corpus pattern: os.MkdirAll before writeCorpusInto for sub/ prefixed fixtures (repo convention from breaking_git_import_test.go)"

requirements-completed: [LINT-ENG-05]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Comment-suppression parity (LINT-ENG-05/D-04): main's 9-cell EOF matrix, byte-exact directive error texts (path:line:column shape), policy independence both ways of AllowCommentIgnores, and all three dialects (easyp:disable/enable, nolint:, buf:lint:ignore with legacy name validation) reproduce main's expected outcomes through LintLinked"
    requirement: LINT-ENG-05
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_suppressions_parity_test.go#TestLintLinkedSuppressionParity/eof_with_newline..nested_closing_brace (9 rows) + malformed_directive_exact_text + unknown_name_exact_text + policy_allow_suppresses + policy_disallow_keeps + nolint_dialect_suppresses + buf_dialect_suppresses + nolint_unknown_name_error"
        status: pass
    human_judgment: false
  - id: D2
    description: "Config-side suppression parity (D-05) and position pin: walk-level Ignore filter, IgnoreOnly exact-file and directory-prefix on per-file findings and ModuleRule Finish output, Finish-side comment re-filter both ways of the policy flag, and the tab directive position pinned to protoast's rune convention with the byte-vs-rune divergence recorded as a Phase 5 observable"
    requirement: LINT-ENG-05
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_suppressions_parity_test.go#TestLintLinkedSuppressionParity/config_ignore_walk_filter + ignore_only_dir_prefix + ignore_only_finish_refilter + finish_comment_refilter/{allow_drops,disallow_keeps}_finish_finding + tab_directive_position_pin"
        status: pass
    human_judgment: false

# Metrics
duration: 19 min
completed: 2026-11-19
status: complete
commits: 3
plan_head_before: 618373630aa8489219c5845e21fa7eb68cf04ec6
plan_head_after: 2016db0804c98baea52d271b80baf6485389a0be
---

# Phase 03 Plan 02: Suppression Parity Summary

**Suppression parity fixture-proven: main's byte-identical suppression fixtures re-driven through LintLinked — 21 green -race rows covering the 9-cell EOF matrix, byte-exact directive error texts, policy independence, all three dialects, ignore/ignore_only on walk/per-file/Finish surfaces, and the rune-column tab pin — zero engine divergences found**

## Performance

- **Duration:** 19 min
- **Started:** 2026-11-19T13:49:53Z
- **Completed:** 2026-11-19T14:09:30Z
- **Tasks:** 2
- **Files modified:** 2 (1 created, 1 modified) + 1 planning file (VALIDATION.md)

## TDD Cycles

- **Task 1 (`4f0f12a`, test):** `TestLintLinkedSuppressionParity` written with 16 named rows (9-cell EOF matrix verbatim, 2 byte-exact error-text rows, 2 policy-independence rows, 3 dialect rows). The first run failed 3 cells — all diagnosed as test-authoring bugs, not engine divergences: two rows expected the wrong fixture line (miscounted `message Valid {}` occupying line 2), and the generic per-message double mis-reported the well-named `Outer` in `nested_closing_brace` because the oracle rule (`rules.MessagePascalCase`) only reports top-level non-PascalCase messages. The double was rewritten to mirror the oracle rule exactly; all 16 rows then passed. No RED evidence JSON was persisted: the MVP+TDD gate predicate returned `is_behavior_adding=false` for this plan's tasks (test-only `<files>`), and no cell failed on a planned engine assertion after the authoring fixes — the green table IS the parity proof, per the plan's "cells that pass immediately are fine".
- **Task 2 (`2016db0`, test-only):** five config-side cells added — `config_ignore_walk_filter`, `ignore_only_dir_prefix`, `ignore_only_finish_refilter`, `finish_comment_refilter` (nested both-ways-of-the-flag subtests), `tab_directive_position_pin`. Two authoring iterations: the `sub/` fixtures needed `os.MkdirAll` before writing (repo convention), and the IgnoreOnly/AllowCommentIgnores acceptance assertions required direct `Options` construction in the config cells (helper parameters don't grep as option usage). 21/21 rows green under -race; no engine defect exposed, so no `fix(03-02)` engine commit was needed.
- **REFACTOR:** none — test-only plan; no production code changed.

## TDD Gate Compliance

Both tasks are test-only (`<files>` lists only `lint_suppressions_parity_test.go`); `gsd_run query task.is-behavior-adding` returned `false`, so the MVP+TDD RED gate is exempt (same precedent as 03-01's Task 2). The plan frontmatter is `type: execute` (not `type: tdd`), so the plan-level RED/GREEN commit-pattern gate does not apply. Commit sequence: `fix(03-02)` 3b62bef (pre-existing flake, below) → `test(03-02)` 4f0f12a → `test(03-02)` 2016db0. No `feat(03-02)` exists because the engine under test already landed in 03-01 and no parity fix was required — the plan scopes engine fixes to `fix(03-02)` only when the table exposes a divergence, which it did not.

## Accomplishments

- ROADMAP criterion 2 ("подавления идентичны текущему main") is now fixture-proven for the new engine: every old-path suppression expectation reproduced through `LintLinked` with LintRule doubles
- The 9-cell EOF matrix (EOF with/without newline, block comment at EOF, orphan enable, unknown trailing rule, unattached disable, unpaired attached disable, marker-in-string, nested closing brace) passes byte-for-byte through the protoast owner walk — the only algorithmic delta vs main's reflection walk is proven equivalent on the whole matrix
- Directive error texts are byte-identical including the `path:line:column` prefix shape (malformed directive, unknown rule); policy independence holds both ways; all three dialects suppress identically with legacy names validated against the known set
- Config-side parity proven on all three surfaces: walk-level `Ignore` filter (rules never run there — observed via the recording double), `IgnoreOnly` exact-file and directory-prefix on per-file findings and on `ModuleRule.Finish` output, and the Finish-side comment re-filter under both `AllowCommentIgnores` values
- The tab fixture pins the new engine's rune-based column explicitly (tab advances by exactly one → marker at column 2) and documents the byte-vs-rune divergence as a Phase 5 observable, not silently accepted (A4)
- Dormancy held: `internal/core/comment_suppressions.go`, `internal/core/lint.go`, `internal/api/lint.go`, and `internal/easypc/` byte-unchanged; `lint_suppressions.go` untouched (no divergence found); go.mod/go.sum diff-free (T-03-SC)

## Task Commits

Each task was committed atomically:

1. **Pre-task fix: flaky fallback assertion** - `3b62bef` (fix, pre-existing bug found at baseline)
2. **Task 1: comment-suppression parity table** - `4f0f12a` (test)
3. **Task 2: config suppression parity and tab position pin** - `2016db0` (test)

_Plan metadata: docs commit (this SUMMARY + state/roadmap/VALIDATION updates)._

_Test-only plan: no RED/GREEN/REFACTOR production cycle — see TDD Gate Compliance._

## Files Created/Modified

- `internal/core/lint_suppressions_parity_test.go` - TestLintLinkedSuppressionParity: 21 named parity rows + messagePascalCase/finishIssueModuleRule doubles + writeCorpusInto helper
- `internal/core/lint_linked_test.go` - fallback_rules_still_run: order-insensitive ElementsMatch for the concurrently-run Validate observations (flake fix)
- `.planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-VALIDATION.md` - LINT-ENG-05 row filled with the real -run pattern and 21 named cells

## Decisions Made

- Parity double mirrors the oracle rule exactly (`messagePascalCaseName` regexp, top-level messages only) instead of reporting every message — the old-path expectations are only comparable when the double's finding set matches `rules.MessagePascalCase`'s
- Config cells construct `Options` directly (literal `Ignore`/`IgnoreOnly`/`AllowCommentIgnores` fields) so the option usage is greppable in the parity file itself
- Tab pin asserts the hand-computed literal `taberr.proto:4:2` with the A4 divergence comment in-code — expectations are never derived from the engine under test
- Count rows assert every finding's `RuleName` and `Path`, so a fixture accidentally routed to COMPILE output cannot pass a count-only assertion

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Pre-existing flaky assertion in fallback_rules_still_run (~50% failure rate)**
- **Found during:** baseline check, before Task 1 (4 of 8 `-race` runs failed)
- **Issue:** the cell asserted `recordingRule.observed()` equals `["a.proto", "b.proto"]` in order, but `Validate` runs inside the parallel window (engine contract, D-02) — observation order is scheduler-dependent; only `Collect` is walk-order serial. The cell's own comment pinned intent as "ran-on-every-parseable-file" (a set), not order
- **Fix:** `assert.ElementsMatch` replaces the order-sensitive `assert.Equal`; walk order stays pinned by the module `Collect` assertion in the same cell; comment documents the split. Verified stable across 10 consecutive `-race` runs
- **Files modified:** internal/core/lint_linked_test.go (outside the plan's declared files_modified — test-only, minimal, required because the whole-package gate is this plan's acceptance criterion; the engine itself is correct and lint_linked.go remains untouched for 03-03)
- **Verification:** 10/10 green `-race` runs after the fix; full package green
- **Committed in:** 3b62bef

---

**Total deviations:** 1 auto-fixed (1 pre-existing test bug, Rule 1)
**Impact on plan:** No scope creep; the fix removes a flake that would have randomly failed this plan's own acceptance gates and CI. Production surface unchanged.

## Issues Encountered

- Initial Task 1 run: 3 cells failed on test-authoring bugs (fixture line miscounts; double not mirroring the oracle rule's top-level/non-PascalCase predicate). Fixed before any commit; none was an engine divergence.
- Initial Task 2 run: 3 config cells failed on the missing `sub/` parent directory (`writeCorpus` writes files flat); resolved with the repo's `os.MkdirAll` convention via a `writeCorpusInto` helper.

## Handoff Findings for 03-03

None on `lint_linked.go`: the Finish re-filter, `shouldIgnoreLinked`, walk filter, and suppression plumbing behaved exactly as specified on all 21 fixtures. No engine fix mandate arises from this plan.

## User Setup Required

None - no external service configuration required.

## Verification Log

- `go build ./... && go test -race -count=1 ./internal/core/ -run 'TestLintLinkedSuppressionParity'` → ok, exit 0, 21 named rows PASS (stable across 3 verbose runs: 21/21/21)
- `go test -race -count=1 ./internal/core/` → ok (whole package, no regressions)
- `./bin/golangci-lint run ./internal/core/...` → 0 issues
- `go mod tidy && git diff --exit-code go.mod go.sum` → zero diff (REQ-1.2/T-03-SC)
- `git diff --stat internal/core/comment_suppressions.go internal/core/lint.go internal/api/lint.go internal/core/dom.go` → empty (old suppression path + old engine byte-unchanged)
- `git status --porcelain internal/easypc/` → empty (sealed tree untouched)
- `grep -c 'reflect\.' internal/core/lint_suppressions.go` → 0 (walk stays reflection-free)
- Flake stability: 10 consecutive `go test -race -count=1 ./internal/core/ -run 'TestLintLinked'` runs green after 3b62bef
- Commits measured: `git rev-list --count 6183736..HEAD` = 3

## Next Phase Readiness

- Ready for 03-03 (lint.use validation parity): the suppression surface is fixture-proven; no engine findings to carry; the parity doubles and `writeCorpusInto` are reusable conventions in package core
- Phase 4 rule porting inherits the proven central-filter semantics (walk, per-file, Finish) and the exact-count named-row style for batch parity gates
- Phase 5 flip inherits the pinned rune-column convention (A4 observable) and the byte-exact directive error texts

---
*Phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya*
*Completed: 2026-11-19*

## Self-Check: PASSED

- internal/core/lint_suppressions_parity_test.go exists on disk
- Commits 3b62bef, 4f0f12a, 2016db0 (tasks) and 2f197d3 (docs) are ancestors of HEAD
- `gsd_run check evaluation-scope --plan 03-02 --commits-only` resolved with 4 plan-subject commits
- All task acceptance criteria re-run after all tasks: green (see Verification Log)
- Plan-level <verification> suite re-run: all green; LINT-ENG-05 marking deferred by the #2388 shared-ID gate pending sibling summaries (03-03/03-04)
