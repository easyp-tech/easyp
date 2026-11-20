# Phase 6: Новые правила и регенерация схем/MCP - Research

**Researched:** 2026-11-21
**Domain:** Go lint-rule porting (compiler-lab etalon), registry/validator wiring, JSON Schema + MCP regeneration on a post-flip single engine
**Confidence:** HIGH (all claims verified against in-repo sources and the lab etalon checkout this session; corpora availability is the one open environment item)

## Summary

Phase 6 adds the three lab-only rules to the post-flip single registry: `PACKAGE_NO_IMPORT_CYCLE` (cross-file `core.ModuleRule`, MINIMAL) and `SYNTAX_SPECIFIED` + `FIELD_NOT_REQUIRED` (per-file `core.Rule`, BASIC). Every symbol the lab rule files reference already exists in this repo with the identical name and shape — the engine contract (lint_model.go), the protoast model, `core.AppendIssue`, `core.ConvertImportPath`, and the reflection-based `core.GetRuleName` were all aligned to the etalon during Phases 1–5. All three files are **verbatim port candidates** with zero seam adaptations; the lab itself carries **no unit tests** for them, so the only lab-side pins are name-presence tests (builder/describe/schema), which port as one-line flips from NotContains to Contains.

The name-plumbing is already a closed loop: `rules.AllLintUseValues()` (fed by `builder.go` `add*` group lists) is the single source for (a) rule-name validation (`selection.go` `ValidateNames`), (b) the JSON Schema `lint.use` enums (`internal/config/v1/schema.go`), (c) MCP description enums (through `v1.SchemaJSON`), and (d) directive validation (`core.Options.KnownLintRules`). Adding the three types to `builder.go`'s `addMinimal`/`addBasic` and to `New`'s `allRules` list in `builder_linked.go` therefore updates schemas, MCP, directives, and migration validation *automatically* — the phase's atomicity requirement is satisfied by one commit covering: three new rule files + two registry lists + special-case removal + guard/test flips + `go run ./cmd/easyp schema-gen` regeneration.

Two structural findings shape the plan. First, the special-deviation is exactly one 4-line branch in `internal/rules/selection.go` (`ValidateNames`), pinned by three test files that must flip in the same commit. Second, **every v1 config seeds `use` with `MINIMAL`** (`policy.go`), so after this phase every `easyp lint` run gains the cross-file cycle rule — a user-visible behavior change that also means the Phase-7 main-vs-new corpus parity must filter the three new rule names out (main does not have them).

**Primary recommendation:** one atomic commit — verbatim-port the three files, extend both registry lists in lab order, delete the `selection.go` special case, flip the five pinned tests, extend `TestModuleRuleSet` to eleven names, regenerate schemas via `go run ./cmd/easyp schema-gen` — then prove new-rule parity against the lab CLI binary on an in-repo "teeth" corpus, with the acceptance-corpus run as the corpora-availability-gated second leg.

## User Constraints (from CONTEXT.md)

No CONTEXT.md exists for this phase (`has_context: false` from init). No locked decisions or discretion areas were recorded — the ROADMAP design note ("механизм проверки паритета новых правил против лабораторного CLI, LINT-NEWRULES-03") is **unresolved** and is resolved as a design note in this document (Architecture Patterns → Pattern 4).

## Project Constraints (from AGENTS.md)

- **v1 YAML schemas**: change types and schema rules in `internal/config/v1`, then regenerate `schemas/` with `task schema:generate`. Do not hand-edit generated schema JSON (`schemas/*.json` — six files).
- **Tests**: use `testify`; mirror patterns in neighboring rules (each rule + colocated `<file>_test.go`).
- **Do not edit anything inside `internal/easypc`** — sealed vendor copy.
- **Do not commit** the local `easyp` binary, `coverage.out`, or generated dependency noise.
- After behavior changes run the relevant tests; for config schema changes run `task schema:check` before finishing.
- New lint rules: `internal/rules/<rule>.go` + colocated test, registered via the existing builder.
- Project skills loaded: `go-code-style`, `go-testing` patterns apply (testify, table-driven cells, exact-literal assertions).

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| LINT-NEWRULES-01 | `MINIMAL` includes `PACKAGE_NO_IMPORT_CYCLE` (cross-file, ModuleRule, lab placement) | Verbatim port + `builder.go` `addMinimal` 5th slot + `New` MINIMAL block 5th slot + `TestModuleRuleSet` 11-name exact-set (Pattern 2) |
| LINT-NEWRULES-02 | `BASIC` includes `SYNTAX_SPECIFIED` and `FIELD_NOT_REQUIRED` (lab placement) | Verbatim ports + `addBasic` tail two slots + `New` BASIC block tail (Pattern 2) |
| LINT-NEWRULES-03 | New-rule findings identical to lab CLI output on corpora | Lab CLI binary builds at 08cef10; `MINIMAL` is auto-seeded by every v1 config so corpora configs already enable the rules; compare (RuleName, Path, Line, Column) multisets filtered to the three names; in-repo teeth corpus guarantees non-vacuity (Pattern 4) |
| LINT-NEWRULES-04 | Schemas include new names in `lint.use` enums; `task schema:check` drift-free; MCP descriptions reflect the set; one atomic change | `AllLintUseValues` → schema enums → MCP `v1.SchemaJSON` closed loop; regen via `go run ./cmd/easyp schema-gen`; five test pins flip (Pattern 5) |

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Rule implementations (3 files) | `internal/rules` | — | Rules own their findings; engine owns suppression |
| Cross-file collection + Finish drain | `internal/core` engine | — | `collectFile`/drain already exist; rules only implement the interface |
| Registry + group membership | `internal/rules` (`builder.go`, `builder_linked.go`) | — | Single name source (`AllLintUseValues`), single constructor (`New`) |
| Rule-name validation | `internal/rules` (`selection.go`) | consumed by `internal/migration` | D-06: one validator surface; no engine-side validation |
| JSON Schema enums | `internal/config/v1/schema.go` → `internal/schemagen` → `schemas/` | — | Enums derive from `rules.AllLintUseValues()` at build time |
| MCP descriptions/enums | `mcp/easypconfig` | reads `v1.SchemaJSON` | No independent enum source; test pin flips only |
| Directive validation (`easyp:disable`) | `internal/api/runtime.go` (`KnownLintRules`) | — | Auto-follows `AllRuleNames()` |
| Parity vs lab CLI | external stand (`compiler-lab/bench/lab.py` + lab binary) | phase script | Comparison is outside the repo; lab.py already snapshots/diffs |

## Standard Stack

### Core (all in-repo — no external packages)

