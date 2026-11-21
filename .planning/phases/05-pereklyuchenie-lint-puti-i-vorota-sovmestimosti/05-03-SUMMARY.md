---
phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti
plan: 03
subsystem: lint-engine
tags: [atomic-flip, engine-switch, rename-consolidation, old-path-deletion, golden-master, class-1-compile, compat-gates, rollback-point]

# Dependency graph
requires:
  - phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti (plan 01)
    provides: the committed 73-finding full-registry golden (every field, D-01 order) and the byte-exact CLI stdout golden — the compat proof the flip must reproduce
  - phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti (plan 02)
    provides: the corrected error-path classifier (WR-01/WR-02/IN-01 fixes) the flipped engine serves from day one
  - phase: 04-perenos-43-pravil-s-batch-paritetom (plan 07)
    provides: the ported 41-rule linked registry and the corpus the golden re-hosts
provides:
  - THE FLIP: easyp lint served exclusively by the compiler-backed linked engine — one atomic commit (4238fbe) whose revert restores main exactly; the old lint contour (44 production + 15 test files: core.Lint, lintProtoInfoRead, 41 old rules, casing.go, the old registry constructor, the reflection suppression walk, CheckIsIgnored, the parity harness) is deleted
  - the rename map collapsed: buildCore and actionV1 textually unchanged (seam comment only); rules.New binds the linked registry, Options.Rules carries the linked contract, apps[key].Lint binds the linked engine — zero stale Linked* symbols outside the three permanent breaking-coexistence exceptions
  - class-1 COMPILE cells through actionV1 (in-process ErrHasLintIssue + subprocess exit 1 with the (COMPILE) line) — Phase-2 flagged assumption A1 executed, RED on the old wiring (machine-validated), GREEN in the flip commit
  - TestLintGoldenFullRegistry as THE golden (old capture drive dead); CLI stdout golden literals byte-identical through the flip
  - the dormancy-audit inversion: breaking-path freeze gate + deletion-completeness greps + subsystem suites replace the Phase-3/4 "old path byte-frozen" audit
affects: [05-04 (mechanical 41-type consolidation), Phase 6 (new rules on the single engine), Phase 7 (external corpus gate)]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 85050        # 340199 patch chars over the flip commit at chars/4 (estimate 52000; deletion context lines dominate the patch)
  tasks: 2
  commits: 1           # MEASURED: git rev-list --count f89d7458ac96753d988d74ee7cbee42937cef10c..HEAD (the task commits; the metadata docs commit is excluded)
  plan_head_before: f89d7458ac96753d988d74ee7cbee42937cef10c
  plan_head_after: 4238fbef11279cd6528593599d08bd0f9818f7d8

# Tech tracking
tech-stack:
  added: []
  patterns: [identity consolidation flip (bindings change behind an unchanged seam), deletion-and-rename-in-one-commit (Pitfall 1), capture-then-flip golden (old drive dies with its oracle, literals survive), grep-gate-with-corrected-plumbing (diff-header-excluding removed-line count), live-reference invariant for deleted symbols]

key-files:
  created: []
  modified:
    - internal/core/lint_linked.go (THE engine, renamed to Lint)
    - internal/core/lint_model.go (Rule/ProtoInfo/AppendIssue/GetRuleName canonical names; rename-map header deleted)
    - internal/core/lint_linked_read.go (gains the moved close verbatim)
    - internal/core/core.go (one rule surface)
    - internal/core/dom.go (old contract symbols deleted)
    - internal/core/comment_suppressions.go (old walk deleted)
    - internal/core/link_test.go, internal/core/proto_info_read_test.go (deleted-oracle callers re-derived)
    - internal/rules/builder.go (old New deleted; add* helpers re-pointed to port types)
    - internal/rules/builder_linked.go (NewLinked renamed to New; THE registry)
    - internal/rules/lint_helpers.go (nine kept-unused helpers deleted)
    - internal/rules/golden_full_registry_test.go (single canonical golden drive)
    - internal/api/lint_link_matrix_test.go (class-1 COMPILE cells)
    - internal/api/lint_stdout_golden_test.go (twin fixture corrected, literals unchanged)
    - internal/api/runtime.go (seam comment only)
    - .golangci.yml (dead lint exclusions removed)

