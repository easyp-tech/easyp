# Pitfalls Research

**Domain:** Parser/engine migration of a shipping protobuf linter (go-protoparser → in-repo `easypc` compiler + `protoast` model) with byte-exact issue parity
**Researched:** 2026-11-17
**Confidence:** HIGH (repo-verified pitfalls from source at `5c9185f`; MEDIUM where only ecosystem sources apply — flagged per item)

Sources of truth: main repo source, the exact pinned `go-protoparser v4.14.2` from the module cache, `.spec/features/lint-easypc-infra/`, `.planning/codebase/`, buf/protocompile/protobuf.dev docs. Every "verified" claim below was read in source, not recalled.

---

## Critical Pitfalls

### Pitfall 1: Position fidelity — 0-based compiler spans vs 1-based legacy output, rune columns, and token choice

**What goes wrong:**
The new engine's issues are systematically off by one line/column, or point at the wrong token (path literal vs `import` keyword, semicolon vs declaration start). Parity looks fine on a hand-checked sample and fails on the corpus.

**Why it happens:**
Three independent conventions collide:
- `go-protoparser v4.14.2` `meta.Position` (the contract users see today): `Line` 1-based, `Column` 1-based **counted in characters (runes)**, `Offset` 0-based bytes (verified in `parser/meta/position.go`; the lexer scans `[]rune`). A byte-offset-derived column diverges on any line containing multi-byte UTF-8 (license headers, non-English comments) even when all ASCII lines match.
- Descriptor/compiler-side coordinates are **0-based**: protobuf.dev documents `SourceCodeInfo.Location` spans as zero-based line/column ("add 1 before displaying"). A linker-derived model (`easypc` syntax trees, protoast) naturally produces 0-based or half-open spans; forgetting the +1 shifts every position.
- `Meta.Pos` of a statement is the position of its **first token** — for an import, the `import` keyword (or `public`/`weak` modifier), not the string literal; `LastPos` is the trailing semicolon (verified in `parser/import.go`). Each rule picks a specific node's position; a rewritten rule that grabs a different node (e.g. the import's location token) emits correct-looking but different columns.

**How to avoid:**
- Make `core.Position = protoast.Position` reproduce go-protoparser semantics exactly: 1-based line, 1-based rune column, statement-first-token anchor. Write a dedicated position-fidelity test: one fixture with a multi-byte UTF-8 comment above each declaration kind, asserting exact line:col for every rule that fires.
- Treat `task`-run corpus lint-diff (`compiler-lab/bench/lab.py`) as the only acceptance for positions; 92-file api-common alone will not exercise the long tail.

**Warning signs:**
- Corpus diff shows uniform +1/-1 deltas, or same-line column shifts only on files with non-ASCII comments.
- New engine's internal positions disagree with `Issue.Position` after adaptation (adapter doing silent arithmetic).

**Phase to address:**
Phase 1/2 boundary — when vendoring `protoast` and wiring `core.Position`, before any rule is ported. Detection must exist before rule batches start, or every batch inherits the offset.

---

### Pitfall 2: Comment-suppression parity — attachment semantics are parser-specific and reflection-coupled

**What goes wrong:**
Suppressions (`easyp:disable`/`easyp:enable`, `nolint:`, `buf:lint:ignore`) stop covering the same lines: issues that main suppressed now appear (corpus noise), or issues main reported vanish. Suppression *validation errors* (malformed directive, unknown rule name in directive, `enable` without `disable`) fire on files that main accepted, turning a lint run into a hard error.

