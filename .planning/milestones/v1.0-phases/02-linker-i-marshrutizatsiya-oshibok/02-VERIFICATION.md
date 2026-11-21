---
phase: 02-linker-i-marshrutizatsiya-oshibok
verified: 2026-11-19T11:33:12Z
status: passed
score: 10/10 must-haves verified
behavior_unverified: 0
overrides_applied: 1
overrides:
  - must_have: "SC-4: fixture matrix of all three error classes passes with CLI exit codes verified through the Action path — class-1 (COMPILE) leg through Action"
    reason: "Human ratified the two-level interpretation (02-UAT.md Test 1, result: pass): class-1 is verified at classifier level (16-cell etalon matrix, re-run green by this verifier) + printIssues golden (3 cells green); the identical matrix re-runs through Action in Phase 5 — the leg is physically unreachable while the linker is dormant (D-04)"
    accepted_by: "user via 02-UAT.md (UAT record, result: pass)"
    accepted_at: "2026-11-19T10:35:38Z"
re_verification:
  previous_status: human_needed
  previous_score: 8/10
  gaps_closed:
    - "G-02-3: Warn diagnostics dropped per the etalon policy (research A3) — the IsWarn branch is now exercised by TestLinkSourcesWarnDrop (4 cells, exact assertions, green under -race, run by this verifier)"
    - "Human item 1 (A1 two-level class-1 interpretation) — RATIFIED by the human in 02-UAT.md (result: pass), recorded here as an override"
    - "Human item 2 (MVP-mode discrepancy) — RATIFIED by the human in 02-UAT.md (result: pass): goal-backward verification per the Phase 1 precedent"
    - "Human item 4 (defensive fallback branch retention) — CONFIRMED by the human in 02-UAT.md (result: pass); branch re-verified present verbatim at internal/core/link.go:48-51"
  gaps_remaining: []
  regressions: []
covered_files: [".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-01-PLAN.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-01-SUMMARY.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-02-PLAN.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-02-SUMMARY.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-03-PLAN.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-03-SUMMARY.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-04-PLAN.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/02-04-SUMMARY.md", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-01-task1.json", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-02-task1.json", ".planning/phases/02-linker-i-marshrutizatsiya-oshibok/tdd-red-evidence-02-04-task1.json", "internal/api/lint_link_matrix_test.go", "internal/core/link.go", "internal/core/link_test.go", "internal/core/testdata/link/multi_file/a.proto", "internal/core/testdata/link/multi_file/b.proto", "internal/core/testdata/link/tab_utf8/broken.proto"]
covered_digest: "v4:sha256:5be5d5fbb8fba45014b3bb51ec0cc55e0addc785c79ca73eea5fd1ee3887b841"
deferred:
  - truth: "Import open/read failures classify as COMPILE instead of the documented fallback (review finding WR-01)"
    addressed_in: "Phase 5"
    evidence: "Phase 5 goal: '`easyp lint` работает на компиляторе; старый lint-контур удалён; формат вывода, пути, коды выхода и все остальные подсистемы не изменились — единственная точка пользовательски видимого изменения'; 02-REVIEW-DISPOSITION.md routes WR-01 to the Phase 5 flip plan (dormant now, real semantic gap only post-flip)"
  - truth: "lintResolver serves corpus files from memory, bypassing importFileAllowed/sourceFileOpen (review finding WR-02)"
    addressed_in: "Phase 5"
    evidence: "02-REVIEW-DISPOSITION.md: 'tie to Phase 5 wiring where importFileAllowed joins the linker path'; dormant now, edge-reachable only with restrictive source_roots"
  - truth: "Full class-1 (COMPILE) matrix re-run through the Action path with CLI exit codes (the ratified-A1 Phase 5 leg)"
    addressed_in: "Phase 5"
    evidence: "Phase 5 goal owns the user-visible flip ('единственная точка пользовательски видимого изменения'); A1 two-level interpretation ratified in 02-UAT.md Test 1 (result: pass)"
  - truth: "Natural-warn fixture re-pin of the IsWarn drop (informational; the drop policy is already proven by the synthetic-warn seam per the user's G-02-3 decision)"
    addressed_in: "Phase 5"
    evidence: "02-UAT.md deferred section and 02-04 SUMMARY 'Next Phase Readiness': the natural-warn re-pin remains a Phase 5 matrix re-run informational item, unchanged by the synthetic closure the user chose"
