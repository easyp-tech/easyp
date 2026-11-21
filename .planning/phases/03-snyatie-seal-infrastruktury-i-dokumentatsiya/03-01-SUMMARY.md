---
phase: 03-snyatie-seal-infrastruktury-i-dokumentatsiya
plan: 01
subsystem: documentation-and-lint-config
tags: [easyprotoc, module-dependency, golangci-lint, staticcheck, agents-md, spec-docs, citations, protoast, d-07]

# Dependency graph
requires:
  - phase: 02-pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii
    provides: "All compiler consumers import github.com/easyp-tech/easyprotoc packages and the 648-file internal/easypc vendor tree is deleted, so only the live seal text and three stale citations remained"
provides:
  - "Seal-free live lint config: the internal/easypc paths exclusion and its compiler-lab/VENDORED comment removed; the finding-scoped ST1005 exemption kept byte-identical"
  - "AGENTS.md module-consumption rule (pinned github.com/easyp-tech/easyprotoc, fixes by version bump, never copying module code) and an internal/protoast Layout row"
  - ".spec documentation of the pinned module as the compiler source and internal/protoast as the local D-07 rune-column seam"
  - "Resolvable citations closing Phase-2 IN-01 (module-qualified driver.go range), IN-02 (three divergence surfaces plus both D-07 gates) and IN-03 (easyprotoc linker)"
  - "Recorded intermediate lint baseline: exactly the pre-existing errcheck at internal/sourceview/session_test.go:29:21, isolated for plan 03-02"
affects: [03-02-errcheck-and-phase-gates, 04-priemka-full-parity]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
# Scale: chars/4 over the realized authored diff (15 added + 10 removed lines, 2914 chars).
actuals:
  tokens: 728
  tasks: 2
  commits: 1

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Docs-only atomic commit: one subject, exactly seven declared paths, comment-only Go edits"
    - "Seal predicate lifted to exact literals: zero easypc/compiler-lab/VENDORED matches over live config and docs"
    - "Lint baseline asserted as unchanged rather than zero, keeping the source fix isolated to its own commit"

key-files:
  created: []
  modified:
    - .golangci.yml
    - AGENTS.md
    - .spec/ARCHITECTURE.md
    - .spec/PACKAGES.md
    - internal/core/link_test.go
    - internal/protoast/build.go
    - internal/rules/package_no_import_cycle_test.go

key-decisions:
  - "All config, agent, documentation and citation edits landed as one seven-file docs commit per D-10 commit (a); the errcheck source fix stays isolated for plan 03-02"
  - "The commit's lint gate asserts the pre-existing single finding is unchanged (1 issues, session_test.go:29:21) instead of requiring zero, because the zero transition belongs to plan 03-02"
  - "Documentation wording states the module contract directly (pin in go.mod, fixes upstream, version bump only) and keeps internal/protoast as the documented D-07 seam rather than implying it migrates into the module"

patterns-established:
  - "Live-surface seal grep: `grep -rInE '\\beasypc\\b|compiler-lab|VENDORED'` over config, agent and spec docs must return zero lines"
  - "Changeset equality gate: commit name-set compared to the declared seven-path list, with comment-only delta proof for Go files"

requirements-completed: [IMPORT-SEAL-01]

coverage:
  - id: D1
    description: "The live lint config carries no internal/easypc path exclusion or seal comment, while the finding-scoped ST1005 exemption and exclusions block remain intact"
    requirement: IMPORT-SEAL-01
    verification:
      - kind: other
        ref: "grep -rInE '\\beasypc\\b|compiler-lab|VENDORED' .golangci.yml AGENTS.md (0 lines; measured baseline 3); awk paths-key count 0; grep exclusions/ST1005 rule lines; bin/golangci-lint run ./... -> 1 issues"
        status: pass
    human_judgment: false
  - id: D2
    description: "AGENTS.md replaces the sealed-vendor prohibition with the pinned-module rule and adds the internal/protoast Layout row"
    requirement: IMPORT-SEAL-01
    verification:
      - kind: other
        ref: "grep 'Copy or vendor the compiler module' / 'github.com/easyp-tech/easyprotoc' / 'internal/protoast' AGENTS.md"
        status: pass
    human_judgment: false
  - id: D3
    description: ".spec/ARCHITECTURE.md and .spec/PACKAGES.md describe the pinned module and the internal/protoast D-07 seam, naming only files that exist on disk"
    requirement: IMPORT-SEAL-01
    verification:
      - kind: other
        ref: "grep module path and internal/protoast in both docs; grep -c '<code>lint.go</code>' == 0; existence loop over protoast and corrected core file lists"
        status: pass
    human_judgment: false
  - id: D4
    description: "The three Phase-2 citations are repointed: module-qualified driver.go range at the pin, both D-07 gate names plus three divergence surfaces, and easyprotoc linker"
    verification:
      - kind: other
        ref: "grep the module-qualified citation and both gate names; grep -c 'vendored LineCol' == 0; grep -c 'easypc linker' == 0; comment-only diff gate over the three Go files"
        status: pass
    human_judgment: false

