---
phase: 04-perenos-43-pravil-s-batch-paritetom
plan: 07
subsystem: rules
tags: [lint-rules, protoast, parity-harness, linked-engine, porting, dormancy, registry-parity, phase-closure]

# Dependency graph
requires:
  - phase: 04-perenos-43-pravil-s-batch-paritetom (plans 01-06)
    provides: requirePorted/assertBatchParity/fingerprint harness, parseLinkedFile + validate() drivers, NewLinked registry (39/41), AppendLinkedIssue + LintRuleName seam, firstExemptMismatches + fileOption helpers, position table, all batch parity suites
provides:
  - RPCNoClientStreamingLinked + RPCNoServerStreamingLinked verbatim-from-lab (rpc.Request.Stream / rpc.Response.Stream, rpc.Pos ≡ rpc.Meta.Pos) — registry complete 41/41 in builder.go order
  - TestRegistryParity — name-set/group-order/config-plumbing/ignoreOnly parity between New and NewLinked (T-04-18/T-04-19 mitigation)
  - TestModuleRuleSet strengthened to EXACT-SET equality (the D-03 ten; PackageNoImportCycle absent until Phase 6)
  - lint_helpers.go re-aligned to the linked contract (mismatches takes core.LintRule via core.AppendLinkedIssue; parser/meta import gone; header delta note records consumed vs D-01-kept sets)
  - .golangci.yml exclusion resolved in place: KEPT with the accurate Phase-4 comment + the Phase-5 removal obligation (supersedes PATTERNS.md's "drop the line")
  - TestLinkedParityFullRegistry — the Phase-4 closure cell: one broad compiling corpus (13 files / 9 packages), all 41 rules firing (coverage assertion), 73 old-side findings pinned before the multiset compare, old-vs-new identical
  - The consolidated deferred-decisions ledger (below) — Phase 5 inherits no silent expectations
affects: [Phase 5 flip (rename-map consolidation, helpers+exclusion deletion, buildCore wiring), Phase 6 (PackageNoImportCycle exact-set change), Phase 7 corpus gates (LINT-PARITY-01, lab.py)]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 9850    # 39396 diff chars over the 11 changed code files at chars/4 (estimate was 70000; two verbatim ports + guards on a proven harness; corpus test dominates)
  tasks: 3
  commits: 4      # MEASURED: git rev-list --count 8a06a11..HEAD (RED + GREEN + finish + closure cell); excludes the metadata docs commit

# Tech tracking
tech-stack:
  added: []
  patterns: [exact-set guard upgrade (ElementsMatch over the constructed full registry), registry-parity guard shape (name-set per use order + per-group expansion order + config plumbing + ignoreOnly unwrap, both constructors driven with identical Use lists), full-registry closure cell (coverage assertion: every rule fires at least once; flat cell, no subtests)]

key-files:
  created:
    - internal/rules/rpc_no_client_streaming_linked.go
    - internal/rules/rpc_no_server_streaming_linked.go
    - internal/rules/parity_unary_rpc_test.go
    - internal/rules/registry_parity_guard_test.go
    - internal/rules/parity_full_registry_test.go
  modified:
    - internal/rules/builder_linked.go
    - internal/rules/lint_helpers.go
    - internal/rules/module_rules_guard_test.go
    - internal/rules/rpc_no_client_streaming_test.go
    - internal/rules/rpc_no_server_streaming_test.go
    - .golangci.yml

key-decisions:
  - "mismatches is re-typed to mismatches(rule core.LintRule) calling core.AppendLinkedIssue(nil, rule, v.pos, v.source) — the 01-04 seam adaptation (5-arg AppendIssue, meta.Position conversion, parser/meta import) is GONE; the helper is lab-faithful again on the linked contract while staying report-everyone and unreachable from the ports (firstExemptMismatches is the ports' Finish)"
  - ".golangci.yml: the lint_helpers.go paths exclusion is KEPT with the accurate Phase-4 comment — consumed helpers (fileOption, groupedValues.add, firstExemptMismatches) are wired, the eight D-01-kept helpers stay verbatim-unused until Phase 5 deletes them together with the old path and the exclusion in one commit; this supersedes PATTERNS.md's drop-the-exclusion line and STATE.md's remove-when-callers condition (documented for the verifier)"
  - "TestModuleRuleSet upgrades from skip-unported membership probing to exact-set ElementsMatch over the constructed full registry — a construction failure is now a registry gap, not a port-in-flight skip; PackageNoImportCycle must stay absent until Phase 6"
  - "TestLinkedParityFullRegistry adds a coverage assertion (every one of the 41 rule names must appear in the old-side findings) on top of the pinned wantOldCount 73 — a silent rule is a coverage hole, not a pass"
  - "The broad corpus declares its rpc payload messages (documented, PascalCase): the linked engine fully resolves type references, so undeclared rpc types route the linked side to COMPILE while the syntax-only old engine reports rules — parity fixtures must compile under BOTH engines' linkers"

patterns-established:
  - "Registry-parity guard: drive New and NewLinked with IDENTICAL Use lists and compare outcomes — never re-derive validation (D-06 single-validator surface preserved)"
  - "Full-registry closure cell: coverage assertion + pinned count + one assertBatchParity over a corpus that compiles under the linked engine's full type resolution"

requirements-completed: [LINT-PARITY-02]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "UNARY_RPC ports: RPCNoClientStreamingLinked and RPCNoServerStreamingLinked verbatim-from-lab (stream predicate on rpc.Request.Stream / rpc.Response.Stream, finding at rpc.Pos ≡ main's rpc.Meta.Pos, byte-verbatim Messages) and the NewLinked registry completed to 41/41 in builder.go order (no placeholders remain)"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_unary_rpc_test.go#TestLinkedParityUnaryRPC"
        status: pass
      - kind: other
        ref: "command: gsd_run check tdd-red-evidence .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-07-task1.json -> RED_EVIDENCE_OK (target test failed on the requirePorted guard, exit 1)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Carried main tests for both streaming rules rewritten onto the linked contract via parseLinkedFile — main's case sets and the shared 300/13/3 rpc-keyword literal byte-verbatim"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/rpc_no_client_streaming_test.go#TestRPCNoClientStreaming"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/rpc_no_server_streaming_test.go#TestRPCNoServerStreaming"
        status: pass
    human_judgment: false
  - id: D3
    description: "Exact-set ModuleRule guard: the constructed full registry holds EXACTLY the ten D-03 cross-file rules as ModuleRules (ElementsMatch, no more no fewer; PackageNoImportCycle absent until Phase 6)"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/module_rules_guard_test.go#TestModuleRuleSet"
        status: pass
    human_judgment: false
  - id: D4
    description: "Registry-parity guard (T-04-18/T-04-19): NewLinked(Use: AllRuleNames()) name list == New's per use order and == AllRuleNames as a set (no _LINKED leak); all five groups MINIMAL/BASIC/DEFAULT/COMMENTS/UNARY_RPC expand to identical name lists in identical order; defaultIfEmpty plumbing (Root/Prefix/Suffix) matches for defaults and user-configured values; ignoreOnly unwrap identical for a group key + bare name sample with shape spot-checks"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/registry_parity_guard_test.go#TestRegistryParity"
        status: pass
    human_judgment: false
  - id: D5
    description: "lint_helpers.go re-aligned: mismatches(rule core.LintRule) via core.AppendLinkedIssue(nil, rule, v.pos, v.source); the parser/meta import is gone; header delta note records consumed (fileOption, groupedValues.add, firstExemptMismatches) vs D-01-kept unused (optionString, enumNumber, leadingCommentDocuments, isLintDirective, packagePosition, mismatches, packageHasVersionSuffix, versionNumber); no helper deleted; .golangci.yml exclusion kept with the accurate Phase-4 comment; golangci 0 issues on internal/rules"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "command: bin/golangci-lint run ./internal/rules/... -> 0 issues; go test -race -count=1 ./internal/rules/ -> ok (the re-typed helper compiles in every suite run)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Full-registry closure cell: one broad compiling corpus (13 files / 9 packages, every batch's edge classes at least once) under cfg.Use = rules.AllRuleNames(); coverage assertion requires all 41 rule names in the old-side findings; 73 old-side findings pinned before the multiset compare; old-vs-new sorted (rule, path, line, col) multisets identical under -race"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_full_registry_test.go#TestLinkedParityFullRegistry"
        status: pass
    human_judgment: false
  - id: D7
    description: "Phase closure gates green: full core+rules -race suite; cumulative TestLinkedParity* (47 PASS: 8 parents + 39 sub-cells, 41/41 rules); suppression centrality + TestReadSourceOnce; go build; tidy diff-empty; schema-gen drift-empty; golangci ./internal/... 0 issues; dormancy git-diff empty over 44 old-path files; easyp lint --help exit 0"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "command: go test -race -count=1 ./internal/core/... ./internal/rules/... && go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity' && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ && bin/golangci-lint run ./internal/rules/... ./internal/core/..."
        status: pass
    human_judgment: false
  - id: D8
    description: "Consolidated deferred-decisions ledger (## Consolidated Deferred-Decisions Ledger below) carries every parked post-parity item from all six prior batch SUMMARYs plus this plan's exclusion resolution — Phase 5/6/7 inherit no silent expectations"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "file: .planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-07-SUMMARY.md##Consolidated-Deferred-Decisions-Ledger"
        status: pass
    human_judgment: false

# Metrics
duration: 25 min
started: 2026-11-20T15:05:00Z
completed: 2026-11-20T15:30:00Z
status: complete
commits: 4
plan_head_before: 8a06a11fb6cc65fc1d751ecc936550d12b5c0f1c
plan_head_after: fda72c70b8f0efeb4c64bab95b0d71361482c0c3
---

# Phase 04 Plan 07: Batch 6 — UNARY_RPC + Phase Finish Summary

**Final two verbatim ports complete the 41/41 linked registry, provably name-parity with rules.New (names, groups, config plumbing, ignoreOnly), helpers re-aligned to the linked contract, the lint exclusion resolved with an honest comment, and all 41 rules parity-green on one broad corpus under the full phase closure gates**

## Performance

- **Duration:** 25 min (this continuation session: Task 2 verify/commit + Task 3 + close-out; Task 1's RED/GREEN cycle ran in the prior session on 2026-11-19 and was already committed)
- **Started:** 2026-11-20T15:05:00Z
- **Completed:** 2026-11-20T15:30:00Z
- **Tasks:** 3
- **Files modified:** 11 code files (+2 planning artifacts updated: 04-VALIDATION.md rows, WINDOWS.md ledger; +this SUMMARY)

## Accomplishments

- **Registry complete — 41/41.** `RPCNoClientStreamingLinked` and `RPCNoServerStreamingLinked` are verbatim-from-lab ports (research inventory #40/#41): the streaming predicate reads `rpc.Request.Stream` / `rpc.Response.Stream` (the etalon's `IsStream` renamed by main's protoast model) and issues at `rpc.Pos` (== main's `rpc.Meta.Pos` — the alias-only test diff the research verified). `builder_linked.go` fills the last two UNARY_RPC slots in builder.go order; the placeholder list is empty.
- **The registry is provably name-parity with the old one.** `TestRegistryParity` (NEW, `registry_parity_guard_test.go`) drives `New` and `NewLinked` with identical Use lists and requires: (a) name-set equality per use order — both registries return exactly `rules.AllRuleNames()`, no `_LINKED` suffix anywhere; (b) group parity — all five groups (MINIMAL/BASIC/DEFAULT/COMMENTS/UNARY_RPC) expand to identical name lists in identical order from both constructors; (c) `defaultIfEmpty` config plumbing — `PACKAGE_DIRECTORY_MATCH` Root/Prefix, `ENUM_ZERO_VALUE_SUFFIX` Suffix, `SERVICE_SUFFIX` Suffix match pairwise for defaults AND user-configured values; (d) `ignoreOnly` unwrap — identical maps for a sample with a group key, a bare name and a second group key, with shape spot-checks. Validation stays single-surface (D-06): both constructors share `validateConfigRuleNames`/`unwrapLintGroups`/`unwrapIgnoreOnly`.
- **TestModuleRuleSet is now exact-set.** The Wave-0 membership guard probes the CONSTRUCTED full registry (`NewLinked(Use: AllRuleNames())` must succeed — a construction failure is a registry gap, not a skip) and `require.ElementsMatch`s the ModuleRule names against the D-03 ten: no more, no fewer. `PACKAGE_NO_IMPORT_CYCLE` must stay absent until Phase 6.
- **lint_helpers.go is re-aligned to the new contract.** `mismatches` takes `core.LintRule` and calls `core.AppendLinkedIssue(nil, rule, v.pos, v.source)` — the 01-04 seam adaptation (5-arg `AppendIssue`, the `meta.Position` field conversion, the `parser/meta` import) is resolved and gone. The header delta note records the settled state: consumed by the ports (`fileOption`, `groupedValues.add`, `firstExemptMismatches`) vs D-01-kept verbatim and unused until Phase 5 (`optionString`, `enumNumber`, `leadingCommentDocuments`, `isLintDirective`, `packagePosition`, `mismatches`, `packageHasVersionSuffix`, `versionNumber`). No helper was deleted (D-01 whole-file vendoring for trivial lab re-syncs).
- **The lint-exclusion state is resolved honestly.** `.golangci.yml` keeps the `internal/rules/lint_helpers.go` paths exclusion with a rewritten comment: consumed helpers are wired; the eight D-01-kept helpers would trip the `unused` check; deleting them now would contradict D-01; the Phase-5 consolidation removes helpers + exclusion together. This supersedes PATTERNS.md's "drop the exclusion at the finishing batch" and STATE.md's "remove when helpers gain callers" (the condition is now partially true and the accurate comment is the honest state) — surfaced here for the verifier.
- **The Phase-4 closure cell is green.** `TestLinkedParityFullRegistry` (NEW, `parity_full_registry_test.go`) runs ONE broad compiling corpus — 13 files across 9 packages exercising every batch's edge classes (MINIMAL dir/package mismatches + packageless 0:0; casing teeth; octal "00" string compare; allow_alias; unused import; all seven PACKAGE_SAME_* on a differing second file; prefix/suffix/filename/uniqueness/standard-name/version/service-suffix teeth; all seven COMMENT_* with the directive-only interplay pin; both-streaming rpcs) — under `cfg.Use = rules.AllRuleNames()`. A coverage assertion requires every one of the 41 rule names to appear in the old-side findings; `wantOldCount` 73 is asserted before the multiset compare; old-vs-new fingerprints are identical under -race.
- **Phase closure gates all green** (see Verification Log): full core+rules suite under -race; cumulative `TestLinkedParity*` 47 PASS lines (8 parents + 39 sub-cells); suppression centrality + `TestReadSourceOnce`; build; `go mod tidy` diff-empty; `schema-gen` drift-empty; `bin/golangci-lint run ./internal/...` 0 issues; dormancy git-diff empty over 44 old-path files (engine files, builder.go, casing.go, lint_helpers.go, every pre-existing rule file); `easyp lint --help` exits 0 — the CLI surface is untouched.
- Standing constraints held at phase close: (a) `TestReadSourceOnce` green; (b) suppression centrality green — rules never self-suppress, the central-filter cells and the corpus-level directive pin coexist.

## Task Commits

Each task was committed atomically (Task 1 in the prior session, Tasks 2-3 in this session):

1. **Task 1 (RED): UNARY_RPC parity suite fails on the unported-rule guard** - `bf2e969` (test, 2026-11-19)
2. **Task 1 (GREEN): port #40/#41 verbatim + registry complete 41/41** - `c0903d7` (feat, 2026-11-19)
3. **Task 2: registry-parity + exact-set guards, helpers re-aligned, exclusion comment resolved** - `262b7da` (feat)
4. **Task 3: full-registry closure cell + validation rows + windows ledger** - `fda72c7` (test)

**Plan metadata:** docs commit (this SUMMARY + STATE/ROADMAP/REQUIREMENTS updates).

## TDD Gate Compliance

- Plan `type: tdd`. RED `bf2e969` (`test(04-07)`) precedes GREEN `c0903d7` (`feat(04-07)`) in ancestry — gate sequence RED -> GREEN satisfied; no REFACTOR needed (the GREEN commit is the minimal verbatim implementation; no cleanup beyond the plan shape was warranted).
- RED evidence machine-validated in the prior session (`gsd_run check tdd-red-evidence` -> `RED_EVIDENCE_OK`, reason `target_test_failed`, format `go-test`): `TestLinkedParityUnaryRPC` executed and failed on its own `requirePorted` assertion naming both unported rules while the package compiled. The record lives at `.planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-07-task1.json` (committed in `bf2e969`).
- Task 1's RED commit touches only `*_test.go` pathspecs plus the evidence JSON; the GREEN commit carries the ports + builder fill with its behavior proof (parity suite + carried tests) run before the commit.
- Tasks 2-3 are not `tdd="true"` tasks; their guards/cells are green-by-construction additions to the proven engine (04-01 Wave-0 precedent).

## Files Created/Modified

- `internal/rules/rpc_no_client_streaming_linked.go` - NEW: verbatim etalon port (Request.Stream predicate, rpc.Pos, canonical name override)
- `internal/rules/rpc_no_server_streaming_linked.go` - NEW: verbatim etalon port (Response.Stream predicate)
- `internal/rules/builder_linked.go` - the two UNARY_RPC placeholders replaced in builder.go order; rename map updated; 41/41 filled
- `internal/rules/parity_unary_rpc_test.go` - NEW: four parity cells (client_streaming 1, server_streaming 1, both_streaming 2, cumulative 4) under the full-registry config; old-side counts probe-confirmed before RED
- `internal/rules/rpc_no_client_streaming_test.go`, `rpc_no_server_streaming_test.go` - carried main tests rewritten onto the linked contract (parseLinkedFile), main's case sets and the 300/13/3 literal byte-verbatim
- `internal/rules/module_rules_guard_test.go` - strengthened to exact-set equality over the constructed full registry
- `internal/rules/registry_parity_guard_test.go` - NEW: TestRegistryParity (names per use order, groups, config plumbing, ignoreOnly)
- `internal/rules/lint_helpers.go` - mismatches retyped to core.LintRule via AppendLinkedIssue; parser/meta import removed; header delta note records the settled consumed/kept split
- `.golangci.yml` - exclusion comment rewritten in place (paths entry STAYS; Phase-5 removal obligation documented)
- `internal/rules/parity_full_registry_test.go` - NEW: the Phase-4 closure cell (13 files / 9 packages, 41-rule coverage assertion, 73 findings pinned)

## Decisions Made

- `mismatches` re-typing resolves the last 01-04 seam adaptation: the vendored helper file is now lab-faithful on the linked contract (one accepted adaptation left in the delta note: `packagePosition` takes the protoast file — still D-01-kept and unused). The ports never reach it (`firstExemptMismatches` is their Finish); the lab-faithful body stays for re-sync fidelity.
- The `.golangci.yml` exclusion is kept, not dropped: after the port, the eight D-01-kept helpers are still unused and would trip `unused`; deleting them would contradict D-01 whole-file vendoring. The comment records exactly which helpers are consumed, which stay, and the Phase-5 removal obligation. PATTERNS.md's drop-line instruction is superseded (documented here and in the comment; PATTERNS.md itself left untouched — it is the phase's pattern record, and the superseding decision lives in the code comment + this SUMMARY).
- `TestLinkedParityFullRegistry` adds a coverage assertion on top of the pinned count: with all 41 rules in the config, every rule must FIRE at least once — a rule that silently stays clean on the broad corpus is a corpus hole, not a pass. The count 73 was probe-derived from the live old engine during fixture development (the established oracle standard).
- The corpus declares its rpc payload messages (documented, PascalCase): the linked engine fully resolves type references while the old engine is syntax-only, so undeclared rpc types route the linked side to COMPILE and make parity vacuous — parity fixtures must compile under BOTH engines' linkers (recorded in the test header).
- Task 2's files were found in the working tree as uncommitted changes from the interrupted prior session (Task 1 committed, Task 2 complete-but-uncommitted). They were verified against the plan spec (all four cells present), proven by the Task-2 gates (guards green, golangci 0 issues, dormancy clean), and committed as `262b7da` rather than rewritten.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Task 3's `<files>` list is empty but the closure acceptance criterion requires a corpus test**
- **Found during:** Task 3 (plan parse)
- **Issue:** the plan's Task-3 action requires "one broad compiling fixture corpus ... driven through assertBatchParity with cfg.Use = rules.AllRuleNames()" and its acceptance criterion says "All 41 rules parity-green on the broad corpus under -race", but the task lists no files — the corpus must live in a test file
- **Fix:** `internal/rules/parity_full_registry_test.go` created as the corpus's home (`TestLinkedParityFullRegistry`, flat cell, no subtests)
- **Files modified:** internal/rules/parity_full_registry_test.go
- **Verification:** cell green under -race; coverage assertion + pinned count + parity compare all pass
- **Committed in:** fda72c7

**2. [Rule 1 - Bug] Corpus fixture: missing directive + unqualified import references read all three imports unused**
- **Found during:** Task 3 (first probe run of the closure cell)
- **Issue:** `dir_only_msg` was drafted WITHOUT its `// easyp:disable MESSAGE_PASCAL_CASE` directive (so MESSAGE_PASCAL_CASE fired instead of being centrally suppressed — the interplay pin was absent), and `UsesDep` referenced `Outer`/`WDepMsg` unqualified — `InstructionParser` parses references against the SOURCE package, so the public/weak imports (packages dep.v1/wdep.v1) could never match and all three imports reported IMPORT_USED
- **Fix:** the directive added above `dir_only_msg` (documenting it for COMMENT_MESSAGE while MESSAGE_PASCAL_CASE stays centrally suppressed); references qualified (`dep.v1.Outer`, `wdep.v1.WDepMsg`) so the public/weak imports mark used and only the truly-unused `unused_dep.proto` import reports
- **Files modified:** internal/rules/parity_full_registry_test.go
- **Verification:** closure cell green; old-side count settled at 73
- **Committed in:** fda72c7

**3. [Rule 1 - Bug] Corpus fixture: undeclared rpc payload types routed the linked side to COMPILE**
- **Found during:** Task 3 (first parity compare — linked side returned 8 COMPILE findings)
- **Issue:** the rpcs referenced GetReq/GetResp/DoStreamRequest/DoStreamResponse/DupRequest/DupResponse/DupAgainResponse without declaring them; the old engine is syntax-only and never resolves rpc types, but the linked engine fully resolves references through the easypc linker — undefined types are compilation errors routed to COMPILE, making the parity compare vacuous
- **Fix:** all seven payload messages declared (PascalCase, each with its own documenting comment — trailing and block-only comments do not document, the 04-06 attachment semantics) so both engines see a compiling corpus and the rule findings are the only divergence candidate
- **Files modified:** internal/rules/parity_full_registry_test.go
- **Verification:** COMPILE findings gone; 73 old-side findings reproduced identically by the linked engine
- **Committed in:** fda72c7

---

**Total deviations:** 3 auto-fixed (1 plan-file-list gap, 2 corpus fixture corrections against the live old-engine oracle)
**Impact on plan:** All three were necessary for the closure cell to be non-vacuous; no scope change, no behavior drift, old path byte-frozen throughout.

## Issues Encountered

- The prior session ended after Task 1 with Task 2 complete-but-uncommitted in the working tree. Handled as a continuation: the four Task-2 files were inspected against the plan spec (all cells present and correct), proven by the Task-2 verification gates (TestModuleRuleSet + TestRegistryParity green under -race; `bin/golangci-lint run ./internal/rules/...` 0 issues), and committed as `262b7da`. No work was redone.
- Unrelated dirty files were present in the working tree (`.agents/skills/sdd/scripts/pipeline.sh`, `.planning/phases/03-.../03-VERIFICATION.md`, untracked runtime dirs). Out of scope per the deviation scope boundary — left untouched, not committed.

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components were introduced. The D-01-kept helpers are verbatim-unused BY DESIGN (dormancy; Phase 5 deletes them) and are recorded in the .golangci.yml comment, not stubs.

## Consolidated Deferred-Decisions Ledger

Every parked post-parity item from the six prior batch SUMMARYs (04-01..04-06) plus this plan's own resolution, consolidated so **Phase 5/6/7 inherit no silent expectations** (research Open Question 2). Each entry: what shipped, what was deliberately excluded, and where the decision is owned.

**Main-vs-lab behavior divergences — main wins, lab variant is the deferred decision (LINT-PARITY-02):**

1. **Resolution-based ImportUsed** (04-04): the lab's `validateResolved` over `Resolved.UsedImports` is excluded; main's stateless syntax walk (`importUsage` + `InstructionParser`, `existInProtoLinked` over top-level imported names) shipped, grep-enforced (zero `Resolved` references in `import_used_linked.go`). The qualified-nested-reference parity cell documents exactly where resolution semantics would diverge (resolution marks it used; main reports the import unused). **Owner:** a deliberate Phase 6/7 behavior change, never silent — the cell would need rewriting in the same commit.
2. **Cross-file RPCRequestResponseUnique** (04-04 pending, 04-05 resolved): the lab's cross-file/qualified/report-everyone ModuleRule over `Resolved.RPCTypes` is excluded; main's per-file first-exempt uniqueness on written type strings shipped (Collect/Finish, insertion-ordered path groups, per-file seen set), grep-enforced (`Resolved|RPCTypes` never read). **Owner:** Phase 6/7 deliberate change.
3. **swift_prefix fix** (04-04): `PackageSameSwiftPrefixLinked` compares the literal `ruby_package` option — main's copy-paste bug, pinned bug-for-bug from both sides (parity cell + carried regression cells + documented type header). The corrected swift_prefix behavior would break Phase-7 corpus parity if applied now. **Owner:** post-parity (Phase 6/7) fix decision, made consciously with the pins rewritten in the same commit.
4. **NamePos positions** (04-03): the lab's declared-name-token positions are excluded; main's `Decl.Pos` keyword positions shipped everywhere, grep-enforced (`NamePos|NumberPos` zero hits over the ten casing ports). **Owner:** none — main wins permanently unless a Phase 6/7 change deliberately reopens it.
5. **Helper-casing adoption** (04-03/04-05): the lab's casing-helper equality (acronym-aware upper-snake, required trailing underscore) and its helpers (`casing.go`) are excluded; main's byte-verbatim regexps and the carried `pascalToUpperSnake` algorithm (as `pascalToUpperSnakeLinked`, rename-mapped) shipped. **Owner:** Phase 5 consolidation deletes `casing.go` together with the old path; the Linked-suffixed carried helper collapses to the canonical name there.
6. **Lab report-everyone `groupedValues.mismatches`** (04-02): untouched, byte-verbatim, and unreachable from all ports (`firstExemptMismatches` — main's first-exempt Finish — is the group ports' mechanism; grep-enforced). **Owner:** Phase 5 deletes it with the old path; if a future rule wants report-everyone semantics it is a deliberate behavior change.
7. **Packageless 1:1 fallback** (04-02): `packagePosition`'s packageless-file fallback (empty value at `protoast.FilePosition` 1:1) is excluded — main iterates Packages only; the `packageless_invisible` cell fails parity if ever ported. **Owner:** none (main wins).
8. **protoast.FilePosition for PACKAGE_DEFINED** (04-02): main's zero Position literal (`Filename`, 0:0) wins (T-04-04), pinned in the parity fingerprint and the carried test; the lab's 1:1 `FilePosition` is excluded. **Owner:** none (main wins).
9. **Standard-name normalization** (04-05): the lab's last-dot-segment pascal-normalized compare (accepts qualified types) is excluded; main's exact written-string compares shipped (qualified types fail), grep-enforced. **Owner:** Phase 6/7 deliberate change.
10. **Directive-filtering comment helper** (04-06): the lab's `isLintDirective`/`leadingCommentDocuments` predicate is unreachable from the seven COMMENT_* ports (grep-enforced) — main's `len(Comments)==0` -> `!HasCommentAbove()` presence predicate shipped, so directive-only comments DOCUMENT. If a post-parity change ever wants the lab's stricter semantics (directive-only => undocumented), it is a deliberate Phase 6/7 behavior change requiring the `directive_only_documents` interplay cell to be rewritten in the same commit.
11. **Empty-enum guard** (04-04/04-05): the ONE documented crash-fix deviation (research A7): the lab's `len(Values)==0` guard is kept in EnumFirstValueZeroLinked/EnumZeroValueSuffixLinked (main panics; protoc rejects empty enums so the engine path never sees one), unit-pinned since 04-04; main's string compare (`Number != "0"`) retained — "00"/"0x0" issue where the lab's numeric compare passes. **Owner:** conscious deviation, documented in the type header; revisit only with a compiler-level rationale.
12. **RPCPascalCase digit-led names** (04-03): the main-vs-lab divergence edge is UNREACHABLE — both parsers reject digit-led rpc names at the rpc-name token, so no compiling corpus (and no Phase-7 lab.py corpus) can distinguish the semantics on this rule. **Owner:** documented for the Phase-7 corpus gate; no action possible.
13. **Adjacent option string literals** (04-04, observation-only): protoast merges `"a" "b"` to `"ab"`; main's `Constant` basis for that edge was never probed, so it stays an observation note, not a parity cell. **Owner:** Phase-7 corpus runs may surface it; treat as a finding then.

**Phase-5 obligations (mechanical, no behavior change):**

14. **Old-path consolidation:** Phase 5 deletes the old rule files, `builder.go`, `casing.go`, the eight D-01-kept helpers and the `.golangci.yml` paths exclusion TOGETHER in one commit (the exclusion comment records this obligation); the rename map in `builder_linked.go`'s header (`pascalToUpperSnake -> pascalToUpperSnakeLinked`, `existInProto -> existInProtoLinked`, `AppendIssue -> AppendLinkedIssue`, `GetRuleName -> LinkedRuleName` + override, locally-declared regexps -> package-level if desired) is the consolidation checklist.
15. **buildCore/Options.LinkedRules production wiring:** deliberately deferred to Phase 5 (research Open Question 3 resolution, restated 04-01): `LinkedRules` is populated only from tests; dormancy is trivially provable and `easyp lint` byte-identical. Phase 5's flip plan owns the wiring + the COMPAT gate.
16. **parseProtoLinked:** removed in 04-02 (no callers, lint gate); re-add WITH callers only if a batch ports the etalon's `module_rules_test` shape — no batch did.
17. **PACKAGE_NO_IMPORT_CYCLE:** the Phase-3 parity-guard pin stays byte-stable; the exact-set `TestModuleRuleSet` asserts its ABSENCE; Phase 6 adds it as a ModuleRule and updates the guard's D-03 set in the same commit.
18. **wazero data race:** pre-existing in `internal/adapters/plugin` + `internal/generation` (-race), Phase-1 deferred item, VERIFIER-corrected scope (Phase 3) — owned before the Phase-7 full-suite gate; explicitly out of scope for every Phase-4 batch (this plan's scoped suite excludes both packages).

## User Setup Required

None - no external service configuration required. The etalon checkout precondition held through the phase: `/tmp/gsd-research/compiler-lab` at SHA `08cef10909c0722f7042ba17b5c11951f9c6b6f6` (the verbatim porting source for #40/#41; verified by the prior session's commits).

## Verification Log

- Task 2 gates (on the inherited working tree): `go test -race -count=1 ./internal/rules/ -run 'TestModuleRuleSet|TestRegistryParity' -v` -> ok (both guards green); `bin/golangci-lint run ./internal/rules/...` -> 0 issues; committed `262b7da`
- Closure-cell probe iterations (pre-commit): first run -> coverage assertion PASSED, count 77 (stale fixtures: no directive, all imports unused); after directive + qualified references -> count 73, parity compare failed on 8 linked-side COMPILE findings; after declaring the seven rpc payload messages -> count 73, parity compare green
- `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityFullRegistry' -v` -> ok — Task 3 verify
- Phase gate (Task 3, on the committed state):
  - `go test -race -count=1 ./internal/core/... ./internal/rules/...` -> ok (core, core/path_helpers, rules)
  - `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity'` -> ok; -v: 47 PASS lines (8 parents: Wave0, Minimal, BasicCasing, BasicSemantic, Default, Comments, UnaryRPC, FullRegistry + 39 sub-cells; the closure cell is flat) — 41/41 rules cumulative
  - `go test -race -count=1 ./internal/core/ -run 'Suppression|ReadSourceOnce'` -> ok (TestReadSourceOnce + 3 suppression cells — standing constraints a+b at phase close)
  - `go build ./...` -> BUILD_OK; `go mod tidy && git diff --exit-code go.mod go.sum` -> TIDY_CLEAN
  - `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` -> SCHEMA_CLEAN (rule names unchanged — drift impossible, proven)
  - `bin/golangci-lint run ./internal/...` -> 0 issues
  - Dormancy audit: `git diff --exit-code -- internal/core/lint.go internal/core/lint_linked.go internal/core/lint_linked_read.go internal/core/lint_suppressions.go internal/core/link.go internal/rules/builder.go internal/rules/casing.go internal/rules/lint_helpers.go <all 38 pre-existing rule files>` -> DORMANCY_CLEAN (44 old-path files byte-frozen; lint_model.go carries only the Wave-0 seam additions asserted since 04-01)
  - `go build -o /tmp/easyp-check ./cmd/easyp && /tmp/easyp-check lint --help` -> exit 0 (CLI surface untouched)
- Commits measured from the persisted ledger: `git rev-list --count 8a06a11..HEAD` = 4 task commits (excludes the metadata docs commit)

## Next Phase Readiness

**Phase 4 is complete (7/7 plans).** Phase 5 (the flip) inherits:

- A fully ported, parity-proven, dormant rule set: 41/41 rules in `NewLinked` in builder.go order, name-parity-proven against `rules.New` (names, groups, config plumbing, ignoreOnly), the exact D-03 ModuleRule set, and byte-frozen old path + engine files
- The parity harness (requirePorted / assertBatchParity / fingerprint / oldFindings), the position table, the suppression + readSource-once cells, and the full-registry closure cell as the regression baseline
- The consolidated deferred-decisions ledger above — every main-vs-lab divergence pinned, every Phase-5 mechanical obligation enumerated (old-path + helpers + exclusion deletion, rename-map collapse, buildCore wiring, COMPAT gate)
- No blockers. The wazero race remains the known pre-Phase-7 item (not a Phase-5 blocker).

Ready for `/gsd:verify-work` on Phase 4; then `/gsd:plan-phase 5` for the flip.

## Self-Check: PASSED

- All 11 task files exist on disk (5 created + 6 modified; verified via `git log --name-status 8a06a11..HEAD` and the per-commit stats)
- All 4 task commits are ancestors of HEAD: bf2e969 (RED test), c0903d7 (GREEN feat), 262b7da (finish feat), fda72c7 (closure test)
- Commits measured from the persisted ledger: `git rev-list --count 8a06a11..HEAD` = 4 (plan_head_before 8a06a11fb6cc65fc1d751ecc936550d12b5c0f1c, plan_head_after fda72c70b8f0efeb4c64bab95b0d71361482c0c3)
- RED evidence record exists (committed in bf2e969) and was machine-validated RED_EVIDENCE_OK in the prior session
- WINDOWS.md ledger: three deviation entries appended (ids 7-9, open); no stubs, no skipped tests, no unrun verifies
- LINT-PARITY-02 marked complete via the shared-ID gate: 04-07 is the LAST declaring plan in the phase and all six sibling SUMMARYs exist

---
*Phase: 04-perenos-43-pravil-s-batch-paritetom*
*Completed: 2026-11-20*
