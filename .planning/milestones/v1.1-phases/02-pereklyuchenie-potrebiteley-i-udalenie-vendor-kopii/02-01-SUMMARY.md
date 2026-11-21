---
phase: 02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii
plan: 01
subsystem: compiler-integration
tags: [easyprotoc, protobuf, protoc, import-rewrite, vendor-removal, go-modules, protoast]

# Dependency graph
requires:
  - phase: 01-merge-main-i-primirenie-putey
    provides: "origin/main #253 merged: go.mod already pins github.com/easyp-tech/easyprotoc directly at v0.0.0-20261104120806-f378f3f3a05f and google.golang.org/protobuf v1.37.0-rc.1, and internal/core/breaking_profiles.go already imports easyprotoc/protoc"
provides:
  - "All compiler consumers (lint linker, its tests, the AST-to-lint-view builder and outline walk) import github.com/easyp-tech/easyprotoc/{protoc,protoc/syntax,parser} instead of vendored internal/easypc paths"
  - "internal/easypc removed from the tracked tree (648 files: 48 .go, 555 .proto, 30 .tokens, 14 .golden, VENDORED.md) including cmd/easyp-protoc and cmd/lexbench"
  - "One-source-of-truth compiler version: the go.mod pin, verified tidy-idempotent with zero replace/vendor"
  - "Proof the seven-file edit is purely mechanical (14 removed / 14 added path lines; stripped-content diff byte-identical)"
affects: [02-02-full-battery-and-parity, 03-seal-infra-and-docs]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
# Scale: chars/4 over the realized authored diff (the 15 inserted lines of the seven-file edit).
# The 22,907 deleted lines are mechanical vendor volume, not authored content; the plan's
# estimate breakdown also priced the deletion as a command, not as content.
actuals:
  tokens: 228
  tasks: 2
  commits: 1

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Versioned public-module dependency as the single compiler source; no vendor copy, no replace"
    - "Scripted path sweep with per-fragment occurrence assertions before each substitution"
    - "Non-behavioral edit proven by a strip gate (remove comments/blanks/import paths, require byte-identical content)"

key-files:
  created: []
  modified:
    - internal/core/link.go
    - internal/core/link_test.go
    - internal/core/lint_linked_matrix_test.go
    - internal/protoast/build.go
    - internal/protoast/outline.go
    - internal/protoast/protoast.go
    - internal/protoast/protoast_test.go

key-decisions:
  - "Tidy is a verified no-op: go.mod/go.sum are byte-identical to PRE_SWAP_BASE, so the pin was already the only source (D-06 confirmed empirically, not assumed)"
  - "Import regrouping places easyprotoc in the third-party group (own group in link.go; alphabetically above testify in the three test files) mirroring main's internal/core/breaking_profiles.go"
  - "The two VENDORED.md comments were rewritten in place (no new lines) so the seven-file diff stays exactly 14 removed / 14 added"
  - "No replacement local cmd binary: upstream cmd/easyprotoc builds from the module (proven with go build -o /dev/null)"

patterns-established:
  - "Non-behavioral migration gate: per-file strip diff must be silent between PRE_SWAP_BASE and SWAP_SHA"
  - "Atomic swap+delete: a single commit whose changeset is exactly 648 D + 7 M + 0 A and whose build is green"

requirements-completed: [IMPORT-CONS-01, IMPORT-VENDRM-01, IMPORT-DEP-01]

