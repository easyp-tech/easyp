# Phase 7: User Setup Required

**Generated:** 2026-11-21
**Phase:** 07-priemka-paritet-korpusov-i-proizvoditelnost
**Status:** Incomplete

Complete these items only if you want **LINT-PARITY-01's letter** (api-common 92 files, googleapis-nocompute 7414 files, run twice, main-vs-main baseline captured first) and **LINT-PERF-01's letter** (`easyp lint` not slower than main on googleapis-nocompute at CI-equivalent flags) executed on the acceptance corpora. The automated core already carries the phase meanwhile: teeth-corpus Gates A-D (`parity-evidence-07.json`), the golden corpus (94 findings byte-exact, 44 rules), and the SC3 battery (full `-race` suite + golangci + schema drift). These items unblock the corpora legs — now or at end-of-phase verification.

## Environment Variables

| Status | Variable | Source | Add to |
|--------|----------|--------|--------|
| [ ] | `GOOGLEAPIS` | Path to a googleapis checkout **COMPATIBLE with the corpus args lists** — a fresh `github.com/googleapis/googleapis` clone is the expected source. Pitfall 7 revision-compatibility: the stale yadro 2024-02-12 checkout (`/home/nil/DiskD/W/yadro/googleapis`) holds 3889 `.proto` files vs the corpus's required ~7416, and `setup.sh` aborts with `missing import ...` on any file the args lists reference but the checkout lacks | shell environment when invoking `bench/setup.sh` |
| [ ] | `CORPUS_ARGS` | Path to `EasyP-Parser-Compiler-Experiment/bench/corpus/` (the `.args` file lists defining api-common and googleapis-nocompute membership) | shell environment when invoking `bench/setup.sh` |

## Account Setup

Not applicable — no new account needed. The only access item is the private repository below.

## Dashboard Configuration

- [ ] **Clone EasyP-Parser-Compiler-Experiment (private, SSH remote — Phase 1-5 clone precedent) or point `GOOGLEAPIS`/`CORPUS_ARGS` at existing compatible checkouts**
  - Location: private easyp-tech repo / local filesystem path; full path + expected results in this file
  - Notes: the corpora `.args` file lists that define api-common/googleapis-nocompute membership live in this repo (not in compiler-lab's git history, not on disk); the executor cannot resolve it (`gh repo view` / clone and the SSH remote both failed on the 2026-11-21 research spike AND at 07-03 execution — records in `parity-evidence-07.json` `corpora_leg.clone_attempts`). Zero findings for a rule on a corpus is the EXPECTED result (proto3-heavy, syntaxed, acyclic — Gate D on the corpora is REPORTED, never a failure).

## Verification

After completing setup, verify the corpora legs can fire (executor/automation does the rest):

```bash
gh repo view easyp-tech/EasyP-Parser-Compiler-Experiment --json name  # resolves = access OK
GOOGLEAPIS=/path/to/googleapis CORPUS_ARGS=/path/to/EasyP-Parser-Compiler-Experiment/bench/corpus bash /tmp/gsd-research/compiler-lab/bench/setup.sh
```

Expected results:
- `bench/projects/api-common/` and `bench/projects/googleapis-nocompute/` materialize with ~92 and ~7414 `.proto` files respectively, each carrying `easyp.yaml` with `linters: default: STANDARD`
- Both parity binaries then run over each corpus (x2 per binary, main-vs-main captured first) and the filtered compares evaluate per 07-02's parity mode: Gate A 0/0, Gate B zero COMPILE, Gate C 4-tuple multiset equal after the NEW3 filter with a non-empty shared set, Gate D reported (expected ~0; nonzero attributed and documented)

**After setup, the executor fires `/tmp/p7/run-parity.sh parity` + the perf protocol unchanged** — the GNU-sed and GNU-time compatibility patches are idempotent (anchors asserted, hard STOP on upstream drift), setup.sh runs first and phase binaries are placed second (Pitfall 6). The perf gate (`median(new) <= median(main)` on googleapis-nocompute) is then measured per `perf-evidence-07.md`; api-common perf is recorded, not gated. Zero plan changes needed.

---

**Once all items complete:** Mark status as "Complete" at top of file. The unblocked day's full sequence: setup.sh (with the env vars above) → executor fires `/tmp/p7/run-parity.sh parity` → perf protocol → evidence extended (`corpora_leg.status` → `executed`, `per_corpus_results` filled; `perf-evidence-07.md` gate section rewritten with measured medians + verdict).
