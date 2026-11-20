# Phase 5: Переключение lint-пути и ворота совместимости - Research

**Researched:** 2026-11-20
**Domain:** Go CLI engine-flip — buildCore seam rewiring, old-path deletion, rename-map consolidation, compatibility gates (output format / paths / exit codes / sibling subsystems) in EasyP `internal/{api,core,rules}`
**Confidence:** HIGH (every load-bearing fact read from source this session; all paths and line ranges verified)

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| LINT-COMPAT-01 (REQ-5.1) | `easyp lint` выдаёт замечания в неизменном формате: пути, формат issues и коды выхода CLI совпадают с main до переключения, включая новые `COMPILE` в существующем формате | Compat-gate design (Architecture Pattern 4): golden fixture corpus captured from the byte-frozen old engine pre-flip, asserted post-flip; Action-path matrix extension adds the class-1 (COMPILE) cells through `actionV1` + subprocess exit codes (completes Phase-2 flagged assumption A1); text/JSON printer goldens already pinned pre-flip in `TestPrintIssuesCompileGolden` [VERIFIED: internal/api/lint_link_matrix_test.go:246-326] |
| LINT-COMPAT-02 (REQ-5.2) | breaking, generate, `mod tidy` и `migrate` работают как до изменения; `go-protoparser` и `protocompile` остаются в `go.mod` | Deletion inventory keeps breaking-owned files untouched (`proto_info_read.go`, `proto_comments.go` — zero-edit constraint); go-protoparser production users survive the deletion (breaking path + `collectDirectiveComments` + dom.go model types + `internal/modules/import_roots_imports.go`) [VERIFIED: grep, see Architecture Pattern 2]; protocompile users are outside the lint path entirely (`generate.go`, `breaking_profiles.go`, `migration`, `modules`, `version`) [VERIFIED: grep]; SC3 same-run gate = scoped `-race` suite (wazero carve-out documented) |
</phase_requirements>

## User Constraints (from CONTEXT.md)

No CONTEXT.md exists for Phase 5 (`has_context: false` per init). No locked discuss-phase decisions constrain this phase. The ROADMAP Design note ("решение «мост vs одно переключение», принятое совместно с Phase 3") is resolved by the Phase-3/4 architecture itself: there is **no bridge** — the flip is a single seam change backed by the Phase-4 batch parity proofs.

## Project Constraints (from AGENTS.md + project skills)

- **Sealed tree:** no edits inside `internal/easypc` (compiler fixes go upstream; re-sync only). A natural-warn fixture still cannot be produced for the same reason (Phase-2 D-04).
- **Do not hand-edit generated schema JSON** (`schemas/*.json`); regenerate via `go run ./cmd/easyp schema-gen` and gate with `task schema:check` equivalent. Rule names do not change in Phase 5 → drift must be empty.
- **Colocated tests** (`<rule>_test.go` next to `<rule>.go`), testify, `-race` mandatory on behavior tests; named table cases; subprocess tests sequential (process-state exception — existing pattern in `TestLintLinkMatrixExitCodes`).
- **Error style:** wrap with callee-only label + `%w` (`fmt.Errorf("Open: %w", err)`); `errors.Is`/`errors.As` for identity.
- **No parallel agent instruction files; do not commit generated noise / coverage.out / local `easyp` binary.**
- **Git workflow:** config uses `branching_strategy: none` + `allow_default_branch_commits: true` — commits continue on the current branch (Phase-2 ratified mode).
- **Schema/MCP:** no new rule names in this phase (Phase 6 owns that) — `internal/config/v1/schema.go:73-75` reads `rules.AllLintUseValues()` [VERIFIED], which survives the deletion, so schema drift is impossible but must be gated.

## Summary

Phase 5 is the milestone's single user-visible change point: `easyp lint` stops calling the old sequential go-protoparser engine and runs the linked compiler engine that Phases 2–4 built and proved dormant. The flip seam is `internal/api/runtime.go buildCore` [VERIFIED: internal/api/runtime.go:62-88] — the **only** production constructor of a lint Core (all other `core.New` callers are breaking [VERIFIED: internal/api/breaking_v1.go:146,463] and generation [VERIFIED: internal/generation/module.go:53], neither of which sets `Options.Rules`), and the **only** production caller of an engine entrypoint is `internal/api/lint_v1.go:205` (`apps[key].Lint(...)`) [VERIFIED]. Phases 2–4 deliberately shaped the tree so that the flip can be realized **without touching `internal/api` at all**: the rename-map consolidation (promised by the headers of `internal/core/lint_model.go:9-33` and `internal/rules/builder_linked.go:20-54`) renames `NewLinked → New`, `LintLinked → Lint`, `LinkedRules → Rules`, `LintRule → Rule`, `LinkedFileInfo → ProtoInfo` in the same commit that deletes the old contract — after which `buildCore` and `actionV1` are textually unchanged and the seam flips because the names now bind to the linked contract. The alternative (minimal flip: buildCore switches to `rules.NewLinked` + `Options.LinkedRules`, `lint_v1.go` calls `.LintLinked`) leaves permanent `Linked*` names and stale rename-map headers and makes the flip two edits in two api files; it is documented as the fallback shape.

The deletion inventory is larger than "the old lint files" but precisely bounded. In `internal/core`: `lint.go` (old engine; but `close` moves out — breaking and the linked read path call it), `lint_proto_info.go` (old reader), the old walk pieces of `comment_suppressions.go` (`commentSuppressions` method + `parsedDirective`), and the lint-owned symbols of `dom.go` (old `Rule`, `ProtoInfo`, 5-arg `AppendIssue`, `buildError`, `GetRuleName` — the last collapses with `LinkedRuleName`). In `internal/rules`: 41 old rule files, `casing.go` + `casing_test.go`, the old `New` in `builder.go` (the file itself SURVIVES trimmed — `AllGroups`/`AllRuleNames`/`AllLintUseValues`/group lists/unwrap helpers feed api, config schemas, and migration [VERIFIED: internal/api/runtime.go:74, internal/api/lint_v1.go:263, internal/config/v1/schema.go:73-75, internal/migration/convert.go:167]), the eight D-01-kept helpers + `groupedValues.mismatches` in `lint_helpers.go` (with the `.golangci.yml:24` paths exclusion removed in the same commit). Test-side: the entire parity harness suite (8 parents, 53 cells) dies with its oracle — the old engine — and is replaced by golden literals; `check_lint_ignore_test.go` dies with `core.CheckIsIgnored`; `init_test.go` is trimmed (`start`/`parseFile` old-model helpers die, `parseLinkedFile`/`validate` survive and feed all 41 carried tests).

Four compat gates prove SC1/SC3: (1) a golden fixture corpus — the Phase-4 full-registry corpus (13 files / 9 packages, 73 findings, all 41 rules firing) upgraded from fingerprint parity to a full byte-total golden including messages and D-01 order, plus the 41 carried per-rule tests and the suppression matrix cells; (2) the Action-path matrix extended with the class-1 COMPILE cells through `actionV1` and subprocess exit codes — executing Phase-2's flagged assumption A1; (3) sibling-subsystem suites (breaking/generate/tidy/migrate) in the same run with the documented wazero `-race` carve-out; (4) mechanical gates (`go mod tidy` diff-empty, schema-gen drift-empty, golangci 0 issues, build, `easyp lint --help` exit 0). The Phase-2 deferred findings land here with concrete dispositions: WR-01 and WR-02 are **fix-now** (they are exactly the error-path/file-selection divergences the compat gate would otherwise catch), IN-01 is fix-now (mechanical), IN-02 is decide-and-pin (COMPILE bypasses suppression — keep, record, pin).