| Surface | Location | Purpose | Why Standard |
|---------|----------|---------|--------------|
| `core.Rule` / `core.ModuleRule` | `internal/core/lint_model.go:9-41` | Per-file / cross-file rule contracts | The only rule contracts since the Phase-5 flip |
| `core.ProtoInfo` | `internal/core/lint_model.go:53-67` | Linked per-file model (Path, ImportPath, File, ImportedFiles, Resolved) | What `Validate`/`Collect` receive |
| `core.AppendIssue` / `GetRuleName` | `internal/core/lint_model.go:70-87` | Issue construction; reflection UPPER_SNAKE naming | Single issue-construction point (etalon 4-arg shape) |
| `protoast` model | `internal/protoast/` | SyntaxSpecified, Import{Location,Pos}, Field{Label,NamePos}, Visitor/Walk, FilePosition | The model the lab rules were written against |
| `rules.New` + `add*` | `internal/rules/builder_linked.go`, `builder.go` | Registry + name lists | THE registry post-flip (single constructor) |
| `schemagen` CLI | `go run ./cmd/easyp schema-gen` | Writes `schemas/*.json` | AGENTS.md: never hand-edit |
| testify | existing dep | assertions | Project convention |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Adapt rule files to Phase-4 "main-wins" conventions (Decl.Pos positions etc.) | Verbatim lab port | **Verbatim wins**: LINT-NEWRULES-03 demands lab-identical output; the main-wins decisions (T-04-03 NamePos, etc.) governed *ports of existing main rules* — these three have no main variant |
| Hand-edit `schemas/*.json` | `go run ./cmd/easyp schema-gen` | Forbidden by AGENTS.md; drift gate exists |
| Modify lab.py for filtered diff | Phase-local compare script | lab.py is the lab's Phase-7 asset; a small script keeps the lab untouched |
| Full-output corpus diff | Filter to the three rule names | Ported rules carry deliberate main-vs-lab divergences (swift_prefix, directive comments…); full diff would fail on rules this phase must not touch |

**Installation:** none — zero new dependencies; `go.mod`/`go.sum` must stay diff-empty after `go mod tidy` (standing gate).

## Package Legitimacy Audit

**No external packages are installed this phase** — all work is in-repo Go code plus regeneration of existing artifacts. Audit skipped as not applicable; no `checkpoint:human-verify` install gates needed.

## Architecture Patterns

### System Architecture Diagram

```
                    lab etalon (compiler-lab/easyp-cli @ 08cef10)
                    package_no_import_cycle.go / syntax_specified.go /
                    field_not_required.go   +   builder.go add* order
                                   │ verbatim port
                                   ▼
   ┌───────────────────────── internal/rules ─────────────────────────┐
   │ builder.go addMinimal/addBasic ──► AllGroups/AllRuleNames/       │
   │        │                          AllLintUseValues               │
   │        │                               │                         │
   │        ▼                               │                         │
   │ builder_linked.go New(allRules)        │ (same commit)           │
   │        │                               │                         │
   │        ▼                               ▼                         │
   │ selection.go ValidateNames ◄── allowed = AllLintUseValues        │
   │ (special case DELETED)                                           │
   └──────────────┬───────────────────────────────┬───────────────────┘
                  │ rules.New feeds                │ AllLintUseValues
                  ▼                               ▼
        internal/core engine                  internal/config/v1/schema.go
        (Collect → Finish drain;               (lint.use enums ×3 paths)
         D-01 total sort)                            │
                  │                        ┌─────────┴──────────┐
                  ▼                        ▼                    ▼
        easyp lint output            go run ./cmd/easyp       mcp/easypconfig
        (JSON lines: RuleName/       schema-gen               (v1.SchemaJSON →
         Path/Position)              ──► schemas/*.json        SchemaByPath/Describe)
                  │                             │                    │
                  └──────────────┬──────────────┘                    │
                                 ▼                                   ▼
                  parity vs lab CLI binary                describe_test pin flips
                  (rule,file,line,col) filter             (NotContains → Contains)
```

### Pattern 1: The three lab rule sources — verbatim-or-adapt classification (Output question 1)

**Classification: all three files are VERBATIM port candidates. Zero seam adaptations required.** Every referenced symbol exists in this repo with the identical name, package, and signature. Evidence per symbol:

| Lab symbol | Easyp counterpart [VERIFIED] |
|------------|------------------------------|
| `core.ModuleRule` (Collect/Finish) | `internal/core/lint_model.go:35-41` — `Rule` + `Collect(ProtoInfo) error` + `Finish() []IssueInfo` |
| `core.Rule` (Message/Validate) | `internal/core/lint_model.go:10-16` — `Validate(ProtoInfo) ([]Issue, error)` |
| `core.ProtoInfo{Path, File, ImportedFiles}` | `internal/core/lint_model.go:53-67` |
| `core.AppendIssue(nil, rule, pos, source)` | `internal/core/lint_model.go:73-80` — same 4-arg etalon shape |
| `core.IssueInfo{Issue, Path}` | `internal/core/dom.go:28-31` |
| `core.ConvertImportPath` | `internal/core/dom.go:153-155` — `ImportPath(strings.Trim(source, "\""))` |
| `core.Position` | `internal/core/dom.go:19` — `type Position = protoast.Position` |
| `protoInfo.File.SyntaxSpecified` | `internal/protoast/protoast.go:79-80` |
| `protoast.FilePosition(path)` | `internal/protoast/protoast.go:27-30` — line 1, column 1 |
| `protoInfo.File.Imports` + `imp.Location` + `imp.Pos` | `internal/protoast/protoast.go:112-117` (`Import{Path, Location, Public, Weak}` embedding `Decl`; `Decl.Pos` at :63-66) |
| `File.PackageName()` | `internal/protoast/protoast.go:95` |
| `File.Walk(protoast.Visitor{Field: …})` | `internal/protoast/walk.go:7-18` |
| `field.Label == "required"` | `internal/protoast/protoast.go:141` — `Label string // "repeated", "optional", "required" or ""` |
| `field.NamePos` | `internal/protoast/protoast.go:66` (`Decl.NamePos`) |
| `sortedKeys`, `packageImport` (package-local) | No collisions in easyp `internal/rules` (grep-verified empty this session) |

Lab rule bodies (quoted from `/tmp/gsd-research/compiler-lab/easyp-cli/internal/rules/`, verified this session):

