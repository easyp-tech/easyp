# BSR dependency resolution implementation plan

**Goal:** Resolve BSR metadata in Git dependencies through a replaceable backend and reuse the existing Git graph, cache and import pipeline.

**Architecture:** Buf readers preserve module identity, requested reference, lock commit/digest and the declaring config. A small `modules.BSRResolver` supplies an existing `v1.Requirement`; the Git adapter records this binding on its parent's lock entry and adds the requirement to its metadata. Frozen and cached operations replay recorded bindings instead of calling a backend. The current backend uses explicit fixed Git snapshots; its `compatibility_snapshot` policy states that BSR revision equivalence and BSR digest verification are not guaranteed. The user approved this policy on 2026-10-03.

**Tech stack:** Existing Go toolchain, yaml.v3, testify, Git cache and local-binary v1 regressions. No Buf API, Buf executable, HTTP client or new dependency.

## Scope and files

- `internal/config/v1/bsr.go`: BSR request and recorded binding values, validation. Reuse `Requirement` for the Git target.
- `internal/config/v1/module.go`, `lock.go`, `schema.go`: internal metadata and optional `bsr` bindings in locked Git entries. Regenerate lock schemas through the existing generator.
- `internal/modules/bsr.go`: the backend interface; dependency graph code remains BSR-agnostic.
- `internal/adapters/bsr/resolver.go`: explicit snapshot table and diagnostic unknown-module error. CLI logging explains the compatibility policy.
- `internal/adapters/module_config/git_dependency_buf*.go`: v1/v1beta1/v2 Buf references and optional locks, including workspace rebasing. Read all locked transitive BSR pins; reject invalid/stale pins and symlink metadata.
- `internal/adapters/gitmodules/bsr.go`, `cache.go`, `checkout.go`, `migration.go`: resolve bindings while fetching, restore exact bindings in Cached, expose backend injection without altering Git fetch/install APIs.
- `internal/api/mod_v1_cache.go`: backend composition point for the future Service implementation.
- `easyp-test/tests/e2e/v1/bsr_dependencies_test.go`: local Git regression fixtures. Git URL rewriting redirects fixed semantic snapshots to isolated local tagged providers; no runtime test configuration is added to EasyP.
- `easyp-test/tests/e2e/v1/live_dependencies_test.go`: grpc-gateway and grpc-federation must work without explicitly adding Googleapis. Preserve existing direct-provider coverage separately.

## Execution

- [x] Add local CLI regressions and verify RED on the current v1.0 executable: known mapping, unknown mapping, plain Git, workspace, pins/digests, several modules, deduplication, frozen replay and no writes on errors.
- [x] Add metadata/domain/backend tests before implementing each corresponding operation. Cover context cancellation, error identity, malformed/stale Buf locks, ambiguous/conflicting pins and invalid backend results.
- [x] Implement parsing, the explicit backend and fetch/cache bindings. Preserve parent provenance; deduplicate equivalent requirements through the existing graph. Frozen operations must work with a backend that always fails.
- [x] Add/regenerate optional lock schema fields. Validate recorded origins against the verified dependency config rather than trusting arbitrary bindings.
- [x] Update regressions whose old expectation intentionally ignored BSR metadata. Preserve manual Git-provider behavior with fixtures that declare no BSR identity; unknown BSR modules now fail explicitly.
- [x] Build the current development binary; run relevant package tests and the full v1 e2e suite with `-race`.
- [x] Run real old/current grpc-gateway and grpc-federation cases, including cold frozen download/generation/vendor, without manual Googleapis requirements.
- [x] Run full Go tests with `-race`, `lint:go`, schema check, vet and diff review. Document actual mappings, the replacement point and compatibility limitations.

## Verification results

- RED: the original development binary passed plain Git and failed the other 12 new CLI cases for the expected missing-resolution behavior.
- GREEN: the new local BSR cases pass, including cold frozen download, generation and vendor, provenance checks and no project writes on errors.
- EasyP: `go test -race -count=1 ./...` passed on the final source tree.
- EasyP-test: `go test -race -tags=v1 ./tests/e2e/v1 -count=1 -parallel=4` passed with `EASYP_BIN=/tmp/easyp-bsr-development` and `EASYP_SOURCE` pointing to this checkout. The first attempt omitted `EASYP_SOURCE` and failed only its three source-dependent checks; both configured full runs passed.
- Public matrix: `go test -tags='v1 live_dependencies' ./tests/e2e/v1 -run '^TestLiveGitDependencies$' -count=1 -parallel=3 -v` passed all 11 cases. grpc-gateway v2.26.1/v2.31.0 and grpc-federation v1.24.0/v1.27.0 acquired Googleapis automatically; cold frozen descriptors matched their initial bytes and vendor succeeded.
- `task lint:go` with the pinned 2.14.0 executable reported zero issues; `task schema:generate`, `task schema:check`, `go vet ./...` and `git diff --check` passed.
- Independent BSR code review found no required changes. Migration delegates to `moduleCache`, keeping backend composition centralized.

No commit or push is part of this request. Earlier unrelated working-tree changes stay intact.
