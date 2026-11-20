---
phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti
plan: 05
subsystem: lint-engine
tags: [broken-import, error-path-fidelity, sentinel-errors, classifier-fallback, tdd, gap-closure, review-findings]

# Dependency graph
requires:
  - phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti (plan 02)
    provides: the WR-01 broken-import discrimination (PathError subset) and the three-outcome linkSources contract this plan extends to the sentinel shapes
  - phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti (plan 03)
    provides: the flip that made SourceRoots.OpenSourceFile / FSWalker.Open the production open funnels the resolver actually calls
  - phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti (review)
    provides: 05-REVIEW findings WR-01/IN-01/IN-02 with the exact locations and evidence, routed by 05-REVIEW-DISPOSITION (all three to 05-05, fix WITH WR-01)
provides:
  - sentinel-typed broken-import discrimination in lintResolver.FindFileByPath: typed errors.Is on sourceview.ErrUnsupported / ErrChanged / ErrNestedRepository forces the classifier fallback for the PathError-free funnel shapes, so a dangling-alias / changed-during-open / nested-repository import failure degrades to a command error carrying the funnel cause — REQ-2.4 parity semantics restored on the edge input (WR-01 closed)
  - sentinel_import_failure_falls_back probe cell in TestLintErrorPathFidelity: three funnel-shape-mirroring carriers, machine-validated RED then GREEN
  - IN-01 closed: the dead post-fallback readErrs re-check removed; hasPerFileSyntaxError cited as the sole owner of the COMPILE-vs-parse precedence
  - IN-02 closed: existInProto and breakingProtoInfoRead docs state the post-consolidation facts; no comment cites a deleted symbol or the pre-flip rename-map state
affects: [Phase 6 (new rules on the single engine), Phase 7 (external corpus gate — edge-input error fidelity), phase verification (05-REVIEW findings closed)]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 2535          # 10138 patch chars over e90dd46..7853240 at chars/4 (estimate 26000; micro gap closure — small realized surface)
  tasks: 3
  commits: 2           # MEASURED: git rev-list --count e90dd4690a31a5978eeda7d76fea33d68ef1fd5b..78532408f286d8a47692ff67757168dc654a9557 (the task commits; the metadata docs commit is excluded)
  plan_head_before: e90dd4690a31a5978eeda7d76fea33d68ef1fd5b
  plan_head_after: 78532408f286d8a47692ff67757168dc654a9557

# Tech tracking
tech-stack:
  added: []
  patterns: [typed sentinel classification at the resolver boundary (errors.Is only, never diagnostic text parsing — D-01), funnel-shape-mirroring test carriers (the production error construction replayed in-table so the probe proves the exact gap)]

key-files:
  created: []
  modified:
    - internal/core/link.go (FindFileByPath broken-flagging extended to the three sourceview sentinels; two-shape contract comment)
    - internal/core/lint_linked.go (dead readErrs re-check removed; hasPerFileSyntaxError ownership comment)
    - internal/core/proto_info_read.go (breakingProtoInfoRead doc restated)
    - internal/rules/import_used_linked.go (existInProto doc restated)
    - internal/core/lint_linked_matrix_test.go (sentinel_import_failure_falls_back probe cell, three carriers)

key-decisions:
  - "05-05: RED probe carriers mirror production funnel shapes exactly — the OpenLocal wrap of SourceRoots.OpenSourceFile over the local.go sentinel joins (:33+:176 dangling alias with the not-exist cause as %v text, :56 ErrChanged, :209 nested-repository ErrUnsupported+ErrNestedRepository join), no PathError in any chain"
  - "05-05: sentinel discrimination is additive typed errors.Is on the three sourceview sentinels with one broken.Store, raw error returned; the plain rejections stay raw — the etalon A4 non-local COMPILE classification preserved and pinned green"
  - "05-05: IN-01 criterion baseline reconciled — the hasPerFileSyntaxError call lives in link.go's linkSources, not lint_linked.go; the ownership citation is stated in-comment so the >=2 gate is met with truthful text"

patterns-established:
  - "Funnel-shape-mirroring carriers: a test carrier replays the production error construction exactly (the OpenLocal wrap over the named local.go join shape), with no PathError anywhere in the chain and text-only interpolation of causes that must not join the chain — the probe proves the gap, not an approximation"
  - "Two-shape contract comment: the discriminator's doc names both carrier shapes (PathError and the PathError-free sentinel joins) with local.go line cites, so the next WR-style review can diff premise against funnel reality"