**Why it happens:**
Main's suppression machinery is deeply coupled to go-protoparser:
- `internal/core/proto_comments.go` parses with `WithBodyIncludingComments(true)` specifically to retain **EOF and pre-closing-brace comments**, collects them from the ordered AST via a reflection walk (`collectDirectiveComments` in `internal/core/comment_suppressions.go`), then strips standalone nodes before unordered interpretation.
- Attachment spans come from each declaration's `Meta` (`Pos.Line..max(Pos.Line, LastPos.Line)`) — the reflection walk reads `Meta`, `Comments`, `InlineComment`, `InlineCommentBehindLeftCurly` fields by name. protoast has none of these fields; its comment attachment must be rebuilt.
- Buf's own documentation shows how subtle the rules are: a comment attaches as "leading" only if it ends on the same line or the line before the token, block trailing comments attach to the **opening** brace, comments after a closing brace are discarded. Two parsers rarely attach identically.
- The `known[name]` check rejects comment directives naming rules the engine doesn't know — if the new engine's known-set differs (e.g. lab's rule list), directives for rules added in main (or the new rules) hard-error.

**How to avoid:**
- Port the suppression fixtures as executable parity tests, not just the code: main already has `comment_suppressions_test.go` and `comment_suppressions_eof_test.go` — mirror every case against protoast-built `ProtoInfo`.
- Decide the attachment algorithm in protoast explicitly (leading = same/previous line, declaration span from node start to node end) and document divergence from the reflection walk.
- Feed `knownLintRules` from the same registry that feeds config validation (see Pitfall 8) so directive validation can't drift from engine reality.

**Warning signs:**
- Corpus diff noise concentrated in files with doc-comment blocks, block comments, or trailing `//` after declarations.
- Suppression-validation errors on corpora that main linted cleanly.

**Phase to address:**
Phase 2 (engine switch) for the machinery; every rule-batch phase re-runs the suppression fixtures, because rules emit the line numbers that suppressions match against.

---

### Pitfall 3: Determinism — output order is walk-order × rule-order with no final sort, and current within-rule order is already map-random

**What goes wrong:**
Two failure directions: (a) the parallel engine emits issues in completion order, so identical corpora produce different byte streams run-to-run — CI diffs and `lint-diff` snapshots become flaky; (b) the new engine *sorts* output and the parity diff against main flags order differences that are not regressions, or conversely the harness ignores order that users actually consume.

**Why it happens:**
- `textPrinter` in `internal/api/lint.go` prints issues strictly in append order; there is no sort anywhere in the lint path. Append order = `WalkDir` order × `c.rules` order × within-rule emission order.
- Main is **already nondeterministic** within a rule: `ImportUsed.Validate` emits unused-import issues by ranging over the `isImportUsed` map (Go randomizes map iteration order on purpose — spec: "The iteration order over maps is not specified"). With 2+ unused imports in one file, main's own output order varies run to run.
- The lab engine parallelizes in windows of 256 files. `internal/parallel.Run` (present in main, verified) returns errors in index order — results must be reassembled the same way, or the windowed speedup leaks completion-order output.
- REQ-3.1 defines parity as set equality (rule, file, line, column) — correct — but the `lab.py` harness and any golden snapshots must implement that comparison (sort both sides or set-compare), and the roadmap must state whether the new engine's emission order is specified.

**How to avoid:**
- Reassemble parallel results by window index before printing; keep walk order and rule order stable across the switch.
- Make `lab.py` lint-diff compare sorted line sets, and run main-vs-main once to measure main's own flakiness baseline before trusting new-vs-main diffs.
- Specify (and test) the new engine's within-rule order — deterministic sorting is an improvement, but it is a visible change vs main's random order; document it rather than discover it in acceptance.

**Warning signs:**
- Two runs of the *same* binary produce different corpus output (order flakiness).
- Parity diffs that vanish on re-run.

**Phase to address:**
Phase 2 (engine design decision) and the acceptance phase (harness semantics); flag as a `/gsd-discuss-phase` topic — it is one of the documented open design questions.

---

### Pitfall 4: Stateful rule structs + parallel engine = data races and cross-file contamination

**What goes wrong:**
Rules carry per-file mutable state on the receiver; a parallel engine invokes the same shared rule instance concurrently → race (crash or corrupted findings), or a restructured lifecycle (ModuleRule `Collect`/`Finish`) silently changes what a rule sees across files.

**Why it happens:**
- Verified: `ImportUsed` (in `internal/rules/import_used.go`) assigns `i.instrParser`, `i.isImportUsed`, `i.pkgToImport` inside `Validate`. Safe today only because main runs rules sequentially. The rules slice holds one instance per rule shared across all files.
- The lab engine's two-pass design and `ModuleRule` change the call contract; a rule ported mechanically keeps per-file state in fields that now live across a `Collect` boundary.
- `task test` runs with `-race`, which catches this — but only if tests drive the *parallel* path with the stateful rules, not the sequential per-rule unit tests.

**How to avoid:**
- Audit all 43 rules for receiver-field writes in `Validate` before the engine switch (grep for `i\.` / receiver assignments); make rules stateless or move state into engine-owned per-file context.
- Add one engine-level test that runs the full rule set over a multi-file fixture through the parallel path under `-race`.
- For `ImportUsed` specifically, decide where per-file maps live in the new two-pass engine and pin behavior with a parity test (it also depends on `ProtoFilesFromImport` semantics — see Pitfall 5).

**Warning signs:**
- `go test -race` failures appearing only after the engine merge; flaky parity diffs that change issue *counts* (not just order).

**Phase to address:**
Phase 2 (engine) + each rule-batch phase (lifecycle audit is per-rule work).

---

### Pitfall 5: Dual-model coexistence — `protoInfoRead` feeds breaking too, and the import boundary must not diverge

**What goes wrong:**
The lint migration accidentally breaks `easyp breaking`, generate, or the module-selection boundary because they share the parse path. Or lint's new linker resolves imports from places main's lint could not read (and vice versa), changing which files compile.

**Why it happens:**
- Verified: `internal/core/breaking_check.go:119` calls the same `c.protoInfoRead` and consumes `Info *unordered.Proto` / `ProtoFilesFromImport`. The lab solved this by migrating breaking too (`compilerbackend`) — explicitly out of scope here (REQ-5.2). So the read path must **fork**: lint builds protoast from linked trees; breaking keeps go-protoparser. Touching the shared `protoInfoRead` signature instead of forking compiles nowhere; weakening it to serve both masters creates a lowest-common-denominator model.
- `proto_info_read.go` embeds policy logic the new linker must reproduce exactly: `importRoots` search order, `wellknownimports` embedded WKT fallback, and `importFileAllowed` — the gate restricting imports to files selected by dependency metadata. A linker that opens any file under a root silently bypasses that boundary (files outside the module's locked selection get read — a correctness *and* data-exposure divergence); one that is stricter makes previously-compiling corpora fail with COMPILE.

**How to avoid:**
- Fork deliberately: lint gets its own linked-tree reader; `protoInfoRead` stays for breaking untouched. Verify REQ-5.2 with breaking/generate/`mod tidy`/`migrate` test suites in the same CI run as the engine switch.
- Wire the linker's resolver through the same `importRoots` + `importFileAllowed` + `wellknownimports` chain (protocompile's `SourceResolver.Accessor` pattern is the natural seam — its docs require thread-safe accessors under concurrent compilation).
- Expect double-parsing cost while both models exist (lint run and breaking run each parse) — accepted by scope, but measure so REQ-6.1 isn't judged on a combined run.

**Warning signs:**
- Breaking/generate tests fail after engine merge; corpus files that main linted fine now produce import-resolution COMPILE errors (or files that main refused now lint).

**Phase to address:**
Phase 2 (fork the read path) + Phase 5 (REQ-5.2 compatibility acceptance as an explicit gate).

---

### Pitfall 6: The COMPILE semantic cliff and three-way error routing

**What goes wrong:**
Exit codes and output volume flip on corpora that don't fully link: users who got lint findings (exit 1, many issues) now get a handful of `COMPILE` lines; users who got exit 2 "Cannot import file" now get exit 1; or vice versa. CI gates keyed on exit codes (documented: lint issues → 1, `*core.OpenImportFileError` → 2 via `errors.As` in `internal/api/lint.go`) change meaning silently.

**Why it happens:**
The new behavior (user-approved) needs a precise three-way router (REQ-2.2/2.3/2.4):
1. Linking failure that is **not** a missing import → `COMPILE` issues for those files, **no rules run on the corpus**.
2. Linking impossible due to **missing import** → fall back to per-file parsing and report exactly as before — which means surfacing the typed `*core.OpenImportFileError` (exit 2, "Cannot import file"), not a wrapped linker error. `.spec/ERRORS.md` is explicit that the mapping is by *type*, "not every error mentioning an import".
3. Read/syntax error → command error as before.

Easy mistakes: a missing import inside a *dependency* file (not the root file) classified as semantic; an import blocked by `importFileAllowed` (surfaces as not-exist — correct routing or a COMPILE?); `wellknownimports` fallback forgotten in the linker; syntax errors inside `Link` reported as COMPILE instead of command errors. Note the cliff's blast radius: semantic errors go-protoparser *never detected* (it is not a linker — unknown types, duplicate symbols linted fine in main) now suppress all lint findings for the corpus.

**How to avoid:**
- Build a fixture matrix for the router: missing root import, missing transitive import, unknown type, duplicate symbol, syntax error, import blocked by module selection, WKT import — assert (issues, exit code, output bytes) for each against the REQ tables.
- Assert exit codes through the CLI `Action` path (`.spec/ERRORS.md` warns internal-function tests cannot establish exit codes).
- Keep `COMPILE` inside the existing issue format `path:line:col:sourceName message (COMPILE)` (REQ-5.1) — decide and test its `SourceName` and message shape before acceptance.

**Warning signs:**
- Corpus output collapsing to a few COMPILE lines where main produced hundreds of findings.
- Exit code 1 where main exited 2 (or lint succeeding where main hard-failed).

**Phase to address:**
Phase 2 (link.go routing) with the fixture matrix; re-verified in acceptance on real corpora (the acceptance corpora all compile — the broken paths are only covered by fixtures).

---

### Pitfall 7: Parity verification gaps — lab copy is older than main, and corpus-only diff cannot localize regressions

**What goes wrong:**
Rules copied from the lab silently drop main-only behavior added after the lab snapshot; a final corpus diff says "43 files differ somewhere" with no per-rule bisection, blocking release and forcing archaeology.

**Why it happens:**
- Verified in explore.md/REQ-3.2: lab rules were rewritten from an older main base; main evolved since (comment suppressions/`allowCommentIgnores` machinery exists in main's `lint.go`/`comment_suppressions.go` today). Copying 1:1 imports stale semantics.
- REQ-3.1's corpus parity is set-equality over 7414+92 files — a diff localizes to files/lines but not to *which rule rewrite* introduced a change when batches of rules land together.
- Harness fragility: corpora are rebuilt via `bench/setup.sh`, 2.9 GB of parity outputs are not in the repo, and access depends on the private compiler-lab (recorded assumption).

**How to avoid:**
- Resolve the open design question in favor of **batch-level parity tests**: for each rule batch, run old rule and new rule over shared fixtures (the colocated `<file>_test.go` corpora plus `internal/rules/testdata/`) and assert identical (rule, file, line, column) sets, before the corpus run.
- Reconcile each ported rule against *current main semantics*, treating lab code as structure and main as truth (REQ-3.2); diff the lab rule against the main rule's git history since the lab snapshot date.
- Reproduce the `lab.py` harness early (it is a flagged spike), commit its invocation docs, and record the compiler-lab commit used.

**Warning signs:**
- A ported rule's test file references fixtures/behavior that don't exist in main.
- Parity failures appearing only at end-of-milestone corpus runs.

**Phase to address:**
Roadmap planning (make batch parity the per-phase verification of each rule-batch phase); every rule-batch phase.

---

### Pitfall 8: Config schema enum drift and the `PACKAGE_NO_IMPORT_CYCLE` special case

**What goes wrong:**
New rules are selectable in config but rejected elsewhere (or accepted where they should be rejected), schemas drift, and MCP consumers get stale validation.

**Why it happens:**
One rule registry feeds four consumers, and main currently carries a trap: `rules.ValidateNames` (verified in `internal/rules/selection.go`) **hard-rejects `PACKAGE_NO_IMPORT_CYCLE` as "not implemented"** — a user-visible error that must flip to valid when the rule lands. The same registry feeds:
- `AllLintUseValues()` → JSON Schema `enum` for `linters.enable/disable` and `issues.exclude-rules.linters` (verified in `internal/config/v1/schema.go:73-75`),
- group membership in `New()` (REQ-4.1: `PACKAGE_NO_IMPORT_CYCLE` in MINIMAL; REQ-4.2: `SYNTAX_SPECIFIED`, `FIELD_NOT_REQUIRED` in BASIC),
- `knownLintRules` for comment-directive validation (Pitfall 2),
- MCP tool field descriptions in `mcp/easypconfig`.

Missing one consumer yields: `task schema:check` drift failure at the end; or worse, config that validates in the schema but errors in the CLI (or vice versa); or `easyp:disable PACKAGE_NO_IMPORT_CYCLE` comments erroring as "unknown rule".

**How to avoid:**
- Land each new rule with its registry edits, `task schema:generate`, and MCP description updates in the same commit; `task schema:check` in CI catches drift (never hand-edit `schemas/`).
- Add a test asserting `ValidateNames` accepts exactly `AllLintUseValues()` minus nothing — killing the special case by construction, not by memory.
- Grep for the rejected-rule string when the rule lands; the special case has no compiler error guarding its removal.

**Warning signs:**
- `task schema:check` failing after rule addition; MCP validation rejecting configs the CLI accepts.

**Phase to address:**
Phase 4 (new rules + REQ-4.4), enforced per-rule-commit rather than at milestone end.

---

### Pitfall 9: Vendored-compiler drift and repo hygiene

**What goes wrong:**
The vendored `internal/easypc` (3.4 MB including `cmd/` and `testdata/`) becomes an unmaintainable fork: local bug fixes land inside the vendored tree, the next sync from compiler-lab overwrites them, and nobody knows which upstream state the copy matches. Alternatively the vendor commit drags new module requirements into `go.mod`/`go.sum`, violating REQ-1.2.

**Why it happens:**
A copied tree has no module-system link to its origin — upstream and fork diverge silently (verified general practice; easyp-specific facts from explore.md). The lab itself rewrote imports to `github.com/easyp-tech/easyp/internal/...`, so the copy is already a transformed snapshot, making textual diff against upstream harder.

**How to avoid:**
- Record provenance in the vendoring commit message and a short `internal/easypc/VENDORED.md` (upstream repo, branch `easyp-cli`, commit SHA, date, import-rewrite performed).
- Keep the vendored tree pristine: put integration glue and fixes in `internal/protoast` / engine code outside `internal/easypc` wherever possible; anything that must change inside gets a marker comment referenced from VENDORED.md.
- Gate REQ-1.2 mechanically: `go mod tidy` must produce an empty diff for `go.mod`/`go.sum` in the vendoring PR; CI lint config must explicitly include or exclude the vendored tree (testdata `.proto` files can trip linters).
- Review-burden mitigation: vendor as a standalone commit with no other changes, mirroring how the lab structured it.

**Warning signs:**
- `go.mod`/`go.sum` diffs in the vendoring PR; lint failures inside `internal/easypc`; the same helper existing in both `internal/rules` and the vendored tree with diverging behavior.

**Phase to address:**
Phase 1 (vendoring) — provenance and hygiene rules are cheapest to establish in the commit that creates the tree; every future compiler sync re-applies them.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Copy lab rules 1:1 without reconciling against current main | Fast batch progress | Silent loss of main-only semantics (REQ-3.2 violations surface at corpus acceptance) | Never — reconciliation is the deliverable |
| Keep suppression filtering inside individual rules instead of the engine | Unblocks a rule batch | Double suppression or none; divergent behavior per rule | Never — engine owns suppression (main's `AppendIssue` already ignores its comments param) |
| Temporary bridge exposing protoast through the old `ProtoInfo` shape | Smaller diff per phase; old rules compile unchanged | Long-lived dual model, three representations of one file, unremovable bridge | Only if a phase is explicitly scoped to remove it before milestone end (open design question — decide in `/gsd-discuss-phase`) |
| Fix parity mismatches by editing the harness (normalizing diffs) | Green acceptance run | Real regressions ship; harness stops being evidence | Never for line/col; sorting both sides is legitimate set-comparison, document it |
| Leave `go-protoparser` positions shimmed via conversion in each rule | Avoids protoast position rework | 43 conversion sites, off-by-one class forever | Never — fix `core.Position = protoast.Position` once (lab already did) |
| Local edits inside `internal/easypc` to unblock lint | Quick fix | Clobbered on next compiler sync; sync becomes manual merge | Only with a VENDORED.md marker + upstream ticket |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| compiler-lab (`bench/lab.py`, corpora) | Assuming the harness exists and works; running it for the first time at acceptance | Reproduce the harness early (it is a flagged spike); record lab commit; keep corpora rebuildable via `bench/setup.sh` |
| `easyp breaking` (shared `protoInfoRead`) | Changing the shared read path signature/semantics for lint | Fork the path; breaking keeps go-protoparser untouched (REQ-5.2); run breaking tests in the same CI as the engine switch |
| Module selection boundary (`importFileAllowed`, `importRoots`) | Linker resolves imports by raw filesystem walk | Route the linker's resolver through the same boundary chain; a corpus compiling "too well" is a boundary breach, not a win |
| JSON Schemas + `mcp/easypconfig` | Forgetting that one registry feeds schemas, CLI validation, completion and MCP | Land registry + `task schema:generate` + MCP descriptions together; `task schema:check` gates drift |
| WKT imports (`wellknownimports`) | New linker missing the embedded-WKT fallback → spurious COMPILE on `google/protobuf/*.proto` imports | Wire `wellknownimports.Content.Open` into the linker resolver; fixture-test one WKT import |
| CI exit-code consumers | Treating exit code as incidental | Assert 0/1/2 through the CLI `Action` path for the router matrix (Pitfall 6); `.spec/ERRORS.md` documents the contract |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|-----------------|
| Judging speed on api-common (92 files) | "Fast enough" verdict that says nothing | Gate REQ-6.1 only on googleapis-nocompute (7414 files), same machine, multiple runs | Immediately at 7k+ file corpora |
| Per-file import re-parsing without cache in the new engine (main's `readFilesFromImport` re-parses every import per file) | Lint slower than main despite faster parser | Keep the lab's import cache + linked-tree reuse; re-link once per corpus, not per file | Corpora with heavy shared imports (googleapis) |
| Unbounded parallel fan-out (windows of 256 × per-rule × per-file ASTs resident) | Memory blowup, CI OOM kills | Keep the windowed design; measure peak RSS on the 7414-file corpus, not just latency | CI runners with tight memory limits |
| Building protoast model for files that rules never visit (e.g. imports only needed for cross-file rules) | Wasted CPU on huge corpora | Lazy/partial model construction per rule need (lab's two-pass design intent) | Corpores where imports >> scanned files |
| `-race` in CI masking parallel speed, or benching without it | Optimistic numbers that don't hold in CI | Bench with the CI-equivalent flags; compare against main built the same way | Acceptance demo vs CI discrepancy |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Linker bypassing `importFileAllowed` | Lint reads files outside the module's locked selection — data exposure and parity divergence vs breaking/generate which respect the boundary | Route all linker source access through the existing boundary functions; fixture-test a blocked import path |
| Compiler panic on malformed input instead of typed error | Untrusted corpora (user-supplied proto files) crash the CLI with a Go panic; exit code becomes 2-by-crash, not the documented contract; CI shows crashes not issues | Fuzz the vendored parser entry points (`go test -fuzz`) on malformed inputs early in Phase 2; recover panics at the file boundary only as a last-resort backstop, with an issue, not a crash |
| Vendored `cmd/easyp-protoc` binary building by default | Extra attack/maintenance surface shipped in every build | Keep it compiling (REQ-1.1 includes `cmd/`) but ensure it is not installed/released by the main build path |
| Trusting compiler-lab checkout scripts (`bench/setup.sh`) | Setup scripts run in the dev environment with repo context | Run corpus rebuild in an isolated directory; treat fetched corpora as untrusted input data |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| COMPILE cliff discovered by users (one broken file silences all lint findings) | CI that gated on specific lint findings passes/vacates unexpectedly; broken file hides other issues | Accepted by decision (buf-compatible), but document it prominently (README/docs, V1 release notes) and make the COMPILE message name the file and error clearly |
| Changing issue text format "while we're in there" (`path:line:col:source message (RULE)`) | Downstream parsers of lint output break | REQ-5.1 is absolute: format is frozen; add a golden-bytes test over the printer for a fixed fixture |
| Renaming a rule's Go struct during the port | Rule ID changes silently in output, suppressions and config (`GetRuleName` derives it from the type name by reflection — verified in `dom.go`) | Keep all 43 struct names byte-identical; add a test snapshotting the rule-name map before/after |
| New rules firing immediately on existing user corpora after upgrade (e.g. `FIELD_NOT_REQUIRED` on thousands of fields) | Upgrade turns green CIs red | Rules land in MINIMAL/BASIC per REQ-4.1/4.2 as decided — but communicate in release notes; verify group membership against the lab CLI, not memory |

## "Looks Done But Isn't" Checklist

- [ ] **Vendoring:** compiles and `go build` passes — but `go mod tidy` diff is empty (REQ-1.2) and provenance (upstream commit) is recorded
- [ ] **Engine:** corpus lint output matches on api-common — but googleapis-nocompute also matches, with sorted-set comparison, run twice (flakiness check)
- [ ] **Positions:** every rule verified on the ASCII fixture set — but the multi-byte UTF-8 fixture also produces identical line:col (rune columns)
- [ ] **Suppressions:** `easyp:disable` works — but EOF comments, pre-closing-brace comments, inline comments, `buf:lint:ignore` legacy form, and the enable-without-disable error all behave as main (mirror `comment_suppressions*_test.go`)
- [ ] **Error routing:** COMPILE issues appear — but missing-import still exits 2 "Cannot import file" via the typed error, syntax errors still hard-error, and a missing *transitive* import is classified correctly
- [ ] **Rule parity:** rule unit tests pass — but batch parity vs the *old* rule implementation on shared fixtures passed first (unit tests were rewritten, they can't prove parity by themselves)
- [ ] **New rules:** lint fires them — but `ValidateNames` no longer rejects `PACKAGE_NO_IMPORT_CYCLE`, schemas regenerated (`task schema:check` green), MCP descriptions updated, and `knownLintRules` accepts them in comment directives
- [ ] **Untouched subsystems:** lint is green — but breaking/generate/`mod tidy`/`migrate` suites ran in the same CI configuration as before
- [ ] **Parallelism:** corpus is fast — but `go test -race` over the parallel path with the stateful rules is clean, and output order is stable across runs
- [ ] **Performance:** "feels faster" — but REQ-6.1 measured against main on the 7414-file corpus with identical build flags

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Position off-by-one shipped into rule batches | MEDIUM | Fix once in protoast position adaptation; re-run full corpus diff; batches don't need re-porting, only re-verification |
| Suppression divergence discovered at acceptance | MEDIUM | Isolate attachment mismatches with per-fixture diff; fix protoast `comments.go` attachment; re-run suppression fixture suite |
| Exit-code regression in the wild | LOW | Router is centralized; add the missing typed-error branch + Action-level exit test; patch release |
| Race in a stateful rule post-merge | LOW | Make the rule stateless or scope state per file; covered by `-race` CI thereafter |
| Vendored tree diverged beyond manual merge | HIGH | Re-vendor from compiler-lab head; replay glue changes kept outside the tree (this is why fixes belong outside `internal/easypc`) |
| Schema/MCP drift after rule addition | LOW | Regenerate + update descriptions; `task schema:check` enforces; single-commit fix |
| Parity failure at end-of-milestone with no bisection | HIGH | Re-run per-batch parity tests by stashing rule batches (or git-bisect the branch) — expensive; prevented by batch-level parity from the start |
| Breaking regression from shared-path edit | LOW | Revert to forked read path; breaking tests localize it quickly if REQ-5.2 CI gate is in place |

## Pitfall-to-Phase Mapping

Assumed phase skeleton follows the requirements' topological order (REQ-1 → REQ-2 → REQ-3 → REQ-4 → REQ-5/6); rename to actual roadmap phase titles as created.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| P1 Position fidelity | Phase 1/2 (vendoring + engine) | UTF-8 position fixture; corpus lint-diff on both corpora |
| P2 Suppression parity | Phase 2 (engine) + every rule batch | Mirrored `comment_suppressions*_test.go` fixtures pass on protoast model |
| P3 Determinism/ordering | Phase 2 (design) + acceptance | Two consecutive runs byte-stable (new engine); harness compares sorted sets; main's own flakiness baseline measured |
| P4 Stateful rules under parallelism | Phase 2 + each rule batch | `-race` engine test; per-rule receiver-write audit checklist in batch PRs |
| P5 Dual-model coexistence | Phase 2 (fork) + Phase 5 (REQ-5.2) | breaking/generate/tidy/migrate suites green in same CI; import-boundary fixture |
| P6 COMPILE cliff / error routing | Phase 2 (router) + acceptance | Fixture matrix with (issues, exit code, bytes) asserted via CLI Action |
| P7 Parity verification gaps | Roadmap design (batch parity) + each rule batch | Per-batch old-vs-new rule parity tests; corpus run only as final gate |
| P8 Schema enum drift | Phase 4 (new rules) | `task schema:check` green; registry↔schema↔MCP consistency test; special-case removal verified |
| P9 Vendored-tree drift | Phase 1 (vendoring) | Provenance recorded; `go mod tidy` empty diff; no edits inside `internal/easypc` without VENDORED.md marker |

## Sources

**Repo-verified (HIGH confidence, read in source at `5c9185f` / module cache):**
- `internal/core/lint.go` — engine loop, suppression filter, no output sorting
- `internal/core/comment_suppressions.go`, `internal/core/proto_comments.go` — reflection-based comment attachment, directive validation, EOF/pre-brace comments
- `internal/core/dom.go`, `internal/core/proto_info_read.go`, `internal/core/core.go` — `Issue`/`ProtoInfo` model, importRoots/importFileAllowed/wellknownimports chain, `GetRuleName` reflection
- `internal/rules/import_used.go` — stateful receiver, map-iteration issue emission
- `internal/rules/selection.go`, `internal/rules/builder.go` — `ValidateNames` `PACKAGE_NO_IMPORT_CYCLE` special case, group registry
- `internal/config/v1/schema.go` — `AllLintUseValues()` → schema enums
- `internal/api/lint.go` — output format string, exit-code branches
- `internal/parallel/parallel.go` — index-ordered parallel results
- `internal/core/breaking_check.go` — shared `protoInfoRead` consumer
- `.spec/features/lint-easypc-infra/explore.md`, `requirements.md` — lab divergence, REQ tables, open design questions
- `.planning/codebase/CONCERNS.md`, `.planning/codebase/TESTING.md` — exit-code test gap, fragile areas, test conventions
- go-protoparser v4.14.2 module cache: `parser/meta/position.go` (1-based line, 1-based rune column, 0-based byte offset), `parser/import.go` (`Meta.Pos` = statement first token), `lexer/scanner/position.go`
- `.spec/ERRORS.md` — exit-code contract, typed-error mappings

**External (confidence per item):**
- protobuf.dev descriptor API reference — SourceCodeInfo spans are 0-based line/column (HIGH, official docs; cross-checked on two protobuf.dev reference pages)
- buf lint docs (buf.build/docs/lint/usage) — comment-ignore placement and v1/v2 defaults (MEDIUM-HIGH, official)
- buf compilation-and-descriptors reference — leading/trailing/detached comment categorization algorithm (MEDIUM-HIGH, official)
- bufbuild/protocompile README — thread-safe `SourceResolver.Accessor`, reporter accumulation semantics (MEDIUM, official repo)
- Go spec / Stack Overflow — map iteration order intentionally randomized (MEDIUM, spec quoted)
- Golden-file and vendored-fork maintenance practice articles (MEDIUM-LOW, general web; used only for framing, not for easyp-specific claims)

---
*Pitfalls research for: lint parser/engine migration (go-protoparser → easypc/protoast)*
*Researched: 2026-11-17*
