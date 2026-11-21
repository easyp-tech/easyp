---
phase: 03-dvuhprohodnyy-dvizhok-i-podavleniya
verified: 2026-11-19T15:09:52Z
status: human_needed
score: 19/19 must-haves verified
covered_files:
  - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-01-PLAN.md
  - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-01-SUMMARY.md
  - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-02-PLAN.md
  - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-02-SUMMARY.md
  - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-03-PLAN.md
  - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-03-SUMMARY.md
  - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-04-PLAN.md
  - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/03-04-SUMMARY.md
  - .planning/phases/03-dvuhprohodnyy-dvizhok-i-podavleniya/tdd-red-evidence-03-01-task1.json
  - internal/core/core.go
  - internal/core/link.go
  - internal/core/lint_linked.go
  - internal/core/lint_linked_matrix_test.go
  - internal/core/lint_linked_read.go
  - internal/core/lint_linked_test.go
  - internal/core/lint_model.go
  - internal/core/lint_suppressions.go
  - internal/core/lint_suppressions_parity_test.go
  - internal/rules/validation_parity_test.go
covered_digest: "v4:sha256:152a70410d39921573eb14dd9954078f12f5ba3eb62e432768a0d18ea1d1f3b1"
behavior_unverified: 0
overrides_applied: 0
flagged_prohibitions: 2 # unverified-prohibition — human review recommended (ADR-550 D4): 2 of 11 plan prohibitions hold on LLM code-reading judgment without direct test evidence; they are the two human_verification items below. The other 9 are verified by executed tests or deterministic git/grep evidence.
deferred:
  - truth: "Representative-rule parity realized with real rules: engine output equals old implementation for the actual 43 rules"
    addressed_in: "Phase 4"
    evidence: "Phase 4 goal: 'Все 43 текущие правила переписаны на protoast с семантикой текущего main... каждый батч переноса верифицирован паритетом'; Phase 4 SC1. Phase 3 proves the engine contract with test doubles per research A2 (flagged in 03-01 PLAN)."
  - truth: "Directive-error rune-vs-byte column divergence (A4) resolved as a user-visible observable"
    addressed_in: "Phase 5"
    evidence: "Phase 5 goal: 'easyp lint выдаёт замечания в неизменном формате: пути, формат issues и коды выхода CLI совпадают с main'; the tab pin (tab_directive_position_pin) documents the divergence in-code as the Phase 5 flip-plan decision."
human_verification:
  - test: "Linked-path no-disk-reread review (plan 03-01 prohibition): read protoInfoFromLinked + the LintLinked pass structure, or (stronger) add a disk-instrumentation cell in Phase 4 asserting readSource is invoked exactly once per corpus file on a compiling corpus"
    expected: "On a compiling corpus the only disk reads are pass-1 readSource, once per file; pass-2 models come exclusively from Linked.Syntax via protoast.Build (tree dropped after build); only imports (never corpus files) are read on demand, through openImportFile"
    why_human: "No test instruments the filesystem; the guarantee is currently code-structural (grep confirms zero disk access in the linked model path), which is an LLM-judge verdict, not behavioral proof"
  - test: "Suppression-centrality standing constraint (plan 03-02 prohibition): at each Phase 4 rule-port batch, confirm ported rules never consult comment directives or ignoreOnly themselves, and keep the central-filter cells (comment_suppression_central, finish_comment_refilter, compile_bypasses_suppression) green"
    expected: "The engine filters centrally (checkFile per-finding filter + Finish re-filter under allowCommentIgnores); rules only return findings and never self-suppress"
    why_human: "Current test doubles are trivially compliant; the prohibition binds future Phase 4 code, which no test written today can pin"
---

# Phase 3: Двухпроходный движок и подавления — Verification Report

**Phase Goal:** Перезаписанный lint-движок на слинкованных деревьях — параллельные окна, ModuleRule для кросс-файловых правил, детерминированный порядок вывода, центральный фильтр подавлений и валидация конфигурации с паритетом main; движок готов и проверен, но ещё не врезан в CLI
**Verified:** 2026-11-19T15:09:52Z
**Status:** human_needed (all 19 must-haves verified; 2 flagged prohibitions routed to human review per ADR-550 D4 — complete with 2 flagged prohibitions)
**Re-verification:** No — initial verification

