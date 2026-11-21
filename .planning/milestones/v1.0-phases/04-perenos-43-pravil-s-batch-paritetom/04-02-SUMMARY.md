---
phase: 04-perenos-43-pravil-s-batch-paritetom
plan: 02
subsystem: rules
tags: [lint-rules, protoast, parity-harness, linked-engine, porting, dormancy]

# Dependency graph
requires:
  - phase: 04-perenos-43-pravil-s-batch-paritetom (plan 01, Wave 0)
    provides: requirePorted/assertBatchParity/fingerprint harness, parseLinkedFile+validate helpers, NewLinked registry with placeholders, AppendLinkedIssue + LintRuleName seam, locked Decl.Pos position table
provides:
  - DirectorySamePackageLinked + PackageSameDirectoryLinked ModuleRules with main's first-exempt Finish semantics (T-04-03 mitigation)
  - PackageDefinedLinked emitting main's zero Position literal 0:0 (T-04-04 mitigation)
  - groupedValues.firstExemptMismatches — the shared main-semantics Finish helper for all 9 main-semantics group rules; lab report-everyone mismatches untouched and unreachable from ports
  - Batch-1 parity suite TestLinkedParityMinimal (five cells) with the pilot re-verified under the cumulative four-rule config
  - NewLinked registry advanced to 4/41 rules (pilot + MINIMAL batch)
affects: [04-03..04-06 batches (harness + firstExemptMismatches reuse), 04-07 (exact-set guard, deferred-decisions consolidation), Phase 5 flip (rename-map consolidation)]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 5078    # 20314 diff chars over internal/rules/ at chars/4 (estimate was 65000; the batch was uniform ports on a proven harness)
  tasks: 3
  commits: 3      # MEASURED: git rev-list --count c7f3654..HEAD

# Tech tracking
tech-stack:
  added: []
  patterns: [first-exempt Finish via groupedValues.firstExemptMismatches (main cache semantics on the Collect/Finish contract), zero Position literal for file-level rules, full-list requirePorted RED guard, packageless files invisible to group Collect]

key-files:
  created:
    - internal/rules/directory_same_package_linked.go
    - internal/rules/package_defined_linked.go
    - internal/rules/package_same_directory_linked.go
    - internal/rules/parity_minimal_test.go
  modified:
    - internal/rules/lint_helpers.go
    - internal/rules/builder_linked.go
    - internal/rules/parity_harness_test.go
    - internal/rules/init_test.go
    - internal/rules/directory_same_package_test.go
    - internal/rules/package_defined_test.go
    - internal/rules/package_same_directory_test.go

key-decisions:
  - "requirePorted probes every given name and reports the FULL unported list in one RED assertion — the plan's acceptance criterion demands the failure name all three unported rules, which the Wave-0 fail-fast guard could not do"
  - "MINIMAL group rules reach only firstExemptMismatches; the lab's report-everyone groupedValues.mismatches stays byte-untouched and unreachable from ports (T-04-03 mitigation, grep-verified)"
  - "PackageDefinedLinked emits main's zero Position literal (Filename only, Line 0 Column 0) pinned in both the parity fingerprint and the carried test; protoast.FilePosition 1:1 is excluded (T-04-04 mitigation)"
  - "Unused parseProtoLinked removed instead of a new lint exclusion — no later plan names it and no natural batch-1 caller exists; any batch porting the etalon module_rules_test shape re-adds it with callers in its own commit"

patterns-established:
  - "Group-rule port shape: Collect adds one groupedValue per package declaration (packageless files contribute nothing), Finish delegates to firstExemptMismatches — every later main-semantics group rule (04-03..04-05) follows it"
  - "Zero-position rule shape: core.Position{Filename: info.Path} through AppendLinkedIssue — never protoast.FilePosition"

