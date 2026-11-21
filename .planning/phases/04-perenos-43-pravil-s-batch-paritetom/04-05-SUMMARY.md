---
phase: 04-perenos-43-pravil-s-batch-paritetom
plan: 05
subsystem: rules
tags: [lint-rules, protoast, parity-harness, linked-engine, porting, dormancy, default-batch, prefix-algorithm]

# Dependency graph
requires:
  - phase: 04-perenos-43-pravil-s-batch-paritetom (plans 01-04)
    provides: requirePorted/assertBatchParity/fingerprint harness, parseLinkedFile + validate() drivers, NewLinked registry (24/41), AppendLinkedIssue + LintRuleName seam, locked Decl.Pos position table, firstExemptMismatches, fileOption helper
provides:
  - Eight DEFAULT ports (7 LintRules + 1 ModuleRule) with main's algorithms byte-faithful: pascalToUpperSnake prefix (underscore before EVERY uppercase rune, bare HasPrefix), first-value exact-equality zero-value suffix with the ENV:"ENUM_ZERO_VALUE_SUFFIX" tag kept, full-filename regexp at the 0:0 zero literal, per-file written-name uniqueness as a ModuleRule Finish, exact written-name compares, unanchored version regexp, keyword-position service suffix (32/41 cumulative)
  - TestLinkedParityDefault — eight cells pinning the batch's main-vs-lab divergences under cumulative 32-rule configs (29 green parity sub-cells cumulative)
  - Eight carried test rewrites with byte-verbatim main literals (zero old-contract tests left for the batch)
  - builder_linked.go DEFAULT slots filled in builder.go order — registry covers MINIMAL+BASIC+DEFAULT (32/41) plus pilot
affects: [04-06..04-07 batches (harness + registry reuse), 04-07 (exact-set guard, deferred-decisions consolidation), Phase 5 flip (rename-map consolidation incl. pascalToUpperSnakeLinked), Phase 7 corpus gate]

# Actuals (#2632) — same scale as the plan's estimate (chars/4 over realized diff)
actuals:
  tokens: 13037    # 52148 diff chars over the 19 changed files at chars/4 (estimate was 82000; five uniform LintRule ports on a proven harness, probe-validated fixtures, no harness changes)
  tasks: 3
  commits: 3       # MEASURED: git rev-list --count 33284989..HEAD (RED test + 2 GREEN); excludes the metadata docs commit

