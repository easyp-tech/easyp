# Phase 1: Вендоринг компилятора и расчленение модели - Pattern Map

**Mapped:** 2026-11-17
**Files analyzed:** 9 (new/modified)
**Analogs found:** 7 / 9 (2 are vendor-copy files with no in-repo analog by design)

All analog paths verified git-tracked (`git ls-files` non-empty).

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/easypc/**` (~48 files + testdata) | vendored library | transform (compile) | none — pure copy-in from compiler-lab (out of scope for analogs; D-04 seals it) | n/a (vendor drop) |
| `internal/protoast/**` (6 files) | vendored library + adaptation seam | transform (parse → model) | none in-repo; adaptation locus documented in RESEARCH (build.go `pos()`) | n/a (vendor drop + 1 edit) |
| `internal/protoast/position_test.go` | test (colocated, D-07a) | transform assertion | `internal/rules/file_lower_snake_case_test.go` (structure) + vendored `protoast_test.go` (subject) | role-match |
| `internal/core/position_parity_test.go` | test (golden, D-07b) | parse-compare | `internal/core/proto_info_read_test.go` | exact (same pkg, same reader under test) |
| `internal/core/testdata/position_contract_utf8.proto` (new fixture location is discretion; `internal/core/testdata/` suggested) | test fixture | file-I/O | `testdata/auth/InvalidName.proto` + `internal/rules/init_test.go` fixture-constant pattern | role-match |
| `internal/easypc/VENDORED.md` | docs/provenance | n/a | none (new doc kind); template in RESEARCH Code Examples | none |
| `.golangci.yml` (edit) | config | n/a | existing `linters.exclusions.rules` block in same file | exact (same file) |
| `AGENTS.md` (edit: D-04 prohibition) | docs | n/a | existing "Do not" section | exact (same file) |
| `internal/core/{dom.go,lint.go,proto_info_read.go,breaking_check.go,breaking_checker.go,...}` model split (D-06) | model/types refactor | request-response (engine) | `internal/core/dom.go` current type block; `generate.go` precedent for core file boundaries | role-match |
| `internal/rules/{casing.go,casing_test.go,lint_helpers.go}` (D-01) | utility (vendored helpers) | transform | `internal/rules/file_lower_snake_case.go` + `_test.go` (colocation/structure only) | role-match |

## Pattern Assignments

### `internal/core/dom.go` model split (D-06 + Position swap)

**Analog:** the file itself; lab precedent `type Position = protoast.Position`.

Current split point — `internal/core/dom.go:34-40, 48-56`:

```go
// Issue contains the information of an issue.
Issue struct {
	Position   meta.Position
	SourceName string
	Message    string
	RuleName   string
}

// ProtoInfo is the information of a proto file.
ProtoInfo struct {
	directiveComments []directiveComment
	Path              string
	ImportPath        string
	Info              *unordered.Proto
	ProtoFilesFromImport map[ImportPath]*unordered.Proto
}
```

Change mechanics (RESEARCH Pattern 2):
- `Issue.Position` → `protoast.Position` via alias `type Position = protoast.Position` placed where `meta.Position` import currently sits (dom.go:11).
- Keep the name `ProtoInfo` for the **lint**-owned type (rules untouched); create a renamed breaking-owned type (naming is planner discretion) holding the same fields for `protoInfoRead`'s current consumers.
- Single conversion point — `AppendIssue` (dom.go:117-122):

```go
// AppendIssue records a rule finding. The lint engine applies per-file suppressions.
func AppendIssue(
	issues []Issue, lintRule Rule, pos meta.Position, sourceName string, comments []*parser.Comment,
) []Issue {
	return append(issues, buildError(lintRule, pos, sourceName))
}
```
Accept `meta.Position` still (or convert at top) so the 48 `.Meta.Pos` call sites in `internal/rules` need zero changes; `buildError` (dom.go:149-156) becomes the conversion boundary (`fromMetaPos`: field-for-field copy — identical field sets `Filename/Offset/Line/Column`).

Mirrored conversion point for breaking — `internal/core/breaking_checker.go:368-377`:

```go
func buildBreakingCheckIssue(path, message string, pos meta.Position) IssueInfo {
	issue := Issue{
		Position:   pos,
		...
```
Convert `meta.Position` → `protoast.Position` here; all ~20 call sites (breaking_checker.go:383-467+) pass `.Meta.Pos` through unchanged.

### `internal/core/lint.go` lint-owned read path

**Analog:** itself, line 34:

```go
protoInfo, err := c.protoInfoRead(ctx, fsWalker, path)
if err != nil {
	return ProtoInfo{}, fmt.Errorf("c.protoInfoRead: %w", err)
}
```
Change: call the new lint-owned read function (renamed copy of `protoInfoRead` returning the lint type). Preserve the error-wrapping style verbatim (`fmt.Errorf("<callee>: %w", err)` — no package/receiver prefix, per `.spec/agent-rules.md`).

### `internal/core/proto_info_read.go` → breaking-owned

**Analog:** itself. Current shape to rename/move (lines 17-42):

```go
func (c *Core) protoInfoRead(ctx context.Context, fs FS, path string) (ProtoInfo, error) {
	f, err := c.openDiskSource(fs, path)
	...
	protoInfo := ProtoInfo{
		directiveComments:    comments,
		Path:                 path,
		Info:                 protoFile,
		ProtoFilesFromImport: protoFilesFromImport,
	}
```
Same constructor pattern (`core.New`-style: full inputs, no post-construction setters) applies to the new lint twin. Keep `readProtoFile` / `readProtoFileWithDirectives` shared or duplicated per planner ruling (D-06 says types split, not necessarily the parse helpers).

### `internal/protoast/position_test.go` (D-07a) and `internal/core/position_parity_test.go` (D-07b)

**Analogs (test conventions):**

Test structure — `internal/rules/file_lower_snake_case_test.go:1-16` (external test package, `t.Parallel()`, `require.New`):

```go
package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestFileLowerSnakeCase_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)
```

Named table-test cases with `wantIssues *core.Issue` holding `meta.Position{...}` literals (file_lower_snake_case_test.go:26-40) — this is also the exact literal shape that the type swap mechanically rewrites (`meta.Position{` → `protoast.Position{`) across 41 rule test files / 49 literals + `breaking_checker_test.go`.

Core-package test harness for the read path — `internal/core/proto_info_read_test.go` (same package `core`, uses the real reader; closest analog for the golden parity test skeleton in RESEARCH Code Examples).

Fixture access — `internal/rules/init_test.go:18,40`:

```go
invalidAuthProto3 = `./../../testdata/auth/InvalidName.proto`
...
invalidAuthProto3: parseFile(t, assert, invalidAuthProto3),
```
Repo-root `testdata/` tree (`testdata/auth/`, `testdata/breaking_check/`, ...) is the existing fixture home. Note RESEARCH Wave-0 gap: zero non-ASCII fixtures exist today; `position_contract_utf8.proto` (Cyrillic/emoji comments+strings, tab-indented decl) is genuinely new — placement in `internal/core/testdata/` (colocated, like `internal/core` package tests) or repo-root `testdata/` both fit precedent; planner picks.

### `.golangci.yml` exclusion (vendor-tree hygiene, not D-02)

**Analog:** existing exclusion block, `.golangci.yml:5-12`:

```yaml
  exclusions:
    rules:
      # Exported callee labels are required by .spec/agent-rules.md.
      ...
      - linters: [staticcheck]
        text: '^ST1005:'
        source: 'fmt\.Errorf\("[A-Z][A-Za-z0-9_]*: '
```
Add sibling key (golangci v2 syntax, empirically verified 22 → 1 findings):

```yaml
    paths:
      - internal/easypc   # sealed vendor tree (D-04); fixes go upstream
```
Do NOT exclude `internal/protoast` (maintained seam; fix its single QF1002 in build.go with the position adaptation).

### `AGENTS.md` D-04 note

**Analog:** existing "Do not" section — append one bullet forbidding edits inside `internal/easypc` (sole exception: one-time import-path rewrite at copy-in), pointing to `internal/easypc/VENDORED.md`.

### `internal/rules/{casing.go, casing_test.go, lint_helpers.go}` (D-01)

**Analog (structure only):** `internal/rules/file_lower_snake_case.go` — `package rules`, `var _ core.Rule = ...` interface assertion, doc comments on Message/Validate; tests follow the colocated external-`rules_test` pattern above. `casing.go`/`casing_test.go` vendored verbatim (stdlib-only). `lint_helpers.go` needs the 3 model-coupled signature adaptations catalogued in RESEARCH Pitfall 2 (`packagePosition`, `groupedValues.add/mismatches`, `groupedValue.pos`); bodies stay verbatim, delta documented.

## Shared Patterns

### Error wrapping (all touched Go files)
**Source:** `internal/core/lint.go:36`, `proto_info_read.go:20-31`
```go
return fmt.Errorf("readProtoFileWithDirectives: %w", err)
```
Callee name only, `%w`, no receiver/package prefix. `Open:` / `Parse:` capitalized labels are exempted from ST1005 via `.golangci.yml` source regex — keep the same shape in new read functions.

### Test conventions (all new tests)
**Source:** `.spec/agent-rules.md` + `file_lower_snake_case_test.go`
- External test package (`package rules_test`) when testing public API; same-package (`package core`) for internals like `position_parity_test.go` (needs `readProtoFileWithDirectives`).
- `t.Parallel()` first line; `require` fatal, `assert` non-fatal; named map-based table cases.
- Values (line/column numbers) must NOT change in the type swap — only the literal type name does (Pitfall 5 stop-condition).

### Position conversion (single boundary)
**Source:** `internal/core/dom.go:117-122` (`AppendIssue`) and `breaking_checker.go:368` (`buildBreakingCheckIssue`)
Both already funnel all Issue construction; convert `meta.Position` → `protoast.Position` at exactly these two functions. 48 rule call sites + ~20 breaking call sites untouched.

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `internal/easypc/**` | vendored compiler | compile/transform | Pure external copy-in; source of truth is compiler-lab main @ 08cef10 dir `easyp-cli/` (RESEARCH) — no in-repo analog applicable |
| `internal/easypc/VENDORED.md` | provenance doc | n/a | No existing provenance doc; use RESEARCH's minimal template (repo/branch/SHA/date/procedure) |
| `internal/protoast/**` adaptation | seam library | transform | New dependency tier; adaptation locus is vendored `build.go` `pos()` (lab lines 447-454, quoted in RESEARCH Pitfall 1) — change rune-column computation there only |

## Metadata

**Analog search scope:** `internal/core`, `internal/rules`, `testdata/`, `.golangci.yml`, `AGENTS.md`
**Files scanned:** ~15 (targeted reads; core model/rules/config files)
**Pattern extraction date:** 2026-11-17
