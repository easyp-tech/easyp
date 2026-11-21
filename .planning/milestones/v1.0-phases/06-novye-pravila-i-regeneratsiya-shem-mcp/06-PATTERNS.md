# Phase 6: Новые правила и регенерация схем/MCP - Pattern Map

**Mapped:** 2026-11-21
**Files analyzed:** 22 (6 create, 16 modify/regen)
**Analogs found:** 21 / 22 (1 no-analog: parity compare script — external mechanism, RESEARCH Pattern 4 is the source)

## External Porting Source (flagged)

The three rule implementation files are **verbatim ports** from the lab etalon:

- **Port source:** `/tmp/gsd-research/compiler-lab/easyp-cli/internal/rules/` (compiler-lab @ `08cef10909c0722f7042ba17b5c11951f9c6b6f6` — SHA re-verified this session)
- This path is **outside this repo's git** — a porting reference, not a tracked analog. Per the tracked-source gate, no plan action may commit to or depend on it as repo source; plan Task 0 must re-verify the SHA (`git -C /tmp/gsd-research/compiler-lab rev-parse HEAD`) and fall back to `gh repo clone easyp-tech/compiler-lab && git checkout 08cef10` if /tmp was wiped.
- Every symbol the lab files reference already exists in this repo with identical name/shape (verified in RESEARCH Pattern 1 and spot-checked below) — zero seam adaptations.
- The lab carries **no unit tests** for these rules; the only lab-side pins are name-presence assertions that port as one-line NotContains→Contains flips (excerpts below).

## File Classification

| New/Modified File | Action | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|--------|------|-----------|----------------|---------------|
| `internal/rules/package_no_import_cycle.go` | create | rule (core.ModuleRule) | event-driven (Collect/Finish batch) | LAB `package_no_import_cycle.go` (verbatim port); repo shape: `internal/rules/package_same_directory_linked.go` | exact (external port) |
| `internal/rules/syntax_specified.go` | create | rule (core.Rule) | request-response (per-file Validate) | LAB `syntax_specified.go` (verbatim port) | exact (external port) |
| `internal/rules/field_not_required.go` | create | rule (core.Rule) | request-response (per-file Validate + visitor walk) | LAB `field_not_required.go` (verbatim port) | exact (external port) |
| `internal/rules/package_no_import_cycle_test.go` | create | test (cross-file semantics) | event-driven | `internal/rules/package_same_directory_test.go` + `internal/rules/init_test.go` harness | role-match |
| `internal/rules/syntax_specified_test.go` | create | test (per-file semantics) | request-response | `internal/rules/package_defined_test.go` | role-match |
| `internal/rules/field_not_required_test.go` | create | test (per-file visitor semantics) | request-response | `internal/rules/package_defined_test.go` (no visitor-field rule test exists yet — walk assertions are new) | role-match |
| `testdata/import_cycle/a.proto` + `b.proto` + same-package negative | create | test fixture | file-I/O | `testdata/import_used/` (multi-file dir precedent; imports resolved from repo-root `testdata/`) | role-match |
| parity script + teeth corpus (throwaway) | create | integration harness | external binary compare | **none in repo** — RESEARCH Pattern 4 + lab `bench/lab.py` JSON-line shape | none |
| `internal/rules/builder.go` | modify | registry (name lists) | — | self: `addMinimal` :132-139, `addBasic` :141-163 | exact |
| `internal/rules/builder_linked.go` | modify | registry (constructor) | — | self: `New` allRules :26-81 | exact |
| `internal/rules/selection.go` | modify | validator | — | self: `ValidateNames` :14-27 (delete 4-line branch) | exact |
| `internal/rules/module_rules_guard_test.go` | modify | test guard | — | self: `moduleRuleNames` map :19-31 (10 → 11 names) | exact |
| `internal/core/lint_model.go` | modify | comment-only (D-03 comment, 41 → 44) | — | self: :27-34 | exact |
| `internal/rules/validation_parity_test.go` | modify | test pin flip | — | self: :70-78 cell | exact |
| `internal/rules/selection_test.go` | modify | test pin flip | — | self: :54-55 cells | exact |
| `internal/rules/builder_test.go` | modify | test pin flip | — | self: :69 `NotContains` → `Contains` (lab mirror `builder_test.go:69`) | exact |
| `internal/config/v1/lint_selection_test.go` | modify | test pin flip | — | self: :18, :32, :57 (lab mirror `lint_selection_test.go:57` `Contains`) | exact |
| `mcp/easypconfig/describe_test.go` | modify | test pin flip | — | self: :296-298 (lab mirror `describe_test.go:298` `Contains`) | exact |
| `schemas/easyp-v1.schema.json`, `schemas/easyp.schema.json` | regen (tool) | generated config | transform | `internal/config/v1/schema.go:69-76` derivation; regenerate via `go run ./cmd/easyp schema-gen` — never hand-edit | exact (tool-driven) |
| `mcp/easypconfig/docs.go` | modify (optional) | doc wording | — | self: :38 ("unimplemented" wording polish) | exact |
| `internal/rules/golden_full_registry_test.go` | probe only | regression guard | — | self: :47 `goldenCorpusUse = rules.AllRuleNames()` — expect green, no edit | exact |

