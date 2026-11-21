---
phase: 01-merge-main-i-primirenie-putey
plan: 01
subsystem: infra
tags: [git-merge, go-modules, mvs, easyprotoc, protobuf, sourceview, blob-gates]

# Dependency graph
requires: []
provides:
  - "Merge commit 994c503: origin/main (#253) joined under feat/lint-easypc-infra; the branch is no longer behind (HEAD..origin/main = 0)"
  - "internal/core/breaking_profiles.go = origin/main blob 6e45a05 (easyprotoc/protoc compile path plus breakingImportResolver)"
  - "internal/core/dom.go = pre-merge branch blob 52bfbd7 (type Position = protoast.Position)"
  - "internal/core/core.go = union blob 34987d1 (main's sourceSession plus the branch's importCache/parsedImports)"
  - "internal/sourceview/{root_tree,session,session_test}.go and local.go byte-identical to origin/main"
  - "go.mod/go.sum MVS pins: easyprotoc v0.0.0-20261104120806-f378f3f3a05f, protobuf v1.37.0-rc.1; zero replace/vendor directives"
affects: [phase-02, phase-03, phase-04]

# Actuals (#2632) — chars/4 over the realized first-parent diff (29628 chars / 4)
actuals:
  tokens: 7407
  tasks: 2
  # Measured as git rev-list --count 79e353a..HEAD = 2: the authored merge commit
  # plus origin/main's #253 commit that the merge brought into this branch's history.
  commits: 2

# Tech tracking
tech-stack:
  added:
    - "github.com/easyp-tech/easyprotoc v0.0.0-20261104120806-f378f3f3a05f (protoc package) as a direct requirement of the breaking path"
    - "google.golang.org/protobuf v1.37.0-rc.1 (MVS lift required by easyprotoc)"
  patterns:
    - "Side-selective semantic merge: each conflicted file resolved from one side wholesale, union only where both sides add disjoint state"
    - "Blob-SHA resolution gates (index blob == parent/union blob) instead of diff review"

key-files:
  created:
    - internal/sourceview/root_tree.go
    - internal/sourceview/session.go
    - internal/sourceview/session_test.go
  modified:
    - go.mod
    - go.sum
    - internal/api/breaking_v1.go
    - internal/core/breaking_profiles.go
    - internal/core/core.go
    - internal/sourceview/local.go

key-decisions:
  - "breaking_profiles.go taken from origin/main wholesale: the branch's only delta there was the Position spelling, subsumed by the kept dom.go alias; taking the branch side would silently drop #253"
  - "dom.go kept from the branch (blob 52bfbd7): main's meta.Position alias is rejected so every rules call site keeps compiling against protoast.Position (v1.0 D-06)"
  - "core.go resolved by hand-editing only the import block; the result matched the rehearsed union blob 34987d1 exactly, proving struct fields, Options and New auto-merged"
  - "One merge commit with subject 'merge: origin/main (#253 Use compiler for breaking check)' and a body naming the semantic rule and all three resolved files"
  - "Full -race battery, acceptance, golden and position gates deliberately excluded here and budgeted once in plan 01-02"

patterns-established:
  - "Merge resolution is proven by content hashes: index blob equality against each parent (or the union blob) plus first-parent 9-file set equality"
  - "Freeze gate as a pathspec diffstat: internal/rules, internal/easypc, internal/protoast, internal/core/lint_*, link.go, proto_info_read.go, dom.go must show zero lines"

requirements-completed: [IMPORT-REC-01]

