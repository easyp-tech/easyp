---
phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti
plan: 01
subsystem: testing
tags: [golden-master, compat-gates, lint-engine, cli-stdout-golden, d01-sort, parity, subprocess-test]

# Dependency graph
requires:
  - phase: 04-perenos-43-pravil-s-batch-paritetom (plan 07)
    provides: the full-registry corpus (13 files / 9 packages, 73 findings, 41-rule coverage assertion), the parity harness drive shape, and both alive engines (old serving CLI, linked dormant)
provides:
  - Gate 1: committed 73-finding golden (every field incl. Message, Position.Filename/Offset, SourceName; D-01 order) in internal/rules/golden_full_registry_test.go with the verbatim corpus re-host, the old-engine capture drive (dies at flip) and the linked parity drive (survives the flip as THE golden)
  - Gate 3: flip-invariant CLI stdout + exit-code golden (TestLintStdoutGolden) over a corpus covering findings, per-path exclusion, an easyp:disable directive and the PACKAGE_DIRECTORY_MATCH SourceName-to-Path rewrite
  - Message-level parity proven pre-flip — the flip's RED net is already wired and green
affects: [Phase 5 flip (05-03 golden re-point/renames), 05-02 error-path fixes, Phase 6/7 corpus gates]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 8064          # 32256 diff chars over the two new test files at chars/4 (estimate was 38000; two goldens, literals dominate)
  tasks: 3
  commits: 3            # MEASURED: git rev-list --count c96dd138867ce99059972b14818f32a9033ccd99..HEAD (task commits; excludes the metadata docs commit)
  plan_head_before: c96dd138867ce99059972b14818f32a9033ccd99
  plan_head_after: 98b9975aaf599ae67589e6f9a7fd322343cd778f

# Tech tracking
tech-stack:
  added: []
  patterns: [golden-master capture-then-flip (committed literals + two drives over one golden), D-01 totality guard (five-term tie halts the freeze), zero-COMPILE compile guard + pinned-count anti-vacuity + 41-name coverage on every golden drive, CLI stdout golden via buildEasypBinary/runEasypLint subprocess reuse]

key-files:
  created:
    - internal/rules/golden_full_registry_test.go
    - internal/api/lint_stdout_golden_test.go
  modified:
    - .planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-VALIDATION.md

key-decisions:
  - "The golden captures Position.Offset too (70/73 findings carry non-zero byte offsets; both engines agree byte-for-byte): full-struct fidelity per must_haves truth 1 ('every field populated'), so the compact literal helper takes eight arguments (path, rule, filename, offset, line, col, source, message), not the plan's sketched six — the planner's count predates the empirical field scan and Filename/Offset are user-visible in JSON output"
  - "The v1 per-path rule-suppression spelling is issues.exclude-rules (path + linters), not ignore_only: v1.Policy decodes with KnownFields(true) and its grammar has NO ignore_only key, so the plan's 'confirm the exact v1 key spelling' loop fired and the confirmed spelling is the issues section, applied by actionV1 after the engine run; engine-level ignore_only stays covered by the rules-package tests"
  - "The linked drive compares the engine's emitted order as-is (no re-sort): LintLinked already applies the D-01 total sort, so re-sorting would mask an engine-order regression; the old drive sorts with a local D-01 replica (slices.SortStableFunc + cmp.Or) because the old engine's output order is nondeterministic"
  - "The corpus is re-hosted verbatim under goldenCorpusFiles (not the parity file's fullRegistryFiles name) so the two files coexist in package rules_test until the flip; verbatimness is mechanically diff-verified against parity_full_registry_test.go lines 49-262"
  - "The PACKAGE_DIRECTORY_MATCH rewrite (lint_v1.go:213-215) is exercised with SourceName == Path: without module import roots sourceImportPath returns empty, so the rewrite fires and renders ToSlash(SourceName) byte-identically — distinguishing SourceName values would need a module policy with import roots"

