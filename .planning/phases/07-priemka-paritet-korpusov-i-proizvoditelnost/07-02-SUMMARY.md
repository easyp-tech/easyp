---
phase: 07-priemka-paritet-korpusov-i-proizvoditelnost
plan: 02
subsystem: testing
tags: [lint-parity, acceptance-gates, baseline-oracle, git-worktree, multiset-compare, perf-protocol, subprocess-tests, lab-py]

# Dependency graph
requires:
  - phase: 05 (the flip, 4238fbe)
    provides: single linked engine at HEAD — main's behavior exists only as a binary from the pre-milestone commit, hence the baseline-oracle design
  - phase: 06-novye-pravila-i-regeneratsiya-shem-mcp
    provides: 44-rule registry (5be9f7e), the teeth fixtures under testdata/ and the 06-03 corpus/exit-code contract parity-evidence-06-03.json
  - phase: 07-01 (1f82b61)
    provides: race-free wazero default factory — the full -race suite (including ./internal/api) runs green on the state this plan tests
provides:
  - TestAcceptanceGates (internal/api) — the four-gate compare machine-asserted on the teeth corpus, x2 per binary, permanent in the normal suite and independent of corpora availability
  - buildBaselineBinary — the RATIFIED 1e9a596 oracle as a reproducible build procedure (commit-presence gate -> detached worktree -> plain no-race build -> t.Cleanup worktree teardown), reusable by 07-03
  - /tmp/p7 instruments (compare.py, perf.py, run-parity.sh — throwaway, NOT committed, rebuildable from the plan spec) self-tested with two negative controls; parity mode encodes the exact unblock-day sequence
  - 07-VALIDATION.md gates row pointing at the real test home (internal/api) and command
affects: [07-03 (SC3 battery + parity-evidence-07.json/perf-evidence-07.md consume these gates and instruments; corpora disposal checkpoint)]

# Actuals (#2632) — pairs with the plan's estimate to calibrate future estimates.
actuals:
  tokens: 4919    # chars/4 over the realized diff (19678 chars, +425/-1 across 2 files)
  tasks: 2
  commits: 1      # MEASURED: git rev-list --count gsd-plan-head-before-07-02..HEAD; Task 2's instruments are throwaway /tmp artifacts (06-03 precedent) — nothing to stage by design
plan_head_before: bc3ac3893a37cd32997b180295ac4a72e9d34ead
plan_head_after: 0669aab

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Baseline oracle pattern: sync.OnceValues worktree build at a ratified commit with a loud commit-presence gate (fail, never skip), plain no-race build, and cleanup-guaranteed worktree teardown (t.Cleanup + failure-path self-removal — OnceValues caches errors, so an unclean failure path would leak forever)"
    - "Portable /usr/bin/time wrapper: probe BSD -l, fall back to GNU -f '%e %M' reading the same wall+RSS quantities (this machine is GNU Time; lab.py's BSD assumption patched compatibly, not semantically)"
    - "Anti-vacuity compare pattern: Gate B zero-COMPILE asserted inside load; Gate C both-direction multiset identity + non-empty shared set; NEW3 filter is one-sided BY DESIGN (only the side with the bigger registry) — same-binary Gate A compares need --no-filter"

key-files:
  created:
    - internal/api/acceptance_gates_test.go
  modified:
    - .planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-VALIDATION.md
  throwaway-not-committed:
    - /tmp/p7/compare.py
    - /tmp/p7/perf.py
    - /tmp/p7/run-parity.sh

key-decisions:
  - "A2 RATIFIED and materialized: baseline = 1e9a596 (verified zero Go-source delta to the first milestone commit 8f969d8; docs-only a0a0ed9/26924b1 intervene) — provenance recorded in the test's comments and echoed by the scripts into /tmp/p7/out"
  - "buildBaselineBinary is a sync.OnceValues returning a result struct (OnceValues is (T, error)-shaped) with the worktree dir surfaced for the test's t.Cleanup; the build-failure path self-removes its worktree registration because a cached OnceValues error would otherwise leak it forever"
  - "Gate A evaluated twice over: raw (snapshots are written sorted, so byte equality IS the sorted multiset equality) and via compare.py --no-filter; the one-sided NEW3 filter phantom-diffs shared NEW3 rows on same-binary compares"
  - "RESEARCH's '[VERIFIED] /usr/bin/time -l' is false on this machine (GNU Time rejects -l) — portable wrapper in the instruments; parity mode carries an idempotent lab.py lint() GNU-time compatibility patch that swaps ONLY the time flag + RSS parse for the GNU form reporting the same quantities (stand semantics untouched; patch anchors pre-validated against a copy)"
  - "Teeth-corpus actual counts are 10 (baseline) / 14 (HEAD); RESEARCH's '18 issues / 20 issues' narrative is not reproducible as stated (its arithmetic contradicts its own only-main-0 evidence) — no gate asserted totals, Gate D held exactly 2/1/1"

