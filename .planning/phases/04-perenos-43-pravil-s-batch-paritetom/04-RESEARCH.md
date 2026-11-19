# Phase 4: Перенос 43 правил с батч-паритетом - Research

**Researched:** 2026-11-19
**Domain:** Porting EasyP's 41 registry lint rules from the go-protoparser `core.Rule` contract to the Phase-3 linked-engine `core.LintRule`/`core.ModuleRule` contract (protoast model), with old-vs-new per-batch parity
**Confidence:** HIGH (every rule file diffed main-vs-lab this session; all in-repo claims read at source)

## Summary

Phase 4 re-implements the current rule set on the Phase-3 linked engine contract: `core.LintRule` (`Validate(LinkedFileInfo)`) for per-file rules, `core.ModuleRule` (`Collect`/`Finish`) for the 10 cross-file rules. The porting source is the compiler-lab etalon (`/tmp/gsd-research/compiler-lab/easyp-cli/internal/rules` @ `08cef10`), where all rules were already rewritten onto protoast. A full file-by-file diff of all 41 main rules against their lab counterparts (run this session) shows the diffs are **not uniformly mechanical**: besides the expected model swap (`protoInfo.Info.ProtoBody.*` → `protoInfo.File.Walk(protoast.Visitor{...})`), the lab changed **observable behavior** in three systematic ways — (1) most casing rules replaced main's regexps with the shared `casing.go` helpers, which disagree with the regexps on acronyms, digits, and underscores; (2) rules reporting positions moved from the declaration **keyword** to the declaration **name token** (`NamePos`, `NumberPos`, `Request.Pos`); (3) group-checking rules switched from main's first-file-wins / skip-empty / per-file semantics to report-everyone / count-empty / cross-file semantics. Per LINT-PARITY-02 (main wins) and the Phase-7 corpus parity gate (LINT-PARITY-01: identical (rule, file, line, col) old-vs-new), **main's semantics win everywhere they differ**: the port is the lab's protoast walk mechanics + main's predicate, position, and grouping behavior, proven by main's own colocated test expectations.

The central structural problem is **two-contract coexistence**: the old engine (`c.rules` + `Lint`) stays the live user path until Phase 5, so for most of the phase the same conceptual rule must exist twice in `internal/rules` — as the old `core.Rule` (untouched, registry names pinned by reflection over type names) and as the new `core.LintRule`. Old type names are immovable (GetRuleName is reflection-based, names are schema/config-pinned), so the new implementations need either a subpackage or suffixed type names plus an explicit-name override for `LinkedRuleName` (recommendation below). The old rule files are never edited.

**Primary recommendation:** Port in config-group batches (MINIMAL → BASIC×2 → DEFAULT → COMMENTS → UNARY_RPC) behind a Wave-0 parity harness that drives the **old engine and the new engine in-process on the same compiling fixture trees** and compares sorted (rule, path, line, col) multisets; corpus lint-diff (`lab.py`) stays the Phase-7 FINISH gate, not the batch mechanism. Each batch carries main's colocated test expectations onto the new implementations and pins the catalogued main-vs-lab divergences (this document's inventory) as explicit test cells.

## User Constraints

No `04-CONTEXT.md` exists at research time (phase dir empty; design notes are scheduled for `/gsd-discuss-phase` per ROADMAP). The binding constraints are:

### From ROADMAP § Phase 4 (goal + success criteria)
- **Goal:** «Все 43 текущих правила переписаны на protoast с семантикой текущего main (main побеждает лабораторную копию); каждый батч переноса верифицирован по паритету до начала следующего» — main semantics beat the lab copy; batch parity verified before the next batch starts; corpus lint-diff stays the FINISH gate, not a debugging session.
- **Criterion 1:** all rules produce findings identical to the old implementation (rule, file, line, col) on shared fixtures; batches verified in order.
- **Criterion 2:** rules whose behavior changed in main after the lab snapshot follow main; **colocated tests are carried over from main**.
- **Criterion 3:** every ported rule has a colocated `<rule>_test.go`; `task test` and `task lint` green at all intermediate states.
- **Design notes to resolve in discuss:** batch grouping/order (natural candidates — config groups MINIMAL → BASIC → DEFAULT → COMMENTS → UNARY_RPC); batch-verification mechanism — per-batch parity tests vs corpus lint-diff (resolved in this document: per-batch parity tests, see below).

### Standing constraints (UAT-ratified Phase 3; ROADMAP backlog 999.1/999.2; STATE.md Blockers)
- **(a) Disk instrumentation:** add a disk-instrumentation cell asserting `readSource` runs exactly once per corpus file on a compiling corpus (behavioral proof of the no-disk-reread guarantee).
- **(b) Suppression centrality at EVERY rule-port batch:** ported rules never consult comment directives or `ignoreOnly` themselves; central-filter cells (`comment_suppression_central`, `finish_comment_refilter`, `compile_bypasses_suppression`) stay green.

### From AGENTS.md / project skills (unchanged)
- Sealed `internal/easypc` — no edits; `internal/protoast` is the maintained adaptation seam.
- Colocated `<rule>.go` + `<rule>_test.go`; testify; `-race` mandatory; handwritten doubles fine.
- No schema changes in this phase (rule names unchanged until Phase 6) → `task schema:check` stays clean.
- Phase-1 lint exclusion for `internal/rules/lint_helpers.go` must be **removed when helpers gain callers** (STATE.md Phase-4 note); re-check QF1001 in `casing.go` at next lab re-sync.

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| LINT-PARITY-02 (REQ-3.2) | Where current-main rule behavior changed after the lab snapshot (e.g. comment suppressions), keep current-main semantics, not the lab copy | Full 41-rule main-vs-lab diff inventory below catalogues every divergence and assigns the main-wins port decision per rule; per-batch parity tests against the live old implementation enforce it mechanically |

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Rule predicates (per-file) | `internal/rules` (new `LintRule` impls) | `internal/protoast` (Visitor walk, Decl positions) | Rules own the "what is a violation" logic; the model walk is protoast's |
| Cross-file grouping (Collect/Finish) | `internal/rules` (10 ModuleRules) | `internal/core` (engine lifecycle, serial drain) | Engine guarantees phase boundary and walk order; rules only accumulate/report |
| Issue construction for the new contract | `internal/core` (new linked AppendIssue/buildError variant) | — | Lab's 4-arg `AppendIssue(LintRule, Position)` has no main equivalent; must be added under a rename-map name |
| Registry / wiring of both contracts | `internal/rules` builder | `internal/core` Options (`Rules` + additive `LinkedRules`) | Old `rules.New` untouched; new constructor mirrors it |
| Suppressions, ignore_only, ordering | `internal/core` engine (Phase 3, done) | — | Rules never self-suppress (standing constraint b); D-01 sort at engine exit |
| CLI wiring / flip | — | — | Out of scope; dormancy until Phase 5 |

## Standard Stack

All stack items are **in-repo** — zero external dependencies, nothing to install.

