---
phase: 02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii
status: complete
reviewed: 2026-11-21T13:35:00Z
depth: deep
commit_reviewed: 5342455410a50ae30f1f4f5eeb1f49f7ac20eb4f
base_commit: df23d64aa107f17024f7cfcd3bb6198fe92f7a91
files_reviewed: 7
deleted_files: 648
blocking: 0
warnings: 0
info: 3
findings:
  critical: 0
  warning: 0
  info: 3
  total: 3
  necessity: 0
files_reviewed_list:
  - internal/core/link.go
  - internal/core/link_test.go
  - internal/core/lint_linked_matrix_test.go
  - internal/protoast/build.go
  - internal/protoast/outline.go
  - internal/protoast/protoast.go
  - internal/protoast/protoast_test.go
---

# Phase 2: Code Review Report

**Reviewed:** 2026-11-21T13:35:00Z
**Depth:** deep
**Commit Reviewed:** `5342455410a50ae30f1f4f5eeb1f49f7ac20eb4f` (parent `df23d64aa107f17024f7cfcd3bb6198fe92f7a91`)
**Files Reviewed:** 7 changed (`internal/core` ×3, `internal/protoast` ×4) + 648 deletions under `internal/easypc/**`
**Status:** complete

## Summary

The changeset is a pure dependency-source swap: eight vendored compiler import lines move to the public `github.com/easyp-tech/easyprotoc` packages, six comments that named deleted paths are repointed, and the 648-file `internal/easypc` tree is deleted in the same atomic commit. I independently re-derived the commit shape, re-ran the strip gate, rebuilt and vetted the tree, compared every deleted `.go` file against the pinned module snapshot, and re-ran the affected unit/golden gates.

The core claims hold:

- **Diff shape is mechanical.** 14 removed and 14 added path-bearing lines; the only other edits are one removed blank line (`internal/protoast/protoast_test.go`) and one added blank line (`internal/core/link.go`) introduced by the import regroup. Per-file strip diff (comments, blank lines, old/new import paths removed) is silent for all seven files, so no logic or fixture changed.
- **No vendored path survives in Go source.** `git grep 'internal/easypc' -- '*.go'` is empty; the tree is absent from the index, from the swap commit, and from the worktree; both deleted `cmd` mains are gone and not referenced by Taskfile, Dockerfile, CI, scripts, or tests.
- **The seam is intact.** `pos()`, `outline.go`, `protoast.go`, and every fixture in `protoast_test.go` are byte-identical apart from the listed import/comment lines. No aliases, shims, `replace`, or `vendor` were introduced.
- **Dependencies are frozen.** `go.mod`/`go.sum` are unchanged versus the base, `easyprotoc` remains a direct exact pin, `go mod tidy -diff` is empty, and `go mod verify` reports all modules verified.
- **Behavior is preserved on the checks I re-ran.** `go build ./...`, `go vet ./...`, `go test -race -count=1 ./internal/protoast ./internal/core`, and `TestLintGoldenFullRegistry` (94 findings) are green. The full `-race` battery, Gates A–D, and the CLI matrix were not re-run per the review instruction; `02-02-SUMMARY.md` carries that evidence.

No BLOCKER and no WARNING findings. Three INFO findings are documentation-drift items around the deleted compiler's old name/layout; none affects behavior, and two of them are pre-declared Phase 3 or accepted residue.

## Findings

