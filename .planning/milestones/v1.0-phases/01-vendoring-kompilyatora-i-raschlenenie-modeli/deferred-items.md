## Deferred Items

### Pre-existing wazero v1.9.0 data race (GetWazeroVersion) in internal/adapters/plugin
- **Status:** resolved
- **What:** `go test -race -count=1 ./internal/adapters/plugin/...` failed with `WARNING: DATA RACE` inside wazero v1.9.0's `internal/version.GetWazeroVersion()` (write in `wazevo.NewEngine` under one `sync.Once` racing a read under another); every `TestBuiltin*` test was poisoned by the detector. Reproduced on pristine HEAD with `internal/easypc`/`internal/protoast` moved out (2026-11-18, plan 01-01) — pre-existing, unrelated to the vendor drop; upstream third-party code.
- **Resolution:** FIXED by Phase 7, plan 07-01 — one-point construction serialization (package `runtimeCreationMu` + `newRuntimeSerialized` wrapping the default `runtimeFactory`) + `TestMain` priming of the lazy version global; mutation check (Lock/Unlock removed → pin RED with the planned message → restore byte-identical). Full `-race ./...` green for the first time in the milestone: 41 packages, 3718 tests, 0 FAIL. Verified by Phase 7 verifier and 07-04 review.
