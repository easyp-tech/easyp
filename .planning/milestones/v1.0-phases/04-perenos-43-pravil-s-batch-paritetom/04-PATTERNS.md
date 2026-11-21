# Phase 4: Перенос 43 правил с батч-паритетом - Pattern Map

**Mapped:** 2026-11-19
**Files analyzed:** 94 (41 new rule impls + 41 rewritten test files + 12 infra/wiring files)
**Analogs found:** 94/94 (all roles have repo or etalon analogs; 5 items are net-new shapes with partial analogs — see No Analog Found)

## Porting-source warning (read first)

Two distinct source kinds appear below. The planner must never treat them the same way:

- **Repo analogs** — git-tracked files under `internal/` (all verified with `git ls-files`). These are patterns to follow *in place*.
- **Etalon porting sources** — `/tmp/gsd-research/compiler-lab/easyp-cli/internal/rules/*.go` @ `08cef10`. These are **external files to port FROM, never repo analogs**. They are not tracked in this repo and must not be referenced as pattern sources after the port; their content lands in new repo files.
- Etalon name mapping (the etalon is the post-flip tree, its names differ): etalon `core.Rule` → main's `core.LintRule`; etalon `core.ProtoInfo` → main's `core.LinkedFileInfo`; etalon 4-arg `core.AppendIssue(issues, rule, pos, sourceName)` → the new linked constructor (Shared Pattern 3); etalon `core.Position` ≡ main `core.Position` (alias of `protoast.Position`, `internal/core/dom.go:20`).

## File Classification

