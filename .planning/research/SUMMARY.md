# Project Research Summary

**Project:** EasyP — milestone `lint-easypc-infra`
**Domain:** Compiler-backed protobuf lint engine (parser/engine migration inside a Go CLI)
**Researched:** 2026-11-17
**Confidence:** HIGH

## Executive Summary

This milestone migrates `easyp lint` from the stalled `go-protoparser` (last push ~12 months ago, parser-only) to the project's own validated in-repo compiler, vendored from `compiler-lab` (branch `easyp-cli`) as `internal/easypc` (~13 packages, ~3.4 MB) plus the purpose-built lint view model `internal/protoast` (~1000 LOC). The hard constraint REQ-1.2 — zero new `go.mod` entries — is satisfied by plain copy-in; `go-protoparser` и `protocompile` stay for breaking/generate/mod-tidy (REQ-5.2), which are explicitly untouched. The lab already proved the approach: identical findings on acceptance corpora and 20 min → 0.08 s on a 4 MB file.

The recommended approach is a hard switch at the command boundary (strangler by command, not by rule), with expand–contract sequencing on the branch: vendor → decouple the shared `protoInfoRead`/`ProtoInfo` model (breaking keeps its own path) → build the new two-pass engine unwired (`link.go` + windowed `lint.go`, `ModuleRule` Collect/Finish for cross-file rules) → batch-port the 43 rules with colocated golden tests → one flip commit rewires `easyp lint` and deletes the old loop. Then add the three new rules (`PACKAGE_NO_IMPORT_CYCLE` in MINIMAL, `SYNTAX_SPECIFIED`/`FIELD_NOT_REQUIRED` in BASIC) with schema/MCP regeneration, and gate on corpus parity (`lab.py lint-diff` on api-common 92 files + googleapis-nocompute 7414 files) plus "not slower than main" (REQ-6.1).

Key risks are parity fidelity, not architecture. The dominant ones: position conventions (main is 1-based line / 1-based **rune** column; compiler spans are 0-based) — detect with a UTF-8 fixture before any rule batch starts; comment-suppression parity (reflection-coupled attachment must be rebuilt in protoast; port `comment_suppressions*_test.go` fixtures); the COMPILE semantic cliff (three-way error router: compile error → COMPILE issues, rules skipped; missing import → legacy fallback + exit 2 via typed `OpenImportFileError`; syntax/read → command error); stateful rule structs under the parallel engine (audit receivers, `-race` engine test); and rule registry drift (one registry feeds config validation, JSON Schemas, MCP, and comment-directive validation; `ValidateNames` currently hard-rejects `PACKAGE_NO_IMPORT_CYCLE` and must flip).

## Key Findings

### Recommended Stack

See [STACK.md](STACK.md). Nothing is installed or added — the milestone's defining property is a zero-diff `go.mod`/`go.sum`. Go 1.26.6, testify v1.11.1, golangci-lint v2.14.0, gotestsum v1.13.0 stay pinned; no mid-milestone bumps.

**Core technologies:**
- `internal/easypc` (vendored from compiler-lab) — lexer/parser/linker/protoc API + embedded WKTs; validated, zero external deps, imports already rewritten
- `internal/protoast` — linter-facing AST model (build/walk/comments/outline); the stability seam so rules never touch compiler internals
- Two-pass lint engine per lab design — `protoc.Link` with `WithSyntaxTrees(true)`, import cache, 256-file parallel windows via existing `internal/parallel`, `ModuleRule` (Collect/Finish)
- protocompile v0.14.1 — stays in go.mod for other subsystems; **parity oracle only** (COMPILE issue shape), never in the lint path
- Go native fuzzing (`testing.F`) — follow-up hardening for the hand-written parser, not a blocker

### Expected Features

See [FEATURES.md](FEATURES.md).

