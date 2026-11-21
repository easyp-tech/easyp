---
phase: 02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii
plan: 02
subsystem: testing
tags: [full-battery, race-detector, acceptance-gates, golden-94, position-parity, expectation-freeze, evidence-only, easyprotoc]

# Dependency graph
requires:
  - phase: 02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii
    provides: "Atomic swap commit 5342455: seven consumer files import public easyprotoc packages and the 648-file internal/easypc tree is deleted"
provides:
  - "Attributable evidence that the tested tree is the swap tree (HEAD is a docs-only descendant; zero non-.planning diff from 5342455)"
  - "One full -race battery green: exit 0, 34 ok packages, 0 FAIL, 0 race blocks, only the known TestGenerateGoHeaderUsesCompilerVersion environment skip, 3 no-test packages"
  - "All D-09 named gates green by exact RUN/PASS lines: TestAcceptanceGates (A-D x2 against 1e9a596), TestLintGoldenFullRegistry (94), TestPositionRuneColumns (19), TestProtoastPositionsMatchLegacyContract (11), core matrix (43), CLI matrix (16)"
  - "Zero-expectation-edit freeze: the swap modified set is exactly the seven D-02 files; internal/rules and internal/api are byte-identical to PRE_SWAP_BASE"
  - "Phase predicates re-proven on the committed tree: grep-empty, no tracked or committed internal/easypc, tidy idempotent, exact pins, zero replace/vendor, upstream cmd/easyprotoc builds as package main"
affects: [03-seal-infra-and-docs, 04-acceptance-full-parity]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
# Scale: chars/4 over the realized authored diff (this SUMMARY; no source or test file changed).
actuals:
  tokens: 5150
  tasks: 2
  commits: 0
  plan_head_before: 7e46374d30282de5b793321bb58c921be93826a5
  plan_head_after: 7e46374d30282de5b793321bb58c921be93826a5

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Evidence-only gate plan: no source or expectation edit; every gate bound to an exact top-level `=== RUN` line plus `--- PASS`, with a /tmp log"
    - "Expectation freeze proven by modified-set equality, path-scoped zero-diff versus the pre-swap base, and single-occurrence pinned constants"
    - "Battery-shape assertions from a `-v` log: package-line counts, FAIL/race/skip counts, and no-test counts, with one allow-listed environment skip"

key-files:
  created:
    - .planning/phases/02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii/02-02-SUMMARY.md
  modified: []

key-decisions:
  - "Evidence-only execution: both tasks are gate runs whose only artifact is this SUMMARY, so no per-task commit exists; 5342455 remains the last production commit and HEAD is a docs-only descendant"
  - "The battery ran exactly once with the ratified `-race -count=1 -timeout 20m -v` flags from a clean tracked tree; no gate was re-run to fish for green"
  - "Every named gate is claimed by its exact top-level `=== RUN` and `--- PASS` line because a `go test -run` filter matching nothing exits 0 with no tests"
  - "The stale plan ledger (c896eee, previous milestone) was replaced with the true base 7e46374 before the metadata commit, following the 02-01 precedent"
  - "The known skip is environmental, not a carve-out: protoc-gen-go is absent from PATH, and TestGenerateGoHeaderUsesCompilerVersion is the only `--- SKIP` line"

patterns-established:
  - "Gate evidence table: command, exit code, exact RUN-line presence, total PASS/FAIL lines, pinned constant observed, log path"
  - "Battery-shape gate: 34 ok / 0 FAIL / 0 race / 3 no-test with exactly one allow-listed environment skip"

requirements-completed: [IMPORT-DEP-01, IMPORT-CONS-01, IMPORT-VENDRM-01]

