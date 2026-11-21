---
phase: 06-novye-pravila-i-regeneratsiya-shem-mcp
reviewed: 2026-11-21T00:47:57Z
depth: deep
files_reviewed: 28
files_reviewed_list:
  - internal/rules/package_no_import_cycle.go
  - internal/rules/syntax_specified.go
  - internal/rules/field_not_required.go
  - internal/rules/selection.go
  - internal/rules/builder.go
  - internal/rules/builder_linked.go
  - internal/core/lint_model.go
  - schemas/easyp-v1.schema.json
  - schemas/easyp.schema.json
  - mcp/easypconfig/describe_test.go
  - internal/rules/builder_test.go
  - internal/rules/init_test.go
  - internal/rules/module_rules_guard_test.go
  - internal/rules/selection_test.go
  - internal/rules/validation_parity_test.go
  - internal/rules/golden_full_registry_test.go
  - internal/rules/package_no_import_cycle_test.go
  - internal/rules/syntax_specified_test.go
  - internal/rules/field_not_required_test.go
  - internal/config/v1/lint_selection_test.go
  - internal/config/v1/policy_merge_test.go
  - testdata/import_cycle/a.proto
  - testdata/import_cycle/a2.proto
  - testdata/import_cycle/b/b.proto
  - testdata/import_cycle/sp1.proto
  - testdata/import_cycle/sp2.proto
  - testdata/no_syntax/no_syntax.proto
  - testdata/required_field/required_field.proto
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
  necessity: 0
status: clean
---

# Phase 06: Code Review Report

**Reviewed:** 2026-11-21T00:47:57Z
**Depth:** deep
**Files Reviewed:** 28
**Status:** clean

## Summary

Deep review of the Phase 06 diff (e89a9b5..HEAD): three verbatim lab rule ports, the registry/validation flip, schema regeneration, MCP pin updates, golden corpus extension, and all supporting tests/fixtures. All six review priorities were independently verified against ground truth (lab checkout @ 08cef10, schema regeneration, test execution); no defects found.

**Priority-1 verbatim ports (byte-empty diff).** `internal/rules/{package_no_import_cycle,syntax_specified,field_not_required}.go` diff clean against `/tmp/gsd-research/compiler-lab/easyp-cli/internal/rules/` @ 08cef10 (diff exit 0 on all three; lab HEAD confirmed 08cef10909c0722f7042ba17b5c11951f9c6b6f6). No hand-edits slipped in.

**Priority-2 selection.go flip.** Exactly the 4-line `PACKAGE_NO_IMPORT_CYCLE` special case was deleted (`git diff` shows only that hunk); post-deletion `ValidateNames` is semantically identical to the lab's `selection.go` (same `AllLintUseValues()` Contains surface, same `ErrInvalidRule` wrapping). `fmt` remains used (lines 19/35/45). The name now validates as a plain registry member because `addMinimal` feeds it into `AllLintUseValues()`; `use`, `except`, and `ignore_only` acceptance is pinned by `selection_test.go` and `validation_parity_test.go`.

**Priority-3 registry placement.** `addMinimal` puts `PackageNoImportCycle` 5th (after `PackageSameDirectory`), `addBasic` puts `SyntaxSpecified` 21st and `FieldNotRequired` 22nd (after `ServicePascalCase`) — byte-for-byte the lab's `builder.go` group order (lab lines 224-255). `builder_linked.go` `allRules` mirrors the same order. Actual counts: 5+22+8+7+2 = 44 in both constructors; both "44" comments (`builder.go` via `lint_model.go`, `builder_linked.go`) are accurate, as is the 11-name ModuleRule guard set in `module_rules_guard_test.go`.

