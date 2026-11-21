# Phase 7 — perf measurement protocol (recorded; gate pending corpora unblock)

**Status:** the gate `median(new) <= median(main)` on googleapis-nocompute is **pending corpora unblock** — measured the day access unblocks. api-common perf is recorded, not gated. This document records the protocol, provenance, machine context, and rehearsal data; **no verdict is claimed**.

## Protocol (committed before any measurement)

- **CI-equivalent flags:** run from the corpus root with the corpus's committed `easyp.yaml` (STANDARD, MINIMAL auto-seed), no rule/config overrides. Primary format `--format json lint` (the lab.py lint shape — makes results comparable with the stand's meta and reuses the stderr `msg="lint completed" issues=N` cross-check); secondary plain `lint` (text) as a format-sensitivity sanity. Identical flags on both binaries — same-flag, same-corpus, same-machine.
- **Repetitions:** 1 warmup + 5 measured runs per binary per corpus, **INTERLEAVED** (m,n,m,n,...) to distribute load/thermal drift.
- **Timing:** `/usr/bin/time` for wall + peak RSS. This machine runs GNU Time (the BSD `-l` form is rejected — 07-02 Rule-3 deviation), so the portable wrapper probes BSD `-l` and falls back to GNU `-f "%e %M"` reading the same wall+RSS quantities.
- **Per-run record fields** reuse lab.py lint's meta verbatim: `{bin, wall_s, rss_mb, issues, rc}` (accepted rc {0,1}).
- **Determinism witness:** issue counts must be equal across all runs of a binary.
- **Gate:** `median(new) <= median(main)` on googleapis-nocompute. api-common recorded too (cheap), not gated.
- **Expected margin (indicative only):** ~3.8x from the saved lab meta (easypc 14.2 s vs main-engine 54.5 s, author's macOS machine) [A3/A6] — the verdict is measured fresh on unblock regardless; noise cannot flip a 3.8x gap, but the protocol keeps the verdict defensible regardless.
- **Anti-noise rules (T-07-05):** plain no-race builds only (detector overhead 5-10x invalidates timing); warmup + median-of-5 interleaved; machine context recorded; single-run verdicts never.

## Provenance

- **baseline ("main"):** commit `1e9a596` (RATIFIED A2 — zero Go-source delta to the first milestone commit `8f969d8`) — `git worktree add <tmp> 1e9a596 && (cd <tmp> && go build -o easyp-baseline ./cmd/easyp)`; worktree removed immediately after the build (no registration leak).
- **new:** commit `9a37392b322fd42f378b2210a0b7ccbd5db78e00` (battery-tested HEAD at 07-03 execution; every later 07-03 commit is .planning docs-only, Go source identical) — `go build -o easyp ./cmd/easyp` from the repo root.
- Both plain builds, **no `-race`**.
- **go version:** go1.27.3 linux/amd64.
- **Placement order (Pitfall 6):** `bench/setup.sh` FIRST (materializes corpora + builds the five lab bins), then overwrite `compiler-lab/bin/{easyp,easyp-baseline}` with the phase binaries. Re-running setup.sh resets them — order matters, and phase binaries are never placed before setup.sh (also: never re-run setup.sh after placing phase binaries).

## Machine context

| Field | Value |
|-------|-------|
| CPU | Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz |
| nproc | 16 |
| load at rehearsal | 3.59 2.28 2.09 |
| kernel | linux 7.3.4-3-t2-noble x64 |

(Load is sampled per measurement; the 07-03 battery-time load was elevated — the full `-race` suite was running concurrently. The gate compares same-machine interleaved medians, so background load affects both binaries' runs symmetrically.)

## Corpus identity (pending unblock)

- **api-common:** 92 files
- **googleapis-nocompute:** 7414 files

(Both from the lab README. Corpora materialize via `bench/setup.sh` from `$GOOGLEAPIS` + the experiment repo's `.args` lists; a stale 2024-02 googleapis checkout is incompatible — 3889 protos vs required ~7416, Pitfall 7. See `07-USER-SETUP.md` for the unblock path.)

## REHEARSAL (teeth mini-corpus, 2026-11-21 — no verdict; machinery validation only)

The `/tmp/p7/perf.py` rehearsal on the in-repo teeth corpus (1 warmup + 3 measured, interleaved m,n,m,n, GNU time wrapper, issue counts equal across each binary's runs — 10/14, matching the parity gates):

| binary | wall | vs main | RSS | issues |
|---|---|---|---|---|
| main | 0.06s (0.05-0.07) | 1.00x | 29MB | 10 |
| new | 0.06s (0.06-0.07) | 1.02x | 30MB | 14 |

Full record with per-run `{wall_s, rss_mb, issues, rc}` rows: `/tmp/p7/out/perf-REHEARSAL-teeth-20261121-065221.{json,md}` (rebuildable via `/tmp/p7/run-parity.sh selftest` per the 07-02 spec if /tmp was wiped). **REHEARSAL = instrument validation only — the mini corpus is ~5 files, not a corpora-scale verdict.**

## Corpora gate state

`corpora: UNREACHABLE (gh + SSH + env + stale local googleapis; research spike 2026-11-21, fresh attempts re-recorded at 07-03 execution in parity-evidence-07.json corpora_leg.clone_attempts) — perf gate pending_user_unblock.`

When unblocked: `/tmp/p7/run-parity.sh parity` fires the parity legs (setup.sh with GOOGLEAPIS/CORPUS_ARGS, phase-binary placement second, lab.py lint x2x2 per corpus, Gate A lint-diff, compare.py Gates B/C/D), then the perf protocol (perf.py median-of-5 interleaved on googleapis-nocompute — the gate; api-common recorded) — unchanged, zero plan changes needed. The gate section of this file is then rewritten with the measured medians + verdict, and `parity-evidence-07.json`'s `corpora_leg.status` flips to `executed` with `per_corpus_results`.

## SC3 battery record (cross-reference)

The SC3 battery legs, exit codes, runner, and the hadolint manual item live in `07-03-SUMMARY.md` (single source of truth for the battery); the provenance above (SHAs, go version, no-race builds, machine context) applies to both the battery and this protocol.
