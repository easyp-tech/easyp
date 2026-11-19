---
phase: 04-perenos-43-pravil-s-batch-paritetom
plan: 01
subsystem: rules
tags: [lint-rules, protoast, parity-harness, linked-engine, porting, dormancy]

# Dependency graph
requires:
  - phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya
    provides: LintLinked two-pass engine, LintRule/ModuleRule contracts, LinkedFileInfo, Options.LinkedRules, central comment suppressions, D-01 total sort
provides:
  - core.AppendLinkedIssue — the single linked issue-construction point for all 41 ports (rename-map tracked)
  - LinkedRuleName with the additive LintRuleName() method preference — Option A coexistence mechanism (suffixed port types report canonical registry names)
  - rules.NewLinked — dormant linked registry mirroring New's validation order (pilot entry + 40 commented batch placeholders)
  - PackageDirectoryMatchLinked pilot rule passing main's carried test literals byte-verbatim
  - Batch parity harness: requirePorted RED guard, assertBatchParity (old-side count asserted before the multiset compare), fingerprint (D-01 key sort)
  - TestDeclPositionProbe — Decl.Pos == Meta.Pos pinned on Offset/Line/Column for 10 declaration kinds (11 anchors) with the mapping table in a comment
  - TestReadSourceOnce — disk-instrumentation cell proving readSource runs exactly once per corpus file (standing constraint 999.1)
  - TestModuleRuleSet — D-03 ten-name ModuleRule membership guard
affects: [04-02 batch 1 (re-verifies pilot in MINIMAL), 04-03..04-06 batches (harness + helpers reuse), 04-07 (exact-set guard strengthening, deferred-decisions consolidation), Phase 5 flip (rename-map consolidation)]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 8513
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns: [two-contract coexistence via suffixed <Name>Linked types + LintRuleName() override, per-batch old-vs-new parity harness with pre-compare old-side count assertion, RED guard by runtime registry probe (requirePorted) instead of a build failure, counting-walker disk instrumentation, position-mapping table locked in a probe comment]

key-files:
  created:
    - internal/rules/builder_linked.go
    - internal/rules/package_directory_match_linked.go
    - internal/rules/parity_harness_test.go
    - internal/rules/parity_wave0_test.go
    - internal/rules/module_rules_guard_test.go
    - internal/core/position_probe_test.go
    - internal/core/lint_linked_readsource_test.go
  modified:
    - internal/core/lint_model.go
    - internal/rules/init_test.go
    - internal/rules/package_directory_match_test.go

key-decisions:
  - "assertBatchParity takes wantOldCount and asserts the expected old-side finding count BEFORE the multiset compare — implements the T-04-02 anti-vacuity requirement in one engine run rather than a duplicate run per cell"
  - "TestModuleRuleSet probes each AllRuleNames entry through NewLinked individually and skips not-yet-ported names — the plan's literal Use: rules.AllRuleNames() cannot be green while the registry holds only the pilot; the guard's stated purpose (catch accidental ModuleRule implementations on landing) is preserved, and 04-07 strengthens it to exact-set equality"
  - "core-package test uses a literal KnownLintRules list instead of the plan's rules.AllRuleNames() — package core cannot import internal/rules (import cycle; 03-01 recorded precedent)"
  - "buildCore/Options.LinkedRules production wiring stays deferred to Phase 5 (planner decision restated): LinkedRules is populated only from tests in this phase, keeping easyp lint byte-identical and dormancy trivially provable"
  - "parseLinkedFile/parseProtoLinked/validate helpers added NEXT TO the kept old start/parseFile helpers (plan explicit): unported rules' tests keep compiling during the transition"

patterns-established:
  - "Port skeleton: <Name>Linked type + LintRuleName() canonical override + Validate(LinkedFileInfo) over protoast walk + core.AppendLinkedIssue — every batch plan follows it"
  - "Parity cell shape: requirePorted(t, names...) RED guard, then assertBatchParity(t, cfg, files, wantOldCount) — later batches add fixture cells only, no bespoke comparisons"

