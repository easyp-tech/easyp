# Phase 7: Приёмка — паритет корпусов и производительность - Research

**Researched:** 2026-11-21
**Domain:** corpus-level lint parity vs pre-milestone main, performance measurement protocol, milestone-gate battery, external-corpus access gating
**Confidence:** HIGH on all mechanism/protocol design (every claim probed live this session); MEDIUM on corpora-dependent outcomes (corpora unreachable — honestly gated); the one genuinely open engineering item is the wazero race fix

## Summary

Phase 7 is the milestone's final acceptance: prove the post-flip single engine emits **identical remark sets vs current main** (rule, file, line, column) on the two acceptance corpora (api-common 92 files, googleapis-nocompute 7414 files), prove it runs **not slower than main** on the large corpus at CI-equivalent flags, and take the **milestone gates green** (`task test`, `task lint`, `task schema:check`).

**SPIKE VERDICT (executed live this session, 2026-11-21): the acceptance corpora are NOT reachable and NOT reconstructible on this machine.** All four unblock routes fail or are absent: `gh repo clone easyp-tech/EasyP-Parser-Compiler-Experiment` → "Could not resolve to a Repository"; SSH `git ls-remote` → "Repository not found"; `GOOGLEAPIS`/`CORPUS_ARGS` unset; the only local googleapis checkout (`/home/nil/DiskD/W/yadro/googleapis`, revision 2024-02-12) holds 3889 `.proto` files vs the corpus's required ~7416 — and the `.args` membership lists that define both corpora exist only inside the unreachable experiment repo (not in compiler-lab's git history, not on disk). The phase plan must therefore be built as **two legs**: a fully automated in-repo core that carries everything it can, plus ready-to-fire corpora scripts behind an explicit `checkpoint:human-verify` gate with the documented unblock path (06-USER-SETUP.md) — never silently passed.

The compare design itself is **proven end-to-end live this session** on a mini fixture: the baseline binary builds from the pre-milestone commit `1e9a596` (verified zero Go-source delta to the first milestone code commit `8f969d8` — two docs-only commits intervene), both binaries run the lab.py lint shape (`--format json lint`, exit 0/1, stderr `msg="lint completed" issues=N` cross-check), JSON issue shape is field-identical (`Message/Path/Position{Column,Filename,Line,Offset}/RuleName/SourceName`), and after filtering the 3 new rule names the output is **byte-identical on shared findings** — the raw-line delta was exactly the 4 new-rule findings. The spike also demonstrated the compare's biggest trap concretely: on a non-fully-compiling corpus the linked engine bails to COMPILE-only (1 COMPILE issue, zero rules run → 19-finding false diff), so a zero-COMPILE precondition gate is mandatory.

