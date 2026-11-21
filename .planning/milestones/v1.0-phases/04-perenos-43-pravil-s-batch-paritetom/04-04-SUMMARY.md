---
phase: 04-perenos-43-pravil-s-batch-paritetom
plan: 04
subsystem: rules
tags: [lint-rules, protoast, parity-harness, linked-engine, porting, dormancy, bug-pin]

# Dependency graph
requires:
  - phase: 04-perenos-43-pravil-s-batch-paritetom (plans 01-03)
    provides: requirePorted/assertBatchParity/fingerprint harness, parseLinkedFile + validate() drivers, NewLinked registry (14/41), AppendLinkedIssue + LintRuleName seam, locked Decl.Pos position table, firstExemptMismatches, fileOption helper
provides:
  - Ten BASIC-semantic ports (EnumFirstValueZero, EnumNoAllowAlias, ImportUsed + the seven PackageSame* option-comparison ModuleRules) — main's predicates byte-faithful: string compare ("00"/"0x0" issue), any allow_alias option at the enum keyword, stateless syntax-walk ImportUsed, raw written constants, skip-optionless, first-exempt (24/41 cumulative)
  - swift_prefix bug-for-bug pin (compares the literal ruby_package option; T-04-09) pinned from both sides — parity cell + carried regression cells + documented type header with the deferred post-parity fix
  - EnumFirstValueZero empty-enum guard kept as the ONE documented deviation (research A7, T-04-10), unit-pinned
  - TestLinkedParityBasicSemantic — eight cells pinning the batch's main-vs-lab divergences under cumulative 24-rule configs (21 green parity sub-cells cumulative)
  - Ten carried test rewrites with byte-verbatim expected literals (zero old-contract tests left for the batch; the four formerly-nil-verdict option tests upgraded to concrete findings)
affects: [04-05..04-06 batches (harness + registry reuse), 04-07 (exact-set guard, deferred-decisions consolidation incl. the swift_prefix fix and resolution-based ImportUsed), Phase 5 flip (rename-map consolidation), Phase 7 corpus gate (swift_prefix/ruby_package corpus note)]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 19774    # 79097 diff chars over the 23 changed files at chars/4 (estimate was 88000; one uniform ModuleRule shape x7 on a proven harness + probe-validated fixtures)
  tasks: 3
  commits: 3       # MEASURED: git rev-list --count 246c750..HEAD (RED test + 2 GREEN); excludes the metadata docs commit

# Tech tracking
tech-stack:
  added: []
  patterns: [option-comparison ModuleRule shape (fileOption + raw opt.Value + firstExemptMismatches, skip-optionless, packageless early return), stateless importUsage extraction with InstructionParser reuse (existInProtoLinked over top-level imported names), bug-for-bug pin pattern (wrong main literal ported + both-side regression pins + deferred-fix header), probe-validated fixture design (old-engine count oracle + zero-rules link check before RED)]

key-files:
  created:
    - internal/rules/enum_first_value_zero_linked.go
    - internal/rules/enum_no_allow_alias_linked.go
    - internal/rules/import_used_linked.go
    - internal/rules/package_same_csharp_namespace_linked.go
    - internal/rules/package_same_go_package_linked.go
    - internal/rules/package_same_java_multiple_files_linked.go
    - internal/rules/package_same_java_package_linked.go
    - internal/rules/package_same_php_namespace_linked.go
    - internal/rules/package_same_ruby_package_linked.go
    - internal/rules/package_same_swift_prefix_linked.go
    - internal/rules/parity_basic_semantic_test.go
  modified:
    - internal/rules/builder_linked.go
    - internal/rules/enum_first_value_zero_test.go
    - internal/rules/enum_no_allow_alias_test.go
    - internal/rules/import_used_test.go
    - internal/rules/package_same_csharp_namespace_test.go
    - internal/rules/package_same_go_package_test.go
    - internal/rules/package_same_java_multiple_files_test.go
    - internal/rules/package_same_java_package_test.go
    - internal/rules/package_same_php_namespace_test.go
    - internal/rules/package_same_ruby_package_test.go
    - internal/rules/package_same_swift_prefix_test.go