**Primary recommendation:** One atomic flip commit — WR-01/WR-02/IN-01 fixes land pre-flip on the dormant linked path, goldens are captured from the frozen old engine pre-flip, then the flip commit deletes the old path, collapses the rename map (except the three documented breaking-coexistence exceptions), and re-runs all gates; a follow-up mechanical commit renames the 41 port types to canonical names and drops the now-redundant `LintRuleName()` overrides. Rollback = revert the flip commit.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Lint engine selection (the flip) | `internal/api` (`buildCore`, runtime.go:62-88) | `internal/core` (Core fields) | buildCore is the only production lint-Core constructor; post-consolidation its source text is unchanged and the seam flips through renamed bindings |
| Lint engine execution | `internal/core` (renamed `Lint` = today's `LintLinked`, lint_linked.go:25) | `internal/parallel`, `internal/easypc/protoc`, `internal/protoast` | The linked engine IS the engine; old `core.Lint` (lint.go:17-77) deleted |
| Rule registry + config validation | `internal/rules` (trimmed `builder.go` + `builder_linked.go` → renamed `New`) | `internal/core` (`ErrInvalidRule`, name derivation) | D-06 single-validator surface (`validateConfigRuleNames`, selection.go) survives unchanged; registry feeds schemas/MCP/api exclusions |
| Error classification (COMPILE / fallback / read) | `internal/core/link.go` (`linkSources`) | `internal/easypc/protoc` | WR-01/WR-02/IN-01 fixes land here pre-flip; IN-02 decision recorded here |
| Breaking / generate / mod tidy / migrate | their existing packages — UNTOUCHED | — | REQ-5.2; verified zero lint-path symbols leak into their files |
| Output printing / exit codes | `internal/api` (`lint.go` printers, `Action`) | — | Untouched by the flip; gates assert byte-identical behavior through it |

## Standard Stack

### Core

All stack items are **in-repo Go packages** — no new external dependencies, nothing to install (REQ-1.2 long enforced; REQ-5.2 keeps the existing external modules).

| Component | Version/Location | Purpose | Why Standard |
|-----------|------------------|---------|--------------|
| Go toolchain | go.mod `go 1.26.6`, installed `go1.27.3` | Language | Pinned by go.mod; verified this session (`go version` → go1.27.3) |
| `internal/core` linked engine | `lint_linked.go` (242 lines), `lint_model.go`, `lint_linked_read.go`, `lint_suppressions.go`, `link.go` | THE engine post-flip | 41/41 rules parity-proven old-vs-new in Phase 4; suppression/main-wins parity proven Phase 3 |
| `internal/rules` linked registry | `builder_linked.go` `NewLinked` + 41 `*_linked.go` | THE registry post-flip | Registry-parity guard proved name/group/config/ignoreOnly parity with `New` (Phase 4) |
| `internal/api` lint command | `lint.go`, `lint_v1.go`, `lint_batch.go` | CLI surface — untouched | Exit codes 0/1/2 and printers pinned by the Phase-2 matrix |
| `slices.SortStableFunc` + `cmp.Or` | Go stdlib | D-01 five-term deterministic sort (lint_linked.go:152-160) | Stable sort + first-non-zero comparison — the engine's output-order contract [VERIFIED: internal/core/lint_linked.go:152-160; CITED: pkg.go.dev slices/cmp] |
| testify | go.mod | Test assertions | Project convention |
| Golden-master practice | method, not code | Compat gate shape | Capture old engine's observable output (stdout, exit codes, issue sets) as committed goldens, swap the implementation behind the seam, require identity, then delete the legacy path [CITED: thedignifiedcatonline/blog golden-master; endoflineblog characterization testing — web-confirmed MEDIUM] |

### Supporting

| Component | Purpose | When to Use |
|-----------|---------|-------------|
| `buildEasypBinary` `sync.OnceValues` pattern (internal/api/lint_link_matrix_test.go:131-149) | Subprocess gate: build once per run, assert real exit codes 0/1/2 | Reuse as-is for the new COMPILE class-1 subprocess cell and the stdout golden test |
| `writeLinkMatrixFixture` + `runLinkMatrixAction` (same file, :21-47) | In-process actionV1 driving over per-case temp fixtures | Extend with the class-1 in-process cell |
| `internal/fs/fs.FSWalker` (`RootPath()`, internal/fs/fs/adapter.go:39) | Engine fixtures; RootPath feeds `importCacheKey` and `importFileAllowed` | Unchanged |
| Phase-4 parity corpus (in `parity_full_registry_test.go`) | The golden corpus (13 files / 9 packages, 73 findings, 41-rule coverage assertion) | Promote to the flip's golden test before the oracle is deleted |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Consolidation flip (rename map collapses; api textually unchanged) | Minimal flip (`buildCore` → `NewLinked`/`LinkedRules`, `lint_v1.go` → `.LintLinked`, keep Linked names) | Consolidation is what the rename-map headers promise (lint_model.go:9-33, builder_linked.go:20-54); leaves zero stale names; api untouched. Minimal flip is a smaller rename sweep but strands permanent `Linked*` debt and stale headers, and SC2's "врезка — одно изменение в шве buildCore" then spans two api files |
| Full collapse incl. 41 port-type renames + override drops | Leave `<Name>Linked` type names + `LintRuleName()` overrides forever | Renames are behavior-neutral (reflection fallback derives the same canonical names — proven by `TestRegistryParity` name-set equality between both registries); leaving suffixes contradicts "старые реализации правил удалены" spirit and keeps 41 override methods. Recommended as a separate mechanical commit after the flip |
| Golden corpus as committed `testdata/` tree | Keep corpus as in-test string maps (Phase-4 pattern) | In-test follows the established batch pattern, keeps the corpus next to its assertions; `testdata/` tree only if the corpus outgrows string maps |
| Keep parity harness alive by keeping the old engine | — | REJECTED: SC2 mandates deletion; the harness cannot compile without its oracle |

**Installation:** none — `go.mod` unchanged. `go-protoparser v4.14.2` and `bufbuild/protocompile v0.14.1` REMAIN (verified users survive; see Pattern 2).

## Package Legitimacy Audit

Not applicable: this phase installs **zero external packages** (pure in-repo Go deletion/renaming work). No registry lookups, no `checkpoint:human-verify` installs.

## Architecture Patterns

### System Architecture Diagram

```
                        easyp lint  (CLI surface — flags, printers, exit codes UNCHANGED)
                             │
                    actionV1 (lint_v1.go:205)      ← textually unchanged
                             │  apps[key].Lint(...)
                             ▼
   ┌────────────────────────────────────────────────────────────────────┐
   │ buildCore (runtime.go:62-88)  ← THE SEAM (textually unchanged)     │
   │   rules.New(cfg)      = was old registry, now renamed NewLinked    │
   │   Options.Rules       = was []oldRule, now the linked contract     │
   └────────────────────────────────────────────────────────────────────┘
                             │
                             ▼
   ┌──────────────────────── THE ENGINE (ex LintLinked) ────────────────┐
   │ walk paths (ignore filter, .proto)                                 │
   │ pass 1: parallel readSource → linkSources classification:          │
   │   ├─ linked trees            → models from Linked.Syntax (windows) │
   │   ├─ COMPILE issues          → returned as issues (rules skipped)  │
   │   └─ (nil,nil) fallback      → parseOnTheirOwn per file            │
   │ pass 2: windowed parallel model build + per-file rules;            │
   │         serial drain in walk order; ModuleRule Collect → Finish    │
   │ central suppressions (shouldIgnore + comment ranges);              │
   │ D-01 total sort (path, rule, line, col, source)                    │
   └────────────────────────────────────────────────────────────────────┘
                             │
                             ▼
        printIssues (text/JSON) → exit 0 / exit 1 (ErrHasLintIssue) / exit 2 (OpenImportFileError)
                             ▲
        DELETED PATH: old core.Lint (go-protoparser walk) + 41 old rules
                      + old registry New + old suppression walk  ✗
```

### Pattern 1: The flip seam — what buildCore must construct post-flip

**Verified seam facts:**

- `buildCore` (internal/api/runtime.go:62-88) is the only production constructor that populates lint rules: `lintRules, ignoreOnly, err := rules.New(cfg.Lint)` → `core.Options{Rules: lintRules, Ignore: ..., IgnoreOnly: ignoreOnly, AllowCommentIgnores: ..., KnownLintRules: rules.AllRuleNames(), Logger, ImportRoots, ImportFileAllowed, OpenSourceFile, CurrentProjectGitWalker, BreakingCheckConfig}` [VERIFIED: read this session].
- The only production engine call: `apps[key].Lint(ctx.Context, newSelectedLintWalker(lintRoot, batchFiles[key]))` [VERIFIED: internal/api/lint_v1.go:205].
- Other `core.New` callers set no Rules: `breaking_v1.go:146,463` (breaking options only) and `internal/generation/module.go:53` + `config.go` (plugins/inputs/managed mode only) [VERIFIED: grep + read].
- `core.Options.LinkedRules` → `Core.linkedRules` is populated only from tests today [VERIFIED: internal/core/core.go:69-72,101-105 + 04-07 ledger item 15].

**Post-flip construction (consolidation shape — recommended):** buildCore's source text is **identical**; every binding behind it changes identity in the flip commit:

| Binding | Pre-flip meaning | Post-flip meaning |
|---------|------------------|-------------------|
| `rules.New(cfg)` | old registry constructor (builder.go:61) | renamed `NewLinked` body (builder_linked.go:70) — same signature `(config.LintConfig) ([]core.Rule, map[string][]string, error)` after `LintRule→Rule` |
| `Options.Rules` | `[]core.Rule` (old contract) | re-typed `[]core.Rule` where `Rule` = etalon interface (`Message` + `Validate(ProtoInfo-protoast)`) |
| `apps[key].Lint` | old sequential engine | renamed `LintLinked` (lint_linked.go:25) |

`core.Core` keeps ONE rule field (renamed `linkedRules → rules`, `[]Rule`), and the dormant `Options.LinkedRules` field is deleted. The suppression known-rules fallback in `lint_suppressions.go:54-58` reads `c.rules` + `GetRuleName` — it compiles after the same rename (hidden cross-reference; see Pitfall 5).

**What buildCore must construct post-flip — nothing new.** All eleven fields it already sets survive with the same names and sources. The flip adds no options, no flags, no config surface. `defaultVendorDir` ignore-append stays. `KnownLintRules: rules.AllRuleNames()` stays (feeds comment-directive validation in the linked engine, lint_suppressions.go:51-58).

**Minimal-flip fallback shape (if the planner rejects the consolidation):** `buildCore` changes to `rules.NewLinked(cfg)` + `Options.LinkedRules: lintRules`; `lint_v1.go:205` changes to `.LintLinked(...)`; old path deleted in the same commit; `Linked*` names become permanent; the rename-map headers in lint_model.go/builder_linked.go must be rewritten (they promise Phase-5 consolidation). Two api-file edits instead of zero.

### Pattern 2: The deletion inventory (files/functions/fields — what dies, what stays and why)

**internal/core — deletions:**

| File | Action | Detail |
|------|--------|--------|
| `lint.go` (125 lines) | DELETE + move 1 symbol | Contents: old `Lint` (:17-77), old `shouldIgnore` (:79-93), `close` (:95-105), `CheckIsIgnored` (:107-124). `close` MOVES to a surviving file — it is called by `breakingProtoInfoRead` (proto_info_read.go), `readSource`/`linkedFileFromImport` (lint_linked_read.go:19,107), `lintResolver.FindFileByPath` (link.go:139) [VERIFIED: grep]. `CheckIsIgnored`'s only user is `internal/rules/check_lint_ignore_test.go:96` [VERIFIED: grep] — that test dies too. `shouldIgnore` (old, Rule-typed) dies; `shouldIgnoreLinked` collapses into its name |
| `lint_proto_info.go` (30 lines) | DELETE whole | `lintProtoInfoRead` — the old lint's read-model twin; the linked engine never calls it (its `fs.WalkDir: c.lintProtoInfoRead:` wrapper at lint_linked.go:116 is a byte-literal parity pin, not a call) [VERIFIED: lint_linked.go:112-116 comment] |
| `comment_suppressions.go` | TRIM | DELETE: `commentSuppressions` method (:41-114, the old reflection-fed walk) + `parsedDirective` type (:33-37). KEEP: `suppressionRange`/`suppressionRanges`/`contains` (:13-26, shared with the linked engine), `directiveComment` (:28-31, *parser.Comment — used by `proto_comments.go` + `BreakingProtoInfo.directiveComments`), `parseDirectivePrefix` (:126-143), `compactDirectiveNames` (:116-124), `collectDirectiveComments` (:147-231 — called by `readProtoFileWithDirectives` in proto_comments.go:21, which the breaking read path uses) [VERIFIED: read both files in full] |
| `dom.go` | TRIM | DELETE: old `Rule` interface (:24-29), `ProtoInfo` struct (:57-64 — verified zero surviving users outside the old lint path and old rules; breaking uses `BreakingProtoInfo` [VERIFIED: grep `\bProtoInfo\b`]), `AppendIssue` 5-arg (:126-130), `buildError` (:157-164 — only AppendIssue calls it [VERIFIED: grep]), `GetRuleName` (:133-135 — collapses with `LinkedRuleName`). KEEP: `Position` alias, `Issue`, `IssueInfo`, `ImportPath`, `PackageName`, `Import`/`Service`/`Message`/`OneOf`/`Enum` wrappers, `Collection`, `ProtoData`, `GetPackageName` (all used by breaking_checker/breaking_check/unstable_package_matcher [VERIFIED: grep]), `fromMetaPos` (:169-176 — used by breaking_checker.go:370 [VERIFIED: grep] — NOT deletable), `OpenImportFileError`, `GitRefNotFoundError`, `ConvertImportPath`, Plugin/Input types |
| `core.go` | TRIM | DELETE `Options.LinkedRules` field (:72-73) + dormancy comments; the `rules`/`linkedRules` pair collapses to one `rules []Rule` field (:19-23) wired from `Options.Rules` |

**internal/core — survives untouched or renamed:**

| File | Action |
|------|--------|
| `lint_linked.go` | KEEP; method `LintLinked → Lint`; `shouldIgnoreLinked → shouldIgnore`; `c.linkedRules → c.rules`; the main-wins wrapper literal `fs.WalkDir: c.lintProtoInfoRead:` (:116) stays byte-identical (it matches main's observable text, lint.go:36) |
| `lint_model.go` | KEEP; `LinkedFileInfo → ProtoInfo`, `LintRule → Rule`, `AppendLinkedIssue → AppendIssue` (4-arg — re-typed; may reuse a re-typed `buildError`), `LinkedRuleName → GetRuleName` (+ keep the optional-name override mechanism or drop it with the type renames — see Pattern 3); the rename-map header comment (:9-33) is DELETED once executed |
| `lint_linked_read.go` | KEEP; `linkedFilesFromImport`/`linkedFileFromImport` KEEP their names (see Pattern 3 exceptions); `readSource`, `protoInfoFromSource`, `protoInfoFromLinked`, `importCacheKey`, `cacheLintedOutline`, `parseOnTheirOwn` unchanged |
| `lint_suppressions.go` | KEEP; `linkedCommentSuppressions → commentSuppressions`, `linkedParsedDirective → parsedDirective`; `linkedDirectiveComment` KEEPS its name (exception); `directiveComments` walk unchanged |
| `link.go` | KEEP whole (linkSources, lintResolver, resolution, usesThrough, linkedFiles, diagCollector, hasPerFileSyntaxError, `CompileRuleName`); gains WR-01/WR-02/IN-01 fixes pre-flip |
| `proto_info_read.go`, `proto_comments.go` | KEEP **byte-frozen** — breaking-owned (breakingProtoInfoRead, readProtoFileWithDirectives, readProtoFile, readFilesFromImport, readFileFromImport, openImportFile, sourceImportPath, openDiskSource). The flip must not edit these files at all (gate: `git diff --exit-code` on both) |
| `fs.go`, `path_helpers/`, breaking_*/generate_*/managed_* files | KEEP untouched |

**internal/rules — deletions:**

| Target | Detail |
|--------|--------|
| 41 old rule files | `comment_enum.go`, `comment_enum_value.go`, `comment_field_message.go`, `comment_message.go`, `comment_one_of.go`, `comment_rpc.go`, `comment_service.go`, `directory_same_package.go`, `enum_first_value_zero.go`, `enum_no_allow_alias.go`, `enum_pascal_case.go`, `enum_value_prefix.go` (owns `pascalToUpperSnake`), `enum_value_upper_snake_case.go`, `enum_zero_value_suffix.go`, `file_lower_snake_case.go`, `import_no_public.go`, `import_no_weak.go`, `import_used.go` (owns `existInProto`), `message_field_lower_snake_case.go`, `message_pascal_case.go`, `oneof_lower_snake_case.go`, `package_defined.go`, `package_directory_match.go`, `package_lower_snake_case.go`, `package_same_csharp_namespace.go`, `package_same_directory.go`, `package_same_go_package.go`, `package_same_java_multiple_files.go`, `package_same_java_package.go`, `package_same_php_namespace.go`, `package_same_ruby_package.go`, `package_same_swift_prefix.go`, `package_version_suffix.go`, `rpc_no_client_streaming.go`, `rpc_no_server_streaming.go`, `rpc_pascal_case.go`, `rpc_request_response_unique.go`, `rpc_request_standard_name.go`, `rpc_response_standard_name.go`, `service_pascal_case.go`, `service_suffix.go` [VERIFIED: ls — 45 non-linked non-test files minus builder/casing/lint_helpers/selection] |
| `casing.go` + `casing_test.go` | DELETE — no surviving consumer: zero references from `*_linked.go` or any other package (core's `managed_mode.go` has its own identically-named helpers in package core) [VERIFIED: grep]. Ledger item 5 obligation |
| `builder.go` | TRIM, NOT delete. DELETE: `New` (:61-142, the old constructor). KEEP: `RuleGroup` + group constants (:14-21,35-52), `AllGroups`/`AllRuleNames`/`AllLintUseValues` (:35-59 — used by api/runtime.go:74, api/lint_v1.go:263, config/v1/schema.go:73-75, policy_merge.go:190, migration/convert.go:167), `unwrapIgnoreOnly`/`unwrapLintGroups`/`removeExcept`/`defaultIfEmpty` (:145-212 — used by builder_linked.go:109-142), `addMinimal/addBasic/addDefault/addComments/addUnary` (:220-276 — used by AllGroups) [VERIFIED: read builder.go + grep users] |
| `lint_helpers.go` | TRIM. DELETE the eight D-01-kept helpers (`optionString`, `enumNumber`, `leadingCommentDocuments`, `isLintDirective`, `packagePosition`, `mismatches`, `packageHasVersionSuffix`, `versionNumber`) + the unreachable report-everyone `groupedValues.mismatches` (:144-169) — ledger items 6+14. KEEP `fileOption` (:56), `groupedValues` + `add` (:120-142), `firstExemptMismatches` (:171) — consumed by the linked ports. DELETE the `.golangci.yml` paths exclusion (line 24) in the same commit — the exclusion's comment (lines 16-23) documents exactly this obligation [VERIFIED: .golangci.yml:16-24] |
| `selection.go` | KEEP whole (`ValidateNames`, `validateConfigRuleNames` — the D-06 single validator shared by both constructors) |

**Test files — die with the oracle (cannot compile post-flip):**

| File | Why |
|------|-----|
| `internal/rules/parity_harness_test.go` | `assertBatchParity` drives `rules.New` + `Options.Rules` + `oldApp.Lint` [VERIFIED: parity_harness_test.go:61-99] — the oracle is deleted |
| 8 parity parents: `parity_wave0/minimal/basic_casing/basic_semantic/default/comments/unary_rpc/full_registry_test.go` (+ `parity_suppression_test.go` — 9 total) | All call `assertBatchParity`/`requirePorted`/`oldFindings` [VERIFIED: grep — 52 files use the harness helpers, but only the parity parents + harness define old-engine drives] |
| `internal/rules/check_lint_ignore_test.go` | Uses `core.CheckIsIgnored` + old rules [VERIFIED: grep] |
| `internal/core/comment_suppressions_test.go`, `comment_suppressions_eof_test.go` | Drive old `app.Lint` [VERIFIED: grep :29,:42,:38] — main's suppression fixtures remain covered through `TestLintLinkedSuppressionParity` (core) + `TestLinkedParitySuppressions` (rules), which drive only the linked engine with pinned expectations [VERIFIED: lint_suppressions_parity_test.go:71-77] |
| old-model helpers in `internal/rules/init_test.go` | `start`/`parseFile` (go-protoparser `core.ProtoInfo` fixtures) — used only by check_lint_ignore_test.go and init_test.go itself [VERIFIED: grep]; they become dead and trip `unused`. TRIM the file: keep `parseLinkedFile`/`validate` (:105,:141 — feed all 41 carried tests) |

**Test files — rewrite onto the surviving constructor (compile unchanged only if `NewLinked → New` lands):**

| File | Detail |
|------|--------|
| `internal/rules/builder_test.go` | `TestNew_ExceptExpandsGroups/ExceptExpandsSingleGroup/DefaultSuffixValues/ExplicitSuffixValuesPreserved` (:87-177) drive `rules.New` — keep working after the rename; `TestAllGroups/TestAllRuleNames/TestAllLintUseValues` (:13-59) unaffected |
| `internal/rules/selection_test.go` | `TestRepeatedLintSelectionIsIdempotent` (:14), `TestInvalidLintNamesAreRejectedBeforeFiltering` (:43) — the D-06 validation parity guards; keep working after the rename |
| `internal/config/v1/lint_selection_test.go:50` | drives `rules.New` as a validation oracle — keeps working after the rename |
| `internal/migration/migration_test.go:46,50` | drives `rules.New` twice (generated legacy config must construct) — keeps working after the rename |
| every test calling `LintLinked` (~30 files in core/rules) | mechanical method-rename sweep (gopls/sed) |
| `internal/rules/registry_parity_guard_test.go` | DELETE — `TestRegistryParity` compares New vs NewLinked; with one registry the guard has no meaning. Its parity guarantee is subsumed: post-consolidation there is exactly one registry, and `TestAllRuleNames`/`TestAllGroups` still pin the name surface |
| `internal/rules/module_rules_guard_test.go` | KEEP — exact-set ModuleRule guard over the (renamed) registry; PackageNoImportCycle stays absent until Phase 6 |

**go.mod:** `go-protoparser/v4` STAYS — post-deletion production users: `breaking_check.go`, `breaking_checker.go`, `comment_suppressions.go` (collectDirectiveComments), `proto_comments.go`, `proto_info_read.go`, `dom.go`, `internal/modules/import_roots_imports.go` [VERIFIED: grep minus dying files]. `bufbuild/protocompile` STAYS — users are entirely outside the lint path: `generate.go`, `breaking_profiles.go`, `migration_bindings.go`, `modules/imports*`, `version` [VERIFIED: grep]. Gate: `go mod tidy && git diff --exit-code go.mod go.sum` plus a grep gate asserting both module paths remain in go.mod.

### Pattern 3: Rename-map consolidation scope (which renames land now)

The binding maps live in `internal/core/lint_model.go:9-33` and `internal/rules/builder_linked.go:20-54` — they promise this consolidation. Sweep is bounded: **no `Linked*` symbol is referenced outside internal/core and internal/rules** [VERIFIED: grep over internal/ cmd/].

**REQUIRED renames (the flip cannot compile or cannot satisfy SC2 without them):**

| Rename | Why required |
|--------|--------------|
| `LintLinked → Lint` (method) | Frees `actionV1:205` unchanged; old `Lint` deleted same commit |
| `NewLinked → New` (constructor) | buildCore text unchanged; `config/v1/lint_selection_test.go` + `migration_test.go` keep compiling; old `New` deleted same commit |
| `Options.LinkedRules → Options.Rules`, `Core.linkedRules → Core.rules` | One rule surface; `lint_suppressions.go:55` fallback reads `c.rules` |
| `LintRule → Rule` (interface) | Old `Rule` deleted same commit; breaking has no `core.Rule` implementations [VERIFIED] |
| `LinkedFileInfo → ProtoInfo` | Old `ProtoInfo` verified dead (only old lint path + old rules referenced it) |
| `LinkedRuleName → GetRuleName` | Old `GetRuleName` deleted same commit; `toUpperSnakeCase` (dom.go:138-154) stays as the reflection fallback |
| `AppendLinkedIssue → AppendIssue` (4-arg) | Old 5-arg deleted same commit |
| `shouldIgnoreLinked → shouldIgnore` | Old method deleted with lint.go |
| `linkedCommentSuppressions → commentSuppressions` | Old method deleted with the old walk |
| `linkedParsedDirective → parsedDirective` | Old type deleted with the old walk |
| `pascalToUpperSnakeLinked → pascalToUpperSnake` | Old definition dies with `enum_value_prefix.go`; shared by `enum_value_prefix_linked.go` + `enum_zero_value_suffix_linked.go` — must land in the deletion commit (Pitfall 1) [VERIFIED: grep] |
| `existInProtoLinked → existInProto` | Old definition dies with `import_used.go` [VERIFIED: import_used.go:140 vs import_used_linked.go:189] |

**PERMANENT EXCEPTIONS (cannot collapse — the breaking path owns the unsuffixed names):**

| Keeps its name | Why |
|----------------|-----|
| `linkedFilesFromImport`, `linkedFileFromImport` (lint_linked_read.go:52,92) | `readFilesFromImport`/`readFileFromImport` survive in proto_info_read.go for the breaking read path (different signatures: unordered proto vs protoast) — same-package name collision [VERIFIED: proto_info_read.go:71,89] |
| `linkedDirectiveComment` (lint_suppressions.go:12) | `directiveComment` (*parser.Comment) survives via proto_comments.go + `BreakingProtoInfo.directiveComments` [VERIFIED: comment_suppressions.go:28, proto_info_read.go:49] |

Renaming the breaking-owned symbols to free these three names is REJECTED: it edits breaking-owned files for zero behavior gain and widens the flip diff into the must-not-touch zone (REQ-5.2).

**OPTIONAL polish (separate mechanical commit, recommended):** 41 port-type renames (`PackageDirectoryMatchLinked → PackageDirectoryMatch`, ... per the builder_linked.go:20-54 map) + dropping the `LintRuleName()` override methods + deleting both rename-map header comments + optional file renames (`lint_linked.go → lint.go`, folding `builder_linked.go` into the trimmed `builder.go`). Name-neutrality is proven: the registry-parity guard established that `New`'s reflection-derived names (computed over the UNSUFFIXED type names) equal `NewLinked`'s override-reported names [VERIFIED: 04-07 registry-parity guard, TestRegistryParity] — after the type renames the reflection fallback reproduces the identical strings; keep `TestAllRuleNames`/`TestAllGroups` green as the post-rename name gate.

### Pattern 4: The compat-gate design (SC1: format/paths/exit codes identical incl. COMPILE)

**Gate 1 — Golden fixture corpus (in-process, full registry).** Before the flip, convert the Phase-4 closure corpus (13 files / 9 packages, 73 old-side findings, coverage assertion requiring all 41 rule names to fire [VERIFIED: 04-07-SUMMARY ledger]) from old-vs-new parity into a **pinned golden**: capture the byte-frozen old engine's complete issue set — (path, rule, line, col, source, message) — as committed literals. Post-flip the renamed engine must reproduce exactly that, including messages (upgrading the Phase-4 fingerprint, which deliberately excluded Message) and the D-01 order. The 41 carried `<rule>_test.go` files (pinned per-rule literals on the linked contract) and the suppression matrix (`TestLintLinkedSuppressionParity` in core, `TestLinkedParitySuppressions` in rules) are the rule-level golden layer and survive with mechanical renames. The corpus must keep compiling (Phase-4 lesson: parity/golden fixtures must compile under the linked engine's full type resolution).

**Gate 2 — Action-path matrix + exit codes (completes Phase-2 A1).** Extend `internal/api/lint_link_matrix_test.go`:
- In-process `TestLintLinkMatrix`: add the class-1 cell — a corpus that fails compilation (not missing-import, not syntax; e.g. proto3 `required` field) must come back through `actionV1` as printed COMPILE issues and `ErrHasLintIssue`. Pre-flip this cell is physically unreachable (the old engine runs; a broken file is a syntax error class) — the cell lands IN the flip commit.
- Subprocess `TestLintLinkMatrixExitCodes`: add `compile_error_code_1` — exit 1 with stdout `path:line:col: message (COMPILE)` (reuse `buildEasypBinary`/`runEasypLint` as-is; sequential cells per the process-state exception).
- Existing cells re-run through the new engine post-flip: missing-import → exit 2 + "Cannot import file" stderr; syntax error → exit 1; clean → exit 0 empty stdout; rule findings → exit 1. These four are the primary SC1 exit-code gate.
- The `TestPrintIssuesCompileGolden` pins (text byte-exact `test.proto:4:3: ... (COMPILE)`; JSON zero Filename/Offset; D-03 multi-issue order preserved through the api path-sort) already exist pre-flip [VERIFIED: lint_link_matrix_test.go:246-326] and bound the printer contract.

**Gate 3 — CLI stdout golden over a rule-finding corpus.** One corpus exercising findings + ignore_only + comment directives through the real binary: pin the exact stdout lines (path:line:col:source message (RULE) order after the engine sort + api stable path sort) and exit code, captured pre-flip from the old binary, asserted post-flip. This is the golden-master layer at the CLI boundary (covers path rendering, `PACKAGE_DIRECTORY_MATCH` SourceName→Path rewrite at lint_v1.go:214-216, and format edge shapes). Implement as a test reusing `buildEasypBinary` with a committed fixture map.

**Gate 4 — Same-run subsystem gates (SC3) + mechanical closure:**
- `go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/api/... ./internal/modules/... ./internal/migration/... ./internal/config/...` — breaking (core+api breaking_* tests), generate (core+api generate_* tests), mod tidy (internal/modules), migrate (internal/migration) all inside this run. EXCLUDE `./internal/adapters/plugin/` and `./internal/generation/` — pre-existing wazero data race poisons both under `-race` [VERIFIED: STATE.md blocker, Phase-3 verifier correction; ledger item 18 — owned before the Phase-7 full-suite gate, explicitly not a Phase-5 blocker]. Document the carve-out in the plan, same as Phases 2–4.
- `go build ./...`; `go mod tidy && git diff --exit-code go.mod go.sum`; `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/`; `bin/golangci-lint run ./internal/...` (0 issues — validates the exclusion removal and the trim of init_test.go/lint_helpers.go); `go build -o /tmp/easyp-check ./cmd/easyp && /tmp/easyp-check lint --help` exit 0.
- Breaking-path freeze gate: `git diff --exit-code -- internal/core/proto_info_read.go internal/core/proto_comments.go` (zero-edit constraint) plus the breaking suites green.
- Deletion-completeness gate: `git status --porcelain` shows only intended deletions; a grep gate asserting zero references to deleted symbols (`core.AppendIssue` 5-arg call shape, `core.CheckIsIgnored`, `lintProtoInfoRead` call sites outside the parity-pin literal, `rules.NewLinked` post-consolidation).

### Pattern 5: Disposition of every Phase-2/Phase-4 deferred finding

| Finding | Disposition | Rationale / action |
|---------|-------------|--------------------|
| Phase-2 WR-01: import open/read failures misclassified as COMPILE (link.go:126-137; the generic `"File not found."` resolution swallows real errors into a misleading COMPILE remark) | **FIX NOW (pre-flip)** | Required by SC1: post-flip an unreadable import would exit 1 as a COMPILE issue with a false "not found" message, while main exits 1 as a command error carrying the OS cause. Fix per 02-REVIEW: add `broken atomic.Bool` to `lintResolver`; non-`ErrNotExist` open/read failures set `broken` and the `linkSources` fallback condition gains `resolver.broken.Load()`. The `filepath.IsLocal` pre-check is ALREADY in `openImportFile` (proto_info_read.go:103-106) and must stay the COMPILE-class boundary for non-local imports (etalon A4 cell must stay green — its fixture never touches the broken path). Pin with an unreadable-import fallback cell |
| Phase-2 WR-02: lintResolver serves corpus files from memory bypassing `importFileAllowed`/`sourceFileOpen` (link.go:126-128 vs openDiskSource proto_info_read.go:170-174) | **FIX NOW (pre-flip)** | Required by SC1: on module-scoped runs with restrictive source roots, a blocked corpus file errors today (command error) but would link cleanly and lint post-flip — silent file-selection divergence. Fix: corpus branch consults `r.core.importFileAllowed` on the rooted full path (mirroring openDiskSource); denial → treat as missing → classifier falls back → per-file readSource → openDiskSource denial yields main's exact `os.ErrNotExist` carrier. Pin with a blocked-corpus-file cell |
| Phase-2 IN-01: `lintResolver.FindFileByPath` ignores ctx (link.go:126) | **FIX NOW (pre-flip)** | Mechanical: add the same `ctx.Err()` guard every sibling resolver has (protoc/resolver.go:27-97); zero behavior change, restores cancellation promptness |
| Phase-2 IN-02: COMPILE diagnostics target files outside the corpus and bypass ignore machinery (link.go:70-82) | **DECIDE + PIN (keep current behavior)** | Keep: COMPILE bypasses `c.ignore`/`ignoreOnly`/comment suppressions — compiler diagnostics are not rule findings; matches buf/etalon; already pinned in code by `compile_bypasses_suppression` (lint_linked_matrix_test.go:269) and research A6. Record the decision in the `linkSources` doc comment at flip; pin with an ignored-broken-dependency-file cell |
| Phase-2 A3: natural-warn fixture re-pin | **NO ACTION (documented non-actionable)** | Producing a natural warn requires editing sealed `internal/easypc` (D-04, Phase-2 decision). The synthetic-warn `diagCollector` pin (warn-drop policy) stays green through the flip; carry the constraint forward |
| Phase-2 A1: class-1 (COMPILE) matrix through Action | **EXECUTED THIS PHASE** | Gate 2 above — the flip is the first state where class-1 through `actionV1` is reachable; ratification action = the new cells pass |
| Phase-4 ledger 1: resolution-based ImportUsed | **DEFER → Phase 6/7** | Deliberate behavior change; would break the flip's golden gate (the qualified-nested-reference cell documents the divergence). Phase 5 keeps main's syntax walk |
| Phase-4 ledger 2: cross-file RPCRequestResponseUnique | **DEFER → Phase 6/7** | Same reasoning |
| Phase-4 ledger 3: swift_prefix ruby_package fix | **DEFER → Phase 6/7** | Bug pinned bug-for-bug; correcting now breaks corpus parity |
| Phase-4 ledger 4: NamePos positions | **CLOSED (main wins permanently)** | No action in any phase unless deliberately reopened |
| Phase-4 ledger 5: helper-casing adoption / casing.go | **EXECUTED THIS PHASE (deletion)** | `casing.go` + `casing_test.go` deleted with the old path; `pascalToUpperSnakeLinked` collapses. Lab acronym-aware helpers are NOT adopted |
| Phase-4 ledger 6: report-everyone `groupedValues.mismatches` | **EXECUTED THIS PHASE (deletion)** | Deleted from lint_helpers.go with the old path |
| Phase-4 ledger 14: old-path consolidation + rename map | **EXECUTED THIS PHASE** | Patterns 2+3 |
| Phase-4 ledger 15: buildCore wiring | **EXECUTED THIS PHASE** | Pattern 1 |
| Phase-4 ledger 16: parseProtoLinked | **STAYS REMOVED** | No batch needed the etalon module_rules_test shape |
| Phase-4 ledger 17: PACKAGE_NO_IMPORT_CYCLE | **DEFER → Phase 6** | Exact-set `TestModuleRuleSet` keeps asserting its absence through the flip |
| Phase-4 ledger 18: wazero data race | **OUT OF SCOPE (carve-out documented)** | Poisoned packages excluded from the -race gate; owned before Phase 7 |
| Standing constraint (a): disk instrumentation `TestReadSourceOnce` | **KEEP GREEN through the flip** | readSource is the flip's read path; the cell survives with renames (lint_linked_readsource_test.go:57) |
| Standing constraint (b): suppression centrality (rules never self-suppress) | **KEEP GREEN through the flip** | Central-filter cells survive; no rule gains suppression logic |

### Anti-Patterns to Avoid

- **Partial deletion commits:** deleting old rule files without the same-commit rename collapses (or vice versa) breaks the build — e.g. `pascalToUpperSnakeLinked` cannot rename before `enum_value_prefix.go` is gone, and `enum_zero_value_suffix_linked.go` breaks if one lands without the other.
- **Keeping the parity harness "just in case":** it references the deleted old engine; it cannot compile. Its value is fully transferred into goldens BEFORE deletion.
- **Editing breaking-owned files during consolidation** to free the three exception names: REQ-5.2 freeze; zero behavior gain.
- **Deleting `builder.go` wholesale:** `AllGroups`/`AllRuleNames`/`AllLintUseValues` + unwrap helpers are load-bearing for api exclusions, config schemas, and migration tests.
- **Skipping the golden capture pre-flip:** after the old engine is deleted there is no oracle left to capture from; the capture must precede or be part of the flip-plan execution order.
- **Treating the dormancy audit as still applicable:** post-flip "old path byte-frozen" is meaningless; the freeze gates invert to breaking-path zero-diff + deletion completeness.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Before/after output comparison | Custom diff scripts over runs | Committed golden literals in tests (Phase-4 pinned-count pattern + `TestPrintIssuesCompileGolden` shape) | Deterministic, CI-visible, revertable; the corpus and its 73-finding pin already exist |
| Subprocess exit-code capture | New harness | `buildEasypBinary`/`runEasypLint` (sync.OnceValues build; sequential cells) | Proven in Phase 2, handles ExitError codes 0/1/2 |
| Rename sweep | Manual per-file edits | `gofmt`/gopls rename or a scripted sed sweep with compile+suite gates | ~52 test files carry the `LintLinked` call rename; Phase 1 already used a scripted sweep precedent with value-unchanged proof |
| Registry name surface after type renames | New naming logic | Existing reflection fallback `toUpperSnakeCase` + `TestAllRuleNames` gate | Provably name-neutral (registry-parity guard) |
| Error-class classification fixes | New classifier | Extend `lintResolver` flags per the 02-REVIEW fix sketches | The three-outcome contract and its 16+ cells are pinned; fixes are additive |

**Key insight:** this phase is deletion + mechanical rename + gates, not design work. Every hard behavior decision was made in Phases 2–4; the research's job is to make the deletion order and the gate coverage exact.

## Runtime State Inventory

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — the lint path holds no persistent state; caches (`importCache`, `linkedFiles.outlines`) are per-run in memory | None |
| Live service config | None — no external services involved | None |
| OS-registered state | None | None |
| Secrets/env vars | None touched | None |
| Build artifacts | `schemas/*.json` — rule names unchanged → drift impossible but gated (`schema-gen` + `git diff --exit-code schemas/`); `bin/golangci-lint` unaffected; local `easyp` binary gitignored; `.golangci.yml` exclusion line 24 REMOVED deliberately (documented obligation) | Remove exclusion with the helpers in one commit; run schema check |

*(Verified by inspection this session: the phase deletes code and renames identifiers; no runtime system stores or registers any deleted name.)*

## Common Pitfalls

### Pitfall 1: Deletion/rename ordering (the phase's central mechanical trap)
**What goes wrong:** The flip commit compiles only if every rename collapse lands atomically with its enabling deletion. Examples verified in source: `pascalToUpperSnakeLinked → pascalToUpperSnake` requires `enum_value_prefix.go` gone (the definition lives there, not in casing.go [VERIFIED: enum_value_prefix.go:62]) while `enum_zero_value_suffix_linked.go` consumes the renamed symbol; `existInProtoLinked → existInProto` requires `import_used.go` gone; `close` must MOVE out of `lint.go` before the file is deleted (breakingProtoInfoRead and three linked-path call sites use it); `GetRuleName` collapse requires `lint_suppressions.go:55`'s `c.rules` fallback to compile against the renamed field.
**Why it happens:** rename maps were recorded per-batch across two packages; consumers cross file boundaries.
**How to avoid:** plan the flip as ONE commit with an explicit symbol-level checklist (Pattern 2 tables are that checklist); gate = `go build ./...` + full suite + golangci. Never split "delete now, rename later" across commits.
**Warning signs:** compile errors referencing `_linked` suffixed symbols after deletion; `close undefined` after removing lint.go.

### Pitfall 2: Test fixtures and helpers shared between old and new suites
**What goes wrong:** Deleting `init_test.go` wholesale kills `parseLinkedFile`/`validate` used by all 41 carried tests; deleting the parity harness removes helpers imported (same-package) by files that must survive; the repo-root `testdata/` corpora (auth/api/import_used/no_lint/...) are referenced by surviving tests and MUST NOT be touched [VERIFIED: init_test.go:14-28 constants; testdata/ tree].
**Why it happens:** same-package test helpers create invisible dependencies.
**How to avoid:** trim, don't delete: init_test.go keeps linked helpers; deletion list is file-precise (Pattern 2 test tables); run the full rules+core suites after every test deletion commit.
**Warning signs:** `parseLinkedFile undefined`; surviving tests failing on missing fixture paths.

### Pitfall 3: Import cycles after rename consolidation
**What goes wrong:** Package `core` cannot import `internal/rules` (rules imports core) — any consolidation that moves registry/naming code from rules into core (or vice versa) in the wrong direction creates a cycle. Core-package tests already use literal `KnownLintRules` lists for exactly this reason (Phase-4 precedent, STATE.md).
**Why it happens:** the name functions (`GetRuleName`/`toUpperSnakeCase`) live in core and are consumed by rules — direction rules→core only.
**How to avoid:** keep `GetRuleName`/`AppendIssue`/`buildError`-equivalent in package core; registry constructors stay in package rules; no new core→rules imports anywhere including tests.
**Warning signs:** `import cycle not allowed` on a seemingly innocent move.

### Pitfall 4: Dormancy-gate inversion forgotten
**What goes wrong:** Phases 3–4 gates assert the old path is byte-frozen (git diff over 44 files). Post-flip those files are deleted and the linked files are heavily edited — running the old dormancy audit fails confusingly, and its absence could mask accidental edits to breaking files.
**How to avoid:** replace the audit with its inversion (Pattern 4): breaking-path zero-diff gate (`proto_info_read.go`, `proto_comments.go`) + deletion-completeness gate + subsystem suites.
**Warning signs:** verifier running stale dormancy commands; unreviewed diffs in breaking-owned files.

### Pitfall 5: Hidden cross-references into the old contract
**What goes wrong:** files that look "new" still touch old symbols: `lint_suppressions.go:54-58` iterates `c.rules` with `GetRuleName` (old-typed); `lint_model.go`'s header comment references old names; `builder_linked.go`'s header references old names; `core.go` comments describe dormancy. Missing one breaks compile or leaves stale documentation contradicting the post-flip tree.
**How to avoid:** after the flip commit, grep gates: zero `LinkedFileInfo|LintRule\b|LinkedRuleName|AppendLinkedIssue|shouldIgnoreLinked|linkedCommentSuppressions|NewLinked|LintLinked` hits outside the three documented exceptions (plus any retained type suffixes if the optional rename is deferred).
**Warning signs:** golangci `unused`/`staticcheck` complaints; stale comments describing "the old engine".

### Pitfall 6: The wazero race poisons the full-suite reading
**What goes wrong:** `go test -race -count=1 ./internal/...` fails in `internal/adapters/plugin` + `internal/generation` for reasons unrelated to the flip — an unreadable SC3 gate.
**Why it happens:** pre-existing third-party wazero version-cache race (ledger item 18, deliberately out of scope).
**How to avoid:** scope the gate explicitly and document the carve-out in the plan and SUMMARY (established Phase 2–4 practice); the flip's own blast radius is covered by the included packages.
**Warning signs:** a reviewer demanding the full-package -race run green.

### Pitfall 7: Golden corpus divergence through linked-engine type resolution
**What goes wrong:** adding fixtures to the golden corpus that don't compile under the linked engine's full type resolution routes them to COMPILE and vacates the rule-finding golden (Phase-4 deviation #3: undeclared rpc payload types produced 8 linked-side COMPILE findings).
**How to avoid:** declare every referenced type; keep the coverage assertion (all 41 rule names must fire) and the pinned count asserted before the compare; COMPILE-class coverage lives in dedicated error-path cells, not the golden corpus.
**Warning signs:** golden test suddenly reporting COMPILE issues.

## Code Examples

### Post-flip buildCore (consolidation shape — textually unchanged, for the plan's seam documentation)
```go
// Source: internal/api/runtime.go:62-88 (current text; identical post-flip)
func buildCore(log logger.Logger, cfg config.Config, importRoots modules.SourceRoots) (*core.Core, error) {
	lintRules, ignoreOnly, err := rules.New(cfg.Lint) // now the renamed NewLinked body
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}
	return core.New(core.Options{
		Rules:                   lintRules, // now []core.Rule (linked contract)
		Ignore:                  append(append([]string(nil), cfg.Lint.Ignore...), defaultVendorDir),
		IgnoreOnly:              ignoreOnly,
		AllowCommentIgnores:     cfg.Lint.AllowCommentIgnores,
		KnownLintRules:          rules.AllRuleNames(),
		Logger:                  log,
		ImportRoots:             importRoots.Paths(),
		ImportFileAllowed:       importRoots.FileAllowed(),
		OpenSourceFile:          importRoots.OpenSourceFile,
		CurrentProjectGitWalker: go_git.New(),
		BreakingCheckConfig: core.BreakingCheckConfig{ /* unchanged */ },
	}), nil
}
```

### WR-01 fix shape (from 02-REVIEW, validated against current source)
```go
// Source: .planning/phases/02-.../02-REVIEW.md WR-01 fix sketch + internal/core/link.go:118-145
type lintResolver struct {
	core    *Core
	disk    FS
	linted  map[string][]byte
	missing atomic.Bool // an import could not be found
	broken  atomic.Bool // an import failed to open or read for a non-notFound reason
}

func (r *lintResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil { // IN-01: same guard as sibling resolvers
		return nil, err
	}
	if src, ok := r.linted[path]; ok {
		// WR-02: mirror openDiskSource's policy on corpus files:
		if rooted, ok := r.disk.(interface{ RootPath() string }); ok && r.core.importFileAllowed != nil {
			if !r.core.importFileAllowed(filepath.Join(rooted.RootPath(), path)) {
				r.missing.Store(true) // fall back; per-file path yields main's os.ErrNotExist carrier
				return nil, protoc.ErrFileNotFound
			}
		}
		return src, nil
	}
	f, err := r.core.openImportFile(r.disk, path)
	if err != nil {
		var notFound *OpenImportFileError
		if errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist) {
			r.missing.Store(true)
			return nil, protoc.ErrFileNotFound
		}
		r.broken.Store(true) // WR-01: real open/read failure — never COMPILE
		return nil, err
	}
	// ... unchanged read path
}

// linkSources fallback condition gains the broken flag:
if resolver.missing.Load() || resolver.broken.Load() || !errors.Is(err, protoc.ErrCompilationFailed) || len(diags) == 0 {
	// existing debug log + (nil, nil) fallback
}
```
Note: `filepath.IsLocal` is NOT added to the resolver — the rejection of non-local imports already happens inside `openImportFile` (proto_info_read.go:103-106) and MUST continue to classify as COMPILE (etalon A4); the WR-01 fix only reroutes *real* open/read failures. The 02-REVIEW sketch's IsLocal mention is satisfied by the existing boundary — verify the A4 cell stays green.

### Golden gate skeleton (post-flip, replaces the parity harness)
```go
// Source: shape follows parity_harness_test.go + parity_full_registry_test.go (captured pre-flip)
func TestLintGoldenFullRegistry(t *testing.T) {
	// corpus: the 13-file/9-package Phase-4 closure corpus (compiles; every rule fires)
	got, err := app.Lint(t.Context(), fs.NewFSWalker(root, ".")) // renamed engine
	require.NoError(t, err)
	require.Len(t, got, 73)                       // pinned old-engine count (oracle captured pre-flip)
	require.Equal(t, goldenIssues, got)           // FULL compare: path, rule, line, col, source, message, order
}
```

### Deletion-completeness + freeze gates (phase closure)
```bash
go build ./... && go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/api/... \
  ./internal/modules/... ./internal/migration/... ./internal/config/...
git diff --exit-code -- internal/core/proto_info_read.go internal/core/proto_comments.go   # breaking freeze
go mod tidy && git diff --exit-code go.mod go.sum && grep -q go-protoparser go.mod && grep -q protocompile go.mod
go run ./cmd/easyp schema-gen && git diff --exit-code schemas/
bin/golangci-lint run ./internal/...
go build -o /tmp/easyp-check ./cmd/easyp && /tmp/easyp-check lint --help   # exit 0
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Dual engines side by side (old serves CLI, linked dormant) | Single engine; old path deleted | This phase | The rename-map coexistence cost (suffixed identifiers, mirrored suppression skeleton) is repaid; ~2,900 lines of old rules + engine removed |
| Old-vs-new parity harness as the regression net | Pinned goldens + carried per-rule tests + suppression matrix | This phase | The oracle is gone by design; Phase 7's corpus lint-diff (lab.py vs main) becomes the external parity net |
| Dormancy audits ("old path byte-frozen") | Freeze gates on breaking-owned files + deletion completeness | This phase | Audit scope inverts |
| `easyp lint` sequential per-file go-protoparser walk | Linked compiler engine: one corpus Link, windowed parallel models, ModuleRule lifecycle, deterministic D-01 output | This phase (built Phases 2–4) | User-visible: same findings format; deterministic order (documented improvement); new COMPILE remarks on non-compiling corpora |

**Deprecated/outdated:** the old `Rule` contract, `ProtoInfo` (go-protoparser lint model), `CheckIsIgnored`, the reflection comment walk for lint — all deleted. go-protoparser itself STAYS (breaking + modules).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The consolidation shape (rename collapse making buildCore/actionV1 textually unchanged) satisfies SC2's "врезка — одно изменение в шве buildCore"; the alternative reading (an explicit buildCore edit is expected) is served by the minimal-flip shape documented in Pattern 1 | Pattern 1 | Planner/user preference; both shapes are gate-equivalent — a checkpoint at plan review settles it cheaply |
| A2 | Dropping `LintRuleName()` overrides after the optional 41 type renames is name-neutral | Pattern 3 | If any override diverged from its reflection-derived name, findings would change — but TestRegistryParity proved override-names == old-registry reflection-names over the unsuffixed names; residual risk caught by TestAllRuleNames + golden gate |
| A3 | The wazero `-race` carve-out remains necessary in Phase 5 CI (not fixed since Phase 1) | Pattern 4 / Pitfall 6 | If it were fixed meanwhile, the gate can widen; if the carve-out is dropped without a fix, the gate reads red spuriously |
| A4 | `BreakingProtoInfo.directiveComments` stays write-only (no behavior impact from the trim of the old lint walk) — we are NOT touching it either way | Pattern 2 | None — the field and its population path are frozen; only relevant because it pins `directiveComment`'s survival |
| A5 | Phase-7's external corpus gate (lab.py vs main binary) does not need anything reserved in Phase 5 beyond the surviving golden corpus and fixtures | Validation Architecture | Phase 7 planning would need to re-derive corpus assets if this is wrong — cheap to check then |

## Open Questions

1. **Consolidation shape ratification (A1)**
   - What we know: both shapes are behaviorally identical; consolidation is what the rename-map headers promise and leaves zero stale names.
   - What's unclear: whether the planner/human wants the flip commit to literally edit buildCore (reviewability) or accepts the identity-flip.
   - Recommendation: consolidation shape; add one explanatory comment at buildCore marking the seam ("this constructor binds the lint engine; the Phase-5 flip changed the bindings, not this file").
2. **Scope of the optional type-rename commit**
   - What we know: 41 type renames + override drops + header deletions are behavior-neutral and gate-verifiable, but enlarge the phase's diff (~52 files touched mechanically).
   - What's unclear: land in Phase 5 (recommended — the phase owns "старые реализации правил удалены" and the rename-map promise) or defer to a Phase-6 prep step.
   - Recommendation: land in Phase 5 as a separate mechanical commit immediately after the flip commit; both commits share the gate suite.
3. **Golden corpus home**
   - What we know: the Phase-4 corpus lives as in-test string maps in parity_full_registry_test.go, which dies with the harness.
   - What's unclear: whether to move it to a committed `testdata/` tree or re-host it in a new in-test file.
   - Recommendation: re-host in-test (established pattern, keeps corpus+assertions adjacent); move to testdata/ only if the corpus gains non-trivial file layout needs (e.g. ignore-path semantics).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | everything | ✓ | go1.27.3 (go.mod 1.26.6) | — |
| `bin/golangci-lint` | lint gate | ✓ | pinned in bin/ | — |
| git | commits, freeze gates | ✓ | — | — |
| `task` / gotestsum | `task test`/`task lint` | ✗ (absent from PATH and bin/) | — | raw equivalents documented in AGENTS.md (`go test -race -count=1 ./internal/...`, `go build -o easyp ./cmd/easyp`) |
| etalon clone | NOT needed (deletion/consolidation need no lab source); present at /tmp/gsd-research/compiler-lab @ 08cef10 if file-layout fidelity questions arise | ✓ | — | `gh repo clone easyp-tech/compiler-lab` |
| External packages | none to install | — | — | — |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** `task`/gotestsum — use raw `go` commands (operating mode since Phase 1).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go standard `testing` + testify; runner gotestsum via `task test` (absent → raw `go test -race`) |
| Config file | none (convention-driven: colocated `*_test.go`, `-race` mandatory; subprocess tests sequential) |
| Quick run command | `go test -race -count=1 ./internal/core/... ./internal/rules/...` |
| Full suite command | `go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/api/... ./internal/modules/... ./internal/migration/... ./internal/config/...` (documented carve-out: `./internal/adapters/plugin/`, `./internal/generation/` excluded — wazero race, ledger item 18) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LINT-COMPAT-01 | Full-registry golden: all 41 rules' findings (path, rule, line, col, source, message) byte-identical to the frozen old engine on the closure corpus | integration golden | `go test -race -count=1 ./internal/rules/ -run TestLintGolden` | ❌ Wave 0 (capture pre-flip; asserted post-flip; replaces parity_full_registry_test.go) |
| LINT-COMPAT-01 | Per-rule literals + suppression matrix survive the flip | unit/integration (carried) | `go test -race -count=1 ./internal/rules/... ./internal/core/... -run 'TestLinkedParity|Suppression'` | ✅ (rename sweep only) |
| LINT-COMPAT-01 | Class-1 COMPILE through actionV1: issues printed in existing format, exit 1 | integration + subprocess golden | `go test -race -count=1 ./internal/api/ -run 'TestLintLinkMatrix'` | ⚠️ extend (new cells land in the flip commit — A1 execution) |
| LINT-COMPAT-01 | Exit codes 0/1/2 through the real binary post-flip (missing-import 2, syntax 1, clean 0, findings 1, COMPILE 1) | subprocess (sequential) | same file `TestLintLinkMatrixExitCodes` | ⚠️ extend (new COMPILE cell) |
| LINT-COMPAT-01 | CLI stdout golden over findings+ignore_only+directives corpus (text + JSON) | subprocess golden | new test reusing `buildEasypBinary` | ❌ Wave 0 (capture pre-flip) |
| LINT-COMPAT-02 | breaking/generate/tidy/migrate suites green in the same run | suite gate | full-suite command above | ✅ existing suites (must stay green through the flip) |
| LINT-COMPAT-02 | go-protoparser + protocompile remain required | mechanical gate | `go mod tidy && git diff --exit-code go.mod go.sum` + grep gate | ✅ command-only |
| (gate) | Schema drift empty (rule names unchanged) | mechanical gate | `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` | ✅ command-only |
| (gate) | Breaking-path files byte-frozen | freeze gate | `git diff --exit-code -- internal/core/proto_info_read.go internal/core/proto_comments.go` | ✅ command-only |
| (gate) | Engine determinism/order preserved through renames | repeated-run cell | `go test -race -count=1 ./internal/core/ -run TestLintLinkedMatrix` (renamed) | ✅ exists |

### Sampling Rate
- **Per task commit:** quick run command + `go build ./...`
- **Per wave merge:** full suite command + `bin/golangci-lint run ./internal/...`
- **Phase gate:** full suite + all mechanical gates (Pattern 4 Gate 4) + `easyp lint --help` exit 0, before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] Golden capture: full-registry corpus goldens (incl. messages) + CLI stdout goldens captured from the byte-frozen old engine BEFORE the flip commit
- [ ] WR-01/WR-02 fallback cells (unreadable import; blocked corpus file) — linked-engine level, pre-flip
- [ ] IN-02 decision comment + ignored-broken-dep cell at flip
- [ ] Class-1 in-process + subprocess cells in lint_link_matrix_test.go (flip commit)
- [ ] Rename sweep plan for ~30 test files calling `LintLinked` (mechanical)

*(Existing infrastructure covers the rest: testify conventions, Phase-2 matrix harness, Phase-4 corpus, suppression fixtures, `buildEasypBinary` pattern.)*

## Security Domain

`security_enforcement: true` (ASVS L1). Internal CLI refactor; the file-access policy surface is the security-relevant part and it is exactly where WR-02 lands.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | N/A — local CLI |
| V3 Session Management | no | N/A |
| V4 Access Control | yes (path-level) | `importFileAllowed` + `filepath.IsLocal` in `openImportFile`/`openDiskSource` [VERIFIED: proto_info_read.go:103-106,170-174]; WR-02 fix extends the SAME policy to the linker's corpus branch — the flip must not open any file outside these two funnels |
| V5 Input Validation | yes | `validateConfigRuleNames` (D-06 single validator) + `knownLintRules` directive validation with error-on-unknown — carried verbatim through the renames |
| V6 Cryptography | no | N/A |
| V14 Config | minimal | No new config surfaces; schema drift gated |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Linker path silently linting files the policy excludes (WR-02) | Elevation/Information Disclosure | Fix-now (Pattern 5): corpus branch consults `importFileAllowed`; pin with blocked-corpus cell |
| Misleading COMPILE masking a real access error (WR-01) | Tampering (diagnostic integrity) | Fix-now: `broken` flag keeps real open/read failures in the command-error class with the OS cause |
| Import path traversal | Tampering | `filepath.IsLocal` at `openImportFile` entry — unchanged; A4 non-local COMPILE classification preserved |
| Suppression bypass via malformed directives | Tampering | Error-on-malformed/unknown directive skeleton carried verbatim (both engines byte-identical machine) |

## Sources

### Primary (HIGH confidence)
- All in-repo reads this session with line ranges cited inline: `internal/api/{runtime.go, lint.go, lint_v1.go, lint_batch.go, lint_link_matrix_test.go, lint_v1_test.go}`, `internal/core/{lint.go, lint_linked.go, lint_linked_read.go, lint_model.go, lint_suppressions.go, comment_suppressions.go, proto_comments.go, proto_info_read.go, dom.go, core.go, link.go, fs.go}`, `internal/rules/{builder.go, builder_linked.go, selection.go, casing.go, lint_helpers.go, init_test.go, parity_harness_test.go, builder_test.go, selection_test.go, check_lint_ignore_test.go}` + full file listings of both packages, `internal/generation/{module.go, config.go}`, `internal/config/v1/schema.go`, `internal/migration/migration_test.go`, `.golangci.yml`, `go.mod`
- Etalon cross-check: `/tmp/gsd-research/compiler-lab/easyp-cli/internal/core/{dom.go, comment_suppressions.go, proto_info_read.go}` @ 08cef10 — confirms the post-flip naming target (`Rule`/`ProtoInfo`/`AppendIssue` 4-arg/`GetRuleName`/`buildError`, one directiveComment)
- Planning: `.planning/ROADMAP.md` § Phase 5, `.planning/REQUIREMENTS.md` (LINT-COMPAT-01/02 + traceability), `.planning/STATE.md` (blockers/decisions), `04-07-SUMMARY.md` (Consolidated Deferred-Decisions Ledger, 18 entries), `04-VERIFICATION.md` (deferred re-matching), `03-RESEARCH.md` (rename map + collision table), `02-REVIEW.md` + `02-REVIEW-DISPOSITION.md` (WR-01/WR-02/IN-01/IN-02 detail), `02-03-SUMMARY.md` (A1/A3 flagged assumptions)

### Secondary (MEDIUM confidence)
- Golden-master / characterization testing practice [CITED: thedigitalcatonline.com/blog/2012/11/11/testing-legacy-code-with-golden-master/, endoflineblog.com, blog.jcarsiquel.com/post/legacy-code-and-golden-master/] — method-level confirmation only; all gate mechanics derive from in-repo precedents
- Go stdlib `slices.SortStableFunc` stability / `cmp.Or` semantics [CITED: pkg.go.dev] — in-repo usage verified at lint_linked.go:152-160

### Tertiary (LOW confidence)
- None — no unverified external claims.

## Metadata

**Confidence breakdown:**
- Flip seam mechanics: HIGH — every constructor/caller enumerated by grep + read; zero unaccounted `core.New`/engine call sites
- Deletion inventory: HIGH — per-file, per-symbol verified by reading the files and grepping every symbol's users this session
- Rename consolidation: HIGH — maps read at source (lint_model.go, builder_linked.go); exceptions derived from verified breaking-path symbol usage; sweep bounded by grep
- Compat-gate design: HIGH — built from existing verified harnesses (Phase-2 matrix, Phase-4 corpus) plus established golden-master practice
- Deferred-findings dispositions: HIGH for mechanical items (fix shapes read from 02-REVIEW against current source); the IN-02 keep-decision is a policy choice flagged for planner confirmation

**Research date:** 2026-11-20
**Valid until:** 2026-12-20 (stable: internal codebase work; all claims are source-read against HEAD 3065960)
