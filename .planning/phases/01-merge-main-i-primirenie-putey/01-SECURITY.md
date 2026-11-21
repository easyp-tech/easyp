---
phase: "01"
slug: "merge-main-i-primirenie-putey"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-11-21"
---

# Phase 01 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| upstream git -> local branch | `origin/main` content enters the branch through the merge; SHA, behind-count, and conflict-set assertions are the barrier against an unseen upstream state | source code (proto compiler breaking path) |
| hand resolution -> committed tree | resolved hunks become permanent history; blob hashes are the only barrier to a silent branch-side drop of #253 | source code |
| working tree -> commit index | broad staging could pull untracked planning/runtime artifacts into a merge commit | repository metadata |
| committed merge -> generated evidence | gates run against the working tree at HEAD; evidence is meaningful only if that tree equals the merge commit | test evidence |
| test code -> pinned oracles | a green run is meaningful only if expectations, golden counts, and the baseline commit were not edited or re-captured | test expectations |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-01-01 | Tampering | internal/core/breaking_profiles.go resolution | high | mitigate | Taken from `origin/main` verbatim; blob equals `6e45a055…`; `git diff origin/main -- <file>` empty; `easyprotoc/protoc` present (1), legacy `protocompile` count 0; build green (01-01 gates) | closed |
| T-01-02 | Repudiation / Tampering | frozen lint contour and expectations | high | mitigate | Zero-diff gate vs pre-merge base over `internal/rules`, `internal/easypc`, `internal/protoast`, `internal/core/lint_*`, `link.go`, `proto_info_read.go`, `dom.go`; 01-02 re-proved parity (41 ok / 0 FAIL, golden 94) | closed |
| T-01-03 | Information disclosure | merge-commit contents | medium | mitigate | Explicit pathspecs only; staged-name set equals the 9 expected files; no `git add -A`; tracked tree clean after gates | closed |
| T-01-04 | Tampering | upstream drift between research and execution | high | mitigate | `origin/main` asserted `5d36266d…` pre-flight; behind count 1; exact 3-file conflict set verified; parent blobs asserted | closed |
| T-01-05 | Tampering / Correctness | shared Position model in internal/core/dom.go | medium | mitigate | Branch `dom.go` blob `52bfbd7…` kept; single `type Position = protoast.Position`; compile + D-07 position gates green | closed |
| T-01-06 | Tampering (supply chain) | go.mod / go.sum pins | high | mitigate | MVS pins adopted as-is (`easyprotoc v0.0.0-20261104120806-f378f3f3a05f`, `protobuf v1.37.0-rc.1`); `go list -m` exact; `go mod tidy -diff` empty; zero replace/vendor; go.sum committed | closed |
| T-01-07 | Repudiation | gate evidence attribution | high | mitigate | Merge ancestry/two-parent order/behind=0 asserted; `git diff MERGE_SHA HEAD -- ':(exclude).planning'` empty; HEAD + MERGE_SHA recorded in SUMMARY | closed |
| T-01-08 | Tampering | expectation/golden/baseline edits or a filtered no-op | high | mitigate | Zero-diff freeze vs `79e353a`; constants `94` and `1e9a596` pinned exactly once; exact `=== RUN` assertions; no edits/re-captures/skips | closed |
| T-01-09 | Denial of service / Correctness | battery timeout or silently dropped package | medium | mitigate | `-timeout 20m`; 41 ok / 0 FAIL / 0 race blocks; acceptance re-run by exact name; no `-short` | closed |
| T-01-SC | Tampering | npm/pip/cargo installs | high | mitigate | No npm/pip/cargo installs; the only dependency arrival is the pinned Go module verified via Go proxy + go.sum; no [ASSUMED]/[SUS] packages | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|

No accepted risks.

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
