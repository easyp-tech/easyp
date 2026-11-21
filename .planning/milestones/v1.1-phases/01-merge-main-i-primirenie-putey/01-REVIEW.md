---
phase: 01-merge-main-i-primirenie-putey
reviewed: 2026-11-21T11:53:54Z
depth: deep
files_reviewed: 10
files_reviewed_list:
  - go.mod
  - go.sum
  - internal/api/breaking_v1.go
  - internal/core/breaking_profiles.go
  - internal/core/core.go
  - internal/core/dom.go
  - internal/sourceview/local.go
  - internal/sourceview/root_tree.go
  - internal/sourceview/session.go
  - internal/sourceview/session_test.go
blocking: 0
warnings: 3
info: 7
status: complete
---

# Phase 1: Code Review Report — merge `origin/main` (#253)

**Reviewed:** 2026-11-21T11:53:54Z
**Depth:** deep
**Files Reviewed:** 10 (9 files in the merge's first-parent diff + `internal/core/dom.go` as resolution context)
**Merge commit:** `994c503968a20ecdc1b4c1ebb2a9f06f23973386` — parents `79e353a5eb1f001b065df89ca4645d962a74ea55` (first) and `5d36266d71e85f8f0fc3fa41f787a03ee5c68a79` (`origin/main`, #253)
**Status:** complete — no merge-blocking defect; 3 warnings and 7 info items, all traced to upstream `origin/main` except where noted

## Summary

The merge itself is correct and provable. The hand edit is confined to the `internal/core/core.go` union and is coherent (both imports, both state fields, `Options.SourceSession` wired through `New`), no side was dropped, `internal/core/dom.go` keeps exactly one `Position` alias, and `internal/core/breaking_profiles.go`, `internal/sourceview/*`, and `internal/api/breaking_v1.go` are byte-identical to `origin/main`. The pins are main's MVS pins with no `replace`/`vendor` and `go mod tidy -diff` is empty.

Every finding below is either the already-known upstream errcheck or a latent defect inherited verbatim from `origin/main` (#253). **Nothing found here is a merge regression**: the adversarial pass did not produce a single defect attributable to the conflict resolution or the hand-edited union. The two warnings about `sourceview.Session`/`rootTree` lifecycle are real but not reachable from any current call site; they must not be "fixed" inside this merge commit because D-03/D-06 require byte fidelity to `origin/main` for these paths.

## Findings

| ID | Severity | File:Line | Finding | Evidence | Recommendation |
|----|----------|-----------|---------|----------|----------------|
| WR-01 | warning | `internal/sourceview/session_test.go:29` | Bare `defer session.Close()` discards the error returned by `(*Session).Close` (which joins `rootTree.Close` errors). Makes the repo's lint gate red. | Independent run: `bin/golangci-lint run ./...` (v2.14.0, per `.golangci.yml`) → `session_test.go:29:21: Error return value of 'session.Close' is not checked (errcheck)`, 1 issue, exit 1. File is byte-identical to `origin/main` (`git diff origin/main 994c503 -- internal/sourceview` empty) → upstream pre-existing, not a merge regression. | Known/upstream; owner Phase 3 per 01-01/01-02 flagged assumptions. Do **not** edit inside the merge commit (D-03, RESEARCH Pitfall 5). Follow-up: `t.Cleanup(func() { if err := session.Close(); err != nil { t.Errorf("Close: %v", err) } })`. |
| WR-02 | warning | `internal/sourceview/session.go:78` (with `:41`), `internal/sourceview/root_tree.go:129` (with `:30-37`) | `Close` is not a terminal-safe state: `Session.Close` sets `s.roots = nil` while `root()` writes `s.roots[rootPath] = entry` unconditionally, so any `OpenLocal` after `Close` — and the first `OpenLocal` on a zero-value `&sourceview.Session{}` (only `NewSession` initializes the map) — panics with `assignment to entry in nil map`. At the tree layer `rootTree.Close` sets `t.dirs = nil`, and `dir()` assumes the map holds `"."`: for `name != "."`, `split` calls `dir(path.Dir(name))`, which reaches `dir(".")` on the nil map and recurses `dir(path.Dir(".")) == dir(".")` forever (unrecoverable stack overflow). | `session.go:41` write vs `:78` nil assignment; `root_tree.go:37` recursive `dir`, `:129` `t.dirs = nil`, `:30-36` no closed/`"."` guard. Go semantics: nil-map read is safe, write panics. **Latent**: both production sessions (`internal/api/breaking_v1.go:74`, `:306`) are closed by `defer` at command end and never reopened, so no live caller hits it; the tree-level recursion is additionally masked by the `Session` panic. | Upstream-only fix (not this merge): lazily initialize the map under `s.mu` and add a `closed bool` returning `fs.ErrClosed`; in `rootTree`, keep the `dirs` map (or add `closed bool`) so `dir(".")` returns an error instead of recursing. |
| WR-03 | warning | `internal/sourceview/session.go:44` and `:73-76` | `Session` is documented "safe for concurrent use" (`:17-18`), but `Close` is not synchronized with an in-flight open: `entry.tree` is written inside `once.Do` (`:44`) without `s.mu`, while `Close` reads `entry.tree` under `s.mu` (`:73-74`). No happens-before edge → data race; if `Close`'s loop passes the entry before the `Once` body assigns, the newly opened `rootTree` is never closed (directory-handle/fd leak). If `Close` reaches `rootTree.Close` while another goroutine is resolving, that goroutine can observe `t.dirs == nil` and hit WR-02's unbounded `dir(".")` recursion. | `session.go:44` (write, outside lock), `:70-78` (locked read + `roots = nil`); `sync.Once` establishes happens-before only with other `Do` calls, not with a mutex-guarded reader; `root_tree.go:120-132`. **Latent**: current call sites are strictly sequential. | Upstream-only fix: either serialize `Close` against `root()` with a `closed` flag checked inside the `Once`/after it, or document explicitly that `Close` must not run concurrently with opens. Do not change in the merge commit. |
| IN-01 | info | `internal/api/breaking_v1.go:306` (vs `:74`) | `checkV1ExtendedBreakingSource` creates a second `sourceview.Session` even though `checkV1Policies` already owns one for the command and passes it to `core.New` at `:155`. On the extended-policy path the caller's session is never used, defeating the documented "one session per command" sharing model (`internal/core/core.go:82-84`, `internal/sourceview/session.go:13-18`) and re-opening root handles for every extended policy source. | `:74` `session := sourceview.NewSession()` + defer close; `:89` calls the extended path, which allocates its own session at `:306`; two `defer Close` scopes. Functional behavior unchanged (sequential, both closed). | Accept the caller's session as a parameter (or hoist the outer creation) so one command uses one root cache. Behavior-neutral cleanup for a later phase; not a merge fix. |
| IN-02 | info | `internal/core/breaking_profiles.go:753` | Duplicated compound condition: `matchingNumber != nil && matchingNumber.Name() != oldPair.name && profile.source \|\| matchingNumber != nil && matchingNumber.Name() != oldPair.name && profile.wireJSON`. Relies on `&&` > `\|\|` precedence; semantically equals `A && B && (source \|\| wireJSON)` but duplicates the whole predicate and invites a future edit to desynchronize the two branches. | Exact expression at `:753`; `jsonProfileEnabled` (`:1082-1084`) already encodes `source \|\| wireJSON`. | Replace with `matchingNumber != nil && matchingNumber.Name() != oldPair.name && (profile.source \|\| profile.wireJSON)`. Upstream file (adopted verbatim) — fix upstream or after the merge. |
| IN-03 | info | `internal/core/core.go:46-53` | The identical sentence ("holds parsed import outlines for lint; files are parsed in parallel, so it is shared and locked") is used for both the `importCache` field and the `parsedImports` type; the field comment adds the field name, the type comment adds the type name. Noisy duplication in the union file's only hand-touched region. | `core.go:47-48` field comment, `:52-53` type comment. This duplication pre-dates the merge (branch delta at `git diff 5c9185f 79e353a -- internal/core/core.go`), so it is not merge fallout. | Keep the type comment and shorten the field comment to "importCache caches lint import outlines per root; nil until Lint initializes it." |
| IN-04 | info | `internal/sourceview/session.go:60-63` | Doc contract mismatch: `Session.OpenLocal` is described as "the package OpenLocal through the session's roots", but the package version documents "Closing the returned file also closes the root handle" (`local.go:17-18`), while the session version returns the raw file from `openSelected` — closing it leaves the session's `rootTree` (and its directory handles) open. A caller that relies on the documented close semantics will hold handles until `Session.Close`. | `session.go:60-63` vs `local.go:17-18`, `local.go:36` (`localFile{..., root: root}` is only produced by the package-level wrapper). Today's sole caller (`internal/core/breaking_profiles.go:186-203`) closes per file and relies on the session for root lifetime, so no leak occurs. | Reword the Session method docs to state that the session, not the returned file, owns the root handles. Upstream doc fix. |
| IN-05 | info | `internal/core/breaking_profiles.go:373-375` | The shared `Position` alias now carries two column contracts: breaking issues fill `Column` from `protoreflect.SourceLocation.StartColumn + 1` (byte-oriented — this repo's own `internal/protoast/protoast.go:5-9` states upstream protoc columns are byte columns with tab stops of 8), while the aliased type's documented contract is "1-based columns counted in runes". The same `Issue.Position` is printed to users for both engines, so non-ASCII breaking findings report different columns than lint findings for the same location. `Offset` is also left at 0 (no consumer reads it today). | `breaking_profiles.go:371-376`; `internal/protoast/protoast.go:5-9`, `:19-24`; consumers only read `Line`/`Column` (`internal/api/breaking_v1.go:181`, `:499`, `internal/core/breaking_profiles.go:1118`, `:1128-1132`). **Pre-existing on the branch**: the pre-merge file built `protoast.Position` the same way (`git diff 5c9185f 79e353a -- internal/core/breaking_profiles.go` shows only the spelling change), so this is not a merge regression and not a Phase 1 action. | Note for the future position-contract work: either convert descriptor columns to rune columns at this boundary or document that breaking positions are byte columns. Do not touch here (D-03). |
| IN-06 | info | `internal/core/breaking_profiles.go:97-105` | The target walk silently drops a physically distinct file whose canonical import path was already seen (`if seen[canonicalPath] { return nil }`): no diagnostic, and `targetPaths` keeps only the first walker path. A second file shadowed by a nested import root is therefore never compiled or compared — a silent false-negative for breaking detection rather than the compiler error the adjacent comment describes. | `breaking_profiles.go:93-105` (`canonicalBreakingPath` picks the *first* matching import root, so two different physical files can yield one identity), comment at `:97-98`. **Pre-existing** (identical lines at `79e353a:internal/core/breaking_profiles.go`), not introduced by #253. Reachability is limited on the v1 path: `CheckSourceCollisions` (`internal/modules/collisions.go:9-37`, called from `internal/api/policy_imports.go:244` and `internal/api/breaking_scope.go:203`) rejects duplicate import paths across the module's roots before compilation; legacy import roots built by `internal/api/runtime.go:65-91` do not run that check. | If the legacy path can present overlapping roots, emit a diagnostic (or a deterministic "shadowed target" error) instead of a silent skip; otherwise state first-wins explicitly in the comment. Upstream-only; no action in this merge. |
| IN-07 | info | `internal/core/breaking_profiles.go:140-143` | Diagnostics are joined multi-line and placed between the callee label and the wrapped sentinel: `fmt.Errorf("Compile: %s: %w", strings.Join(diagnostics, "\n"), err)`. CLI/log output becomes a multi-line message where the wrapped error is buried at the end of the last diagnostic line, which conflicts with the project's `"<callee>: %w"` convention and makes single-line log assertions brittle. | `breaking_profiles.go:139-144`; `AGENTS.md` convention (`fmt.Errorf("<callee>: %w", err)`). Diagnostic-producing path is upstream #253 by design (surfacing compiler diagnostics). | Keep the sentinel adjacent: `fmt.Errorf("Compile: %w: %s", err, strings.Join(diagnostics, "; "))` (or log diagnostics separately). Upstream-only. |

## Merge-Specific Verification (independently re-run)

**Resolution correctness**

- `git rev-parse 994c503:internal/core/core.go` = `34987d1824df1c4c14337df4b04c2625761e7b65` — equals the rehearsed union blob; the diff vs the first parent is exactly four additive hunks (`sourceview` import, `sourceSession *sourceview.Session`, `Options.SourceSession` + doc, `New` assignment) with zero deletions (`git diff 79e353a 994c503 -- internal/core/core.go`).
- Both sides survive the union: the branch's 13-line delta (`sync`, `internal/protoast`, `importCache`, `parsedImports` — `git diff 5c9185f 79e353a -- internal/core/core.go`) and main's 6-line delta are all present; no duplicated field, method, or import. `sourceSession` is consumed at `breaking_profiles.go:220`; `importCache` is lazily initialized at `internal/core/lint_linked.go:47-49` and nil-guarded at `lint_linked_read.go:81/95/115`.
- `git rev-parse 994c503:internal/core/dom.go` = `52bfbd7d645a4c2d587ea84d08c9bbf27e7e0f67` (pre-merge branch blob). Exactly one `type Position = protoast.Position` (`dom.go:19`); the only remaining `meta.Position` references are the conversion boundary (`dom.go:128`, `breaking_checker.go:368`) and a test fixture — no stale alias.
- `git grep -n -E '^(<<<<<<< |>>>>>>> )' -- '*.go'` → no output.
- `go build ./...` → exit 0; `go vet ./...` → exit 0 (no findings).

**Adoption fidelity (vs `origin/main` at `994c503`)**

- `git diff origin/main 994c503 -- internal/core/breaking_profiles.go internal/sourceview internal/api/breaking_v1.go` → empty (byte-identical: `breaking_profiles.go`, `local.go`, `root_tree.go`, `session.go`, `session_test.go`, `breaking_v1.go`).
- The branch never touched those paths: `git diff 5c9185f 79e353a -- internal/api/breaking_v1.go internal/sourceview` → empty; its only `breaking_profiles.go` delta was the two-line `Position` spelling, which the kept alias makes a no-op (`breaking_profiles.go:373` now resolves to `protoast.Position`).
- Lint-contour freeze: `git diff --stat 79e353a 994c503 -- internal/rules internal/easypc internal/protoast 'internal/core/lint_*' internal/core/link.go internal/core/proto_info_read.go internal/core/dom.go` → 0 lines.
- Ancestry/aliveness: `git merge-base --is-ancestor origin/main HEAD` → true; `git rev-list --count HEAD..origin/main` → 0. First-parent changeset is exactly the 9 expected files (456 insertions, 46 deletions).

**Dependency pins**

- `go list -m` → `github.com/easyp-tech/easyprotoc v0.0.0-20261104120806-f378f3f3a05f`, `google.golang.org/protobuf v1.37.0-rc.1`.
- `go mod tidy -diff` → empty (exit 0). No `replace` or `vendor` directive; no `vendor/` directory.
- `go.sum` contains the easyprotoc pair and the protobuf `v1.37.0-rc.1` pair, and the diff vs the first parent removes the `v1.36.11` pair — consistent with the merged build list. `github.com/go-git/go-billy/v5 v5.6.2` is direct exactly once (main's promotion) and absent from the indirect block.

**Other checks**

- `bin/golangci-lint run ./...` → exactly 1 issue (WR-01), i.e. the merge adds no new lint finding.
- No `TODO`/`FIXME`/`XXX`/`HACK`, no `console.log`/`fmt.Print` debug artifacts in the 10 reviewed files.
- Not a race (verified, not assumed): the unlocked `len(diagnostics)` read at `breaking_profiles.go:140` is safe because easyprotoc's driver joins every worker (`.../easyprotoc/internal/compiler/driver.go:156-186`) before `protoc.Compile` returns.
- Coexistence of `internal/easypc/protoc` and `github.com/easyp-tech/easyprotoc/protoc` in one binary is intentional dormancy (D-06, Phase 2 removes the vendored copy); not a defect of this phase.

## Out of Scope Here (recorded, not actioned)

- Full `-race` battery, acceptance A–D ×2, golden 94, position gates: not re-run per review instructions; evidence lives in `01-02-SUMMARY.md` and was not contradicted by any independent check above (build/vet/tidy/lint).
- Push of the merge commit to `origin/feat/lint-easypc-infra`: out of Phase 1 scope.
- `internal/easypc`, `internal/protoast`, `internal/core/lint_*`, `link.go`, `proto_info_read.go`: byte-frozen; reviewed only as call-chain context.

---

_Reviewed: 2026-11-21T11:53:54Z_
_Reviewer: ZCode (gsd-code-reviewer)_
_Depth: deep_