| ID | Severity | File:Line | Finding | Evidence | Recommendation |
|---|---|---|---|---|---|
| IN-01 | INFO | `internal/core/link_test.go:288` | The recursive-import case cites a bare `driver.go:319-323`; after this commit no `driver.go` exists in the repository, so the citation is unresolvable locally. | `git ls-files` has no `driver.go`; the cited literal `Line: 2, Col: 1` now lives at `github.com/easyp-tech/easyprotoc/internal/compiler/driver.go:319-323` in the module cache at the go.mod pin. | Qualify the path with the module prefix (and ideally the pin), or drop the line numbers and name the diagnostic. |
| IN-02 | INFO | `internal/protoast/build.go:449-452` | The re-sync note was shortened to "re-apply the divergence" and no longer names the three surfaces that encode it or the two D-07 gates; the deleted `VENDORED.md` carried that inventory. | `git show df23d64:internal/easypc/VENDORED.md` §"protoast divergence (re-apply on every sync)" lists `build.go` `pos()`, `protoast.go` wording, `protoast_test.go` tab expectations, and the D-07 gates; the new comment names only the divergence. Mitigated: `internal/protoast/position_test.go:13-21` and `internal/core/position_parity_test.go:16-21` independently document the contract. | Extend the comment to name the three surfaces and `TestPositionRuneColumns` / `TestProtoastPositionsMatchLegacyContract`. |
| IN-03 | INFO | `internal/rules/package_no_import_cycle_test.go:28` | Prose still says "The easypc linker hard-rejects…", naming a component that no longer exists in-tree; it is the last in-repo prose use of the old compiler name. | `git grep -n 'easypc'` outside `.planning`/`.spec` returns only this line plus `.golangci.yml` and `AGENTS.md` (accepted Phase 3 seal items); the path grep over `*.go` is empty. | Reword to "the easyprotoc linker" or "the compiler's linker" when Phase 3 updates the seal docs. Pre-declared as known residue in `02-01-SUMMARY.md`. |

## Narrative Findings (AI reviewer)

### IN-01: Dangling `driver.go` line citation

**File:** `internal/core/link_test.go:288`
**Category:** code quality (documentation accuracy)

The `import_cycle` case explains its `Line: 2, Col: 1` assertion with:

```go
// The driver hard-codes the recursive-import diagnostic at
// Line 2, Col 1 (driver.go:319-323); do not "fix" the 2:1.
```

Before this phase the only `driver.go` was `internal/easypc/pkg/compiler/driver.go`. This commit deletes it, and no `driver.go` remains anywhere in the repository (`git ls-files` returns nothing for `driver.go`). The exact lines still exist in the pinned module at `github.com/easyp-tech/easyprotoc/internal/compiler/driver.go:319-323` and still contain `Line: 2, Col: 1` — the numbers coincidentally survived the `pkg/` → `internal/` reorganization, so the comment is not currently *wrong*, only unverifiable from the repo and fragile to a pin bump.

**Fix:** `// Line 2, Col 1 (github.com/easyp-tech/easyprotoc/internal/compiler/driver.go:319-323 at the go.mod pin); do not "fix" the 2:1.`

### IN-02: The protoast re-sync note lost its divergence inventory

**File:** `internal/protoast/build.go:449-452` (also `internal/protoast/protoast.go:2,10`)
**Category:** code quality (maintainability)

The old comment ended with "re-apply the divergence on every compiler-lab re-sync (see internal/easypc/VENDORED.md)". The deleted `VENDORED.md` spelled out the full procedure: re-apply the rune-column `pos()` computation in `build.go`, the position wording in `protoast.go`, and the tab-as-1 expectations in `protoast_test.go`, then re-run both D-07 gates. The replacement says only "re-apply the divergence whenever the github.com/easyp-tech/easyprotoc version in go.mod is bumped", leaving the surface area to be rediscovered from git history.

Two mitigations keep this at INFO: `internal/protoast/position_test.go:13-21` documents the rune/tab-1 contract on its own, and `internal/core/position_parity_test.go:16-21` pins the legacy equality, so a careless bump fails the gates. Separately, `internal/protoast/protoast.go:2` still calls the compiler "easyp-protoc" while citing the `easyprotoc` module path — a harmless name/path mismatch that a wording pass could normalize.

**Fix:** Name the three files/expectations and the two D-07 gate names in the comment; optionally use "easyprotoc" consistently in the package doc.

### IN-03: Stale "easypc linker" prose in the rule test

**File:** `internal/rules/package_no_import_cycle_test.go:28`
**Category:** code quality (stale naming)

