---
phase: "02"
slug: "pereklyuchenie-potrebiteley-i-udalenie-vendor-kopii"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-11-21"
---

# Phase 02 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| vendored source -> public module | build inputs move from an in-tree copy to a pinned external module | source code (proto compiler) |
| working tree -> commit index | the swap and the 648-file deletion must land as one atomic, path-scoped commit | repository metadata |
| test code -> pinned oracles | parity evidence is meaningful only if expectations/goldens were not edited | test expectations |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-02-11 | Tampering / Repudiation | vendored import path left in any tracked Go file | high | mitigate | Scripted sweep with per-mapping assertions; `git grep 'internal/easypc' -- '*.go'` empty; exact 9/8 module-import counts; compile gate | closed |
| T-02-12 | Tampering | deletion of a still-referenced file (build break) | high | mitigate | Baseline: only the seven D-02 files referenced the tree; atomic deletion; `go build`, `go vet`, quick `-race` tests; package count 37/34/3 | closed |
| T-02-13 | Repudiation | expectation/golden edit masking a regression | high | mitigate | Changeset 648 D + 7 M + 0 A; strip-gate byte-identity; golden 94 / Gates A–D / D-07 re-run green | closed |
| T-02-14 | Tampering (supply chain) | go.mod/go.sum drift, replace/vendor escape hatch | high | mitigate | Tidy asserted no-op; `go mod tidy -diff` empty; `go list -m` exact; zero replace/vendor; go.mod/go.sum zero-diff | closed |
| T-02-15 | Tampering | partial split of swap and deletion across commits | medium | mitigate | Single-commit changeset gate (648/7/0); subject lookup; build green on the commit | closed |
| T-02-16 | Information disclosure / Repudiation | unrelated paths staged into the atomic commit | medium | mitigate | Explicit `git add` pathspecs; staged name-set equality; `.planning`/`.golangci.yml`/`AGENTS.md`/`.spec` absent | closed |
| T-02-17 | Tampering / Correctness | API drift between the deleted vendor copy and the pinned module | medium | mitigate | Planning-time source diff at the exact pseudo-version (identical modulo imports/comments/gofmt); build+vet; full parity green | closed |
| T-02-18 | Denial of service | stale `.golangci.yml` exclusion of a deleted path | low | accept | D-07 assigns seal removal to Phase 3; lint not a Phase 2 gate; an exclusion for a nonexistent path cannot affect build/tests | closed (accepted) |
| T-02-19 | Repudiation | empty `go test -run` filter passing vacuously | high | mitigate | Every gate asserted by exact top-level `=== RUN` plus `--- PASS`; exit status alone never accepted | closed |
| T-02-20 | Tampering | golden/acceptance expectation moved to mask a regression | high | mitigate | Modified-set equality; `internal/rules`/`internal/api` zero-diff; pinned constants single-occurrence; no re-capture | closed |
| T-02-21 | Tampering | testing an uncommitted or drifted tree | high | mitigate | Zero non-`.planning` diff from `SWAP_SHA` to HEAD; clean tree before/after; one battery on the committed tree | closed |
| T-02-22 | Tampering (supply chain) | module-cache substitution or pin drift during the run | high | mitigate | Exact `go list -m` versions re-checked; tidy diff empty; zero replace/vendor; go.sum unchanged | closed |
| T-02-23 | Denial of service | battery rerun/relaxation hiding a real failure | medium | mitigate | One invocation with fixed flags; exact package-line counts; failure reported as a swap defect | closed |
| T-02-24 | Repudiation | unattributed evidence | medium | mitigate | SUMMARY records SHAs, commands, exit codes, RUN-line counts, log paths; tree clean at end | closed |
| T-02-SC | Tampering | npm/pip/cargo installs | high | mitigate | No package-manager installs; no dependency change; only the pinned Go module (Package Legitimacy Audit: Approved) | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-02-01 | T-02-18 | The `.golangci.yml` exclusion of the now-deleted `internal/easypc` path remains until Phase 3 owns its removal; it cannot affect build, tests, or the module graph | orchestrator (pre-ratified D-07) | 2026-11-21 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-11-21 | 15 | 15 | 0 | orchestrator (ASVS L1 short-circuit; register authored at plan time) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-11-21
