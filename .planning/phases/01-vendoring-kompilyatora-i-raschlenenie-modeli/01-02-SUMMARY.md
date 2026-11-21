---
phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli
plan: 02
subsystem: testing
tags: [protobuf, protoast, positions, utf8, runes, tdd, golangci-lint, vendoring]

requires:
  - phase: 01-01
    provides: vendored internal/easypc (sealed) and internal/protoast (adaptation seam) from compiler-lab main@08cef10, with protoast left lint-visible carrying exactly one known QF1002
provides:
  - builder.pos() in internal/protoast/build.go computing 1-based rune columns (tab = +1) — the single adaptation locus for main's position contract
  - internal/protoast/position_test.go — D-07(a) unit proof (18 hand-computed anchors incl. name anchors)
  - internal/core/position_parity_test.go — D-07(b) golden proof (10 first-token anchors, legacy go-protoparser vs protoast, package core)
  - internal/core/testdata/position_contract_utf8.proto — shared UTF-8+tab fixture (first non-ASCII fixture in the repo)
  - protoast lint-clean with QF1002 fixed in code (no exclusion added)
  - internal/easypc/VENDORED.md protoast divergence section with re-sync re-apply instruction
affects: [01-03, 01-04, phase-4-rules, phase-3-lint-engine]

actuals:
  tokens: 3212   # chars/4 over the realized diff (12,850 diff chars across 7 files) — the 45,000 estimate budgeted exploration the executor could skip because 01-01's read_first trail was precise
  tasks: 3
  commits: 3
  plan_head_before: 96d1747faf563aec4411c41cfed3d1315f31f0d6
  plan_head_after: 31885b481d481de8c860cdd4fc96ef8730ee516b

tech-stack:
  added: []   # zero new dependencies (LINT-VEND-02 preserved; tidy zero-diff re-proven)
  patterns:
    - "Position-contract-in-one-place: the rune-column contract lives exclusively in protoast builder.pos(); namePos and attachComments inherit it by routing through pos()"
    - "Two-level position proof (D-07): hand-computed unit expectations (D-07a) + model-vs-model golden parity against the legacy parser (D-07b) — the legacy parser is the contract's definition, so the golden test needs no position literals"

key-files:
  created:
    - internal/core/testdata/position_contract_utf8.proto
    - internal/protoast/position_test.go
    - internal/core/position_parity_test.go
  modified:
    - internal/protoast/build.go
    - internal/protoast/protoast.go
    - internal/protoast/protoast_test.go
    - internal/easypc/VENDORED.md

key-decisions:
  - "Rune columns are computed in builder.pos() by scanning b.ast.Src back to the previous newline and counting runes with utf8.RuneCount; the line number still comes from the vendored b.syn.LineCol because line starts are byte offsets in both conventions"
  - "The two tab-8-dependent assertions in the vendored protoast_test.go are lines 93 and 95 (enum value Pos 9→2, NumberPos 20→13), not ~93/~96 as research pinned: line 96's outer.NamePos==9 is space-based (message Outer) and identical under both conventions — value unchanged"
  - "D-07(a) hand-computed expectations were cross-checked against the legacy go-protoparser parser (the contract's definition) before the RED commit; one off-by-one (package column 40→39, 'option' is 6 chars) was corrected pre-commit"
  - "Commits landed on main per the sequential-executor dispatch (branching_strategy 'none', 01-01 precedent 8f969d8/690a9cc/96d1747), although the default-branch guard query reports main as protected — documented override, config not modified"

patterns-established:
  - "Contract proof before porting: every cross-model convention gap gets a fixture that makes the two conventions provably diverge, then a model-level golden test with zero position literals"
  - "Vendored-test expectation updates ride the adaptation commit, never a separate silent edit; the divergence and its re-sync re-apply steps are recorded next to the vendor provenance"

requirements-completed: [LINT-VEND-01]

