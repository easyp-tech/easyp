---
phase: 06-novye-pravila-i-regeneratsiya-shem-mcp
plan: 02
subsystem: lint-rules
tags: [lint-rules, registry, json-schema, mcp, golden-corpus, tdd, atomic-commit]

# Dependency graph
requires:
  - phase: 06-novye-pravila-i-regeneratsiya-shem-mcp (06-01)
    provides: three byte-verbatim rule types (SyntaxSpecified, FieldNotRequired, PackageNoImportCycle) dormant in internal/rules + package-cycle/no-syntax/required-field fixture shapes
  - phase: 05-flip-i-konsolidatsiya
    provides: single lint engine, canonical unsuffixed names, reflection GetRuleName, rules.New as THE validator, schema-gen derivation loop
provides:
  - 44-rule registry with the three new rules at lab-verbatim placement (PACKAGE_NO_IMPORT_CYCLE 5th MINIMAL as core.ModuleRule; SYNTAX_SPECIFIED/FIELD_NOT_REQUIRED 21st/22nd BASIC)
  - One validation surface (D-06): the PACKAGE_NO_IMPORT_CYCLE "is not implemented" special case deleted; deny-by-default slices.Contains path serves every name
  - Regenerated schemas/easyp-v1.schema.json + schemas/easyp.schema.json (3 enum spots x 3 new names) and MCP Describe enums through the closed derivation loop — zero MCP/v1 schema code changes
  - Extended golden corpus (18 files) with re-captured 94-finding golden: every one of the 44 registry names fires at least once
  - Machine-validated RED evidence (6 records, RED_EVIDENCE_OK) at tdd-red-evidence-06-02-task1.json
affects: [06-03 lab parity teeth corpus, phase-7 main-vs-new corpus compare (filter the three names), migration/directive surfaces]

# Actuals (#2632) — chars/4 over the realized diff (29453 chars).
actuals:
  tokens: 7363
  tasks: 3
  commits: 2
  plan_head_before: 79a00706543164c0fbf05a14302e8c093cde3257
  plan_head_after: 70bbaff9837fa5e928b8fff73cf4397c07266dee

# Tech tracking
tech-stack:
  added: []  # zero new dependencies; go.mod/go.sum diff-empty after go mod tidy
  patterns:
    - "atomic registration commit: registry + validator + pin flips + golden extension + schema regen in ONE revertable commit (05-03 ratified precedent)"
    - "golden re-capture via a temporary engine-dump probe test (deleted pre-commit): existing literals proven byte-identical, new findings transcribed in engine D-01 order"
    - "RED evidence per-record machine validation: the tdd-red-evidence checker parses one record per file, so the combined array artifact is validated by extracting each record to a temp file"

key-files:
  created:
    - .planning/phases/06-novye-pravila-i-regeneratsiya-shem-mcp/tdd-red-evidence-06-02-task1.json
  modified:
    - internal/rules/builder.go
    - internal/rules/builder_linked.go
    - internal/rules/selection.go
    - internal/rules/builder_test.go
    - internal/rules/module_rules_guard_test.go
    - internal/rules/validation_parity_test.go
    - internal/rules/selection_test.go
    - internal/rules/golden_full_registry_test.go
    - internal/core/lint_model.go
    - internal/config/v1/lint_selection_test.go
    - internal/config/v1/policy_merge_test.go
    - mcp/easypconfig/describe_test.go
    - schemas/easyp-v1.schema.json
    - schemas/easyp.schema.json

key-decisions:
  - "ONE feat commit (5be9f7e, exactly the plan's 13 files) carries the registry wiring, validator deletion, six pin flips, golden extension/re-capture and regenerated schemas — the ratified atomic-flip precedent (05-03); RED evidence machine-validated but uncommitted until this commit"
  - "Golden corpus extension is mandatory (A2 resolved-by-design as false-as-stated): requireGoldenPurity's coverage invariant went RED once the registry hit 44 with three silent names; 5 fixtures + re-captured 94-finding golden landed in the same commit"
  - "The three new-rule golden literals (cycle chains 'a.v1 -> b.v1 -> a.v1'/'b.v1 -> a.v1 -> b.v1' at import keyword 6:1, SYNTAX_SPECIFIED at 1:1 with SourceName=path, FIELD_NOT_REQUIRED at NamePos 7:18 source 'id') are engine-captured and pending the 06-03 lab CLI cross-check per plan"
  - "MCP docs.go:38 'unimplemented' wording stays AS-IS (resolved prohibition): the lab carries the identical wording; polishing would create a lab-vs-easyp divergence for zero parity gain"