key-decisions:
  - "The flip is ONE commit (feat 4238fbe, 135 files, +538/−6355): the RATIFIED file-surface excess assumption executed as planned — never split deletion from rename (research Pitfall 1); rollback = git revert 4238fbe"
  - "RED evidence for the class-1 cells captured pre-commit against the old wiring and machine-validated (two records, both RED_EVIDENCE_OK) instead of a committed test(05-03) RED commit — the plan's atomicity mandate (one commit containing everything; a committed RED would carry failing tests on a tree state that never ships) supersedes the per-cycle commit split"
  - "The Gate-3 twin fixture was corrected (Excluded_File.proto's message renamed CleanFile -> ExcludedMsg): the 05-01 corpus declared the same message in two same-package files, which the old per-file engine never noticed but the linked engine's full corpus linking rejects, routing the run to COMPILE and vacating the golden (research Pitfall 7 applied to Gate 3); the pinned stdout literals were re-captured from the byte-frozen pre-flip binary (built from flip-HEAD via git archive) over the corrected fixture and are byte-identical to the 05-01 literals"
  - "Two test files outside the plan's declared surface call the deleted lintProtoInfoRead directly: proto_info_read_test.go's twin-mirror and error-label tests died with their oracle; link_test.go's two fallback cells re-pointed to protoInfoFromSource (same openImportFile funnel, same typed OpenImportFileError carrier — assertions preserved verbatim in intent)"
  - "The lintProtoInfoRead acceptance criterion's EXPECT==1 is reconciled to the achieved invariant — zero LIVE references: the parity-pin comment+literal (2 lines, byte-identical per canonical-six C1), the frozen proto_info_read.go comment (1 line, zero-edit constraint), and the five exact-string proof cells (5 lines, the Phase-3 main-wins proof) are all mandated by the plan's own prohibitions; the raw runtime.go removed-line grep counts the diff header, corrected plumbing is grep '^-[^-]' = 0"

patterns-established:
  - "Identity consolidation flip: the seam (buildCore) keeps its text; the bindings behind it change identity when the old contract dies in the same commit — zero api edits, zero stale suffixes (except the three documented permanent exceptions)"
  - "Deletion-completeness gate: zero-LIVE-reference greps (callable call sites) rather than raw string counts, because byte-frozen parity pins and their proof cells legitimately carry the deleted names as text"

