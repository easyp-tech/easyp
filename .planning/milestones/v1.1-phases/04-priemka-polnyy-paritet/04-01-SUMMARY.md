---
phase: 04-priemka-polnyy-paritet
plan: 01
subsystem: testing
tags: [acceptance, full-battery, race-detector, acceptance-gates, golden-94, position-parity, lint-gate, schema-drift, seal, evidence-only, easyprotoc]

# Dependency graph
requires:
  - phase: 03-snyatie-seal-infrastruktury-i-dokumentatsiya
    provides: "Phase-3-final production/test tree at 6f52840 (seal infrastructure removed, live docs updated) from which HEAD has a zero non-.planning diff"
  - phase: 02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii
    provides: "Public easyprotoc module consumption and the measured 34-test-package / 3-no-test battery baseline"
provides:
  - "Attributable acceptance evidence that HEAD (740bce9, docs-only descendant) carries the phase-3-final code tree 6f52840: zero non-.planning diff, clean tracked tree before and after every run"
  - "One full -race battery green: exit 0, 34 ok packages, 0 FAIL, 0 '--- FAIL', 0 race blocks, exactly one known environment skip, 3 no-test packages"
  - "All eleven named parity gates green by exact top-level RUN/PASS: TestAcceptanceGates (Gates A-D x2 against 1e9a596), TestLintGoldenFullRegistry (94), both D-07 position gates, the 43-line core matrix and the 16-line CLI matrix"
  - "Quality predicates green: bin/golangci-lint exactly `0 issues.`, zero schema drift from schema-gen, empty tidy diff, exact pins, zero replace/vendor, empty live-surface seal grep"
  - "Zero-expectation-edit proof: internal/api, internal/rules, internal/core and internal/protoast are byte-identical to 6f52840; both pinned oracle constants occur exactly once"
  - "Honest v1.0 delta record: 41 -> 34 test packages (7 deleted with the vendored tree by design) and 3718 gotestsum events -> 3602 verbose PASS / 3603 RUN, with both counting methods stated"
affects: [milestone-close-v1.1]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
# Scale: chars/4 over the realized authored diff (this SUMMARY; no source, test, config or schema file changed).
actuals:
  tokens: 6372
  tasks: 3
  commits: 0
  plan_head_before: 740bce983c84aebe6a924524ec3795ab6f08ac9c
  plan_head_after: 740bce983c84aebe6a924524ec3795ab6f08ac9c

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Acceptance-evidence plan: no source or expectation edit; every gate claimed by exact top-level `=== RUN` plus `--- PASS`, with a distinct /tmp log and measured counts"
    - "Attribution by exact subject plus a path-qualified zero-diff (`:(exclude).planning`) rather than by counting commits"
    - "Battery-shape assertion from one `-v` log: package-line counts, FAIL/race/skip counts, and no-test counts, with a single allow-listed environment skip"
    - "Oracle-freeze proof: single-occurrence pinned constants plus a zero diff of the four gate packages versus the phase-3-final commit"

key-files:
  created:
    - .planning/phases/04-priemka-polnyy-paritet/04-01-SUMMARY.md
  modified: []

key-decisions:
  - "Evidence-only execution: all three tasks are gate runs whose only artifact is this SUMMARY, so there are no per-task commits and `commits` measures 0 at SUMMARY-write time (02-02 precedent); the plan's single metadata commit carries this file"
  - "The battery ran exactly once with the ratified `-race -count=1 -timeout 20m -v` flags from a clean tracked tree; no second ./... run was performed, including for the auto-mode tracer re-check"
  - "The auto-mode tracer feedback gate was satisfied by the Task-1 verify corpus that had just executed and passed; re-running it would have violated the plan's one-battery prohibition and T-04-28"
  - "The stale plan ledger (21f6b96, plan-creation commit) was replaced with the true execution base 740bce9 before any evidence work, following the 02-02 precedent"
  - "`.claude/CLAUDE.md` is recorded as an accepted Phase-3 D-03 exclusion (live GSD-generated mirror, tracked, line 16 still mentions `compiler-lab/bench/lab.py`); the seal gate claims only the live surfaces it actually covers"
  - "The ROADMAP Phase-4 criterion's `41+ packages / >=3718 tests` phrasing predates the deletion; the phase's own acceptance note (D-06) and the plan's must-have (exactly 34) govern, and the delta is recorded rather than treated as a regression"

