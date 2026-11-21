---
phase: 01-merge-main-i-primirenie-putey
plan: 02
subsystem: testing
tags: [full-battery, race-detector, acceptance-gates, golden-94, position-parity, merge-attribution, expectation-freeze]

# Dependency graph
requires:
  - phase: 01-merge-main-i-primirenie-putey
    provides: "Merge commit 994c503 (origin/main #253 joined; blob-exact internal/core resolution; lint contour zero-diff)"
provides:
  - "Attributable evidence that HEAD's code tree equals the 994c503 merge tree (docs-only descendant, behind=0)"
  - "Full -race battery green: exit 0, 41 ok packages, 0 FAIL, 0 race-warning blocks, 0 skips"
  - "D-09 parity gates green by exact test name: TestAcceptanceGates (A-D x2, baseline 1e9a596), TestLintGoldenFullRegistry (94), TestPositionRuneColumns, TestProtoastPositionsMatchLegacyContract, core matrix/suppression, CLI stdout/link/exit-code/COMPILE"
  - "Zero-diff expectation freeze versus PRE_MERGE_BASE for internal/rules, internal/protoast and every listed core/API expectation file"
affects: [phase-02, phase-03, phase-04]

# Actuals (#2632) — chars/4 over the realized diff (SUMMARY + planning metadata). Never a harness token count.
actuals:
  tokens: 5725
  tasks: 2
  # Measured at SUMMARY write: git rev-list --count c9308f6..HEAD = 0. This is an evidence-only
  # plan (files_modified is just this SUMMARY); the docs-only metadata commit carrying it is not
  # a task commit.
  commits: 0
  plan_head_before: c9308f6a856a323927d271a8d69d92b5773b788a
  plan_head_after: c9308f6a856a323927d271a8d69d92b5773b788a

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Evidence-only gate plan: no source edit, no expectation edit, one SUMMARY artifact; every gate bound to an exact test name plus a log under /tmp"
    - "Attribution by path-qualified diff: HEAD may be a docs-only descendant, so code-tree equality is proven by `git diff MERGE_SHA HEAD -- ':(exclude).planning'` being empty"

key-files:
  created:
    - .planning/phases/01-merge-main-i-primirenie-putey/01-02-SUMMARY.md
  modified: []

key-decisions:
  - "Evidence-only execution: both tasks are gate runs whose only artifact is this SUMMARY, so no per-task commit exists; the plan's own prohibition keeps HEAD on the merge commit or a docs-only descendant, and no source or expectation file was touched"
  - "The battery was executed exactly once, after build/vet/tidy and before any named gate, from a clean tracked tree; the acceptance package was then re-run alone by exact name"
  - "Named gates assert the exact top-level `=== RUN` line in a `-v` log because a filter matching nothing exits 0 with no tests run; zero exit status alone is not accepted"
  - "Observed timings were faster than the cold estimate (internal/api 211.1 s vs 309 s; TestAcceptanceGates 9.64 s with a warm build cache), and no rerun was needed or performed"

patterns-established:
  - "Gate evidence table: command, exit code, exact RUN-line count, PASS/FAIL counts, pinned constant observed, log path"
  - "Expectation freeze as a pathspec diffstat against the merge commit's first parent, paired with single-occurrence pinned-constant greps"

requirements-completed: [IMPORT-REC-01]

