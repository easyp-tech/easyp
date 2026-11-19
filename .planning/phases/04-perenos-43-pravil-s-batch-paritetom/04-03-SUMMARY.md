---
phase: 04-perenos-43-pravil-s-batch-paritetom
plan: 03
subsystem: rules
tags: [lint-rules, protoast, parity-harness, linked-engine, porting, dormancy]

# Dependency graph
requires:
  - phase: 04-perenos-43-pravil-s-batch-paritetom (plans 01-02)
    provides: requirePorted/assertBatchParity/fingerprint harness, parseLinkedFile helper, NewLinked registry (4/41), AppendLinkedIssue + LintRuleName seam, locked Decl.Pos position table, firstExemptMismatches
provides:
  - Ten BASIC-casing ports (EnumPascalCase, EnumValueUpperSnakeCase, FieldLowerSnakeCase, ImportNoPublic, ImportNoWeak, MessagePascalCase, OneofLowerSnakeCase, PackageLowerSnakeCase, RPCPascalCase, ServicePascalCase) — main's regexps byte-verbatim, Decl.Pos keyword positions, shallow main walk shapes (14/41 cumulative)
  - ImportNoPublic/ImportNoWeak verbatim-from-lab (imp.Public/imp.Weak, imp.Pos, quoted Location)
  - TestLinkedParityBasicCasing — five cells pinning casing-edge regexp semantics (T-04-05), keyword positions (T-04-06), shallow walks and the top-level-message enum boundary (T-04-07), and import flags, under cumulative 14-rule configs
  - Ten carried main test rewrites with byte-verbatim expected literals (zero old-contract tests left for the batch)
  - Probe-documented finding: digit-led rpc names are rejected by BOTH parsers — the main-vs-lab RPCPascalCase divergence edge is unreachable in compiling corpora
affects: [04-04..04-06 batches (harness + registry reuse), 04-07 (exact-set guard, deferred-decisions consolidation), Phase 5 flip (rename-map consolidation), Phase 7 corpus gate (RPCPascalCase corpus note)]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 10671    # 42684 diff chars over internal/rules/ at chars/4 (estimate was 78000; uniform ports on a proven harness, one shape x10)
  tasks: 3
  commits: 4       # MEASURED: git rev-list --count 7fc5aae..HEAD minus the metadata docs commit — 3 task commits (RED + 2 GREEN) + 1 fixture-fix commit (5ef30ab)

# Tech tracking
tech-stack:
  added: []
  patterns: [regexp-wins casing ports (casing.go helpers excluded and grep-enforced), keyword positions via Decl.Pos with carried offset/line/col literals as oracle, shallow direct-field iteration replacing the recursion-capable Visitor walk, cumulative cell configs (ten batch names + four MINIMAL names), old-side position-literal assertions through a local oldFindings helper]

key-files:
  created:
    - internal/rules/enum_pascal_case_linked.go
    - internal/rules/enum_value_upper_snake_case_linked.go
    - internal/rules/message_field_lower_snake_case_linked.go
    - internal/rules/import_no_public_linked.go
    - internal/rules/import_no_weak_linked.go
    - internal/rules/message_pascal_case_linked.go
    - internal/rules/oneof_lower_snake_case_linked.go
    - internal/rules/package_lower_snake_case_linked.go
    - internal/rules/rpc_pascal_case_linked.go
    - internal/rules/service_pascal_case_linked.go
    - internal/rules/parity_basic_casing_test.go
  modified:
    - internal/rules/builder_linked.go
    - internal/rules/enum_pascal_case_test.go
    - internal/rules/enum_value_upper_snake_case_test.go
    - internal/rules/message_field_lower_snake_case_test.go
    - internal/rules/import_no_public_test.go
    - internal/rules/import_no_weak_test.go
    - internal/rules/message_pascal_case_test.go
    - internal/rules/oneof_lower_snake_case_test.go
    - internal/rules/package_lower_snake_case_test.go
    - internal/rules/rpc_pascal_case_test.go
    - internal/rules/service_pascal_case_test.go

