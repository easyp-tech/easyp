---
phase: 01
review: 01-REVIEW.md
titles: json
findings:
  - id: WR-01
    severity: warning
    disposition: open
    title: "Vacuous invalid-case assertions — the whole `package_same_*` test family can never fail on a rule regression"
  - id: WR-02
    severity: warning
    disposition: open
    title: "`shouldIgnore` prefix match has no path-separator boundary — config ignores silently over-match"
  - id: WR-03
    severity: warning
    disposition: open
    title: "D-07b parity proof does not assert `Offset` (or `Filename`) although it claims full `Position` equality"
  - id: IN-01
    severity: info
    disposition: open
    title: "Phantom parameter — `AppendIssue` accepts `comments` it never reads"
  - id: IN-02
    severity: info
    disposition: open
    title: "`core.CheckIsIgnored` is exported production code with no production caller"
  - id: IN-03
    severity: info
    disposition: open
    title: "Two divergent `toUpperSnakeCase` implementations in the same module"
  - id: IN-04
    severity: info
    disposition: open
    title: "`decl()` computes `span.End - 1` without the empty/synthetic guard its caller `register()` has"
  - id: IN-05
    severity: info
    disposition: open
    title: "`directory_same_package_test.go` valid branch dereferences a nil `wantIssues` on regression"
  - id: IN-06
    severity: info
    disposition: open
    title: "Unparenthesized `&&`/`||` precedence in the enum-value comparison reads as two different rules"
  - id: IN-07
    severity: info
    disposition: open
    title: "Legacy `BreakingChecker.Check()` emits issues in map-iteration order — nondeterministic CLI output"
open: 10
total: 10
recorded: 2026-11-21T05:55:51.308Z
---

# Phase 01: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| WR-01 | warning | open | - |
| WR-02 | warning | open | - |
| WR-03 | warning | open | - |
| IN-01 | info | open | - |
| IN-02 | info | open | - |
| IN-03 | info | open | - |
| IN-04 | info | open | - |
| IN-05 | info | open | - |
| IN-06 | info | open | - |
| IN-07 | info | open | - |

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved.