coverage:
  - id: D1
    description: "protoast positions follow main's rune-column contract: 1-based line, 1-based rune column, tab advances by exactly 1 — proven by 18 hand-computed anchors (decl + name) over a Cyrillic/emoji/tab fixture"
    requirement: LINT-VEND-01
    verification:
      - kind: unit
        ref: "go test -count=1 ./internal/protoast/ -run 'TestPosition' → ok (TestPositionRuneColumns, 18 subtests)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Golden anchor-by-anchor parity: for every first-token anchor both models expose (package, message, 3 fields, enum, 2 enum values, service, rpc), protoast Position.Line/.Column equal legacy go-protoparser Meta.Pos on the same fixture bytes"
    requirement: LINT-VEND-01
    verification:
      - kind: integration
        ref: "go test -count=1 ./internal/core/ -run 'TestProtoastPositions' → ok (TestProtoastPositionsMatchLegacyContract, 10 anchors)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Vendored protoast suite green with tab expectations updated to tab-as-1 semantics, and protoast lint-clean: QF1002 fixed in code (tagged switch), zero golangci findings, no exclusion added"
    requirement: LINT-VEND-01
    verification:
      - kind: unit
        ref: "go test -count=1 ./internal/protoast/... → ok (full package)"
        status: pass
      - kind: other
        ref: "bin/golangci-lint run ./... → 0 issues (grep -c 'multiple of 8' protoast_test.go = 0; grep -c 'internal/protoast' .golangci.yml = 0)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Upstream divergence recorded for future syncs: VENDORED.md protoast section states the rune-column divergence in build.go pos() and mandates re-applying it (plus tab expectations and package doc) on every compiler-lab re-sync"
    requirement: LINT-VEND-01
    verification:
      - kind: other
        ref: "grep -c 'protoast' internal/easypc/VENDORED.md → 10; 're-apply this divergence' present"
        status: pass
    human_judgment: false
  - id: D5
    description: "Phase-gate sweep: golangci fully clean, go mod tidy zero-diff, -race green on all in-scope packages; the full-module -race run fails only on the pre-existing documented wazero race in internal/adapters/plugin (deferred in 01-01, WINDOWS.md ledger id 1) — re-confirmed pre-existing on pristine HEAD there"
    verification:
      - kind: other
        ref: "bin/golangci-lint run ./... → 0 issues; go mod tidy && git diff --exit-code go.mod go.sum → empty"
        status: pass
      - kind: unit
        ref: "go test -race -count=1 ./internal/core/ ./internal/protoast/... → ok, ok"
        status: pass
    human_judgment: false

duration: 17 min
completed: 2026-11-18
status: complete
---

# Phase 1 Plan 02: Rune-column position contract Summary

**protoast positions now follow main's output contract (1-based rune columns, tab = +1) via a single builder.pos() adaptation, proven at two levels — 18 hand-computed anchors plus anchor-by-anchor golden parity against the legacy go-protoparser path — with QF1002 fixed and the divergence recorded for future compiler-lab syncs.**

## Performance

- **Duration:** 17 min (2026-11-18T13:33:24Z → 2026-11-18T13:50:53Z)
- **Started:** 2026-11-18T13:33:24Z
- **Completed:** 2026-11-18T13:50:53Z
- **Tasks:** 3 (RED → GREEN → divergence record + phase-gate sweep)
- **Files modified:** 7 (3 created, 4 modified; 211 insertions, 9 deletions)

## TDD Gate Compliance

**RED** — `test(01-02)` commit `ce1cc8d`: fixture + both D-07 tests, failing on column arithmetic against the vendored byte/tab-8 behavior. D-07(a) `TestPositionRuneColumns`: 13 of 18 subtests failed (e.g. package expected 39 / actual 51, enum value after tab+emoji expected 26 / actual 50, rpc after two tabs expected 3 / actual 17). D-07(b) `TestProtoastPositionsMatchLegacyContract`: 6 of 10 anchors failed (exactly the tab/multi-byte anchors; the ASCII/space-indented anchors passed, matching the behavior bullet that pure-ASCII columns are identical). Both runs captured as complete `go test` console output (package summary lines present, no build/setup failures) and classified by `gsd_run check tdd-red-evidence`: **RED_EVIDENCE_OK / target_test_failed** for both records. Semantic assessment: the targets executed and failed on the planned Column assertions for the intended reason — byte-vs-rune and tab-8-vs-tab-1 arithmetic — not on compilation, discovery, or unrelated assertions. Line assertions passed everywhere, confirming line numbering is convention-independent.

