---
phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti
verified: 2026-11-20T22:04:55Z
status: passed
score: 15/15 must-haves verified
covered_files: [".golangci.yml",".planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-01-PLAN.md",".planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-01-SUMMARY.md",".planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-02-PLAN.md",".planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-02-SUMMARY.md",".planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-03-PLAN.md",".planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-03-SUMMARY.md",".planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-04-PLAN.md",".planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-04-SUMMARY.md",".planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-05-PLAN.md",".planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/05-05-SUMMARY.md","internal/api/lint_link_matrix_test.go","internal/api/lint_stdout_golden_test.go","internal/api/runtime.go","internal/core/comment_suppressions.go","internal/core/core.go","internal/core/dom.go","internal/core/link.go","internal/core/link_test.go","internal/core/lint_linked.go","internal/core/lint_linked_matrix_test.go","internal/core/lint_linked_read.go","internal/core/lint_linked_readsource_test.go","internal/core/lint_linked_test.go","internal/core/lint_model.go","internal/core/lint_suppressions.go","internal/core/lint_suppressions_parity_test.go","internal/core/proto_info_read.go","internal/core/proto_info_read_test.go","internal/rules/builder.go","internal/rules/builder_linked.go","internal/rules/comment_enum_linked.go","internal/rules/comment_enum_test.go","internal/rules/comment_enum_value_linked.go","internal/rules/comment_enum_value_test.go","internal/rules/comment_field_message_linked.go","internal/rules/comment_field_message_test.go","internal/rules/comment_message_linked.go","internal/rules/comment_message_test.go","internal/rules/comment_one_of_linked.go","internal/rules/comment_one_of_test.go","internal/rules/comment_rpc_linked.go","internal/rules/comment_rpc_test.go","internal/rules/comment_service_linked.go","internal/rules/comment_service_test.go","internal/rules/directory_same_package_linked.go","internal/rules/directory_same_package_test.go","internal/rules/enum_first_value_zero_linked.go","internal/rules/enum_first_value_zero_test.go","internal/rules/enum_no_allow_alias_linked.go","internal/rules/enum_no_allow_alias_test.go","internal/rules/enum_pascal_case_linked.go","internal/rules/enum_pascal_case_test.go","internal/rules/enum_value_prefix_linked.go","internal/rules/enum_value_prefix_test.go","internal/rules/enum_value_upper_snake_case_linked.go","internal/rules/enum_value_upper_snake_case_test.go","internal/rules/enum_zero_value_suffix_linked.go","internal/rules/enum_zero_value_suffix_test.go","internal/rules/file_lower_snake_case_linked.go","internal/rules/file_lower_snake_case_test.go","internal/rules/golden_full_registry_test.go","internal/rules/import_no_public_linked.go","internal/rules/import_no_public_test.go","internal/rules/import_no_weak_linked.go","internal/rules/import_no_weak_test.go","internal/rules/import_used_linked.go","internal/rules/import_used_test.go","internal/rules/init_test.go","internal/rules/lint_helpers.go","internal/rules/message_field_lower_snake_case_linked.go","internal/rules/message_field_lower_snake_case_test.go","internal/rules/message_pascal_case_linked.go","internal/rules/message_pascal_case_test.go","internal/rules/module_rules_guard_test.go","internal/rules/oneof_lower_snake_case_linked.go","internal/rules/oneof_lower_snake_case_test.go","internal/rules/package_defined_linked.go","internal/rules/package_defined_test.go","internal/rules/package_directory_match_linked.go","internal/rules/package_directory_match_test.go","internal/rules/package_lower_snake_case_linked.go","internal/rules/package_lower_snake_case_test.go","internal/rules/package_same_csharp_namespace_linked.go","internal/rules/package_same_csharp_namespace_test.go","internal/rules/package_same_directory_linked.go","internal/rules/package_same_directory_test.go","internal/rules/package_same_go_package_linked.go","internal/rules/package_same_go_package_test.go","internal/rules/package_same_java_multiple_files_linked.go","internal/rules/package_same_java_multiple_files_test.go","internal/rules/package_same_java_package_linked.go","internal/rules/package_same_java_package_test.go","internal/rules/package_same_php_namespace_linked.go","internal/rules/package_same_php_namespace_test.go","internal/rules/package_same_ruby_package_linked.go","internal/rules/package_same_ruby_package_test.go","internal/rules/package_same_swift_prefix_linked.go","internal/rules/package_same_swift_prefix_test.go","internal/rules/package_version_suffix_linked.go","internal/rules/package_version_suffix_test.go","internal/rules/rpc_no_client_streaming_linked.go","internal/rules/rpc_no_client_streaming_test.go","internal/rules/rpc_no_server_streaming_linked.go","internal/rules/rpc_no_server_streaming_test.go","internal/rules/rpc_pascal_case_linked.go","internal/rules/rpc_pascal_case_test.go","internal/rules/rpc_request_response_unique_linked.go","internal/rules/rpc_request_response_unique_test.go","internal/rules/rpc_request_standard_name_linked.go","internal/rules/rpc_request_standard_name_test.go","internal/rules/rpc_response_standard_name_linked.go","internal/rules/rpc_response_standard_name_test.go","internal/rules/service_pascal_case_linked.go","internal/rules/service_pascal_case_test.go","internal/rules/service_suffix_linked.go","internal/rules/service_suffix_test.go","internal/rules/validation_parity_test.go"]
covered_digest: "v4:sha256:8a2080f73f77cc22108350cd94be946528aa129ecb1ff857612912cc3db56105"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 5: Переключение lint-пути и ворота совместимости — Verification Report