patterns-established:
  - "Gate evidence table: command, exit code, exact RUN-line presence, PASS/FAIL line counts, observed pinned constant, log path"
  - "Battery-shape gate: 34 ok / 0 FAIL / 0 race / 3 no-test with exactly one allow-listed environment skip"
  - "Zero-expectation-edit proof: gate-package zero-diff versus the phase-3-final commit plus single-occurrence oracle constants"

requirements-completed: [IMPORT-PARITY-01]

# Coverage metadata (#1602) — one entry per shipped evidence deliverable.
coverage:
  - id: D1
    description: "The tested tree is attributable: FINAL_CODE_SHA resolves from the exact subject `test(03-02): check session close error`, is an ancestor of HEAD, and the non-.planning diff to HEAD is empty; tracked tree clean and no untracked path outside .planning"
    requirement: IMPORT-PARITY-01
    verification:
      - kind: integration
        ref: "git log --grep='^test(03-02): check session close error$' -1 -> 6f52840c5a90a7b88633d6296154a73a8dcd44ed; git diff --name-only 6f52840 HEAD -- ':(exclude).planning' -> empty; git status --porcelain --untracked-files=no -> empty"
        status: pass
    human_judgment: false
  - id: D2
    description: "Build, vet, tidy idempotence and exact module pins are clean on the attributable tree"
    requirement: IMPORT-PARITY-01
    verification:
      - kind: integration
        ref: "go build ./... rc 0; go vet ./... rc 0 with 0-byte stderr; go mod tidy -diff -> empty; easyprotoc v0.0.0-20261104120806-f378f3f3a05f; protobuf v1.37.0-rc.1; replace=0, vendor=0"
        status: pass
    human_judgment: false
  - id: D3
    description: "One full -race battery is green with the expected shape: 34 ok, 0 FAIL, 0 races, exactly the known environment skip, 3 no-test packages"
    requirement: IMPORT-PARITY-01
    verification:
      - kind: integration
        ref: "go test -race -count=1 -timeout 20m -v ./... rc 0 (log /tmp/04-01-full-battery.log: 34 ok / 0 FAIL / 0 'WARNING: DATA RACE' / 1 SKIP / 3 '?')"
        status: pass
    human_judgment: false
  - id: D4
    description: "Acceptance Gates A-D pass twice per binary against the ratified 1e9a596 baseline, with baseline presence asserted, never skipped, constant unmoved"
    requirement: IMPORT-PARITY-01
    verification:
      - kind: integration
        ref: "internal/api/acceptance_gates_test.go#TestAcceptanceGates (exact RUN; PASS 8.94s; 1e9a596^{commit} present; constant single occurrence)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Golden registry gate passes byte-exactly with the pinned 94 findings and no re-capture"
    requirement: IMPORT-PARITY-01
    verification:
      - kind: unit
        ref: "internal/rules/golden_full_registry_test.go#TestLintGoldenFullRegistry (exact RUN; PASS 0.02s; goldenFindingCount = 94 exactly one occurrence)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Both D-07 position gates pass unchanged (protoast rune columns and core legacy contract)"
    requirement: IMPORT-PARITY-01
    verification:
      - kind: unit
        ref: "TestPositionRuneColumns (19 PASS lines); TestProtoastPositionsMatchLegacyContract (11 PASS lines)"
        status: pass
    human_judgment: false
  - id: D7
    description: "Core and CLI parity matrices pass by exact test name with zero test failures"
    requirement: IMPORT-PARITY-01
    verification:
      - kind: integration
        ref: "TestLintMatrix + TestLintErrorPathFidelity + TestLintSuppressionParity (43 PASS lines); TestLintStdoutGolden + TestLintLinkMatrix + TestLintLinkMatrixExitCodes + TestPrintIssuesCompileGolden (16 PASS lines)"
        status: pass
    human_judgment: false
  - id: D8
    description: "Lint, schema-drift, tidy, pin, replace/vendor and seal predicates are green without touching config, schema or documentation"
    requirement: IMPORT-PARITY-01
    verification:
      - kind: integration
        ref: "bin/golangci-lint run ./... rc 0, exactly `0 issues.`; go run ./cmd/easyp schema-gen rc 0, git diff -- schemas/ 0 bytes; go.mod/go.sum/schemas/.golangci.yml zero diff versus 6f52840; live seal grep 0 matches"
        status: pass
    human_judgment: false
  - id: D9
    description: "Zero expectations moved and the v1.0 delta is recorded honestly with both counting methods and the deletion-by-design explanation"
    requirement: IMPORT-PARITY-01
    verification:
      - kind: integration
        ref: "git diff --name-only 6f52840 HEAD -- ':(exclude).planning' -> empty; gate-package zero-diff; census 37/34/3; battery RUN 3603 / PASS 3602 / SKIP 1 versus v1.0 41 packages / 3718 gotestsum events"
        status: pass
    human_judgment: false