requirements-completed: [LINT-COMPAT-01, LINT-COMPAT-02]  # copied verbatim from the plan frontmatter; the REQUIREMENTS.md checkboxes stay unmarked by the shared-ID gate (05-04 also declares both and has no SUMMARY yet)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "THE FLIP: one atomic commit deletes the old lint contour (59 files), collapses the rename map behind the unchanged buildCore/actionV1 text, trims the five surviving files, removes the dead lint exclusions — deletions, trims and renames never split across commits"
    requirement: LINT-COMPAT-01
    verification:
      - kind: other
        ref: "command: git show --stat 4238fbe (135 files, +538/−6355, 59 deletions) && grep -rnE 'NewLinked|LintLinked|LinkedFileInfo|LinkedRuleName|AppendLinkedIssue|shouldIgnoreLinked|linkedCommentSuppressions|linkedParsedDirective|LinkedRules|\\bLintRule\\b|CheckIsIgnored' internal/ --include='*.go' | wc -l (0)"
        status: pass
    human_judgment: false
  - id: D2
    description: "SC1 golden identity: the renamed engine reproduces the committed 73-finding golden exactly — every field including Message, Position.Filename/Offset, SourceName, in the D-01 order — through TestLintGoldenFullRegistry (the former linked parity drive, re-pointed by the sweep)"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/rules/golden_full_registry_test.go#TestLintGoldenFullRegistry"
        status: pass
    human_judgment: false
  - id: D3
    description: "Phase-2 flagged assumption A1 executed: the class-1 COMPILE cells through actionV1 — a compiling-syntax/non-compiling corpus (proto3 required) prints COMPILE issues and routes to ErrHasLintIssue in process, and the real binary exits 1 with the (COMPILE) line on stdout; RED on the old wiring (machine-validated RED_EVIDENCE_OK on both targets), GREEN in the flip commit"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrix/compile_issue_exit1_class"
        status: pass
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrixExitCodes/compile_error_code_1"
        status: pass
    human_judgment: false
  - id: D4
    description: "SC1 CLI boundary: the real binary's pinned stdout (two path-sorted (RULE) lines) and exit code 1 reproduce byte for byte through the flipped engine over the findings/exclusion/directive/rewrite corpus"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/api/lint_stdout_golden_test.go#TestLintStdoutGolden"
        status: pass
    human_judgment: false
  - id: D5
    description: "SC1 exit-code surface: missing-import exit 2 with the Cannot-import-file carrier, syntax error exit 1, clean corpus exit 0 with empty stdout, rule findings exit 1 — the four pre-existing exit-code cells re-run green through the new engine in the same run as the new class-1 cell"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/api/lint_link_matrix_test.go#TestLintLinkMatrixExitCodes"
        status: pass
    human_judgment: false
  - id: D6
    description: "SC3 subsystem isolation: breaking, generate, mod tidy and migrate suites green inside ONE six-root -race run (core, rules, api, modules, migration, config) with the documented wazero carve-out (./internal/adapters/plugin/, ./internal/generation/ — pre-existing third-party race, ledger item 18, owned before Phase 7); go-protoparser and protocompile remain required in go.mod; breaking-owned files (proto_info_read.go, proto_comments.go) byte-frozen"
    requirement: LINT-COMPAT-02
    verification:
      - kind: other
        ref: "command: go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/api/... ./internal/modules/... ./internal/migration/... ./internal/config/... (all ok, api 186.8s) && go mod tidy && git diff --exit-code go.mod go.sum && grep -c 'go-protoparser' go.mod (1) && grep -c 'protocompile' go.mod (1) && git diff HEAD -- internal/core/proto_info_read.go internal/core/proto_comments.go (empty)"
        status: pass
    human_judgment: false
  - id: D7
    description: "Mechanical closure on the flip commit: schema-gen drift empty (rule names unchanged), golangci-lint 0 issues on ./internal/... with the exclusions removed, easyp lint --help exits 0, working tree carries no unintended artifacts"
    requirement: LINT-COMPAT-02
    verification:
      - kind: other
        ref: "command: go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ && bin/golangci-lint run ./internal/... (0 issues) && go build -o /tmp/easyp-check ./cmd/easyp && /tmp/easyp-check lint --help (exit 0)"
        status: pass
    human_judgment: false
  - id: D8
    description: "Standing constraints survive the flip: TestReadSourceOnce (one readSource per corpus file) green, the suppression-centrality matrix green, the compile_bypasses_suppression IN-02 pin green — all under their swept names inside the suite runs"
    requirement: LINT-COMPAT-01
    verification:
      - kind: integration
        ref: "tests/internal/core/lint_linked_readsource_test.go#TestReadSourceOnce"
        status: pass
      - kind: integration
        ref: "tests/internal/core/lint_linked_matrix_test.go#TestLintMatrix/compile_bypasses_suppression"
        status: pass
      - kind: integration
        ref: "tests/internal/core/lint_suppressions_parity_test.go#TestLintSuppressionParity"
        status: pass
    human_judgment: false

# Metrics
duration: 16 min
started: 2026-11-20T20:14:08Z
completed: 2026-11-20T20:30:06Z
status: complete
---

# Phase 05 Plan 03: The Atomic Flip Summary

**easyp lint now runs only the compiler-backed engine: one atomic commit (4238fbe, 135 files, +538/−6355) deletes the entire old lint contour, collapses the rename map behind the textually unchanged buildCore/actionV1 seam, and lands the class-1 COMPILE cells — with the 73-finding golden, the CLI stdout golden, the exit-code matrix and the six-root subsystem battery all green on the committed flip state**

## Performance

- **Duration:** 16 min
- **Started:** 2026-11-20T20:14:08Z
- **Completed:** 2026-11-20T20:30:06Z
- **Tasks:** 2
- **Files modified:** 135 changed in the flip commit (59 deletions, 76 modified) + the planning artifacts of this SUMMARY

## Accomplishments

