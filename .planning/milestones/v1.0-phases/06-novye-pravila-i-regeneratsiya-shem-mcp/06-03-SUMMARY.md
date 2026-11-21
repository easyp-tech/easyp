---
phase: 06-novye-pravila-i-regeneratsiya-shem-mcp
plan: 03
subsystem: lint-rules
tags: [parity, lab-cli, teeth-corpus, filtered-projection, evidence, corpora-deferral]

# Dependency graph
requires:
  - phase: 06-novye-pravila-i-regeneratsiya-shem-mcp (06-01)
    provides: tracked teeth fixtures at repo-root testdata/ (no_syntax/, required_field/, import_cycle/ 3-file/2-package cycle)
  - phase: 06-novye-pravila-i-regeneratsiya-shem-mcp (06-02)
    provides: 44-rule registry with the three new rules live end-to-end; repo binary built from the 06-02 final state; engine-captured golden literals pending this cross-check
  - phase: compiler-lab etalon @ 08cef10
    provides: the lab CLI binary as the parity oracle (bench/lab.py invocation shape)
provides:
  - Committed auditable parity evidence (parity-evidence-06-03.json): binary provenance (lab checkout SHA 08cef10..., repo commit d0525b4), corpus file list, per-rule finding counts on both binaries, multiset-equality verdict
  - Proven teeth-corpus parity: filtered 3-name projection (RuleName, Path, Position.Line, Position.Column) identical between lab and repo binaries; per-rule counts PACKAGE_NO_IMPORT_CYCLE 2/2, SYNTAX_SPECIFIED 1/1, FIELD_NOT_REQUIRED 1/1 — cross-validates 06-02's engine-captured golden literals
  - Corpora-leg resolution record: deferred with the access-gate outcome (clone attempts, options presented); the leg stays the manual-only row in 06-VALIDATION.md and rides Phase 7's lab.py stand spike
  - 06-USER-SETUP.md documenting exactly what the user can provide to unblock the corpora leg
affects: [phase-7 main-vs-new corpus gate (filter the three names), phase-7 lab.py stand spike (corpora leg), LINT-NEWRULES-03 traceability]

# Actuals (#2632) — chars/4 over the realized diff (evidence JSON 2323 chars + summary + user-setup ≈ 9800 chars).
actuals:
  tokens: 2450
  tasks: 2
  commits: 2
  plan_head_before: d0525b4f9e6882ef8670ce9ea49c34448ecc9cab
  plan_head_after: de8259568c98dfbd97b986b62f47d02ffea30d19

# Tech tracking
tech-stack:
  added: []  # zero new dependencies; no repo source changes (binaries built to /tmp, never committed)
  patterns:
    - "Filtered-projection parity: compare (RuleName, Path, Position.Line, Position.Column) multisets filtered to the three new rule names — never a raw full-output diff (deliberate main-vs-lab divergences in ported rules stay untouched)"
    - "Teeth corpus from tracked fixtures preserving subdirectory layout: corpus-root-relative imports resolve; MINIMAL auto-seed (policy.go:109) enables all three rules on both binaries with the committed STANDARD config"
    - "Committed evidence over /tmp artifacts: binary SHAs + corpus list + per-rule counts make the parity claim auditable after /tmp wipe"

key-files:
  created:
    - .planning/phases/06-novye-pravila-i-regeneratsiya-shem-mcp/parity-evidence-06-03.json
    - .planning/phases/06-novye-pravila-i-regeneratsiya-shem-mcp/06-USER-SETUP.md
  modified: []

key-decisions:
  - "Corpora leg DEFERRED per the plan's pre-resolved outcome path: EasyP-Parser-Compiler-Experiment inaccessible (gh clone under authenticated repo-scope token: repository not resolvable; SSH remote refused); teeth parity stands as the automated LINT-NEWRULES-03 record; the leg remains the 06-VALIDATION.md manual-only row and rides Phase 7's lab.py stand spike (same private-repo access, STATE.md blocker)"
  - "Throwaway compare script at /tmp/p6-parity/compare.mjs (not committed) — its OUTPUT is the committed evidence, consistent with prior phases' probe style; lab.py, the lab checkout and the tracked fixtures untouched"
  - "Teeth corpus assembled from the 06-01 tracked fixtures preserving subdirectory layout so import locations resolve corpus-root-relative; config unsurgical (MINIMAL auto-seed at policy.go:109 enables all three rules on both binaries)"
  - "Non-JSON or key-incomplete stdout lines are a HARD failure (A4 output-contract drift surfaced, never skipped); zero findings for a rule on acceptance corpora would be the EXPECTED result there, never 'fixed'"

