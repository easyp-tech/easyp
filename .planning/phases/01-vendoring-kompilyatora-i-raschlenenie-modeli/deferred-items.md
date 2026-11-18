## Deferred Items

- Pre-existing wazero data race fails `internal/adapters/plugin` under `-race` on this machine (go1.27.3)
  status: open
  **What:** `go test -race -count=1 ./internal/adapters/plugin/...` fails with `WARNING: DATA RACE` inside `github.com/tetratelabs/wazero@v1.9.0/internal/version.GetWazeroVersion()` (write in `wazevo.NewEngine` under one `sync.Once` racing a read under another); every `TestBuiltin*` test in the package is poisoned by the detector.
  **Attribution:** Reproduced on pristine HEAD with `internal/easypc` and `internal/protoast` moved out (2026-11-18, plan 01-01 execution) — pre-existing and unrelated to the vendor drop. Upstream third-party code; nothing in `internal/adapters/plugin` imports the vendored trees.
  **Disposition:** Out of scope per executor scope-boundary rule; not fixed in 01-01. Note `task test` (CI, go 1.26.6) may or may not surface the same window — worth a look before the next `-race` gate is treated as blocking.
