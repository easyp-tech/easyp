---
phase: 02-linker-i-marshrutizatsiya-oshibok
verified: 2026-11-19T05:44:31Z
status: human_needed
score: 8/10 must-haves verified
behavior_unverified: 2
overrides_applied: 0
covered_files: [".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-01-PLAN.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-01-SUMMARY.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-02-PLAN.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-02-SUMMARY.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-03-PLAN.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-03-SUMMARY.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-01-task1.json", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-02-task1.json", "internal/api/lint_link_matrix_test.go", "internal/core/link.go", "internal/core/link_test.go", "internal/core/testdata/link/multi_file/a.proto", "internal/core/testdata/link/multi_file/b.proto", "internal/core/testdata/link/tab_utf8/broken.proto"]
covered_digest: "v4:sha256:f5fec5fd6f85cbe758e1d7e9eb0cbc38816638385038cabeeac79b837f1a0502"
behavior_unverified_items:
  - truth: "SC-4: the fixture matrix of all three error classes passes with CLI exit codes verified through the Action path — its class-1 (COMPILE) leg through Action is not exercised by any test"
    test: "After the Phase 5 flip, run the built easyp binary on a compile-error corpus (no missing import, no syntax error)"
    expected: "COMPILE remarks render in the existing issues format and the exit code follows the ErrHasLintIssue/COMPILE class contract pinned by the golden"
    why_human: "Physically unreachable while the linker is dormant (D-04: internal/core/lint.go and internal/api/lint.go must stay byte-unchanged this phase); ratifying the A1 two-level interpretation (classifier cells + printIssues golden now, identical matrix re-run through Action in Phase 5) is a policy decision, not a testable behavior"
  - truth: "Warn diagnostics are dropped per the etalon policy (research A3) — the IsWarn branch has no exercising test"
    test: "When a natural warn-producing fixture exists (Phase 5 matrix re-run), link a corpus that yields only warn diagnostics and confirm zero COMPILE issues plus fallback-to-COMPILE behavior matches the etalon"
    expected: "IsWarn diagnostics never become COMPILE issues"
    why_human: "No natural warn fixture exists in this phase (02-02 flagged assumption A3); the branch is code-present (link.go:45-47) but no test drives a warn diagnostic through it"
human_verification:
  - test: "Ratify flagged assumption A1 (02-03 coverage D5, human_judgment: true): class-1 (COMPILE) verification before the Phase 5 flip is satisfied by the two-level matrix — classifier cells (02-01/02-02) + printIssues golden (02-03) — with the identical matrix re-run through Action in Phase 5"
    expected: "Explicit acceptance (or rejection and a gap-closure round) of the two-level interpretation of ROADMAP SC-4"
    why_human: "Policy interpretation of the roadmap contract under the D-04 dormancy invariant; automation cannot decide it (the orchestrator instructed this be disposed as an explicit human item, not a silent pass)"
  - test: "Resolve the MVP-mode discrepancy: ROADMAP sets Mode: mvp but the phase goal is not a User Story (user-story.validate -> false); this verification proceeded goal-backward against the four ROADMAP success criteria per the Phase 1 precedent"
    expected: "Either ratify the precedent for this engineering milestone or run /gsd:mvp-phase 02 to set a proper User Story goal and re-verify"
    why_human: "Mode/gate policy decision surfaced per verify-mvp-mode; inventing a user story for an internal compiler-migration phase would be fabrication"
  - test: "Re-pin the warn-drop policy when a natural warn fixture appears (Phase 5 matrix re-run) — flagged assumption A3 from plan 02-02"
    expected: "A test cell exercising the IsWarn drop branch with a real warn diagnostic"
    why_human: "No warn-producing fixture exists in this phase; the branch is unexercised code"
  - test: "Confirm retention of the etalon-verbatim defensive fallback branch (!errors.Is(err, protoc.ErrCompilationFailed) -> (nil, nil)) that no unit can trigger with a complete linted map"
    expected: "Branch stays with its not-unit-triggerable comment (plan 02-02 documents it as deliberate, never deleted)"
    why_human: "Unexercised defensive path; acceptance of its documentation-only coverage is a judgment call"
deferred:
  - truth: "Import open/read failures classify as COMPILE instead of the documented fallback (review finding WR-01)"
    addressed_in: "Phase 5"
    evidence: "Phase 5 goal: 'easyp lint работает на компиляторе; старый lint-контур удалён; формат вывода, пути, коды выхода... не изменились'; 02-REVIEW-DISPOSITION.md routes WR-01 to the Phase 5 flip plan (dormant now, real semantic gap only post-flip)"
  - truth: "lintResolver serves corpus files from memory, bypassing importFileAllowed/sourceFileOpen (review finding WR-02)"
    addressed_in: "Phase 5"
    evidence: "02-REVIEW-DISPOSITION.md: 'tie to Phase 5 wiring where importFileAllowed joins the linker path'; dormant now, edge-reachable only with restrictive source_roots"
  - truth: "Full class-1 (COMPILE) matrix re-run through the Action path with CLI exit codes"
    addressed_in: "Phase 5"
    evidence: "Phase 5 goal owns the user-visible flip ('единственная точка пользовательски видимого изменения'); research A1 interpretation recorded in 02-03 PLAN/SUMMARY flagged assumptions"
---

# Phase 02: Линкер и маршрутизация ошибок — Verification Report

**Phase Goal:** Линковка корпусов через `easypc` с корректной трёхсторонней классификацией отказов: COMPILE-замечания, откат на пофайловый разбор при отсутствующем импорте, ошибки чтения/синтаксиса как в main
**Verified:** 2026-11-19T05:44:31Z
**Status:** human_needed
**Re-verification:** No — initial verification

## MVP Mode Discrepancy

ROADMAP sets `Mode: mvp` for this phase, but the goal is not in User Story format (`user-story.validate` → `false`). Per `verify-mvp-mode.md` this is surfaced as a discrepancy rather than silently absorbed. Following the documented Phase 1 precedent (02-01 PLAN, "Flagged Assumptions"): this milestone is an internal compiler migration; inventing a story framing would be fabrication. Verification proceeded goal-backward against the four ROADMAP success criteria. No "User Flow Coverage" section is emitted (no user story exists to decompose). **A human decision is requested** (human_verification item 2): ratify the precedent or run `/gsd:mvp-phase 02`.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC-1: Corpus failing linking with compile errors (not missing import) yields COMPILE remarks (position + compiler message) in the existing issues format; rules do not run | ✓ VERIFIED | `TestLinkSources/compile_error` pins the etalon cell verbatim (test.proto 4:3, "Required fields are not allowed in proto3.", zero Filename/Offset/SourceName); dep-file, multi-file, tab/UTF-8, cycle and order cells all green; `TestPrintIssuesCompileGolden` pins the existing text/JSON printer format byte-for-byte. All run by this verifier: 16/16 cells pass under `-race` |
| 2 | SC-2: Missing import falls back to per-file parsing; typed `OpenImportFileError` reaches the existing exit-2 route | ✓ VERIFIED | `missing_import`, `transitive_missing_depth2`, `blocked_import` cells assert `(nil, nil)` fallback plus `require.ErrorAs *OpenImportFileError` with the import FileName; in-process `TestLintLinkMatrix/missing_import_exit2_class` through `actionV1`; subprocess `missing_import_code_2` proves actual exit 2 with "Cannot import file" + `missing.proto` on stderr via the built binary |
| 3 | SC-3: Read/syntax error returns a command execution error exactly as before | ✓ VERIFIED | `TestLintLinkMatrix/syntax_error_command_class` (plain error, `require.NotErrorAs *OpenImportFileError`); subprocess `syntax_error_code_1` proves actual exit 1 through the built binary |
| 4 | SC-4: Fixture matrix of all three error classes passes with CLI exit codes verified through the Action path — before engine insertion | ⚠️ PRESENT_BEHAVIOR_UNVERIFIED | Matrix proven through `actionV1` + subprocess for the reachable classes (missing→2, syntax→1, lint-issue→1, clean→0; 10/10 cells green). The class-1 (COMPILE) leg through Action is physically unreachable while the linker is dormant (D-04) — covered at component level (classifier cells + printIssues golden) per flagged assumption A1, deferred to Phase 5. Routes to human ratification — see behavior_unverified_items / human_verification |
| 5 | Compiling corpus (incl. WKT import) returns non-nil `*protoc.Linked` + nil issues; missing-import detection goes only through the typed resolver seam | ✓ VERIFIED | `compiling_with_wkt` + `wkt_resolves` cells; `link.go:116-135` — `FindFileByPath` maps typed errors (`errors.As *OpenImportFileError` / `errors.Is os.ErrNotExist`) to the `missing` atomic flag + `protoc.ErrFileNotFound`; no diagnostic-text parsing anywhere in the file |
| 6 | COMPILE issues deterministically sorted by Path, Line, Col (D-03), stable across runs and across files | ✓ VERIFIED | `deterministic_order_twice` (require.Equal on full slices of two runs) + `multi_file_matrix_order` (natural dep-first fixture order emits a.proto-then-b.proto) — both green in the verifier's run |
| 7 | Missing-import matrix completeness: transitive at dep depth 2, blocked by `importFileAllowed`, WKT resolves, non-local "../x.proto" classifies COMPILE per etalon (research A4) | ✓ VERIFIED | Four named cells green: `transitive_missing_depth2`, `blocked_import` (empty-importRoots Core construction honored), `wkt_resolves`, `non_local_import_compile` (COMPILE issue referencing the path) |
| 8 | COMPILE positions follow the compiler LineCol convention, pinned by a fixture that provably diverges from the Phase 1 D-07 rune contract | ✓ VERIFIED | `tab_utf8_position`: fixture `testdata/link/tab_utf8/broken.proto` — Line 6 / Column 37 asserted (byte columns, tab stop 8), comment records the rune-column value 19; green in the verifier's run |
| 9 | Import-cycle diagnostics pinned at Line 2 / Col 1 with the "recursively imports itself" chain (driver hard-coded 2:1) | ✓ VERIFIED | `import_cycle` cell asserts the exact position and message substring; green in the verifier's run |
| 10 | Warn diagnostics dropped per the etalon policy (research A3) | ⚠️ PRESENT_BEHAVIOR_UNVERIFIED | Branch present at `link.go:45-47` (`if d.IsWarn { return }`) but no test drives a warn diagnostic through it — no natural warn fixture exists in this phase (02-02 flagged assumption A3; Phase 5 re-pin). See behavior_unverified_items |

**Score:** 8/10 truths verified (2 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/core/link.go` | CompileRuleName const, `(*Core).linkSources` three-outcome classifier, `lintResolver` typed seam, `hasPerFileSyntaxError` | ✓ VERIFIED | All four symbols present and substantive (135 lines); wired to tests (`link_test.go` exercises every branch but the two noted unexercised ones); deliberately NOT wired into `Lint()` — that is the D-04 dormancy requirement, verified byte-unchanged since phase base `b5c26df` |
| `internal/core/link_test.go` | TestLinkSources with the full 16-cell matrix | ✓ VERIFIED | 16 named cells, all passing under `-race` in the verifier's own run (17 including parent) |
| `internal/core/testdata/link/tab_utf8/broken.proto` | Tab+multi-byte-UTF-8 fixture with provably divergent LineCol | ✓ VERIFIED | Tracked in git; drives `tab_utf8_position` (6:37 vs rune 19) |
| `internal/core/testdata/link/multi_file/a.proto`, `b.proto` | Cross-file D-03 ordering corpus | ✓ VERIFIED | Tracked in git; drive `multi_file_matrix_order` |
| `internal/api/lint_link_matrix_test.go` | TestLintLinkMatrix (4 in-process cells), TestLintLinkMatrixExitCodes (3 subprocess cells), TestPrintIssuesCompileGolden (3 golden cells) | ✓ VERIFIED | 10/10 cells green in the verifier's own run, including subprocess exit codes 2/1/0 through a freshly built binary |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| internal/core/link.go | internal/easypc/protoc/link.go | `protoc.Link(ctx, paths, WithResolver, WithSyntaxTrees(true), WithErrorHandler)` | WIRED | link.go:41-51; options exactly as planned, no `WithStandardImports` anywhere |
| internal/core/link.go | internal/core/proto_info_read.go | `lintResolver.FindFileByPath -> c.openImportFile` (typed missing-import seam D-01) | WIRED | link.go:120; typed mapping at link.go:122-126 |
| internal/core/link.go | internal/core/dom.go | `IssueInfo{RuleName: CompileRuleName}` | WIRED | link.go:79-86; `Position = protoast.Position` fields Line/Column set, Filename/Offset zero |
| internal/core/link.go | internal/protoast/build.go | `hasPerFileSyntaxError -> protoast.Parse` | WIRED | link.go:95-102, invoked only on the failed-Link branch (link.go:63) before sort/issue construction |
| internal/core/link_test.go | internal/core/testdata/link/ | fixture loading via `os.ReadFile` | WIRED | `mustReadFixture` helper; tab_utf8 + multi_file cells |
| internal/api/lint.go | internal/core/dom.go | `errors.As *OpenImportFileError -> errExit(2, "Cannot import file")` | WIRED | lint.go:85-91 (unchanged this phase — verified byte-identical since `b5c26df`) |
| internal/api/lint.go | internal/flags/flags.go | printIssues TextFormat/JSONFormat | WIRED | golden cells exercise both formats through the unchanged printer |
| internal/api/lint_link_matrix_test.go | internal/api/lint_v1_test.go | actionV1 + cli.NewContext harness precedent | WIRED | `runLinkMatrixAction` mirrors the harness (private FlagSet, ctx.Context = t.Context()) |

### Data-Flow Trace (Level 4)

Not applicable to rendering surfaces — this phase ships no UI/API route that renders dynamic data. The data-flow equivalents were traced: classifier inputs are real in-memory/testdata corpora (not hardcoded expectations of the system under test); golden expected values derive from the etalon (external lab oracle) and the sealed `token.go` LineCol contract, not from the system being tested (no circularity). The subprocess matrix flows real fixtures through the real binary.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Classification matrix (16 cells) | `go test -race -count=1 -v ./internal/core/ -run 'TestLinkSources'` | ok, 16/16 named cells PASS | ✓ PASS |
| Action-path error classes (4 cells) | `go test -race -count=1 -v ./internal/api/ -run 'TestLintLinkMatrix\|TestLintLinkMatrixExitCodes\|TestPrintIssuesCompileGolden'` | ok, 10/10 cells PASS | ✓ PASS |
| Subprocess CLI exit codes 2/1/0 | included above (`TestLintLinkMatrixExitCodes`) | missing_import_code_2 (exit 2 + "Cannot import file" + missing.proto), syntax_error_code_1 (exit 1), compiling_clean_code_0 (exit 0, empty stdout) | ✓ PASS |
| printIssues COMPILE golden text/JSON/order | included above (`TestPrintIssuesCompileGolden`) | byte-exact `test.proto:4:3: ... (COMPILE)` line; JSON zero Filename/Offset; order preserved | ✓ PASS |
| Whole-workspace suite | not run | pre-existing `-race` failure in `internal/adapters/plugin` (wazero), verified identical at phase base `a77a13d` per deferred-items.md — not a phase regression; targeted suites green | ? SKIP (documented) |

Full-suite constraint honored: targeted runs only; no per-truth full-suite filtering.

### Probe Execution

No probes declared (`scripts/*/tests/probe-*.sh` absent; no probe references in PLAN/SUMMARY). The plan's own verification suite (targeted `-race` runs, lint, dependency gate) was re-executed by this verifier instead — see spot-checks above. SKIPPED (no probes exist for this phase).

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| LINT-ENG-02 (REQ-2.2) | 02-01, 02-02, 02-03 | Compile-failed files (not missing import) get COMPILE remarks; rules do not run on the corpus | ✓ SATISFIED | Classifier behavior test-proven (etalon cell + matrix); existing format pinned by golden; `linked == nil` when issues are returned so no rule can consume trees. Marked Complete in REQUIREMENTS.md traceability; end-to-end user-visible rendering lands with the Phase 5 flip per flagged assumption A1 (ratification item 1) |
| LINT-ENG-03 (REQ-2.3) | 02-01, 02-02, 02-03 | Missing import -> per-file fallback, read/syntax errors reported as before, no COMPILE issues; `OpenImportFileError` reaches exit 2 | ✓ SATISFIED | Fallback proven for direct/transitive/blocked variants; typed carrier asserted in-process and actual exit 2 proven via subprocess through the built binary |
| LINT-ENG-04 (REQ-2.4) | 02-02, 02-03 | Read/syntax error file returns command execution error as before | ✓ SATISFIED | In-process class assertion (plain error, NOT the carrier) + actual exit 1 via subprocess |

Orphaned requirements: none. REQUIREMENTS.md traceability maps exactly LINT-ENG-02/03/04 to Phase 2; all three are declared in plan frontmatter. All three plan-declared IDs accounted for.

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts. (`check.decision-coverage-verify`: total 5, honored 5, not_honored [] — gate non-blocking.)

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| internal/core/link_test.go | LINT-ENG-02/03/04 | 16 | 0 | 0 | Value + behavioral (exact positions, verbatim messages, typed error carriers, twice-run equality) | VALID |
| internal/api/lint_link_matrix_test.go | LINT-ENG-02/03/04 | 10 | 0 | 0 | Value + behavioral (error types through actionV1; numeric exit codes from exec.ExitError; byte-exact golden) | VALID |

Disabled tests on requirements: 0. Circular patterns: 0 (expected values come from the external etalon/lab oracle and the sealed token.go contract, hand-computed then cross-checked — not from the system under test). Insufficient assertions: 0.

### Prohibition Dispositions (flagged items, ADR-550 D4)

All 7 plan-flagged prohibitions were judged by this verifier (non-authoritative LLM-judge verdicts; deterministic evidence cited). **unverified-prohibition — human review recommended:** completion reads "complete with 7 flagged prohibitions", none silently passed.

| # | Prohibition | Verdict | Evidence |
|---|-------------|---------|----------|
| 1 | Classification never parses compiler diagnostic message text (02-01) | SATISFIED (judge) | Full read of link.go + link_test.go: classification uses the typed `missing` flag and behavioral `protoast.Parse`; grep for "was not found or had errors" clean in both files |
| 2 | Resolver performs no filesystem access of its own; `protoc.WithStandardImports` never used (02-01) | SATISFIED (judge) | link.go resolver delegates exclusively to `c.openImportFile`; grep clean for `WithStandardImports`/`os.Stat`/`filepath.Walk` in link.go and link_test.go |
| 3 | Linker dormant: internal/core/lint.go and internal/api/lint.go unchanged (02-01, D-04) | SATISFIED (deterministic) | `git diff --stat b5c26df..HEAD` empty for internal/core/lint.go, internal/api/lint.go, internal/api/lint_v1.go, internal/core/proto_info_read.go, internal/core/dom.go, internal/easypc; `linkSources` referenced only by tests |
| 4 | Precedence is behavioral (per-file parse), never a message-keyword scan (02-02) | SATISFIED (judge) | `hasPerFileSyntaxError` calls `protoast.Parse` (link.go:95-102); no keyword matching in tests |
| 5 | Resolver gained no filesystem search to "fix" blocked/transitive cells (02-02) | SATISFIED (judge) | Same evidence as #2; blocked_import cell proves the denial flows to the ErrNotExist fall-through with empty importRoots |
| 6 | Matrix did not wire the classifier into `Lint()` or alter exit routes (02-03) | SATISFIED (deterministic) | Dormancy diff empty; exit-route code at lint.go:85-91 unchanged since phase base; zero production commits in 02-03 (all three commits are `test(02-03)`) |
| 7 | In-process tests never call os.Exit / os.Chdir / os.Setenv on the parent (02-03) | SATISFIED (deterministic) | grep clean in lint_link_matrix_test.go; subprocess isolation via per-child `exec.Cmd.Dir`/inherited Env |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | Debt markers (TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER), placeholders, empty implementations, skipped tests: all clean across the six phase files | — | — |

Dependency gate: `go mod tidy && git diff --exit-code go.mod go.sum` exits 0 (re-run by this verifier). `./bin/golangci-lint run ./internal/core/... ./internal/api/...` → 0 issues. `git status --porcelain cmd/ internal/` clean after the subprocess run (no binary artifacts).

### Known Issues (not phase gaps)

- Pre-existing `-race` data race + ~10 test failures in `internal/adapters/plugin` (third-party wazero version cache), verified identical at phase base `a77a13d` (deferred-items.md, Phase 1 Acknowledged Gaps). Not a Phase 2 regression; needs a dedicated fix before the milestone's final acceptance gate (Phase 7).
- Review findings IN-01 (resolver ignores ctx) and IN-02 (COMPILE diagnostics bypass ignore mechanism) — advisory, recorded open in 02-REVIEW-DISPOSITION.md, routed to the first Phase 5 code touch.

### Human Verification Required

### 1. Ratify flagged assumption A1 — class-1 (COMPILE) verification level before the Phase 5 flip

**Test:** Accept or reject the two-level interpretation of ROADMAP SC-4: class-1 verified at classifier level (02-01/02-02 cells) + `printIssues` golden (02-03); the identical matrix re-runs through Action in Phase 5 (02-03 coverage D5, `human_judgment: true`).
**Expected:** Explicit ratification, or rejection converts SC-4's unexercised leg into a gap for a Phase 5-scoped closure plan.
**Why human:** Policy interpretation under the D-04 dormancy invariant; automation cannot decide it (orchestrator instructed explicit disposition, not a silent pass).

### 2. Resolve the MVP-mode discrepancy

**Test:** Ratify proceeding goal-backward although `Mode: mvp` and the goal is not a User Story (`user-story.validate` → false), following the Phase 1 precedent — or run `/gsd:mvp-phase 02` and re-verify.
**Expected:** A recorded decision on the phase's mode framing.
**Why human:** Mode/gate policy; surfaced per verify-mvp-mode.md.

### 3. Warn-drop re-pin (flagged assumption A3)

**Test:** When a natural warn-producing fixture exists (Phase 5 matrix re-run), exercise the `IsWarn` drop branch with a real warn diagnostic.
**Expected:** Warn diagnostics never become COMPILE issues.
**Why human:** No warn fixture exists in this phase; the branch is unexercised code (link.go:45-47).

### 4. Retention of the etalon-verbatim defensive fallback branch

**Test:** Confirm the `!errors.Is(err, protoc.ErrCompilationFailed) -> (nil, nil)` branch (link.go:55) stays with its documented not-unit-triggerable comment.
**Expected:** Branch retained, documentation-only coverage accepted deliberately.
**Why human:** Unexercised defensive path; plan 02-02 documents it as deliberate (research Pitfall 1).

### Gaps Summary

No FAILED truths, no MISSING/STUB artifacts, no NOT_WIRED key links, no blockers. Every behavioral claim I re-ran passed: 16/16 classifier cells under `-race`, 10/10 Action-path/subprocess/golden cells including real exit codes 2/1/0 through a freshly built binary. Dormancy (D-04) and the zero-dependency invariant are deterministic-verified against the phase base commit. The phase is not `gaps_found`.

It is `human_needed` for four items: the A1 ratification (SC-4's class-1-through-Action leg is real but unreachable while dormant — precisely the flagged assumption the phase carried), the MVP-mode framing decision, the warn-drop re-pin, and the defensive-branch retention. Three scope items are legitimately deferred to Phase 5 (WR-01, WR-02, the full class-1 matrix re-run) with matching Phase 5 goal text; none blocks this phase's classification-contract goal.

---

_Verified: 2026-11-19T05:44:31Z_
_Verifier: Claude (gsd-verifier)_
