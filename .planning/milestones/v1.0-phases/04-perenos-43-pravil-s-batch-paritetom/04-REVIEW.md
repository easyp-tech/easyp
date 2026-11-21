---
phase: 04-perenos-43-pravil-s-batch-paritetom
reviewed: 2026-11-20T15:58:12Z
depth: deep
files_reviewed: 58
files_reviewed_list:
  - internal/core/lint_model.go
  - internal/rules/builder_linked.go
  - internal/rules/lint_helpers.go
  - internal/rules/comment_enum_linked.go
  - internal/rules/comment_enum_value_linked.go
  - internal/rules/comment_field_message_linked.go
  - internal/rules/comment_message_linked.go
  - internal/rules/comment_one_of_linked.go
  - internal/rules/comment_rpc_linked.go
  - internal/rules/comment_service_linked.go
  - internal/rules/directory_same_package_linked.go
  - internal/rules/enum_first_value_zero_linked.go
  - internal/rules/enum_no_allow_alias_linked.go
  - internal/rules/enum_pascal_case_linked.go
  - internal/rules/enum_value_prefix_linked.go
  - internal/rules/enum_value_upper_snake_case_linked.go
  - internal/rules/enum_zero_value_suffix_linked.go
  - internal/rules/file_lower_snake_case_linked.go
  - internal/rules/import_no_public_linked.go
  - internal/rules/import_no_weak_linked.go
  - internal/rules/import_used_linked.go
  - internal/rules/message_field_lower_snake_case_linked.go
  - internal/rules/message_pascal_case_linked.go
  - internal/rules/oneof_lower_snake_case_linked.go
  - internal/rules/package_defined_linked.go
  - internal/rules/package_directory_match_linked.go
  - internal/rules/package_lower_snake_case_linked.go
  - internal/rules/package_same_csharp_namespace_linked.go
  - internal/rules/package_same_directory_linked.go
  - internal/rules/package_same_go_package_linked.go
  - internal/rules/package_same_java_multiple_files_linked.go
  - internal/rules/package_same_java_package_linked.go
  - internal/rules/package_same_php_namespace_linked.go
  - internal/rules/package_same_ruby_package_linked.go
  - internal/rules/package_same_swift_prefix_linked.go
  - internal/rules/package_version_suffix_linked.go
  - internal/rules/rpc_no_client_streaming_linked.go
  - internal/rules/rpc_no_server_streaming_linked.go
  - internal/rules/rpc_pascal_case_linked.go
  - internal/rules/rpc_request_response_unique_linked.go
  - internal/rules/rpc_request_standard_name_linked.go
  - internal/rules/rpc_response_standard_name_linked.go
  - internal/rules/service_pascal_case_linked.go
  - internal/rules/service_suffix_linked.go
  - internal/rules/init_test.go
  - internal/rules/package_same_swift_prefix_test.go
  - internal/rules/rpc_request_response_unique_test.go
  - internal/rules/comment_enum_test.go
  - internal/rules/file_lower_snake_case_test.go
  - internal/rules/parity_harness_test.go
  - internal/rules/parity_full_registry_test.go
  - internal/rules/parity_unary_rpc_test.go
  - internal/rules/parity_wave0_test.go
  - internal/rules/registry_parity_guard_test.go
  - internal/rules/module_rules_guard_test.go
  - internal/core/position_probe_test.go
  - internal/core/lint_linked_readsource_test.go
  - .golangci.yml
findings:
  critical: 1
  warning: 4
  info: 1
  total: 6
  necessity: 0
status: issues_found
---

# Phase 04: Code Review Report

**Reviewed:** 2026-11-20T15:58:12Z
**Depth:** deep
**Files Reviewed:** 58
**Status:** issues_found

## Summary

Reviewed the Phase-4 production diff at deep depth: `internal/core/lint_model.go` (seam), `internal/rules/builder_linked.go` (registry), all 41 `<rule>_linked.go` ports, `lint_helpers.go`, and a sample of carried/parity/guard tests. Every port was diffed against its byte-frozen main counterpart (`<rule>.go`), the two parsers' models (go-protoparser `unordered.Proto` vs `protoast.File`), and the engine drain in `core/lint_linked.go`. The full `internal/rules` + `internal/core` suites pass, `go vet` is clean, `gofmt` flags one test file, all 41 `Message()` strings and regexps are byte-identical to main, every rule carries `LintRuleName()`, suppression centrality holds (no port touches directives or `ignoreOnly`), `AppendLinkedIssue` is the sole construction point (the only direct `IssueInfo{}` wrap is the ModuleRule `Finish` contract shape), registry completeness/order/group-expansion/`ignoreOnly`-unwrap parity is guarded by `TestRegistryParity`, and the ModuleRule exact set is guarded (D-03 ten).

