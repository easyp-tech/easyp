---
phase: 03-snyatie-seal-infrastruktury-i-dokumentatsiya
plan: 02
subsystem: testing-and-quality-gates
tags: [errcheck, sourceview, t-cleanup, golangci-lint, staticcheck, seal-grep, changeset-attribution, d-08, d-09, d-10]

# Dependency graph
requires:
  - phase: 03-snyatie-seal-infrastruktury-i-dokumentatsiya
    provides: "Commit (a) 9320361: seal-free live lint config and docs plus the recorded intermediate lint baseline — exactly the pre-existing errcheck at internal/sourceview/session_test.go:29:21"
provides:
  - "Source-level close-error fix: the bare deferred Session.Close() is replaced by a t.Cleanup closure that reports a non-nil close error through t.Errorf"
  - "Repository lint gate at exactly `0 issues.` (exit 0) with no suppression, no exclusion and no config change in commit (b)"
  - "Phase-3 gate evidence: build, four scoped -race packages, empty tidy diff, empty live-surface seal grep, two-commit identity/size/ancestry and the exact eight-file non-planning changeset"
  - "Proof that historical records survive: .spec/plans and .spec/features zero diff in range, .planning/milestones zero diff, 9 tracked files under .spec/features/lint-easypc-infra/**"
affects: [04-priemka-full-parity]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
# Scale: chars/4 over the realized authored source diff (5 added + 1 removed lines, 137 chars).
# Excludes this SUMMARY, matching the 03-01 convention.
actuals:
  tokens: 34
  tasks: 2
  commits: 1

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Close-error reporting in tests: t.Cleanup closure with `if err := session.Close(); err != nil { t.Errorf(...) }`"
    - "Zero-issue lint proven at source: one-file commit equality plus a suppression-free test file and an unchanged config in commit (b)"
    - "Two-commit attribution by exact subject with a non-.planning range equality against the declared eight-file changeset"

key-files:
  created: []
  modified:
    - internal/sourceview/session_test.go

key-decisions:
  - "The close error is handled in t.Cleanup and reported via t.Errorf instead of suppressed, so the 1 -> 0 lint transition is attributable to the source fix (D-08)"
  - "The evidence task ran only the D-09 scoped gates; the full -race battery, schema generation/check and acceptance remain Phase 4"
  - "Commit identity and scope are asserted by exact subject and non-.planning range equality, so intermediate .planning-only commits cannot mask or break the changeset claim"

patterns-established:
  - "One-file lint-fix commit: subject, name-set equality, clean tracked tree, and exact `0 issues.` output all asserted in one gate block"
  - "Phase evidence block: build + scoped -race + tidy -diff + rc-aware seal grep + subject/ancestry/size/range attribution"

requirements-completed: [IMPORT-SEAL-01]

coverage:
  - id: D1
    description: "The discarded Session.Close error is reported through t.Cleanup + t.Errorf, the session is still closed once at test end, and TestSessionSharesRootsAndKeepsChecks passes under -race"
    requirement: IMPORT-SEAL-01
    verification:
      - kind: unit
        ref: "go test -race -count=1 ./internal/sourceview (TestSessionSharesRootsAndKeepsChecks)"
        status: pass
      - kind: other
        ref: "awk: t.Cleanup count 1, bare deferred close count 0, 'close session: %v' count 1; gofmt -l empty"
        status: pass
    human_judgment: false
  - id: D2
    description: "The repository lint gate prints exactly `0 issues.` with exit 0, reached without any suppression directive or config change in commit (b)"
    requirement: IMPORT-SEAL-01
    verification:
      - kind: other
        ref: "bin/golangci-lint run ./... -> `0 issues.`; commit (b) changes only internal/sourceview/session_test.go; no nolint directive in the test file"
        status: pass
    human_judgment: false
  - id: D3
    description: "Phase-3 gate battery green on the committed tree: build, four scoped -race packages, empty tidy diff, empty live-surface seal grep, and the exact eight-file two-commit changeset with historical records intact"
    requirement: IMPORT-SEAL-01
    verification:
      - kind: integration
        ref: "go build ./...; go test -race -count=1 ./internal/sourceview ./internal/core ./internal/protoast ./internal/rules; go mod tidy -diff (empty); seal grep rc=1 over .golangci.yml AGENTS.md .spec/*.md README.md"
        status: pass
      - kind: other
        ref: "git merge-base --is-ancestor 9320361 6f52840; (a)=7 files, (b)=1 file; non-.planning range == declared 8 files; .spec/plans/.spec/features/.planning/milestones zero diff"
        status: pass
    human_judgment: false

duration: 3 min
started: 2026-11-21T14:36:19Z
completed: 2026-11-21T14:38:57Z
plan_head_before: bca520ef464403ecbccefe529f87e28d86405453
plan_head_after: 6f52840c5a90a7b88633d6296154a73a8dcd44ed
status: complete
---

# Phase 3 Plan 02: Fix the Pre-existing Close Errcheck and Prove the Phase Gates Summary

**The last pre-existing lint finding was fixed at its source — `internal/sourceview/session_test.go` now reports `Session.Close()` failures through `t.Cleanup` + `t.Errorf` — taking `bin/golangci-lint run ./...` to exactly `0 issues.` and proving the phase's build, scoped-race, tidy, seal-grep and eight-file changeset gates on the committed tree.**

## Performance

- **Duration:** 3 min
- **Started:** 2026-11-21T14:36:19Z
- **Completed:** 2026-11-21T14:38:57Z
- **Tasks:** 2 (1 test edit+commit, 1 evidence-only)
- **Files modified:** 1 source file (plus this SUMMARY)