# Metrics
duration: 9min
completed: 2026-11-21
status: complete
---

# Phase 4 Plan 1: Acceptance — full parity evidence on the modular dependency Summary

**The v1.0 acceptance set passes unchanged on the public `easyprotoc` module: one 34-package `-race` battery green (0 FAIL, 0 races, one known environment skip), all eleven named parity gates proven by exact RUN/PASS (Gates A–D ×2 against `1e9a596`, golden 94 byte-exact, both D-07 position gates, the 43-line core and 16-line CLI matrices), exactly `0 issues.` lint, zero schema drift, exact pins, an empty live-surface seal grep, and a zero non-`.planning` diff from `6f52840`.**

## Performance

- **Duration:** 9 min (start 2026-11-21T15:44:14Z; full battery wall 220 s, log written at 15:47:54Z)
- **Tasks:** 3 (all evidence-only gate runs; no authored code, test, config, schema or doc change)
- **Files modified:** 1 (this SUMMARY); planning metadata `STATE.md`, `ROADMAP.md`, `REQUIREMENTS.md` updated by the executor
- **Battery:** `internal/api` 206.889 s dominates; `internal/adapters/plugin` 56.090 s; `internal/generation` 38.072 s; `internal/modules` 25.190 s; every other package under 14 s (sum of per-package times 382.62 s)

## Commit under test

| Item | Value |
|---|---|
| Phase-3-final subject | `test(03-02): check session close error` |
| `FINAL_CODE_SHA` | `6f52840c5a90a7b88633d6296154a73a8dcd44ed` |
| HEAD during evidence (`HEAD_SHA`) | `740bce983c84aebe6a924524ec3795ab6f08ac9c` |
| HEAD descendants since final | 6 `.planning`-only docs commits (`a4defab`, `feb452e`, `c3c098b`, `eb493c5`, `638d0ff`, `779a530`, `375e683`, `740bce9`) — proven harmless by the empty non-.planning diff, not by counting |
| Attribution | `git diff --name-only 6f52840 HEAD -- ':(exclude).planning'` empty, before and after all gates |
| Tracked tree | clean before and after every run; no untracked path outside `.planning/` |
| Branch | `feat/lint-easypc-infra` kept as-is; no push (D-08) |
| Rollback cost | none — this plan changes no tracked file except this evidence SUMMARY |

Pre-existing untracked artifacts under `.planning/` (research cache, `state.json`, `ui-reviews/.gitignore`) were present before the plan started and were intentionally left uncommitted, as the plan instructs.

## Gate Evidence

Every command ran on the committed tree from a clean tracked tree; exit codes are as observed. All logs are under `/tmp/04-01-*`.

