# Phase 1: Merge main и примирение путей - Research

**Researched:** 2026-11-21
**Domain:** Git merge integration + Go module/MVS reconciliation (EasyP CLI, branch `feat/lint-easypc-infra`)
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Стиль слияния
- **D-01:** Слияние `origin/main` выполняется **merge-коммитом** (не rebase). PR #258 с 40+ коммитами открыт; rebase переписал бы историю и сломал review. Merge сохраняет SHA и делает последующий мерж PR #258 в main тривиальным. — **Reversibility:** costly — откат переписывает публичную историю PR-ветки.
- **D-02:** Конфликты `internal/core` правятся **внутри merge-коммита** (один коммит `merge origin/main`), не отдельными follow-up коммитами: PR-мерж не должен требовать ручного разрешения конфликтов.

#### Семантика разрешения конфликтов
- **D-03:** Правило приоритета: **main wins** для поведения breaking/sourceview (#253 — `breaking_profiles.go`, `core.go`, `dom.go`, `internal/sourceview/*`); **ветка wins** для переработанного lint-контура (новые файлы `lint_linked.go`, `link.go`, `proto_info_read.go` и т.д. остаются как есть). Пересечение (общие структуры `Core`, `dom.go`, `core.go`) — semantic merge: обе стороны сосуществуют, breaking использует `easyprotoc/protoc`, lint остаётся на вендор-пути `internal/easypc/protoc`.
- **D-04:** Если после merge падает тест из-за изменения поведения main (#253), разбор точечный: поведение main — истина для breaking/sourceview (прецедент REQ-3.2 v1.0: «main wins over lab copy»); lint-ожидания (golden, parity) не правятся — падение означает ошибку разрешения конфликта, а не новую норму.

#### Зависимости
- **D-05:** `go.mod` принимает main-пин `github.com/easyp-tech/easyprotoc v0.0.0-20261104120806-f378f3f3a05f` и `google.golang.org/protobuf v1.37.0-rc.1` **как есть** (MVS, без понижения, без `replace`). Тот же пин переиспользуется в Phase 2 — никаких повторных go get.

#### Дормация вендор-пути
- **D-06:** `internal/easypc` и все импорты `internal/easypc/...` в Phase 1 **не трогаются** (byte-frozen): фаза проверяет, что merge не сломал существующий lint-путь. Удаление/переключение — строго Phase 2.
- **D-07:** `internal/protoast` в Phase 1 не меняется (D-07a/D-07b position-тесты — часть гейта).

#### Ветка и коммиты
- **D-08:** Работа идёт в текущей ветке `feat/lint-easypc-infra` (GSD `branching_strategy: none`); коммиты атомарные, дефолтная ветка не затрагивается.

#### Ворота фазы
- **D-09:** Гейт после merge: `go build ./...` + `go test -race -count=1 ./...` (полная батарея), плюс корпусный lint-паритет Gates A–D ×2 и golden 94 findings байт-точно без правки ожиданий; `git merge-base --is-ancestor origin/main HEAD` истинно.

### Claude's Discretion
Конкретные hunks при разрешении конфликтов, порядок прогонов, формулировки комментариев в merge-коммите — на усмотрение исполнителя в рамках D-03/D-04.

### Deferred Ideas (OUT OF SCOPE)

- Взятие upstream-ветки `feat/easyp-integration` (generation metadata fix) — не в v1.1, ждёт мержа upstream / codegen-серии.
- Bump easyprotoc при появлении тегов — v2 requirements.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| IMPORT-REC-01 | WHEN ветка v1.1 готова к переключению импортов, the system SHALL содержать merge `origin/main` (#253 «Use compiler for breaking check») с семантически разрешёнными конфликтами `internal/core` (breaking-путь на `easyprotoc/protoc`, protobuf `v1.37.0-rc.1`); после слияния ветка не отстаёт от `origin/main` и полная батарея тестов зелёная | Exact merge recipe + deterministic blob gates (§Code Examples); rehearsed full `-race` battery 41 pkgs/0 FAIL, acceptance gates, golden 94, position gates (§Validation Architecture); ancestry/aliveness checks (§Code Examples step 4); pin verification via `go list -m` (§Standard Stack) |

</phase_requirements>

## Summary

The phase is a single merge of `origin/main` (`5d36266`, PR #253 "Use compiler for breaking check") into `feat/lint-easypc-infra` (`96b8c6a`, 278 commits ahead, 1 behind). The merge is conflict-reproducible and was **rehearsed end-to-end in a throwaway worktree**: exactly three content conflicts appear (`internal/core/breaking_profiles.go`, `internal/core/core.go`, `internal/core/dom.go`); everything else auto-merges. The branch never touched `internal/sourceview` (`git diff 5c9185f..HEAD -- internal/sourceview` is empty) and only added a comment to `internal/api/runtime.go`, so all of main's #253 sourceview/API changes apply cleanly. The correct resolution is mechanical and side-selectable per file: take main's `breaking_profiles.go` wholesale (the branch's only delta there was a `Position` spelling that is subsumed by the branch's alias), keep the branch's `dom.go` wholesale (byte-identical to pre-merge), and hand-merge only the import block of `core.go` (struct fields, `Options`, and `New` auto-merge). The resulting merge commit touches exactly 9 files versus its first parent — main's 10-file commit minus `dom.go` — and **zero** files of the lint contour (`internal/rules`, `internal/easypc`, `internal/protoast`, `internal/core/lint_*`, `link.go`, `lint_model.go`).

The rehearsed resolution was proven green: `go build ./...` OK, `go test -race -count=1 ./...` → **41 packages ok, 0 FAIL** (api 308.9 s), `TestLintGoldenFullRegistry` (94 findings) PASS, `TestAcceptanceGates` (Gates A–D ×2, baseline `1e9a596` vs HEAD) PASS, position gates PASS, `go vet ./...` clean, `go mod tidy -diff` empty, schema generation drift-free. Three resolution blobs are deterministic and can be asserted directly: `breaking_profiles.go` = origin/main's blob, `dom.go` = branch's blob, `core.go` = the union blob.

Two phase-level risks dominate the plan. First, **`git merge` refuses to run while the index is dirty** (empirically confirmed, including the exact staged-rename shape currently present in the repo: 156 staged `R100` planning renames + modified `.planning/STATE.md`), so the merge must start from `git status --porcelain --untracked-files=no` empty. Second, `origin/main` itself carries **one pre-existing golangci-lint finding** in `internal/sourceview/session_test.go:29` (errcheck `defer session.Close()`), identical with the pinned 2.14.0 binary; the merge adds no new findings, but `task lint:go` is already red on main, so it must not be treated as a merge regression.

**Primary recommendation:** Execute `git merge --no-ff --no-commit origin/main`, resolve in this order — `git checkout origin/main -- internal/core/breaking_profiles.go`, `git checkout <pre-merge-base> -- internal/core/dom.go`, hand-add both imports in `internal/core/core.go` — stage, assert the three blob SHAs plus "no conflict markers" plus the 9-file changed set, build, commit one merge commit, then run the full `-race` battery and the named golden/acceptance/position gates.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Breaking descriptor compilation | `internal/core` (engine) | external module `github.com/easyp-tech/easyprotoc/protoc` | #253 replaces `bufbuild/protocompile` with `protoc.Compile`; main owns this behavior (D-03) |
| Bounded source opens + session cache | `internal/sourceview` (IO library) | `internal/api` (session lifecycle), `internal/core` (consumer) | main adds `Session`/`rootTree`; `api/breaking_v1.go` creates one session per command, `core.breakingSourceAccessor` consumes it |
| Lint engine, rules, linked model | `internal/core` + `internal/rules` | `internal/easypc` (sealed vendor copy) | Phase 1 freezes this path byte-identically; vendor path stays dormant until Phase 2 (D-06) |
| Issue position model | `internal/core/dom.go` | `internal/protoast` | `Position` must remain `protoast.Position` (v1.0 D-06); main's `meta.Position` alias must be dropped |
| Dependency pins / build list | `go.mod` / `go.sum` (build system) | Go MVS | main's pins adopted as-is; no `replace`/`vendor` (D-05) |
| CLI issue output (text/JSON) | `internal/api` printers | `core.Issue` | explicit field formatting; alias choice must not change output (pinned by stdout golden) |
| Git integration / merge commit | repository working tree | CI (`tests.yml`) | one merge commit on the feature branch; CI runs `lint:go`, cross-build, `task test`, `proto:check` |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/easyp-tech/easyprotoc` | `v0.0.0-20261104120806-f378f3f3a05f` | Proto compiler consumed by the breaking path (`protoc` package) | The exact pin already used by `origin/main` #253; same module the lint path migrates to in Phase 2. `[VERIFIED: go.mod after merge line 10; git show origin/main:go.mod]` |
| `google.golang.org/protobuf` | `v1.37.0-rc.1` | Protobuf runtime/reflection; required by easyprotoc | easyprotoc's own `go.mod` requires `v1.37.0-rc.1`, which wins MVS. `[VERIFIED: /home/nil/go/pkg/mod/cache/download/github.com/easyp-tech/easyprotoc/@v/v0.0.0-20261104120806-f378f3f3a05f.mod]` |
| `github.com/go-git/go-billy/v5` | `v5.6.2` | Filesystem abstraction (gitmodules, acceptance test) | Promoted to a direct requirement by main #253 (also direct on the branch) — auto-merges. `[VERIFIED: go.mod:11 pre-merge; merged go.mod direct block]` |

### Supporting (retained, not removed)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/bufbuild/protocompile` | `v0.14.1` | Still used by `internal/core/generate.go`, `internal/modules/*`, `internal/migration/*` | Stays in `go.mod`; only `breaking_profiles.go` stops using it. `go mod tidy -diff` is empty after the merge, proving it is still required `[VERIFIED: go mod tidy -diff empty on rehearsed merge]` |
| `github.com/yoheimuta/go-protoparser/v4` | `v4.14.2` | Legacy proto AST for lint rule model + breaking read models | Unchanged; `meta.Position` remains used only inside `fromMetaPos` |
| `github.com/stretchr/testify` | `v1.11.1` | All test assertions | Unchanged |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Merge commit (D-01) | Rebase onto `origin/main` | Rejected by D-01: rewrites 278 public PR commits and invalidates review history |
| Taking main's `breaking_profiles.go` | Hand-merging both sides | Hand-merging would re-import `protoast` and reintroduce `protocompile` usage; the branch's only delta there is a spelling subsumed by the `Position` alias |
| Keeping branch's `dom.go` `Position` alias | main's `type Position = meta.Position` | Compile error at every rules call site passing `protoast.Position` to `core.AppendIssue`; also silently changes formatting behavior if it ever compiled (`protoast.Position` has `String()`, `meta.Position` does not) |
| `easyprotoc/protoc` for breaking | Stay on `bufbuild/protocompile` | Rejected by #253 and D-03 (main wins for breaking) |

**Installation:**

```bash
# No install command: the merge brings the pins. Warm the module cache explicitly:
go mod download github.com/easyp-tech/easyprotoc google.golang.org/protobuf
```

**Version verification (run before/after merge):**

```bash
go list -m github.com/easyp-tech/easyprotoc google.golang.org/protobuf
# expected after merge:
# github.com/easyp-tech/easyprotoc v0.0.0-20261104120806-f378f3f3a05f
# google.golang.org/protobuf v1.37.0-rc.1
```

Registry provenance (Go proxy, authoritative):

```json
{"Path":"github.com/easyp-tech/easyprotoc","Version":"v0.0.0-20261104120806-f378f3f3a05f",
 "Time":"2026-11-04T12:08:06Z",
 "Origin":{"VCS":"git","URL":"https://github.com/easyp-tech/easyprotoc",
           "Hash":"f378f3f3a05febd98f54aff2be9a570de239a59a"}}
```
`[VERIFIED: go list -m -json github.com/easyp-tech/easyprotoc@v0.0.0-20261104120806-f378f3f3a05f]`

## Package Legitimacy Audit

> The `package-legitimacy` seam supports npm/PyPI/crates; this phase is Go-only, so the Go registry (`proxy.golang.org`) was used as the ecosystem-appropriate authority, cross-checked with the module's VCS origin and `go.sum`.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| `github.com/easyp-tech/easyprotoc` | Go proxy (`proxy.golang.org`) | pseudo-version of commit `f378f3f` dated 2026-11-04; module public since 2026-11-20 | n/a (private→public Go module, same org) | `https://github.com/easyp-tech/easyprotoc` | OK | Approved — pinned by commit hash; already used by `origin/main` #253 |
| `google.golang.org/protobuf` v1.37.0-rc.1 | Go proxy | rc published 2026-10-20 | Very high (official Go protobuf) | `https://go.googlesource.com/protobuf` | OK | Approved — required transitively by easyprotoc (MVS) |
| `github.com/go-git/go-billy/v5` v5.6.2 | Go proxy | Pre-existing dependency, unchanged version | High | `https://github.com/go-git/go-billy` | OK | Approved — already direct on the branch pre-merge |

**Packages removed due to SLOP verdict:** none.
**Packages flagged as suspicious [SUS]:** none.
**Notes:** No `replace` or `vendor` directives are introduced. `go.sum` after merge contains exactly the easyprotoc pair and the protobuf rc.1 pair, and drops the v1.36.11 pair `[VERIFIED: rehearsed merge go.sum diff]`.

## Architecture Patterns

### System Architecture Diagram

```text
                    git origin
                        │  fetch
                        ▼
        origin/main 5d36266 (#253) ─────┐
                                        │ merge --no-ff
   feat/lint-easypc-infra 96b8c6a ──────┤
   (278 ahead, 1 behind)                ▼
                              ┌─────────────────────┐
                              │ merge commit         │
                              │ 9 files changed      │
                              └──────────┬───────────┘
                                         │
        ┌────────────────────────────────┼──────────────────────────────┐
        ▼                                ▼                              ▼
 breaking path (main wins)        shared model (union)          lint path (branch wins,
 ┌──────────────────────┐        ┌────────────────────┐        byte-frozen)
 │ api/breaking_v1.go   │        │ core/core.go       │        ┌──────────────────────┐
 │  NewSession()        │        │  sourceSession     │        │ rules/* (44)         │
 │  SourceSession       │        │  importCache       │        │ core/lint_linked.go  │
 └─────────┬────────────┘        └─────────┬──────────┘        │ core/link.go         │
           ▼                               │                   │ internal/easypc/...  │
 ┌──────────────────────┐                  │                   │ internal/protoast/.. │
 │ core/breaking_       │  Position alias  │                   └──────────┬───────────┘
 │ profiles.go          │◄─────────────────┘                              │
 │  protoc.Compile      │        dom.go: Position = protoast.Position     │
 │  breakingImportRes.  │                                                 ▼
 │  sourceSession.Open  │                                     output printers + golden
 └─────────┬────────────┘
           ▼
 sourceview.Session → rootTree (per-root dir handles)
           │
           ▼
 google.golang.org/protobuf (descriptors) ; easyprotoc/protoc (compiler)
```

### Recommended Project Structure (no new source files)

```text
internal/
├── api/breaking_v1.go          # main: creates sourceview.Session, passes SourceSession
├── core/
│   ├── breaking_profiles.go    # main's file verbatim (protoc.Compile + import resolver)
│   ├── core.go                 # union: sourceSession (main) + importCache (branch)
│   └── dom.go                  # branch's file verbatim (Position = protoast.Position)
└── sourceview/                 # main's files verbatim (local.go, root_tree.go, session.go, session_test.go)
```

### Pattern 1: Side-selection semantic merge (per-file, not per-hunk)

**What:** resolve each conflicted file by selecting one side wholesale where the other side's delta is provably subsumed, and hand-merge only the file where both sides add disjoint fields.
**When to use:** when one side's entire diff to a file is a rename/spelling that the other side's new type alias makes equivalent.
**Example (verified):**

```bash
git checkout origin/main -- internal/core/breaking_profiles.go   # main's rewrite; branch delta subsumed
git checkout "$PRE_MERGE_BASE" -- internal/core/dom.go           # branch's alias; main's delta is only the alias we reject
# core.go: remove conflict markers, keep BOTH imports:
#   "github.com/easyp-tech/easyp/internal/protoast"
#   "github.com/easyp-tech/easyp/internal/sourceview"
```

Why this is semantically correct:
- `breaking_profiles.go` on the branch differed from merge-base only by `-meta` import and `Position: meta.Position{` → `Position: protoast.Position{`. main changed the same line to `Position: Position{` and removed `meta`. With `type Position = protoast.Position`, main's `Position{...}` **is** `protoast.Position{...}` — the branch delta is fully preserved. `[VERIFIED: git diff 5c9185f..HEAD -- internal/core/breaking_profiles.go]`
- `dom.go` on the branch replaced `Position = meta.Position` with `Position = protoast.Position` and moved `Rule`/`ProtoInfo`/`AppendIssue`/`GetRuleName`/`buildError` out. main's only change was adding `type Position = meta.Position`. Keeping the branch file is therefore "branch wins for the model, main's addition is redundant". `[VERIFIED: internal/core/dom.go:15-19]`
- `core.go` auto-merges everywhere except imports; both sides add disjoint state. `[VERIFIED: rehearsed merge, core.go conflict markers only in import block]`

### Pattern 2: Mechanical blob gates instead of reading the diff

**What:** assert resolution correctness by content hash, not by eyeballing hunks.
**When to use:** every time a conflict resolves to "take one side" or "small union".

```bash
test "$(git rev-parse :internal/core/breaking_profiles.go)" = "$(git rev-parse origin/main:internal/core/breaking_profiles.go)"
test "$(git rev-parse :internal/core/dom.go)" = "$(git rev-parse "$PRE_MERGE_BASE":internal/core/dom.go)"
test "$(git rev-parse :internal/core/core.go)" = "34987d1824df1c4c14337df4b04c2625761e7b65"
```

The three expected values were reproduced twice from identical parents:
`breaking_profiles.go` → `6e45a055c1e12efb25d14d210fdfa5ee0c52ba90` (== origin/main blob),
`dom.go` → `52bfbd7d645a4c2d587ea84d08c9bbf27e7e0f67` (== branch blob),
`core.go` → `34987d1824df1c4c14337df4b04c2625761e7b65` (the union blob).
`[VERIFIED: two independent worktree rehearsals produced identical blobs]`

### Anti-Patterns to Avoid

- **Resolving `breaking_profiles.go` with `--ours`:** builds and likely passes the branch's own tests while silently discarding #253 (`easyprotoc` unused in the build, `go.mod` orphaned). This is the phase's only *silent* failure mode — gate it by blob/diff, not by green tests.
- **Taking main's `dom.go` alias:** `type Position = meta.Position` makes every `core.AppendIssue(..., protoast.Position, ...)` call site fail to compile (rules pass `protoast.FilePosition`, `imp.Pos`, `field.NamePos`, `field.NamePos`). Compile-caught, but a late surprise; fix by keeping the branch file.
- **Unioning the two `Position` aliases by hand:** duplicate declaration in package `core` → compile error.
- **`git add -A` into the merge commit:** the tree contains 25 untracked `.planning/*` files (23 `.planning/research/.cache/*.json`, `.planning/state.json`, `.planning/ui-reviews/.gitignore`); they must not enter the merge commit.
- **Running `git merge` with the current dirty index:** it aborts (see Pitfalls).
- **Fixing main's pre-existing errcheck finding inside the merge commit:** pollutes an otherwise pure merge; handle as a separate decision (Open Questions).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Proto compilation for breaking | Any local compiler/resolver wrapper | `protoc.Compile` + custom `protoc.Resolver` from #253 | The resolver shim (`breakingImportResolver`) is already main's code and preserves import-path order + `ErrFileNotFound` semantics |
| Session/root caching for local opens | New caching layer in `core` | main's `sourceview.Session` + `rootTree` | Nil-safe session (`if s == nil` at `session.go:50/67`) keeps existing tests and all non-session callers working |
| Merge verification | Manual diff review of hunks | blob SHA gates + `git diff origin/main -- <file>` + full battery | Deterministic and repeatable; catches the silent `--ours` failure |
| Conflict-marker detection | Custom parser | `git grep -n -E '^(<<<<<<< |>>>>>>> )' -- '*.go'` | Git already knows |
| Dependency reconciliation | `go get` / `go mod tidy` / `replace` | MVS from the merge | D-05; `go mod tidy -diff` is already empty after merge |
| Deciding "who is behind" | eyeballing log | `git rev-list --count HEAD..origin/main` and `git merge-base --is-ancestor` | Exact, scriptable, matches D-09 |

**Key insight:** every question in this phase is answerable by a content hash, a git ancestry check, or a test that already exists. No new code is needed — the phase's product is a correctly-resolved commit plus evidence.

## Runtime State Inventory

> This is a merge/migration phase (breaking path changes compiler); the five categories were checked explicitly.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — no database, cache key, collection name or user_id encodes the merged code paths. Verified by grepping for persistent-string usage of "easypc"/"protocompile" outside Go source (none). | None |
| Live service config | None — no external service (CI, Datadog, Tailscale, n8n) references the breaking implementation. CI only runs `task test`/`lint:go` from the repo. `[VERIFIED: .github/workflows/tests.yml]` | None |
| OS-registered state | None — no scheduler/pm2/launchd/systemd registration. Test-suite git worktrees are temporary and removed by `t.Cleanup` (`removeWorktreeRegistration`); a killed run can leave one, cleaned by `git worktree prune`. | None (cleanup note only) |
| Secrets/env vars | None — no env var or secret key names the compiler modules. `EASYP_PROXY`/`EASYP_GIT_TIMEOUT`/`EASYPPATH` are unrelated. | None |
| Build artifacts / installed packages | Go module cache gains `github.com/easyp-tech/easyprotoc@v0.0.0-20261104120806-f378f3f3a05f` (not cached before this session; download ~10 s). `go.sum` replaces protobuf v1.36.11 hashes with rc.1. `bin/` is gitignored. No Docker image or compiled binary carries the old path. | None — `go mod download` / first build fetches it; `go.sum` is committed by the merge |

## Common Pitfalls

### Pitfall 1: `git merge` aborts on a dirty index — including unrelated staged renames

**What goes wrong:** `git merge origin/main` fails with `error: Your local changes to the following files would be overwritten by merge: ... Merge with strategy ort failed.` even when the staged paths are nowhere near the merge's files.
**Why it happens:** while materializing the merged tree, git requires the index to match HEAD for every path it writes; a staged change — even one wholly unrelated to the merge's input paths — makes the write a refusal instead of an overwrite. The current repo has **156 staged `R100` renames** (`.planning/phases/*` → `.planning/milestones/v1.0-phases/*`) plus a modified `.planning/STATE.md` — exactly this shape.
**How to avoid:** before merging, commit (recommended, `commit_docs: true`) or stash the planning artifacts; gate with `test -z "$(git status --porcelain --untracked-files=no)"`. Untracked files (25 `.planning/*` files, mostly the research cache) do **not** block the merge.
**Warning signs:** `git status --porcelain` shows `R `, `M `, `A ` entries; merge exits before any conflict is reported.
`[VERIFIED: reproduced in a scratch repo with a staged rename and with an unrelated staged modification; unstaged-only changes merge fine]`

### Pitfall 2: Taking the branch's `breaking_profiles.go` silently drops #253

**What goes wrong:** `git checkout --ours -- internal/core/breaking_profiles.go` compiles (branch code was green pre-merge) and may pass the branch's tests, but the breaking path stays on `protocompile`, `easyprotoc` becomes an unused requirement, and Success Criterion 2 fails.
**Why it happens:** the conflict looks like "both changed the same import block and one line"; taking ours is the intuitive 'keep my work' move, but the branch had nothing to keep there.
**How to avoid:** take **theirs** (or explicit `origin/main`), then gate: `git diff origin/main -- internal/core/breaking_profiles.go` empty **and** `grep -q 'easyprotoc/protoc' internal/core/breaking_profiles.go` **and** blob equality.
**Warning signs:** `go mod tidy -diff` suddenly wants to remove `github.com/easyp-tech/easyprotoc`; `breaking_profiles.go` still imports `bufbuild/protocompile`.

### Pitfall 3: Choosing the wrong `Position` alias — compile-caught, but confusing mid-merge

**What goes wrong:** keeping main's `type Position = meta.Position` (or declaring both aliases) breaks compilation in `internal/rules` call sites that pass `protoast.Position` values to `core.AppendIssue`, or duplicates the declaration.
**Why it happens:** both aliases have identical field layouts (`Filename string; Offset int; Line int; Column int`), so the struct literal `Position{Filename: ..., Line: ..., Column: ...}` compiles under either alias; only the call sites reveal the type identity.
**How to avoid:** keep the branch's `dom.go` verbatim; verify with `grep -n 'type Position = protoast.Position' internal/core/dom.go` and the dom.go blob gate.
**Warning signs:** errors like `cannot use protoast.FilePosition(...) (value of type protoast.Position) as core.Position value in argument`; duplicate `Position` declaration.
`[VERIFIED: internal/protoast/protoast.go:19-24 and go-protoparser v4.14.2 parser/meta/position.go:8-17 have identical field sets; internal/rules call sites pass protoast values (protoast.FilePosition, imp.Pos, field.Pos, field.NamePos, val.Pos, enum.Pos)]`

### Pitfall 4: main's pre-existing lint finding is mistaken for a merge regression

**What goes wrong:** `golangci-lint run ./...` reports `internal/sourceview/session_test.go:29:21: Error return value of 'session.Close' is not checked (errcheck)`; the executor treats it as merge fallout.
**Why it happens:** the file is new to the branch (from #253) but **already fails on `origin/main` with the pinned v2.14.0 binary** — main's CI is red on this gate independent of the merge. The full scan on a pristine `origin/main` worktree yields exactly this 1 finding; the merged tree yields the identical single finding.
**How to avoid:** record as pre-existing/upstream; do not gate Phase 1 on `lint:go` (D-09 does not include it). If CI must be green for PR #258, fix it in a **separate** follow-up commit (Open Questions).
**Warning signs:** lint output references only `internal/sourceview/session_test.go`.
`[VERIFIED: golangci-lint 2.14.0 (Taskfile LINTER_VERSION) on worktree at origin/main → 1 issue; same on merged tree]`

### Pitfall 5: Editing main's new sourceview files to "fix" lint

**What goes wrong:** changing `session_test.go` inside the merge commit breaks the invariant "main wins for sourceview" and makes the merge commit non-pure.
**Why it happens:** the ergonomic temptation to make every local gate green in the merge commit.
**How to avoid:** keep the merge commit byte-faithful to `origin/main` for all sourceview files; any fix is a separate commit (or Phase 3, which owns seal/lint config edits).
**Warning signs:** `git diff origin/main -- internal/sourceview` non-empty.

### Pitfall 6: Running gates before committing (tests build the working tree)

**What goes wrong:** `TestAcceptanceGates`, `TestLintStdoutGolden`, `TestLintLinkMatrix*` call `buildEasypBinary`, which runs `go build ./cmd/easyp` from the repo root — i.e. the **working tree**, not a commit. Running them mid-resolution tests uncommitted code and pollutes evidence.
**Why it happens:** the tests are designed as subprocess gates and cache the build in `sync.OnceValues`.
**How to avoid:** run the battery after the merge commit, from a clean tree; record `git rev-parse HEAD` with the evidence.
**Warning signs:** gates pass pre-commit but fail after commit (or vice versa).

### Pitfall 7: Budgeting too little time for the full battery

**What goes wrong:** `go test -race -count=1 ./...` looks hung; `internal/api` alone takes ~309 s cold (acceptance gates build baseline+HEAD binaries and run four lint passes), default `-test.timeout` is 10 m.
**Why it happens:** per-package parallelism hides the heavy package until it finishes.
**How to avoid:** budget ~7 min for the battery; run the fast named gates separately for evidence; avoid re-running the whole battery more than once.
**Warning signs:** only `internal/adapters/*` lines appear for minutes.
`[VERIFIED: rehearsed full battery: 41 ok / 0 FAIL; internal/api 308.948s]`

### Pitfall 8: Expecting `task` to exist

**What goes wrong:** AGENTS.md/Taskfile references `task test`/`task lint`, but `task` is not installed in this environment; `docker` is absent too, so `task lint` (which includes hadolint) cannot pass.
**Why it happens:** the pinned helpers live in `bin/` (present) but the `task` binary does not.
**How to avoid:** run plain `go build`/`go test`/`bin/golangci-lint`; do not run `task lint` (Docker dependency). Use `bin/gotestsum` only if the exact `task test` shape is wanted.
**Warning signs:** `task: command not found`.
`[VERIFIED: command -v task/docker → missing; bin/golangci-lint 2.14.0, bin/gotestsum 1.13.0 present]`

## Code Examples

### 1. Pre-flight (clean tree, fetch, pins, dry-run conflict set)

```bash
cd /home/nil/DiskD/W/Djarvur/easyp

# 1a. Clean the index — the 156 staged planning renames block the merge.
#     Commit (recommended; commit_docs=true) or stash BOTH staged and unstaged planning edits:
git add -u .planning
git commit -m "docs: archive v1.0 phase artifacts; sync v1.1 state"
test -z "$(git status --porcelain --untracked-files=no)" || { echo "dirty tree"; exit 1; }

# 1b. Fetch and pin the expected upstream.
git fetch origin
test "$(git rev-parse origin/main)" = "5d36266d71e85f8f0fc3fa41f787a03ee5c68a79"

# 1c. Record the first-parent base for diff gates.
PRE_MERGE_BASE=$(git rev-parse HEAD)

# 1d. Dry-run: conflict set must be exactly these three files.
git merge-tree --write-tree HEAD origin/main | tail -n +2
# expected:
# internal/core/breaking_profiles.go
# internal/core/core.go
# internal/core/dom.go
```

### 2. Merge + resolution (exact, deterministic)

```bash
git merge --no-ff --no-commit origin/main   # expected exit 1, 3 "UU" paths

git checkout origin/main -- internal/core/breaking_profiles.go
git checkout "$PRE_MERGE_BASE" -- internal/core/dom.go
# core.go: remove the conflict markers in the import block, keeping both lines:
#   "github.com/easyp-tech/easyp/internal/protoast"
#   "github.com/easyp-tech/easyp/internal/sourceview"

git add internal/core/breaking_profiles.go internal/core/dom.go internal/core/core.go
```

### 3. Assert the resolution before building

```bash
# 3a. No conflict markers anywhere.
test -z "$(git grep -n -E '^(<<<<<<< |>>>>>>> )' -- '*.go')"

# 3b. Blob gates (commit-independent).
test "$(git rev-parse :internal/core/breaking_profiles.go)" = "$(git rev-parse origin/main:internal/core/breaking_profiles.go)"
test "$(git rev-parse :internal/core/dom.go)" = "$(git rev-parse "$PRE_MERGE_BASE":internal/core/dom.go)"
test "$(git rev-parse :internal/core/core.go)" = "34987d1824df1c4c14337df4b04c2625761e7b65"

# 3c. The merge must change exactly main's 9 surviving files (dom.go excluded).
#     This IS the complete freeze gate: any additional path is a resolution error.
git diff --cached --name-only | sort
# expected:
# go.mod
# go.sum
# internal/api/breaking_v1.go
# internal/core/breaking_profiles.go
# internal/core/core.go
# internal/sourceview/local.go
# internal/sourceview/root_tree.go
# internal/sourceview/session.go
# internal/sourceview/session_test.go

# 3d. Pins.
grep -q 'github.com/easyp-tech/easyprotoc v0.0.0-20261104120806-f378f3f3a05f' go.mod
grep -q 'google.golang.org/protobuf v1.37.0-rc.1' go.mod

# 3e. Pre-commit smoke build.
go build ./...
```

### 4. Commit + post-merge gates

```bash
git commit -m "merge: origin/main (#253 Use compiler for breaking check)" -m "..."
# Suggested body (rehearsed wording, CONTEXT leaves exact text at discretion):
#   Resolve internal/core semantically: main wins for breaking/sourceview (#253);
#   the branch lint contour stays byte-identical (dormant vendor path until Phase 2).
#
#   Conflicts resolved:
#   - internal/core/breaking_profiles.go: taken from origin/main (easyprotoc/protoc
#     compile path; the branch delta was only the Position spelling, subsumed by the
#     dom.go alias)
#   - internal/core/dom.go: kept from the branch (Position = protoast.Position, v1.0 D-06)
#   - internal/core/core.go: union of both sides (branch importCache + main sourceSession/SourceSession)

# Ancestry / aliveness (D-09, Success Criterion 1).
git merge-base --is-ancestor origin/main HEAD
test "$(git rev-list --count HEAD..origin/main)" -eq 0

# Merge commit shape.
git log -1 --format='%P'      # two parents: <PRE_MERGE_BASE> 5d36266...
git diff --name-only HEAD^1 HEAD   # the same 9 files as step 3c

# Lint contour must be byte-identical to the pre-merge branch.
git diff --stat "$PRE_MERGE_BASE" HEAD -- \
  internal/rules internal/easypc internal/protoast \
  internal/core/lint_linked.go internal/core/lint_linked_read.go \
  internal/core/lint_model.go internal/core/lint_suppressions.go \
  internal/core/link.go internal/core/proto_info_read.go internal/core/dom.go
# expected: empty

# Breaking file must equal main verbatim.
git diff origin/main -- internal/core/breaking_profiles.go
# expected: empty

# Full battery + named gates.
go build ./...
go test -race -count=1 ./...
go test -race -count=1 ./internal/api -run '^TestAcceptanceGates$' -v
go test -race -count=1 ./internal/api -run 'TestLintStdoutGolden|TestLintLinkMatrix$|TestLintLinkMatrixExitCodes|TestPrintIssuesCompileGolden' -v
go test -race -count=1 ./internal/rules -run 'TestLintGoldenFullRegistry|TestLintUseValidationParity' -v
go test -race -count=1 ./internal/core -run 'TestLintMatrix|TestLintErrorPathFidelity|TestLintSuppressionParity|TestProtoastPositionsMatchLegacyContract' -v
go test -race -count=1 ./internal/protoast -run 'TestPositionRuneColumns' -v

# Optional cheap gates.
go vet ./...
go mod tidy -diff        # expected: no output, rc 0
```

### 5. Rollback

```bash
# Before the merge commit exists:
git merge --abort

# After the commit, not yet pushed:
git reset --hard "$PRE_MERGE_BASE"      # discards the merge commit
# If pushed (not planned in this phase):
# git revert -m 1 <merge-sha>
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Breaking compiled with `bufbuild/protocompile` | `github.com/easyp-tech/easyprotoc/protoc.Compile` + custom `Resolver`, `protodesc.ToFileDescriptorProto` | #253, 2026-11-21 (`5d36266`) | Breaking path shares the compiler module that Phase 2 brings to lint; `protocompile` remains for generate/modules/migration |
| Per-open `os.Root` per local source | `sourceview.Session` + `rootTree` reused root/dir handles for one command; nil session falls back to package functions | #253 | Faster repeated opens; API layer owns session lifetime (`NewSession`/`Close` in `breaking_v1.go`) |
| `google.golang.org/protobuf v1.36.11` | `v1.37.0-rc.1` (required by easyprotoc) | #253 | MVS selects rc.1; `go mod tidy -diff` empty; vendor copy compiles against rc.1 (proven by battery) |
| Issues carried `meta.Position` | `Position = protoast.Position` alias (v1.0) with `fromMetaPos` boundary | v1.0 Phase 01 | main #253 now spells breaking positions as `Position{...}`; the alias keeps both sides compatible |

**Deprecated/outdated:** `.claude/CLAUDE.md`/`.spec` still say protobuf v1.36.11 and "protocompile is the heart of internal/core" — stale after this merge, owned by Phase 3 (SEAL/docs), not Phase 1.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `origin/main` will still be `5d36266` at execution time | Pre-flight | New upstream commits could add conflicts; plan asserts the SHA and re-runs the merge-tree check, aborting on mismatch |
| A2 | The pending `.planning` archive move (156 renames + STATE.md) should become a docs commit rather than be stashed | Pitfalls / pre-flight | If the user prefers stash, the pre-flight step changes but the merge recipe does not; commit_docs=true supports committing |
| A3 | PR #258's CI must end green, so the pre-existing errcheck finding needs a decision | Open Questions | If CI red is acceptable, no follow-up needed; if not, a separate fix commit is required |
| A4 | The phase does not push to `origin/feat/lint-easypc-infra` | Open Questions | If push is expected, add a final `git push` task; the merge itself is local |
| A5 | `go mod tidy` need not run in Phase 1 | Standard Stack | Evidence shows `tidy -diff` is already empty; if a later change makes it drift, tidy belongs to Phase 2/4 per IMPORT-DEP-01 |

## Open Questions

1. **Fix main's pre-existing errcheck finding in Phase 1?**
   - What we know: `origin/main` fails `golangci-lint run ./...` with exactly one issue at `internal/sourceview/session_test.go:29`; the merge adds none; Phase 1 gates (D-09) do not include lint.
   - What's unclear: whether PR #258's CI must be green for this phase or whether a red upstream finding is acceptable to carry.
   - Recommendation: keep the merge pure; add an optional follow-up commit `test(sourceview): check session.Close error` **after** the merge commit (or defer to Phase 3, which owns seal/lint config). Do not touch it inside the merge commit.

2. **Push the merge commit to `origin/feat/lint-easypc-infra`?**
   - What we know: the local branch is 5 commits ahead of its remote; the phase produces a 6th (the merge). CONTEXT/Roadmap do not mention push.
   - What's unclear: whether publishing the merge is part of "ветка не отстаёт от origin/main" for review purposes.
   - Recommendation: treat push as out of scope unless the user asks; the local ancestry/aliveness gates are the phase contract.

3. **Commit or stash the pending planning-artifact move?**
   - What we know: merge refuses with a dirty index; the artifacts are legitimate GSD archive work.
   - What's unclear: whether the user wants those renames committed now.
   - Recommendation: commit with a `docs:` prefix (A2); fallback is `git stash push` + `git stash pop` after the merge.

4. **What if `origin/main` moved past `5d36266`?**
   - What we know: remote currently equals local tracking `5d36266`; the recipe is proven against it.
   - What's unclear: future upstream commits.
   - Recommendation: abort and re-research if the SHA differs; do not improvise resolutions against a new conflict set.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | build/test | ✓ | go1.27.3 (go.mod declares `go 1.26.6`; newer toolchain satisfies it) | — |
| git | merge, ancestry gates, acceptance tests | ✓ | 2.43.0 | — |
| Network to GitHub (SSH) | `git fetch`, `git ls-remote` | ✓ | origin `git@github.com:easyp-tech/easyp.git`; `ls-remote main` = `5d36266` | — |
| Network to Go proxy | download easyprotoc/protobuf | ✓ | proxy.golang.org; easyprotoc downloaded in ~10 s | — |
| `github.com/easyp-tech/easyprotoc` module | build/test | ✓ (fetched this session; was not in cache before) | `v0.0.0-20261104120806-f378f3f3a05f` | — |
| `google.golang.org/protobuf` rc.1 module | build/test | ✓ (already in cache) | `v1.37.0-rc.1` | — |
| `golangci-lint` | optional `lint:go` | ✓ | `bin/golangci-lint` v2.14.0 (matches Taskfile pin) | plain `go vet` |
| `gotestsum` | optional task-equivalent test run | ✓ | `bin/gotestsum` v1.13.0 | `go test -race -count=1 ./...` |
| `task` (Taskfile runner) | `task test/lint/schema:check` | ✗ | — | run the underlying `go`/`bin/` commands directly |
| Docker / Hadolint | `task lint:docker` | ✗ | — | not a Phase 1 gate; skip (`hadolint` remains a v1.0 deferred item) |

**Missing dependencies with no fallback:** none — every Phase 1 gate has a direct `go`/`git` equivalent.
**Missing dependencies with fallback:** `task` (use `go`/`bin/*`), Docker/Hadolint (out of phase scope).

## Validation Architecture

> `workflow.nyquist_validation` is `true` in `.planning/config.json`, so this section is included.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go `testing` + `github.com/stretchr/testify` v1.11.1 |
| Config file | none — package-level `go test`; race/coverage flags come from Taskfile |
| Quick run command | `go test -count=1 ./internal/core ./internal/rules ./internal/protoast` (≈5 s) |
| Full suite command | `go test -race -count=1 ./...` (≈7 min; `internal/api` 308.9 s cold) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|--------------|
| IMPORT-REC-01 | Merge applied; branch not behind | integration (shell) | `git merge-base --is-ancestor origin/main HEAD`; `test "$(git rev-list --count HEAD..origin/main)" -eq 0` | ✅ (plan tasks) |
| IMPORT-REC-01 | Conflicts resolved semantically (main breaking, branch lint) | integration (shell) | blob gates in §Code Examples step 3b; `git diff origin/main -- internal/core/breaking_profiles.go` empty | ✅ (plan tasks) |
| IMPORT-REC-01 | Breaking on `easyprotoc/protoc`, pins present | unit (module) | `go list -m github.com/easyp-tech/easyprotoc google.golang.org/protobuf`; `grep -q easyprotoc/protoc internal/core/breaking_profiles.go` | ✅ |
| IMPORT-REC-01 | Full battery green | unit/integration | `go build ./...`; `go test -race -count=1 ./...` | ✅ |
| IMPORT-REC-01 | Lint parity not shifted (Gates A–D ×2) | e2e subprocess | `go test -race -count=1 ./internal/api -run '^TestAcceptanceGates$' -v` (`acceptanceBaselineCommit = "1e9a596"`) | ✅ `internal/api/acceptance_gates_test.go:284` |
| IMPORT-REC-01 | Golden 94 findings byte-exact | unit | `go test -race -count=1 ./internal/rules -run '^TestLintGoldenFullRegistry$' -v` (`goldenFindingCount = 94`) | ✅ `internal/rules/golden_full_registry_test.go:529` |
| IMPORT-REC-01 | D-07 positions | unit | `go test -race -count=1 ./internal/protoast -run '^TestPositionRuneColumns$' -v`; `go test -race -count=1 ./internal/core -run '^TestProtoastPositionsMatchLegacyContract$' -v` | ✅ `internal/protoast/position_test.go:22`, `internal/core/position_parity_test.go:21` |
| IMPORT-REC-01 | Suppression/matrix parity | unit | `go test -race -count=1 ./internal/core -run 'TestLintMatrix|TestLintErrorPathFidelity|TestLintSuppressionParity' -v` | ✅ `internal/core/lint_linked_matrix_test.go:69,359`, `internal/core/lint_suppressions_parity_test.go:77` |
| IMPORT-REC-01 | CLI stdout/JSON shape | e2e subprocess | `go test -race -count=1 ./internal/api -run 'TestLintStdoutGolden|TestLintLinkMatrix|TestPrintIssuesCompileGolden' -v` | ✅ `internal/api/lint_stdout_golden_test.go:124`, `internal/api/lint_link_matrix_test.go:67,204,292` |

### Sampling Rate

- **Per task commit:** `go build ./...` + quick package tests (`internal/core`, `internal/rules`, `internal/protoast`).
- **Per wave merge:** not applicable (single merge commit); use the full battery once after the commit.
- **Phase gate:** `go build ./...`, full `-race` battery green, named golden/acceptance/position gates green, ancestry/aliveness true, none of the lint-contour paths changed.

### Wave 0 Gaps

- None — existing test infrastructure covers every phase requirement. No new Go tests are required; the shell-level blob/ancestry gates live in the plan's task verify blocks.

## Security Domain

> `security_enforcement: true`, ASVS level 1. The phase changes no user-facing input handling; the security-relevant surface is dependency/supply-chain integrity and merge correctness.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | — |
| V3 Session Management | no | `sourceview.Session` is an in-process file-handle cache, not a user session; no auth semantics |
| V4 Access Control | no | Path-bounded opens are enforced by existing `sourceview` checks; the merge keeps them (main's code) |
| V5 Input Validation | no | No new inputs or parsers are introduced by the merge; the compiler is a trusted pinned module |
| V6 Cryptography | no | No crypto; `go.sum` hashes are the integrity mechanism |
| V14 Configuration / supply chain (informative) | yes | Pin exact pseudo-version + VCS origin verified via Go proxy; `go.sum` committed; no `replace`; no `vendor` |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Dependency / module substitution | Tampering | Exact pseudo-version pinned in `go.mod`, hash in `go.sum`, proxy origin resolved to `github.com/easyp-tech/easyprotoc` @ `f378f3f...` |
| Silent behavior regression during conflict resolution | Tampering/Repudiation | Blob-SHA gates + `git diff origin/main` + golden/acceptance gates (not just green compile) |
| Accidental commit of untracked local artifacts | Information disclosure | Explicit pathspecs; never `git add -A` in the merge; check `git diff --cached --name-only` = expected 9 |
| Unverified upstream drift between research and execution | Tampering | Assert `origin/main` SHA `5d36266`; abort and re-research on mismatch |

## Project Constraints (from AGENTS.md / .claude/CLAUDE.md)

No `CLAUDE.md` exists at the repo root; `.claude/CLAUDE.md` exists (GSD-generated mirror) and `AGENTS.md` is the declared source of truth. Directives that bear on this phase:

- **Build/test/lint:** prefer Task targets, but the underlying commands are `go build -o easyp ./cmd/easyp`, `go test -race -count=1 ./...`, `bin/golangci-lint run ./...`, `go run ./cmd/easyp schema-gen` (Taskfile `LINTER_VERSION: "2.14.0"`). Task binary is absent; run equivalents directly.
- **Generated schema:** never hand-edit `schemas/*.json`; regenerate via `task schema:generate`/`go run ./cmd/easyp schema-gen`. Phase 1 must not change schemas (verified drift-free).
- **Sealed vendor copy:** do not edit anything inside `internal/easypc` (D-06). Fixes go upstream and arrive by re-sync.
- **Commits:** no secrets/credentials/unrelated local artifacts; do not commit `coverage.out`, the root `easyp` binary, or generated dependency noise. Untracked `.planning/research/.cache` must stay out of the merge commit.
- **Go conventions:** error wrapping is `fmt.Errorf("<callee>: %w", err)`; never bare `defer x.Close()` (the pre-existing errcheck finding is exactly this rule in main's test); testify with named cases; tests use `-race`.
- **Planning docs:** do not add parallel agent instruction files; keep `AGENTS.md` + `.spec/` authoritative. Docs describing the merge (stack pins, "protocompile is the heart") are stale but belong to Phase 3.
- **`.spec/agent-rules.md`:** after behavior changes run tests for the touched packages; after schema changes run `task schema:check` (not needed here).

## Sources

### Primary (HIGH confidence)

- Git object database of `/home/nil/DiskD/W/Djarvur/easyp`: `git show 5d36266`, `git diff 5c9185f..HEAD`, `git merge-tree --write-tree HEAD origin/main`, `git diff origin/main --`, `git rev-parse origin/main:<path>` — exact merge content, conflict set, resolved blobs, parents.
- Rehearsed merge in throwaway git worktrees at `96b8c6a` + `origin/main`: two independent reproductions of the three resolution blobs; `go build ./...`; `go test -race -count=1 ./...` (41 ok / 0 FAIL, `internal/api` 308.948 s); `go vet ./...`; `go mod tidy -diff` (empty); `go run ./cmd/easyp schema-gen` + `diff -r schemas` (no drift).
- Go module proxy metadata: `go list -m -json github.com/easyp-tech/easyprotoc@v0.0.0-20261104120806-f378f3f3a05f` → Origin URL/VCS hash `f378f3f3a05febd98f54aff2be9a570de239a59a`; easyprotoc `go.mod` requires `google.golang.org/protobuf v1.37.0-rc.1`; `go list -m` after merge → `v1.37.0-rc.1`.
- Module cache source: `.../easyprotoc@v0.0.0-.../protoc/compile.go:22` (`func Compile(ctx context.Context, files []string, opts ...Option) (*Result, error)`), `protoc/resolver.go:14` (`ErrFileNotFound`), `protoc/result.go:52` (`Files() []protoreflect.FileDescriptor`), `protoc/options.go:28,44,51,59,73` (option constructors).
- Repository source opened this session: `internal/core/dom.go:15-19` (`type Position = protoast.Position`), `internal/core/core.go:45-55,63-120` (branch `importCache`/`parsedImports`/`Options`), `internal/core/link.go:86` (`Position: Position{Line: d.Line, Column: d.Col}`), `go.mod:1-25` (pre-merge pins), `internal/protoast/protoast.go:19-36`, `internal/api/acceptance_gates_test.go:29,70,138-180,284`, `internal/api/lint_link_matrix_test.go:159`, `internal/rules/golden_full_registry_test.go:60,529`, `internal/core/position_parity_test.go:21`, `internal/protoast/position_test.go:22`.
- Empirical environment/tooling checks: `git ls-remote origin main` = `5d36266...`; `command -v task/docker` missing; `bin/golangci-lint --version` 2.14.0; `bin/gotestsum` present; scratch-repo merge experiments (dirty index refusal, staged-rename refusal, unstaged-change acceptance).

### Secondary (MEDIUM confidence)

- `[CITED: https://go.dev/ref/mod]` — SemVer ordering ("Pre-release versions sort before the corresponding release versions… v1.36.11 < v1.37.0-rc.1 < v1.37.0") and MVS selects the highest **required** version (easyprotoc requires rc.1 → rc.1 selected).
- `.spec/ARCHITECTURE.md`, `.spec/agent-rules.md`, `.github/workflows/tests.yml` — layering, conventions, CI command sequence (`task init` → `task lint:go` → cross-build → `task test` → `proto:check` → `dev-tools:check`).
- `.planning/PROJECT.md`, `.planning/milestones/v1.0-*` — v1.0 baseline (41 packages/3718 tests, D-06/D-07 invariants, main-wins precedent).

### Tertiary (LOW confidence)

- None. All claims above were verified by tool output in this session or cited from official docs.

## Metadata

**Confidence breakdown:**

- Standard stack: HIGH — pins read from `origin/main`'s `go.mod`/`go.sum`, module provenance from the Go proxy, API symbols read from the module cache, MVS selection reproduced with `go list -m`.
- Architecture: HIGH — the merge was rehearsed twice with identical resolution blobs; the resulting tree built and passed the full `-race` battery, acceptance gates and golden/position gates.
- Pitfalls: HIGH — dirty-index refusal, pre-existing lint finding, and silent `--ours` failure were reproduced; the working-tree build behavior was read at source (`build.Dir = repoRoot`).

**Research date:** 2026-11-21
**Valid until:** until `origin/main` moves past `5d36266` or the branch's pre-merge base changes; re-validate the conflict set with `git merge-tree` before execution.