patterns-established:
  - "Registry-growth pin sweep: BASIC/MINIMAL group-size pins (assert.Len literals) must be swept when the registry grows — policy_merge_test.go held two the plan's verify pattern did not cover"
  - "Engine-dump probe pattern for golden re-capture: temp test dumps findings as golden() lines, existing entries verified byte-identical via multiset diff, probe deleted before commit"

requirements-completed: [LINT-NEWRULES-01, LINT-NEWRULES-02, LINT-NEWRULES-04]  # copied verbatim from plan frontmatter; REQUIREMENTS.md checkboxes: LINT-NEWRULES-04 completes this plan; 01/02 stay open (shared with 06-03, shared-ID gate #2388)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Registry holds 44 rules at lab-verbatim placement: addMinimal 5th entry PackageNoImportCycle, addBasic 21st/22nd SyntaxSpecified+FieldNotRequired, mirrored in New's allRules; TestModuleRuleSet extends to the D-03 eleven"
    requirement: LINT-NEWRULES-01
    verification:
      - kind: unit
        ref: "internal/rules/module_rules_guard_test.go#TestModuleRuleSet (exact eleven-name ElementsMatch, -race)"
        status: pass
      - kind: unit
        ref: "internal/rules/validation_parity_test.go#TestLintUseValidationParity/package_no_import_cycle_constructs (New(Use:[PACKAGE_NO_IMPORT_CYCLE]) constructs; except/ignore_only accept the name)"
        status: pass
      - kind: other
        ref: "command: grep -c 'core.GetRuleName(&PackageNoImportCycle{})' internal/rules/builder.go = 1; grep -c '&PackageNoImportCycle{}' internal/rules/builder_linked.go = 1 (name-list/constructor agreement)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Validation is one deny-by-default surface (D-06): the 4-line PACKAGE_NO_IMPORT_CYCLE special case deleted from selection.go; the name validates through the generic slices.Contains path in use/except/ignore_only"
    requirement: LINT-NEWRULES-01
    verification:
      - kind: unit
        ref: "internal/rules/selection_test.go#TestInvalidLintNamesAreRejectedBeforeFiltering/package_no_import_cycle_is_valid (-race)"
        status: pass
      - kind: unit
        ref: "internal/rules/validation_parity_test.go#TestLintUseValidationParity (unknown-name rejection cells still green — validation not loosened)"
        status: pass
      - kind: other
        ref: "command: grep -c 'is not implemented' internal/rules/selection.go = 0"
        status: pass
    human_judgment: false
  - id: D3
    description: "Schemas and MCP updated through the closed derivation loop: go run ./cmd/easyp schema-gen adds the 3 new names at exactly 3 enum spots each in easyp-v1.schema.json and easyp.schema.json; MCP Describe enums contain all three via v1.SchemaJSON -> AllLintUseValues (zero MCP/v1 code changes)"
    requirement: LINT-NEWRULES-04
    verification:
      - kind: other
        ref: "command: go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ (drift empty on committed state; regen idempotent)"
        status: pass
      - kind: other
        ref: "command: grep -c PACKAGE_NO_IMPORT_CYCLE/SYNTAX_SPECIFIED/FIELD_NOT_REQUIRED per schema file = 3 each; git show --stat HEAD lists exactly 2 schema files"
        status: pass
      - kind: unit
        ref: "mcp/easypconfig/describe_test.go#TestDescribeLintSelectionMatchesExecutableRules (all 3 paths, -race)"
        status: pass
      - kind: unit
        ref: "internal/config/v1/lint_selection_test.go#TestLintSelectionValidationMatchesRuntimeCatalog (schema Contains + every AllLintUseValues name validates, -race)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Golden corpus extended (5 Phase-6 fixtures: 3-file package cycle over an acyclic file DAG, no-syntax file, proto2 required field) and goldenIssues re-captured 73 -> 94 with every one of the 44 registry names firing at least once"
    requirement: LINT-NEWRULES-04
    verification:
      - kind: unit
        ref: "internal/rules/golden_full_registry_test.go#TestLintGoldenFullRegistry (94-finding golden compare incl. Filename/Offset, requireGoldenPurity coverage + requireTotalOrder, -race)"
        status: pass
    human_judgment: false
  - id: D5
    description: "RED machine-validated: 6 name-absence pins fail on planned assertions before the atomic commit; evidence records at tdd-red-evidence-06-02-task1.json with RED_EVIDENCE_OK for every record"
    requirement: LINT-NEWRULES-01
    verification:
      - kind: other
        ref: "command: gsd-tools check tdd-red-evidence per record — 6x verdict=RED_EVIDENCE_OK reason=target_test_failed format=go-test"
        status: pass
    human_judgment: false

