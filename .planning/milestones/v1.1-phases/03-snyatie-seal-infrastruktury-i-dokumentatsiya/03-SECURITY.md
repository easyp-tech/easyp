---
phase: "03"
slug: "snyatie-seal-infrastruktury-i-dokumentatsiya"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-11-21"
---

# Phase 03 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| live config -> lint gate | removing an exclusion must not weaken the gate it configured | lint configuration |
| live docs -> operative rule | AGENTS.md/.spec are the operative agent guidance; stale vendoring text misdirects future work | instruction text |
| history -> provenance | archived v1.0 records must survive the docs pass | historical records |
| test -> oracle | the errcheck fix must not change test semantics | test code |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-03-01 | Tampering / Repudiation | `AGENTS.md`/`.spec` still describe vendoring or mismatch the pin | high | mitigate | Exact-literal seal grep 3→0; module-path grep; go.mod cross-check; `driver.go:319-323` verified in the pinned module cache | closed |
| T-03-02 | Tampering | lint gate weakened while removing the exclusion | high | mitigate | Config diff limited to the four-line block; `exclusions:` + `ST1005` guards; no suppressions; lint 1→0 at source | closed |
| T-03-03 | Tampering | stale citation names a deleted file or wrong line range | medium | mitigate | Module-qualified citation verified against the pinned module; both D-07 gate names asserted | closed |
| T-03-04 | Repudiation | historical records rewritten | medium | mitigate | `.spec/plans/**`/`.spec/features/**` zero-diff; `.planning/milestones` zero-diff in the two-commit range; `.claude/CLAUDE.md` untouched | closed |
| T-03-05 | Tampering | unintended behavior/test change inside a docs commit | medium | mitigate | Comment-only diff gate over the three Go files; gofmt/build/targeted `-race`; seven-file changeset equality | closed |
| T-03-06 | Tampering | zero-issue result achieved by suppression or exclusion | high | mitigate | Commit (b) exactly one test file; `.golangci.yml` zero diff in the range; no suppression directive; output exactly `0 issues.` | closed |
| T-03-07 | Repudiation / Correctness | close-error check changes test semantics or masks a failure | medium | mitigate | `t.Cleanup` + `t.Errorf` exact shape; session still closed once at test end; `-race` run; one-file commit | closed |
| T-03-08 | Repudiation | gates reported from the wrong tree/commit or a partial run | medium | mitigate | Gates run after commits on a clean tree; subjects/paths/sizes asserted; combined range equality; historical-dir zero-diff | closed |
| T-03-SC | Tampering | npm/pip/cargo installs | high | mitigate | No package-manager installs; no dependency change; `go mod tidy -diff` empty; go.mod/go.sum untouched in range | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|

No accepted risks.

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-11-21 | 9 | 9 | 0 | orchestrator (ASVS L1 short-circuit; register authored at plan time) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-11-21
