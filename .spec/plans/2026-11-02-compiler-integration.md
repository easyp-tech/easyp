# Native Compiler Integration Plan

**Goal:** Integrate `github.com/easyp-tech/easyp-test-compiler` as a separate Go dependency in EasyP and pass the full v1 regression suite. Do not copy or vendor compiler implementation into EasyP. Work in `feat/compiler-integration`; necessary compiler fixes belong to its separate `feat/easyp-integration` branch.

**Architecture:** One EasyP `internal/compiler` adapter invokes the native public library with caller-authoritative source resolution and complete descriptors. Replace all four true compilation sites, retaining the independent syntax/import parsers used for lint and source-preserving edits. Keep no implicit old-compiler fallback. Use a temporary external Go workspace during two-repository development; final dependency metadata must identify the tested compiler revision without absolute local replacements.

**Contract:** Keep logical import identities, ordered roots/targets, source boundaries/aliases, proposed in-memory sources, deterministic custom options and output, source locations, managed mode, independent consumer graphs, cancellation, and contextual errors. Standard imports are explicit resolver composition. Native compilation may never bypass a resolver denial by opening the filesystem.

## 1. Dependency and shared adapter (orchestrator)

- [x] Resolve/pin the compiler Go module using command-local private/SSH access, preserving standalone module layout.
- [x] Add behavior tests first for authorized virtual inputs, dependency/target order, denied/missing imports and no fallback, descriptor errors, custom extensions and source locations, cancellation and repeated deterministic output.
- [x] Implement `internal/compiler.Compile(ctx, targets, Options)` returning target-ordered `[]protoreflect.FileDescriptor`, with `Options{Resolve, SourceInfo, MaxParallelism}` and an ordered `SourceResolver(importPaths, accessor, standardImports)` helper.
- [x] Collect native diagnostics as returned errors, avoid stderr effects, check descriptor construction errors and preserve interpreted custom options across serialization/cloning.

## 2. Native public API (compiler repository)

- [x] Write failing tests for resolver permission bypass and missing errors, checked complete reflection and interpreted custom options.
- [x] Add `protoc.WithResolverOnly(resolver)`: use only supplied resolvers (plus explicitly requested standard imports); no directory/direct filesystem fallback. Fatal resolver errors must survive with `errors.Is`.
- [x] Add `Result.FilesWithError() ([]protoreflect.FileDescriptor,error)` without breaking existing Files API; construct the full registry and dynamically hydrate custom option extensions. Adapter always requests includeImports.
- [x] Preserve context and deterministic diagnostic behavior; cover valid default API behavior in existing tests.

## 3. Native semantic compatibility (compiler repository)

- [x] Add focused red tests for found blockers: invalid or unknown custom options, malformed aggregates/types/oneofs, Editions features/source locations, explicit empty defaults, nested extensions and unresolved RPC types.
- [x] Repair parser/linker/options/serialization/sourceinfo behavior necessary for current EasyP contract and supported protobuf inputs; reject invalid input rather than silently dropping data.
- [x] Keep bounded worker scheduling and deterministic descriptors; run package/race tests and parity fixtures. Do not implement an old-compiler fallback.

## 4. EasyP consumers (integration worker)

- [x] Route `prepareGeneration` and `compileBreakingGraph` through the shared adapter; accept reflection slices and use standard descriptor conversion without losing options.
- [x] Route `tidySourceView.compileChanges` and `proveMigrationImportBindings` through authoritative resolver callbacks, preserving per-target closures and migration serial behavior.
- [x] Update compiler identity/build metadata and compiler-sensitive tests to identify the native compiler truthfully. Update the one external v1 header assertion if required, preserving all preexisting easyp-test edits.
- [x] Migrate plugin deterministic-request compiler fixture; preserve independent lint/import syntax parsers.

## 5. Verification and delivery

- [x] Run independent spec and code quality reviews of adapter, source safety, native fixes and consumer wiring.
- [x] Run full Go suites with `-race` for both repositories, pinned EasyP Go lint, formatting and diff checks; fix substantive failures and recheck.
- [x] Build fresh CLI/MCP from final sources; record both repositories' revisions/patch fingerprints and dependency mode.
- [x] Run `easyp-test` full `-tags=v1 -race` regress with fresh binaries; inspect every failure and preserve unrelated local changes.
- [x] Finalize a reproducible separate dependency pin and ensure tested code equals that dependency; document any expected compiler-header changes and actual limits. No source copying, no global credentials/config changes.


Compiler fixes are published on `feat/easyp-integration` as `36db74fdd1eb73d52a93f29d88371d52c937d7c4`; EasyP pins `v0.0.0-20261101222746-36db74fdd1eb`. The compiler repository remains a separate dependency; final verification uses `GOWORK=off`.

## 6. Producer Buf source exclusions (corrected scope)

Do not add an excludes directive to protobuf.mod. Its existing grammar remains
unchanged. Source filters come automatically from the dependency's own Buf
metadata (v1/v1beta1 build.excludes, v2 modules[].excludes and v1 workspaces).
A repository without buf.yaml, including googleapis, is valid and gets no
invented exclusions. New consumer-level exclusion controls are out of scope.

- [x] Remove the rejected native directive, MCP/schema-reference additions, consumer-only scaffolding and the synthetic googleapis native-excludes script.
- [x] Verify real Buf-only producer metadata through Fetch/install/cold cache, mod tidy and actual native compiler descriptor generation; retain conflicting/malformed excluded-source, alias and absent-metadata controls.
- [x] Run affected race checks, Go lint and fresh standalone v1 regressions; document evidence with exact code/binary identity.

The prior benchmark in workspace compiler-bench/RESULTS.md is explicitly a
historical result for integration-49f78b316371. Its synthetic native directive
fixture does not establish the current Buf metadata contract and is retired.

Final Buf-only verification: 3528 source tests/subtests with race detector,
0 lint issues, 18 explicit CLI checks (including cold frozen install), and
976 v1 regress tests/subtests / 210 top-level / no failures or skips.
Evidence: workspace compiler-bench/BUF-EXCLUDES.md.
