---
last_mapped_commit: 5c9185fd17b72af60e0b7c50d6d713d466722555
last_mapped_at: 2026-11-17
---
# Codebase Concerns

**Analysis Date:** 2026-11-17

## Tech Debt

**Breaking-change detection ignores RPC streaming changes:**
- Issue: Both `TODO: check stream` markers are unfilled; adding/removing/changing `stream` on an RPC request or response is not reported as a breaking change.
- Files: `internal/core/breaking_checker.go` (lines ~109 and ~116)
- Impact: `easyp breaking` can pass on WIRE-breaking streaming changes, giving false confidence versus the buf-compatible contract.
- Fix approach: Compare `againstRPC.RPCRequest.IsStream` / `RPCResponse.IsStream` against the current RPC and emit an issue per direction; mirror the existing `getRPCRequestChangedTypeIssue` pattern and add a colocated test in `breaking_checker_test.go`.

**Insertion-point generation does not account for line terminators:**
- Issue: Explicit TODO: generated-file line terminator (CRLF vs LF) handling is skipped.
- Files: `internal/core/generate_insertion_point.go` (line ~27)
- Impact: Incorrect insertion offsets in generated files on Windows checkouts with CRLF.
- Fix approach: Detect the file's dominant line ending and adjust offset math; test with a CRLF fixture.

**Console error output is minimal:**
- Issue: `TODO: extend error output` — `console.RunError` prints command/error/stderr but has no `Unwrap`, so the underlying error is invisible to `errors.Is`/`errors.As`.
- Files: `internal/adapters/console/error.go`
- Impact: Callers cannot classify plugin/subprocess failures by cause; documented in `.spec/ERRORS.md`.
- Fix approach: Add `Unwrap() error { return r.Err }` and extend formatting; update dependents that match on message text.

**Large files with mixed responsibilities:**
- Issue: `internal/core/breaking_profiles.go` (1192 lines) and `internal/core/managed_mode.go` (1008 lines) concentrate profile definitions and managed-mode orchestration; `internal/adapters/gitmodules/snapshot.go` (617 lines) and `internal/migration/apply.go` (600 lines) are similarly dense.
- Impact: High change-risk, harder review, slow onboarding.
- Fix approach: Split breaking profiles into per-category files (`breaking_profile_file.go`, `breaking_profile_wire.go`, …); extract transaction staging from `apply.go` into a helper package.

**Russian-language comments in production code:**
- Issue: Non-English comments in gRPC plugin client, DOM helper and legacy config parser.
- Files: `internal/adapters/plugin/remote.go` (lines 42–89), `internal/core/dom.go` (~135), `internal/config/config.go` (~169)
- Impact: Inconsistent with the rest of the English-commented codebase; hampers contributor review.
- Fix approach: Translate comments in place; no behavior change.

**Static BSR compatibility mapping instead of a real BSR client:**
- Issue: `StaticResolver` maps exactly five hardcoded `buf.build/*` modules to pinned Git snapshots; every other BSR module fails with `UnsupportedModuleError`. BSR commit/digest equivalence is explicitly not verified (warned at runtime).
- Files: `internal/adapters/bsr/resolver.go`
- Impact: Users depending on any BSR module outside the list cannot migrate; pin equality to the requested BSR revision is unverified.
- Fix approach: Documented intentional boundary (`.spec/ERRORS.md`, `V1_RELEASE_NOTES.md`). Extending means adding entries or implementing a real BSR API client; keep the explicit-failure behavior.

**No automatic unknown-import-to-module discovery:**
- Issue: `resolveV1Lock` returns `cannot resolve imports` rather than inferring the owning Git module from a missing import path (X20 boundary).
- Files: `internal/modules/` (`locked_sources.go`, `resolve.go`), documented in `.spec/ERRORS.md`
- Impact: Users must declare every dependency explicitly in `protobuf.mod`.
- Fix approach: Intentional gap; any future discovery must not guess repositories (project rule). UX mitigation only (better error messages).

