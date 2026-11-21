---
phase: 04-perenos-43-pravil-s-batch-paritetom
plan: 06
subsystem: rules
tags: [lint-rules, protoast, parity-harness, linked-engine, porting, dormancy, comments-batch, presence-predicate]

# Dependency graph
requires:
  - phase: 04-perenos-43-pravil-s-batch-paritetom (plans 01-05)
    provides: requirePorted/assertBatchParity/fingerprint harness, parseLinkedFile + validate() drivers, NewLinked registry (32/41), AppendLinkedIssue + LintRuleName seam, locked Decl.Pos position table
provides:
  - Seven COMMENTS ports (7 LintRules) with main's presence predicate: undocumented iff !HasCommentAbove() (Leading+Detached) — directive-only comments and detached comments document, inline trailing comments do not (39/41 cumulative)
  - Shallow main walk shapes pinned: top-level messages' plain fields/oneofs, top-level enums + one level (enums of top-level messages), services/RPCs; proto2 group fields never enter the loop (no guard needed — A4 resolved at source in both models)
  - TestLinkedParityComments — six cells pinning the batch's main-vs-lab divergences under cumulative 39-rule configs, incl. the pivotal directive-only interplay cell (suppression AND documentation coexist) (35 green parity sub-cells cumulative)
  - Seven carried test rewrites with byte-verbatim main literals (zero old-contract tests left for the batch)
  - builder_linked.go COMMENTS slots filled in builder.go order — registry covers MINIMAL+BASIC+DEFAULT+COMMENTS (39/41); only batch 6's two UNARY_RPC placeholders remain
affects: [04-07 (final batch + exact-set guard + deferred-decisions consolidation), Phase 5 flip (rename-map consolidation), Phase 7 corpus gate (suppression-heavy corpora note)]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 9732    # 38927 diff chars over the 16 changed task files at chars/4 (estimate was 68000; one uniform presence-predicate port x7 on a proven harness, probe-validated fixtures, no harness changes)
  tasks: 3
  commits: 3       # MEASURED: git rev-list --count 4fd382dc..HEAD (RED test + 2 GREEN); excludes the metadata docs commit