| # | Gate | Command | Expected | Observed | Exit |
|---|---|---|---|---|---|
| 1 | Phase-3-final resolvable | `git log --format=%H --grep='^test(03-02): check session close error$' -1` | non-empty | `6f52840c5a90a7b88633d6296154a73a8dcd44ed` | 0 |
| 2 | Final is ancestor of HEAD | `git merge-base --is-ancestor 6f52840 HEAD` | true | true | 0 |
| 3 | Attribution | `git diff --name-only 6f52840 HEAD -- ':(exclude).planning'` | empty | empty | 0 |
| 4 | Tracked tree before | `git status --porcelain --untracked-files=no` | empty | empty | 0 |
| 5 | No untracked outside `.planning` | `git ls-files --others --exclude-standard \| grep -v '^\.planning/'` | empty | empty | 0 |
| 6 | Build | `go build ./...` | rc 0 | rc 0 | 0 |
| 7 | Vet | `go vet ./...` | rc 0, empty stderr | rc 0, 0 bytes stderr | 0 |
| 8 | Tidy idempotence | `go mod tidy -diff` | empty | empty (0 bytes) | 0 |
| 9 | easyprotoc pin | `go list -m -f '{{.Version}}' github.com/easyp-tech/easyprotoc` | `v0.0.0-20261104120806-f378f3f3a05f` | exact | 0 |
| 10 | protobuf pin | `go list -m -f '{{.Version}}' google.golang.org/protobuf` | `v1.37.0-rc.1` | exact | 0 |
| 11 | Package census | `go list ./...` / test / no-test filters | 37 total / 34 test / 3 no-test | 37 / 34 / 3 | 0 |
| 12 | Full battery | `go test -race -count=1 -timeout 20m -v ./...` (log `/tmp/04-01-full-battery.log`) | rc 0, 34 ok, 0 FAIL, 0 `--- FAIL`, 0 race, 0 unexpected skips, 3 no-test | all exact (shape table below) | 0 |
| 13 | Acceptance Gates A-D x2 | `go test -race -count=1 ./internal/api -run '^TestAcceptanceGates$' -v` (log `/tmp/04-01-acceptance.log`) | exact RUN + PASS | `=== RUN   TestAcceptanceGates`; `--- PASS: TestAcceptanceGates (8.94s)` | 0 |
| 14 | Baseline commit present | `git cat-file -e 1e9a596^{commit}` | rc 0 | rc 0 | 0 |
| 15 | Acceptance constant | `grep -c 'acceptanceBaselineCommit = "1e9a596"' internal/api/acceptance_gates_test.go` | 1 | 1 | 0 |
| 16 | Golden 94 | `go test -race -count=1 ./internal/rules -run '^TestLintGoldenFullRegistry$' -v` (log `/tmp/04-01-golden.log`) | exact RUN + PASS | `=== RUN   TestLintGoldenFullRegistry`; `--- PASS` (0.02s) | 0 |
| 17 | Golden constant | `grep -c 'goldenFindingCount = 94' internal/rules/golden_full_registry_test.go` | 1 | 1 | 0 |
| 18 | D-07a positions | `go test -race -count=1 ./internal/protoast -run '^TestPositionRuneColumns$' -v` (log `/tmp/04-01-pos-protoast.log`) | exact RUN + PASS | `--- PASS: TestPositionRuneColumns`; 19 PASS lines (18 subtests) | 0 |
| 19 | D-07b positions | `go test -race -count=1 ./internal/core -run '^TestProtoastPositionsMatchLegacyContract$' -v` (log `/tmp/04-01-pos-core.log`) | exact RUN + PASS | `--- PASS`; 11 PASS lines (10 subtests) | 0 |
| 20 | Core matrix | `go test -race -count=1 ./internal/core -run 'TestLintMatrix\|TestLintErrorPathFidelity\|TestLintSuppressionParity' -v` (log `/tmp/04-01-core-matrix.log`) | 3 RUN + 3 PASS, 0 FAIL | 3 RUN + 3 top-level PASS, 43 PASS lines, 0 `--- FAIL` | 0 |
| 21 | CLI matrix | `go test -race -count=1 ./internal/api -run 'TestLintStdoutGolden\|TestLintLinkMatrix$\|TestLintLinkMatrixExitCodes\|TestPrintIssuesCompileGolden' -v` (log `/tmp/04-01-api-matrix.log`) | 4 RUN + 4 PASS, 0 FAIL | 4 RUN + 4 top-level PASS, 16 PASS lines, 0 `--- FAIL` | 0 |
| 22 | Gate-file freeze | `git diff --name-only 6f52840 HEAD -- internal/api internal/rules internal/core internal/protoast` | empty | empty | 0 |
| 23 | Lint gate (`task lint:go` equivalent) | `bin/golangci-lint run ./...` (log `/tmp/04-01-lint.log`) | rc 0, exactly `0 issues.` | rc 0, one non-empty line: `0 issues.` (10-byte log) | 0 |
| 24 | Schema generation | `go run ./cmd/easyp schema-gen` (log `/tmp/04-01-schema-gen.log`) | rc 0 | rc 0, empty output | 0 |
| 25 | Schema drift | `git diff -- schemas/` + `git diff --exit-code -- schemas/` | empty / rc 0 | 0 bytes / rc 0 | 0 |
| 26 | Module-surface freeze | `git diff --name-only 6f52840 HEAD -- go.mod go.sum schemas .golangci.yml` | empty | empty | 0 |
| 27 | Tidy re-proof | `go mod tidy -diff` | empty | empty | 0 |
| 28 | Escape hatches | `grep -c '^replace' go.mod` / `grep -c '^vendor' go.mod` | 0 / 0 | 0 / 0 | 0 |
| 29 | Seal — live surfaces | `grep -rInE '\beasypc\b\|compiler-lab\|VENDORED' .golangci.yml AGENTS.md .spec/*.md README.md` (log `/tmp/04-01-seal-live.txt`) | empty | 0 matches | 1 (no match) |
| 30 | Seal — recursive `.spec` | `grep -rInE '\beasypc\b\|compiler-lab\|VENDORED' .spec --exclude-dir=plans --exclude-dir=features` (log `/tmp/04-01-seal-spec.txt`) | empty | 0 matches | 1 (no match) |
| 31 | Zero source change | `git diff --name-only 6f52840 HEAD -- ':(exclude).planning'` | empty | empty | 0 |
| 32 | Tracked tree after | `git status --porcelain --untracked-files=no` | empty | empty | 0 |
| 33 | No untracked outside `.planning` | `git ls-files --others --exclude-standard \| grep -v '^\.planning/'` | empty | empty | 0 |