patterns-established:
  - "Anti-vacuity floors as first-class evidence: per-rule >=1 findings on BOTH binaries asserted (cycle rule >=2 — the two cycle edges); a 0-count rule is a failure, not a pass (Pitfall 5 guard)"
  - "Projection to the four-tuple from each line's own keys: missing keys on either side hard-fail instead of being skipped (A4 mitigation)"

requirements-completed: [LINT-NEWRULES-03]  # copied verbatim from plan frontmatter; shared with 06-01/06-02 — both siblings carry summaries, so the shared-ID gate releases it with this plan

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Teeth-corpus parity proven non-vacuously: lab binary @ 08cef10909c0722f7042ba17b5c11951f9c6b6f6 vs repo binary @ d0525b4 (06-02 final state) over the corpus assembled from the tracked 06-01 fixtures — filtered 3-name projection multisets equal, per-rule floors met (PACKAGE_NO_IMPORT_CYCLE 2/2, SYNTAX_SPECIFIED 1/1, FIELD_NOT_REQUIRED 1/1), evidence committed with full provenance"
    requirement: LINT-NEWRULES-03
    verification:
      - kind: other
        ref: "command: git -C /tmp/gsd-research/compiler-lab rev-parse HEAD | grep -qx 08cef10909c0722f7042ba17b5c11951f9c6b6f6 && test -x /tmp/p6-parity/bin/easyp-lab && test -x /tmp/p6-parity/bin/easyp-new && node /tmp/p6-parity/compare.mjs (PARITY_OK, exit 0)"
        status: pass
      - kind: other
        ref: "command: node evidence check — multiset_equal true + per-rule >=1 both sides + cycle >=2 -> PARITY_EVIDENCE_OK (exit 0)"
        status: pass
      - kind: unit
        ref: "internal/rules/{package_no_import_cycle,syntax_specified,field_not_required}_test.go (06-01 suites green; message semantics pinned by unit tests, positions cross-validated by this corpus run)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Corpora leg (api-common, googleapis-nocompute) consciously resolved: clone attempted per the Phase 1-5 precedent (gh + SSH), access gate fired, DEFERRAL recorded with reason and options in the evidence (corpora_leg.status=deferred) and in this SUMMARY; no lab.py/lab-checkout/fixture modification"
    requirement: LINT-NEWRULES-03
    verification: []
    human_judgment: true
    rationale: "The gate outcome is a human decision by design — only the user can PROVIDE the private corpora source (clone it themselves or point GOOGLEAPIS/CORPUS_ARGS at local checkouts); the deferral's adequacy and any later execution of the corpora leg are reviewed by the end-of-phase verifier (see the human-check block and 06-USER-SETUP.md)"

# Metrics
duration: 3 min
completed: 2026-11-21
status: complete
---

# Phase 6 Plan 3: Parity vs lab CLI Summary

**Teeth-corpus parity proven non-vacuously against the lab CLI @ 08cef10 — the filtered 3-name projection (RuleName, Path, Line, Column) is identical between binaries (2/2, 1/1, 1/1) with committed auditable evidence; the acceptance-corpora leg deferred with the recorded access-gate outcome.**

## Performance

- **Duration:** 3 min
- **Started:** 2026-11-21T00:19:10Z
- **Completed:** 2026-11-21T00:22:05Z
- **Tasks:** 2
- **Files modified:** 3 (evidence JSON + SUMMARY + USER-SETUP; no repo source changes — binaries built to /tmp, never committed)

