---
phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli
plan: 03
subsystem: core
tags: [protobuf, protoast, positions, refactoring, read-model, tdd, architecture]

requires:
  - phase: 01-01
    provides: vendored internal/protoast (adaptation seam) whose Position type Issue now aliases
  - phase: 01-02
    provides: protoast rune-column position contract and readProtoFileWithDirectives golden test (helper preserved untouched)
provides:
  - core.Position = protoast.Position alias in internal/core/dom.go — Issue.Position already speaks the type later phases emit
  - fromMetaPos(meta.Position) Position — the single field-for-field conversion, applied only in AppendIssue/buildError and buildBreakingCheckIssue
  - internal/core/lint_proto_info.go — lint-owned (c *Core) lintProtoInfoRead(ctx, fs, path) (ProtoInfo, error)
  - internal/core/proto_info_read.go — breaking-owned BreakingProtoInfo type + (c *Core) breakingProtoInfoRead; shared parse helpers unchanged
  - twin-mirror parity tests (TestProtoInfoReadTwinsMirroring + error-label suites) proving both read models behaviorally identical in Phase 1
affects: [01-04, phase-3-lint-engine, phase-4-rules, phase-5-flip]

actuals:
  tokens: 12820   # chars/4 over the realized diff (51,283 diff chars, 51 files, +289/-122)
  tasks: 2
  commits: 3
  plan_head_before: 6563fabe09f0eacad64673fc76d87b5a52e19ab9
  plan_head_after: 0d7d8cb

tech-stack:
  added: []   # zero new dependencies (LINT-VEND-02 preserved; tidy zero-diff re-proven)
  patterns:
    - "Single conversion boundary: go-protoparser positions enter as meta.Position at exactly two funnels (AppendIssue/buildError, buildBreakingCheckIssue) and become protoast.Position there — callers pass .Meta.Pos unchanged"
    - "Twin-mirror split (D-06): two behaviorally identical read constructors with distinct owned types; shared parse helpers stay put so both paths and the 01-02 golden test keep one parse surface"