## Pattern Assignments

### `internal/rules/package_no_import_cycle.go` (rule, event-driven ModuleRule)

**Analog (verbatim port source):** LAB `/tmp/gsd-research/compiler-lab/easyp-cli/internal/rules/package_no_import_cycle.go` (109 lines — copy entire file)
**Repo structural analog:** `internal/rules/package_same_directory_linked.go` (same Message/Validate-stub/Collect/Finish shape)

**Verbatim lab source (copy whole file):**

```go
// Source: LAB internal/rules/package_no_import_cycle.go:1-109 (PORT VERBATIM)
package rules

import (
	"slices"
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.ModuleRule = (*PackageNoImportCycle)(nil)

// PackageNoImportCycle this rule checks that packages do not import each other in a cycle.
// The package graph covers linted files and the packages of their direct
// imports, so a cycle that passes only through dependencies is not seen.
type PackageNoImportCycle struct {
	// package -> imported package -> import statements of linted files
	edges map[string]map[string][]packageImport
}

type packageImport struct {
	path, location string
	pos            core.Position
}

// Message implements lint.Rule.
func (p *PackageNoImportCycle) Message() string {
	return "package import cycle"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (p *PackageNoImportCycle) Validate(core.ProtoInfo) ([]core.Issue, error) { return nil, nil }

// Collect implements core.ModuleRule.
func (p *PackageNoImportCycle) Collect(protoInfo core.ProtoInfo) error {
	if p.edges == nil {
		p.edges = make(map[string]map[string][]packageImport)
	}
	pkg := protoInfo.File.PackageName()
	for _, imp := range protoInfo.File.Imports {
		imported, ok := protoInfo.ImportedFiles[core.ConvertImportPath(imp.Location)]
		if !ok || imported.PackageName() == pkg {
			continue
		}
		if p.edges[pkg] == nil {
			p.edges[pkg] = make(map[string][]packageImport)
		}
		target := imported.PackageName()
		p.edges[pkg][target] = append(p.edges[pkg][target], packageImport{path: protoInfo.Path, location: imp.Location, pos: imp.Pos})
	}
	return nil
}

// Finish implements core.ModuleRule. Like buf, it reports every import of a
// package that leads into a cycle back to that package.
func (p *PackageNoImportCycle) Finish() []core.IssueInfo {
	var res []core.IssueInfo
	for _, pkg := range sortedKeys(p.edges) {
		if pkg == "" {
			continue
		}
		for _, imported := range sortedKeys(p.edges[pkg]) {
			if imported == "" {
				continue
			}
			cycle := p.cycle(imported, map[string]bool{pkg: true}, []string{pkg})
			if len(cycle) == 0 {
				continue
			}
			for _, imp := range p.edges[pkg][imported] {
				for _, issue := range core.AppendIssue(nil, p, imp.pos, strings.Join(cycle, " -> ")) {
					res = append(res, core.IssueInfo{Issue: issue, Path: imp.path})
				}
			}
		}
	}
	p.edges = nil
	return res
}

// cycle returns the package path back to path[0] through pkg, if any.
func (p *PackageNoImportCycle) cycle(pkg string, used map[string]bool, path []string) []string {
	path = append(path, pkg)
	if used[pkg] {
		if path[0] == path[len(path)-1] {
			return path
		}
		return nil
	}
	used[pkg] = true
	for _, next := range sortedKeys(p.edges[pkg]) {
		if next == "" {
			continue
		}
		if found := p.cycle(next, used, path); len(found) > 0 {
			return found
		}
	}
	delete(used, pkg)
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
```

