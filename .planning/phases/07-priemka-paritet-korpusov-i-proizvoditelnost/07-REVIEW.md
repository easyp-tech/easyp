---
phase: 07-priemka-paritet-korpusov-i-proizvoditelnost
reviewed: 2026-11-21T04:45:00Z
depth: deep
files_reviewed: 4
files_reviewed_list:
  - internal/adapters/plugin/builtin_run.go
  - internal/adapters/plugin/main_test.go
  - internal/api/acceptance_gates_test.go
  - .planning/phases/07-priemka-paritet-korpusov-i-proizvoditelnost/parity-evidence-07.json
findings:
  critical: 0
  warning: 2
  info: 0
  total: 2
  necessity: 0
status: issues_found
---

# Phase 07: Code Review Report

**Reviewed:** 2026-11-21T04:45:00Z
**Depth:** deep
**Files Reviewed:** 4
**Status:** issues_found

## Summary

The phase fixes the pre-existing wazero v1.9.0 lazy-global data race (`internal/version.GetWazeroVersion` writes an unsynchronized package global, `version = ret`, first-call only) via construction serialization. The production fix is **correct, complete, and load-bearing** — verified at four levels:

1. **Coverage:** the only production construction path in the repo is `BuiltinRun.runtimeFactory` (default `newRuntimeSerialized`, `builtin_run.go:74`), reached from exactly two call sites — `builtin_run.go:159` and `builtin_impl.go:143` — both behind the same package-level `runtimeCreationMu`. No other production `wazero.NewRuntime*` exists (grep incl. the sealed `internal/easypc`, which uses no wazero). `wazero.NewCompilationCacheWithDir` (the other `GetWazeroVersion` trigger, cache.go:58) is never used; easyp uses the in-memory cache.
2. **Mechanism:** the wazero v1.9.0 source confirms the race, and the committed RED evidence (`tdd-red-evidence-07-01-task1.json`) shows both racing stacks enter through `(*cache).initEngine` → `sync.Once` — each `EnsureBuiltinRun` scope owns its own `NewCompilationCache()`, so the per-cache `sync.Once` does not order constructions across scopes. Package-level serialization is therefore the right shape.
3. **Lock scope:** the mutex wraps only `wazero.NewRuntimeWithConfig`; `CompileModule`, `Instantiate`, and `Run` stay outside it. Per-invocation concurrency is preserved. Lock ordering (`r.mu` → `runtimeCreationMu`) has no reverse path, so no deadlock is possible.
4. **Load-bearing proof (experiments in a throwaway worktree, since removed):** HEAD as-is under `-race` is green (39.8s); with BOTH guards removed the original race reproduces (RED, same test set as the committed evidence) — so the mutex genuinely protects production, where no priming exists.

The three frozen `runtimeFactory` test seams (`builtin_run_test.go:289/335/364`) are untouched — the phase diff to `builtin_run_test.go` is purely additive (the new pin test at line 406). `TestMain` runs before `m.Run()` on the TestMain goroutine, so the priming write has a fork-level happens-before edge to every test goroutine; it constructs one throwaway runtime and closes it — test-only, no production impact.