- **The milestone's single user-visible change is committed.** `feat(05-03)` `4238fbe` — ONE atomic commit: 44 production files and 15 test files deleted (the old engine `core.Lint`, the old read twin `lintProtoInfoRead`, all 41 old rule implementations, `casing.go`, the old registry constructor `New`, the old reflection comment-suppression walk, `CheckIsIgnored`, and the whole parity harness), the rename map collapsed (12 required renames), the five surviving files trimmed, the two dead `.golangci.yml` exclusions removed, the golden re-pointed, and the class-1 cells landed. `git revert 4238fbe` restores main exactly — the flip is the phase's designed rollback point.
- **The seam flipped without touching the api surface.** `buildCore` keeps its text (one seam comment added, zero removed lines vs f27d2d1), `lint_v1.go` is zero-diff vs f27d2d1, and `actionV1`'s `apps[key].Lint(...)` now binds the linked engine because the old `Lint` died in the same commit that renamed `LintLinked` to `Lint`. `rules.New(cfg)` binds the linked registry the same way. One rule surface in `core.Core` (`rules []Rule` wired from `Options.Rules`; the dormant `Options.LinkedRules` deleted), one name function (`GetRuleName`, the override mechanism retained until 05-04), one registry constructor.
- **The compat gates prove identity.** The 73-finding full-registry golden reproduces through the renamed engine (`TestLintGoldenFullRegistry`, full-struct compare incl. Message/Filename/Offset, D-01 order, purity guards); the CLI stdout golden is byte-identical (`TestLintStdoutGolden`); the exit-code matrix re-runs green (missing-import 2, syntax 1, clean 0, findings 1); the 41 carried per-rule tests and the suppression matrix survive under their swept names; `TestReadSourceOnce`, suppression centrality, and the `compile_bypasses_suppression` IN-02 pin are green.
- **Phase-2 flagged assumption A1 is executed.** The class-1 COMPILE cells (in-process `compile_issue_exit1_class` asserting `ErrHasLintIssue` through `actionV1`; subprocess `compile_error_code_1` asserting exit 1 with the `(COMPILE)` line) were written first, run against the old wiring, and failed exactly on their own planned assertions — RED evidence machine-validated (`RED_EVIDENCE_OK` on both records) — then went GREEN in the flip commit. The old engine parsed proto3-required without complaint (exit 0, no findings); the linked engine classifies it as COMPILE in the existing print format.
- **SC3 holds.** The six-root same-run `-race` battery (`./internal/core/... ./internal/rules/... ./internal/api/... ./internal/modules/... ./internal/migration/... ./internal/config/...`) is fully green — breaking, generate, mod tidy and migrate suites all inside one run — with the documented wazero carve-out (`./internal/adapters/plugin/`, `./internal/generation/` excluded: pre-existing third-party race, ledger item 18, owned before Phase 7, established Phase 2–4 practice). `go.mod` keeps go-protoparser and protocompile; `go mod tidy` diff-empty; schema-gen drift empty; `bin/golangci-lint run ./internal/...` reports 0 issues with the exclusions removed; `easyp lint --help` exits 0.
- **The dormancy audit inverted.** The Phase-3/4 audit ("old path byte-frozen over 44 files") is meaningless post-flip — the old path no longer exists. The verifier's audit contract is now: (a) the breaking-path freeze gate (`git diff --exit-code -- internal/core/proto_info_read.go internal/core/proto_comments.go`), (b) the deletion-completeness greps (zero LIVE references to every deleted symbol; the grep-gate over `NewLinked|LintLinked|LinkedFileInfo|LinkedRuleName|AppendLinkedIssue|shouldIgnoreLinked|linkedCommentSuppressions|linkedParsedDirective|LinkedRules|\bLintRule\b|CheckIsIgnored` = 0 across internal/), and (c) the subsystem suites. **Note for the verifier: running the old dormancy audit command will fail confusingly by design** — the files it diffs are deleted.

## Task Commits

Each task was committed atomically:

1. **Task 1: The atomic flip commit — delete old path, collapse rename map, land class-1 cells** - `4238fbe` (feat; 135 files: 59 deletions, 76 modified, +538/−6355)
2. **Task 2: Post-flip battery — SC3 subsystem suites + mechanical closure on the flip commit** - verification-only (six-root -race suite, schema drift, golangci, tidy, freeze, lint --help — all green on the committed state; no file changes, no commit)

**Plan metadata:** docs commit (this SUMMARY + STATE/ROADMAP/REQUIREMENTS/WINDOWS + the two RED evidence records).

## Files Created/Modified

