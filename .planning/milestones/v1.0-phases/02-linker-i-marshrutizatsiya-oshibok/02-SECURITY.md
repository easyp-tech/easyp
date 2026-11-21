---
phase: "02"
slug: "linker-i-marshrutizatsiya-oshibok"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-11-19"
---

# Phase 02 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| proto corpus → vendored linker | Corpus files (user-owned proto trees) are parsed and linked by the vendored `easypc` compiler during lint; malformed/crafted input is the attack surface | .proto source bytes; import paths |
| test subprocess → built binary | The exit-code matrix drives a freshly built `easyp` binary over hostile fixtures (tab/UTF-8, cycles, crafted import paths) in temp dirs | CLI arguments, exit codes, stderr |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-02-01 | Information Disclosure | internal/core/link.go lintResolver (plan 02-01) | high | mitigate | Resolver delegates only to `c.openImportFile` (zero own filesystem access); blocked/transitive missing imports surface as `*OpenImportFileError` → exit 2 (main-identical route) — proven by blocked_import/transitive cells | closed |
| T-02-02 | Tampering (DoS) | protoc.Link on malformed corpora (vendored compiler panic) (plan 02-01) | medium | accept | Malformed fixtures (tab/UTF-8, cycles) driven through the classifier under `-race`; recover boundary is a deliberate Phase 3/5 decision; fuzzing deferred (FUTURE-HARDENING-01) | closed |
| T-02-03 | Elevation of Privilege / Information Disclosure | crafted import paths ("../x.proto", absolute) (plan 02-01) | low | mitigate | `filepath.IsLocal` guard inside openImportFile (existing control); non_local_import_compile cell pins the COMPILE classification; resolver adds no path handling | closed |
| T-02-04 | Information Disclosure | lintResolver (plan 02-02) | high | mitigate | Enforced by construction (delegates only to c.openImportFile); plan lands the proof: blocked_import + transitive_missing_depth2 cells assert fallback → OpenImportFileError, no COMPILE leakage | closed |
| T-02-05 | Tampering (DoS) | protoc.Link malformed corpora (plan 02-02) | medium | accept | tab_utf8 and syntax-error fixtures through the classifier under `-race`; fuzzing deferred (FUTURE-HARDENING-01) | closed |
| T-02-06 | Elevation of Privilege / Information Disclosure | crafted import paths (plan 02-02) | low | mitigate | non_local_import_compile cell pins the etalon COMPILE classification; no resolver-side path handling added | closed |
| T-02-07 | Information Disclosure | lintResolver (plan 02-03) | high | mitigate | Resolver performs zero filesystem access outside c.openImportFile; blocked_import cell asserts blocked → missing → fallback → exit-2 route | closed |
| T-02-08 | Tampering (DoS) | protoc.Link malformed corpora (plan 02-03) | medium | accept | Matrix drives malformed fixtures through the real binary under `-race`; recover boundary deliberate Phase 3/5 decision; fuzzing deferred | closed |
| T-02-09 | Elevation of Privilege / Information Disclosure | crafted import paths (plan 02-03) | low | mitigate | Subprocess matrix exercises the CLI end-to-end; non-local import classification pinned at core level (plan 02-02) | closed |
| T-02-10 | Tampering | diagnostic handler concurrency (plan 02-04, gap closure) | low | mitigate | `diagCollector` mutex-guarded collection; handler_race_mixed cell pins exact counts under 4-goroutine contention; matrix green ×2 under `-race -count=2` | closed |
| T-02-SC | Tampering | npm/pip/cargo/Go-module installs | high | mitigate | Zero-install phase; every plan gates on `go mod tidy && git diff --exit-code go.mod go.sum` — any new dependency fails the plan | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-02-01 | T-02-02/05/08 | Compiler-panic recover boundary is a deliberate Phase 3/5 decision; crashes surface as `-race` test failures now; fuzzing deferred to FUTURE-HARDENING-01 | Daniel Podolsky (UAT ratification session) | 2026-11-19 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-11-19 | 11 | 11 | 0 | Claude (gsd-security-auditor flow, ASVS L1, registers authored at plan time, renumbered per #4683) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-11-19