**Evidence parity (parity-evidence-07.json vs. gates' actual outputs):** `TestAcceptanceGates` was re-run independently under `-race` on a pristine HEAD worktree: **PASS** (9.2s), and the raw-line diagnostic prints `only-main 0, only-head 0` — exactly the committed `raw_line_diagnostic` claim. Gate D's 2/1/1 delta and Gate A's double-run equalities are asserted by the test and passed. The git claims verified: `1e9a596^{commit}` exists; `1e9a596..8f969d8` is `.planning/*` + `go.mod` metadata only; `9a37392..HEAD` (the evidence's head SHA) is `.planning/*`-only, so the Go source state matches the evidence. The `shared 10 / 14` counts are arithmetically forced by the passing gates (Gate C identity + Gate D's 4 NEW3 findings ⇒ head raw = baseline + 4). Not independently re-verified: the milestone-wide "41 packages green under -race" claim (only the two packages central to this phase were re-run).

Two warnings, both about the *evidence narrative* rather than shipped behavior: the `-race` suite cannot actually detect loss of the serialization (priming masks the race — proven empirically), and the gates' comment misstates how many registry rules the corpus config enables.

## Warnings

### WR-01: False proof claim — TestMain priming masks the race, so `-race` cannot detect loss of the serialization; the pin does not pin

**File:** `internal/adapters/plugin/main_test.go:21-23` (claim), `internal/adapters/plugin/builtin_run_test.go:406-454` (pin)
**Issue:** The comment states: *"the -race detector then proves the production serialization for everything easyp constructs."* This is demonstrably false. `GetWazeroVersion` writes the global **once**; after TestMain primes it, every later call takes the `len(version) != 0` early return and never writes again — so the race signal is gone for the whole process. Measured in a throwaway worktree at HEAD:

| Variant | `-race` result |
|---|---|
| HEAD as-is (mutex + priming) | green, 39.8s |
| **mutex deleted, priming kept** | **green, 38.9s** |
| mutex deleted, priming deleted | RED (original race, same test set as tdd-red-evidence-07-01-task1.json) |

Consequence: `TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes` asserts only functional echo behavior across scopes — it passes with or without the mutex — and a future refactor that removes `runtimeCreationMu` (e.g., "the race is primed away anyway") reintroduces the **production** data race while the entire suite stays green. The mutex currently has zero regression detection; the comment asserts a proof that does not exist.

**Fix:**
1. Correct the claim in `main_test.go` — priming *removes* the race signal; the mutex protects production (which has no priming) and must be pinned structurally, not by the race detector.
2. Make the pin detect overlap directly instead of relying on `-race`. A test-only wrapper around the real constructor measures construction concurrency while still exercising the production path:

```go
var inFlight, peak atomic.Int64
run.runtimeFactory = func(ctx context.Context, cfg wazero.RuntimeConfig) wazero.Runtime {
    n := inFlight.Add(1)
    for {
        p := peak.Load()
        if n <= p || peak.CompareAndSwap(p, n) {
            break
        }
    }
    defer inFlight.Add(-1)
    return newRuntimeSerialized(ctx, cfg) // real production constructor
}
// ... after both scopes finish:
require.EqualValues(t, 1, peak.Load(), "runtime construction must never overlap across scopes")
```

This fires whenever constructions actually overlap — which the experiment shows happens reliably in this package's parallel test set — instead of being permanently blinded by the priming. Update both the pin's name/comment and the `main_test.go` comment so the evidence narrative matches what the suite can prove.

### WR-02: False coverage claim — the corpus config does not enable "every rule of each registry"; STANDARD enables 35 of 44 on HEAD (32 of 41 at baseline)

**File:** `internal/api/acceptance_gates_test.go:33-36`
**Issue:** The `acceptanceCorpusConfig` comment claims: *"The MINIMAL auto-seed (policy.go:109) then enables every rule of each registry: all 44 on HEAD, all 41 on the baseline."* Source-verified false. `policy.go:109` seeds only `MINIMAL`, and `default: STANDARD` appends exactly `BASIC, DEFAULT` (`internal/config/v1/policy.go:106-115`). `builder_linked.go` selects only the unwrapped `use` names (`res := make([]core.Rule, len(use))`), so the engine runs: MINIMAL 5 + BASIC 22 + DEFAULT 8 = **35 rules on HEAD**; the registry's remaining 9 names — COMMENTS (7) and UNARY_RPC (2) — are constructed into the registry map but never enabled by this config. At baseline 1e9a596 the same preset enables 32 of 41.

This is the same defect class the phase itself polices elsewhere (lint_model.go:28-29 flagging the stale "43" REQUIREMENTS figure; "the stale figure is never asserted"). Gate D is unaffected — all three NEW3 rules sit inside MINIMAL/BASIC (`PACKAGE_NO_IMPORT_CYCLE` in `addMinimal`, builder.go:137; `SYNTAX_SPECIFIED`/`FIELD_NOT_REQUIRED` in `addBasic`) — but any future evidence-time disposition reasoning from "all 44 rules active" would start from a wrong premise, and the comment contradicts the parity-evidence registry-vs-enabled distinction. For contrast, the golden test that genuinely fires all 44 names does so via `rules.AllRuleNames()` (golden_full_registry_test.go:55), not via `default: STANDARD`.

**Fix:** Reword the comment to the verified fact, e.g.:

```go
// convention. STANDARD expands to MINIMAL+BASIC+DEFAULT (policy.go:109): 35 of
// the 44-name HEAD registry, 32 of the 41-name baseline registry; the COMMENTS
// (7) and UNARY_RPC (2) registry rules are never enabled by this config. All
// three NEW3 rules are inside MINIMAL/BASIC, so Gate D is unaffected.
```

---

_Reviewed: 2026-11-21T04:45:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: deep_
