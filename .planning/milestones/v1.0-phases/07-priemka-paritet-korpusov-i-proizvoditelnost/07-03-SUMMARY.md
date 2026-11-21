---
phase: 07-priemka-paritet-korpusov-i-proizvoditelnost
plan: 03
subsystem: testing
tags: [sc3-battery, gotestsum, lint-parity, perf-protocol, evidence, corpora-disposal, anti-ratification, hadolint-manual]

# Dependency graph
requires:
  - phase: 07-01 (1f82b61)
    provides: race-free wazero default factory — the full -race ./... suite (SC3 leg a) runs green for the first time since Phase 1
  - phase: 07-02 (0669aab)
    provides: TestAcceptanceGates permanent machinery, buildBaselineBinary (RATIFIED 1e9a596 oracle), /tmp/p7 instruments + out/ gate states and REHEARSAL data
provides:
  - SC3 milestone battery green on the final state through the documented direct equivalents (full -race ./... via bin/gotestsum v1.13.0, golangci zero issues, schema drift empty; hadolint honest manual item)
  - parity-evidence-07.json — committed auditable parity record: baseline RATIFIED 1e9a596 (registry 41), head 9a37392 (registry 44), automated-core Gates A-D states on the teeth corpus, corpora_leg pending_user_unblock with fresh clone_attempts + anti-ratification rule, salvage cross-reference
  - perf-evidence-07.md — committed measurement protocol (CI-equivalent flags, warmup 1 + median-of-5 interleaved, no-race provenance, machine context, {bin, wall_s, rss_mb, issues, rc} fields, REHEARSAL data, pending gate definition)
  - 07-USER-SETUP.md — the corpora unblock path with the Pitfall 7 revision-compatibility nuance and the one-command verification path
  - 07-VALIDATION.md final rows (battery green; corpora/perf MANUAL-ONLY pending)
affects: [milestone close (verify-work/UAT reviews the evidence + corpora disposal), any future corpora unblock (scripts fire unchanged)]

# Actuals (#2632)
actuals:
  tokens: 10886   # chars/4 over the realized diff (43543 chars: three evidence docs + VALIDATION rows + SUMMARY; no production code)
  tasks: 3
  commits: 2      # MEASURED: git rev-list --count gsd-plan-head-before-07-03..HEAD at SUMMARY write (Task 1 battery = no repo commit by design; docs commits carry the evidence)

plan_head_before: 9a37392b322fd42f378b2210a0b7ccbd5db78e00
plan_head_after: 0cc549c

# Tech tracking
tech-stack:
  added: [gotestsum v1.13.0 (Taskfile-pinned; installed to bin/, gitignored, never committed)]
  patterns:
    - "SC3 direct-equivalents battery: conditional runner (bin/gotestsum when present, else documented go test fallback), one-tmp-dir schema-drift gate, pathspec-limited hygiene checks"
    - "Anti-ratification evidence contract: corpora_leg.status stays pending_user_unblock until per-corpus results exist; fresh access attempts re-recorded at execution time (never the stale research snapshot)"
    - "MANUAL-ONLY verification rows carry the exact one-command path next to the automated core's demonstrated evidence"

key-files:
  created:
    - .planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/parity-evidence-07.json
    - .planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/perf-evidence-07.md
    - .planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-USER-SETUP.md
    - .planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-03-SUMMARY.md
  modified:
    - .planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-VALIDATION.md
  never-committed:
    - coverage.out (gitignored battery artifact)
    - bin/gotestsum (gitignored Taskfile-pinned helper)

key-decisions:
  - "Battery record lives in this SUMMARY (one source of truth); perf-evidence-07.md carries a cross-reference and the shared provenance (SHAs, go version, no-race, machine context)"
  - "head.commit_sha in the evidence = 9a37392 (the battery-tested SHA); later 07-03 commits are .planning docs-only — the Go source state is identical, noted in the evidence"
  - "Corpora legs disposed honest: fresh access attempts (gh exit 1, SSH exit 128, env vars unset) recorded in corpora_leg.clone_attempts at 2026-11-21T04:15:09Z; status stays pending_user_unblock; LINT-PARITY-01's and LINT-PERF-01's letters ride as MANUAL-ONLY rows with the exact one-command path from 07-USER-SETUP.md"
  - "Gate D on corpora is REPORTED, never asserted (A1: zero findings for a rule on proto3-heavy corpora is the EXPECTED result); rule counts stated 41/44 from the registries — the stale '43' figure is never asserted in any artifact"