# Coverage metadata (#1602) — one entry per shipped evidence deliverable.
coverage:
  - id: D1
    description: "The tested tree is the swap tree: SWAP_SHA resolved by subject, an ancestor of HEAD, and zero non-.planning diff from SWAP_SHA to HEAD; tracked tree clean before and after every run"
    requirement: IMPORT-CONS-01
    verification:
      - kind: integration
        ref: "git diff --name-only 5342455410a50ae30f1f4f5eeb1f49f7ac20eb4f HEAD -- ':(exclude).planning' -> empty; git status --porcelain --untracked-files=no -> empty"
        status: pass
    human_judgment: false
  - id: D2
    description: "Build, vet, tidy idempotence and exact pins are clean on the committed swap tree"
    requirement: IMPORT-DEP-01
    verification:
      - kind: integration
        ref: "go build ./... rc 0; go vet ./... rc 0 with 0-byte stderr; go mod tidy -diff -> empty; easyprotoc v0.0.0-20261104120806-f378f3f3a05f; protobuf v1.37.0-rc.1; replace=0, vendor=0"
        status: pass
    human_judgment: false
  - id: D3
    description: "One full -race battery is green with the measured post-deletion shape: 34 ok, 0 FAIL, 0 race blocks, only the known environment skip, 3 no-test packages"
    requirement: IMPORT-CONS-01
    verification:
      - kind: integration
        ref: "go test -race -count=1 -timeout 20m -v ./... rc 0 (log /tmp/02-02-full-battery.log: 34 ok / 0 FAIL / 0 'WARNING: DATA RACE' / 1 SKIP / 3 '?')"
        status: pass
    human_judgment: false
  - id: D4
    description: "Acceptance Gates A-D pass twice per binary against the ratified 1e9a596 baseline, with baseline presence asserted and never skipped"
    requirement: IMPORT-CONS-01
    verification:
      - kind: integration
        ref: "internal/api/acceptance_gates_test.go#TestAcceptanceGates (exact RUN line; PASS 9.42s; 1e9a596^{commit} present; constant single occurrence; raw diagnostic only-main 0 / only-head 0)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Golden registry gate passes byte-exactly with the pinned 94 findings and no re-capture"
    requirement: IMPORT-CONS-01
    verification:
      - kind: unit
        ref: "internal/rules/golden_full_registry_test.go#TestLintGoldenFullRegistry (exact RUN line; PASS 0.02s; goldenFindingCount = 94 exactly one occurrence)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Both D-07 position gates pass unchanged (protoast rune columns and core legacy contract)"
    requirement: IMPORT-CONS-01
    verification:
      - kind: unit
        ref: "TestPositionRuneColumns (19 PASS lines); TestProtoastPositionsMatchLegacyContract (11 PASS lines)"
        status: pass
    human_judgment: false
  - id: D7
    description: "Core and CLI parity matrices pass by exact test name with zero test failures"
    requirement: IMPORT-CONS-01
    verification:
      - kind: integration
        ref: "TestLintMatrix + TestLintErrorPathFidelity + TestLintSuppressionParity (43 PASS lines); TestLintStdoutGolden + TestLintLinkMatrix + TestLintLinkMatrixExitCodes + TestPrintIssuesCompileGolden (16 PASS lines)"
        status: pass
    human_judgment: false
  - id: D8
    description: "Zero expectations moved: the swap modified set is exactly the seven D-02 files, internal/rules and internal/api are byte-identical to the pre-swap base, and the phase predicates are re-proven"
    requirement: IMPORT-VENDRM-01
    verification:
      - kind: integration
        ref: "git diff --diff-filter=M df23d64 5342455 -> exactly seven files; zero diff for internal/rules+internal/api; grep/ls-files/cat-file absence; tidy diff empty; go list -m exact; cmd/easyprotoc package name main"
        status: pass
    human_judgment: false

# Metrics
duration: 7min
completed: 2026-11-21
status: complete
commits: 0
plan_head_before: 7e46374d30282de5b793321bb58c921be93826a5
plan_head_after: 7e46374d30282de5b793321bb58c921be93826a5
---