- `syntax_specified.go` (27 lines): `Message()` returns `"syntax should be specified"`; `Validate` appends `core.AppendIssue(res, s, protoast.FilePosition(protoInfo.Path), protoInfo.Path)` when `!protoInfo.File.SyntaxSpecified`.
- `field_not_required.go` (29 lines): `Message()` returns `"field should not be required"`; `Validate` walks `protoInfo.File.Walk(protoast.Visitor{Field: func(field *protoast.Field) { if field.Label == "required" { res = core.AppendIssue(res, f, field.NamePos, field.Name) } }})` — SourceName is the field name, position is the declared-name token.
- `package_no_import_cycle.go` (109 lines): `Message()` returns `"package import cycle"`; stateless `Validate` stub; `Collect` builds `map[pkg]map[importedPkg][]packageImport` keyed via `protoInfo.ImportedFiles[core.ConvertImportPath(imp.Location)]`, skipping same-package imports; `Finish` walks package/import keys in sorted order, runs a recursive DFS (`cycle`) from each imported package back to the source package, and reports **every import** leading into a cycle at the import-statement position (`imp.Pos`) with `SourceName = strings.Join(cycle, " -> ")` and `IssueInfo.Path = imp.path`; state reset (`p.edges = nil`) after Finish.

Position-semantics contrast to document in the plan: the ported 41 rules use main-wins positions (Decl.Pos keyword positions, T-04-03); `FIELD_NOT_REQUIRED` uses the lab's `NamePos` and `SYNTAX_SPECIFIED` the lab's `FilePosition` — correct here because lab parity (LINT-NEWRULES-03) is the oracle for **new** rules, not main parity (LINT-PARITY-02).

**Test porting:** the lab has **no unit tests** for these rules (verified: no `*_test.go` in the lab mentions `SyntaxSpecified`/`FieldNotRequired`/`PackageNoImportCycle`). The portable lab pins are name-presence only:
- lab `internal/rules/builder_test.go` `TestAllLintUseValues`: `require.Contains(t, values, "PACKAGE_NO_IMPORT_CYCLE")` + uniqueness.
- lab `mcp/easypconfig/describe_test.go:287-303` `TestDescribeLintSelectionMatchesExecutableRules`: Describe on `linters.enable`, `linters.disable`, `issues.exclude-rules[].linters` must contain the new names.
- lab `internal/config/v1/lint_selection_test.go:57`: schema data contains the name.

Rule-behavior tests must be **authored fresh** in easyp style (testify, `parseLinkedFile` harness from `internal/rules/init_test.go:32-60`, colocated `<rule>_test.go`), with the lab CLI run as the behavioral oracle. For `PACKAGE_NO_IMPORT_CYCLE`, unit tests need a two-file import-cycle fixture under `internal/rules/testdata/` (parseLinkedFile resolves direct imports from that dir; a cycle requires both files present there) plus an acyclic negative and a same-package-import negative. The Finish issue must pin: every-edge reporting, sorted iteration, `SourceName` = the `"A -> B -> A"` chain, position at the import keyword, `IssueInfo.Path` = the importing file.

### Pattern 2: ModuleRule set extension and registry wiring (Output question 2)

The exact-set guard today pins **ten** names — verbatim from `internal/rules/module_rules_guard_test.go:19-31`:

```go
// moduleRuleNames is the D-03 ten-name set as registry names.
var moduleRuleNames = map[string]bool{
	"DIRECTORY_SAME_PACKAGE":           true,
	"PACKAGE_SAME_DIRECTORY":           true,
	"PACKAGE_SAME_CSHARP_NAMESPACE":    true,
	"PACKAGE_SAME_GO_PACKAGE":          true,
	"PACKAGE_SAME_JAVA_MULTIPLE_FILES": true,
	"PACKAGE_SAME_JAVA_PACKAGE":        true,
	"PACKAGE_SAME_PHP_NAMESPACE":       true,
	"PACKAGE_SAME_RUBY_PACKAGE":        true,
	"PACKAGE_SAME_SWIFT_PREFIX":        true,
	"RPC_REQUEST_RESPONSE_UNIQUE":      true,
}
```

Phase 6 adds `"PACKAGE_NO_IMPORT_CYCLE": true` → the D-03 **eleven**; the file header comment (lines 3-7: "PackageNoImportCycle arrives with Phase 6 and must be absent until then") and `internal/core/lint_model.go:27-34` (the D-03 comment listing the ten and "PackageNoImportCycle arrives with Phase 6") update in the same commit. The name is derived by pure reflection: `toUpperSnakeCase("PackageNoImportCycle")` → `PACKAGE_NO_IMPORT_CYCLE` (`internal/core/dom.go:107-123`, camel-hump splitting — same mechanism as all 41 existing names).

**Registry wiring — three lists must change in one commit:**

1. `internal/rules/builder.go` `addMinimal` (:132-139): append `core.GetRuleName(&PackageNoImportCycle{})` as the **5th** entry, after `PackageSameDirectory` — lab placement verbatim (lab `builder.go:223-231` has exactly this order).
2. `internal/rules/builder.go` `addBasic` (:141-163): append `SyntaxSpecified`, `FieldNotRequired` as the **21st/22nd** entries, after `ServicePascalCase` — lab placement verbatim (lab `builder.go:252-255`).
3. `internal/rules/builder_linked.go` `New` `allRules` (:26-81): insert `&PackageNoImportCycle{}` after `&PackageSameDirectory{}` in the MINIMAL block, `&SyntaxSpecified{}` + `&FieldNotRequired{}` after `&ServicePascalCase{}` in the BASIC block, mirroring the same order.

`builder.go`'s `add*` lists feed `AllGroups`/`AllRuleNames`/`AllLintUseValues`/`unwrapLintGroups`/`unwrapIgnoreOnly`; `New`'s list feeds construction. Membership must agree in both or `New(Use: [MINIMAL])` fails its name lookup. Order between the two files is cosmetic (findings are engine-sorted; `TestAllRuleNames` only checks concat equality) but mirroring the lab exactly keeps the diff reviewable. Group count stays 5 (`TestAllGroups` :13-46 has no per-group size pins); registry count goes **41 → 44** — update the "41" doc comments in `lint_model.go:28` and `builder_linked.go:14`. The REQUIREMENTS/phase-title figure "43" is the known documentation discrepancy already flagged in `lint_model.go:28-29` ("REQUIREMENTS.md says 43, research Pitfall 7") — do not propagate it into new comments.

The Phase-4 `TestRegistryParity` guard died with the old engine at the flip (file no longer exists); group-expansion order is now covered by `TestAllGroups` (order/keys/uniqueness) and `TestAllRuleNames` (concat equality), both auto-adapting.

### Pattern 3: The rule-name validation special-deviation (Output question 3)

**Where it lives:** `internal/rules/selection.go:14-27`, verbatim:

```go
// ValidateNames checks every rule/group, including names that will be disabled.
// The executable registry is also the source for schemas and completion values.
func ValidateNames(names []string) error {
	allowed := AllLintUseValues()
	var issues []error
	for _, name := range names {
		if name == "PACKAGE_NO_IMPORT_CYCLE" {
			issues = append(issues, fmt.Errorf("%w: %s is not implemented", core.ErrInvalidRule, name))
			continue
		}
		if !slices.Contains(allowed, name) {
			issues = append(issues, fmt.Errorf("%w: %s", core.ErrInvalidRule, name))
		}
	}
	return errors.Join(issues...)
}
```

