---
phase: "01"
slug: "vendoring-kompilyatora-i-raschlenenie-modeli"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-11-19"
---

# Phase 01 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| compiler-lab clone → repo | Content fetched from the (private, easyp-owned) upstream over gh-authenticated HTTPS crosses into the tracked tree; treated as source to review, never executed during vendoring | Vendored Go source + testdata (easyp-owned, pinned SHA `08cef10`) |
| fixture bytes → vendored parser | New fixtures and vendored testdata are parsed by tests only; no CLI runtime path executes easypc/protoast in Phase 1 | easyp-owned fixture files |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-01-01 | Tampering | vendor copy step (plan 01-01 Task 1) | medium | mitigate | Clone HEAD asserted equal to pinned SHA `08cef10909c0722f7042ba17b5c11951f9c6b6f6` before copy; SHA recorded in internal/easypc/VENDORED.md:14; drop landed as one standalone reviewable commit | closed |
| T-01-02 | Tampering | internal/easypc (sealed tree) | medium | mitigate | D-04 seal: AGENTS.md Do-not bullet + VENDORED.md prohibition; lint exclusion scoped to `internal/easypc` only so drift outside it stays visible; fixes go upstream by re-sync | closed |
| T-01-03 | Tampering/DoS | protoast/easypc parsers fed fixture input | low | accept | Phase 1 parses only easyp-owned fixtures in tests; no untrusted input crosses the vendored code in any runtime path. Parser hardening/fuzzing scoped out as FUTURE-HARDENING-01 (requirements v2) | closed |
| T-01-04 | Tampering | Issue output construction (AppendIssue / buildBreakingCheckIssue) | low | mitigate | Single conversion point (fromMetaPos) with identical field sets; diff-of-diffs numeric-values gate proved no position value changed; full `-race` suite covers breaking/generate/mod/migrate outputs | closed |
| T-01-05 | Tampering | vendored helper drift (casing / lint_helpers) | low | mitigate | Casing files gated byte-identical by diff; lint_helpers delta bounded to the 3 catalogued declarations + header (every diff hunk read by acceptance criterion); same pinned SHA asserted before copy | closed |
| T-01-SC | Tampering | dependency supply chain | high | mitigate | Zero registry installs this phase by design (LINT-VEND-02): `go mod tidy` zero-diff gate + `grep -c '^replace' go.mod` == 0 (verified: 0) re-run on every plan; vendored trees' only external imports are already-direct requires (research-verified import scan) | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-01-01 | T-01-03 | Phase 1 has no runtime path executing vendored parser code on untrusted input; hardening/fuzzing deferred to FUTURE-HARDENING-01 | Daniel Podolsky (UAT ratification session) | 2026-11-19 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-11-19 | 6 | 6 | 0 | Claude (gsd-security-auditor flow, ASVS L1, register authored at plan time) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-11-19