patterns-established:
  - "Evidence documents carry machine-proven gate states + explicit pending states side by side — the milestone closes on the automated core with honestly pending corpora letters (A5)"
  - "Fresh-attempt-before-disposal: every execution re-probes the blocked resource and records commands/outputs/exit codes/timestamps rather than trusting the research-date snapshot"

requirements-completed: [LINT-PARITY-01, LINT-PERF-01]  # verbatim from plan frontmatter; shared-ID gate with 07-01/07-02 (both summaries exist) — the letters' corpora legs ride as MANUAL-ONLY pending rows, the automated core + disposal is this phase's record

coverage:
  - id: D1
    description: "SC3 milestone battery green on the final state: full -race ./... suite via bin/gotestsum v1.13.0 (3718 tests, 1 pre-existing skip, 0 FAIL, incl. plugin/generation, TestAcceptanceGates, TestLintGoldenFullRegistry, TestLintSuppressionParity/TestLintMatrix/TestReadSourceOnce), golangci zero issues, schema drift empty via one tmp dir, hygiene gates clean"
    requirement: LINT-PARITY-01
    verification:
      - kind: integration
        ref: "command: bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./... (exit 0; verify-chain re-run also exit 0, 3718 tests)"
        status: pass
      - kind: other
        ref: "command: bin/golangci-lint run ./... -> 0 issues; one-tmp-dir schema-gen diff -> empty; targeted -race runs name the standing-constraint cells PASS"
        status: pass
    human_judgment: false
  - id: D2
    description: "parity-evidence-07.json committed and auditable: baseline RATIFIED 1e9a596 (registry 41), head 9a37392 (registry 44), automated-core gate states (Gate A equal x2 both sides, Gate B zero COMPILE, Gate C multiset-equal with non-empty shared set, Gate D 2/1/1 on teeth), salvage block, stale-figure note"
    requirement: LINT-PARITY-01
    verification:
      - kind: other
        ref: "command: python3 JSON gate asserting ratified_sha/registry_counts/gate_d counts/gate_c multiset_equal/stale_figure_note (all pass)"
        status: pass
    human_judgment: false
  - id: D3
    description: "perf-evidence-07.md records the full protocol before any measurement: CI-equivalent flags, warmup 1 + median-of-5 interleaved, no-race provenance (1e9a596 / 9a37392), machine context, corpus identity 92/7414 (pending unblock), REHEARSAL teeth data (no verdict), pending gate median(new) <= median(main) on googleapis-nocompute"
    requirement: LINT-PERF-01
    verification:
      - kind: other
        ref: "command: grep gates (REHEARSAL present; {bin, wall_s, rss_mb, issues, rc} fields verbatim; gate state explicit)"
        status: pass
    human_judgment: false
  - id: D4
    description: "07-USER-SETUP.md carries the corpora unblock path with the Pitfall 7 revision-compatibility nuance (fresh public googleapis clone; stale 2024-02 checkout = 3889 protos vs required ~7416) and the one-command verification path"
    requirement: LINT-PARITY-01
    verification:
      - kind: other
        ref: "command: grep gates ('Pitfall 7' + 'revision-compatibility' present; no stale-figure patterns in any of the three docs)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Corpora legs honestly disposed: fresh access attempts recorded (gh exit 1, SSH exit 128, env unset) in corpora_leg.clone_attempts; status pending_user_unblock; LINT-PARITY-01/LINT-PERF-01 letters ride as MANUAL-ONLY rows with the exact one-command path; automated disposition gate passes (clone_attempts non-empty + MANUAL-ONLY rows in SUMMARY)"
    requirement: LINT-PERF-01
    verification:
      - kind: other
        ref: "command: Task 3 automated disposition gate (executed per_corpus_results non-empty OR clone_attempts non-empty; pending leg carries MANUAL-ONLY rows) — pass"
        status: pass
    human_judgment: true
    rationale: "The corpora legs cannot fire without user action (private repo unreachable — gh/SSH/env all fail, re-verified at execution); the end-of-phase human check reviews the disposal honesty and, on unblock, the same scripts fire unchanged (07-USER-SETUP.md path)"

# Metrics
duration: 16min
completed: 2026-11-21
status: complete
---

# Phase 07 Plan 03: SC3 milestone battery + committed evidence + corpora legs honest disposal Summary

**SC3 green on the final state (full `-race` suite 3718 tests via Taskfile-pinned gotestsum, golangci 0 issues, schema drift empty) with the milestone's parity/perf claims committed as auditable evidence and the corpora legs honestly disposed as pending_user_unblock with a ready-to-fire one-command path**