---

# Phase 02: Линкер и маршрутизация ошибок — Verification Report

**Phase Goal:** Линковка корпусов через `easypc` с корректной трёхсторонней классификацией отказов: COMPILE-замечания, откат на пофайловый разбор при отсутствующем импорте, ошибки чтения/синтаксиса как в main
**Verified:** 2026-11-19T11:33:12Z
**Status:** passed
**Re-verification:** Yes — after gap closure (plan 02-04 closed G-02-3; the human ratified UAT items 1, 2 and 4)

## MVP Mode Disposition

ROADMAP sets `Mode: mvp` but the goal is not a User Story (`user-story.validate` → false). This discrepancy was surfaced in the previous verification and **ratified by the human** (02-UAT.md Test 2, result: pass): this milestone is an internal compiler migration; the Phase 1 precedent applies — goal-backward verification against the four ROADMAP success criteria. No "User Flow Coverage" section is emitted (no user story exists to decompose). Disposition recorded; no open decision remains.

## Re-Verification Scope

Prior verification (2026-11-19T05:44:31Z, human_needed, 8/10) carried two behavior-unverified truths and four human items. Since then: (a) the human ratified UAT items 1 (A1 two-level interpretation), 2 (MVP-mode precedent) and 4 (defensive-branch retention) and flagged item 3 as a gap (G-02-3, synthetic warn now instead of Phase 5 deferral); (b) plan 02-04 executed (RED `7be8b07` → GREEN `2066ef5` → contention pin `dc49ec1` → docs `bb10845`) extracting the `diagCollector` seam with `TestLinkSourcesWarnDrop`. This re-verification: full 3-level check of the changed code (`internal/core/link.go`, `internal/core/link_test.go`), behavioral re-runs of every matrix, regression proof on the 16 pre-existing cells, dormancy/seal re-audit, and disposition of all four prior human items.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC-1: Corpus failing linking with compile errors (not missing import) yields COMPILE remarks (position + compiler message) in the existing issues format; rules do not run | ✓ VERIFIED | `TestLinkSources/compile_error` pins the etalon cell verbatim (test.proto 4:3, "Required fields are not allowed in proto3.", zero Filename/Offset/SourceName); re-run by this verifier — 16/16 cells + parent PASS under `-race`; `TestPrintIssuesCompileGolden` pins the printer format byte-for-byte (3 cells green) |
| 2 | SC-2: Missing import falls back to per-file parsing; typed `OpenImportFileError` reaches the existing exit-2 route | ✓ VERIFIED | `missing_import`, `transitive_missing_depth2`, `blocked_import` cells assert `(nil, nil)` fallback plus `require.ErrorAs *OpenImportFileError`; in-process `TestLintLinkMatrix/missing_import_exit2_class` and subprocess `missing_import_code_2` (actual exit 2, "Cannot import file" + missing.proto on stderr) — all green in this verifier's run |
| 3 | SC-3: Read/syntax error returns a command execution error exactly as before | ✓ VERIFIED | `TestLintLinkMatrix/syntax_error_command_class` (plain error, `require.NotErrorAs *OpenImportFileError`) + subprocess `syntax_error_code_1` (actual exit 1 through the built binary) — green in this verifier's run |
| 4 | SC-4: Fixture matrix of all three error classes passes with CLI exit codes verified through the Action path — before engine insertion | ✓ VERIFIED (override) | Reachable legs re-proven by this verifier (missing→2, syntax→1, lint-issue→1, clean→0; 10/10 API cells green). The class-1 (COMPILE)-through-Action leg is physically unreachable while dormant (D-04); the human ratified the two-level interpretation (classifier cells + printIssues golden now, identical matrix through Action in Phase 5) — 02-UAT.md Test 1, result: pass. Recorded in frontmatter `overrides:`; the Phase 5 leg stays in `deferred:` |
| 5 | Compiling corpus (incl. WKT import) returns non-nil `*protoc.Linked` + nil issues; missing-import detection goes only through the typed resolver seam | ✓ VERIFIED | `compiling_with_wkt` + `wkt_resolves` cells green; `link.go:126-137` maps typed errors (`errors.As *OpenImportFileError` / `errors.Is os.ErrNotExist`) to the `missing` atomic + `protoc.ErrFileNotFound`; no diagnostic-text parsing anywhere in the file |
| 6 | COMPILE issues deterministically sorted by Path, Line, Col (D-03), stable across runs and across files | ✓ VERIFIED | `deterministic_order_twice` (require.Equal on full slices of two runs) + `multi_file_matrix_order` — green in this verifier's run |
| 7 | Missing-import matrix completeness: transitive at dep depth 2, blocked by `importFileAllowed`, WKT resolves, non-local "../x.proto" classifies COMPILE per etalon (research A4) | ✓ VERIFIED | `transitive_missing_depth2`, `blocked_import`, `wkt_resolves`, `non_local_import_compile` — all green in this verifier's run |
| 8 | COMPILE positions follow the compiler LineCol convention, pinned by a fixture that provably diverges from the Phase 1 D-07 rune contract | ✓ VERIFIED | `tab_utf8_position`: fixture `testdata/link/tab_utf8/broken.proto` — Line 6 / Column 37 (byte columns, tab stop 8) vs rune-column 19; green in this verifier's run |
| 9 | Import-cycle diagnostics pinned at Line 2 / Col 1 with the "recursively imports itself" chain | ✓ VERIFIED | `import_cycle` cell asserts the exact position and message substring; green in this verifier's run |
| 10 | Warn diagnostics dropped per the etalon policy (research A3) — was G-02-3, the phase's only gap | ✓ VERIFIED | Closed by executed plan 02-04: `diagCollector` (internal/core/link.go:103-112) is the exact handler `linkSources` registers (`protoc.WithErrorHandler(diagCollector(&diagsMu, &diags))`, link.go:44) — the test drives the production handler value, not a copy. `TestLinkSourcesWarnDrop` 4/4 cells green under `-race` in this verifier's run: `warn_dropped` (require.Empty after a warn post), `non_warn_collected` (drop is IsWarn-selective, exact one-element match), `warn_mix_order` (exactly the 2 error diagnostics in arrival order), `handler_race_mixed` (4 goroutines × 25 pairs: exact count 100, exact per-identity Line tallies, no warn leakage, `-race` clean). The collector slice is diagnostics' sole channel into the D-03 sort and COMPILE construction (link.go:40→62→70-80), so a dropped warn structurally cannot surface as COMPILE |

