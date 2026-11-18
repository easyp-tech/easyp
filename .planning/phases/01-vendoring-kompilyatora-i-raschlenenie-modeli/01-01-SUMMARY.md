---
phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli
plan: 01
subsystem: infra
tags: [protobuf, compiler, vendoring, go, golangci-lint, supply-chain]

requires: []
provides:
  - internal/easypc — sealed vendored compiler tree (pkg/, protoc/ incl. syntax + embedded wkt, cmd/, testdata/), imports as github.com/easyp-tech/easyp/internal/easypc/...
  - internal/protoast — vendored lint view model (6 files), the maintained adaptation seam
  - internal/easypc/VENDORED.md — provenance record (repo/branch/dir/SHA/date) + re-sync procedure
  - .golangci.yml linters.exclusions.paths scoped to internal/easypc (protoast deliberately NOT excluded)
  - AGENTS.md D-04 seal — edit prohibition inside internal/easypc
  - tidy-clean go.mod baseline (go-billy correctly marked direct)
affects: [01-02, 01-03, 01-04, phase-2-linker, phase-3-lint-engine, phase-4-rules]

actuals:
  tokens: 223806   # chars/4 over the realized production diff (895,225 chars across 657 files) — dominated by the vendor copy itself; the plan's 25,000-token estimate measured authored work, not landed bytes
  tasks: 2
  commits: 2
  plan_head_before: 26924b14f432abdcee1c27a01ec9317d82f3ecad
  plan_head_after: 690a9cc1b07b3b137f6a10742d2692e20bf2e508

tech-stack:
  added: []   # zero new dependencies by design (LINT-VEND-02); golangci-lint v2.14.0 provisioned into repo-local bin/ per existing Taskfile pin
  patterns:
    - "Sealed vendor tree: copy-in from a pinned SHA, provenance in VENDORED.md, edits forbidden (AGENTS.md), lint excluded by path — fixes flow upstream and arrive by re-sync (D-03/D-04)"
    - "Standalone atomic vendor commit: the pure-addition property (build green, vendored tests green, tidy zero-diff) is provable only on a commit that mixes in nothing else (RESEARCH Pattern 1)"