coverage:
  - id: D1
    description: "origin/main (#253) merged under feat/lint-easypc-infra as one two-parent merge commit; the branch no longer lags upstream"
    requirement: IMPORT-REC-01
    verification:
      - kind: integration
        ref: "git merge-base --is-ancestor origin/main HEAD; git rev-list --count HEAD..origin/main == 0; git log -1 --format=%P -> 79e353a5eb1f001b065df89ca4645d962a74ea55 5d36266d71e85f8f0fc3fa41f787a03ee5c68a79"
        status: pass
    human_judgment: false
  - id: D2
    description: "Three internal/core conflicts resolved byte-exactly: breaking_profiles from main, dom.go from the branch, core.go as the import union"
    requirement: IMPORT-REC-01
    verification:
      - kind: integration
        ref: "git rev-parse HEAD:internal/core/breaking_profiles.go == origin/main blob 6e45a05; HEAD:internal/core/dom.go == HEAD^1 blob 52bfbd7; HEAD:internal/core/core.go == 34987d1"
        status: pass
    human_judgment: false
  - id: D3
    description: "Lint contour untouched by the merge and main's breaking/sourceview/API code byte-identical to origin/main"
    requirement: IMPORT-REC-01
    verification:
      - kind: integration
        ref: "git diff --stat 79e353a 994c503 -- internal/rules internal/easypc internal/protoast 'internal/core/lint_*' internal/core/link.go internal/core/proto_info_read.go internal/core/dom.go (empty); git diff origin/main 994c503 -- internal/core/breaking_profiles.go internal/sourceview internal/api/breaking_v1.go (empty)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Dependency pins adopted as-is from main with no replace/vendor escape hatch and no tidy drift"
    requirement: IMPORT-REC-01
    verification:
      - kind: integration
        ref: "go list -m github.com/easyp-tech/easyprotoc == v0.0.0-20261104120806-f378f3f3a05f; google.golang.org/protobuf == v1.37.0-rc.1; go mod tidy -diff empty; grep -c '^replace'/'^vendor' go.mod == 0"
        status: pass
    human_judgment: false
  - id: D5
    description: "Commit provenance: subject and body name #253 and all three resolved files"
    requirement: IMPORT-REC-01
    verification:
      - kind: integration
        ref: "git log -1 --format=%B 994c503 contains #253, internal/core/breaking_profiles.go, internal/core/dom.go, internal/core/core.go"
        status: pass
    human_judgment: false

# Metrics
duration: 3min
completed: 2026-11-21
status: complete
---

# Phase 1 Plan 1: Merge origin/main (#253) with semantic internal/core resolution Summary

**Merged origin/main #253 onto feat/lint-easypc-infra with blob-exact resolution: main's easyprotoc breaking path, the branch's protoast Position model, a union core.go, and a byte-untouched lint contour.**

## Performance

- **Duration:** 3 min
- **Started:** 2026-11-21T11:22:10Z
- **Completed:** 2026-11-21T11:24:54Z
- **Tasks:** 2 (Task 1: merge + resolution + commit; Task 2: post-commit purity/provenance gates, no code delta)
- **Files modified:** 9 versus the first parent (6 modified, 3 added)

## Accomplishments