- `internal/core/lint.go`, `internal/core/lint_proto_info.go`, `internal/rules/casing.go` + 41 old rule files - DELETED (the old lint contour)
- 10 parity test files, `check_lint_ignore_test.go`, `registry_parity_guard_test.go`, `casing_test.go`, `internal/core/comment_suppressions_test.go`, `comment_suppressions_eof_test.go` - DELETED (died with their oracle)
- `internal/core/lint_linked.go` - THE engine: `LintLinked` renamed `Lint`, `shouldIgnoreLinked` renamed `shouldIgnore`, reads the single `rules` field; the parity-pin wrapper literal byte-identical
- `internal/core/lint_model.go` - canonical contract names (`Rule`, `ProtoInfo`, `AppendIssue`, `GetRuleName`); the executed rename-map header deleted
- `internal/core/lint_linked_read.go` - gains the moved `close` (body verbatim from the deleted lint.go)
- `internal/core/core.go` - one rule surface (`rules []Rule` from `Options.Rules`)
- `internal/core/dom.go` - old `Rule` interface, old `ProtoInfo`, 5-arg `AppendIssue`, `buildError`, old `GetRuleName` deleted; everything breaking consumes kept
- `internal/core/comment_suppressions.go` - old reflection walk + `parsedDirective` deleted; the breaking-path helpers kept
- `internal/core/link_test.go`, `internal/core/proto_info_read_test.go` - deleted-oracle callers re-derived (see deviations)
- `internal/rules/builder.go` - old `New` deleted; the five `add*` helpers re-pointed to the 41 port types (name output unchanged via the `LintRuleName()` overrides); `config` import dropped
- `internal/rules/builder_linked.go` - `NewLinked` renamed `New` — THE registry; the 41-port-type rename map kept for 05-04
- `internal/rules/lint_helpers.go` - nine kept-unused helpers deleted; `groupedValues`+`add`+`firstExemptMismatches` kept
- `internal/rules/init_test.go` - `start`/`parseFile` deleted; `parseLinkedFile`/`validate` and the testdata constants kept
- `internal/rules/golden_full_registry_test.go` - single canonical golden drive `TestLintGoldenFullRegistry`; the old capture drive and its local D-01 replica sort deleted
- `internal/api/lint_link_matrix_test.go` - the class-1 COMPILE cells (in-process + subprocess)
- `internal/api/lint_stdout_golden_test.go` - twin fixture corrected (unique message names); pinned literals unchanged
- `internal/api/runtime.go` - one seam comment; zero removed lines vs f27d2d1
- `.golangci.yml` - `internal/rules/lint_helpers.go` paths entry + comment block and the casing.go QF1001 rule removed; the easypc exclusion and ST1005 rule kept
- 41 `*_linked.go` port files + carried per-rule tests + engine tests - mechanical symbol sweep (13 rename patterns + the four `TestLintLinked*` function renames)

## TDD Gate Compliance

- Plan `type: execute` (not `type: tdd`); no task carries `tdd="true"`. The orchestrator's tdd_mode directive asked for RED evidence on the class-1 COMPILE Action cells before the flip commit.
- **RED:** captured pre-commit against the old wiring — `go test -race -count=1 ./internal/api/ -run TestLintLinkMatrix -v` exit 1 with exactly the two new cells failing on their own planned assertions (`compile_issue_exit1_class`: expected `ErrHasLintIssue`, got nil; `compile_error_code_1`: expected exit 1, got 0) while all pre-existing matrix cells passed. Both evidence records machine-validated `RED_EVIDENCE_OK` (`gsd_run check tdd-red-evidence`): `.planning/phases/05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti/tdd-red-evidence-05-03-task1.json` and `...-task1-subprocess.json`.
- **GREEN:** the flip commit `feat(05-03)` `4238fbe` — all class-1 cells green through the linked engine, pre-existing cells unchanged.
- **REFACTOR:** not needed — the sweep is the mechanical pass, gated by the same suite.
- **Deviation from the canonical commit split (documented):** no separate `test(05-03)` RED commit exists. The plan's RATIFIED atomic-flip mandate (research Pitfall 1 + the file-surface excess assumption: "ONE commit — never split deletion from rename across commits"; "the commit is ONE atomic commit containing all deletions, trims and renames") requires the cells, the deletions and the renames to land together — the cells are "physically unreachable until this commit exists" (the plan's own words), so a committed RED would carry failing tests on a tree state that never ships. The machine-validated pre-commit RED records are the evidence trail.

## Decisions Made