The cycle-shape explanation still says "The easypc linker hard-rejects a direct file A<->B import cycle". No `easypc` component exists after this commit. The line was outside D-02's exact seven-file set and outside the D-09 path grep (it contains no `internal/easypc` path), and `02-01-SUMMARY.md` records it as known non-blocking residue, so it is correctly excluded from the phase's gates — but it is now the only surviving in-repo prose that names the deleted compiler.

**Fix:** Reword during Phase 3's seal/documentation pass; no behavioral consequence.

## Accepted (not defects)

These items are stale references to the deleted tree that the phase deliberately deferred; they are recorded here so they are not mistaken for review misses:

| Item | Location | Disposition |
|---|---|---|
| Excluded-path entry for the deleted tree | `.golangci.yml:7` (comment naming `internal/easypc/VENDORED.md`) and `:9` (`internal/easypc`) | Accepted by D-07; golangci-lint is not a Phase 2 gate, only `staticcheck` is enabled, and an exclusion for a nonexistent path matches nothing. Owner: Phase 3. |
| "Do not edit anything inside internal/easypc" prohibition | `AGENTS.md:92` | Accepted by D-07; agent-facing seal text, owned by Phase 3. Harmless because the directory cannot be edited — it does not exist. |
| Feature/spec documentation describing the vendored tree | `.spec/features/lint-easypc-infra/explore.md`, `requirements.md`, `approved/explore.md`, and four `revisions/*` files | Phase 3 documentation ownership; historical feature record. |
| Planning history | `.planning/**` | Explicitly out of scope for this review. |

## Verification Evidence

All checks were run against the current tree `82d59ca`, which has zero non-`.planning` diff from the swap commit (`git diff --name-only 5342455 HEAD -- ':(exclude).planning'` empty), so it is the commit under review.

| Claim | Command | Result |
|---|---|---|
| Atomic single commit | `git rev-list --count df23d64..5342455` | 1 |
| Changeset shape | `git diff --name-status df23d64 5342455` | 648 `D`, 7 `M`, 0 `A` (15 insertions, 22921 deletions) |
| No vendored path in Go | `git grep -n 'internal/easypc' -- '*.go'` | empty (rc 1) |
| Module import census | `git grep -h -E '^[[:space:]]*"github\.com/easyp-tech/easyprotoc/' -- '*.go'` | 9 lines: 8 swapped (`link.go:17`, `link_test.go:10`, `lint_linked_matrix_test.go:15`, `build.go:10-11`, `outline.go:4-5`, `protoast_test.go:6`) + pre-existing `breaking_profiles.go:14` |
| Import grouping | read the five import blocks | third-party group separated from `github.com/easyp-tech/easyp/...`; `easyprotoc` sorts above `testify` in the three test files; matches `.spec/agent-rules.md:12` and `.golangci.yml` (staticcheck only) |
| Tree absent from index/commit/worktree | `git ls-files internal/easypc`, `git ls-tree -r HEAD`, `git cat-file -e 5342455:internal/easypc` | 0 paths, 0 paths, rc 128; `cmd/` holds only `easyp` and `easyp-mcp` |
| Diff shape | `git diff -U0` + awk gate over the seven files | removed=14 badremoved=0 added_nonblank=14 badadded=0; plus one blank line removed (`protoast_test.go`) and one added (`link.go`) |
| Strip gate (seam untouched) | per-file diff of comment/blank/import-path-stripped content, base vs swap | silent for all seven files |
| Formatting/whitespace | `gofmt -l` (seven files), `git diff --check` | empty, rc 0 |
| Build and vet | `go build ./...`, `go vet ./...` | rc 0 (2.0 s), rc 0 |
| Package graph | `go list ./...`, test-package filter | 37 packages, 34 with tests (matches the claimed post-state) |
| Affected tests | `go test -race -count=1 ./internal/protoast ./internal/core` | `ok` both (1.013 s / 1.501 s) |
| Golden gate | `go test -race -count=1 ./internal/rules -run '^TestLintGoldenFullRegistry$'` | PASS (0.01 s), 94-finding constant unchanged |
| Dependency freeze | `git diff --name-only df23d64 5342455 -- go.mod go.sum`; `git diff 5342455 HEAD -- go.mod go.sum` | empty; empty |
| Pins | `go list -m -f '{{.Path}} {{.Version}} {{.Indirect}}'` | `easyprotoc v0.0.0-20261104120806-f378f3f3a05f direct`; `protobuf v1.37.0-rc.1 direct` |
| Tidy idempotence | `go mod tidy -diff` | empty (rc 0) |
| Escape hatches | `grep -nE '^(replace\|vendor)' go.mod`; `ls go.work vendor` | no match; neither exists |
| Module integrity | `go mod verify` | all modules verified |
| Upstream replacement binary | `go list -f '{{.Name}}' .../cmd/easyprotoc`; `go build -o /dev/null ...` | `main`; rc 0 |
| Deleted-fixture isolation | `git grep -E 'audit_probes\|audit_perturb\|audit_m2\|audit_m4\|m3_adversarial\|m4_adversarial\|m5_adversarial\|testdata/semantic\|testdata/e2e\|testdata/errors\|testdata/lexer'` outside planning/spec | empty (rc 1) |
| Build-tooling isolation | grep `easypc`/`easyp-protoc`/`lexbench` in `Taskfile.yml`, `Dockerfile`, `.github/`, `scripts/` | no references |