# Tech tracking
tech-stack:
  added: []
  patterns: [carried-helper pattern (main's private algorithm re-declared byte-verbatim under a rename-mapped Linked name to dodge the frozen-file collision), ModuleRule per-file grouping Finish (insertion-ordered path groups + per-file first-exempt seen set on written names), regexp-declared-inside-Validate (no package-level collision with frozen old files), probe-driven fixture correction before RED (distinct symbols per same-package file + live-oracle verdicts)]

key-files:
  created:
    - internal/rules/parity_default_test.go
    - internal/rules/enum_value_prefix_linked.go
    - internal/rules/enum_zero_value_suffix_linked.go
    - internal/rules/file_lower_snake_case_linked.go
    - internal/rules/rpc_request_response_unique_linked.go
    - internal/rules/rpc_request_standard_name_linked.go
    - internal/rules/rpc_response_standard_name_linked.go
    - internal/rules/package_version_suffix_linked.go
    - internal/rules/service_suffix_linked.go
  modified:
    - internal/rules/builder_linked.go
    - internal/rules/enum_value_prefix_test.go
    - internal/rules/enum_zero_value_suffix_test.go
    - internal/rules/file_lower_snake_case_test.go
    - internal/rules/package_version_suffix_test.go
    - internal/rules/service_suffix_test.go
    - internal/rules/rpc_request_response_unique_test.go
    - internal/rules/rpc_request_standard_name_test.go
    - internal/rules/rpc_response_standard_name_test.go

key-decisions:
  - "pascalToUpperSnake is carried byte-verbatim as pascalToUpperSnakeLinked (single declaration in enum_value_prefix_linked.go, called from enum_zero_value_suffix_linked.go): the frozen old file enum_value_prefix.go still declares the original in the same package, so a verbatim redeclaration was a compile error — the body is byte-identical, the name is rename-mapped (pascalToUpperSnake -> pascalToUpperSnakeLinked) for the Phase-5 consolidation, and the lab's acronym-aware helpers stay grep-clean"
  - "The v1alpha.proto verdict follows the live old engine (the oracle), not the plan's 'clean' annotation: a digit directly after the leading lowercase run fails main's full-filename regexp, and the file IS the main-vs-lab tooth — the lab strips .proto and compares the already-lower-snake stem, so it would pass the file main fails (probe-confirmed before RED)"
  - "Fixture corrections from the pre-RED probes enforced the compiling precondition: same-package files needed distinct enum names (protoc duplicate symbols would route the linked engine to COMPILE), and the second rpc file's package ru2.v1 failed PACKAGE_LOWER_SNAKE_CASE (digit after the leading lowercase run) — renamed rw.v1; old-side counts 7/2/3/10/2/3/1/0 probe-confirmed per cell"
  - "An eighth parity cell, central_filter_suppressible, mechanizes standing constraint (b) inside the batch suite: a buf:lint:ignore SERVICE_SUFFIX directive silences the batch finding (0) while the unsuppressed fixture reports it (1) — the ported rules never consult comments themselves"
  - "FileLowerSnakeCaseLinked/PackageVersionSuffixLinked declare main's regexps inside Validate (batch-2 precedent): the frozen old files own the package-level matchLowerSnakeCase/matchVersionSuffix declarations, so package-level redeclaration was impossible; behavior is identical"

patterns-established:
  - "Per-file-uniqueness ModuleRule Finish: insertion-ordered path groups over accumulated occurrences, per-file first-exempt seen set on written names, findings wrapped in IssueInfo{Path: occ.path}, state reset — main's lo.Contains-per-file semantics on the Collect/Finish lifecycle"
  - "Carried-helper pattern: a frozen old file's private helper is re-declared byte-verbatim under a Linked-suffix name in the ported file (rename-map entry) so the port owns its algorithm without touching the old path"

requirements-completed: [LINT-PARITY-02]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Batch-4 parity suite TestLinkedParityDefault: eight cells proving old-vs-new sorted (rule, path, line, col) multiset parity for the eight DEFAULT rules under cumulative 32-rule configs, with probe-confirmed expected old-side counts (7/2/3/10/2/3/1/0/0-1) asserted before each compare"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_default_test.go#TestLinkedParityDefault"
        status: pass
      - kind: other
        ref: "command: gsd_run check tdd-red-evidence .planning/phases/04-perenos-43-pravil-s-batch-paritetom/tdd-red-evidence-04-05-task1.json -> RED_EVIDENCE_OK (target test failed on the requirePorted guard, exit 1)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Prefix algorithm pinned (T-04-12): enum HTTPCode values HTTPCODE_OK/OTHER both issue under main's H_T_T_P_CODE prefix (the lab's acronym-aware helper would pass HTTPCODE_OK), the name-only value FOB passes the bare HasPrefix (the lab's trailing-underscore requirement would issue), the two-level walk reaches top-level-message enums and stops at nested-message enums"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_default_test.go#TestLinkedParityDefault/prefix_algorithm"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/enum_value_prefix_test.go#TestEnumValuePrefix_Validate"
        status: pass
    human_judgment: false
  - id: D3
    description: "First-value exact-equality suffix pinned (T-04-15): FOO_UNSPECIFIED clean, FOB_OTHER issues at the first value's position while the second value's wrong suffix is invisible, BARX_UNSPECIFIED issues despite passing a HasSuffix reading (exact equality only); the ENV:\"ENUM_ZERO_VALUE_SUFFIX\" struct tag carried byte-verbatim and the suffix defaults to UNSPECIFIED via NewLinked's defaultIfEmpty"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_default_test.go#TestLinkedParityDefault/zero_value_suffix"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/enum_zero_value_suffix_test.go#TestEnumZeroValueSuffix_Validate"
        status: pass
      - kind: other
        ref: "command: grep -n 'ENV:' internal/rules/enum_zero_value_suffix_linked.go -> tag present"
        status: pass
    human_judgment: false
  - id: D4
    description: "Full-filename check pinned: InvalidName.proto/aB.proto/v1alpha.proto issue at the zero Position literal (path, 0, 0) against the FULL name including .proto, good_name.proto/good_name2.proto stay clean (digits allowed after a separator) — no extension stripping, never the lab's 1:1 file position"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_default_test.go#TestLinkedParityDefault/file_name_full"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/file_lower_snake_case_test.go#TestFileLowerSnakeCase_Validate"
        status: pass
    human_judgment: false
  - id: D5
    description: "Per-file uniqueness pinned (T-04-13): the second use of ReqT and the input==output OneX duplicate issue within one file at their rpc keyword positions with the written type as source (first user exempt), while the same written name ReqT in a different file stays clean; the ModuleRule Finish groups by file and never reads the resolution table (grep-clean)"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_default_test.go#TestLinkedParityDefault/rpc_unique_per_file"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/rpc_request_response_unique_test.go#TestRPCRequestResponseUnique_Validate"
        status: pass
      - kind: other
        ref: "command: grep -n 'Resolved\\|RPCTypes' internal/rules/rpc_request_response_unique_linked.go (code lines) -> no hits"
        status: pass
    human_judgment: false
  - id: D6
    description: "Written-name compares pinned (T-04-14): the qualified .pkg.v1.GetFooRequest and GetFooReply fail the exact compares (both standard-name rules issue at the rpc keyword position) while the unqualified exact form and the service-qualified SecondServiceGetBarResponse stay clean; no pascal normalization or dot-segment splitting exists in the ports (grep-clean)"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_default_test.go#TestLinkedParityDefault/standard_name_written"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/rpc_request_standard_name_test.go#TestRPCRequestStandardName_Validate"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/rpc_response_standard_name_test.go#TestRPCResponseStandardName_Validate"
        status: pass
    human_judgment: false
  - id: D7
    description: "Version and service suffixes pinned: packages example/example.foo issue under main's unanchored regexp while the single-component v1 passes (the lab's two-component helper would fail it; PACKAGE_LOWER_SNAKE_CASE co-fires on v1 — digit in the first component), and service User issues at the keyword position column 1 while UserService stays clean with the default Service suffix"
    requirement: LINT-PARITY-02
    verification:
      - kind: integration
        ref: "tests/internal/rules/parity_default_test.go#TestLinkedParityDefault/version_suffix"
        status: pass
      - kind: integration
        ref: "tests/internal/rules/parity_default_test.go#TestLinkedParityDefault/service_suffix"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/package_version_suffix_test.go#TestPackageVersionSuffix_Validate"
        status: pass
      - kind: unit
        ref: "tests/internal/rules/service_suffix_test.go#TestServiceSuffix_Validate"
        status: pass
    human_judgment: false
  - id: D8
    description: "Batch gate green on ten gates (ROADMAP Phase-4 criterion 1 — verified before 04-06 starts): full core+rules -race suite; cumulative TestLinkedParity* (29 green sub-cells, 32/41 rules); suppression centrality + readSource-once + TestModuleRuleSet; go build; tidy diff-empty; schema drift-empty; dormancy git-diff empty over the frozen old-path files; golangci-lint 0 issues; grep-clean on forbidden helpers/resolution in all eight ports"
    requirement: LINT-PARITY-02
    verification:
      - kind: other
        ref: "command: go test -race -count=1 ./internal/core/... ./internal/rules/... && go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity' && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ && bin/golangci-lint run ./internal/rules/... ./internal/core/..."
        status: pass
    human_judgment: false

# Metrics
duration: 13 min
started: 2026-11-19T20:36:43Z
completed: 2026-11-19T20:50:05Z
status: complete
commits: 3
plan_head_before: 33284989a9f2bc7e63724c1ea7909bf53605f42a
plan_head_after: 6789cfb4f5effb2936c129658d83be608c02d9f4
---

# Phase 04 Plan 05: Batch 4 — DEFAULT Summary

**Eight DEFAULT rules ported with main's algorithms byte-faithful (pascalToUpperSnake prefix, first-value exact-equality suffix with the ENV tag kept, full-filename 0:0, per-file written-name uniqueness as a ModuleRule Finish, exact written-name compares, unanchored version regexp), parity-proven old-vs-new through eight cumulative cells (32/41 rules ported)**

## Performance

- **Duration:** 13 min
- **Started:** 2026-11-19T20:36:43Z
- **Completed:** 2026-11-19T20:50:05Z
- **Tasks:** 3
- **Files modified:** 18 source/test files (+1 planning artifact: RED evidence JSON; +VALIDATION.md task rows)

## Accomplishments

- The eight DEFAULT rules are ported as `<Name>Linked` types with main's semantics: ENUM_VALUE_PREFIX carries the private pascalToUpperSnake algorithm byte-verbatim (underscore before EVERY uppercase rune — "HTTPCode" becomes "H_T_T_P_CODE", so acronym enum values issue where the lab's acronym-aware helper would pass them) with a bare strings.HasPrefix (name-only values pass; the lab's trailing-underscore requirement is not ported), ENUM_ZERO_VALUE_SUFFIX checks ONLY the first value for EXACT equality and keeps main's `ENV:"ENUM_ZERO_VALUE_SUFFIX"` struct tag byte-verbatim, FILE_LOWER_SNAKE_CASE matches the FULL filename including .proto at the 0:0 zero literal, RPC_REQUEST_RESPONSE_UNIQUE is per-file first-exempt uniqueness on written type strings, the standard-name rules compare written strings exactly (qualified types fail), PACKAGE_VERSION_SUFFIX uses the unanchored regexp (single-component "v1" passes), SERVICE_SUFFIX is HasSuffix at the keyword position
- RPCRequestResponseUniqueLinked keeps the etalon's ModuleRule Collect/Finish lifecycle but groups accumulated occurrences by file in Finish, reproducing main's within-file lo.Contains first-exempt semantics (including input==output duplicates); the resolution table is never read (grep-clean)
- `TestLinkedParityDefault` proves old-vs-new parity in eight cells under cumulative 32-rule configs; every corpus was probe-validated before the RED run (live old-engine count oracle + a zero-rules linked-engine link check proving no COMPILE routing)
- All eight carried test files rewritten onto the linked contract with byte-verbatim literals (917/51/3 + 667/33/5 twice, 20/3/1, 197/10/1, InvalidName 0:0, 375/14/3, 214/11/3 twice); the ModuleRule carried test drives Collect/Finish through the validate() driver
- builder_linked.go DEFAULT slots filled in builder.go order with defaultIfEmpty config plumbing (UNSPECIFIED / Service); registry covers MINIMAL+BASIC+DEFAULT — 32/41 rules; batch gate green on ten gates before 04-06 starts

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): DEFAULT parity suite fails on the unported-rule guard** - `16f46c2` (test)
2. **Task 2 (GREEN part 1): port #25/#26/#27/#31/#32 — prefix algorithm, suffix, filename, version, service suffix** - `7ab23e0` (feat)
3. **Task 3 (GREEN part 2): port #28 ModuleRule + #29/#30 standard names + builder fill + batch gate** - `6789cfb` (feat)

**Plan metadata:** docs commits (this SUMMARY + state/roadmap/validation updates).

## TDD Gate Compliance

- RED `16f46c2` (test(04-05)) precedes both GREEN commits `7ab23e0`/`6789cfb` (feat(04-05)) in ancestry (merge-base verified: RED is an ancestor of the first GREEN); no REFACTOR needed — the GREEN commits are the minimal implementations and no cleanup beyond the plan shape was warranted.
- RED evidence machine-validated (`gsd_run check tdd-red-evidence` → `RED_EVIDENCE_OK`, reason `target_test_failed`): TestLinkedParityDefault executed and failed on its own requirePorted assertion naming all eight unported rules while the package compiled; the semantic assessment is recorded in the evidence file.
- Task 1's RED commit touches only the `*_test.go` pathspec plus the evidence JSON planning artifact; Tasks 2-3 carry the GREEN implementations with their behavior proof run before each commit.
- Note: the plan splits the batch across two GREEN tasks with builder registration deferred to Task 3 (the checker-mandated rebalance), so between `7ab23e0` and `6789cfb` the parity suite still failed on the requirePorted guard (it probes all eight names and reports the full unported list — the 04-02 ratified decision). The five Task-2 rules were probe-verified old-vs-new through their carried tests at Task 2 time and re-proven under the full 32-name config at the Task 3 gate.

## Files Created/Modified

- `internal/rules/{enum_value_prefix,enum_zero_value_suffix,file_lower_snake_case,package_version_suffix,service_suffix}_linked.go` - NEW: the five LintRule ports (carried prefix algorithm, first-value exact equality + ENV tag, full-filename 0:0, unanchored version regexp, HasSuffix at keyword pos)
- `internal/rules/{rpc_request_response_unique,rpc_request_standard_name,rpc_response_standard_name}_linked.go` - NEW: the ModuleRule port (per-file first-exempt grouping Finish) and the two exact written-compare LintRules
- `internal/rules/builder_linked.go` - eight DEFAULT placeholders replaced in builder.go order; rename map updated (32/41 filled; COMMENTS/UNARY_RPC placeholders remain for 04-06)
- `internal/rules/parity_default_test.go` - NEW: eight parity cells + defaultUse config (32 names cumulative) + probe-evidence header
- `internal/rules/{<five LintRules>,<three RPC rules>}_test.go` - rewritten onto the linked contract (parseLinkedFile + validate() Collect/Finish for the ModuleRule), main's case sets and literals carried verbatim

## Decisions Made

- pascalToUpperSnake carried as `pascalToUpperSnakeLinked`: the frozen old file declares the same name in this package, so a verbatim redeclaration could not compile; the body is byte-identical, the declaration lives once in enum_value_prefix_linked.go (called from enum_zero_value_suffix_linked.go per the plan), and the rename is recorded for Phase 5
- The v1alpha.proto cell verdict follows the live old engine (the oracle), not the plan's "clean" annotation — the file is documented in the cell comment as the main-vs-lab tooth (the lab strips the extension; main fails the digit after the leading lowercase run of the full name)
- Fixture corrections came out of the pre-RED probes: same-package files got distinct enum names (protoc duplicate symbols would route the linked engine to COMPILE — the research precondition), and the second rpc file's package was renamed ru2.v1 → rw.v1 (it failed PACKAGE_LOWER_SNAKE_CASE)
- An eighth cell `central_filter_suppressible` mechanizes standing constraint (b) inside the batch suite: the batch finding is silenced ONLY by the central `buf:lint:ignore` filter, never by the rules themselves
- The two regexp ports declare their patterns inside Validate (batch-2 precedent) because the frozen old files own the package-level declarations of the same names — behavior identical, no collision

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] pascalToUpperSnake redeclaration impossible — carried under a rename-mapped name**
- **Found during:** Task 2 (porting the prefix rule)
- **Issue:** the plan says to carry main's private pascalToUpperSnake byte-verbatim into enum_value_prefix_linked.go, but the byte-frozen old file enum_value_prefix.go still declares that exact name in the same package — a verbatim redeclaration is a compile error, and deleting/renaming it in the old file would violate the dormancy invariant
- **Fix:** the helper is declared once in the ported file as `pascalToUpperSnakeLinked` with a byte-identical body (renames: rules call it from both ports; the name is rename-mapped `pascalToUpperSnake -> pascalToUpperSnakeLinked` for the Phase-5 consolidation and documented in the file header)
- **Files modified:** internal/rules/enum_value_prefix_linked.go, internal/rules/enum_zero_value_suffix_linked.go
- **Verification:** prefix_algorithm and zero_value_suffix parity cells green (behavior identical to main by construction); `grep -c "func pascalToUpperSnakeLinked"` = 1
- **Committed in:** 7ab23e0