patterns-established:
  - "Golden-master capture-then-flip: committed literals + two drives (capture drive dies with the oracle, parity drive survives) over ONE golden; captured via a temporary in-package generator (drive old engine, sort D-01, print literals, delete generator)"
  - "Golden purity guards: pinned count asserted BEFORE the compare, zero COMPILE findings (corpus compiles under full type resolution), all-41-name coverage, five-term totality guard on adjacent findings"

requirements-completed: [LINT-COMPAT-01]  # copied verbatim from the plan frontmatter; the REQUIREMENTS.md checkbox flips when the LAST declaring plan finishes (shared-ID gate: 05-02/03/04 also declare it)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Gate 1: the 73-finding full-registry golden (every field incl. Message, Position.Filename/Offset, SourceName; D-01 order) is committed, captured from the byte-frozen old engine, and the old drive reproduces it under -race with purity guards (pinned count, zero COMPILE, 41-name coverage, totality)"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/rules/golden_full_registry_test.go#TestLintGoldenFullRegistryOld"
        status: pass
    human_judgment: false
  - id: D2
    description: "Message-level parity: the dormant linked engine reproduces the old-captured golden exactly (full-struct compare incl. Message, Position.Filename/Offset and the D-01 order) on the same corpus — the flip's RED net is wired and green pre-flip, and the drive survives the flip as THE golden"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/rules/golden_full_registry_test.go#TestLintGoldenFullRegistryLinked"
        status: pass
    human_judgment: false
  - id: D3
    description: "Gate 3: the real binary's exact stdout (two path-sorted lines with (RULE) suffixes) and exit code 1 are pinned over a corpus whose four behaviors are each observable (finding prints, per-path exclusion suppresses one twin, easyp:disable suppresses the other mismatch, PACKAGE_DIRECTORY_MATCH renders the rewritten path)"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/api/lint_stdout_golden_test.go#TestLintStdoutGolden"
        status: pass
    human_judgment: false
  - id: D4
    description: "Zero production files modified: only the two new test files were committed; the Task-1 git-status guard (scoped to internal/rules/ + internal/core/) and go build ./... stay clean"
    verification:
      - kind: other
        ref: "command: go build ./... && go vet ./internal/rules/ ./internal/api/ && git status --porcelain -- internal/rules/ internal/core/ (count 0 at Task-1; internal/ shows test files only at Tasks 2-3)"
        status: pass
    human_judgment: false

# Metrics
duration: 20 min
started: 2026-11-20T19:00:48Z
completed: 2026-11-20T19:21:00Z
status: complete
---

# Phase 05 Plan 01: Golden Capture from the Frozen Old Engine Summary

**Gate 1 + Gate 3 captured pre-flip: a committed 73-finding full-registry golden (every field, D-01 order) that the old engine reproduces and the dormant linked engine already matches message-for-message, plus a byte-exact CLI stdout/exit-code golden through the real binary — zero production files touched**

## Performance

- **Duration:** 20 min
- **Started:** 2026-11-20T19:00:48Z
- **Completed:** 2026-11-20T19:21:00Z
- **Tasks:** 3
- **Files modified:** 2 new test files (+1 planning artifact: 05-VALIDATION.md rows; +this SUMMARY)

## Accomplishments