coverage:
  - id: D1
    description: "Eight vendored import lines in six consumer files rewritten to the public easyprotoc packages; zero aliases, shims, or intermediate packages"
    requirement: IMPORT-CONS-01
    verification:
      - kind: other
        ref: "git grep -n 'internal/easypc' -- '*.go'  (empty); 9 easyprotoc import lines total, 8 of them in the six swapped files; go build ./... exit 0"
        status: pass
    human_judgment: false
  - id: D2
    description: "internal/easypc deleted in the same atomic commit: all 648 tracked files including VENDORED.md and both cmd binaries"
    requirement: IMPORT-VENDRM-01
    verification:
      - kind: other
        ref: "git ls-files internal/easypc (empty); git ls-tree -r SWAP_SHA -- internal/easypc (empty); git cat-file -e SWAP_SHA:internal/easypc fails; per-path absence checks for VENDORED.md, cmd/easyp-protoc/main.go, cmd/lexbench/main.go"
        status: pass
    human_judgment: false
  - id: D3
    description: "Dependency resolution unchanged: exact easyprotoc and protobuf pins, byte-identical go.mod/go.sum, empty tidy diff, zero replace/vendor"
    requirement: IMPORT-DEP-01
    verification:
      - kind: other
        ref: "go list -m -f '{{.Version}}' github.com/easyp-tech/easyprotoc == v0.0.0-20261104120806-f378f3f3a05f; google.golang.org/protobuf == v1.37.0-rc.1; go mod tidy -diff empty; git diff PRE_SWAP_BASE SWAP_SHA -- go.mod go.sum empty"
        status: pass
    human_judgment: false
  - id: D4
    description: "protoc-compatible binary remains reachable without a local copy: upstream cmd/easyprotoc builds from the module"
    requirement: IMPORT-VENDRM-01
    verification:
      - kind: other
        ref: "go build -o /dev/null github.com/easyp-tech/easyprotoc/cmd/easyprotoc (exit 0); go list -f '{{.Name}}' reports main"
        status: pass
    human_judgment: false
  - id: D5
    description: "The seven-file edit is provably mechanical: 14 removed lines all naming internal/easypc, 14 added non-blank lines all naming easyprotoc, stripped content byte-identical"
    requirement: IMPORT-CONS-01
    verification:
      - kind: other
        ref: "git diff -U0 PRE_SWAP_BASE SWAP_SHA -- <seven files> awk gate: removed=14 added=14 bad=0; per-file strip diff (comments/blanks/import paths removed) silent"
        status: pass
    human_judgment: false

# Metrics
duration: 4min
completed: 2026-11-21
status: complete
commits: 1
plan_head_before: df23d64aa107f17024f7cfcd3bb6198fe92f7a91
plan_head_after: 5342455410a50ae30f1f4f5eeb1f49f7ac20eb4f
---

# Phase 2 Plan 1: Consume the public easyprotoc module and drop the vendored easypc tree

**Eight vendored compiler imports in six files rewritten to `github.com/easyp-tech/easyprotoc`, six stale comments repaired, and the entire 648-file `internal/easypc` tree — VENDORED.md and both cmd binaries included — deleted in one atomic `feat` commit whose build, vet, gofmt, tidy and quick `-race` tests are green.**

## Performance

- **Duration:** 4 min
- **Started:** 2026-11-21T13:03:07Z
- **Completed:** 2026-11-21T13:07:13Z
- **Tasks:** 2 (Task 1 = tracer implement+commit; Task 2 = post-commit invariant evidence, evidence-only, no separate code commit)
- **Files modified:** 7 (+648 deleted)

## Accomplishments

- Exactly three import-path substitutions, no aliases and no shim packages: `.../easypc/protoc` -> `easyprotoc/protoc` (6 lines, covering the `/protoc/syntax` imports), `.../easypc/pkg/parser` -> `easyprotoc/parser` (2 lines). Post-state import census: 9 `easyprotoc/` import lines total, 8 of them in the six swapped consumer files plus the pre-existing one in `internal/core/breaking_profiles.go`; per-file counts 1/1/1/2/2/0/1 for the D-02 seven.
- Six comment lines that named deleted paths rewritten: three `easypc/protoc/resolver.go` references, one `easypc/pkg/lexer/token.go:121-123` reference, and the two `internal/easypc/VENDORED.md` sync pointers (now `build.go`'s `pos()` divergence text points at the go.mod easyprotoc pin, and `protoast.go`'s package doc points at the builder plus the version pin).
- The whole `internal/easypc` tree deleted with `git rm -r`: 48 `.go`, 555 `.proto`, 30 `.tokens`, 14 `.golden`, 1 `.md` (`VENDORED.md`) = 648 files, including `cmd/easyp-protoc/main.go` and `cmd/lexbench/main.go`. Nothing was copied out and no replacement local binary was created.
- `go mod tidy` ran exactly once and changed nothing: go.mod/go.sum hashes identical before/after, `go mod tidy -diff` prints nothing. easyprotoc remains a direct exact pin, protobuf resolves to `v1.37.0-rc.1`, zero `replace`/`vendor` lines.
- Non-behavioral edit proven, not asserted: the seven-file diff removes exactly 14 lines (each containing `internal/easypc`) and adds exactly 14 non-blank lines (each containing `easyprotoc`), and stripping comments, blank lines and old/new import-path lines from both sides leaves byte-identical content for all seven files. `internal/protoast` logic — including the rune-column `pos()` computation and every fixture expectation — is untouched (D-05).
- Post-state package graph matches the measured prediction exactly: 37 packages, 34 with test files, 3 without.