**2. [Rule 1 - Bug] v1alpha.proto verdict corrected against the live oracle**
- **Found during:** Task 1 (fixture-design probe: old-engine count run)
- **Issue:** the plan's cell spec annotates `v1alpha.proto` as "clean — dots/digits allowed by the regexp", but main's regexp `^[a-z]+([_|[.][a-z0-9]+)*$` fails a digit directly after the leading lowercase run (no separator), so the FULL filename "v1alpha.proto" ISSUES — the old engine (the oracle standard, same as 04-04's literal correction) reports it at 0:0
- **Fix:** the cell keeps the file and asserts the oracle verdict (count includes it); the cell comment documents it as exactly the main-vs-lab tooth (the lab strips .proto and compares the already-lower-snake stem, so it would pass the file main fails)
- **Files modified:** internal/rules/parity_default_test.go (fileNameFiles comment + count)
- **Verification:** file_name_full cell green old-vs-new with 3 old-side findings (InvalidName, aB, v1alpha — all at 0:0)
- **Committed in:** 16f46c2

**3. [Rule 3 - Blocking] Parity fixtures did not compile as first drafted — probe-driven corrections before RED**
- **Found during:** Task 1 (fixture-design probes: zero-rules linked-engine link check + old-engine count run)
- **Issue:** three same-package files reused the enum name Foo (and its value names), which protoc rejects as duplicate symbols — the linked engine would route the corpus to COMPILE and parity would be vacuous; separately, the second rpc file's package ru2.v1 failed the already-ported PACKAGE_LOWER_SNAKE_CASE (a digit directly after the leading lowercase run), adding an unplanned 11th old-side finding to the rpc cell
- **Fix:** distinct enum/value names per same-package file (Fob/FOB, Foa/FOA_BAR, Foe/FOE_* in the prefix and suffix cells) and the rpc cell's second package renamed to rw.v1; both probes re-run clean (counts 7/2/3/10/2/3/1/0 + link check empty) before the RED run
- **Files modified:** internal/rules/parity_default_test.go (fixture maps + comments)
- **Verification:** TestZZProbe old-count and link probes green on all eight fixture sets (probe file deleted before commit, 04-04 precedent)
- **Committed in:** 16f46c2

---

**Total deviations:** 3 auto-fixed (1 blocking helper rename, 1 oracle-verdict correction, 1 blocking fixture redesign)
**Impact on plan:** All three necessary for parity fidelity and the compiling-corpus precondition; no scope change, no behavior drift, old path byte-frozen throughout.

## Issues Encountered

- Task 2's batch-parity state between the two GREEN commits is structurally RED (requirePorted probes all eight names; builder registration is deliberately rebalanced into Task 3 per the plan's checker note) — handled exactly as in 04-04: the five Task-2 rules were verified through their carried tests plus the suppression/readSource cells at Task 2 time, and the full parity suite was proven at the Task 3 gate.
- The plan's Task-2 action anticipated this exact mid-plan RED ("expected; the carried tests ... are this task's green oracle") — no plan change needed.

