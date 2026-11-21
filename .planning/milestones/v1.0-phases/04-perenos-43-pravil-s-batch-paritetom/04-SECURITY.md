---
phase: "04"
slug: "perenos-43-pravil-s-batch-paritetom"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-11-20"
---

# Phase 04 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| proto corpus → linked rules | 41 ported rules walk protoast trees from the linked engine; fixture corpora drive old-vs-new parity in-process | .proto source bytes, positions, written constants |
| parity harness → both engines | The harness drives the live old engine and the dormant new engine over one fixture tree, comparing sorted multisets (incl. SourceName as fifth term) | findings tuples |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-04-01 | Tampering | parity harness blind to suppression state (plan 04-01) | high | mitigate | WR-01 closed in 04-08: harness threads NewLinked's ignoreOnly into the linked app; 3 suppression cells; standing constraint b cells green | closed |
| T-04-02 | Information Disclosure | casing-helper adoption diverging from main (plan 04-03) | medium | mitigate | main regexps byte-verbatim in all casing ports; grep gates zero hits for helper names; T-04-05/06/07 mitigations | closed |
| T-04-03 | Tampering | keyword-position divergence (lab NAME-token vs main KEYWORD) (plan 04-03) | medium | mitigate | Decl.Pos discipline with Wave-0 probe pinning the 10-kind table; zero NamePos/NumberPos/FilePosition (verifier grep) | closed |
| T-04-04 | Tampering | swift_prefix lab fix leaking in (plan 04-04) | low | mitigate | bug-for-bug pin: ruby_package comparison from both sides; deferred fix ledgered | closed |
| T-04-05 | Tampering | casing helper name leakage (plan 04-03) | low | mitigate | grep gates on helper names across the ten ports | closed |
| T-04-06 | Tampering | NamePos/NumberPos leakage (plan 04-03) | low | mitigate | grep gates; verifier re-proved zero code hits | closed |
| T-04-07 | Tampering | recursive-walk over-report on nested decls (plan 04-03) | low | mitigate | main's shallow shapes directed per plan; Wave-0 nested-decl parity cell backstop | closed |
| T-04-08 | Information Disclosure | swift_prefix wrong-target comparison (plan 04-04) | low | mitigate | bug-for-bug pin with deferred fix documented (T-04-08..T-04-11 mitigations per plan) | closed |
| T-04-09 | Tampering | resolution-based ImportUsed divergence (plan 04-04) | low | mitigate | stateless syntax-walk; Resolved grep-clean; qualified-nested-reference cell pins the divergence | closed |
| T-04-10 | Tampering | cross-file RPCRequestResponseUnique divergence (plan 04-04) | low | mitigate | per-file lo.Contains kept; lab cross-file variant deferred (ledger) | closed |
| T-04-11 | Tampering | empty-enum guard deviation scope (plan 04-04) | low | mitigate | THE ONE allowed deviation (crash fix); documented, pinned by tests | closed |
| T-04-12 | Tampering | pascalToUpperSnake helper divergence (plan 04-05) | medium | mitigate | carried byte-verbatim as rename-mapped pascalToUpperSnakeLinked; frozen old file owns the original | closed |
| T-04-13 | Tampering | ENUM_ZERO_VALUE_SUFFIX tag loss (plan 04-05) | low | mitigate | ENV tag kept verbatim | closed |
| T-04-14 | Tampering | full-filename position divergence (plan 04-05) | low | mitigate | 0:0 literal per main | closed |
| T-04-15 | Tampering | cross-file uniqueness divergence (plan 04-05) | low | mitigate | per-file first-exempt on written names via path-grouped Finish | closed |
| T-04-16 | Tampering | version regexp drift (plan 04-05) | low | mitigate | unanchored regexp carried from main | closed |
| T-04-17 | Tampering | comment predicate drift (plan 04-06) | medium | mitigate | len(Comments)==0 → !HasCommentAbove() mapping grounded in parser source (ParseComments attachment semantics verified) | closed |
| T-04-18 | Tampering | recursive walk over-report (plan 04-06) | low | mitigate | shallow walks; groups live in MessageBody.Groups both models (research A4 closed) | closed |
| T-04-19 | Tampering | proto2 group classification (plan 04-06) | low | mitigate | group pin cell | closed |
| T-04-20 | Tampering | registry completeness/order drift (plan 04-07) | medium | mitigate | TestRegistryParity (names per use order, five groups' expansion order, defaultIfEmpty, ignoreOnly unwrap) + TestModuleRuleSet exact-set; full-registry closure cell — 41 rules fire on broad corpus, 73 findings pinned | closed |
| T-04-21 | Tampering | gap-closure regression (plan 04-08) | medium | mitigate | 6 gaps closed with RED-before/GREEN-after cells; suite 47→53 cells; verifier re-proved each closure with own runs | closed |
| T-04-SC | Tampering | npm/pip/cargo/Go-module installs | high | mitigate | Zero-install phase; every plan gates on `go mod tidy && git diff --exit-code go.mod go.sum`; dormancy byte-clean over 44 old-path files | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-04-01 | T-04-02/03 | Corpus lint-diff (api-common, googleapis-nocompute) stays the Phase-7 FINISH gate; fixture-level parity is proven, corpus-level reconciliation deferred by design (ROADMAP criterion 1) | Daniel Podolsky (chain directive; goal-backward verification) | 2026-11-20 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-11-20 | 23 | 23 | 0 | Claude (gsd-security-auditor flow, ASVS L1, registers authored at plan time incl. gap closure) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-11-20