patterns-established:
  - "Oracle presence is the gate: git cat-file -e <sha>^{commit} fails the test loudly; never skip"
  - "Snapshot shape contract: argv [binary, --format, json, lint] with cwd=corpus root, exit 0/1 accepted, exit 2 = corpus does not link = STOP, strict JSON-object stdout, stderr msg=\"lint completed\" issues=N cross-check"

requirements-completed: [LINT-PARITY-01, LINT-PERF-01]  # verbatim from plan frontmatter; the shared-ID gate keeps the REQUIREMENTS checkboxes pending until 07-03's summary exists — this plan carries PARITY-01's automated core (teeth Gates A-D) and PERF-01's protocol machinery (rehearsed, no verdict)

coverage:
  - id: D1
    description: "TestAcceptanceGates green under -race: Gate A both sides x2, Gate B zero COMPILE (golden purity rationale cited), Gate C filtered 4-tuple multiset identity both directions with non-empty shared set, Gate D exact 2/1/1, stderr issues=N cross-check on all four runs; raw JSON-line delta reported via t.Logf only (only-main 0, only-head 0)"
    requirement: LINT-PARITY-01
    verification:
      - kind: integration
        ref: "command: go test -race -count=1 ./internal/api/ -run TestAcceptanceGates (green 4/4 runs incl. post-commit re-verification)"
        status: pass
      - kind: other
        ref: "command: bin/golangci-lint run ./internal/api/ -> 0 issues"
        status: pass
    human_judgment: false
  - id: D2
    description: "Baseline oracle reproducible and provenance-tracked: built from a detached worktree at the ratified 1e9a596 with commit-presence gate, plain no-race build, worktree registration removed (t.Cleanup + failure-path self-removal); git worktree list holds at 1 after every run"
    requirement: LINT-PARITY-01
    verification:
      - kind: other
        ref: "command: git worktree list | wc -l == 1 after test + selftest runs; git cat-file -e 1e9a596^{commit} gate exercised every build"
        status: pass
    human_judgment: false
  - id: D3
    description: "/tmp/p7 instruments self-tested on the teeth corpus: gate states match TestAcceptanceGates (independent implementations agree), both negative controls fail as designed (diverged pair -> nonzero; COMPILE row -> Gate B trip) and the expect-mode confirms the residual exactly 1/1; perf rehearsal produces the full record shape labeled REHEARSAL"
    requirement: LINT-PARITY-01
    verification:
      - kind: integration
        ref: "command: /tmp/p7/run-parity.sh selftest (green 4/4 runs incl. the formal verify-chain run; gates A-D + negative controls + perf rehearsal)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Parity mode encodes the exact unblock-day sequence (GNU-sed patch, GNU-time lab.py compat patch, setup.sh FIRST, phase binaries SECOND, x2x2 lab.py lint per corpus, lint-diff Gate A, compare.py Gates B/C/D with Gate D reported, perf gated on googleapis-nocompute / recorded on api-common) and refuses to run without GOOGLEAPIS/CORPUS_ARGS"
    requirement: LINT-PERF-01
    verification:
      - kind: other
        ref: "command: env -u GOOGLEAPIS -u CORPUS_ARGS /tmp/p7/run-parity.sh parity -> fast-fails at the guard; patch anchors validated against a lab.py COPY; patched-form lint snapshot green (10 issues, RSS parsed); real lab checkout untouched (git status empty)"
        status: pass
    human_judgment: true
    rationale: "The parity leg cannot fire without the corpora (07-RESEARCH spike: all four unblock routes failed 2026-11-21); the sequence is validated structurally here, and its actual firing is 07-03's end-of-phase human check with the documented unblock path"
  - id: D5
    description: "VALIDATION.md gates row points at internal/api (the pre-seeded internal/core path was a placeholder) and the real automated command; file-exists flipped to yes; every other row untouched"
    requirement: LINT-PARITY-01
    verification:
      - kind: other
        ref: "command: grep -c 'internal/api/ -run' 07-VALIDATION.md >= 1 (verify-gate element)"
        status: pass
    human_judgment: false

