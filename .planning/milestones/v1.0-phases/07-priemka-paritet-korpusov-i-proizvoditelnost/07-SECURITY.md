---
phase: "07"
slug: "priemka-paritet-korpusov-i-proizvoditelnost"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-11-21"
---

# Phase 07 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| wazero runtime construction | The pre-existing lazy-global race fixed via construction serialization — production and test seams now provably race-free under `-race` | in-process |
| acceptance evidence | parity-evidence-07.json / perf-evidence-07.md carry the milestone's central claims; provenance blocks pin both binaries (baseline 1e9a596/41, head 9a37392/44) | committed evidence |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-07-01 | Tampering | runtimeFactory seam freeze (plan 07-01) | medium | mitigate | 3 test seam overrides freeze-guarded (grep gates); builtin_impl.go zero-diff in phase 7 | closed |
| T-07-02 | DoS | wazero v1.9.0 lazy-global race GetWazeroVersion (plan 07-01) | high | mitigate | FIXED: one-point serialization (package runtimeCreationMu + newRuntimeSerialized wrapping the default factory; both creation sites through the shared field) + TestMain priming; mutation check: Lock/Unlock removed → pin RED, restore byte-identical; first full green -race ./... (3718 tests) | closed |
| T-07-03 | Tampering | probe snippet empirically broken both ways (plan 07-04) | medium | mitigate | Replaced with deterministic held-mutex ordering probe + retained peak==1; pin extended to BOTH production creation sites; 07-04 review verified probe sound (no false-RED, no early synctest.Wait) | closed |
| T-07-04 | Tampering | corpus-config "every rule" misstatement (plan 07-04) | low | mitigate | Corrected to measured STANDARD=35/44 HEAD (32/41 baseline); COMMENTS+UNARY_RPC never enabled; Gate D unaffected | closed |
| T-07-05 | Repudiation | parity-evidence-07.json provenance | medium | mitigate | Baseline block pins both binaries (1e9a596/41, 9a37392/44 — measured chain: zero Go-source delta to first milestone commit, two docs-only commits intervene); verifier independently re-ran TestAcceptanceGates + selftest | closed |
| T-07-06 | Repudiation | corpora legs anti-ratification (no results claimed without corpora) | medium | mitigate | Fresh access attempts recorded at execution (gh exit 1, SSH exit 128, env unset — 2026-11-21T04:15:09Z) into clone_attempts; status pending_user_unblock; MANUAL-ONLY rows; automated disposition gate | closed |
| T-07-SC | Tampering | npm/pip/cargo installs | low | accept | Zero-install phase; tidy diff-empty gates | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-07-01 | T-07-02 residual | wazero upgraded at v1.9.0 stays; serialization is easyp-side; upstream upgrade needs verification + ratification (REQ-5.2 surface) — deferred past milestone | Daniel Podolsky (UAT ratified in absentia per chain directive) | 2026-11-21 |
| AR-07-02 | T-07-06 | Corpora legs pending_user_unblock — obligation alive in milestone backlog with one-command path; automated core (teeth Gates A-D ×2, golden 94, SC3) carried the machine-verifiable substance | Daniel Podolsky (UAT ratified in absentia per chain directive) | 2026-11-21 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-11-21 | 7 | 7 | 0 | Claude (gsd-security-auditor flow, ASVS L1, registers authored at plan time incl. gap closure) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-11-21
