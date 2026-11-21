# Phase 1: Вендоринг компилятора и расчленение модели - Research

**Researched:** 2026-11-17
**Domain:** Go in-repo compiler vendoring (copy-in, zero new deps) + lint/breaking read-model split + position-contract proof
**Confidence:** HIGH (nearly every load-bearing claim was empirically executed or read verbatim this session: throwaway copy-in build/tidy/test run, golangci-lint run, both parsers' position code read from source, lab repo cloned and diffed across branches)

## Summary

Phase 1 is a copy-in vendoring plus a type-level model split, and it is **pre-verified end to end in this session**: `internal/easypc` + `internal/protoast` were copied from `compiler-lab` (main @ `08cef10`) into a throwaway copy of this repo; `go build ./...` passed, `go mod tidy` left `go.mod`/`go.sum` byte-identical, and every vendored package's own tests passed. The import paths in the lab copy already point at `github.com/easyp-tech/easyp/internal/easypc/...`, so the copy is mechanical with **zero import rewriting**.

Three findings materially change what the plan must contain:

1. **"Ветка `easyp-cli`" не существует.** `easyp-cli` is a *directory* on compiler-lab's `main` branch (its history was merged under `easyp-cli/` in `08cef10`, 2026-11-04). The newest branch `compile_from_internal` (`eaa9a1a3d225`, 2026-11-17) moved the compiler OUT to a standalone module `github.com/easyp-tech/compiler-lab/easypc` with re-rewritten imports — content is otherwise identical (verified: 6 files differ only in gofmt realignment after the path rewrite). The zero-rewrite source of truth is **main @ `08cef10909c0722f7042ba17b5c11951f9c6b6f6`, directory `easyp-cli/`**. VENDORED.md must record exactly that.
2. **The position contract does NOT hold as vendored — and that is the point of the D-07 fixture.** protoast/easypc columns are **1-based, counted in bytes, tab advances to next multiple of 8** (protoc convention, pinned by the lab's own `protoast_test.go`); main's go-protoparser columns are **1-based, counted in runes, tab counts as 1** (both read verbatim from source). On any line with multi-byte UTF-8 before the anchor (or a tab), they diverge. Success criterion 4 therefore requires a **position adaptation inside `internal/protoast`** (allowed — D-04 seals only `internal/easypc`), plus updating the lab's tab-8 test expectation. This deviates from "pure addition" for protoast and must be a separate commit after the vendor commit.
3. **`task lint` will be red on the vendored tree without a path exclusion.** golangci-lint v2.14.0 (the pinned version, same config as the lab) reports **22 issues** on the vendored copy (21 in `internal/easypc` — errcheck/ineffassign/staticcheck QF/ST1005 — and 1 QF1002 in `internal/protoast/build.go:242`). D-04 forbids fixing code inside `internal/easypc`, so a `linters.exclusions.paths: [internal/easypc]` entry is unavoidable (empirically verified to reduce 22 → 1). This is vendor-tree hygiene, distinct from D-02's "no exclusions for helpers".

The model split (D-06) has a quantified ripple: `Issue.Position` → `protoast.Position` touches 5 non-test files (2 rules construct literals, breaking constructs in 2 sites), 49 `meta.Position` literals across 41 rule test files + `breaking_checker_test.go`, and `AppendIssue`'s signature. The smallest-diff mechanics: keep the name `ProtoInfo` for the **lint** type (rules untouched), give breaking its own renamed type/files, and convert `meta.Position` → `protoast.Position` inside `AppendIssue`/`buildBreakingCheckIssue` (single conversion point). `lint_helpers.go` (D-01) does **not compile as-in-the-lab** against main's types — 3 of its declarations consume the lab's `core.ProtoInfo.File` and 4-arg `AppendIssue`; the plan must adapt those signatures (see Open Questions).

**Primary recommendation:** Vendor from compiler-lab **main @ `08cef10`**, dir `easyp-cli/`, as a standalone pure-addition commit (empirically known-green); add `linters.exclusions.paths: [internal/easypc]` in the same commit; then land the position-contract work (rune columns in `internal/protoast` + two-level D-07 tests + tab-expectation update) as a reviewed follow-up commit; then the model split with lint keeping the `ProtoInfo` name and breaking receiving renamed ownership; run the full `go test -race -count=1 ./...` and the `go mod tidy` zero-diff gate at every commit.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Вендоринг хелперов**
- **D-01:** `lint_helpers.go` и `casing.go` переносятся в Phase 1 ЦЕЛИКОМ, как в лаборатории (не подмножество, не отложено в Phase 4). Обоснование: будущий синк с компилятором тривиален, дифф при переносе правил в Phase 4 минимален. Допустимый временный «мёртвый» код до Phase 4.
- **D-02:** Исключения в golangci-lint для хелперов НЕ настраиваются. Экспортируемые функции не должны триггерить линтер; если `task lint` всё же упадёт — исполнитель решает по факту (минимальное точечное исключение).

**Гигиена вендор-дерева**
- **D-03:** Провенанс фиксируется в `internal/easypc/VENDORED.md`: upstream репозиторий (compiler-lab), ветка (`easyp-cli`), SHA, дата вкопирования, процедура обновления (синка). — **Reversibility:** reversible — один файл, удаляется/обновляется тривиально.
- **D-04:** Правки внутри `internal/easypc` ЗАПРЕЩЕНЫ, кроме единственного исключения: переписывание импорт-путей на `github.com/easyp-tech/easyp/internal/...` один раз при вкопировании. Любые фиксы компилятора идут вверх по потоку (compiler-lab) и приходят синком. Запрет отражается в AGENTS.md и VENDORED.md. — **Reversibility:** costly — правило предотвращает drift вендоренной копии; отмена ломает трассируемость синков.
- **D-05:** Собственные тесты easypc ВКЛЮЧАЮТСЯ в `task test` как обычные пакеты; `testdata/` остаётся в репозитории (нужна этим тестам).

**Механика расщепления модели**
- **D-06:** Полное расщепление сразу: lint получает собственные типы модели чтения, breaking забирает legacy `protoInfoRead`/`ProtoInfo` в единоличное владение. Без общего типа/alias — чистая граница, Phase 3/5 не тянут наследие. — **Reversibility:** costly — откат затрагивает все места конструирования в `internal/core`.
- **D-07:** Контракт позиций доказывается ДВУМЯ уровнями тестов: (а) юнит-тест на конверсию спанов компилятора → `protoast.Position` (1-based строка, 1-based колонка в рунах); (б) golden-тест сквозного пути: proto-файл с многобайтовым UTF-8 (кириллица/эмодзи в комментариях и строках) → замечания текущего линтера vs позиции protoast. Ловит off-by-one и byte-vs-rune до переноса правил.

**Эталонная ревизия**
- **D-08:** Вендоринг выполняется с ЗАФИКСИРОВАННОГО SHA ветки `easyp-cli` (определяется при планировании/исполнении, записывается в VENDORED.md до вкопирования). Новые синки — только осознанно, с новым SHA.
- **D-09:** Доступ executor'а к compiler-lab — через `gh repo clone` с этой машины (gh-аутентификация уже работает, explore-исследование клонировало так же). Не риск, стандартный путь.

### Claude's Discretion
- Точная механика переименования/расщепления типов в `internal/core` (какие имена, где границы файлов) — planner решает по фактическому коду.
- Выбор конкретных UTF-8 фикстур и их placement (testdata-структура).
- Формат VENDORED.md (минимально достаточный: repo/branch/SHA/date/procedure).

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| LINT-VEND-01 (REQ-1.1) | WHEN исходники фичи собраны, the system SHALL содержать `easypc` (компилятор целиком, включая `cmd/` и `testdata/`) и `protoast`, вкопированные из `compiler-lab/easyp-cli` с импортами на `github.com/easyp-tech/easyp/internal/...` | Vendoring procedure empirically executed this session: copy from compiler-lab main@08cef10 dir `easyp-cli/` → `go build ./...` green, all vendored package tests green. Import paths already match main's module layout — no rewrite. Sizes/counts and the branch-name correction documented below. |
| LINT-VEND-02 (REQ-1.2) | WHEN выполнен `go mod tidy`, the system SHALL не содержать новых внешних зависимостей для компилятора (без внешнего модуля и без `replace`) | Empirically executed: `go mod tidy` on the vendored copy left `go.mod` and `go.sum` byte-identical. Import scan confirms the vendored trees' only external imports are `google.golang.org/protobuf` (5 subpackages) and `testify` (protoast test), both direct requires of main's go.mod. |
</phase_requirements>

## Project Constraints (from CLAUDE.md / AGENTS.md)

`.claude/CLAUDE.md` (GSD-managed) and `AGENTS.md` carry these actionable directives for Phase 1:

- **No new dependencies for the compiler** (REQ-1.2) — vendoring into `internal/`; nothing may add a `go.mod` entry.
- **Colocated tests**: `<file>.go` + `<file>_test.go`, testify (`require` fatal, `assert` non-fatal), named table-test cases, `t.Parallel()` where stateless (`.spec/agent-rules.md:18,31-35`).
- **Error wrapping**: `fmt.Errorf("<callee>: %w", err)` — callee name only, no package/receiver prefix.
- **No hand-edits to generated `schemas/*.json`** — not touched in this phase.
- **GSD workflow enforcement**: repo edits go through GSD commands.
- **AGENTS.md is the agent-instruction source of truth** — Phase 1 itself edits it (D-04 prohibition note); do not create parallel agent files.
- **Test/lint gates**: `task test` (gotestsum, `-race -count=1`, coverage), `task lint` (golangci-lint v2.14.0 + hadolint-in-Docker), `task build`.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Compiler (lexer/parser/linker/sourceinfo/serialize/WKT) | `internal/easypc` (sealed vendor tree) | — | Self-contained; imports nothing from easyp, only stdlib + `google.golang.org/protobuf`. Sealed so compiler-lab syncs stay mechanical (D-04). |
| Lint view model (AST, positions, comments, walk) | `internal/protoast` | — | The adaptation seam: easyp-specific position contract (rune columns) lives here, NOT in easypc. Depends only on `easypc/pkg/parser` + `easypc/protoc/syntax`. |
| Legacy read model (`protoInfoRead`/`ProtoInfo`, go-protoparser) | `internal/core` breaking-owned files | — | After the split, breaking is sole owner (D-06); lint never touches it. |
| Lint read model (until Phase 4: still go-protoparser; type separate) | `internal/core` lint-owned files | — | Same parser, different type — the split is ownership/type-level now, parser-level in Phase 3/5. |
| `Issue`/`IssueInfo` output type + `Position` | `internal/core` (dom) | `internal/api` (printing only) | Single output type keeps CLI format layer untouched; `Position` becomes `protoast.Position` (lab precedent: type alias). |
| Output formatting / exit codes | `internal/api` (lint.go, breaking_v1.go) | — | Reads only `.Position.Line`/`.Column` by field name — type-swap safe (verified by grep: no other consumers). |
| Engine wiring | `internal/api/runtime.go` (`buildCore`) | — | Passes through unchanged; Phase 1 must not need new Options knobs. |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `internal/easypc` (vendored from compiler-lab main @ `08cef10`, dir `easyp-cli/`) | copy-in; 48 .go files (15 tests), 3.4 MB total (testdata 2.6 MB, embedded WKT 204 KB / 20 .proto) | The compiler: `pkg/{compiler,lexer,linker,metrics,options,parser,serialize,sourceinfo}`, `protoc` (+`syntax`, `wkt` via `go:embed`), `cmd/{easyp-protoc,lexbench}`, `testdata/` | Already validated in the lab; imports already `github.com/easyp-tech/easyp/internal/easypc/...` — zero rewrite. Only external import: `google.golang.org/protobuf` (5 subpackages), already a direct require. [VERIFIED: import scan + empirical copy-in build] |
| `internal/protoast` (vendored same source) | copy-in; 6 .go files (1 test), 48 KB | Lint view model: `protoast.go` (types+Position), `build.go` (span→model), `comments.go` (attachment), `walk.go` (Visitor), `outline.go` | Depends only on `easypc/pkg/parser` + `easypc/protoc/syntax` (verified). The deliberate adaptation point for the position contract. |
| `google.golang.org/protobuf` | v1.36.11 (already in go.mod) | The vendored compiler's ONLY external dependency (descriptorpb/pluginpb/protodesc/protoreflect/proto) | Present and used by main; no version change. [VERIFIED: main go.mod + lab import scan] |
| Go toolchain | go.mod `go 1.26.6`; local toolchain go1.27.3 | Build/test | Toolchain newer than directive is fine; CI uses `go-version-file: go.mod`. [VERIFIED: local build green] |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/stretchr/testify` | v1.11.1 (already in go.mod) | Vendored `protoast_test.go` + new D-07 tests | All new tests follow `.spec/agent-rules.md` (require/assert). |
| golangci-lint | v2.14.0 (Taskfile pin, `bin/`) | Lint gate | Needs `linters.exclusions.paths: [internal/easypc]` (empirically verified required, see Pitfall 3). |
| gotestsum | v1.13.0 (Taskfile pin, `bin/`) | `task test` runner | Provision via `task init`; raw `go test -race -count=1 ./...` is the fallback (verified working). |

### Alternatives Considered

Rejected by project research (`.planning/research/STACK.md`) — do not revisit in this phase:

| Instead of (copy-in) | Could Use | Tradeoff |
|----------------------|-----------|----------|
| Copy-in under `internal/` | Separate module + `replace` | Forbidden by REQ-1.2 (new module edge); unnecessary — imports already match. |
| Copy-in | `git subtree` / `git submodule` | No foreign upstream to track; the lab IS easyp-owned; subtree pulls add merge noise. |
| Copy-in | `go mod vendor` | Mechanism caches declared module deps; does not own forked code. |

**Installation:**

```bash
# Nothing installs from any registry. The entire procedure:
gh repo clone easyp-tech/compiler-lab /tmp/compiler-lab   # default branch main
cd /tmp/compiler-lab && git checkout 08cef10909c0722f7042ba17b5c11951f9c6b6f
cp -r easyp-cli/internal/easypc   <repo>/internal/easypc
cp -r easyp-cli/internal/protoast <repo>/internal/protoast
# no import rewriting needed — paths already github.com/easyp-tech/easyp/internal/...
go build ./... && go mod tidy && git diff --exit-code go.mod go.sum   # expect empty
```

**Version verification:** performed empirically this session on a throwaway copy of main: `go build ./...` exit 0; `go mod tidy` → `go.mod`/`go.sum` no diff; `go test ./internal/easypc/... ./internal/protoast/...` all `ok` (fastest package 0.002s, slowest `protoc` 0.725s — no runtime concern for `task test`).

## Package Legitimacy Audit

This phase installs **zero external packages** — the defining property of REQ-1.2. The gate is therefore not triggered; no registry lookups were performed or needed. The vendored trees' complete external import set (verified by grep over all .go files):

- `internal/easypc`: stdlib + `google.golang.org/protobuf/{proto, reflect/protodesc, reflect/protoreflect, types/descriptorpb, types/pluginpb}` — all provided by the existing `google.golang.org/protobuf v1.36.11` require. `cmd/*` import only internal packages.
- `internal/protoast`: stdlib + the two easypc packages above; `protoast_test.go` adds `testify` (existing require).

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| (none — no new packages) | — | — | — | — | — | N/A |

## Architecture Patterns

### System Architecture Diagram

State after Phase 1 (lint still reads go-protoparser at runtime; types split; compiler vendored but unwired):

```
                     gh repo clone compiler-lab (main @ 08cef10, dir easyp-cli/)
                                   │  copy-in, zero import rewrite
                                   ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ easyp repo (module github.com/easyp-tech/easyp)                         │
│                                                                         │
│  internal/easypc          [SEALED VENDOR TREE, D-04]                    │
│    pkg/lexer → pkg/parser → pkg/linker → pkg/{options,sourceinfo,...}   │
│    protoc (+syntax, +wkt go:embed)   cmd/{easyp-protoc, lexbench}       │
│    testdata/ (2.6 MB, consumed by its own tests, D-05)                  │
│         ▲                                                               │
│         │ imports only easypc/pkg/parser + easypc/protoc/syntax         │
│  internal/protoast         [ADAPTATION SEAM — position contract lives  │
│         │                  here: rune columns, tabs = 1]                │
│         ▲  (Phase 1: imported only by tests + core.Position alias;      │
│         │   NOT wired into any runtime path yet)                        │
│  internal/core                                                            │
│    lint-owned files:  lint read fn → lint ProtoInfo (go-protoparser     │
│                      fields, type owned by lint) → rules                 │
│    breaking-owned files: protoInfoRead → breaking ProtoInfo (renamed)   │
│                      → collect → ProtoData → BreakingChecker            │
│    dom.go: Issue.Position = protoast.Position  (shared output type;     │
│             meta.Position → protoast.Position conversion at the         │
│             AppendIssue/buildBreakingCheckIssue boundary)                │
│  internal/api: text/JSON printers read .Line/.Column only — unchanged   │
└─────────────────────────────────────────────────────────────────────────┘
```

No runtime behavior flows through easypc/protoast in Phase 1 — they are compiled and tested, not wired. The first runtime wiring is Phase 2/3.

### Recommended Project Structure (delta only)

```
internal/
├── easypc/                  # vendored whole (pkg/, protoc/, cmd/, testdata/)
│   └── VENDORED.md          # NEW: provenance + sync procedure (D-03)
├── protoast/                # vendored (6 files)
│   ├── build.go             # position adaptation lands here (b.pos())
│   └── position_test.go     # NEW: D-07(a) unit test (colocated)
├── core/
│   ├── dom.go               # EDIT: Issue.Position → protoast.Position;
│   │                        #   lint/breaking type split (D-06)
│   ├── lint.go              # EDIT: calls lint-owned read, not protoInfoRead
│   ├── proto_info_read.go   # MOVE/RENAME: breaking-owned (sole owner)
│   ├── breaking_check.go    # stays; consumer of breaking-owned read
│   └── position_parity_test.go  # NEW: D-07(b) golden test + fixture
├── rules/
│   ├── casing.go            # NEW: vendored verbatim (stdlib-only)
│   ├── casing_test.go       # NEW: vendored verbatim (testify)
│   └── lint_helpers.go      # NEW: vendored with minimal seam adaptation
│                            #   (see Open Question 1)
└── (test fixtures: UTF-8+tab .proto files — placement is discretion)
```

### Pattern 1: Standalone vendor commit, then adaptation commits

**What:** Land the vendored trees as one commit containing ONLY the copy + `.golangci.yml` path exclusion + `VENDORED.md` + AGENTS.md note (D-04). The protoast position adaptation and the model split are separate, individually reviewable commits.

**When to use:** Always for vendor drops (P9). The pure-addition property (`go mod tidy` zero-diff, `go build` green, tests green) is provable only on a commit that mixes in nothing else.

### Pattern 2: Single conversion point for positions

**What:** Keep `core.AppendIssue` accepting `meta.Position` (or convert at its top) so the 48 `.Meta.Pos` pass-through call sites in `internal/rules` need zero changes; breaking converts in `buildBreakingCheckIssue`. `Issue.Position` becomes `protoast.Position` (type alias `type Position = protoast.Position`, as the lab did).

**When to use:** Phase 1's split is type-level; the physical sources of positions (go-protoparser `Meta.Pos`) don't change until Phase 4/5. One conversion point beats 50 scattered ones.

**Example:**
```go
// internal/core/dom.go (mechanics indicative — final naming is planner's discretion per CONTEXT)
type Position = protoast.Position // matches lab: lab internal/core/dom.go:14

func fromMetaPos(p meta.Position) Position {
	return Position{Filename: p.Filename, Offset: p.Offset, Line: p.Line, Column: p.Column}
}
```
Field sets are identical in name, type, and order (`Filename string; Offset int; Line int; Column int`) — [VERIFIED: `internal/core/dom.go` import of `meta` + go-protoparser `parser/meta/position.go` struct vs `internal/protoast/protoast.go:17-22`], so the conversion is a field-for-field copy and the JSON output form is unchanged (lab comment: "Its JSON form matches earlier releases", lab dom.go:13).

### Pattern 3: Two-level position-contract proof (D-07)

**What:**
- **(a) Unit level** — `internal/protoast/position_test.go`: `protoast.Parse` on a fixture with (i) Cyrillic/emoji comments and string literals before anchors, (ii) a tab-indented declaration; assert exact `Position{Line, Column}` values computed by hand as 1-based runes.
- **(b) Golden level** — `internal/core` test: parse the same fixture with BOTH `readProtoFile` (go-protoparser — the thing whose positions current lint issues carry) and `protoast.Read`; assert anchor-by-anchor equality of positions (package statement, message/field/enum/enum-value/service/rpc first-token anchors). Model-level comparison is equivalent to "замечания текущего линтера vs позиции protoast" because current rule positions ARE go-protoparser `Meta.Pos` values of exactly these anchors — and it runs in milliseconds without CLI/config scaffolding.

**When to use:** Before ANY rule porting (Phase 4) — this is the phase's reason for existing (P1 detection-before-propagation).

### Anti-Patterns to Avoid

- **Fixing lint findings inside `internal/easypc`** — violates D-04; the next sync clobbers it. Use the path exclusion instead.
- **Wiring protoast into the lint runtime path in Phase 1** (e.g., populating a `File *protoast.File` field during lint reads "to help Phase 4") — changes error surfaces (protoast parse failures differ from go-protoparser's) and breaks the zero-behavior-diff criterion. Phase 1 types are split; paths are not.
- **Adapting positions by post-processing `Issue`s** — the contract belongs in `protoast.Position` construction (D-07 tests target protoast directly).
- **Excluding `internal/protoast` from golangci** to silence its single QF1002 — protoast is easyp-maintained seam code and must stay lint-clean; fix the tagged switch (it can ride along with the position adaptation edit in `build.go`).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Compiler sync provenance | Memory / commit messages | `internal/easypc/VENDORED.md` (D-03) + standalone vendor commit | Copy has no module-system link to origin; provenance is the only drift guard (P9). |
| Position conversion | Per-rule conversion shims | Single `fromMetaPos`/`AppendIssue` boundary + protoast-level contract | 43 rules × shims = permanent off-by-one surface (P1). |
| Column computation | Byte-counting with tab-8 in new code | protoast `pos()` computing rune columns from `b.ast.Src` | The contract is main's rune-column behavior; implement once at the seam. |
| Lint-noise suppression in vendor tree | Inline `//nolint` comments inside easypc | `linters.exclusions.paths` in `.golangci.yml` | D-04 forbids editing files inside easypc, nolint comments included. |
| Test runner/linter provisioning | Ad-hoc global installs | `task init` (pins gotestsum 1.13.0, golangci-lint 2.14.0, mockery 2.53.7 into `bin/`) | Pinned versions match CI; local `bin/` is currently EMPTY (see Environment). |

**Key insight:** every hand-rolled substitute for these either violates a locked decision (D-04) or re-creates a validated lab mechanism with fresh bug surface.

## Runtime State Inventory

> Phase involves type renames/moves within one repo. Categories checked explicitly:

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — no databases/caches keyed on these type names. | None. |
| Live service config | None — CLI tool; no external services configured from these names. | None. |
| OS-registered state | None — no scheduled tasks/launchd/systemd referencing the repo. | None. |
| Secrets/env vars | None — `EASYP_PROXY`, `EASYP_GIT_TIMEOUT`, `EASYPPATH` are unrelated to renamed types; untouched. | None. |
| Build artifacts | `bin/` (empty, untracked tools dir), `coverage.out`, root `easyp` binary — all gitignored/untracked; unaffected by type renames. Vendor drop adds 3.4 MB under `internal/easypc` (tracked). | None (do not commit artifacts; AGENTS.md already forbids). |

## Common Pitfalls

### Pitfall 1: Column semantics mismatch — bytes+tab-8 (vendored) vs runes+tab-1 (main's contract)
**What goes wrong:** protoast positions are right on ASCII/space-indented files (that's why lab corpus parity held) and wrong on any line with multi-byte UTF-8 or a tab before the anchor.
**Why it happens:** The vendored code implements protoc's convention, verbatim:
- `internal/easypc/pkg/lexer/token.go:126-131` (lab): `// LineCol computes 1-based line and 1-based column for a given byte offset. // Tab characters ('\t') advance the 0-based column to the next multiple of 8: col0 = (col0/8 + 1) * 8. All other bytes advance col0 by 1.`
- `internal/protoast/protoast.go:6-8` (lab): `// Positions follow protoc: 1-based lines, 1-based columns counted in bytes // with tabs advancing to the next multiple of 8, and 0-based byte offsets.`
- pinned by `internal/protoast/protoast_test.go:93`: `assert.Equal(t, 9, values[1].Pos.Column, "a tab advances to the next multiple of 8, as in protoc")`

while main's contract (go-protoparser v4.14.2, module cache, verbatim):
- `parser/meta/position.go`: `// Column is a column number, starting at 1 (character count per line)`
- `lexer/scanner/position.go:35-47`: `func (pos *Position) Advance(r rune) { length := utf8.RuneLen(r); pos.Offset += length; if r == '\n' { ...; pos.Column = 1 } else { pos.Column++ } }` — every rune (tab included) advances Column by exactly 1.

The lab's own `internal/core/dom.go:12-13` comment claims `// Position is a location in a proto file: 1-based line, 1-based rune column` — the lab INTENDED main's contract but shipped byte columns (latent, untested on UTF-8). Phase 1's D-07 fixture surfaces exactly this.
**How to avoid:** Adapt `builder.pos()` in `internal/protoast/build.go` (lab line 448-454, currently `line, column := b.syn.LineCol(off)`) to compute the column as a rune count from `b.ast.Src` (line number can still come from `LineCol`). Update `protoast_test.go`'s two tab-8 expectations (lines ~93, ~96: `outer.NamePos.Column == 9` → tab-as-1 value). Record the divergence in VENDORED.md so a future sync re-applies it.
**Warning signs:** corpus diffs with same-line column shifts only on files with non-ASCII comments; golden D-07 test failing with byte-vs-rune deltas (e.g., column 34 vs 25 after a 3-rune 8-byte emoji).

### Pitfall 2: `lint_helpers.go` does not compile against main's types (D-01 letter vs reality)
**What goes wrong:** Copying `internal/rules/lint_helpers.go` "как в лаборатории" breaks the build: `packagePosition(info core.ProtoInfo) (core.Position, string)` reads `info.File.Packages` — the LAB's protoast-based `ProtoInfo` shape; main's (and Phase-1 lint's) model has `Info *unordered.Proto` and no `File`. Also `groupedValues.mismatches` calls the lab's 4-arg `core.AppendIssue(nil, rule, v.pos, v.source)`; main's is 5-arg (with `comments []*parser.Comment` — a parameter main's implementation already ignores).
**Why it happens:** The helpers were written against the lab's post-migration core API.
**How to avoid:** Vendor `casing.go` + `casing_test.go` verbatim (stdlib-only; compiles as-is). Vendor `lint_helpers.go` with minimal seam adaptation of the 3 coupled declarations (`packagePosition`, `groupedValues.add/mismatches` + `groupedValue.pos` type): keep every function body verbatim, change only the model-coupled signatures (e.g., `packagePosition(file *protoast.File, path string)`); document the delta next to D-01. The pure helpers (`enumNumber`, `leadingCommentDocuments`, `isLintDirective`, `packageHasVersionSuffix`, `versionNumber`) compile verbatim.
**Warning signs:** build errors referencing `info.File` or "too many arguments in call to core.AppendIssue" during the helpers task.
(Also note: ALL helper functions are unexported — contrary to D-02's "экспортируемые функции" premise. staticcheck since 2020.2 considers all exported identifiers used [CITED: dominikh/go-tools 2020.2 changelog, via Context7], so the actual U1000 risk is these unexported helpers sitting unused until Phase 4. `casing_test.go` exercises casing.go; `lint_helpers.go` ships with no test file. If `unused` fires, D-02's "минимальное точечное исключение" applies — most likely one path/file exclusion for `lint_helpers.go`.)

### Pitfall 3: golangci-lint red on the vendored tree (22 findings)
**What goes wrong:** `task lint:go` fails after the drop. Empirically measured this session (pinned v2.14.0, repo `.golangci.yml`, on the vendored copy): **22 issues** — 21 inside `internal/easypc` (errcheck 3, ineffassign 2, staticcheck 16 incl. QF1003 ×3, S1017, S1011, ST1005 at `protoc/compile.go:250` `"%s: File not found."` which does NOT match the existing ST1005 identifier-wrapper exclusion) and 1 in `internal/protoast/build.go:242` (QF1002).
**Why it happens:** The lab pins the same linter version and carries an identical `.golangci.yml` with no easypc exclusion — the lab's own Go lint gate is red on these trees (inferred from config+version identity and reproduced findings). golangci v2's default set (errcheck/govet/ineffassign/staticcheck/unused) stays active alongside `enable: [staticcheck]`.
**How to avoid:** Add to `.golangci.yml` (verified syntax and effect this session — reduces findings 22 → 1):
```yaml
linters:
  exclusions:
    paths:
      - internal/easypc   # sealed vendor tree (D-04); fixes go upstream
```
Fix protoast's single QF1002 in the position-adaptation commit. Frame the exclusion in the plan as vendor-tree hygiene (P9), NOT a D-02 helper exclusion.
**Warning signs:** CI lint failures listing `internal/easypc/...` paths.

### Pitfall 4: Branch/SHA pinning — "ветка easyp-cli" does not exist
**What goes wrong:** Executor runs `gh repo clone easyp-tech/compiler-lab -- --branch easyp-cli` (the command printed in `.planning/research/STACK.md`) → `fatal: Remote branch easyp-cli not found`. Or worse, pins `compile_from_internal` (newest) and silently takes the standalone-module layout whose imports (`github.com/easyp-tech/compiler-lab/easypc/...`) need rewriting.
**Why it happens:** `easyp-cli` became a directory when its history was merged under `easyp-cli/` on main (commit `08cef10`, "lab: merge easyp-cli history under easyp-cli/", 2026-11-04). Branch heads: `main=08cef10`, `compile_from_internal=eaa9a1a3d225` (2026-11-17, moved compiler to top-level module `github.com/easyp-tech/compiler-lab/easypc` — content verified identical modulo import paths + gofmt realignment), `breaking-concurrent-compile=24aeda2` (older).
**How to avoid:** Clone default branch, `git checkout 08cef10909c0722f7042ba17b5c11951f9c6b6f6`, copy `easyp-cli/internal/{easypc,protoast}`. VENDORED.md records: repo `easyp-tech/compiler-lab`, branch `main` (dir `easyp-cli/`, former `easyp-cli` branch), SHA, date, procedure.
**Warning signs:** any import-rewrite step in the plan (means the wrong revision was pinned).

### Pitfall 5: Test-literal churn from the Position type swap
**What goes wrong:** The swap `Issue.Position: meta.Position → protoast.Position` breaks 49 `meta.Position{...}` literals in 41 rule test files + `breaking_checker_test.go` (5+ sites), and 5 non-test files (`dom.go`, `breaking_checker.go:368 buildBreakingCheckIssue`, `breaking_profiles.go:300` literal, `rules/file_lower_snake_case.go:30`, `rules/package_defined.go:24`).
**Why it happens:** Expected positions in tests are struct literals of the position type.
**How to avoid:** Mechanical replace `meta.Position{` → `protoast.Position{` (+ import swap) in tests; convert at `AppendIssue`/`buildBreakingCheckIssue` in non-test code (Pattern 2). Run the full suite — the VALUES must not change (same numbers, new type).
**Warning signs:** any diff that changes numeric line/column values in tests — values changing means a real behavior change, stop.

### Pitfall 6: `task lint`/`task test` cannot run unprovisioned (and Docker is down)
**What goes wrong:** Gates fail on missing tools, not on code: `task` binary absent from PATH; `bin/` empty (no gotestsum/golangci-lint); Docker daemon inactive → `task lint` (which runs lint:go AND lint:docker and fails if either fails) can never go green as-is.
**How to avoid:** Plan an environment-provisioning task first (`task` install + `task init`), or use raw equivalents (verified working this session): `go build ./...`, `go test -race -count=1 ./...`, `/tmp/.../golangci-lint run ./...`. For `task lint` green specifically, Docker must be started (hadolint runs in a container) — or the plan states `task lint:go` as the phase gate with hadolint explicitly out of scope (no Dockerfile changes in this phase).
**Warning signs:** precondition messages "Missing local golangci-lint; run task install_linters" / "Docker is required".

## Code Examples

### Vendoring gate script (per-commit verification)

```bash
# Source: this session's empirically verified procedure (REQ-1.1/1.2 gates)
go build ./... || exit 1
go test -count=1 ./internal/easypc/... ./internal/protoast/... || exit 1
cp go.mod /tmp/go.mod.before && cp go.sum /tmp/go.sum.before
go mod tidy
git diff --exit-code go.mod go.sum || { echo "REQ-1.2 violated"; exit 1; }
diff /tmp/go.mod.before go.mod && diff /tmp/go.sum.before go.sum
```

### VENDORED.md minimal template (D-03; format is discretion, content is not)

```markdown
# Vendored: easypc compiler

- Upstream: easyp-tech/compiler-lab (private)
- Branch: main (the former `easyp-cli` branch was merged under `easyp-cli/` in this SHA)
- Directory in upstream: easyp-cli/internal/easypc
- SHA: 08cef10909c0722f7042ba17b5c11951f9c6b6f6
- Vendored: <date>
- Import paths: already github.com/easyp-tech/easyp/internal/easypc/... — no rewrite performed
- Edits inside this tree are FORBIDDEN (see AGENTS.md); fixes go upstream and arrive by re-sync
- Lint: excluded via .golangci.yml linters.exclusions.paths (findings are upstream's)
- Re-sync procedure: clone compiler-lab, checkout <new SHA>, diff easyp-cli/internal/easypc,
  copy over, re-run the gate script above, update this file's SHA/date
```

### Position adaptation locus (internal/protoast/build.go, lab lines 447-454)

```go
// Source: lab internal/protoast/build.go:447-454 (current, byte+tab-8 columns)
// pos converts a byte offset into a Position.
func (b *builder) pos(off uint32) Position {
	if off > b.ast.SrcLen {
		off = b.ast.SrcLen
	}
	line, column := b.syn.LineCol(off)
	return Position{Offset: int(off), Line: line, Column: column}
}
```
Adaptation (rune columns, tab = 1 — main's contract): keep `line` from `LineCol` (line starts are byte offsets; line numbering identical in both conventions), replace `column` with a rune count over `b.ast.Src[lineStart:off]` + 1. `b.ast.Src` is already held by the builder. `namePos` and `attachComments` route through `pos()`, so one change covers all positions. [VERIFIED: build.go read in full this session]

### D-07(b) golden test skeleton (model-level)

```go
// internal/core/position_parity_test.go — mechanics indicative
// Current lint positions ARE go-protoparser Meta.Pos values of these anchors,
// so model-level anchor comparison is equivalent to comparing issue positions.
func TestProtoastPositionsMatchLegacyContract(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile("testdata/position_contract_utf8.proto") // Cyrillic/emoji comments+strings, tab-indented decl
	require.NoError(t, err)

	legacy, _, err := readProtoFileWithDirectives(bytes.NewReader(src)) // go-protoparser path
	require.NoError(t, err)
	got, err := protoast.Read("position_contract_utf8.proto", bytes.NewReader(src))
	require.NoError(t, err)

	assert.Equal(t, legacyLCol(...), got.Packages[0].Pos.Line)   // package keyword anchor
	// ... messages/fields/enums/enum-values/services/rpcs: first-token anchors, Line+Column
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| go-protoparser as lint parser (stalled: last release 2025-05, last push 2025-11) | Own compiler `easypc` vendored in-repo | lab validated 2026 | This phase lands the compiler; lint keeps go-protoparser until Phase 4/5 (REQ-5.2 keeps it in go.mod). |
| `Issue.Position meta.Position` | `Issue.Position protoast.Position` (type alias) | this phase | Output format/JSON unchanged (identical field sets); one conversion boundary. |

**Deprecated/outdated:** nothing in main is deprecated by this phase; `STACK.md`'s printed clone command (`--branch easyp-cli`) is outdated — see Pitfall 4.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The lab's own `task lint` gate is red on these trees (inferred from identical config+version+trees and 22 reproduced findings; the lab module itself was not lint-run) | Pitfall 3 | Low — irrelevant to main's plan either way; main's exclusion is justified by its own measured findings. |
| A2 | D-07 fixtures should also cover TAB indentation (CONTEXT names only кириллица/эмодзи; tabs are the second divergence axis per P1 and the lab test pins tab-8) | Patterns/D-07 | Low — extra fixture lines; if excluded, tab divergence ships to Phase 4 discovery. |
| A3 | "Lab history merged under easyp-cli/" reading of commit 08cef10 (commit title read; deeper history not audited) | Pitfall 4 | Low — the directory+SHA are what the plan consumes; both verified directly. |
| A4 | `linters.exclusions.paths` behavior verified only for the exact v2.14.0 pin used here | Pitfall 3 | Low — plan re-runs lint after the config edit anyway. |
| A5 | golangci's `unused` may flag `lint_helpers.go`'s unexported helpers pre-Phase-4 (U1000 semantics cited from staticcheck docs; the helpers-in-main scenario not lint-run because helpers aren't in the throwaway copy) | Pitfall 2 | Low — D-02 pre-authorizes a minimal point exclusion; executor decides on actual failure. |

## Open Questions (RESOLVED)

All four questions are resolved — Q1/Q2 by explicit user confirmation (2026-11-17, recorded in `01-CONTEXT.md` § "Подтверждённые уточнения букв решений"), Q3/Q4 by the phase plans. Original wording kept below for traceability; each item carries its resolution.

1. **D-01 mechanics for `lint_helpers.go`** — "как в лаборатории" is not compilable (Pitfall 2). Recommended: vendor with the 3 model-coupled signatures adapted, bodies verbatim, delta documented for the Phase-4 sync. Needs planner ruling (and it slightly reinterprets a locked decision's letter — worth one line in the plan's risks).
   - What we know: exact coupled declarations catalogued (Pitfall 2); casing.go/casing_test.go verbatim-safe.
   - What's unclear: whether the user prefers verbatim-but-uncompilable deferral (contradicts D-01's "in Phase 1") vs adapted-now (recommended).
   - Recommendation: adapt now; record delta in VENDORED.md-adjacent note or plan.
   - **(RESOLVED)** User confirmed 2026-11-17 (CONTEXT.md "Подтверждённые уточнения букв решений", D-01 refinement): vendor now with the 3 adapted signatures, bodies verbatim, delta enumerated. Decided by plan 01-04 (Assumptions + Task 2).
2. **Reference SHA confirmation** — D-08 says "ветка easyp-cli"; no such branch exists (Pitfall 4). Recommended pin: compiler-lab **main @ 08cef10909c0722f7042ba17b5c11951f9c6b6f6**, dir `easyp-cli/`. Confirm during plan review (it is the only zero-rewrite source, so risk is minimal).
   - **(RESOLVED)** User confirmed 2026-11-17 (CONTEXT.md D-08 refinement): reference = main @ 08cef10909c0722f7042ba17b5c11951f9c6b6f6, dir `easyp-cli/`. Decided by plan 01-01 (Task 1 pin assertion + VENDORED.md, Assumptions block).
3. **`task lint` gate vs Docker unavailability** — success criterion 3 requires `task lint` green; hadolint needs Docker which is down (Environment). Either provision Docker or scope the gate to `task lint:go` + explicit no-Dockerfile-changes note.
   - **(RESOLVED)** Scoped to the Go half: plans gate on `bin/golangci-lint run ./...` with an explicit note that the hadolint half is vacuous this phase (no Dockerfile change; Docker daemon down per Environment). Decided by plan 01-01 (verification section).
4. **`protoast_test.go` tab-expectation update ownership** — updating a vendored test's assertion is a protoast-tree edit (allowed; not easypc). Confirm the plan treats protoast as the adaptation seam (recommended) rather than sealed.
   - **(RESOLVED)** protoast is the maintained adaptation seam: NOT excluded in `.golangci.yml`, tab expectations updated in-tree, divergence recorded for syncs. Decided by plan 01-02 (Task 2 + Task 3 VENDORED.md note; easypc stays sealed per D-04).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | everything | ✓ | go1.27.3 (go.mod `go 1.26.6`) | — |
| gh (authed) | cloning compiler-lab | ✓ | logged in as onokonem; clone verified this session | — |
| compiler-lab access | vendoring source | ✓ | private repo, main @ 08cef10 | — |
| `task` binary | Taskfile gates | ✗ | — | install Task, or raw go commands (verified) |
| gotestsum (`bin/`) | `task test` | ✗ (bin/ empty) | — | `task init`, or `go test -race -count=1 ./...` (verified) |
| golangci-lint (`bin/`) | `task lint:go` | ✗ (bin/ empty; global mise shim unset) | v2.14.0 pinned | `task init` / `GOBIN=... go install ...@v2.14.0` (verified this session) |
| Docker | `task lint` (hadolint), `task docker:build` | ✗ (daemon inactive; no docker/podman binary) | — | start Docker, or gate on `task lint:go` (no Dockerfile changes in phase) |
| Network (module proxy) | dep downloads | ✓ | deps downloaded during verification | — |

**Missing dependencies with no fallback:** none that block the phase's substance (build/test/lint:go all verified working with raw tools).

**Missing dependencies with fallback:** `task`, gotestsum, golangci-lint (raw equivalents verified); Docker (lint:go fallback; no Dockerfile change in scope).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify v1.11.1; runner gotestsum v1.13.0 (`task test`: `-race -count=1 -coverprofile=coverage.out ./...`) |
| Config file | `Taskfile.yml` targets `test` / `lint:go` (`.golangci.yml` v2) |
| Quick run command | `go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/protoast/...` |
| Full suite command | `go test -race -count=1 ./...` (or `task test` once provisioned) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LINT-VEND-01 | Vendored trees present & compiling; own tests pass | build + unit | `go build ./... && go test -count=1 ./internal/easypc/... ./internal/protoast/...` | Vendored tests exist in lab source; ❌ in main until vendor commit (Wave 0 of phase) |
| LINT-VEND-02 | `go mod tidy` zero-diff | gate | `go mod tidy && git diff --exit-code go.mod go.sum` | ❌ Wave 0 (script/checkpoint in plan) |
| SC-3 (behavior) | breaking/generate/mod/migrate unchanged | regression (existing suites) | `go test -race -count=1 ./...` | ✅ existing suites (verified green baseline: `./internal/core` passes) |
| SC-4 (positions) | span→Position conversion, UTF-8+tabs | unit | `go test -count=1 ./internal/protoast/ -run TestPosition` | ❌ Wave 0 — new `position_test.go` |
| SC-4 (positions) | golden legacy-vs-protoast anchor parity | integration (model-level) | `go test -count=1 ./internal/core/ -run TestProtoastPositions` | ❌ Wave 0 — new `position_parity_test.go` + fixture |

### Sampling Rate
- **Per task commit:** quick run command above + `go build ./...`; vendor commit additionally runs the tidy zero-diff gate.
- **Per wave merge:** full suite command + `golangci-lint run ./...` (with the easypc path exclusion).
- **Phase gate:** full suite green, lint green (`task lint` or documented `task lint:go` scope), tidy zero-diff, and both D-07 tests green — before `/gsd-verify-work`.

### Wave 0 Gaps
- [ ] `internal/easypc/**` + `internal/protoast/**` — the vendor drop itself (REQ-1.1 carrier)
- [ ] `internal/protoast/position_test.go` — D-07(a), with UTF-8+tab fixture (placement discretion)
- [ ] `internal/core/position_parity_test.go` + `testdata/position_contract_utf8.proto` — D-07(b); note: 0 non-ASCII fixtures exist in main's `internal/rules/testdata` today (verified), so this fixture is genuinely new
- [ ] `.golangci.yml` exclusion + `VENDORED.md` + AGENTS.md D-04 note
- [ ] Environment provisioning if Taskfile gates are used: `task` binary + `task init`

## Security Domain

ASVS level 1 (config `security_asvs_level: 1`, `security_enforcement: true`).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|------------------|
| V2 Authentication | no | CLI tool, no auth surface |
| V3 Session Management | no | — |
| V4 Access Control | no | — |
| V5 Input Validation | marginal | Proto sources are user input; Phase 1 does not change any runtime parsing path (vendored compiler is unwired). Parser hardening/fuzzing is FUTURE-HARDENING-01. |
| V6 Cryptography | no | — |

### Known Threat Patterns for this phase

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Supply-chain injection via new dependencies | Tampering | The zero-new-deps invariant IS the control: import scan + tidy zero-diff gate (REQ-1.2). Nothing installs from any registry. |
| Vendor-tree drift / trojaned re-sync | Tampering | D-03/D-04: pinned SHA in VENDORED.md, no in-tree edits, standalone vendor commit, sync is a deliberate diff review. |
| Untrusted content in fetched repo | Tampering/EoP | compiler-lab is easyp-owned private inner source (same trust domain); clone treated as source, not executed. `testdata/` is data for tests only; Go toolchain ignores `testdata/` dirs for package loading. |
| Embedded WKT tampering | Tampering | `protoc/wkt` embeds 20 upstream-shaped .proto files via `go:embed` (wkt.go:9); they are data, reviewed as part of the vendor diff. |

## Sources

### Primary (HIGH confidence — read/executed this session)
- Throwaway copy-in experiment (this session): main repo copy + lab trees → `go build ./...` green; `go mod tidy` go.mod/go.sum byte-identical; `go test ./internal/easypc/... ./internal/protoast/...` all ok; golangci-lint v2.14.0 runs (22 findings raw, 1 after verified `linters.exclusions.paths: [internal/easypc]`).
- compiler-lab clone (main @ `08cef10909c0722f7042ba17b5c11951f9c6b6f6`): branch heads via `git ls-remote` (`main=08cef10`, `compile_from_internal=eaa9a1a3d225` "feat: Move compiler from internal" 2026-11-17, `breaking-concurrent-compile=24aeda2`); no `easyp-cli` branch (clone attempt failed); cross-branch content diff = import-rewrite + gofmt only; `internal/easypc` tree listing, sizes (3.4M/2.6M testdata/204K wkt), 48+6 .go files; import scans; `internal/protoast/{protoast,build,comments,walk,outline,protoast_test}.go` read in full; lab `internal/core/dom.go:12-14,80`; lab `internal/rules/{lint_helpers,casing,casing_test}.go` read in full; lab `.golangci.yml`, `Taskfile.yml` pins.
- Main repo source: `internal/core/dom.go` (Issue/ProtoInfo/AppendIssue/GetRuleName), `lint.go` (engine loop, `c.protoInfoRead` at :34), `proto_info_read.go` (:17 def; callers lint.go:34, breaking_check.go:119), `breaking_check.go` (readProtoFiles/collect), `core.go` (Options/Core), `comment_suppressions.go`, `proto_comments.go`, `internal/api/runtime.go` (buildCore), `internal/api/lint.go` (text :125 / JSON :145-149 printers), `internal/api/breaking_v1.go:172,483` (`.Position.Line/.Column` reads), `go.mod`, `Taskfile.yml`, `.golangci.yml`.
- Greps quantifying the split ripple: 5 non-test `meta.Position` files; 41 rule test files / 49 literals; 48 `.Meta.Pos|.Meta.LastPos` sites; 41 `core.AppendIssue` caller files; 45 non-test `ProtoInfo` consumers.
- go-protoparser v4.14.2 module cache: `parser/meta/position.go` (rune-column doc), `lexer/scanner/position.go:35-47` (`Advance`, verbatim quoted in Pitfall 1).
- Environment probes: go1.27.3, `bin/` empty, docker inactive, gh authed; baseline `go build ./...` and `go test ./internal/core/` green on main.

### Secondary (MEDIUM confidence — per classify-confidence seam for context7)
- go.dev/ref/mod via Context7 `/golang/website` — `go mod tidy` add/remove semantics.
- staticcheck 2020.2 changelog via Context7 `/dominikh/go-tools` — whole-program U1000 removed; exported identifiers always used.

### Tertiary (LOW confidence)
- None used.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — everything executed or read from source; no external claims.
- Architecture: HIGH — split surface fully quantified from greps + reads; mechanics recommendation grounded in lab precedent.
- Pitfalls: HIGH — position divergence and lint findings empirically reproduced; A1/A5 inference flagged in Assumptions.

**Research date:** 2026-11-17
**Valid until:** 2026-12-17 (stable domain; re-check only if compiler-lab `compile_from_internal` branch gets merged somewhere relevant or main gains new commits touching easyp-cli/internal/easypc)