**Symbol-availability verification (why verbatim works):** every referenced symbol exists in-repo:
- `core.ModuleRule` — `internal/core/lint_model.go:35-41`
- `core.ProtoInfo{Path, File, ImportedFiles}` — `internal/core/lint_model.go:53-67`
- `core.AppendIssue` 4-arg — `internal/core/lint_model.go:73-80`
- `core.Position` — `internal/core/dom.go:19` (alias of `protoast.Position`)
- `core.IssueInfo{Issue, Path}` — `internal/core/dom.go:28-31`
- `core.ConvertImportPath` — `internal/core/dom.go:153-155`
- `protoast.File.PackageName()/.Imports/.Walk` — `internal/protoast/protoast.go:95,112-117`, `walk.go:7-18`
- `sortedKeys`/`packageImport` package-local — no name collisions in repo `internal/rules` (verified this session)

**Filename note:** the repo's Phase-4 ported rules still carry legacy `*_linked.go` filenames (`package_same_directory_linked.go`, …). AGENTS.md's new-rule convention is `internal/rules/<rule>.go` + `<rule>_test.go`, and the lab files are unsuffixed — the three new files use **unsuffixed lab filenames** (`package_no_import_cycle.go`, `syntax_specified.go`, `field_not_required.go`). No unsuffixed name collides (only `builder.go`, `builder_linked.go`, `selection.go`, `lint_helpers.go` are unsuffixed non-test files today).

---

### `internal/rules/syntax_specified.go` (rule, request-response)

**Analog (verbatim port source):** LAB `internal/rules/syntax_specified.go` (27 lines — copy entire file)

```go
// Source: LAB internal/rules/syntax_specified.go:1-27 (PORT VERBATIM)
package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*SyntaxSpecified)(nil)

// SyntaxSpecified this rule checks that every file has a syntax or edition statement.
type SyntaxSpecified struct{}

// Message implements lint.Rule.
func (s *SyntaxSpecified) Message() string {
	return "syntax should be specified"
}

// Validate implements lint.Rule.
func (s *SyntaxSpecified) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	if !protoInfo.File.SyntaxSpecified {
		res = core.AppendIssue(res, s, protoast.FilePosition(protoInfo.Path), protoInfo.Path)
	}

	return res, nil
}
```

**Symbol availability:** `protoInfo.File.SyntaxSpecified` — `internal/protoast/protoast.go:79-80`; `protoast.FilePosition(path)` (line 1, column 1) — `internal/protoast/protoast.go:27-30`.

---

### `internal/rules/field_not_required.go` (rule, request-response + visitor walk)

**Analog (verbatim port source):** LAB `internal/rules/field_not_required.go` (29 lines — copy entire file)

```go
// Source: LAB internal/rules/field_not_required.go:1-29 (PORT VERBATIM)
package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*FieldNotRequired)(nil)

// FieldNotRequired this rule checks that fields are not required.
type FieldNotRequired struct{}

// Message implements lint.Rule.
func (f *FieldNotRequired) Message() string {
	return "field should not be required"
}

// Validate implements lint.Rule.
func (f *FieldNotRequired) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{Field: func(field *protoast.Field) {
		if field.Label == "required" {
			res = core.AppendIssue(res, f, field.NamePos, field.Name)
		}
	}})

	return res, nil
}
```

**Symbol availability:** `field.Label` (`"repeated"/"optional"/"required"/""`) — `internal/protoast/protoast.go:141`; `field.NamePos` — `internal/protoast/protoast.go:66` (`Decl.NamePos`); `File.Walk(Visitor{Field: …})` — `internal/protoast/walk.go:7-18`.

**Position-semantics note for the plan:** the 41 ported rules use main-wins positions (`Decl.Pos` keyword positions, T-04-03); these two per-file rules keep the lab's `NamePos`/`FilePosition` because lab parity (LINT-NEWRULES-03) is the oracle for **new** rules, not main parity.