| Component | Version/Location | Purpose | Why Standard |
|-----------|------------------|---------|--------------|
| Go toolchain | go.mod `go 1.26.6` | Language | Pinned [VERIFIED: go.mod] |
| `core.LintRule`/`ModuleRule`/`LinkedFileInfo`/`Resolution` | `internal/core/lint_model.go:35-91` | The new rule contract (Phase 3) | Locked engine contract; engine drives `c.linkedRules` only [VERIFIED: internal/core/lint_linked.go:173-212] |
| `core.LintLinked` | `internal/core/lint_linked.go:25-161` | Dormant engine the ported rules plug into | Two-pass, D-01 sort, `shouldIgnoreLinked`, central suppressions — all done |
| `protoast.File.Walk(Visitor)` | `internal/protoast/walk.go:18` | Per-file traversal (the lab rules' mechanic) | Replaces main's hand-written nested loops; recursion coverage pinned in pilot [ASSUMED: full nested-kind coverage — pin via pilot fixture] |
| `protoast.Decl.Pos` / `Decl.NamePos` | `internal/protoast/protoast.go:63-71` | Position mapping: `Pos` = "first token", `NamePos` = "the declared name" | See Position Mapping pitfall — main's keyword positions map to `Decl.Pos` (test-verified for Import, Field, Message, Package, RPC, Option kinds) |
| `rules.New` / builder groups | `internal/rules/builder.go:61-143` | Old registry — source of names, groups, config plumbing | Untouched; new `NewLinked` mirrors it |
| `casing.go` helpers | `internal/rules/casing.go` (byte-identical main/lab) | Lab rules' casing oracle | Main rules use regexps instead — helpers stay for lab-faithful rules and Phase 6 |
| `lint_helpers.go` | `internal/rules/lint_helpers.go` | Vendored lab helpers, 3 catalogued seam adaptations | Header mandates Phase-4 re-alignment (`mismatches` takes old `core.Rule`; must take the new contract) [VERIFIED: lint_helpers.go:10-27] |
| testify | go.mod | Tests | Convention |

**Version verification:** n/a — no registry packages involved.

## Package Legitimacy Audit

Not applicable: zero external packages installed (pure in-repo Go; REQ-1.2 prohibits new external deps on the compiler path). No `checkpoint:human-verify` items.

## The 41-Rule Inventory: Port Classes and Main-vs-Lab Divergences

**Registry truth:** the builder constructs **41** rules — MINIMAL 4 + BASIC 20 + DEFAULT 8 + COMMENTS 7 + UNARY_RPC 2 [VERIFIED: internal/rules/builder.go:65-120]. The "43" in REQUIREMENTS.md/ROADMAP/phase title is the spec-document artifact flagged as Phase-3 Pitfall 7; no rules exist outside the builder. The lab tree additionally contains `field_not_required.go`, `package_no_import_cycle.go`, `syntax_specified.go` — these are the **Phase 6 new rules**, not Phase 4 scope.

**Port classes** (every rule also gets mechanical renames: `core.ProtoInfo`→`core.LinkedFileInfo`, `var _ core.Rule`→`var _ core.LintRule` (or `ModuleRule`), lab 4-arg `AppendIssue`→the new linked variant, position field per main):
- **verbatim-from-lab** — lab logic ≡ main logic; port = copy + mechanical renames + position-field swap back to main's.
- **adapt** — lab's walk shape kept; hand-adjustment beyond renames (ModuleRule lifecycle, local-state extraction, helper re-alignment); main semantics preserved.
- **main-diverged** — lab changed observable behavior; port = lab's protoast walk + **main's predicate/position/grouping**, proven by main's carried tests.

| # | Rule | Group | File (rule name) | Class | Main-wins decisions the port must encode | Position note |
|---|------|-------|------------------|-------|------------------------------------------|---------------|
| 1 | DirectorySamePackage | MINIMAL | directory_same_package.go | main-diverged + adapt | Main: cache dir→package, **first file of a dir exempt**, only later differing files issue. Lab `groupedValues.mismatches` reports everyone — do NOT reuse as-is for this rule | package pos: main `pack.Meta.Pos`; lab `packagePosition` → `pkg.Pos` — same kind mapping, pin |
| 2 | PackageDefined | MINIMAL | package_defined.go | main-diverged (position) | Main emits `meta.Position{Filename: path}` → **Line 0, Column 0** [VERIFIED: main package_defined_test.go expects 0:0]; lab emits `protoast.FilePosition` → 1:1. Port emits the zero Position literal | 0:0 has no protoast helper — construct `core.Position{Filename: path}` directly |
| 3 | PackageDirectoryMatch | MINIMAL | package_directory_match.go | **verbatim-from-lab** | 4-line diff, model access only; Root/Prefix logic identical | pin `pkgInfo.Pos` ≡ main `Meta.Pos` |
| 4 | PackageSameDirectory | MINIMAL | package_same_directory.go | main-diverged + adapt | Main: keyed by package, first directory exempt, later files issue at package pos. Lab reports everyone | package pos as #1 |
| 5 | EnumFirstValueZero | BASIC | enum_first_value_zero.go | main-diverged (+1 deliberate deviation) | Main: **string compare** `Number != "0"` (only literal "0" passes; "00"/"0x0" issue). Lab: numeric `enumNumber != 0`. Main wins. Main indexes `EnumFields[0]` unguarded → **panics on empty enum**; keep lab's `len(Values)==0` guard as a documented crash fix (unreachable on compiling corpora: protoc rejects empty enums → COMPILE path) | main reports at value `Meta.Pos` (name); lab at `NumberPos` (the number token) — use main's |
| 6 | EnumNoAllowAlias | BASIC | enum_no_allow_alias.go | main-diverged | Main: **any** option named `allow_alias` (any value, incl. `= false`) at **enum** position. Lab: only `Value=="true"` at **option** position. Main wins | enum `Meta.Pos` ≡ `enum.Pos` (Decl.Pos = first token) |
| 7 | EnumPascalCase | BASIC | enum_pascal_case.go | main-diverged | Main regexp `^[A-Z][a-z]+(?:[A-Z][a-z]+)*$` — acronym enums ("ABC") **fail** main. Lab `toPascalCase` equality passes them. Keep regexp | main `Meta.Pos` (keyword); lab `NamePos` — use main's |
| 8 | EnumValueUpperSnakeCase | BASIC | enum_value_upper_snake_case.go | main-diverged | Main regexp `^[A-Z0-9]+(_[A-Z0-9]+)*$` — leading/trailing underscore and empty **fail** main; lab `toUpperSnakeCase` equality passes them. Keep regexp | enum-value Meta.Pos (name start) ≈ `value.Pos`/`NamePos` — pin |
| 9 | FieldLowerSnakeCase | BASIC | message_field_lower_snake_case.go | main-diverged | Main regexp `^[a-z0-9]+(_[a-z0-9]+)*$` — leading `_` fails main; lab helper passes it. Keep regexp | main field `Meta.Pos` = type-token start (col 3) [VERIFIED: main test col 3]; lab `NamePos` col 10 — use main's ≡ `field.Pos` |
| 10 | ImportNoPublic | BASIC | import_no_public.go | **verbatim-from-lab** | `imp.Modifier == ImportModifierPublic` → `imp.Public`; logic identical | test-verified same position (3-line test diff = alias only) |
| 11 | ImportNoWeak | BASIC | import_no_weak.go | **verbatim-from-lab** | `imp.Modifier == ImportModifierWeak` → `imp.Weak` | same as #10 |
| 12 | ImportUsed | BASIC | import_used.go | adapt + parity-risk | Lab extracts receiver state into local `importUsage` (stateless rule — required for parallel windows). Lab **fallback** walk (no `Resolved`) mirrors main's `existInProto`/`InstructionParser` mechanics. Lab adds **`validateResolved`** (resolution-based) used whenever `Resolved != nil` — i.e. on every compiling corpus — NEW behavior main never had. **Recommendation:** port main's syntax-walk semantics unconditionally for Phase 4 parity; treat `validateResolved` as a lab improvement gated behind a deliberate post-parity decision (same bucket as #28 cross-file) | import pos same (test-verified) |
| 13 | MessagePascalCase | BASIC | message_pascal_case.go | main-diverged | Main regexp `^[A-Z][a-zA-Z0-9]+(?:[A-Z][a-zA-Z0-9]+)*$` — single-char message ("A") **fails** main; lab passes it. Keep regexp | main `Meta.Pos` keyword col 1 vs lab `NamePos` col 9 [VERIFIED: test diff 536/1 → 544/9] — use main's ≡ `message.Pos` |
| 14 | OneofLowerSnakeCase | BASIC | oneof_lower_snake_case.go | main-diverged | Main regexp `^[a-z]+(_[a-z]+)*$` — **digits fail** main ("oneof1"); lab helper passes. Keep regexp | main `Meta.Pos`; lab `NamePos` — use main's |
| 15 | PackageLowerSnakeCase | BASIC | package_lower_snake_case.go | main-diverged | Main regexp `^[a-z]+([_|[.][a-z0-9]+)*$` (char class `{_,\|,[, .}`; digits fail in first component). Lab: per-component `toLowerSnakeCase`. Keep regexp | package pos same (3-line test diff) |
| 16 | PackageSameCsharpNamespace | BASIC | package_same_csharp_namespace.go | main-diverged + adapt | Main: files **without the option are skipped entirely** (not counted as a value, never reported); first option value per package wins; later differing files issue at option pos. Lab: optionless file counts as its own (empty) value at FilePosition and everyone is reported. Main wins | main option `Meta.Pos` ≡ `Option.Pos` (Decl.Pos = first token = `option` keyword) |
| 17 | PackageSameGoPackage | BASIC | package_same_go_package.go | main-diverged + adapt | as #16 | as #16 |
| 18 | PackageSameJavaMultipleFiles | BASIC | package_same_java_multiple_files.go | main-diverged + adapt | as #16 | as #16 |
| 19 | PackageSameJavaPackage | BASIC | package_same_java_package.go | main-diverged + adapt | as #16 | as #16 |
| 20 | PackageSamePHPNamespace | BASIC | package_same_php_namespace.go | main-diverged + adapt | as #16 | as #16 |
| 21 | PackageSameRubyPackage | BASIC | package_same_ruby_package.go | main-diverged + adapt | as #16 | as #16 |
| 22 | PackageSameSwiftPrefix | BASIC | package_same_swift_prefix.go | main-diverged + adapt + **BUG PIN** | as #16 **plus**: main has a copy-paste bug — it compares **`ruby_package`** options, not `swift_prefix` [VERIFIED: main package_same_swift_prefix.go option-name literal]. Decision needed: bug-for-bug parity (recommended for Phase 4 — pin with a regression test + explicit note) vs fix (behavior change, parity risk) | as #16 |
| 23 | RPCPascalCase | BASIC | rpc_pascal_case.go | main-diverged | Main regexp `^[A-Z][a-zA-Z0-9]*$` (single char OK; leading digit fails). Lab `toPascalCase` equality differs on digit-led names ("2Fast": main issues, lab passes). Keep regexp | main `Meta.Pos` keyword col 3 vs lab `NamePos` col 7 [VERIFIED: test diff 300/3 → 304/7] — use main's ≡ `rpc.Pos` |
| 24 | ServicePascalCase | BASIC | service_pascal_case.go | main-diverged | Main regexp `^[A-Z][a-z]+([A-Z]|[a-z]+)*$` — "AB", "ABC", digits **fail** main. Keep regexp | main `Meta.Pos` col 1 vs lab `NamePos` col 9 — use main's |
| 25 | EnumValuePrefix | DEFAULT | enum_value_prefix.go | main-diverged | Main prefix = private `pascalToUpperSnake(name)` (underscore before **every** uppercase rune) and `HasPrefix(ident, prefix)` **without** trailing `_` (value == enum name passes). Lab: `toUpperSnakeCase(name)+"_"` (acronym-aware, requires `_`). Two divergences: acronym enums and name-only values. Keep main's algorithm — carry `pascalToUpperSnake` into the ported file | enum-value pos ≈ same (name start); pin |
| 26 | EnumZeroValueSuffix | DEFAULT | enum_zero_value_suffix.go | main-diverged | Main: checks **only the first** value and requires **exact equality** with `pascalToUpperSnake(enumName)+"_"+Suffix`. Lab: checks **every** value numbered 0 for suffix-only ending (underscore-normalized). Keep main's (first-value + full-string) | main first-value `Meta.Pos`; lab `NamePos` — use main's |
| 27 | FileLowerSnakeCase | DEFAULT | file_lower_snake_case.go | main-diverged (position + regexp) | Main: checks the **full filename including `.proto`** against `^[a-z]+([_|[.][a-z0-9]+)*$`; emits `meta.Position{Filename: path}` → **0:0**. Lab: strips extension, helper equality, FilePosition 1:1. Keep main's regexp and 0:0 | zero Position literal as #2 |
| 28 | RPCRequestResponseUnique | DEFAULT | rpc_request_response_unique.go | main-diverged + adapt | Main: **per-file** uniqueness (written type strings, `lo.Contains` per file), **first user exempt**, later users issue at rpc keyword pos; input==output reported as a duplicate within the file. Lab: **cross-file** ModuleRule over **qualified** names, **every** user reported. Main wins on all three. Keep the ModuleRule shape (D-03 list), but `Finish` groups by file and reproduces main's within-file first-exempt logic on **written** names (`Resolved.RPCTypes` not used for parity) | main `rpc.Meta.Pos` ≡ `rpc.Pos` (test-verified kind) |
| 29 | RPCRequestStandardName | DEFAULT | rpc_request_standard_name.go | main-diverged | Main: written `MessageType` must equal `RPCName+"Request"` or `ServiceName+RPCName+"Request"` **exactly** (a qualified written type `.pkg.GetFooRequest` FAILS main); lab: last dot-segment vs `toPascalCase(rpc.Name)+"Request"` (accepts qualified, pascal-normalizes). Keep main's written-string compare | main rpc `Meta.Pos` (keyword); lab `rpc.Request.Pos` (type token) — use main's |
| 30 | RPCResponseStandardName | DEFAULT | rpc_response_standard_name.go | main-diverged | as #29 with `Response` | as #29 |
| 31 | PackageVersionSuffix | DEFAULT | package_version_suffix.go | main-diverged | Main regexp `.*v\d+\|.*v\d+test.*\|.*v\d+(alpha\|beta)\d*\|.*v\d+p\d+(alpha\|beta)\d*$` (unanchored; single-component "v1" passes). Lab `packageHasVersionSuffix` (2+ components, major ≥ 1) — "v1" fails lab. Keep regexp | package pos same (3-line test diff) |
| 32 | ServiceSuffix | DEFAULT | service_suffix.go | **verbatim-from-lab** | `HasSuffix(name, Suffix)` identical | main `Meta.Pos` keyword col 1 vs lab `NamePos` col 9 [VERIFIED: test diff 197/1 → 205/9] — use main's ≡ `service.Pos` |
| 33 | CommentEnum | COMMENTS | comment_enum.go | main-diverged | Main: issue iff **no comment attached at all** (`len(Comments)==0`; a directive-only comment counts as documentation). Lab: `!leadingCommentDocuments` (directive-only ⇒ undocumented ⇒ issue). Main wins — port the presence predicate as "no Leading/Detached/Trailing-equivalent attached", NOT the lab's directive-filtering helper | main `Meta.Pos` ≡ lab `.Pos` (comment_* tests verified alias-only diffs) |
| 34 | CommentEnumValue | COMMENTS | comment_enum_value.go | main-diverged | as #33 | as #33 |
| 35 | CommentField | COMMENTS | comment_field_message.go | main-diverged (+group question) | as #33. Lab adds `!field.Group` ("group fields documented on the group itself"); main has no group guard — verify how main's go-protoparser model surfaces proto2 group fields before dropping the guard [ASSUMED: group fields exotic; encode main behavior literally and pin with a proto2 group fixture if reachable] | as #33 |
| 36 | CommentMessage | COMMENTS | comment_message.go | main-diverged | as #33 | as #33 (test-verified same position) |
| 37 | CommentOneof | COMMENTS | comment_one_of.go | main-diverged | as #33 | as #33 |
| 38 | CommentRPC | COMMENTS | comment_rpc.go | main-diverged | as #33 | as #33 |
| 39 | CommentService | COMMENTS | comment_service.go | main-diverged | as #33 | as #33 |
| 40 | RPCNoClientStreaming | UNARY_RPC | rpc_no_client_streaming.go | **verbatim-from-lab** | `Request.IsStream`→`Request.Stream` | test-verified same position (alias-only diff) — `rpc.Pos` ≡ `Meta.Pos` |
| 41 | RPCNoServerStreaming | UNARY_RPC | rpc_no_server_streaming.go | **verbatim-from-lab** | `Response.IsStream`→`Response.Stream` | as #40 |

**Tally:** verbatim-from-lab 6 (#3, 10, 11, 32, 40, 41 — all with position pins where applicable); adapt 3 (#12, plus the ModuleRule lifecycle aspect of #1/#4/#16-22/#28 folded into their main-diverged entries); main-diverged 32. Every rule: the `AppendIssue` call and `Validate` signature change mechanically.

### Position mapping — the systematic finding (read this before planning any rule task)

Main rules emit go-protoparser `Meta.Pos` = the **declaration keyword/first token**; the lab deliberately moved casing/suffix/standard-name rules to finer tokens (`NamePos` = name, `NumberPos` = number literal, `Request.Pos` = type name "after any stream keyword" [VERIFIED: protoast.go:188-191]). Evidence (byte-diffs of colocated tests, same fixtures): message 536/col 1 → 544/col 9; rpc 300/col 3 → 304/col 7; service 197/col 1 → 205/col 9; field 560/col 3 → 567/col 10 [VERIFIED: test diffs this session]. Meanwhile `protoast.Decl.Pos` is documented "first token" and is **test-verified equal to main's Meta.Pos** for Import, Field, Message, Package, RPC, Option, Enum kinds (the non-casing lab rules that use `.Pos` show alias-only test diffs). **Resolution:** ported rules emit `Decl.Pos` (not `NamePos`) wherever main used `Meta.Pos`, and a zero `core.Position{Filename: path}` for the two file-level rules (#2, #27). The carried main test expectations are the oracle — any wrong mapping fails the batch gate immediately. Wave 0 adds a one-shot position probe (parse one fixture per decl kind through both models) to lock the table above into a comment.

## Recommended Batches

Group-per-config-group (ROADMAP's natural candidate), with semantic-risk rules placed inside their groups and a Wave-0 infrastructure batch first. Batches are plans; order is fixed (each verified before the next starts):

| Batch | Rules | Content |
|-------|-------|---------|
| **0 — Wave 0 infra + pilot** | PackageDirectoryMatch (#3) | Linked issue-construction helper in core (`AppendLinkedIssue`/`buildError` for `LintRule` + `Position`, rename-map name); `init_test.go` helpers re-based on the lab's (`validate()` Collect/Finish driver, `parseProto`); position probe; parity harness skeleton; pilot rule proves the harness end-to-end |
| **1 — MINIMAL (4)** | #1, #2, #3*, #4 | First two ModuleRules with first-exempt Finish logic; the two zero-position file-level rules |
| **2 — BASIC casing (9)** | #7, #8, #9, #13, #14, #15, #23, #24 + EnumValueUpperSnakeCase already in #8… (exact split planner's discretion) | All regexp-wins rules — uniform shape: lab walk + main regexp; fixtures for acronym/digit/underscore edges |
| **3 — BASIC semantic (11)** | #5, #6, #12, #16–#22 | EnumFirstValueZero guard, EnumNoAllowAlias, ImportUsed (stateless + resolved-path decision), the 7 option-comparison ModuleRules incl. swift_prefix bug pin |
| **4 — DEFAULT (8)** | #25–#32 | The heaviest divergences: EnumValuePrefix algorithm, EnumZeroValueSuffix first-value semantics, RPCRequestResponseUnique per-file grouping, standard-name written-string compares |
| **5 — COMMENTS (7)** | #33–#39 | One uniform presence-predicate port; shared fixture set incl. directive-only-comment case |
| **6 — UNARY_RPC (2) + finish** | #40, #41 + wiring | Builder `NewLinked` wiring for all 41, helpers re-alignment (`mismatches` → new contract + main first-exempt variant), lint-exclusion removal, registry-count guard, full gates |

(Batch 0's pilot #3 is re-verified inside batch 1; batches 2/3 split BASIC to keep a single batch's parity surface reviewable.)

## Batch-Verification Mechanism (design note — RESOLVED)

**Recommendation: per-batch parity tests against the old implementation, in-process; corpus lint-diff stays the FINISH gate.** Rationale:

1. **The oracle is alive the whole phase.** The old engine (`c.rules` + `Lint`) remains the user path until Phase 5 — the parity harness can run both engines in one test process on the same fixture tree and compare findings. Phase 3 already established the both-engines-on-one-fixture pattern (parity double, `MessagePascalCase` oracle + mirror) [CITED: 03-RESEARCH.md Pattern/Validation sections].
2. **Corpus lint-diff is the wrong granularity for debugging.** `lab.py` requires the private compiler-lab checkout and corpus rebuilds (ROADMAP blocker notes a spike is needed even for Phase 7); a corpus diff failure localizes nothing. ROADMAP itself says the corpus gate must stay "воротами, а не сессией отладки".
3. **Preconditions:** parity fixtures must be **compiling** corpora (LINT-PARITY-01's precondition "все файлы компилируются") — on broken files the engines legitimately diverge (COMPILE routing vs per-file parse).
4. **Compare sorted multisets** of (RuleName, Path, Position.Line, Position.Column) — old engine output order is nondeterministic (map iteration, D-01 research); the new engine is D-01-sorted; sort both sides before comparing. Messages may be compared too but (rule, file, line, col) is the contract.
5. **Fixture coverage must encode the catalogued divergences** — acronym enums, digit names, underscores, single-char messages, optionless files, qualified written types, cross-file type sharing, directive-only comments, mixed packages/directories. A parity test that only replays happy-path fixtures would pass while corpora diverge. Each inventory row's "Main-wins decisions" column names the fixture cells its batch needs.
6. **Standing constraints ride every batch:** suppression-centrality cells green; `readSource`-once disk-instrumentation cell added in Wave 0 and kept green.

## Two-Contract Coexistence (the structural decision for the planner)

During Phase 4 the same rule name must exist as an old `core.Rule` (live, reflection-named, immovable) and a new `core.LintRule`. `LinkedRuleName` derives names by reflection over the type name [VERIFIED: lint_model.go:95-97], so a suffixed new type (`MessagePascalCaseLinked`) would yield `MESSAGE_PASCAL_CASE_LINKED` unless names are overridden.

| Option | Mechanics | Tradeoff |
|--------|-----------|----------|
| **A (recommended): in-package suffixed types + explicit-name override** | New types `<Name>Linked` in `internal/rules`; `LinkedRuleName` gains an additive preference for an optional `interface{ LintRuleName() string }` before falling back to reflection; each new type returns its canonical name. Phase 5 deletes old types, drops suffixes and the override | One package → shared `casing.go`/`lint_helpers.go` stay unexported and shared; old files untouched; engine change is additive (evolves Phase-3 code, documented in the rename map) |
| B: subpackage `internal/rules/linked` | Clean names from day one | Must export or duplicate the unexported helpers (`casing.go`, `groupedValues`, `packagePosition`…); Phase 5 must migrate everything back up; diverges from the etalon's single-package shape |
| C: shared predicates + thin dual wrappers | Extract each rule's logic into package funcs consumed by both contracts | 41 extraction refactors touch old-path-adjacent code — dormancy risk; rejected (anti-pattern: "mutating the old path while we're here") |

Supporting wiring (all additive): `rules.NewLinked(cfg) ([]core.LintRule, map[string][]string, error)` mirroring `New` (reuses `validateConfigRuleNames`, `unwrapLintGroups`, `unwrapIgnoreOnly`); `Options.LinkedRules` already exists [VERIFIED: core.go:68-76]; whether `buildCore` populates `LinkedRules` in Phase 4 (harmless — engine dormant) or waits for Phase 5 is planner discretion.

## Architecture Patterns

### System Architecture Diagram

```
                easyp lint (CLI) ── UNCHANGED ──> old engine c.Lint
                                                    │ c.rules = rules.New(cfg)   [41 old core.Rule]
                                                    └─ LIVE user path until Phase 5

 NEW (dormant, constructed in tests; buildCore wiring = planner discretion):

  rules.NewLinked(cfg) ──> Options.LinkedRules []core.LintRule   [41 new impls, both contracts coexist]
     │
     ▼
  c.LintLinked (Phase 3 engine — unchanged)
     walk → readSource (once per file — standing constraint a) → linkSources
     → windows: model build + checkFile
            │  per-file LintRule.Validate(LinkedFileInfo)   ← ported rules land here
            │  central comment suppressions + shouldIgnoreLinked   (rules never self-suppress)
     → serial drain: ModuleRule.Collect in walk order
     → ModuleRule.Finish + shouldIgnoreLinked + suppressed filter   ← cross-file rules land here
     → D-01 sort (path, rule, line, col)

  PARITY HARNESS (per batch, test-only):
     fixture tree (compiling) ──> old engine .Lint ─┐
                    └──────────────> new engine .LintLinked ─┴─> sorted (rule,path,line,col) multiset compare
```

### Recommended Project Structure

```
internal/rules/
├── builder.go               # OLD registry — untouched (names, groups, config)
├── builder_linked.go         # NEW: NewLinked(cfg) mirroring New (planner may fold into builder.go)
├── casing.go                 # shared, untouched (byte-identical main/lab)
├── lint_helpers.go           # re-aligned in Wave 0/6: mismatches → new contract + main first-exempt variant
├── <rule>.go                 # OLD rule impls — untouched until Phase 5
├── <rule>_test.go            # REWRITTEN per port: new impl + main's carried expectations
└── parity_<batch>_test.go    # NEW per batch: both engines, shared fixtures
internal/core/
├── lint_model.go             # += linked AppendIssue/buildError variant (rename-map name)
└── (rest untouched)
```

### Pattern 1: Rule port shape (per-file, main-diverged example)

**What:** lab's protoast walk + main's predicate/position.
**Example** — MessagePascalCase (regexp wins, keyword position wins):
```go
// Port: lab walk (Visitor) + main semantics (regexp + Meta.Pos ≡ Decl.Pos)
var _ core.LintRule = (*MessagePascalCaseLinked)(nil)

func (m *MessagePascalCaseLinked) LintRuleName() string { return "MESSAGE_PASCAL_CASE" }

var pascalCase = regexp.MustCompile("^[A-Z][a-zA-Z0-9]+(?:[A-Z][a-zA-Z0-9]+)*$")

func (m *MessagePascalCaseLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
    var res []core.Issue
    info.File.Walk(protoast.Visitor{Message: func(message *protoast.Message) {
        if !pascalCase.MatchString(message.Name) {
            res = core.AppendLinkedIssue(res, m, message.Pos, message.Name) // Decl.Pos = keyword, NOT NamePos
        }
    }})
    return res, nil
}
```
(Final names — `AppendLinkedIssue`, `LintRuleName`, `<Name>Linked` — planner's discretion under Option A; keep the rename map comment updated.)

### Pattern 2: ModuleRule port with main's first-exempt semantics

Main's group rules are first-wins caches; the lab's `groupedValues.mismatches` reports everyone. The port keeps the Collect/Finish shape but Finish reproduces main:
```go
// Finish: first value per group (walk order) exempt; later differing files issue at their own pos.
func (d *DirectorySamePackageLinked) Finish() []core.IssueInfo {
    var res []core.IssueInfo
    for _, dir := range d.files.order {           // add() preserves insertion order (existing helper)
        first := d.files.groups[dir][0]
        for _, v := range d.files.groups[dir][1:] {
            if v.value != first.value {
                for _, issue := range core.AppendLinkedIssue(nil, d, v.pos, v.source) {
                    res = append(res, core.IssueInfo{Issue: issue, Path: v.path})
                }
            }
        }
    }
    d.files = groupedValues{}
    return res
}
```
Optionless/skip rules (#16–22): Collect adds a value **only when the option exists** (main skips silently); PackageSameDirectory/DirectorySamePackage always add. #28 groups by file and exempts the first user of each written type per file.

### Anti-Patterns to Avoid
- **Porting lab semantics because the code is prettier** (helper casing, report-everyone, cross-file uniqueness, NamePos): every one is a Phase-7 corpus parity break. Main wins; improvements are deferred decisions, never silent.
- **Reusing `groupedValues.mismatches` as-is for group rules** — its report-everyone behavior contradicts main's first-exempt for all 9 group rules.
- **Self-suppression in rules** — standing constraint (b); the engine filters centrally.
- **Editing old rule files "to share code"** — old path frozen; new impls are separate types.
- **Non-compiling parity fixtures** — engines legitimately diverge on broken files (COMPILE routing); parity fixtures must compile.
- **Hand-rolled comment scanning in comment rules** — attachment facts come from protoast (`Decl.Leading` etc.); the suppression machinery stays in the engine.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Per-file traversal | Nested hand-written recursion | `protoast.File.Walk(Visitor)` | The lab's proven mechanic; recursion coverage pinned by pilot fixture (nested enums/messages) |
| Issue construction (new contract) | Ad-hoc `core.Issue` literals scattered per rule | One `AppendLinkedIssue`/`buildError` variant in core | Single naming point (`LinkedRuleName`), rename-map tracked for Phase 5 |
| Collect/Finish test driver | Per-test lifecycle code | `validate()` helper from lab `init_test.go` (ported) | One lifecycle driver for all 10 ModuleRule tests |
| Parity comparison | Per-rule bespoke comparisons | The batch harness (sorted multiset) | Uniform, localizes failures to the batch |
| Registry/group/config plumbing | Second validation path | Reuse `validateConfigRuleNames`, `unwrapLintGroups`, `unwrapIgnoreOnly` in `NewLinked` | Phase 3 D-06: `rules.New` remains THE validator |

## Runtime State Inventory

(rename/migration phase — all five categories answered explicitly)

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — rules are stateless per run; no persistent state involved | None |
| Live service config | None — no external services | None |
| OS-registered state | None | None |
| Secrets/env vars | `EnumZeroValueSuffix.Suffix` carries `ENV:"ENUM_ZERO_VALUE_SUFFIX"` tag in main (lab dropped it) — keep main's tags; no key renames | None (keep main tags) |
| Build artifacts | `bin/golangci-lint` pre-existing; the Phase-1 `internal/rules/lint_helpers.go` golangci paths-exclusion in lint config must be **removed** when helpers gain callers (STATE.md Phase-4 note); no schema artifacts change (`task schema:check` clean) | Remove exclusion in the finishing batch; re-check QF1001 casing note at next lab re-sync |

## Common Pitfalls

### Pitfall 1: Position drift (keyword → name token) — the systematic parity killer
**What goes wrong:** porting lab rules verbatim shifts columns/offsets by the keyword length (`message `=8, `rpc `=4, `service `=8, type-name width for fields) on nearly every finding; also 0:0 → 1:1 for PackageDefined/FileLowerSnakeCase.
**Why:** lab rules report `NamePos`/`NumberPos`/`Request.Pos`; main reported `Meta.Pos` (keyword/first token).
**How to avoid:** emit `Decl.Pos` where main used `Meta.Pos` (test-verified equality for Import/Field/Message/Package/RPC/Option kinds); zero Position literal for the two file-level rules; carried main test expectations are the oracle; Wave-0 probe locks the mapping table.
**Warning signs:** parity test diffs with constant column deltas; batch gate fails on position-only mismatches.

### Pitfall 2: Casing-helper vs regexp semantics
**What goes wrong:** silently "upgrading" regexps to `casing.go` helpers changes findings on acronyms (`EnumPascalCase`, `ServicePascalCase`), digits (`OneofLowerSnakeCase`, `PackageLowerSnakeCase`, `RPCPascalCase`), underscores/empty (`EnumValueUpperSnakeCase`, `MessageFieldLowerSnakeCase`), and the prefix algorithm (`EnumValuePrefix` acronym + name-only-value cases).
**How to avoid:** keep main's regexp/algorithm per inventory; fixtures for each edge class in the casing batch.
**Warning signs:** new findings on googleapis-style acronym names at Phase 7.

### Pitfall 3: Comment-presence predicate divergence
**What goes wrong:** lab's `leadingCommentDocuments` treats directive-only comments as "not documented" → new issues on files main accepted.
**Why:** main checked `len(Comments)==0` on the go-protoparser model; the lab rewrote the predicate.
**How to avoid:** port the presence predicate ("no attached comment"), not the lab's directive filter; fixture: declaration whose only comment is a directive.
**Warning signs:** COMMENT_* findings appearing on suppression-heavy fixtures.

### Pitfall 4: Group-rule semantics (first-exempt, optionless skip, everyone-reported)
**What goes wrong:** `groupedValues.mismatches` reports every file of a disagreeing group and counts optionless files as empty values — main exempts the first file and skips optionless files entirely.
**How to avoid:** main-semantics Finish (Pattern 2); Collect adds only option-present values for the 7 option rules; fixture sets with 3-file groups where only the 2nd/3rd diverge.
**Warning signs:** extra findings on the first file of each group.

### Pitfall 5: Deliberate lab improvements leaking in (cross-file uniqueness, resolved imports, swift_prefix fix, empty-enum guard)
**What goes wrong:** four lab behaviors are genuine improvements that main lacks — cross-file `RPCRequestResponseUnique`, resolution-based `ImportUsed.validateResolved`, correct `swift_prefix` option name (main compares `ruby_package`), empty-enum guard. Adopting any silently is a behavior change.
**How to avoid:** Phase 4 ships main behavior (#28 per-file first-exempt on written names; #12 syntax walk unconditionally; #22 bug-for-bug with a pinning regression test; #5 keeps the guard as the single documented crash-fix deviation). Each improvement gets an explicit deferred-decision note for post-parity phases.
**Warning signs:** parity cells failing on multi-file fixtures; corpus diffs on packages with divergent `ruby_package`/`swift_prefix`.

### Pitfall 6: ModuleRule state and tests
**What goes wrong:** ported ModuleRule tests keep calling `rule.Validate(per file)` — a no-op stub → zero findings, tests "pass" vacuously or fail confusingly; receiver-state writes in `Validate` would race under parallel windows.
**How to avoid:** the 10 ModuleRule tests switch to the `validate()` Collect/Finish driver (lab `init_test.go` pattern); per-file rules stay stateless (local structs like the lab's `importUsage`); Collect mutates only in the serial drain.
**Warning signs:** a ModuleRule test asserting nothing; `-race` failures only in engine-level tests.

### Pitfall 7: Two-contract name collisions
**What goes wrong:** adding `Validate(LinkedFileInfo)` alongside the old `Validate(ProtoInfo)` on one type is impossible (same method name); renaming old types breaks reflection-derived registry names (schema/config-pinned).
**How to avoid:** Option A (suffixed new types + `LintRuleName()` override, additive `LinkedRuleName` preference); document every new name in the rename-map comment for Phase 5 consolidation.
**Warning signs:** compile errors on duplicate type/method names; registry names with `_LINKED` suffixes leaking into schemas.

### Pitfall 8: Dormancy leak
**What goes wrong:** edits to old rule files, old engine files, or shared helpers change old-path output mid-phase.
**How to avoid:** old files byte-frozen; all changes additive (new files/types/constructor); full `-race` suite + old-path behavior-diff-empty at every intermediate state; `schema:check` clean (no name changes).
**Warning signs:** any diff in old-engine test outputs before their file is ported.

### Pitfall 9: `protoast.Walk` recursion coverage
**What goes wrong:** main walked nested messages/enums explicitly; the lab relies on `File.Walk` recursing into nested declarations — if any nested kind is missed, findings understate main.
**How to avoid:** pilot fixture with nested enum-in-message, oneof fields, extends, map fields; compare old-vs-new counts [ASSUMED: Walk covers all kinds — pin in Wave 0].
**Warning signs:** missing findings on nested declarations.

### Pitfall 10: Option value quoting
**What goes wrong:** main compared `option.Constant` (go-protoparser, as-written) vs lab `optionString(opt.Value)` (unquoted) — mixed quoting styles in a package could compare differently.
**How to avoid:** pick main's comparison basis; fixture with `'single'` vs `"double"` quoted options only if reachable; note as edge case.
**Warning signs:** none expected on normal corpora.

## Code Examples

### Parity harness cell (per batch)
```go
// parity_minimal_test.go (sketch)
func assertBatchParity(t *testing.T, cfg config.LintConfig, files map[string]string) {
    t.Helper()
    root := writeFixtures(t, files) // compiling fixtures in t.TempDir()
    oldRules, ignoreOnly, err := rules.New(cfg); require.NoError(t, err)
    oldApp := core.New(core.Options{Logger: logger.NewNop(), Rules: oldRules,
        IgnoreOnly: ignoreOnly, AllowCommentIgnores: true, KnownLintRules: rules.AllRuleNames()})
    oldIssues, err := oldApp.Lint(t.Context(), disk.NewFSWalker(root, ".")); require.NoError(t, err)

    newRules, _, err := rules.NewLinked(cfg); require.NoError(t, err)
    newApp := core.New(core.Options{Logger: logger.NewNop(), LinkedRules: newRules,
        AllowCommentIgnores: true, KnownLintRules: rules.AllRuleNames()})
    newIssues, err := newApp.LintLinked(t.Context(), disk.NewFSWalker(root, ".")); require.NoError(t, err)

    require.Equal(t, fingerprint(oldIssues), fingerprint(newIssues),
        "old-vs-new parity: sorted (rule, path, line, col) multisets")
}
// fingerprint sorts by the D-01 key; old-engine output order is nondeterministic (map iteration).
```

### Lab test lifecycle driver (port into `init_test.go`)
```go
// Source: lab internal/rules/init_test.go (validate runs rule over infos the way lint does)
func validate(rule core.LintRule, infos ...core.LinkedFileInfo) ([]core.Issue, error) {
    if moduleRule, ok := rule.(core.ModuleRule); !ok {
        var issues []core.Issue
        for _, info := range infos {
            found, err := rule.Validate(info)
            if err != nil { return nil, err }
            issues = append(issues, found...)
        }
        return issues, nil
    } else {
        for _, info := range infos {
            if err := moduleRule.Collect(info); err != nil { return nil, err }
        }
        var issues []core.Issue
        for _, found := range moduleRule.Finish() { issues = append(issues, found.Issue) }
        return issues, nil
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Regexps per rule for casing | Shared `casing.go` helpers (buf xstrings port) | lab rewrite | Helpers already vendored (Phase 1); Phase 4 keeps main regexps for parity; helpers remain for Phase 6 rules |
| Receiver-state caches in group rules (`cache map` + `lazyInit`) | Stateless per-file rules + ModuleRule Collect/Finish | lab engine (Phase 3 contract) | 10 rules lose receiver state; first-exempt semantics move into Finish |
| Per-file uniqueness (RPCRequestResponseUnique), syntax-based ImportUsed | Cross-file/qualified, resolution-based | lab rewrite | Deferred behind parity (deliberate post-parity decisions, Pitfall 5) |
| Keyword positions in findings | Name-token positions | lab rewrite | Deferred — main positions win until a deliberate post-parity change |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `protoast.Decl.Pos` equals main's `Meta.Pos` for all keyword-decl kinds (test-verified for Import, Field, Message, Package, RPC, Option; Enum/EnumValue expected same) | Position mapping | Wrong mapping shifts columns; caught immediately by carried tests — cheap to fix |
| A2 | `protoast.File.Walk` recurses into nested messages/enums/oneofs/extends/map fields | Pitfall 9 | Understated findings on nested decls; pinned by Wave-0 pilot fixture |
| A3 | Main's `PackageSameSwiftPrefix` comparing `ruby_package` is a bug to pin, not intended behavior | #22 / Pitfall 5 | If intended, the pin test documents the wrong thing — surfaced at discuss |
| A4 | go-protoparser surfaces proto2 group fields such that main's CommentField effectively checks them (lab added `!field.Group`) | #35 | Divergence only on proto2 group declarations; pin with a proto2 fixture if reachable |
| A5 | Parity corpora (api-common, googleapis-nocompute) do not exercise the divergent edges in a way that makes main-wins vs lab-wins indistinguishable at Phase 7 (e.g., they compile; option quoting is consistent) | Batch mechanism | If corpora DO exercise an edge the fixtures missed, Phase 7 lint-diff catches it — the gate still holds, later than ideal |
| A6 | `NewLinked` returning `(rules, ignoreOnly-map, error)` mirroring `New` is the right shape for `Options.LinkedRules` wiring | Coexistence | Cosmetic; planner may reshape |
| A7 | Keeping `EnumFirstValueZero`'s empty-enum guard (lab) while keeping main's string-compare is acceptable as the one deliberate deviation | #5 | If strict bug-for-bug is demanded, the port re-introduces a panic on a reachable-in-fallback input — recommend keeping the guard |

## Open Questions (RESOLVED 2026-11-19 — dispositions recorded during plan revision; each implemented in the phase plans as cited)

1. **swift_prefix bug-for-bug vs fix** (#22) — RESOLVED: bug-for-bug pin ratified; 04-04 Task 3 ports `PackageSameSwiftPrefixLinked` comparing the literal `ruby_package` option with documented pin comment, regression cells in the batch parity suite (`swift_prefix_bug_pin`) and carried test, threat T-04-09 mitigates a silent "fix"; the corrected behavior is a deferred post-parity decision (ledgered in the 04-07 SUMMARY).
2. **Deferred lab improvements** (cross-file RPCRequestResponseUnique, resolution-based ImportUsed, NamePos positions, helper casing) — RESOLVED: parked as explicit deferred-decision items, never adopted silently — 04-04 Task 2 (ImportUsed syntax-walk unconditionally, prohibition + grep), 04-05 Tasks 2-3 (main algorithms byte-verbatim, per-file uniqueness), and the phase-close deferred-decisions ledger in 04-07 Task 3 item 7, so Phase 7 inherits no silent expectations.
3. **buildCore wiring timing** — RESOLVED (planner discretion): deferred to Phase 5; `Options.LinkedRules` is populated only from tests during Phase 4 (dormancy invariant stated in every batch plan's standing constraints; 04-01 records it as the planner decision), keeping `easyp lint` byte-identical until the flip.
4. **Position mapping probe results** (A1) — RESOLVED: closed by the Wave-0 probe `TestDeclPositionProbe` (04-01 Task 3) which pins `Decl.Pos == Meta.Pos` for all ten declaration kinds and records the mapping table in a comment before batch 1 starts; carried main test literals remain the per-batch oracle.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | everything | ✓ | go1.27.3 (go.mod 1.26.6) | — |
| `task` / gotestsum | `task test`/`task lint` | ✗ not on PATH | — | raw `go test -race -count=1 ./internal/...`, `go build -o easyp ./cmd/easyp` |
| golangci-lint | `task lint:go` + exclusion removal | ✓ bin/golangci-lint | pinned | — |
| etalon clone (porting source) | all rule ports | ✓ /tmp/gsd-research/compiler-lab @ 08cef10 | — | `gh repo clone easyp-tech/compiler-lab` |
| testify | tests | ✓ (go.mod) | — | — |
| Parity corpora + `lab.py` | Phase 7 gate only — NOT needed per batch | ✗ private repo + rebuild | — | per-batch fixture parity harness (this document's recommendation) |

**Missing dependencies with no fallback:** none for this phase (corpora are a Phase-7 concern with its own spike).
**Missing dependencies with fallback:** `task`/gotestsum → raw `go` commands (operating mode since Phase 1).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go standard `testing` + testify; runner gotestsum via `task test` (absent → raw go test) |
| Config file | none (convention-driven: colocated `*_test.go`, `-race` mandatory) |
| Quick run command | `go test -race -count=1 ./internal/rules/... ./internal/core/...` |
| Full suite command | `go test -race -count=1 ./internal/...` (known pre-existing wazero race in internal/adapters/plugin + internal/generation — deferred, poisons full-suite `-race` [VERIFIED: STATE.md:132]) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LINT-PARITY-02 | Ported rule reproduces main semantics on divergence-edge fixtures (per inventory row) | parity (both engines, one fixture tree) | `go test -race -count=1 ./internal/rules/ -run 'Parity'` | ❌ Wave 0 + per batch |
| LINT-PARITY-02 | Ported rule passes main's carried colocated expectations | unit (colocated) | `go test -race -count=1 ./internal/rules/ -run Test<Rule>` | rewritten per port |
| Standing (a) | `readSource` invoked exactly once per corpus file on compiling corpus | integration (disk-instrumented walker) | `go test -race -count=1 ./internal/core/ -run 'ReadSourceOnce'` | ❌ Wave 0 |
| Standing (b) | Ported rules never self-suppress; `comment_suppression_central`, `finish_comment_refilter`, `compile_bypasses_suppression` stay green | regression | `go test -race -count=1 ./internal/core/ -run 'Suppression'` | ✅ Phase-3 cells; kept green per batch |
| Criterion 1 | Each batch verified before the next; corpus gate untouched | batch gate | batch parity suite green | ❌ per batch |
| Criterion 3 | All intermediate states green; schemas untouched | suite + schema | full suite; `go run ./cmd/easyp schema-gen` diff-empty | ✅ infra exists |
| D-03 list | Exactly the 10 documented rules implement `ModuleRule` (no more, no fewer) | guard (compile-time asserts + count test) | `go test ./internal/rules/ -run TestModuleRuleSet` | ❌ Wave 0 (cheap guard) |

### Sampling Rate
- **Per task commit:** quick run (rules + core packages) + the batch's parity suite.
- **Per wave merge:** full suite (`-race`) + `golangci-lint run` (exclusion removed at finishing batch).
- **Phase gate:** full suite + build + `schema:check`-equivalent drift-free + all batch parity suites green. Corpus lint-diff NOT required at this phase's gate (Phase 7 owns LINT-PARITY-01).

### Wave 0 Gaps
- [ ] `internal/core` — linked issue-construction helper (`AppendLinkedIssue`/buildError variant for `LintRule`), rename-map comment updated
- [ ] `internal/rules/init_test.go` — protoast-based `parse`/`parseProto` + `validate()` lifecycle driver (old-path parse helper retained for still-unported rules)
- [ ] `internal/rules/parity_*_test.go` — harness skeleton + fingerprint comparator
- [ ] Position-mapping probe test (one fixture per decl kind, old-vs-new positions)
- [ ] `TestModuleRuleSet` guard (10 rules, D-03 list)
- [ ] Disk-instrumentation cell (`readSource` once per corpus file)
- [ ] Pilot rule port end-to-end (PackageDirectoryMatch) proving harness + helpers + carryover

## Security Domain

`security_enforcement: true` (ASVS L1) — internal CLI refactor; no auth/session/network/crypto surfaces. Rules are pure in-memory predicates over parsed trees.

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|------------------|
| V2/V3/V6 | no | N/A — local CLI, no crypto |
| V4 Access Control | inherited | All file access remains in the engine's `openImportFile`/`openDiskSource` path (Phase 3); rules perform no I/O |
| V5 Input Validation | unchanged | `rules.New`/`NewLinked` validate `lint.use` names via the existing single validator (D-06); no new config surfaces; no schema changes |
| V14 Config | minimal | `EnumZeroValueSuffix`/`ServiceSuffix` config plumbed through builder as before; keep main's struct tags |

### Known Threat Patterns for this stack
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| None new — rules add no I/O, parsing, or process boundaries | — | Engine-level controls (path guards, directive validation) carried from Phases 1–3 |

## Sources

### Primary (HIGH confidence)
- Full main-vs-lab diff of all 41 rule files + all colocated tests, run this session (evidence quoted inline): `/home/nil/DiskD/W/Djarvur/easyp/internal/rules/*.go` vs `/tmp/gsd-research/compiler-lab/easyp-cli/internal/rules/*.go` @ `08cef10`
- In-repo (Read this session, line ranges cited inline): `internal/core/lint_linked.go`, `lint_model.go`, `dom.go`, `core.go`; `internal/rules/{builder.go, lint_helpers.go, casing.go}`; `internal/protoast/protoast.go`, `walk.go`
- Planning: `.planning/ROADMAP.md` § Phase 4 + backlog 999.1/999.2; `.planning/REQUIREMENTS.md`; `.planning/STATE.md` (Phase-3 decisions, standing constraints, deferred items); `.planning/phases/03-.../03-RESEARCH.md` (engine contract, Pitfall 7, rename map)
- Lab tree (Read this session): `internal/core/dom.go`, `internal/rules/lint_helpers.go`, `init_test.go` (lab variants)

### Secondary (MEDIUM confidence)
- None — no external web sources needed; all claims are source-read.

### Tertiary (LOW confidence)
- Marked `[ASSUMED]` inline: A1 (partial), A2, A3, A4, A5.

## Metadata

**Confidence breakdown:**
- Rule inventory & divergences: HIGH — every file diffed byte-level this session; divergences quoted from actual code
- Position mapping: HIGH for verified kinds (test-diff evidence), MEDIUM overall (A1 remainder pinned in Wave 0)
- Batch mechanism: HIGH — grounded in the live-oracle fact and ROADMAP constraints
- Coexistence design: MEDIUM — engineering judgment; three options costed, planner ratifies

**Research date:** 2026-11-19
**Valid until:** 2026-12-19 (stable: internal codebase work; etalon pinned SHA)
