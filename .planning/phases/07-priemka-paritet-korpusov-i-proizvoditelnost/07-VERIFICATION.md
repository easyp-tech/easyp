---
phase: 07-priemka-paritet-korpusov-i-proizvoditelnost
verified: 2026-11-21T06:23:03Z
status: passed
score: 10/12 must-haves verified
covered_files: [".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-01-PLAN.md", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-01-SUMMARY.md", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-02-PLAN.md", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-02-SUMMARY.md", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-03-PLAN.md", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-03-SUMMARY.md", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-04-PLAN.md", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-04-SUMMARY.md", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-USER-SETUP.md", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/07-VALIDATION.md", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/parity-evidence-07.json", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/perf-evidence-07.md", ".planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/tdd-red-evidence-07-01-task1.json", "internal/adapters/plugin/builtin_run.go", "internal/adapters/plugin/builtin_run_test.go", "internal/adapters/plugin/main_test.go", "internal/api/acceptance_gates_test.go"]
covered_digest: "v4:sha256:f9104f9816c5cc5bf79bf00cb93526a3e1ff27ee14c8ad66e8c0bfd76d29ac8a"
behavior_unverified: 0
overrides_applied: 0
unverified_prohibitions: 16  # judgment-tier (flagged) prohibitions assessed non-authoritatively below — human review recommended (ADR-550 D4)
mvp_mode_note: "ROADMAP marks Phase 7 mode: mvp but the goal is not a User Story (user-story.validate = false). Verification proceeded goal-backward on the merged must_haves per the ratified Phases 1-6 precedent recorded in every 07 plan ('MVP framing note'). No User Flow Coverage section — this phase is a milestone acceptance gate, not a user-story slice."
human_verification:
  - test: "Corpora parity legs (LINT-PARITY-01 letter, SC1): complete 07-USER-SETUP.md — export GOOGLEAPIS (fresh public github.com/googleapis/googleapis clone; the stale yadro 2024-02 checkout with 3889 protos is incompatible, Pitfall 7) and CORPUS_ARGS (EasyP-Parser-Compiler-Experiment/bench/corpus/), run bench/setup.sh, then fire /tmp/p7/run-parity.sh parity unchanged (or DEFER: accept pending_user_unblock as the milestone-close disposition per A5)."
    expected: "EITHER executed per-corpus blocks (api-common ~92 files, googleapis-nocompute ~7414 files; Gate A 0/0 x2 per binary per corpus with main-vs-main captured first; Gate B zero COMPILE; Gate C 4-tuple multiset equal after the NEW3 filter with non-empty shared set; Gate D reported — ~0 new-rule findings is the EXPECTED result, never 'fixed') appended to parity-evidence-07.json, OR the deferral explicitly ratified as adequate."
    why_human: "The corpora source is a private repo unreachable from this machine (spike-verified 2026-11-21; re-confirmed at verification time: gh exit 1, SSH exit 128, GOOGLEAPIS/CORPUS_ARGS unset) — only the user can provide access or ratify the deferral; no programmatic check can execute the leg."
  - test: "Perf gate measurement (LINT-PERF-01 letter, SC2): after corpora unblock, fire the committed perf protocol (perf.py median-of-5 interleaved, no-race, CI-equivalent flags) on googleapis-nocompute; api-common recorded, not gated. Or ratify the pending state alongside item 1."
    expected: "perf-evidence-07.md gate section rewritten with measured medians and the verdict median(new) <= median(main) on googleapis-nocompute (saved-lab margin ~3.8x is indicative only [A3/A6]); issue counts equal across each binary's runs."
    why_human: "The gate compares real wall-clock on a corpus that cannot materialize without user-provided access; the protocol is committed but the measurement physically cannot run here."
  - test: "Hadolint leg of `task lint` (SC3): run `task lint:docker` (hadolint on the root Dockerfile) once Docker is available on this machine."
    expected: "hadolint clean — the root Dockerfile is untouched by the milestone (last green stands; recorded as an honest environment-unavailable manual item in 07-03, never silently passed)."
    why_human: "Docker is not installed on this machine (binary absent) — the check cannot run here; a human with Docker runs it or accepts the recorded disposition."
  - test: "End-of-phase evidence review (harvested from 07-03 Task 3 human-check): review parity-evidence-07.json gate states + corpora_leg honesty, the MANUAL-ONLY rows in 07-VALIDATION.md, and the fact that REQUIREMENTS.md marks LINT-PARITY-01/LINT-PERF-01 [x] while their corpora letters remain pending_user_unblock."
    expected: "The automated core (teeth Gates A-D x2, golden corpus 94 byte-exact, SC3 battery, race fix) accepted as the milestone record, with the corpora letters explicitly pending — or unblocked and executed per items 1-2."
    why_human: "Accepting 'automated core + honestly pending letters' as the milestone-close disposition is a product decision (assumption A5), not a machine-checkable property."
---

# Phase 7: Приёмка — паритет корпусов и производительность Verification Report

**Phase Goal:** Финальные ворота milestone: идентичность замечаний текущему main на приёмочных корпусах и производительность не хуже main на большом корпусе
**Verified:** 2026-11-21T06:23:03Z
**Status:** human_needed (no gaps; 4 human items — corpora parity letter, perf gate measurement, hadolint/Docker, evidence-disposition review)
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

Merged must-haves: 3 ROADMAP Success Criteria + the four PLAN frontmatter truth sets (07-01..07-04). SC1/SC2 split into their automated-core and corpora-letter halves, mirroring the ratified two-leg disposal design (07-03) and the Phase-6 verification precedent.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC1 automated core: four-gate compare proven end-to-end on the teeth corpus — Gate A main-vs-main determinism x2 both sides (sorted raw JSON-line multisets), Gate B zero-COMPILE, Gate C 4-tuple multiset identity after the NEW3 filter, both directions, non-empty shared set (10 shared), Gate D exact 2/1/1; baseline oracle built from the RATIFIED 1e9a596 | ✓ VERIFIED | `TestAcceptanceGates` re-run by verifier under -race: ok 9.829s, worktree hygiene held (list == 1 after baseline rebuild); gate code read in full (acceptance_gates_test.go:266-420: gate order, purity rationale cited, anti-vacuity, raw-line diagnostic t.Logf-reported); independent script stack agrees — `/tmp/p7/run-parity.sh selftest` re-run exit 0 ("gate states match TestAcceptanceGates; both negative controls fail as designed") |
| 2 | SC1 corpora letter: identical (rule, file, line, column) multisets vs main on api-common (92 files) and googleapis-nocompute (7414 files), x2 runs, main-vs-main base captured | ? UNCERTAIN — needs human | Corpora source (private EasyP-Parser-Compiler-Experiment + compatible googleapis) unreachable: research spike 2026-11-21 + fresh attempts recorded in parity-evidence-07.json `corpora_leg.clone_attempts` (gh exit 1, SSH exit 128, env unset); re-probed by verifier this session — same result (gh exit 1, SSH exit 128). Not a code defect — external resource blocker with a ready-to-fire path. Anti-ratification honored: status is `pending_user_unblock`, never green. → Human Verification 1 |
| 3 | SC2 protocol recorded before any measurement: CI-equivalent flags, warmup 1 + median-of-5 INTERLEAVED, per-run {bin, wall_s, rss_mb, issues, rc}, no-race provenance (1e9a596 / 9a37392), machine context, corpus identity 92/7414, REHEARSAL data labeled no-verdict, pending gate definition median(new) <= median(main) | ✓ VERIFIED | perf-evidence-07.md read in full — all sections present; {wall_s, rss_mb} fields verbatim in perf.py (8 matches) and the doc; full rehearsal JSON record at /tmp/p7/out/perf-REHEARSAL-teeth-20261121-065221.json (keys: interleaved, machine, rows, time_wrapper, warmup, runs=3); REHEARSAL explicitly carries no verdict |
| 4 | SC2 gate: easyp lint on googleapis-nocompute not slower than main at CI-equivalent flags, measurement recorded | ? UNCERTAIN — needs human | Same corpora blocker as truth 2; the gate physically cannot run without the corpora. Protocol committed and rehearsed; verdict pending unblock. → Human Verification 2 |
| 5 | SC3 milestone battery green on the final state: full -race ./... suite (incl. plugin/generation, TestAcceptanceGates, golden corpus, standing-constraint cells), golangci zero issues, schema drift empty, hygiene clean, hadolint honest manual item | ✓ VERIFIED | Verifier re-ran the entire battery on HEAD 229c176: `bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...` → exit 0, **DONE 3718 tests, 1 skipped in 236.9s** (the pre-existing TestGenerateGoHeaderUsesCompilerVersion skip: protoc-gen-go not in PATH), 41 packages ok incl. internal/adapters/plugin (64.054s), internal/generation (43.364s), internal/api (3m54.6s); `bin/golangci-lint run ./...` → **0 issues**; one-tmp-dir schema-drift gate → diff empty (exit 0); standing-constraint cells `TestLintSuppressionParity|TestLintMatrix|TestReadSourceOnce` → ok; hygiene: coverage.out/easyp unstaged, internal/easypc + schemas/ last touched pre-phase-7 (5be9f7e / 690a9cc); hadolint honestly manual (Docker absent — binary not found this session) |
| 6 | 07-01: default runtime construction serialized at exactly ONE point (runtimeCreationMu + newRuntimeSerialized wrapping only the default factory value at builtin_run.go:74); both creation sites (builtin_run.go:159, builtin_impl.go:143) route through the shared field; three frozen seam overrides byte-untouched; builtin_impl.go zero-diff | ✓ VERIFIED | Code read: mutex wraps only `wazero.NewRuntimeWithConfig` (builtin_run.go:26-30), default assignment `runtimeFactory: newRuntimeSerialized` :74; greps: newRuntimeSerialized=3, runtimeCreationMu=4, raw default `runtimeFactory: wazero`=0, seam overrides `run.runtimeFactory = func`=4 (3 frozen + probe), `rt := run.runtimeFactory` in builtin_impl.go=1; `git log -- builtin_impl.go` ends at 5c9185f (pre-phase) — zero phase-7 diff |
| 7 | 07-01: previously poisoned packages green under -race (plugin + generation); full -race ./... suite green for the first time since Phase 1 | ✓ VERIFIED | Verifier's own suite run (truth 5): internal/adapters/plugin ok 1m4.054s, internal/generation ok 43.364s — both inside the 41-package green run; matches 07-01/07-03 claims |
| 8 | 07-01: two-scope concurrency pin exercises the PRODUCTION default factory (no seam override) and machine-validated RED precedes GREEN | ✓ VERIFIED | `go test -race -count=1 -run '^TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes$'` re-run: ok 1.228s; RED evidence re-validated through the machine checker this session: **RED_EVIDENCE_OK** (target_test_failed; 8 package tests failing with DATA RACE at wazero internal/version version.go:23 read vs :46 write); probe delegates to `newRuntimeSerialized(ctx, cfg)` (builtin_run_test.go:467, :525) — production constructor exercised |
| 9 | 07-02: baseline oracle reproducible and provenance-tracked — commit-presence gate (fail loudly, never skip), detached worktree at 1e9a596, plain no-race build, t.Cleanup + failure-path self-removal teardown | ✓ VERIFIED | `git cat-file -e` equivalent verified (commit exists; `1e9a596` referenced 7× in acceptance_gates_test.go incl. provenance comments); buildBaselineBinary = sync.OnceValues at acceptance_gates_test.go:138; worktree list == 1 after every run (checked after TestAcceptanceGates and selftest re-runs) |
| 10 | 07-02: /tmp/p7 instruments (compare.py, perf.py, run-parity.sh — throwaway, rebuildable) self-tested on the teeth corpus WITHOUT corpora, agreeing with TestAcceptanceGates, with non-vacuity negative controls | ✓ VERIFIED | Instruments present (/tmp/p7: 6837/9755/20899 bytes); selftest re-run by verifier: exit 0 — "SELFTEST PASSED — gate states match TestAcceptanceGates; both negative controls fail as designed"; negative-control logs show the designed failures (control (a) diverged pair → only-main 1/only-new 1 residual exactly 1/1; control (b) COMPILE row → "Gate B failed: COMPILE issues present … disposition, never proceed"); gate logs show Gate C shared 10 / only 0-0, Gate D delta 4 (2/1/1) |
| 11 | 07-03: parity/perf evidence committed and auditable; corpora legs honestly disposed (pending_user_unblock + anti-ratification + fresh attempts + MANUAL-ONLY rows); nothing invalid committed | ✓ VERIFIED | JSON gate re-run: baseline 1e9a596/41, head 9a37392/44, gate_c multiset_equal true, gate_d 2/1/1, corpora_leg.status=pending_user_unblock with 3 structured clone_attempts, stale_figure_note present; committed (docs(07-03) 0f3efec + 0cc549c in HEAD ancestry); evidence structure is a superset of the 06-03 contract (corpora_leg carries status/reason/clone_attempts from the precedent + spike/anti_ratification/unblock_path deltas); MANUAL-ONLY rows present in 07-VALIDATION.md:60-66 with the one-command path; stale '43' grep over all evidence docs + gates test: clean; hygiene clean (truth 5) |
| 12 | 07-04: WR-01/WR-02 closed — TestMain narrative states priming removes the race signal + structural pinning; the pin mechanically detects loss of serialization (held-mutex ordering probe, mutation-proven); corpus-config comment states the measured 35/44 (32/41) truth; production byte-identical | ✓ VERIFIED | Grep gates: 'then proves the production serialization'=0, 'removes the race signal'=1, 'pinned structurally'=1, 'every rule of'=0, '35 of the 44-name HEAD registry'=1, '32 of the 41-name baseline registry'=1, UNARY_RPC=1; WR-02 claims source-verified by verifier: policy.go STANDARD → MINIMAL seed + BASIC + DEFAULT; builder.go group counts MINIMAL 5 (incl. PACKAGE_NO_IMPORT_CYCLE) + BASIC 22 (incl. SYNTAX_SPECIFIED, FIELD_NOT_REQUIRED) + DEFAULT 8 = **35 of 44**; COMMENTS 7 + UNARY_RPC 2 never enabled; baseline 41 − NEW3-within-enabled ⇒ 32/41. Mutation check re-run by verifier: Lock/Unlock pair removed → pin **RED** with the exact planned message ("runtime construction must never overlap across scopes: a construction completed while the probe held runtimeCreationMu"); guard restored → `git diff --exit-code` clean; builtin_run.go last commit 1f82b61 (production byte-identical through 07-04) |

**Score:** 10/12 truths verified (2 uncertain — both corpora-blocked external legs; 0 present-but-behavior-unverified)

### MVP Mode Note

ROADMAP marks Phase 7 `Mode: mvp`, but the goal is not a User Story (`user-story.validate` → false). Per the ratified Phases 1-6 precedent recorded in every 07 plan ("goal-backward truths in must_haves carry the framing"), verification proceeded goal-backward on the merged must-haves above. No User Flow Coverage section: this phase is the milestone acceptance gate, not a user-story slice. If a literal User Story goal is wanted for auditability, run `/gsd:mvp-phase 7` and re-verify.

### Deferred Items

None possible — Phase 7 is the final phase of the milestone; no later phase exists to defer to. The two corpora-blocked truths stay as human verification items (honestly pending), not deferrals.

### Advisory (New Scope, Unevidenced)

None — this was an initial verification (no prior pass), and the anti-pattern scan found no new-scope findings.

### Required Artifacts

Note: `verify.artifacts` parsed 0 items from all four plans (frontmatter shape unsupported by the verb — exit 1 with total 0, nothing evaluated); artifacts were verified manually against each plan's `must_haves.artifacts` list. Same for `verify.key-links` (1/3, 1/4, 0/4, 0/3 — filename-grep false negatives on Go package contracts and planning-doc links, the known limitation documented in Phase 6's report); all links verified manually below.

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/adapters/plugin/builtin_run.go` | one-point serialization fix | ✓ VERIFIED | substantive (mutex-guarded constructor, godoc names root cause); wired (default factory + both creation sites through the shared field) |
| `internal/adapters/plugin/builtin_run_test.go` | two-scope pin + probe, frozen seams | ✓ VERIFIED | pin at :407 with held-mutex probe + peak==1; 4 seam assignments (3 frozen + probe); mutation-verified RED |
| `internal/adapters/plugin/main_test.go` | TestMain priming, corrected narrative | ✓ VERIFIED | primes lazy wazero global before m.Run; corrected WR-01 narrative |
| `internal/api/acceptance_gates_test.go` | permanent four-gate machinery + baseline oracle | ✓ VERIFIED | 425+ lines; Gates A-D in order with anti-vacuity; buildBaselineBinary reuses the package's buildEasypBinary pattern (lint_link_matrix_test.go:159 → acceptance_gates_test.go:287); green under -race |
| `.planning/.../tdd-red-evidence-07-01-task1.json` | machine-validated RED record | ✓ VERIFIED | checker re-run: RED_EVIDENCE_OK (target_test_failed), DATA RACE frames inside wazero internals |
| `.planning/.../parity-evidence-07.json` | auditable parity record | ✓ VERIFIED | all blocks present with executed values; JSON gate passes; committed |
| `.planning/.../perf-evidence-07.md` | committed perf protocol | ✓ VERIFIED | protocol/provenance/machine context/REHEARSAL/pending gate all present; {bin, wall_s, rss_mb, issues, rc} fields verbatim |
| `.planning/.../07-USER-SETUP.md` | corpora unblock path | ✓ VERIFIED | env-var table with Pitfall 7 revision-compatibility nuance + one-command verification path + expected results |
| `.planning/.../07-VALIDATION.md` | gates row + battery row + MANUAL-ONLY rows | ✓ VERIFIED | :43 gates row (internal/api, real command), :45 battery row ✅ green with composite command, :46 corpora/perf row ⬜ pending MANUAL-ONLY, Manual-Only table :60-66 |
| `.planning/.../07-0{1..4}-SUMMARY.md` | per-plan summaries | ✓ VERIFIED | all four exist with self-checks; commit SHAs verified in HEAD ancestry (4f20f59, 1f82b61, 0669aab, 0f3efec, 0cc549c, bc3f853, f1baa65, 3da252a) |
| `/tmp/p7/{compare.py,perf.py,run-parity.sh}` | throwaway instruments (not committed by design, 06-03 precedent) | ✓ VERIFIED | present, selftest exit 0 re-run; out/ carries gate logs + negative controls + rehearsal records; rebuildable per 07-02 spec |

No MISSING / STUB / ORPHANED artifacts.

### Key Link Verification

All links verified manually (the verb cannot see Go package contracts or planning-doc links — same false-negative limitation as Phase 6).

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| builtin_run.go | builtin_impl.go | shared run.runtimeFactory field covering both creation sites | ✓ WIRED | default `newRuntimeSerialized` :74; builtin_run.go:159 and builtin_impl.go:143 (`rt := run.runtimeFactory`) route through it; builtin_impl.go zero-diff in phase 7 |
| builtin_run_test.go probe | builtin_run.go newRuntimeSerialized (:26) | wrapper delegation through the lazy creation site | ✓ WIRED | probe calls `newRuntimeSerialized(ctx, cfg)` at :467 (phase 1) and :525 (phase 2 wrapper) |
| tdd-red-evidence-07-01-task1.json | internal/adapters/plugin | machine checker verdict | ✓ WIRED | RED_EVIDENCE_OK re-validated this session |
| builtin_run.go fix | 07-03 SC3 battery | full -race ./... gate | ✓ WIRED | suite green incl. plugin (1m4s) + generation (43s) on the fixed state |
| acceptance_gates_test.go | lint_link_matrix_test.go | buildEasypBinary reuse (house pattern) | ✓ WIRED | same package; defined :159, used :287 |
| acceptance_gates_test.go Gate B | golden_full_registry_test.go purity rationale | cited verbatim in the failure message | ✓ WIRED | Gate B message cites "golden_full_registry_test.go requireGoldenPurity" |
| testdata fixtures | teeth corpus (Go t.TempDir + /tmp/p7/corpus) | tracked fixtures copied preserving layout + two-key STANDARD config | ✓ WIRED | both implementations measure main 10 / head 14 with delta exactly 2/1/1 — the fixture shapes carried from 06-03 |
| parity-evidence-07.json | parity-evidence-06-03.json | evidence-block contract (corpora_leg status/reason/clone_attempts) extended with Phase-7 deltas | ✓ WIRED | structure compared: 07 corpora_leg = superset of 06-03's keys; baseline/head/automated_core blocks added per plan |
| 07-USER-SETUP.md | 06-USER-SETUP.md shape | unblock path + revision-compatibility nuance | ✓ WIRED | env-var table + dashboard item + verification commands with expected results; 'Pitfall 7' + 'revision-compatibility' present |
| perf-evidence-07.md | lab.py lint meta | {bin, wall_s, rss_mb, issues, rc} fields verbatim | ✓ WIRED | fields present in the doc and in perf.py (8 matches); provenance cross-references the battery record |
| /tmp/p7/run-parity.sh | compiler-lab bench (lab.py, setup.sh) | parity mode drives the stand from outside; setup.sh first, phase binaries second | ✓ WIRED | 26 lab.py references, 18 setup.sh references, GNU-sed patch present, GOOGLEAPIS guard present; lab checkout clean at 08cef10 (never modified) |
| run.runtimeFactory field | test seam overrides | plain assignable field preserved | ✓ WIRED | 3 frozen overrides intact + probe wrapper; field is not a singleton or unassignable type |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| TestAcceptanceGates gate inputs | parsed JSON issues | both binaries' real lint runs over the assembled corpus (subprocess, exit 0/1 accepted, 2 = STOP) | Yes | ✓ FLOWING |
| Gate D counts | NEW3 per-rule counts | HEAD binary output on the teeth corpus, pinned against parity-evidence-06-03.json | Yes | ✓ FLOWING (re-run reproduced 2/1/1) |
| Gate C projection | (RuleName, Path, Line, Column) multisets | baseline-oracle binary vs HEAD binary outputs (external pre-milestone oracle = real data source) | Yes | ✓ FLOWING |
| parity-evidence-07.json | gate states, SHAs, counts | executed battery + gates + fresh clone attempts | Yes | ✓ FLOWING |
| perf rehearsal | wall_s/rss_mb/issues | real timed runs (portable time wrapper) over the teeth corpus | Yes | ✓ FLOWING (labeled REHEARSAL, no verdict) |
| corpora_leg.per_corpus_results | — | does not exist by design (anti-ratification: no corpora → no results) | N/A | ✓ HONEST ABSENCE (not a hollow prop — the container is explicitly pending, never populated with mock data) |

No hardcoded-literal or mock-sourced rendered values found.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Full -race suite (SC3 leg a) | `bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...` | DONE 3718 tests, 1 skipped (pre-existing), exit 0, 41 packages ok, 236.9s | ✓ PASS |
| Serialization pin (single named test) | `go test -race -count=1 -run '^TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes$' ./internal/adapters/plugin/` | ok 1.228s | ✓ PASS |
| Four-gate acceptance compare (single named test) | `go test -race -count=1 -run '^TestAcceptanceGates$' ./internal/api/` | ok 9.829s; worktree list == 1 after | ✓ PASS |
| Mutation check (guard removed → pin RED → restore) | sed remove Lock/Unlock pair; plain `go test -run '^TestBuiltinRun...$'`; `git restore` | FAIL with "runtime construction must never overlap across scopes: a construction completed while the probe held runtimeCreationMu" (exit 1); restore diff clean | ✓ PASS (RED as designed) |
| Standing-constraint cells | `go test -count=1 -run 'TestLintSuppressionParity|TestLintMatrix|TestReadSourceOnce' ./internal/core/` | ok | ✓ PASS |
| Script-stack agreement (selftest) | `bash /tmp/p7/run-parity.sh selftest` | exit 0; gate states match TestAcceptanceGates; both negative controls fail as designed | ✓ PASS |
| Golden corpus gate | inside the full suite (TestLintGoldenFullRegistry, goldenFindingCount=94) | package ok | ✓ PASS |
| Schema drift gate (SC3 leg c) | one tmp dir: schema-gen → diff -r schemas | diff empty, exit 0 | ✓ PASS |
| RED evidence machine check | `gsd-tools check tdd-red-evidence … --raw` | RED_EVIDENCE_OK (target_test_failed) | ✓ PASS |
| Worktree hygiene | `git worktree list | wc -l` after builds | 1 | ✓ PASS |

Full-suite discipline: one full-suite run total this verification; everything else single named tests/instruments.

### Probe Execution

No `scripts/*/tests/probe-*.sh` probes are declared or conventional for this repo (discovery: none found; the phase's executable checks are the Go tests, the /tmp/p7 instruments, and the schema-drift gate — all executed above). The phase-declared "probe rows" in the plans were planning-time probe reconciliations, not executable probe scripts.

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| (none declared) | — | — | N/A |

### Requirements Coverage

REQUIREMENTS.md maps exactly LINT-PARITY-01 and LINT-PERF-01 to Phase 7 (lines 109, 117); both are claimed in plan frontmatter (07-01: PARITY-01 enabler; 07-02: both; 07-03: both; 07-04: PARITY-01). No orphaned requirements.

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| LINT-PARITY-01 (REQ-3.1) | 07-01, 07-02, 07-03, 07-04 | Identical finding sets (rule, file, line, column) vs current main on api-common and googleapis-nocompute, all current rules | ✓ SATISFIED (automated core) / ? corpora letter needs human | Automated core: teeth-corpus Gates A-D x2 with the 1e9a596 main oracle (truths 1, 6-10, 12); corpora letter: pending_user_unblock with ready-to-fire path (truth 2 → Human Verification 1). Note: REQUIREMENTS.md marks the checkbox [x] and its text carries the stale "43 правила" figure — the evidence documents flag the figure as stale (lint_model.go:28-29) and state counts from the registries (41/44) without asserting it |
| LINT-PERF-01 (REQ-6.1) | 07-02, 07-03 | easyp lint on googleapis-nocompute not slower than current main | ✓ SATISFIED (protocol) / ? measurement needs human | Protocol committed before measurement with machine context, provenance, interleaved median-of-5, no-race rules (truth 3); gate measurement blocked on the same corpora (truth 4 → Human Verification 2); rehearsal machinery validated on the teeth corpus |

### Decision Coverage

Skipped cleanly — no 07-CONTEXT.md exists in the phase directory ("No CONTEXT.md - nothing to check.", `{skipped: true, blocking: false, total: 0, honored: 0, not_honored: []}`). Status impact: none.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| internal/adapters/plugin/builtin_run_test.go (pin + probe) | LINT-PARITY-01 (enabler) | yes | 0 | no | Behavioral (mutation-verified RED, peak==1, held-mutex ordering) | VALID |
| internal/adapters/plugin/main_test.go (TestMain priming) | LINT-PARITY-01 (enabler) | yes | 0 | no | Behavioral (race-signal removal narrative matches falsification experiments) | VALID |
| internal/api/acceptance_gates_test.go (TestAcceptanceGates) | LINT-PARITY-01 | yes | 0 | no | Behavioral (4 gates, value-level multiset identity, anti-vacuity, stderr cross-check) | VALID |
| internal/rules/golden_full_registry_test.go (golden corpus) | LINT-PARITY-01 (context) | yes | 0 | no | Value (94 findings pinned, 44-name coverage invariant computed at runtime) | VALID |

**Disabled tests on requirements:** 0
**Circular patterns detected:** 0 — the gates test writes only the corpus fixture (copied tracked testdata + two-key config), never expected values; expected values come from the baseline-oracle binary (external pre-milestone source) and the 06-03 committed counts (independent capture). Provenance: VALID.
**Insufficient assertions:** 0

### Prohibition Assessment (non-authoritative — human review recommended)

All 16 must_haves.prohibitions across the four plans are judgment-tier (`status: flagged`; no `verification: test`). Per the ADR-550 D4 soft-gate, this is a NON-AUTHORITATIVE LLM-judge assessment, each backed by the evidence above; none silently passes. Summary of the 16 (07-01: 3, 07-02: 5, 07-03: 3, 07-04: 4 — the set also documented in the 07-04 plan as "11 kept + the 07-04 additions"):

1. **07-01 one-point wrap only (no singleton/shared-runtime rewrite, seams untouched, call sites untouched)** — honored: mutex wraps only the default factory value; builtin_impl.go zero-diff; 3 frozen seams intact (code read).
2. **07-01 never upgrade wazero** — honored: go.mod still wazero v1.9.0.
3. **07-01 never weaken/skip the reproducer or carve out -race** — honored: TestBuiltinRunCompilesOnceAndIsolatesInvocations present and ran in the full green suite; no carve-out (full ./... ran).
4. **07-02 never compare unsorted / never raw-diff unfiltered (raw diff = reported diagnostic, never a gate)** — honored: sorted multisets + NEW3 projection in test and compare.py; raw diagnostic via t.Logf reported-not-gated.
5. **07-02 Gate B before Gate C; never modify lab.py/fixtures to force parity** — honored: Gate B asserted before compares (test order + compare.py load path); lab checkout clean at 08cef10, status empty.
6. **07-02 never assert the stale '43'** — honored: grep clean across all phase artifacts.
7. **07-02 never time a -race binary / single-run verdicts** — honored: protocol mandates no-race + warmup + median-of-5 interleaved + issue-count equality; only labeled REHEARSAL runs exist.
8. **07-02 never modify lab.py/bench; never place phase binaries before setup.sh** — honored: lab clean; run-parity.sh encodes setup-first ordering + GNU-sed patch on the throwaway checkout.
9. **07-03 never mark a corpora leg green without corpora (anti-ratification absolute)** — honored: status pending_user_unblock; no per_corpus_results fabricated.
10. **07-03 never commit coverage.out/easyp binaries/noise/hand-edited schemas; never re-run setup.sh after placing binaries** — honored: hygiene greps clean; schemas/easypc last touched pre-phase-7.
11. **07-03 never propagate the stale '43'** — honored (see 6).
12. **07-04 no production code changes** — honored: builtin_run.go last commit 1f82b61; mutation restore proven byte-identical.
13. **07-04 pin test name NOT renamed** — honored: TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes present (committed RED evidence references it).
14. **07-04 peak assertion never weakened** — honored: `require.EqualValues(t, 1, peak.Load(), "runtime construction must never overlap across scopes")` retained verbatim (:555).
15. **07-04 no schemas/coverage.out/binaries committed** — honored (see 10).
16. **07-01 prohibition 1's Close/error-joining semantics untouched** — honored: Close/error-joining code path read unchanged (close sync.Once, join semantics at builtin_run.go:84-104).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | no TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER debt markers; no placeholder/coming-soon/not-yet-implemented strings; no empty implementations; no hardcoded-empty props in any phase-7 file | — | — |

Hygiene notes (non-gaps, verified): the `.golangci.yml` lint_helpers exclusion is ABSENT — matching the documented state ("the exclusion was removed in the Phase-6 flip"; golangci runs 0 issues on ./... without it); `git status` carries pre-existing unrelated entries (.agents skill script, 03-VERIFICATION.md, untracked .planning/intel etc.) untouched by this phase; the REQUIREMENTS.md "43" figure and its [x] checkboxes are bookkeeping owned by the shared-ID gate, with the stale figure flagged (never asserted) in the evidence.

### Human Verification Required

### 1. Corpora parity legs (LINT-PARITY-01 letter / SC1)

**Test:** Complete 07-USER-SETUP.md (export GOOGLEAPIS = fresh public googleapis clone — Pitfall 7 revision-compatibility; CORPUS_ARGS = experiment repo bench/corpus/), run bench/setup.sh, then fire `/tmp/p7/run-parity.sh parity` unchanged. Or explicitly ratify the pending_user_unblock deferral as the milestone-close disposition (A5).
**Expected:** Either executed per-corpus blocks appended to parity-evidence-07.json (Gate A 0/0 x2 per binary per corpus with main-vs-main captured first; Gate B zero COMPILE; Gate C 4-tuple multiset equal after the NEW3 filter, non-empty shared set; Gate D reported — ~0 expected, never "fixed"), or the deferral ratified.
**Why human:** The corpora source is a private repo unreachable from this machine (spike + fresh attempts + verifier's own re-probe: gh exit 1, SSH exit 128, env unset) — only the user can provide access or accept the deferral.

### 2. Perf gate measurement (LINT-PERF-01 letter / SC2)

**Test:** After corpora unblock, run the committed perf protocol (perf.py, median-of-5 interleaved, no-race, CI-equivalent flags) on googleapis-nocompute; api-common recorded, not gated.
**Expected:** perf-evidence-07.md gate section rewritten with measured medians; verdict median(new) <= median(main) on googleapis-nocompute; issue counts equal across each binary's runs.
**Why human:** The gate physically cannot run without the corpora; the protocol is committed but the measurement is blocked on the same external resource.

### 3. Hadolint leg of `task lint` (SC3)

**Test:** Run `task lint:docker` (hadolint on the root Dockerfile) once Docker is available.
**Expected:** Clean — the Dockerfile is untouched by the milestone; the recorded manual-item disposition ("environment-unavailable, last green stands") holds.
**Why human:** Docker is not installed on this machine; the check cannot execute here.

### 4. End-of-phase evidence + disposition review (harvested from 07-03 Task 3 human-check)

**Test:** Review parity-evidence-07.json gate states + corpora_leg honesty, the MANUAL-ONLY rows in 07-VALIDATION.md, and that REQUIREMENTS.md marks LINT-PARITY-01/LINT-PERF-01 [x] while their corpora letters ride pending.
**Expected:** The automated core accepted as the milestone record (teeth Gates A-D x2, golden 94 byte-exact, SC3 battery, race fix) with the corpora letters explicitly pending — or unblocked and executed per items 1-2. Also review the 16 judgment-tier prohibitions (section above) — all evidence-supported as honored, non-authoritative.
**Why human:** Accepting "automated core + honestly pending letters" as the milestone close is a product decision (A5); prohibition sign-off requires explicit human resolution per the ADR-550 D4 soft-gate.

### Gaps Summary

No gaps. Every automatable claim was re-verified independently by this verification (not trusted from SUMMARYs): the SC3 battery re-run green end-to-end (3718 tests, -race, golangci 0 issues, schema drift empty), the four-gate acceptance machinery re-run green, the serialization pin re-run green AND mutation-proven RED by the verifier's own hands, the RED evidence re-validated through the machine checker, the script instruments re-run in agreement, the WR-01/WR-02 narrative corrections grep- and source-verified (35/44, 32/41 counts recomputed from builder.go), hygiene and anti-ratification contracts honored, and the stale-figure and lab.py-non-modification prohibitions mechanically checked.

The two remaining truths (SC1 corpora letter, SC2 gate measurement) fail only on an external resource: the acceptance corpora live in a private repo and a googleapis checkout this machine cannot reach — spike-verified at research time, freshly re-attempted at 07-03 execution, and re-probed again during this verification (same failures). The failure mode is a blocked resource, not missing or broken code: the ready-to-fire instruments, the unblock path, the committed protocol, and the anti-ratification evidence contract all exist and were exercised. Per the phase's ratified disposal design (A5) and the Phase-6 precedent, this is honestly surfaced as human verification — never silently passed, and never a code gap.

---

_Verified: 2026-11-21T06:23:03Z_
_Verifier: Claude (gsd-verifier)_