requirements-completed: [LINT-COMPAT-01]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "WR-01 closed: an import whose open failure carries a sourceview sentinel shape (dangling source alias, changed-during-open target, nested-repository crossing — all PathError-free) forces the classifier fallback; the corpus degrades to the per-file path and the same funnel carrier re-surfaces through openImportFile's wrap chain as a command error satisfying errors.Is for each sentinel — never the misleading not-found-shaped COMPILE remark"
    requirement: LINT-COMPAT-01
    verification:
      - kind: unit
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintErrorPathFidelity/sentinel_import_failure_falls_back"
        status: pass
      - kind: integration
        ref: "tests/internal/core (whole package green under -race with the fix: A4 non-local COMPILE cell, missing-import matrix, warn-drop, compile_bypasses_suppression, ignored_broken_dependency_not_suppressed, blocked_corpus_file_errors, unreadable_import_falls_back, find_by_path_honors_ctx all green)"
        status: pass
    human_judgment: false
  - id: D2
    description: "IN-01 closed: the dead post-fallback readErrs re-check is gone from lint_linked.go; the COMPILE-vs-per-file-parse precedence is owned solely by linkSources' hasPerFileSyntaxError sweep, with the ownership comment in place (behavior-neutral deletion proven by the green suites)"
    requirement: LINT-COMPAT-01
    verification:
      - kind: other
        ref: "command: grep -c 'len(compileIssues) > 0 &&' internal/core/lint_linked.go (0) && grep -c 'hasPerFileSyntaxError' internal/core/lint_linked.go (2) && grep -c 'hasPerFileSyntaxError' internal/core/link.go (3, the call) && go test -race -count=1 ./internal/core/ (ok)"
        status: pass
    human_judgment: false
  - id: D3
    description: "IN-02 closed: existInProto's doc states the post-consolidation reclaim of the etalon name and breakingProtoInfoRead's doc references protoInfoFromSource and the deleted-twin history — no comment cites a deleted symbol or the pre-flip rename-map state"
    requirement: LINT-COMPAT-01
    verification:
      - kind: other
        ref: "command: grep -c 'lintProtoInfoRead' internal/core/proto_info_read.go (0) && grep -c 'protoInfoFromSource' internal/core/proto_info_read.go (>=1) && grep -c 'Linked suffix (rename map)' internal/rules/import_used_linked.go (0) && grep -c 'reclaimed the etalon name' internal/rules/import_used_linked.go (1)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Zero behavioral collateral: the 73-finding golden TestLintGoldenFullRegistry byte-identical, the api COMPILE/exit-code matrix and stdout/print goldens unchanged, the four breaking-owned files byte-frozen, vet and golangci clean on the touched packages"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/rules/golden_full_registry_test.go#TestLintGoldenFullRegistry"
        status: pass
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrix (+ExitCodes) and tests/internal/api/lint_stdout_golden_test.go#TestLintStdoutGolden"
        status: pass
      - kind: other
        ref: "command: git diff --exit-code -- internal/core/breaking_check.go internal/core/breaking_checker.go internal/core/breaking_profiles.go internal/core/proto_comments.go (empty) && go vet ./internal/core/... ./internal/rules/... (0) && bin/golangci-lint run ./internal/core/... ./internal/rules/... (0 issues)"
        status: pass
    human_judgment: false

# Metrics
duration: 10 min
started: 2026-11-20T21:42:21Z
completed: 2026-11-20T21:52:45Z
status: complete
---

# Phase 05 Plan 05: WR-01 Sentinel Gap Closure Summary

**Typed sentinel discrimination closes WR-01: PathError-free open-funnel failures (dangling source alias, changed-during-open target, nested-repository crossing) now force the classifier fallback into a command error carrying the funnel cause — never a false not-found-shaped COMPILE remark — with the dead readErrs re-check removed and both stale comments refreshed, goldens byte-identical and the breaking-owned files frozen**

## Performance

- **Duration:** 10 min
- **Started:** 2026-11-20T21:42:21Z
- **Completed:** 2026-11-20T21:52:45Z
- **Tasks:** 3
- **Files modified:** 5 (4 production + 1 test) + planning artifacts (RED evidence record, VALIDATION row)

## Accomplishments