## Task Commits

Each task was committed atomically:

1. **Task 1: Swap the 14 vendored path references, delete internal/easypc, and land one atomic feat commit** - `5342455` (feat)
2. **Task 2: Post-commit invariant evidence — path absence, deletion completeness, diff shape, pins, upstream replacement binary** - evidence-only; no code commit (all gates below, recorded in this SUMMARY). Precedent: Phase 01 plan 01-02.

**Plan metadata:** docs commit carrying this SUMMARY, STATE.md and ROADMAP.md (see the completion message for its hash).

_Note: the plan mandates exactly one `feat` commit; Task 2 re-derives `SWAP_SHA` by subject, never from a bare HEAD assumption._

## Commit under test

| Item | Value |
|---|---|
| Subject | `feat(lint): consume public easyprotoc module; drop vendored easypc tree` |
| `SWAP_SHA` | `5342455410a50ae30f1f4f5eeb1f49f7ac20eb4f` |
| `PRE_SWAP_BASE` | `df23d64aa107f17024f7cfcd3bb6198fe92f7a91` |
| Branch | `feat/lint-easypc-infra` (no push, D-08) |
| Changeset | 655 paths: 648 `D` + 7 `M` + 0 `A` |
| Diffstat | 15 insertions(+), 22921 deletions(-) |
| Rollback | `git revert 5342455` (single commit, branch unpushed) |

## Gate Evidence

Every command ran on the committed tree; exit codes are as observed.