coverage:
  - id: D1
    description: "The tested tree is the 994c503 merge tree: two parents in the recorded order, origin/main an ancestor, behind count 0, and no diff outside .planning from MERGE_SHA to HEAD; tracked tree clean before and after"
    requirement: IMPORT-REC-01
    verification:
      - kind: integration
        ref: "git log --grep '^merge: origin/main (#253 Use compiler for breaking check)' -1; git rev-list --count HEAD..origin/main == 0; git diff --name-only 994c503 HEAD -- ':(exclude).planning' -> empty"
        status: pass
    human_judgment: false
  - id: D2
    description: "go build ./..., go vet ./... and go mod tidy -diff are clean on the merge tree"
    requirement: IMPORT-REC-01
    verification:
      - kind: integration
        ref: "go build ./... rc 0; go vet ./... rc 0 (empty stderr); go mod tidy -diff -> empty output"
        status: pass
    human_judgment: false
  - id: D3
    description: "One full -race battery is green: exit 0, 41 ok package lines, 0 FAIL lines, 0 race-detector warning blocks, 0 skips"
    requirement: IMPORT-REC-01
    verification:
      - kind: integration
        ref: "go test -race -count=1 -timeout 20m ./... (log /tmp/01-02-full-battery.log: 41 ok / 0 FAIL / 0 'WARNING: DATA RACE')"
        status: pass
    human_judgment: false
  - id: D4
    description: "Acceptance Gates A-D x2 pass against the ratified 1e9a596 baseline with baseline presence asserted, never skipped"
    requirement: IMPORT-REC-01
    verification:
      - kind: integration
        ref: "internal/api/acceptance_gates_test.go#TestAcceptanceGates (exact RUN line present, PASS 9.64s, baseline commit 1e9a596 present in git)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Golden registry gate passes with the pinned 94 findings and the constant unchanged; no golden re-capture"
    requirement: IMPORT-REC-01
    verification:
      - kind: unit
        ref: "internal/rules/golden_full_registry_test.go#TestLintGoldenFullRegistry; goldenFindingCount = 94 exactly one occurrence"
        status: pass
    human_judgment: false
  - id: D6
    description: "D-07 position gates and the core/CLI parity matrix pass by exact test name; expectations are byte-frozen versus the pre-merge base"
    requirement: IMPORT-REC-01
    verification:
      - kind: unit
        ref: "TestPositionRuneColumns (19 PASS lines), TestProtoastPositionsMatchLegacyContract (11), TestLintMatrix/TestLintErrorPathFidelity/TestLintSuppressionParity (43), TestLintStdoutGolden/TestLintLinkMatrix/TestLintLinkMatrixExitCodes/TestPrintIssuesCompileGolden (16); freeze diffstat empty"
        status: pass
    human_judgment: false

# Metrics
duration: 20min
completed: 2026-11-21
status: complete
---

# Phase 1 Plan 2: Post-merge full battery and parity-gate evidence Summary

**Committed merge 994c503 proven behavior-preserving: a 41-package `-race` battery green with 0 FAIL and 0 races, plus Gates A–D ×2, golden 94, both D-07 position gates and the core/CLI parity matrix all passing with expectations byte-frozen against the pre-merge base.**

## Performance

- **Duration:** ~20 min (battery 3.7 min wall, `internal/api` 211.1 s; acceptance 9.6 s with a warm build cache)
- **Started:** 2026-11-21T11:23:00Z (approx. — derived from the first battery log timestamp 11:28:27Z)
- **Completed:** 2026-11-21T11:43:00Z (approx.)
- **Tasks:** 2 (both evidence-only gate runs; no authored code or test change)
- **Files modified:** 1 (this SUMMARY); planning metadata `STATE.md`, `ROADMAP.md`, `REQUIREMENTS.md` updated by the executor

## Accomplishments

- Proved the tested code tree is exactly the merge tree: `MERGE_SHA=994c503968a20ecdc1b4c1ebb2a9f06f23973386` with parents `79e353a5eb1f001b065df89ca4645d962a74ea55` (first) and `5d36266d71e85f8f0fc3fa41f787a03ee5c68a79` (origin/main), `HEAD=c9308f6a856a323927d271a8d69d92b5773b788a` a docs-only descendant, behind count 0, and zero diff versus `MERGE_SHA` outside `.planning`.
- Ran the single budgeted full `-race` battery on a clean tracked tree: exit status 0, 41 `ok` package lines, 0 `FAIL` lines, 0 `WARNING: DATA RACE` blocks, 0 skips, 9 packages with no test files.
- Verified build/vet/tidy cleanliness on the merged tree: `go build ./...` rc 0, `go vet ./...` rc 0 with empty stderr, `go mod tidy -diff` printed nothing (no dependency drift after the merge).
- Re-ran every D-09 gate by exact test name with `-v`: acceptance Gates A–D ×2 against baseline `1e9a596` (present, never skipped), golden `94`, both D-07 position gates, core matrix/error-path/suppression, and CLI stdout/link/exit-code/COMPILE — all green with the exact `=== RUN` line asserted so a no-op filter could not pass.
- Froze expectations: zero diffstat versus `PRE_MERGE_BASE` for `internal/rules`, `internal/protoast` and every listed core/API expectation file; `goldenFindingCount = 94` and `acceptanceBaselineCommit = "1e9a596"` each appear exactly once; the tracked tree stayed clean.

## Gate Evidence

### Identifiers