**Must have (table stakes):**
- COMPILE diagnostics inside lint output in the frozen format `path:line:col:source message (COMPILE)` — REQ-2.2
- Missing-import fallback preserving legacy behavior and exit code 2 — REQ-2.3/2.4
- Byte-identical 43-rule findings (rule, file, line, column) on acceptance corpora — REQ-3.1, reconciled against **current main** (main wins over lab per REQ-3.2)
- Suppression/ignore/config-validation parity (`easyp:disable`/`nolint`/`buf:lint:ignore` dialects, `ignore`/`ignoreOnly`, unknown-name rejection) — REQ-2.5/2.6
- Deterministic, walk-order-stable output under parallelism; frozen exit codes 1/2 — REQ-2.1/5.1
- Cross-file rules via linking (the milestone's reason to exist; protolint cannot do them)

**Should have (differentiators):**
- Zero new external deps / single binary; lab's 250x speed as upside, not a gate
- Richer suppression vocabulary than buf (block scopes)
- Buf-matching category placement for new rules so shared policies get equivalent defaults

**Defer (v2+):** breaking/generate/tidy/migrate migration (then delete go-protoparser), config-ignore-yaml output, extra output formats, auto-fix/rule plugins/persistent cache (anti-features for now).

### Architecture Approach

See [ARCHITECTURE.md](ARCHITECTURE.md). Three-layer shape mirroring buf/protocompile: engine orchestration → rule layer (rules report only, never suppress/print) → compiler/view layer. Direction rule: `api → core → protoast → easypc`; rules see only protoast; только линкер-модуль в core касается `easypc/protoc`.

**Major components:**
1. `internal/easypc` — sealed vendor tree, imports nothing from easyp; provenance in VENDORED.md
2. `internal/protoast` — lint view model from linked syntax trees (buf's `bufprotosource` role)
3. Linker module (`link.go`, новый файл в `internal/core`) — resolver over import roots + `OpenSourceFile` + WKT, import cache, error classification (COMPILE vs missing-import vs read)
4. `internal/core/lint.go` (rewritten) — two-pass windowed engine, ModuleRule phases, central suppression filter, deterministic ordering
5. `internal/rules` — 43 rewritten + 3 new rules, colocated testify tests
6. Legacy path retained — `protoInfoRead`/`ProtoInfo` become breaking-owned; `Issue.Position` becomes `protoast.Position`

**Anti-patterns to enforce:** no bridge adapter, no runtime dual-engine flag, no rule reading go-protoparser or easypc directly, no suppression logic in rules, no completion-order output.

### Critical Pitfalls

See [PITFALLS.md](PITFALLS.md) (9 pitfalls, all repo-verified).

1. **Position fidelity** — 0-based compiler spans vs 1-based legacy output with **rune** columns; fix `core.Position = protoast.Position` once, with a multi-byte UTF-8 fixture, before any rule batch
2. **Comment-suppression parity** — main's attachment is reflection-coupled to go-protoparser; port fixtures as executable parity tests, feed `knownLintRules` from the real registry
3. **COMPILE semantic cliff** — three-way error router with a fixture matrix; assert exit codes through the CLI Action path; missing import must still surface typed `OpenImportFileError` (exit 2)
4. **Stateful rules + parallel engine** — audit all 43 receivers for field writes; `-race` engine test over the parallel path
5. **Parity verification gaps / lab staleness** — batch-level old-vs-new parity tests per rule batch (corpus diff localizes to files, not rules); reproduce `lab.py` harness early as a spike
6. **Registry drift** — one registry feeds config validation, schemas, MCP, comment directives; land rule + `task schema:generate` + MCP descriptions in one commit; remove the `PACKAGE_NO_IMPORT_CYCLE` rejection special case
7. **Dual-model boundary** — fork lint's read path from breaking's; route the linker through `importRoots` + `importFileAllowed` + `wellknownimports` (bypassing it is a data-exposure divergence)

## Implications for Roadmap

The build order below follows the dependency chain REQ-1 → REQ-2 → REQ-3 → REQ-4 → REQ-5/6, with each step leaving `task test`/`task lint` green.

### Phase 1: Vendor the compiler and decouple the shared model
**Rationale:** Pure addition first (`go.mod` unchanged, REQ-1.1/1.2), then cut the lint↔breaking coupling while behavior is unchanged — this makes REQ-5.2 ("breaking untouched") reviewable as an empty diff before any engine work.
**Delivers:** `internal/easypc` + `internal/protoast` copy-in with provenance (VENDORED.md, upstream SHA), `go mod tidy` empty diff, `Issue.Position → protoast.Position` with breaking adapted, breaking-owned `protoInfoRead`, position-fidelity UTF-8 fixture proving 1-based/rune-column semantics.
**Addresses:** REQ-1.1/1.2; foundation for everything.
**Avoids:** P1 (position offsets caught before rule batches), P9 (vendor drift — hygiene established in the creating commit), P5 (fork is deliberate).

### Phase 2: New engine, unwired — link.go + two-pass lint.go
**Rationale:** The engine must exist and be test-constructed before rules can port against it; the error router is the highest-risk correctness surface and needs its fixture matrix before the flip.
**Delivers:** linker-модуль в `internal/core` (resolver chain incl. `importFileAllowed` + WKT, import cache, three-way error routing with COMPILE/missing-import/read fixtures), rewritten `lint.go` (256-file windows, ModuleRule Collect/Finish, index-ordered result assembly, central suppression filter re-based on protoast), `lint_helpers.go`/`casing.go` vendored, suppression fixture suite mirrored, `-race` parallel engine test.
**Uses:** easypc `protoc.Link`, `internal/parallel`, protoast.
**Avoids:** P2, P3, P4, P6.
**Note:** resolve the documented open design question — new engine's within-rule emission order (main is map-random; deterministic sorting is a visible change) — via `/gsd-discuss-phase`.

### Phase 3–k: Batch-port the 43 rules (grouped MINIMAL → BASIC → DEFAULT → COMMENTS → UNARY_RPC)
**Rationale:** Rules port in batches mapped to config groups (schema surfaces, complexity tiers, review checkpoints); each batch carries old-vs-new parity tests on shared fixtures so corpus acceptance at the end is a gate, not a debugging session. Colocated tests port from **main**, not the lab (REQ-3.2).
**Delivers:** All 43 rules on protoast with identical findings, reconciled against current main semantics; per-rule receiver-write audit (statelessness).
**Avoids:** P4, P7.
**Suggestion:** one roadmap phase per group (5 phases) or fewer if the roadmapper prefers coarser batches — the invariant is batch-level parity verification, not the batch size.

### Phase 4: The flip — wire new engine, delete old lint path
**Rationale:** Single commit at the `buildCore` seam; the only user-visible change point; rollback = revert.
**Delivers:** `easyp lint` on the compiler; old lint loop and old rule implementations deleted; breaking/generate/mod-tidy/migrate suites green in the same CI (REQ-5.2 gate).

### Phase 5: New rules + schema/MCP regeneration
**Rationale:** Rides on the finished engine; registry edits must land atomically with schemas and MCP descriptions.
**Delivers:** `PACKAGE_NO_IMPORT_CYCLE` (MINIMAL, ModuleRule-based), `SYNTAX_SPECIFIED` + `FIELD_NOT_REQUIRED` (BASIC), `ValidateNames` special-case removal, `task schema:generate`/`schema:check`, MCP description updates, parity vs lab CLI (REQ-4.1–4.4).
**Avoids:** P8.

### Phase 6: Acceptance — corpus parity + performance gate
**Rationale:** Final gate only works after all rules are ported (earlier measurement compares a hybrid path).
**Delivers:** `lab.py lint-diff` on api-common + googleapis-nocompute with sorted-set comparison (run twice for flakiness), main-vs-main baseline, REQ-6.1 not-slower measurement on 7414 files with CI-equivalent flags, `task test`/`lint`/`schema:check`.
**Avoids:** P3, P7 residual, performance traps.

### Phase Ordering Rationale

- Dependencies are strict: linking (Phase 2) precedes COMPILE and cross-file rules; all rules ported precede meaningful perf measurement; schema enums precede config validation of new rules.
- Grouping by architecture seam: vendor tree / view model / engine / rules / flip / new rules / acceptance each change one boundary, keeping every intermediate state green.
- Batch-level parity (Phase 3–k) is the pitfall-research answer to "corpus diff can't localize" — it must be a per-phase verification, decided at roadmap time.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 2 (engine):** hardest design decisions — error-router edge cases, suppression attachment algorithm in protoast, within-rule ordering spec, ModuleRule concurrency contract. Recommend `/gsd-plan-phase --research-phase`.
- **Phase 6 (acceptance):** `lab.py` harness reproduction is a flagged spike (private repo access, corpora rebuild); plan a spike within or before this phase.

Phases with standard patterns (skip research-phase):
- **Phase 1 (vendoring/decoupling):** mechanical copy + provenance; procedure fully specified in STACK.md.
- **Phase 3–k (rule batches):** per-rule reconcile against main; convention-driven (colocated tests), no unknowns beyond those Phase 2 settles.
- **Phases 4, 5:** single-commit flip and additive rule landing; patterns established.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Library status from GitHub API/official notes; approach validated by the in-repo lab decision record. Web-search survey breadth was LOW (DuckDuckGo empty) but the recommendation doesn't depend on it |
| Features | HIGH | buf behaviors cross-verified docs vs source (`bufprotocompile.go`, exit code 100); EasyP current behavior read at first-party commits |
| Architecture | HIGH | Project structure verified against source at `5c9185f`; external patterns corroborated against protocompile/buf source and strangler/parity literature |
| Pitfalls | HIGH | Nearly all pitfalls repo-verified in source/module cache; positions verified against pinned go-protoparser v4.14.2 |

**Overall confidence:** HIGH

## Gaps to Address

- **`lab.py` harness availability:** requires private compiler-lab `gh` access and corpus rebuilds — validate in a spike before the acceptance phase depends on it.
- **Within-rule emission order of the new engine:** main is nondeterministic (map iteration); a deterministic sort is a visible change. Open design question — decide in `/gsd-discuss-phase` for Phase 2.
- **`internal/parallel` main-vs-lab compatibility:** flagged assumption in explore.md; verify during Phase 2 design.
- **Editions support in easypc:** unverified (acceptance corpora are proto2/proto3). If user corpora adopt editions, that is a future milestone — flag in the roadmap, do not scope in.
- **Peak RSS on the 7414-file corpus:** memory under retained syntax trees unmeasured at main scale; measure during acceptance.

## Sources

### Primary (HIGH confidence)
- Repo source at `5c9185f`/`5ebab04`: файлы lint / comment_suppressions / dom / proto_info_read / breaking_check / generate в `internal/core`, `internal/rules/*`, `internal/config/v1/schema.go`, `internal/api/lint.go`, `internal/parallel/parallel.go`, `.spec/features/lint-easypc-infra/*`, `.spec/ERRORS.md`
- go-protoparser v4.14.2 module cache (`meta/position.go`, `parser/import.go`) — position contract
- GitHub API: bufbuild/protocompile, yoheimuta/go-protoparser, jhump/protoreflect releases/metadata; buf CLI public go.mod
- jhump/protoreflect v2 release notes (parser removed; Go 1.26 floor)
- bufbuild/buf source: `bufprotocompile.go` (COMPILE annotations), `lint.go` command flow, `bufctl.go` (exit 100)

### Secondary (MEDIUM confidence)
- buf.build docs (via Context7): lint usage, rule categories, error-format semantics
- Context7 `/bufbuild/protocompile`: Compiler API, resolver/reporter semantics, editions status
- protolint README (single primary source)
- protobuf.dev: SourceCodeInfo 0-based spans
- Strangler fig (MS Learn, AWS), Fowler feature-parity, MS shadow-testing guidance

### Tertiary (LOW confidence)
- DuckDuckGo ecosystem-survey searches returned empty — survey breadth unverified; recommendation independent of it

---
*Research completed: 2026-11-17*
*Ready for roadmap: yes*