**GREEN** — `feat(01-02)` commit `7fd1af4`: `builder.pos()` keeps the line from the vendored `b.syn.LineCol(off)` (line starts are byte offsets in both conventions) and recomputes the column as `utf8.RuneCount(b.ast.Src[lineStart:off]) + 1`, with lineStart found by scanning back to the previous newline. `namePos` and `attachComments` route through `pos()`, so this is the only change needed. Same commit updates the package doc in `protoast.go`, the two tab-8 expectations in `protoast_test.go`, and fixes QF1002 (tagless switch on a `TypeKind` comparison → tagged switch on `n.TypeKind`). Both D-07 tests and the full vendored protoast suite green.

**REFACTOR** — none needed; the adaptation is minimal and documented (no refactor commit, per discipline).

**Gate sequence:** RED `ce1cc8d` is an ancestor of GREEN `7fd1af4` (verified via `git merge-base --is-ancestor`); no feat commit precedes the test commit for this plan.

## Accomplishments
- Fixture `internal/core/testdata/position_contract_utf8.proto` (19 lines): Cyrillic/emoji comments AND a multi-byte string literal placed before anchored tokens on the same line (package after `option ... = "привет, мир 😀";`, names after inline block comments), tab-indented field/enum-value/rpc declarations, anchors covering package, message, field, enum, enum value, service, rpc.
- D-07(a) `internal/protoast/position_test.go`: 18 named table cases asserting exact hand-computed 1-based rune Line/Column per anchor, including NamePos anchors and pure-ASCII regression anchors; field access only, zero `meta.Position` literals (same-wave invariant for 01-03 preserved).
- D-07(b) `internal/core/position_parity_test.go` (package `core`): reads the fixture once, parses with unexported `readProtoFileWithDirectives` (the source of current lint issue positions) and `protoast.Read`, asserts Line+Column equality per first-token anchor — no position literals.
- QF1002 at `internal/protoast/build.go:242` fixed in code; `bin/golangci-lint run ./...` reports **0 issues** with protoast NOT excluded in `.golangci.yml`.
- `internal/easypc/VENDORED.md` records the protoast divergence (rune columns in `pos()`, tab-as-1 test expectations) with a mandatory re-apply step in the re-sync procedure; easypc itself untouched.

## Task Commits

Each task was committed atomically:

1. **Task 1: RED — fixture + D-07(a) + D-07(b), both failing on column arithmetic** - `ce1cc8d` (test) — 3 files, 165 insertions
2. **Task 2: GREEN — rune columns in builder.pos(), doc + tab expectations, QF1002** - `7fd1af4` (feat) — 3 files, 24 insertions, 9 deletions
3. **Task 3: divergence record + phase-gate sweep** - `31885b4` (docs) — VENDORED.md +22 lines; gates run before commit

**Plan metadata:** see docs commit below.

## Files Created/Modified
- `internal/core/testdata/position_contract_utf8.proto` - shared UTF-8+tab fixture; first non-ASCII fixture in the repo (4 tab-indented lines, 9 multi-byte lines)
- `internal/protoast/position_test.go` - D-07(a) unit proof, hand-computed rune columns
- `internal/core/position_parity_test.go` - D-07(b) golden parity, package core
- `internal/protoast/build.go` - `pos()` rune-column adaptation (single locus) + QF1002 tagged-switch fix
- `internal/protoast/protoast.go` - package doc now states the easyp output contract (rune columns, tab = +1) and points at VENDORED.md for the sync note
- `internal/protoast/protoast_test.go` - tab expectations 9→2 and 20→13 with rune-convention wording
- `internal/easypc/VENDORED.md` - protoast divergence section + re-sync re-apply instruction

