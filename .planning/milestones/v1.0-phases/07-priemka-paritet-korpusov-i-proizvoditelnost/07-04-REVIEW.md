---
phase: 07-priemka-paritet-korpusov-i-proizvoditelnost
reviewed: 2026-11-21T05:57:12Z
depth: deep
files_reviewed: 3
files_reviewed_list:
  - /home/nil/DiskD/W/Djarvur/easyp/internal/adapters/plugin/main_test.go
  - /home/nil/DiskD/W/Djarvur/easyp/internal/adapters/plugin/builtin_run_test.go
  - /home/nil/DiskD/W/Djarvur/easyp/internal/api/acceptance_gates_test.go
findings:
  critical: 0
  warning: 1
  info: 2
  total: 3
  necessity: 1
status: issues_found
---

# Phase 07 Plan 04: Code Review Report (gap-closure diff 9e8ce79..HEAD)

**Reviewed:** 2026-11-21T05:57:12Z
**Depth:** deep
**Files Reviewed:** 3
**Status:** issues_found

## Summary

The 9e8ce79..HEAD range touches exactly the three in-scope test files (+138/−21) plus planning docs; **production code is byte-identical to 9e8ce79** (`git diff --name-only 9e8ce79..HEAD -- internal/` yields only `_test.go` files; `builtin_run.go` and `builtin_impl.go` untouched). No structural-findings pre-pass was provided.

Every load-bearing narrative claim was independently verified against source, and the mutation matrix was independently reproduced:

- **WR-01 TestMain narrative — verified accurate.** wazero v1.9.0 `internal/version.GetWazeroVersion()` writes the package-global `version` exactly once on the slow path and every later call takes the `len(version) != 0` early return (wazero `version.go:24,44`); it is reached from `NewRuntimeWithConfig` via the wazevo engine constructor (`internal/engine/wazevo/engine.go:113`); the two frozen-seam doubles in `builtin_run_test.go:292` and `:337` call `wazero.NewRuntimeWithConfig` directly, bypassing `newRuntimeSerialized`; the priming write on the TestMain goroutine has a happens-before edge to every test goroutine forked by `m.Run`. The corrected comment's central claim — priming *removes* the race signal, so `-race` cannot detect loss of `runtimeCreationMu`, and regression protection must be structural — is correct on all points.
- **Probe mechanism — verified sound.** Phase 1 (real time, outside the bubble): the mutex is held continuously from `<-holderLocked` until `close(releaseHolder)`, so a `constructed` send inside the dwell window is impossible with the pair intact (green is structural); the only unlock triggers are `releaseHolder` and `defer close(bodyDone)`, so a `require` Goexit anywhere in phase 1 still releases the holder (no mutex leak, no deadlock); `constructed` is buffered, so completions never park. The phase-1 comment's synctest rationale was verified against the go1.27.3 toolchain docs: mutex locking is explicitly non-durable, so holding the mutex across `synctest.Wait` would deadlock the bubble — the real-time phase is required. Phase 2 (bubble): `synctest.Wait` cannot return early with fewer than 4 results — the compile goroutine stays runnable/running until after `close(compilation.ready)` readies the waiter (defer order in `compile`: close-then-Done), and `close(gate)` readies scope B before scope A exits, so quiescence is reached only when all bubble goroutines have exited. `results`/`closes` are created outside the bubble but capacity-matched to their sends, so they never block and no bubble-isolation rule is violated.
- **Both production creation sites covered.** One echo invocation on a fresh scope constructs twice through the overridden factory — the lazy compiler runtime at `builtin_run.go:159` and the per-invocation runtime at `builtin_impl.go:143` — exactly as the comment states.
- **WR-02 comment — verified accurate.** `internal/rules/builder.go` groups: MINIMAL 5 + BASIC 22 + DEFAULT 8 = 35 of 44 total; COMMENTS = 7, UNARY_RPC = 2; STANDARD → `["MINIMAL","BASIC","DEFAULT"]` at `internal/config/v1/policy.go:109-116`. All three NEW3 names sit in MINIMAL/BASIC (`PACKAGE_NO_IMPORT_CYCLE` MINIMAL, `SYNTAX_SPECIFIED`/`FIELD_NOT_REQUIRED` BASIC), so Gate D is unaffected; baseline 41 = 44 − 3, enabled 32.
- **Mutation matrix independently reproduced** in a throwaway `/tmp` worktree (repo source untouched): Lock/Unlock pair removed → **RED 2/2** under `-race` with the exact planned message, *with TestMain priming in place* — independently confirming "no amount of priming can mask it"; var+pair removed → **compile RED** (`undefined: runtimeCreationMu` at `builtin_run_test.go:445,451`); restore byte-identical.
- **Empirical green:** probe 3/3 under `-race`; full plugin package `ok` (42s) under `-race`; `TestAcceptanceGates` PASS under `-race` (only-main 0, only-head 0); `go vet` clean; `bin/golangci-lint run` on both packages: 0 issues.

Three findings remain, all minor (0 Critical). The probe and the comment corrections are sound; the residuals are an unreachable defensive branch in phase 2, one overstated clause about the peak counter's evidentiary role, and one garbled sentence in the corrected TestMain narrative.

## Warnings

### WR-01: Impossible-state guard — phase-2 `run == nil` branch in the serialization probe