## Accomplishments

- `internal/sourceview/session_test.go`: `defer session.Close()` (line 29) replaced by a `t.Cleanup` closure that checks `session.Close()` and calls `t.Errorf("close session: %v", err)` on a non-nil error; the session is still closed once at test end and no other statement, assertion or helper changed.
- Commit (b) `6f52840` has subject `test(03-02): check session close error`, changes exactly one file (5 insertions, 1 deletion), and contains no suppression directive; `.golangci.yml` is untouched by it.
- Lint transition is attributable: baseline at `bca520e` was exactly 1 issue (`internal/sourceview/session_test.go:29:21`, errcheck); after commit (b) `bin/golangci-lint run ./...` prints exactly `0 issues.` with exit 0.
- Phase-3 gates on the committed tree: `go build ./...` green; `go test -race -count=1 ./internal/sourceview ./internal/core ./internal/protoast ./internal/rules` all ok; `go mod tidy -diff` empty; the combined seal grep over `.golangci.yml`, `AGENTS.md`, `.spec/*.md`, `README.md` returns zero lines (rc=1).
- Attribution: commit (a) `9320361` is an ancestor of commit (b), changes exactly 7 files; the combined non-`.planning` changeset from `(a)^` to `(b)` is exactly the eight declared files; `go.mod`/`go.sum` have zero diff in the range.
- Historical record intact: `.spec/plans`, `.spec/features` and `.planning/milestones` have zero diff in the range, and 9 files under `.spec/features/lint-easypc-infra/**` remain tracked.
- The tracked tree was clean before and after the gate battery; no push occurred (D-10).

## Task Commits

Each task was committed atomically:

1. **Task 1: Handle the session close error in the sourceview test and commit it** - `6f52840` (test)
2. **Task 2: Prove the Phase-3 gate battery on the committed tree** - no commit (evidence-only; no tracked file changed)

**Plan metadata:** committed with this SUMMARY (`docs(03-02): complete check session close error plan`).

## Files Created/Modified

- `internal/sourceview/session_test.go` - `t.Cleanup` closure checking `Session.Close()` and reporting a failure via `t.Errorf("close session: %v", err)`; no other change
- `.planning/phases/03-snyatie-seal-infrastruktury-i-dokumentatsiya/03-02-SUMMARY.md` - This summary with the gate evidence

## Gate Evidence

| Gate | Command | Result |
|------|---------|--------|
| Build | `go build ./...` | exit 0 |
| Lint | `bin/golangci-lint run ./...` | exactly `0 issues.`, exit 0 (baseline 1 at `bca520e`) |
| Scoped race | `go test -race -count=1 ./internal/sourceview ./internal/core ./internal/protoast ./internal/rules` | 4/4 ok |
| Tidy | `go mod tidy -diff` | empty |
| Seal grep | `grep -rInE '\beasypc\b\|compiler-lab\|VENDORED' .golangci.yml AGENTS.md .spec/*.md README.md` | rc=1, zero lines |
| Commit (a) | `9320361` `docs(03-01): remove seal infrastructure and document easyprotoc dependency` | 7 files; ancestor of (b) |
| Commit (b) | `6f52840` `test(03-02): check session close error` | exactly `internal/sourceview/session_test.go` |
| Changeset | `git diff --name-only 9320361^ 6f52840 -- ':(exclude).planning'` | exactly the 8 declared files |
| History | `.spec/plans`, `.spec/features`, `.planning/milestones` diff in range | empty; 9 lint-easypc-infra files tracked |

The eight-file combined changeset: `.golangci.yml`, `AGENTS.md`, `.spec/ARCHITECTURE.md`, `.spec/PACKAGES.md`, `internal/core/link_test.go`, `internal/protoast/build.go`, `internal/rules/package_no_import_cycle_test.go`, `internal/sourceview/session_test.go`.

## Decisions Made

- Fixed the close error at source in `t.Cleanup` with `t.Errorf`, matching D-08 and the repository's "never discard a close error" convention, rather than adding an exclusion or suppression.
- Kept the evidence task read-only over the committed tree; only the D-09 scoped gates ran (no full `-race` battery, schema generation/check or acceptance gates — Phase 4).
- Located the two commits by exact subject and asserted range equality excluding `.planning`, so intermediate planning-only commits neither mask nor break the source changeset claim.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The executor's plan commit ledger (`.git/gsd-plan-head-before-03-02`) contained a stale base `6183736` from a prior run, 215 commits behind HEAD. It was reset to the actual pre-plan HEAD `bca520ef464403ecbccefe529f87e28d86405453`, so the SUMMARY's measured `commits` value (1) reflects this plan's real commit range. (Same failure mode documented in `03-01-SUMMARY.md`.)

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- IMPORT-SEAL-01 is satisfied through observable predicates: live config/docs are seal-free, the module is documented, the lint fix is at source, and all D-09 gates are green.
- Phase 3 is complete at 2/2 plans; Phase 4 can run the full parity battery, Gates A-D, schema gates and acceptance against a tree whose only remaining claim is parity, not seal cleanup.
- No blockers; the tracked tree was clean at `6f52840` and no push occurred.

## Self-Check: PASSED

- `internal/sourceview/session_test.go` and this SUMMARY exist on disk.
- Commits `6f52840` (test) and `9320361` (docs, plan 03-01) are ancestors of HEAD.
- `internal/sourceview/session_test.go` contains zero `nolint` directives; the tracked tree is clean.