# Metrics
duration: 19min
completed: 2026-11-21
status: complete
---

# Phase 07 Plan 02: Acceptance gates tracer — baseline oracle + Gates A-D + corpora instruments Summary

**The ratified 1e9a596 baseline oracle plus the four-gate compare are permanent suite machinery (TestAcceptanceGates green under -race on the teeth corpus, x2 per binary), and the corpora/perf legs are ready-to-fire self-tested /tmp/p7 instruments with two passing negative controls**

## Performance

- **Duration:** 19 min
- **Started:** 2026-11-21T03:34:07Z
- **Completed:** 2026-11-21T03:53:35Z
- **Tasks:** 2
- **Files modified:** 2 committed (+3 throwaway instruments under /tmp/p7)

## Accomplishments

- TestAcceptanceGates (internal/api/acceptance_gates_test.go) is green under `-race` 4/4 runs: Gate A main-vs-main determinism on BOTH sides (sorted raw JSON-line multisets, x2), Gate B zero `RuleName==COMPILE` in all four runs (the golden_full_registry_test.go purity rationale cited verbatim in the failure message), Gate C filtered 4-tuple (RuleName, Path, Position.Line, Position.Column) multiset identity in BOTH directions with a non-empty shared set (10 shared findings), Gate D exact PACKAGE_NO_IMPORT_CYCLE=2 / SYNTAX_SPECIFIED=1 / FIELD_NOT_REQUIRED=1 on both run pairs
- The baseline oracle is ratified, reproducible, and provenance-tracked: `buildBaselineBinary` (sibling sync.OnceValues) verifies `git cat-file -e 1e9a596^{commit}` loudly (the oracle's presence IS the gate — never skipped), builds from a detached worktree at 1e9a596 with a plain no-race build, and guarantees worktree teardown via the test's t.Cleanup plus a failure-path self-removal; `git worktree list` holds at 1 after every run
- The raw JSON-line diagnostic after the NEW3 filter is EMPTY (only-main 0, only-head 0) — byte-identical shared findings, exactly matching the RESEARCH spike; reported via t.Logf only, never a gate
- /tmp/p7 instruments self-tested end-to-end 4/4 runs: compare.py (Gate B inside load, Gate C both directions + non-empty shared, Gate D reported, raw-line diagnostic reported, --expect-only-main/--expect-only-new modes, --no-filter Gate A mode), perf.py (interleaved m,n,m,n warmup+measured, wall/RSS/issues per run, stderr issues=N cross-check, equal-issue-counts determinism witness, median table + full JSON record + provenance + machine context), run-parity.sh (selftest green; parity mode encodes the exact unblock-day sequence and fast-fails without GOOGLEAPIS/CORPUS_ARGS)
- Both negative controls fail as designed: a synthetic pair differing in one 4-tuple makes compare.py exit nonzero, and a synthetic COMPILE row trips Gate B specifically — the compare is not vacuous
- Independent implementations agree: the Go test and the script stack read the same corpus through different code paths and report identical gate states (main 10 findings / HEAD 14, delta exactly 2/1/1)
- VALIDATION.md gates row filled: test home internal/api, command `go test -race -count=1 ./internal/api/ -run TestAcceptanceGates`, File Exists -> yes, status pending; every other row untouched

## Task Commits

Each task committed atomically:

1. **Task 1 (tracer): baseline oracle from ratified 1e9a596 + TestAcceptanceGates — Gates A-D on teeth corpus, x2** - `0669aab` (test; includes the VALIDATION.md gates row)
2. **Task 2: corpora/perf instruments at /tmp/p7** - no repo commit BY DESIGN: the instruments are throwaway /tmp artifacts (06-03 precedent — the committed artifact is the evidence record, which 07-03 owns); the repo working tree is untouched by Task 2 (verified: git status shows only pre-existing unrelated entries)

**Plan metadata:** see the docs commit (SUMMARY + STATE + ROADMAP).

_Tracer feedback gate: auto mode re-ran the full verify chain end-to-end after Task 1 (composite rc=0, 10.9s) before expanding to Task 2 — `⚡ Tracer verified end-to-end — expanding`._

## Files Created/Modified

- `internal/api/acceptance_gates_test.go` - TestAcceptanceGates + buildBaselineBinary + teeth-corpus assembly + lab.py-shape lint runner + the four gates (permanent suite machinery, 425 lines)
- `.planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-VALIDATION.md` - gates row filled (Task ID, plan 02, test home internal/api, real command)
- `/tmp/p7/compare.py` - throwaway Gate B/C/D compare with --no-filter and --expect modes (rebuildable from the plan spec; NOT committed)
- `/tmp/p7/perf.py` - throwaway interleaved perf protocol with portable time wrapper (NOT committed)
- `/tmp/p7/run-parity.sh` - throwaway selftest/parity orchestrator (NOT committed)

## Decisions Made

- A2 RATIFIED here as planned: baseline = 1e9a596 — verified zero Go-source delta to the first milestone commit 8f969d8 (two docs-only commits a0a0ed9/26924b1 intervene); recorded in the test's provenance comments and handed to 07-03's evidence
- buildBaselineBinary returns a result struct because sync.OnceValues is (T, error)-shaped; the worktree dir is surfaced to the test's t.Cleanup AND self-removed on the build-failure path (a cached OnceValues error would otherwise leak the registration forever)
- Gate A is evaluated twice: raw (snapshots written sorted, so `cmp` byte equality IS the sorted multiset equality) and via compare.py --no-filter (the pipeline path). The one-sided NEW3 filter is correct for main-vs-HEAD compares and WRONG for same-binary compares — head-vs-head needs --no-filter or it phantom-diffs the shared NEW3 rows
- The time wrapper probes BSD `-l` first (the lab's own mechanism) and falls back to GNU `-f "%e %M"` reading the same wall+RSS quantities; parity mode patches lab.py's lint() compatibly (idempotent, anchors asserted, pre-patch copy kept) so the stand runs unchanged on GNU-time machines — semantics never modified
- Teeth-corpus counts recorded from measurement (baseline 10, HEAD 14) rather than the RESEARCH narrative's 18/20; no gate asserted totals, so nothing needed re-deriving

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `/usr/bin/time -l` does not exist on this machine (GNU Time)**
- **Found during:** Task 2 (first selftest run — snapshot failed with "invalid option -- 'l'")
- **Issue:** RESEARCH's Environment Availability row "[VERIFIED] /usr/bin/time -l present" is false here: `/usr/bin/time` is GNU Time (2018), which rejects `-l` and would also crash lab.py's lint() RSS parse on any corpora firing
- **Fix:** portable time wrapper in perf.py + snapshot (probe BSD `-l`, fall back to GNU `-f "%e %M"`, RSS parsed per mode — same wall+RSS quantities); parity() gained `patch_lab_py_gnu_time` — an idempotent compatibility patch of lab.py's lint() invocation + RSS parse (anchors asserted with a hard STOP if the stand changed upstream; pre-patch copy kept for evidence; count=1 so timed() is untouched)
- **Files modified:** /tmp/p7/perf.py, /tmp/p7/run-parity.sh (throwaway; no repo files)
- **Verification:** selftest green end-to-end; patch anchors validated against a lab.py COPY (real checkout untouched, git status empty); patched-form lint snapshot green (10 issues, RSS parsed)
- **Committed in:** no repo commit (throwaway instruments; documented here per the 06-03 evidence-artifact precedent)

**2. [Rule 3 - Blocking] compare.py needed a --no-filter mode and expectation-mode exit semantics**
- **Found during:** Task 2 (selftest Gate A — head-vs-head phantom diff)
- **Issue:** (a) The plan's Gate A evaluation "with compare.py" cannot work with the one-sided NEW3 filter on same-binary compares: both sides carry the NEW3 rows and only the NEW side is filtered, so the shared new-rule findings report as residual. (b) The --expect-only-main/--expect-only-new modes (mandated "for negative controls") still applied the plain Gate C failure rule, so a control asserting the KNOWN nonzero residual could never exit zero
- **Fix:** added --no-filter (disables NEW3 filtering on BOTH sides; used for Gate A r1-vs-r2 compares only) and made the expectation modes replace the plain Gate C exit rule with the expectation check; Gate D shell check aligned to compare.py's actual `", "` join format
- **Files modified:** /tmp/p7/compare.py, /tmp/p7/run-parity.sh (throwaway; no repo files)
- **Verification:** selftest green 4/4: Gate A raw + pipeline both sides; control (a) plain -> nonzero AND expect-mode -> zero with residual exactly 1/1; control (b) Gate B trip confirmed by message
- **Committed in:** no repo commit (throwaway instruments)

---

**Total deviations:** 2 auto-fixed (2 blocking, both in throwaway instruments — zero production-code surface)
**Impact on plan:** Both fixes were required for the plan's own acceptance criteria (selftest green; parity mode ready-to-fire). No repo file beyond the plan's file list was touched; lab.py's semantics and the tracked fixtures are untouched.

## Issues Encountered

- The RESEARCH spike's quoted teeth-corpus totals ("Main: rc=1, 18 issues ... Head: rc=1, 20 issues") are not reproducible as stated — its own only-main-0 evidence implies main=16 on a 20-finding head; the actual measured counts on the assembled corpus are main=10, HEAD=14 (delta exactly the 4 NEW3 findings, matching parity-evidence-06-03.json's per-rule counts). No gate asserted totals, so the discrepancy cost nothing; recorded here so 07-03's evidence states counts from measurement
- Selftest needed 4 iterations to green: GNU time form (deviation 1), Gate A --no-filter (deviation 2), a Gate D expectation format string mismatch (my own check, fixed in-place), and the expect-mode exit semantics (deviation 2) — all diagnosed from captured outputs under /tmp/p7/out, none silent

## User Setup Required

None - no external service configuration required. (The corpora unblock path is 06-USER-SETUP.md and stays with 07-03's end-of-phase human check.)

## Next Phase Readiness

- 07-03 can reuse buildBaselineBinary (same package) and TestAcceptanceGates as the automated-core evidence datapoint; the /tmp/p7 instruments and their outputs under /tmp/p7/out (gate logs, negative-control transcripts, perf rehearsal records, provenance files) are the raw material for parity-evidence-07.json / perf-evidence-07.md
- The parity mode fires unchanged on unblock day: GNU-sed patch + GNU-time lab.py compat patch are idempotent and pre-validated (anchors asserted with hard STOP on drift), setup.sh-first/binary-placement-second ordering is enforced, Gate A/B/C/D and the perf gate are wired exactly per the ratified design
- LINT-PARITY-01 / LINT-PERF-01 REQUIREMENTS checkboxes unblock via the shared-ID gate when 07-03's summary exists (ready-ids returned blocked:[] ready for both this plan — correct, 07-03 declares them too)
- Rule counts are stated 41 (baseline registry) / 44 (HEAD registry) in every artifact; the stale "43" figure appears only in flagging mentions

---
*Phase: 07-priemka-paritet-korpusov-i-proizvoditelnost*
*Completed: 2026-11-21*

## Self-Check: PASSED

- Files created/modified verified on disk: `internal/api/acceptance_gates_test.go`, `07-VALIDATION.md`, `07-02-SUMMARY.md`, `/tmp/p7/{compare.py,perf.py,run-parity.sh}` — all FOUND
- Commits verified in HEAD ancestry: `0669aab` (test(07-02) Task 1), plus the docs metadata commit carrying this SUMMARY/STATE/ROADMAP (verified in ancestry as c5fd927 before the Self-Check append; the append was folded in via a single amend, so the current docs hash is the amended HEAD — described by scope rather than re-pinning a self-referential hash)
- Ledger measurement (`gsd-plan-head-before-07-02` = bc3ac38): production commits = 1 (`0669aab`); the post-docs `git rev-list --count bc3ac38..HEAD` reads 2 (1 production + 1 metadata — `actuals.commits: 1` measures the production state at SUMMARY write per #3968; Task 2 carries no repo commit by design)
- Plan-level verification re-run green: `go test -race -count=1 ./internal/api/ -run TestAcceptanceGates` (4/4 runs), `git worktree list | wc -l == 1`, `bin/golangci-lint run ./internal/api/` 0 issues, `/tmp/p7/run-parity.sh selftest` green 4/4 with both negative controls failing as designed, VALIDATION row grep pass, no stale-43 assertions in any artifact