duration: 3 min
started: 2026-11-21T14:28:46Z
completed: 2026-11-21T14:31:19Z
plan_head_before: e49fdebb62475bcefe32545093ce604e6eb90f6f
plan_head_after: 93203614f44da92eaac47ac76e2002e0710f8753
status: complete
---

# Phase 3 Plan 01: Remove the Seal Infrastructure and Document the Pinned easyprotoc Module Summary

**The internal/easypc lint exclusion, the AGENTS.md sealed-vendor prohibition and three stale compiler citations were removed in one seven-file docs commit, while `.spec` now documents the pinned `github.com/easyp-tech/easyprotoc` module and the `internal/protoast` D-07 rune-column seam.**

## Performance

- **Duration:** 3 min
- **Started:** 2026-11-21T14:28:46Z
- **Completed:** 2026-11-21T14:31:19Z
- **Tasks:** 2 (1 tracer edit+commit, 1 evidence-only)
- **Files modified:** 7 (plus this SUMMARY)

## Accomplishments

- `.golangci.yml`: deleted the `paths: [internal/easypc]` exclusion and its compiler-lab/`VENDORED.md` comment; `exclusions:` and the finding-scoped `ST1005` staticcheck rule stayed byte-identical.
- `AGENTS.md`: the final `## Do not` bullet now forbids copying or vendoring the compiler module and states that `github.com/easyp-tech/easyprotoc` is consumed from the `go.mod` pin and updated by a version bump; the Layout table gained an `internal/protoast` row naming the local seam and divergence D-07.
- `.spec/ARCHITECTURE.md`: the `internal/core` row now names the three compiler packages as coming from the pinned module, and a new paragraph records the bump path and the `internal/protoast` seam.
- `.spec/PACKAGES.md`: added the `internal/protoast` row and replaced the deleted `lint.go` token with the three live `lint_linked.go` / `lint_linked_read.go` / `lint_model.go` files.
- Citations: `internal/core/link_test.go` cites `github.com/easyp-tech/easyprotoc/internal/compiler/driver.go:319-323` at the pin; `internal/protoast/build.go` names the three divergence surfaces and both D-07 gates and calls `LineCol` the pinned compiler's; `internal/rules/package_no_import_cycle_test.go` says `easyprotoc linker`.
- Gates on the committed slice: seal grep over `.golangci.yml`, `AGENTS.md`, `.spec/*.md`, `README.md` returns 0 (baseline 3); `go build ./...` green; `go test -race -count=1 ./internal/core ./internal/protoast ./internal/rules` green; `gofmt -l` empty; `go mod tidy -diff` empty; lint still exactly one issue at `internal/sourceview/session_test.go:29:21`.

## Task Commits

Each task was committed atomically:

1. **Task 1: Remove the seal, document the module, repoint the citations, and land the seven-file docs commit** - `9320361` (docs)
2. **Task 2: Prove the committed documentation slice and its gate baseline** - no commit (evidence-only; no tracked file changed)

**Plan metadata:** committed with this SUMMARY (`docs(03-01): complete remove seal infrastructure and document easyprotoc dependency plan`).

## Files Created/Modified

- `.golangci.yml` - Removed the `internal/easypc` path exclusion and seal comment; retained the `ST1005` finding-scoped rule
- `AGENTS.md` - Module-consumption rule in `## Do not`; `internal/protoast` Layout row
- `.spec/ARCHITECTURE.md` - Pinned-module compiler dependency in the `internal/core` row plus the seam/bump paragraph
- `.spec/PACKAGES.md` - Live `internal/core` lint file list; new `internal/protoast` row
- `internal/core/link_test.go` - Module-qualified `driver.go:319-323` citation (comment only)
- `internal/protoast/build.go` - Re-sync note naming three surfaces and both D-07 gates; `pinned compiler's LineCol` (comment only)
- `internal/rules/package_no_import_cycle_test.go` - `easyprotoc linker` naming (comment only)
- `.planning/phases/03-snyatie-seal-infrastruktury-i-dokumentatsiya/03-01-SUMMARY.md` - This summary

## Decisions Made

- All config/docs/citation edits form one atomic seven-file docs commit (D-10 commit (a)); no per-file commits.
- The lint gate for this commit pins the unchanged baseline (exactly the known errcheck) rather than zero, isolating the transition to plan 03-02.
- The documentation states the module contract and the preserved `internal/protoast` seam explicitly, so a future version bump does not read as a migration of protoast into the module.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The executor's plan commit ledger (`.git/gsd-plan-head-before-03-01`) contained a stale base 219 commits behind HEAD from a prior run. It was reset to the actual pre-plan HEAD `e49fdebb62475bcefe32545093ce604e6eb90f6f`, so the SUMMARY's measured `commits` value (1) reflects this plan's real commit range.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 03-02 remains: fix the close-error errcheck at `internal/sourceview/session_test.go:29` and prove the phase-final gates (lint 0 issues, seal grep, tidy, targeted tests).
- The intermediate lint baseline is recorded here so the final 1 -> 0 transition has a single, attributable cause.
- No blockers; the tracked tree was clean after commit `9320361` and no push occurred.

## Self-Check: PASSED

All seven modified paths exist on disk and commit `9320361` is an ancestor of HEAD.
