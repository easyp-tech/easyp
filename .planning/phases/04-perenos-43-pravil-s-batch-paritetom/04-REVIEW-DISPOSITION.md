# Phase 04 — Code Review Disposition

**Review:** 04-REVIEW.md (deep, 2026-11-20, probe-backed via go test -overlay)
**Result:** 1 critical / 4 warnings / 1 info
**Recorded:** 2026-11-20 (advisory ledger — one row per finding)

| Finding | Severity | Title | Disposition | Notes |
|---------|----------|-------|-------------|-------|
| CR-01 | critical | ignore_only changes group membership in main (ignored files never seed exempt-first value) but linked engine Collects every file, filtering findings only | gap-closure REQUIRED | Probe-proven (old=0/new=1); violates REQ-2.5 parity semantics; affects all nine group ModuleRules; must be fixed before phase close — route: gap-closure plan this phase (engine: skip ignored files in Collect/Finish membership per shouldIgnoreLinked, or path-gate the exempt-first seeding) |
| WR-01 | warning | assertBatchParity discards NewLinked's ignoreOnly return (harness gap that hid CR-01) | fix WITH CR-01 | Harness must wire IgnoreOnly into the linked side + a dedicated ignore_only-group-membership parity cell |
| WR-02 | warning | Duplicate file-level option: main iterates every matching occurrence; ports collapse via fileOption (last-wins) — old=2/new=0 | fix WITH CR-01 | Port-side: iterate all matching occurrences |
| WR-03 | warning | Duplicate import: main reports last written occurrence, port the first (line 6 vs 5) | fix WITH CR-01 | Port-side: match main's occurrence choice |
| WR-04 | warning | Parity fingerprint omits SourceName (user-visible in text+JSON output) — harness gap, all 41 ports currently match | fix WITH CR-01 | Add SourceName to fingerprint compare + regression-guard cell |
| IN-01 | info | registry_parity_guard_test.go fails gofmt -l | fix WITH CR-01 | gofmt -w as part of the fix commit |

**Summary:** 0 open-after-plan / 6 routed to gap-closure / 0 deferred / 0 rejected. CR-01 is a real behavioral divergence on the phase's own parity contract — phase cannot pass with it open.