key-decisions:
  - "Every ported casing rule keeps main's regexp byte-verbatim and declares it inside Validate (main's style) — casing.go helpers are excluded by prohibition and grep-enforced over the ten new files (zero hits for the three helper names)"
  - "Positions emit Decl.Pos (keyword/first token) everywhere main used Meta.Pos — NamePos/NumberPos are grep-excluded; the carried main literals (536/26/1 message keyword, 560/27/3 field type token, 300/13/3 rpc keyword) are the per-rule oracle and all survive verbatim"
  - "Walk shapes are main's shallow shapes via direct field iteration (info.File.Messages -> m.Fields/m.Oneofs/m.Enums, info.File.Enums/Services/Imports/Packages) — the recursion-capable protoast Visitor is not used anywhere in this batch; the shallow_walk (0 findings) and top_level_message_enum (2 findings) cells pin the exact boundary"
  - "Digit-led rpc edge (plan's rpc 2Fast) is unreachable in a compiling corpus — both parsers reject it at the rpc-name token (probe evidence in the test header); the RPC regexp semantics ride the shared get_thing finding plus carried tests"
  - "All five cells run cfg.Use = ten batch names + four MINIMAL names, so every cell re-verifies the whole ported registry cumulatively (14 rules)"

patterns-established:
  - "Casing-batch port shape: <Rule>Linked struct + LintRuleName() canonical override + Message() byte-verbatim from main + Validate(LinkedFileInfo) iterating info.File fields directly with main's regexp declared inline + core.AppendLinkedIssue(res, rule, <decl>.Pos, <name>)"
  - "Cumulative parity cell: one assertBatchParity call per cell with hand-derived old-side count, optional oldFindings() pre-run pinning exact (rule,line,col) literals on the old side before the both-engines compare"

requirements-completed: [LINT-PARITY-02]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Batch-2 parity suite TestLinkedParityBasicCasing: five cells proving old-vs-new sorted (rule, path, line, col) multiset parity for the ten BASIC-casing rules under cumulative 14-rule configs, with expected old-side counts (9/6/0/2/2) asserted before each compare"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_basic_casing_test.go#TestLinkedParityBasicCasing"
        status: pass
      - kind: other
        ref: "command: gsd_run check tdd-red-evidence .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-03-task1.json -> RED_EVIDENCE_OK (target test failed on the requirePorted guard, exit 1)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Casing-edge semantics pinned (T-04-05): main's regexps reject acronym enum ABC, leading-underscore value _LEADING, fields _field/oneTwo, single-char message A, digit oneof1, digit-first package foo1bar and acronym service AB while their well-named counterparts report nothing — casing.go helpers are excluded from all ten ports (grep zero hits)"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_basic_casing_test.go#TestLinkedParityBasicCasing/casing_edges"
        status: pass
      - kind: other
        ref: "command: grep -n 'toPascalCase|toUpperSnakeCase|toLowerSnakeCase|NamePos|NumberPos' internal/rules/*_linked.go (ten batch files) -> no hits"
        status: pass
    human_judgment: false
  - id: D3
    description: "Keyword positions pinned (T-04-06): the keyword_positions cell asserts six old-side (rule,line,col) literals (message/service col 1, rpc col 3 keyword, field col 3 type token, enum keyword col 1, enum value col 3) and the parity compare forces the new engine to reproduce them; a name-token port would shift columns and fail"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_basic_casing_test.go#TestLinkedParityBasicCasing/keyword_positions"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/message_pascal_case_test.go#TestMessagePascalCase_Validate"
        status: pass
    human_judgment: false
  - id: D4
    description: "Shallow walk shapes pinned (T-04-07): nested message/enum/oneof/field/map declarations report nothing (asserted zero = main's shape), while an enum inside a TOP-LEVEL message reports both the enum and its value (main's +1 level, 2 findings)"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_basic_casing_test.go#TestLinkedParityBasicCasing/shallow_walk"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_basic_casing_test.go#TestLinkedParityBasicCasing/top_level_message_enum"
        status: pass
    human_judgment: false
  - id: D5
    description: "Carried main tests: all ten rewritten <rule>_test.go suites green with main's expected core.Issue literals byte-verbatim (positions, source names, messages, rule names); zero old-contract test files remain for the batch's ten rules"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/message_pascal_case_test.go#TestMessagePascalCase_Validate"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/enum_pascal_case_test.go#TestEnumPascalCase_Validate"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/enum_value_upper_snake_case_test.go#TestEnumValueUpperSnakeCase_Validate"
        status: pass
      - kind: other
        ref: "command: go test -race -count=1 ./internal/rules/ -run 'TestEnumPascalCase|TestEnumValueUpperSnakeCase|TestFieldLowerSnakeCase|TestImportNoPublic|TestImportNoWeak|TestMessagePascalCase|TestOneofLowerSnakeCase|TestPackageLowerSnakeCase|TestRPCPascalCase|TestServicePascalCase' -> ok"
        status: pass
    human_judgment: false
  - id: D6
    description: "Batch gate (9 gates): quick suite -race green; cumulative TestLinkedParity* 13 green sub-cells (14/41 rules); suppression centrality + readSource-once green; build ok; tidy diff-empty; schema diff-empty; dormancy git-diff empty (ten old rule files + builder.go + engine files + casing.go + lint_helpers.go); golangci-lint 0 issues; TestModuleRuleSet green; NewLinked BASIC section filled in builder.go order"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "command: go test -race -count=1 ./internal/core/... ./internal/rules/... && go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity' && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && bin/golangci-lint run ./internal/rules/... ./internal/core/..."
        status: pass
    human_judgment: false