## Goal Achievement

All behavioral evidence below was produced by this verifier's own runs (not taken from SUMMARY claims): three core suites under `-race` (49 named cells PASS, 0 FAIL, 0 DATA RACE), two rules suites under `-race` (15 PASS, 0 FAIL), plus the deterministic gates (dormancy diffs, tidy, schema drift, golangci-lint).

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Engine builds each file's model from linked easypc syntax trees; walk order preserved via indexed assembly; output deterministic across runs under parallel windows (`-race` green) — ROADMAP SC1 | ✓ VERIFIED | Own run: `TestLintLinked/repeated_runs_byte_identical` (5 runs, full-slice require.Equal), `race_shared_doubles_identity` (exact per-file counts + line tallies), `engine_sort_order`, `linked_corpus_deterministic` all green under -race; code: indexed drain (lint_linked.go:111-127), unconditional `slices.SortStableFunc` by (path, rule, line, col) at exit (149-156). Parallel≡sequential follows: byte-identity across scheduler variation + walk-order drain + total sort |
| 2 | Suppressions identical to main: `ignore`/`ignore_only` + `allowCommentIgnores`, all three dialects, full fixture set passes — ROADMAP SC2 | ✓ VERIFIED | Own run: `TestLintLinkedSuppressionParity` 21/21 rows green under -race. Fixture provenance VALID: `declarations` constant and row bodies verbatim from the old-path oracle `comment_suppressions_eof_test.go`; expectations carried from main's tests; error texts asserted byte-exact incl. `path:line:column` prefix |
| 3 | Unknown rule or group in `lint.use` returns main's exact validation error — ROADMAP SC3 | ✓ VERIFIED | Own run: `TestLintUseValidationParity` 7/7 cells green (errors.Is `core.ErrInvalidRule` + exact `validateConfigRuleNames: use: / except: / ignore_only: ` prefixes); the pre-existing main oracle `TestInvalidLintNamesAreRejectedBeforeFiltering` green untouched (6/6 subtests) |
| 4 | Cross-file rules via ModuleRule (Collect/Finish) complete correctly in the parallel engine; statelessness audited inside the engine race test — ROADMAP SC4 | ✓ VERIFIED | Own run: `module_rule_lifecycle` (Validate never called, Collect once per file in walk order, Finish finding lands), `collect_serial_walk_order` (unsynchronized recording slice — the race detector itself proves serial Collect), `race_shared_doubles_identity` (shared instances, exact counts, no cross-file bleed), `ignore_only_finish_refilter` — all green under -race |
| 5 | Engine is dormant — ready and verified but NOT wired into easyp lint (goal clause; 03-01 truth) | ✓ VERIFIED | `grep -rn "LintLinked\|LinkedRules" internal/api/` → no matches; `git diff e19599c -- internal/core/lint.go internal/api/lint.go internal/api/lint_v1.go internal/core/dom.go internal/core/proto_info_read.go internal/core/comment_suppressions.go` empty; `git diff e19599c -- internal/easypc/` empty; core.go/link.go diffs additive-only (core.go's 2 deletions are gofmt realignment around the new `linkedRules` field; `rules: options.Rules` assignment intact) |
| 6 | Model built exactly once per file from Linked.Syntax; outline kept, tree dropped; no disk re-read of corpus files (D-07) | ✓ VERIFIED | Code: lint_linked_read.go:143 `protoast.Build(file.Syntax)`, :145 `file.Syntax = nil`; pass-1 `readSource` once per path (lint_linked.go:55-58); linked model path has zero disk access (grep: no readSource/os/ReadAll in protoInfoFromLinked); behavior exercised by `linked_corpus_deterministic` + `fallback_outline_cache` (linked outlines serve sibling imports) |
| 7 | Windows of 256 via internal/parallel (D-02); indexed serial drain in walk order; D-01 total sort | ✓ VERIFIED | `const lintWindow = 256` (lint_linked.go:19); `parallel.Run` at 3 sites (pass 1, windowed pass 2, parseOnTheirOwn); drain loop indexed by walk order; sort at exit (149-156) |
| 8 | Central suppression filter on per-file findings AND Finish findings; COMPILE issues bypass suppression machinery entirely (D-04/A6) | ✓ VERIFIED | Own run: `comment_suppression_central` (both sides of AllowCommentIgnores), `finish_comment_refilter/{allow_drops,disallow_keeps}_finish_finding`, `compile_error_rules_skipped`, `compile_bypasses_suppression` (arms ignoreOnly + directive fixture + AllowCommentIgnores; COMPILE issues return unchanged) — green under -race; code: lint_linked.go:73-79 returns compileIssues before suppression machinery, with the A6 comment |
| 9 | Per-file errors carry main's exact user-visible wrapper chain `fs.WalkDir: c.lintProtoInfoRead:` (main-wins pin, REQ-2.4) | ✓ VERIFIED | Code: literal present once (lint_linked.go:116) with the in-code parity-pin comment; `grep -c 'c\.protoInfoRead'` = 0 (etalon callee literal absent); own run: `fallback_missing_import_command_error` (errors.As `*OpenImportFileError` + exact prefix), `fallback_syntax_error_wrapper`, `fallback_read_error` green |
| 10 | Fallback read path: missing import → exit-2 carrier; syntax/read errors as command errors; rules still run on fell-back corpora; cached import outlines serve siblings (LINT-ENG-03 semantics through the new engine) | ✓ VERIFIED | Own run: `fallback_missing_import_command_error`, `fallback_syntax_error_wrapper`, `fallback_read_error`, `fallback_rules_still_run`, `fallback_outline_cache` green under -race; cache wiring verified (importCache mutex map, core.go:51 + lint_linked_read.go:79-117) |
| 11 | allowCommentIgnores policy independence holds per file and at Finish re-filter (D-04/D-05) | ✓ VERIFIED | Own run: `policy_allow_suppresses` / `policy_disallow_keeps` and the two Finish-refilter sub-cells green |
| 12 | Three directive dialects (easyp:disable/enable, nolint:, buf:lint:ignore) suppress identically; malformed/unknown names error with main's byte-identical texts | ✓ VERIFIED | Own run: `nolint_dialect_suppresses`, `buf_dialect_suppresses`, `nolint_unknown_name_error`, `malformed_directive_exact_text` (`test.proto:3:1: malformed lint directive "easyp:disable"`), `unknown_name_exact_text` — green; the four error formats verified in lint_suppressions.go:69,74,94,112, carried verbatim from the old skeleton |
| 13 | Config-side parity: walk-level `ignore`, `ignore_only` exact-file and dir-prefix on per-file and Finish findings (D-05) | ✓ VERIFIED | Own run: `config_ignore_walk_filter`, `ignore_only_dir_prefix`, `ignore_only_finish_refilter` green; `shouldIgnoreLinked` (lint_linked.go:218-232) mirrors main's shouldIgnore semantics (exact match then HasPrefix) |
| 14 | Tab position pinned to protoast's rune convention with the byte-vs-rune divergence documented as a Phase 5 observable (A4) | ✓ VERIFIED | Own run: `tab_directive_position_pin` green; hand-computed literal `taberr.proto:4:2` asserted (expectation not derived from the engine under test); A4 divergence comment in-code |
| 15 | Per-file errors surface in walk order, not completion order; cancelled context surfaces as context error | ✓ VERIFIED | Own run: `walk_order_error_precedence` (a_bad.proto beats z_bad.proto — pins the indexed drain), `ctx_cancel_surfaces` (errors.Is context.Canceled) green under -race |
| 16 | COMPILE class routes correctly: mixed-corpus isolation with rules skipped corpus-wide; dep diagnostics carry the dep's import path; compile output deterministic twice | ✓ VERIFIED | Own run: `mixed_corpus_compile_isolation` (exact count/path, etalon shape, bare CompileRuleName), `dep_file_diagnostic_through_engine` (Path=b.proto, verbatim message), `compile_deterministic_twice` green under -race |
| 17 | PACKAGE_NO_IMPORT_CYCLE special case (ErrInvalidRule + "is not implemented") pinned until Phase 6 | ✓ VERIFIED | Own run: `package_no_import_cycle_special_case` green (sentinel identity + exact text); `git diff e19599c -- internal/rules/` shows only the new test file — validation path untouched |
| 18 | No second lint.use validator; one registry surface feeding config validation, directive validation (knownLintRules), and future schemas (D-06) | ✓ VERIFIED | Own run: `valid_config_still_constructs` (core.New with LinkedRules exposes no error surface — the absence IS the contract), `registry_covers_all_names` green; chain verified: runtime.go:74 `KnownLintRules: rules.AllRuleNames()` → core.go:109 → lint_suppressions.go:51 |
| 19 | Phase closure: dormancy, dependency tidiness, schema drift-freedom, golangci-lint clean, full -race suite green in phase-scope packages | ✓ VERIFIED | Own runs: dormancy diffs empty (see #5); `go mod tidy && git diff --exit-code go.mod go.sum` zero diff; schema-gen vs schemas/ diff exit 0; `./bin/golangci-lint run ./...` → "0 issues."; full `go test -race -count=1 ./internal/...`: core, rules, api, protoast, easypc/*, config, migration, modules and all other phase-scope packages ok; failing packages are `internal/adapters/plugin` and `internal/generation` — both poisoned by the SAME pre-existing wazero `GetWazeroVersion` race (generation's race stack goes through `adapters/plugin.(*BuiltinPluginExecutor).Execute()`); neither package touched by any phase-3 commit (`git log e19599c..HEAD -- internal/adapters/ internal/generation/` empty); the race is the Phase-1 deferred item, reproduced pre-vendor-drop on pristine HEAD |

**Score:** 19/19 truths verified (0 present, behavior-unverified)

**Prohibition dispositions (11 flagged items):** 9 verified by executed tests or deterministic evidence (P-completion-order: walk_order_error_precedence + sort; P-collect-serial: collect_serial_walk_order; P-no-reflection/no-substring: reflect-grep 0 + marker_in_string row; P-old-path-unchanged: git diff empty; P-no-weak-assertions: source read — full-slice require.Equal + exact maps; P-compile-unfiltered: compile_bypasses_suppression; P-no-second-validator: valid_config_still_constructs; P-validation-unweakened: git diff + existing oracle test; P-no-silent-unknown: unknown-name cells on both surfaces). 2 carry non-authoritative LLM code-reading verdicts and route to human review (see human_verification): no-disk-reread on the linked path, suppression-centrality as a Phase-4 standing constraint. **unverified-prohibition — human review recommended (2).**

### Deferred Items

| # | Item | Addressed In | Evidence |
|---|------|-------------|----------|
| 1 | Real-rule parity on the linked engine (representative rules are test doubles in Phase 3, per research A2) | Phase 4 | Phase 4 goal/SC1: all 43 rules ported with per-batch old-vs-new parity |
| 2 | Directive-error rune-vs-byte column divergence (A4) as a user-visible observable | Phase 5 | Phase 5 goal: unchanged output format/exit codes; divergence documented in-code at the tab pin |

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/core/lint_model.go` | LintRule/ModuleRule/LinkedFileInfo/Resolution contracts, LinkedRuleName, rename-map header | ✓ VERIFIED | 97 lines; all symbols present; rename map (etalon→new) in header for Phase 5; D-03 ModuleRule list godoc |
| `internal/core/lint_linked.go` | LintLinked engine: windows of 256, indexed drain, D-01 sort, checkFile/collectFile, shouldIgnoreLinked | ✓ VERIFIED | 232 lines; structural claims confirmed line-by-line; wired (called by all 42 core test cells) |
| `internal/core/lint_linked_read.go` | readSource/protoInfoFromSource/protoInfoFromLinked/fallback imports/cache/parseOnTheirOwn | ✓ VERIFIED | 162 lines; Build over Linked.Syntax with tree drop; openImportFile security boundary preserved; mutex-guarded cache |
| `internal/core/lint_suppressions.go` | protoast Comments/Owner walk + main's pairing machine + byte-identical error texts | ✓ VERIFIED | 118 lines; reflection-free (grep 0); reuses parseDirectivePrefix/compactDirectiveNames/suppressionRanges in place |
| `internal/core/link.go` (additive) | linkedFiles/linkedOutline/outline (mutex+sync.Once), resolution, usesThrough | ✓ VERIFIED | +75 lines appended, zero deletions; linkSources/hasPerFileSyntaxError/diagCollector untouched |
| `internal/core/core.go` (additive) | linkedRules + importCache fields, parsedImports, Options.LinkedRules + New assignment | ✓ VERIFIED | Additive fields + gofmt realignment; old `rules` contract intact |
| `internal/core/lint_linked_test.go` | 12 named engine cells + doubles | ✓ VERIFIED | 508 lines; 12/12 cells green under -race in this verifier's run |
| `internal/core/lint_suppressions_parity_test.go` | 21 parity rows + oracle-mirroring doubles | ✓ VERIFIED | 288 lines; 21/21 green; fixtures verbatim from the old-path oracle |
| `internal/core/lint_linked_matrix_test.go` | 9 parallel-correctness/COMPILE cells | ✓ VERIFIED | 340 lines; 9/9 green; exact-count assertions, no range bounds |
| `internal/rules/validation_parity_test.go` | 7 validation-parity cells | ✓ VERIFIED | 126 lines; 7/7 green; zero production changes in internal/rules |

### Key Link Verification

Note: `gsd-tools verify.key-links` reported false negatives for same-package symbol reuse and fixture-echo links (it string-matches the target filename in the source). Each was verified manually.

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| internal/core/lint_linked.go | internal/core/link.go | LintLinked → c.linkSources wrapped in linkedFiles | ✓ WIRED | `c.linkSources(ctx, fsWalker, paths, sources)` at lint_linked.go:68; wrapped as linkedFiles at :70 |
| internal/core/lint_linked_read.go | internal/protoast | protoast.Build(file.Syntax); OutlineOf; Parse | ✓ WIRED | Build at :143, `file.Syntax = nil` at :145, outline serving at :149 |
| internal/core/lint_linked.go | internal/parallel | parallel.Run (pass 1, windows, parseOnTheirOwn) | ✓ WIRED | lint_linked.go:55,91; lint_linked_read.go:125 |
| internal/core/lint_suppressions.go | internal/core/comment_suppressions.go | skeleton reused in place | ✓ WIRED | parseDirectivePrefix (:63), compactDirectiveNames (:71), suppressionRanges (:48,80) — tool false negative |
| internal/core/core.go | internal/core/lint_linked_read.go | importCache lazily initialized in LintLinked | ✓ WIRED | core.go:51 field; lazy init lint_linked.go:47-48; mutex use lint_linked_read.go:79-117 — tool false negative |
| internal/core/core.go | internal/core/lint_linked.go | linkedRules consumed by the new engine only | ✓ WIRED | core.go:23,72,105; consumed at lint_linked.go:130,179,204 — tool false negative |
| lint_suppressions_parity_test.go | comment_suppressions_eof_test.go | fixtures verbatim as parity oracle | ✓ WIRED | `declarations` constant and row bodies byte-identical; expectations carried from main's table — tool false negative (fixture echo) |
| lint_linked_matrix_test.go | lint_linked.go / link_test.go | drives LintLinked; mirrors Phase 2 cells | ✓ WIRED | TestLintLinkedMatrix green through LintLinked; Phase 2 cells (compile_error/dep_file_semantic_error/deterministic_order_twice) mirrored one level up |
| validation_parity_test.go | internal/rules/selection.go | ValidateNames → ErrInvalidRule via rules.New | ✓ WIRED | Test calls New() and pins the wrapping chain end-to-end (green) |
| validation_parity_test.go | internal/rules/selection_test.go | existing main oracle stays green untouched | ✓ WIRED | TestInvalidLintNamesAreRejectedBeforeFiltering 6/6 subtests green in own run; selection_test.go untouched |
| internal/api/runtime.go | internal/core/lint_suppressions.go | KnownLintRules: rules.AllRuleNames() → known set | ✓ WIRED | runtime.go:74 → core.go:109 → lint_suppressions.go:51 (pre-existing wiring, read-only) |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| LintLinked return | []IssueInfo | corpus files → readSource → linkSources (real easypc compile) → Linked.Syntax → protoast.Build → rule.Validate | Yes — linked syntax trees of the actual corpus | ✓ FLOWING |
| COMPILE issues | []IssueInfo | real compiler diagnostics from linkSources | Yes — verbatim compiler messages/positions (pinned by cells) | ✓ FLOWING |
| linkedCommentSuppressions | suppressionRanges | protoast File.Comments/Owner of the file's built model | Yes — parsed comment positions, no static data | ✓ FLOWING |
| protoInfoFromLinked | LinkedFileInfo.ImportedFiles | linked.outline (OutlineOf over linked trees) / cached outlines | Yes | ✓ FLOWING |

No static returns, hardcoded findings, or mock data anywhere in the engine path; test doubles live only in test files.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Engine + fallback matrix (12 cells) | `go test -race -count=1 -v ./internal/core/ -run 'TestLintLinked'` | 49 PASS (12 + 21 + 9 cells + parents/subtests), 0 FAIL, 0 DATA RACE, ok 1.071s | ✓ PASS |
| Suppression parity (21 rows) | (same run, prefix match) | 21/21 named rows PASS | ✓ PASS |
| Parallel-correctness/COMPILE matrix (9 cells) | (same run, prefix match) | 9/9 named cells PASS | ✓ PASS |
| Validation parity guard (7 cells) | `go test -race -count=1 -v ./internal/rules/ -run 'TestLintUseValidationParity\|TestInvalidLintNamesAreRejectedBeforeFiltering'` | 15 PASS (7 + 6 oracle subtests + parents), 0 FAIL | ✓ PASS |
| Whole-core-package regressions | `go test -race -count=1 ./internal/...` (single full run) | internal/core ok 2.270s; internal/rules ok 3.273s; internal/api ok 239.912s | ✓ PASS |
| Dormancy | `git diff e19599c -- <sealed files>` + `git diff e19599c -- internal/easypc/` | empty | ✓ PASS |
| Dependency tidiness | `go mod tidy && git diff --exit-code go.mod go.sum` | zero diff | ✓ PASS |
| Schema drift | `go run ./cmd/easyp schema-gen --out-dir $tmp && diff -r $tmp schemas/` | exit 0 | ✓ PASS |
| Lint | `./bin/golangci-lint run ./...` | "0 issues." | ✓ PASS |
| TDD RED authenticity | `tdd-red-evidence-03-01-task1.json` | exitCode 1; all 7 cells failed on planned require.Len assertions against the (nil,nil) stub; compiles + discovers all cells | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| (none) | `find scripts -path '*/tests/probe-*.sh'` → none; no probe declared in any PLAN/SUMMARY | not applicable | SKIP |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| LINT-ENG-01 | 03-01, 03-03 | REQ-2.1: model per file from linked easypc trees, all applicable rules, walk order unchanged | ✓ SATISFIED (phase scope: dormant engine test-proven; CLI visibility lands at Phase 5 flip per the milestone's phase mapping) | Truths 1,2,6,7,10,15,16; 21 core cells green under -race |
| LINT-ENG-05 | 03-02 | REQ-2.5: same suppression rules as main (ignore/ignore_only/comment directives) | ✓ SATISFIED | Truths 2,11,12,13,14; 21/21 parity rows on main's own fixture table |
| LINT-ENG-06 | 03-04 | REQ-2.6: unknown rule/group in lint.use returns the same validation error | ✓ SATISFIED | Truths 3,17,18; 7/7 guard cells + untouched main oracle |

Orphaned requirements: none — REQUIREMENTS.md maps exactly LINT-ENG-01/05/06 to Phase 3, and all three are claimed by plans (03-01/03-03, 03-02, 03-04 respectively).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | No TBD/FIXME/XXX/TODO/PLACEHOLDER markers, no stub returns, no hardcoded empty data, no console-log-only code in any phase file | — | — |

ℹ️ Info — SUMMARY reporting nuance (not a code defect): 03-04-SUMMARY states the full -race suite has "the single failing package internal/adapters/plugin". This verifier's run shows `internal/generation` also fails — poisoned by the same wazero race through `adapters/plugin.(*BuiltinPluginExecutor).Execute()` (race stack in the generation section of the log names wazero, not generation code). Generation is untouched by phase-3 commits and the root cause is the already-deferred Phase-1 item; classification ratified, count corrected here.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| internal/core/lint_linked_test.go | LINT-ENG-01 | 12 cells | 0 | 0 | Behavioral (full-slice require.Equal, exact counts, errors.As + exact prefix) | VALID |
| internal/core/lint_suppressions_parity_test.go | LINT-ENG-05 | 21 rows | 0 | 0 | Value/behavioral; provenance VALID (fixtures + expectations from main's old-path tests, not from the system under test; tab column hand-computed) | VALID |
| internal/core/lint_linked_matrix_test.go | LINT-ENG-01 | 9 cells | 0 | 0 | Behavioral (exact maps, full-slice identity, race detector as serialization proof) | VALID |
| internal/rules/validation_parity_test.go | LINT-ENG-06 | 7 cells | 0 | 0 | Value (errors.Is sentinel + exact wrapped text prefixes); oracle = untouched main test, green | VALID |

**Disabled tests on requirements:** 0. **Circular patterns:** 0. **Insufficient assertions:** 0. Coverage quantities meet every plan gate (>=12, >=20, >=9, >=7).

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts: 7/7 (D-01 through D-07), 0 not honored. (Gate: `check.decision-coverage-verify`, non-blocking.)

### Human Verification Required

**unverified-prohibition — human review recommended (2 items).** The phase is otherwise fully verified; these two plan-flagged prohibitions hold on code-reading judgment without direct test evidence and must not silently pass.

### 1. Linked-path no-disk-reread (plan 03-01 prohibition)

**Test:** Review `internal/core/lint_linked_read.go` `protoInfoFromLinked` + the LintLinked pass structure; or add a disk-instrumentation cell in Phase 4 asserting readSource runs exactly once per corpus file on a compiling corpus.
**Expected:** Only pass-1 readSource touches corpus files; pass-2 models come exclusively from Linked.Syntax via protoast.Build (tree dropped after build); only imports are read on demand, via openImportFile.
**Why human:** No test instruments the filesystem; current evidence is code-structural (grep: zero disk access in the linked model path) — an LLM-judge verdict, not behavioral proof.

### 2. Suppression-centrality as a standing Phase-4 constraint (plan 03-02 prohibition)

**Test:** At each Phase 4 rule-port batch, confirm ported rules never consult comment directives or ignoreOnly themselves; keep the central-filter cells green.
**Expected:** The engine filters centrally (checkFile per-finding + Finish re-filter); rules only return findings.
**Why human:** Current doubles are trivially compliant; the prohibition binds future Phase 4 code that no test written today can pin.

### Gaps Summary

No gaps. Every roadmap success criterion and every plan must-have is verified with evidence produced by this verification's own runs: the two-pass linked engine is real, substantive, wired, and behaviorally proven (determinism under parallel windows with byte-identical repeated runs, walk-order indexed drain and error precedence, serial ModuleRule lifecycle proven by the race detector, central suppression filtering on all three surfaces, COMPILE routing per the Phase 2 contract); suppression parity holds on main's own byte-identical fixture table (21/21); lint.use validation parity is pinned through the untouched rules.New chain; and the engine is provably dormant — no API references, sealed files byte-unchanged since the phase base, additive-only core.go/link.go changes, zero new dependencies, schemas drift-free, golangci-lint clean.

Status is `human_needed` solely because two of the eleven plan-flagged prohibitions (no-disk-reread on the linked path; suppression-logic-never-in-rules as a standing constraint) rest on code-reading judgment rather than a direct behavioral test — per the ADR-550 D4 contract they are surfaced for human review rather than silently passed. Both carry favorable LLM-judge verdicts with concrete code evidence; neither blocks Phase 4 planning.

---

_Verified: 2026-11-19T15:09:52Z_
_Verifier: Claude (gsd-verifier)_
