---
phase: 06-novye-pravila-i-regeneratsiya-shem-mcp
plan: 01
subsystem: lint-rules
tags: [lint-rules, verbatim-port, compiler-lab, protoast, module-rule, testify, testdata]

# Dependency graph
requires:
  - phase: 05-flip-i-konsolidatsiya
    provides: single lint engine with core.ModuleRule (Collect/Finish) + core.AppendIssue 4-arg + reflection GetRuleName; canonical unsuffixed rule type names
  - phase: 02-dvukhprokhod... (linker phase)
    provides: protoast model (FilePosition, Decl.NamePos, Import.Pos, Field.Label) and parseLinkedFile harness semantics
provides:
  - Three lab-only rule types under unsuffixed filenames, byte-verbatim @ 08cef10, UNREGISTERED (dormant): SyntaxSpecified, FieldNotRequired, PackageNoImportCycle
  - Fresh colocated testify tests pinning lab semantics (messages, FilePosition 1:1, NamePos declared-name token, import-keyword Pos, SourceName chains, IssueInfo.Path, state reset)
  - Shared repo-root testdata fixtures reused later by 06-02 (golden corpus) and 06-03 (teeth corpus): no_syntax/, required_field/, import_cycle/ (3-file/2-package cycle + same-package negative)
affects: [06-02 registry flip, 06-03 parity teeth corpus, Phase 7 corpus gate]

# Actuals (#2632) — chars/4 over the realized diff (14204 added chars).
actuals:
  tokens: 3551
  tasks: 3
  commits: 4
  plan_head_before: 9d04ab2827718a8bd9dce60720ce09c261d83797
  plan_head_after: 65d81bc6eeeefa6c4c7dc7d7575038b17665629d

# Tech tracking
tech-stack:
  added: []  # zero new dependencies; go.mod/go.sum diff-empty after go mod tidy
  patterns:
    - "verbatim port diff-gate: cp from the lab checkout + empty diff against pinned SHA 08cef10 per file"
    - "package-cycle fixture shape: file DAG stays acyclic (3 files / 2 packages) while the package graph cycles — the only shape the real linker accepts"
    - "hand-computed position anchors documented in test comments (Phase-1 precedent), validated by first-run green"

key-files:
  created:
    - internal/rules/syntax_specified.go
    - internal/rules/syntax_specified_test.go
    - internal/rules/field_not_required.go
    - internal/rules/field_not_required_test.go
    - internal/rules/package_no_import_cycle.go
    - internal/rules/package_no_import_cycle_test.go
    - testdata/no_syntax/no_syntax.proto
    - testdata/required_field/required_field.proto
    - testdata/import_cycle/a.proto
    - testdata/import_cycle/b/b.proto
    - testdata/import_cycle/a2.proto
    - testdata/import_cycle/sp1.proto
    - testdata/import_cycle/sp2.proto
  modified:
    - internal/rules/init_test.go (fixture consts only: +7 lines)
    - .planning/phases/06-.../06-VALIDATION.md (task rows filled with real -run patterns)

key-decisions:
  - "Verbatim-or-nothing: each rule file copied with cp from the lab checkout @ 08cef10 and gated by an empty diff; zero seam adaptations (lab positions FilePosition/NamePos/imp.Pos kept — lab parity is the oracle for NEW rules, not main-wins)"
  - "Fixtures live at repo-root testdata/ (parseLinkedFile resolves imports from ./../../testdata), not internal/rules/testdata; the cycle corpus is a 3-file/2-package PACKAGE cycle because the easypc linker hard-rejects direct file A<->B cycles"
  - "TDD compile-gap RED per plan: type-referencing tests cannot be machine-validated RED before the verbatim port lands; machine-validated RED deferred to 06-02 Task 1 (registry guard flip) — 06-01 tasks land as single atomic feat commits"
  - "Registry untouched is proven, not assumed: TestModuleRuleSet exact ten-name set + TestAllRuleNames + TestAllGroups green and schema-gen drift empty on the final state"