Two additional live findings shape the plan. First, the **wazero data race still reproduces** (package-level `-race` run of `internal/adapters/plugin` fails: `TestBuiltinRunCompilesOnceAndIsolatesInvocations`, race inside wazero v1.9.0's `internal/version.GetWazeroVersion()` package-global, triggered by concurrent `NewRuntimeWithConfig` from `builtin_run.go:144` and `builtin_impl.go:143`) — it poisons SC3's full-suite gate and needs a fix or a ratified disposition **in this phase**. Second, the compiler-lab checkout ships **tracked saved corpus outputs** (`results/lint/{api-common,googleapis-nocompute,compute}/{go-protoparser,easypc}.jsonl` + meta, Nov 2 2026, author's machine) — the lab's own engines' complete issue sets on the canonical corpora (477 and 69985 issues; main-engine 54.5 s vs easypc 14.2 s on googleapis-nocompute) — an indicative reference for expected counts and the perf ratio, but NOT an oracle (lab lineage ≠ easyp main).

**Primary recommendation:** build the phase as (1) an automated core — baseline binary from `1e9a596` via `git worktree`, a filter+project compare script validated on an in-repo mini corpus, golden-corpus byte-exact gate, main-vs-main determinism runs, wazero race fix, and the SC3 battery via direct go commands — plus (2) corpora legs implemented as complete, ready-to-run scripts (setup.sh → lab.py lint ×2 per binary → lint-diff → perf protocol) that fire the same-day the user unblocks the experiment repo, with every corpora-gated claim honestly disposed as manual-verification pending unblock.

## Project Constraints (from AGENTS.md)

- **Do not** commit secrets, the local `easyp` binary, `coverage.out`, or generated dependency noise; corpora artifacts live under gitignored paths outside the repo.
- **Do not** hand-edit generated schema JSON (`schemas/*.json` — six files); regenerate via `go run ./cmd/easyp schema-gen` and verify with the schema:check equivalent.
- **Do not** edit anything inside `internal/easypc` (sealed vendor copy).
- **Tests:** use `testify`; mirror neighboring patterns.
- After behavior changes run the relevant tests (at least touched packages); for schema changes run the `schema:check` equivalent before finishing.
- No parallel agent instruction files; this research adds none.
- Task targets preferred (`task test`/`task lint`/`task schema:check`) — `task` is absent on this machine, so RESEARCH documents exact direct equivalents (see SC3 Battery).
- Project skills loaded: `go-code-style`, `go-testing` (testify, table-driven cells, exact-literal assertions) apply to any in-repo code the phase adds (the wazero race fix is production Go code and must follow them).

## User Constraints

No CONTEXT.md exists for this phase (`has_context: false` from init; `context_scope: complete`). No locked decisions or discretion areas were recorded. The ROADMAP research flag ("воспроизведение стенда lab.py — валидировать спайком до фазы или в её начале") is discharged by this document's SPIKE section.

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| LINT-PARITY-01 | Identical remark sets (rule, file, line, column) vs current main for all current rules on api-common (92 files) and googleapis-nocompute (7414 files), run twice, main-vs-main baseline captured first | Compare design validated live (Pattern 1): baseline binary from `1e9a596`, lab.py lint/lint-diff reuse, 4-tuple multiset gate + raw-line diagnostic, 3-new-rules filter (MINIMAL auto-seed), zero-COMPILE precondition gate, ×2 determinism runs, main-vs-main baseline = same-binary double run. Corpora legs gated on the access spike (UNREACHABLE — Pattern 5 disposal design) |
| LINT-PERF-01 | `easyp lint` on googleapis-nocompute not slower than current main at CI-equivalent flags; measurement recorded | Perf protocol (Pattern 3): same-machine interleaved runs, warmup + median-of-5, wall/RSS/issues recorded, both formats, both binaries identical flags; expectation ~3.8x margin from saved lab meta. Corpora-gated; protocol scripted now, executed on unblock |

## SPIKE RESULT — corpora reachability (question 1, answered live)

| Probe | Command | Result (2026-11-21) |
|-------|---------|---------------------|
| gh clone | `gh repo clone easyp-tech/EasyP-Parser-Compiler-Experiment /tmp/gsd-research/spike-experiment` | FAIL — "GraphQL: Could not resolve to a Repository with the name 'easyp-tech/EasyP-Parser-Compiler-Experiment'" (matches 06-USER-SETUP.md exactly) |
| SSH | `git ls-remote git@github.com:easyp-tech/EasyP-Parser-Compiler-Experiment.git HEAD` | FAIL — "ERROR: Repository not found." |
| Env vars | `echo $GOOGLEAPIS $CORPUS_ARGS` | both `<unset>` |
| Default lab path | `/tmp/gsd-research/EasyP-Parser-Compiler-Experiment` | absent |
| Local googleapis | `/home/nil/DiskD/W/yadro/googleapis` | exists but revision 2024-02-12 (`7d73737e0`), 3889 `.proto` total / 3888 excluding compute dirs — incompatible with the corpus (README: googleapis corpus = **7416** files, nocompute = **7414**) |
| `.args` lists anywhere | `find` over `/tmp/gsd-research`, compiler-lab git history (`--all --name-only`) | none on disk; **not in compiler-lab history**; they live only in the experiment repo (`bench/corpus/*.args`) |
| `bench/projects` | materialized corpora | absent (setup.sh never ran here) |

**Conclusion:** LINT-PARITY-01's letter and LINT-PERF-01's letter are **unreachable without user action**. The unblock path is 06-USER-SETUP.md (clone EasyP-Parser-Compiler-Experiment or point `GOOGLEAPIS`/`CORPUS_ARGS` at compatible checkouts). Additional nuance discovered this session: `GOOGLEAPIS` must be a checkout **compatible with the args lists** — setup.sh exits ("missing import ...") on any missing file, so a fresh public `github.com/googleapis/googleapis` clone (not the stale yadro checkout) is the likely required source; the experiment repo may pin the revision (unknown — unreachable).

**What survives on disk (salvage, [VERIFIED: compiler-lab repo tracked files]):** `/tmp/gsd-research/compiler-lab/results/lint/` carries the lab's own complete corpus outputs — `api-common`: 477 issues (main-engine 0.72 s), `googleapis-nocompute`: 69985 issues (main-engine 54.5 s vs easypc 14.2 s), `compute`: 5872 issues (main 1212.8 s vs easypc 0.08 s — the quadratic-comment story that motivated the corpus split). `results/` is **tracked** in the compiler-lab repo (came with the clone; timings are the author's macOS machine — indicative, not this machine's baseline). After unblock, these files double as a historical cross-check: if easyp-main's fresh corpus output matches the saved `go-protoparser.jsonl`, the lab's old engine == easyp main on corpora and the saved `easypc.jsonl` becomes a third reference; any delta needs disposition against the Phase-4 main-wins ledger (lab divergences are known and deliberate).

**Feasibility proven in the same spike (what IS reachable):**
- compiler-lab checkout `/tmp/gsd-research/compiler-lab` @ `08cef10909c0722f7042ba17b5c11951f9c6b6f6` present and intact (SHA re-verified).
- Baseline binary builds: `git worktree add /tmp/... 1e9a596 && go build -o easyp-main ./cmd/easyp` → OK (module fetch worked; binary 50.3 MB). Worktree removed after probe; binary kept at `/tmp/gsd-research/spike-easyp-main`.
- HEAD binary builds → `/tmp/gsd-research/spike-easyp-head` (51.4 MB).
- Mini-fixture compare validated end-to-end (Pattern 1, evidence quoted there).
- `sed -i ''` BSD-ism in setup.sh confirmed broken on GNU sed (exit 2) — Pitfall 3.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Baseline "main" binary | git worktree @ `1e9a596` → `go build ./cmd/easyp` | — | The pre-flip engine is deleted from HEAD (Phase-5 flip 4238fbe); main's behavior only exists at the pre-milestone commit |
| Corpora materialization | external stand (`compiler-lab/bench/setup.sh`) | user checkpoint | Corpora are derived from $GOOGLEAPIS + args lists; outside the repo by design |
| Parity run + diff | external stand (`compiler-lab/bench/lab.py lint`, `lint-diff`) | phase filter/projection script | The stand is the ROADMAP-named instrument; identity projection (4-tuple) + new-rule filter is phase-local |
| Perf measurement | phase script over `/usr/bin/time -l` | lab.py lint meta | Same shape as the stand (`--format json lint`), plus repetitions the stand's `lint` verb lacks |
| Milestone gates (SC3) | repo toolchain (go test/golangci-lint/schema-gen) | — | Direct equivalents of task targets; `task` absent |
| wazero race fix | `internal/adapters/plugin` (production code) | possibly `internal/generation` | Race is in easyp's concurrent `NewRuntimeWithConfig` call sites; SC3's `-race ./...` gate is poisoned until fixed |
| New-rules delta accounting | phase compare script | — | MINIMAL auto-seed emits 3 rules main lacks; filter + attributed delta report |

## Standard Stack

### Core (no new external packages — tooling only)

| Surface | Location/Version | Purpose | Why Standard |
|---------|------------------|---------|--------------|
| compiler-lab checkout | `/tmp/gsd-research/compiler-lab` @ `08cef10` [VERIFIED: rev-parse this session] | lab.py stand, setup.sh, saved reference outputs | ROADMAP names the stand; Phase-1–6 precedent |
| lab.py `lint`/`lint-diff` | `/tmp/gsd-research/compiler-lab/bench/lab.py` [VERIFIED: read lines 161–218] | snapshot `--format json lint` (sorted JSONL + meta) and Counter diff report | The instrument REQUIREMENTS' verification table points at; sorts lines (neutralizes main's emission-order nondeterminism) and cross-checks stderr `issues=N` |
| setup.sh | same dir [VERIFIED: read] | materialize corpora from $GOOGLEAPIS + args; build lab bins + buf v1.57.0 + fdsdump | Only corpora source; writes `easyp.yaml` STANDARD per corpus |
| Baseline binary | `git worktree` @ **`1e9a596`** → `go build` [VERIFIED: builds this session] | "current main" side of parity + perf | Last commit before the first milestone code commit `8f969d8` (`chore(01-01)` go.mod-only); docs-only commits after it don't affect the binary |
| HEAD binary | `go build -o ... ./cmd/easyp` [VERIFIED: builds] | "new" side | The phase-under-test artifact |
| `/usr/bin/time -l` | present [VERIFIED] | wall + peak RSS per run | lab.py's own mechanism (lines 63–72) |
| python3 3.15.2 / rsync 3.2.7 / gh / go 1.27.3 | present [VERIFIED] | setup.sh + lab.py prerequisites | — |
| `bin/golangci-lint` 2.14.0 | present in repo [VERIFIED: `ls bin/`] | `task lint` Go leg | Taskfile-pinned |
| gotestsum v1.13.0 | **absent** from `bin/` [VERIFIED]; installable `GOBIN=$PWD/bin go install gotest.tools/gotestsum@v1.13.0` | `task test` exact command | Taskfile-pinned version; Go module proxy reachable (two builds this session) |
| testify / existing deps | in go.mod | any in-repo code (race fix) | Project convention |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| lab.py stand | hand-written phase runner | Stand is the named instrument, already sorts/counts/cross-checks; only a thin filter/projection wrapper is new (06-03 `compare.mjs` precedent) |
| Build baseline from `8f969d8` | build from `1e9a596` | `8f969d8` is go.mod-only; `1e9a596` is the cleaner "milestone start" definition — planner should ratify the commit (Open Question 3) |
| wazero side-fix only | wazero upgrade v1.10–v1.12 | Upgrade changes a direct dep touching the generate path (REQ-5.2 compatibility surface); fix-locally is smaller, upgrade needs upstream verification + ratification (A4) |
| Perf via lab.py `lint` only | small repetition loop around it | `lab.py lint` is single-run; the perf requirement needs warmup+median-of-N — a ~15-line loop calling the same shape |

**Installation:** none required for the automated core. On unblock: `setup.sh` installs `buf` v1.57.0 via `go install` (network) and builds lab binaries + fdsdump. Optional: `GOBIN=$PWD/bin go install gotest.tools/gotestsum@v1.13.0` to run the exact `task test` command.

## Package Legitimacy Audit

**No new external packages are installed by this phase.** Go modules are outside the seam's npm/pypi/crates coverage; the two package-adjacent items:

| Package | Registry | Status | Disposition |
|---------|----------|--------|-------------|
| `gotest.tools/gotestsum@v1.13.0` | Go modules | already pinned by Taskfile.yml (project-vetted, used by `task test` in every prior phase) | Approved — install only to reproduce the exact task command; plain `go test -race` is the documented fallback |
| `github.com/tetrabelabs/wazero` v1.9.0 → v1.10.0/v1.11.0/v1.12.0 exist [VERIFIED: `go list -m -versions`] | Go modules | upgrade is a **candidate** race fix, not installed by research; whether upstream fixed the version-global race is unverified [ASSUMED: A4] | Flagged — planner must gate any upgrade behind `checkpoint:human-verify` + upstream changelog check, and re-run the REQ-5.2 compatibility surfaces (breaking/generate) |

**Packages removed due to [SLOP] verdict:** none. **Flagged [SUS]:** none (no installs).

## Architecture Patterns

### System Architecture Diagram

```
                ┌────────────────────────────────────────────────────────────┐
                │ USER UNBLOCK (checkpoint:human-verify)                     │
                │  EasyP-Parser-Compiler-Experiment (args lists)  ✗ UNREACHABLE (spike)
                │  GOOGLEAPIS checkout compatible with args       ✗ absent    │
                └───────────────┬────────────────────────────────────────────┘
                                │ when provided (env vars)
                                ▼
   $GOOGLEAPIS ──► bench/setup.sh ──► bench/projects/{api-common,googleapis-nocompute}
                   (builds lab bins, buf v1.57.0, fdsdump; writes easyp.yaml STANDARD)
                                │                          then place phase binaries:
                                ▼                          bin/easyp = HEAD build
             ┌────────────────────────────────┐            bin/easyp-baseline = 1e9a596 build
             │ lab.py lint (cwd=corpus,       │◄───────────┘
             │  --format json, exit 0/1)      │   ×2 runs per binary per corpus (determinism)
             │  → {label}.jsonl + meta.json   │
             └───────────────┬────────────────┘
                             ▼
             ┌────────────────────────────────┐
             │ lab.py lint-diff + phase       │
             │ filter/projection script:      │
             │  Gate A main-vs-main: 0 diff   │
             │  Gate B zero COMPILE (new run) │
             │  Gate C 4-tuple multiset ==    │  identity = (RuleName, Path,
             │   after 3-new-rules filter     │   Line, Column) per LINT-PARITY-01
             │  Gate D delta accounting (3)   │
             └───────────────┬────────────────┘
                             ▼                 ┌──────────────────────────────┐
                parity evidence (committed)    │ PERF (same corpora, same     │
                                               │ machine): interleaved runs,  │
 Automated core (no corpora):                  │ warmup 1 + median of 5,      │
   golden corpus 94 findings byte-exact        │ /usr/bin/time, wall+RSS,     │
   mini-fixture compare (validated live)       │ gate: median(new)<=median(m) │
   teeth corpus (06-03 precedent)              └──────────────┬───────────────┘
   main-vs-main determinism on mini corpus                     ▼
   wazero race fix (internal/adapters/plugin)     perf record committed to phase dir
   SC3: go test -race ./... + golangci run ./...  │
        + schema-gen diff gate                    ▼
                                             milestone green (SC3)
```

### Pattern 1: The lint-diff compare design vs current main (question 2) — validated live

**What "main" is.** The pre-flip engine was deleted at the Phase-5 flip (4238fbe, 135 files), so main's behavior exists only as a **binary built from the pre-milestone commit**. Definition: `1e9a596` ("fix(01): revise plans per checker feedback", docs-only) — verified zero Go-source delta to the first milestone commit `8f969d8` (`chore(01-01): promote go-billy`, go.mod-only); two docs-only commits (`a0a0ed9`, `26924b1`) intervene between them and git diff 1e9a596..8f969d8 touches only .planning/* + go.mod metadata (go-billy v5.6.2 indirect->direct at the same version). [MEASURED: git log 1e9a596..8f969d8 + git diff --stat, re-verified at revision 2026-11-21; worktree build of `1e9a596` succeeded.] Procedure:

```bash
git worktree add /tmp/p7/main-worktree 1e9a596
(cd /tmp/p7/main-worktree && go build -o /tmp/p7/bin/easyp-baseline ./cmd/easyp)
git worktree remove /tmp/p7/main-worktree          # registration must not leak
go build -o /tmp/p7/bin/easyp ./cmd/easyp          # HEAD
```

Place both at `compiler-lab/bin/{easyp,easyp-baseline}` (lab.py `--bin` choices are restricted to exactly these two names [VERIFIED: lab.py line 351]; `main()` requires all five bins to exist — run setup.sh first, then overwrite the two). Build **without** `-race` for perf runs (detector overhead invalidates timing).

**Rule counts (discrepancy resolution).** Main's registry = **41 rules** [VERIFIED: 41 distinct `GetRuleName(&T{})` entries in 1e9a596's `internal/rules/builder.go`, full list captured this session — Comment* ×7, casing/enum/field/import/message/oneof/package/rpc/service sets]; HEAD registry = **44** (41 + PACKAGE_NO_IMPORT_CYCLE, SYNTAX_SPECIFIED, FIELD_NOT_REQUIRED). REQUIREMENTS' "43" is the documented stale figure (`lint_model.go:28-29` already flags it). The gate therefore reads: for **every rule name either side emits**, the (rule, file, line, column) multisets are equal.