### Cross-snapshot source audit (extra check, beyond the phase's own evidence)

The deleted vendored copy was vendored from compiler-lab SHA `08cef10…` (`VENDORED.md`), while `go.mod` pins `easyprotoc@f378f3f…`; they are different snapshots and the pinned module reorganized private packages (`pkg/*` → `internal/*`, `pkg/parser` → `parser`, `pkg/lexer` → `lexer`). I compared all 48 vendored `.go` files against the module cache after normalizing the import path:

- No vendored file lacks a module counterpart; the module adds only `protoc/example_test.go` (a test), and both deleted `cmd` mains have no replacement in the repo by design (D-04).
- All remaining deltas are import-path/gofmt/doc-comment changes or four semantically equivalent refactors: `internal/linker/resolve.go` drops dead `symID` zeroing on the `ok = false` path; `internal/sourceinfo/comments.go` uses `TrimPrefix`; `internal/sourceinfo/generator.go` uses `append(…, subTags...)`; `lexer/token.go` uses `fmt.Fprintf` instead of `WriteString(fmt.Sprintf(…))`.
- The 20 embedded WKT `.proto` files are byte-identical, and `protoc/syntax/syntax.go` differs only in one doc word.
- `protoc/resolver.go`, `protoc/options.go`, `protoc/result.go`, `protoast`'s consumed `parser` surface, and the recursive-import diagnostic at `driver.go:319-323` are equivalent, which is consistent with the green golden/position/core tests.

No security findings: the new trust surface is a single module dependency pinned by exact pseudo-version and `go.sum` hashes (`go mod verify` clean, Apache-2.0), with no `replace`/`vendor`/`go.work`, and no new network, auth, or file-access path.

No necessity findings: this changeset introduces no runtime construct (no guards, abstractions, flags, or requirements) — only import paths, comments, and deletions — so there is nothing whose guarded state could be impossible.

## Review Notes

- The phase framing quotes "14 removed / 14 added lines"; the literal diffstat is 15/15. The extra pair is whitespace only — a blank line removed where the project import group collapsed in `protoast_test.go`, and a blank line added to separate the new third-party group in `link.go`. The "14 path-bearing lines" claim and the strip gate are accurate; only the unqualified prose is loose.
- Full `-race` battery, Acceptance Gates A–D ×2, and the CLI matrix were not re-executed (review instruction; `02-02-SUMMARY.md` records them). The targeted runs above independently exercise the swapped lint path, the golden corpus, and the D-07 position contracts.
- `.planning/**` history references and the pre-existing tracked `.claude/CLAUDE.md` are outside this phase's scope and were not reviewed.

---

_Reviewed: 2026-11-21T13:35:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: deep_