# Phase 2 Plan 2: Full battery and parity evidence on the committed swap tree Summary

**The public-module swap proven behavior-preserving on commit 5342455: one 34-package `-race` battery green with 0 FAIL, 0 races and one known environment skip, plus Gates A-D x2 against baseline 1e9a596, golden 94 byte-exact, both D-07 position gates, the 43-line core matrix and the 16-line CLI matrix all passing with expectations provably unmoved.**

## Performance

- **Duration:** 7 min (start 13:10:35Z, completed 13:18:35Z; battery wall ~4 min)
- **Tasks:** 2 (both evidence-only gate runs; no authored code or test change)
- **Files modified:** 1 (this SUMMARY); planning metadata `STATE.md`, `ROADMAP.md`, `REQUIREMENTS.md` updated by the executor
- **Battery:** `internal/api` 221.603 s dominates; `internal/adapters/plugin` 64.752 s; every other package under 4 s

## Commit under test

| Item | Value |
|---|---|
| Subject | `feat(lint): consume public easyprotoc module; drop vendored easypc tree` |
| `SWAP_SHA` | `5342455410a50ae30f1f4f5eeb1f49f7ac20eb4f` |
| `PRE_SWAP_BASE` | `df23d64aa107f17024f7cfcd3bb6198fe92f7a91` |
| HEAD during evidence | `7e46374d30282de5b793321bb58c921be93826a5` (docs-only descendant: `docs(02-01): refresh STATE current-position prose`) |
| Attribution | `git diff --name-only 5342455 HEAD -- ':(exclude).planning'` empty, before and after all gates |
| Changeset | 655 paths: 648 `D` + 7 `M` + 0 `A`; 15 insertions, 22921 deletions |
| Branch | `feat/lint-easypc-infra` kept as-is; no push (D-08) |
| Rollback | `git revert 5342455` (single commit, branch unpushed) |

## Gate Evidence

Every command ran on the committed tree from a clean tracked tree; exit codes are as observed.

