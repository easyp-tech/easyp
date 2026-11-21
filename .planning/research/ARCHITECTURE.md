# Architecture Research

**Domain:** Compiler-backed lint engine integrated into an existing Go protobuf CLI (EasyP, milestone `lint-easypc-infra`)
**Researched:** 2026-11-17
**Confidence:** HIGH for project-specific structure (verified against source at `5c9185f`, lab facts from validated explore.md); MEDIUM for external pattern corroboration (buf/protocompile sources fetched from upstream `main`; strangler/parity literature)

## Standard Architecture

### How compiler-backed lint engines are structured

Every mature protobuf lint engine follows the same three-layer shape. Buf — the system EasyP is drop-in compatible with — is the canonical reference, and the compiler-lab implementation already matches it.

```
┌───────────────────────────────────────────────────────────────────────┐
│  Lint engine (per-command orchestration)                              │
│  walk targets → compile/link unit → build rule views → run rules →    │
│  apply suppressions → ordered findings                                 │
├───────────────────────────────────────────────────────────────────────┤
│  Rule layer                                                            │
│  rule catalog (specs/constructors) + scoped execution adapters:       │
│  per-file · per-declaration · per-dir · per-package · whole-module    │
│  rules write annotations; they never print, never apply ignores       │
├───────────────────────────────────────────────────────────────────────┤
│  Compiler / view layer                                                 │
│  resolver (import roots + thread-safe source accessor + WKT)          │
│  → parser (AST with positions & comments)                              │
│  → linker (symbol resolution, type checks; diagnostics via reporter)  │
│  → semantic view model over linked trees (rules' only data surface)   │
└───────────────────────────────────────────────────────────────────────┘
```

Verified structural facts (bufbuild/protocompile README + `compiler.go`/`resolver.go`, bufbuild/buf `bufcheckserverutil`):

1. **Compilation is unit-scoped, not file-scoped.** `Compiler.Compile(ctx, files...)` links all targets plus transitive imports into one graph with built-in parallelism (`MaxParallelism`). The compiler-lab equivalent is `protoc.Link` with `WithSyntaxTrees(true)` over the whole corpus window.
2. **The resolver is the integration seam.** `SourceResolver{ImportPaths, Accessor}` with a thread-safe accessor, `CompositeResolver` for fallback chains, `WithStandardImports` for embedded WKTs, `SourceAccessorFromMap` for tests. EasyP's generate path *already* uses exactly this shape — `internal/core/generate.go:151` builds `wellknownimports.WithStandardImports(&protocompile.SourceResolver{ImportPaths, Accessor: c.openSourceFile})`. The lint migration replicates a pattern the codebase already contains.
3. **Syntax trees are retained for lint.** protocompile's `RetainASTs`; the lab's `WithSyntaxTrees(true)`. Descriptors erase what EasyP's 43 rules need (comments, casing, declaration order, `syntax` presence), so rules must see trees, not just descriptors.
4. **A view model sits between trees and rules.** Buf wraps linked descriptors as `bufprotosource` (locations, comments); the lab's `protoast` (~1000 lines: `build.go`, `walk.go` Visitor, `comments.go`) is the same idea over syntax trees. This layer is what makes rules portable across compiler internals later.
5. **Cross-file rules are first-class via scoped handlers.** Buf's factories: `NewLintFileRuleHandler` (per file), `NewLintFilesRuleHandler` (whole module, imports excluded), `NewLintPackageToFilesRuleHandler`, `NewLintDirPathToFilesRuleHandler`, per-declaration handlers (Enum/Field/Message/…). The lab's `ModuleRule` (`Collect`/`Finish`) is the map-reduce form of buf's whole-module handler: `Collect` runs per file inside parallel windows, `Finish` aggregates serially per module. This shape fits a windowed engine better than "handler receives all files", because the map phase parallelizes and only the reduce is serial.
6. **The framework owns suppression; rules only report.** Buf applies comment ignores (`allow_comment_ignores`) and path ignores after rules run, resolving comment locations from source metadata (`SourceCodeInfo` + `protosourcepath` associated paths). EasyP's engine already does this (`comment_suppressions.go` filters `Validate` results in `lint.go:57-64`); the new engine must preserve that ownership split (REQ-2.5).
7. **Compile diagnostics surface as findings under a synthetic rule name.** Buf reports compilation failure instead of running rules; the lab's `COMPILE` issues with `RuleName: "COMPILE"` (`internal/core/link.go` in lab) are the same contract, now formalized in REQ-2.2/2.3.

