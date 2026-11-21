# Milestones

## v1.1 import-easyprotoc (Shipped: 2026-11-21)

**Phases completed:** 4 phases, 7 plans, 15 tasks

**Key accomplishments:**
- Merged origin/main #253 onto feat/lint-easypc-infra with blob-exact resolution: main's easyprotoc breaking path, the branch's protoast Position model, a union core.go, and a byte-untouched lint contour.
- Committed merge 994c503 proven behavior-preserving: a 41-package `-race` battery green with 0 FAIL and 0 races, plus Gates A–D ×2, golden 94, both D-07 position gates and the core/CLI parity matrix all passing with expectations byte-frozen against the pre-merge base.
- The public-module swap proven behavior-preserving on commit 5342455: one 34-package `-race` battery green with 0 FAIL, 0 races and one known environment skip, plus Gates A-D x2 against baseline 1e9a596, golden 94 byte-exact, both D-07 position gates, the 43-line core matrix and the 16-line CLI matrix all passing with expectations provably unmoved.
- The internal/easypc lint exclusion, the AGENTS.md sealed-vendor prohibition and three stale compiler citations were removed in one seven-file docs commit, while `.spec` now documents the pinned `github.com/easyp-tech/easyprotoc` module and the `internal/protoast` D-07 rune-column seam.
- The last pre-existing lint finding was fixed at its source — `internal/sourceview/session_test.go` now reports `Session.Close()` failures through `t.Cleanup` + `t.Errorf` — taking `bin/golangci-lint run ./...` to exactly `0 issues.` and proving the phase's build, scoped-race, tidy, seal-grep and eight-file changeset gates on the committed tree.
- The v1.0 acceptance set passes unchanged on the public `easyprotoc` module: one 34-package `-race` battery green (0 FAIL, 0 races, one known environment skip), all eleven named parity gates proven by exact RUN/PASS (Gates A–D ×2 against `1e9a596`, golden 94 byte-exact, both D-07 position gates, the 43-line core and 16-line CLI matrices), exactly `0 issues.` lint, zero schema drift, exact pins, an empty live-surface seal grep, and a zero non-`.planning` diff from `6f52840`.

---

**Closeout notes:** verified_closeout with one recorded override — Phase 1/2 `verification.status` read `stale` at close because Phase 3 edited covered files (`internal/sourceview/session_test.go` test-only errcheck fix; `internal/core/link_test.go` and `internal/protoast/build.go` comment-only, proven by the Phase-3 comment-only diff gate). The behavioral acceptance was re-proven on the final tree by Phase 4 (`IMPORT-PARITY-01`: full `-race` battery, Gates A–D ×2, golden 94, D-07, lint `0 issues.`, zero schema drift); no behavioral coverage was lost. Known environment skip: `TestGenerateGoHeaderUsesCompilerVersion` (protoc-gen-go absent). Branch not pushed (push/PR is the milestone-close UAT gate).

**Post-close fixes (2026-11-21, same branch):** (1) upstream `easyprotoc` 32-bit lexer overflow (`math.MaxUint32` vs 32-bit `int`) broke the CI `linux/arm7` cross-build — pre-existing on the branch, not caused by v1.1; fixed upstream as `easyprotoc` PR #2 (`uint64(len(src))`) and the pin was bumped to `v0.0.0-20261121194146-14f27ad9ce15` (re-pin to upstream main after merge). (2) `TestAcceptanceGates` builds the v1.0 baseline from `1e9a596` and failed in the depth-1 CI checkout; the workflow now fetches that commit explicitly by full SHA. CI `tests` green on `f3f7fac`; PR #258 is MERGEABLE and blocked only on the required human review.