key-files:
  created:
    - internal/easypc/** (656 files incl. VENDORED.md)
    - internal/protoast/{protoast,build,comments,walk,outline}.go + protoast_test.go
  modified:
    - .golangci.yml
    - AGENTS.md
    - go.mod

key-decisions:
  - "Vendored from compiler-lab main @ 08cef10909c0722f7042ba17b5c11951f9c6b6f6, dir easyp-cli/ — the easyp-cli branch does not exist (D-08 premise correction, user-confirmed); import paths already targeted github.com/easyp-tech/easyp so zero rewriting was performed"
  - "go-billy promoted indirect→direct in a dedicated pre-vendor commit (8f969d8) instead of folding it into the vendor commit — keeps 690a9cc a provable pure addition while making the tidy zero-diff gate pass"
  - "golangci exclusion scoped to internal/easypc only; internal/protoast stays lint-clean — its single known QF1002 (build.go:242) is intentionally left for plan 01-02"
  - "Pre-existing wazero data race in internal/adapters/plugin documented in deferred-items.md, not fixed (scope boundary)"

patterns-established:
  - "Vendor-drop seal: VENDORED.md + AGENTS.md bullet + .golangci.yml paths exclusion always land together with the copy"
  - "Re-sync procedure: clone, checkout new SHA, diff, copy, re-run the four gates (build / vendored tests / tidy zero-diff / lint), update VENDORED.md SHA+date"

requirements-completed: [LINT-VEND-01, LINT-VEND-02]

coverage:
  - id: D1
    description: "easypc compiler (whole tree incl. cmd/ and testdata/) and protoast view model vendored from the pinned SHA, compiling as part of the module with every vendored package's own tests green"
    requirement: LINT-VEND-01
    verification:
      - kind: unit
        ref: "go test -count=1 ./internal/easypc/... ./internal/protoast/... → 8 ok, 0 FAIL"
        status: pass
      - kind: other
        ref: "go build ./... → exit 0"
        status: pass
    human_judgment: false
  - id: D2
    description: "Zero new external dependencies: go mod tidy leaves go.mod/go.sum byte-identical, no replace directive, vendored trees' only external imports are already-direct requires (protobuf 5 subpackages + testify)"
    requirement: LINT-VEND-02
    verification:
      - kind: other
        ref: "go mod tidy && git diff --exit-code go.mod go.sum → empty (after 8f969d8 baseline fix)"
        status: pass
      - kind: other
        ref: "grep -c '^replace' go.mod → 0"
        status: pass
    human_judgment: false
  - id: D3
    description: "Vendor hygiene seal: VENDORED.md provenance with pinned SHA, easypc-only lint exclusion, AGENTS.md D-04 prohibition bullet, all in one atomic commit whose scope is exactly the vendor drop"
    requirement: LINT-VEND-01
    verification:
      - kind: other
        ref: "bin/golangci-lint run ./... → exactly one finding, internal/protoast/build.go:242 QF1002 (known, fixed in 01-02); zero findings elsewhere"
        status: pass
      - kind: other
        ref: "git show --name-only 690a9cc → only internal/easypc/**, internal/protoast/**, .golangci.yml, AGENTS.md"
        status: pass
    human_judgment: false

duration: 14h 15m
completed: 2026-11-18
status: complete
---

# Phase 1 Plan 01: Vendor easypc+protoast Summary

**easypc compiler (656 files: pkg/, protoc/ with embedded WKT, cmd/, testdata/) and the protoast view model vendored verbatim from compiler-lab main@08cef10 with zero new dependencies, sealed by VENDORED.md provenance, a path-scoped golangci-lint exclusion, and the AGENTS.md edit prohibition.**

## Performance

- **Duration:** 14h 15m wall-clock (2026-11-17T20:17Z → 2026-11-18T10:33Z). Includes two mid-execution environment interruptions (an API rate-limit pause and an overnight Orca environment restart) during which no work progressed; active execution time was roughly 75 minutes.
- **Started:** 2026-11-17T20:17:22Z
- **Completed:** 2026-11-18T10:33:18Z
- **Tasks:** 2 (both complete; Task 1 tracer + Task 2 auto landed as one atomic vendor commit per RESEARCH Pattern 1)
- **Files modified:** 657 (656 created in the vendor drop + go.mod marker fix)

## Accomplishments
- Vendored `internal/easypc` (whole compiler: `pkg/{compiler,lexer,linker,metrics,options,parser,serialize,sourceinfo}`, `protoc` + syntax + 20 embedded WKT protos, `cmd/{easyp-protoc,lexbench}`, `testdata/` 2.6 MB) and `internal/protoast` (6 files) from compiler-lab **main @ `08cef10909c0722f7042ba17b5c11951f9c6b6f6`**, dir `easyp-cli/`, verified by exact `git rev-parse` assertion before copying (supply-chain gate T-01-01).
- Import paths in the lab copy already targeted `github.com/easyp-tech/easyp/internal/...` — the copy needed **zero import rewriting**, and a repo-wide scan found zero wrong-prefix imports.
- All gates green: `go build ./...`; every vendored package's own tests `ok` (8 packages, D-05); `go mod tidy` zero-diff (LINT-VEND-02); golangci-lint v2.14.0 reports exactly the one expected finding (`internal/protoast/build.go:242` QF1002, deferred to 01-02 by design).
- Vendor hygiene sealed: `internal/easypc/VENDORED.md` (provenance + re-sync procedure, D-03/D-08), `.golangci.yml` `linters.exclusions.paths: [internal/easypc]` (D-04; protoast deliberately not excluded), AGENTS.md "Do not" bullet forbidding edits inside `internal/easypc` (D-04).
- Landed as ONE atomic vendor commit (690a9cc) whose file list is exactly the vendor scope — the pure-addition property is provable on it in isolation.

## Task Commits

Each task was committed atomically (per plan: Tasks 1+2 form a single vendor commit):

1. **Task 1 + Task 2: vendor drop + hygiene seal** - `690a9cc` (feat) — 656 files, 23,915 insertions
2. **Deviation (pre-vendor baseline fix): go-billy direct promotion** - `8f969d8` (chore)

**Plan metadata:** see below (docs commit).

## Files Created/Modified
- `internal/easypc/**` - vendored compiler tree, sealed (656 files incl. `VENDORED.md` provenance + re-sync procedure)
- `internal/protoast/{protoast,build,comments,walk,outline}.go`, `protoast_test.go` - vendored lint view model; the adaptation seam where 01-02 lands the position contract
- `.golangci.yml` - added `linters.exclusions.paths: [internal/easypc]` sibling to the existing `rules:` block
- `AGENTS.md` - appended D-04 prohibition bullet to "Do not"
- `go.mod` - go-billy promoted from `// indirect` to direct (marker-only change, commit 8f969d8)

## Decisions Made
- Pinned source honored exactly as user-confirmed in 01-CONTEXT.md D-08 refinement: main @ 08cef10, dir `easyp-cli/` (no `easyp-cli` branch exists; clone used the default branch + explicit SHA checkout).
- The one-time import-path rewrite exemption (D-04's sole exception) was **not exercised** — imports already matched; VENDORED.md records the exemption as exhausted.
- `internal/protoast` kept lint-visible on purpose: it is the maintained seam; its single QF1002 is 01-02's first fix target.
- golangci-lint provisioned via direct `go install ...@v2.14.0` into `bin/` (Taskfile pin) because the `task` binary is not installed on this machine (research-verified grounded path).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] go.mod go-billy promotion required before the tidy zero-diff gate could pass — FLAGGED for verifier attention re REQ-1.2**
- **Found during:** Task 1 (vendoring gate `go mod tidy && git diff --exit-code go.mod go.sum`)
- **Issue:** The plan (and research) expected `go mod tidy` to leave go.mod/go.sum byte-identical. Instead, tidy promoted `github.com/go-git/go-billy/v5 v5.6.2` from `// indirect` to direct: `internal/adapters/gitmodules` imports `go-billy/v5/osfs` directly, so main's committed go.mod was stale. Attribution proven empirically: with the vendored trees moved out, pristine HEAD produces the *identical* tidy diff — the condition is pre-existing and not caused by the vendor drop.
- **Fix:** Landed the tidy correction as its own dedicated commit (8f969d8) *before* the vendor commit, so 690a9cc remains a provable pure addition and its tidy gate passes byte-identically. Tidy verified idempotent afterwards.
- **Files modified:** go.mod
- **Verification:** pristine-HEAD tidy reproduction (identical diff without vendor trees); post-fix `go mod tidy && git diff --exit-code go.mod go.sum` → empty, repeated runs idempotent.
- **Committed in:** 8f969d8
- **REQ-1.2 assessment:** substance holds — no new module, no new require entry (same requirement, same version, marker-only move between blocks), no replace directive, go.sum untouched. The byte-identical expectation in the plan's flagged assumption was falsified by a pre-existing repo condition, not by the vendoring.

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Minimal and contained: one extra commit outside the (intact) atomic vendor scope. No scope creep; the vendor commit itself mixes in nothing but the drop.

## Issues Encountered
- **Pre-existing wazero data race (out of scope, deferred):** `go test -race -count=1 ./...` fails only in `internal/adapters/plugin` — a race inside `github.com/tetratelabs/wazero@v1.9.0/internal/version.GetWazeroVersion()` poisons the package's `TestBuiltin*` tests under the race detector. Reproduced on pristine HEAD with the vendor trees removed; unrelated to this plan. Recorded in [deferred-items.md](./deferred-items.md) and the windows ledger; not fixed per the executor scope boundary. All other 40 packages pass `-race`.
- Two execution interruptions (API rate-limit, environment restart) paused wall-clock time; on-disk state was verified intact after each and no work was lost.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Ready for 01-02 (position adaptation in `internal/protoast/build.go` `pos()` + D-07 two-level tests + the QF1002 tagged-switch fix) and, in parallel wave terms, 01-03/01-04 — the vendored trees, provenance seal, and lint scoping they build on are in place at 690a9cc.
- LINT-VEND-01 remains checked-open in REQUIREMENTS.md by the shared-ID gate (01-02/03/04 also declare it); it flips when the last sibling summary lands.
- No blockers for the remaining plans. The deferred wazero race is worth eyes before any later phase treats a full `-race` run as a hard gate on this machine.

## Self-Check: PASSED

All key artifacts verified present on disk (internal/easypc/{VENDORED.md, pkg/parser, protoc/wkt, testdata}, internal/protoast/protoast.go, .golangci.yml, AGENTS.md, this SUMMARY, deferred-items.md). Commits 8f969d8 and 690a9cc confirmed ancestors of HEAD; ledger-measured commit count 2 (nonzero, matching code changes). VENDORED.md is 36 lines (min 12).

---
*Phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli*
*Completed: 2026-11-18*
