---
phase: 04-perenos-43-pravil-s-batch-paritetom
verified: 2026-11-20T16:11:01Z
status: gaps_found
score: 6/10 must-haves verified
covered_files: [".golangci.yml", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-01-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-01-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-02-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-02-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-03-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-03-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-04-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-04-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-05-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-05-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-06-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-06-SUMMARY.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-07-PLAN.md", ".planning/phases/04-perenos-43-pravil-s-batch-paritetom/04-07-SUMMARY.md", "internal/core/lint_linked_readsource_test.go", "internal/core/lint_model.go", "internal/core/position_probe_test.go", "internal/rules/builder_linked.go", "internal/rules/comment_enum_linked.go", "internal/rules/comment_enum_test.go", "internal/rules/comment_enum_value_linked.go", "internal/rules/comment_enum_value_test.go", "internal/rules/comment_field_message_linked.go", "internal/rules/comment_field_message_test.go", "internal/rules/comment_message_linked.go", "internal/rules/comment_message_test.go", "internal/rules/comment_one_of_linked.go", "internal/rules/comment_one_of_test.go", "internal/rules/comment_rpc_linked.go", "internal/rules/comment_rpc_test.go", "internal/rules/comment_service_linked.go", "internal/rules/comment_service_test.go", "internal/rules/directory_same_package_linked.go", "internal/rules/directory_same_package_test.go", "internal/rules/enum_first_value_zero_linked.go", "internal/rules/enum_first_value_zero_test.go", "internal/rules/enum_no_allow_alias_linked.go", "internal/rules/enum_no_allow_alias_test.go", "internal/rules/enum_pascal_case_linked.go", "internal/rules/enum_pascal_case_test.go", "internal/rules/enum_value_prefix_linked.go", "internal/rules/enum_value_prefix_test.go", "internal/rules/enum_value_upper_snake_case_linked.go", "internal/rules/enum_value_upper_snake_case_test.go", "internal/rules/enum_zero_value_suffix_linked.go", "internal/rules/enum_zero_value_suffix_test.go", "internal/rules/file_lower_snake_case_linked.go", "internal/rules/file_lower_snake_case_test.go", "internal/rules/import_no_public_linked.go", "internal/rules/import_no_public_test.go", "internal/rules/import_no_weak_linked.go", "internal/rules/import_no_weak_test.go", "internal/rules/import_used_linked.go", "internal/rules/import_used_test.go", "internal/rules/init_test.go", "internal/rules/lint_helpers.go", "internal/rules/message_field_lower_snake_case_linked.go", "internal/rules/message_field_lower_snake_case_test.go", "internal/rules/message_pascal_case_linked.go", "internal/rules/message_pascal_case_test.go", "internal/rules/module_rules_guard_test.go", "internal/rules/oneof_lower_snake_case_linked.go", "internal/rules/oneof_lower_snake_case_test.go", "internal/rules/package_defined_linked.go", "internal/rules/package_defined_test.go", "internal/rules/package_directory_match_linked.go", "internal/rules/package_directory_match_test.go", "internal/rules/package_lower_snake_case_linked.go", "internal/rules/package_lower_snake_case_test.go", "internal/rules/package_same_csharp_namespace_linked.go", "internal/rules/package_same_csharp_namespace_test.go", "internal/rules/package_same_directory_linked.go", "internal/rules/package_same_directory_test.go", "internal/rules/package_same_go_package_linked.go", "internal/rules/package_same_go_package_test.go", "internal/rules/package_same_java_multiple_files_linked.go", "internal/rules/package_same_java_multiple_files_test.go", "internal/rules/package_same_java_package_linked.go", "internal/rules/package_same_java_package_test.go", "internal/rules/package_same_php_namespace_linked.go", "internal/rules/package_same_php_namespace_test.go", "internal/rules/package_same_ruby_package_linked.go", "internal/rules/package_same_ruby_package_test.go", "internal/rules/package_same_swift_prefix_linked.go", "internal/rules/package_same_swift_prefix_test.go", "internal/rules/package_version_suffix_linked.go", "internal/rules/package_version_suffix_test.go", "internal/rules/parity_basic_casing_test.go", "internal/rules/parity_basic_semantic_test.go", "internal/rules/parity_comments_test.go", "internal/rules/parity_default_test.go", "internal/rules/parity_full_registry_test.go", "internal/rules/parity_harness_test.go", "internal/rules/parity_minimal_test.go", "internal/rules/parity_unary_rpc_test.go", "internal/rules/parity_wave0_test.go", "internal/rules/registry_parity_guard_test.go", "internal/rules/rpc_no_client_streaming_linked.go", "internal/rules/rpc_no_client_streaming_test.go", "internal/rules/rpc_no_server_streaming_linked.go", "internal/rules/rpc_no_server_streaming_test.go", "internal/rules/rpc_pascal_case_linked.go", "internal/rules/rpc_pascal_case_test.go", "internal/rules/rpc_request_response_unique_linked.go", "internal/rules/rpc_request_response_unique_test.go", "internal/rules/rpc_request_standard_name_linked.go", "internal/rules/rpc_request_standard_name_test.go", "internal/rules/rpc_response_standard_name_linked.go", "internal/rules/rpc_response_standard_name_test.go", "internal/rules/service_pascal_case_linked.go", "internal/rules/service_pascal_case_test.go", "internal/rules/service_suffix_linked.go", "internal/rules/service_suffix_test.go"]
covered_digest: "v4:sha256:f54135493f81a0325c80a54b7c950bcaac99b885952a1b9754d4e4e0bc476768"
behavior_unverified: 0
overrides_applied: 0
gaps:
  - truth: "Suppression parity with main: ignore_only behaves identically to main (REQ-2.5 semantics carried into the linked engine)"
    status: failed
    reason: "CR-01 (probe re-proven by verifier, go test -overlay TestScratchIgnoreOnlyGroupParityProbe): main's engine skips Validate for a file listed in ignore_only, so an ignored file never seeds a group rule's exempt-first value; the linked engine's collectFile hands EVERY file to ModuleRule.Collect and only filters findings by path afterwards. On a clean two-file corpus with the exempt-first file ignored: old=0 findings, new=1 finding at opts/v1/beta.proto:5:1. Affects all nine group ModuleRules. Violates the phase goal's 'main wins' semantics and the suppression-parity must-have."
    artifacts:
      - path: "internal/core/lint_linked.go"
        issue: "collectFile calls ModuleRule.Collect for every file regardless of shouldIgnoreLinked; ignore_only applied only to findings by path in the drain filter"
      - path: "internal/rules/package_same_go_package_linked.go"
        issue: "one of the nine affected group ports (Collect/Finish design assumes complete groups)"
      - path: "internal/core/lint_model.go"
        issue: "ModuleRule doc claims 'suppressions and ignore_only apply to the findings by path' — contradicted by main's membership behavior"
    missing:
      - "Mirror main's shouldIgnore-before-Validate at collection time: skip moduleRule.Collect for files ignored for that rule (per 04-REVIEW CR-01 fix sketch)"
      - "Rewrite the collectFile 'group stays complete' comment and the lint_model.go suppression claim to the corrected semantics"
      - "Add parity cells: exempt-first file ignored (currently red) and mismatching file ignored (green regression pin)"
  - truth: "The parity harness compares identical suppression state old-vs-new (IgnoreOnly wired into both engines)"
    status: failed
    reason: "WR-01: assertBatchParity discards NewLinked's second return value (`newRules, _, err := rules.NewLinked(cfg)`) and never sets IgnoreOnly on the linked app, so even a parity cell configuring ignore_only would compare nil-vs-set suppression state. This is the structural blind spot that let CR-01 through 39 green parity cells."
    artifacts:
      - path: "internal/rules/parity_harness_test.go"
        issue: "newRules, _, err := rules.NewLinked(cfg) discards newIgnoreOnly; core.New(Options{LinkedRules: ...}) without IgnoreOnly"
    missing:
      - "Thread newIgnoreOnly into the linked app (or add assertBatchParityWithIgnoreOnly variant) + the three suppression cells (group exempt-first ignored, mismatching file ignored, per-file rule ignored)"
  - truth: "Duplicate file-level option occurrences behave as main (every matching occurrence iterated; first seeds, later occurrences compared)"
    status: failed
    reason: "WR-02 (probe re-proven: old=2 findings, new=0): all seven option ModuleRules use fileOption(info.File, name), which is last-wins over file.Options; main's Validate loops over every matching occurrence (first seeds the per-package cache, every later occurrence compared). Both engines accept the corpus (no COMPILE issues), so 'protoc rejects duplicates' is not a shield."
    artifacts:
      - path: "internal/rules/lint_helpers.go"
        issue: "fileOption overwrites `found` per match — returns the LAST matching option"
      - path: "internal/rules/package_same_go_package_linked.go"
        issue: "Collect uses fileOption (line 50) — first-occurrence seeding and per-occurrence comparisons lost"
      - path: "internal/rules/package_same_java_package_linked.go"
        issue: "same fileOption pattern (line 50)"
      - path: "internal/rules/package_same_java_multiple_files_linked.go"
        issue: "same fileOption pattern (line 50)"
      - path: "internal/rules/package_same_csharp_namespace_linked.go"
        issue: "same fileOption pattern (line 51)"
      - path: "internal/rules/package_same_php_namespace_linked.go"
        issue: "same fileOption pattern (line 50)"
      - path: "internal/rules/package_same_ruby_package_linked.go"
        issue: "same fileOption pattern (line 50)"
      - path: "internal/rules/package_same_swift_prefix_linked.go"
        issue: "same fileOption pattern (line 67)"
    missing:
      - "Replace fileOption with main's per-occurrence iteration in all seven option ports' Collect; firstExemptMismatches then reproduces first-seed/rest-compare"
      - "Add a duplicate-option parity cell (expect old=2 on the probe corpus)"
  - truth: "Duplicate import statements report at main's position (last written occurrence)"
    status: failed
    reason: "WR-03 (probe re-proven: old reports line 6, linked line 5 — same count, different position, fingerprint mismatch on an input both engines accept): main builds importInfo by overwriting per occurrence (last wins); the port iterates info.File.Imports in source order with a reported dedup set (first wins). The port's justifying comment ('rejected by protoc') is factually unsupported and immaterial."
    artifacts:
      - path: "internal/rules/import_used_linked.go"
        issue: "report loop reports first occurrence (reported map); comment at lines 36-40 claims protoc rejects duplicates"
    missing:
      - "Mirror main's last-wins selection (importInfo map overwrite semantics) and correct the :36-40 comment"
      - "Add a duplicate-import parity cell (expect the last-occurrence position on both sides)"
  - truth: "The parity fingerprint covers every user-visible field, including SourceName"
    status: failed
    reason: "WR-04: fingerprint sorts and compares (path, rule, line, col) only; SourceName is user-visible in both output printers (internal/api/lint.go text and JSON) and carries semantic payload (written option constant, declared names, import location). No divergence exists today (review verified all 41 ports pass main's exact source), but a future source-argument drift would pass all parity cells while changing CLI output."
    artifacts:
      - path: "internal/rules/parity_harness_test.go"
        issue: "fingerprint tuple lacks SourceName (sort key and compare)"
    missing:
      - "Add SourceName to the fingerprint tuple and the engine D-01 sort key; update the harness contract comment"
  - truth: "Repo formatting convention holds on all phase files (gofmt)"
    status: failed
    reason: "IN-01 (re-proven: gofmt -l flags internal/rules/registry_parity_guard_test.go): the requireConfigParity literal uses two extra alignment spaces. golangci (staticcheck-only) does not fail on it, so task lint stays green, but gofmt is the repository convention (.agents/skills/go-code-style)."
    artifacts:
      - path: "internal/rules/registry_parity_guard_test.go"
        issue: "gofmt -l flags the file (requireConfigParity literal alignment)"
    missing:
      - "gofmt -w internal/rules/registry_parity_guard_test.go (as part of the CR-01 fix commit)"
deferred:
  - truth: "Resolution-based ImportUsed (lab validateResolved) is excluded"
    addressed_in: "Phase 6/7"
    evidence: "Ledger item 1 (04-07-SUMMARY Consolidated Deferred-Decisions Ledger): deliberate post-parity behavior change, owned by Phase 6/7 with the syntax_walk_only cell rewritten in the same commit"
  - truth: "Cross-file RPCRequestResponseUnique (lab shape) is excluded; main's per-file first-exempt shipped"
    addressed_in: "Phase 6/7"
    evidence: "Ledger item 2: deliberate post-parity behavior change owner Phase 6/7"
  - truth: "swift_prefix ruby_package bug is pinned bug-for-bug, not fixed"
    addressed_in: "Phase 6/7"
    evidence: "Ledger item 3: corrected behavior would break Phase-7 corpus parity; fix decision owned post-parity with pins rewritten in the same commit"
  - truth: "Lab casing helpers (casing.go acronym-aware algorithms) not adopted"
    addressed_in: "Phase 5"
    evidence: "Ledger item 5 + Phase 5 goal 'старый lint-контур удалён': Phase 5 deletes casing.go with the old path and collapses pascalToUpperSnakeLinked to the canonical name"
  - truth: "Lab report-everyone groupedValues.mismatches stays verbatim-but-unreachable"
    addressed_in: "Phase 5"
    evidence: "Ledger item 6: Phase 5 deletes it with the old path"
  - truth: "Old-path consolidation (old rule files, builder.go, D-01-kept helpers, .golangci.yml exclusion) not done in Phase 4"
    addressed_in: "Phase 5"
    evidence: "Ledger items 14-15 + Phase 5 goal: old lint contour removal and buildCore/LinkedRules production wiring owned by Phase 5 (dormancy is a deliberate Phase-4 state, grep/git-diff verified)"
  - truth: "PACKAGE_NO_IMPORT_CYCLE not a ModuleRule yet (absence pinned by TestModuleRuleSet)"
    addressed_in: "Phase 6"
    evidence: "Ledger item 17 + Phase 6 goal 'PACKAGE_NO_IMPORT_CYCLE в MINIMAL': Phase 6 adds it as a ModuleRule and updates the guard in the same commit"
  - truth: "Adjacent option string literals ('a' 'b' merge) edge unprobed"
    addressed_in: "Phase 7"
    evidence: "Ledger item 13: observation note; Phase 7 corpus runs own the finding"
coincidental_reliance_items:
  - truth: "All 41 rules produce identical sorted (rule, path, line, col) multisets old-vs-new on the parity corpora"
    reason: fixture-only
    harden: "Parity holds on the harness fixture grammar only; inputs outside it (ignore_only configs, duplicate options/imports — the probe corpora) diverge. The gap-closure cells (suppression cells via WR-01 fix, duplicate-option cell, duplicate-import cell) plus the SourceName fingerprint extension are exactly the hardening that promotes the hidden fixture-scope assumption into a declared, guarded precondition."
---

# Phase 04: Перенос 43 правил с батч-паритетом — Verification Report

**Phase Goal:** Все 43 текущих правила переписаны на protoast с семантикой текущего main (main побеждает лабораторную копию); каждый батч переноса верифицирован паритетом старое-vs-новое до перехода к следующему (registry truth: 41 rules — Pitfall 7)
**Verified:** 2026-11-20T16:11:01Z
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

The port itself is real and high quality: 41/41 rules registered with canonical names, carried tests on every rule, position discipline, dormancy, and the full parity harness green on its corpora — all independently re-verified here (own test runs, own greps, own probe re-runs). **But the phase's core contract is "main wins" parity, and one probe-proven behavioral divergence from main survives (CR-01: ignore_only changes group membership in main but not in the linked engine), plus three parity holes in duplicate-input semantics and two harness-contract gaps that structurally hide exactly this class of divergence.** The code review (04-REVIEW.md, disposition 04-REVIEW-DISPOSITION.md) routed all six findings to gap closure this phase; none are fixed at HEAD. The verdict is honestly `gaps_found`.

### Observable Truths

| #   | Truth   | Status     | Evidence       |
| --- | ------- | ---------- | -------------- |
| 1   | All 41 rules produce identical sorted (rule, path, line, col) multisets old-vs-new on the parity corpora; per-batch parity verified cumulatively (SC1, corpus-scoped) | ✓ VERIFIED (coincidental-reliance) | Own run: `go test -race -count=1 ./internal/rules/ ./internal/core/` green; TestLinkedParity* = 8 parents + 39 cells; TestRegistryParity 41/41 name parity. Holds on the harness fixture grammar; inputs outside it diverge (see coincidental_reliance_items and truths 6-9) |
| 2   | Batch-parity process: each batch gated RED-before-GREEN and re-verified cumulatively before the next | ✓ VERIFIED | tdd-red-evidence-04-0{2..7}-task1.json (RED suites naming the unported rules, old-side counts probe-confirmed); VALIDATION.md per-batch green rows |
| 3   | Registry truth: 41/41 rules in NewLinked under canonical names; names/groups/config plumbing/ignoreOnly unwrap parity with rules.New; exact D-03 ModuleRule set (10) | ✓ VERIFIED | 41 `&...Linked{` registrations in builder_linked.go; own run: TestRegistryParity + TestModuleRuleSet PASS; grep: no `_LINKED` suffix in non-comment code |
| 4   | Main-wins semantics for changed-in-main behavior (SC2): keyword Decl.Pos positions, main regexps, written-constant compares, ruby_package bug pin, ImportUsed stateless walk, comment presence predicate, first-exempt groups, packageless invisibility | ✓ VERIFIED | Own greps: zero NamePos/NumberPos/FilePosition code hits in 41 ports (2 comment-only hits); zero suppression-state references in ports (central filter); carried tests green in own suite run; deep-diff cross-check against 04-REVIEW verified-clean checklist |
| 5   | Every ported rule carries a colocated `<rule>_test.go` with main's case sets (SC3); task test/lint-equivalents green at HEAD | ✓ VERIFIED | 41/41 colocated test files present (enumeration loop, zero missing); own runs: go build ./... exit 0, go vet clean, -race suites green, `bin/golangci-lint run ./internal/rules/... ./internal/core/...` = 0 issues, schema-gen drift-empty |
| 6   | Suppression parity with main: ignore_only/comments behave identically to main (REQ-2.5 semantics into the linked engine) | ✗ FAILED | CR-01 re-proven by own probe run (`go test -overlay`, TestScratchIgnoreOnlyGroupParityProbe): old=0 findings, new=1 at opts/v1/beta.proto:5:1 with the exempt-first file in ignore_only. Linked engine collects ignored files into groups; main never lets them reach Validate. Affects all nine group ModuleRules |
| 7   | Parity harness compares identical suppression state (IgnoreOnly wired into both sides) | ✗ FAILED | parity_harness_test.go: `newRules, _, err := rules.NewLinked(cfg)` — newIgnoreOnly discarded; linked app runs with nil IgnoreOnly (read of source, line ~77) |
| 8   | Duplicate-input parity on inputs both engines accept: file-level options per-occurrence (main), duplicate imports last-wins (main) | ✗ FAILED | Own probe re-runs: TestScratchDuplicateOptionProbe old=2/new=0; TestScratchDuplicateImportProbe old line 6 / new line 5. fileOption is last-wins (lint_helpers.go:50-58); import_used report loop reports first occurrence |
| 9   | Parity fingerprint covers every user-visible field incl. SourceName | ✗ FAILED | fingerprint sorts (path, rule, line, col) only — SourceName absent (read of source); SourceName is user-visible in internal/api/lint.go printers |
| 10  | Dormancy: easyp lint unchanged, old path byte-frozen, LinkedRules test-only; schemas drift-free | ✓ VERIFIED | buildCore (internal/api/runtime.go) uses rules.New/Rules — LinkedRules never set in production wiring; `git log` on old-path files (core/lint.go, builder.go, old rule files) shows last touch in Phase-1 commits; schema-gen produced zero diffs |

**Score:** 6/10 truths verified (0 present, behavior-unverified)

The four failed truths share one root cause class: parity was proven on a fixture grammar that cannot express suppression state (WR-01) and does not include duplicate-occurrence inputs, and the linked engine's collect-everything design diverges from main's membership semantics exactly there (CR-01). WR-02/WR-03 are the same class surfacing in per-file rules.

### Deferred Items

Items not yet met but explicitly addressed in later milestone phases (frontmatter `deferred:` — from the 04-07-SUMMARY Consolidated Deferred-Decisions Ledger, matched against Phase 5/6/7 roadmap goals). These are **deliberate main-wins or lab-exclusion decisions**, not gaps: the shipped behavior IS main's; the deferred item is the lab alternative or the post-parity change.

| # | Item | Addressed In | Evidence |
|---|------|-------------|----------|
| 1 | Resolution-based ImportUsed (lab validateResolved) | Phase 6/7 | Ledger item 1; Phase 7 goal (corpus parity gates) |
| 2 | Cross-file RPCRequestResponseUnique (lab shape) | Phase 6/7 | Ledger item 2 |
| 3 | swift_prefix ruby_package bug fix | Phase 6/7 | Ledger item 3; would break Phase-7 corpus parity if applied now |
| 4 | casing.go / lab helper deletion, old-path consolidation, exclusion removal | Phase 5 | Ledger items 5/6/14 + Phase 5 goal "старый lint-контур удалён" |
| 5 | buildCore LinkedRules production wiring (dormancy is deliberate) | Phase 5 | Ledger item 15 + Phase 5 goal (flip + COMPAT gate) |
| 6 | PACKAGE_NO_IMPORT_CYCLE as ModuleRule | Phase 6 | Ledger item 17 + Phase 6 goal |
| 7 | Adjacent option string literals edge | Phase 7 | Ledger item 13; corpus runs own it |

### Required Artifacts

| Artifact | Expected    | Status | Details |
| -------- | ----------- | ------ | ------- |
| `internal/core/lint_model.go` | Linked seam (LinkedFileInfo, AppendLinkedIssue, ModuleRule contract) | ✓ VERIFIED | Wired by all 41 ports + engine; suppression doc claim needs correction with CR-01 fix |
| `internal/rules/builder_linked.go` | NewLinked registry mirroring New | ✓ VERIFIED | 41 rules, builder.go group order, TestRegistryParity green |
| `internal/rules/<rule>_linked.go` (41) | Ported rules with main semantics | ✓ VERIFIED for parity surface / ✗ FAILED for 9 group ports (ignore_only membership, CR-01) + 7 option ports (duplicate options, WR-02) + import_used (duplicate imports, WR-03) | See gaps |
| `internal/rules/lint_helpers.go` | Shared Collect/Finish helpers on the linked contract | ✓ VERIFIED with issue | firstExemptMismatches wired from all 9 group ports; fileOption last-wins is WR-02's mechanism |
| `internal/rules/parity_harness_test.go` | Old-vs-new sorted multiset compare harness | ✓ exists / ✗ contract gaps | WR-01 (IgnoreOnly discarded), WR-04 (SourceName absent) |
| `internal/rules/<rule>_test.go` (41) | Carried main tests on the linked contract | ✓ VERIFIED | 41/41 present, green in own suite run |
| `internal/core/position_probe_test.go` | Decl.Pos == Meta.Pos pin | ✓ VERIFIED | Green in own run (TestDeclPositionProbe) |
| `internal/core/lint_linked_readsource_test.go` | readSource-once cell | ✓ VERIFIED | Green in own run (TestReadSourceOnce) |
| `internal/rules/module_rules_guard_test.go` / `registry_parity_guard_test.go` | Exact-set + registry-parity guards | ✓ VERIFIED with issue | Both green in own run; registry_parity_guard_test.go fails gofmt (IN-01) |
| `.golangci.yml` | Phase-4 exclusion comment resolution | ✓ VERIFIED | Comment records wired helpers vs D-01-kept helpers and the Phase-5 deletion obligation |

### Key Link Verification

| From | To  | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| `internal/rules/builder_linked.go` | `internal/core/core.go` | NewLinked -> Options.LinkedRules (test-only) | ✓ WIRED | Dormancy verified: production buildCore uses rules.New; linked rules consumed by core.LintLinked in tests |
| `internal/rules/*_linked.go` (41) | `internal/core/lint_model.go` | AppendLinkedIssue sole construction point | ✓ WIRED | Direct IssueInfo wrap only in the ModuleRule Finish contract shape (review-verified, spot-checked) |
| `internal/rules/parity_harness_test.go` | `internal/core` | rules.New+Lint vs rules.NewLinked+LintLinked multiset compare | ✓ WIRED / ✗ contract gap | Harness runs both engines (39 cells green) but drops ignoreOnly (WR-01) and SourceName (WR-04) |
| `internal/rules/*_linked.go` (group ports) | `internal/rules/lint_helpers.go` | Collect -> groupedValues.add; Finish -> firstExemptMismatches | ✓ WIRED | Mechanism correct for the no-suppression case; membership-vs-main wrong under ignore_only (CR-01) |
| `internal/rules/comment_*_linked.go` (7) | `internal/protoast` | HasCommentAbove() presence predicate | ✓ WIRED | Directive-only-documents cell green; zero leadingCommentDocuments references |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| All 41 ports | Decl positions/names/values | protoast.File (linked easypc trees) | Yes — no static returns, no mocked model | ✓ FLOWING |
| Parity harness | issues old/new | Both engines on a real temp-dir corpus walked by fs.NewFSWalker | Yes — old-side counts asserted before compare (anti-vacuity) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Full rules+core suite green | `go test -race -count=1 ./internal/rules/ ./internal/core/` | ok (both packages) | ✓ PASS |
| Build green | `go build ./...` | exit 0 | ✓ PASS |
| Vet clean | `go vet ./internal/rules/ ./internal/core/` | clean | ✓ PASS |
| Lint clean | `bin/golangci-lint run ./internal/rules/... ./internal/core/...` | 0 issues | ✓ PASS |
| gofmt convention | `gofmt -l internal/rules/` | flags registry_parity_guard_test.go | ✗ FAIL (IN-01) |
| Schema drift | `go run ./cmd/easyp schema-gen` | zero working-tree diffs in schemas/ | ✓ PASS |
| Registry + exact-set guards | `go test -run 'TestRegistryParity\|TestModuleRuleSet' -v ./internal/rules/` | both PASS | ✓ PASS |
| CR-01 ignore_only group membership | `go test -overlay /tmp/scratch-probe/overlay.json -run TestScratchIgnoreOnlyGroupParityProbe` | old=0, new=1 (opts/v1/beta.proto:5:1) | ✗ DIVERGENT → gap |
| WR-02 duplicate option | `go test -overlay ... -run TestScratchDuplicateOptionProbe` | old=2, new=0 | ✗ DIVERGENT → gap |
| WR-03 duplicate import position | `go test -overlay ... -run TestScratchDuplicateImportProbe` | old line 6, new line 5 | ✗ DIVERGENT → gap |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| Verifier re-run of 04-REVIEW probe (CR-01 + WR-02 + WR-03) | `go test -race -count=1 -overlay /tmp/scratch-probe/overlay.json -run 'TestScratch(IgnoreOnlyGroupParity\|DuplicateImport\|DuplicateOption)Probe' ./internal/rules/` | All three probes reproduce the review's numbers exactly; no repository file modified (overlay injection) | PASS (probes ran; outputs prove the gaps) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| LINT-PARITY-02 (REQ-3.2) | 04-01..04-07 (all seven declare it) | Rules whose behavior changed in main after the lab copy follow current-main semantics, not the lab's | ✗ BLOCKED (gap closure required) | Main-wins pins verified for positions/regexps/compares/bug pins/presence predicate (truth 4); BUT CR-01 is a main-semantics violation on the suppression/group interplay, and WR-02/WR-03 are main-semantics divergences on duplicate inputs — the shipped engine does not fully follow main where it matters most for the phase contract. Fixable set routed by 04-REVIEW-DISPOSITION.md |

No orphaned requirements: REQUIREMENTS.md maps exactly LINT-PARITY-02 to Phase 4; every plan declares it. (REQ-2.5/LINT-ENG-05 is a Phase-3 requirement whose semantics the CR-01 fix restores for the linked path.)

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| internal/rules/parity_harness_test.go | ~77 | Discarded return value (newIgnoreOnly) — suppression state untestable | 🛑 Blocker (harness gap that hid CR-01) | Parity contract blind spot |
| internal/rules/parity_harness_test.go | 97-109 | Fingerprint tuple missing user-visible field (SourceName) | ⚠️ Warning | Source drift would pass all cells |
| internal/core/lint_linked.go | collectFile drain | Collect-everything group membership diverging from main under ignore_only | 🛑 Blocker (CR-01, probe-proven) | Wrong user-visible findings for all nine group rules once flipped |
| internal/rules/lint_helpers.go | 50-58 | fileOption last-wins collapse vs main's per-occurrence loop | ⚠️ Warning (WR-02, probe-proven old=2/new=0) | Missing findings on duplicate-option inputs |
| internal/rules/import_used_linked.go | 99-110 (+36-40 comment) | First-occurrence report + unsupported protoc-rejection justification | ⚠️ Warning (WR-03, probe-proven line 6 vs 5) | Position mismatch on accepted inputs |
| internal/rules/registry_parity_guard_test.go | 89-93 | gofmt-unformatted literal | ⚠️ Warning (IN-01; golangci green, repo convention violated) | Minor; fix with CR-01 commit |

No TBD/FIXME/XXX debt markers in any phase file; no placeholder/stub returns in the 41 ports (grep-checked).

### Human Verification Required

None. The phase is a dormant engine port: no user-visible surface changed (dormancy grep/git-verified), every behavioral claim is covered by a runnable test or probe, and the divergences are probe-reproducible without human judgment. The gap-closure fix set (CR-01 + WR-01..04 + IN-01) is deterministic engineering work, not a human-check item.

### Gaps Summary

One coherent, fixable set — all six items routed to gap closure by 04-REVIEW-DISPOSITION.md and still open at HEAD (ad24bcf/c9b64b4 are docs-only commits):

1. **CR-01 (critical, probe re-proven):** the linked engine's collect-everything ModuleRule design changes group membership under `ignore_only` — main never lets an ignored file seed the exempt-first value, the linked engine does, so old=0/new=1 findings on a clean corpus. Violates the phase goal's "main wins" semantics (and REQ-2.5 semantics once flipped). Fix: skip `moduleRule.Collect` for files ignored for that rule; rewrite the two contradictory doc claims; add the ignore_only parity cells.
2. **WR-01 (harness gap that hid CR-01):** assertBatchParity discards `newIgnoreOnly` — thread it through and add the three suppression cells; without this, suppression parity stays untestable by construction.
3. **WR-02 (probe re-proven old=2/new=0):** seven option ports collapse duplicate file-level options via last-wins `fileOption`; main iterates every occurrence. Fix per-occurrence iteration + parity cell.
4. **WR-03 (probe re-proven line 6 vs 5):** import_used reports the first duplicate-import occurrence; main reports the last. Fix last-wins selection + correct the unsupported comment + parity cell.
5. **WR-04:** fingerprint omits SourceName (user-visible in both printers). Add to tuple and D-01 sort key.
6. **IN-01:** gofmt the guard test file.

Everything else the phase claims holds under independent re-verification: 41/41 registry with carried tests, batch-gate process with RED evidence, position discipline, suppression centrality, dormancy, byte-frozen old path, green build/vet/lint/schema gates. Group the six fixes into one gap-closure plan/commit set on this phase, then re-verify: the acceptance signal is the new ignore_only parity cell red-before-fix/green-after, the duplicate-input cells green, and the existing 47-cell suite still green.

---

_Verified: 2026-11-20T16:11:01Z_
_Verifier: Claude (gsd-verifier)_