- The flip executed directly on the working tree with per-step verification and a full git-restore safety net (every mutation touches tracked files only), rather than a scratch-clone rehearsal: the risky mechanical piece (the 13-pattern sweep) was validated by exact per-pattern match counts (8/41/11/3/140/105/9/44/7/4/5/8/4) plus immediate `go build ./...`, and the sweep's `\b` word boundaries were verified to protect `LintRuleName` (the override mechanism survives with 88 lines intact until 05-04).
- The three permanent exceptions keep their names as planned: `linkedFilesFromImport`/`linkedFileFromImport` (breaking owns the unsuffixed read twin) and `linkedDirectiveComment` (breaking owns `directiveComment` on `*parser.Comment`). The breaking-owned symbols were NOT renamed to free them (REQ-5.2 freeze respected — the frozen files are zero-diff).
- The `parity-pin` wrapper literal `fs.WalkDir: c.lintProtoInfoRead:` stays byte-identical with its comment, and the five exact-string proof cells keep asserting it — the main-wins chain text is user-visible behavior, unchanged by the flip.
- Light comment refreshes where the flip made prose factually wrong (the `shouldIgnore` mirror justification and its dead rename-map reference; `builder_linked.go`'s dormancy paragraph — the rename MAP inside it kept verbatim for 05-04; the golden file's two-drive header rewritten to the single-drive reality). Comment-only, no behavior text.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Gate-3 twin fixture cannot compile under the linked engine — corrected with a pre-flip re-capture**
- **Found during:** Task 1 (commit gate — `TestLintStdoutGolden` failed)
- **Issue:** the 05-01 CLI stdout golden corpus declares the same `CleanFile` message in two same-package files (`Bad_File.proto` and `Excluded_File.proto`). The old per-file engine lints files independently and never notices; the linked engine links the whole corpus, hits the redeclaration, routes the run to COMPILE (`"CleanFile" is already defined.`), and the golden diverges completely — research Pitfall 7 ("golden corpus divergence through linked-engine type resolution") applied to Gate 3, whose fixture the 05-01 capture never needed to compile.
- **Fix:** renamed `Excluded_File.proto`'s message to `ExcludedMsg` (unique names; the file-NAME violation the golden pins is untouched). To keep the golden-master methodology intact, the pinned stdout was re-captured from the byte-frozen PRE-flip binary — built from flip-HEAD (`git archive HEAD | tar -x`) before the flip commit — over the corrected fixture: the output is byte-identical to the 05-01 literals (md5-stable across 3 runs), and the flipped binary reproduces it. The golden literals did not change; only the fixture did.
- **Files modified:** internal/api/lint_stdout_golden_test.go (fixture string + explanatory comment)
- **Verification:** TestLintStdoutGolden green post-flip; old-binary capture md5 89616e4f0353f5c063da8e5ff2699f97 ×3
- **Committed in:** 4238fbe