## Performance

- **Duration:** 16 min
- **Started:** 2026-11-21T04:04:14Z
- **Completed:** 2026-11-21T04:19:38Z
- **Tasks:** 3
- **Files modified:** 5 (4 created evidence/SUMMARY docs, 1 VALIDATION.md row update) — zero production code by design

## Accomplishments

- **SC3 milestone battery green on the final committed state** (the ROADMAP's final gate, first unblocked full `-race` run of the milestone): leg (a) `bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...` exit 0 — 3718 tests, 1 skipped (pre-existing `TestGenerateGoHeaderUsesCompilerVersion` skip: `protoc-gen-go` not in PATH), 0 FAIL, 222 s; includes `internal/adapters/plugin` + `internal/generation` (post-07-01), `TestAcceptanceGates` (07-02, builds the 1e9a596 baseline), `TestLintGoldenFullRegistry`, and the standing-constraint cells — confirmed by name via targeted `-race` runs (`TestAcceptanceGates` ok 10.039s, `TestLintGoldenFullRegistry` ok 1.120s, `TestLintSuppressionParity`/`TestLintMatrix`/`TestReadSourceOnce` PASS)
- Leg (b) `bin/golangci-lint run ./...` → exit 0, **0 issues** (5 s). Leg (c) schema-drift gate with ONE temp dir → exit 0, diff empty (3 s)
- Composite verify chain re-run end-to-end (tracer feedback gate): COMPOSITE_RC=0, HYGIENE_OK — full suite green again (218.8 s), lint clean, drift empty, coverage.out/easyp binary unstaged, `internal/easypc` zero-diff
- **parity-evidence-07.json committed**: baseline block (RATIFIED `1e9a596`, registry 41, worktree build procedure, no-race), head block (`9a37392`, registry 44), automated core (golden corpus green; gates runner green 4/4; Gate A equal x2 on both sides; Gate B zero COMPILE; Gate C multiset-equal after the NEW3 filter with a non-empty shared set — 10 shared findings; Gate D exact 2/1/1 on the teeth corpus), corpora_leg (pending_user_unblock + anti-ratification + fresh attempts), salvage cross-reference (indicative only [A3]), rule_counts 41/44 with the stale-figure note, raw-line diagnostic reported-not-gated
- **perf-evidence-07.md committed**: the protocol recorded before any measurement (CI-equivalent flags, warmup 1 + median-of-5 INTERLEAVED, portable GNU-time wrapper, `{bin, wall_s, rss_mb, issues, rc}` fields verbatim, issue-count determinism witness), provenance (SHAs, build commands, no `-race`, go1.27.3), machine context (i9-9980HK, nproc 16, load), corpus identity 92/7414 (pending unblock), REHEARSAL teeth table (no verdict), the pending gate definition `median(new) <= median(main)` on googleapis-nocompute (api-common recorded-not-gated), explicit corpora-gate state
- **07-USER-SETUP.md committed**: the unblock path mirroring the 06-USER-SETUP shape, with the **Pitfall 7 revision-compatibility nuance** (GOOGLEAPIS must be a fresh public googleapis clone; the stale yadro 2024-02 checkout holds 3889 protos vs the corpus's required ~7416 and setup.sh aborts with `missing import ...`) and the one-command verification path
- **Corpora legs honestly disposed**: fresh access attempts re-run at 07-03 execution (`gh repo view` exit 1 — Could not resolve to a Repository; `git ls-remote` SSH exit 128 — Repository not found; `GOOGLEAPIS`/`CORPUS_ARGS` both unset) and recorded in `corpora_leg.clone_attempts` with timestamps; status stays `pending_user_unblock`; **LINT-PARITY-01's letter (api-common 92 / googleapis-nocompute 7414, x2 runs, main-vs-main base captured first) and LINT-PERF-01's letter ride as MANUAL-ONLY verification rows** carrying the exact one-command path from 07-USER-SETUP.md, alongside the automated core's demonstrated evidence
- **07-VALIDATION.md final rows**: battery row ✅ green (with the exact composite command), corpora/perf row ⬜ pending MANUAL-ONLY (pointing at parity-evidence-07.json + 07-USER-SETUP.md), Manual-Only Verifications table updated to the Phase-7 unblock path + the perf gate definition
- hadolint recorded as an **honest environment-unavailable manual item** (Docker absent — `docker` binary not installed, rc=127; hadolint absent): the root Dockerfile is untouched by the milestone and the last green stands; never silently passed

## Task Commits

1. **Task 1 (tracer): SC3 milestone battery on the final state** — no repo commit BY DESIGN: the battery is a run (gotestsum helper lives gitignored in `bin/`; `coverage.out` gitignored); all results carried into the Task 2 evidence. Tracer feedback gate passed in auto mode: the composite `<verify>` chain re-ran end-to-end green (COMPOSITE_RC=0) before expansion — `⚡ Tracer verified end-to-end — expanding`
2. **Task 2: committed evidence — parity-evidence-07.json, perf-evidence-07.md, 07-USER-SETUP.md** - `0f3efec` (docs)
3. **Task 3: corpora/perf legs — fresh access attempt + honest disposal** - (docs commit with the clone_attempts extension, VALIDATION final rows, and this SUMMARY)

**Plan metadata:** see the docs commit (SUMMARY + STATE/ROADMAP + REQUIREMENTS).

## SC3 Battery Record (single source of truth)

| Leg | Command | Exit | Duration | Result |
|-----|---------|------|----------|--------|
| a. full suite | `bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...` | 0 | 222 s | 3718 tests, 1 skipped (pre-existing PATH-dependent cell), 0 FAIL; runner = gotestsum v1.13.0 per Taskfile pin (no fallback needed) |
| a-verify rerun | same (composite chain) | 0 | 218.8 s | same totals — tracer gate evidence |
| b. go lint | `bin/golangci-lint run ./...` | 0 | 5 s | 0 issues |
| c. schema drift | `go run ./cmd/easyp schema-gen --out-dir "$tmpd" && diff -r schemas "$tmpd"` (one tmp dir) | 0 | 3 s | diff empty |
| hygiene | `git status --porcelain` pathspec checks | 0 | — | coverage.out + easyp binary unstaged (gitignored); internal/easypc zero-diff; bin/ gitignored |
| hadolint (task lint docker leg) | `docker run --rm -i ghcr.io/hadolint/hadolint:... < Dockerfile` | n/a | — | **MANUAL ITEM — environment-unavailable**: Docker not installed on this machine (rc=127), hadolint absent; Dockerfile untouched by the milestone; last green stands |

Battery tested SHA `9a37392` (recorded in the evidence as `head.commit_sha`; later 07-03 commits are `.planning` docs-only).

## Files Created/Modified

- `.planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/parity-evidence-07.json` - the committed parity record (baseline/head/automated_core/corpora_leg/salvage/rule_counts/raw_line_diagnostic/timestamp)
- `.planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/perf-evidence-07.md` - the recorded perf protocol + provenance + machine context + REHEARSAL + pending gate definition
- `.planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-USER-SETUP.md` - the corpora unblock path (env vars, clone item, verification commands, expected results)
- `.planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-VALIDATION.md` - final battery + corpora/perf rows, Manual-Only Verifications updated to the Phase-7 path
- `.planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-03-SUMMARY.md` - this file

## Decisions Made

- Battery record lives HERE (single source of truth); perf-evidence-07.md carries the shared provenance and a cross-reference — per the plan's "either is fine, keep one source of truth"
- `head.commit_sha` in the evidence = `9a37392` (what the battery actually tested), with the note that later 07-03 commits are docs-only — auditable without implying a rebuild
- Fresh access attempts recorded as structured objects (command/output/exit_code/attempted_at) in `clone_attempts` — stronger audit trail than the 06-03 string-array form while staying a superset of it
- Rule counts stated 41/44 from the registries everywhere; the stale '43' figure appears only in the flagging note (`rule_counts.stale_figure_note`), never asserted (grep-gated)
- Gate D disposition language pinned in the evidence: on the acceptance corpora, zero findings for a rule is the EXPECTED result (A1) — never treated as a failure or "fixed"

## Deviations from Plan

**1. [Orchestrator-directed] 07-VALIDATION.md final rows**
- **Found during:** Task 3
- **Issue/Note:** the plan's `files_modified` list did not name 07-VALIDATION.md, but the orchestrator notes explicitly directed "VALIDATION.md: final rows" — the two "(заполняется планами)" placeholder rows and the Manual-Only Verifications table needed the phase's final states
- **Fix:** battery row marked ✅ green with the composite command; corpora/perf row marked ⬜ pending MANUAL-ONLY with pointers to the evidence + 07-USER-SETUP.md; manual table updated to the Phase-7 unblock path and the perf gate definition (frontmatter status stays draft — validate-phase owns it)
- **Files modified:** .planning/phases/07-.../07-VALIDATION.md
- **Verification:** rows carry the executed states; no status/nyquist flips (verifier's job)
- **Committed in:** the Task 3 docs commit

**Total deviations:** 1 documented orchestrator-directed addition (0 auto-fixed bugs — plan executed as written otherwise)
**Impact on plan:** bookkeeping only; the audit trail now reflects the executed state in the phase's own validation contract.

## Issues Encountered

- Load average was elevated (16.05) during the battery window because the background verify chain ran concurrently with other work — noted in perf-evidence-07.md's machine-context section; the perf gate compares same-machine interleaved medians, so background load affects both binaries symmetrically. No action needed.
- The SSH probe's exit code was masked by a pipe on first capture; re-run standalone and recorded with the true rc (128). No evidence impact.

## User Setup Required

**External access requires manual configuration.** See [07-USER-SETUP.md](./07-USER-SETUP.md) for:
- `GOOGLEAPIS` (fresh public googleapis clone — Pitfall 7 revision-compatibility) and `CORPUS_ARGS` environment variables
- Clone EasyP-Parser-Compiler-Experiment (private, SSH) or point the env vars at compatible checkouts
- One-command verification path + expected results (~92 / ~7414 .proto materialized with easyp.yaml STANDARD; both binaries run; filtered compares evaluate per 07-02's parity mode)

After setup the executor fires `/tmp/p7/run-parity.sh parity` + the perf protocol unchanged — zero plan changes needed.

## END-OF-PHASE HUMAN CHECK (for the verifier/UAT)

Review in one batch:
1. `parity-evidence-07.json` gate states + `corpora_leg` honesty — EITHER executed per-corpus results (multiset_equal true, file counts ~92/~7414, perf verdict) OR `pending_user_unblock` with the fresh attempt records (gh exit 1 / SSH exit 128 / env unset, 2026-11-21T04:15:09Z) and the 07-USER-SETUP.md one-command path. **This phase's state: pending_user_unblock.**
2. MANUAL-ONLY rows: LINT-PARITY-01's letter and LINT-PERF-01's letter ride pending with the exact unblock path (above); the automated core (teeth Gates A-D 2/1/1, golden corpus 94 byte-exact, SC3 battery) carries the phase meanwhile.
3. The hadolint manual item disposition (Docker unavailable — never silently passed).
4. If the user unblocks at verification time: complete 07-USER-SETUP.md, then the executor fires `/tmp/p7/run-parity.sh parity` + the perf protocol unchanged — evidence extends (`corpora_leg.status` → `executed` with `per_corpus_results`; perf gate section rewritten with measured medians + verdict), zero plan changes needed.

## Next Phase Readiness

- **This was the LAST execution plan of the milestone** — phase verification gates run next (`/gsd:verify-work 07`): the full suite, golangci, schema drift are green on the final state; the evidence is committed and auditable; the corpora letters are explicitly pending with a ready-to-fire path (A5: milestone closes on the automated core + honestly pending items)
- The `/tmp/p7` instruments + their outputs remain the raw material if corpora unblock (rebuildable per the 07-02 spec if /tmp was wiped)
- Requirements LINT-PARITY-01 / LINT-PERF-01 close via the shared-ID gate (07-01/07-02/07-03 summaries all exist); their corpora-letter halves remain MANUAL-ONLY rows in 07-VALIDATION.md until unblock

---
*Phase: 07-priemka-paritet-korpusov-i-proizvoditelnost*
*Completed: 2026-11-21*

## Self-Check: PASSED

- Files created/modified verified on disk: `parity-evidence-07.json`, `perf-evidence-07.md`, `07-USER-SETUP.md`, `07-03-SUMMARY.md`, `07-VALIDATION.md` (rows) — all FOUND
- Commits verified in HEAD ancestry: `0f3efec` (Task 2 evidence), `0cc549c` (Task 3 disposal + SUMMARY) — all FOUND
- Measured from ledger `gsd-plan-head-before-07-03` (`9a37392`): 2 docs commits, diff 43543 chars (+407/−4 across 5 files); Task 1 carries no repo commit by design (battery run; gotestsum helper + coverage.out gitignored)
- Task gates re-run green: Task 2's JSON/doc grep gates (ratified_sha, 41/44, gate_d 2/1/1, gate_c multiset_equal, no stale-figure patterns) and Task 3's automated disposition gate (clone_attempts non-empty + MANUAL-ONLY rows in SUMMARY)
- Plan-level verification re-run green: composite SC3 chain exit 0 (3718 tests -race, golangci 0 issues, schema drift empty), hygiene clean, `internal/easypc` zero-diff