`git diff --quiet 6f52840 HEAD -- .golangci.yml` is clean; the blob hash is `0e70ffefd709b41b7d334350d7efca3c1be121e8` on both commits.

### Battery shape (`/tmp/04-01-full-battery.log`, 13,891 lines, 927,170 bytes, sha256 `6400160e…`)

| Metric | Expected | Observed |
|---|---|---|
| `ok` package lines | 34 | **34** |
| `FAIL` lines | 0 | **0** |
| `--- FAIL` lines | 0 | **0** |
| `WARNING: DATA RACE` blocks | 0 | **0** |
| `--- SKIP` lines | exactly `TestGenerateGoHeaderUsesCompilerVersion` | **exactly that one** (`--- SKIP: TestGenerateGoHeaderUsesCompilerVersion (0.00s)`; protoc-gen-go absent from PATH) |
| `?` no-test packages | 3 | **3** |
| Total packages (`go list ./...`) | 37 | **37** |

No-test packages: `cmd/easyp-mcp`, `internal/logger`, `wellknownimports`.

Derived counts from the same one log: `=== RUN` **3603**, top-level `--- PASS` **1015**, all-level `--- PASS` **3602**, `--- SKIP` **1** (3603 = 3602 PASS + 1 SKIP). The sum of per-package `ok` durations is 382.62 s; the wall time is 220 s because packages run in parallel.

### Gate-count parity versus prior phases

| Gate | Phase 1 (01-02) | Phase 2 (02-02) | Phase 4 (this plan) |
|---|---|---|---|
| `TestAcceptanceGates` | PASS, baseline `1e9a596` | PASS 9.42 s, same baseline | PASS 8.94 s, same baseline |
| `TestLintGoldenFullRegistry` | 94, byte-exact | 94, byte-exact, 0.02 s | 94, byte-exact, 0.02 s |
| `TestPositionRuneColumns` | 19 PASS | 19 PASS | 19 PASS |
| `TestProtoastPositionsMatchLegacyContract` | 11 PASS | 11 PASS | 11 PASS |
| Core matrix | 43 PASS | 43 PASS | 43 PASS |
| CLI matrix | 16 PASS | 16 PASS | 16 PASS |

Timing differences are machine noise; the pass-line counts are identical, so no oracle moved.