key-decisions:
  - "swift_prefix bug-for-bug pin implemented: PackageSameSwiftPrefixLinked reads the literal ruby_package option (main's copy-paste bug); fixing it is forbidden here and ledgered as a deferred post-parity decision — pinned by the swift_prefix_bug_pin parity cell (2 then 0) and carried regression cells (differing ruby_package issues at 5:1 with the raw constant; differing swift_prefix stays clean)"
  - "ImportUsed ports main's syntax walk UNCONDITIONALLY and statelessly (etalon importUsage shape + InstructionParser reuse): LinkedFileInfo.Resolved never read (grep-clean), existInProtoLinked keeps main's top-level-only imported-file reading, and the walk covers services (rpc types/options), recursive messages (field types/options, oneof fields), file-level extends and file-level options — maps/groups/message-level extends/message-level options are outside main's walk (unordered keeps them in MessageBody; main's checkInExtends receives ProtoBody.Extends only)"
  - "The real testdata/import_used corpus does NOT link (undefined extend target import_used.thrd_party.MethodOptions, undefined MethodOption, undeclared extension number, duplicate TestService) — engine-level IMPORT_USED cells use compiling same-package variants; the real fixtures stay in the carried test driving the linked rule directly over parseLinkedFile infos (all 11 imports marked used, clean)"
  - "EnumFirstValueZero keeps the lab's len(Values)==0 guard as the single documented crash-fix deviation (main panics on empty enums; protoc rejects empty enums so the engine path never sees one), unit-pinned by TestEnumFirstValueZero_EmptyEnumGuard, while the string compare (Number != \"0\") stays main's — \"00\" and \"0x0\" issue where the lab's numeric compare passes"
  - "Every parity cell runs cfg.Use = ten batch-3 names + fourteen already-ported names (cumulative 24 rules), so each cell doubles as a full-registry regression; cell fixtures were probe-validated before RED (old-engine count oracle + zero-rules link check proving every corpus links)"

patterns-established:
  - "Option-comparison ModuleRule port shape: files groupedValues; Collect early-returns on packageless files, uses fileOption (last-match), returns nil when the option is absent, adds groupedValue{path, value: opt.Value, source: opt.Value, pos: opt.Pos}; Finish delegates to groupedValues.firstExemptMismatches"
  - "Bug-pin pattern: the wrong main literal is ported verbatim, the type header documents the bug + pin decision + deferred fix, and both-side regression cells (bug fires / fixed-behavior stays silent) make any silent 'fix' fail the suite"

