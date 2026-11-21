---
phase: 04-perenos-43-pravil-s-batch-paritetom
verified: 2026-11-20T17:06:49Z
status: passed
score: 10/10 must-haves verified
covered_files: [".golangci.yml", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-01-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-01-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-02-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-02-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-03-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-03-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-04-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-04-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-05-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-05-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-06-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-06-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-07-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-07-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-08-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-08-SUMMARY.md", "internal/core/lint_linked.go", "internal/core/lint_linked_readsource_test.go", "internal/core/lint_model.go", "internal/core/position_probe_test.go", "internal/rules/builder_linked.go", "internal/rules/builder_test.go", "internal/rules/casing_test.go", "internal/rules/check_lint_ignore_test.go", "internal/rules/comment_enum_linked.go", "internal/rules/comment_enum_test.go", "internal/rules/comment_enum_value_linked.go", "internal/rules/comment_enum_value_test.go", "internal/rules/comment_field_message_linked.go", "internal/rules/comment_field_message_test.go", "internal/rules/comment_message_linked.go", "internal/rules/comment_message_test.go", "internal/rules/comment_one_of_linked.go", "internal/rules/comment_one_of_test.go", "internal/rules/comment_rpc_linked.go", "internal/rules/comment_rpc_test.go", "internal/rules/comment_service_linked.go", "internal/rules/comment_service_test.go", "internal/rules/directory_same_package_linked.go", "internal/rules/directory_same_package_test.go", "internal/rules/enum_first_value_zero_linked.go", "internal/rules/enum_first_value_zero_test.go", "internal/rules/enum_no_allow_alias_linked.go", "internal/rules/enum_no_allow_alias_test.go", "internal/rules/enum_pascal_case_linked.go", "internal/rules/enum_pascal_case_test.go", "internal/rules/enum_value_prefix_linked.go", "internal/rules/enum_value_prefix_test.go", "internal/rules/enum_value_upper_snake_case_linked.go", "internal/rules/enum_value_upper_snake_case_test.go", "internal/rules/enum_zero_value_suffix_linked.go", "internal/rules/enum_zero_value_suffix_test.go", "internal/rules/file_lower_snake_case_linked.go", "internal/rules/file_lower_snake_case_test.go", "internal/rules/import_no_public_linked.go", "internal/rules/import_no_public_test.go", "internal/rules/import_no_weak_linked.go", "internal/rules/import_no_weak_test.go", "internal/rules/import_used_linked.go", "internal/rules/import_used_test.go", "internal/rules/init_test.go", "internal/rules/lint_helpers.go", "internal/rules/message_field_lower_snake_case_linked.go", "internal/rules/message_field_lower_snake_case_test.go", "internal/rules/message_pascal_case_linked.go", "internal/rules/message_pascal_case_test.go", "internal/rules/module_rules_guard_test.go", "internal/rules/oneof_lower_snake_case_linked.go", "internal/rules/oneof_lower_snake_case_test.go", "internal/rules/package_defined_linked.go", "internal/rules/package_defined_test.go", "internal/rules/package_directory_match_linked.go", "internal/rules/package_directory_match_test.go", "internal/rules/package_lower_snake_case_linked.go", "internal/rules/package_lower_snake_case_test.go", "internal/rules/package_same_csharp_namespace_linked.go", "internal/rules/package_same_csharp_namespace_test.go", "internal/rules/package_same_directory_linked.go", "internal/rules/package_same_directory_test.go", "internal/rules/package_same_go_package_linked.go", "internal/rules/package_same_go_package_test.go", "internal/rules/package_same_java_multiple_files_linked.go", "internal/rules/package_same_java_multiple_files_test.go", "internal/rules/package_same_java_package_linked.go", "internal/rules/package_same_java_package_test.go", "internal/rules/package_same_php_namespace_linked.go", "internal/rules/package_same_php_namespace_test.go", "internal/rules/package_same_ruby_package_linked.go", "internal/rules/package_same_ruby_package_test.go", "internal/rules/package_same_swift_prefix_linked.go", "internal/rules/package_same_swift_prefix_test.go", "internal/rules/package_version_suffix_linked.go", "internal/rules/package_version_suffix_test.go", "internal/rules/parity_basic_casing_test.go", "internal/rules/parity_basic_semantic_test.go", "internal/rules/parity_comments_test.go", "internal/rules/parity_default_test.go", "internal/rules/parity_full_registry_test.go", "internal/rules/parity_harness_test.go", "internal/rules/parity_minimal_test.go", "internal/rules/parity_suppression_test.go", "internal/rules/parity_unary_rpc_test.go", "internal/rules/parity_wave0_test.go", "internal/rules/registry_parity_guard_test.go", "internal/rules/rpc_no_client_streaming_linked.go", "internal/rules/rpc_no_client_streaming_test.go", "internal/rules/rpc_no_server_streaming_linked.go", "internal/rules/rpc_no_server_streaming_test.go", "internal/rules/rpc_pascal_case_linked.go", "internal/rules/rpc_pascal_case_test.go", "internal/rules/rpc_request_response_unique_linked.go", "internal/rules/rpc_request_response_unique_test.go", "internal/rules/rpc_request_standard_name_linked.go", "internal/rules/rpc_request_standard_name_test.go", "internal/rules/rpc_response_standard_name_linked.go", "internal/rules/rpc_response_standard_name_test.go", "internal/rules/selection_test.go", "internal/rules/service_pascal_case_linked.go", "internal/rules/service_pascal_case_test.go", "internal/rules/service_suffix_linked.go", "internal/rules/service_suffix_test.go", "internal/rules/validation_parity_test.go"]
covered_digest: "v4:sha256:c8e96e7851581648844c2b46a020719d8d2cff4373d71654e121c57854a29e42"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 6/10
  gaps_closed:
    - "Suppression parity with main: ignore_only behaves identically to main (CR-01 — collectFile skips ModuleRule.Collect for shouldIgnoreLinked files; old=0 AND new=0 on the exempt-first-ignored corpus)"
    - "The parity harness compares identical suppression state old-vs-new (WR-01 — NewLinked's ignoreOnly feeds core.Options.IgnoreOnly; three suppression cells)"
    - "Duplicate file-level option occurrences behave as main (WR-02 — per-occurrence iteration in all seven option ports; old=2 AND new=2)"
    - "Duplicate import statements report at last written occurrence (WR-03 — importInfo overwrite map; line-6 parity)"
    - "The parity fingerprint covers every user-visible field including SourceName (WR-04 — fifth cmp.Or term in harness fingerprint and engine D-01 sort; tiebreak cell)"
    - "Repo formatting convention holds on all phase files (IN-01 — gofmt -l internal/rules/ empty)"
  gaps_remaining: []
  regressions: []
deferred:
  - truth: "Resolution-based ImportUsed (lab validateResolved) is excluded"
    addressed_in: "Phase 6/7"
    evidence: "Ledger item 1 (04-07-SUMMARY Consolidated Deferred-Decisions Ledger); Phase 7 owns corpus parity gates (LINT-PARITY-01). Reconfirmed: LinkedFileInfo.Resolved never read by the port (only the explanatory comment names it)"
  - truth: "Cross-file RPCRequestResponseUnique (lab shape) is excluded; main's per-file first-exempt shipped"
    addressed_in: "Phase 6/7"
    evidence: "Ledger item 2; deliberate post-parity behavior change owner Phase 6/7"
  - truth: "swift_prefix ruby_package bug is pinned bug-for-bug, not fixed"
    addressed_in: "Phase 6/7"
    evidence: "Ledger item 3; corrected behavior would break Phase-7 corpus parity. Reconfirmed: BUG PIN comments (T-04-09) and the ruby_package literal survived the WR-02 sweep verbatim"
  - truth: "Lab casing helpers (casing.go acronym-aware algorithms) not adopted"
    addressed_in: "Phase 5"
    evidence: "Ledger item 5 + Phase 5 goal 'старый lint-контур удалён': Phase 5 deletes casing.go with the old path"
  - truth: "Lab report-everyone groupedValues.mismatches stays verbatim-but-unreachable"
    addressed_in: "Phase 5"
    evidence: "Ledger item 6: Phase 5 deletes it with the old path"
  - truth: "Old-path consolidation (old rule files, builder.go, D-01-kept helpers incl. the now-unconsumed fileOption, .golangci.yml exclusion) not done in Phase 4"
    addressed_in: "Phase 5"
    evidence: "Ledger items 14-15 + Phase 5 SC 2 'Старый lint-контур и старые реализации правил удалены'; fileOption is documented kept-unused (lint_helpers.go + .golangci.yml comment updated in mirror by the gap closure) with the Phase-5 deletion obligation"
  - truth: "PACKAGE_NO_IMPORT_CYCLE not a ModuleRule yet (absence pinned by TestModuleRuleSet)"
    addressed_in: "Phase 6"
    evidence: "Ledger item 17 + Phase 6 SC 1 'PACKAGE_NO_IMPORT_CYCLE в MINIMAL ... на ModuleRule'"
  - truth: "Adjacent option string literals ('a' 'b' merge) edge unprobed"
    addressed_in: "Phase 7"
    evidence: "Ledger item 13: observation note; Phase 7 corpus runs own the finding"
---

# Phase 04: Перенос 43 правил с батч-паритетом — Verification Report

**Phase Goal:** Все 43 текущих правила переписаны на protoast с семантикой текущего main (main побеждает лабораторную копию); каждый батч переноса верифицирован паритетом старое-vs-новое до перехода к следующему (registry truth: 41 rules — Pitfall 7)
**Verified:** 2026-11-20T17:06:49Z
**Status:** passed
**Re-verification:** Yes — after gap closure (plan 04-08 closed the six gaps CR-01, WR-01, WR-02, WR-03, WR-04, IN-01 from the previous gaps_found report)

**MVP-mode disposition (milestone-ratified):** ROADMAP sets `Mode: mvp` but the phase goal is not a User Story (`user-story.validate` → false). This discrepancy was surfaced at Phase 1 and **ratified by the human** at Phase 2 (02-UAT.md Test 2, result: pass): this milestone is an internal compiler migration; goal-backward verification against the ROADMAP success criteria applies. No "User Flow Coverage" section is emitted (no user story exists to decompose). Disposition recorded; no open decision remains.

## Goal Achievement

The gap closure is real and complete. All six previously-failing must-haves now verify against the actual code with my own runs: the engine skips `Collect` for ignored files exactly like main's skip of `Validate` (CR-01), the harness drives both engines with identical suppression state (WR-01), duplicate file options compare per occurrence across all seven ports (WR-02), duplicate imports report at the last written occurrence (WR-03), the fingerprint and the engine's D-01 sort are total through the user-visible SourceName (WR-04), and `internal/rules` is gofmt-clean (IN-01). Each fix is gated RED-before/GREEN-after with machine-validated evidence records, the parity suite grew 47 → 53 green cells with no regression, and every previously-passing truth still holds (registry, positions, dormancy, old path byte-frozen). The phase goal is achieved.

### Observable Truths

| #   | Truth   | Status     | Evidence       |
| --- | ------- | ---------- | -------------- |
| 1   | All 41 rules produce identical sorted (path, rule, line, col, source) multisets old-vs-new on the parity corpora; per-batch parity verified cumulatively (SC1, corpus-scoped) | ✓ VERIFIED | Own run: full `go test -race -count=1 ./internal/core/... ./internal/rules/...` green (exit 0); TestLinkedParity = 9 parents, 53 PASS cells (was 47 + 6 new). The previous report's coincidental-reliance (fixture-only) is retired: the suite now covers suppression state (3 cells), duplicate options, duplicate imports, and source totality — the hidden fixture-scope assumption was promoted into declared, guarded preconditions exactly as the previous report's harden note prescribed |
| 2   | Batch-parity process: each batch gated RED-before-GREEN and re-verified cumulatively before the next | ✓ VERIFIED | Prior batches' RED evidence unchanged (files untouched since prior pass). Gap-closure RED evidence re-validated by my own gsd_run calls: all three tdd-red-evidence-04-08-task{1,2,3}.json → RED_EVIDENCE_OK (target_test_failed) with the exact planned failures (old=0/new=1 at opts/v1/beta.proto:5:1; old=2/new=0; line 6 vs line 5) |
| 3   | Registry truth: 41/41 rules in NewLinked under canonical names; names/groups/config plumbing/ignoreOnly unwrap parity with rules.New; exact D-03 ModuleRule set (10) | ✓ VERIFIED | Own run: TestRegistryParity + TestModuleRuleSet PASS; 41 `&...Linked{` registrations in builder_linked.go (grep count) |
| 4   | Main-wins semantics for changed-in-main behavior (SC2): keyword Decl.Pos positions, main regexps, written-constant compares, ruby_package bug pin, ImportUsed stateless walk, comment presence predicate, first-exempt groups, packageless invisibility | ✓ VERIFIED | Own greps: zero NamePos/NumberPos/FilePosition code hits in the ports; `.Resolved` referenced only in an explanatory comment (never read); BUG PIN comments and the ruby_package literal survived the WR-02 sweep verbatim; carried tests green in the full-suite run |
| 5   | Every ported rule carries a colocated `<rule>_test.go` with main's case sets (SC3); task test/lint-equivalents green at HEAD | ✓ VERIFIED | 41/41 colocated test files present; own runs: go build ./... OK, full -race suite green, `bin/golangci-lint run ./internal/rules/... ./internal/core/...` = 0 issues, schema-gen drift-free |
| 6   | Suppression parity with main: ignore_only behaves identically to main (REQ-2.5 semantics carried into the linked engine) — **CR-01** | ✓ VERIFIED (was FAILED) | Source: collectFile (internal/core/lint_linked.go:209-221) skips moduleRule.Collect when shouldIgnoreLinked(rule, info.Path) — mirroring main's shouldIgnore-before-Validate (core/lint.go:48); both contradicted doc claims rewritten (grep gates 'stays complete' = 0, 'apply to the findings by' = 0). Behavioral (own run): TestLinkedParitySuppressions/first_exempt_file_ignored green — the cell asserts the old side = 0 before the compare and requires fingerprint equality, so old=0 AND new=0 (pre-fix new=1 at opts/v1/beta.proto:5:1). Affects all nine group ModuleRules via the one engine fix |
| 7   | Parity harness compares identical suppression state (IgnoreOnly wired into both sides) — **WR-01** | ✓ VERIFIED (was FAILED) | Source: parity_harness_test.go:81-86 captures `newRules, newIgnoreOnly` and sets `core.Options.IgnoreOnly: newIgnoreOnly` on the linked app; grep gate 'newRules, _, err' = 0. Behavioral: the suppression cells only pass because both engines run the same state — first_exempt_file_ignored exercises the identical ignore_only config on both sides |
| 8   | Duplicate-input parity on inputs both engines accept: file-level options per-occurrence (main), duplicate imports last-wins (main) — **WR-02 + WR-03** | ✓ VERIFIED (was FAILED) | WR-02 source: `for _, opt := range info.File.Options` per-occurrence loop in all seven option ports (grep: 7 files); `fileOption(info.File` = 0 files; fileOption doc moved to kept-unused with the .golangci.yml mirror updated. Behavioral (own run): duplicate_file_option green — old side = 2 asserted, fingerprint equality ⇒ new = 2 at matching positions (pre-fix new=0). WR-03 source: importInfo overwrite map keyed by ImportPath (import_used_linked.go:93-98), report loop over the map (:111-116); grep gates 'rejected by protoc' = 0, 'first written occurrence' = 0. Behavioral: duplicate_import_last_occurrence green — old = 1 asserted, fingerprint equality proves the same position (old is main's last-written line 6; pre-fix new reported line 5) |
| 9   | Parity fingerprint covers every user-visible field incl. SourceName — **WR-04** | ✓ VERIFIED (was FAILED) | Source: SourceName is the fifth cmp.Or term in BOTH the harness fingerprint (parity_harness_test.go:116) and the engine's final D-01 sort (lint_linked.go:158); contract comments updated to (path, rule, line, col, source) with Message deliberately outside. Behavioral (own run): TestFingerprintSourceNameTiebreak green — ties on (path, rule, line, col) order by SourceName, and one-element fingerprints differing only in SourceName are NotEqual |
| 10  | Dormancy: easyp lint unchanged, old path byte-frozen, LinkedRules test-only; schemas drift-free | ✓ VERIFIED | Old-path files (core/lint.go, lint_suppressions.go, link.go, core.go, rules/builder.go, old `<rule>.go` files) last touched pre-Phase-4 per git log; working-tree dormancy stages 1+2 (exclude-pathspec diffs) exit 0 in my run; the seven gap-closure commits touch only the plan's declared files; schema-gen drift-free |

**Score:** 10/10 truths verified (0 present, behavior-unverified)

### Deferred Items

Deliberate main-wins or lab-exclusion decisions from the 04-07 Consolidated Deferred-Decisions Ledger, re-matched against the Phase 5/6/7 roadmap goals — all still validly deferred (frontmatter `deferred:`). These are not gaps: the shipped behavior IS main's.

| # | Item | Addressed In | Evidence |
|---|------|-------------|----------|
| 1 | Resolution-based ImportUsed (lab validateResolved) | Phase 6/7 | Ledger item 1; Phase 7 corpus parity gates |
| 2 | Cross-file RPCRequestResponseUnique (lab shape) | Phase 6/7 | Ledger item 2 |
| 3 | swift_prefix ruby_package bug fix | Phase 6/7 | Ledger item 3; bug pin byte-verified surviving the gap closure |
| 4 | casing.go / lab helper deletion, old-path consolidation, exclusion removal (incl. unconsumed fileOption) | Phase 5 | Ledger items 5/6/14 + Phase 5 SC 2 |
| 5 | buildCore LinkedRules production wiring (dormancy is deliberate) | Phase 5 | Ledger item 15 + Phase 5 goal "врезка — одно изменение в шве buildCore" |
| 6 | PACKAGE_NO_IMPORT_CYCLE as ModuleRule | Phase 6 | Ledger item 17 + Phase 6 SC 1 |
| 7 | Adjacent option string literals edge | Phase 7 | Ledger item 13; corpus runs own it |

### Advisory (New Scope, Unevidenced)

None. No new-scope unevidenced findings on any phase file.

Out-of-scope observation (recorded, non-blocking): `gofmt -l internal/core/` flags `internal/core/link_test.go` — a Phase-2 file last modified 2026-11-19 14:15 (commit dc49ec1, before Phase 4 started), never part of any Phase-4 plan's file set, not in this phase's covered_files, and outside the IN-01 gap's scope (`internal/rules/`). It predates the phase entirely; the 04-08-SUMMARY documented it as pre-existing and out of scope. Not a Phase-4 gap.

### Required Artifacts

| Artifact | Expected    | Status | Details |
| -------- | ----------- | ------ | ------- |
| `internal/core/lint_linked.go` | Linked engine: collectFile membership + Finish drain + D-01 sort | ✓ VERIFIED | collectFile skips Collect for ignored rule/file pairs (CR-01 fix); SourceName fifth sort term (WR-04); corrected comments; wired by all 41 ports |
| `internal/core/lint_model.go` | Linked seam (LinkedFileInfo, AppendLinkedIssue, ModuleRule contract) | ✓ VERIFIED | Suppression doc claim rewritten to the corrected REQ-2.5 semantics; grep gate 0 |
| `internal/rules/builder_linked.go` | NewLinked registry mirroring New | ✓ VERIFIED | 41 rules, TestRegistryParity green in own run |
| `internal/rules/<rule>_linked.go` (41) | Ported rules with main semantics | ✓ VERIFIED | Per-occurrence iteration in all 7 option ports (WR-02); last-wins import reporting (WR-03); BUG PIN comments verbatim |
| `internal/rules/lint_helpers.go` | Shared Collect/Finish helpers on the linked contract | ✓ VERIFIED | firstExemptMismatches reproduces main's first-seed/rest-compare over the per-occurrence lists; fileOption documented kept-unused (nine D-01-kept helpers) |
| `internal/rules/parity_harness_test.go` | Old-vs-new sorted multiset compare harness | ✓ VERIFIED | Threads NewLinked's ignoreOnly into the linked app (WR-01); fingerprint total through SourceName (WR-04); TestFingerprintSourceNameTiebreak green |
| `internal/rules/parity_suppression_test.go` | NEW: three suppression parity cells | ✓ VERIFIED | Green in own run with provenance headers recording the CR-01 probe numbers |
| `internal/rules/<rule>_test.go` (41) | Carried main tests on the linked contract | ✓ VERIFIED | Green in own full-suite run |
| `internal/core/position_probe_test.go` / `lint_linked_readsource_test.go` | Decl.Pos pin + readSource-once cell | ✓ VERIFIED | TestReadSourceOnce green (constraint a) in own run; position greps clean |
| `internal/rules/module_rules_guard_test.go` / `registry_parity_guard_test.go` | Exact-set + registry-parity guards | ✓ VERIFIED | Both green in own run; registry guard now gofmt-clean (IN-01) |
| `.golangci.yml` | Exclusion comment matching reality | ✓ VERIFIED | fileOption moved to the kept-unused list in mirror with lint_helpers.go |

### Key Link Verification

| From | To  | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| `internal/core/lint_linked.go` | `internal/core/lint.go` | shouldIgnoreLinked-before-Collect mirrors main's shouldIgnore-before-Validate (core/lint.go:48) | ✓ WIRED | Read of source: collectFile lines 209-221 match the CR-01 fix sketch; behavior proven by first_exempt_file_ignored |
| `internal/rules/parity_harness_test.go` | `internal/rules/builder_linked.go` | NewLinked's second return feeds core.Options.IgnoreOnly | ✓ WIRED | Read of source: lines 81-86; suppression cells prove both engines run the same state |
| `internal/rules/package_same_*_linked.go` (7) | their main counterparts | Per-occurrence option iteration == main's ProtoBody.Options loop | ✓ WIRED | 7/7 ports carry the loop; firstExemptMismatches reproduces first-seed/rest-compare; duplicate_file_option old=2/new=2 |
| `internal/rules/import_used_linked.go` | `internal/rules/import_used.go` | Last-wins importInfo overwrite == main's map overwrite | ✓ WIRED | Map keyed by ImportPath, overwritten per occurrence; duplicate_import_last_occurrence line-6 parity |
| `internal/rules/parity_harness_test.go` | `internal/core/lint_linked.go` | SourceName as fifth cmp.Or term in BOTH sort keys | ✓ WIRED | Both sites read-verified; TestFingerprintSourceNameTiebreak pins ordering + value compare |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| All 41 ports | Decl positions/names/values | protoast.File (linked easypc trees) | Yes — no static returns, no mocked model | ✓ FLOWING |
| Parity harness | issues old/new | Both engines on real temp-dir corpora walked by fs.NewFSWalker | Yes — old-side counts asserted before compare (anti-vacuity) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Full rules+core suite green | `go test -race -count=1 ./internal/core/... ./internal/rules/...` | ok (core, path_helpers, rules), exit 0 | ✓ PASS |
| CR-01 suppression membership | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParitySuppressions'` -v | first_exempt_file_ignored + mismatching_file_ignored + per_file_rule_ignored all PASS (old=0/0/1 asserted before compare) | ✓ PASS |
| WR-04 fingerprint totality | `go test -race -count=1 ./internal/rules/ -run 'TestFingerprintSourceNameTiebreak'` -v | PASS (order + NotEqual pins) | ✓ PASS |
| WR-02/WR-03 duplicate inputs | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityBasicSemantic'` -v | 10/10 subtests PASS incl. duplicate_file_option (old=2) and duplicate_import_last_occurrence (old=1) | ✓ PASS |
| Parity suite total | `go test -race -count=1 ./internal/rules/ -run TestLinkedParity` -v | 9 parents, 53 PASS cells | ✓ PASS |
| Registry + exact-set guards | `go test -race -count=1 ./internal/rules/ -run 'TestRegistryParity\|TestModuleRuleSet'` -v | both PASS | ✓ PASS |
| Constraints a+b | `go test -race -count=1 ./internal/core/ -run 'TestReadSourceOnce\|TestLintLinkedSuppressionParity'` -v | both PASS | ✓ PASS |
| Build | `go build ./...` | exit 0 | ✓ PASS |
| gofmt convention (IN-01) | `gofmt -l internal/rules/` | empty | ✓ PASS |
| Schema drift | `go run ./cmd/easyp schema-gen` + `git diff --exit-code schemas/` | clean | ✓ PASS |
| Lint | `bin/golangci-lint run ./internal/rules/... ./internal/core/...` | 0 issues | ✓ PASS |
| Dormancy stages | `git diff --exit-code -- internal/rules/ ':(exclude)*_linked.go' ...` × 2 | both exit 0 | ✓ PASS |
| RED evidence validity | `gsd_run check tdd-red-evidence` × 3 | RED_EVIDENCE_OK × 3 | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| Gap-closure acceptance re-runs (supersede the 04-REVIEW -overlay probes) | The three probe corpora are now permanent in-suite cells (TestLinkedParitySuppressions, duplicate_file_option, duplicate_import_last_occurrence) run green by this verifier | All three previously divergent probes now agree old-vs-new (old=0/new=0, old=2/new=2, line-6/line-6) | PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| LINT-PARITY-02 (REQ-3.2) | 04-01..04-08 (all eight declare it) | Rules whose behavior changed in main after the lab copy follow current-main semantics, not the lab's | ✓ SATISFIED | Main-wins pins verified (positions/regexps/compares/bug pins/presence predicate/first-exempt groups) AND the last main-semantics divergences are closed: suppression membership (CR-01), duplicate-option per-occurrence iteration (WR-02), duplicate-import last-wins position (WR-03) — each proven old-vs-new identical by my own runs of the pinning cells |

No orphaned requirements: REQUIREMENTS.md maps exactly LINT-PARITY-02 to Phase 4 (status Complete); every plan declares it. LINT-PARITY-01 (corpus parity) is Phase 7's requirement, deferred by design.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| internal/rules/parity_suppression_test.go | LINT-PARITY-02 | 3 cells | 0 | No | Value (old count asserted before exact multiset compare incl. positions+source) | VALID |
| internal/rules/parity_basic_semantic_test.go | LINT-PARITY-02 | 10 cells | 0 | No | Value (hand-derived old counts; fingerprint equality) | VALID |
| internal/rules/parity_harness_test.go (+ tiebreak) | LINT-PARITY-02 | harness + 1 pin | 0 | No | Value/Behavioral (old engine = byte-frozen main is the independent oracle) | VALID |
| internal/rules/parity_* (other 6 parents) | LINT-PARITY-02 | 39 cells | 0 | No | Value | VALID |
| internal/rules/registry_parity_guard_test.go / module_rules_guard_test.go | LINT-PARITY-02 | 2 guards | 0 | No | Value (41 names, exact D-03 set) | VALID |

**Disabled tests on requirements:** 0. **Circular patterns detected:** 0 (expected values come from the old engine — main's byte-frozen implementation, an independent oracle relative to the linked engine under test; old-side counts are hand-derived per cell and asserted before the compare). **Insufficient assertions:** 0.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| — | — | No debt markers (TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER), no placeholder text, no stub returns in any of the 16 gap-closure files (grep-checked) | — | Clean |

Note: `return nil, nil` in the ModuleRule ports' `Validate` is the documented contract shape (the engine calls Collect/Finish on ModuleRules and never Validate — lint_model.go contract doc + checkFile skip at lint_linked.go:187), not a stub. The D-01-kept helpers (incl. the now-unconsumed fileOption) are verbatim-unused BY DESIGN — documented dormancy with a gate-enforced Phase-5 deletion obligation (lint_helpers.go header + .golangci.yml comment), explicitly accepted by the prior necessity review.

### Human Verification Required

None. Infrastructure/foundation phase: a dormant engine port with no user-facing surface changed (dormancy grep/git-verified; the CLI flip is Phase 5's). Every behavioral claim is covered by a runnable test that this verifier executed; no `<verify><human-check>` blocks exist in any phase plan; no behavior-unverified truths remain.

### Gaps Summary

None. All six gaps from the previous report are closed and re-proven by this verifier's own runs:

1. **CR-01** — engine-side fix in collectFile (skip Collect for ignored rule/file pairs), mirroring main's shouldIgnore-before-Validate; the exempt-first-ignored corpus now yields old=0 AND new=0 (was old=0/new=1). All nine group ModuleRules ride the one engine fix; the Finish drain's path filter stays as defense in depth; both contradicted doc claims rewritten.
2. **WR-01** — assertBatchParity threads NewLinked's unwrapped ignoreOnly into the linked app; suppression parity is testable by construction; three suppression cells added (one CR-01 probe reproduction, two regression pins).
3. **WR-02** — per-occurrence option iteration in all seven option ports; duplicate_file_option old=2 AND new=2 at matching positions (was new=0); fileOption correctly demoted to kept-unused with mirrored comments.
4. **WR-03** — importInfo overwrite map keyed by ImportPath; duplicate imports report at the last written occurrence on both engines (was line 5 vs 6); the unsupported protoc-rejection justification deleted.
5. **WR-04** — SourceName as the fifth cmp.Or term in both the harness fingerprint and the engine's D-01 sort; tiebreak cell green.
6. **IN-01** — gofmt -l internal/rules/ empty.

The previously-flagged coincidental reliance (parity proven on a fixture grammar only) is hardened: the suite now declares and guards suppression state, duplicate inputs, and source totality — the parity contract is (path, rule, line, col, source) with Message equality separately proven by the carried per-rule tests (declared in the harness comment). The phase goal — all 41 registry rules on protoast with main's semantics, each batch verified old-vs-new before the next — is achieved at HEAD. Phase 5 (the flip) inherits a corrected, fully-parity-pinned engine.

---

_Verified: 2026-11-20T17:06:49Z_
_Verifier: Claude (gsd-verifier)_