- One merge commit (`994c503`) with parents `79e353a` (pre-merge branch base) and `5d36266` (origin/main #253); ancestry and aliveness gates now hold (`behind = 0`).
- The three `internal/core` conflicts were resolved byte-exactly and proven by blob SHA rather than by a green build: `breaking_profiles.go` = main's `6e45a05` (easyprotoc/protoc), `dom.go` = the branch's `52bfbd7` (`Position = protoast.Position`), `core.go` = the union `34987d1`.
- The first-parent changeset is exactly the 9 expected files; the staged set was asserted equal before commit, so no untracked planning artifact entered the merge.
- The lint contour and `internal/easypc` show a zero-line diff against the pre-merge base; `internal/sourceview`, `internal/api/breaking_v1.go` and `internal/core/breaking_profiles.go` are byte-identical to origin/main.
- Dependency provenance is main's MVS as-is: `easyprotoc v0.0.0-20261104120806-f378f3f3a05f`, `protobuf v1.37.0-rc.1`, `go mod tidy -diff` empty, zero replace/vendor lines.

## Task Commits

Each task was committed atomically:

1. **Task 1: Pre-flight, merge origin/main, three-file semantic resolution, one merge commit** - `994c503` (merge)
2. **Task 2: Post-commit purity, provenance, and lint-contour freeze gates** - no authored commit (gate-only task; all evidence recorded below)

**Plan metadata:** the `docs(01-01)` commit created immediately after this file.

## Files Created/Modified

- `internal/core/breaking_profiles.go` - replaced wholesale with origin/main's `easyprotoc/protoc` implementation plus `breakingImportResolver`
- `internal/core/dom.go` - restored byte-for-byte from the branch (`type Position = protoast.Position`); zero diff versus the pre-merge base
- `internal/core/core.go` - union: `github.com/easyp-tech/easyp/internal/protoast` and `internal/sourceview` imports; `sourceSession` and `importCache`/`parsedImports` coexist
- `internal/sourceview/local.go` - main's `rootTree`-based open path
- `internal/sourceview/root_tree.go` - new from main: per-directory root handles
- `internal/sourceview/session.go` - new from main: per-command root cache with nil-receiver fallback
- `internal/sourceview/session_test.go` - new from main, verbatim (carries main's pre-existing errcheck finding, deferred to Phase 3)
- `internal/api/breaking_v1.go` - main's `sourceview.NewSession()` lifecycle passed through `core.Options.SourceSession`
- `go.mod`, `go.sum` - MVS pins adopted from main

## Gate Evidence

Identifiers:

- `PRE_MERGE_BASE` = `79e353a5eb1f001b065df89ca4645d962a74ea55`
- `MERGE_SHA` = `994c503968a20ecdc1b4c1ebb2a9f06f23973386`
- `origin/main` = `5d36266d71e85f8f0fc3fa41f787a03ee5c68a79`

Pre-flight (read-only until the merge):

- `git fetch origin` rc 0; `git rev-parse origin/main` = `5d36266...` (match).
- `git rev-list --count HEAD..origin/main` = 1 before the merge; `git merge-base --is-ancestor origin/main HEAD` rc 1 (merge genuinely required).
- `git merge-tree --write-tree HEAD origin/main` rc 1 with exactly three stage lines per file; ours/theirs blobs: breaking_profiles `8a3a10d4`/`6e45a055`, core.go `9b634722`/`a2419c88`, dom.go `52bfbd7d`/`3a35dc5e` — all as researched.
- `git cat-file -e 34987d1824df1c4c14337df4b04c2625761e7b65` rc 0 (rehearsed union blob present).
- `go mod download github.com/easyp-tech/easyprotoc@... google.golang.org/protobuf@v1.37.0-rc.1` rc 0.

Resolution and pre-commit gates:

- Conflict-marker sweep over `*.go` for both prefixes: no output.
- `git rev-parse :internal/core/breaking_profiles.go` = `6e45a055c1e12efb25d14d210fdfa5ee0c52ba90` (= origin/main).
- `git rev-parse :internal/core/dom.go` = `52bfbd7d645a4c2d587ea84d08c9bbf27e7e0f67` (= `HEAD` pre-merge blob).
- `git rev-parse :internal/core/core.go` = `34987d1824df1c4c14337df4b04c2625761e7b65` (union blob).
- `git diff --cached --name-only | sort` = exactly the 9 expected paths (9 lines).
- Pins present; `grep -c '^replace'` = 0; `grep -c '^vendor'` = 0.
- `go build ./...` rc 0.

Post-commit gates:

- `git diff --stat 79e353a 994c503 -- internal/rules internal/easypc internal/protoast 'internal/core/lint_*' internal/core/link.go internal/core/proto_info_read.go internal/core/dom.go` -> empty.
- `git diff origin/main 994c503 -- internal/core/breaking_profiles.go internal/sourceview internal/api/breaking_v1.go` -> empty.
- `go list -m -f '{{.Version}}' github.com/easyp-tech/easyprotoc` -> `v0.0.0-20261104120806-f378f3f3a05f`; same for `google.golang.org/protobuf` -> `v1.37.0-rc.1`.
- `go mod tidy -diff` -> empty; zero replace/vendor lines.
- `git log -1 --format=%B 994c503` contains `#253`, `internal/core/breaking_profiles.go`, `internal/core/dom.go`, `internal/core/core.go`.
- `git status --porcelain --untracked-files=no` -> empty after all gates.
- `go vet ./...` -> clean (optional evidence).
- Acceptance greps: `easyprotoc/protoc` in breaking_profiles = 1; `protocompile` there = 0; `internal/sourceview` in core.go = 1; `type Position = protoast.Position` in dom.go = 1.
- First-parent changeset: `go.mod`, `go.sum`, `internal/api/breaking_v1.go`, `internal/core/breaking_profiles.go`, `internal/core/core.go`, `internal/sourceview/local.go`, `internal/sourceview/root_tree.go`, `internal/sourceview/session.go`, `internal/sourceview/session_test.go`.

## Decisions Made

- `breaking_profiles.go` from origin/main wholesale (branch delta was a subsumed Position spelling); the branch side was never considered because it would silently drop #253.
- `dom.go` kept from the branch; main's `meta.Position` alias rejected, preserving every `core.AppendIssue(..., protoast.Position, ...)` call site.
- `core.go` hand-edited in the import block only, then proven equal to the rehearsed union blob `34987d1` — the struct/`Options`/`New` portions auto-merged as researched.
- Merge commit subject/body follow CONTEXT D-01/D-02 and the `8bd9626` body shape; validation line points at `go build ./...`, the full `-race` battery, and the named 01-02 gates.
- Requirements marking: `IMPORT-REC-01` is intentionally copied from the plan frontmatter but not yet marked complete in REQUIREMENTS.md — it is shared with plan 01-02, so the shared-ID gate defers the checkbox until 01-02's SUMMARY exists.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- `gsd_run` is not on `PATH` in this environment; state/roadmap verbs were invoked from `$HOME/.zcode/gsd-core/bin/gsd_run`. Environment note only, no plan deviation.
- main's pre-existing golangci-lint errcheck finding at `internal/sourceview/session_test.go:29` is inherited unchanged and intentionally not fixed (D-09 excludes `lint:go`; owner Phase 3).
- Push to `origin/feat/lint-easypc-infra` remains out of scope for Phase 1 per CONTEXT/orchestrator decision 2.
- `gsd_run check evaluation-scope --plan 01-01` resolves by plan-subject prefixes and therefore returns the archived v1.0 Phase 01 commits (`docs(01-01)`, `feat(01-01)`) rather than this plan's merge commit, whose subject is mandated verbatim by the plan. The authoritative commit evidence for this plan is the explicit `MERGE_SHA` / parent / blob gates above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 01-02 can pin its full-battery evidence to `MERGE_SHA = 994c503968a20ecdc1b4c1ebb2a9f06f23973386` with `PRE_MERGE_BASE = 79e353a5eb1f001b065df89ca4645d962a74ea55` (for the lint-contour freeze comparison).
- The merge is proven pure: main's breaking/sourceview/API code is byte-identical to origin/main and the lint contour is untouched, so any parity change in 01-02 is a resolution error, not a new norm.
- Budget for the full `-race` battery in 01-02: ~7 min, `internal/api` is the long pole (~309 s cold). Run it after the merge commit from a clean tree.

---
*Phase: 01-merge-main-i-primirenie-putey*
*Completed: 2026-11-21*

## Self-Check: PASSED

Verified: all three created sourceview files exist on disk; the SUMMARY exists; commit `994c503968a20ecdc1b4c1ebb2a9f06f23973386` is an ancestor of `HEAD`.