requirements-completed: [LINT-PARITY-02]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Batch-3 parity suite TestLinkedParityBasicSemantic: eight cells proving old-vs-new sorted (rule, path, line, col) multiset parity for the ten BASIC-semantic rules under cumulative 24-rule configs, with probe-confirmed expected old-side counts (3/2/1/1/7/1/2-0/2) asserted before each compare"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_basic_semantic_test.go#TestLinkedParityBasicSemantic"
        status: pass
      - kind: other
        ref: "command: gsd_run check tdd-red-evidence .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-04-task1.json -> RED_EVIDENCE_OK (target test failed on the requirePorted guard, exit 1)"
        status: pass
    human_judgment: false
  - id: D2
    description: "EnumFirstValueZero string-compare pinned (T-04-10): first value \"00\" (octal zero) and \"0x0\" (hex zero) issue where the lab's numeric parse-and-compare passes them, literal \"0\" is clean, second value never checked, top-level-message enums behave the same (+1 level), nested-message enums report nothing; the empty-enum guard is unit-pinned as the one documented deviation"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_basic_semantic_test.go#TestLinkedParityBasicSemantic/first_value_string_compare"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/enum_first_value_zero_test.go#TestEnumFirstValueZero_EmptyEnumGuard"
        status: pass
    human_judgment: false
  - id: D3
    description: "EnumNoAllowAlias any-option semantics pinned: allow_alias = false issues at the ENUM keyword position with the enum name as source (the lab's Value==true-at-option-position would report neither); optionless enums clean"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_basic_semantic_test.go#TestLinkedParityBasicSemantic/allow_alias_any_value"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/enum_no_allow_alias_test.go#TestEnumNoAllowAlias_Validate"
        status: pass
    human_judgment: false
  - id: D4
    description: "ImportUsed syntax-walk verdicts pinned (T-04-08): field-type and rpc request/response references to imported top-level names mark the import used; an unused import issues at the import keyword with the quoted Location; a qualified reference into an imported NESTED message (which resolution would mark used) stays unused and issues — Resolved is never read (grep-clean)"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_basic_semantic_test.go#TestLinkedParityBasicSemantic/import_used_verdicts"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_basic_semantic_test.go#TestLinkedParityBasicSemantic/syntax_walk_only"
        status: pass
      - kind: other
        ref: "command: grep -n '\\.Resolved\\|validateResolved' internal/rules/import_used_linked.go (code lines) -> no hits"
        status: pass
    human_judgment: false
  - id: D5
    description: "Seven option-comparison ModuleRules pinned (T-04-11): raw written constants compared (\"x\" vs 'x' diverges; unquoting would pass), optionless files invisible, first value per package exempt, packageless files ignored; PackageSameSwiftPrefix compares ruby_package — bug-for-bug pin from both sides"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_basic_semantic_test.go#TestLinkedParityBasicSemantic/option_groups_first_exempt"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_basic_semantic_test.go#TestLinkedParityBasicSemantic/option_written_compare"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_basic_semantic_test.go#TestLinkedParityBasicSemantic/swift_prefix_bug_pin"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_basic_semantic_test.go#TestLinkedParityBasicSemantic/packageless_files_ignored"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/package_same_swift_prefix_test.go#TestPackageSameSwiftPrefix_RubyPackageBugPin"
        status: pass
    human_judgment: false
  - id: D6
    description: "Carried main tests: all ten rewritten <rule>_test.go suites green with main's case sets and expected literals byte-verbatim (ENUM_FIRST_VALUE_ZERO 917/51/3 + 667/33/5, ENUM_NO_ALLOW_ALIAS 864/49/1 + 610/31/3, IMPORT_USED 20/3/1, and the seven option rules' oracle-extracted findings: csharp 142/9/1, go_package 208/10/1, java_multiple_files 283/11/1, java_package 319/12/1, php_namespace 385/13/1, ruby_package 452/14/1, swift_prefix 452/14/1 at ruby_package's position); zero old-contract test files remain for the batch"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/package_same_csharp_namespace_test.go#TestPackageSameCsharpNamespace_Validate"
        status: pass
      - kind: other
        ref: "command: go test -race -count=1 ./internal/rules/ -run 'TestEnumFirstValueZero|TestEnumNoAllowAlias|TestImportUsed|TestPackageSame' -> ok"
        status: pass
    human_judgment: false
  - id: D7
    description: "Batch gate green on ten gates (ROADMAP Phase-4 criterion 1 — verified before 04-05 starts): full core+rules -race suite; cumulative TestLinkedParity* (21 green sub-cells, 24/41 rules); suppression centrality + readSource-once; go build; tidy diff-empty; schema drift-empty; dormancy git-diff empty over the frozen old-path files; golangci-lint 0 issues; TestModuleRuleSet green; grep-clean on optionString/mismatches over the seven ports"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "command: go test -race -count=1 ./internal/core/... ./internal/rules/... && go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity' && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ && bin/golangci-lint run ./internal/rules/... ./internal/core/..."
        status: pass
    human_judgment: false

# Metrics
duration: 34 min
started: 2026-11-19T19:48:14Z
completed: 2026-11-19T20:23:11Z
status: complete
commits: 3
plan_head_before: 246c75034e1ede2edadbb9a0f454e0acf6111ab6
plan_head_after: e0fc3f464d269bcc2dab8149efe30b5cd4a1f773
---

# Phase 04 Plan 04: Batch 3 — BASIC Semantic Summary