## Expectation freeze

- `internal/api`, `internal/rules`, `internal/core` and `internal/protoast` have zero diff versus `6f52840`.
- `acceptanceBaselineCommit = "1e9a596"` and `goldenFindingCount = 94` each occur exactly once.
- Baseline commit `1e9a596` exists, so Gates A–D could not have been skipped.
- No golden, parity fixture, acceptance expectation, rule, engine, API, test, config or schema file was edited, regenerated or re-captured. The plan's only authored artifact is this SUMMARY.

## Seal status

| Surface | Matches | Verdict |
|---|---|---|
| `.golangci.yml` | 0 | clean (no `easypc` path exclusion) |
| `AGENTS.md` | 0 | clean (Phase-3 rewrite: compiler consumed as pinned module, never copied) |
| `README.md` | 0 | clean |
| `.spec/*.md` (top level) | 0 | clean |
| `.spec` recursive minus `plans`/`features` | 0 | clean |
| Archived `.planning`, `.spec/plans`, `.spec/features` | excluded | historical records, out of scope by Phase-3 decision |
| `.claude/CLAUDE.md` | 1 (line 16) | **accepted exclusion** — live GSD-generated instruction mirror deliberately left untouched by Phase-3 D-03; its `compiler-lab/bench/lab.py` mention is recorded here and is **not** claimed as clean |

## v1.0 delta (honest record)

| Metric | v1.0 closeout | v1.1 observed | Reading |
|---|---|---|---|
| Test packages | 41 | **34** | 7 test packages (`pkg/compiler`, `pkg/lexer`, `pkg/linker`, `pkg/options`, `pkg/parser`, `pkg/serialize`, `protoc`) were deleted with the vendored tree by design — not a regression |
| No-test packages | 9 | **3** | 6 no-test packages (`pkg/metrics`, `pkg/sourceinfo`, `protoc/syntax`, `protoc/wkt`, both cmd mains) deleted with the same tree |
| Total packages | 44 | **37** | 41 − 7 = 34 test, 9 − 6 = 3 no-test |
| Test-count figure | 3,718 gotestsum events | **3,602 `--- PASS` + 1 `--- SKIP` = 3,603 verbose results** (1,015 top-level tests) | The methods differ: gotestsum emits events per test/subtest plus package events; the observed figure is derived from `-v` RUN/PASS/SKIP lines. The raw difference is **not** a comparison of equal instruments |

The test-count difference is not evidence of lost coverage: behavioral parity is carried by the byte-exact golden (94 findings), Gates A–D ×2 against the frozen `1e9a596` baseline, and both D-07 position gates — never by the census. The ROADMAP's Phase-4 success criterion 1 still reads "41+ packages / ≥3718 tests"; that phrasing predates the deletion and is superseded by the phase's own acceptance note (CONTEXT D-06) and this plan's must-have of exactly 34 test packages.

## Files Created/Modified

- `.planning/phases/04-priemka-polnyy-paritet/04-01-SUMMARY.md` — this evidence record. No source, test, expectation, golden, config, schema or documentation file was modified.

## Decisions Made

- **No per-task commit.** All three tasks are evidence-only gate runs; `files_modified` names only this SUMMARY, matching the 01-02/02-02 precedent. The last production commit remains `6f52840`, and the single metadata docs commit carries this file.
- **Battery executed once.** The single invocation used the ratified flags from a clean tree. No second `./...` run was performed to obtain a greener result.
- **Auto-mode tracer gate reconciled with the one-battery budget.** Auto mode (`_auto_chain_active: true`) would normally re-run the tracer's `<verify>` end-to-end. That verify corpus includes the full battery, so re-running it would violate the plan's prohibition and T-04-28. The gate was satisfied by the Task-1 run that had just executed and passed; the reconciliation is recorded rather than silently skipped.
- **Exact-name gate claims.** Because `go test -run` exits 0 when its filter matches nothing, every gate is claimed by its top-level `=== RUN` and `--- PASS` lines, not by exit status alone; `TestLintLinkMatrix$` is anchored so the larger `TestLintLinkMatrixExitCodes` cannot satisfy it.
- **Ledger base corrected.** `.git/gsd-plan-head-before-04-01` held `21f6b96` (the plan-creation commit). It was overwritten with the true execution base `740bce9` before any evidence work, following the 02-02 precedent, so `commits` measures this plan only.
- **Skip allow-list is one exact name.** `TestGenerateGoHeaderUsesCompilerVersion` is the only `--- SKIP`; the cause (protoc-gen-go absent from PATH) was confirmed and no other skip was tolerated.
- **`.claude/CLAUDE.md` is an exclusion, not a clean surface.** Its known `compiler-lab/bench/lab.py` mention is recorded as an accepted Phase-3 D-03 exclusion.