| Item | Value |
|------|-------|
| `HEAD` | `c9308f6a856a323927d271a8d69d92b5773b788a` (`docs(01-01): complete merge origin/main plan`) |
| `MERGE_SHA` | `994c503968a20ecdc1b4c1ebb2a9f06f23973386` (`merge: origin/main (#253 Use compiler for breaking check)`) |
| Merge parents | `79e353a5eb1f001b065df89ca4645d962a74ea55`, `5d36266d71e85f8f0fc3fa41f787a03ee5c68a79` |
| `PRE_MERGE_BASE` | `79e353a5eb1f001b065df89ca4645d962a74ea55` (= `MERGE_SHA^1`) |
| `origin/main` | `5d36266d71e85f8f0fc3fa41f787a03ee5c68a79` |
| Commits from merge to HEAD | 1 (docs-only, inside `.planning`) |
| Module pins | `easyprotoc v0.0.0-20261104120806-f378f3f3a05f`; `google.golang.org/protobuf v1.37.0-rc.1` |

### Task 1 — attribution, build/vet/tidy, full battery

| Gate | Command | Exit | Result |
|------|---------|------|--------|
| Merge identity | `git log --format=%H --grep='^merge: origin/main (#253 Use compiler for breaking check)' -1` | 0 | `994c503…`, exactly 2 parents in recorded order |
| Ancestry / aliveness | `git merge-base --is-ancestor origin/main HEAD`; `git rev-list --count HEAD..origin/main` | 0 | ancestor true; behind = 0 |
| Tree attribution | `git diff --name-only 994c503 HEAD -- ':(exclude).planning'` | 0 | empty (code tree = merge tree) |
| Clean tree (before) | `git status --porcelain --untracked-files=no` | 0 | empty |
| Build | `go build ./...` | 0 | clean |
| Vet | `go vet ./...` | 0 | clean; stderr 0 bytes |
| Tidy | `go mod tidy -diff` | 0 | empty output |
| Full battery | `go test -race -count=1 -timeout 20m ./...` | 0 | 41 `ok`, 0 `FAIL`, 0 `WARNING: DATA RACE`, 0 `--- SKIP`; `internal/api` 211.120 s; 9 packages `[no test files]`; log `/tmp/01-02-full-battery.log`, `rc` also in `/tmp/01-02-battery-rc.txt` |
| Clean tree (after) | `git status --porcelain --untracked-files=no` | 0 | empty |

### Task 2 — named parity gates (each claimed by exact `=== RUN` line)

| # | Gate | Command (`go test -race -count=1 …`) | Exit | RUN lines | PASS / FAIL | Log |
|---|------|-------------------------------------|------|-----------|-------------|-----|
| 1 | Acceptance Gates A–D ×2 | `./internal/api -run '^TestAcceptanceGates$' -v` | 0 | 1 | 1 / 0 (9.64 s) | `/tmp/01-02-acceptance.log` |
| 2 | Golden 94 | `./internal/rules -run '^TestLintGoldenFullRegistry$' -v` | 0 | 1 | 1 / 0 | `/tmp/01-02-golden.log` |
| 3a | D-07a rune columns | `./internal/protoast -run '^TestPositionRuneColumns$' -v` | 0 | 1 | 19 / 0 | `/tmp/01-02-pos-protoast.log` |
| 3b | D-07b legacy contract | `./internal/core -run '^TestProtoastPositionsMatchLegacyContract$' -v` | 0 | 1 | 11 / 0 | `/tmp/01-02-pos-core.log` |
| 4 | Core matrix / error path / suppression | `./internal/core -run 'TestLintMatrix\|TestLintErrorPathFidelity\|TestLintSuppressionParity' -v` | 0 | 1 each | 43 / 0 | `/tmp/01-02-core-matrix.log` |
| 5 | CLI stdout / link / exit codes / COMPILE | `./internal/api -run 'TestLintStdoutGolden\|TestLintLinkMatrix$\|TestLintLinkMatrixExitCodes\|TestPrintIssuesCompileGolden' -v` | 0 | 1 each | 16 / 0 | `/tmp/01-02-api-matrix.log` |

Baseline presence: `git cat-file -e 1e9a596^{commit}` rc 0; the acceptance log records the baseline comparison ran (raw JSON-line diagnostic after NEW3 filter: only-main 0, only-head 0), so the 1e9a596 worktree build was executed, not skipped.

Pinned constants observed:

- `internal/rules/golden_full_registry_test.go`: `goldenFindingCount = 94` — exactly 1 occurrence of the assignment, value unchanged.
- `internal/api/acceptance_gates_test.go:29`: `acceptanceBaselineCommit = "1e9a596"` — exactly 1 occurrence, value unchanged.