requirements-completed: [LINT-PARITY-02]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Batch-1 parity suite: five cells proving old-vs-new sorted (rule, path, line, col) multiset parity for DIRECTORY_SAME_PACKAGE / PACKAGE_DEFINED / PACKAGE_SAME_DIRECTORY with expected old-side counts asserted before each compare, plus the Wave-0 pilot re-verified under the cumulative four-rule config"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_minimal_test.go#TestLinkedParityMinimal"
        status: pass
      - kind: other
        ref: "command: gsd_run check tdd-red-evidence .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-02-task1.json -> RED_EVIDENCE_OK (target test failed on the requirePorted guard, exit 1)"
        status: pass
    human_judgment: false
  - id: D2
    description: "DirectorySamePackageLinked port: first-exempt directory semantics (main_exempt first file, later differing files issue at their own package positions), packageless files contribute nothing"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/directory_same_package_test.go#TestDirectorySamePackage_Validate"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_minimal_test.go#TestLinkedParityMinimal/first_exempt_directory"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_minimal_test.go#TestLinkedParityMinimal/packageless_invisible"
        status: pass
    human_judgment: false
  - id: D3
    description: "PackageDefinedLinked port: packageless file reported at the zero Position literal (Filename, 0:0) with main's message; packageful files report nothing"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/package_defined_test.go#TestPackageDefined_Validate"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_minimal_test.go#TestLinkedParityMinimal/package_defined_zero_pos"
        status: pass
    human_judgment: false
  - id: D4
    description: "PackageSameDirectoryLinked port: package-keyed first-exempt (first directory wins, later differing directories issue at their own package positions)"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/package_same_directory_test.go#TestPackageSameDirectory_Validate"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_minimal_test.go#TestLinkedParityMinimal/first_exempt_package_dir"
        status: pass
    human_judgment: false
  - id: D5
    description: "Shared firstExemptMismatches Finish helper exists once in lint_helpers.go; the lab's report-everyone mismatches stays byte-untouched and unreachable from any ported rule file"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "command: grep -n '\\.mismatches(' internal/rules/*_linked.go -> no hits; git diff lint_helpers.go shows mismatches body unchanged"
        status: pass
    human_judgment: false
  - id: D6
    description: "Batch gate: cumulative TestLinkedParity* (8 green sub-cells), quick suite -race green, suppression centrality + readSource-once green, build/tidy/schema diff-empty, dormancy audit clean (old rule files + builder.go + engine files byte-frozen), golangci-lint clean"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "command: go test -race -count=1 ./internal/core/... ./internal/rules/... && go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity' && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && bin/golangci-lint run ./internal/rules/... ./internal/core/..."
        status: pass
    human_judgment: false

# Metrics
duration: 10 min
started: 2026-11-19T19:05:39Z
completed: 2026-11-19T19:15:31Z
status: complete
commits: 3
plan_head_before: c7f36545e6e50786428a9300d6b09c18ad31db29
plan_head_after: 9b1c123f9833678ae1f90ea063781c46a56dbed9
---

# Phase 04 Plan 02: Batch 1 — MINIMAL Group Summary

**DirectorySamePackage, PackageDefined and PackageSameDirectory ported to the linked contract with main's first-exempt and zero-position semantics, parity-proven old-vs-new through five batch cells with the pilot re-verified cumulatively (4/41 rules ported)**

## Performance

- **Duration:** 10 min
- **Started:** 2026-11-19T19:05:39Z
- **Completed:** 2026-11-19T19:15:31Z
- **Tasks:** 3
- **Files modified:** 11 source/test files (+3 planning artifacts)

## Accomplishments

- The two MINIMAL group rules are ported as ModuleRules with main's first-exempt Finish semantics: `Collect` adds one `groupedValue` per package declaration (packageless files contribute nothing — main iterates Packages only), `Finish` delegates to the new shared `groupedValues.firstExemptMismatches`; the first file/value of every disagreeing group never issues
- `PackageDefinedLinked` emits main's zero Position literal (`core.Position{Filename: info.Path}`, Line 0 Column 0) — the lab's 1:1 `protoast.FilePosition` is excluded by prohibition, pinned in the parity fingerprint and the carried test
- `firstExemptMismatches` is the single main-semantics Finish mechanism in `lint_helpers.go`; the lab-faithful report-everyone `mismatches` stays byte-untouched and grep-verified unreachable from all `*_linked.go` files
- `builder_linked.go` replaces the three MINIMAL placeholders in builder.go's group order (registry now 4/41: pilot + this batch); the rename map records the three old-to-new type names
- TDD cycle machine-validated: RED evidence `tdd-red-evidence-04-02-task1.json` returned `RED_EVIDENCE_OK` (TestLinkedParityMinimal failed on the requirePorted guard naming all three unported rules, exit 1, package compiled) before GREEN
- Batch gate green on all eight gates: cumulative `TestLinkedParity*` (8 green sub-cells), quick suite `-race`, suppression centrality + readSource-once, build, tidy diff-empty, schema diff-empty, dormancy audit, golangci-lint clean
- Standing constraints held: (a) `TestReadSourceOnce` green after every task; (b) the three ported rules consult no comment directives or `ignoreOnly` — central-filter cells green throughout

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): MINIMAL parity suite fails on the unported-rule guard** - `674aa18` (test)
2. **Task 2 (GREEN): port #1/#2/#4 with main semantics + firstExemptMismatches helper + carried tests** - `910a5f8` (feat)
3. **Task 3: batch gate — lint-gate fix (drop unused helper)** - `9b1c123` (fix)

**Plan metadata:** docs commit (this SUMMARY + state/roadmap/validation updates).

## TDD Gate Compliance

