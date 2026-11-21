# Phase 02 Deferred Items

Out-of-scope discoveries logged during execution (scope boundary: pre-existing issues
unrelated to the executing plan's changes are not auto-fixed).

## 2026-11-19 — Plan 02-03

| Item | Where | Evidence | Why deferred |
|------|-------|----------|--------------|
| Pre-existing `-race` data race + test failures in the plugin package (wazero `GetWazeroVersion` unsynchronized read, ~10 failing `TestBuiltin*`/`TestLocalPluginExecution*`/`TestMarshalRequest*` tests) | `internal/adapters/plugin` (third-party `tetratelabs/wazero@v1.9.0` internal version cache) | `go test -race -count=1 ./internal/adapters/plugin/` fails identically at plan 02-03's base commit `a77a13d` (verified in a throwaway worktree) and at plan HEAD `35fc26b`; plan 02-03 touches only `internal/api/lint_link_matrix_test.go` (test-only) | Unrelated to the linker phase's test-only plan; belongs to the plugin/WASM subsystem. Blocks the `go test -race -count=1 ./...` phase-gate reading — needs a dedicated fix (e.g. wazero version-cache sync or upstream re-sync) before the milestone's final acceptance gate. All plan 02-03 targeted suites (`./internal/api/`, `./internal/core/`) are green under `-race`. |
