# Roadmap: EasyP

## Shipped Milestones

- [x] **v1.0: lint-easypc-infra** — `easyp lint` migrated to the vendored `easypc` compiler (7 phases, 32 plans, shipped 2026-11-21, PR #258) — archived: `.planning/milestones/v1.0-ROADMAP.md`
- [x] **v1.1: import-easyprotoc** — vendored `internal/easypc` replaced by the public `github.com/easyp-tech/easyprotoc` module; branch reconciled with `main` #253, seal infrastructure removed, v1.0 acceptance set re-proven (4 phases, 7 plans, shipped 2026-11-21, PR #258) — archived: `.planning/milestones/v1.1-ROADMAP.md`

## Backlog

### Phase 999.1: Follow-up — Phase 03 deferred UAT follow-up: Test 1 (BACKLOG)

**Goal:** Resolve the UAT checkpoint deferred during Phase 03 verification
**Source phase:** 03
**Follow-ups:**
- [ ] Test 1: Add a disk-instrumentation cell in Phase 4 asserting readSource is invoked exactly once per corpus file on a compiling corpus (behavioral proof for the no-disk-reread guarantee, currently structural) (deferred 2026-11-19)

### Phase 999.2: Follow-up — Phase 03 deferred UAT follow-up: Test 2 (BACKLOG)

**Goal:** Resolve the UAT checkpoint deferred during Phase 03 verification
**Source phase:** 03
**Follow-ups:**
- [ ] Test 2: Standing constraint for Phase 4: at each rule-port batch confirm ported rules never consult comment directives or ignoreOnly themselves; keep central-filter cells green (deferred 2026-11-19)

### Corpora letters (LINT-PARITY-01 / LINT-PERF-01, from Phase 7)

- [ ] Unblock private corpora (GOOGLEAPIS fresh clone per Pitfall 7 + CORPUS_ARGS per `07-USER-SETUP.md`), fire `/tmp/p7/run-parity.sh parity` + perf protocol; verdict `median(new) <= median(main)` (deferred 2026-11-21, pending_user_unblock)

### Hadolint

- [ ] Run `task lint:docker` once Docker is available (Dockerfile untouched by v1.0) (deferred 2026-11-21)

### easyprotoc version pin

- [ ] Bump `github.com/easyp-tech/easyprotoc` when upstream publishes tags; re-evaluate the unmerged `feat/easyp-integration` branch (generation sources/descriptor metadata fix, changes `Link` error semantics — needed for the codegen series) (deferred 2026-11-21, v2)