| # | Gate | Command | Expected | Observed | Exit |
|---|---|---|---|---|---|
| 1 | Swap commit resolvable | `git log --format=%H --grep='^feat(lint): consume public easyprotoc module; drop vendored easypc tree$' -1` | non-empty | `5342455…` | 0 |
| 2 | Swap is ancestor of HEAD | `git merge-base --is-ancestor 5342455 HEAD` | true | true | 0 |
| 3 | Attribution | `git diff --name-only 5342455 HEAD -- ':(exclude).planning'` | empty | empty | 0 |
| 4 | Tracked tree before | `git status --porcelain --untracked-files=no` | empty | empty | 0 |
| 5 | Build | `go build ./...` | 0 | 0 | 0 |
| 6 | Vet | `go vet ./...` | 0, empty stderr | 0, 0 bytes stderr | 0 |
| 7 | Tidy idempotence | `go mod tidy -diff` | empty | empty | 0 |
| 8 | Pins | `go list -m -f '{{.Version}}' github.com/easyp-tech/easyprotoc` / `google.golang.org/protobuf` | `v0.0.0-20261104120806-f378f3f3a05f` / `v1.37.0-rc.1` | exact match | 0 |
| 9 | Package census | `go list ./...` / test-package filter | 37 total / 34 test | 37 / 34 | 0 |
| 10 | Full battery | `go test -race -count=1 -timeout 20m -v ./...` (log `/tmp/02-02-full-battery.log`) | rc 0, 34 ok, 0 FAIL, 0 `--- FAIL`, 0 race, 0 unexpected skips, 3 no-test | all exact (see shape table) | 0 |
| 11 | Acceptance Gates A-D x2 | `go test -race -count=1 ./internal/api -run '^TestAcceptanceGates$' -v` (log `/tmp/02-02-acceptance.log`) | exact RUN + PASS | `=== RUN   TestAcceptanceGates`; `--- PASS: TestAcceptanceGates (9.42s)` | 0 |
| 12 | Baseline commit present | `git cat-file -e 1e9a596^{commit}` | rc 0 | rc 0 | 0 |
| 13 | Acceptance constant | `grep -c 'acceptanceBaselineCommit = "1e9a596"' internal/api/acceptance_gates_test.go` | 1 | 1 | 0 |
| 14 | Golden 94 | `go test -race -count=1 ./internal/rules -run '^TestLintGoldenFullRegistry$' -v` (log `/tmp/02-02-golden.log`) | exact RUN + PASS | `=== RUN   TestLintGoldenFullRegistry`; `--- PASS` (0.02s) | 0 |
| 15 | Golden constant | `grep -c 'goldenFindingCount = 94' internal/rules/golden_full_registry_test.go` | 1 | 1 | 0 |
| 16 | D-07a positions | `go test -race -count=1 ./internal/protoast -run '^TestPositionRuneColumns$' -v` (log `/tmp/02-02-pos-protoast.log`) | exact RUN + PASS | `--- PASS: TestPositionRuneColumns`; 19 PASS lines (18 subtests) | 0 |
| 17 | D-07b positions | `go test -race -count=1 ./internal/core -run '^TestProtoastPositionsMatchLegacyContract$' -v` (log `/tmp/02-02-pos-core.log`) | exact RUN + PASS | `--- PASS`; 11 PASS lines (10 subtests) | 0 |
| 18 | Core matrix | `go test -race -count=1 ./internal/core -run 'TestLintMatrix\|TestLintErrorPathFidelity\|TestLintSuppressionParity' -v` (log `/tmp/02-02-core-matrix.log`) | 3 RUN + 3 PASS, 0 FAIL | 43 PASS lines, 3 top-level PASS, 0 `--- FAIL` | 0 |
| 19 | CLI matrix | `go test -race -count=1 ./internal/api -run 'TestLintStdoutGolden\|TestLintLinkMatrix$\|TestLintLinkMatrixExitCodes\|TestPrintIssuesCompileGolden' -v` (log `/tmp/02-02-api-matrix.log`) | 4 RUN + 4 PASS, 0 FAIL | 16 PASS lines, 4 top-level PASS, 0 `--- FAIL` | 0 |
| 20 | Modified set | `git diff --name-only --diff-filter=M df23d64 5342455` | exactly the seven D-02 files | exactly seven (see list) | 0 |
| 21 | Rules/API freeze | `git diff --name-only df23d64 5342455 -- internal/rules internal/api` | empty | empty | 0 |
| 22 | Vendored grep | `git grep -n 'internal/easypc' -- '*.go'` | empty | empty | no match |
| 23 | Vendored tracked | `git ls-files internal/easypc` | empty | empty | 0 |
| 24 | Vendored in swap tree | `git cat-file -e 5342455:internal/easypc` | fails | fails | non-zero |
| 25 | Tidy re-proof | `go mod tidy -diff` | empty | empty | 0 |
| 26 | Escape hatches | `grep -c '^replace' go.mod` / `'^vendor'` | 0 / 0 | 0 / 0 | 0 |
| 27 | Upstream cmd builds | `go build -o /dev/null github.com/easyp-tech/easyprotoc/cmd/easyprotoc` | 0 | 0 | 0 |
| 28 | Upstream cmd package | `go list -f '{{.Name}}' github.com/easyp-tech/easyprotoc/cmd/easyprotoc` | `main` | `main` | 0 |
| 29 | go.mod/go.sum freeze | `git diff --name-only df23d64 5342455 -- go.mod go.sum` | empty | empty | 0 |
| 30 | Tracked tree after | `git status --porcelain --untracked-files=no` | empty | empty | 0 |

### Battery shape (`/tmp/02-02-full-battery.log`, 13891 lines, rc 0)