### System Overview — EasyP after this milestone

```text
┌──────────────────────────────────────────────────────────────────────────┐
│ CLI entrypoints: cmd/easyp, cmd/easyp-mcp                                │
├──────────────────────────────────────────────────────────────────────────┤
│ internal/api (command wiring; buildCore constructs engines per command)  │
│   easyp lint ──────────┐            easyp breaking ─────────────┐        │
├────────────────────────┼─────────────────────────────────────────┼───────┤
│ NEW lint path          ▼          LEGACY paths (unchanged)       ▼       │
│ internal/core/lint.go (rewritten, 2-pass, windows of 256)   breaking_    │
│ internal/core/link.go (new: resolver+link+COMPILE+fallback)  check.go    │
│ internal/rules (43 rewritten + 4 new, protoast-based)        readProto-  │
│        │                                                     Files→collect │
│        ▼                                                     (go-proto-    │
│ internal/protoast (File model, walk, comments) ◄──────────── parser)      │
│        │                                  generate: protocompile (as-is) │
│        ▼                                  mod tidy / migrate: as-is      │
│ internal/easypc (vendored compiler: lexer, parser, linker,               │
│  sourceinfo, protoc API + WKT; imports nothing from easyp)               │
├──────────────────────────────────────────────────────────────────────────┤
│ Shared: internal/config(+v1), internal/parallel, internal/adapters/*     │
└──────────────────────────────────────────────────────────────────────────┘
```

The two models coexist **across command boundaries, never within one path**: lint reads `protoast.File`; breaking/generate keep go-protoparser/protocompile. `go-protoparser` and `protocompile` stay in `go.mod` (REQ-5.2).

### Component Responsibilities

| Component | Responsibility | Boundary rule |
|-----------|----------------|---------------|
| `internal/easypc` (vendored) | Compiler: lexer, parser, linker, sourceinfo, `protoc` API, embedded WKTs | Self-contained; imports **nothing** from easyp packages. Sync point with upstream compiler-lab, so treat as vendor code: no easyp-specific logic inside |
| `internal/protoast` | Lint view model: `File` AST, `build.go` from linked trees, `walk.go` Visitor, `comments.go`, `outline.go` | Depends only on `easypc/pkg/parser` + `easypc/protoc/syntax`. Rules and core depend on it; it depends on neither |
| `internal/core/link.go` (new) | Unit linking: resolver over import roots (thread-safe open via `core.Options.OpenSourceFile`, WKT imports), link with syntax trees, import cache, link-error → `COMPILE` issues, missing-import → per-file fallback (REQ-2.2/2.3) | Only core code that talks to `easypc/protoc`; owns the "buf-compatible failure mode" contract |
| `internal/core/lint.go` (rewritten) | Two-pass engine: pass 1 windows of 256 files via `internal/parallel` (link + protoast build), pass 2 rules (file rules parallel, `ModuleRule.Collect`/`Finish`), suppression filter, deterministic output order (REQ-2.1, 5.1) | No rule logic, no parsing details; owns traversal order and result assembly |
| `internal/rules` | 43 rewritten + 4 new rules (`casing` helpers in `lint_helpers.go`/`casing.go`) | Rules see `protoast` types only — never `easypc` internals. Colocated `_test.go` per rule (AGENTS.md) |
| `internal/core/breaking_*`, `proto_info_read.go` | Legacy model: `protoInfoRead` → `unordered.Proto` → `collect` → `ProtoData` → `BreakingChecker` | **Untouched this milestone** (REQ-5.2). Its continued compilation is the acceptance criterion for the decoupling work |
| `internal/core` shared types | `IssueInfo`/`Issue` (output), `Rule`/`ModuleRule` (contracts), config plumbing | `Issue.Position` becomes `protoast.Position` (lab decision) — shared by breaking output construction, see Decoupling below |
| `internal/api` | `buildCore` wiring, per-command engine construction | The **switch point**: one commit flips `easyp lint` from old loop to new engine |
| Parity harness (dev-time) | Two binaries (main vs branch, plus lab CLI for new rules), corpora `api-common` (92) + `googleapis-nocompute` (7414), diff key = (rule, file, line, column, message) | Lives in compiler-lab (`bench/lab.py`), not in the easyp binary |