# Metrics
duration: 18 min
started: 2026-11-19T19:21:45Z
completed: 2026-11-19T19:39:50Z
status: complete
commits: 4
plan_head_before: 7fc5aae57d53c76402ca8a49d83aa475fce5814a
plan_head_after: 5ef30ab14e2bf1c50d2e1cae67297d3bb6ca1cc5
---

# Phase 04 Plan 03: Batch 2 — BASIC Casing Summary

**Ten BASIC-casing rules ported with main's regexps, keyword positions and shallow walks, parity-proven old-vs-new through five cumulative cells and main's carried literals byte-verbatim (14/41 rules ported)**

## Performance

- **Duration:** 18 min
- **Started:** 2026-11-19T19:21:45Z
- **Completed:** 2026-11-19T19:39:50Z
- **Tasks:** 3
- **Files modified:** 22 source/test files (+3 planning artifacts)

## Accomplishments

- The ten BASIC-casing rules are ported as `<Name>Linked` LintRules: main's regexp stays the predicate in every casing rule (declared inline in Validate, main's style) — acronym enums/services ("ABC", "AB"), digit names ("oneof1", "foo1bar"), leading underscores ("_field", "_LEADING") and single-char messages ("A") keep failing exactly as main fails them, where the lab's casing helpers would have passed several
- Positions emit Decl.Pos (the declaration keyword / first token) everywhere main used Meta.Pos: message/service col 1, rpc col 3, field at the type token, enum at its keyword, enum value at its name start, imports at the keyword; the lab's declared-name token is excluded and grep-enforced
- Walk shapes are main's shallow shapes through direct field iteration: top-level messages only (MessagePascalCase), plain fields and oneofs of top-level messages only (FieldLowerSnakeCase, OneofLowerSnakeCase), enums at top level plus one level deep (EnumPascalCase, EnumValueUpperSnakeCase); ImportNoPublic/ImportNoWeak port verbatim-from-lab with imp.Public/imp.Weak
- `TestLinkedParityBasicCasing` proves old-vs-new parity in five cells under cumulative 14-rule configs (ten batch names + four MINIMAL names): casing_edges (9 findings — the Pitfall-2 teeth), keyword_positions (six old-side position literals asserted before the compare), shallow_walk (asserted zero = main's shape), top_level_message_enum (2 findings — the exact +1-level boundary), import_flags (2 findings on resolving public/weak imports)
- All ten carried main test files rewritten onto the linked contract with expected literals byte-verbatim (536/26/1, 864/49/1, 610/31/3, 917/51/3, 667/33/5, 560/27/3, 81/6/1, 38/5/1, 748/39/3, 20/3/1, 300/13/3, 197/10/1); zero old-contract tests remain for the batch
- Batch gate green on nine gates (see Verification Log); registry now 14/41 with the BASIC-casing slots filled in builder.go's exact order; standing constraints (a) readSource-once and (b) suppression centrality held after every task

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): BASIC-casing parity suite fails on the unported-rule guard** - `1def52e` (test)
2. **Task 2 (GREEN part 1): port the ten casing rules — main regexps + keyword positions + shallow walks** - `cd7f5aa` (feat)
3. **Task 3 (GREEN part 2 + batch gate): carried main tests + cumulative gate** - `3065a7a` (feat)
4. **Fixture fix: keyword_positions literals corrected against the old-engine oracle** - `5ef30ab` (fix)

**Plan metadata:** docs commits (this SUMMARY + state/roadmap/validation updates; `0f7dda8` + the SUMMARY-correction commit).

## TDD Gate Compliance