### Infrastructure files

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/core/lint_model.go` (MODIFY: add linked issue constructor) | model/utility | transform | `internal/core/dom.go:126-164` (AppendIssue/buildError) | exact (rename-map variant) |
| `internal/rules/builder_linked.go` (NEW) | registry/config | CRUD | `internal/rules/builder.go:61-143` (`New`) | exact |
| `internal/rules/lint_helpers.go` (MODIFY: re-align `mismatches` to new contract + main first-exempt variant) | utility | transform | itself (lines 104-152) + etalon `lint_helpers.go` (external) | exact |
| `internal/rules/init_test.go` (REWRITE: protoast parse + `validate()` driver) | test helper | file-I/O + transform | etalon `init_test.go` (external); repo `init_test.go` keeps fixture constants | exact |
| `internal/rules/parity_<batch>_test.go` ×7 (NEW) | test | request-response (both engines on one tree) | `internal/core/lint_suppressions_parity_test.go` + `internal/rules/validation_parity_test.go` | role-match |
| position-probe test (Wave 0, placement planner's discretion) | test | transform | `internal/core/position_parity_test.go` | role-match |
| `TestModuleRuleSet` guard + readSource-once cell (Wave 0) | test | event-driven | `internal/core/lint_linked_test.go` cells | role-match |
| `.golangci.yml` (MODIFY: drop line 14 `- internal/rules/lint_helpers.go` exclusion — finishing batch) | config | — | itself | exact |
| `internal/core/core.go` (possibly untouched — `Options.LinkedRules` already exists at lines 68-76; `buildCore` wiring is planner discretion) | provider | — | — | n/a |

### Rule files (41 impls + 41 tests)

Role for all: **rule** (lint predicate over parsed trees; pure in-memory transform). Data flow: **transform** (per-file: `LinkedFileInfo` → `[]core.Issue`) for 31 rules; **event-driven** (engine-driven Collect/Finish lifecycle) for the 10 ModuleRules. The 10 ModuleRules (D-03 list, pinned at `internal/core/lint_model.go:50-64`): #1, #4, #16-22, #28.

Naming/coexistence (RESEARCH Option A, planner ratifies): new types are suffixed (`<Name>Linked`) in NEW files (`<rule>_linked.go` or folded per planner) — old `<rule>.go` files are **byte-frozen until Phase 5**; each new type overrides its registry name via `LintRuleName() string` (see Shared Pattern 1). Existing `<rule>_test.go` files are rewritten per port to drive the new type while carrying main's expectations verbatim.

| Batch | Rules (impl file = new `<rule>_linked.go`) | Contract | Analog pair (repo predicate oracle / etalon walk source) |
|-------|--------------------------------------------|----------|----------------------------------------------------------|
| 0 | #3 PackageDirectoryMatch (pilot, re-verified in batch 1) | LintRule | `internal/rules/package_directory_match.go` / etalon `package_directory_match.go` — verbatim-from-lab, only model access differs |
| 1 | #1 DirectorySamePackage, #2 PackageDefined, #3 (re-verify), #4 PackageSameDirectory | #1, #4: ModuleRule; #2, #3: LintRule | #1: `internal/rules/directory_same_package.go` + etalon; #2: `internal/rules/package_defined.go` + etalon (zero-position divergence); #4: `internal/rules/package_same_directory.go` + etalon |
| 2 | #7 EnumPascalCase, #8 EnumValueUpperSnakeCase, #9 FieldLowerSnakeCase (file `message_field_lower_snake_case.go`), #13 MessagePascalCase, #14 OneofLowerSnakeCase, #15 PackageLowerSnakeCase, #23 RPCPascalCase, #24 ServicePascalCase, plus verbatim #10 ImportNoPublic / #11 ImportNoWeak (exact split planner's discretion) | LintRule | repo `<rule>.go` (regexp + keyword pos) / etalon `<rule>.go` (Visitor walk); #10/#11: etalon verbatim (`imp.Public`/`imp.Weak`) |
| 3 | #5 EnumFirstValueZero, #6 EnumNoAllowAlias, #12 ImportUsed, #16-22 PackageSame{CsharpNamespace,GoPackage,JavaMultipleFiles,JavaPackage,PHPNamespace,RubyPackage,SwiftPrefix} | #16-22: ModuleRule; rest: LintRule | #5: `internal/rules/enum_first_value_zero.go` + etalon (string compare + empty-enum guard); #6: `internal/rules/enum_no_allow_alias.go` + etalon; #12: `internal/rules/import_used.go` + etalon (stateless `importUsage`); #16-22: `internal/rules/package_same_go_package.go` + etalon (uniform shape ×7) |
| 4 | #25 EnumValuePrefix, #26 EnumZeroValueSuffix, #27 FileLowerSnakeCase, #28 RPCRequestResponseUnique, #29 RPCRequestStandardName, #30 RPCResponseStandardName, #31 PackageVersionSuffix, #32 ServiceSuffix | #28: ModuleRule; rest: LintRule | #25: `internal/rules/enum_value_prefix.go` + etalon (carry `pascalToUpperSnake`); #26: `internal/rules/enum_zero_value_suffix.go` + etalon (keep ENV tag); #27: `internal/rules/file_lower_snake_case.go` + etalon (0:0 position); #28: `internal/rules/rpc_request_response_unique.go` + etalon; #29/#30: `internal/rules/rpc_{request,response}_standard_name.go` + etalon; #31: `internal/rules/package_version_suffix.go` + etalon (keep regexp, not `packageHasVersionSuffix`); #32: etalon verbatim + `internal/rules/service_suffix.go` position |
| 5 | #33 CommentEnum, #34 CommentEnumValue, #35 CommentField (file `comment_field_message.go`), #36 CommentMessage, #37 CommentOneof (file `comment_one_of.go`), #38 CommentRPC, #39 CommentService | LintRule | `internal/rules/comment_enum.go` + etalon `comment_enum.go` — one uniform presence-predicate port ×7 |
| 6 | #40 RPCNoClientStreaming, #41 RPCNoServerStreaming + wiring finish | LintRule | etalon `rpc_no_client_streaming.go` / `rpc_no_server_streaming.go` verbatim (`Request.IsStream`→`Request.Stream`); wiring analog `internal/rules/builder.go` |

Every rule's colocated test rewrite (`<rule>_test.go`) has a dual analog: **main's current colocated test is the expectations oracle** (carry cases verbatim, including positions), and the etalon's colocated test shows the protoast-era structure. Exemplar pair: `internal/rules/message_pascal_case_test.go` vs etalon `message_pascal_case_test.go`.

## Pattern Assignments

### Batch 0 — Wave 0 infra + pilot

**Files:** `internal/core/lint_model.go` (+constructor), `internal/rules/builder_linked.go`, `internal/rules/init_test.go` (rewrite), `internal/rules/parity_wave0_test.go`, position probe, `TestModuleRuleSet` guard, disk-instrumentation cell, pilot rule #3.

**Linked issue constructor** — analog `internal/core/dom.go:126-164`, plus etalon `internal/core/dom.go:80-82` (external, the linked shape):
```go
// internal/core/dom.go (main, old contract) lines 126-130, 156-164:
func AppendIssue(
	issues []Issue, lintRule Rule, pos meta.Position, sourceName string, comments []*parser.Comment,
) []Issue {
	return append(issues, buildError(lintRule, fromMetaPos(pos), sourceName))
}
func buildError(lintRule Rule, pos Position, sourceName string) Issue {
	return Issue{Position: pos, SourceName: sourceName, Message: lintRule.Message(), RuleName: GetRuleName(lintRule)}
}
```
The new linked variant mirrors the etalon's 4-arg shape — `AppendLinkedIssue(issues []Issue, rule LintRule, pos Position, sourceName string) []Issue` — reusing `buildError`'s field assembly. Because `core.Position = protoast.Position` (dom.go:20) there is **no conversion**: etalon rules' `core.AppendIssue(res, c, message.NamePos, message.Name)` maps mechanically. Update the rename-map comment at `internal/core/lint_model.go:6-27`. Exact constructor name is planner's discretion (rename-map tracked).

**`builder_linked.go`** — copy `internal/rules/builder.go:61-143` structure exactly: validate via `validateConfigRuleNames` (D-06: one validator), same 41-rule list in the same group order, `unwrapLintGroups`/`removeExcept`/`unwrapIgnoreOnly` reused as-is; the element type becomes `[]core.LintRule` and names come from `core.LinkedRuleName` (or the explicit override, Shared Pattern 1). Signature mirrors `New`: `NewLinked(cfg config.LintConfig) ([]core.LintRule, map[string][]string, error)`.

**`init_test.go` rewrite** — port etalon `init_test.go` (external) wholesale: same fixture constants (repo file's lines 15-30 carry over), `protoast.Parse`-based `parseFile` (etalon lines 56-88), `validate()` Collect/Finish driver (etalon lines 92-115), and `parseProto` for focused cases (etalon lines 118-123). Keep the repo's `start(t)` signature shape (`internal/rules/init_test.go:32-55`) so unported rules' tests still compile during the transition.

**Parity harness skeleton** — analog `internal/core/lint_suppressions_parity_test.go:63-69` (`writeCorpusInto`) and its oracle-mirror doubles pattern (lines 16-42); in-package style per `internal/rules/validation_parity_test.go:26-110`. Comparison is the RESEARCH sketch: sorted (RuleName, Path, Position.Line, Position.Column) multisets via `require.Equal` — old-engine output order is nondeterministic, sort both sides.

**Position probe** — analog `internal/core/position_parity_test.go`; one fixture per decl kind through old (`Meta.Pos`) vs new (`Decl.Pos`) model, locking RESEARCH's mapping table (Shared Pattern 4) into a comment.

**Pilot rule #3 PackageDirectoryMatch** — verbatim-from-lab. Etalon `package_directory_match.go` lines 39-53 vs repo `package_directory_match.go` lines 39-53 — the entire diff is `protoInfo.Info.ProtoBody.Packages` → `protoInfo.File.Packages`, `pkgInfo.Meta.Pos` → `pkgInfo.Pos`, and the 5-arg→4-arg AppendIssue. Keep main's struct tags verbatim (`Root` env/json/yaml tags, lines 15-16).

### Batch 1 — MINIMAL (4)

**`<Name>Linked` skeleton (uniform for all 41 ports)** — exemplar: pilot's shape applied to a main-diverged rule. Excerpt is RESEARCH Pattern 1, grounded in the two real files:
```go
// Walk mechanics: etalon message_pascal_case.go lines 21-25
protoInfo.File.Walk(protoast.Visitor{Message: func(message *protoast.Message) {
    if message.Name != toPascalCase(message.Name) { ... }
}})
// Predicate + position (main wins): internal/rules/message_pascal_case.go lines 23-27
pascalCase := regexp.MustCompile("^[A-Z][a-zA-Z0-9]+(?:[A-Z][a-zA-Z0-9]+)*$")
for _, message := range protoInfo.Info.ProtoBody.Messages {
    if !pascalCase.MatchString(message.MessageName) {
        res = core.AppendIssue(res, c, message.Meta.Pos, message.MessageName, message.Comments)
    }
}
```
Ported = walk shape of the first + regexp/position of the second: `if !pascalCase.MatchString(message.Name) { res = core.AppendLinkedIssue(res, m, message.Pos, message.Name) }` — `message.Pos` (Decl.Pos = first token), **not** `NamePos`.

**#1 DirectorySamePackage (ModuleRule, main-diverged+adapt)** — repo `internal/rules/directory_same_package.go:29-45` holds main's predicate (dir-keyed cache, `if d.cache[directory] == "" { store; continue }` first-exempt, issue at `pack.Meta.Pos` only on later differing files). Etalon `directory_same_package.go:24-31` holds the ModuleRule shape (`Collect` via `packagePosition` + `groupedValues.add`, `Finish` → `mismatches`). **Do NOT reuse `groupedValues.mismatches` as-is** (reports everyone) — Finish needs the first-exempt variant (Shared Pattern 5).

**#2 PackageDefined** — repo `internal/rules/package_defined.go` emits the zero position (`core.Position{Filename: path}` → 0:0, pinned by main's test); etalon emits `protoast.FilePosition` (1:1). Construct the zero literal directly; no protoast helper exists.

**#4 PackageSameDirectory** — same pair as #1: repo `package_same_directory.go` (package-keyed first-exempt cache) + etalon ModuleRule shape.

### Batch 2 — BASIC casing (8 regexp rules + 2 verbatim imports)

Uniform shape: etalon's `Visitor` walk + main's regexp + `Decl.Pos`. Per-rule regexps and position pins (repo file line numbers of the regexp to copy):

| Rule | Regexp source (repo) | Position pin |
|------|---------------------|--------------|
| #7 EnumPascalCase | `enum_pascal_case.go` `^[A-Z][a-z]+(?:[A-Z][a-z]+)*$` | `enum.Pos` (Meta.Pos) |
| #8 EnumValueUpperSnakeCase | `enum_value_upper_snake_case.go` `^[A-Z0-9]+(_[A-Z0-9]+)*$` | `value.Pos` |
| #9 FieldLowerSnakeCase | `message_field_lower_snake_case.go` `^[a-z0-9]+(_[a-z0-9]+)*$` | `field.Pos` (main col 3) |
| #13 MessagePascalCase | `message_pascal_case.go` (excerpt above) | `message.Pos` (main col 1) |
| #14 OneofLowerSnakeCase | `oneof_lower_snake_case.go` `^[a-z]+(_[a-z]+)*$` | `oneof.Pos` |
| #15 PackageLowerSnakeCase | `package_lower_snake_case.go` `^[a-z]+([_|[.][a-z0-9]+)*$` | `pkg.Pos` |
| #23 RPCPascalCase | `rpc_pascal_case.go` `^[A-Z][a-zA-Z0-9]*$` | `rpc.Pos` (main col 3) |
| #24 ServicePascalCase | `service_pascal_case.go` `^[A-Z][a-z]+([A-Z]|[a-z]+)*$` | `service.Pos` (main col 1) |

Walk callback per kind: `Visitor{Enum:...}`, `Visitor{EnumValue: func(e *protoast.Enum, v *protoast.EnumValue)}`, `Visitor{Field:...}`, `Visitor{Message:...}`, `Visitor{Oneof:...}`, `Visitor{RPC: func(s *protoast.Service, r *protoast.RPC)}` — see `internal/protoast/walk.go:7-15`. **Never** replace a regexp with `casing.go` helpers (`toPascalCase`/`toUpperSnakeCase`/`toLowerSnakeCase`) — they disagree on acronyms, digits, underscores (Pitfall 2).

**#10/#11 verbatim** — etalon `import_no_public.go`/`import_no_weak.go` copy with `imp.Public`/`imp.Weak` (repo main used `imp.Modifier == ImportModifierPublic/Weak`); position test-verified equal.

### Batch 3 — BASIC semantic

**#5 EnumFirstValueZero** — repo `enum_first_value_zero.go`: main's **string compare** (`Number != "0"`) and position at value `Meta.Pos`. Port etalon's walk (`Visitor{EnumValue:...}`), keep the string compare, keep etalon's `len(Values)==0` guard as the one documented deviation (main panics on empty enums), emit at `value.Pos` not `NumberPos`.

**#6 EnumNoAllowAlias** — repo `enum_no_allow_alias.go`: main reports when **any** enum-level option named `allow_alias` exists (any value) at enum position. Etalon checks `Value=="true"` at option position — do not port. Walk `Visitor{Enum:...}`, iterate `enum.Options` for the name, emit at `enum.Pos`.

**#12 ImportUsed** — repo `internal/rules/import_used.go` (main's syntax-walk `existInProto` semantics, unconditional) is the behavior oracle; etalon `import_used.go` contributes the stateless `importUsage` local struct (receiver-state extraction for parallel windows) and the fallback-walk mechanics. Port main's syntax walk unconditionally; the etalon's resolution-based `validateResolved` is a **deferred decision**, not Phase 4 behavior.

**#16-22 option-comparison ModuleRules (×7, one shape)** — repo `internal/rules/package_same_go_package.go:27-48` is main's oracle (first option value per package wins via cache; skip files without the option entirely; compare `option.Constant` **as written**; issue at `option.Meta.Pos`):
```go
for _, option := range protoInfo.Info.ProtoBody.Options {
    if option.OptionName == "go_package" {
        if p.cache[packageName] == "" { p.cache[packageName] = option.Constant; continue }
        if p.cache[packageName] != option.Constant {
            res = core.AppendIssue(res, p, option.Meta.Pos, option.Constant, option.Comments)
        }
    }
}
```
Etalon `package_same_go_package.go:24-32` contributes the Collect shape (`fileOption(file, "go_package")` + `optionString(opt.Value)` + `opt.Pos`, add **only when the option exists**). Port: etalon Collect + repo's skip-optionless + first-exempt Finish + **main's comparison basis** (written constant vs etalon's unquoted — Pitfall 10). #22 PackageSameSwiftPrefix: main compares the literal `"ruby_package"` — bug-for-bug pin with a regression test (Open Question 1).

### Batch 4 — DEFAULT (8)

- **#25 EnumValuePrefix** — repo `enum_value_prefix.go`: carry main's private `pascalToUpperSnake` (underscore before every uppercase rune) into the new file; `HasPrefix(ident, prefix)` **without** trailing `_`. Etalon's `toUpperSnakeCase(name)+"_"` is not portable (acronym + name-only-value divergences).
- **#26 EnumZeroValueSuffix** — repo `enum_zero_value_suffix.go`: first value only, exact equality with `pascalToUpperSnake(enumName)+"_"+Suffix`; **keep main's `ENV:"ENUM_ZERO_VALUE_SUFFIX"` struct tag** (etalon dropped it). Position: first value's `Meta.Pos` ≡ `value.Pos`.
- **#27 FileLowerSnakeCase** — repo `file_lower_snake_case.go`: main checks the **full filename including `.proto`** against the package-style regexp and emits `core.Position{Filename: path}` (0:0). Etalon strips the extension and emits 1:1 — do not port.
- **#28 RPCRequestResponseUnique (ModuleRule)** — repo `rpc_request_response_unique.go`: main is per-file uniqueness on **written** type strings (`lo.Contains` per file), **first user exempt**, input==output counts as a duplicate within the file; position at `rpc.Meta.Pos`. Etalon contributes the ModuleRule shape; `Finish` groups by file and reproduces first-exempt on written names (`Resolved.RPCTypes` unused for parity).
- **#29/#30 RPCRequest/ResponseStandardName** — repo `rpc_request_standard_name.go`: main requires written `MessageType == RPCName+"Request"` or `ServiceName+RPCName+"Request"` **exactly** (qualified written type FAILS). Emit at `rpc.Pos` (Meta.Pos), not etalon's `rpc.Request.Pos`.
- **#31 PackageVersionSuffix** — repo `package_version_suffix.go` regexp (unanchored; "v1" passes). Etalon's `packageHasVersionSuffix` helper (in `lint_helpers.go:157-190`) is **not used** by this port.
- **#32 ServiceSuffix** — etalon `service_suffix.go` verbatim (`HasSuffix(name, Suffix)`); position at `service.Pos` (main col 1); keep main's `Suffix` config plumbing/tag.

### Batch 5 — COMMENTS (7, one uniform port)

**Exemplar pair** — repo `internal/rules/comment_enum.go:21-33` (main's predicate: two-level nested loop, `len(enum.Comments) == 0` presence check) vs etalon `comment_enum.go:21-25` (walk shape, wrong predicate):
```go
// Main predicate (wins): presence, directive-only comments count as documentation
if len(enum.Comments) == 0 {
    res = core.AppendIssue(res, c, enum.Meta.Pos, enum.EnumName, enum.Comments)
}
// Etalon walk (mechanics only):
protoInfo.File.Walk(protoast.Visitor{Enum: func(enum *protoast.Enum) {
    if !leadingCommentDocuments(&enum.Decl) { ... }  // DO NOT port this predicate
}})
```
Port: `Visitor` walk + "no attached comment" predicate over protoast attachment facts (`Decl.Leading`/`Detached`/`Trailing`, `internal/protoast/protoast.go:63-73` — port "comment attached at all", i.e. `HasCommentAbove()`-plus-trailing semantics matching main's `len(Comments)==0`; pin the exact predicate against main's carried tests). **Never** use `leadingCommentDocuments` (`lint_helpers.go:72-83`) — its directive-only⇒undocumented reading diverges from main (Pitfall 3). Apply to #33-#39 with the matching `Visitor` callback (`Enum`, `EnumValue`, `Field`, `Message`, `Oneof`, `RPC`, `Service`); #35 CommentField: encode main behavior literally, no `!field.Group` guard (assumption A4). Shared fixture must include a directive-only-comment declaration.

### Batch 6 — UNARY_RPC (2) + finish

- **#40/#41** — etalon `rpc_no_client_streaming.go`/`rpc_no_server_streaming.go` verbatim: `if rpc.Request.IsStream { ... }` → `rpc.Request.Stream`; emit at `rpc.Pos` (≡ main `Meta.Pos`, alias-only test diff).
- **Finish:** complete `builder_linked.go` wiring for all 41; finish `lint_helpers.go` re-alignment (`mismatches` takes the new contract, plus a main-first-exempt variant per Shared Pattern 5); remove `.golangci.yml` line 14 exclusion (`- internal/rules/lint_helpers.go`); finalize `TestModuleRuleSet` count guard (exactly 10 ModuleRules); full gates (`go test -race -count=1 ./internal/...`, build, schema drift-empty).

## Shared Patterns

### 1. Two-contract coexistence naming (applies to all 41 new types)
**Source:** `internal/core/lint_model.go:93-97` (LinkedRuleName reflection) + RESEARCH Option A.
```go
// LinkedRuleName derives names by reflection over the type name — a suffixed
// type would yield NAME_LINKED unless overridden. Additive preference:
// if r, ok := rule.(interface{ LintRuleName() string }); ok { return r.LintRuleName() }
// return toUpperSnakeCase(reflect.TypeOf(rule).Elem().Name())
```
Each new type: `var _ core.LintRule = (*MessagePascalCaseLinked)(nil)` + `func (m *MessagePascalCaseLinked) LintRuleName() string { return "MESSAGE_PASCAL_CASE" }`. Every new identifier goes in the rename-map comment (`lint_model.go:6-27`) for the Phase-5 consolidation. Old files are never edited (Pitfall 7/8).

### 2. Rule impl imports block
**Source:** etalon `message_pascal_case.go:3-6` (external) + repo `message_pascal_case.go:3-7`:
```go
import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)
```
New rule files always import both; tests additionally import `internal/rules` (external `rules_test` package, per repo convention). Regexp rules declare the compiled regexp inside `Validate` (main's style, lines 23 of `message_pascal_case.go`) — keep it, no package-level caching needed for parity.

### 3. Linked issue construction
**Source:** `internal/core/dom.go:126-164`; etalon `internal/core/dom.go:80-82` (external shape).
4-arg linked variant reusing `buildError`; no meta.Position conversion (core.Position is the protoast alias). Rules never build `core.Issue` literals by hand (Don't-Hand-Roll) — except the two zero-position rules, which pass `core.Position{Filename: info.Path}` through the same constructor.

### 4. Position mapping (systematic — check before every rule)
**Source:** `internal/protoast/protoast.go:63-71` (`Decl.Pos` = first token, `NamePos` = name), verified per-kind in RESEARCH. Main's `Meta.Pos` → new `Decl.Pos` (`.Pos` on the walked decl). **Forbidden:** `NamePos`, `EnumValue.NumberPos`, `RPCType.Pos` (etalon's finer tokens), `protoast.FilePosition` (1:1) for #2/#27 (main wants 0:0 zero literal). Carried main test expectations are the oracle — `internal/rules/message_pascal_case_test.go:36-46` expects `Offset: 536, Line: 26, Column: 1` (keyword), not the etalon's 544/9.

### 5. ModuleRule first-exempt Finish (all 10 cross-file rules)
**Source:** repo cache predicates (`directory_same_package.go:34-43`, `package_same_go_package.go:37-48`) + RESEARCH Pattern 2:
```go
// Finish: first value per group (walk order) exempt; later differing files issue at their own pos.
func (d *DirectorySamePackageLinked) Finish() []core.IssueInfo {
	var res []core.IssueInfo
	for _, dir := range d.files.order {           // add() preserves insertion order (lint_helpers.go:116-124)
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
Option-comparison rules (#16-22): `Collect` adds a value **only when the option exists**; #1/#4 always add. #28 groups by file. `Collect` mutates receiver state only in the engine's serial drain (`internal/core/lint_linked.go` — engine drives `c.linkedRules`, Collect in walk order, Finish after last file); per-file rules stay stateless.

### 6. ModuleRule test driver
**Source:** etalon `internal/rules/init_test.go:92-115` (external) — port verbatim into the rewritten `init_test.go`:
```go
func validate(rule core.LintRule, infos ...core.LinkedFileInfo) ([]core.Issue, error) {
	moduleRule, ok := rule.(core.ModuleRule)
	if !ok { /* loop Validate per info */ }
	for _, info := range infos { if err := moduleRule.Collect(info); err != nil { return nil, err } }
	var issues []core.Issue
	for _, found := range moduleRule.Finish() { issues = append(issues, found.Issue) }
	return issues, nil
}
```
The 10 ModuleRule tests must drive Collect/Finish this way — calling `Validate` is a no-op stub (Pitfall 6).

### 7. Batch parity harness (every batch)
**Source:** `internal/core/lint_suppressions_parity_test.go:63-69` (`writeCorpusInto`, fixture-map pattern), `internal/rules/validation_parity_test.go:108` (in-package `core.New(core.Options{...})` construction), RESEARCH sketch:
- Fixtures: compiling `.proto` trees in `t.TempDir()` (broken files legitimately diverge — COMPILE routing).
- Old side: `rules.New(cfg)` → `core.New(core.Options{Rules: ..., IgnoreOnly: ..., AllowCommentIgnores: true, KnownLintRules: rules.AllRuleNames()})` → `.Lint`.
- New side: `rules.NewLinked(cfg)` → `core.New(core.Options{LinkedRules: ..., AllowCommentIgnores: true, KnownLintRules: rules.AllRuleNames()})` → `.LintLinked`.
- Compare sorted (RuleName, Path, Line, Column) multisets; messages optional.
- Each batch's fixture cells come from the RESEARCH inventory's "Main-wins decisions" column (acronym/digit/underscore edges, optionless files, qualified written types, directive-only comments, 3-file groups where only 2nd/3rd diverge, swift_prefix ruby_package pin).
- Standing constraints ride every batch: suppression-centrality cells green; readSource-once cell green (Wave 0 onward).

### 8. Test carryover shape
**Source:** `internal/rules/message_pascal_case_test.go:26-73` (external `rules_test` package, `start(t)`, named map cases, `r.Contains(issues, *tc.wantIssues)`). Rewritten tests keep main's case set and expected `core.Issue` literals verbatim, swapping `rule.Validate(protos[tc.fileName])` for the new type (and `validate()` for ModuleRules). testify `require`/`assert`; `t.Parallel()` per project go-testing skill.

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `AppendLinkedIssue` in `internal/core/lint_model.go` | model/utility | transform | Main has no LintRule issue constructor (old `AppendIssue` pins the old `Rule` contract and converts from meta.Position); etalon's 4-arg is the external shape to rename-map. Partial analog: `dom.go:126-164`. |
| `internal/rules/parity_<batch>_test.go` | test | request-response | No existing test runs both engines on one fixture tree and multiset-compares. Partial analogs: `lint_suppressions_parity_test.go` (fixture + oracle-carryover pattern), `validation_parity_test.go` (in-package construction cells). |
| `NewLinked` in `internal/rules/builder_linked.go` | registry/config | CRUD | No linked registry exists; exact structural analog is `builder.go:61-143` with element-type swap. |
| Position-probe test, `TestModuleRuleSet` guard, readSource-once cell | test | transform/event-driven | Net-new Wave-0 cells; analogs are style references (`position_parity_test.go`, `lint_linked_test.go`). |
| 32 main-diverged rule ports (the port itself) | rule | transform | The *combination* (etalon walk + main predicate) exists nowhere; both halves have exact analogs per batch sections above. |

## Metadata

**Analog search scope:** `internal/rules/`, `internal/core/`, `internal/protoast/`, `.golangci.yml` (repo, all git-tracked via `git ls-files`); `/tmp/gsd-research/compiler-lab/easyp-cli/internal/rules/` + `internal/core/dom.go` (external etalon @ 08cef10 — porting source only)
**Files scanned:** ~25 in depth (6 rule pairs, both init_test.go, both lint_helpers.go, builder, model, engine contract, protoast model/walk, 2 parity tests, 1 test oracle, lint config)
**Pattern extraction date:** 2026-11-19