**Score:** 10/10 truths verified (9 direct + 1 override; 0 present-behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/core/link.go` | `CompileRuleName`, `(*Core).linkSources` three-outcome classifier, `diagCollector` warn-drop seam, `lintResolver` typed seam, `hasPerFileSyntaxError` | ✓ VERIFIED | All five symbols present and substantive (146 lines); `git show 2066ef5` confirms the extraction moved the inline closure body verbatim (early return on `d.IsWarn`, lock/defer-unlock/append) with no other classifier line changed; deliberately NOT wired into `Lint()` — D-04 dormancy re-verified: `git diff --stat b5c26df..HEAD` empty on internal/core/lint.go, internal/api/lint.go, internal/api/lint_v1.go, internal/core/proto_info_read.go, internal/core/dom.go, internal/easypc |
| `internal/core/link_test.go` | `TestLinkSources` 16-cell matrix + `TestLinkSourcesWarnDrop` (4 cells) | ✓ VERIFIED | 20 named cells; all green under `-race` in this verifier's own runs (16-cell regression proof is the zero-behavior-change evidence for the extraction; exact-count assertions throughout the new cells) |
| `internal/core/testdata/link/tab_utf8/broken.proto` | Tab+multi-byte-UTF-8 fixture with divergent LineCol | ✓ VERIFIED | Tracked in git; drives `tab_utf8_position` (6:37 vs rune 19) |
| `internal/core/testdata/link/multi_file/a.proto`, `b.proto` | Cross-file D-03 ordering corpus | ✓ VERIFIED | Tracked in git; drive `multi_file_matrix_order` |
| `internal/api/lint_link_matrix_test.go` | TestLintLinkMatrix (4 in-process), TestLintLinkMatrixExitCodes (3 subprocess), TestPrintIssuesCompileGolden (3 golden) | ✓ VERIFIED | 10/10 cells green in this verifier's run, including subprocess exit codes 2/1/0 through the built binary |
| `.planning/phases/02-.../tdd-red-evidence-02-04-task1.json` | Machine-validated RED evidence for the gap-closure plan | ✓ VERIFIED | Exists on disk; RED commit `7be8b07` verified as ancestor of GREEN `2066ef5` (`git merge-base --is-ancestor` → OK, re-run by this verifier) |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| internal/core/link.go | internal/easypc/protoc/link.go | `protoc.Link(ctx, paths, WithResolver, WithSyntaxTrees(true), WithErrorHandler)` | WIRED | link.go:41-44; no `WithStandardImports` anywhere |
| internal/core/link.go | internal/easypc/protoc/options.go | `protoc.WithErrorHandler(diagCollector(&diagsMu, &diags))` — identical function value at runtime and in test | WIRED | link.go:44; sole handler registration (02-04 key link, re-verified) |
| internal/core/link.go | internal/core/proto_info_read.go | `lintResolver.FindFileByPath -> c.openImportFile` (typed missing-import seam D-01) | WIRED | link.go:130; typed mapping at link.go:132-136 |
| internal/core/link.go | internal/core/dom.go | collected diags map to COMPILE `IssueInfo` (D-03); the handler is diagnostics' only entry into issue construction | WIRED | link.go:39/40 (sole channel) → 62 (sort) → 70-80 (issue construction) |
| internal/core/link_test.go | internal/core/link.go | same-package `TestLinkSourcesWarnDrop` drives synthetic `protoc.Diagnostic` values through `diagCollector` | WIRED | link_test.go:350-449; warn dropped, non-warn collected, mixed order, contention pin |
| internal/core/link_test.go | internal/core/testdata/link/ | fixture loading via `os.ReadFile` (`mustReadFixture`) | WIRED | tab_utf8 + multi_file cells |
| internal/core/link.go | internal/protoast/build.go | `hasPerFileSyntaxError -> protoast.Parse` | WIRED | link.go:88-95, invoked only on the failed-Link branch (link.go:56) |
| internal/api/lint.go | internal/core/dom.go | `errors.As *OpenImportFileError -> errExit(2, "Cannot import file")` | WIRED | unchanged this phase — byte-identical since phase base `b5c26df` (re-verified) |
| internal/api/lint_link_matrix_test.go | internal/api/lint_v1_test.go | actionV1 + cli.NewContext harness precedent | WIRED | subprocess cells build the real binary and assert numeric exit codes |

### Data-Flow Trace (Level 4)

Not applicable to rendering surfaces — the phase ships no UI/API route that renders dynamic data. The data-flow equivalents were re-traced: classifier inputs are real in-memory/testdata corpora; the synthetic warn fixtures in `TestLinkSourcesWarnDrop` are constructed values (the test is the producer by design, per the user's gap-closure decision — no natural producer exists in the sealed tree; grep re-confirmed: only `IsWarn` reads/propagation in internal/easypc, zero construction sites set it true). Golden expected values derive from the external etalon/lab oracle, not the system under test. The subprocess matrix flows real fixtures through the real binary.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Warn-drop seam (G-02-3 closure, 4 cells incl. race) | `go test -race -count=1 -v ./internal/core/ -run 'TestLinkSourcesWarnDrop'` | ok; all 4 named cells PASS (warn_dropped, non_warn_collected, warn_mix_order, handler_race_mixed) | ✓ PASS |
| 16-cell classifier matrix (extraction regression proof) | `go test -race -count=1 -v ./internal/core/ -run 'TestLinkSources$'` | ok; 17 PASS lines (parent + 16 cells), cell count unchanged | ✓ PASS |
| Action-path error classes + subprocess exit codes + golden | `go test -race -count=1 -v ./internal/api/ -run 'TestLintLinkMatrix\|TestPrintIssuesCompileGolden'` | ok; 10/10 cells PASS (exit 2/1/0 through built binary; byte-exact text/JSON golden) | ✓ PASS |
| Dormancy + seal + dependency invariant | `git diff --stat b5c26df..HEAD -- <dormant files>` empty; `git status --porcelain internal/easypc/` empty; `go mod tidy -diff` exit 0 (non-mutating); worktree clean under cmd/ internal/ go.mod go.sum | all clean | ✓ PASS |
| Whole-workspace suite | not run | pre-existing `-race` failure in `internal/adapters/plugin` (wazero), acknowledged at Phase 1 (deferred-items.md) — not a phase regression; targeted suites green | ? SKIP (documented) |

Full-suite constraint honored: targeted runs only; no per-truth full-suite filtering.

### Probe Execution

No probes declared (`scripts/*/tests/probe-*.sh` absent; no probe references in any PLAN/SUMMARY, re-checked for 02-04). SKIPPED (no probes exist for this phase); the plan's own verification suite was re-executed by this verifier — see spot-checks above.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| LINT-ENG-02 (REQ-2.2) | 02-01, 02-02, 02-03, 02-04 | Compile-failed files (not missing import) get COMPILE remarks; rules do not run on the corpus | ✓ SATISFIED | Classifier behavior test-proven (etalon cell + 16-cell matrix); warn-drop policy now executed-proof (TestLinkSourcesWarnDrop, closing G-02-3 under this same requirement ID — plan 02-04 frontmatter `requirements: [LINT-ENG-02]`); existing format pinned by golden; `linked == nil` when issues are returned so no rule can consume trees. REQUIREMENTS.md: Complete |
| LINT-ENG-03 (REQ-2.3) | 02-01, 02-02, 02-03 | Missing import -> per-file fallback, read/syntax errors reported as before, no COMPILE issues; `OpenImportFileError` reaches exit 2 | ✓ SATISFIED | Fallback proven for direct/transitive/blocked variants; typed carrier asserted in-process and actual exit 2 via subprocess. REQUIREMENTS.md: Complete |
| LINT-ENG-04 (REQ-2.4) | 02-02, 02-03 | Read/syntax error file returns command execution error as before | ✓ SATISFIED | In-process class assertion (plain error, NOT the carrier) + actual exit 1 via subprocess. REQUIREMENTS.md: Complete |

Orphaned requirements: none — REQUIREMENTS.md maps exactly LINT-ENG-02/03/04 to Phase 2; all three appear in plan frontmatter (02-04 adds the shared-ID completion for LINT-ENG-02). LINT-ENG-01/05 remain Phase 3 (Pending) as expected.

### Prohibition Dispositions (ADR-550 D4)

All 9 plan-flagged prohibitions carry a non-authoritative LLM-judge verdict with deterministic evidence cited. **unverified-prohibition — human review recommended:** completion reads "complete with 9 flagged prohibitions", none silently passed. Prohibitions 1-7 are unchanged from the prior verification (all re-audited: dormancy diff empty, greps clean). The two added by plan 02-04:

| # | Prohibition | Verdict | Evidence |
|---|-------------|---------|----------|
| 8 | The synthetic warn must not be produced by editing internal/easypc (sealed), and classification must never parse diagnostic message text (D-01); the only sanctioned drive is the extracted handler with synthetic values (02-04) | SATISFIED (judge + deterministic) | `git status --porcelain internal/easypc/` empty; `git diff b5c26df..HEAD -- internal/easypc` empty; grep confirms zero `IsWarn=true` construction sites in the sealed tree; link_test.go posts literal Diagnostic values, no text matching |
| 9 | The extraction must not alter classifier behavior: no new production parameters, no test-only injection API on linkSources, three-outcome classification/precedence/D-03 sort unchanged — the 16 pre-existing cells are the regression proof (02-04) | SATISFIED (judge + deterministic) | `git show 2066ef5 -- internal/core/link.go` shows only the closure→constructor move; `linkSources` signature unchanged (link.go:32, same unexported params); 16-cell matrix re-run green by this verifier |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | Debt markers (TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER), placeholders, empty implementations, skipped tests: clean across internal/core/link.go, internal/core/link_test.go, internal/api/lint_link_matrix_test.go | — | — |

`./bin/golangci-lint run ./internal/core/... ./internal/api/...` → 0 issues (re-run by this verifier). Dependency gate re-verified non-mutatingly (`go mod tidy -diff` exit 0). `git status --porcelain cmd/ internal/ go.mod go.sum` clean.

### Known Issues (not phase gaps)

- Pre-existing `-race` data race + test failures in `internal/adapters/plugin` (third-party wazero version cache) — Phase 1 acknowledged gap (deferred-items.md); needs a dedicated fix before the milestone's final acceptance gate (Phase 7). Not a Phase 2 regression.
- Review findings IN-01 (resolver ignores ctx) and IN-02 (COMPILE diagnostics bypass ignore mechanism) — advisory, recorded open in 02-REVIEW-DISPOSITION.md, routed to the first Phase 5 code touch.

### Human Verification Dispositions (all resolved — recorded, none open)

| Prior item | Disposition | Evidence |
| ---------- | ----------- | -------- |
| 1. Ratify flagged assumption A1 (two-level class-1 verification pre-flip) | RATIFIED by the human (result: pass) | 02-UAT.md Test 1; recorded as the frontmatter override for truth 4; Phase 5 leg in `deferred:` |
| 2. MVP-mode discrepancy (Mode: mvp, goal not a User Story) | RATIFIED by the human (result: pass) — Phase 1 precedent stands | 02-UAT.md Test 2 |
| 3. Warn-drop re-pin (G-02-3) | RESOLVED by executed gap-closure plan 02-04 (`gap_ids: [G-02-3]`, SUMMARY exists, RED→GREEN→contention commits verified); synthetic warn now per the user's decision; natural-warn re-pin remains a Phase 5 informational item | 02-UAT.md Test 3 (result: issue → gap), 02-04-PLAN/SUMMARY, `TestLinkSourcesWarnDrop` green in this verifier's run |
| 4. Retention of the etalon-verbatim defensive fallback branch | CONFIRMED by the human (result: pass); branch re-verified present verbatim at internal/core/link.go:48-51 | 02-UAT.md Test 4; this verifier's read of link.go |

### Gaps Summary

None. The phase's only gap (G-02-3, minor — the unexercised IsWarn drop branch) is closed by an executed RED→GREEN plan whose test I re-ran green under `-race`, with the 16 pre-existing cells proving the extraction behavior-preserving. Dormancy (D-04), the internal/easypc seal, and the zero-dependency invariant are deterministically re-verified against the phase base commit. All four prior human items carry recorded dispositions (three ratified by the human, one closed by executed plan). Three scope items remain legitimately deferred to Phase 5 (WR-01, WR-02, the ratified class-1-through-Action matrix re-run) with matching Phase 5 goal text; the natural-warn re-pin is informational and likewise Phase 5. Nothing blocks this phase's classification-contract goal: all ten must-haves are verified, no stubs, no broken links, no blockers, no open human items.

---

_Verified: 2026-11-19T11:33:12Z_
_Verifier: Claude (gsd-verifier)_