- **WR-01 closed (T-05-08 mitigation).** `lintResolver.FindFileByPath`'s openImportFile error branch now sets `broken` for the PathError-free sentinel shapes the production open funnels return — typed `errors.Is` on `sourceview.ErrUnsupported`, `sourceview.ErrChanged` and `sourceview.ErrNestedRepository` (one store for any of the three, the raw error still returned). A dangling-alias / non-regular / changed / nested-repository import failure now degrades the run exactly like main before the flip: the classifier falls back to the per-file path and the same carrier re-surfaces through `openImportFile`'s wrap chain as a command error. The misleading not-found-shaped COMPILE remark that skipped every rule corpus-wide is unreachable for these inputs.
- **The etalon A4 classification stays pinned.** The plain rejections (openImportFile's "invalid import path" and the missing classification) still flow raw; the A4 non-local COMPILE cell and the full missing-import matrix stayed green through the fix. Typed checks only — no diagnostic text parsing, no path-locality re-check added to the resolver.
- **IN-01 closed.** The post-fallback `readErrs` re-check at lint_linked.go:73 is deleted; the else-if keeps only the compileIssues term, with the ownership comment citing `hasPerFileSyntaxError` (the sweep over identical sources plus parseOnTheirOwn's identical re-parse leaves readErrs all-nil on the only path where the branch evaluates).
- **IN-02 closed.** `existInProto`'s doc states the post-consolidation reclaim of the etalon name; `breakingProtoInfoRead`'s doc references `protoInfoFromSource` and the deleted-twin history. No comment cites a deleted symbol or the pre-flip rename-map state.
- **The RED probe is machine-validated.** `sentinel_import_failure_falls_back` (three carriers) failed on its own `require.Error` at HEAD with `RED_EVIDENCE_OK` from `gsd_run check tdd-red-evidence`, then went GREEN in the fix commit; every pre-existing cell in `TestLintErrorPathFidelity` passed in both runs.

## Sentinel-shape map, type-alias fact, IN-01 deadness proof, comment texts

(TASK 3 record per the plan's action block.)

**Sentinel-shape map** — which local.go line produces which carrier, all wrapped by SourceRoots.OpenSourceFile's `OpenLocal: %w` (sources.go:151):

| Carrier (test row) | local.go shape | Chain |
|---|---|---|
| dangling_alias_unsupported | :33 (Resolve branch) + :176 (localResolutionError) | `OpenLocal:` over `errors.Join(Resolve: <dangling-alias message>, nil)`; the message wraps `ErrUnsupported` with `%w` and interpolates the not-exist cause as `%v` text only |
| changed_during_open | :56 (SameFile mismatch) | `OpenLocal:` over `errors.Join(ErrChanged, nil, nil)` — the nil Close errors are discarded by errors.Join |
| nested_repository_crossing | :209 (CheckLocalResolution) | `OpenLocal:` over the nested-repository message wrapping `errors.Join(ErrUnsupported, ErrNestedRepository)` |

**Type-alias fact bounding the gap.** `os.PathError` is a type ALIAS for `io/fs.PathError` (`go doc os.PathError`), so the existing `errors.As(err, &pathErr)` already caught the funnels' PathError-wrapped sentinel shapes (sourceview/view.go:248/:267). The gap was exactly the PathError-free shapes: the `errors.Join(ErrUnsupported, ...)` non-regular-target join (local.go:42), the dangling-alias Resolve wrap (:176), the `errors.Join(ErrChanged, ...)` joins (:56/:165), and the nested-repository join (:209). The probe carriers carry no PathError — otherwise the cell would pass at HEAD and prove nothing.

**IN-01 deadness proof chain.** Reaching the else-if requires `linkSources` to have returned `(nil, compileIssues>0)`, which per link.go:58-69 requires `hasPerFileSyntaxError(paths, sources) == false` — a `protoast.Parse` over the identical (unmutated until line 102) `sources` that `parseOnTheirOwn` then re-parses. So `ContainsFunc(readErrs, ...)` was always false there and only `len(compileIssues) > 0` decided; the concern is fully owned by `hasPerFileSyntaxError` inside linkSources and by the line-67 entry guard for read errors.

**The two comment refresh texts.**

1. `internal/rules/import_used_linked.go` (existInProto): "the etalon named it existInProto, and the port reclaimed the etalon name once main's byte-frozen import_used.go was deleted in the 05-03 flip. The reading is main's: the extend fields, the top-level message names and the top-level enum names."
2. `internal/core/proto_info_read.go` (breakingProtoInfoRead): "formerly the twin of the lint read constructor deleted in the 05-03 flip — the lint path now reads through protoInfoFromSource. Behaviorally identical in Phase 1; only the read-model type differs."

## TDD Gate Compliance

- Plan `type: tdd`; Tasks 1 and 2 carry `tdd="true"`.
- **RED:** `test(05-05)` `dbf6005` — test-only commit (99 insertions in internal/core/lint_linked_matrix_test.go, zero production edits). `go test -race -count=1 ./internal/core/ -run TestLintErrorPathFidelity -v` exited 1 with all three probe rows failing on their planned `require.Error` ("An error is expected but got nil" — Lint returned the COMPILE misclassification with a nil error) while all four pre-existing cells passed. Evidence record `.planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/tdd-red-evidence-05-05-task1.json` machine-validated `RED_EVIDENCE_OK` (go-test format, target `TestLintErrorPathFidelity/sentinel_import_failure_falls_back`, 9 tests: 4 pass / 5 fail). Semantic assessment: VALID RED — the failure mode matches flagged assumption 1 exactly; no build/fixture/load/unrelated failure.
- **GREEN:** `feat(05-05)` `7853240` — the four production edits (sentinel discrimination, dead-term removal, two comment refreshes); the probe cell green and every standing core cell green under `-race`.
- **REFACTOR:** not needed — no cleanup beyond the plan's enumerated edits.
- **Gate sequence:** `test(05-05)` precedes `feat(05-05)` in ancestry ✓.

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): sentinel probe cell fails at HEAD** - `dbf6005` (test; lint_linked_matrix_test.go only, 99 insertions)
2. **Task 2 (GREEN): sentinel discrimination + dead-term removal + comment refresh** - `7853240` (feat; 4 production files, +32/−14)
3. **Task 3: closure battery — golden byte-identity, breaking dormancy, standing gates** - verification-only (core+rules `-race` incl. the golden, api matrix/goldens, vet, freeze diff, golangci — all green on the committed state; no file changes, no commit)

**Plan metadata:** docs commit (this SUMMARY + STATE/ROADMAP/REQUIREMENTS/VALIDATION + the RED evidence record).

## Files Created/Modified

- `internal/core/link.go` - FindFileByPath's openImportFile error branch: broken-flagging extended to the three sourceview sentinels via typed errors.Is (one store, raw error returned); the stale single-shape premise comment rewritten to the two-shape contract with local.go line cites; `internal/sourceview` import added
- `internal/core/lint_linked.go` - the COMPILE-vs-parse else-if reduced to `len(compileIssues) > 0`; the dead readErrs re-check removed; hasPerFileSyntaxError ownership comment added (entry guard and slices import untouched)
- `internal/core/proto_info_read.go` - breakingProtoInfoRead doc restated to the post-consolidation fact (openImportFile's A4 boundary untouched)
- `internal/rules/import_used_linked.go` - existInProto doc restated to the post-consolidation reclaim fact
- `internal/core/lint_linked_matrix_test.go` - sentinel_import_failure_falls_back probe cell: three funnel-shape-mirroring carriers with per-row funnel-origin comments (local.go line refs), per-row assertions require.Error / require.Nil / require.ErrorIs / assert.Contains(dep.proto on the dangling row)
- `.planning/phases/05-…/tdd-red-evidence-05-05-task1.json` - machine-validated RED evidence record (RED_EVIDENCE_OK)
- `.planning/phases/05-…/05-VALIDATION.md` - 05-05 task row filled

## Decisions Made

- The probe carriers replay the production error construction exactly (the OpenLocal wrap over the named local.go join shape) instead of approximating with bare sentinels: a bare `sourceview.ErrUnsupported` would exercise the fallback but not prove the real funnels' shape reaches it. The not-exist cause of the dangling row is interpolated as text only (`%v`), mirroring localResolutionError — a wrapped `fs.ErrNotExist` would trigger the missing branch and prove the wrong path (flagged assumption 3; `errors.Join` discards nil Close errors, verified).
- The sentinel extension is additive: the existing missing classification and PathError classification are byte-untouched; the raw error is still returned after `broken.Store(true)`, because the classifier's fallback condition already reads `resolver.broken.Load()`.
- IN-01's comment states the ownership decision explicitly ("hasPerFileSyntaxError owns the COMPILE-vs-parse precedence solely: the dead readErrs re-check (IN-01) is gone") so a future editor cannot read the removed term back as load-bearing — the review's fix intent.
- Zero edits to the four breaking-owned files; no change to the linkSources fallback condition, the COMPILE issue construction, the D-03 sort, or any issue text — the goldens stayed byte-identical by construction.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] IN-01 acceptance-criterion baseline reconciled — the hasPerFileSyntaxError call is not in lint_linked.go**
- **Found during:** Task 2 (criteria re-run)
- **Issue:** the criterion expects `grep -c 'hasPerFileSyntaxError' internal/core/lint_linked.go` >= 2 "(the call plus the ownership comment)" [HEAD=1], but the call lives in `internal/core/link.go`'s linkSources (:66) — HEAD's lint_linked.go carries 0 occurrences, and the planned comment adds exactly 1. The "(the call plus the ownership comment)" parenthetical mis-located the call.
- **Fix:** stated the ownership decision in-comment a second time ("hasPerFileSyntaxError owns the COMPILE-vs-parse precedence solely: the dead readErrs re-check (IN-01) is gone") — truthful, informative text that satisfies the >= 2 gate with the review's intended ownership citation, not grep gaming. The intent-true invariant holds: comment cites the owner (2 mentions in lint_linked.go), the call is present in link.go (3 mentions: call + 2 doc references).
- **Files modified:** internal/core/lint_linked.go (comment text only)
- **Verification:** grep counts 2 (lint_linked.go) / 3 (link.go); core suite green under -race; CHAIN_EXIT=0 on the full verify chain
- **Committed in:** 7853240

**2. [Rule 3 - Blocking] RED carrier's text-interpolated cause uses *os.PathError (type-identical to the plan's *fs.PathError) to avoid an import collision**
- **Found during:** Task 1 (first RED run — build failure)
- **Issue:** the plan's behavior block sketches the dangling carrier's interpolated cause as `&fs.PathError{…}` (io/fs), but the test file already imports `internal/fs/fs` as `fs` — adding `"io/fs"` collides and the package does not build. The repo convention (generate.go) aliases it `stdfs`.
- **Fix:** used `&os.PathError{Op: "lstat", Path: "dep.proto", Err: os.ErrNotExist}` — `os.PathError` is a type ALIAS for `io/fs.PathError` (verified via `go doc os.PathError`), so the value is type-identical to production's Resolve cause, and because it is interpolated as `%v` TEXT it never joins the wrap chain (the flagged-assumption-3 constraint holds either way). No alias needed, one fewer import.
- **Files modified:** internal/core/lint_linked_matrix_test.go (carrier literal + import block)
- **Verification:** the second RED run built and failed exactly on the three rows' require.Error (RED_EVIDENCE_OK on the recorded run)
- **Committed in:** dbf6005

---

**Total deviations:** 2 auto-fixed (2 Rule 3 blocking: 1 criterion-baseline reconciliation, 1 test-authoring import collision)
**Impact on plan:** Both were verification-plumbing/test-authoring corrections inside the declared surface; the production behavior, the probe's semantics, and every gate outcome are exactly as planned. No scope creep.

## Issues Encountered

None — the RED failed on the intended assertion class on the second run (the first attempt was the documented import-collision build failure), and the GREEN chain passed on the first full run.

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components; the probe cell drives the real engine end to end.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- **All three 05-REVIEW findings are closed** per 05-REVIEW-DISPOSITION (0 open-after-plan): WR-01 with typed sentinel discrimination + machine-validated probe, IN-01 dead-term removal, IN-02 comment refreshes. The phase is ready for verification; the 05-REVIEW open-finding counter is now 0.
- **Standing constraints intact:** TestReadSourceOnce and the suppression-centrality matrix green in the same `-race` runs; the 73-finding golden byte-identical; the four breaking-owned files byte-frozen; REQ-2.4 error-path parity now holds on both carrier shapes (PathError subset and sentinel subset).
- No blockers.

## Self-Check: PASSED

- All 5 created/modified files exist on disk (4 production + 1 test) plus the SUMMARY and the RED evidence record.
- Both task commits are ancestors of HEAD: `dbf6005` (test RED), `7853240` (feat GREEN); `git rev-list --count e90dd46..7853240` = 2, matching the frontmatter `commits:`.
- `gsd_run check evaluation-scope --plan 05-05 --commits-only` resolves both commits on this branch (source: plan-subjects), no missing files.
- RED evidence machine-validated `RED_EVIDENCE_OK` before GREEN; the plan-level `<verification>` battery (core+rules -race, api matrix/goldens, vet, freeze diff, golangci) all green on the committed state.

---
*Phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti*
*Completed: 2026-11-20*