**Why it existed:** main deliberately hard-excluded the name while the lab placed it in `MINIMAL` — the current `builder.go` `add*` lists hold only the 41 ported rules, so `PACKAGE_NO_IMPORT_CYCLE` is *not* in `allowed` today, and the branch turns "not in the allowed set" into the friendlier "is not implemented" error. Phase 3/4 pinned this as main's deliberate exception ("PACKAGE_NO_IMPORT_CYCLE special case pinned by parity guard until Phase 6 flip removes it" — STATE.md Phase-3 decision) so validation behavior stayed byte-stable during the port. It is a "name known but unimplemented" guard, not a members-of-allowed set issue.

**How it's removed:** delete the `if name == "PACKAGE_NO_IMPORT_CYCLE" { … continue }` branch (4 lines). After the registry change the name *is* in `AllLintUseValues()`, so the generic `slices.Contains` path accepts it. The removal and the registry addition are **one commit**: removing the branch first would let the name pass validation and then fail construction in `New` (:94-98) with a different `ErrInvalidRule` wrapping — incoherent intermediate state.

Callers that change behavior automatically: `validateConfigRuleNames` (use/except/ignore_only — `selection.go:29-52`, called by `New` :23), legacy-config migration validation (`internal/migration/convert.go:163` — a legacy config naming the rule now validates instead of erroring: intended, document it), directive validation (`internal/api/runtime.go:77` `KnownLintRules: rules.AllRuleNames()` — `easyp:disable PACKAGE_NO_IMPORT_CYCLE` becomes legal), and the v1 policy validation surface (`internal/config/v1` collects all invalid-name locations; the name leaves that set).

**Pins that must flip in the same commit** (all verified this session):

| File | Line(s) | Current pin | Becomes |
|------|---------|-------------|---------|
| `internal/rules/validation_parity_test.go` | 70-78 | cell `package_no_import_cycle_special_case` expects `"PACKAGE_NO_IMPORT_CYCLE is not implemented"` | Replace with a valid-construction cell (name in `use` constructs; also `except`/`ignore_only` accept it) |
| `internal/rules/selection_test.go` | 54-55 | `"unimplemented enabled/disabled"` cells expect `"not implemented"` | Rewrite as valid-selection cells or drop |
| `internal/rules/builder_test.go` | 69 | `require.NotContains(t, values, "PACKAGE_NO_IMPORT_CYCLE")` | `Contains` (mirrors lab `builder_test.go:69`) |
| `internal/config/v1/lint_selection_test.go` | 18, 32, 57 | fixture lists the name among 4 expected-invalid locations; schema data asserts `NotContains` | Swap the fixture name for another invalid name (keeps 4 issues), keep the loop, flip :57 to `Contains` (mirrors lab :57) |
| `mcp/easypconfig/describe_test.go` | 296-298 | `NotContains(t, values, "PACKAGE_NO_IMPORT_CYCLE")` | `Contains` (mirrors lab `describe_test.go:298`) |

### Pattern 4: Parity mechanism vs lab CLI on acceptance corpora (Output question 4 — ROADMAP design note resolved)

**The lab CLI run:**
1. Binary: `go build -o <dir>/easyp-lab ./cmd/easyp` inside `/tmp/gsd-research/compiler-lab/easyp-cli` (checkout `08cef10909c0722f7042ba17b5c11951f9c6b6f6` — verified present and **builds** this session, `BUILD_EXIT=0`, Go 1.27.3, module proxy reachable). If `/tmp` was wiped: `gh repo clone easyp-tech/compiler-lab` + checkout 08cef10 (SSH remote; Phase 1-5 precedent).
2. New binary: `go build -o <dir>/easyp-new ./cmd/easyp` in this repo.
3. Corpus run (either via the lab stand or directly): cwd = corpus project; `<bin> --format json lint`; stdout JSON lines carry `RuleName`, `Path`, `Position.Line/Column` (lab.py parses exactly these keys; `Issue` has no JSON tags — Go default capital keys, same lineage both sides). Exit codes 0/1 accepted; exit 2 = config/compile error path.

**Config discovery — no surgery needed:** every materialized corpus project carries `easyp.yaml` with `linters: default: STANDARD` (lab `bench/setup.sh` writes it verbatim). `Policy.LintConfig()` **always seeds MINIMAL** and STANDARD adds BASIC+DEFAULT — verbatim from `internal/config/v1/policy.go:104-116`:

```go
	use := []string{"MINIMAL"}
	switch p.Linters.Default {
	case "BASIC":
		use = append(use, "BASIC")
	case "STANDARD":
		use = append(use, "BASIC", "DEFAULT")
	case "COMMENTS":
		use = append(use, "BASIC", "DEFAULT", "COMMENTS")
	}
```

The lab's v1 code seeds MINIMAL identically (lab `internal/config/v1/policy.go:109`). Post-change, both binaries therefore run all three new rules on the corpora **with their committed configs unchanged**.

**What's compared:** per corpus, parse each JSON line, keep only `RuleName ∈ {PACKAGE_NO_IMPORT_CYCLE, SYNTAX_SPECIFIED, FIELD_NOT_REQUIRED}`, project to the tuple `(RuleName, Path, Position.Line, Position.Column)`, and `Counter`-compare lab vs new (identity definition inherited from LINT-PARITY-01: правило, файл, строка, колонка). Message equality is not part of the corpus identity — it is pinned by the carried unit tests (Messages ported verbatim). Raw-JSON-line comparison (what `lab.py lint-diff` does) is **not** suitable here: (a) it would include the ~41 ported rules, whose lab-vs-easyp divergences are *deliberate* main-wins decisions owned by the Phase-4 ledger (swift_prefix, directive filtering, ImportUsed…), and (b) field-set/serialization drift between lab and easyp CLIs could false-positive. Filter + project avoids both.

**Anti-vacuity requirement:** the acceptance corpora are proto3-heavy (googleapis lineage) — all files have syntax statements, `required` is essentially absent, and package import cycles are unlikely [ASSUMED — cannot be verified without the corpora]. The parity claim must therefore be backed by an **in-repo teeth corpus**: a temp project dir (or `internal/rules/testdata` fixtures driven through the real `core` entry) containing (a) a file with no syntax statement, (b) a proto2 file with a `required` field, (c) a two-package import cycle A→B→A, (d) clean negative files. Run **both binaries** on the teeth corpus with a v1 config enabling exactly the three rules; assert each rule fires ≥1 finding on both sides and the projected multisets are equal. This makes LINT-NEWRULES-03 verifiable regardless of corpora availability; the acceptance-corpus run then extends it to the letter of the requirement.