The high-concern pins from the brief all check out: `PackageSameSwiftPrefixLinked` is the ruby_package bug-for-bug pin (double-pinned by parity cell + carried regression), `ImportUsedLinked` never reads `Resolved`, walks main's exact sections, and iterates imports in source order, `RPCRequestResponseUniqueLinked` is per-file first-exempt on written names, `FileLowerSnakeCaseLinked`/`PackageDefinedLinked` use the zero Position literal, casing regexps are main's, and no map iteration feeds output order before the engine's D-01 sort. Position discipline holds: zero `NamePos`/`NumberPos`/`FilePosition` in the ports (the two grep hits are comments explaining why those tokens are forbidden), and `TestDeclPositionProbe` locks the `Meta.Pos ≡ Decl.Pos` mapping for 11 declaration kinds.

However, deep cross-tracing found one requirement violation and three parity holes that the fixture corpus structurally cannot catch. I proved all four with scratch probe tests injected via `go test -overlay` (no repository file was modified; probe source at `/tmp/scratch-probe/zz_scratch_probe_test.go`):

1. **ignore_only changes group membership in main but not in the linked engine** — probe: old=0 findings, new=1 on a clean two-file corpus with `ignore_only` pointing at the exempt-first file. Violates REQ-2.5 (LINT-ENG-05: same suppression behavior as main). Affects all nine group ModuleRules.
2. **The parity harness never wires `IgnoreOnly` into the linked side** (`newIgnoreOnly` discarded) — the gap that let (1) through 47 green cells.
3. **Duplicate file-level option divergence** — probe: old=2, new=0 (main iterates every matching option occurrence; the port collapses via `fileOption` last-wins).
4. **Duplicate-import position divergence** — probe: old reports the last written occurrence (line 6), the port the first (line 5); the port's justifying comment claims protoc rejects duplicates, but the linked engine itself accepts such files and runs rules on them.

Necessity review: read the rubric. Zero necessity findings — the deliberately-unreachable D-01 helpers in `lint_helpers.go` are unreachable code (category 3), not unnecessary runtime guards, and their retention is a documented, gate-enforced Phase-5 decision (file header + `.golangci.yml` exclusion), so they are recorded as checked-and-accepted, not as defects.

Oracle files read for comparison (byte-frozen, not under review): the 36 old `<rule>.go` counterparts, `core/lint.go`, `core/lint_linked.go`, `core/lint_linked_read.go`, `core/dom.go`, `core/proto_comments.go`, `internal/protoast/*`, go-protoparser v4.14.2 option/lexer sources, and `internal/api/lint.go` (user-visible output formats).

## Critical Issues

### CR-01: ignore_only changes group membership in main but not in the linked engine (REQ-2.5 violation, affects all nine group ModuleRules)

**File:** `internal/core/lint_linked.go:203-212` (engine drain) with the nine group ports' Collect/Finish design, e.g. `internal/rules/package_same_go_package_linked.go:44-68`; semantics pinned as a doc claim in `internal/core/lint_model.go:50-70`

**Issue:** Main's group rules (`PackageSameGoPackage`, `PackageSameDirectory`, `DirectorySamePackage`, the six other `PACKAGE_SAME_*`) are per-file `Rule`s whose cache is seeded inside `Validate`; the old engine skips `Validate` entirely for a file listed in `ignore_only` for that rule (`core/lint.go:48`), so an ignored file never contributes its value to the group — in particular it can never be the exempt first value. The linked engine deliberately calls `Collect` for every file (`lint_linked.go:203-212`, comment: "a group stays complete even when findings for some of its files are ignored") and applies `ignore_only` only to the findings by path afterwards. When the ignored file is the exempt-first contributor, the engines disagree.