**Ten BASIC-semantic rules ported with main's predicates (string compare, any-allow_alias, stateless syntax-walk imports, raw written constants, skip-optionless, first-exempt) and the swift_prefix ruby_package bug pinned bug-for-bug, parity-proven old-vs-new through eight cumulative cells (24/41 rules ported)**

## Performance

- **Duration:** 34 min
- **Started:** 2026-11-19T19:48:14Z
- **Completed:** 2026-11-19T20:23:11Z
- **Tasks:** 3
- **Files modified:** 22 source/test files (+1 planning artifact: RED evidence JSON; +VALIDATION.md task rows)

## Accomplishments

- The ten BASIC-semantic rules are ported as `<Name>Linked` types: ENUM_FIRST_VALUE_ZERO keeps main's string compare (`Number != "0"` — octal "00" and hex "0x0" issue where the lab's numeric compare passes), ENUM_NO_ALLOW_ALIAS reports ANY allow_alias option at the enum keyword position, IMPORT_USED walks written references statelessly (the etalon's importUsage shape + InstructionParser reuse) and never reads LinkedFileInfo.Resolved
- The seven option-comparison ModuleRules share one uniform shape: Collect reads the shared fileOption helper, skips packageless files and optionless files entirely, adds the RAW written constant (`opt.Value` ≡ main's `option.Constant`), and Finish delegates to `firstExemptMismatches` — the lab's add-always, optionString-unquoting and report-everyone shapes are unreachable (grep-enforced over the seven ports)
- PackageSameSwiftPrefixLinked compares the literal `ruby_package` option — main's bug, pinned bug-for-bug per the orchestrator-ratified resolution: the type header documents the bug, the pin decision and the deferred post-parity fix, and the pin holds from both sides (parity cell: differing ruby_package → finding, differing swift_prefix → silence; carried regression cells assert the same)
- `TestLinkedParityBasicSemantic` proves old-vs-new parity in eight cells under cumulative 24-rule configs; every corpus was probe-validated before the RED run (old-engine count oracle + a zero-rules linked-engine link check proving no COMPILE routing)
- All ten carried test files rewritten onto the linked contract with byte-verbatim literals; the four formerly-nil-verdict option tests (csharp/java_multiple_files/php_namespace/ruby_package) now assert the concrete findings main's fixture pairs produce (extracted from the live old engine — the oracle standard)
- Batch gate green on ten gates (see Verification Log): 24/41 rules ported cumulatively with carried tests for every ported rule; standing constraints (a) readSource-once and (b) suppression centrality held after every task

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): BASIC-semantic parity suite fails on the unported-rule guard** - `c68b51d` (test)
2. **Task 2 (GREEN part 1): port #5/#6/#12 — string compare, any-allow_alias, stateless syntax-walk ImportUsed** - `a6aa2c3` (feat)
3. **Task 3 (GREEN part 2): seven option-comparison ModuleRules + carried tests + swift_prefix bug pin + batch gate** - `e0fc3f4` (feat)

**Plan metadata:** docs commits (this SUMMARY + state/roadmap/validation updates).

## TDD Gate Compliance

- RED `c68b51d` (test(04-04)) precedes GREEN `a6aa2c3`/`e0fc3f4` (feat(04-04)) in ancestry (merge-base verified: RED is an ancestor of both GREEN commits); no REFACTOR needed — the GREEN commits are the minimal implementations and no cleanup beyond the plan shape was warranted.
- RED evidence machine-validated (`gsd_run check tdd-red-evidence` → `RED_EVIDENCE_OK`, reason `target_test_failed`): TestLinkedParityBasicSemantic executed and failed on its own requirePorted assertion naming all ten unported rules while the package compiled; the semantic assessment is recorded in the evidence file.
- Task 1's RED commit touches only the `*_test.go` pathspec plus the evidence JSON planning artifact; Tasks 2-3 carry the GREEN implementations with their behavior proof (parity suite + carried tests) run before each commit.
- Note: the plan splits the batch across two GREEN tasks, so between `a6aa2c3` and `e0fc3f4` the parity suite still failed on the requirePorted guard (it probes all ten names and reports the full unported list — the 04-02 ratified decision). The three Task-2 cells were probe-verified old-vs-new under a three-name config at Task 2 time and re-proven under the full 24-name config at the Task 3 gate.

## Files Created/Modified

- `internal/rules/{enum_first_value_zero,enum_no_allow_alias,import_used}_linked.go` - NEW: the three LintRule ports (string compare + guard, any-option at enum pos, stateless syntax walk with existInProtoLinked)
- `internal/rules/package_same_{csharp_namespace,go_package,java_multiple_files,java_package,php_namespace,ruby_package,swift_prefix}_linked.go` - NEW: the seven ModuleRule ports (fileOption + raw opt.Value + firstExemptMismatches; swift_prefix reads ruby_package per the bug pin)
- `internal/rules/builder_linked.go` - ten BASIC-semantic placeholders replaced in builder.go order; rename map updated (28/41 filled; DEFAULT/COMMENTS/UNARY_RPC placeholders remain for 04-05/04-06)
- `internal/rules/parity_basic_semantic_test.go` - NEW: eight parity cells + basicSemanticUse config + probe-evidence header
- `internal/rules/{<three LintRules>,<seven option rules>}_test.go` - rewritten onto the linked contract (parseLinkedFile + validate() Collect/Finish), main's case sets and literals carried verbatim, the four nil-verdict tests upgraded to concrete findings, + TestEnumFirstValueZero_EmptyEnumGuard and TestPackageSameSwiftPrefix_RubyPackageBugPin pins

## Decisions Made

- swift_prefix bug pin: the option-name literal is main's `ruby_package`; "fixing" it here would be a behavior change that breaks Phase-7 corpus parity — the corrected behavior is a deferred post-parity decision for the 04-07 ledger (research Open Question 1 / Pitfall 5 / T-04-09)
- ImportUsed walk shape follows main's ACTUAL walk, not the plan's descriptive shorthand: file-level extends only (go-protoparser's unordered model keeps message-level extends inside MessageBody.Extends, so main's `checkInExtends(ProtoBody.Extends)` never sees them), and maps/groups/message-level options are likewise outside — documented in the ported file's header; main wins per LINT-PARITY-02
- Engine-level IMPORT_USED parity cells use compiling same-package corpus variants; the real testdata/import_used fixtures stay in the carried test (parseLinkedFile has no linking step, so the full 11-import marking semantics are proven there)
- Deterministic reporting in ImportUsedLinked: unused imports reported in source order, each distinct path once at its first occurrence (duplicated imports are rejected by protoc, so first-vs-last never differs on a compiling corpus); the engine's D-01 sort normalizes order either way

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] The real testdata/import_used corpus does not link — engine-level cells use compiling variants**
- **Found during:** Task 1 (fixture-design probe: old-engine run + zero-rules linked-engine link check)
- **Issue:** the plan's `import_used_verdicts` cell drives "the import_used fixtures (testdata/import_used/used.proto + not_used.proto) through the engines", but the linked engine routes that corpus to COMPILE before any rule runs: protoc reports the undefined extend target `import_used.thrd_party.MethodOptions`, the undefined `MethodOption`, the undeclared extension number 10000, and the duplicate `TestService` (used.proto and not_used.proto both define it in package import_used) — parity fixtures must compile (research precondition), so the cell as specified is unrunnable
- **Fix:** the engine-level cells use compiling same-package variants exercising the same verdict classes (field-type + rpc request/response marking → clean; nothing-matches → finding; qualified nested reference → finding where resolution would be clean), and the real fixtures stay in import_used_test.go, which drives the linked rule directly over parseLinkedFile infos without linking — main's full 11-import marking semantics are asserted there (clean)
- **Files modified:** internal/rules/parity_basic_semantic_test.go (cell fixtures + probe-evidence header), internal/rules/import_used_test.go (unchanged fixture set)
- **Verification:** import_used_verdicts and syntax_walk_only cells green; carried TestImportUsed_Validate green with the real fixtures (used.proto clean, not_used.proto 20/3/1)
- **Committed in:** c68b51d