**Corpora availability:** `bench/projects/` is gitignored in compiler-lab; corpora are materialized by `bench/setup.sh` from `$GOOGLEAPIS` (default `$LAB/../EasyP-Parser-Compiler-Experiment/googleapis`) + corpus `.args` file lists. That source repo is **not present on disk** (`/tmp/gsd-research/` has no EasyP-Parser-Compiler-Experiment; a `googleapis` checkout exists at `/home/nil/DiskD/W/yadro/googleapis` but the `.args` file lists that define api-common/googleapis-nocompute membership live in the experiment repo). This is a `checkpoint:human-verify` item: clone EasyP-Parser-Compiler-Experiment (or point `GOOGLEAPIS`/`CORPUS_ARGS` at existing checkouts) before the corpora leg; the teeth corpus is the fallback that keeps the phase unblocked.

### Pattern 5: Schema/MCP atomic update chain (Output question 5)

The chain is already wired; the phase only feeds it. What exactly changes, per hop:

| Hop | File(s) | Change |
|-----|---------|--------|
| Name source | `internal/rules/builder.go` `addMinimal`/`addBasic` | +3 names (Pattern 2) — this is the **only** hand-written name change |
| Registry | `internal/rules/builder_linked.go` `New` | +3 constructions (same commit) |
| Validation | `internal/rules/selection.go` | special case deleted (Pattern 3) |
| JSON Schema enums | `internal/config/v1/schema.go:72-76` — `policy.Properties["linters"].Properties[field].Items.Enum = rules.AllLintUseValues()` (enable+disable) and `…exclude-rules…linters…Enum = rules.AllLintUseValues()` | **No code change** — enums re-derive at generation time |
| Generated artifacts | `schemas/easyp-v1.schema.json`, `schemas/easyp.schema.json` | Each gains the 3 names at its 3 enum spots (enable, disable, exclude-rules[].linters) — verified: exactly these two files carry `PACKAGE_DIRECTORY_MATCH` ×3 today; the other four schemas (easyp.gen*, protobuf.lock*) are untouched. Regenerated by `go run ./cmd/easyp schema-gen` (Taskfile: `task schema:generate`); verify `git diff --exit-code schemas/` after regen. **Never hand-edit** |
| MCP | `mcp/easypconfig/schema.go:12-14,34-37` — `MarshalConfigJSONSchema`/`SchemaByPathFor` call `v1.SchemaJSON(...)` | **No code change** — Describe/SchemaByPath enum values flow through `v1.SchemaJSON` → `documents()` → `AllLintUseValues()`. Only `describe_test.go:296-298` flips to `Contains`. `docs.go` field descriptions are generic ("Implemented lint rules or groups…") and stay accurate; the `issues.exclude-rules[].linters` wording "unknown or unimplemented names are rejected" (docs.go:38) becomes slightly stale ("unimplemented" set is empty) — optional one-word polish, lab carries the same wording |
| Directive validation | `internal/api/runtime.go:77` | No change — `KnownLintRules: rules.AllRuleNames()` auto-includes the new names |
| Migration | `internal/migration/convert.go:163` | No change — legacy configs naming the rule now validate (behavior change, intended) |

**Atomicity:** schemas derive from the registry, so a registry change without regen trips `task schema:check`; the MCP test pins the same set. One commit containing rules + registry + validator + test flips + regenerated `schemas/` satisfies "правила, схемы и MCP-описания меняются согласованно, одним изменением". The lab proves the target end state: lab `schemas/easyp*.schema.json` contain `PACKAGE_NO_IMPORT_CYCLE` at exactly the same 3×2 spots (verified in lab tree).

**Golden corpus impact:** `internal/rules/golden_full_registry_test.go:47` drives the golden corpus with `goldenCorpusUse = rules.AllRuleNames()` — the new rules will run over it. Verified this session: all 13 corpus files declare `syntax = "proto3"` (13 `syntax` statements), zero `required` fields, and the corpus's only import is `dep/v1/unused_dep.proto` (one direction — acyclic), so all three rules are expected to fire **zero** findings and the golden literals should survive byte-identical [ASSUMED — verify with a probe run of `TestLintGoldenFullRegistry` immediately after the registry change; if red, the corpus must be extended or re-captured **in the same commit**, never silently].

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Schema enum updates | Hand-editing `schemas/*.json` | `go run ./cmd/easyp schema-gen` + `task schema:check` | Generated artifacts; drift gate exists; AGENTS.md forbids hand edits |
| Rule-name derivation | Manual name constants | `core.GetRuleName` reflection (`dom.go:107-123`) | Single naming law for registry, findings, config |
| Import-path keying | String trimming ad hoc | `core.ConvertImportPath` (`dom.go:153-155`) | The engine's `ImportedFiles` keys are `ImportPath` |
| Deterministic map iteration in Finish | Ad-hoc sort | Lab's `sortedKeys` helper, ported verbatim | Lab-identical iteration order; engine re-sorts anyway |
| Suppression/ignore handling in rules | Per-rule suppression logic | Engine machinery (`collectFile` skips ignored files; drain path filter; comment suppressions) | Rules never self-suppress (standing constraint (b) since Phase 3/4) |
| Corpora snapshot/diff | New stand | `lab.py lint --bin X --label Y` + `lint-diff`, or a phase script reusing its JSON-line shape | The stand exists, is proven, and Phase 7 reuses it |
| Cross-file collection plumbing | New module-rule machinery | Existing `core.ModuleRule` + `collectFile` + Finish drain (`lint_linked.go:137-153, 213-228`) | Proven by the D-03 ten since Phase 4 |

**Key insight:** this phase is mostly *wiring into loops that already close*. The only genuinely new logic is the ~140 lines of rule bodies — and those are copied, not designed.

## Common Pitfalls

### Pitfall 1: Go import cycle in the rules→core direction
**What goes wrong:** `PackageNoImportCycle` needs `core.ModuleRule`, `core.AppendIssue`, `core.ConvertImportPath`; tests need `core` and `protoast`. `internal/core` cannot import `internal/rules` (package-core precedent since 03-01; `core.New` takes `[]Rule` values, never rule constructors).
**Why it happens:** wanting the validator to double-check rule implementations.
**How to avoid:** rules import `core` + `protoast` only; `TestModuleRuleSet` lives in `rules_test` package and probes the constructed registry (`rules.New(config.LintConfig{Use: rules.AllRuleNames()})`), never the reverse.
**Warning signs:** an import of `internal/rules` from `internal/core`, or a new interface declared in `core` to "help" one rule.