| Metric | Pre-swap HEAD | Expected | Observed |
|---|---|---|---|
| `ok` package lines | 41 | 34 | **34** |
| `FAIL` lines | 0 | 0 | **0** |
| `--- FAIL` lines | 0 | 0 | **0** |
| `WARNING: DATA RACE` blocks | 0 | 0 | **0** |
| `--- SKIP` lines | — | only `TestGenerateGoHeaderUsesCompilerVersion` | **exactly that one** (protoc-gen-go absent from PATH) |
| `?` no-test packages | 9 | 3 | **3** |
| Total packages | 44 | 37 | **37** |

No-test packages: `cmd/easyp-mcp`, `internal/logger`, `wellknownimports`.

Deletion accounting matches the measurement: removing `internal/easypc` removes 7 test packages (`pkg/compiler`, `pkg/lexer`, `pkg/linker`, `pkg/options`, `pkg/parser`, `pkg/serialize`, `protoc`) and 6 no-test packages (`pkg/metrics`, `pkg/sourceinfo`, `protoc/syntax`, `protoc/wkt`, both cmd mains) — 41-7=34 and 9-6=3.

### Gate-count parity versus the Phase 1 oracle

| Gate | Phase 1 (01-02) | Phase 2 (this plan) |
|---|---|---|
| `TestAcceptanceGates` | PASS, baseline `1e9a596` | PASS 9.42 s, same baseline, raw diagnostic only-main 0 / only-head 0 |
| `TestLintGoldenFullRegistry` | 94, byte-exact | 94, byte-exact, 0.02 s |
| `TestPositionRuneColumns` | 19 PASS | 19 PASS |
| `TestProtoastPositionsMatchLegacyContract` | 11 PASS | 11 PASS |
| Core matrix | 43 PASS | 43 PASS |
| CLI matrix | 16 PASS | 16 PASS |

## Expectation freeze

The swap commit's modified set is exactly the seven D-02 consumer files:

- `internal/core/link.go`
- `internal/core/link_test.go`
- `internal/core/lint_linked_matrix_test.go`
- `internal/protoast/build.go`
- `internal/protoast/outline.go`
- `internal/protoast/protoast.go`
- `internal/protoast/protoast_test.go`

`internal/rules` and `internal/api` are byte-identical to `PRE_SWAP_BASE` (`git diff` empty), both pinned constants occur exactly once, and no golden, parity fixture, rule, engine, or API file was edited by this plan. No gate was relaxed, re-run, or re-captured.

## Phase predicates re-proven

- No tracked Go file references `internal/easypc` (`git grep` empty).
- `internal/easypc` is absent from the index and from the swap commit tree, including `VENDORED.md` and both cmd binaries.
- `go mod tidy -diff` is empty; the easyprotoc pin is exact and direct; protobuf is `v1.37.0-rc.1`; zero `replace`/`vendor` directives.
- `go.mod`/`go.sum` have zero diff versus `PRE_SWAP_BASE`.
- Upstream `github.com/easyp-tech/easyprotoc/cmd/easyprotoc` builds from the module and reports package `main`.

## Files Created/Modified

- `.planning/phases/02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii/02-02-SUMMARY.md` — this evidence record. No source, test, expectation, golden, config, or documentation file was modified.

## Decisions Made

- **No per-task commit.** Both tasks are evidence-only gate runs; `files_modified` names only this SUMMARY, matching the Phase 1 plan 01-02 precedent. The last production commit remains `5342455` and the metadata docs commit carries this file.
- **Battery executed once.** The single invocation used the ratified flags on a warm cache; a second run solely to obtain a greener result was explicitly not performed.
- **Exact-name gate claims.** Because `go test -run` exits 0 when the filter matches nothing, every gate is claimed by its top-level `=== RUN` and `--- PASS` lines, not by exit status alone.
- **Skip allow-list is one exact name.** `TestGenerateGoHeaderUsesCompilerVersion` is the only `--- SKIP` line; the cause (protoc-gen-go absent from PATH) was confirmed, and no other skip was tolerated.

