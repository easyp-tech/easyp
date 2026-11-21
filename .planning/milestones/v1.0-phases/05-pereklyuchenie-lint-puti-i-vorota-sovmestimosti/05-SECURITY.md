---
phase: "05"
slug: "pereklyuchenie-lint-puti-i-vorota-sovmestimosti"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-11-20"
---

# Phase 05 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| proto corpus → flipped engine | `easyp lint` now serves user corpora exclusively through the linked compiler engine; error-path classification (command error vs COMPILE remark) is the security-relevant routing | .proto source bytes, OS open/read failures, import paths |
| breaking path freeze | breaking/generate/tidy/migrate subsystems untouched — freeze gates prove byte-equality whole-phase | none (isolation boundary) |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-05-01 | Tampering | golden drift (golden captured from stale engine state) (plan 05-01) | high | mitigate | Golden captured pre-flip from the frozen old engine with purity guards (zero production diffs); byte-identity re-proven through the renamed engine post-flip | closed |
| T-05-02 | Information Disclosure | error-path masking (open/read failures classified as COMPILE) (plan 05-02) | high | mitigate | WR-01: typed *os.PathError discrimination landed pre-flip; extended in 05-05 to sourceview sentinels (ErrUnsupported/ErrChanged/ErrNestedRepository) — command error carries the OS cause as main did; probe cells RED-proven | closed |
| T-05-03 | Elevation of Privilege | importFileAllowed bypass for corpus files (plan 05-02) | medium | mitigate | WR-02 closed: resolver mirrors openDiskSource policy; end-to-end engine denial cell green | closed |
| T-05-04 | Tampering | partial flip (non-atomic commit leaves hybrid state) (plan 05-03) | high | mitigate | Single atomic commit 4238fbe verified (135 files, exactly 59 deletions); rollback = git revert; compile gate inside the commit | closed |
| T-05-05 | Tampering | dangling references to deleted symbols (plan 05-03) | medium | mitigate | Zero live refs grep gate over all deleted symbols; go build/vet/golangci clean; registry trim matches schema enums exactly | closed |
| T-05-06 | Tampering | breaking-path damage hidden by two-file freeze gate (plan 05-03) | medium | mitigate | Freeze extended whole-phase over four breaking-owned files; verifier re-proved empty diff; trimmed comment_suppressions.go keeps shared primitives | closed |
| T-05-07 | Tampering | rename consolidation breaking external name expectations (plan 05-04) | medium | mitigate | Reflection name-neutrality proven (TestAllRuleNames/TestAllGroups/TestModuleRuleSet); golden byte-identical post-rename; git diff -M reviewability | closed |
| T-05-08 | Information Disclosure | sentinel misclassification residue (plan 05-05) | medium | mitigate | WR-01 gap closure: errors.Is on the three sourceview sentinels; A4 non-local COMPILE classification stays pinned; probe cell RED-proven | closed |
| T-05-SC | Tampering | npm/pip/cargo/Go-module installs | high | mitigate | Zero-install phase; go.mod gates require BOTH legacy modules present (no new deps); tidy diff-empty every plan | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-05-01 | T-05-04 | The flip's rollback is git revert of one commit — no runtime rollback mechanism beyond git; accepted as the roadmap-mandated design (откат = revert) | Daniel Podolsky (chain directive; roadmap SC wording) | 2026-11-20 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-11-20 | 9 | 9 | 0 | Claude (gsd-security-auditor flow, ASVS L1, registers authored at plan time incl. gap closure) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-11-20