- Plan `type: tdd`; Task 1 is the plan's RED task and its commit `674aa18` (`test(04-02)`) precedes GREEN `910a5f8` (`feat(04-02)`) in ancestry — gate sequence RED -> GREEN satisfied, no REFACTOR needed (no cleanup beyond the plan shape was warranted; the GREEN commit is the minimal implementation).
- RED evidence is machine-validated (`gsd_run check tdd-red-evidence` -> `RED_EVIDENCE_OK`, reason `target_test_failed`, format `go-test`) with the semantic assessment recorded in the evidence file: the target test executed and failed on the planned requirePorted assertion, not a build/setup/fixture fault.
- Tasks 1's RED commit touches only `*_test.go` pathspecs plus the evidence JSON planning artifact.

## Files Created/Modified

- `internal/rules/directory_same_package_linked.go` - NEW: ModuleRule port (dir-keyed groups, first-exempt Finish, package pos = Decl.Pos)
- `internal/rules/package_defined_linked.go` - NEW: LintRule port with the zero Position literal (0:0)
- `internal/rules/package_same_directory_linked.go` - NEW: ModuleRule port (package-keyed groups, first-exempt Finish)
- `internal/rules/lint_helpers.go` - added `firstExemptMismatches` (main first-exempt Finish); `mismatches` untouched; header delta note extended
- `internal/rules/builder_linked.go` - three MINIMAL placeholders replaced in group order; rename map updated
- `internal/rules/parity_minimal_test.go` - NEW: five parity cells with expected old-side counts (2/0/1/1/2)
- `internal/rules/parity_harness_test.go` - requirePorted now probes every name and reports the full unported list in one assertion
- `internal/rules/init_test.go` - removed the unused `parseProtoLinked` helper (lint-gate fix)
- `internal/rules/directory_same_package_test.go`, `package_defined_test.go`, `package_same_directory_test.go` - rewritten onto the linked contract (parseLinkedFile + validate Collect/Finish driver), main's case names and literals carried

## Decisions Made

- `requirePorted` reports all unported names in one failure (Wave-0 guard adjusted): the RED evidence must name the batch's full unported list per the plan's acceptance criterion; single-name fail-fast could not
- Group ports never reach `groupedValues.mismatches` — `firstExemptMismatches` is added beside it, keeping the vendored file lab-faithful for re-sync while the ports get main semantics
- `parseProtoLinked` removed rather than excluded from lint: no caller in this batch, no later plan references it; re-add with callers when a batch ports the etalon's `module_rules_test` shape
- PackageSameDirectory carried test asserts the real engine literal at the package-keyword position (see Deviations 2)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] requirePorted fail-fast could not satisfy the RED acceptance criterion**
- **Found during:** Task 1 (design-time check of the RED evidence)
- **Issue:** the plan requires the RED failure output to name all three unported rules, but the Wave-0 guard Fatals on the first unported name
- **Fix:** the guard probes every given name and reports the full unported list in one `t.Fatalf`; single-name Wave-0 usage is unchanged
- **Files modified:** internal/rules/parity_harness_test.go
- **Verification:** RED run output lists DIRECTORY_SAME_PACKAGE, PACKAGE_DEFINED and PACKAGE_SAME_DIRECTORY; checker verdict RED_EVIDENCE_OK
- **Committed in:** 674aa18 (Task 1 commit)

**2. [Rule 1 - Bug] PackageSameDirectory carried literal was dead text**
- **Found during:** Task 2 (carried-test rewrite)
- **Issue:** main's carried test never asserted `wantIssues` (errors-only assertions), so its literal (0:0, empty source/message) contradicts the engine's real finding; asserting it verbatim would fail against main's own behavior, while keeping the errors-only shape would leave a vacuous test and violate the acceptance criterion "pkg keyword positions for the two group rules"
- **Fix:** the rewritten test drives the ModuleRule through `validate()` and asserts the real engine literal (offset 20 / line 3 / col 1, SourceName "Queue", full message) — the same shape main's DirectorySamePackage test already pins
- **Files modified:** internal/rules/package_same_directory_test.go
- **Verification:** TestPackageSameDirectory green under -race; parity cell first_exempt_package_dir independently pins the same position
- **Committed in:** 910a5f8 (Task 2 commit)

**3. [Rule 3 - Blocking] Unused parseProtoLinked tripped the batch lint gate**
- **Found during:** Task 3 (gate 8, pre-existing from 04-01: the focused-parse helper landed pre-wired and batch 1 has no natural caller)
- **Issue:** `golangci-lint run ./internal/rules/...` failed on `unused` — Task 3's gate requires lint clean before 04-03 starts
- **Fix:** removed the helper (smallest blast radius — no new lint exclusion; the etalon stays the re-add reference with callers landing in the same commit as the need)
- **Files modified:** internal/rules/init_test.go
- **Verification:** gate 8 re-run clean (`0 issues`); all rule tests still green
- **Committed in:** 9b1c123 (Task 3 commit)