# Tech tracking
tech-stack:
  added: []
  patterns: [presence-predicate port shape (!HasCommentAbove() = Leading+Detached as the exact protoast image of main's len(Comments)==0 over go-protoparser ParseComments), interplay pin cell (a directive comment simultaneously documents for COMMENT_* verdicts and suppresses its own rule centrally — both behaviors asserted in one fixture pair), group-free field walks (both models list groups separately; no !field.Group guard)]

key-files:
  created:
    - internal/rules/parity_comments_test.go
    - internal/rules/comment_enum_linked.go
    - internal/rules/comment_enum_value_linked.go
    - internal/rules/comment_field_message_linked.go
    - internal/rules/comment_message_linked.go
    - internal/rules/comment_one_of_linked.go
    - internal/rules/comment_rpc_linked.go
    - internal/rules/comment_service_linked.go
  modified:
    - internal/rules/builder_linked.go
    - internal/rules/comment_enum_test.go
    - internal/rules/comment_enum_value_test.go
    - internal/rules/comment_field_message_test.go
    - internal/rules/comment_message_test.go
    - internal/rules/comment_one_of_test.go
    - internal/rules/comment_rpc_test.go
    - internal/rules/comment_service_test.go

key-decisions:
  - "Presence predicate is main's len(Comments)==0 mapped to !HasCommentAbove() (Leading+Detached): grounded at source in go-protoparser v4.14.2 — the parse loop attaches ALL preceding comments (including across blank lines) to element.Comments and keeps the same-line trailing comment in the separate InlineComment field, so directive-only comments and detached comments document exactly as in main; the lab's directive-filtering helper is never referenced (grep-enforced over the seven ports)"
  - "Proto2 group fields need NO guard: go-protoparser's unordered model keeps them in MessageBody.Groups and protoast keeps them in Message.Groups (both verified at source), so the plain-field loops of both engines never see them — the plan's proto2_group_field cell pins the zero-findings behavior (research A4 resolved at source)"
  - "The directive-only cell uses // easyp:disable MESSAGE_PASCAL_CASE (a DIFFERENT rule) so the central filter does not mask the COMMENT_* verdict; parseDirectivePrefix verified at source for both engines (easyp:disable -> action disable); the undirected sibling corpus (2 findings) proves the directive actually does central work — 0 vs 2 coexist in one fixture pair"
  - "Oneof member fields are deliberately undocumented in the presence_basics fixture: they sit outside BOTH engines' field walks (main MessageBody.Fields / port Message.Fields), so they never fire COMMENT_FIELD — the fixture proves the boundary instead of avoiding it"

patterns-established:
  - "Presence-predicate port: <Rule>Linked.Validate iterates main's shallow shapes via direct field iteration and issues at Decl.Pos iff !decl.HasCommentAbove() — comment content is never read, suppression stays central (standing constraint b)"
  - "Interplay pin cell: a directive-only comment corpus (0 findings) + its undirected sibling (2 findings) in one parity cell proves central suppression and documentation work together, killing the lab's directive-filtering predicate from both sides"

requirements-completed: [LINT-PARITY-02]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Batch-5 parity suite TestLinkedParityComments: six cells proving old-vs-new sorted (rule, path, line, col) multiset parity for the seven COMMENTS rules under cumulative 39-rule configs, with probe-confirmed expected old-side counts (8/0-2/0/3/2/0) asserted before each compare"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_comments_test.go#TestLinkedParityComments"
        status: pass
      - kind: other
        ref: "command: gsd_run check tdd-red-evidence .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-06-task1.json -> RED_EVIDENCE_OK (target test failed on the requirePorted guard, exit 1)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Directive-only interplay pinned (T-04-16, LINT-PARITY-02's motivating case): a declaration whose only comment is // easyp:disable MESSAGE_PASCAL_CASE is DOCUMENTED for COMMENT_MESSAGE (0 findings) while the same directive still suppresses MESSAGE_PASCAL_CASE centrally — the undirected sibling produces 2 findings (casing + comment), proving both behaviors coexist; a ported directive-filter would fabricate a COMMENT_MESSAGE finding and fail the parity compare"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_comments_test.go#TestLinkedParityComments/directive_only_documents"
        status: pass
      - kind: other
        ref: "command: grep -n 'leadingCommentDocuments|isLintDirective|buf:lint:ignore|nolint|easyp:' internal/rules/comment_*_linked.go -> no hits (GREP_CLEAN)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Presence predicate composition pinned: a blank-line-separated comment documents (detached_comment_documents, 0 findings — the lab's Leading-only helper would call it undocumented) while an inline trailing comment does not (trailing_only_undocumented, 3 findings at message/enum/value keyword positions — go-protoparser keeps same-line comments in the separate InlineComment field)"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_comments_test.go#TestLinkedParityComments/detached_comment_documents"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_comments_test.go#TestLinkedParityComments/trailing_only_undocumented"
        status: pass
    human_judgment: false
  - id: D4
    description: "Shallow walk shapes + proto2 groups pinned: undocumented message/enum/field/oneof children two levels deep report nothing while an enum inside a TOP-LEVEL message reports (main's +1 level, 2 findings); a proto2 group field and its body field produce zero COMMENT_FIELD findings (groups listed separately in both models — no guard needed, A4 at source)"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_comments_test.go#TestLinkedParityComments/shallow_shapes"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_comments_test.go#TestLinkedParityComments/proto2_group_field"
        status: pass
    human_judgment: false
  - id: D5
    description: "Carried main tests: all seven rewritten <rule>_test.go suites green with main's case sets and expected core.Issue literals byte-verbatim (COMMENT_ENUM 864/49/1 + 610/31/3, COMMENT_ENUM_VALUE 917/51/3 + 667/33/5, COMMENT_FIELD 447/18/3, COMMENT_MESSAGE 425/17/1, COMMENT_ONEOF 748/39/3, COMMENT_RPC 214/11/3, COMMENT_SERVICE 197/10/1); zero old-contract test files remain for the batch's seven rules"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/comment_enum_test.go#TestCommentEnum_Validate"
        status: pass
      - kind: other
        ref: "command: go test -race -count=1 ./internal/rules/ -run 'TestComment' -> ok (14 tests)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Batch gate green on ten gates (ROADMAP Phase-4 criterion 1 — verified before 04-07 starts): full core+rules -race suite; cumulative TestLinkedParity* (35 green sub-cells, 39/41 rules); TestComment carried suites; suppression centrality + readSource-once; TestModuleRuleSet; go build; tidy diff-empty; schema drift-empty; dormancy git-diff empty over the frozen old-path files; golangci-lint 0 issues; grep-clean on the directive-filtering helper and directive markers in all seven ports"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "command: go test -race -count=1 ./internal/core/... ./internal/rules/... && go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity' && go test -race -count=1 ./internal/rules/ -run 'TestComment' && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ && bin/golangci-lint run ./internal/rules/... ./internal/core/..."
        status: pass
    human_judgment: false

# Metrics
duration: 9 min
started: 2026-11-19T20:56:43Z
completed: 2026-11-19T21:05:34Z
status: complete
commits: 3
plan_head_before: 4fd382dcb58c879c61f39ea6d8ee1bf8de0af5f4
plan_head_after: 67d960f5a65e34d61ff2d65c8f102c74dabecf7c
---

# Phase 04 Plan 06: Batch 5 — COMMENTS Summary

**Seven COMMENT_* rules ported with main's presence predicate (!HasCommentAbove = Leading+Detached, grounded in go-protoparser ParseComments attachment), shallow walk shapes and the proto2-group no-guard behavior, parity-proven old-vs-new through six cumulative cells including the directive-only interplay pin (39/41 rules ported)**

## Performance

- **Duration:** 9 min
- **Started:** 2026-11-19T20:56:43Z
- **Completed:** 2026-11-19T21:05:34Z
- **Tasks:** 3
- **Files modified:** 16 source/test files (+1 planning artifact: RED evidence JSON; +VALIDATION.md task rows; +2 WINDOWS.md ledger entries)

## Accomplishments

- The seven COMMENTS rules are ported as `<Name>Linked` types with ONE uniform presence predicate: a declaration is undocumented iff `!HasCommentAbove()` (protoast Leading+Detached empty) — the exact image of main's `len(Comments)==0`, grounded at source in go-protoparser v4.14.2 (the parse loop attaches ALL preceding comments including across blank lines to `element.Comments`; same-line trailing comments land in the separate `InlineComment` field). Directive-only comments and detached comments therefore document exactly as in main; the lab's directive-filtering helper is never referenced (grep-enforced)
- Walk shapes are main's shallow shapes via direct field iteration: top-level messages' plain fields and oneofs, top-level enums plus one level (enums of top-level messages), services and their RPCs; positions at the declaration keyword (`Decl.Pos` ≡ main `Meta.Pos`); Message texts byte-verbatim
- Proto2 group behavior pinned with a compiling proto2 fixture: group fields and their body fields produce zero COMMENT_FIELD findings in BOTH engines — groups are listed separately from plain fields in both models (go-protoparser `MessageBody.Groups`, protoast `Message.Groups`), so no `!field.Group` guard is needed (research A4 resolved at source)
- `TestLinkedParityComments` proves old-vs-new parity in six cells under cumulative 39-rule configs; every corpus was probe-validated before the RED run (old-engine count oracle 8/0-2/0/3/2/0 + a zero-rules linked-engine link check proving no COMPILE routing)
- The pivotal LINT-PARITY-02 interplay cell: `// easyp:disable MESSAGE_PASCAL_CASE` (a DIFFERENT rule) both documents the declaration for COMMENT_MESSAGE and suppresses MESSAGE_PASCAL_CASE centrally — 0 findings — while the undirected sibling produces 2, proving the two behaviors coexist in one fixture pair
- All seven carried test files rewritten onto the linked contract with byte-verbatim literals (864/49/1 + 610/31/3, 917/51/3 + 667/33/5, 447/18/3, 425/17/1, 748/39/3, 214/11/3, 197/10/1); zero old-contract tests remain for the batch
- builder_linked.go COMMENTS slots filled in builder.go order; registry covers MINIMAL+BASIC+DEFAULT+COMMENTS — 39/41 rules; batch gate green on ten gates before 04-07 starts

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): COMMENTS parity suite fails on the unported-rule guard** - `cc2697b` (test)
2. **Task 2 (GREEN part 1): port the seven comment rules — presence predicate + shallow walks + builder fill** - `d94293f` (feat)
3. **Task 3 (GREEN part 2 + batch gate): carried main tests + cumulative gate** - `67d960f` (feat)

**Plan metadata:** docs commits (this SUMMARY + state/roadmap/validation updates).

## TDD Gate Compliance

- RED `cc2697b` (test(04-06)) precedes both GREEN commits `d94293f`/`67d960f` (feat(04-06)) in ancestry (merge-base verified: RED is an ancestor of the first GREEN); no REFACTOR needed — the GREEN commits are the minimal implementations and no cleanup beyond the plan shape was warranted.
- RED evidence machine-validated (`gsd_run check tdd-red-evidence` → `RED_EVIDENCE_OK`, reason `target_test_failed`): TestLinkedParityComments executed and failed on its own requirePorted assertion naming all seven unported rules while the package compiled; the semantic assessment is recorded in the evidence file.
- Task 1's RED commit touches only the `*_test.go` pathspec plus the evidence JSON planning artifact; Tasks 2-3 carry `tdd="true"`-cycle GREEN implementations with their behavior proof (parity suite + carried tests) run before each commit.
- The whole batch went green at Task 2 (builder registration is IN Task 2 per this plan's checker-mandated split — unlike 04-04/04-05 there was no mid-plan RED window).

## Files Created/Modified

- `internal/rules/{comment_enum,comment_enum_value,comment_field_message,comment_message,comment_one_of,comment_rpc,comment_service}_linked.go` - NEW: the seven LintRule ports (shared !HasCommentAbove() predicate, shallow walks, keyword positions, byte-verbatim Messages)
- `internal/rules/builder_linked.go` - seven COMMENTS placeholders replaced in builder.go order; rename map updated (39/41 filled; only UNARY_RPC placeholders remain for 04-07)
- `internal/rules/parity_comments_test.go` - NEW: six parity cells + commentsUse config (39 names cumulative) + probe-evidence header
- `internal/rules/{comment_enum,comment_enum_value,comment_field_message,comment_message,comment_one_of,comment_rpc,comment_service}_test.go` - rewritten onto the linked contract (parseLinkedFile), main's case sets and literals carried verbatim

## Decisions Made

- Presence predicate composition locked at source before any fixture was written: go-protoparser v4.14.2's parse loop (`parser/message.go`) assigns `comments := p.ParseComments()` to EVERY element unconditionally and same-line trailing comments go to the separate `InlineComment` field via `SetInlineComment` — so main's `len(Comments)==0` maps exactly to `!HasCommentAbove()` (Leading+Detached); the probe confirmed all four presence behaviors empirically before RED
- The directive-only cell uses `easyp:disable MESSAGE_PASCAL_CASE`: `parseDirectivePrefix` (verified at source, shared by both engines' central machinery) recognizes it as action "disable"; using a DIFFERENT rule than the COMMENT_* verdicts keeps the comment verdict unmasked by the central filter
- The proto2 group fixture relies on both models listing groups separately (verified at source in `interpretMessageBody` and `internal/protoast/build.go`); the probe confirmed zero old-side findings, closing research A4 without any guard in the port
- Oneof member fields stay undocumented in the presence_basics fixture (never fire COMMENT_FIELD in either engine) — the boundary is proven, not avoided

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] presence_basics fixture left DOC_ENUM_OTHER undocumented**
- **Found during:** Task 1 (fixture-design probe: old-engine count run)
- **Issue:** the hand-derived count assumed all of DocEnum's values documented, but only the first value carried the `// documented value` comment — the old engine (the oracle) produced 9 findings instead of the planned 8 (COMMENT_ENUM_VALUE co-fired on DOC_ENUM_OTHER)
- **Fix:** added the documenting comment to DOC_ENUM_OTHER (the plan's intent: the documented enum's values all documented); old-side count re-probed at 8
- **Files modified:** internal/rules/parity_comments_test.go (presenceFiles)
- **Verification:** probe re-run green: 8/0/2/0/3/2/0 per corpus; parity cell green after GREEN
- **Committed in:** cc2697b

**2. [Rule 1 - Bug] shallow_shapes fixture had the nested enum one level too shallow**
- **Found during:** Task 1 (fixture-design probe: old-engine count run)
- **Issue:** `NestedEnum` was written as a direct child of the TOP-LEVEL message OuterHolder (after NestedInner closed), so main's +1-level walk correctly reported it AND its value — old-side 4 instead of the planned 2, duplicating the toplevel_enum.proto case instead of isolating the two-level zero-findings boundary
- **Fix:** moved `NestedEnum` inside `NestedInner` so the cell exercises exactly the plan's spec: a NESTED message with undocumented message/enum/field/oneof children (two levels deep) reports NOTHING, while toplevel_enum.proto pins the +1-level boundary; old-side count re-probed at 2
- **Files modified:** internal/rules/parity_comments_test.go (shallowFiles)
- **Verification:** probe re-run green: 2 for the shallow corpus; parity cell green after GREEN
- **Committed in:** cc2697b

---

**Total deviations:** 2 auto-fixed (2 fixture bugs corrected against the live old-engine oracle)
**Impact on plan:** Both were test-fixture corrections in the pre-RED probe phase (the 04-03/04-05 precedent); the port code itself needed zero deviations. No scope change, no behavior drift, old path byte-frozen throughout.

## Issues Encountered

None beyond the two probe-caught fixture corrections above; every verify passed on its first run after the probes.

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components were introduced.

## Deferred Decisions (for 04-07 consolidation ledger)

- No lab-improvement behaviors were adopted or excluded in this batch beyond the catalogued presence-predicate divergence itself (the lab's directive-filtering helper stays unreachable — grep-enforced). The remaining catalogued divergences are already owned by earlier batch ledgers or need no decision here.
- The seven `<Name>Linked` type names and the builder rename-map entries collapse into canonical names at the Phase-5 consolidation (same as all prior batches).
- If a post-parity change ever wants the lab's stricter comment documentation (directive-only => undocumented), it is a deliberate behavior change to make at Phase 6/7, never silently — the interplay cell would need rewriting in the same commit.

## User Setup Required

None - no external service configuration required. The etalon checkout precondition held: `/tmp/gsd-research/compiler-lab` at SHA `08cef10909c0722f7042ba17b5c11951f9c6b6f6` (verified at execution start).

## Verification Log

- Fixture-design probes (TestZZProbeComments, deleted before commit): old-engine counts 8/0/2/0/3/2/0 probe-confirmed per corpus after the two fixture corrections (first probe exposed 9 and 4 — see Deviations); zero-rules linked runs empty on every corpus (no COMPILE routing; the proto2 group corpus links)
- `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityComments'` → RED: exit 1, `--- FAIL: TestLinkedParityComments` naming all seven unported rules — Task 1 verify; evidence record validated `RED_EVIDENCE_OK`
- `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityComments' -v` → ok (6 cells) — Task 2 verify, first run
- `go test -race -count=1 ./internal/core/ -run 'Suppression|ReadSourceOnce'` → ok (standing constraints a+b, after Task 2)
- Grep-clean: the directive-filtering helper identifier and directive markers over the seven ports → no hits; `HasCommentAbove` present in all seven (uniform predicate)
- Dormancy: `git diff --exit-code` over the seven old COMMENT rule files + builder.go + casing.go + lint_helpers.go + engine files → DORMANCY_CLEAN
- `go test -race -count=1 ./internal/rules/ -run 'TestComment'` → ok (14 tests, all literals verbatim) — Task 3 verify
- Batch gate (Task 3, on the committed state): full core+rules `-race` ok (core, core/path_helpers, rules); `TestLinkedParity -v` → 35 green sub-cells / 5 parents (3 Wave-0 + 5 MINIMAL + 5 BASIC casing + 8 BASIC semantic + 8 DEFAULT + 6 COMMENTS), 39/41 rules; `go build ./...` ok; `go mod tidy && git diff --exit-code go.mod go.sum` → TIDY_CLEAN; `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` → SCHEMA_CLEAN; `TestModuleRuleSet` → ok; `bin/golangci-lint run ./internal/rules/... ./internal/core/...` → 0 issues
- TDD gate: `git log -E --grep '^test\((0*4)-(0*6)\):'` → cc2697b; feat commits d94293f/67d960f; merge-base ancestry RED→GREEN verified
- Commits measured: `git rev-list --count 4fd382dc..HEAD` = 3 (excludes the metadata docs commit)

## Next Phase Readiness

- Ready for 04-07 (final batch — UNARY_RPC 2 rules + wiring finish): the COMMENTS group is fully ported and parity-verified before the next batch starts (ROADMAP Phase-4 criterion 1); the registry holds 39/41; the remaining placeholders are exactly batch 6's scope (RPCNoClientStreaming, RPCNoServerStreaming) plus the finish work (exact-set TestModuleRuleSet guard, lint_helpers re-alignment, lint-exclusion removal, deferred-decisions consolidation)
- The deferred-decisions list above carries this batch's parked items into the 04-07 consolidation ledger
- No harness changes were needed; no blockers; standing constraints (a) and (b) remain mechanized green cells (TestReadSourceOnce + the core Suppression cells + the batch's directive_only_documents interplay cell)

## Self-Check: PASSED

- All 16 task files exist on disk (8 created + 8 modified per git log --name-status of the three task commits; verified with `[ -f ]` below)
- All 3 code commits are ancestors of HEAD: cc2697b (RED test), d94293f (GREEN part 1), 67d960f (GREEN part 2 + gate)
- Commits measured from the persisted ledger: `git rev-list --count 4fd382dc..HEAD` = 3 (plan_head_before 4fd382dcb58c879c61f39ea6d8ee1bf8de0af5f4, plan_head_after 67d960f5a65e34d61ff2d65c8f102c74dabecf7c)
- RED evidence record exists and validates RED_EVIDENCE_OK
- Broken-windows ledger: the two documented fixture-correction deviations appended to .planning/WINDOWS.md as open entries 5/6; no stubs, no skipped tests, no unrun verifies

---
*Phase: 04-perenos-43-pravil-s-batch-paritetom*
*Completed: 2026-11-19*
