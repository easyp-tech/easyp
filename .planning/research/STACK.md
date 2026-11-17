# Stack Research

**Domain:** In-repo protobuf compiler + lint engine for a Go CLI (vendoring, parsing/linking, lint infrastructure)
**Researched:** 2026-11-17
**Confidence:** HIGH for library-status facts and the vendoring/engineering approach; MEDIUM for ecosystem-survey breadth (DuckDuckGo search backend returned empty results; evidence is from GitHub API, official release notes, Context7 docs, and buf's public go.mod)

**Context:** Milestone `lint-easypc-infra` — migrate `easyp lint` from go-protoparser to the in-repo compiler vendored as `internal/easypc` + AST model `internal/protoast`, sourced from the validated `easyp-tech/compiler-lab` (branch `easyp-cli`). Hard constraint REQ-1.2: **no new external dependencies** — nothing below adds a `go.mod` entry.

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.6 (repo pin) | Language for compiler, AST, lint engine | Go 1.26 is the ecosystem floor in late 2026 (jhump/protoreflect v2.0.0 notes: "requires Go 1.26 or higher. That is the same minimum version supported by the Go team"). Native fuzzing, generics, and `testing` are all the toolchain this milestone needs. |
| `internal/easypc` (vendored from compiler-lab, branch `easyp-cli`) | copy-in, ~13 packages, ~3.4 MB with testdata | The compiler: `pkg/` (lexer, parser, linker, sourceinfo, serialize…), `protoc/` (public API + embedded WKT), `cmd/easyp-protoc`, testdata | Already written, imported, and validated in the lab: same findings on acceptance corpora, 20 min → 0.08 s on a 4 MB `compute.proto`. Vendoring the whole thing (including `cmd/` and `testdata/`, user decision 2026-11-17) preserves the tested shape and keeps a path for the future breaking/generate migration. Zero external dependencies; imports already rewritten to `github.com/easyp-tech/easyp/internal/easypc/...`. |
| `internal/protoast` (vendored from compiler-lab) | copy-in, ~1000 LOC (`build.go`, `protoast.go`, `walk.go` Visitor, `comments.go`, `outline.go`) | Linter-facing AST model with positions and comment attachment | Purpose-built for lint rules: `core.Position = protoast.Position`, `ProtoInfo.File *protoast.File` + `ImportedFiles`. Decouples the 43 rules from compiler internals so rules never touch lexer/linker types. Depends only on `easypc/pkg/parser` and `easypc/protoc/syntax` — clean seam. |
| Lint engine per lab design (`internal/core/lint.go`, `link.go`) | lab revision | Two-pass lint: `protoc.Link` with `WithSyntaxTrees(true)`, import cache, buf-compatible `COMPILE` issues on link failure, `ModuleRule` (`Collect`/`Finish`) for cross-file rules | This is the only design that delivers both goals: buf-parity `COMPILE` behavior and cross-file rules (`PACKAGE_NO_IMPORT_CYCLE`). Linking with syntax trees retained is what makes comment suppressions and position fidelity survive into the linked model. |
| `internal/parallel` (already in main) | existing | Parallel lint windows of 256 files | Already exists in both main and lab (verify compatibility in Design phase — flagged assumption in explore.md). No new concurrency dependency needed. |

### Supporting Libraries (all already in `go.mod` — nothing added)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/yoheimuta/go-protoparser/v4` | v4.14.2 (latest tag, 2025-05-18; last repo push 2025-11-24) | Stays for breaking-change detection (`breaking_check.go` → `collect` → `unordered.Proto`) and any other non-lint consumers | Keep untouched this milestone. Removal is deferred until breaking migrates to the compiler (v2 of the series). |
| `github.com/bufbuild/protocompile` | v0.14.1 (latest tag, 2024-08-30; main actively developed, pushes as of 2026-11-17; buf CLI consumes main via pseudo-version `v0.14.2-0.20261110…`) | Stays for its current role in the non-lint subsystems (mod-tidy imports etc.). **Secondary role this milestone: parity oracle.** | Use as a reference implementation when shaping `COMPILE`-issue text/positions to match buf behavior — it is the engine buf itself runs. Do NOT wire it into the lint path. |
| `github.com/stretchr/testify` | v1.11.1 (latest: v1.12.1, 2026-08-19 — compatible, no mid-milestone bump needed) | Colocated rule tests `<rule>_test.go`, doubles | All 43 migrated rules + 3 new rules keep the existing testify colocated-test convention (AGENTS.md). |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| Task (Taskfile.yml) | All build/test/lint entrypoints | Unchanged. `task test`, `task lint`, `task schema:check` are the milestone gates. |
| gotestsum v1.13.0 | Test runner with `-race -count=1` + coverage | Pinned in `bin/`, and v1.13.0 **is** the latest release (2025-09-11) — current. |
| golangci-lint v2.14.0 | Static analysis incl. vendored compiler code | Latest is v2.15.0 (2026-10-16), same v2 major — optional post-milestone bump; do not churn mid-milestone. Configure exclusions for vendored `internal/easypc` testdata if needed. |
| Go native fuzzing (`testing.F`) | Hardening for the hand-written lexer/parser | Stdlib since Go 1.18 (verified against golang/go source): `f.Add` seed corpus from `easypc` testdata + acceptance corpora, fuzz parse-then-invariants (no panic, positions monotonic, AST builds). Cheap to add, catches the crash class `-race` cannot. Recommended as a follow-up hardening task, not a milestone blocker. |
| `compiler-lab/bench/lab.py` | Acceptance harness: `lint` / `lint-diff` snapshots on `api-common` (92 files) and `googleapis-nocompute` (7414 files) | Requires `gh` access to the private compiler-lab repo. This is the parity + performance gate; corpora rebuild from googleapis via `bench/setup.sh`. |
| benchstat (optional) | Comparing benchmark runs for the perf gate | Only if "not slower than main" needs numeric evidence beyond `lab.py` wall-clock; `testing.B` benchmarks are stdlib. |

## Installation

Nothing to install for production — the milestone's defining property is that `go.mod` gains **zero** entries. The vendoring procedure:

```bash
# 1. Materialize the reference implementation
gh repo clone easyp-tech/compiler-lab /tmp/compiler-lab -- --branch easyp-cli

# 2. Copy the compiler and the AST model (whole trees, incl. cmd/ and testdata/)
cp -r /tmp/compiler-lab/easyp-cli/internal/easypc   internal/easypc
cp -r /tmp/compiler-lab/easyp-cli/internal/protoast internal/protoast
# (imports in the lab copy already point at github.com/easyp-tech/easyp/internal/...)

# 3. Prove the no-new-deps invariant
go mod tidy
git diff --stat go.mod go.sum        # expected: no new require lines

# 4. Gates
task test && task lint
```

Dev tools are already provisioned by `task init` (gotestsum, golangci-lint, mockery pinned into `bin/`).

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| Vendored `easypc` copy-in under `internal/` | `bufbuild/protocompile` as the lint engine | Only if the lab compiler did not exist or failed parity. protocompile is the ecosystem reference (buf CLI itself runs it; pure-Go full pipeline: parser → linker → options → sourceinfo; editions 2023 supported; `Compiler{Resolver, MaxParallelism, Reporter, SourceInfoMode, RetainASTs}` API). But it is descriptor-centric: syntax-level rules (casing, comment suppressions, `SYNTAX_SPECIFIED`) would fight the descriptor/SourceInfo model, the lab's validated 43-rule rewrite would be discarded, and lint behavior would track buf's release decisions instead of EasyP's own cadence. It stays in `go.mod` for other subsystems and as the COMPILE-parity oracle. |
| Vendored `easypc` copy-in | Continue on `yoheimuta/go-protoparser` | Never for new lint capability. Parser-only (no linking, no descriptors, no COMPILE parity), and maintenance has stalled: last release v4.14.2 2025-05-18, last push 2025-11-24 (~12 months stale). Acceptable only as the frozen incumbent for breaking until that subsystem migrates. |
| Vendored `easypc` copy-in | `jhump/protoreflect` v2 | Never as a parser. v2.0.0/v2.0.1 (2026-10/11) is actively maintained but its release notes confirm `desc/protoparse` (deprecated since v1.17.0) was **removed**, redirecting source-IDL parsing to protocompile. v2's remaining scope is gRPC reflection + dynamic invocation. Not a lint-path option at any version. |
| Plain copy-in, imports rewritten | Separate module + `replace` directive | Only when you need frequent upstream re-syncs of a foreign upstream. easypc is easyp-owned (the lab IS the upstream), imports are already rewritten, and REQ-1.2 forbids new module edges. |
| Plain copy-in | `git subtree` / `git submodule` | Only for tracking a foreign upstream's history. Here the code has permanently diverged (protoast seam, rewritten imports); subtree pulls would add merge noise with no upstream to track. |
| Hand-written recursive-descent lexer/parser (as in `easypc`) | ANTLR/javacc-style generated parser | Only in greenfield projects without Go-native grammar tooling constraints. Generated parsers drag a runtime + generator toolchain, complicate exact position/comment fidelity (the core of lint), and neither protoc, protocompile, nor the lab took that path. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `jhump/protoreflect/v2` parsing APIs | The parser no longer exists in v2; official notes point to protocompile | `internal/easypc`; protocompile as oracle |
| New external dependencies on the lint path (any parser/AST/graph lib) | Prohibited by REQ-1.2; also pointless — import-cycle detection is a DFS over `ImportedFiles` maps, ~tens of lines | Hand-rolled cycle detection in `ModuleRule.Finish` (lab already implements `package_no_import_cycle`) |
| `go mod vendor` for the compiler | That mechanism caches *declared module dependencies*; it does not own or internalize forked code | Copy-in under `internal/` (chosen) or module+`replace` (rejected above) |
| Deleting `go-protoparser` / `protocompile` from `go.mod` in this milestone | breaking/generate/mod-tidy still consume them (REQ-5.2) | Leave both; removal lands with the later breaking/generate migration features |
| Editing `schemas/*.json` by hand when the 3 new rules ship config | Generated artifacts (AGENTS.md) | `internal/config/v1` types → `task schema:generate` → `task schema:check` |
| Mid-milestone bumps of testify / golangci-lint / gotestsum | Churn risk with zero milestone value; current pins are latest or one compatible minor behind | Stay on pinned versions; batch bumps after milestone close |

## Stack Patterns by Variant

**This milestone (lint-only migration):**
- Dual models coexist deliberately: lint reads `protoast` linked trees; breaking keeps reading `unordered.Proto` (go-protoparser). `core.ProtoInfo` gains `File *protoast.File` + `ImportedFiles` without removing the old field.
- Because: the breaking↔ProtoInfo coupling (explore.md "key finding") is out of scope; premature unification recreates Option C's unacceptable diff size.

**Later milestones (breaking/generate/tidy/migrate on the compiler):**
- Extend `protoast`/`ImportedFiles` and the descriptor export already in `internal/generation`; then go-protoparser exits `go.mod`.
- Because: the compiler is already fully vendored; subsystem migration becomes model plumbing, not compiler work.

**If `easypc` needs a fix found during rule migration:**
- Patch in-repo; compiler-lab stays the reference. Re-sync is a deliberate manual diff, never VCS plumbing.
- Because: the in-repo copy is now the source of truth for main (REQ-3.2: main-side behavior drift wins over the lab copy).

**Forward-looking gap — editions:** acceptance corpora (`api-common`, `googleapis-nocompute`) are proto2/proto3; easypc's editions support is unverified. protocompile main already handles editions 2023 (`edition` keyword, feature gating) and is building an experimental next-gen parser/IR. If user corpora adopt editions, that becomes a new milestone for `easypc` — flag it in the roadmap, do not scope-creep it here.

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| Go 1.26.6 (go.mod) | protoreflect v2's Go 1.26+ floor | Confirms the repo's pin matches the ecosystem minimum for late 2026; no toolchain change needed. |
| protocompile v0.14.1 (tag) | buf CLI's `v0.14.2-0.20261110…` pseudo-version of main | Tag is 2024-stale while main moves daily; buf tracks main via pseudo-versions. For the oracle role, running the pinned v0.14.1 locally + `buf` CLI against the same corpora is sufficient; do not chase main. |
| testify v1.11.1 | latest v1.12.1 | Drop-in compatible; defer the bump. |
| golangci-lint v2.14.0 | latest v2.15.0 | Same v2 config schema; vendored `internal/easypc` may need path exclusions for its testdata — verify `task lint` passes before wiring rule batches. |
| gotestsum v1.13.0 | — | Is the latest release; nothing to do. |

## Sources

- GitHub API (releases + repo metadata, primary source): [bufbuild/protocompile](https://github.com/bufbuild/protocompile) — v0.14.1 tag 2024-08-30, pushes through 2026-11-17; [yoheimuta/go-protoparser](https://github.com/yoheimuta/go-protoparser) — v4.14.2 tag 2025-05-18, last push 2025-11-24; [jhump/protoreflect](https://github.com/jhump/protoreflect) — v2.0.0 2026-10-10, v2.0.1 2026-11-14. **HIGH** (authoritative first-party data, cross-corroborated).
- jhump/protoreflect [v2.0.0-beta.2 release notes](https://github.com/jhump/protoreflect/releases/tag/v2.0.0-beta.2) — removed packages list, `desc/protoparse` → protocompile redirect; [v2.0.0 notes](https://github.com/jhump/protoreflect/releases/tag/v2.0.0) — Go 1.26+ floor. **HIGH** (official release notes).
- buf CLI public `go.mod` (main) — `protocompile v0.14.2-0.20261110…` + `protoreflect/v2 v2.0.1` direct dependencies. **HIGH** (primary evidence of who runs what).
- Context7 `/bufbuild/protocompile` — Compiler API surface (`Resolver`, `MaxParallelism`, `Reporter`, `SourceInfoMode`, `RetainASTs`), package layout (parser/ast/linker/options/sourceinfo/reporter/walk), `WithStandardImports` embedded WKT, editions testdata, experimental parser/IR. **MEDIUM** (per seam tier; corroborated by buf's dependency).
- Context7 `/golang/go` — native fuzzing mechanics (`testing.F`, `f.Add`, `-fuzz`/`-fuzztime`, corpus in `testdata/fuzz` + `GOCACHE/fuzz`, `go clean -fuzzcache`). **MEDIUM** (per seam tier; stable stdlib surface since Go 1.18).
- Web search (DuckDuckGo via cc-websearch skill) — returned empty results for both ecosystem-survey and vendoring-pattern queries; those sections above rest on the primary sources listed plus the project's own validated decision record (PROJECT.md, explore.md). **LOW for survey breadth** — noted honestly; the recommendation does not depend on the missing survey.
- In-repo ground truth: `go.mod` (Go 1.26.6, protocompile v0.14.1, go-protoparser v4.14.2, testify v1.11.1), `Taskfile.yml` (gotestsum v1.13.0, golangci-lint v2.14.0, mockery v2.53.7 pins), `.planning/PROJECT.md`, `.spec/features/lint-easypc-infra/explore.md`. **HIGH**.

---
*Stack research for: in-repo protobuf compiler + lint engine (lint-easypc-infra)*
*Researched: 2026-11-17*