| # | Gate | Command | Expected | Observed | Exit |
|---|---|---|---|---|---|
| 1 | Pre-flight tracked files | `git ls-files internal/easypc \| wc -l` | 648 | 648 | 0 |
| 2 | Pre-flight vendored refs (scoped) | `git grep -n 'internal/easypc' -- '*.go' ':(exclude)internal/easypc' \| wc -l` | 14 | 14 | 0 |
| 3 | Pre-flight module import lines | `git grep -h -E '^[[:space:]]*"github\.com/easyp-tech/easyprotoc/' -- '*.go' \| wc -l` | 1 | 1 (breaking_profiles.go) | 0 |
| 4 | Fragment census before sweep | `grep -o <fragment> <seven files> \| wc -l` | 6 / 2 / 6 / 1 / 2 | 6 / 2 / 6 / 1 / 2 | 0 |
| 5 | Comment fragment after import swap | `grep -o 'internal/easypc/protoc/' <seven files> \| wc -l` | 3 | 3 | 0 |
| 6 | gofmt | `gofmt -l <seven files>` | empty | empty | 0 |
| 7 | Path gate (scoped) | `git grep -n 'internal/easypc' -- '*.go' ':(exclude)internal/easypc'` | empty | empty | 0 |
| 8 | Module import census (post) | `git grep -n -E '^[[:space:]]*"github\.com/easyp-tech/easyprotoc/' -- '*.go' ':(exclude)internal/easypc'` | 9 total (8 swapped + 1 pre-existing) | 9 total | 0 |
| 9 | Per-file import counts | `grep -c <module prefix> <file>` | 1/1/1/2/2/0/1 | 1/1/1/2/2/0/1 | 0 |
| 10 | Diff shape | `git diff -U0 PRE_SWAP_BASE SWAP_SHA -- <seven files> \| awk ...` | removed=14 added=14 bad=0 | removed=14 added=14 bad=0 | 0 |
| 11 | Strip gate | per-file `diff` of comment/blank/import-path-stripped content | silent | silent for all 7 | 0 |
| 12 | Delete tree | `git rm -r internal/easypc` | 648 deletions staged | 648 `D` | 0 |
| 13 | Tidy run | `go mod tidy` | no drift | `sha256sum -c` OK for go.mod and go.sum | 0 |
| 14 | Tidy idempotence | `go mod tidy -diff` | empty | empty | 0 |
| 15 | Build | `go build ./...` | 0 | 0 | 0 |
| 16 | Vet | `go vet ./...` | 0 | 0 | 0 |
| 17 | Quick consumer tests | `go test -race -count=1 ./internal/core ./internal/protoast ./internal/rules` | all ok | core 1.391s / protoast 1.018s / rules 1.088s, all `ok` | 0 |
| 18 | Staged changeset | `git diff --cached --name-status` | 648 D + 7 M + 0 A | 648 D + 7 M + 0 A | 0 |
| 19 | Staged modified set | `git diff --cached --name-status --diff-filter=M` | exactly the D-02 seven | exactly the D-02 seven | 0 |
| 20 | Forbidden paths in commit | `git diff --name-only PRE_SWAP_BASE SWAP_SHA \| grep -E '^(\.planning/\|\.golangci\.yml\|AGENTS\.md\|\.spec/)'` | none | NONE | 1 (no match) |
| 21 | Working tree after commit | `git status --porcelain --untracked-files=no` | empty | empty | 0 |
| 22 | Task 1 verify block | plan `<verify>` (subject, single commit, 655 paths, tree absent, clean, build) | PASS | PASS | 0 |
| 23 | Task 2 block A | path absence + 9/8 imports + tree/cmd absence | PASS | PASS | 0 |
| 24 | Task 2 block B | diff shape + strip gate | PASS | `removed=14 added=14 bad=0`, strip silent | 0 |
| 25 | Pins | `go list -m -f '{{.Version}}' ...` | `v0.0.0-20261104120806-f378f3f3a05f` / `v1.37.0-rc.1` | exact match | 0 |
| 26 | go.mod/go.sum vs base | `git diff --name-only PRE_SWAP_BASE SWAP_SHA -- go.mod go.sum` | empty | empty | 0 |
| 27 | Escape hatches | `grep -c '^replace' go.mod` / `'^vendor'` | 0 / 0 | 0 / 0 | 0 |
| 28 | Upstream replacement binary | `go build -o /dev/null github.com/easyp-tech/easyprotoc/cmd/easyprotoc` | 0 | 0 | 0 |
| 29 | Package graph | `go list ./... \| wc -l` | 37 | 37 | 0 |
| 30 | Test-package count | `go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./... \| grep -n . \| wc -l` | 34 | 34 | 0 |
| 31 | Re-run build/vet/tests/gofmt on commit | `go build ./...`, `go vet ./...`, quick `-race`, `gofmt -l` | all clean | all clean | 0 |

## Files Created/Modified

- `internal/core/link.go` - `protoc` import swapped to the public module and moved into its own third-party group above the project imports; resolver comment repointed at `easyprotoc/protoc/resolver.go`
- `internal/core/link_test.go` - `protoc` import swapped and placed at the top of the testify third-party group; LineCol comment repointed at `easyprotoc/lexer/token.go:121-123`
- `internal/core/lint_linked_matrix_test.go` - same import regroup; ctx-guard comment repointed at `easyprotoc/protoc/resolver.go`
- `internal/protoast/build.go` - two imports swapped (`easyprotoc/parser`, `easyprotoc/protoc/syntax`); `pos()` divergence comment now says to re-apply the divergence when the go.mod easyprotoc version is bumped
- `internal/protoast/outline.go` - the same two imports swapped; no other change
- `internal/protoast/protoast.go` - package doc references the public `protoc/syntax` path and points the sync note at `build.go` + the go.mod pin; no imports change
- `internal/protoast/protoast_test.go` - `protoc/syntax` import swapped and placed above the testify imports
- `internal/easypc/**` - deleted in full: `VENDORED.md`, `cmd/easyp-protoc/main.go`, `cmd/lexbench/main.go`, 45 other `.go` files, 555 `.proto`, 30 `.tokens`, 14 `.golden`
- `go.mod` / `go.sum` - **not modified** (tidy no-op; zero-diff versus `PRE_SWAP_BASE`)

## Decisions Made

