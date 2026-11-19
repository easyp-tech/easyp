---
status: testing
phase: 03-Двухпроходный движок и подавления
source: [03-VERIFICATION.md]
started: 2026-11-19T15:20:00Z
updated: 2026-11-19T15:20:00Z
---

## Current Test

number: 1
name: Linked-path no-disk-reread review (plan 03-01 prohibition)
expected: |
  Review `protoInfoFromLinked` + the LintLinked pass structure (or stronger: add a
  disk-instrumentation cell in Phase 4). On a compiling corpus the only disk reads are pass-1
  readSource, once per file; pass-2 models come exclusively from Linked.Syntax via protoast.Build
  (tree dropped after build); only imports (never corpus files) are read on demand, through
  openImportFile.
awaiting: user response

## Tests

### 1. Linked-path no-disk-reread review (plan 03-01 prohibition)
expected: |
  On a compiling corpus the only disk reads are pass-1 readSource, once per file; pass-2 models
  come exclusively from Linked.Syntax via protoast.Build (tree dropped after build); only imports
  (never corpus files) are read on demand, through openImportFile.
result: [pending]

### 2. Suppression-centrality standing constraint (plan 03-02 prohibition)
expected: |
  At each Phase 4 rule-port batch, confirm ported rules never consult comment directives or
  ignoreOnly themselves, and keep the central-filter cells (comment_suppression_central,
  finish_comment_refilter, compile_bypasses_suppression) green. The engine filters centrally;
  rules only return findings and never self-suppress.
result: [pending]

## Summary

total: 2
passed: 0
issues: 0
pending: 2
skipped: 0
blocked: 0

## Gaps

None — 19/19 must-haves verified with verifier's own runs (49 core cells + 15 rules cells under
-race, dormancy sealed, tidy/schema/lint gates green); status is human_needed solely for the two
flagged prohibitions above (ADR-550 D4: favorable LLM-judge verdicts with code evidence, but no
direct behavioral test). Known correction: wazero race poisons TWO packages
(internal/adapters/plugin AND internal/generation) — both pre-existing, Phase 1 deferred root
cause.