requirements-completed: [LINT-PARITY-02]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Linked issue-construction seam in core: AppendLinkedIssue assembles Position/SourceName/Message/RuleName in one point, and LinkedRuleName prefers an implemented LintRuleName() so the suffixed pilot type reports the canonical registry name (no _LINKED leak)"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/package_directory_match_test.go#TestPackageDirectoryMatch_Validate"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/package_directory_match_test.go#TestPackageDirectoryMatch_Message"
        status: pass
    human_judgment: false
  - id: D2
    description: "Dormant linked registry NewLinked mirrors New's single validateConfigRuleNames call (D-06) and registers exactly the pilot entry with 40 batch placeholders; unported names fail construction with core.ErrInvalidRule-wrapped errors (the per-batch RED guard mechanism)"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/module_rules_guard_test.go#TestModuleRuleSet"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/validation_parity_test.go#TestLintUseValidationParity"
        status: pass
    human_judgment: false
  - id: D3
    description: "Pilot rule PackageDirectoryMatchLinked passes main's carried expectations byte-verbatim (offset 20/line 3/col 1, SourceName, message, module-qualified prefix cells) and old-vs-new parity on the harness: mismatch cell reports exactly one finding, match cell an explicitly asserted zero, nested-decls corpus shape agreed zero"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/package_directory_match_test.go#TestPackageDirectoryMatch_Validate"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/package_directory_match_test.go#TestPackageDirectoryMatch_ModuleQualifiedPackage"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_wave0_test.go#TestLinkedParityWave0"
        status: pass
    human_judgment: false
  - id: D4
    description: "Position mapping locked by probe: protoast Decl.Pos equals fromMetaPos(Meta.Pos) on Offset, Line and Column for all ten declaration kinds (11 anchors incl. plain vs oneof-member fields), mapping table recorded in the probe file comment; NamePos/NumberPos/RPCType.Pos/FilePosition documented as forbidden for parity"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/core/position_probe_test.go#TestDeclPositionProbe"
        status: pass
    human_judgment: false
  - id: D5
    description: "Standing constraint (a) behavioral proof: on a compiling two-file corpus with an intra-corpus import, readSource opens every corpus file exactly once from disk (total opens == 2) through the counting-walker wrapper"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/core/lint_linked_readsource_test.go#TestReadSourceOnce"
        status: pass
    human_judgment: false
  - id: D6
    description: "D-03 membership guard green: every ported rule implementing core.ModuleRule reports a registry name inside the ten-name set; accidental ModuleRule implementations fail the moment they land (04-07 strengthens to exact-set)"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/module_rules_guard_test.go#TestModuleRuleSet"
        status: pass
    human_judgment: false
  - id: D7
    description: "Dormancy provable: old engine and old rule files byte-frozen (git diff empty on internal/core/lint.go, internal/api/, builder.go, all pre-existing rule files); Options.LinkedRules populated only from tests; go.mod/go.sum and schemas/ diff-empty; suppression centrality cells green"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "command: git diff --exit-code internal/core/lint.go internal/api/ internal/rules/builder.go; go mod tidy && git diff --exit-code go.mod go.sum; go run ./cmd/easyp schema-gen && git diff --exit-code schemas/; go test -race ./internal/core/ -run 'Suppression'"
        status: pass
    human_judgment: false

# Metrics
duration: 13 min
started: 2026-11-19T18:47:35Z
completed: 2026-11-19T19:00:18Z
status: complete
commits: 3
plan_head_before: 21f6b9609b3457d91b2e8989ed34bcbfced246a2
plan_head_after: f5ee4e6bd8516a4a2ace39f8699c62ca5dbe6bdf
---

# Phase 04 Plan 01: Wave-0 Infra + Pilot Tracer Summary

**Linked-issue seam + dormant NewLinked registry + pilot PackageDirectoryMatchLinked proven end-to-end on a two-engine parity harness, with the position map locked for 10 declaration kinds and the readSource-once disk cell green**

## Performance

- **Duration:** 13 min
- **Started:** 2026-11-19T18:47:35Z
- **Completed:** 2026-11-19T19:00:18Z
- **Tasks:** 3
- **Files modified:** 10 source/test files (+3 planning artifacts)

## Accomplishments

- The linked issue-construction seam exists in core: `AppendLinkedIssue` (the rename-mapped etalon 4-arg shape) is the single construction point for all 41 ports, and `LinkedRuleName` prefers an implemented `LintRuleName() string` — the Option A mechanism that lets suffixed port types report canonical registry names
- `rules.NewLinked` mirrors `New` structurally (one `validateConfigRuleNames` call per D-06, group order, use/except/ignore_only plumbing) and registers exactly the pilot entry; the other 40 sit as commented placeholders each batch replaces; the constructor is test-reachable only (dormancy)
- The pilot rule `PackageDirectoryMatchLinked` passes main's carried test literals byte-verbatim (position Offset 20 / Line 3 / Column 1, SourceName, both message formats, module-qualified prefix cells)
- The batch parity mechanism is proven end-to-end: `requirePorted` RED guard fails on its own assertion while the package compiles; `assertBatchParity` runs both engines in-process on one compiling fixture tree and asserts the expected old-side count BEFORE the multiset compare (T-04-02 anti-vacuity); `fingerprint` sorts both sides by the D-01 key
- Position mapping locked by `TestDeclPositionProbe`: `Decl.Pos == fromMetaPos(Meta.Pos)` on Offset/Line/Column for package, import, file-level option, message, plain field, oneof, oneof-member field, enum, enum value, service, rpc — the research A1 remainder is closed and the mapping table is recorded in the probe comment
- Standing constraint (a) has its behavioral proof: `TestReadSourceOnce` shows every corpus file opened exactly once (total 2) on a compiling two-file corpus with an intra-corpus import; suppression centrality cells stayed green (standing constraint b)
- `TestModuleRuleSet` membership guard green; exact-set strengthening lands in 04-07