---

### `internal/rules/package_no_import_cycle_test.go` (test, cross-file)

**Analog:** `internal/rules/package_same_directory_test.go` (whole file, 80 lines) + `internal/rules/init_test.go` harness

**Harness pattern** (`internal/rules/init_test.go:35-94` — reuse, do not duplicate):

```go
// Source: internal/rules/init_test.go:35-67 — parseLinkedFile builds a ProtoInfo
// for a testdata file; its direct imports are parsed from ./../../testdata
// (repo-root testdata/), skipping missing ones.
func parseLinkedFile(t testing.TB, path string) core.ProtoInfo { … }

// Source: internal/rules/init_test.go:71-94 — validate dispatches on the rule's
// contract: ModuleRule ⇒ Collect(every file) + Finish(); Rule ⇒ per-file Validate.
func validate(rule core.Rule, infos ...core.ProtoInfo) ([]core.Issue, error) { … }
```

**Table-test pattern** (`internal/rules/package_same_directory_test.go:26-80`):

```go
// Source: internal/rules/package_same_directory_test.go:26-78
func TestPackageSameDirectory_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames  []string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			fileNames: []string{invalidAuthProto2, invalidAuthProto4}, // multi-file fixture consts from init_test.go:14-29
			wantIssues: &core.Issue{
				Position:   protoast.Position{Filename: "", Offset: 20, Line: 3, Column: 1},
				SourceName: "Queue",
				Message:    "different proto files in the same package should be in the same directory",
				RuleName:   "PACKAGE_SAME_DIRECTORY",
			},
			wantErr: nil,
		},
		"valid": {
			fileNames: []string{validAuthProto, validAuthProto2},
			wantErr:   nil,
		},
	}
	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule := &rules.PackageSameDirectory{}
			var infos []core.ProtoInfo
			for _, fileName := range tc.fileNames {
				infos = append(infos, parseLinkedFile(t, fileName))
			}
			issues, err := validate(rule, infos...)
			require.ErrorIs(t, err, tc.wantErr)
			switch {
			case tc.wantIssues != nil:
				require.Contains(t, issues, *tc.wantIssues)
			case len(issues) > 0:
				require.Empty(t, issues)
			}
		})
	}
}
```

**Message-test pattern** (`package_same_directory_test.go:13-24`): every rule test opens with a `Test<Rule>_Message` asserting the verbatim message string.

**Cycle-rule pins to author (from RESEARCH Code Examples):** fixture consts follow the `init_test.go:14-29` const block (add e.g. `importCycleA = "./../../testdata/import_cycle/a.proto"`); assert every edge of the cycle is reported, `SourceName == "a.v1 -> b.v1 -> a.v1"` chain, position at the import keyword, `IssueInfo.Path` == importing file; plus an acyclic negative and a same-package-import negative. **Note:** `validate` collapses `IssueInfo` to `Issue` (loses `Path`) — for Path-specific assertions call `Collect`/`Finish` directly as init_test.go:84-93 does internally, or assert via the returned Issues plus a direct Finish call.

**Fixture location correction (important):** RESEARCH Wave 0 says `internal/rules/testdata/import_cycle/`, but `parseLinkedFile` resolves direct imports from `./../../testdata` — the **repo-root** `testdata/` (verified: `init_test.go:15-28` consts and `init_test.go:49`). The cycle fixture belongs at `testdata/import_cycle/` (mirroring `testdata/import_used/`, a tracked multi-file fixture dir), with `a.proto` importing `"import_cycle/b.proto"` and vice versa.

---

### `internal/rules/syntax_specified_test.go` and `internal/rules/field_not_required_test.go` (tests, per-file)

**Analog:** `internal/rules/package_defined_test.go` (whole file, 72 lines) — single-file rule calling `rule.Validate(parseLinkedFile(t, tc.fileName))` directly:

```go
// Source: internal/rules/package_defined_test.go:26-71
func TestPackageDefined_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: invalidAuthProtoEmptyPkg,
			wantIssues: &core.Issue{
				Position:   protoast.Position{Filename: "./../../testdata/auth/empty_pkg.proto", Offset: 0, Line: 0, Column: 0},
				SourceName: "./../../testdata/auth/empty_pkg.proto",
				Message:    "package should be defined",
				RuleName:   "PACKAGE_DEFINED",
			},
			wantErr: nil,
		},
		"valid": { fileName: validAuthProto, wantErr: nil },
	}
	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule := rules.PackageDefined{}
			issues, err := rule.Validate(parseLinkedFile(t, tc.fileName))
			require.ErrorIs(t, err, tc.wantErr)
			// same require.Contains / require.Empty switch as the module test above
		})
	}
}
```

Fixtures needed: a testdata file with no syntax statement (for SYNTAX_SPECIFIED) and a proto2 file with a `required` field (for FIELD_NOT_REQUIRED); add consts to the `init_test.go:14-29` block, mirroring `invalidAuthProto…` naming. Pin exact `Position`/`SourceName`/`Message`/`RuleName` per the lab semantics (`FilePosition` = line 1 col 1; `NamePos` = declared-name token).

---

### `internal/rules/builder.go` (modify — name lists)

**Analog:** self. Insertions mirror LAB `internal/rules/builder.go` order (verified this session: lab `builder.go:227-228` and `:253-255`).

```go
// Source: internal/rules/builder.go:132-139 — addMinimal: append 5th entry
// after PackageSameDirectory (lab builder.go:227-228 order)
func addMinimal(res []string) []string {
	res = append(res, core.GetRuleName(&DirectorySamePackage{}))
	res = append(res, core.GetRuleName(&PackageDefined{}))
	res = append(res, core.GetRuleName(&PackageDirectoryMatch{}))
	res = append(res, core.GetRuleName(&PackageSameDirectory{}))
	res = append(res, core.GetRuleName(&PackageNoImportCycle{})) // NEW

	return res
}

// Source: internal/rules/builder.go:141-163 — addBasic: append 21st/22nd entries
// after ServicePascalCase (lab builder.go:253-255 order)
func addBasic(res []string) []string {
	// … 19 existing entries unchanged …
	res = append(res, core.GetRuleName(&RPCPascalCase{}))
	res = append(res, core.GetRuleName(&ServicePascalCase{}))
	res = append(res, core.GetRuleName(&SyntaxSpecified{}))     // NEW
	res = append(res, core.GetRuleName(&FieldNotRequired{}))    // NEW
	return res
}
```

`add*` lists feed `AllGroups`/`AllRuleNames`/`AllLintUseValues` (:24-55), `unwrapLintGroups` (:95-116), `unwrapIgnoreOnly` (:57-93) — no other edits in this file. Group count stays 5; registry count 41 → 44.

---

### `internal/rules/builder_linked.go` (modify — constructor)

**Analog:** self; insertion order mirrors LAB `builder.go:73-74` (`&PackageSameDirectory{}, &PackageNoImportCycle{},`) and `:96-98` (`&ServicePascalCase{}, &SyntaxSpecified{}, &FieldNotRequired{},`).

```go
// Source: internal/rules/builder_linked.go:26-81 — New's allRules; insertions only
allRules := []core.Rule{
	//	minGroup
	&DirectorySamePackage{},
	&PackageDefined{},
	&PackageDirectoryMatch{Root: ".", Prefix: cfg.PackageDirectoryPrefix},
	&PackageSameDirectory{},
	&PackageNoImportCycle{}, // NEW — 5th in MINIMAL, lab placement

	//	basicGroup
	&EnumFirstValueZero{},
	// … 18 more BASIC entries unchanged …
	&RPCPascalCase{},
	&ServicePascalCase{},
	&SyntaxSpecified{},    // NEW — 21st
	&FieldNotRequired{},   // NEW — 22nd
	// … DEFAULT / COMMENTS / UNARY_RPC unchanged …
}
```

Membership must agree with `add*` or `New(Use: [MINIMAL])` fails name lookup (:94-98). Update the "41" doc comment at `builder_linked.go:14` to 44 in the same commit (do not propagate the "43" REQUIREMENTS discrepancy — see `lint_model.go:28-29`).

---

### `internal/rules/selection.go` (modify — special-case deletion)

