# Phase 6: User Setup Required

**Generated:** 2026-11-21
**Phase:** 06-novye-pravila-i-regeneratsiya-shem-mcp
**Status:** Incomplete

Complete these items only if you want the LINT-NEWRULES-03 **corpora leg** (api-common 92 files, googleapis-nocompute 7414 files) executed. The automated requirement is already carried by the teeth-corpus parity evidence (`parity-evidence-06-03.json`) — these items unblock the acceptance-corpora extension, now or at Phase 7's lab.py stand spike.

## Environment Variables

| Status | Variable | Source | Add to |
|--------|----------|--------|--------|
| [ ] | `GOOGLEAPIS` | Path to a googleapis checkout usable by compiler-lab `bench/setup.sh` (lab default: `$LAB/../EasyP-Parser-Compiler-Experiment/googleapis`) | shell environment when invoking `bench/setup.sh` |
| [ ] | `CORPUS_ARGS` | Path to the directory holding the corpus `.args` file lists (`api-common.args`, `googleapis-nocompute.args`) — lives in EasyP-Parser-Compiler-Experiment (`bench/corpus/`) | shell environment when invoking `bench/setup.sh` |

## Account Setup

Not applicable — no new account needed. The only access item is the private repository below.

## Dashboard Configuration

- [ ] **Clone EasyP-Parser-Compiler-Experiment (private, SSH remote — Phase 1-5 clone precedent) or point `GOOGLEAPIS`/`CORPUS_ARGS` at existing checkouts**
  - Location: private easyp-tech repo / local filesystem path provided to the executor at the checkpoint
  - Set to: a checkout readable by the executor (e.g. `/tmp/gsd-research/EasyP-Parser-Compiler-Experiment` or any stable path)
  - Notes: the corpora `.args` file lists that define api-common/googleapis-nocompute membership live in this repo; the executor cannot resolve it (`gh repo clone` under an authenticated token and the SSH remote both failed on 2026-11-21). Zero findings for a rule on a corpus is the EXPECTED result (proto3-heavy, syntaxed, acyclic) — do not treat it as a failure.

## Verification

After completing setup, verify the corpora leg runs (executor/automation does the rest):

```bash
gh repo view easyp-tech/EasyP-Parser-Compiler-Experiment --json name  # resolves = access OK
GOOGLEAPIS=/path/to/googleapis CORPUS_ARGS=/path/to/EasyP-Parser-Compiler-Experiment/bench/corpus bash /tmp/gsd-research/compiler-lab/bench/setup.sh
```

Expected results:
- `bench/projects/api-common/` and `bench/projects/googleapis-nocompute/` materialize with ~92 and ~7414 `.proto` files respectively, each carrying `easyp.yaml` with `linters: default: STANDARD`
- Both parity binaries then run over each corpus and the filtered 3-name projection compares equal (see `parity-evidence-06-03.json` for the binary provenance and compare shape)

---

**Once all items complete:** Mark status as "Complete" at top of file.