## Task Commits

Each task was committed atomically:

1. **Task 1 (tracer): linked seam + NewLinked registry + pilot rule** - `6961a41` (feat)
2. **Task 2: batch parity harness + Wave-0 pilot cells** - `afd37c8` (test)
3. **Task 3: position probe, readSource-once cell, ModuleRule-set guard** - `f5ee4e6` (test)

**Plan metadata:** docs commit (this SUMMARY + state/roadmap/validation updates).

## TDD Cycles

- **Task 1 (tracer, production code):** not a `tdd="true"` task and the plan frontmatter is `type: execute`, so the RED/GREEN gate does not fire; the tracer's behavior proof is main's carried test oracle re-driven through the new contract, green on the first `-race` run. The tracer feedback gate was re-run end-to-end after expansion (auto-mode row 2) and passed: `go test -race -count=1 ./internal/rules/ -run 'TestPackageDirectoryMatch' && go build ./...` → OK.
- **Tasks 2-3 (test-only):** exempt from the RED gate; the cells are designed to pass against the proven engine (03-02/03-03 precedent). All green on first run.

## TDD Gate Compliance

- Plan `type: execute`; no task carries `tdd="true"` → MVP+TDD gate not applicable to this plan.
- Production commit `6961a41` (feat) carries its behavior proof (carried-oracle tests) in the same commit by plan design (Wave-0 cells target the proven Phase-3 engine, not new engine behavior).
- No `test(...)`-before-`feat(...)` sequence is required or claimed for this plan; the `test(04-01)` commits are additive test infrastructure.

## Files Created/Modified

- `internal/core/lint_model.go` - AppendLinkedIssue + LinkedRuleName LintRuleName() preference; rename-map header extended (AppendIssue 5-arg -> 4-arg linked, GetRuleName -> LinkedRuleName + override)
- `internal/rules/builder_linked.go` - NEW: NewLinked registry (pilot entry + 40 placeholders) with the rule-type rename map in the header
- `internal/rules/package_directory_match_linked.go` - NEW: pilot rule port (etalon walk mechanics, main semantics, byte-verbatim struct tags, canonical name override)
- `internal/rules/init_test.go` - kept old start/parseFile; added parseLinkedFile, parseProtoLinked, validate (Collect/Finish driver)
- `internal/rules/package_directory_match_test.go` - rewritten onto the linked contract; main's expected literals carried verbatim
- `internal/rules/parity_harness_test.go` - NEW: requirePorted / assertBatchParity / fingerprint
- `internal/rules/parity_wave0_test.go` - NEW: mismatch_dir_package, match_dir_package, nested_decls_shallow cells
- `internal/rules/module_rules_guard_test.go` - NEW: TestModuleRuleSet D-03 membership guard
- `internal/core/position_probe_test.go` - NEW: TestDeclPositionProbe, 11 anchors + mapping-table comment
- `internal/core/lint_linked_readsource_test.go` - NEW: countingWalker + TestReadSourceOnce

## Decisions Made

- `assertBatchParity` signature carries `wantOldCount`: the anti-vacuity count assertion precedes the multiset compare inside one engine run (T-04-02), instead of each cell running the old engine twice
- `TestModuleRuleSet` shape: per-name registry probing with skip for unported names (see Deviations 2)
- `TestReadSourceOnce` uses a literal `KnownLintRules` list (see Deviations 3)
- Linked helpers (`parseLinkedFile`/`parseProtoLinked`/`validate`) live beside the old helpers, not replacing them — unported rules' tests keep compiling

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] go-protoparser MessageBody field name corrected in the probe**
- **Found during:** Task 3 (first compile of position_probe_test.go)
- **Issue:** `MessageBody.OneOfs` does not exist in go-protoparser v4.14.2 — the field is `Oneofs`
- **Fix:** renamed the accessor (model content unchanged; all 11 anchors then green)
- **Files modified:** internal/core/position_probe_test.go
- **Verification:** TestDeclPositionProbe passes all 11 sub-cells under -race
- **Committed in:** f5ee4e6

