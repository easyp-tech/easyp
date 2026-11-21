---
phase: "03"
slug: "dvuhprohodnyy-dvizhok-i-podavleniya"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-11-19"
---

# Phase 03 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| proto corpus → linked engine | Corpus files flow through the dormant `LintLinked` engine (pass-1 readSource once per file; models from linked trees); malformed input is the crash surface | .proto source bytes, import paths |
| rule doubles → engine | Shared test-double rule instances exercise cross-file state isolation under `-race` | in-memory findings |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-03-01 | Tampering | parallel windows shared state (plan 03-01) | medium | mitigate | Statelessness audited inside `race_shared_doubles_identity` (exact per-file counts, no cross-file bleed); serial Collect proven by unsynchronized recording slice under `-race`; engine immutable per-file inputs | closed |
| T-03-02 | DoS | unbounded corpus through parallel engine (plan 03-01) | medium | accept | Windows of 256 bound memory; small-corpus threshold avoids goroutine overhead; recover boundary is a deliberate Phase 3/5 decision; fuzzing deferred (FUTURE-HARDENING-01) | closed |
| T-03-03 | Information Disclosure | suppression bypass via rules self-suppressing (plan 03-02) | low | mitigate | Central filter: checkFile per-finding filter + Finish re-filter under allowCommentIgnores; 21/21 parity rows green; standing constraint ratified in UAT (rules never self-suppress — Phase 4 batches re-check) | closed |
| T-03-04 | Tampering | fallback path divergence (plan 03-01) | low | mitigate | main-wins wrapper text `fs.WalkDir: c.lintProtoInfoRead: %w` pinned by exact-string cells; negative gate on the etalon callee literal | closed |
| T-03-05 | Tampering | config validation drift (plan 03-04) | low | mitigate | Parity guard through untouched `rules.New` chain: 7 cells (errors.Is ErrInvalidRule + exact prefixes); main oracle test green untouched | closed |
| T-03-SC | Tampering | npm/pip/cargo/Go-module installs | high | mitigate | Zero-install phase; every plan gates on `go mod tidy && git diff --exit-code go.mod go.sum`; dormancy seals byte-verified | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-03-01 | T-03-02 | Engine recover boundary and corpus-size hardening deferred to Phase 5 flip; crashes surface as `-race` failures now; fuzzing deferred to FUTURE-HARDENING-01 | Daniel Podolsky (UAT ratification session) | 2026-11-19 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-11-19 | 6 | 6 | 0 | Claude (gsd-security-auditor flow, ASVS L1, registers authored at plan time) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-11-19