requirements-completed: [LINT-NEWRULES-01, LINT-NEWRULES-02, LINT-NEWRULES-03]  # copied verbatim from plan frontmatter; REQUIREMENTS.md checkboxes stay open this plan — all three IDs are shared with 06-02/06-03 and complete when those plans finish (shared-ID gate, #2388)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "SYNTAX_SPECIFIED ported byte-verbatim @ 08cef10 under internal/rules/syntax_specified.go (unsuffixed) with a fresh test pinning the lab FilePosition semantics (1:1, Filename+SourceName=path) and the proto3-valid negative"
    requirement: LINT-NEWRULES-02
    verification:
      - kind: other
        ref: "command: diff /tmp/gsd-research/compiler-lab/easyp-cli/internal/rules/syntax_specified.go internal/rules/syntax_specified.go (empty, SHA 08cef10909c0722f7042ba17b5c11951f9c6b6f6 verified)"
        status: pass
      - kind: unit
        ref: "internal/rules/syntax_specified_test.go#TestSyntaxSpecified_Message"
        status: pass
      - kind: unit
        ref: "internal/rules/syntax_specified_test.go#TestSyntaxSpecified_Validate (invalid + valid cells, -race)"
        status: pass
    human_judgment: false
  - id: D2
    description: "FIELD_NOT_REQUIRED ported byte-verbatim with a fresh visitor-walk test pinning the declared-name token position (NamePos 6:18/offset 71, hand-computed anchor), bare field name SourceName, and optional/repeated zero-finding cells"
    requirement: LINT-NEWRULES-02
    verification:
      - kind: other
        ref: "command: diff /tmp/gsd-research/compiler-lab/easyp-cli/internal/rules/field_not_required.go internal/rules/field_not_required.go (empty)"
        status: pass
      - kind: unit
        ref: "internal/rules/field_not_required_test.go#TestFieldNotRequired_Message"
        status: pass
      - kind: unit
        ref: "internal/rules/field_not_required_test.go#TestFieldNotRequired_Validate (invalid + valid_proto3 + optional_and_repeated cells, -race)"
        status: pass
    human_judgment: false
  - id: D3
    description: "PACKAGE_NO_IMPORT_CYCLE ported byte-verbatim with 3-file/2-package package-cycle + same-package fixtures and a fresh Collect/Finish test pinning both cycle directions (SourceName chains, import-keyword positions 5:1/35, IssueInfo.Path), acyclic/same-package negatives, state reset and the Validate stub"
    requirement: LINT-NEWRULES-01
    verification:
      - kind: other
        ref: "command: diff /tmp/gsd-research/compiler-lab/easyp-cli/internal/rules/package_no_import_cycle.go internal/rules/package_no_import_cycle.go (empty)"
        status: pass
      - kind: unit
        ref: "internal/rules/package_no_import_cycle_test.go#TestPackageNoImportCycle_Message"
        status: pass
      - kind: unit
        ref: "internal/rules/package_no_import_cycle_test.go#TestPackageNoImportCycle_CollectFinish (package_cycle/acyclic_negative/same_package_negative/state_reset/validate_stub cells, -race -v)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Registry dormancy provable on the final state: the three types are NOT registered — exact ten-name TestModuleRuleSet guard, TestAllRuleNames, TestAllGroups stay green and schema-gen drift is empty"
    requirement: LINT-NEWRULES-01
    verification:
      - kind: unit
        ref: "internal/rules/module_rules_guard_test.go#TestModuleRuleSet (exact ten-name set, -race)"
        status: pass
      - kind: other
        ref: "command: go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ (empty)"
        status: pass
    human_judgment: false

# Metrics
duration: 6 min
completed: 2026-11-20
status: complete
---

# Phase 6 Plan 1: Verbatim rule ports with fresh colocated tests Summary

**Three lab-only lint rules (PACKAGE_NO_IMPORT_CYCLE, SYNTAX_SPECIFIED, FIELD_NOT_REQUIRED) ported byte-verbatim from compiler-lab @ 08cef10 with fresh colocated testify tests and shared package-cycle fixtures — the registry stays provably untouched for 06-02's atomic registration flip.**

## Performance