**Priority-4 schemas.** Each schema file gained exactly 3 enum spots × 3 names (9 added lines each; the two files are identical blobs before and after, pre-existing), with enum ordering matching `AllLintUseValues()` (`PACKAGE_NO_IMPORT_CYCLE` after `PACKAGE_SAME_DIRECTORY`; `SYNTAX_SPECIFIED`/`FIELD_NOT_REQUIRED` after `SERVICE_PASCAL_CASE`). Nothing else drifted: `go run ./cmd/easyp schema-gen` regenerates both files byte-identical (zero diff), and the enum source is the executable registry (`internal/config/v1/schema.go:73-75`), so schema and runtime cannot diverge. Gen schemas correctly untouched.

**Priority-5 MCP pin flip.** `describe_test.go` `NotContains` → `Contains` verified against the real pipeline: the test drives `Describe` over all three enum paths (`linters.enable`, `linters.disable`, `issues.exclude-rules[].linters`), and no MCP source hardcodes rule or group lists (everything derives from the registry/schema). No other stale pins exist repo-wide.

**Priority-6 golden corpus.** The 73 pre-existing `golden()` entries are byte-identical (set-diff: 0 removed, 21 added; 94 total). The 21 new findings live in 5 teeth-corpus-shaped entries (package cycle over an acyclic file DAG, no-syntax file, proto2 required field) with hand-verifiable positions: offsets/lines/columns of every new finding recompute exactly from the embedded fixtures (the +1 offset/+1 line shift vs the `testdata/` twins is the leading newline of the raw-string literals). `goldenFindingCount = 94` is asserted before the golden compare; the coverage invariant is computed dynamically over `rules.AllRuleNames()` (so "44" cannot go stale); the D-01 total-order guard still runs on engine-emitted order.

**Deep cross-file verification (beyond the six priorities).**
- Engine drain (`internal/core/lint_linked.go`): `Finish()` findings honor `IssueInfo.Path` through `shouldIgnore` and comment-ignore suppression before the D-01 sort; `collectFile` skips ignored paths, matching REQ-2.5 parity.
- Concurrency: `Collect` is invoked sequentially in walk order (after the parallel check window), so the rule's `edges` map has no data race; `-race` passes on the full-registry golden drive.
- Import-key consistency: the rule's `ImportedFiles[ConvertImportPath(imp.Location)]` lookup matches both population sites (`lint_linked_read.go:56-61,148-154`); outline files preserve package names (`protoast.go:95-100`), so cycle detection works on linked imports.
- SYNTAX_SPECIFIED false-positive check: the parser emits `DeclSyntax` for `edition` statements too (`internal/easypc/pkg/parser/parser.go:275-276`), so editions files do not fire.
- FIELD_NOT_REQUIRED coverage: `protoast.Visitor.Field` covers message/map/group/oneof/extension fields; `Label` is populated from source text (`build.go:237`); proto3 `optional` does not match `"required"`.
- SourceName parity: the parity evidence's projection omits SourceName, so this was checked separately — lab's `core.AppendIssue` (dom.go:80) has the identical 4-arg `sourceName` shape, so the cycle-string rendering matches lab.
- Interface/type consistency: `var _ core.ModuleRule` / `var _ core.Rule` assertions hold; `core.Position` is `protoast.Position` (tests compare them directly and pass).
- `go build ./...`, `go vet` (4 touched packages), `gofmt -l` (clean), `golangci-lint run internal/rules/...` (0 issues), and `go test -race -count=1` over `internal/rules`, `internal/core`, `internal/config/v1`, `mcp/easypconfig` all pass.

**Necessity:** 0 findings. All decision logic in the three rule files is a verbatim lab port under the project's byte-identical parity mandate, so no guarded state in them is assessable for removal; the repo-authored deltas (special-case deletion, registry additions, test flips) introduce no impossible-state guards, duplicated defenses, or cargo-culted patterns.

**Observation (not a finding, out of scope — planning artifact):** REQUIREMENTS.md now marks LINT-NEWRULES-03 Complete while `parity-evidence-06-03.json` records the corpora leg (api-common/googleapis) as deferred to a manual row in 06-VALIDATION.md (private-repo access gate). The deferral is transparently recorded in the evidence file itself; flagging for the end-of-phase human check awareness only.

All reviewed files meet quality standards. No issues found.