## Accomplishments
- Lab provenance re-verified at run time (`git -C /tmp/gsd-research/compiler-lab rev-parse HEAD` == `08cef10909c0722f7042ba17b5c11951f9c6b6f6`); both binaries built OUTSIDE the repos' trees (`/tmp/p6-parity/bin/easyp-lab` from the lab checkout, `/tmp/p6-parity/bin/easyp-new` from this repo @ d0525b4)
- Teeth corpus assembled at `/tmp/p6-parity/corpus/` from the tracked 06-01 fixtures preserving subdirectory layout (import_cycle/a.proto, import_cycle/b/b.proto, import_cycle/a2.proto, no_syntax/no_syntax.proto, required_field/required_field.proto) with the committed-config-shaped `easyp.yaml` (`version: v1 / linters: default: STANDARD`) — no config surgery needed on either binary (MINIMAL auto-seed, policy.go:109)
- Both binaries ran linker-clean (exit 1, findings present — never exit 2): filtered 3-name projection multisets EQUAL; per-rule counts lab/new: PACKAGE_NO_IMPORT_CYCLE 2/2 (the two cycle edges a.v1->b.v1->a.v1 and b.v1->a.v1->b.v1 at import keyword 5:1), SYNTAX_SPECIFIED 1/1 (no_syntax at 1:1), FIELD_NOT_REQUIRED 1/1 (required `id` at 6:18, SourceName bare `id`) — anti-vacuity floors met on both sides
- The run cross-validates 06-02's engine-captured new-rule golden literals (same Message/Position/SourceName semantics on the same fixture shapes)
- Corpora leg: clone attempted two ways before any deferral (gh repo clone under the authenticated token — repository not resolvable; SSH remote — refused); DEFERRAL recorded in the evidence (`corpora_leg` field) with the reason, the attempts, and the PROVIDE/DEFER options; 06-USER-SETUP.md generated
- The throwaway compare script (`/tmp/p6-parity/compare.mjs`) enforces the shared output contract: any non-JSON or key-incomplete stdout line is a HARD failure (A4 drift surfaced, never skipped)

## Task Commits

Each task was committed atomically:

1. **Task 1: Tracer — teeth-corpus parity with committed evidence** - `7db2ee0` (docs)
2. **Task 2: Corpora-leg outcome record (deferred with access-gate outcome)** - `de82595` (docs)

_Plan metadata commit: see final docs commit (SUMMARY/STATE/ROADMAP/REQUIREMENTS/USER-SETUP)._

**Tracer feedback gate:** auto mode active — the full `<automated>` verify chain re-ran end-to-end green (`PARITY_EVIDENCE_OK`, CHAIN_EXIT=0) before Task 2 expansion.

## Files Created/Modified
- `.planning/phases/06-novye-pravila-i-regeneratsiya-shem-mcp/parity-evidence-06-03.json` - the committed parity evidence: lab {checkout_sha, binary_path, exit_code 1, 9 findings}, new {commit_sha d0525b4, binary_path, exit_code 1, 14 findings}, corpus file list, config, projection, per_rule_counts, multiset_equal: true, teeth_anti_vacuity, corpora_leg {status: deferred, reason, clone_attempts, options_presented, choice, carries_requirement}, timestamp
- `.planning/phases/06-novye-pravila-i-regeneratsiya-shem-mcp/06-USER-SETUP.md` - how to provide the private corpora source (GOOGLEAPIS/CORPUS_ARGS env vars + clone/point-at-checkout task) if the corpora leg is to be executed now or at Phase 7
- `/tmp/p6-parity/{bin,corpus,compare.mjs,lab.out,new.out}` - throwaway run artifacts (NOT committed; their output is the evidence; wiped with /tmp without loss)

## Decisions Made
- **Deferral over blocking:** the access gate fired exactly as the plan's action item 2 prescribed; with no human available mid-flight in auto mode, the pre-resolved DEFER record (teeth parity carries the automated requirement) was taken and the gate outcome fully documented for the end-of-phase human check rather than stalling the phase. The corpora leg is consciously unresolved-until-provided, not silently dropped.
- **Throwaway script, committed output:** the compare mechanism lives at /tmp (not committed) per the plan's resolution of RESEARCH Open Question 2; the evidence JSON is the auditable artifact (binary SHAs, corpus list, per-rule counts).
- **Hard-fail output contract:** a non-JSON line on either side would abort the compare (A4: missing keys surfaced, not skipped) — the run had zero parse failures on both sides.

## Deviations from Plan

None - plan executed exactly as written. (The corpora-leg deferral is a planned outcome defined by the plan's own action item 2, recorded via the plan's own artifacts; not a deviation. No lab.py, lab-checkout or tracked-fixture modification occurred; no repo source changed.)

## TDD Gate Compliance

`workflow.tdd_mode` is ON, but this plan's frontmatter is `type: execute` (not `type: tdd`) and both tasks are compare/parity + evidence-recording tasks with no `<behavior>` blocks — the canonical predicate exempts them; no RED/GREEN/REFACTOR sequence applies. The parity claim is carried by the machine-checked evidence chain (Task 1 `<automated>`, re-run green at the tracer gate) plus the standing 06-01/06-02 unit suites.

## Authentication Gates