- **Duration:** 6 min
- **Started:** 2026-11-20T23:40:17Z
- **Completed:** 2026-11-20T23:45:45Z
- **Tasks:** 3
- **Files modified:** 14 (13 created, init_test.go modified; +532 insertions, 0 deletions)

## Accomplishments
- Byte-verbatim ports of all three lab rule files (per-file `diff` gates empty against the pinned checkout SHA 08cef10909c0722f7042ba17b5c11951f9c6b6f6), unsuffixed filenames per AGENTS.md new-rule convention
- Fresh colocated tests pin the lab semantics the plan demands: `SYNTAX_SPECIFIED` at FilePosition 1:1 with SourceName=path, `FIELD_NOT_REQUIRED` at the declared-name token (hand-computed NamePos 6:18/offset 71) with the bare field name, `PACKAGE_NO_IMPORT_CYCLE` reporting every cycle edge at the import keyword with `a.v1 -> b.v1 -> a.v1` / `b.v1 -> a.v1 -> b.v1` chains and IssueInfo.Path = the importing file
- Shared fixtures under repo-root `testdata/`: `no_syntax/`, `required_field/`, and `import_cycle/` (3-file/2-package package cycle + same-package negative) — the shapes 06-02's golden-corpus extension and 06-03's teeth corpus reuse
- Dormancy proven on the final state: full `./internal/rules/` suite green under -race including the exact ten-name `TestModuleRuleSet` guard, and `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` empty

## Task Commits

Each task was committed atomically:

1. **Task 1: Tracer — SYNTAX_SPECIFIED verbatim port with fresh colocated test** - `bb27941` (feat)
2. **Task 2: FIELD_NOT_REQUIRED verbatim port with fresh colocated test** - `014d049` (feat)
3. **Task 3: PACKAGE_NO_IMPORT_CYCLE verbatim port with package-cycle fixtures and fresh cross-file test** - `dc5bfc6` (feat)
4. **Index hygiene: legacy duplicate 06-RESEARCH.md path dedup** - `65d81bc` (chore — orchestrator-staged cleanup, see Issues Encountered)

_Plan metadata commit: see final docs commit (SUMMARY/STATE/ROADMAP/REQUIREMENTS/VALIDATION)._

## Files Created/Modified
- `internal/rules/syntax_specified.go` - verbatim lab rule (core.Rule): fires "syntax should be specified" at protoast.FilePosition when SyntaxSpecified is false
- `internal/rules/field_not_required.go` - verbatim lab rule (core.Rule): visitor-walk fires "field should not be required" at Decl.NamePos for `required` labels
- `internal/rules/package_no_import_cycle.go` - verbatim lab rule (core.ModuleRule): package->import graph keyed via core.ConvertImportPath, DFS cycle detection, every-edge reporting, state reset in Finish
- `internal/rules/{syntax_specified,field_not_required,package_no_import_cycle}_test.go` - fresh tests (lab carries none); map-driven tables + Collect/Finish subtests per go-testing skill
- `testdata/no_syntax/no_syntax.proto` - no-syntax-statement fixture (proto2 default, SyntaxSpecified false)
- `testdata/required_field/required_field.proto` - proto2 fixture: required/optional/repeated fields
- `testdata/import_cycle/{a.proto,b/b.proto,a2.proto,sp1.proto,sp2.proto}` - 3-file/2-package package cycle + same-package negative (file DAG acyclic)
- `internal/rules/init_test.go` - +7 fixture consts (noSyntaxProto, requiredFieldProto, importCycle{A,B,A2,SamePkg1,SamePkg2})
- `06-VALIDATION.md` - per-task verification map filled with the real -run patterns; Wave-0 fixtures checkbox closed