- **Tidy was proven rather than assumed.** D-06 predicted no drift; the measurement confirms it — the swap does not reach any old dependency because `internal/core/breaking_profiles.go` already kept `easyprotoc/protoc` a direct requirement. No `go.mod` edit was needed and none was made.
- **Import placement mirrors main's `internal/core/breaking_profiles.go`.** `easyprotoc` is third-party, `github.com/easyp-tech/easyp/...` is project, so the two groups are separated; in the three test files the module import sorts above `github.com/stretchr/testify/assert`. In `build.go`/`outline.go` the two module imports already occupied the non-stdlib group, so no move was needed.
- **The two `VENDORED.md` comments were rewritten in place**, preserving line count. This keeps the seven-file diff at exactly 14 removed / 14 added path-bearing lines, which is what makes the non-behavioral claim mechanically checkable. Wording follows the CONTEXT discretion grant for `protoast.go`.
- **The vendored copy's provenance is now the go.mod pin only.** No archive, no port, no re-export: `git revert` is the rollback and a future easyprotoc bump must re-verify the API mapping (plan 02-02 re-proves behavior).

## Deviations from Plan

None - plan executed exactly as written. No deviation rule (1-4) was triggered: the sweep, deletion, tidy, build, vet, quick tests and every gate behaved exactly as the plan predicted, including the rehearsed 648 D + 7 M + 0 A changeset and the 37/34/3 package graph.

## Issues Encountered

1. **Stale plan commit ledger from an earlier milestone.** `.git/gsd-plan-head-before-02-01` already existed (dated Nov 19) from a previous milestone's phase 02 plan 01 and pointed at an unrelated commit (`b5c26df`). The protocol's `[ -f ] || write` guard therefore kept the stale base, which would have made the SUMMARY's measured `commits:` count wrong. Resolved by overwriting the ledger with the true `PRE_SWAP_BASE` (`df23d64`) before committing; the measured count is now `git rev-list --count df23d64..5342455` = 1.
2. **Shell abort during the scripted sweep (no code impact).** The first sweep invocation used `set -euo pipefail` together with `grep -o <fragment> | wc -l` post-check; once a fragment reached zero occurrences `grep` exits 1, `pipefail` propagated it and the script stopped after step 1 (the `parser` substitution, which had applied). Steps 2-4 were re-run with a counting helper that tolerates a no-match; every step re-asserted its pre-substitution occurrence count first, so no fragment was applied twice or missed. Final census verifies the exact expected counts (0 residual vendored refs, 9 module imports).
3. **`task` and `docker` are absent in this environment.** Every gate used `go`, `git`, and shell directly (Phase 1 precedent). golangci-lint is not a Phase 2 gate per D-07, and the stale `.golangci.yml` exclusion of the now-deleted path is intentionally left for Phase 3.

## Known Stubs

None. No placeholder values, TODOs, unwired components, or skipped tests were introduced. The deleted `internal/easypc` tree carried no interface that this plan stubbed out: its consumers now resolve the same packages through the module, proven by a green build/vet/test run on the commit.

## Threat Flags

None. No new network endpoint, auth path, file-access pattern, or trust-boundary schema change was introduced. The one trust boundary this plan moves — compiler source entering the build from the Go module cache — is covered by the existing exact pseudo-version pin plus `go.sum`, verified unchanged (T-02-14).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- **Atomic slice is green before expansion**, exactly as the tracer gate requires: `5342455` builds on its own, so any failure in plan 02-02's full battery or parity gates is a swap error rather than a new norm.
- **Plan 02-02 can start immediately.** It owns the full `-race` battery, golden 94, Gates A-D x2, the D-07 position gates, the core/CLI matrix, and re-proving these invariants by named test.
- **Deferred by design to Phase 3 (D-07):** `.golangci.yml`'s exclusion of `internal/easypc`, the compiler-lab/VENDORED.md prose in `AGENTS.md`, and `.spec/**`. A nonexistent exclusion path cannot fail a Phase 2 gate (golangci-lint is not one), and T-02-18 accepts this for one phase.
- **Not pushed (D-08):** the commit sits on `feat/lint-easypc-infra`, extending PR #258.
- **Known non-blocking residue:** `internal/rules/package_no_import_cycle_test.go:28` still says "the easypc linker" in prose (no path), which is outside D-02's exact file set and outside the D-09 grep — flagged at planning time and left untouched.

---
*Phase: 02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii*
*Completed: 2026-11-21*

## Self-Check: PASSED

- SUMMARY.md exists at the plan path
- Commit 5342455 is an ancestor of HEAD
- All seven consumer files exist; `internal/easypc` is absent