**Probe proof** (scratch test via `go test -overlay`, suite `TestScratchIgnoreOnlyGroupParityProbe`): corpus `opts/v1/alpha.proto` (go_package `"github.com/alpha/v1"`) + `opts/v1/beta.proto` (go_package `"github.com/beta/v1"`), config `PACKAGE_SAME_GO_PACKAGE` with `ignore_only: {PACKAGE_SAME_GO_PACKAGE: ["opts/v1/alpha.proto"]}`:
- old engine: **0 findings** (alpha never seeds the cache; beta becomes the exempt first value);
- linked engine: **1 finding** at `opts/v1/beta.proto:5:1` (alpha still seeds the group; beta's finding path is not ignored, so it survives the drain filter).

This violates REQUIREMENTS LINT-ENG-05 (REQ-2.5: "the system SHALL применять те же правила подавления, что и текущий main"). The divergence is invisible to the parity harness by construction (see WR-01) and will produce user-visible wrong findings the moment Phase 5 flips the CLI. `RPC_REQUEST_RESPONSE_UNIQUE` is not affected (per-file seen-set; ignoring a file removes exactly its own findings in both engines); comment directives are not affected (both engines still process the file and only filter the finding by path).

**Fix:** Mirror main's shouldIgnore-before-Validate at collection time so an ignored file never joins the group:

```go
// collectFile hands the file to module rules, in walk order — except files
// ignored for that rule, which main never lets reach Validate and therefore
// never join the group (REQ-2.5 parity; see 04-REVIEW CR-01).
func (c *Core) collectFile(info LinkedFileInfo) error {
	for _, rule := range c.linkedRules {
		if moduleRule, ok := rule.(ModuleRule); ok {
			if c.shouldIgnoreLinked(rule, info.Path) {
				continue
			}
			if err := moduleRule.Collect(info); err != nil {
				return fmt.Errorf("rule.Collect: %w", err)
			}
		}
	}
	return nil
}
```

Then rewrite the "a group stays complete even when findings for some of its files are ignored" comment in `collectFile` and the "Suppressions and ignore_only apply to the findings by path" claim in `lint_model.go` (ModuleRule doc), and add parity cells that run `assertBatchParity`-style old-vs-new comparisons with an `ignore_only` entry on a group rule (exempt-first file ignored, and separately a mismatching file ignored — the latter already agrees, pin it so it cannot regress). If the orchestrator instead rules that the lab's collect-everything semantics are the intended post-parity behavior, that must be a ledgered deferred decision plus an explicit REQ-2.5 scope amendment — not a silent doc claim.

## Warnings

### WR-01: Parity harness never wires IgnoreOnly into the linked engine — ignore_only behavior is untestable by construction

**File:** `internal/rules/parity_harness_test.go:75-82`

**Issue:** `assertBatchParity` constructs the old app with `IgnoreOnly: oldIgnoreOnly` but drops the linked constructor's second return value (`newRules, _, err := rules.NewLinked(cfg)`), so the linked app always runs with nil `ignoreOnly`. Even if a parity cell configured `ignore_only`, the two sides would not be comparing the same suppression state. This is the structural gap that let CR-01 pass 47 cumulative parity cells (and `TestRegistryParity` only proves the unwrap maps are identical pre-engine, not that the engine applies them identically).

**Fix:** Thread it through:

```go
newRules, newIgnoreOnly, err := rules.NewLinked(cfg)
require.NoError(t, err)
newApp := core.New(core.Options{
	Logger:         logger.NewNop(),
	LinkedRules:    newRules,
	IgnoreOnly:     newIgnoreOnly, // parity must include suppression state
	AllowCommentIgnores: true,
	KnownLintRules: rules.AllRuleNames(),
})
```

…or add an explicit `assertBatchParityWithIgnoreOnly(cfg, ignoreOnly, files, wantOldCount)` variant, plus cells: (a) `ignore_only` on a group rule with the exempt-first file ignored (currently red — CR-01), (b) a mismatching file ignored (green, regression pin), (c) a per-file rule ignored (green, pins per-path Validate skip).

### WR-02: Duplicate file-level option: main iterates every matching option occurrence, the port collapses to the last one (old=2, new=0 proven)

**File:** `internal/rules/package_same_go_package_linked.go:50` and the six sibling option ports (`package_same_java_package_linked.go:50`, `package_same_java_multiple_files_linked.go:50`, `package_same_csharp_namespace_linked.go:51`, `package_same_php_namespace_linked.go:50`, `package_same_ruby_package_linked.go:50`, `package_same_swift_prefix_linked.go:67`)

**Issue:** Main's Validate loops over **all** `ProtoBody.Options` with the compared name: the first occurrence seeds the per-package cache, every later occurrence (including later occurrences inside the same file) is compared against it. The ports use `fileOption(info.File, name)`, which returns the **last** matching option — losing both the first-occurrence seeding and all per-occurrence comparisons. Probe proof (`TestScratchDuplicateOptionProbe`): `opts/v1/first.proto` declaring `option go_package` twice (`"github.com/one/v1"` then `"github.com/two/v1"`) plus `opts/v1/second.proto` with `"github.com/two/v1"`: old engine reports **2** findings (first.proto:7 seeding mismatch and second.proto:5), linked engine reports **0** (last occurrence `"github.com/two/v1"` seeds, second.proto matches). The linked engine emits no COMPILE issues for this corpus — it accepts the file and runs rules on it, so "protoc rejects duplicates" is not a valid parity shield here.

**Fix:** In each of the seven option ports' `Collect`, replace `fileOption` with main's iteration shape and let `firstExemptMismatches` reproduce the first-seed/rest-compare semantics:

```go
for _, opt := range info.File.Options {
	if opt.Name != "go_package" {
		continue
	}
	p.files.add(pkg.Name, groupedValue{path: info.Path, value: opt.Value, source: opt.Value, pos: opt.Pos})
}
```

The accumulated value list then has exactly main's per-occurrence order; `firstExemptMismatches` seeds the first and reports every later differing occurrence at its own position. Add a parity cell with a duplicated option (probe corpus above) — expect old=2.

### WR-03: Duplicate import: main reports the last written occurrence, the port the first (proven position divergence; justifying comment is factually unsupported)

**File:** `internal/rules/import_used_linked.go:99-110` (report loop) and `:36-40` (comment)

**Issue:** Main builds `importInfo map[ImportPath]*parser.Import` by overwriting per occurrence, so a duplicated `import "a.proto";` reports at the **last** statement's position; the port iterates `info.File.Imports` in source order with a `reported` dedup set, reporting the **first** occurrence. Probe proof (`TestScratchDuplicateImportProbe`): `dup/v1/dup.proto` importing `"dep/v1/dep.proto"` twice — old: line 6 col 1, linked: line 5 col 1 (same count, different position → fingerprint mismatch). The port comment justifies first-occurrence with "a duplicated import statement is rejected by protoc" — that claim is unverified, and it is immaterial anyway: the linked engine itself accepts this file and runs rules on it (probe shows no COMPILE issues), so old-vs-new parity is broken on inputs both engines accept. The old engine's user-visible output for this input is the line-6 position, full stop.

**Fix:** Mirror main's last-wins selection:

```go
importInfo := make(map[core.ImportPath]*protoast.Import, len(info.File.Imports))
for _, imp := range info.File.Imports {
	importPath := core.ConvertImportPath(imp.Location)
	u.isImportUsed[importPath] = false
	importInfo[importPath] = imp // last written occurrence wins — main's overwrite
}
...
for importPath, imp := range importInfo {
	if u.isImportUsed[importPath] {
		continue
	}
	res = core.AppendLinkedIssue(res, i, imp.Pos, imp.Location)
}
```

(Map iteration order is normalized by the engine's D-01 sort; each path reports once, at its last occurrence.) Either way, correct the `:36-40` comment — the protoc-rejection premise must not stand as the only justification — and add a duplicate-import parity cell (probe corpus, expect the line-6 position on both sides).

### WR-04: Parity fingerprint omits SourceName, which is user-visible in both output formats — source drift is unprovable by the corpus

**File:** `internal/rules/parity_harness_test.go:97-109` (fingerprint) against `internal/api/lint.go:120-154` (printers)

**Issue:** `fingerprint` sorts and compares `(path, rule, line, column)` only, with an explicit "Messages are not part of the parity contract" note. But `SourceName` **is** user-visible: the text printer emits `%s:%d:%d:%s %s (%s)` (SourceName between the colons) and the JSON printer serializes `IssueInfo` whole. For these rules SourceName carries semantic payload — the written option constant, the enum/field/rpc name, the import location, the written rpc type. I verified all 41 ports pass main's exact source argument (including `opt.Value ≡ option.Constant`: both parsers keep the raw literal with quotes — go-protoparser `lexer.ReadConstant` returns the raw TSTRLIT text, `protoast.builder.constant` accumulates raw tokens), so no actual divergence exists today; but a future edit that reports, say, the unquoted constant or the enum keyword instead of the name would pass all 47 parity cells and change CLI output.

**Fix:** Add SourceName to the fingerprint tuple and to the engine's D-01 sort key so the comparison stays total:

```go
slices.SortStableFunc(out, func(a, b core.IssueInfo) int {
	return cmp.Or(
		strings.Compare(a.Path, b.Path),
		strings.Compare(a.RuleName, b.RuleName),
		cmp.Compare(a.Position.Line, b.Position.Line),
		cmp.Compare(a.Position.Column, b.Position.Column),
		strings.Compare(a.SourceName, b.SourceName),
	)
})
```

(Update the harness comment: rule, file, line, col, **source** is the parity contract; Message stays out because `Message()` equality is separately proven by the carried per-rule tests.)

## Info

### IN-01: gofmt-unformatted test file in the phase diff

**File:** `internal/rules/registry_parity_guard_test.go:89-93`

**Issue:** `gofmt -l` flags the file — the `requireConfigParity` literal uses two extra alignment spaces (`Use:                     allNames,` should be `Use:                    allNames,`). `gofmt` is a repository convention (.agents/skills/go-code-style: "Format Go code with gofmt").

**Fix:** Run `gofmt -w internal/rules/registry_parity_guard_test.go`.

## Verified-clean checklist (deep)

- **Position discipline:** zero `NamePos`/`NumberPos`/`FilePosition` in the 41 ports (2 grep hits are explanatory comments); `TestDeclPositionProbe` locks `Meta.Pos ≡ Decl.Pos` for 11 kinds; the two zero-literal rules (`PACKAGE_DEFINED`, `FILE_LOWER_SNAKE_CASE`) match main's `meta.Position{Filename: path}` exactly.
- **Suppression centrality:** no port references `directiveComment`, `isLintDirective`, `ignoreOnly`, `shouldIgnore`, or `CheckIsIgnored`; comment rules use presence-only `HasCommentAbove()`; suppression/`ignore_only` live in `core/lint_linked.go` and `core/lint_suppressions.go`.
- **main-wins pins:** ruby_package bug pin verified in source + parity cell + carried regression; `ImportUsedLinked` is stateless per call (local `importUsage` struct) and walks exactly main's sections (rpc req/resp/option names, plain field types + options, oneof member types, file-level extends, file-level options — no maps/groups/message-options/message-extends, matching the `unordered` model's separate listings); uniqueness is per-file first-exempt on written names; all 10 regexps byte-identical; `pascalToUpperSnakeLinked` byte-identical to main's helper; option compare basis `opt.Value ≡ option.Constant` verified at parser level.
- **D-01 determinism:** all rule-side orderings are slice-order (source order / insertion order); the only map fed into output (`ImportUsed` report in main) is deterministic in the port; `groupedValues` keeps an insertion-order `order` slice; the engine's final sort covers Finish findings.
- **Registry:** 41/41 in builder.go group order; `NewLinked` structurally mirrors `New` (same `validateConfigRuleNames`, `unwrapLintGroups`, `removeExcept`, `unwrapIgnoreOnly`, `defaultIfEmpty`); `TestRegistryParity` proves names-per-use-order, per-group expansion, config plumbing, ignoreOnly unwrap; `TestModuleRuleSet` pins the exact D-03 ten.
- **Issue construction:** `AppendLinkedIssue` is the only issue factory; the single direct `core.IssueInfo{...}` (rpc_request_response_unique_linked.go:97) is the ModuleRule `Finish` return shape, identical to `lint_helpers.go:151/174`.
- **Concurrency:** per-file rules are stateless (empty structs or read-only config), so the concurrent `checkFile` window drain is race-free; ModuleRules only mutate state in the serial `collectFile` drain; `init_test.go`'s `validate` helper reproduces the Collect/Finish lifecycle faithfully.
- **Dormancy:** no production caller of `NewLinked`/`LintLinked` wiring outside tests (`Options.LinkedRules` unset by `buildCore`); old path byte-frozen.
- **Tests:** full `internal/rules` + `internal/core` suites green with `-count=1`; `go vet` clean; parity cells assert old-side counts before the multiset compare (anti-vacuity) and the full-registry cell asserts every rule fires; carried tests assert exact positions/SourceNames against main's oracle literals.

---

_Reviewed: 2026-11-20T15:58:12Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: deep_