### Pitfall 2: Schema drift ordering
**What goes wrong:** registry changed but `schemas/` regenerated in a later commit (or forgotten) → `task schema:check` red; or schemas regenerated before the registry change → transient drift.
**Why it happens:** schemas are build artifacts of `AllLintUseValues`, one hop removed from the edit.
**How to avoid:** regen inside the same atomic commit, immediately after the registry edit; the commit's verify chain is `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` (Phase-4/5 precedent, D7 gate).
**Warning signs:** `schema:check` red in CI; `git status` showing `schemas/` dirty.

### Pitfall 3: Group membership vs validation ordering
**What goes wrong:** `add*` (name lists) and `New` (type list) updated asymmetrically — e.g. name added to `addMinimal` but type forgotten in `New` → `New(Use: [MINIMAL])` fails construction ("invalid rule: PACKAGE_NO_IMPORT_CYCLE") despite validation passing; or the special case deleted before the type exists → validation passes, construction fails with a *different* error text.
**Why it happens:** two parallel lists in two files with no compile-time link.
**How to avoid:** one commit; order of edits: rule files → `New` list → `add*` lists → special-case deletion → test flips → regen. `TestModuleRuleSet` (constructs `New(AllRuleNames())`) catches membership asymmetry immediately.
**Warning signs:** any `ErrInvalidRule` from a `New` call in tests; `TestModuleRuleSet` red.