**2. [Rule 1 - Bug] Carried-test literals corrected against the live oracle**
- **Found during:** Task 3 (first run of the rewrites — the OLD oracle output disagreed with two hand-written literals)
- **Issue:** the php_namespace raw-string literal carried four backslashes per separator where the written constant has two (`"ZergsLaw\\BackTemplate\\Api\\Session\\V2"`), and the swift-prefix pin cell expected byte offset 76 where the fixture's option keyword sits at 34
- **Fix:** corrected both literals to the engine's actual output (the old engine / real finding is the oracle — the test was fixed, not the code); same pattern as 04-03's keyword_positions correction
- **Files modified:** internal/rules/package_same_php_namespace_test.go, internal/rules/package_same_swift_prefix_test.go
- **Verification:** all seven carried option suites green with concrete literals; cumulative parity suite re-run green on the committed state
- **Committed in:** e0fc3f4

**3. [Rule 1 - Bug] ImportUsed extend walk follows main's actual code, not the plan's "file-level + message-level" shorthand**
- **Found during:** Task 2 (porting the walk shape)
- **Issue:** the plan's action text says to walk "extends (file-level + message-level Extend types)", but main's `checkInExtends` receives `ProtoBody.Extends` only, and go-protoparser's unordered model keeps message-level extends inside `MessageBody.Extends` (verified at source, v4.14.2) — main never walks message-level extends, so walking them would silently diverge from the oracle on proto2 corpora
- **Fix:** the port walks file-level extends only, exactly like main; message-level extends/maps/groups/message-level options are documented as outside main's walk in the ported file's header
- **Files modified:** internal/rules/import_used_linked.go (header + walk shape)
- **Verification:** parity cells and the carried used/not_used verdicts green under main's semantics
- **Committed in:** a6aa2c3