## Recommended Project Structure

```text
internal/
├── easypc/                  # vendored compiler (as in lab: pkg/, protoc/, cmd/, testdata)
│   ├── pkg/lexer, pkg/parser, pkg/linker, pkg/sourceinfo, pkg/serialize …
│   └── protoc/              # Link API + WKT
├── protoast/                # lint view model (build, walk, comments, outline)
├── core/
│   ├── lint.go              # REWRITTEN: 2-pass windowed engine
│   ├── link.go              # NEW: resolver + link + COMPILE + fallback
│   ├── comment_suppressions.go   # re-based onto protoast comments (same semantics)
│   ├── dom.go               # SPLIT: lint-facing vs breaking-facing types
│   ├── proto_info_read.go   # RETAINED for breaking only
│   └── breaking_*.go        # unchanged
└── rules/                   # 43 rewritten + 4 new, each with colocated test
```

### Structure Rationale

- **`easypc` as sealed vendor tree:** keeps the future sync with compiler-lab mechanical (copy + import rewrite); no easyp logic leaks in, so the next migration (breaking, generate) reuses it unchanged.
- **`protoast` between compiler and rules:** the stability seam. When the compiler evolves (options interpretation, sourceinfo), rules don't change — exactly buf's `bufprotosource` role.
- **`link.go` separate from `lint.go`:** linking policy (resolver chain, error classification: compile error vs missing import vs read error) is the buf-compatibility contract; the engine (windows, ordering, suppressions) is EasyP policy. Different rates of change.
- **Breaking keeps its files:** zero-diff for breaking is verifiable by review; the legacy read path becomes obviously "breaking-owned" once lint stops calling `protoInfoRead`.

## Architectural Patterns

### Pattern 1: Hard switch at the command boundary (strangler by command)

**What:** The CLI facade routes per command, not per config flag. `easyp lint` moves to the new engine in one release; `easyp breaking`/`generate`/`mod`/`migrate` keep their parsers. Engine + model + rules switch **together** in a single commit; there is no runtime dual-engine mode and no user-facing feature flag.

**When to use:** When the new and old data models are not interface-compatible and the seam (command dispatch) is already clean. Both hold here.

**Trade-offs:** The flip commit is large-by-nature (all rules rewritten); mitigated by the batch-porting work that precedes it (each rule lands with its colocated test on the branch, the flip only rewires). Rollback = revert the flip commit; the legacy path still exists in-tree until cleanup.

**Why not the alternatives** (this is the core design question from REQUIREMENTS.md):