- **Task 2 corpora leg:** the private-repo access gate fired as a dynamic gate (the plan's anticipated `checkpoint:human-verify` condition): `gh repo clone easyp-tech/EasyP-Parser-Compiler-Experiment` failed under an authenticated token (account onokonem, scopes repo/read:org/workflow/gist) with "Could not resolve to a Repository", and the SSH remote (`git@github.com:easyp-tech/EasyP-Parser-Compiler-Experiment.git`) was refused ("Could not read from remote repository"). Documented as normal flow (not a deviation): the deferral record + 06-USER-SETUP.md + the `<human-check>` block below carry the outcome; PROVIDE options remain open for the end-of-phase verifier or Phase 7.

## Issues Encountered
- None beyond the access gate above. Both binaries' total finding counts differ as expected (lab 9 vs new 14 — deliberate main-vs-lab divergences in the ~41 ported rules, e.g. PACKAGE_DIRECTORY_MATCH text; the comparison is filtered+projected by design, Pitfall 6 respected).

## User Setup Required

**The corpora leg needs human-provided access.** See [06-USER-SETUP.md](./06-USER-SETUP.md) for:
- Environment variables to add (`GOOGLEAPIS`, `CORPUS_ARGS`)
- The clone/point-at-checkout task for the private EasyP-Parser-Compiler-Experiment repo
- Verification commands and expected corpus sizes (92 / 7414 files)

The automated LINT-NEWRULES-03 core (teeth parity) requires no setup.

## Next Phase Readiness
- LINT-NEWRULES-03's automated core is proven: identical projected findings on the teeth corpus, non-vacuously, with committed provenance. All four phase requirements (01-04) now have their plans complete — phase verification can close REQUIREMENTS.md checkboxes.
- **Phase 7 inherits:** (a) the corpora leg deferred to the lab.py stand spike — same private-repo access gap already flagged in STATE.md blockers; 06-USER-SETUP.md documents the unblock path; (b) the parity mechanism shape (filter+project, never raw-diff) reusable for the main-vs-new corpus gate; (c) reminder: filter PACKAGE_NO_IMPORT_CYCLE/SYNTAX_SPECIFIED/FIELD_NOT_REQUIRED out of the Phase-7 main-vs-new compare (main lacks them; MINIMAL auto-seed makes them run on every v1 config); (d) binaries in /tmp are disposable — rebuild from the recorded SHAs.
- No blockers introduced by this plan.

## End-of-Phase Human Check (harvested by the verifier)

The corpora leg is the phase's one manual-only verification (06-VALIDATION.md Manual-Only row): review the corpora-leg outcome in this SUMMARY and in `parity-evidence-06-03.json` (`corpora_leg`). As DEFERRED: confirm the deferral reason (private-repo access: both clone paths failed) and that the teeth-corpus evidence remains the LINT-NEWRULES-03 automated record. If you can PROVIDE access (clone the repo yourself or point GOOGLEAPIS/CORPUS_ARGS at local checkouts — see 06-USER-SETUP.md), the corpora leg can be executed at Phase 7's stand spike with the same compare mechanism; corpora results with ZERO findings for a rule are the EXPECTED outcome there (proto3-heavy, syntaxed, acyclic) and must not be "fixed".

## Self-Check: PASSED

- All 3 created files exist on disk: parity-evidence-06-03.json, 06-03-SUMMARY.md, 06-USER-SETUP.md (3/3 FOUND)
- All 3 commits are ancestors of HEAD: 7db2ee0 (task 1), de82595 (task 2), plus the plan metadata docs commit carrying this file (3/3 FOUND)
- Task 1 verify chain re-ran green at self-check: PARITY_OK + PARITY_EVIDENCE_OK (multiset equal, per-rule floors, cycle >= 2)
- Task 2 verify chain green on the committed SUMMARY: EVIDENCE_STATE_OK corpora_leg=deferred, 'corpora leg' recorded in the SUMMARY
- Lab checkout clean at `git -C /tmp/gsd-research/compiler-lab status` (0 modified files — the T-06-06/lab.py prohibition held); repo diff d0525b4..HEAD touches .planning only (no repo source changes, binaries never committed)
- Measured commits from the ledger: d0525b4..de82595 = 2 task commits; +1 metadata commit (a12e864); plan_head_after = de8259568c98dfbd97b986b62f47d02ffea30d19
- Broken-windows ledger: no entries added (no stubs, no skipped tests, no unrun verifies — both `<automated>` chains ran green; a probe entry written while testing ledger availability was reverted)

---
*Phase: 06-novye-pravila-i-regeneratsiya-shem-mcp*
*Completed: 2026-11-21*