**2. [Rule 3 - Blocking] TestModuleRuleSet construction shape adjusted to stay green today**
- **Found during:** Task 3 (design-time verification of the plan snippet)
- **Issue:** the plan's literal `rules.NewLinked(config.LintConfig{Use: rules.AllRuleNames()})` errors today (the registry holds only the pilot, so the first unported name fails lookup) — contradicting the plan's own "(green today)" requirement
- **Fix:** the guard probes each AllRuleNames entry individually and skips not-yet-ported names; every ported ModuleRule must still report a D-03 set name; 04-07 strengthens to exact-set equality as planned
- **Files modified:** internal/rules/module_rules_guard_test.go
- **Verification:** TestModuleRuleSet green under -race; a hand-probe confirmed requirePorted-style failures stay assertion-based
- **Committed in:** f5ee4e6

**3. [Rule 3 - Blocking] KnownLintRules literal list in the core test**
- **Found during:** Task 3 (design-time check of TestReadSourceOnce construction)
- **Issue:** the plan's `KnownLintRules: rules.AllRuleNames()` is unimplementable in package core — internal/rules imports internal/core (import cycle), recorded as the 03-01 engine-test precedent
- **Fix:** literal two-name list; corpus has no directives so validation behavior is unaffected
- **Files modified:** internal/core/lint_linked_readsource_test.go
- **Verification:** TestReadSourceOnce green under -race with exact per-file Open counts
- **Committed in:** f5ee4e6

---

**Total deviations:** 3 auto-fixed (3 blocking, Rule 3)
**Impact on plan:** All three were mechanical plan-snippet corrections (one field name, two import-cycle/green-today impossibilities); no scope change, no behavior drift, old path byte-frozen throughout.

## Issues Encountered

None beyond the deviations above; every task's verify passed on its first or second run.

## Deferred Decisions (for 04-07 consolidation ledger)

- No lab-improvement behaviors were adopted or excluded in Wave 0 — the pilot rule is verbatim-from-lab; the catalogued main-vs-lab divergences (NamePos positions, helper casing, report-everyone grouping, cross-file uniqueness, resolution-based ImportUsed, swift_prefix fix) are owned by batch plans 04-02..04-06 per the research inventory
- buildCore/Options.LinkedRules production wiring: deliberately deferred to Phase 5 (dormancy invariant; reaffirmed here as the planner decision 04-RESEARCH Open Question 3 resolution)
- The spec-less edge-probe flag (LINT-PARITY-02, one unresolved unclassified row) remains a phase-level flag — surfaced in the plan frontmatter, inherited unchanged by the batch plans

## User Setup Required

The plan's `user_setup` item (etalon checkout availability) was verified met at execution start: `/tmp/gsd-research/compiler-lab` at SHA `08cef10909c0722f7042ba17b5c11951f9c6b6f6`. No action needed now; if `/tmp` is wiped before later batch plans, re-clone the private repo (see [04-USER-SETUP.md](./04-USER-SETUP.md)).

## Verification Log

- `go test -race -count=1 ./internal/rules/ -run 'TestPackageDirectoryMatch'` → ok (3 tests, -v) — Task 1 verify
- `go build ./...` → ok
- `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityWave0' -v` → ok (3 cells) — Task 2 verify; checker advisory followed (-v cell counts: 1 / 0 / 0 asserted before compares)
- `go test -race -count=1 ./internal/core/ -run 'TestDeclPositionProbe|TestReadSourceOnce' -v` → ok (11 probe sub-cells + readSource cell) — Task 3 verify
- `go test -race -count=1 ./internal/rules/ -run 'TestModuleRuleSet'` → ok
- `go mod tidy && git diff --exit-code go.mod go.sum` → TIDY_CLEAN
- `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` → SCHEMA_CLEAN
- `go test -race -count=1 ./internal/core/ -run 'Suppression'` → ok (standing constraint b)
- `go test -race -count=1 ./internal/core/... ./internal/rules/...` → ok (quick suite)
- `git diff -- internal/core/lint.go internal/api/ internal/rules/builder.go <old rule files>` → empty (dormancy, T-04-01)
- RED-guard probe: an unported name fails with `rule not ported to the linked registry yet: MESSAGE_PASCAL_CASE: invalid rule: ...` — assertion failure, package still compiles (probe file removed after the check)
- Commits measured: `git rev-list --count 21f6b96..HEAD` = 3

## Next Phase Readiness

- Ready for 04-02 (batch 1 — MINIMAL): the harness, helpers, position table, and RED guard are proven; batch 1 needs only its rule ports, carried tests, and fixture cells; the pilot is re-verified inside batch 1 per the research batch table
- Every later batch consumes: `requirePorted` (RED guard), `assertBatchParity` (comparison), `parseLinkedFile`/`parseProtoLinked`/`validate` (test helpers), `AppendLinkedIssue` + `LintRuleName` (port skeleton), the locked position table
- No blockers; standing constraints (a) and (b) are mechanized as green cells that every batch keeps green

---
*Phase: 04-perenos-43-pravil-s-batch-paritetom*
*Completed: 2026-11-19*