## Decisions Made
- Line number still comes from the vendored `LineCol` (byte-offset line starts are convention-independent); only the column is recomputed — keeps the divergence surface minimal and the O(line-length) backward scan local to `pos()`.
- Research pin corrected: the tab-8-dependent assertions are protoast_test.go lines 93 and 95; line 96 (`outer.NamePos.Column == 9`) is space-based and identical under both conventions, so its value stays 9 (no rewording needed — it carries no convention-specific message).
- D-07(a) expectations cross-checked against the legacy parser before committing RED (the legacy parser defines the contract); this caught one hand-arithmetic slip (`option` is 6 chars → package rune column 39, not 40) before it could masquerade as a contract value.
- RED evidence JSON records were classified (both RED_EVIDENCE_OK) and then removed as transient gate artifacts; the verdicts, commands, and representative diffs are recorded in this SUMMARY and reproducible from the committed tests.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Research-pinned tab-8 assertion sites were off by one line**
- **Found during:** Task 2 (updating vendored tab expectations)
- **Issue:** Plan/research said the two tab-8 expectations sit at protoast_test.go ~93 and ~96 (`outer.NamePos.Column == 9`). Line 96 is space-based (`message Outer {` puts `Outer` at column 9 under both conventions); the actual second tab-dependent assertion is line 95 (`values[1].NumberPos.Column == 20`, which depends on the tab stop).
- **Fix:** Updated exactly the two tab-8-dependent assertions: line 93 (enum value Pos 9→2, message reworded to "a tab advances the column by exactly one") and line 95 (NumberPos 20→13). Line 96 left unchanged — its value is convention-independent.
- **Files modified:** internal/protoast/protoast_test.go
- **Verification:** `grep -c 'multiple of 8' internal/protoast/protoast_test.go` → 0; full protoast suite green
- **Committed in:** 7fd1af4 (part of Task 2 GREEN commit)

---

**Total deviations:** 1 auto-fixed (1 bug — research citation correction; the plan's "two tab-8 expectations" count was correct)
**Impact on plan:** None on scope or contract; the plan's intent (update every tab-8-dependent expectation to tab-as-1) was satisfied exactly.

## Issues Encountered
- **Full-module `-race` sweep fails only in `internal/adapters/plugin`** (wazero `GetWazeroVersion` data race, 1 occurrence) — the pre-existing defect documented in 01-01's deferred-items.md and the WINDOWS.md ledger (id 1); the environment brief pre-authorized treating exactly this as the known deferred item. All other packages pass `-race`, including `internal/core` and `internal/protoast` (the plan's acceptance criterion). Not fixed here (executor scope boundary); no duplicate ledger entry appended since id 1 already tracks it.
- One hand-computation off-by-one in D-07(a) (package column) was caught by the legacy-parser cross-check before the RED commit — the RED run still failed for the right reason on all planned divergent anchors.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- The position contract is proven at both D-07 levels before any rule porting (Phase 4 precondition met); Phase 4's expected columns can rely on `protoast.Position` matching what users see today.
- For 01-03 (same wave): both new test files assert positions via `.Line`/`.Column` field access and introduce zero `meta.Position` value literals, so the HEAD=56-literal / 42-file sweep baselines are unaffected by this plan.
- LINT-VEND-01 remains governed by the shared-ID gate (01-01/02/03/04 all declare it); it flips when the last sibling summary lands.
- The deferred wazero race remains the only red spot in a full `-race` run on this machine.

## Self-Check: PASSED

All key files verified present on disk (fixture, both test files, build.go, protoast.go, protoast_test.go, VENDORED.md). Commits ce1cc8d, 7fd1af4, 31885b4 confirmed ancestors of HEAD; ledger-measured commit count 3 (matches per-task commits); no file deletions in any commit; no stray untracked files from this plan; RED precedes GREEN in ancestry.

---
*Phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli*
*Completed: 2026-11-18*