### Task 2 gate 6 — expectation freeze

`git diff --stat PRE_MERGE_BASE HEAD -- internal/rules internal/protoast internal/core/position_parity_test.go internal/core/lint_linked_matrix_test.go internal/core/lint_suppressions_parity_test.go internal/api/acceptance_gates_test.go internal/api/lint_link_matrix_test.go internal/api/lint_stdout_golden_test.go` → empty.

Tracked tree after all gates: `git status --porcelain --untracked-files=no` → empty. No expectation, golden, or parity fixture was edited, regenerated, or re-captured.

## Task Commits

This plan is evidence-only: `files_modified` is just this SUMMARY, and no source, test, expectation, or commit was added by it.

1. **Task 1: Commit attribution, build/vet/tidy, and the single full -race battery** — no authored commit (gate-only task; evidence recorded above)
2. **Task 2: Named parity gates** — no authored commit (gate-only task; evidence recorded above)

**Plan metadata:** the docs commit carrying this SUMMARY and the planning state updates (created after this file).

Measured at SUMMARY write: `git rev-list --count c9308f6a856a323927d271a8d69d92b5773b788a..HEAD` = 0.

## Files Created/Modified

- `.planning/phases/01-merge-main-i-primirenie-putey/01-02-SUMMARY.md` — this evidence record (gate tables, identifiers, deferred items)

No production source, test, or expectation file changed. Logs live under `/tmp` (`/tmp/01-02-*.log`, `/tmp/01-02-battery-rc.txt`, `/tmp/01-02-vet.err` is empty).

## Decisions Made

- Evidence-only execution: since both tasks produce only this SUMMARY, there are no per-task commits; the plan's prohibition ("no follow-up fix, tidy, or bump commit") and attribution gate ("HEAD stays the merge commit or a docs-only descendant") are both honored.
- Both tasks' gates were run strictly sequentially, never in parallel: build/vet/tidy first, then the battery alone, then the acceptance gate alone, then the remaining named gates.
- Exact-name evidence: each gate log was checked for its top-level `=== RUN` line, because `go test -run` on a non-matching pattern exits 0 with no tests run.
- The battery was not re-run. The observed wall time was shorter than the cold planning estimate (warm build cache), which is an observation, not a reason to re-run or to relax any assertion.
- Requirement `IMPORT-REC-01` is marked complete now (the shared-ID deferral from plan 01-01 is satisfied by this SUMMARY).

## Deviations from Plan

None - plan executed exactly as written. Command forms were equivalent to the verify blocks (plain redirection for the battery log instead of `tee`; identical assertions and log contents). No gate failed, so no recorded-failure path was taken.

## Issues Encountered

- `gsd_run` is not on `PATH` in this environment; it was invoked as `$HOME/.zcode/gsd-core/bin/gsd_run` (same environment note as plan 01-01).
- `task` and `docker` are absent (orchestrator decision 5); direct `go`/`git` equivalents were used, and no Docker/Hadolint work was attempted.
- `golangci-lint` is not a Phase 1 gate (D-09). Main's pre-existing errcheck finding at `internal/sourceview/session_test.go:29` is inherited unchanged and remains owned by Phase 3.
- Push to `origin/feat/lint-easypc-infra` remains out of scope for Phase 1.
- Timing observations vs the planning oracle: `internal/api` took 211.1 s (vs 309 s cold) and `TestAcceptanceGates` 9.64 s (vs ~5 min) because the Go build cache was warm from the battery; the exact RUN assertions confirm the tests actually executed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- IMPORT-REC-01's behavior-preservation clause is proven and attributable: the merge tree builds, vets, tidies clean, is green under one full `-race` battery, and every pinned parity oracle still passes byte-identically.
- Phase 2 can start from the reconciled branch; the import flip has a recorded behavior baseline (41 ok / 0 FAIL, golden 94, acceptance A–D on 1e9a596) to regress against.
- No blockers introduced by this plan. The deferred items above (errcheck ownership, push scope, `task`/`docker` absence) are recorded, not silently dropped.

## Known Stubs

None.

---
*Phase: 01-merge-main-i-primirenie-putey*
*Completed: 2026-11-21*

## Self-Check: PASSED

Verified: `01-02-SUMMARY.md` exists; `MERGE_SHA 994c503968a20ecdc1b4c1ebb2a9f06f23973386` is an ancestor of `HEAD`; all seven gate logs under `/tmp/01-02-*.log` exist and are non-empty.
