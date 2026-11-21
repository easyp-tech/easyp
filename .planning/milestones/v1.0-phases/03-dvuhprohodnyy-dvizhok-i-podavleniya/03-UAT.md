---
status: complete
phase: 03-Двухпроходный движок и подавления
source: [03-VERIFICATION.md]
started: 2026-11-19T15:20:00Z
updated: 2026-11-19T16:54:07.771Z
---

## Current Test

[testing complete]

## Tests

### 1. Linked-path no-disk-reread review (plan 03-01 prohibition)
expected: |
  On a compiling corpus the only disk reads are pass-1 readSource, once per file; pass-2 models
  come exclusively from Linked.Syntax via protoast.Build (tree dropped after build); only imports
  (never corpus files) are read on demand, through openImportFile.
result: pass

### 2. Suppression-centrality standing constraint (plan 03-02 prohibition)
expected: |
  At each Phase 4 rule-port batch, confirm ported rules never consult comment directives or
  ignoreOnly themselves, and keep the central-filter cells (comment_suppression_central,
  finish_comment_refilter, compile_bypasses_suppression) green. The engine filters centrally;
  rules only return findings and never self-suppress.
result: pass

## Summary

total: 2
passed: 2
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

None — 19/19 must-haves verified with verifier's own runs; both flagged prohibitions ratified.

## Deferred Follow-Ups

- test: 1
  idea: "Add a disk-instrumentation cell in Phase 4 asserting readSource is invoked exactly once per corpus file on a compiling corpus (behavioral proof for the no-disk-reread guarantee, currently structural)"
  deferred_at: 2026-11-19
- test: 2
  idea: "Standing constraint for Phase 4: at each rule-port batch confirm ported rules never consult comment directives or ignoreOnly themselves; keep central-filter cells (comment_suppression_central, finish_comment_refilter, compile_bypasses_suppression) green"
  deferred_at: 2026-11-19