key-files:
  created:
    - internal/core/lint_proto_info.go
  modified:
    - internal/core/dom.go
    - internal/core/breaking_checker.go
    - internal/core/breaking_profiles.go
    - internal/core/proto_info_read.go
    - internal/core/breaking_check.go
    - internal/core/lint.go
    - internal/core/proto_info_read_test.go
    - internal/core/breaking_target_paths_test.go
    - internal/core/breaking_checker_test.go
    - internal/rules/*_test.go (41 files, literal type swap only)

key-decisions:
  - "Commits landed on main per the sequential-executor dispatch (branching_strategy 'none'; 01-01/01-02 precedent), although the default-branch guard query reports main as protected — documented override, config not modified"
  - "RED used zero-value stubs (lintProtoInfoRead/breakingProtoInfoRead returning zero values) so the twin tests compile and fail ON planned assertions — canonical RED step 4 ('make the test compile and fail on its assertion'); a build-failure run is INVALID_RED and cannot authorize GREEN"
  - "GREEN commit uses feat(01-03) instead of the plan's stated refactor(01-03) message — the canonical tdd.md commit-scope contract (test → feat → refactor) owns TDD-plan commit types; the plan's description is retained verbatim"
  - "meta.Position keeps a String() method vs protoast.Position's identical-output String() — both render Filename:Line:Column with <input> fallback, so no formatting behavior changes (checked before the swap)"

patterns-established:
  - "Compile-atomic scripted sweep: a type change that breaks 56 test literals lands with ONE grep|xargs sed pipeline over the located file list + scripted import fix; per-file manual editing is prohibited"
  - "Read-model twin ownership: lint and breaking each construct their own type via identically-shaped constructors; equality between twins is asserted by a test so future divergence is a deliberate act"

requirements-completed: [LINT-VEND-01]

coverage:
  - id: D1
    description: "Issue.Position is protoast.Position via the core.Position alias with a single fromMetaPos conversion boundary (AppendIssue/buildError + buildBreakingCheckIssue); AppendIssue keeps its meta.Position parameter so all 48 rule call sites and ~20 breaking call sites are untouched"
    requirement: LINT-VEND-01
    verification:
      - kind: other
        ref: "grep counts: type Position = protoast.Position ==1, fromMetaPos ==4, meta.Position in dom.go ==2 (AppendIssue+fromMetaPos params), meta.Position in breaking_profiles.go ==0"
        status: pass
      - kind: unit
        ref: "go test -race -count=1 ./internal/core/... ./internal/rules/... → all ok"
        status: pass
      - kind: other
        ref: "go build ./... → exit 0 (whole module, both funnels compile)"
        status: pass
    human_judgment: false
  - id: D2
    description: "All 56 expected-position test literals (41 internal/rules test files + breaking_checker_test.go) swapped meta.Position{ → protoast.Position{ by one scripted sweep with provably identical numeric values — diff-of-diffs over Line/Column/Offset numbers is empty plan-wide"
    requirement: LINT-VEND-01
    verification:
      - kind: other
        ref: "grep -rn 'meta\\.Position{' internal --include='*_test.go' | wc -l → 0; protoast.Position{ → 56; git diff --name-only -- '*_test.go' → exactly 42 files"
        status: pass
      - kind: other
        ref: "diff-of-diffs gate on plan diff 6563fab..HEAD: multiset of removed == added numeric (Line|Column|Offset) values → empty diff"
        status: pass
      - kind: unit
        ref: "go test -race -count=1 ./internal/rules/... → ok (all 41 rule test files with identical expectations)"
        status: pass
    human_judgment: false
  - id: D3
    description: "D-06 split: lint owns ProtoInfo + lintProtoInfoRead, breaking solely owns BreakingProtoInfo + breakingProtoInfoRead; no shared read-model type or alias between the paths; bare protoInfoRead name gone; api layer needed no new knobs"
    requirement: LINT-VEND-01
    verification:
      - kind: unit
        ref: "go test -count=1 ./internal/core/ -run 'TestProtoInfoReadTwinsMirroring|TestLintProtoInfoReadErrorLabels|TestBreakingProtoInfoReadErrorLabels' → 3/3 PASS (twin equality + error-label parity both twins)"
        status: pass
      - kind: other
        ref: "grep counts: bare protoInfoRead ==0; lintProtoInfoRead ==7; breakingProtoInfoRead ==7; api/ references ==0; readProtoFileWithDirectives file untouched (empty git diff for proto_comments.go)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Empty behavior diff proven: every package suite green under -race with identical numeric expectations; deps and lint clean (tidy zero-diff, golangci 0 issues). Full ./... -race run has exactly one red spot: the pre-existing documented wazero race in internal/adapters/plugin (01-01 deferred item, WINDOWS ledger id 1), reproduced on pristine HEAD and unrelated to this plan"
    requirement: LINT-VEND-01
    verification:
      - kind: unit
        ref: "go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/protoast/... → ok; full ./... run → 40 packages ok, only internal/adapters/plugin fails (documented pre-existing wazero data race)"
        status: pass
      - kind: other
        ref: "bin/golangci-lint run ./... → 0 issues; go mod tidy && git diff --exit-code go.mod go.sum → empty"
        status: pass
    human_judgment: false

duration: 30 min
completed: 2026-11-18
status: complete
---

# Phase 1 Plan 03: Вендоринг компилятора и расчленение модели — расщепление read-моделей Summary

**Issue.Position теперь protoast.Position за одним конверсионным рубежом (fromMetaPos), а lint и breaking больше не делят модель чтения: lint владеет lintProtoInfoRead→ProtoInfo, breaking — BreakingProtoInfo/breakingProtoInfoRead, и всё это доказано пустым диффом поведения — 56 тестовых литералов поменялись только именем типа при побитово равных числах, полный -race набор зелёный.**

## Performance

- **Duration:** 30 min (2026-11-18T20:57:48Z → 2026-11-18T21:27:40Z)
- **Started:** 2026-11-18T20:57:48Z
- **Completed:** 2026-11-18T21:27:40Z
- **Tasks:** 2 (Task 1 auto — one compile-atomic commit; Task 2 tdd — RED + GREEN commits)
- **Files modified:** 51 (10 in internal/core, 41 internal/rules test files)

## TDD Gate Compliance

**RED** — `test(01-03)` commit `d24e8a8`: twin-mirror tests added to `internal/core/proto_info_read_test.go` (TestProtoInfoReadTwinsMirroring + error-label suites for both twins) plus minimal zero-value stubs so the package compiles. Run `go test -race -count=1 ./internal/core/`: all 9 tests (3 top-level + 6 subtests) failed ON planned assertions — Path expected `item.proto` actual `""`; "An error is expected but got nil" in every failure-path subtest. Classified by `gsd_run check tdd-red-evidence`: **RED_EVIDENCE_OK / target_test_failed** (exit 1, format go-test). Semantic assessment: targets executed and failed because the twin read functions were unimplemented stubs — the intended reason; no build failure, zero-test discovery, or fixture crash involved. Evidence record removed after classification as a transient gate artifact (01-02 precedent).

**GREEN** — `feat(01-03)` commit `0d7d8cb`: real twin implementations (lint twin in new `lint_proto_info.go`, breaking twin replacing `protoInfoRead`), callers and labels updated, `readProtoFiles`/`collect` moved to `[]BreakingProtoInfo`. All twin tests green; one test-side type fix during GREEN (`PackageName("twins")` instead of bare string in the equality). Same commit switches `breaking_target_paths_test.go` to the breaking type (forced by `collect`'s signature — see deviations).

**REFACTOR** — none needed; the GREEN implementation is the minimal split (no refactor commit, per discipline).

**Gate sequence:** RED `d24e8a8` verified ancestor of GREEN `0d7d8cb` (`git merge-base --is-ancestor`). No `feat(01-03)` commit precedes the RED commit.

## Accomplishments
- `core.Position = protoast.Position` alias + `fromMetaPos` field-for-field conversion landed in `dom.go`; `AppendIssue` keeps its `meta.Position` parameter, so all 48 rule call sites (`internal/rules/*`) and both rule-file literals (`file_lower_snake_case.go:30`, `package_defined.go:24`) needed zero changes.
- Breaking conversion mirrored: `buildBreakingCheckIssue` keeps its `meta.Position` parameter and converts once; the `breaking_profiles.go` descriptor-path literal swapped to `protoast.Position` with identical values.
- One scripted sweep (`grep|xargs sed` + scripted import fix + `gofmt`) swapped all 56 `meta.Position{` test literals across exactly 42 test files — zero manual per-file edits, multiset of numeric values provably unchanged.
- D-06 split landed: `lintProtoInfoRead` (lint-owned, new file) and `BreakingProtoInfo`/`breakingProtoInfoRead` (breaking-owned) with twin-mirror parity tests; shared parse helpers (`readProtoFile`, `readProtoFileWithDirectives`, `readFilesFromImport`, `readFileFromImport`, `openImportFile`, `sourceImportPath`, `openDiskSource`) untouched — 01-02's golden test and `proto_info_read_test.go` harness keep compiling.
- Full gates green: module builds, `golangci-lint` 0 issues, `go mod tidy` zero-diff (LINT-VEND-02 control re-proven), 40 packages ok under `-race` (single pre-existing exception below).

## Task Commits

Each task was committed atomically:

1. **Task 1: Issue.Position becomes protoast.Position via single conversion boundary** - `9fb2341` (refactor) — 45 files, +124/-104
2. **Task 2 RED: twin-mirror read-model parity tests** - `d24e8a8` (test) — 3 files, +124 (incl. compile-enabling stubs)
3. **Task 2 GREEN: split lint/breaking read models (D-06)** - `0d7d8cb` (feat) — 6 files, +50/-27

**Plan metadata:** docs commit (see below).

## Files Created/Modified
- `internal/core/dom.go` - `type Position = protoast.Position`, `fromMetaPos`, `Issue.Position` swapped, `buildError` takes Position; `AppendIssue` parameter list unchanged
- `internal/core/breaking_checker.go` - `buildBreakingCheckIssue` converts via `fromMetaPos`; ~20 `.Meta.Pos` call sites untouched
- `internal/core/breaking_profiles.go` - descriptor-path Issue literal swapped to `protoast.Position` (values identical); imports adjusted
- `internal/core/lint_proto_info.go` - NEW: lint-owned `(c *Core) lintProtoInfoRead(ctx, fs, path) (ProtoInfo, error)`
- `internal/core/proto_info_read.go` - `BreakingProtoInfo` type + breaking-owned `breakingProtoInfoRead`; `protoInfoRead` deleted; shared helpers unchanged
- `internal/core/lint.go` - lint caller: `c.lintProtoInfoRead` with label `c.lintProtoInfoRead: %w`
- `internal/core/breaking_check.go` - breaking caller: `c.breakingProtoInfoRead` with label `c.breakingProtoInfoRead: %w`; `readProtoFiles`→`[]BreakingProtoInfo`, `collect` takes `[]BreakingProtoInfo`
- `internal/core/proto_info_read_test.go` - twin-mirror parity + error-label tests (both twins)
- `internal/core/breaking_target_paths_test.go` - switched to `BreakingProtoInfo` (collect signature)
- `internal/core/breaking_checker_test.go` + 41 `internal/rules/*_test.go` - literal type swap only (`meta.Position{` → `protoast.Position{`), values identical

## Decisions Made
- Zero-value stubs as RED scaffolding (see TDD Gate Compliance) — canonical RED requires a compiling package that fails on planned assertions.
- GREEN commit prefixed `feat(01-03)` per canonical tdd.md commit-scope contract, plan description kept.
- Pre-swap behavior check: `meta.Position.String()` and `protoast.Position.String()` produce identical output (`Filename:Line:Column`, `<input>` fallback) — confirmed before relying on the alias for output-critical types.
- Commits on main per sequential-executor dispatch (documented override, matching 01-01/01-02; branching_strategy "none"; config not modified).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] breaking_target_paths_test.go switched to BreakingProtoInfo**
- **Found during:** Task 2 GREEN (compile after `collect` signature change)
- **Issue:** `TestBreakingTargetPathsTakePrecedenceOverImportAliases` constructs `[]ProtoInfo` and passes it to `collect`, whose planned signature change to `[]BreakingProtoInfo` breaks that test's compilation. The plan's `files_modified` did not list this test.
- **Fix:** The test now constructs `[]BreakingProtoInfo` — it tests the breaking-owned collector, so the breaking type is the correct owner under D-06.
- **Files modified:** internal/core/breaking_target_paths_test.go
- **Verification:** `go build ./...`, `go vet ./internal/...`, full core suite green
- **Committed in:** 0d7d8cb (part of Task 2 GREEN commit)

**2. [Process - TDD discipline] GREEN commit message prefix feat(01-03) instead of plan's refactor(01-03)**
- **Found during:** Task 2 GREEN
- **Issue:** The plan's action states commit message `refactor(01-03): split lint/breaking read models (D-06)`, but the plan's own `tdd="true"` makes the canonical `gsd-core/references/tdd.md` commit-scope contract authoritative: RED→`test(...)`, GREEN→`feat(...)`.
- **Fix:** GREEN committed as `feat(01-03): split lint/breaking read models (D-06)` — plan description verbatim, prefix per TDD contract; keeps the RED→GREEN gate sequence machine-checkable.
- **Files modified:** none (commit metadata only)
- **Verification:** gate validation greps find `test(01-03)` `d24e8a8` before `feat(01-03)` `0d7d8cb`, ancestry confirmed
- **Committed in:** 0d7d8cb

---

**Total deviations:** 2 auto-fixed (1 blocking Rule 3, 1 process/TDD-contract)
**Impact on plan:** None on scope or behavior; both adjustments are forced or mandated by the plan's own TDD gate contract.

## Issues Encountered
- **Acceptance-criterion count artifact (A4, Task 2):** `grep -c 'readProtoFileWithDirectives' internal/core/proto_comments.go` measures **2** (godoc comment line + definition line), not the planned `== 1`; the HEAD baseline measures 2 as well, so the criterion's expectation was a count-vs-lines artifact at planning time. The criterion's intent — the shared helper is untouched (parallel-wave safety with 01-02) — is proven by an **empty git diff** for `proto_comments.go`. No code change needed.
- **Full-module `-race` sweep fails only in `internal/adapters/plugin`** (wazero `GetWazeroVersion` data race) — the pre-existing defect documented in 01-01's deferred-items.md and the WINDOWS.md ledger (id 1), re-confirmed pre-existing on pristine HEAD in both prior plans. All other 40 packages pass `-race`, including everything this plan touches. Not fixed (executor scope boundary); no duplicate ledger entry since id 1 already tracks it.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 3/5 inherit a clean boundary: lint and breaking no longer share a read-model type, and `Issue.Position` already speaks `protoast.Position` — the parser-level flip (Phase 3/5) needs no type unwind.
- For 01-04 (same wave): `protoInfoRead` is gone (its 5 HEAD references now split across the twins); 01-04 should reference `lintProtoInfoRead`/`breakingProtoInfoRead` names.
- LINT-VEND-01 remains governed by the shared-ID gate (01-04 also declares it and has no SUMMARY yet); it flips when the last sibling summary lands.
- The deferred wazero race remains the only red spot in a full `-race` run on this machine.

---
*Phase: 01-vendoring-kompilyatora-i-raschlenenie-modeli*
*Completed: 2026-11-18*

## Self-Check: PASSED

SUMMARY on disk; internal/core/lint_proto_info.go present; commits 9fb2341, d24e8a8, 0d7d8cb all ancestors of HEAD; ledger-measured commit count 3 (nonzero, matches code changes); plan-wide numeric values gate empty; no file deletions in any plan commit; no stray untracked files from this plan.