**2. [Rule 3 - Blocking] Two test files outside the plan's declared surface call the deleted `lintProtoInfoRead` directly**
- **Found during:** Task 1 (pre-flight grep of every deleted symbol's references)
- **Issue:** `internal/core/proto_info_read_test.go` and `internal/core/link_test.go` call `engine.lintProtoInfoRead(...)` — the old read twin the flip deletes — but neither file is in the plan's `files_modified` list; without edits the package does not compile (the plan's grep-gate criterion also requires their references gone).
- **Fix:** `TestProtoInfoReadTwinsMirroring` and `TestLintProtoInfoReadErrorLabels` deleted with their oracle (the D-06 twin-mirror proof's value is fully transferred — the linked engine has its own read path with `TestReadSourceOnce` and the suppression matrix); `link_test.go`'s two fallback cells (`transitive_missing_depth2`, `blocked_import`) re-pointed from `lintProtoInfoRead` to `protoInfoFromSource` — the per-file read path that survives the flip — preserving the assertions verbatim in intent: same `openImportFile` funnel, same typed `*OpenImportFileError` carrier (`missing.proto` / `b.proto`), source read from `walker.RootPath()`.
- **Files modified:** internal/core/proto_info_read_test.go, internal/core/link_test.go
- **Verification:** ./internal/core/... green under -race; both cells assert the same OpenImportFileError carriers as before
- **Committed in:** 4238fbe

**3. [Rule 3 - Blocking] Acceptance-criterion baselines reconciled with the plan's own prohibitions (verification plumbing)**
- **Found during:** Task 1 (criteria re-run)
- **Issue:** (a) the `lintProtoInfoRead` criterion expects exactly 1 occurrence ("the parity-pin literal only"), but the plan simultaneously mandates the byte-identical pin COMMENT (which names the callee), the frozen `proto_info_read.go` (whose comment says "the breaking path's own twin of lintProtoInfoRead" — zero-edit constraint), and the exact-string proof cells that would fail if the pin text changed; the reachable minimum is 8 lines. (b) the runtime.go removed-lines gate greps `grep -c '^-'`, which counts the diff header line `--- a/internal/api/runtime.go` itself — impossible to satisfy at 0 with the mandated seam comment present. (c) the zero-ref grep's recorded HEAD baseline (384) predates 05-02's cells (actual 399).
- **Fix:** (a) enforced the stronger, intent-true invariant — zero LIVE references: `grep -rn 'lintProtoInfoRead' internal/ --include='*.go' | grep -vE 'fs\.WalkDir: c\.lintProtoInfoRead:|callee name c\.lintProtoInfoRead|twin of lintProtoInfoRead'` = 0 (the 8 remaining lines are exactly: pin comment + pin literal, the frozen-file comment, and 5 proof-cell assertions); (b) corrected plumbing `grep -c '^-[^-]'` = 0 (comment-only addition proven); (c) EXPECT==0 achieved regardless of baseline drift. All three recorded in the WINDOWS ledger.
- **Files modified:** none (verification plumbing + documentation only)
- **Verification:** all criteria re-run green on the committed state
- **Committed in:** n/a — no code change

**4. [Rule 3 - Blocking] Unused imports left behind by the trims (mechanical fallout)**
- **Found during:** Task 1 (first `go build ./...` after the trims)
- **Issue:** deleting old `New` (builder.go) and the old suppression walk (comment_suppressions.go) orphaned their `fmt`/`reflect`/`strconv`/`strings`/`protoast` imports.
- **Fix:** dropped the orphaned imports; gofmt clean.
- **Files modified:** internal/rules/builder.go, internal/core/comment_suppressions.go, internal/rules/init_test.go
- **Verification:** go build ./... green; golangci 0 issues
- **Committed in:** 4238fbe

---

**Total deviations:** 4 auto-fixed (4 Rule 3 blocking: 1 fixture correction with re-capture proof, 1 undeclared-surface compile fix, 1 criterion-baseline reconciliation, 1 mechanical import fallout)
**Impact on plan:** All four were necessary for the flip's gates to hold with the plan's own prohibitions intact; the golden literals, the parity-pin text and the frozen files are untouched; no scope creep beyond the flip's surface.

## Issues Encountered

- `TestLintStdoutGolden` failed at the first post-sweep commit-gate run — the flip's compat net doing exactly its job. Root-caused to the Gate-3 fixture's duplicate message name (deviation 1) within one probe cycle: the old binary re-capture over the corrected fixture proved the divergence was fixture-borne, not engine behavior drift.
- The pre-existing dirty files in the working tree (`.agents/skills/sdd/scripts/pipeline.sh`, `.planning/phases/03-…/03-VERIFICATION.md`, untracked `.planning/intel/`, `.planning/research/.cache/`, `.planning/state.json`, `.spec/features/`) — out of scope per the deviation scope boundary; left untouched and uncommitted.
- The sealed vendored tree `internal/easypc` is not gofmt-clean in its committed state (pre-existing, untouched, out of scope — AGENTS.md forbids edits there); every package this plan touched is gofmt-clean.

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components. The flip deletes code and renames identifiers; every surviving test drives real engine behavior.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- **05-04 (mechanical consolidation) inherits the post-flip tree:** one engine, one registry (`rules.New` = the linked constructor), one rule surface, the `LintRuleName()` override mechanism still in place (88 lines, protected by the sweep), and the 41-port-type rename map kept verbatim in `builder_linked.go`'s header. Its work is purely mechanical: 41 type renames, override drops, and the remaining header deletions — name-neutrality proven by `TestRegistryParity` (its guard died with the flip; `TestAllRuleNames`/`TestAllGroups` + the golden gate are the post-rename name surface).
- **The verifier's audit contract is inverted** (documented above): breaking-path freeze + deletion-completeness + subsystem suites; the old dormancy audit command fails by design.
- **Rollback remains one command:** `git revert 4238fbe` restores main exactly; the golden literals and carried tests prove behavior identity on both sides of the commit.
- No blockers.

---
*Phase: 05-pereklyuchenie-lint-puti-i-vorota-sovmestimosti*
*Completed: 2026-11-20*