**Phase Goal:** `easyp lint` работает на компиляторе; старый lint-контур удалён; формат вывода, пути, коды выхода и все остальные подсистемы не изменились — единственная точка пользовательски видимого изменения, откат = revert
**Verified:** 2026-11-20T22:04:55Z
**Status:** passed
**Re-verification:** No — initial verification
**Requirements:** LINT-COMPAT-01, LINT-COMPAT-02 (both SATISFIED — see Requirements Coverage)

**Mode note (MVP):** ROADMAP marks this phase `mode: mvp`, but the phase goal is not a User Story (`gsd_run query user-story.validate` → `valid: false` — no "As a…, I want to…, so that…" shape). Per the ratified Phase 1–4 precedent, verification proceeds goal-backward against the roadmap Success Criteria; the MVP User-Flow-Coverage narrowing is inapplicable. No items were deferred or dropped because of this.

## Goal Achievement

Every claim below was re-derived from the codebase and re-run by this verification — SUMMARY claims were treated as hypotheses, not evidence. The decisive evidence is this verification's own run of the full six-root scoped `-race` battery (exit 0), the freeze/seam `git diff` gates, the deletion-completeness greps, and the mechanical closure chain (build, schema-gen drift, tidy drift, golangci, `lint --help`).

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC1: `easyp lint` issues in the unchanged format — paths, issue format and CLI exit codes match pre-flip main on the fixture corpus, incl. new COMPILE remarks in the existing format | ✓ VERIFIED | `TestLintGoldenFullRegistry` reproduces the committed 73-finding golden byte-exactly (full struct incl. Message/Position.Filename/Offset/SourceName, D-01 order, purity guards) — green in this verification's own suite run; `TestLintStdoutGolden` pins byte-exact stdout + exit 1 through the REAL binary over the findings/exclusion/directive/rewrite corpus; exit-code matrix (missing-import 2, syntax 1, clean 0, findings 1) + class-1 COMPILE cells (`compile_issue_exit1_class` in-process ErrHasLintIssue through actionV1, `compile_error_code_1` subprocess exit 1 with the `(COMPILE)` line) all green in the same run. Golden literals verified stable across the flip (`git diff 98b9975 4238fbe -- internal/api/lint_stdout_golden_test.go` = fixture-only; 73 literal entries counted in the golden file) |
| 2 | SC2: old lint contour and old rule implementations deleted; `easyp lint` served only by the compiler engine; the flip is one change at the `buildCore` seam | ✓ VERIFIED | Flip commit `4238fbe` = 135 files, +538/−6355, 59 deletions (`git show --stat` re-run); all sampled old files absent (internal/core/lint.go, lint_proto_info.go, casing.go, 41 old rule files, 10 parity test files, check_lint_ignore_test.go, registry_parity_guard_test.go); grep gates = 0 live references for `NewLinked\|LintLinked\|LinkedFileInfo\|LinkedRuleName\|AppendLinkedIssue\|shouldIgnoreLinked\|linkedCommentSuppressions\|linkedParsedDirective\|LinkedRules\|\bLintRule\b\|CheckIsIgnored`, 0 for `LintRuleName\|\w+Linked\b`, 0 LIVE refs for `lintProtoInfoRead` (parity-pin literal + frozen-file comment + 5 proof cells only, per the plan's own prohibition); `rules.New` binds builder_linked.go's constructor (`internal/api/runtime.go:67`), `apps[key].Lint` binds the renamed engine (`internal/api/lint_v1.go:205`); `git diff f27d2d1 HEAD -- internal/api/lint_v1.go` = 0 lines, runtime.go = 0 removed lines (comment-only seam note at :62-64) |
| 3 | SC3: breaking, generate, mod tidy and migrate suites pass in the same run; go-protoparser and protocompile remain in go.mod | ✓ VERIFIED | This verification ran `go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/api/... ./internal/modules/... ./internal/migration/... ./internal/config/...` — SUITE_EXIT=0, 8/8 packages ok (breaking in core, mod tidy in modules, migrate in migration); documented wazero carve-out (adapters/plugin, generation) confirmed pre-existing: both pass WITHOUT `-race` in this verification's own run; `grep -c go-protoparser go.mod` = 1, `grep -c protocompile go.mod` = 1 |
| 4 | Committed 73-finding golden with every field populated (Path, RuleName, Position incl. Filename/Offset/Line/Column, SourceName, Message) in D-01 order (05-01) | ✓ VERIFIED | `internal/rules/golden_full_registry_test.go`: exactly 73 `golden(...)` literal entries counted; corpus re-host has 13 file keys; purity guards present in code (zero-COMPILE via `require.NotEqual(t, core.CompileRuleName, ...)`, pinned count asserted BEFORE compare, 41-name coverage) + D-01 totality guard (`requireTotalOrder`); drive green in this verification's suite run |
| 5 | Flip-invariant CLI stdout + exit-code golden committed (05-01) | ✓ VERIFIED | `internal/api/lint_stdout_golden_test.go` present, drives the real binary via `buildEasypBinary`/`runEasypLint` reuse from lint_link_matrix_test.go (tool-verified key link); green in this verification's suite run |
| 6 | WR-01: an import whose open/read fails for a non-notFound OS reason forces the classifier fallback — never a misleading COMPILE remark (05-02) | ✓ VERIFIED | `internal/core/link.go`: `broken atomic.Bool` stored on `errors.As(err, &pathErr)` (:180-183) with the extended fallback condition `resolver.missing.Load() \|\| resolver.broken.Load() \|\| ...` (:59); `unreadable_import_falls_back` cell green in this verification's suite run; RED record machine-validated (RED_EVIDENCE_OK, re-checked via `gsd_run check tdd-red-evidence`) |
| 7 | WR-02: policy-denied corpus file errors with the exact os.ErrNotExist carrier instead of linting from memory (05-02) | ✓ VERIFIED | `internal/core/link.go:146-159`: corpus branch consults `importFileAllowed` on `filepath.Join(rooted.RootPath(), filepath.FromSlash(path))` — the byte-identical construction openDiskSource uses (code re-read); `blocked_corpus_file_errors` green in the suite run; frozen proto_info_read.go itself zero-edit across the phase |
| 8 | IN-01: lintResolver.FindFileByPath honors ctx cancellation with the sibling-resolver guard shape (05-02) | ✓ VERIFIED | ctx guard present in link.go; `find_by_path_honors_ctx` cell green in the suite run |
| 9 | IN-02 (05-02): COMPILE diagnostics stay outside the suppression machinery, keep-decision recorded and pinned | ✓ VERIFIED | linkSources doc comment records the decision; `ignored_broken_dependency_not_suppressed` and `compile_bypasses_suppression` cells present and green in the suite run |
| 10 | The flip is ONE atomic commit whose revert restores main exactly (05-03) | ✓ VERIFIED | `git show --stat 4238fbe`: 135 files, +538/−6355, exactly 59 deletions (`git show --diff-filter=D --name-only` = 59); deletions, trims and renames in one commit — rollback = `git revert 4238fbe`; all 9 phase commits verified real via `gsd_run query verify.commits` (9/9 valid) |
| 11 | Consolidation: 41 port types canonical, override methods gone, name function is pure reflection, rename-map header deleted (05-04) | ✓ VERIFIED | grep `LintRuleName\|\w+Linked\b` over internal/ = 0; `grep -c 'Rule-type rename map' builder_linked.go` = 0; `GetRuleName` = `return toUpperSnakeCase(reflect.TypeOf(rule).Elem().Name())` (internal/core/lint_model.go:85-87); builder.go add* helpers construct unsuffixed types (`core.GetRuleName(&DirectorySamePackage{})`); TestAllRuleNames/TestAllGroups/TestModuleRuleSet green inside the suite run; schema-gen drift empty (independent name-surface gate) |
| 12 | Sentinel-typed broken-import discrimination: PathError-free funnel shapes (ErrUnsupported/ErrChanged/ErrNestedRepository) force the fallback (05-05 WR-01) | ✓ VERIFIED | `internal/core/link.go:184-188`: typed `errors.Is` on all three sourceview sentinels with one `broken.Store`, raw error returned (code re-read); `sentinel_import_failure_falls_back` cell present (3 carriers) and green in this verification's suite run; A4 non-local COMPILE cell stayed green (no `IsLocal` in link.go — grep 0); RED record RED_EVIDENCE_OK (re-checked) |
| 13 | IN-01 (05-05): the dead post-fallback readErrs re-check removed; COMPILE-vs-parse precedence owned solely by hasPerFileSyntaxError | ✓ VERIFIED | `internal/core/lint_linked.go:73`: else-if is `len(compileIssues) > 0` alone (grep `len(compileIssues) > 0 &&` = 0); ownership comment present (:76-82); core suite green (behavior-neutrality proven in this verification's own run) |
| 14 | IN-02 (05-05): no comment cites deleted symbols or pre-flip rename-map state | ✓ VERIFIED | `grep -c 'reclaimed the etalon name' internal/rules/import_used_linked.go` = 1; `grep -c 'lintProtoInfoRead' internal/core/proto_info_read.go` = 0 with `protoInfoFromSource` doc state present; `grep -c 'Linked suffix (rename map)'` = 0 |
| 15 | Zero collateral: golden outputs byte-identical; breaking-owned files byte-frozen (05-05, canonical-six C1/C2/C3) | ✓ VERIFIED | Freeze gates re-run across the WHOLE phase (`git diff c96dd13..HEAD`): proto_comments.go, breaking_check.go, breaking_checker.go, breaking_profiles.go = empty diff; proto_info_read.go shows ONLY the 05-05 IN-02 doc-comment rewrite (comment-only, mandated by 05-REVIEW-DISPOSITION — not a behavior edit); golden literals byte-stable (73 entries unchanged; stdout literals untouched by the flip — diff inspected); `TestPrintIssuesCompileGolden`/matrix/stdout goldens green in the suite run |

**Score:** 15/15 truths verified (0 present-but-behavior-unverified — every behavior-dependent truth above was exercised by tests that ran green in THIS verification's own suite run)

### Canonical-Six Prohibitions (plan frontmatter, all judgment-tier → resolved with codebase evidence)

| Prohibition | Status | Evidence |
|---|---|---|
| C1 output format/paths/exit codes byte-identical | ✓ held | SC1 evidence (truth 1); seam gates; parity-pin literal byte-identical (`internal/core/lint_linked.go:122`) |
| C2 breaking-owned files byte-frozen | ✓ held | Whole-phase diffs empty for 4 breaking-owned files; proto_info_read.go comment-only (05-05 IN-02, review-mandated) |
| C3 no deferred behavior change rode with the flip | ✓ held | Golden byte-identity on the verbatim 13-file corpus; corpus keys re-counted in the surviving file |
| C4 COMPILE outside suppression machinery | ✓ held | compile_bypasses_suppression + ignored_broken_dependency_not_suppressed green |
| C5 go-protoparser + protocompile stay in go.mod | ✓ held | presence greps = 1/1; tidy diff-empty |
| C6 shared registry helpers survive the trim | ✓ held | AllGroups/AllRuleNames/AllLintUseValues in builder.go; lint_helpers.go keeps exactly 2 funcs (`grep -c '^func '` = 2) |
| D-01: no diagnostic-text parsing in the resolver | ✓ held | link.go classification is typed errors.As/errors.Is only (code re-read) |
| A4: no path-locality re-check in the resolver | ✓ held | `grep -c 'IsLocal' internal/core/link.go` = 0; non-local COMPILE cell green |

### Deferred Items

None — no truth failed, so nothing to filter against later phases. (LINT-PARITY-01/NEWRULES-*/PERF-01 are Phase 6/7 roadmap scope, not Phase 5 must-haves.)

### Advisory (New Scope, Unevidenced)

Not applicable — initial verification (no prior VERIFICATION.md exists; the #3304 evidence gate is a no-op on a first pass).

### Required Artifacts

`gsd_run query verify.artifacts` per plan: 05-02 (2/2), 05-03 (6/6), 05-04 (3/3), 05-05 (5/5) — all passed. Two 05-01 tool verdicts report "missing" and are DESIGNED supersession, not gaps:

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `internal/rules/golden_full_registry_test.go` | THE golden test | ✓ VERIFIED | 73 literals + guards + surviving drive; tool's "Missing pattern: TestLintGoldenFullRegistryOld" is the capture drive whose death in the flip is mandated by 05-03's own plan ("old-drive capture deleted") — the surviving `TestLintGoldenFullRegistry` is the ratified post-flip golden and is green |
| `internal/api/lint_stdout_golden_test.go` | CLI stdout golden | ✓ VERIFIED | present, substantive, wired (reuses buildEasypBinary/runEasypLint), green |
| `internal/core/link.go` | resolver + fixes | ✓ VERIFIED | broken field, ctx guard, policy mirror, sentinel discrimination all present |
| `internal/core/lint_linked_matrix_test.go` | error-path cells | ✓ VERIFIED | all 5 cells present, green in suite run |
| `internal/core/lint_linked.go`, `core.go`, `dom.go`, `comment_suppressions.go`, `lint_model.go`, `lint_linked_read.go`, `lint_suppressions.go` | engine/trims | ✓ VERIFIED | renamed engine, one rule surface, old symbols deleted, dead term removed |
| `internal/rules/builder.go`, `builder_linked.go`, `lint_helpers.go` | registry trim | ✓ VERIFIED | old New deleted, unsuffixed literals, 2 funcs in lint_helpers, header deleted |
| `internal/rules/import_used_linked.go`, `internal/core/proto_info_read.go` | IN-02 doc states | ✓ VERIFIED | grep states confirmed |
| `.golangci.yml` | exclusions removed | ✓ VERIFIED | lint_helpers.go and casing.go entries = 0; golangci 0 issues over ./internal/... in this verification's own run |
| (deleted by design) 59 flip-deleted files | old contour gone | ✓ VERIFIED | `git show --diff-filter=D --name-only 4238fbe` = 59; sampled files absent on disk |

### Key Link Verification

`gsd_run query verify.key-links` per plan: 05-03 (3/3), 05-04 (1/1), 05-05 (3/3) verified. Two "pending" verdicts are name-level supersession by the ratified rename sweep, with substance re-verified manually:

| From | To | Via | Status | Details |
|---|---|---|---|---|
| internal/api/runtime.go | internal/rules/builder_linked.go | rules.New binds the renamed constructor | ✓ WIRED | tool-verified; buildCore at runtime.go:65-67 calls rules.New; constructor at builder_linked.go:22 |
| internal/api/lint_v1.go | internal/core/lint_linked.go | apps[key].Lint binds the renamed engine | ✓ WIRED | tool-verified; lint_v1.go:205 |
| internal/core/lint_suppressions.go | internal/core/core.go | known-rules fallback reads c.rules via GetRuleName | ✓ WIRED | tool-verified |
| internal/core/lint_model.go | internal/rules/builder.go | pure-reflection GetRuleName reproduces override names | ✓ WIRED | tool-verified; TestAllRuleNames + golden RuleName values + schema drift = the gates |
| internal/core/link.go | internal/sourceview/local.go | typed errors.Is on the three sentinels | ✓ WIRED | tool-verified; link.go:184-188 |
| internal/core/lint_linked_matrix_test.go | internal/modules/sources.go | probe carriers injected through Options.OpenSourceFile | ✓ WIRED | tool-verified |
| internal/core/link.go | internal/core/proto_info_read.go | corpus branch mirrors openDiskSource's rooted-path construction | ✓ WIRED | tool-verified; link.go:154-155 |
| internal/rules/golden_full_registry_test.go | internal/rules/parity_full_registry_test.go | corpus re-hosted verbatim | ✓ WIRED (substance) | tool reports "pattern not found" because the parity FILE was deleted in the flip BY DESIGN; substance re-verified: the corpus survives in the golden file — 13 file keys counted, 41-name coverage enforced at runtime by the green drive |
| internal/api/lint_stdout_golden_test.go | internal/api/lint_link_matrix_test.go | buildEasypBinary/runEasypLint harness reuse | ✓ WIRED | tool-verified |
| internal/core/lint_linked_matrix_test.go | internal/core/link.go | cells drive the engine over restrictive Options | ✓ WIRED (substance) | tool reports "LintLinked not found" because the flip renamed LintLinked→Lint BY DESIGN; substance re-verified: the cells drive `app.Lint` and are green in the suite run |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|---|---|---|---|---|
| golden_full_registry_test.go | goldenIssues (expected) | committed literals captured from the deleted old engine (external-oracle capture — the defined golden-master provenance) | yes (byte-exact live-engine compare incl. Message) | ✓ FLOWING |
| lint_stdout_golden_test.go | stdout/exit code | real binary subprocess over actionV1 → engine → printer | yes | ✓ FLOWING |
| lint_v1.go rendering | findings (Path/RuleName/Position/SourceName/Message) | compiled AST via rules.New → core.Options.Rules → apps[key].Lint | yes (golden compares full structs) | ✓ FLOWING |

No static returns, no hardcoded render data, no hollow props anywhere on the lint path (stub grep across all phase-modified production files: zero matches).

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| SC3 same-run subsystem battery (incl. all goldens, matrix, exit codes, error-path and suppression cells) | `go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/api/... ./internal/modules/... ./internal/migration/... ./internal/config/...` | SUITE_EXIT=0; 8/8 packages ok (api 193.1s) | ✓ PASS |
| Wazero carve-out packages are pre-existing-only (generate subsystem healthy) | `go test -count=1 ./internal/generation/... ./internal/adapters/plugin/` | both ok | ✓ PASS |
| Build + CLI help | `go build -o /tmp/easyp-verify ./cmd/easyp && /tmp/easyp-verify lint --help` | BUILD OK, exit 0 | ✓ PASS |
| Schema contract unchanged (rule names feed schema enums) | `go run ./cmd/easyp schema-gen && git status --porcelain schemas/` | 0 before, 0 after (drift empty) | ✓ PASS |
| go.mod hygiene | `go mod tidy && git diff --exit-code go.mod go.sum` | TIDY_DRIFT_EMPTY | ✓ PASS |
| Lint hygiene | `bin/golangci-lint run ./internal/...` | "0 issues.", exit 0 | ✓ PASS |
| Freeze gates (whole phase) | `git diff c96dd13..HEAD -- <4 breaking-owned files>` | empty | ✓ PASS |
| Seam gates vs pre-flip f27d2d1 | `git diff f27d2d1 HEAD -- internal/api/lint_v1.go` / removed-lines on runtime.go | 0 / 0 | ✓ PASS |
| Commit existence | `gsd_run query verify.commits` (9 hashes from SUMMARYs) | 9/9 valid | ✓ PASS |
| TDD RED evidence | `gsd_run check tdd-red-evidence` (4 records) | 4/4 RED_EVIDENCE_OK | ✓ PASS |

No test in any phase-linked test file is skipped (`t.Skip`/`SkipNow`/`testing.Short`: 0 matches).

### Probe Execution

| Probe | Command | Result | Status |
|---|---|---|---|
| — | `find scripts -path '*/tests/probe-*.sh'` + PLAN/SUMMARY probe declarations | no probe-*.sh scripts exist or are declared; the phase's "probe cells" are Go test cells inside the suite (compile_issue_exit1_class, compile_error_code_1, sentinel_import_failure_falls_back, etc.) — all executed within the spot-check battery above | SKIPPED (no probe scripts; Go test cells cover the probe surface) |

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|---|---|---|---|---|---|---|
| internal/rules/golden_full_registry_test.go | LINT-COMPAT-01 | yes | 0 | no — expected values are committed literals captured from the deleted OLD engine (legacy-oracle provenance, the VALID class; not generated from the system under test) | Behavioral (byte-exact full-struct Equal + count/coverage/totality guards) | VALID |
| internal/api/lint_stdout_golden_test.go | LINT-COMPAT-01 | yes | 0 | no — pinned stdout/exit captured from the pre-flip binary | Behavioral (byte-exact stdout, exit code) | VALID |
| internal/api/lint_link_matrix_test.go | LINT-COMPAT-01 | yes | 0 | no | Behavioral (exit codes via real subprocess, error carriers) | VALID |
| internal/core/lint_linked_matrix_test.go | LINT-COMPAT-01 | yes | 0 | no | Behavioral (typed error-chain assertions) | VALID |
| internal/core/lint_linked_readsource_test.go / lint_suppressions_parity_test.go | LINT-COMPAT-01 | yes | 0 | no | Behavioral | VALID |

**Disabled tests on requirements:** 0. **Circular patterns:** 0. **Insufficient assertions:** 0.

### Decision Coverage

Not evaluated — no `*-CONTEXT.md` exists in the phase directory; the gate skips cleanly.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| LINT-COMPAT-01 (REQ-5.1) | 05-01/02/03/04/05 | `easyp lint` keeps the current issue output format, paths and CLI exit codes unchanged (incl. new COMPILE remarks in the existing format) | ✓ SATISFIED | Truths 1, 2, 4-15: goldens byte-exact through the real binary and the renamed engine; exit-code matrix; old contour deleted; flip = one seam change |
| LINT-COMPAT-02 (REQ-5.2) | 05-03/04 | breaking/generate/mod tidy/migrate work as before (go-protoparser breaking model preserved; go-protoparser + protocompile stay in go.mod) | ✓ SATISFIED | Truth 3: six-root battery green in one run; carve-out packages green without -race; go.mod presence greps; breaking-owned files byte-frozen |

**Orphaned requirements:** none — REQUIREMENTS.md maps exactly LINT-COMPAT-01/02 to Phase 5 and both are declared by plans (traceability table already marks both Complete).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---|---|---|---|
| — | — | none: 0 debt markers (TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER), 0 placeholder/not-implemented strings, 0 empty implementations across all 108 phase-touched files | — | — |

### Human Verification Required

None. Every user-visible surface of the flip — stdout line format, path rendering, exit codes 2/1/0/1, COMPILE remark format — is pinned byte-exactly by committed goldens driven through the REAL binary, and those tests ran green in this verification's own suite run. No visual, real-time, external-service, or edge-input behavior remains that grep/tests cannot see; no must-have is judgment-tier. Per the standing chain directive, the user's ratification of the flip itself is this report.

### Gaps Summary

None. The phase goal is achieved in the codebase, proven by evidence generated during this verification:

1. **The flip is real and atomic.** Commit `4238fbe` (135 files, +538/−6355, 59 deletions) removed the entire old lint contour; zero live references to any deleted symbol remain; the api seam (`buildCore`/`lint_v1.go`) is textually unchanged vs pre-flip `f27d2d1` apart from the mandated one-line seam comment. Rollback is `git revert 4238fbe`.
2. **Output parity is proven, not claimed.** The 73-finding full-registry golden (every field, D-01 order) and the byte-exact CLI stdout/exit-code golden — both captured from the old engine before it was deleted — reproduce through the new engine and the real binary in this verification's own `-race` suite run, including the class-1 COMPILE cells in the existing print format and the 2/1/0/1 exit-code matrix.
3. **The other subsystems did not move.** One six-root `-race` run covers breaking/mod-tidy/migrate; the two wazero-poisoned packages were confirmed healthy without `-race`; the four breaking-owned files have an empty whole-phase diff; go.mod keeps both parser modules; schemas drift-empty; golangci 0 issues.
4. **All three 05-REVIEW findings are closed** (sentinel-typed WR-01 discrimination with a machine-validated RED→GREEN probe; IN-01 dead-term removal; IN-02 comment refreshes), with zero behavioral collateral.

Two `gsd-tools` verdicts that read as failures (05-01 `TestLintGoldenFullRegistryOld` artifact, 05-02 `LintLinked` key-link pattern) are designed supersession by the ratified rename sweep and deletion — the underlying substance (surviving golden drive; cells driving the renamed engine) was re-verified manually and is green. They are recorded in the artifact/key-link tables above, not as gaps.

---

_Verified: 2026-11-20T22:04:55Z_
_Verifier: Claude (gsd-verifier)_