## Deviations from Plan

None - plan executed exactly as written. No deviation rule (1-4) was triggered: attribution, build, vet, tidy, the one battery, all named gates, the freeze and all phase predicates behaved exactly as the plan predicted, including the measured 34-ok / 3-no-test shape.

## Issues Encountered

1. **Stale plan commit ledger from an earlier milestone.** `.git/gsd-plan-head-before-02-02` already contained `c896eee` from a previous milestone's phase 02 plan 02. Following the 02-01 precedent, it was overwritten with the true current base (`7e46374`) before the metadata commit. This plan produces no task commits, so the measured count is `0`.
2. **HEAD advanced to a docs-only descendant before the gates.** The plan loaded with HEAD at `40124d4`; a `docs(02-01): refresh STATE current-position prose` commit moved HEAD to `7e46374` before the first gate. Attribution was re-proven at that HEAD (zero non-.planning diff from `5342455`) and every gate ran on it.
3. **`task`, `docker`, `golangci-lint`, and `protoc-gen-go` are absent in this environment.** Every gate used `go` and `git` directly (Phase 1 precedent). `lint:go`/`schema:check` are not Phase 2 gates per D-09 (lint/seal ownership is Phase 3, full acceptance Phase 4), and the missing `protoc-gen-go` is the documented cause of the one allowed skip.

## Known Stubs

None. This plan introduced no placeholder values, TODOs, unwired components, or skipped tests. The one `--- SKIP` line is a pre-existing environment guard with an explicit skip reason, not a stub created here.

## Threat Flags

None. No new network endpoint, auth path, file-access pattern, or trust-boundary schema change was introduced. This plan is evidence-only and exercised the planned mitigations: T-02-19 (exact RUN-line claims), T-02-20 (freeze plus single-occurrence constants), T-02-21 (attribution and clean tree), T-02-22 (exact pins and tidy), T-02-23 (one fixed-flag battery), T-02-24 (SHAs, logs, timestamps).

## Deferred Items (honest record)

- **Phase 3 (seal/docs):** `.golangci.yml`'s `internal/easypc` exclusion, the compiler-lab/VENDORED.md prose in `AGENTS.md`, and `.spec`/README updates remain out of this phase by D-07.
- **Phase 3/4 quality gates:** `task lint:go`, `task schema:check`, and `task build` are not Phase 2 gates; full acceptance is IMPORT-PARITY-01 in Phase 4.
- **v1.0 corpora letters:** LINT-PARITY-01 and LINT-PERF-01 remain `pending_user_unblock` (private compiler-lab corpora inaccessible); the ready-to-fire scripts and `07-USER-SETUP.md` path stay alive in the milestone backlog.
- **No push (D-08):** the work stays on `feat/lint-easypc-infra`, extending PR #258.
- **Known non-blocking residue:** `internal/rules/package_no_import_cycle_test.go:28` still says "the easypc linker" in prose (no path). It is outside D-02's exact file set and outside the D-09 grep; left untouched by this evidence-only plan.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- **Phase 2 evidence is complete and attributable.** The committed swap tree is provably behavior-preserving against the v1.0 automated core, with zero expectation movement.
- **Phase 3 can start immediately:** remove the seal infrastructure and update active documentation, then run its own lint gates.
- **Phase 4 inherits a green baseline:** the full `-race` battery, Gates A-D x2, golden 94 and both D-07 position gates are already proven on the modular dependency.

---

*Phase: 02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii*
*Completed: 2026-11-21*

## Self-Check: PASSED

- SUMMARY.md exists at the plan path
- Commit 5342455 is an ancestor of HEAD, with zero non-.planning diff to HEAD
- All seven consumer files and every gate test file exist; `internal/easypc` is absent