---

**Total deviations:** 3 auto-fixed (1 blocking harness criterion, 1 dead-literal bug, 1 blocking lint gate)
**Impact on plan:** All three mechanical corrections to plan/tooling details; no scope change, no behavior drift, old path byte-frozen throughout.

## Issues Encountered

None beyond the deviations above; every gate passed on its first or second run (the lint gate was the only failure and is Deviation 3).

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components were introduced.

## Deferred Decisions (for 04-07 consolidation ledger)

- Lab behavior deliberately excluded: `groupedValues.mismatches` report-everyone Finish semantics — not reached by any ported group rule; `firstExemptMismatches` (main first-exempt) is the ports' mechanism. The lab shape stays in the vendored file for re-sync fidelity.
- Lab behavior deliberately excluded: `packagePosition`'s packageless-file fallback (empty value at `protoast.FilePosition` 1:1) — main iterates Packages only; the pinned `packageless_invisible` cell makes the divergence fail parity if ever ported.
- Lab behavior deliberately excluded: `protoast.FilePosition` (1:1) for PACKAGE_DEFINED — main's zero Position literal wins (T-04-04).
- No lab-improvement adoption occurred in this batch; the catalogued main-vs-lab divergences for later batches (casing helpers, NamePos positions, ImportUsed resolution, swift_prefix bug, cross-file uniqueness) remain owned by plans 04-03..04-06.

## User Setup Required

None - no external service configuration required. The etalon checkout precondition held: `/tmp/gsd-research/compiler-lab` at SHA `08cef10909c0722f7042ba17b5c11951f9c6b6f6` (verified at execution start).

## Verification Log

- `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityMinimal'` → RED: exit 1, `--- FAIL: TestLinkedParityMinimal` naming all three unported rules — Task 1 verify; evidence record validated `RED_EVIDENCE_OK`
- `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityMinimal|TestDirectorySamePackage|TestPackageDefined|TestPackageSameDirectory' -v` → ok (all 5 parity cells + 3 carried tests green) — Task 2 verify
- `go test -race -count=1 ./internal/core/ -run 'Suppression|ReadSourceOnce'` → ok (standing constraints a+b)
- `grep -n '\.mismatches(' internal/rules/*_linked.go` → no hits (acceptance: mismatches unreachable from ports)
- `git diff --exit-code -- internal/rules/builder.go internal/rules/directory_same_package.go internal/rules/package_defined.go internal/rules/package_same_directory.go internal/rules/package_directory_match.go` → empty (dormancy)
- Batch gate (Task 3): quick suite `-race` ok (core, core/path_helpers, rules); `TestLinkedParity -v` → 8 green sub-cells; `go build ./...` ok; `go mod tidy && git diff --exit-code go.mod go.sum` → TIDY_CLEAN; `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` → SCHEMA_CLEAN; dormancy audit (lint.go, lint_linked.go, lint_model.go, builder.go + three old rule files) → empty; `bin/golangci-lint run ./internal/rules/... ./internal/core/...` → 0 issues (after Deviation 3 fix)
- Commits measured: `git rev-list --count c7f3654..HEAD` = 3

## Next Phase Readiness

- Ready for 04-03 (batch 2 — BASIC casing): the MINIMAL batch is parity-verified before any later batch starts (ROADMAP Phase-4 criterion 1); `firstExemptMismatches` is in place for the option-comparison ModuleRules of batch 3; the harness consumed unchanged apart from the full-list RED guard
- Every later batch consumes the same harness surfaces as before (`requirePorted` now reports full unported lists — better RED evidence); `parseProtoLinked` is no longer pre-wired — re-add it with callers if a batch ports the etalon's `module_rules_test` shape
- No blockers; standing constraints (a) and (b) remain mechanized green cells

## Self-Check: PASSED

- All 9 task files + the RED evidence record exist on disk (11/11 FOUND)
- All 4 commits are ancestors of HEAD: 674aa18 (RED test), 910a5f8 (GREEN feat), 9b1c123 (gate fix), 90ee86d (docs metadata)
- Commits measured from the persisted ledger: `git rev-list --count c7f3654..HEAD` = 3
- STATE.md at Plan 3 of 7, stopped_at "Completed 04-02-PLAN.md"; ROADMAP.md phase row updated (2/7 summaries, In Progress)
- LINT-PARITY-02 correctly NOT marked complete (shared-ID gate: sibling plans 04-03..04-07 declare it without SUMMARYs)
- No WINDOWS.md entries: no stubs, no skipped tests, no unrun verifies

---
*Phase: 04-perenos-43-pravil-s-batch-paritetom*
*Completed: 2026-11-19*