## Deviations from Plan

None - plan executed exactly as written. No deviation rule (1–4) was triggered: attribution, build, vet, tidy, the one battery, all eleven named gates, the freeze, lint, schema, seal and census behaved exactly as predicted. The two process notes (tracer/battery reconciliation and the stale ledger base) are recorded under Decisions Made; neither changed a file, a gate, or an expectation.

## Issues Encountered

None. Every gate was green on its first invocation. The battery shape matched the Phase-2 measurement exactly (34 ok / 0 FAIL / 0 race / 1 known skip / 3 no-test), and `internal/api` remained the dominant package at 206.889 s (Phase-2: 221.603 s).

## Known Stubs

None. This plan introduced no placeholder values, TODOs, unwired components or skipped tests. The one `--- SKIP` line is a pre-existing environment guard with an explicit skip reason, not a stub created here.

## Threat Flags

None. No new network endpoint, auth path, file-access pattern or trust-boundary schema change was introduced. This plan is evidence-only and exercised the planned mitigations: T-04-22 (exact RUN-line claims), T-04-23 (freeze plus single-occurrence constants), T-04-24 (attribution and clean tree), T-04-25 (exact `0 issues.` plus config freeze), T-04-26 (schema-gen plus diff gate), T-04-27 (exact pins, tidy, zero replace/vendor), T-04-28 (one fixed-flag battery), T-04-29 (SHAs, logs, counts) and T-04-30 (dual-method delta record).

## Deferred Items (honest record)

- **Push/PR and milestone close:** out of scope by D-08; the work stays on `feat/lint-easypc-infra`, and the publication decision is a milestone-close UAT gate with the user.
- **v1.0 corpora letters:** `LINT-PARITY-01` and `LINT-PERF-01` remain `pending_user_unblock` (private compiler-lab corpora inaccessible); the ready-to-fire scripts and `07-USER-SETUP.md` path stay alive in the milestone backlog.
- **easyprotoc bump:** deferred to v2; v1.1 is pinned to `v0.0.0-20261104120806-f378f3f3a05f`.
- **`.claude/CLAUDE.md` compiler-lab mention:** accepted Phase-3 exclusion, to be resolved the next time the generated mirror is refreshed.
- **Untracked `.planning` runtime artifacts** (research cache, `state.json`, `ui-reviews/.gitignore`) were pre-existing and intentionally not committed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- **Phase 4 acceptance is complete and attributable.** IMPORT-PARITY-01 is verified end-to-end on the phase-3-final code tree with zero expectation movement.
- **Milestone v1.1 is ready for close:** behavior preservation is proven by the golden, Gates A–D ×2 and both D-07 gates; the shape predicates (lint, schema, tidy, pins, seal) are green.
- **Publication remains a user decision:** no push or PR was attempted; the deferred items above are named rather than silently dropped.

---

*Phase: 04-priemka-polnyy-paritet*
*Completed: 2026-11-21*

## Self-Check: PASSED

- SUMMARY.md exists at the plan path (`.planning/phases/04-priemka-polnyy-paritet/04-01-SUMMARY.md`)
- `FINAL_CODE_SHA` 6f52840c5a90a7b88633d6296154a73a8dcd44ed is an ancestor of HEAD, with a zero non-.`.planning` diff to HEAD
- `HEAD_SHA` 740bce983c84aebe6a924524ec3795ab6f08ac9c resolved and recorded; tracked tree clean
- Baseline commit 1e9a596 present; both oracle constants single-occurrence
- All eleven named gates, the battery, lint, schema and seal logs exist under `/tmp/04-01-*`
