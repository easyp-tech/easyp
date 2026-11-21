# Phase 02 Deferred Items

Out-of-scope discoveries logged during execution (scope boundary: pre-existing issues
unrelated to the executing plan's changes are not auto-fixed).

## Deferred Items

### Pre-existing wazero v1.9.0 data race + ~10 failing tests in internal/adapters/plugin
- **Status:** resolved
- **What:** wazero `GetWazeroVersion` unsynchronized read poisoned ~10 `TestBuiltin*`/`TestLocalPluginExecution*`/`TestMarshalRequest*` tests under `-race`; failed identically at plan 02-03's base commit `a77a13d` and at HEAD `35fc26b` (verified in a throwaway worktree). Plan 02-03 was test-only.
- **Resolution:** FIXED by Phase 7, plan 07-01 — one-point construction serialization + `TestMain` priming; full `-race ./...` green (41 packages, 3718 tests, 0 FAIL); plugin + generation packages green 4/4 consecutive runs.
