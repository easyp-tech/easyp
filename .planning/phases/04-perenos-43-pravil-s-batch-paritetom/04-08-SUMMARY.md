---
phase: 04-perenos-43-pravil-s-batch-paritetom
plan: 08
subsystem: rules
tags: [lint-rules, linked-engine, parity-harness, ignore-only, suppression-parity, duplicate-inputs, fingerprint-totality, gap-closure, tdd]

# Dependency graph
requires:
  - phase: 04-perenos-43-pravil-s-batch-paritetom (plans 01-07)
    provides: the dormant linked engine + 41/41 registry, the parity harness (requirePorted/assertBatchParity/fingerprint), the 47-cell parity suite, and 04-VERIFICATION's six probe-proven gaps (CR-01, WR-01..04, IN-01) with 04-REVIEW's fix sketches
provides:
  - CR-01 closed: collectFile skips ModuleRule.Collect for files ignored for that rule — group membership under ignore_only matches main's shouldIgnore-before-Validate on all nine group ModuleRules; old=0 AND new=0 on the exempt-first-ignored corpus
  - WR-01 closed: assertBatchParity threads NewLinked's unwrapped ignoreOnly into core.Options.IgnoreOnly — the harness can no longer compare nil-vs-set suppression state; three suppression parity cells (first_exempt_file_ignored, mismatching_file_ignored, per_file_rule_ignored) added as TestLinkedParitySuppressions
  - WR-02 closed: per-occurrence file-option iteration in all seven option ports (main's first-seed/rest-compare); duplicate_file_option cell old=2 AND new=2 at matching positions; fileOption is kept-unused (nine D-01-kept helpers)
  - WR-03 closed: duplicate imports report at the LAST written occurrence (importInfo overwrite map keyed by ImportPath); duplicate_import_last_occurrence cell line-6 parity; unsupported protoc-rejection comment deleted
  - WR-04 closed: SourceName as the fifth cmp.Or term in both the harness fingerprint and the engine's D-01 sort — the parity contract is (path, rule, line, col, source), total through the user-visible field; TestFingerprintSourceNameTiebreak
  - IN-01 closed: gofmt -l internal/rules/ empty (registry_parity_guard_test.go literal collapsed)
  - three machine-validated RED evidence records (tdd-red-evidence-04-08-task{1,2,3}.json)
affects: [Phase 5 flip (the fixed engine is what flips; COMPAT gate inherits the suppression cells), Phase 6 (exact-set ModuleRule guard unchanged), Phase 7 corpus gates (the parity contract now includes source; duplicate-input corpora are declared parity surfaces)]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 10426    # 41704 diff chars over internal/ + .golangci.yml at chars/4 (estimate was 32000; comment rewrites + test cells dominate, engine changes are small)
  tasks: 4
  commits: 7      # MEASURED: git rev-list --count 8a42a8f..HEAD (3 RED + 3 GREEN + 1 gofmt chore); excludes the metadata docs commit
  plan_head_before: 8a42a8fa6bd406507fcc77d75e47a8d01f23aa50
  plan_head_after: fe11d9d30a33abe2e0755f71419d1958b360f3b1

# Tech tracking
tech-stack:
  added: []
  patterns: [probe-cell-in-suite (04-REVIEW scratch probes reproduced as permanent parity cells with in-file provenance headers), harness suppression-state threading (both engines driven with NewLinked's unwrapped ignoreOnly), per-occurrence option iteration (first-seed/rest-compare via groupedValues.firstExemptMismatches), overwrite-map last-wins selection (importInfo keyed by ImportPath), five-term total sort ((path, rule, line, col, source) in fingerprint AND engine D-01)]

key-files:
  created:
    - internal/rules/parity_suppression_test.go
    - .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-08-task1.json
    - .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-08-task2.json
    - .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-08-task3.json
  modified:
    - internal/core/lint_linked.go
    - internal/core/lint_model.go
    - internal/rules/parity_harness_test.go
    - internal/rules/parity_basic_semantic_test.go
    - internal/rules/package_same_go_package_linked.go
    - internal/rules/package_same_java_package_linked.go
    - internal/rules/package_same_java_multiple_files_linked.go
    - internal/rules/package_same_csharp_namespace_linked.go
    - internal/rules/package_same_php_namespace_linked.go
    - internal/rules/package_same_ruby_package_linked.go
    - internal/rules/package_same_swift_prefix_linked.go
    - internal/rules/import_used_linked.go
    - internal/rules/lint_helpers.go
    - internal/rules/registry_parity_guard_test.go
    - .golangci.yml

key-decisions:
  - "CR-01 fixed in the ENGINE, not the rules: collectFile (internal/core/lint_linked.go) skips moduleRule.Collect when shouldIgnoreLinked(rule, info.Path) — mirroring main's shouldIgnore-before-Validate at core/lint.go:48 — so all nine group ModuleRules get correct membership from one engine change; the Finish drain's path filter stays as defense in depth (and for the per-file RPC_REQUEST_RESPONSE_UNIQUE findings)"
  - "WR-01 fixed by threading, not a variant: assertBatchParity captures NewLinked's second return value and sets core.Options.IgnoreOnly, so every existing cell gets identical unwrapped maps (nil for no-ignore_only configs) and suppression parity is testable by construction — no assertBatchParityWithIgnoreOnly fork"
  - "WR-02 keeps the empty-loop semantics: a file without the compared option adds nothing (main's skip-optionless survives as the loop that never matches); the swift-prefix port reads the same ruby_package literal per occurrence — the 04-08 sweep changes only how many occurrences are read, never which option"
  - "fileOption moves to the kept-unused list (nine helpers) with doc reworded away from the gated 'returns the last file-level option called name' phrase; header comment list and .golangci.yml exclusion comment updated in mirror so the documented state matches reality; deletion stays a Phase-5 obligation"
  - "WR-03 reports through the overwrite map (importInfo[importPath] = imp per occurrence) instead of a reported dedup set; map iteration order is normalized by the engine's D-01 sort (now total through SourceName), so each unused path reports exactly once at its LAST written occurrence"
  - "The two regression-pin cells (mismatching_file_ignored, per_file_rule_ignored) were expected red pre-fix through the WR-01 nil-vs-set gap — documented in the task-1 RED evidence as anticipated, not as unplanned failures"

patterns-established:
  - "Probe-provenance headers: each new cell's file records the 04-REVIEW gap id, the verifier's go test -overlay re-proof, and the old/new numbers, so the suite carries its own evidence trail"
  - "Suppression-state parity cells: every new ignore_only behavior gets a parity cell through the threaded harness — never a rule-local suppression path"

requirements-completed: [LINT-PARITY-02]

# Coverage metadata (#1602) — one entry per shipped deliverable.
coverage:
  - id: D1
    description: "CR-01 engine membership fix: collectFile skips ModuleRule.Collect for files ignored for that rule (mirrors main's shouldIgnore-before-Validate); on the exempt-first-ignored corpus old=0 AND new=0 (pre-fix new=1 at opts/v1/beta.proto:5:1)"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_suppression_test.go#TestLinkedParitySuppressions/first_exempt_file_ignored"
        status: pass
      - kind: other
        ref: "command: gsd_run check tdd-red-evidence .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-08-task1.json -> RED_EVIDENCE_OK (old len 0 vs new 1 at opts/v1/beta.proto:5:1, target_test_failed)"
        status: pass
    human_judgment: false
  - id: D2
    description: "WR-01 harness threading: assertBatchParity feeds NewLinked's unwrapped ignoreOnly into core.Options.IgnoreOnly; both engines run identical suppression state by construction; mismatching_file_ignored and per_file_rule_ignored are green regression pins after the fix"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_suppression_test.go#TestLinkedParitySuppressions/mismatching_file_ignored"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_suppression_test.go#TestLinkedParitySuppressions/per_file_rule_ignored"
        status: pass
      - kind: other
        ref: "command: grep -c 'newRules, _, err' internal/rules/parity_harness_test.go -> 0"
        status: pass
    human_judgment: false
  - id: D3
    description: "WR-04 SourceName totality: fifth cmp.Or term in the harness fingerprint AND the engine's D-01 sort; ties on (path, rule, line, col) order by SourceName; the parity contract comment updated to (path, rule, line, col, source) with Message deliberately outside"
    requirement: LINT-PARITY-02
    verification:
      - kind: unit
        ref: "tests/internal/rules/parity_harness_test.go#TestFingerprintSourceNameTiebreak"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_suppression_test.go#TestLinkedParitySuppressions (engine sort extension regressions nothing)"
        status: pass
    human_judgment: false
  - id: D4
    description: "WR-02 per-occurrence file-option iteration in all seven option ports; duplicate_file_option cell old=2 AND new=2 at matching positions (pre-fix new=0); optionless files contribute nothing; swift-prefix bug pin (ruby_package literal + BUG PIN comments) verbatim"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_basic_semantic_test.go#TestLinkedParityBasicSemantic/duplicate_file_option"
        status: pass
      - kind: other
        ref: "command: gsd_run check tdd-red-evidence .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-08-task2.json -> RED_EVIDENCE_OK (old=2 vs new=0); grep -lc 'fileOption(info.File' internal/rules/*_linked.go -> 0 files"
        status: pass
    human_judgment: false
  - id: D5
    description: "WR-03 last-wins duplicate-import reporting: importInfo overwrite map keyed by core.ImportPath, each unused path reports once at its LAST written occurrence; duplicate_import_last_occurrence cell line-6 parity (pre-fix new line 5); protoc-rejection justification deleted from the comments"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_basic_semantic_test.go#TestLinkedParityBasicSemantic/duplicate_import_last_occurrence"
        status: pass
      - kind: other
        ref: "command: gsd_run check tdd-red-evidence .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-08-task3.json -> RED_EVIDENCE_OK (line 6 vs line 5); grep -c 'rejected by protoc' internal/rules/import_used_linked.go -> 0"
        status: pass
    human_judgment: false
  - id: D6
    description: "IN-01 gofmt closure: gofmt -l internal/rules/ empty (registry_parity_guard_test.go requireConfigParity literal collapsed, no semantic change)"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "command: test -z \"$(gofmt -l internal/rules/)\" -> pass"
        status: pass
    human_judgment: false
  - id: D7
    description: "Gap-closure gates at the fixed state: full rules+core -race suite; TestRegistryParity + TestModuleRuleSet guards; TestReadSourceOnce (constraint a) + TestLintLinkedSuppressionParity (constraint b — centrality holds with the engine-side membership fix); build; tidy diff-empty; schema-gen drift-empty; golangci 0 issues; dormancy audit clean (working-tree and range-based exclude-pathspec stages; .planning additions are exactly the three RED evidence records)"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "command: go test -race -count=1 ./internal/core/... ./internal/rules/... && go test -race -count=1 ./internal/rules/ -run 'TestRegistryParity|TestModuleRuleSet' && go test -race -count=1 ./internal/core/ -run 'TestReadSourceOnce|TestLintLinkedSuppressionParity' && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ && bin/golangci-lint run ./internal/rules/... ./internal/core/..."
        status: pass
    human_judgment: false

# Metrics
duration: 15 min
started: 2026-11-20T16:36:13Z
completed: 2026-11-20T16:50:41Z
status: complete
commits: 7
plan_head_before: 8a42a8fa6bd406507fcc77d75e47a8d01f23aa50
plan_head_after: fe11d9d30a33abe2e0755f71419d1958b360f3b1
---

# Phase 04 Plan 08: Gap Closure — Suppression Membership, Duplicate Inputs, Fingerprint Totality Summary

**All six verification gaps closed with RED-before/GREEN-after probes: ignore_only now changes group membership exactly like main (CR-01 engine fix), the parity harness drives identical suppression state (WR-01), duplicate file options compare per occurrence across all seven ports (WR-02), duplicate imports report at the last written occurrence (WR-03), the fingerprint and engine D-01 sort are total through SourceName (WR-04), and internal/rules is gofmt-clean (IN-01) — the 47-cell suite grew to 53 green cells with no regression**

## Performance

- **Duration:** 15 min
- **Started:** 2026-11-20T16:36:13Z
- **Completed:** 2026-11-20T16:50:41Z
- **Tasks:** 4
- **Files modified:** 15 code files (+3 planning artifacts: the RED evidence records; +this SUMMARY)

## Accomplishments

- **CR-01 closed in the engine, benefiting all nine group ModuleRules at once.** `collectFile` (internal/core/lint_linked.go) now skips `moduleRule.Collect` for a rule/file pair where `shouldIgnoreLinked(rule, info.Path)` is true — the exact shape of 04-REVIEW's fix sketch, mirroring main's shouldIgnore-before-Validate loop at core/lint.go:48. An ignored file never joins a group and never seeds the exempt-first value. The Finish drain's `shouldIgnoreLinked` path filter stays as-is (defense in depth; it also gates the per-file RPC_REQUEST_RESPONSE_UNIQUE findings). The contradicted doc claims are rewritten: collectFile's "group stays complete" comment and the ModuleRule contract sentence in lint_model.go now state the corrected REQ-2.5 semantics.
- **WR-01 closed by threading, making suppression parity testable by construction.** `assertBatchParity` captures NewLinked's second return value and sets `core.Options.IgnoreOnly` on the linked app. All pre-existing cells get identical unwrapped maps (no ignore_only configured) and stayed green untouched; the three new suppression cells compare real old-vs-new suppression state.
- **Three suppression parity cells (TestLinkedParitySuppressions).** `first_exempt_file_ignored` reproduces the CR-01 probe in-suite (old=0 asserted; new=1 at opts/v1/beta.proto:5:1 pre-fix, new=0 after); `mismatching_file_ignored` pins the drain path filter; `per_file_rule_ignored` pins the per-path Validate skip in checkFile. All three red-before/green-after via the validated task-1 RED record.
- **WR-02 closed across all seven option ports.** Each Collect now iterates `info.File.Options` in source order and appends a groupedValue per matching occurrence (go_package, java_package, java_multiple_files, csharp_namespace, php_namespace, ruby_package — including the pinned `ruby_package` literal in the swift-prefix port), reproducing main's first-seed/rest-compare through `firstExemptMismatches`. `duplicate_file_option` cell: old=2 AND new=2 at matching positions (pre-fix new=0). Optionless files contribute nothing via the empty loop; the swift_prefix bug pin text is byte-untouched.
- **WR-03 closed with main's last-wins selection.** `ImportUsedLinked.Validate` builds `importInfo` as an overwrite map keyed by `core.ImportPath` and reports each unused path once at the stored (last written) occurrence's Pos; the first-occurrence dedup set is gone. `duplicate_import_last_occurrence` cell: old=1 and new=1 at line 6 on both sides (pre-fix new reported line 5). The unsupported "rejected by protoc" justification is deleted; the determinism note states the actual selection and points at the pinning cell.
- **WR-04 closed in both sort sites.** `SourceName` is the fifth `cmp.Or` term in the harness fingerprint AND in LintLinked's final D-01 sort, with both comments updated: the parity contract is (path, rule, line, col, source); Message stays out because Message() equality is separately proven by the carried per-rule tests; the engine sort remains the one deliberate delta vs the etalon, now total through the user-visible SourceName. `TestFingerprintSourceNameTiebreak` pins ordering on ties plus the value-compare NotEqual pin.
- **IN-01 closed.** `gofmt -w internal/rules/registry_parity_guard_test.go` collapsed the requireConfigParity literal's two extra alignment spaces; `gofmt -l internal/rules/` prints nothing.
- **Helper/comment accuracy.** `fileOption` lost its last consumer, so its doc moved to the kept-unused state (nine D-01-kept helpers), and the lint_helpers.go header plus the .golangci.yml exclusion comment were updated in mirror so the documented state matches reality. The Phase-5 deletion obligation (helpers + exclusion together) is unchanged.
- **Gap-closure gates all green at the fixed state** (see Verification Log): full core+rules -race suites, registry/exact-set guards, TestReadSourceOnce + suppression centrality (standing constraints a+b), build, tidy diff-empty, schema drift-empty, golangci 0 issues, and the dormancy audit — working-tree AND range-based exclude-pathspec stages exit 0; the plan's diff touches only the declared files; the .planning additions are exactly the three RED evidence records.

## Task Commits

Each task was committed atomically (TDD RED→GREEN per task):

1. **Task 1 (RED): suppression cells + tiebreak pin fail** - `ad0e5d3` (test)
2. **Task 1 (GREEN): engine membership + harness threading + SourceName sort** - `b58ccc7` (feat)
3. **Task 2 (RED): duplicate_file_option fails old=2 vs new=0** - `2a045eb` (test)
4. **Task 2 (GREEN): per-occurrence iteration in the seven ports + helper/lint comment accuracy** - `9ee1121` (feat)
5. **Task 3 (RED): duplicate_import_last_occurrence fails line 6 vs line 5** - `3043dfb` (test)
6. **Task 3 (GREEN): last-wins importInfo + comment correction** - `509e78a` (feat)
7. **Task 4: gofmt the registry guard literal (IN-01)** - `fe11d9d` (chore)

**Plan metadata:** docs commit (this SUMMARY + STATE/ROADMAP/REQUIREMENTS updates).

## TDD Gate Compliance

- Plan `type: tdd`; Tasks 1-3 are `tdd="true"`. Each has a RED commit (`test(04-08)`) that precedes its GREEN commit (`feat(04-08)`) in ancestry; no REFACTOR commits were needed (each GREEN is the minimal implementation of its fix sketch).
- All three RED evidence records are machine-validated (`gsd_run check tdd-red-evidence` → `RED_EVIDENCE_OK`, reason `target_test_failed`, format `go-test`): task 1 `TestLinkedParitySuppressions` (first_exempt_file_ignored compare failure old len 0 vs new 1 at opts/v1/beta.proto:5:1 + tiebreak order failure), task 2 `TestLinkedParityBasicSemantic/duplicate_file_option` (old=2/new=0), task 3 `TestLinkedParityBasicSemantic/duplicate_import_last_occurrence` (line 6 vs line 5, same count). In every record the package compiled and the old-side count assertions passed before the compare — the failures are the planned assertion failures.
- The RED commits touch only `*_test.go` pathspecs plus their evidence JSON; each GREEN carries the behavior proof (new cells green + the full parity suite) run before the commit.
- Task 4 is closure (IN-01 gofmt + gates), not `tdd="true"` — no RED required; its gate set is fully green.

## Files Created/Modified

- `internal/core/lint_linked.go` - CR-01 collectFile membership fix + SourceName in the D-01 sort + comment rewrites
- `internal/core/lint_model.go` - ModuleRule contract doc: the suppression claim rewritten to the corrected REQ-2.5 semantics
- `internal/rules/parity_harness_test.go` - WR-01 threading; fingerprint fifth sort term + contract comments; TestFingerprintSourceNameTiebreak
- `internal/rules/parity_suppression_test.go` - NEW: TestLinkedParitySuppressions with the three suppression cells and the CR-01 probe provenance header
- `internal/rules/parity_basic_semantic_test.go` - duplicate_file_option + duplicate_import_last_occurrence cells; gap-closure probe-provenance notes
- `internal/rules/package_same_{go_package,java_package,java_multiple_files,csharp_namespace,php_namespace,ruby_package}_linked.go` - per-occurrence Collect loops; type-header bullet wording per-occurrence
- `internal/rules/package_same_swift_prefix_linked.go` - same sweep with the ruby_package literal and BUG PIN comments preserved verbatim
- `internal/rules/import_used_linked.go` - last-wins importInfo map + report loop rewrite; determinism note corrected
- `internal/rules/lint_helpers.go` - fileOption doc moved to the kept-unused state; header consumed/kept lists updated (nine kept helpers)
- `internal/rules/registry_parity_guard_test.go` - gofmt (IN-01)
- `.golangci.yml` - exclusion comment mirrors the fileOption move

## Decisions Made

- CR-01 is fixed once, in the engine (`collectFile`), not per-rule: one change gives all nine group ModuleRules main's membership semantics, and the rules stay suppression-free (standing constraint b preserved — centrality cells green).
- WR-01 is fixed by threading rather than an `assertBatchParityWithIgnoreOnly` variant: every cell — existing and future — compares identical suppression state by construction, and existing cells needed no edits.
- The regression-pin cells' pre-fix RED failures (mismatching_file_ignored, per_file_rule_ignored failing while the harness ran nil-vs-set state) were anticipated consequences of the WR-01 gap and are documented in the task-1 evidence record as such — they are the proof the threading was load-bearing, not unplanned failures.
- The pinned probe position (opts/v1/beta.proto:5:1) is reproduced exactly in-suite by writing the cell corpus without the leading blank line the house style otherwise uses, so the fingerprint mismatch names the review's exact position.
- No deferred item was implemented: resolution-based ImportUsed, cross-file RPCRequestResponseUnique, the swift_prefix fix decision, lab casing helpers, and old-path consolidation all stay with their owning phases (Phase 5/6/7); the old path is byte-frozen.

## Deviations from Plan

None - plan executed exactly as written. (The two regression-pin cells failing in the task-1 RED run were the documented, anticipated WR-01 consequence — recorded in the RED evidence, not a deviation.)

## Issues Encountered

None. Every RED run reproduced its planned failure on the first try; every GREEN run passed its gate set on the first complete pass. Pre-existing dirty files in the working tree (`.agents/skills/sdd/scripts/pipeline.sh`, `.planning/phases/03-.../03-VERIFICATION.md`, untracked runtime dirs) and the pre-existing gofmt flag on `internal/core/link_test.go` are out of scope per the deviation scope boundary — untouched, uncommitted.

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components were introduced. The D-01-kept helpers (now including the unconsumed `fileOption`) are verbatim-unused BY DESIGN (dormancy; Phase 5 deletes them with the old path) and are documented in the .golangci.yml comment, not stubs.

## User Setup Required

None - no external service configuration required.

## Verification Log

- Task 1 RED: `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParitySuppressions|TestFingerprintSourceNameTiebreak'` -> exit 1 with first_exempt_file_ignored compare failure (old len 0 vs new 1 at opts/v1/beta.proto:5:1) + tiebreak order failure; evidence validated RED_EVIDENCE_OK; committed `ad0e5d3`
- Task 1 GREEN: targets all green; grep gates 0 ('stays complete', 'apply to the findings by', 'newRules, _, err'); `TestLinkedParity` all parents green; full core+rules -race green; committed `b58ccc7`
- Task 2 RED: `TestLinkedParityBasicSemantic/duplicate_file_option` -> exit 1 (old 2 vs new 0, old-side count passing); evidence validated; committed `2a045eb`
- Task 2 GREEN: `TestLinkedParityBasicSemantic` 9/9 subtests green; grep gates 0 (fileOption(info.File across *_linked.go = 0 files; 'returns the last file-level option called name' = 0); full parity suite green; TestReadSourceOnce green; golangci 0 issues on internal/rules; committed `9ee1121`
- Task 3 RED: `duplicate_import_last_occurrence` -> exit 1 (same count, line 5 vs line 6); evidence validated; committed `3043dfb`
- Task 3 GREEN: cell green; grep gates 0 ('rejected by protoc', 'first written occurrence'); carried `TestImportUsed` green; full parity suite green; vet clean; committed `509e78a`
- Task 4: gofmt -l internal/rules/ empty (IN-01); full scoped suite `go test -race -count=1 ./internal/core/... ./internal/rules/...` ok; TestRegistryParity + TestModuleRuleSet green; TestReadSourceOnce + TestLintLinkedSuppressionParity green; `go build ./...` ok; `go mod tidy` + go.mod/go.sum diff-empty; `go run ./cmd/easyp schema-gen` + schemas/ diff-empty; `bin/golangci-lint run ./internal/rules/... ./internal/core/...` 0 issues; dormancy audit: working-tree stages 1+2 and range-based stages over `8a42a8f..HEAD` all exit 0; .planning additions = exactly the three evidence records; committed `fe11d9d`
- Cell-count gate with -v (checker advisory): 9 TestLinkedParity parents PASS, 53 total PASS lines (parents + sub-cells; was 47 — +3 suppression cells, +duplicate_file_option, +duplicate_import_last_occurrence, +TestFingerprintSourceNameTiebreak)
- Commits measured from the persisted ledger: `git rev-list --count 8a42a8f..HEAD` = 7 (excludes the metadata docs commit)

## Next Phase Readiness

**Phase 4 is gap-free.** The verifier's six gaps are closed with probe-verifiable acceptance (old=0/new=0 on the exempt-first-ignored corpus, old=2/new=2 on the duplicate-option corpus, line-6 parity on the duplicate-import corpus, tiebreak cell green), the parity contract now includes the user-visible SourceName, and the fixed state passes the full gate set. Phase 5 (the flip) inherits the corrected engine — the COMPAT gate now runs against main-parity suppression membership — plus the 53-cell parity suite and the consolidated deferred-decisions ledger from 04-07, unchanged in its obligations (old-path + helpers + exclusion deletion, rename-map collapse, buildCore wiring, COMPAT gate).

Ready for `/gsd:verify-work` re-verification of Phase 4.

---
*Phase: 04-perenos-43-pravil-s-batch-paritetom*
*Completed: 2026-11-20*