**File:** `/home/nil/DiskD/W/Djarvur/easyp/internal/adapters/plugin/builtin_run_test.go:511-515`
**Category:** necessity

**Issue:** The phase-2 `runScope` body defends against `EnsureBuiltinRun(baseCtx)` returning a nil owner, but that state cannot occur: `baseCtx` is `t.Context()`, which carries no `builtinRunContextKey` value (the phase-1 scope value lives only in the derived `probeCtx`, never in `t.Context()`), and `EnsureBuiltinRun`'s miss path always constructs and returns a non-nil `*BuiltinRun` (`internal/adapters/plugin/builtin_run.go:61-78` — the `ctx.Value` hit path returns the nil owner only when the value already exists). The branch is unreachable. It is also, on one of its two call sites, not actually protective: if it fired for the gate-opening scope (`"a"`, `openGate=true`), the early return skips `close(gate)`, parking the gated scope (`"bb"`) forever on a bubble channel and turning the test into a synctest deadlock-panic instead of the graceful result-err failure the channel-balancing (`closes <- nil`) is designed to produce. Phase 1 already uses the idiomatic fatal form (`require.NotNil(t, probeRun)`).

**Fix:** Replace the branch with `require.NotNil(t, run)` (mirroring phase 1), or delete it and rely on the factory's own behavior. If a nil-owner path is ever added to `EnsureBuiltinRun`, the factory assignment below would panic loudly anyway — no silent-path comment is needed for a state the production signature excludes.

## Info

### IN-01: "Secondary overlap detector" clause overstates the peak counter's evidentiary role

**File:** `/home/nil/DiskD/W/Djarvur/easyp/internal/adapters/plugin/builtin_run_test.go:494-498` (assertion at `:558`)

**Issue:** The phase-2 comment calls the retained `require.EqualValues(t, 1, peak.Load(), ...)` assertion "the secondary overlap detector behind the held-mutex ordering probe", but under the current test structure its failure condition is unreachable: phase-1's two constructions are sequential within a single invocation goroutine, and phase-2's wrapper windows are serialized by the test's own gate (scope B parks until scope A's whole loop closes the gate), so `peak` stays 1 regardless of whether production serialization exists. The counter's real residual value is pinning the *gate-staggering* — the property that keeps phase 2 deterministic (a future refactor that drops `waitGate` would push `inFlight` to 2 under `-race`, exactly the false-RED mechanism the deviation record describes). It contributes nothing to the production-serialization pin, which phase 1 owns alone. The comment's first clause ("constructions are deterministically ordered … so the wrapper windows never overlap and the peak counter stays at 1") is honest; only the "detector" label misstates what would be detected.

**Fix:** Reword to match the counter's actual role, e.g. "the peak==1 assertion pins the gate-staggered ordering that keeps this phase deterministic; the production-serialization pin is the held-mutex probe above." No code change required.

### IN-02: Garbled sentence in the corrected TestMain narrative

**File:** `/home/nil/DiskD/W/Djarvur/easyp/internal/adapters/plugin/main_test.go:16-18`

**Issue:** "the frozen runtimeFactory seam doubles in builtin_run_test.go call the raw wazero constructor directly" has a subject/verb mismatch and requires the reader to guess that "doubles" means "the two test doubles" and that they are the subject of "call". The factual content is correct (both `builtin_run_test.go:292` and `:337` bypass `newRuntimeSerialized`), but this phase's deliverable is narrative precision, and this is the one sentence a future evidence-time reader must decode.

**Fix:** e.g. "the two frozen-runtimeFactory doubles in builtin_run_test.go call the raw wazero constructor directly, so no easyp-side mutex can order those calls against each other or against the serialized default."

---

## Verification record

| Check | Result |
|---|---|
| Production code byte-identical to 9e8ce79 | PASS (only `_test.go` + `.planning/` in range) |
| TestMain narrative vs wazero v1.9.0 source | PASS (`version.go:24,44`; `engine.go:113`; raw-constructor sites `builtin_run_test.go:292,337`) |
| Probe phase-1 structural green / no false-RED | PASS (mutex held across whole dwell window; `bodyDone` covers require-Goexit) |
| Probe phase-1 no deadlock; synctest rationale | PASS (go1.27.3 docs: mutex lock non-durable; no `synctest.Wait` in held window) |
| Probe phase-2 quiescence (Wait ⇒ results==4) | PASS (compile goroutine runnable until after `close(ready)`; gate readies B before A exits) |
| Mutation: pair-only removal → RED | PASS (independently reproduced, 2/2, planned message, priming in place) |
| Mutation: var+pair removal → compile RED | PASS (`undefined: runtimeCreationMu`, `builtin_run_test.go:445,451`) |
| WR-02 comment figures (35/44, 32/41, COMMENTS 7, UNARY_RPC 2, NEW3 in MINIMAL/BASIC) | PASS (counted from `internal/rules/builder.go`; `policy.go:109-116`) |
| `go test -race -count=3 -run …SerializedAcrossScopes` | PASS 3/3 (0.12–0.19s each) |
| `go test -race -count=1 ./internal/adapters/plugin/` | PASS (42.1s) |
| `go test -race -run TestAcceptanceGates ./internal/api/` | PASS (6.3s; only-main 0, only-head 0) |
| `go vet` both packages | PASS |
| `bin/golangci-lint run` both packages | PASS (0 issues) |

---

_Reviewed: 2026-11-21T05:57:12Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: deep_
