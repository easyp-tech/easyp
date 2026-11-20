---
phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti
plan: 04
subsystem: lint-engine
tags: [mechanical-rename, name-consolidation, pure-reflection, override-drop, golden-master, compat-gates, final-battery]

# Dependency graph
requires:
  - phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti (plan 03)
    provides: THE FLIP (4238fbe) — the single-engine tree whose rename map (kept verbatim in builder_linked.go's header) this plan executes, plus the surviving golden corpus and fixtures the battery re-runs
  - phase: 04-perenos-43-pravil-s-batch-paritetom (plan 07)
    provides: the 41 port types and the registry-parity proof (override names == reflection names over the unsuffixed spellings) that made the rename name-neutral by construction
provides:
  - the consolidation commit (263f067): all 41 port types carry their canonical unsuffixed names (PackageDirectoryMatchLinked -> PackageDirectoryMatch, et al.), all 41 LintRuleName override methods dropped, GetRuleName collapsed to the etalon's pure-reflection shape, the builder_linked.go rename-map header deleted — zero stale consolidation debt, zero LIVE \w+Linked\b / LintRuleName references in internal/
  - the phase-final battery evidence on the committed final state: six-root scoped -race suite green, schema drift-empty, golangci 0 issues, tidy diff-empty, go-protoparser + protocompile still required, lint --help exit 0, breaking freeze + api seam gates intact, SC1 (golden + matrix incl. class-1 COMPILE + stdout golden) green
affects: [Phase 6 (new rules on the single engine), Phase 7 (external corpus gate inherits the golden corpus and fixtures unchanged — research A5)]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 32250        # 129020 patch chars over the consolidation commit at chars/4 (estimate 26000)
  tasks: 2
  commits: 1           # MEASURED: git rev-list --count 7b11f91006fde8b419129c44fda4d9e53f378c57..HEAD (the task commit; the metadata docs commit is excluded)
  plan_head_before: 7b11f91006fde8b419129c44fda4d9e53f378c57
  plan_head_after: 263f067a947541839a31e3fcd6ada637b874b407

# Tech tracking
tech-stack:
  added: []
  patterns: [one-atomic-mechanical-rename-commit (partial rename does not compile), scripted word-boundary sweep with per-pattern occurrence accounting, pure-reflection name derivation (no per-type overrides), zero-LIVE-reference grep as the rename-completeness gate]

key-files:
  created: []
  modified:
    - internal/rules/*_linked.go (41 port files: types renamed to canonical unsuffixed spellings, override methods dropped, headers de-self-referentialized)
    - internal/rules/builder_linked.go (rename-map header deleted; registry doc refreshed; constructor literals re-pointed to the unsuffixed names)
    - internal/rules/builder.go (add* helpers read the renamed unsuffixed names — name surface byte-identical)
    - internal/core/lint_model.go (GetRuleName collapsed to pure reflection; override branch + LintRuleName doc deleted)
    - internal/core/lint_linked.go, internal/core/lint_linked_read.go (protoInfoFromLinked -> protoInfoFromLink, call + def + comments)
    - 41 internal/rules/<rule>_test.go files (test references swept)

key-decisions:
  - "The consolidation is ONE mechanical commit (refactor 263f067, 87 files, +400/−669): a partial rename does not compile — deletion of overrides, the 41 renames, the name-function collapse and the header deletion land together (the plan's RATIFIED file-surface excess assumption executed)"
  - "No file renames: lint_linked.go, builder_linked.go and the *_linked.go file names stay per the plan's resolved discretion — pure churn avoided, four phases of verification references stay stable; a file-name polish is a deliberate later decision, not silent debt"
  - "protoInfoFromLinked renamed to protoInfoFromLink (Rule 3): the plan's action list named only the 41 port types, but its own zero-reference criterion (LintRuleName|\\w+Linked\\b == 0 over internal/) is unmeetable without renaming this symbol — mechanical word-boundary rename of the def, its two doc comments and the single engine call site"
  - "Comment-only factual refreshes forced by the same criterion and by the sweep: the 41 port-file headers ('X is the linked-engine port of X' would be self-referential after the rename), builder_linked.go's registry doc (override mention), package_directory_match's registry-name prose, GetRuleName's doc comment — no behavior text touched"
  - "Name neutrality was proven before committing, not after: TestAllRuleNames + TestAllGroups + TestModuleRuleSet green, TestLintGoldenFullRegistry byte-identical (full-struct compare incl. RuleName), schema-gen drift empty — three independent gates pin the name surface (threat T-05-07 mitigation)"

patterns-established:
  - "Rename-map execution discipline: a kept-for-later rename map is deleted in the same commit that executes it — the criterion grep (zero \\w+Linked\\b occurrences) proves no stale debt remains"
  - "Post-rename name gate: with overrides gone, TestAllRuleNames/TestAllGroups + the golden's RuleName values + schema drift are the standing name-surface contract"

requirements-completed: [LINT-COMPAT-01, LINT-COMPAT-02]  # copied verbatim from the plan frontmatter

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "The consolidation commit: 41 port types renamed to canonical unsuffixed names, 41 LintRuleName overrides dropped, GetRuleName collapsed to the etalon's pure-reflection shape, the builder_linked.go rename-map header deleted — one atomic mechanical commit (263f067, 87 files, +400/−669), reviewable as rename-only via git show --stat + git diff -M"
    requirement: LINT-COMPAT-01
    verification:
      - kind: other
        ref: "command: grep -rnE 'LintRuleName|\\w+Linked\\b' internal/ --include='*.go' | wc -l (0) && grep -c 'Rule-type rename map' internal/rules/builder_linked.go (0) && git show --stat 263f067 (87 files)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Name neutrality on the renamed types: the registry derives identical names through pure reflection — TestAllRuleNames, TestAllGroups and TestModuleRuleSet (the exact D-03 ten) all green post-rename"
    requirement: LINT-COMPAT-01
    verification:
      - kind: unit
        ref: "tests/internal/rules/builder_test.go#TestAllRuleNames"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/builder_test.go#TestAllGroups"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/builder_test.go#TestModuleRuleSet"
        status: pass
    human_judgment: false
  - id: D3
    description: "Golden byte-identity through the renamed types: the 73-finding full-registry golden reproduces with identical RuleName values (full-struct compare incl. Message/Position/SourceName, D-01 order)"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/rules/golden_full_registry_test.go#TestLintGoldenFullRegistry"
        status: pass
    human_judgment: false
  - id: D4
    description: "SC1 evidence re-run on the final state: the api matrix incl. the class-1 COMPILE cells, the four exit-code cells, the CLI stdout golden and the printer goldens all green"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrix/compile_issue_exit1_class"
        status: pass
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrixExitCodes/compile_error_code_1"
        status: pass
      - kind: integration
        ref: "tests/internal/api/lint_stdout_golden_test.go#TestLintStdoutGolden"
        status: pass
    human_judgment: false
  - id: D5
    description: "Phase-final battery on the committed final state: the six-root scoped -race suite green in ONE run (core, rules, api, modules, migration, config; api 217.4s) with the documented wazero carve-out (./internal/adapters/plugin/, ./internal/generation/ excluded — ledger item 18, owned before Phase 7); go build ./... green; go mod tidy diff-empty with go-protoparser and protocompile still required (greps = 1 each); schema-gen drift empty; golangci-lint 0 issues on ./internal/...; easyp lint --help exit 0"
    requirement: LINT-COMPAT-02
    verification:
      - kind: other
        ref: "command: go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/api/... ./internal/modules/... ./internal/migration/... ./internal/config/... (CHAIN_EXIT=0 verbatim plan verify chain)"
        status: pass
      - kind: other
        ref: "command: go mod tidy && git diff --exit-code go.mod go.sum && grep -c go-protoparser go.mod (1) && grep -c protocompile go.mod (1) && go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ && bin/golangci-lint run ./internal/... (0 issues) && /tmp/easyp-check lint --help (exit 0)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Standing constraints survive consolidation: breaking-owned files byte-frozen (proto_info_read.go, proto_comments.go zero-diff), the api seam intact (lint_v1.go zero-diff vs f27d2d1 with git rc 0 PIPESTATUS-captured; runtime.go comment-only, removed lines 0), the full core+rules scoped suites green under -race"
    requirement: LINT-COMPAT-02
    verification:
      - kind: other
        ref: "command: git diff --exit-code -- internal/core/proto_info_read.go internal/core/proto_comments.go && git diff f27d2d1 HEAD -- internal/api/lint_v1.go | wc -l (0; PIPESTATUS 0) && git diff f27d2d1 HEAD -- internal/api/runtime.go | grep -c '^-[^-]' (0)"
        status: pass
      - kind: other
        ref: "command: go test -race -count=1 ./internal/core/... ./internal/rules/... (all ok)"
        status: pass
    human_judgment: false

# Metrics
duration: 16 min
started: 2026-11-20T20:43:38Z
completed: 2026-11-20T20:59:50Z
status: complete
---

# Phase 05 Plan 04: Mechanical Consolidation + Final Battery Summary

**All 41 lint port types now carry their canonical unsuffixed names with the override mechanism deleted — one mechanical commit (263f067, 87 files, +400/−669) collapses the rename map, the name function drops to the etalon's pure-reflection shape, and the full phase battery is green on the committed final state**

## Performance

- **Duration:** 16 min
- **Started:** 2026-11-20T20:43:38Z
- **Completed:** 2026-11-20T20:59:50Z
- **Tasks:** 2
- **Files modified:** 87 in the consolidation commit (46 production + 41 test files) + the planning artifacts of this SUMMARY

## Accomplishments

- **The rename map is fully executed and deleted.** `refactor(05-04)` `263f067` — ONE mechanical commit: the 41 port types renamed to their canonical unsuffixed spellings (`PackageDirectoryMatchLinked -> PackageDirectoryMatch` and the other forty, across type declarations, the `builder_linked.go` constructor literals, the `builder.go` add* literals, interface assertions and every test reference), the 41 `LintRuleName()` override methods dropped, `GetRuleName` collapsed to the etalon's pure-reflection shape (the override-preference branch deleted; the `toUpperSnakeCase` reflection fallback byte-identical), and the `builder_linked.go` rename-map header deleted. The zero-reference gate holds: `grep -rnE 'LintRuleName|\w+Linked\b' internal/ --include='*.go' | wc -l` = 0, and `grep -c 'Rule-type rename map' internal/rules/builder_linked.go` = 0. No stale consolidation debt remains.
- **Name neutrality proven by three independent gates before committing** (threat T-05-07 mitigation): `TestAllRuleNames` + `TestAllGroups` + `TestModuleRuleSet` (the exact D-03 ten) green, `TestLintGoldenFullRegistry` byte-identical (full-struct compare incl. RuleName), schema-gen drift empty. A RuleName shift would have failed all three; none shifted.
- **The phase-final battery is green on the committed final state.** The verbatim plan verify chain ran end-to-end with CHAIN_EXIT=0: six-root scoped `-race` suite (core 2.2s, rules 1.1s, api 217.4s, modules 16.3s, migration 5.1s, config 1.0s + config/v1 1.9s) with the documented wazero carve-out (`./internal/adapters/plugin/`, `./internal/generation/` excluded — pre-existing third-party race, ledger item 18, owned before Phase 7, Phase 2–5 practice); `go build ./...`; `go mod tidy` diff-empty with `go-protoparser` and `protocompile` still required (presence greps = 1 each); schema-gen drift empty; `bin/golangci-lint run ./internal/...` 0 issues; `easyp lint --help` exit 0.
- **Standing constraints intact on the final tree.** Breaking-owned files byte-frozen (`proto_info_read.go`, `proto_comments.go` zero-diff); the api seam intact (`lint_v1.go` zero-diff vs f27d2d1 with PIPESTATUS-captured git rc 0; `runtime.go` comment-only, removed lines 0); SC1 re-run green — the api matrix including both class-1 COMPILE cells (`compile_issue_exit1_class` in-process, `compile_error_code_1` subprocess), the four exit-code cells, the CLI stdout golden, and the printer goldens.
- **The three permanent breaking-coexistence exceptions survive untouched:** `linkedFilesFromImport`/`linkedFileFromImport` and `linkedDirectiveComment` keep their names (the breaking path owns the unsuffixed twins; REQ-5.2 freeze respected — the criterion pattern `\w+Linked\b` never matched these lowercase-prefixed names).

## Task Commits

Each task was committed atomically:

1. **Task 1: The consolidation commit — 41 type renames, override drops, name-function collapse** - `263f067` (refactor; 87 files: 46 production + 41 test, +400/−669; rename-only, reviewable via git show --stat + git diff -M)
2. **Task 2: Final phase battery — the closing evidence on the final state** - verification-only (the verbatim plan verify chain CHAIN_EXIT=0 on the committed state; no file changes, no commit)

**Plan metadata:** docs commit (this SUMMARY + STATE/ROADMAP/REQUIREMENTS/WINDOWS).

## Files Created/Modified

- 41 `internal/rules/*_linked.go` port files - type renamed to the canonical unsuffixed spelling; `LintRuleName()` override dropped; header de-self-referentialized (`// X (research inventory #N, ...)`)
- `internal/rules/builder_linked.go` - the rename-map header deleted (every mapped name now canonical); registry doc refreshed (no override mention); constructor literals re-pointed to the unsuffixed names
- `internal/rules/builder.go` - the five add* helpers construct `core.GetRuleName(&<Unsuffixed>{})` — the produced name list byte-identical (TestAllRuleNames/TestAllGroups pin it)
- `internal/core/lint_model.go` - `GetRuleName` is the etalon's pure-reflection shape: `return toUpperSnakeCase(reflect.TypeOf(rule).Elem().Name())` (fallback line byte-identical), the override branch and its doc deleted
- `internal/core/lint_linked_read.go`, `internal/core/lint_linked.go` - `protoInfoFromLinked -> protoInfoFromLink` (def + 2 doc comments + the single engine call)
- 41 `internal/rules/<rule>_test.go` - references swept to the unsuffixed names (word-boundary sweep; assertions untouched)

## Decisions Made

- **Executed rename-only, no file renames** (the plan's resolved discretion): `lint_linked.go`, `builder_linked.go` and the `*_linked.go` file names stay — a file-name polish is a deliberate later decision, not silent debt; four phases of verification references remain stable.
- **The sweep was scripted with per-pattern accounting**: 41 word-boundary rename pairs + `protoInfoFromLink`, 41 uniform override blocks dropped (each verified as the exact 4-line shape before the script ran), 41 self-referential headers fixed — per-pattern match counts asserted in the script, `go build ./...` immediately after, gofmt clean, and the zero-reference grep as the completeness gate.
- **Name neutrality was verified pre-commit** (the plan's action step 5): the name gates, the golden and the freeze gate all ran before the single commit was created.
- **PATTERNS planner decision points — dispositions (as required by Task 2):**
  - **(a) fileOption disposition** — DELETED in the flip commit (4238fbe, one of the nine kept-unused helpers; zero consumers, header said delete, and the exclusion removal made keeping it fail the lint gate). Confirmed still absent: `golangci-lint` 0 issues on this plan's final state.
  - **(b) builder.go add* helpers re-point** — executed in the flip commit (4238fbe, re-pointed to the port types); this plan re-pointed them AGAIN to the unsuffixed names — the name surface byte-identical (TestAllRuleNames/TestAllGroups green).
  - **(c) .golangci.yml QF1001 block** — removed in the flip commit together with the lint_helpers.go paths exclusion; golangci reports 0 issues on the final state.
  - **(d) Consolidation vs minimal flip** — the consolidation shape executed in the flip commit (buildCore/actionV1 text unchanged; lint_v1.go zero-diff vs f27d2d1 re-proven on this plan's final state).
  - **(e) Optional 41-type-rename commit** — RESOLVED BY THIS PLAN: landed as the separate mechanical consolidation commit immediately after the flip (263f067), gated by the name-surface tests. The registry-parity guard died with the flip by design; its guarantee is carried by TestAllRuleNames/TestAllGroups + the golden + schema drift on the renamed types.
- **Phase 7's corpus gate inherits the surviving golden corpus and fixtures unchanged** (research A5): this plan touched no testdata, no fixture, no golden literal — the consolidation is symbols + comments only, so the Phase-7 lab corpus comparison runs against exactly the assets 05-01/05-03 committed.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `protoInfoFromLinked` renamed to `protoInfoFromLink` (not in the plan's action list)**
- **Found during:** Task 1 (post-sweep zero-reference criterion check)
- **Issue:** the plan's action step 1 named only the 41 port types, but its acceptance criterion `grep -rnE 'LintRuleName|\w+Linked\b' internal/ --include='*.go' | wc -l` == 0 is unmeetable without renaming `protoInfoFromLinked` (4 occurrences: the def in `lint_linked_read.go:139`, two doc comments naming it, and the engine call in `lint_linked.go:98`). The symbol is capital-L `Linked`-suffixed — exactly the consolidation debt the criterion exists to prove gone.
- **Fix:** mechanical word-boundary rename `protoInfoFromLinked -> protoInfoFromLink` (mirrors the sibling `protoInfoFromSource`; name-neutral — no behavior, ordering or signature change).
- **Files modified:** internal/core/lint_linked_read.go, internal/core/lint_linked.go
- **Verification:** zero-reference grep = 0; core suite green under -race; recorded in the WINDOWS ledger
- **Committed in:** 263f067 (Task 1 commit)

**2. [Rule 3 - Blocking] Comment-only factual refreshes forced by the criterion and the sweep**
- **Found during:** Task 1 (sweep fallout)
- **Issue:** (a) after the type renames, the 41 port-file headers read "X is the linked-engine port of X" — self-referential and factually wrong; (b) `builder_linked.go`'s registry doc said names come "via LintRuleName" and `package_directory_match_linked.go`'s prose said the name "is pinned through LintRuleName" — both factually wrong after the override drop; (c) `GetRuleName`'s doc described the override mechanism that no longer exists. The plan mandated the header (rename-map) deletion but not these; leaving them would keep `LintRuleName` text matches alive (criterion > 0) and leave wrong prose.
- **Fix:** comment-only refreshes: 41 headers collapsed to `// X (research inventory #N, ...)`, the two LintRuleName prose spots rewritten to state the reflection derivation, `GetRuleName`'s doc rewritten for the pure-reflection shape. No behavior text touched; `git diff -M` shows them as comment-line-only hunks.
- **Files modified:** the 41 port files, internal/rules/builder_linked.go, internal/rules/package_directory_match_linked.go, internal/core/lint_model.go
- **Verification:** zero-reference grep = 0; all suites green (comments cannot affect behavior; the gates re-prove it anyway)
- **Committed in:** 263f067 (Task 1 commit)

---

**Total deviations:** 2 auto-fixed (2 Rule 3 blocking: 1 symbol rename forced by the plan's own criterion, 1 comment-only factual-refresh fallout)
**Impact on plan:** Both are direct consequences of the plan's own zero-reference acceptance criterion; both are name/comment-surface only. The name surface, the golden, the fixtures and the breaking-owned files are untouched. No scope creep.

## Issues Encountered

- The sweep script's first run aborted on an internal assertion bug (a count-expectation conflated grep line counts with substitution occurrences) AFTER the override-drop step had applied; the sweep itself then applied fully on the corrected run and the tree state was verified against the plan's criteria (zero-reference grep, build, gates) rather than the script's own tally. No partial-file writes were possible by construction (per-file write happens only after that file's full substitution loop).
- The pre-existing dirty files (`.agents/skills/sdd/scripts/pipeline.sh`, `.planning/phases/03-…/03-VERIFICATION.md`, untracked `.planning/intel/`, `.planning/research/.cache/`, `.planning/state.json`, `.spec/features/`) remain out of scope per the deviation scope boundary; left untouched and uncommitted.
- `TestModuleRuleSet` still asserts the exact D-03 ten with PackageNoImportCycle absent — unchanged, as designed until Phase 6.

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components. The consolidation renames identifiers and deletes dead override code; every surviving test drives real engine behavior.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- **The phase's final state is the verifier's starting point:** one engine, one registry (`rules.New` in builder_linked.go), one rule surface, canonical unsuffixed type names, pure-reflection name derivation, zero consolidation debt (the zero-reference grep is now permanently clean, not just for this plan).
- **Verifier notes:** (a) the phase audit contract from 05-03 still applies (breaking-path freeze + deletion-completeness greps + subsystem suites); (b) the wazero carve-out remains documented for the `-race` gate (`internal/adapters/plugin/`, `internal/generation/`) — owned before Phase 7; (c) `TestModuleRuleSet`'s exact-set expectation (D-03 ten, PackageNoImportCycle absent) is a designed Phase-6 gate, not a gap.
- **Phase 7 inheritance (research A5):** the golden corpus, fixtures and pinned literals pass through this plan untouched — the external corpus gate can rely on the 05-01/05-03 assets byte-for-byte.
- No blockers. The phase is ready for /gsd:verify-work.

---
*Phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti*
*Completed: 2026-11-20*