## Known Stubs

None — no placeholder data sources, TODOs, or unwired components were introduced.

## Deferred Decisions (for 04-07 consolidation ledger)

- Cross-file RPCRequestResponseUnique (the lab's cross-file/qualified/report-everyone ModuleRule over Resolved.RPCTypes) — main's per-file first-exempt semantics on written names shipped; the lab variant is the deferred post-parity decision, grep-enforced (zero references to the resolution table in rpc_request_response_unique_linked.go)
- Standard-name normalization (the lab's last-dot-segment pascal-normalized compare that accepts qualified types) — main's exact written compares shipped; the lab variant is a deferred decision
- Acronym-aware casing helpers (the lab's upper-snake helper with its required trailing underscore) — main's carried algorithm shipped; helper substitution stays a deferred decision (casing.go untouched, grep-clean)
- pascalToUpperSnakeLinked carries the Linked suffix until Phase 5 deletes the old rule files (rename-map entry in the ported file header); the two locally-declared regexps collapse into package-level declarations at the same consolidation if desired
- EnumZeroValueSuffix/EnumFirstValueZero keep the len(Values)==0 guard as the documented crash-safety deviation (research A7; unit-pinned since 04-04) — carried unchanged into this batch's suffix port

## User Setup Required

None - no external service configuration required. The etalon checkout precondition held: `/tmp/gsd-research/compiler-lab` at SHA `08cef10909c0722f7042ba17b5c11951f9c6b6f6` (verified at execution start).

## Verification Log

- `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityDefault'` → RED: exit 1, `--- FAIL: TestLinkedParityDefault` naming all eight unported rules — Task 1 verify; evidence record validated `RED_EVIDENCE_OK`
- Fixture-design probes (TestZZProbe, deleted before commit): old-engine counts 7/2/3/10/2/3/1/0 probe-confirmed per cell (the rpc cell's 11-finding first draft exposed the ru2.v1 PACKAGE_LOWER_SNAKE_CASE co-fire; the link probe exposed the duplicate-symbol fixture fault); zero-rules linked runs empty on every corpus after the fixes; the first probe run also corrected the v1alpha.proto verdict (issues at 0:0)
- `go test -race -count=1 ./internal/rules/ -run 'TestEnumValuePrefix|TestEnumZeroValueSuffix|TestFileLowerSnakeCase|TestPackageVersionSuffix|TestServiceSuffix'` → ok (all literals verbatim) — Task 2 verify
- `go test -race -count=1 ./internal/core/ -run 'Suppression|ReadSourceOnce'` → ok (standing constraints a+b, after every task)
- Grep-clean: forbidden identifiers (version helper/acronym helpers/name+number tokens/file-level position) in the five LintRule ports → no hits (headers reworded per the 04-04 precedent so the acceptance grep is unambiguous); `Resolved|RPCTypes` and normalization helpers in the three RPC ports → no hits
- Batch gate (Task 3, re-run on the committed state): full core+rules `-race` ok; `TestLinkedParity -v` → 29 green sub-cells / 34 PASS lines incl. 5 parents (3 Wave-0 + 5 MINIMAL + 5 BASIC casing + 8 BASIC semantic + 8 DEFAULT); `go build ./...` ok; `go mod tidy && git diff --exit-code go.mod go.sum` → TIDY_CLEAN; `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` → SCHEMA_CLEAN; dormancy `git diff --exit-code` over the 16 frozen old-path files (8 old DEFAULT rules + builder.go + casing.go + lint_helpers.go + engine files) → DORMANCY_CLEAN; `bin/golangci-lint run ./internal/rules/... ./internal/core/...` → 0 issues; `TestModuleRuleSet` → ok
- TDD gate: `git log -E --grep '^test\((0*4)-(0*5)\):'` → 16f46c2; feat commits 7ab23e0/6789cfb; merge-base ancestry RED→GREEN verified
- Commits measured: `git rev-list --count 33284989..HEAD` = 3 (excludes the metadata docs commit)

## Next Phase Readiness

- Ready for 04-06 (batch 5 — COMMENTS): the DEFAULT group is fully ported and parity-verified before the next batch starts (ROADMAP Phase-4 criterion 1); the registry holds 32/41; the remaining placeholders are exactly batch 5's scope (CommentEnum, CommentEnumValue, CommentField, CommentMessage, CommentOneof, CommentRPC, CommentService) plus batch 6's two UNARY_RPC rules
- The deferred-decisions list above carries this batch's parked items into the 04-07 consolidation ledger (cross-file uniqueness, standard-name normalization, acronym helpers, helper-name consolidation)
- No harness changes were needed; no blockers; standing constraints (a) and (b) remain mechanized green cells (TestReadSourceOnce + the core Suppression cells + the new central_filter_suppressible cell)

## Self-Check: PASSED

- All 11 task-created/modified files exist on disk (9 created + 9 modified per git log --name-status of the three task commits, 18 source/test files total; verified with `[ -f ]` above)
- All 4 commits are ancestors of HEAD: 16f46c2 (RED test), 7ab23e0 (GREEN part 1), 6789cfb (GREEN part 2 + gate), b509c29 (SUMMARY + VALIDATION)
- Commits measured from the persisted ledger: `git rev-list --count 33284989..HEAD` = 3 (code commits; plan_head_before 33284989a9f2bc7e63724c1ea7909bf53605f42a, plan_head_after 6789cfb4f5effb2936c129658d83be608c02d9f4)
- RED evidence record exists and validates RED_EVIDENCE_OK
- Broken-windows ledger: the two documented deviations (helper rename, oracle verdict) appended to .planning/WINDOWS.md as open entries 3/4; no stubs, no skipped tests, no unrun verifies

---
*Phase: 04-perenos-43-pravil-s-batch-paritetom*
*Completed: 2026-11-19*
