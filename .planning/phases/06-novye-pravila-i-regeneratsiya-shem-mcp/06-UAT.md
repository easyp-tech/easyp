---
status: testing
phase: 06-Новые правила и регенерация схем/MCP
source: [06-VERIFICATION.md]
started: 2026-11-21T03:30:00Z
updated: 2026-11-21T03:30:00Z
---

## Current Test

number: 1
name: Ratify the corpora-leg deferral (LINT-NEWRULES-03 corpus review)
expected: |
  Ratify the deferral: the api-common/googleapis corpora leg could not run (private repo
  access gate fired; clone attempted via gh + SSH). The automated teeth-corpus parity record
  carries LINT-NEWRULES-03 (PACKAGE_NO_IMPORT_CYCLE 2/2, SYNTAX_SPECIFIED 1/1,
  FIELD_NOT_REQUIRED 1/1 vs lab CLI @ 08cef10, non-vacuous per-rule floors). The corpora leg
  moves to the Phase-7 lab.py spike (its roadmap goal owns corpus parity) with the unblock
  path documented in 06-USER-SETUP.md — OR provide repo access now and run the corpora with
  the same filtered 3-name projection.
awaiting: user response

## Tests

### 1. Ratify the corpora-leg deferral (LINT-NEWRULES-03 corpus review)
expected: |
  Ratify the deferral (teeth parity carries the requirement; corpora leg → Phase 7 lab.py
  spike with unblock path in 06-USER-SETUP.md) OR provide repo access now and run the corpora
  with the same filtered 3-name projection.
result: [pending]

### 2. Ratify the TDD gate override (plan 06-02)
expected: |
  The override recorded in STATE.md: plan 06-02 has no RED commit (RED pins lived in the
  working tree per the plan's ratified atomic-flip precedent — SC-3 «одним изменением»
  mandates one commit for rules+schemas+MCP). RED SUBSTANCE machine-proven: 6/6 per-record
  RED_EVIDENCE_OK (committed 5fb8467) + independent worktree reconstruction reproduced
  TestModuleRuleSet failure on the pre-registry tree. The commit-pattern gate cannot see
  evidence records. Ratify the override — OR reject and require a history rework (test RED
  commit split out of the atomic commit, violating SC-3).
result: [pending]

## Summary

total: 2
passed: 0
issues: 0
pending: 2
skipped: 0
blocked: 0

## Gaps

None — 8/9 must-haves verified with verifier's own runs (byte-verbatim ports diff ×3,
registry 44 = 5+22+8+7+2 matching lab line-for-line, schemas 3 enum spots × 3 names × 2
files drift-free, MCP through the real Describe pipeline, atomic commit 5be9f7e = exactly
13 files, teeth parity re-executed by the verifier). Code review CLEAN (0/0/0). Status is
human_needed solely for the two ratification items above.
