---
phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli
reviewed: 2026-11-19T00:00:00Z
depth: standard
files_reviewed: 64
files_reviewed_list:
  - AGENTS.md
  - .golangci.yml
  - go.mod
  - internal/core/breaking_checker.go
  - internal/core/breaking_checker_test.go
  - internal/core/breaking_check.go
  - internal/core/breaking_profiles.go
  - internal/core/breaking_target_paths_test.go
  - internal/core/dom.go
  - internal/core/lint.go
  - internal/core/lint_proto_info.go
  - internal/core/position_parity_test.go
  - internal/core/proto_info_read.go
  - internal/core/proto_info_read_test.go
  - internal/core/testdata/position_contract_utf8.proto
  - internal/easypc/VENDORED.md
  - internal/protoast/build.go
  - internal/protoast/position_test.go
  - internal/protoast/protoast.go
  - internal/protoast/protoast_test.go
  - internal/rules/casing.go
  - internal/rules/casing_test.go
  - internal/rules/comment_enum_test.go
  - internal/rules/comment_enum_value_test.go
  - internal/rules/comment_field_message_test.go
  - internal/rules/comment_message_test.go
  - internal/rules/comment_one_of_test.go
  - internal/rules/comment_rpc_test.go
  - internal/rules/comment_service_test.go
  - internal/rules/directory_same_package_test.go
  - internal/rules/enum_first_value_zero_test.go
  - internal/rules/enum_no_allow_alias_test.go
  - internal/rules/enum_pascal_case_test.go
  - internal/rules/enum_value_prefix_test.go
  - internal/rules/enum_value_upper_snake_case_test.go
  - internal/rules/enum_zero_value_suffix_test.go
  - internal/rules/file_lower_snake_case_test.go
  - internal/rules/import_no_public_test.go
  - internal/rules/import_no_weak_test.go
  - internal/rules/import_used_test.go
  - internal/rules/lint_helpers.go
  - internal/rules/message_field_lower_snake_case_test.go
  - internal/rules/message_pascal_case_test.go
  - internal/rules/oneof_lower_snake_case_test.go
  - internal/rules/package_defined_test.go
  - internal/rules/package_directory_match_test.go
  - internal/rules/package_lower_snake_case_test.go
  - internal/rules/package_same_csharp_namespace_test.go
  - internal/rules/package_same_directory_test.go
  - internal/rules/package_same_go_package_test.go
  - internal/rules/package_same_java_multiple_files_test.go
  - internal/rules/package_same_java_package_test.go
  - internal/rules/package_same_php_namespace_test.go
  - internal/rules/package_same_ruby_package_test.go
  - internal/rules/package_same_swift_prefix_test.go
  - internal/rules/package_version_suffix_test.go
  - internal/rules/rpc_no_client_streaming_test.go
  - internal/rules/rpc_no_server_streaming_test.go
  - internal/rules/rpc_pascal_case_test.go
  - internal/rules/rpc_request_response_unique_test.go
  - internal/rules/rpc_request_standard_name_test.go
  - internal/rules/rpc_response_standard_name_test.go
  - internal/rules/service_pascal_case_test.go
  - internal/rules/service_suffix_test.go
findings:
  critical: 0
  warning: 3
  info: 7
  total: 10
  necessity: 1
status: issues_found
---

# Phase 01: Code Review Report

**Reviewed:** 2026-11-19T00:00:00Z
**Depth:** standard
**Files Reviewed:** 64
**Status:** issues_found

## Summary

Reviewed all 64 in-scope files at standard depth, cross-referencing their collaborators (`core.go` Core fields, `comment_suppressions.go`, `unstable_package_matcher.go`, `protoast/comments.go`/`outline.go`/`walk.go`, the vendored `go-protoparser` v4.14.2 scanner, and the rules test harness `init_test.go`), and diffed every modified file against `diff_base` `ec65bc1`.

The phase's hard invariant holds on the evidence I could check:

- **D-06 split** (`lint_proto_info.go` / `breakingProtoInfoRead` + `BreakingProtoInfo`): the git diff shows the twins are mechanical renames of the former `protoInfoRead`; the read paths in `breaking_check.go` and `lint.go` keep identical error labels and walk predicates. `TestProtoInfoReadTwinsMirroring` pins equality.
- **D-07 position conversion**: `fromMetaPos` (dom.go:169-176) copies all four fields of `meta.Position` v4.14.2 (`Filename/Offset/Line/Column`, identical types and 1-based rune-column semantics — verified against the vendored module's `lexer/scanner/position.go`, which also advances the column by 1 per rune including `\r`, so CRLF files convert losslessly too). I hand-recomputed the rune columns in `position_contract_utf8.proto` (package at 4:39 after the multi-byte string literal, `KIND_AFTER_EMOJI` at 14:26, rpc at 18:3 after two tabs) and they match `position_test.go`. `Issue.Position = protoast.Position` is a type alias, so `AppendIssue`/`buildBreakingCheckIssue` changes are value-preserving.
- **Vendoring**: `casing.go`/`casing_test.go` contain no findings that would require edits (keeping D-01 byte-verbatim intact); `lint_helpers.go` carries exactly the 3 documented seam adaptations; no findings were raised inside `internal/easypc` (sealed).
- Tests pass under `-race -count=1` for `internal/core`, `internal/protoast`, `internal/rules` (re-run during this review).

All findings below are pre-existing defects surfaced by the review (the phase diffs themselves are mechanical) except WR-03 and IN-04, which live in phase-introduced files. None are blockers; no security vulnerabilities, data-loss paths, or crashes were found (`openImportFile` correctly rejects non-local import paths with `filepath.IsLocal`, and `canonicalBreakingPath`/`breakingSourceAccessor` re-check locality before opening local files).

## Narrative Findings (AI reviewer)

## Critical Issues

None.

## Warnings

### WR-01: Vacuous invalid-case assertions — the whole `package_same_*` test family can never fail on a rule regression

**Files:**
`internal/rules/package_same_directory_test.go:63-72`,
`internal/rules/package_same_go_package_test.go:62-70`,
`internal/rules/package_same_java_multiple_files_test.go:62-69`,
`internal/rules/package_same_java_package_test.go:62-69`,
`internal/rules/package_same_php_namespace_test.go:62-69`,
`internal/rules/package_same_ruby_package_test.go:62-69`,
`internal/rules/package_same_swift_prefix_test.go:62-69`

**Issue:** In these seven tests the runner discards rule issues entirely and asserts only `r.ErrorIs(errors.Join(got...), tc.wantErr)`. The rules in this family always return `(issues, nil)` — they never produce an error (e.g. `package_same_directory.go:29-47`, `package_same_go_package.go:27-52`, `package_same_java_multiple_files.go:27-51` return `res, nil` unconditionally) — so `wantErr` is always nil and the assertion is a tautology. The `wantIssues` table field is dead in all seven files: the conflicting fixtures do produce real mismatches (`testdata/auth/queue.proto` and `testdata/invalid_pkg/queue.proto` both declare `package Queue;`; the `testdata/invalid_options/{queue,session}.proto` pair shares `package same_package;` with differing `go_package`/`java_*`/`php_namespace`/`ruby_package`/`swift_prefix`/`csharp_namespace` values), yet if the rules regressed to emit zero issues these tests would still pass. In `package_same_java_package_test.go`, `package_same_ruby_package_test.go`, and `package_same_swift_prefix_test.go` the case keys are additionally swapped: the entry named `"valid"` holds the conflicting fixtures and `"invalid"` holds the agreeing ones, further obscuring intent. This weakens exactly the safety net the Phase-4 rule port (which must preserve lint output) will rely on.

**Fix:** Assert the issues with the same switch the sibling tests (`comment_enum_test.go:80-86`, `package_same_csharp_namespace_test.go:72-78`) already use, and correct the swapped case keys:

```go
rule := rules.PackageSameGoPackage{}
var issuesAll []core.Issue
for _, fileName := range tc.fileNames {
    issues, err := rule.Validate(protos[fileName])
    r.ErrorIs(err, tc.wantErr)
    issuesAll = append(issuesAll, issues...)
}
switch {
case tc.wantIssues != nil:
    r.Contains(issuesAll, *tc.wantIssues)
default:
    r.Empty(issuesAll)
}
```

### WR-02: `shouldIgnore` prefix match has no path-separator boundary — config ignores silently over-match

**File:** `internal/core/lint.go:82-93`

**Issue:** `shouldIgnore` accepts an `ignore_only` entry when `strings.HasPrefix(path, fileOrDir)`. A rule ignore entry meant as a directory (`"auth"`) therefore also ignores `authfoo.proto` and `auth_v2/service.proto`, and an entry `"service"` ignores `service_backup.proto`. Over-matching hides lint findings — the exact failure mode an `ignore_only` escape hatch must not have. (Pre-existing code; the phase diff touched only the read-model call. Fixing it changes legacy output for anyone whose config relies on the loose match, so it must be coordinated with the phase invariant — reasonable to defer to Phase 4, but it should be tracked.)

**Fix:**

```go
for _, fileOrDir := range ignoreFilesOrDirs {
    switch {
    case fileOrDir == path:
        return true
    case strings.HasPrefix(path, fileOrDir+"/"):
        return true
    }
}
```

### WR-03: D-07b parity proof does not assert `Offset` (or `Filename`) although it claims full `Position` equality

**File:** `internal/core/position_parity_test.go:45-71`

**Issue:** The test's doc states "protoast's Position must equal the legacy go-protoparser Meta.Pos on the same multi-byte UTF-8 + tab fixture", but the anchor loop asserts only `Line` and `Column`. `Offset` parity is left unproven, while downstream tests pin exact offsets (`breaking_checker_test.go` expects `Offset: 39/65/649...`, most rules tests expect e.g. `Offset: 864/917`) — so when Phase 4 swaps engines, an offset divergence would surface as failures in unrelated tests instead of being caught by this dedicated proof. Both models use byte offsets (go-protoparser advances `Offset` by `utf8.RuneLen` per rune; `protoast` stores the byte offset directly), so the assertion should hold trivially — it just is not made.

**Fix:** Carry the whole legacy position in the anchor and assert every field:

```go
anchors := []struct {
    name    string
    legacy  meta.Position
    current protoast.Position
}{
    {"package after a multi-byte string literal", legacyBody.Packages[0].Meta.Pos, got.Packages[0].Pos},
    // ...
}
// in the loop:
assert.Equal(t, a.legacy.Offset, a.current.Offset, "byte offset")
assert.Equal(t, a.legacy.Line, a.current.Line, "line")
assert.Equal(t, a.legacy.Column, a.current.Column, "column")
```

## Info

### IN-01: Phantom parameter — `AppendIssue` accepts `comments` it never reads

**File:** `internal/core/dom.go:126-130`
**Category:** necessity (already-defended-elsewhere / phantom parameter)

**Issue:** The `comments []*parser.Comment` parameter is accepted but unused: the body only calls `buildError(lintRule, fromMetaPos(pos), sourceName)`. Every one of the ~30 production callers passes `X.Comments` (e.g. `package_same_directory.go:42`, `enum_pascal_case.go:25`) that has no effect — per-file suppression is solely owned by `commentSuppressions` applied in `core.Lint` (lint.go:39, lint.go:58), and `TestRuleDiagnosticsAreFilteredByTheEngine` confirms rules must emit even for directive-annotated files. The parameter exists as a documented Phase-4 seam for lab-ported rules (`lint_helpers.go:22-27` records that the lab's rules call a comments-consuming `AppendIssue`). Today it is generality without instantiation; whether Phase 4 actually instantiates it is a judgment call.

**Fix:** Keep only if the Phase-4 port demonstrably needs the lab signature; otherwise drop the parameter and update the ~30 call sites, recording the decision next to D-01 so the seam does not silently survive as dead API.

### IN-02: `core.CheckIsIgnored` is exported production code with no production caller

**File:** `internal/core/lint.go:107-124`

**Issue:** Repo-wide grep finds only the definition and one consumer: `internal/rules/check_lint_ignore_test.go:96` — a test in another package exercising a `core` helper. Its concern (recognizing ignore directives in declaration comments) is implemented for real by `comment_suppressions.go` (`parseDirectivePrefix`). This is dead production surface plus misplaced test coverage: the rules package suite silently owns the only coverage of an exported core function, so deleting `CheckIsIgnored` would break `internal/rules` tests, not `internal/core` tests.

**Fix:** Either delete `CheckIsIgnored` and move `TestIsIgnored` coverage onto `commentSuppressions`/`parseDirectivePrefix` in `internal/core`, or (if the lab API parity requires keeping it) wire a consumer for it in Phase 4.

### IN-03: Two divergent `toUpperSnakeCase` implementations in the same module

**Files:** `internal/core/dom.go:138-154`, `internal/rules/casing.go:30-54`

**Issue:** `GetRuleName` uses `dom.go`'s hand-rolled byte-indexed converter, while the vendored buf-port `rules.toUpperSnakeCase` (built on `toSnakeCase`) implements the same PascalCase→UPPER_SNAKE_CASE conversion with a different algorithm (delimiter-aware, rune-trimmed). They agree on the identifier corpus today, but nothing pins their agreement — when Phase 4 wires the vendored casing helpers into the rules, a future edit to one implementation can silently diverge rule names (which are user-facing config keys in `ignore_only` and `disable`) from casing behavior.

**Fix:** Not a change to `casing.go` (byte-verbatim, D-01). Add a small parity test asserting `core.toUpperSnakeCase(name) == rules.toUpperSnakeCase(name)` for the full registered rule-name corpus, or route `GetRuleName` through one documented owner after Phase 4 lands.

### IN-04: `decl()` computes `span.End - 1` without the empty/synthetic guard its caller `register()` has

**File:** `internal/protoast/build.go:96-98` (guard at 102-111; clamp at 453-457)

**Issue:** `register()` explicitly skips `span.IsEmpty() || b.syn.IsSynthetic(span)` spans, acknowledging such spans occur, but `decl()` runs first for the same spans and computes `End: b.pos(span.End - 1)` unguarded. An empty span makes `span.End - 1` underflow to `0xFFFFFFFF`; `pos()` then silently clamps it to `SrcLen`, so the declaration is emitted with a garbage `End` at end-of-file instead of being rejected or zeroed. Latent only for synthetic/error-recovery spans that still reach the typed builders, hence Info, but the asymmetry with `register()` is the smell.

**Fix:**

```go
func (b *builder) decl(span, _ parser.TextSpan) Decl {
	if span.IsEmpty() {
		return Decl{}
	}
	return Decl{Pos: b.pos(span.Start), End: b.pos(span.End - 1)}
}
```

### IN-05: `directory_same_package_test.go` valid branch dereferences a nil `wantIssues` on regression

**File:** `internal/rules/directory_same_package_test.go:69-71`

**Issue:** The runner does `if len(res) > 0 { r.Contains(res, *tc.wantIssues) }`. In the `"valid"` case `tc.wantIssues` is nil, so if the valid fixture pair ever regresses to producing issues, the test panics with a nil-pointer dereference instead of failing with a useful diff. (The `"invalid"` case here does assert `wantIssues`, so unlike WR-01 the assertion is not wholly dead — only the failure mode on the valid side is defective.)

**Fix:** Mirror the standard switch:

```go
switch {
case tc.wantIssues != nil:
    r.Contains(res, *tc.wantIssues)
default:
    r.Empty(res)
}
```

### IN-06: Unparenthesized `&&`/`||` precedence in the enum-value comparison reads as two different rules

**File:** `internal/core/breaking_profiles.go:680`

**Issue:** `matchingNumber != nil && matchingNumber.Name() != oldPair.name && profile.source || matchingNumber != nil && matchingNumber.Name() != oldPair.name && profile.wireJSON` relies on `&&` binding tighter than `||`. It evaluates correctly to `cond && (profile.source || profile.wireJSON)`, but the duplicated subexpression and unparenthesized mix make a future edit likely to change semantics unnoticed.

**Fix:**

```go
if matchingNumber != nil && matchingNumber.Name() != oldPair.name && (profile.source || profile.wireJSON) {
```

### IN-07: Legacy `BreakingChecker.Check()` emits issues in map-iteration order — nondeterministic CLI output

**Files:** `internal/core/breaking_checker.go:23-27`, `internal/core/breaking_checker.go:32-61`

**Issue:** `Check()` ranges over `b.against` (a map) and each `Collection`'s maps without sorting, so the issue order (and therefore the printed `easyp breaking-check` output) varies run to run. The profiled engine sorts its output (`deduplicateBreakingIssues`, breaking_profiles.go:1041-1067) but the legacy path does not; tests compensate with `require.ElementsMatch` (breaking_checker_test.go:158), which confirms rather than fixes the nondeterminism. Pre-existing legacy behavior — flagging so the eventual engine consolidation inherits the sorted contract.

**Fix:** Sort `res` by (Path, Position.Line, Position.Column, RuleName, Message) before returning, mirroring `deduplicateBreakingIssues`' comparator; coordinate with the behavior-invariant constraint since output order is observable.

---

_Vendoring constraints respected: no findings require edits inside `internal/rules/casing.go` (D-01 byte-verbatim), `internal/rules/lint_helpers.go` (3 documented seams, verified to match its header), or `internal/easypc` (sealed; only integration seams were checked — `protoast` imports only `easypc/pkg/parser` and `easypc/protoc/syntax`, both declared adaptation surfaces)._

_Reviewed: 2026-11-19T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