## Known Bugs

**No confirmed open bugs.** Recent fixes (see git log) addressed: repeated dependency walks and snapshot construction (#247), Windows native separators in relative-name checks (#249), nondeterministic plugin request encoding (#244), remote plugin dialing through passthrough (#246). The RPC-stream breaking-check gap above is the closest to a functional bug.

## Security Considerations

**Remote plugin TLS heuristic:**
- Risk: TLS selection in `RemotePluginExecutor.Execute` is host-string-based: any host containing `localhost` or `127.0.0.1` anywhere in the string, or any `http://` URL, gets `insecure.NewCredentials()`. A non-`http://` URL for a plain host also silently upgrades to TLS by heuristic rather than explicit opt-in.
- Files: `internal/adapters/plugin/remote.go` (lines 59–68, 122–127)
- Current mitigation: `https://` prefix forces system-pool TLS; plugin sources come from the project's `easyp.gen.yaml`.
- Recommendations: Match `localhost`/`127.0.0.1` as exact hosts (or use `url.Parse` hostname equality) rather than substring; log clearly when connecting insecurely.

**Git subprocess surface:**
- Risk: `exec.CommandContext(ctx, "git", args...)` in `internal/adapters/gitcommand/command.go` and `git ls-files` parsing in `internal/adapters/module_config/git_dependency_path.go` (this one uses `exec.Command` without context). Shell commands in `internal/adapters/console/bash.go` / `powershell.go` join parameters into a single `-c`/`-Command` string.
- Files: above; `internal/adapters/console/bash.go:20`, `powershell.go:20`
- Current mitigation: Git args are constructed from parsed config, not raw user shell input; five-minute per-operation timeout via `EASYP_GIT_TIMEOUT` on Git/cache-lock paths.
- Recommendations: Add context propagation to `git_dependency_path.go`; audit console command joining for arguments containing quotes when invoked with externally supplied parameter values.

**BSR snapshots trust pinned Git commits only:**
- Risk: BSR revision/digest are not verified against the Git snapshot (warned, not enforced).
- Files: `internal/adapters/bsr/resolver.go`
- Recommendations: If stronger guarantees are needed, record and compare BSR digests at migration time; currently an accepted documented trade-off.

## Performance Bottlenecks

**Module cache snapshot construction:**
- Problem: Snapshot construction over large Git mirrors was expensive enough to warrant a dedicated fix (#247) avoiding repeated dependency walks; immutable Git tree/blob metadata indexing exists to avoid repeated object reads.
- Files: `internal/adapters/gitmodules/snapshot.go`, `snapshot_reuse.go`, `git_object_sizes.go`, `object_cache.go`
- Cause: Repeated tree walks and object reads in large mirrors.
- Improvement path: Largely addressed; further work should reuse the immutable-tree index (`git_object_sizes.go`) rather than adding new walks.

**Remote plugin execution latency:**
- Problem: Hard-coded 30-second gRPC timeout per remote plugin call; no retry/backoff policy anywhere in the codebase.
- Files: `internal/adapters/plugin/remote.go` (~line 61)
- Improvement path: Make the timeout configurable; codegen of large modules may exceed 30s.

## Fragile Areas

**Migration apply transaction:**
- Files: `internal/migration/apply.go` (600 lines), `migration.go`
- Why fragile: Multiple sequential filesystem replacements are not crash-atomic; rollback failures are joined with original errors and depend on retained backup/recovery directories. Legacy `easyp.lock` and backups are the only recovery inputs.
- Safe modification: Preserve the stage/recheck/restore ordering; always `errors.Join` cleanup failures with the primary error; extend tests in `apply_test.go` covering rollback-failure paths.
- Test coverage: Good for ordinary failures; crash-mid-apply is untested (and untestable without injection points).

**`sourceview` unsupported-spelling boundaries:**
- Files: `internal/sourceview/view.go`, `local.go`, `internal/modules/resolved_file_observations.go`
- Why fragile: Many pointer spellings and file types return `ErrUnsupported`; each new legacy layout may add a new failure path surfaced only at runtime.
- Safe modification: Add a new spelling → add a case plus a fixture in `view_test.go` (526 lines of existing cases to mirror).

**`internal/core/managed_mode.go` (1008 lines):**
- Files: `internal/core/managed_mode.go`, `managed_mode_test.go`
- Why fragile: Concentrates consumer-project discovery and module selection logic consumed by generation; changes ripple into `internal/generation`.
- Safe modification: Pin behavior with tests before refactor; `managed_mode_test.go` (804 lines) is the safety net.

**No central exit-code mapping:**
- Files: `internal/api/runtime.go` (`errExit`), `internal/api/lint.go`, `breaking_check.go`, `ls_files_v1.go`, `cmd/easyp/main.go`
- Why fragile: Exit codes 0/1/2 are decided per handler branch; `.spec/ERRORS.md` warns that calling internal functions in tests cannot establish exit codes.
- Safe modification: When adding a command, mirror an existing handler's typed-error branch; never introduce a new sentinel merely to mirror a message.

## Scaling Limits

**Cache layout and lock contention:**
- Current capacity: v1 cache under `os.UserCacheDir` with per-object file locks (`object_lock_unix.go`/`object_lock_windows.go`) and object-lock wait loops honoring context cancellation.
- Limit: Parallel `easyp` invocations on the same cache serialize on object locks; full-history Git fallback (after failed shallow fetch) can be very slow on huge monorepos, bounded by the 5-minute Git timeout.
- Scaling path: `EASYP_GIT_TIMEOUT` tuning; immutable-tree index reuse (`git_object_sizes.go`); the download path verifies hashes of existing content (`download.go`) so cache growth is integrity-checked, but no cache pruning/GC exists.

## Dependencies at Risk

**Single-version hash policy, no compatibility namespace:**
- Risk: There is exactly one v1 hash policy; existing nightly locks whose hashes covered unrelated files are intentionally rejected. Users must back up `protobuf.lock` and rerun `easyp mod tidy`.
- Files: `V1_RELEASE_NOTES.md`, `internal/modules/immutable_versions.go`, `internal/adapters/gitmodules/cache_verification.go`
- Impact: Forced lock regeneration on CLI upgrade; immutable-hash guards fail until then.
- Migration plan: Documented manual backup + `easyp mod tidy`; no code migration planned (intentional).

## Missing Critical Features

**Breaking profiles beyond FILE/PACKAGE/WIRE_JSON/WIRE:**
- Problem: Only these four profiles are implemented; unknown categories fail shared validation, empty defaults to FILE.
- Files: `internal/core/breaking_profiles.go`, `.spec/config/breaking-profiles.md`
- Blocks: buf-compatible finer-grained category sets.

**No retry/backoff framework:**
- Problem: No shared retryable-error classification; only the shallow-fetch-then-full-history fallback in `object_cache.go`.
- Files: codebase-wide, noted in `.spec/ERRORS.md`

## Test Coverage Gaps

**Process exit codes:**
- What's not tested: Handler-level exit-code branches (1 vs 2) are only provable through `Action`; tests calling internal worker functions cannot establish them.
- Files: `internal/api/*`, warned in `.spec/ERRORS.md`
- Risk: Exit-code regressions (CI scripts keying on 1/2) land unnoticed.
- Priority: Medium

**RPC streaming breaking changes:**
- What's not tested: The unimplemented stream checks (see Tech Debt) have no tests by construction.
- Files: `internal/core/breaking_checker.go`, `breaking_checker_test.go`
- Risk: Silent false-negative breaking checks.
- Priority: High

**Crash-mid-migration recovery:**
- What's not tested: Interruption between staged replacements in `internal/migration/apply.go`.
- Risk: Half-migrated projects rely on manual recovery-directory restoration.
- Priority: Low

---

*Concerns audit: 2026-11-17*
