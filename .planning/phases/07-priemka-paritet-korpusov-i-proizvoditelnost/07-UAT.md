---
status: testing
phase: 07-Приёмка — паритет корпусов и производительность
source: [07-VERIFICATION.md]
started: 2026-11-21T06:30:00Z
updated: 2026-11-21T06:30:00Z
---

## Current Test

number: 1
name: Corpora parity legs (LINT-PARITY-01 letter) — unblock or ratify deferral
expected: |
  Complete 07-USER-SETUP.md (GOOGLEAPIS fresh clone per Pitfall 7 + CORPUS_ARGS), fire
  /tmp/p7/run-parity.sh parity unchanged; or ratify the pending_user_unblock deferral (A5).
  Expected on unblock: Gate A 0/0 x2 per corpus, Gate B zero COMPILE, Gate C multiset equal,
  Gate D reported (~0 expected).
awaiting: user response

## Tests

### 1. Corpora parity legs (LINT-PARITY-01 letter) — unblock or ratify deferral
expected: |
  Complete 07-USER-SETUP.md (GOOGLEAPIS fresh clone per Pitfall 7 + CORPUS_ARGS), fire
  /tmp/p7/run-parity.sh parity unchanged; or ratify the pending_user_unblock deferral (A5).
  Expected on unblock: Gate A 0/0 x2 per corpus, Gate B zero COMPILE, Gate C multiset equal,
  Gate D reported (~0 expected).
result: [pending]

### 2. Perf gate measurement (LINT-PERF-01 letter) — unblock or ratify deferral
expected: |
  After unblock, committed perf protocol on googleapis-nocompute; verdict median(new) <=
  median(main); api-common recorded. On ratify: pending_user_unblock with the same path.
result: [pending]

### 3. Hadolint (task lint:docker) — when Docker is available
expected: |
  Dockerfile untouched by the milestone; hadolint recorded as an honest environment-unavailable
  manual item (Docker absent, rc=127). Run when Docker is available, or ratify as
  environment-constrained.
result: [pending]

### 4. End-of-phase evidence + disposition review (16 judgment-tier prohibitions)
expected: |
  Accept the automated core (Gates A-D x2 on teeth, golden 94 findings byte-exact, SC3 full
  battery 3718 tests green, wazero race fixed with two-site pin) + the honestly pending
  corpora letters; review the 16 judgment-tier prohibitions (all evidence-supported as
  honored, non-authoritative per ADR-550 D4).
result: [pending]

## Summary

total: 4
passed: 0
issues: 0
pending: 4
skipped: 0
blocked: 0

## Gaps

None — 10/12 must-haves verified with verifier's own runs (SC3 battery 3718 tests green
end-to-end re-run; race fix verified at code level + mutation check re-run RED→restore;
TestAcceptanceGates re-run green; /tmp instruments selftest exit 0; corpora unreachable
re-probed live). Status is human_needed solely for the four items above — the corpora legs
are blocked on an external resource (spike-verified twice) with a ready-to-fire one-command
path (07-USER-SETUP.md).