### Pitfall 4: Guard tests red-before-green must be flipped in the same commit
**What goes wrong:** five test pins (Pattern 3 table) assert the *current* behavior; landing the registry change without them makes the suite red; landing the flips first makes them red.
**Why it happens:** the pins were deliberately written as Phase-6 tripwires.
**How to avoid:** treat the five flips as part of the atomic change (TDD: the flipped pins are the phase's RED evidence — e.g. `builder_test.go:69` Contains fails pre-registry).
**Warning signs:** `go test ./internal/rules/ ./internal/config/v1/ ./mcp/...` red on any of the five files.

### Pitfall 5: Vacuous corpus parity
**What goes wrong:** acceptance corpora produce zero findings for one or more of the new rules → "identical output" passes trivially and proves nothing.
**Why it happens:** googleapis-lineage corpora are proto3, syntaxed, acyclic [ASSUMED].
**How to avoid:** the teeth corpus (Pattern 4) with per-rule ≥1-finding assertions on both binaries; report per-rule counts in the parity evidence.
**Warning signs:** parity report with `issues 0` for a rule.

### Pitfall 6: Full-output diff poisoned by deliberate main-wins divergences
**What goes wrong:** comparing *all* findings lab-vs-easyp on corpora fails on ported rules (swift_prefix compares `ruby_package` by main's pinned bug; directive-only comments document per main; ImportUsed is syntax-walk per main) — none of which this phase may touch.
**Why it happens:** the lab carries the pre-parity semantics; the Phase-4 deferred-decisions ledger (04-07-SUMMARY items 1-13) froze main's variants.
**How to avoid:** filter the comparison to the three new rule names (Pattern 4); full-output identity is Phase 7's LINT-PARITY-01 concern *against main*, not the lab.
**Warning signs:** diff entries with `RuleName` outside the three.

### Pitfall 7: MINIMAL auto-seed makes the change user-visible everywhere
**What goes wrong:** every v1 policy seeds `use` with `MINIMAL` (`policy.go:104`) — post-change every `easyp lint` gains the cross-file cycle rule even with `default: MINIMAL`-only configs; Phase 7's main-vs-new lint-diff would then show extra findings attributed to "new rules" unless filtered (main lacks them).
**Why it happens:** that's the buf-style preset semantics; the lab behaves identically.
**How to avoid:** document the behavior change in the plan (compat framing: new rules are additive, format unchanged — LINT-COMPAT-01 intact); note for Phase 7: filter the three names out of the main-vs-new compare (or pin a main-vs-main baseline that predates them).
**Warning signs:** Phase-7 diff showing `PACKAGE_NO_IMPORT_CYCLE`/`SYNTAX_SPECIFIED`/`FIELD_NOT_REQUIRED` entries.

### Pitfall 8: Golden corpus silently extended by the new rules
**What goes wrong:** `TestLintGoldenFullRegistry` runs `AllRuleNames()` over the golden corpus; any new-rule finding shifts the byte-exact golden.
**How to avoid:** verified this session that the corpus is clean for all three rules (13/13 syntax statements, zero `required`, single acyclic import); still probe-run the golden test right after the registry change and extend/re-capture in the same commit if red [ASSUMED-clean → verify].
**Warning signs:** golden test red with new-rule findings on `pkgless/orphan.proto` (the only packageless file) or any fixture.

### Pitfall 9: Unbounded DFS recursion in the cycle rule
**What goes wrong:** `PackageNoImportCycle.cycle` is a recursive DFS over the package graph; a deep import chain (thousands of packages) recurses proportionally — pathological corpora could stack-overflow the CLI.
**Why it happens:** the lab accepted the recursion; googleapis is ~20 deep.
**How to avoid:** port verbatim (parity first); note the bound as a known characteristic — same exposure as the lab CLI, so parity holds even on adversarial input.
**Warning signs:** stack overflow on huge corpora (would equally affect the lab binary — a parity-preserving non-issue).

### Pitfall 10: /tmp volatility of the etalon checkout
**What goes wrong:** `/tmp/gsd-research/compiler-lab` is outside git; a machine restart drops the porting source and parity oracle mid-phase.
**How to avoid:** plan Task 0 re-verifies `git -C /tmp/gsd-research/compiler-lab rev-parse HEAD == 08cef10909c0722f7042ba17b5c11951f9c6b6f6`; fallback `gh repo clone easyp-tech/compiler-lab && git checkout 08cef10` (recorded in the task context).
**Warning signs:** missing directory or divergent SHA.

## Code Examples

### The registry insertion points (lab-verbatim order)

```go
// Source: internal/rules/builder_linked.go:26-34 (easyp) after change;
// order per lab builder.go:223-231
allRules := []core.Rule{
	//	minGroup
	&DirectorySamePackage{},
	&PackageDefined{},
	&PackageDirectoryMatch{
		Root:   ".",
		Prefix: cfg.PackageDirectoryPrefix,
	},
	&PackageSameDirectory{},
	&PackageNoImportCycle{},          // NEW — 5th in MINIMAL, lab placement

	//	basicGroup
	&EnumFirstValueZero{},
	// … 18 more BASIC entries unchanged …
	&RPCPascalCase{},
	&ServicePascalCase{},
	&SyntaxSpecified{},               // NEW — 21st in BASIC
	&FieldNotRequired{},              // NEW — 22nd in BASIC
	// …
```

```go
// Source: lab builder.go:228,254-255 — the name lists mirror via reflection
res = append(res, core.GetRuleName(&PackageNoImportCycle{}))   // addMinimal
res = append(res, core.GetRuleName(&SyntaxSpecified{}))        // addBasic
res = append(res, core.GetRuleName(&FieldNotRequired{}))
```

### The special-case deletion

```go
// Source: internal/rules/selection.go:14-27 — AFTER: the loop keeps only the
// generic Contains path; PACKAGE_NO_IMPORT_CYCLE validates through it.
func ValidateNames(names []string) error {
	allowed := AllLintUseValues()
	var issues []error
	for _, name := range names {
		if !slices.Contains(allowed, name) {
			issues = append(issues, fmt.Errorf("%w: %s", core.ErrInvalidRule, name))
		}
	}
	return errors.Join(issues...)
}
```

### ModuleRule shape the port must satisfy

```go
// Source: internal/core/lint_model.go:35-41 (verbatim)
ModuleRule interface {
	Rule
	// Collect hands one file to the rule, in walk order.
	Collect(ProtoInfo) error
	// Finish reports the cross-file findings, once after the last file.
	Finish() []IssueInfo
}
```

The engine contract around it (verified, no engine changes needed): `collectFile` hands files in walk order and **skips files ignored for the rule** (`lint_linked.go:213-228`); the Finish drain honors `issue.Path` and applies comment suppressions (`lint_linked.go:137-153`); the D-01 total sort (path, rule, line, col, source) orders the final output (`lint_linked.go:156-166`).

### Unit-test harness for the new rules

```go
// Source: internal/rules/init_test.go:32-60 — existing harness
info := parseLinkedFile(t, "./../../testdata/import_cycle/a.proto") // resolves direct imports from testdata
rule := &PackageNoImportCycle{}
require.NoError(t, rule.Collect(infoA))
require.NoError(t, rule.Collect(infoB))
issues := rule.Finish()
// pin: every edge of the cycle reported, SourceName == "a.v1 -> b.v1 -> a.v1",
// Position == the import keyword position, Path == the importing file
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Old go-protoparser engine + `NewLinked` suffix registry | Single engine, canonical names, `rules.New` THE registry/validator | Phase 5 flip (4238fbe) + consolidation (263f067) | This phase wires into the consolidated surface — no `_Linked` suffixes, no dual contracts |
| `PACKAGE_NO_IMPORT_CYCLE` = validation exception | Implemented ModuleRule | Phase 6 | Migration/directive/schema surfaces accept the name |
| 41-rule registry | 44-rule registry | Phase 6 | "43" in REQUIREMENTS/phase titles remains the documented discrepancy (`lint_model.go:28-29`) |

**Deprecated/outdated:** the Phase-4 parity harness (`assertBatchParity`, old-side oracle) died at the flip — corpus comparisons are now external (lab CLI / lab.py), not engine-vs-engine.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Acceptance corpora yield zero (or near-zero) new-rule findings — proto3, syntaxed, acyclic, no `required` | Pattern 4 / Pitfall 5 | Parity leg vacuous; mitigated by the mandatory teeth corpus |
| A2 | Golden corpus stays clean under the extended registry | Pitfall 8 | `TestLintGoldenFullRegistry` red → corpus extension/re-capture in the same commit |
| A3 | `EasyP-Parser-Compiler-Experiment` is privately cloneable / the user can supply corpora | Pattern 4 / Environment | Corpora leg blocked → teeth-corpus fallback + Phase-7 stand inherits the same gap (STATE.md blocker already flags this) |
| A4 | Lab and easyp CLIs serialize identical JSON field sets for equal findings (no tags, same lineage) | Pattern 4 | Raw-line compare unusable → already mitigated: comparison projects to (rule,file,line,col) |
| A5 | `toUpperSnakeCase` yields the expected three names | Pattern 2 | Wrong registry names — trivially probe-verified at plan time via `core.GetRuleName(&PackageNoImportCycle{})` |

## Open Questions

1. **Corpora materialization path**
   - What we know: `bench/setup.sh` needs `$GOOGLEAPIS` + `CORPUS_ARGS` from `EasyP-Parser-Compiler-Experiment` (absent); lab remote is SSH (`git@github.com:easyp-tech/compiler-lab.git`); a googleapis checkout exists at `/home/nil/DiskD/W/yadro/googleapis` but the corpus `.args` file lists do not.
   - What's unclear: whether the user can clone/provide the experiment repo for this phase, or the corpora leg defers to Phase 7's stand spike.
   - Recommendation: `checkpoint:human-verify` in the plan; teeth corpus carries LINT-NEWRULES-03 meanwhile.

2. **Where the parity compare script lives**
   - What we know: lab.py must stay untouched (Phase-7 asset); a ~30-line Python/Go script suffices for filter+project+Counter compare.
   - What's unclear: phase script in the repo vs throwaway under `/tmp`.
   - Recommendation: throwaway script committed as plan evidence (its output, not the script, is the artifact), consistent with prior phases' probe style.

3. **`docs.go` "unimplemented" wording** (minor)
   - What we know: `mcp/easypconfig/docs.go:38` says "unknown or unimplemented names are rejected"; after Phase 6 nothing is "unimplemented".
   - Recommendation: optional same-commit wording polish; lab carries the identical wording, so parity does not force it.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | everything | ✓ | 1.27.3 | — |
| compiler-lab checkout @ 08cef10 | rule sources, parity oracle, lab builder order | ✓ (`/tmp/gsd-research/compiler-lab`) | SHA verified | `gh repo clone easyp-tech/compiler-lab` + checkout 08cef10 |
| Lab CLI build | parity leg | ✓ (built this session, exit 0) | — | — |
| `EasyP-Parser-Compiler-Experiment` | corpora materialization (`setup.sh`) | ✗ not on disk | — | teeth corpus; corpora leg gated on user |
| `task` runner | `task schema:generate` / `task schema:check` / `task test` | ✗ not on PATH | — | direct equivalents: `go run ./cmd/easyp schema-gen`, `git diff --exit-code schemas/`, `go test -race -count=1 ./...` |
| `bin/golangci-lint` | lint gate | ✓ | repo-pinned | — |
| testify | tests | ✓ (existing dep) | — | — |
| buf / protoc-gen-go | not needed for lint parity | — | — | only lab.py bench/parity (FDSet) commands need them |

**Missing dependencies with no fallback:** none blocking — the corpora leg is the only gap and has the teeth-corpus fallback.
**Missing dependencies with fallback:** `task` (direct go commands), `EasyP-Parser-Compiler-Experiment` (teeth corpus + checkpoint).

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | go test + testify (existing); `gotestsum` via `task test` (task absent → `go test -race -count=1 ./...`) |
| Config file | none needed (Go convention) |
| Quick run command | `go test -race -count=1 ./internal/rules/ -run 'TestModuleRuleSet|TestAllLintUseValues|TestLintUseValidationParity'` |
| Full suite command | `go test -race -count=1 ./...` (wazero race carve-out: `internal/adapters/plugin` + `internal/generation`, pre-existing, pre-Phase-7) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LINT-NEWRULES-01 | MINIMAL contains the cycle rule as ModuleRule; exactly 11 ModuleRules | unit | `go test -race -count=1 ./internal/rules/ -run 'TestModuleRuleSet|TestAllGroups'` | ✅ (module_rules_guard_test.go — extend map to 11) |
| LINT-NEWRULES-01 | `Use: [MINIMAL]` constructs 5 rules incl. the cycle rule | unit | `go test -race -count=1 ./internal/rules/ -run TestLintUseValidationParity` | ✅ (validation_parity_test.go — flip cell) |
| LINT-NEWRULES-02 | BASIC contains SyntaxSpecified + FieldNotRequired; construction works | unit | `go test -race -count=1 ./internal/rules/ -run 'TestAllLintUseValues'` + new rule tests | ✅ builder_test.go (flip) / ❌ new `<rule>_test.go` files (this phase) |
| LINT-NEWRULES-02/03 | Per-rule finding semantics (messages, positions, sources) | unit | `go test -race -count=1 ./internal/rules/ -run 'TestSyntaxSpecified|TestFieldNotRequired|TestPackageNoImportCycle'` | ❌ Wave 0 (authored fresh; lab has none) |
| LINT-NEWRULES-03 | New-rule findings identical to lab CLI | integration (external) | phase parity script over teeth corpus (both binaries) + corpora when available | ❌ Wave 0 (script + evidence) |
| LINT-NEWRULES-04 | Schema enums include the names; drift-free | unit + gate | `go test ./internal/config/v1/ -run TestLintSelection` && `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` | ✅ (lint_selection_test.go — flip) |
| LINT-NEWRULES-04 | MCP Describe enums include the names | unit | `go test ./mcp/easypconfig/ -run TestDescribeLintSelection` | ✅ (describe_test.go — flip) |
| — (regression) | Golden corpus byte-identical under extended registry | unit | `go test -race -count=1 ./internal/rules/ -run TestLintGoldenFullRegistry` | ✅ (probe: expect green; if red see Pitfall 8) |

### Sampling Rate

- **Per task commit:** quick rules+v1+mcp run + `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/`
- **Per wave merge:** `go test -race -count=1 ./internal/...` + golangci on touched packages
- **Phase gate:** full `-race` suite (with wazero carve-out), `schema:check` equivalent, parity evidence (teeth corpus mandatory; corpora leg per checkpoint)

### Wave 0 Gaps

- [ ] `internal/rules/syntax_specified_test.go`, `field_not_required_test.go`, `package_no_import_cycle_test.go` — fresh unit tests (REQ-01/02/03 semantics)
- [ ] cycle fixture files under `internal/rules/testdata/import_cycle/` (a.proto ↔ b.proto + same-package negative)
- [ ] parity script + teeth corpus harness for LINT-NEWRULES-03 (external-binary comparison)
- [ ] none for framework/config — existing infrastructure covers all else

## Security Domain

### Applicable ASVS Categories (Level 1)

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | CLI tool, no auth surface |
| V3 Session Management | no | no sessions |
| V4 Access Control | no | no multi-tenant surface |
| V5 Input Validation | yes | rule/group names validated through `ValidateNames` (single surface, D-06); config YAML strict-decoded with JSON Schema enums; the phase *removes* an allow-list bypass special case — validation stays deny-by-default via `slices.Contains(allowed, name)` |
| V6 Cryptography | no | no crypto |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Config injection of arbitrary rule names | Tampering | deny-by-default `ValidateNames` → `ErrInvalidRule`; schemas constrain editors |
| Recursive DFS on adversarial deep package graph (cycle rule) | DoS | verbatim lab parity bound; identical exposure in the etalon (Pitfall 9) |
| Generated schema tampering | Tampering | `task schema:check` drift gate; generated-only edits via schemagen |

No new attack surface: the phase adds findings to an existing CLI output channel and enum values to existing config surfaces.

## Sources

### Primary (HIGH confidence — all read this session)

- In-repo: `internal/core/lint_model.go` (Rule/ModuleRule/ProtoInfo/AppendIssue/GetRuleName, D-03 comment), `internal/core/lint_linked.go` (checkFile/collectFile/drain/sort), `internal/core/dom.go` (Position alias, IssueInfo, ConvertImportPath, toUpperSnakeCase)
- In-repo: `internal/rules/builder.go`, `builder_linked.go`, `selection.go` + the five pinned test files (`module_rules_guard_test.go`, `validation_parity_test.go`, `selection_test.go`, `builder_test.go`, `internal/config/v1/lint_selection_test.go`, `mcp/easypconfig/describe_test.go`)
- In-repo: `internal/protoast/protoast.go`, `walk.go` (all model surfaces the rules touch)
- In-repo: `internal/config/v1/schema.go` + `policy.go` (enum wiring, MINIMAL seed), `mcp/easypconfig/schema.go` + `docs.go`, `internal/api/runtime.go` + `lint_v1.go`, `internal/migration/convert.go`, `internal/rules/golden_full_registry_test.go`, `internal/rules/init_test.go`
- Lab etalon: `/tmp/gsd-research/compiler-lab/easyp-cli` @ 08cef10 — the three rule files, `builder.go` add* order, `builder_test.go`, `mcp/easypconfig/describe_test.go`, `internal/config/v1/policy.go`, `schemas/`, `bench/lab.py`, `bench/setup.sh`

### Secondary (MEDIUM confidence)

- None needed — no external documentation was consulted; the domain is fully code-internal. The research-plan seam (Context7/web providers) was evaluated and skipped as inapplicable: there are no external libraries, and every claim's authoritative source is the repo or the lab etalon, both read directly.

### Tertiary (LOW confidence)

- None.

## Metadata

**Confidence breakdown:**
- Rule ports / registry wiring: HIGH — every symbol verified at source with line ranges; verbatim port surface enumerated
- Special-case removal: HIGH — exact location, exact pins, exact flip table
- Parity mechanism: HIGH on mechanism design (lab.py shape verified, config seeding verified), MEDIUM on corpora-specific outcomes (corpora absent — A1/A3 assumptions logged)
- Schema/MCP chain: HIGH — the derivation loop is closed code, verified end to end; lab's end-state schemas confirm

**Research date:** 2026-11-21
**Valid until:** 2026-12-21 (stable — in-repo domain; re-verify only the lab checkout SHA and corpora availability at plan time)
