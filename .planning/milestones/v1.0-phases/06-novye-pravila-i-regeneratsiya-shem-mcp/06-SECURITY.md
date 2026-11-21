---
phase: "06"
slug: "novye-pravila-i-regeneratsiya-shem-mcp"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-11-21"
---

# Phase 06 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| lab checkout → repo ports | Three rule files cross from the lab clone into the tracked tree; byte-verbatim diff gates prove no hand edits | Go source (3 files) |
| registry → derived surfaces | Rule names flow from the executable registry into schema enums, MCP descriptions, and migration validation — one atomic commit keeps them consented | name lists |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-06-01 | Tampering | internal/rules/{syntax_specified,field_not_required,package_no_import_cycle}.go port fidelity | medium | mitigate | Byte-verbatim diff gates vs lab @ 08cef10 (empty ×3); code review re-proved diff-empty; unsuffixed filenames per AGENTS.md | closed |
| T-06-02 | DoS | PackageNoImportCycle.cycle recursive DFS over the package graph | low | accept | Corpus graphs are bounded; recursion depth = package count; fuzzing deferred (FUTURE-HARDENING-01) | closed |
| T-06-03 | Tampering | schemas/easyp-v1.schema.json, schemas/easyp.schema.json hand-edit drift | medium | mitigate | Regenerated ONLY via `go run ./cmd/easyp schema-gen`; drift gate empty on committed state; code review verified exactly 3 enum spots × 3 names × 2 files | closed |
| T-06-04 | Tampering | internal/rules/selection.go ValidateNames special-case residue | medium | mitigate | 4-line branch deleted in the atomic commit; post-deletion semantics identical to lab's ValidateNames; pins flipped NotContains→Contains | closed |
| T-06-05 | Repudiation | parity-evidence-06-03.json provenance | medium | mitigate | Committed evidence records both binaries' pinned versions (lab @ 08cef10, repo @ d0525b4); verifier independently re-executed the compare (MULTISET_EQUAL=true) | closed |
| T-06-06 | Tampering | /tmp/gsd-research/compiler-lab checkout as porting source + oracle | low | mitigate | SHA re-verified per use; lab checkout 0 dirty files; fallback clone procedure documented | closed |
| T-06-SC | Tampering | npm/pip/cargo installs | low | accept | Zero-install phase; tidy diff-empty gates every plan | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-06-01 | T-06-02 | Recursive DFS bounded by real-world package-graph sizes; hardening deferred to FUTURE-HARDENING-01 | Daniel Podolsky (UAT ratified in absentia per chain directive) | 2026-11-21 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-11-21 | 7 | 7 | 0 | Claude (gsd-security-auditor flow, ASVS L1, registers authored at plan time) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-11-21
