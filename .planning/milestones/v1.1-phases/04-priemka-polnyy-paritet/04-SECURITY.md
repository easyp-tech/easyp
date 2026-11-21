---
phase: "04"
slug: "priemka-polnyy-paritet"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-11-21"
---

# Phase 04 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| committed tree -> gate evidence | a green gate is meaningful only if the tested tree equals the accepted code tree | source/test evidence |
| test code -> pinned oracles | parity claims hold only if goldens/constants/baseline were not edited | test expectations |
| dependency pins -> build inputs | acceptance must run on the pinned module, not a substituted cache | module graph |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-04-22 | Repudiation | empty or non-matching `go test -run` filter passing vacuously | high | mitigate | Every gate claimed by exact `=== RUN` + `--- PASS`; `$` anchors separate `TestLintLinkMatrix` from `…ExitCodes` | closed |
| T-04-23 | Tampering | golden/acceptance/parity expectation moved to mask a regression | high | mitigate | Constants single-occurrence; zero diff in `internal/{api,rules,core,protoast}` vs phase-3-final; red gate stops, never edited | closed |
| T-04-24 | Tampering | testing an uncommitted or drifted tree | high | mitigate | Zero non-`.planning` diff `FINAL_CODE_SHA..HEAD` before/after runs; clean tracked tree; no untracked paths outside `.planning` | closed |
| T-04-25 | Tampering | lint gate weakened by suppression/exclusion/disabled linter | high | mitigate | Exact single-line `0 issues.`; `.golangci.yml` zero-diff vs phase-3-final | closed |
| T-04-26 | Tampering | schema drift masked by skipping/reverting | high | mitigate | `schema-gen` run; `git diff --exit-code -- schemas/`; clean tree; drift would stop execution | closed |
| T-04-27 | Tampering (supply chain) | pin drift or module-cache substitution | high | mitigate | Exact `go list -m` pins; go.mod/go.sum zero-diff; `go mod tidy -diff` empty; zero replace/vendor | closed |
| T-04-28 | Denial of service | battery rerun or flag relaxation hiding a real failure | medium | mitigate | Exactly one `-race -count=1 -timeout 20m -v ./...` invocation; counts from that one log | closed |
| T-04-29 | Repudiation | unattributed or fabricated evidence counts | medium | mitigate | `FINAL_CODE_SHA`, `HEAD_SHA`, commands, exit codes, RUN/PASS counts, log paths recorded; counts derived from files | closed |
| T-04-30 | Repudiation | v1.0 test-count delta misrepresented as parity/regression | medium | mitigate | 41→34 packages and observed PASS/SKIP counts recorded with both methods; parity confined to golden/Gates A–D/D-07 | closed |
| T-04-SC | Tampering | npm/pip/cargo installs | high | mitigate | No package-manager installs; no dependency change; only pinned Go modules covered by go.sum | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-04-01 | — (Phase-3 inheritance) | `.claude/CLAUDE.md` is a live generated mirror with a known `compiler-lab/bench/lab.py` mention; excluded from the seal predicate by Phase-3 D-03 and recorded as an accepted exclusion, never as a clean surface | orchestrator (Phase-3 D-03) | 2026-11-21 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-11-21 | 10 | 10 | 0 | orchestrator (ASVS L1 short-circuit; register authored at plan time) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-11-21
