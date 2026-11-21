---
phase: 01
review: 01-REVIEW.md
titles: json
findings:
  - id: WR-01
    severity: warning
    disposition: deferred
    title: "Bare defer session.Close() discards error; upstream pre-existing, repo lint gate red on main too"
  - id: WR-02
    severity: warning
    disposition: deferred
    title: "Close is not terminal-safe (nil-map panic / dir('.') recursion); latent upstream, no live caller"
  - id: WR-03
    severity: warning
    disposition: deferred
    title: "Close unsynchronized with in-flight opens despite 'safe for concurrent use'; latent upstream"
  - id: IN-01
    severity: info
    disposition: deferred
    title: "Second sourceview.Session created in extended breaking path; behavior-neutral cleanup"
  - id: IN-02
    severity: info
    disposition: deferred
    title: "Duplicated precedence-dependent condition in breaking_profiles.go; upstream file"
  - id: IN-03
    severity: info
    disposition: deferred
    title: "Duplicated importCache/parsedImports comment; pre-dates the merge"
  - id: IN-04
    severity: info
    disposition: deferred
    title: "Session.OpenLocal close-semantics doc mismatch; upstream doc fix"
  - id: IN-05
    severity: info
    disposition: deferred
    title: "Byte-vs-rune column contracts under one Position alias; pre-existing, future position work"
  - id: IN-06
    severity: info
    disposition: deferred
    title: "Silent first-wins target dedup in breaking walk; pre-existing, upstream"
  - id: IN-07
    severity: info
    disposition: deferred
    title: "Multi-line diagnostics inside wrapped Compile error; upstream style"
open: 0
total: 10
---

# Phase 01: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| WR-01 | warning | deferred | deferred: upstream pre-existing (`origin/main` #253); owner Phase 3 (lint gate) |
| WR-02 | warning | deferred | deferred: latent upstream (sourceview lifecycle); upstream fix, not a merge action |
| WR-03 | warning | deferred | deferred: latent upstream (Close/open race); upstream fix, not a merge action |
| IN-01 | info | deferred | deferred: behavior-neutral cleanup; future phase |
| IN-02 | info | deferred | deferred: upstream file adopted verbatim; upstream fix |
| IN-03 | info | deferred | deferred: pre-dates the merge; cleanup later |
| IN-04 | info | deferred | deferred: upstream doc fix |
| IN-05 | info | deferred | deferred: pre-existing position-contract note; future work |
| IN-06 | info | deferred | deferred: pre-existing upstream behavior; upstream decision |
| IN-07 | info | deferred | deferred: upstream error-wrapping style |

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved.