---

**Total deviations:** 3 auto-fixed (1 blocking fixture redesign, 2 correctness corrections against the live oracle)
**Impact on plan:** All three necessary for parity fidelity; no scope change, no behavior drift, old path byte-frozen throughout.

## Issues Encountered

- Task 2's `<verify>` second command (`go test -run 'TestLinkedParityBasicSemantic'`) structurally cannot pass between the two GREEN tasks: the requirePorted guard probes all ten batch names and fails while any is unported (the 04-02 full-list decision). Handled by probe-verifying the three Task-2 cells under a three-name config (assertBatchParity on the same corpora: 3/2/1/1 findings old-vs-new) and re-proving everything under the full config at the Task 3 gate. No plan change needed; recorded for the verifier.
- The batch-2 cumulative rules caught two fixture-design slips during probing (single-char messages fail MESSAGE_PASCAL_CASE; digit-led package components fail PACKAGE_LOWER_SNAKE_CASE) and the ruby-differs corpus co-fires PACKAGE_SAME_RUBY_PACKAGE — all folded into the final cell fixtures and counts before the RED run.

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components were introduced.

## Deferred Decisions (for 04-07 consolidation ledger)

- Lab behavior deliberately excluded: resolution-based ImportUsed (`validateResolved` over `Resolved.UsedImports`) — main's syntax walk wins; grep-enforced (zero code references to Resolved in import_used_linked.go); the qualified-nested-reference parity cell documents exactly where resolution-based semantics would diverge
- Main bug deliberately pinned: PackageSameSwiftPrefix comparing `ruby_package` — bug-for-bug per the ratified resolution; the corrected behavior is a deferred post-parity decision (fixing it now would break Phase-7 corpus parity); pinned from both sides
- Cross-file RPCRequestResponseUnique (the lab's cross-file/qualified ModuleRule vs main's per-file first-exempt on written names) — owned by 04-05, still pending
- Adjacent string literals in options (`"a" "b"`) merge to `"ab"` in protoast (probe observation); main's Constant basis for that edge was not probed, so it stays an observation-only note, not a parity cell

## User Setup Required