# Metrics
duration: 20 min
completed: 2026-11-21
status: complete
---

# Phase 6 Plan 2: Registry atomic commit Summary

**PACKAGE_NO_IMPORT_CYCLE, SYNTAX_SPECIFIED and FIELD_NOT_REQUIRED registered in one atomic commit (5be9f7e): lab-verbatim registry placement, single-surface validation, regenerated schemas/MCP enums, and a 94-finding golden over the extended 18-file corpus.**

## Performance

- **Duration:** 20 min
- **Started:** 2026-11-20T23:52:25Z
- **Completed:** 2026-11-21T00:12:48Z
- **Tasks:** 3
- **Files modified:** 14 (1 evidence record created; 13 production/test/schema files in the feat commit + 1 missed-pin fix)

## Accomplishments
- THE atomic commit `5be9f7e` (exactly 13 files, +193/-52): registry wiring in `builder.go` add* lists and `builder_linked.go` New's allRules at lab-verbatim order @ 08cef10 (verified: lab addMinimal :227-228, addBasic :253-255, allRules :73-74/:96-98), the 4-line validation special-case deletion, six test-pin flips, the golden-corpus extension with re-captured 94-finding golden, and the regenerated schemas
- Validation is one deny-by-default surface (D-06): `PACKAGE_NO_IMPORT_CYCLE` now flows through the generic `slices.Contains` path in use/except/ignore_only; every unknown-name rejection cell still passes (validation not loosened — threat T-06-04 mitigation held)
- Schemas/MCP moved by the closed derivation loop only: `go run ./cmd/easyp schema-gen` regenerated exactly `schemas/easyp-v1.schema.json` and `schemas/easyp.schema.json` (3 enum spots x 3 new names each; other four schema files untouched; regen idempotent, drift gate empty on the committed state); MCP Describe enums carry all three names with zero MCP code changes
- Golden corpus extended by 5 Phase-6 fixtures (18 files total) and `goldenIssues` re-captured 73 → 94: the existing 73 literals proven byte-identical via multiset diff before transcription; `TestLintGoldenFullRegistry` green under -race with requireGoldenPurity's 44-name coverage invariant and requireTotalOrder clean
- Battery on the committed final state: six scoped -race suites green (wazero carve-out respected), suppression-centrality (TestLintSuppressionParity, TestLintMatrix — 9 matrix cells) and disk-instrumentation (TestReadSourceOnce) standing cells green, golangci 0 issues on touched packages, tidy diff-empty, collateral freeze gate empty

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): six name-absence pins flipped, machine-validated** — no commit by design (ratified atomic-flip precedent; evidence record committed with plan metadata)
2. **Task 2 (tracer/GREEN): THE atomic registration commit** - `5be9f7e` (feat, 13 files)
3. **Task 3: standing battery + missed-pin fix + docs** - `70bbaff` (fix — Rule 1, see Deviations) + plan metadata commit

_Plan metadata commit: see final docs commit (SUMMARY/STATE/ROADMAP/REQUIREMENTS/VALIDATION + RED evidence record)._

## RED Evidence (Task 1)

Six targets failed pre-implementation on their planned assertions (name absent / set mismatch), all six records machine-validated:

| # | Target | Planned assertion | Verdict |
|---|--------|-------------------|---------|
| 1 | `TestAllLintUseValues` | values does not contain PACKAGE_NO_IMPORT_CYCLE | RED_EVIDENCE_OK (target_test_failed) |
| 2 | `TestModuleRuleSet` | D-03 eleven vs constructed ten — elements differ | RED_EVIDENCE_OK (target_test_failed) |
| 3 | `TestLintUseValidationParity` | package_no_import_cycle_constructs: got "is not implemented" error | RED_EVIDENCE_OK (target_test_failed) |
| 4 | `TestInvalidLintNamesAreRejectedBeforeFiltering` | package_no_import_cycle_is_valid: same error | RED_EVIDENCE_OK (target_test_failed) |
| 5 | `TestLintSelectionValidationMatchesRuntimeCatalog` | schema data does not contain the name | RED_EVIDENCE_OK (target_test_failed) |
| 6 | `TestDescribeLintSelectionMatchesExecutableRules` | Describe enums lack the name (all 3 paths) | RED_EVIDENCE_OK (target_test_failed) |

Semantic assessment per record: every target executed and failed on its planned assertion — no build failures, zero-test discovery, or unrelated failures; the pre-existing unknown-name error cells stayed green, isolating each failure to the planned behavior change. Format: go-test adapter (complete console output incl. package summary lines).

**Checker note:** `gsd-tools check tdd-red-evidence` parses ONE record object per invocation; the plan's combined array artifact was validated by extracting each record to a temp file and running the checker per record (6/6 RED_EVIDENCE_OK). The combined file at the canonical path remains the committed artifact.

## TDD Gate Compliance

`workflow.tdd_mode` is ON and this is a `type: tdd` plan. Compliance status, per the plan's own ratified resolution (06-02-PLAN objective):