**Analog:** self. Delete exactly the 4-line branch (:18-21):

```go
// Source: internal/rules/selection.go:14-27 — BEFORE (delete branch); AFTER,
// the loop keeps only the generic Contains path
func ValidateNames(names []string) error {
	allowed := AllLintUseValues()
	var issues []error
	for _, name := range names {
		if name == "PACKAGE_NO_IMPORT_CYCLE" {          // DELETE lines 18-21
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

Same-commit ordering matters: registry first, then this deletion (removing the branch first leaves an incoherent intermediate state — validation passes, construction fails with a different `ErrInvalidRule` wrap).

---

### Flip pins (5 test files, same commit)

**1. `internal/rules/builder_test.go:69`** — `require.NotContains(t, values, "PACKAGE_NO_IMPORT_CYCLE")` → `require.Contains` (mirror: LAB `internal/rules/builder_test.go:69` is `require.Contains`).

**2. `internal/rules/module_rules_guard_test.go:19-31`** — add `"PACKAGE_NO_IMPORT_CYCLE": true` to the map (10 → 11, the D-03 eleven); update the header comment :3-7 ("PackageNoImportCycle arrives with Phase 6 and must be absent until then") and the "exactly the ten" wording. Mirror: LAB has no guard — the lab's `builder_test.go:69` Contains is the end-state evidence.

```go
// Source: internal/rules/module_rules_guard_test.go:19-31 (extend map)
var moduleRuleNames = map[string]bool{
	"DIRECTORY_SAME_PACKAGE":           true,
	// … 9 existing entries unchanged …
	"RPC_REQUEST_RESPONSE_UNIQUE":      true,
	"PACKAGE_NO_IMPORT_CYCLE":          true, // NEW — D-03 eleven
}
```

**3. `internal/rules/validation_parity_test.go:70-78`** — replace the `package_no_import_cycle_special_case` cell (expects `"PACKAGE_NO_IMPORT_CYCLE is not implemented"`) with valid-construction cells (name in `use` constructs; `except`/`ignore_only` accept it); update the :71-73 comment. Table shape at :80-90 (`require.ErrorIs(core.ErrInvalidRule)` + `ErrorContains`).

**4. `internal/rules/selection_test.go:54-55`** — drop/rewrite the two `"unimplemented enabled/disabled"` cells (`cfg: config.LintConfig{Use: []string{"PACKAGE_NO_IMPORT_CYCLE"}}, want: "not implemented"`) as valid-selection cells or removals.

**5. `internal/config/v1/lint_selection_test.go`** — three flips:
- `:18` — fixture raw YAML: swap `PACKAGE_NO_IMPORT_CYCLE` in `enable: [NO_SUCH_RULE, PACKAGE_NO_IMPORT_CYCLE]` for another invalid name (keeps `require.Len(t, issues, 4)` at :20 and the 4-name loop at :32 intact with 4 invalid locations);
- `:32` — same swap in the `require.ErrorContains` name list;
- `:57` — `assert.NotContains(t, string(data), "PACKAGE_NO_IMPORT_CYCLE")` → `assert.Contains` (mirror: LAB `internal/config/v1/lint_selection_test.go:57` is `assert.Contains`).

**6. `mcp/easypconfig/describe_test.go:296-298`** — `assert.NotContains(t, values, "PACKAGE_NO_IMPORT_CYCLE")` → `assert.Contains` (mirror: LAB `mcp/easypconfig/describe_test.go:298` is `assert.Contains`); keep the surrounding `TestDescribeLintSelectionMatchesExecutableRules` loop over `linters.enable`, `linters.disable`, `issues.exclude-rules[].linters` (:287) unchanged.

---

### `schemas/easyp-v1.schema.json` + `schemas/easyp.schema.json` (regenerate — no code change)

**Analog (derivation source):** `internal/config/v1/schema.go:69-76` — enums re-derive at generation time from `rules.AllLintUseValues()`; **no code change**:

```go
// Source: internal/config/v1/schema.go:72-75 — the three enum spots that gain
// the new names automatically when the registry grows
for _, field := range []string{"enable", "disable"} {
	policy.Properties["linters"].Properties[field].Items.Enum = rules.AllLintUseValues()
}
policy.Properties["issues"].Properties["exclude-rules"].Items.Properties["linters"].Items.Enum = rules.AllLintUseValues()
```

Regenerate with `go run ./cmd/easyp schema-gen` (Taskfile equivalent of `task schema:generate`; `task` is not on PATH this environment). Only these two of the six schema files change (each gains the 3 names at its 3 enum spots — today exactly these two carry `PACKAGE_DIRECTORY_MATCH` ×3). Verify with `git diff --exit-code schemas/`. Never hand-edit (AGENTS.md). End-state proof: LAB `schemas/easyp*.schema.json` carry `PACKAGE_NO_IMPORT_CYCLE` at the same 3×2 spots.

**MCP flow (no code change):** `mcp/easypconfig/schema.go:12-14` (`MarshalConfigJSONSchema` → `v1.SchemaJSON`) and `:34-37` (`SchemaByPathFor` → `v1.SchemaJSON`) flow through automatically; only the `describe_test.go` pin flips. Optional one-word polish at `mcp/easypconfig/docs.go:38` ("unknown or unimplemented names are rejected" — the "unimplemented" set becomes empty).

### `internal/rules/golden_full_registry_test.go` (probe only — no edit expected)

`goldenCorpusUse = rules.AllRuleNames()` (:47) drives the golden corpus with the extended registry automatically. RESEARCH verified the corpus is clean for all three rules (13/13 syntax statements, zero `required`, single acyclic import); probe-run `TestLintGoldenFullRegistry` immediately after the registry change — if red, extend/re-capture the corpus **in the same commit**.

---

### Parity script + teeth corpus (create — no repo analog)

**Source:** RESEARCH Pattern 4 (the resolved design note for LINT-NEWRULES-03); throwaway script under /tmp whose *output* is the plan evidence. Mechanism, condensed:

1. Build both binaries: `go build -o <dir>/easyp-lab ./cmd/easyp` in the lab checkout; `go build -o <dir>/easyp-new ./cmd/easyp` in this repo.
2. Run `<bin> --format json lint` in each corpus project dir (exit 0/1 accepted; 2 = config/compile error).
3. Parse stdout JSON lines; keep only `RuleName ∈ {PACKAGE_NO_IMPORT_CYCLE, SYNTAX_SPECIFIED, FIELD_NOT_REQUIRED}`; project to `(RuleName, Path, Position.Line, Position.Column)`; `Counter`-compare lab vs new. Do NOT raw-diff full JSON lines (deliberate main-wins divergences in the 41 ported rules would poison it — RESEARCH Pitfall 6).
4. Configs need no surgery: every v1 policy seeds `use` with MINIMAL (`internal/config/v1/policy.go:104-116`, `use := []string{"MINIMAL"}` + switch), so committed corpus configs already enable the new rules on both binaries.
5. Teeth corpus (mandatory, anti-vacuity): temp project with (a) no-syntax-statement file, (b) proto2 `required` field, (c) two-package import cycle A→B→A, (d) clean negatives; assert each rule fires ≥1 finding on BOTH binaries and multisets are equal.

## Shared Patterns

### Rule contract + interface assertion
**Source:** `internal/core/lint_model.go:9-41`; repo usage `internal/rules/package_same_directory_linked.go:9`; lab files :8/:10
**Apply to:** all three new rule files
```go
var _ core.Rule = (*SyntaxSpecified)(nil)          // per-file rules
var _ core.ModuleRule = (*PackageNoImportCycle)(nil) // cross-file rule
```

### Issue construction — `core.AppendIssue` 4-arg, single construction point
**Source:** `internal/core/lint_model.go:73-80`
**Apply to:** all three rule bodies (already verbatim in the lab sources)
```go
func AppendIssue(issues []Issue, rule Rule, pos Position, sourceName string) []Issue {
	return append(issues, Issue{Position: pos, SourceName: sourceName, Message: rule.Message(), RuleName: GetRuleName(rule)})
}
```

### Rule naming — reflection only, never name constants
**Source:** `internal/core/lint_model.go:85-87` (`GetRuleName` → `toUpperSnakeCase`, `dom.go:107-123`)
**Apply to:** registry lists (`builder.go` `add*`) and any test asserting registry names. `toUpperSnakeCase("PackageNoImportCycle")` → `PACKAGE_NO_IMPORT_CYCLE` — same mechanism as all 41 existing names.

### Import-path keying
**Source:** `internal/core/dom.go:153-155` (`ImportPath(strings.Trim(source, "\""))`)
**Apply to:** the cycle rule's `ImportedFiles` lookups and its test's fixture wiring (`init_test.go:47-48` uses the same converter).

### Test harness — reuse, don't duplicate
**Source:** `internal/rules/init_test.go:14-29` (fixture const block), `:35-67` (`parseLinkedFile`), `:71-94` (`validate` dispatch)
**Apply to:** all three new `<rule>_test.go` files. Imports resolve from **repo-root `testdata/`** — fixture dir is `testdata/import_cycle/`, not `internal/rules/testdata/` (corrects RESEARCH Wave 0).

### Test style (go-testing skill)
**Sources:** `internal/rules/package_same_directory_test.go`, `internal/rules/package_defined_test.go`
**Apply to:** all new tests: `t.Parallel()` at top + subtests; map-driven (`tests := map[string]struct{…}`) or slice-with-`name`-first tables; lowercase underscore scenario names; `require` for fatal preconditions, `require.Contains` on expected `core.Issue` values; verbatim message assertions via a `Test<Rule>_Message` test.

### Schema regen gate (D7)
**Source:** AGENTS.md (`task schema:generate` / `schema:check`), Phase-4/5 precedent
**Apply to:** the atomic commit's verify chain
```sh
go run ./cmd/easyp schema-gen && git diff --exit-code schemas/
```

### Atomic-commit edit ordering (Pitfall 3/4)
**Apply to:** the single phase commit: rule files → `builder_linked.go` `New` list → `builder.go` `add*` lists → `selection.go` branch deletion → 5 test-pin flips (+ `module_rules_guard_test.go`, `lint_model.go` comments) → schema regen. Quick red-green probe set:
```sh
go test -race -count=1 ./internal/rules/ ./internal/config/v1/ ./mcp/easypconfig/ -run 'TestModuleRuleSet|TestAllLintUseValues|TestLintUseValidationParity|TestLintSelection|TestDescribeLintSelection'
```

## No Analog Found

| File | Role | Data Flow | Reason | Substitute |
|------|------|-----------|--------|-----------|
| parity compare script + teeth corpus harness | integration test | external-binary compare | No external-CLI comparison exists in-repo; lab `bench/lab.py` is the lab's Phase-7 asset and must stay untouched | RESEARCH Pattern 4 (verbatim mechanism above); throwaway script, output is the evidence |

## Metadata

**Analog search scope:** `internal/rules/` (registry, selection, harness, rule tests), `internal/core/` (contracts), `internal/protoast/` (symbol verification via RESEARCH + spot checks), `internal/config/v1/` (schema derivation, pins), `mcp/easypconfig/` (pins, schema flow), `testdata/` (fixture precedents), LAB `/tmp/gsd-research/compiler-lab/easyp-cli/internal/rules/` + `internal/config/v1/lint_selection_test.go` + `mcp/easypconfig/describe_test.go` + `builder.go` + `schemas/` (port sources and mirrors)
**Files scanned:** ~25 targeted reads/greps (stop at 5 strong analog families per RESEARCH's enumerated surfaces)
**Git-tracked verification:** all named repo analogs confirmed via `git ls-files` (internal/rules/*, internal/core/lint_model.go, internal/config/v1/*, mcp/easypconfig/*, schemas/ six files, testdata/import_used/*). Lab paths are untracked external references — flagged above.
**Pattern extraction date:** 2026-11-21
**Corrections vs RESEARCH worth carrying into the plan:**
1. Fixture dir is repo-root `testdata/import_cycle/` (harness resolves imports from `./../../testdata`), not `internal/rules/testdata/`.
2. Repo rule files carry legacy `*_linked.go` filename suffixes, but new rules follow AGENTS.md's unsuffixed `<rule>.go` convention (lab-verbatim filenames, no collisions).
3. Repo-side `package_same_directory.go` does not exist — the ModuleRule structural analog is `package_same_directory_linked.go`.