**The four gates (per corpus), in order:**

- **Gate A — main-vs-main baseline (captured first, per ROADMAP):** run the SAME baseline binary twice under two labels; `lab.py lint-diff` must report zero only-base/only-new. This captures the ROADMAP's "база main-vs-main снята" and proves the pipeline (sort + Counter) is noise-free and main's output is content-deterministic across runs (main's *emission order* is nondeterministic — lab.py sorts, so order noise is already neutralized [VERIFIED: lab.py line 180 `sorted(...)`]; content nondeterminism would surface here).
- **Gate B — zero-COMPILE precondition:** the new binary's run must contain **zero** `RuleName == "COMPILE"` issues. LINT-PARITY-01's precondition is "все файлы компилируются"; a COMPILE issue means the linked engine bailed to compile-error reporting and ran **no rules** — demonstrated live on the spike mini fixture (mixed testdata dirs → 1 COMPILE issue, 0 rule findings, 19-finding false diff vs main). If Gate B fires, either the corpus isn't in precondition state or the engine broke — do not proceed to Gate C on a failing corpus.
- **Gate C — identity:** primary gate per the requirement's letter: project every issue to the tuple **(RuleName, Path, Position.Line, Position.Column)**, multiset-compare main vs new **after filtering the 3 new rule names from the new side**; require empty diff both directions. Secondary (diagnostic, reported alongside): raw-JSON-line Counter diff (byte identity incl. Message + Offset). The spike proved raw-line equality holds on shared findings once the 3 names are filtered — if raw-line diff is ever nonempty while Gate C passes, disposition each delta line explicitly before sign-off (message text or Offset drift must be explained, never waved through).
- **Gate D — delta accounting:** report the filtered 3-name findings per corpus (name, count, sample). Expected 0 on proto3-heavy corpora [ASSUMED: A1 — syntaxed, no `required`, acyclic]; any nonzero is attributed and documented, NOT a failure (main lacks the rules — the delta is the expected MINIMAL-auto-seed effect, STATE.md Phase-6 note).

**Why the 3-name filter is mandatory:** every v1 config seeds `use` with MINIMAL (`policy.go:104` [VERIFIED in 06-RESEARCH with verbatim quote]), so the new binary runs all 44 rules on the corpora while main runs 41. Unfiltered, the compare false-fails on the new rules' findings. Filter validated live: spike residual diff was exactly the 4 new-rule findings, zero shared-finding divergence.

**Determinism (SC1 "прогон повторён дважды"):** each binary runs ×2 per corpus; Gates A/C/D evaluate on both run pairs (r1 and r2 must agree with each other and cross-side).

**Validated mechanics (spike evidence, 2026-11-21, quoting the run):** mini corpus = repo `testdata/{import_cycle,no_syntax,required_field}` trees + `easyp.yaml` STANDARD. Main: rc=1, 18 issues, stderr `msg="lint completed" issues=18`. Head: rc=1, 20 issues, `issues=20`. New-rule findings emitted: `PACKAGE_NO_IMPORT_CYCLE` ×2 (both files' import keyword, 5:1), `FIELD_NOT_REQUIRED` ×1 (6:18), `SYNTAX_SPECIFIED` ×1 (1:1). After filter: `only-main: 0`, `only-head: 0` on the common tree; raw-line Counter diff on common paths = exactly the 4 new-rule lines (`only-main: 0, only-head: [('PACKAGE_NO_IMPORT_CYCLE','import_cycle/a.proto'), ('PACKAGE_NO_IMPORT_CYCLE','import_cycle/b/b.proto'), ('FIELD_NOT_REQUIRED','required_field/required_field.proto'), ('SYNTAX_SPECIFIED','no_syntax/no_syntax.proto')]`). JSON field shape both sides: keys `Message, Path, Position{Column, Filename(=""), Line, Offset}, RuleName, SourceName` [VERIFIED: saved lab jsonl + live runs].

### Pattern 2: The compare/projection script shape (phase-local, thin)

```python
# /tmp/p7/compare.py — throwaway; its OUTPUT (evidence JSON) is the committed artifact
import json, collections, sys
NEW3 = {"PACKAGE_NO_IMPORT_CYCLE", "SYNTAX_SPECIFIED", "FIELD_NOT_REQUIRED"}
def load(label):
    rows = [json.loads(l) for l in open(f"{label}.jsonl") if l.startswith("{")]
    assert not any(r["RuleName"] == "COMPILE" for r in rows), "Gate B failed: COMPILE issues present"
    key = lambda r: (r["RuleName"], r["Path"], r["Position"]["Line"], r["Position"]["Column"])
    proj = collections.Counter(key(r) for r in rows if r["RuleName"] not in NEW3)
    delta = collections.Counter(key(r) for r in rows if r["RuleName"] in NEW3)
    return proj, delta, rows
base, base_d, _ = load("main-r1"); new, new_d, newrows = load("new-r1")
print("only-main:", len(base - new), "only-new:", len(new - base),
      "delta3:", sum(new_d.values()), dict(new_d))
```

(06-03 precedent: throwaway script under /tmp, not committed; the evidence record is the artifact.)

### Pattern 3: Perf measurement protocol (question 3)

- **Flags (CI-equivalent):** the plain invocation a CI/user makes — run from the corpus root with the committed `easyp.yaml` (STANDARD), no rule/config overrides. Primary format `--format json lint` (the lab shape; makes results comparable with the stand's meta and reuses lab.py's stderr `issues=N` cross-check); secondary plain `lint` (text format) as a format-sensitivity sanity. Identical flags on both binaries — the comparison is same-flag, same-corpus, same-machine.
- **Repetitions:** 1 warmup + 5 measured runs per binary per corpus, **interleaved** (m,n,m,n,...) to distribute load/thermal drift. Aggregate: median, min, max wall seconds + median RSS + issue count (issue counts must match across all runs of a binary — another determinism witness).
- **Recorded (the "зафиксирован" artifact):** per-run wall_s/rss_mb/issues, median table, binary provenance (commit SHAs, build flags, `go version`, no-race), machine context (CPU model, `nproc`, load), corpus identity (file counts 92/7414), date. Committed as `perf-evidence-07.md/json` in the phase dir.
- **Gate:** `median(new) <= median(main)` on googleapis-nocompute. api-common perf recorded too (cheap), not gated. Expectation from saved lab meta: ~3.8x margin (14.2 s vs 54.5 s, author's machine) [VERIFIED as the lab's numbers; ASSUMED as qualitative transfer: A6] — noise cannot flip a 3.8x gap, but the protocol keeps the verdict defensible regardless.
- lab.py `lint` is single-run — wrap it in a small repetition loop (or call the same argv under a loop) rather than modifying the stand.

### Pattern 4: SC3 milestone-gates battery (question 4)

`task` is absent on PATH [VERIFIED] and Docker is unavailable [VERIFIED: `docker info` fails], `bin/` holds golangci-lint but not gotestsum [VERIFIED]. Direct equivalents (from Taskfile.yml, read verbatim this session):

| Gate | Taskfile command | This-machine equivalent | Status |
|------|------------------|------------------------|--------|
| `task test` | `bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...` (precondition: gotestsum in bin/) | (a) install: `GOBIN=$PWD/bin go install gotest.tools/gotestsum@v1.13.0` then the exact command; or (b) `go test -race -count=1 ./...` | **Blocked by wazero race until fixed** (Pitfall 8). `coverage.out` is gitignored — never commit |
| `task lint` | `bin/golangci-lint run ./...` + `docker run --rm -i ghcr.io/hadolint/hadolint:... < Dockerfile` | `bin/golangci-lint run ./...` green; hadolint leg **unavailable** (no Docker) | Go leg mandatory; hadolint = honest manual-verification item (Dockerfile untouched by the milestone; last green stands) |
| `task schema:check` | `go run ./cmd/easyp schema-gen --out-dir "$tmp" && diff -r schemas "$tmp"` | identical (task adds nothing but the tmp dir + diff) | Direct equivalent, runnable now |

SC3 is satisfied when: full `-race ./...` suite green (after the wazero fix), golangci zero issues, schema diff empty. Any remaining environment-unreachable leg (hadolint) is recorded as such, never silently passed.

### Pattern 5: Honest disposal design if corpora stay unreachable (question 5)

Two-leg structure, neither leg silent:

**Automated core (runs regardless):**
1. Golden corpus gate — `TestLintGoldenFullRegistry`: 44 rules over the in-repo corpus, 94 findings byte-exact (Position incl. Filename/Offset) [VERIFIED: STATE.md Phase-6, golden re-captured in 5be9f7e].
2. Mini-fixture cross-binary compare — the exact Gates A–D machinery executed on an in-repo corpus (testdata trees + easyp.yaml), validating the script AND giving a real (small) main-vs-new parity datapoint. Live-validated this session.
3. Teeth corpus (06-03) parity record — carried evidence for the 3 new rules.
4. Main-vs-main determinism double-run on the mini corpus (Gate A rehearsed).
5. wazero race fix + full SC3 battery.
6. Salvage cross-reference: commit a summary of the saved lab corpus outputs (issue counts, timings, provenance) as indicative context.

**Corpora legs (implemented, gated, honestly disposed):**
- Build the complete runner now: setup.sh invocation (with `GOOGLEAPIS`/`CORPUS_ARGS`), binary placement, lab.py lint ×2 ×2, lint-diff, Gates A–D evaluation, perf protocol. Validate as much as possible **without** the corpora (mini corpus as the runner's test fixture).
- Every corpora-dependent claim lands as a `checkpoint:human-verify` item with the unblock path (06-USER-SETUP.md + the revision-compatibility nuance) and a PROVIDE/DEFER outcome gate (06-03 precedent: `parity-evidence-06-03.json` recorded clone attempts + options). If the user unblocks mid-phase or at verification, the same scripts fire unchanged; if deferred, VERIFICATION.md / UAT record carries LINT-PARITY-01-letter and LINT-PERF-01-letter as **pending human verification with the exact one-command path**, alongside the automated core's demonstrated parity evidence.

**Anti-ratification rule:** a corpora leg that "passed" without corpora does not exist. Evidence records must state the gate state explicitly (`corpora: UNREACHABLE (gh+SSH, 2026-11-21 spike) — leg pending user unblock`).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSONL snapshot + sorted diff | a new diff engine | `lab.py lint` + `lint-diff` (sorts lines, Counter compares, stderr cross-check, writes reports) | Proven instrument named by REQUIREMENTS; reimplementing risks subtle sort/dedup bugs |
| 4-tuple projection + 3-name filter | heavy framework | ~30-line throwaway script (Pattern 2) | Trivial, phase-local, evidence is the output not the code |
| Corpora assembly | hand-copying proto trees | `bench/setup.sh` from $GOOGLEAPIS + args (transitive-import closure, git init, configs) | setup.sh encodes corpus definition; hand assembly wouldn't be THE corpus |
| Timing | ad-hoc stopwatch | `/usr/bin/time -l` + warmup/median loop | RSS for free, matches lab meta semantics |
| Test runner formatting | custom reporter | gotestsum (install) or plain `go test -race` | Taskfile-pinned tooling |
| Race mitigation | replacing wazero / wrapper runtime | small synchronization fix at the two creation sites (or ratified dep upgrade) | Root cause is a package-global init inside wazero; the fix is contained |
| Baseline binary capture | pre-flip binary resurrected from artifacts | build from `1e9a596` worktree | No pre-flip binary survives (Phase-5 capture binaries were temp); the commit builds cleanly today [VERIFIED] |

**Key insight:** every mechanism this phase needs already exists — the only new code is a thin filter/projection script, a repetition loop, and the wazero synchronization fix. The work is *protocol discipline and honest gating*, not invention.

## Common Pitfalls

### Pitfall 1: COMPILE bail-out false-fails the compare (demonstrated live)
**What goes wrong:** one non-linking file → linked engine emits COMPILE issues and runs NO rules; main (per-file parser) still runs rules → massive false diff. Observed in the spike: mixed testdata corpus → head 1 issue (COMPILE 7:9) vs main 18 issues.
**Why it happens:** LINT-ENG-02 behavior (correct by design) meets a corpus violating LINT-PARITY-01's "все файлы компилируются" precondition.
**How to avoid:** Gate B (zero-COMPILE assertion) before Gate C; on failure, stop and disposition (corpus state vs engine bug).
**Warning signs:** new-side output with RuleName=COMPILE, or issue count collapsing to single digits.

### Pitfall 2: lab.py binary-name and existence constraints
**What goes wrong:** `lab.py lint --bin X` accepts only `easyp|easyp-baseline` (argparse choices, line 351), and `main()` exits unless **all five** `bin/{buf,easyp,easyp-baseline,easyp-protoc,fdsdump}` exist — a lint-only session still needs buf+fdsdump+easyp-protoc present.
**How to avoid:** run setup.sh first (builds all five), then overwrite `bin/easyp`/`bin/easyp-baseline` with the phase binaries. Re-running setup.sh resets them — order matters (Pitfall 6).
**Warning signs:** `missing .../bin/buf; run bench/setup.sh`.

### Pitfall 3: setup.sh `sed -i ''` is BSD-only
**What goes wrong:** line 101 (`sed -i '' 's#^module .*#...'`) fails on GNU sed ("can't read s#^module...: No such file or directory", exit 2 — verified live), aborting the googleapis-nocompute block after rsync but before git init/commit.
**Why it happens:** the script was written on macOS (results/ is tracked and came from the author's machine).
**How to avoid:** patch the throwaway checkout (`sed -i 's#...#...#'` GNU form) or pre-create the module line; for lint+perf legs the damage is nil (easyp.yaml is inherited from the googleapis project before the sed), for any breaking leg it's fatal.
**Warning signs:** setup.sh exiting mid-way with a sed error; nocompute project missing `.git`.

### Pitfall 4: Unfiltered 3-new-rules delta
**What goes wrong:** MINIMAL auto-seed (`policy.go:104`) makes the new binary emit `PACKAGE_NO_IMPORT_CYCLE`/`SYNTAX_SPECIFIED`/`FIELD_NOT_REQUIRED` findings main can never emit → compare false-fails.
**How to avoid:** filter the three names from the new side (Gate C) and account for them separately (Gate D). Validated live.
**Warning signs:** diff entries with those three RuleNames.

### Pitfall 5: Order noise misread as content diff
**What goes wrong:** main's emission order is nondeterministic (map iteration — REQUIREMENTS open question); naive line-by-line diff reports phantom changes.
**How to avoid:** lab.py sorts JSONL before writing; compares are Counter-based. Never compare unsorted streams.
**Warning signs:** identical multisets but "different" files.

### Pitfall 6: Binary placement vs setup.sh ordering
**What goes wrong:** placing phase binaries then re-running setup.sh overwrites them with lab binaries (silently wrong oracle).
**How to avoid:** setup.sh first, phase binaries second; record binary SHAs + build commands in evidence.
**Warning signs:** `bin/easyp` mtime predating the phase.

### Pitfall 7: GOOGLEAPIS revision incompatibility on unblock
**What goes wrong:** user points `GOOGLEAPIS` at an old checkout (e.g. the yadro one: 2024-02, 3889 protos) → setup.sh exits "missing import ..." because the args lists reference files that don't exist there.
**How to avoid:** unblock instructions must state the compatibility requirement; a fresh `github.com/googleapis/googleapis` clone is the expected source; the experiment repo may pin a revision (unknown while unreachable).
**Warning signs:** setup.sh aborts in the `.files` python step with `missing import X`.

### Pitfall 8: wazero race poisons SC3 (live-verified, still present)
**What goes wrong:** `go test -race -count=1 ./internal/adapters/plugin/` fails: `TestBuiltinRunCompilesOnceAndIsolatesInvocations` — race inside **wazero v1.9.0's own** `internal/version.GetWazeroVersion()` (package-global write at version.go:46 vs read at :23), triggered by two concurrent `wazero.NewRuntimeWithConfig` calls (easyp sites: `builtin_run.go:144` under BuiltinRun's mutex and `builtin_impl.go:143` on another path/instance). `internal/generation` passed this session's probe but was previously poisoned by the same root cause (Phase-3 verifier correction) — detection is scheduling-dependent (single-test isolation showed no race; package-level run did).
**How to avoid:** fix in this phase before the SC3 gate. Options: (a) easyp-side serialization of runtime creation (package-level `sync.Once`/mutex around `NewRuntimeWithConfig`, or a shared runtime); (b) upgrade wazero (v1.10.0–v1.12.0 available [VERIFIED: go list]) — but whether upstream fixed the version-global race is UNVERIFIED [ASSUMED: A4] and a dep upgrade touches the generate surface (REQ-5.2 compat) → ratify first. A carve-out is NOT acceptable per STATE ("fix owned before Phase 7 gate").
**Warning signs:** `WARNING: DATA RACE` referencing wazero internals under `-race`; suite green in isolation, red in package run.

### Pitfall 9: The "43 rules" figure
**What goes wrong:** plans/evidence assert "43 rules" parity; actual registries are 41 (main) and 44 (new); the number is a stale REQUIREMENTS figure already flagged in `lint_model.go:28-29`.
**How to avoid:** state rule counts from the registries (41/44) and define the gate as per-rule identity over the union of emitted names; propagate no "43".
**Warning signs:** any new doc/evidence repeating "43 правил" as an asserted count.

### Pitfall 10: /tmp volatility and worktree leakage
**What goes wrong:** binaries, corpora, and the baseline worktree all live in /tmp (reboot = gone); a forgotten `git worktree add` registration pollutes the repo's worktree list.
**How to avoid:** document rebuild commands in the plan; `git worktree remove` immediately after the baseline build (build binary, drop worktree); keep only binaries + evidence.
**Warning signs:** `git worktree list` showing /tmp entries; missing /tmp/p7 between tasks.

### Pitfall 11: Perf measured with wrong builds or noise
**What goes wrong:** timing a `-race` binary (5–10x overhead) or a single run verdict.
**How to avoid:** plain builds for perf (build once, reuse for parity too); warmup + median-of-5, interleaved; record machine context; issue-count equality across runs as sanity.
**Warning signs:** perf numbers wildly off the lab's order of magnitude; RSS missing.

### Pitfall 12: coverage.out committed
**What goes wrong:** `task test` equivalent writes `coverage.out` at repo root; AGENTS.md forbids committing it (and the local `easyp` binary).
**How to avoid:** leave gitignored, never stage; clean before commits.
**Warning signs:** `git status` showing coverage.out / easyp binary staged.

## Code Examples

### Baseline build + placement (the whole oracle setup)

```bash
# Source: git history verified 2026-11-21; lab.py main()/choices verified (lines 341–369)
git worktree add /tmp/p7/main-worktree 1e9a596
(cd /tmp/p7/main-worktree && go build -o /tmp/p7/bin/easyp-baseline ./cmd/easyp)   # VERIFIED: builds
git worktree remove /tmp/p7/main-worktree
go build -o /tmp/p7/bin/easyp ./cmd/easyp                                          # HEAD side
# after setup.sh materialized corpora + built the five lab bins:
cp /tmp/p7/bin/easyp           /tmp/gsd-research/compiler-lab/bin/easyp
cp /tmp/p7/bin/easyp-baseline  /tmp/gsd-research/compiler-lab/bin/easyp-baseline
```

### Corpus legs (fire the day the gate opens)

```bash
GOOGLEAPIS=/path/to/googleapis CORPUS_ARGS=/path/to/EasyP-Parser-Compiler-Experiment/bench/corpus \
  bash /tmp/gsd-research/compiler-lab/bench/setup.sh
# then place phase binaries (Pitfall 6), then, per corpus (api-common googleapis-nocompute):
python3 -I /tmp/gsd-research/compiler-lab/bench/lab.py lint --bin easyp-baseline --label main-r1 --corpora api-common googleapis-nocompute
python3 -I /tmp/gsd-research/compiler-lab/bench/lab.py lint --bin easyp-baseline --label main-r2 --corpora api-common googleapis-nocompute
python3 -I /tmp/gsd-research/compiler-lab/bench/lab.py lint --bin easyp          --label new-r1  --corpora api-common googleapis-nocompute
python3 -I /tmp/gsd-research/compiler-lab/bench/lab.py lint --bin easyp          --label new-r2  --corpora api-common googleapis-nocompute
python3 -I /tmp/gsd-research/compiler-lab/bench/lab.py lint-diff main-r1 main-r2 --corpora api-common googleapis-nocompute   # Gate A: expect 0/0
python3 -I /tmp/gsd-research/compiler-lab/bench/lab.py lint-diff main-r1 new-r1  --corpora api-common googleapis-nocompute   # + Gate C/D via compare.py
```

### SC3 battery (direct equivalents, Taskfile.yml verbatim shapes)

```bash
GOBIN="$PWD/bin" go install gotest.tools/gotestsum@v1.13.0   # once; else use plain go test
bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...
bin/golangci-lint run ./...
tmpd=$(mktemp -d) && go run ./cmd/easyp schema-gen --out-dir "$tmpd" && diff -r schemas "$tmpd"; rm -rf "$tmpd"
```

### The wazero race (root-cause evidence for the fix task)

```
WARNING: DATA RACE
Write at 0x...2453020 by goroutine 35:
  github.com/tetratelabs/wazero/internal/version.GetWazeroVersion()  version.go:46
  ... wazevo.NewEngine() → cache.initEngine → sync.Once → NewRuntimeWithConfig
  → easyp internal/adapters/plugin.(*BuiltinRun).compiledModule()    builtin_run.go:144
Previous read at 0x...2453020 by goroutine 17: same wazero path
Other easyp creation site: internal/adapters/plugin/builtin_impl.go:143 (run.runtimeFactory)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Old engine in-tree (parity = engine-vs-engine harness) | Baseline = binary from pre-milestone commit `1e9a596`; comparisons external | Phase-5 flip 4238fbe deleted the old path | This phase compares *binaries*, not packages |
| 41-rule registry | 44-rule registry (3 new, MINIMAL/BASIC) | Phase 6 (5be9f7e) | Compare must filter the 3 names; "43" is stale |
| lab.py `lint` single-run labels | phase adds ×2 determinism + median-of-5 perf loop | this phase's protocol | Determinism and perf become recorded evidence, not anecdotes |
| wazero race "deferred to Phase 7" | fix owned here | STATE.md blocker | SC3's full-suite gate depends on it |

**Deprecated/outdated:** REQUIREMENTS' "43 текущих правил" (stale figure); pre-flip parity harnesses (`assertBatchParity` died at the flip); any assumption that corpora outputs can be reconstructed without the args lists (3889 vs 7416 files proves it can't).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Acceptance corpora yield zero findings for the 3 new rules (proto3-heavy, syntaxed, acyclic, no `required`) | Gate D / Pitfall 4 | Gate D records nonzero delta — attributed, not fatal; teeth corpus already carries the rules' parity |
| A2 | `1e9a596` is the ratified "current main" baseline (milestone start) | Pattern 1 | Wrong baseline shifts the parity oracle — one-commit re-run to correct; planner should ratify (Open Question 3) |
| A3 | Lab saved outputs (author's machine, Nov 2) are indicative only; easyp-main may differ from lab's go-protoparser engine | SPIKE salvage / Gate E notion | A post-unblock cross-check delta needs disposition against the Phase-4 main-wins ledger — designed for, not assumed equal |
| A4 | Upgrading wazero (v1.10–v1.12) fixes the version-global race | Pitfall 8 | Wrong fix choice wastes a task; verify upstream changelog before ratifying; easyp-side serialization (option a) doesn't depend on this |
| A5 | User can eventually unblock the corpora (clone experiment repo + compatible googleapis) | Pattern 5 | LINT-PARITY-01/LINT-PERF-01 letters stay pending human verification — honestly recorded, milestone closes on automated core + pending items |
| A6 | The ~3.8x perf margin (lab meta) transfers qualitatively to this machine | Pattern 3 | Verdict still measured fresh on unblock; protocol doesn't depend on the expectation |
| A7 | CI-equivalent flags = plain `lint`/`--format json lint` with the corpus's committed config | Pattern 3 | If the milestone means different flags, the protocol re-runs cheaply; ratio robust to format choice |

## Open Questions (RESOLVED)

1. **Corpora unblock (the only true gate)**
   - What we know: spike failed on all routes 2026-11-21 (gh + SSH + env + stale local googleapis); unblock path documented in 06-USER-SETUP.md; revision-compatibility nuance added (Pitfall 7).
   - What's unclear: whether/when the user can provide the experiment repo (and which googleapis revision the args expect).
   - Recommendation: `checkpoint:human-verify` with PROVIDE/DEFER outcome recording (06-03 precedent); scripts built and mini-validated regardless.
   - **(RESOLVED)** 07-03 Task 3's corpora disposal design + end-of-phase human check (workflow.human_verify_mode=end-of-phase): fresh access attempt each run, honest pending_user_unblock with the 07-USER-SETUP.md one-command path; scripts built and self-tested in 07-02 Task 2 regardless.
2. **wazero fix: localize vs upgrade**
   - What we know: root cause = wazero v1.9.0's unsynchronized `GetWazeroVersion()` global hit by concurrent `NewRuntimeWithConfig` from easyp's two call sites; upstream has v1.10–v1.12.
   - What's unclear: whether upstream fixed it (needs changelog/issue-tracker check per the two-source verification rule); which fix the user prefers given REQ-5.2's compatibility surface.
   - Recommendation: default to easyp-side serialization (small, local, no dep change); upgrade only if ratified after upstream verification.
   - **(RESOLVED)** 07-01 flagged_assumptions A4 resolved-by-decision: easyp-side serialization of the default runtime factory chosen; the wazero upgrade stays documented-but-rejected this phase (reversible — a later ratified re-sync can revisit).
3. **Baseline commit ratification**
   - What we know: `1e9a596` = the baseline-oracle commit (measured: two docs-only commits intervene before the first milestone commit `8f969d8` — see RESOLVED below); `8f969d8` (first milestone commit) is go.mod-only.
   - What's unclear: none technically — this is a formality.
   - Recommendation: planner records `1e9a596` as the baseline SHA in the plan; UAT sees it in evidence provenance.
   - **(RESOLVED)** 07-02 flagged_assumptions A2 RATIFIED: baseline = `1e9a596`, kept as oracle — verified zero Go-source delta to the first milestone commit `8f969d8` (two docs-only commits intervene — a0a0ed9, 26924b1; git diff 1e9a596..8f969d8 = .planning/* + go.mod metadata only), superseding the research-time "last pre-milestone commit" wording above.
4. **Raw-line gate placement**
   - What we know: the requirement's letter is the 4-tuple; raw-line equality is strictly stronger and held on the spike.
   - What's unclear: whether a raw-line-only delta (e.g. Message text drift) should hard-fail or be dispositioned.
   - Recommendation: 4-tuple = gate; raw-line = reported diagnostic where every delta line must carry an explicit disposition before sign-off.
   - **(RESOLVED)** 07-02 prohibition 1: the 4-tuple is the gate; the raw JSON-line diff is a reported diagnostic whose every delta line needs an explicit Message/Offset disposition before sign-off — never a gate, never a silent pass.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | everything | ✓ | 1.27.3 | — |
| compiler-lab @ 08cef10 | stand, setup.sh, reference outputs | ✓ (SHA re-verified) | — | `gh repo clone easyp-tech/compiler-lab` + checkout 08cef10 |
| Baseline build (`1e9a596` worktree) | parity + perf oracle | ✓ (built this session) | — | — |
| EasyP-Parser-Compiler-Experiment | corpora args lists | **✗ UNREACHABLE** (spike) | — | checkpoint:human-verify; automated core carries what it can |
| GOOGLEAPIS checkout compatible with args | setup.sh | ✗ (yadro 2024-02 incompatible: 3889 vs 7416) | — | fresh public googleapis clone on unblock |
| `task` runner | SC3 commands | ✗ | — | direct equivalents (Pattern 4) |
| `bin/gotestsum` | `task test` exact command | ✗ (bin/ has only golangci-lint) | — | install v1.13.0 via go install, or `go test -race -count=1 ./...` |
| `bin/golangci-lint` | lint gate | ✓ | 2.14.0 pinned | — |
| Docker + hadolint | `task lint` docker leg | ✗ (daemon down) | — | Go leg only; hadolint = manual item (Dockerfile untouched) |
| buf v1.57.0 | lab.py existence check (not lint itself) | ✗ | — | setup.sh `go install`s it on unblock |
| python3 / rsync / `/usr/bin/time` / gh | setup.sh, lab.py, perf | ✓ | 3.15.2 / 3.2.7 / — / — | — |

**Missing dependencies with no fallback:** the acceptance corpora themselves (blocked on user; the phase is designed around it).
**Missing dependencies with fallback:** `task` (direct commands), gotestsum (install or plain go test), Docker/hadolint (Go lint leg + manual item), buf/fdsdump (setup.sh on unblock).

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | go test + testify; gotestsum via task test (absent → install or plain go test) |
| Config file | none needed (Go convention) |
| Quick run command | `go test -race -count=1 ./internal/rules/ -run 'TestLintGoldenFullRegistry'` + `go build ./cmd/easyp` |
| Full suite command | `go test -race -count=1 ./...` (SC3; currently RED on internal/adapters/plugin wazero race until fixed) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LINT-PARITY-01 | Golden corpus: 44 rules, 94 findings byte-exact | unit (in-repo scale gate) | `go test -race -count=1 ./internal/rules/ -run TestLintGoldenFullRegistry` | ✅ (golden_full_registry_test.go, post-5be9f7e) |
| LINT-PARITY-01 | Mini-corpus main-vs-new Gates A–D machinery | integration (external binaries, in-repo corpus) | `/tmp/p7` runner script over testdata mini corpus (build both binaries, run, filter, project, Counter) | ❌ Wave 0 (script; design validated live this session) |
| LINT-PARITY-01 | Corpora identity (92/7414 files, ×2 runs, main-vs-main first) | integration (external, **gated**) | lab.py lint ×2×2 + lint-diff + compare.py on real corpora | ❌ gated on checkpoint — runner scripted in Wave 0, executed on unblock |
| LINT-PERF-01 | Not-slower-than-main on googleapis-nocompute | measurement (external, **gated**) | perf loop (warmup 1 + median of 5, `/usr/bin/time -l`, both binaries, recorded) | ❌ gated on checkpoint — protocol scripted in Wave 0 |
| — (SC3) | Full suite green incl. race detector | unit/integration | `go test -race -count=1 ./...` | ✅ exists; RED until wazero fix (Wave 0 task) |
| — (SC3) | Go lint clean | static | `bin/golangci-lint run ./...` | ✅ |
| — (SC3) | Schema drift empty | gate | `go run ./cmd/easyp schema-gen --out-dir "$tmp" && diff -r schemas "$tmp"` | ✅ direct equivalent |
| — (regression) | Teeth corpus parity for 3 new rules (carried) | integration | 06-03 runner (parity-evidence-06-03.json) | ✅ carried |

### Sampling Rate

- **Per task commit:** build both binaries + golden test + scoped package tests of touched packages.
- **Per wave merge:** `go test -race -count=1 ./internal/...` (wazero fix included) + golangci on touched packages.
- **Phase gate:** full `-race ./...` (post-fix), golangci, schema diff, mini-corpus Gates A–D evidence, main-vs-main determinism evidence; corpora legs per checkpoint state (fired or honestly pending).

### Wave 0 Gaps

- [ ] Baseline build procedure task (`1e9a596` worktree → binary → worktree removed) — plan task, not a test file
- [ ] `/tmp/p7` runner: compare.py (filter/projection/Counter + Gate B assert) + perf loop + runner fixture (mini corpus) — validated against the live spike results before real use
- [ ] wazero race fix in `internal/adapters/plugin` (+ regression pin) — prerequisite for any full-suite gate
- [ ] gotestsum install (or documented plain-go-test fallback) for the exact task-test command
- [ ] Evidence templates: parity-evidence-07.json (gate states incl. explicit corpora gate state), perf-evidence-07.md
- [ ] none for framework/config — existing test infrastructure covers the rest

## Security Domain

### Applicable ASVS Categories (Level 1)

| ASVS Category | Applies | Standard Control |
|---------------|---------|------------------|
| V2 Authentication | no | CLI + git tooling; no auth surface |
| V3 Session Management | no | no sessions |
| V4 Access Control | no | private-repo access is the user's action, not agent code |
| V5 Input Validation | no (unchanged) | no new config/validation surfaces; corpora are parsed by the same easyp pipeline as any user project |
| V6 Cryptography | no | no crypto |

### Known Threat Patterns for this phase

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Untrusted downloaded content (googleapis clone, corpora) executed/parsed from unsafe cwd | Tampering/Info Disclosure | clone into a dedicated empty dir; run python with `-I`; pass paths as arguments; never run interpreters/build tools from inside downloaded trees (environment guidance) |
| Spoofed oracle binary (stale/wrong lab bin at bin/easyp-baseline) | Spoofing | record commit SHAs + build commands in evidence; rebuild from SHAs at phase start (Pitfall 6) |
| Silent gate bypass (corpora legs marked pass without corpora) | Repudiation | evidence records carry explicit gate state; checkpoint outcomes recorded (Pattern 5 anti-ratification rule) |
| wazero race as latent correctness hazard in production generate path | DoS/correctness | fix owned this phase (Pitfall 8); race detector as the proof |

No new attack surface: the phase measures and fixes; it adds no user-facing surface.

## Sources

### Primary (HIGH confidence — all read/probed this session)

- In-repo: `.planning/REQUIREMENTS.md` (LINT-PARITY-01, LINT-PERF-01, verification table), `.planning/STATE.md` (Phase-6 corpora-leg deferral, MINIMAL auto-seed note, wazero blocker), `.planning/ROADMAP.md` Phase 7 (SC1–SC3, research flag), `Taskfile.yml` (test/lint/schema:check verbatim), `go.mod` (wazero v1.9.0), `internal/adapters/plugin/builtin_run.go`/`builtin_impl.go` (runtime creation sites), repo `testdata/` (mini corpus source)
- Lab checkout: `/tmp/gsd-research/compiler-lab` @ 08cef10 — `bench/lab.py` (lint/lint-diff semantics, binary choices, existence check), `bench/setup.sh` (corpora derivation, BSD sed), `README.md` (corpus sizes 7416/7414, corpora location), tracked `results/lint/*/` (saved go-protoparser/easypc jsonl + meta for api-common 477, googleapis-nocompute 69985, compute 5872)
- Live probes (2026-11-21): gh clone + SSH ls-remote failures; env vars unset; yadro googleapis revision/count (3889); baseline worktree build at `1e9a596`; both binaries' lint runs + compare on mini corpus (numbers quoted in Pattern 1); GNU sed failure; `go test -race` package run reproducing the wazero race with full report; `go list -m -versions github.com/tetratelabs/wazero`; `git worktree` lifecycle

### Secondary (MEDIUM confidence)

- 06-RESEARCH.md and 06-USER-SETUP.md (lab.py lint shape, corpora materialization, MINIMAL seed quote, unblock path) — cross-checked against live probes where touched this session; no conflicts.

### Tertiary (LOW confidence)

- None. The research-plan seam (Context7/web providers) was skipped as inapplicable: the domain is fully code-internal (repo + local lab checkout + live probes); no external library documentation is needed, and config.json has all search providers disabled. The only externally-verifiable open item (whether wazero upstream fixed the race) is logged as A4 and deliberately NOT asserted.

## Metadata

**Confidence breakdown:**
- Compare design (baseline, gates, filter): HIGH — every mechanism verified live end-to-end this session, including the false-failure modes
- Corpora availability: MEDIUM-HIGH on the negative (four independent failing probes) — the positive outcome is genuinely unknowable until unblock
- Perf protocol: HIGH on design (instrument reuse, same-machine pairing), MEDIUM on absolute numbers (author's machine as expectation)
- SC3 battery: HIGH on equivalents; the wazero race is a verified live blocker with a characterized root cause
- Pitfalls: HIGH — 1, 2, 3, 4, 8 demonstrated or reproduced this session; the rest carried from verified state

**Research date:** 2026-11-21
**Valid until:** 2026-12-21 (stable in-repo domain; re-verify only the lab checkout SHA, `/tmp` artifacts, and the corpora gate at plan time)