- RED `1def52e` (test(04-03)) precedes GREEN `cd7f5aa`/`3065a7a` (feat(04-03)) in ancestry — gate sequence RED -> GREEN satisfied; no REFACTOR needed (the GREEN commits are the minimal implementations; no cleanup beyond the plan shape was warranted).
- RED evidence machine-validated (`gsd_run check tdd-red-evidence` -> `RED_EVIDENCE_OK`, reason `target_test_failed`): TestLinkedParityBasicCasing executed and failed on its own requirePorted assertion naming all ten unported rules while the package compiled; the semantic assessment is recorded in the evidence file.
- Task 1's RED commit touches only the `*_test.go` pathspec plus the evidence JSON planning artifact; Tasks 2-3 carry `tdd="true"`-cycle GREEN implementations with their behavior proof (parity suite + carried tests) run before each commit. The `5ef30ab` fix commit is a test-fixture correction that lands after GREEN — the RED -> GREEN order is unaffected (RED `1def52e` is an ancestor of both GREEN commits; 04-02's trailing fix-commit precedent).

## Files Created/Modified

- `internal/rules/{enum_pascal_case,enum_value_upper_snake_case,message_field_lower_snake_case,import_no_public,import_no_weak,message_pascal_case,oneof_lower_snake_case,package_lower_snake_case,rpc_pascal_case,service_pascal_case}_linked.go` - NEW: the ten ports (main regexp + Decl.Pos + shallow walks; imports verbatim-from-lab)
- `internal/rules/builder_linked.go` - ten BASIC-casing placeholders replaced in builder.go order; rename map updated (EnumFirstValueZero/EnumNoAllowAlias/ImportUsed/PackageSame* remain placeholders until 04-04)
- `internal/rules/parity_basic_casing_test.go` - NEW: five parity cells + basicCasingUse config + oldFindings/posKey helpers
- `internal/rules/{<ten rule>}_test.go` - rewritten onto the linked contract (parseLinkedFile), main's case sets and literals carried verbatim

## Decisions Made