None - no external service configuration required. The etalon checkout precondition held: `/tmp/gsd-research/compiler-lab` at SHA `08cef10909c0722f7042ba17b5c11951f9c6b6f6` (verified at execution start).

## Verification Log

- `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityBasicSemantic'` → RED: exit 1, `--- FAIL: TestLinkedParityBasicSemantic` naming all ten unported rules — Task 1 verify; evidence record validated `RED_EVIDENCE_OK`
- Fixture-design probes (deleted before commit): protoast accepts `00`/`0x0` enum numbers, single-quoted `'x'` option literals, adjacent literals (merged to `"ab"`); old-engine counts 3/2/1/1/7/1/2/0/2 confirmed per cell; zero-rules link check clean on every cell corpus, and the real import_used corpus reported 5 COMPILE diagnostics (the Rule 3 redesign evidence)
- `go test -race -count=1 ./internal/rules/ -run 'TestEnumFirstValueZero|TestEnumNoAllowAlias|TestImportUsed'` → ok (all literals verbatim) — Task 2 verify
- Three-name parity probe (Test-2 cells) → ok (3/2/1/1) — Task 2 verify
- `go test -race -count=1 ./internal/rules/ -run 'TestPackageSame|TestLinkedParityBasicSemantic'` → ok — Task 3 verify
- `go test -race -count=1 ./internal/core/ -run 'Suppression|ReadSourceOnce'` → ok (standing constraints a+b, after every task)
- Grep-clean: `Resolved|validateResolved` (code lines) in import_used_linked.go → no hits; `optionString|mismatches` in the seven option ports → no hits (headers reworded per the 04-03 precedent so the acceptance grep is unambiguous)
- Batch gate (Task 3, re-run on the final state): full core+rules `-race` ok; `TestLinkedParity -v` → 21 green sub-cells (3 Wave-0 + 5 MINIMAL + 5 BASIC casing + 8 BASIC semantic); `go build ./...` ok; `go mod tidy && git diff --exit-code go.mod go.sum` → TIDY_CLEAN; `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` → SCHEMA_CLEAN; dormancy `git diff --exit-code` over the 13 frozen old-path files → DORMANCY_CLEAN; `bin/golangci-lint run ./internal/rules/... ./internal/core/...` → 0 issues; `TestModuleRuleSet` → ok
- TDD gate: `git log -E --grep '^test\((0*4)-(0*4)\):'` → c68b51d; feat commits a6aa2c3/e0fc3f4; merge-base ancestry RED→GREEN verified
- Commits measured: `git rev-list --count 246c750..HEAD` = 3 (excludes the metadata docs commit)

## Next Phase Readiness

- Ready for 04-05 (batch 4 — DEFAULT): the BASIC group is fully ported and parity-verified before the next batch starts (ROADMAP Phase-4 criterion 1); the registry holds 24/41; the remaining placeholders are exactly batch 4's scope (EnumValuePrefix, EnumZeroValueSuffix, FileLowerSnakeCase, RPCRequestResponseUnique, RPC{Request,Response}StandardName, PackageVersionSuffix, ServiceSuffix)
- The deferred-decisions list above carries this batch's parked items into the 04-07 consolidation ledger (resolution-based ImportUsed, swift_prefix fix; cross-file uniqueness rides 04-05)
- No harness changes were needed; no blockers; standing constraints (a) and (b) remain mechanized green cells

## Self-Check: PASSED

- All 22 task files exist on disk (11 created + 11 modified; verified via git log --name-status of the three task commits)
- All 3 code commits are ancestors of HEAD: c68b51d (RED test), a6aa2c3 (GREEN part 1), e0fc3f4 (GREEN part 2 + gate)
- Commits measured from the persisted ledger: `git rev-list --count 246c750..HEAD` = 3
- RED evidence record exists and validates RED_EVIDENCE_OK
- No WINDOWS.md entries needed: no stubs, no skipped tests, no unrun verifies (every `<verify>` command was executed and logged above)

---
*Phase: 04-perenos-43-pravil-s-batch-paritetom*
*Completed: 2026-11-19*