## Decisions Made
- **Verbatim-or-nothing with mechanical proof:** each port is a `cp` from the lab checkout gated by an empty `diff`; no retype of positions, no renames, no suppression logic added. Lab positions (FilePosition 1:1, NamePos, import-keyword Pos) are kept because lab parity (LINT-NEWRULES-03) is the oracle for NEW rules.
- **Fixture locality and shape:** repo-root `testdata/` (the harness resolves imports from `./../../testdata`, correcting RESEARCH Wave 0's `internal/rules/testdata/`), and the cycle corpus is the 3-file/2-package shape (06-PATTERNS correction) because the linker rejects direct file cycles (pinned 2:1 in link_test.go).
- **Hand-computed position anchors documented in test comments** (Phase-1 precedent) and validated by first-run green — no literals copied from failure output.
- **Single atomic feat commit per task** (objective: "Commit each task atomically"), matching the Phase-4 batch-port precedent; see TDD Gate Compliance for the RED handling.

## Deviations from Plan

None - plan executed exactly as written. (All 13 planned artifacts + init_test.go consts landed; every `<automated>` verify chain ran green; no seam adaptations were needed.)

## TDD Gate Compliance

`workflow.tdd_mode` is ON and all three tasks carry `tdd="true"` with `<behavior>` blocks, so the runtime gate applies. Compliance status, per the plan's own ratified resolution (06-01-PLAN Task 1 action):

- **RED commit gate:** no `test(06-01)` commits exist — all three tasks land as single atomic `feat(06-01)` commits per the plan objective's "Commit each task atomically".
- **Why:** the plan explicitly defines the task-level RED as the compile gap — each fresh test references `rules.<Type>{}` and cannot compile until the verbatim port lands (`undefined: rules.SyntaxSpecified` / `rules.FieldNotRequired` / `rules.PackageNoImportCycle`, captured for all three tasks under `-race`). Per `tdd.md`'s format-based evidence rules that is `build_failure` = INVALID_RED, i.e. type-referencing tests CANNOT be machine-validated RED here; committing them alone would leave the package build broken at an intermediate commit.
- **Machine-validated RED deferred to 06-02 Task 1** (the plan's stated owner): the registry-surface guard flip (`builder_test.go:69` NotContains→Contains) fails on an assertion pre-change and can be machine-validated (`gsd_run check tdd-red-evidence`).
- **GREEN:** every behavior cell is green on the final state under `-race -v` (message cells, position/SourceName cells, both cycle directions, four negative/state cells).
- **REFACTOR:** not applicable — verbatim ports forbid cleanup; the diff gates would catch any body change.

## Issues Encountered
- **Index sweep at the first task commit:** the session started with a pre-staged deletion of the legacy duplicate `.planning/phases/06-...-skhem-mcp/06-RESEARCH.md` path (orchestrator-side dedup from the phase-directory rename). `git commit` commits the whole index, so task commit `bbd4986` accidentally included it. Handled immediately: local-only redo (`git reset HEAD~1`, working tree untouched) producing the clean `bb27941`, and the dedup landed separately as `chore(06-01)` `65d81bc`. Task commits are exactly-atomic; content loss: none (the canonical `-shem-mcp` path remains tracked and unchanged).
- **Pre-existing unrelated working-tree modifications** (`.agents/skills/sdd/scripts/pipeline.sh`, `.planning/03-VERIFICATION.md`, `.planning/STATE.md` baseline edits) were left untouched and are not part of any commit from this plan.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- All three rule types exist verbatim and compile against the single engine contracts (`var _ core.ModuleRule = (*PackageNoImportCycle)(nil)` / `var _ core.Rule = ...`); ready for 06-02's atomic registration commit (builder.go add* lists + New + selection.go special-case deletion + five pin flips + schema regen).
- Fixture shapes are in place for 06-02's golden-corpus extension and 06-03's teeth corpus; the lab checkout SHA was re-verified this session (Pitfall 10 covered).
- No blockers. Requirement IDs LINT-NEWRULES-01/02/03 remain open in REQUIREMENTS.md by design (shared with 06-02/06-03; the shared-ID gate blocks early completion).

## Self-Check: PASSED

- All 13 created files exist on disk (13/13 FOUND)
- All 4 commits are ancestors of HEAD: bb27941, 014d049, dc5bfc6, 65d81bc (4/4 FOUND)
- init_test.go carries exactly the 7 planned fixture consts
- Measured commits 9d04ab2..HEAD = 4; plan_head_after = 65d81bc6eeeefa6c4c7dc7d7575038b17665629d

---
*Phase: 06-novye-pravila-i-regeneratsiya-shem-mcp*
*Completed: 2026-11-20*