- Regexps stay the predicates; the lab's casing-helper equality is documented as excluded in every ported file's header comment (reworded to avoid the helper identifiers so the acceptance grep is unambiguous) — T-04-05 mitigation is grep-verified
- `oldFindings` local helper runs the old engine alone so keyword_positions can pin exact (rule, line, col) literals on the old side before the both-engines compare — the harness in parity_harness_test.go stays untouched (plan file list respected)
- All cells use the cumulative 14-name config so each cell doubles as a full-registry regression (the plan's "cumulative re-verification" requirement read into every cell, not just one)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Digit-led rpc edge (rpc 2Fast) is unreachable in a compiling corpus**
- **Found during:** Task 1 (pre-test parse probe through both parsers)
- **Issue:** the plan's casing_edges cell specifies rpc `2Fast` as the RPC_PASCAL_CASE divergence edge, but BOTH go-protoparser and the easypc protoast parser reject a digit-led rpc name at the rpc-name token ("expected rpc method name"), so the fixture cannot compile — parity fixtures must compile (research precondition)
- **Fix:** the cell carries a compiling rpc edge instead (shared `get_thing` finding, regexp-rejected under both engines) and the probe evidence is documented in the test file header and the ported file's comment; main's RPCPascalCase regexp remains the pinned predicate (keyword_positions cell + carried test)
- **Files modified:** internal/rules/parity_basic_casing_test.go, internal/rules/rpc_pascal_case_linked.go (comment)
- **Verification:** casing_edges cell green with hand-derived count 9; rpc regexp semantics still pinned by TestRPCPascalCase literals (300/13/3) and keyword_positions (18:3)
- **Committed in:** 1def52e (cell) + cd7f5aa (comment)

**2. [Rule 1 - Bug] keyword_positions fixture had a leading blank line and a value that passes main's regexp**
- **Found during:** Task 2 (first green run — the OLD side failed the assertions)
- **Issue:** the fixture's raw string began with a newline (every expected line number off by one) and the enum value `BAD_V` actually matches main's UPPER_SNAKE regexp (only 5 old-side findings instead of 6) — the hand-derived literals did not match the old engine's real output
- **Fix:** renamed the value to `bad_value` (fails the regexp) and recomputed all six literals from the old engine's actual findings (6:1, 7:3, 13:1, 14:3, 17:1, 18:3) — the old engine is the oracle, so the test was fixed, not the code
- **Files modified:** internal/rules/parity_basic_casing_test.go
- **Verification:** keyword_positions cell green; all six old-side literals asserted and reproduced by the new engine; cumulative parity suite re-run green on the committed state
- **Committed in:** 5ef30ab

---

**Total deviations:** 2 auto-fixed (1 blocking fixture redesign, 1 fixture bug)
**Impact on plan:** Both were test-fixture corrections against the live old-engine oracle; no scope change, no behavior drift, old path byte-frozen throughout.

## Issues Encountered

None beyond the deviations above; every gate passed on its first run after the Task 2 fixture fix.

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components were introduced.

## Deferred Decisions (for 04-07 consolidation ledger)

- Lab behavior deliberately excluded: declared-name-token positions (the lab's NamePos/NumberPos) and casing-helper equality — main's keyword positions and regexps win; grep-verified exclusion over all ten ported files
- RPCPascalCase main-vs-lab divergence (digit-led names): unreachable in compiling corpora — both parsers reject digit-led rpc names at parse time, so no Phase-7 corpus can distinguish the two semantics on this rule; documented for the corpus gate
- No lab-improvement adoption occurred in this batch; the catalogued main-vs-lab divergences for later batches (ImportUsed resolution, swift_prefix bug, cross-file uniqueness, comment predicates) remain owned by plans 04-04..04-06

## User Setup Required

None - no external service configuration required. The etalon checkout precondition held: `/tmp/gsd-research/compiler-lab` at SHA `08cef10909c0722f7042ba17b5c11951f9c6b6f6` (verified at execution start).

## Verification Log

- `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityBasicCasing'` → RED: exit 1, `--- FAIL: TestLinkedParityBasicCasing` naming all ten unported rules — Task 1 verify; evidence record validated `RED_EVIDENCE_OK`
- `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityBasicCasing' -v` → ok (5 cells) — Task 2 verify (after the fixture fix)
- `go test -race -count=1 ./internal/core/ -run 'Suppression|ReadSourceOnce'` → ok (standing constraints a+b, after every task)
- `grep -n 'toPascalCase|toUpperSnakeCase|toLowerSnakeCase|NamePos|NumberPos'` over the ten ported files → no hits (T-04-05/T-04-06 grep mitigation)
- `git diff --exit-code` over the ten old rule files + builder.go + lint.go + lint_linked.go + lint_model.go + casing.go + lint_helpers.go → empty (dormancy)
- Carried tests: `go test -race -count=1 ./internal/rules/ -run 'TestEnumPascalCase|...|TestServicePascalCase'` → ok (all ten suites, -v per-suite verified)
- Batch gate (Task 3): quick suite `-race` ok (core, core/path_helpers, rules); `TestLinkedParity -v` → 13 green sub-cells (3 Wave-0 + 5 MINIMAL + 5 BASIC casing); `go build ./...` ok; `go mod tidy && git diff --exit-code go.mod go.sum` → TIDY_CLEAN; `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` → SCHEMA_CLEAN; `bin/golangci-lint run ./internal/rules/... ./internal/core/...` → 0 issues; `TestModuleRuleSet` → ok
- Commits measured: `git rev-list --count 7fc5aae..HEAD` = 5 (3 task commits + 1 fixture fix + 1 metadata docs commit); task-commits frontmatter count 4 excludes the docs commit per 04-02 precedent

## Next Phase Readiness

- Ready for 04-04 (batch 3 — BASIC semantic): the casing batch is parity-verified before the next batch starts (ROADMAP Phase-4 criterion 1); the registry holds 14/41; the remaining BASIC placeholders (EnumFirstValueZero, EnumNoAllowAlias, ImportUsed, PackageSame*) are exactly batch 3's scope
- Every later batch consumes the same harness surfaces; no harness changes were needed this batch; `parseProtoLinked` remains absent (re-add with callers if a batch ports the etalon's module_rules_test shape)
- No blockers; standing constraints (a) and (b) remain mechanized green cells

## Self-Check: PASSED

- All 22 task files exist on disk (11 created + 11 modified verified via git log --name-status)
- All 4 code commits are ancestors of HEAD: 1def52e (RED test), cd7f5aa (GREEN feat), 3065a7a (GREEN feat + gate), 5ef30ab (fixture fix)
- Commits measured from the persisted ledger: `git rev-list --count 7fc5aae..HEAD` = 5 incl. docs; 4 excluding the metadata docs commit
- RED evidence record exists and validates RED_EVIDENCE_OK
- STATE.md at Plan 4 of 7, stopped_at "Completed 04-03-PLAN.md"; ROADMAP.md phase row updated (3/7 summaries, In Progress)
- LINT-PARITY-02 correctly NOT marked complete (shared-ID gate: sibling plans 04-04..04-07 declare it without SUMMARYs)
- No WINDOWS.md entries: no stubs, no skipped tests, no unrun verifies

---
*Phase: 04-perenos-43-pravil-s-batch-paritetom*
*Completed: 2026-11-19*