- **RED commit gate:** no `test(06-02)` commit exists — the plan's RATIFIED atomic-flip mandate (05-03 precedent: "the plan's RATIFIED atomic-flip mandate requires the cells, deletions and renames in ONE commit") supersedes the strict test-commit-before-feat gate. Task 1's pin flips cannot compile-satisfy any intermediate commit: flipping Contains pins without the registry would leave the suite red at a committed state, and the plan explicitly orders "Work in the working tree WITHOUT committing" for Task 1.
- **Machine-validated RED before GREEN (the gate's substance):** satisfied — six records RED_EVIDENCE_OK + semantic inspection (see RED Evidence) BEFORE the `feat(06-02)` commit landed. An INVALID_RED verdict or failed semantic assessment would have halted Task 2.
- **GREEN:** the single `feat(06-02)` commit turned every RED cell green; the full probe set is green under -race on the committed state.
- **REFACTOR:** not applicable — registry wiring and verbatim-ordered lists admit no cleanup; the count-comment refreshes are plan-mandated comment edits inside the same atomic commit.

## Files Created/Modified
- `internal/rules/builder.go` - addMinimal 5th entry (PackageNoImportCycle), addBasic 21st/22nd (SyntaxSpecified, FieldNotRequired) — lab-verbatim order
- `internal/rules/builder_linked.go` - New's allRules mirrors the same order; doc comment 41 → 44
- `internal/rules/selection.go` - ValidateNames special case deleted; generic Contains path is the only surface
- `internal/rules/builder_test.go` - :69 pin NotContains → Contains (lab mirror)
- `internal/rules/module_rules_guard_test.go` - D-03 eleven-name exact-set guard; header comment rewritten
- `internal/rules/validation_parity_test.go` - special-case error row deleted; package_no_import_cycle_constructs cell added (use/except/ignore_only)
- `internal/rules/selection_test.go` - two "unimplemented" rows deleted; package_no_import_cycle_is_valid cell added
- `internal/rules/golden_full_registry_test.go` - 5 new corpus fixtures, count 73 → 94, re-captured literals, comment refreshes (94/44, never 43)
- `internal/core/lint_model.go` - D-03 comment 41 → 44 with the eleven-name list; REQUIREMENTS-discrepancy note kept
- `internal/config/v1/lint_selection_test.go` - :57 Contains flip; fixture swap (ALSO_UNKNOWN_RULE) keeps the 4-location collection test
- `internal/config/v1/policy_merge_test.go` - BASIC-size pins refreshed 20 → 22 / 19 → 21 (Rule 1 fix)
- `mcp/easypconfig/describe_test.go` - :298 Contains flip (lab mirror)
- `schemas/easyp-v1.schema.json`, `schemas/easyp.schema.json` - regenerated via schema-gen; 3 enum spots x 3 names each
- `tdd-red-evidence-06-02-task1.json` - 6 machine-validated RED records (committed with plan metadata)

## Decisions Made
- **Atomicity over intermediate commits:** Task 1's flips stayed uncommitted until Task 2's single feat commit (ratified 05-03 precedent), keeping every committed state green and the registration revertable as one unit (`git revert 5be9f7e`).
- **Golden re-capture method:** a temporary engine-dump probe test dumped the extended corpus's findings as `golden()` lines; a multiset diff proved the existing 73 literals byte-identical before transcription, and the probe was deleted before the commit. No literals were hand-derived from failure output.
- **Per-record RED validation:** the checker's one-record-per-file contract honored via 6 temp extractions; the combined array remains the canonical committed evidence.
- **docs.go wording untouched** per the resolved prohibition (lab parity).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Refreshed two BASIC-group size pins in policy merge tests**
- **Found during:** Task 3 (standing battery — six scoped suites)
- **Issue:** `TestMergeLinterSectionBaseGroupDisableExpandsEveryRule` (Disable 19) and `TestMergeLinterSectionDedupsRepeatedSelections` (Enable 20) pin the BASIC group's rule count; the registry growth (20 → 22 BASIC rules) made both red. The plan's file list and verify patterns did not cover `policy_merge_test.go`.
- **Fix:** literals refreshed to 21/22 with comments naming the Phase-6 expansion; repo-wide scan found no other registry-size pins.
- **Files modified:** internal/config/v1/policy_merge_test.go
- **Verification:** full `./internal/config/v1/` suite green under -race after the fix
- **Committed in:** `70bbaff` (fix(06-02), separate commit to preserve the feat commit's exactly-13-files acceptance criterion)

---

**Total deviations:** 1 auto-fixed (1 bug — missed pin, directly caused by the registry growth)
**Impact on plan:** No scope creep — the fix is the mechanical fallout of the planned registry growth caught by the battery the plan mandated.

## Issues Encountered
- The Task 2 verify chain's `git diff --exit-code schemas/` step cannot pass pre-commit (the working tree legitimately differs from HEAD). Equivalent gates run instead: regen idempotence (second run diff-empty) pre-commit, and the literal drift gate on the committed state post-commit (passed: `DRIFT_EMPTY_ON_COMMITTED_STATE`).
- The tdd-red-evidence checker accepts one record per file; the combined-array artifact was validated per record (documented above). Not a RED-validity issue — all six verdicts RED_EVIDENCE_OK.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Registry is 44 rules with the three new names live end-to-end: validation, schemas, MCP Describe, directive validation (`easyp:disable PACKAGE_NO_IMPORT_CYCLE` legal via `KnownLintRules = rules.AllRuleNames()`), and migration validation (legacy configs naming the rule now accept — intended behavior change).
- 06-03 inherits: the engine-captured new-rule golden literals to cross-check against the lab CLI (chains `a.v1 -> b.v1 -> a.v1` / `b.v1 -> a.v1 -> b.v1` at import keyword 6:1 offset 36; SYNTAX_SPECIFIED at 1:1 with SourceName=path; FIELD_NOT_REQUIRED at NamePos 7:18/offset 72 source `id`); the lab checkout @ 08cef10 re-verified this session; the extended corpus shapes.
- **Phase-7 note:** filter PACKAGE_NO_IMPORT_CYCLE / SYNTAX_SPECIFIED / FIELD_NOT_REQUIRED out of the main-vs-new corpus compare — main lacks them, and the MINIMAL auto-seed (`policy.go:104`) makes them run on every v1 config.
- No blockers. Requirement IDs LINT-NEWRULES-01/02 remain open in REQUIREMENTS.md by design (shared with 06-03; shared-ID gate); LINT-NEWRULES-04 completes with this plan.

## Self-Check: PASSED

- All 14 modified/created files exist on disk; the feat commit lists exactly the plan's 13 files
- Commits are ancestors of HEAD: 5be9f7e, 70bbaff (2/2 FOUND); measured 79a0070..HEAD = 2 commits
- RED evidence: 6/6 records RED_EVIDENCE_OK; schema gates 3 occurrences x 3 names x 2 files; freeze gate empty
- Measured commits 79a0070..HEAD = 2; plan_head_after = 70bbaff9837fa5e928b8fff73cf4397c07266dee

---
*Phase: 06-novye-pravila-i-regeneratsiya-shem-mcp*
*Completed: 2026-11-21*