| Alternative | Verdict | Reason |
|---|---|---|
| **Bridge**: adapter making protoast satisfy old `Rule.Validate(ProtoInfo)` | Rejected | The old contract is direct field access (`protoInfo.Info.ProtoBody.Messages`), not a narrow interface. A bridge must fake a go-protoparser tree from protoast — a second compiler, lossy and bug-prone, kept alive only to be deleted. Bridges pay off when the contract is narrow; this one is not |
| **Runtime feature flag / dual-run in the shipped binary** | Rejected | Right tool for services with live traffic (Diffy-style shadowing), wrong for a CI CLI: doubles user runtime, risks leaking experimental output format (REQ-5.1 makes format a contract), and CI users cannot act on "shadow findings". Shadow comparison belongs at development time (Pattern 3) |
| **Per-rule strangler inside one lint run** (old and new rules mixed per batch) | Rejected | Requires the bridge; also mixes two position/comment models in one output stream during migration |
| **Migrate breaking too** (lab's approach) | Out of scope | Explore.md Option C — REQ-5.2 pins breaking; diff size unacceptable |

The expand-contract sequencing (below) keeps `main` green while the branch carries the batch work — that is where the "incremental" property of strangler lives, not in the shipped binary.

### Pattern 2: Expand–contract sequencing (parallel change on the branch)

**What:** Land changes in an order where every intermediate state compiles and passes `task test`:

1. **Expand:** vendor `internal/easypc` + `internal/protoast` (pure addition; `go.mod` unchanged — REQ-1.2).
2. **Cut the coupling:** give breaking its own (unchanged) read path; introduce lint-facing types; decide the shared `Issue.Position` type. No behavior change yet.
3. **Build new engine unwired:** `link.go` + rewritten `lint.go` + first ported rules, constructed only in tests.
4. **Batch-port the 43 rules** with colocated tests, reconciling against **current main semantics** (REQ-3.2 — main outranks the lab where they diverge).
5. **Contract (the flip):** `buildCore`/lint command uses the new engine; old lint loop and old rule implementations are deleted in the same commit.
6. **Cleanup deferred:** `go-protoparser` removal waits for breaking's own future migration (Out of Scope).

**When to use:** any in-binary engine replacement. This is the standard strangler-fig/parallel-change discipline (MS Learn, AWS Prescriptive Guidance, Fowler's feature-parity note: characterize the old behavior first, prove the new one matches, then delete).

**Trade-offs:** the branch lives for the duration of the batch porting; main-evolves-during-porting risk is real (mitigation: rebase discipline + REQ-3.2 rule "main wins"; the lab copy is a *reference*, not a source of truth for semantics).

### Pattern 3: Two-level parity harness (development-time shadow)

**What:** Prove equivalence by running old and new implementations on identical inputs and diffing normalized output — never by reviewing the rewrite diff.

- **Level 1 — per-rule golden tests (localize):** colocated `<rule>_test.go` ported alongside each rule. Corpus-level diffs cannot attribute a mismatch to a rule; unit goldens can. This answers the open design question "batch parity tests vs corpus-only lint-diff": **both, with different jobs** — goldens localize, corpus proves coverage of the long tail.
- **Level 2 — corpus lint-diff (accept):** compiler-lab `bench/lab.py`: run main binary and branch binary over `api-common` + `googleapis-nocompute`, diff on the parity key (rule, file, line, column) per the REQUIREMENTS.md glossary; message text additionally stable per REQ-5.1. New rules (REQ-4.3) compare against the **lab CLI**, not main (they don't exist in main).

**When to use:** any rewrite whose output is a contract. Fowler's feature-parity pattern and MS shadow-testing guidance converge on "compare outcomes, inspect disagreements, delete the old path only when mismatch count is zero."

**Trade-offs:** lab.py needs private compiler-lab access and corpus rebuilds (explore.md flags this as a spike); position/column conventions (1- vs 0-based, tab handling) are the classic silent diff source — normalize deliberately.

### Pattern 4: Map-reduce cross-file rules (ModuleRule)

**What:** `ModuleRule` with `Collect(per-file) error` invoked inside parallel windows and `Finish(module) []Issue` invoked once after the window sweep. State accumulates in the rule instance; the engine guarantees the phase boundary.

**When to use:** rules whose predicate spans files (`PACKAGE_NO_IMPORT_CYCLE` — the import graph; buf's analogs are the whole-module/package/dir handler scopes). Keeps the parallel window structure intact: only the reduce phase is serial.

**Trade-offs:** `Collect` must be safe to call concurrently on distinct files but mutate shared rule state — requires the rule to own synchronization or the engine to shard by key (import graph edges shard naturally by file). Determinism of `Finish` output must not depend on `Collect` completion order — sort findings at the end (engine responsibility).

## Data Flow

### Lint (new path) — direction: sources → linked trees → views → findings

```
api.Lint (flags, policy config)
  ↓
buildCore: rules.New(cfg.Lint) → []core.Rule / ModuleRule   [config validation: REQ-2.6]
  ↓
core.Lint(ctx, fsWalker)
  ↓ walk (ignore paths, .proto filter; traversal order preserved — REQ-2.1)
  ↓
┌─ Pass 1 (windows of 256, internal/parallel) ────────────────────────────┐
│ link.go: resolve imports (roots + OpenSourceFile + WKT, cached)         │
│   easypc parse → link with syntax trees                                  │
│   ├─ compile error (not missing import) → COMPILE IssueInfo, skip rules │
│   ├─ missing import → per-file fallback read, legacy error path (2.3/2.4)│
│   └─ ok → protoast.File (targets) + ImportedFiles map                   │
└──────────────────────────────────────────────────────────────────────────┘
  ↓
comment suppressions from protoast comments (same directives, REQ-2.5)
  ↓
┌─ Pass 2 ────────────────────────────────────────────────────────────────┐
│ per file: ignoreOnly filter → File rules Validate → ModuleRule.Collect  │
│ after sweep: ModuleRule.Finish (e.g. PACKAGE_NO_IMPORT_CYCLE)           │
│ filter: allowCommentIgnores × suppressions.contains(rule, line)         │
└──────────────────────────────────────────────────────────────────────────┘
  ↓ deterministic ordering (same output shape as main — REQ-5.1)
[]IssueInfo → api output/exit code (unchanged format)
```

### Breaking (legacy path, unchanged) — diverges at the model layer

```
api.Breaking → buildCore (breaking config only)
  ↓
readProtoFiles (WalkDir) → protoInfoRead (go-protoparser)   ← shared entry RETAINED for breaking
  ↓ collect → ProtoData (unordered.Proto collections)
BreakingChecker.Check → []IssueInfo (constructs Position values)
```

### The divergence point (critical boundary)

Both paths used to share `protoInfoRead` + `ProtoInfo{Info *unordered.Proto}`. After this milestone they must not. The cut happens **before** the engine flip:

| Shared artifact | Today | After decoupling |
|---|---|---|
| `protoInfoRead` | lint + breaking | breaking only (file/ownership becomes explicit) |
| `ProtoInfo` | lint + breaking | split: breaking keeps it (or a breaking-local rename); lint uses a lint-facing struct with `File *protoast.File` + `ImportedFiles` |
| `Issue.Position` (`meta.Position`) | lint + breaking output | becomes `protoast.Position` (lab: `core.Position = protoast.Position`). Plain line/column struct → breaking's construction sites adapt mechanically; keep `IssueInfo` the single output type so the CLI format layer stays untouched |
| `CheckIsIgnored(comments []*parser.Comment, …)` | exported, called by rules | protoast-comment equivalent; suppression semantics (incl. main's newer `allowCommentIgnores` behaviors) reconciled per REQ-3.2/2.5 |
| `internal/protosource`, path helpers | audit | verify no lint-path dependence on go-protoparser types remains |

**Direction rule:** `internal/api → core → protoast → easypc`, and `rules → core + protoast` (never rules → easypc; never protoast → core). `easypc` imports nothing above it.

## Scaling Considerations

| Scale | Architecture adjustments |
|-------|--------------------------|
| ~100 files (api-common) | Windows degenerate to one; sequential-equivalent behavior; correctness dominated by parity, not perf |
| ~7.5k files (googleapis-nocompute) | Windows of 256 + import cache are the design point (lab: full corpus seconds-scale; main: ~20 min on 4 MB `compute.proto`). Perf gate is "not slower than main" (REQ-6.1) |
| 10k+ / monorepo corpora | Memory becomes the constraint: syntax trees retained for the whole linked graph. If needed later: link-per-window with a shared immutable import cache (targets' trees released after rules run per window) — do **not** build this now (premature; lab numbers don't require it) |

Scaling priorities: (1) import cache hit rate — resolution dominates; (2) window size vs. allocation churn (256 is lab-validated; tune only with corpus benchmarks); (3) `Finish`-phase graph rules on huge module sets — fine at current corpus sizes.

## Anti-Patterns

### Anti-Pattern 1: Bridging the old rule interface with a converted AST

**What people do:** Write an adapter that builds a fake `*unordered.Proto` from protoast so old rules compile against the new engine during migration.
**Why it's wrong:** The old contract is *direct field access* into go-protoparser's tree; the adapter is a second parser with its own position/comment bugs — and it exists only to be deleted. It doubles the parity surface instead of shrinking it.
**Do this instead:** Hard switch; port rules in batches with colocated goldens (Patterns 1–3).

### Anti-Pattern 2: Runtime dual-engine flag in the shipped CLI

**What people do:** `--lint-engine=legacy|new` (or env var) so users can fall back.
**Why it's wrong:** Output format is a contract (REQ-5.1); two engines mean twice the suppression/positioning bug surface in the field, and CI pipelines pin whichever flag they first saw. Rollback via git revert of one flip commit is strictly safer.
**Do this instead:** Dev-time shadow harness; atomic flip; revert = rollback.

### Anti-Pattern 3: Long-lived dual models inside one path

**What people do:** New engine, but a few "hard" rules keep secretly reading the old parser "temporarily".
**Why it's wrong:** Dual models *across* command boundaries are the accepted, temporary cost of this milestone; dual models *within* the lint path never converge and drift apart (each bug fix lands on one model only).
**Do this instead:** All 43 rules port before the flip; nothing in the lint path touches go-protoparser afterward.

### Anti-Pattern 4: Rules importing easypc directly

**What people do:** Rules walk `easypc` syntax trees because it's "right there".
**Why it's wrong:** Couples 43+ rules to compiler internals; the next compiler sync from compiler-lab (or the breaking/generate migration) rewrites every rule.
**Do this instead:** Rules see only `protoast`; missing view capabilities get added to protoast (walk/comments/outline).

### Anti-Pattern 5: Suppression logic leaking into rules

**What people do:** Each rule checks `ignore`/comment directives itself.
**Why it's wrong:** Suppression semantics are the engine's contract (REQ-2.5); per-rule reimplementations diverge and parity-testing them individually is impossible.
**Do this instead:** Engine filters `Validate`/`Finish` results centrally (buf does exactly this via the framework + SourceCodeInfo), as `lint.go:57-64` already does today.

### Anti-Pattern 6: Non-deterministic finding order from parallelism

**What people do:** Append findings from worker goroutines in completion order.
**Why it's wrong:** Output diff vs. main becomes flaky → parity harness noise; REQ-2.1/5.1 effectively pin traversal-order-stable output.
**Do this instead:** Per-window indexed results merged in walk order (mirror `internal/parallel.Run`'s index-ordered error contract), final sort at assembly.

### Anti-Pattern 7: Trusting lab rule semantics over main

**What people do:** Copy lab rules 1:1 (they're "already proven").
**Why it's wrong:** Lab rules were cut from an older main; main evolved (e.g., `allowCommentIgnores` comment-suppression work — see `check_lint_ignore_test.go`). REQ-3.2: main wins.
**Do this instead:** Port each rule by reconciling lab's protoast mechanics with main's current semantics; the colocated tests port from **main**, positions/messages included.

## Integration Points

### External Systems

| System | Integration pattern | Notes |
|--------|---------------------|-------|
| compiler-lab (private) | Source of vendored `easypc`/`protoast`; reference CLI for new-rule parity (REQ-4.3); `bench/lab.py` harness + corpora | Access via `gh` assumed for the milestone duration; corpora rebuildable from googleapis (`bench/setup.sh`); 2.9 GB parity outputs not stored in repo |
| Buf compatibility contract | COMPILE issues, exit codes, output format mirror `buf lint` failure modes | The reason missing-import falls back to legacy per-file behavior (buf treats unresolvable imports differently from compile errors) |
| WKT (well-known types) | Embedded imports via easypc/protoc (lab) | Same role as `wellknownimports.WithStandardImports` in the generate path |

### Internal Boundaries

| Boundary | Communication | Considerations |
|----------|---------------|----------------|
| api ↔ core | Constructor injection (`core.New(Options)`, `buildCore`) | Flip commit lives here; no new Options knobs should be needed |
| core ↔ protoast | Direct calls (`protoast.Build` from linked trees) | One-directional; protoast must not import core |
| core/link.go ↔ easypc | `protoc.Link` + resolver callbacks | Only core file touching easypc/protoc; error classification (compile vs missing-import vs read) stays here |
| lint path ↔ breaking path | **None after decoupling** — shared only via `IssueInfo` output type + config plumbing | The deliberate temporary duplication (two parsers, two models) is the price of REQ-5.2; removed by future breaking migration |
| engine ↔ rules | `Rule`/`ModuleRule` interfaces + `IssueInfo` results | Contracts in `dom.go`; keep narrow so future rule authors need only protoast |
| parallel windows ↔ import cache | Concurrent map reads after warm-up, synchronized inserts | Cache shared across windows; correctness detail for the engine phase |

## Build Order (roadmap implications)

Dependency-driven order; each step leaves `task test`/`task lint` green:

1. **Vendor compiler** — `internal/easypc` + `internal/protoast` copied from lab, imports rewritten, `go.mod` unchanged (REQ-1.1/1.2). Pure addition; nothing wired.
2. **Decouple the shared model** — breaking retains `protoInfoRead`/`ProtoInfo`; lint-facing types introduced; `Issue.Position` → `protoast.Position` with breaking construction sites adapted; `CheckIsIgnored`/suppression helpers re-based. Zero behavior change; REQ-5.2 becomes reviewable as "breaking diff is empty".
3. **New engine, unwired** — `link.go` (resolver, COMPILE classification, missing-import fallback) + two-pass `lint.go` (windows, ModuleRule, deterministic order) + `lint_helpers.go`/`casing.go` vendored; exercised by engine tests with a handful of ported rules.
4. **Batch-port 43 rules** — recommended batching by config group (MINIMAL → BASIC → DEFAULT → COMMENTS → UNARY_RPC): groups map to schema surfaces, complexity tiers, and natural review checkpoints; each batch = rules + colocated goldens reconciled against main (REQ-3.1/3.2). Cross-file/new rules excluded (they're new, not ports).
5. **The flip** — single commit wires the new engine into `easyp lint` and deletes the old lint loop + old rule implementations. This is the only user-visible behavior change point (plus new COMPILE issues, already user-approved).
6. **New rules + schemas** — `PACKAGE_NO_IMPORT_CYCLE` (MINIMAL), `SYNTAX_SPECIFIED`, `FIELD_NOT_REQUIRED` (BASIC); `task schema:generate` + MCP description updates (REQ-4.1–4.4; parity vs lab CLI per REQ-4.3).
7. **Acceptance** — corpus parity via `lab.py lint-diff` (api-common + googleapis-nocompute), perf gate on googleapis-nocompute (REQ-6.1), `task test`/`task lint`/`task schema:check`.

Steps 1–2 are safe to merge early and independently; 3–4 carry on the feature branch; 5 is the release point. Defer to future milestones: go-protoparser removal, breaking/generate/tidy/migrate migration, streaming lint.

## Sources

Project sources (HIGH — read directly, commit `5c9185f`):
- `/home/nil/DiskD/W/Djarvur/easyp/internal/core/lint.go`, `dom.go`, `core.go`, `breaking_check.go`, `proto_info_read.go`, `generate.go` (esp. protocompile resolver pattern at generate.go:151)
- `/home/nil/DiskD/W/Djarvur/easyp/internal/rules/builder.go`, `internal/parallel/parallel.go`
- `/home/nil/DiskD/W/Djarvur/easyp/.spec/features/lint-easypc-infra/explore.md` and `requirements.md` (REQ numbering used throughout)
- `/home/nil/DiskD/W/Djarvur/easyp/.planning/codebase/ARCHITECTURE.md`, `/home/nil/DiskD/W/Djarvur/easyp/.planning/PROJECT.md`

External sources (MEDIUM per confidence seam — content is primary-source where noted):
- bufbuild/protocompile README, `compiler.go`, `resolver.go` (via Context7 + pkg.go.dev) — pipeline stages, `Compiler`/`RetainASTs`, resolver chain
- bufbuild/buf `private/bufpkg/bufcheck/bufcheckserver/internal/bufcheckserverutil/lint.go`, `rule_spec_builder.go`, `bufcheckserverbuild.go`, `protosourcepath/README.md` (fetched from upstream main) — rule handler scopes, framework-owned suppression
- Strangler fig: learn.microsoft.com/en-us/azure/architecture/patterns/strangler-fig; docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/strangler-fig.html
- Feature parity: martinfowler.com/articles/patterns-legacy-displacement/feature-parity.html
- Shadow testing: microsoft.github.io/code-with-engineering-playbook/automated-testing/shadow-testing/

Tooling note: built-in WebSearch/WebFetch were hook-blocked in this session; the cc-websearch plugin search script worked, its webfetch script failed (missing plugin asset) — primary sources were fetched via curl/gh instead.

---
*Architecture research for: EasyP lint-easypc-infra (compiler-backed lint engine integration)*
*Researched: 2026-11-17*