- **The capture window is closed with both gates on disk.** `internal/rules/golden_full_registry_test.go` re-hosts the Phase-4 closure corpus verbatim (13 files / 9 packages, declared rpc payload types, directive placements — mechanically diff-verified) and holds `goldenIssues`: 73 `core.IssueInfo` literals with every field populated (Path, RuleName, Position incl. Filename/Offset/Line/Column, SourceName, Message) in the engine's D-01 order, captured from the byte-frozen old engine while the oracle still exists. The compact `golden(...)` helper keeps the block reviewable, one finding per line.
- **The capture drive is green.** `TestLintGoldenFullRegistryOld` drives the old engine exactly like the Phase-4 harness (rules.New + Options.Rules + app.Lint over fs.NewFSWalker), sorts with a local D-01 replica (slices.SortStableFunc + cmp.Or — the exact sort of internal/core/lint_linked.go:152-160), and requires Equal against the literals. It dies at the 05-03 flip by design.
- **Message-level parity is proven pre-flip — the flip's RED net is wired and green.** `TestLintGoldenFullRegistryLinked` drives the dormant linked engine (NewLinked + Options.LinkedRules + LintLinked) on the same corpus and requires the full-struct equality including Message and the D-01 order — upgrading the Phase-4 fingerprint, which deliberately excluded Message and whole-struct equality (Filename/Offset included). The engine's emitted order is compared as-is (no re-sort), so the test also pins the engine's own deterministic sort. The 05-03 consolidation sweep renames this function's symbols and it becomes THE post-flip golden.
- **Purity guards on every drive:** the pinned 73 count asserted BEFORE the compare (T-04-02), zero findings named COMPILE (the corpus compiles under the linked engine's full type resolution — Pitfall 7), the 41-name coverage assertion (a silent rule is a corpus hole), and the D-01 totality guard halting on any five-term tie (flagged assumption 3 — the captured set is total, verified: 73 distinct keys).
- **Gate 3 is pinned through the real binary.** `internal/api/lint_stdout_golden_test.go` reuses `buildEasypBinary`/`runEasypLint` and pins the byte-exact stdout (two path-sorted lines with "(RULE)" suffixes) and exit code 1 over ONE corpus covering all four observable behaviors: FILE_LOWER_SNAKE_CASE finding printed; the per-path exclusion suppressing exactly one of two twins; an `easyp:disable PACKAGE_DIRECTORY_MATCH` directive suppressing the second mismatch; and the PACKAGE_DIRECTORY_MATCH SourceName-to-Path rewrite (lint_v1.go:213-215) exercised. The test is flip-invariant — no re-point at the flip.
- **The v1 spelling question is settled with evidence.** The v1 policy grammar has no `ignore_only` key (KnownFields strict decode would reject it); the confirmed v1 mechanism for "suppress rule R on exactly one of two files" is `issues.exclude-rules` with path + linters, applied by actionV1 after the engine run. Both suppression channels are observable in the pinned stdout.
- Phase-5 obligations untouched: no production file modified, issue output format/paths/exit codes unchanged (C1), no deferred behavior change smuggled into the corpus (C3), both golden verify chains green, `go build ./...` and `go vet` clean, and the full `internal/rules` + `internal/api` packages green under `-race`.

## Task Commits

Each task was committed atomically:

1. **Task 1: Golden capture from the frozen old engine (Gate 1, tracer)** - `afc9a2f` (test)
2. **Task 2: Linked-engine message-parity proof (the flip's RED net)** - `8f88380` (test)
3. **Task 3: CLI stdout golden from the real binary (Gate 3)** - `98b9975` (test)

**Plan metadata:** docs commit (this SUMMARY + STATE/ROADMAP/VALIDATION updates).

## Files Created/Modified

- `internal/rules/golden_full_registry_test.go` - NEW: verbatim corpus re-host + 73-finding golden literals + capture drive (dies at flip) + linked parity drive (survives) + sort/totality/purity helpers
- `internal/api/lint_stdout_golden_test.go` - NEW: CLI stdout + exit-code golden over the findings/exclusion/directive/rewrite corpus, reusing the Phase-2 subprocess harness
- `.planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-VALIDATION.md` - Task rows 1-2 filled with the real -run patterns (golden + subprocess goldens), Wave-0 golden-capture checkbox checked

## TDD Gate Compliance

- Plan `type: execute` (not `type: tdd`); no task carries `tdd="true"`, so no RED/GREEN gate sequence applies. The capture and parity cells are green-by-construction additions to the proven engines (04-07 Task 2/3 precedent for non-tdd tasks).

## Decisions Made

- The golden captures `Position.Offset` (70/73 findings carry non-zero byte offsets, identical on both engines) — full-struct fidelity per the plan's own "every field populated"; the literal helper therefore takes eight arguments, not the sketched six. Filename/Offset are user-visible in JSON output, so strictness is the right bias for a compat net.
- v1 per-path rule suppression is spelled `issues.exclude-rules` (path + linters): `v1.Policy` has no `ignore_only` key and decodes with KnownFields(true), so the plan's confirm-the-spelling loop fired; the corpus exercises the confirmed spelling while engine-level `ignore_only` stays covered by the rules-package tests.
- The linked drive asserts the engine's emitted order directly (no re-sort) — a stronger contract than re-sorting, which would mask an engine-order regression.
- The corpus re-host uses new names (`goldenCorpusFiles`, `goldenCorpusUse`) so the golden file coexists with the living parity file in package rules_test until the flip; verbatimness is enforced by a mechanical diff against the parity block.
- Temporary capture tooling (a probe dump + a literal generator test) lived in-package only during development and was deleted before the Task-1 commit; the committed file holds literals, not a generator, per the plan.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] The plan's Task-1 verify chain cannot exit 0 in its own passing case**
- **Found during:** Task 1 (tracer feedback gate re-run, post-commit)
- **Issue:** the `<automated>` guard `git status --porcelain … | grep -vc '^?? …golden_full_registry_test.go' | grep -q '^0$'` under `set -o pipefail` always exits 1: `grep -c` exits 1 when zero lines are selected, which is exactly the guard's PASSING state (empty porcelain post-commit). The intended condition (count == 0, i.e. no tracked edits under internal/rules/ or internal/core/) held every time.
- **Fix:** ran the gate with corrected plumbing preserving the exact intended semantics: capture `COUNT=$(… | grep -vc '…')` via command substitution (which ignores the inner exit status) and test `[ "$COUNT" = "0" ]`. Test green + count 0 on the committed state.
- **Files modified:** none (verification plumbing only)
- **Verification:** `go test -race -count=1 ./internal/rules/ -run 'TestLintGoldenFullRegistryOld'` pass + `GUARD_COUNT=0`
- **Committed in:** n/a — no code change; recorded in .planning/WINDOWS.md as a deviation entry

**2. [Rule 3 - Blocking] The golden literal helper needs eight arguments, not the plan's six**
- **Found during:** Task 1 (capture probe)
- **Issue:** the plan sketched a six-argument helper (path, rule, line, col, source, message), but the captured data has non-empty `Position.Filename` on 3 findings (FILE_LOWER_SNAKE_CASE ×2, PACKAGE_DEFINED) and non-zero `Position.Offset` on 70 — six args would silently lose fields the plan's own must_haves truth requires ("Position (Filename/Line/Column) … for every finding").
- **Fix:** helper signature `golden(path, rule, filename string, offset, line, column int, source, message string)`; every field explicit in the committed literals.
- **Files modified:** internal/rules/golden_full_registry_test.go
- **Verification:** both golden drives green; full-struct equality between engines holds on the corpus
- **Committed in:** afc9a2f

**3. [Rule 3 - Blocking] `ignore_only` is not a v1 policy key — the confirmed spelling is issues.exclude-rules**
- **Found during:** Task 3 (fixture design, per the plan's "confirm the exact v1 key spelling" instruction)
- **Issue:** the plan's fixture sketch named an `ignore_only` key, but `internal/config/v1` Policy/LinterPolicy has no such field and decodes with KnownFields(true) — the sketch's key would fail strict decode as the plan predicted; the v1 grammar's per-path rule suppression lives in `issues.exclude-rules` (path + linters), applied by actionV1 after the engine run.
- **Fix:** the fixture uses `issues.exclude-rules` with path + linters; the semantic the plan wanted (rule R suppressed on exactly one of two files while printing on the other) is observable in the pinned stdout; engine-level ignore_only remains covered by rules-package tests. Documented in the test header.
- **Files modified:** internal/api/lint_stdout_golden_test.go
- **Verification:** TestLintStdoutGolden green; excluded twin's finding absent, twin's printed
- **Committed in:** 98b9975

---

**Total deviations:** 3 auto-fixed (3 Rule 3 blocking: 1 verification plumbing, 2 plan-spec reconciliations against the live source/data)
**Impact on plan:** All three were necessary for the gates to hold with full fidelity; no scope change, no production edits, the capture window closed with both goldens green.

## Issues Encountered

- The Task-1 git-status guard's grep exit-status artifact (deviation 1) initially misread as GUARD_FAIL on a clean tree; dissected step by step (porcelain empty → `grep -vc` prints 0 and exits 1 → pipefail propagates) before correcting the plumbing.
- PACKAGE_DIRECTORY_MATCH fired at 3:1 (not 4:1) in the first Task-3 probe and a second PACKAGE_SAME_DIRECTORY finding appeared because the directive-suppressed file initially shared the other mismatch's package; the fixture was restructured (distinct `sup_pkg` package) so each file carries exactly one intended finding and the stdout stays deterministic (md5-stable across 3 runs).
- Unrelated pre-existing dirty files in the working tree (`.agents/skills/sdd/scripts/pipeline.sh`, `.planning/STATE.md`, `.planning/phases/03-…/03-VERIFICATION.md`, untracked `.planning/intel/`, `.planning/research/.cache/`, `.planning/state.json`, `.spec/features/`) — out of scope per the deviation scope boundary; left untouched and not committed.
- The full `internal/api` -race suite takes ~186s (subprocess tests); both affected packages were run green before close-out.

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components. The committed goldens are complete literals; both drives assert them.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- **The flip has its byte-level net.** 05-02 (pre-flip error-path work) and 05-03 (the flip: old-path deletion + rename-map consolidation, re-running `TestLintGoldenFullRegistry` with the linked drive renamed to New/Rules/Lint and TestLintGoldenFullRegistryOld deleted) inherit: the committed corpus (now in its flip-surviving home), the full-struct golden including messages, and the flip-invariant CLI stdout gate.
- **Parity facts the flip can rely on:** both engines' full structs are identical on the corpus (verified during capture: 73/73 findings byte-identical after the 5-key sort, incl. Filename/Offset), so the golden compare is expected GREEN through the flip; any post-flip divergence surfaces as a test failure naming the exact field.
- No blockers. The plan's flagged assumption 3 (D-01 totality) is verified on the captured set: 73 distinct five-term keys.

---
*Phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti*
*Completed: 2026-11-20*

## Self-Check: PASSED

- All created files exist on disk: internal/rules/golden_full_registry_test.go, internal/api/lint_stdout_golden_test.go, this SUMMARY
- All 3 task commits are ancestors of HEAD: afc9a2f (test, capture), 8f88380 (test, linked parity), 98b9975 (test, CLI stdout golden)
- Commits measured from the persisted ledger: `git rev-list --count c96dd138867ce99059972b14818f32a9033ccd99..HEAD` = 3 (plan_head_before c96dd138867ce99059972b14818f32a9033ccd99, plan_head_after 98b9975aaf599ae67589e6f9a7fd322343cd778f)
- Final verify runs on the committed state: `go test -count=1 ./internal/rules/ -run 'TestLintGoldenFullRegistry'` ok (both drives); `go test -count=1 ./internal/api/ -run TestLintStdoutGolden` ok; `go build ./...` ok
- WINDOWS.md ledger: one deviation entry appended (the Task-1 verify-chain grep-exit-status artifact); no stubs, no skipped tests, no unrun verifies
- LINT-COMPAT-01 left unmarked by the shared-ID gate: 05-02/05-03/05-04 also declare it and have no SUMMARYs yet — the LAST declaring plan flips the REQUIREMENTS.md checkbox
