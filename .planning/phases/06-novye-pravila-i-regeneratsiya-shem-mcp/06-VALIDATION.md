---
phase: "06"
slug: "novye-pravila-i-regeneratsiya-shem-mcp"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: true) (#2117)
status: validated
nyquist_compliant: false
wave_0_complete: true
created: "2026-11-20"
---

# Phase 06 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (testify) + schema-gen drift gates + lab CLI parity compare |
| **Config file** | none |
| **Quick run command** | `go test -count=1 ./internal/rules/... ./internal/config/... ./mcp/...` |
| **Full suite command** | `go test -race -count=1 ./... && bin/golangci-lint run ./... && go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` |
| **Estimated runtime** | ~180 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test -count=1 ./internal/rules/... ./internal/config/... ./mcp/...`
- **After every plan wave:** Run `go test -race -count=1 ./...` + schema drift gate
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 180 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 06-01 T1 | 01 | 1 | LINT-NEWRULES-02 | T-06-01 | SYNTAX_SPECIFIED verbatim port; lab FilePosition 1:1 + SourceName=path pinned pre-registration (dormancy) | unit | `go test -race -count=1 ./internal/rules/ -run 'TestSyntaxSpecified'` | ✅ | ✅ green |
| 06-01 T2 | 01 | 1 | LINT-NEWRULES-02 | T-06-01 | FIELD_NOT_REQUIRED verbatim port; NamePos declared-name token (6:18/71) + bare field name + optional/repeated negatives pinned | unit | `go test -race -count=1 ./internal/rules/ -run 'TestFieldNotRequired'` | ✅ | ✅ green |
| 06-01 T3 | 01 | 1 | LINT-NEWRULES-01 | T-06-01 | PACKAGE_NO_IMPORT_CYCLE verbatim port; both cycle directions + Path pinned; dormancy: guard ten-name set green, schema drift empty | unit/guard | `go test -race -count=1 ./internal/rules/ -run 'TestPackageNoImportCycle\|TestModuleRuleSet\|TestAllRuleNames\|TestAllGroups' && go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` | ✅ | ✅ green |
| 06-02 T1 | 02 | 1 | LINT-NEWRULES-01 | T-06-* | RED machine-validated: 6 name-absence pins fail on planned assertions (RED_EVIDENCE_OK x6); uncommitted per ratified atomic-flip precedent | unit/guard | `go test -count=1 ./internal/rules/ ./internal/config/v1/ ./mcp/easypconfig/` (targets per 06-02-PLAN Task 1 verify; evidence tdd-red-evidence-06-02-task1.json) | ✅ | ✅ green (RED) |
| 06-02 T2 | 02 | 1 | LINT-NEWRULES-01/02 | T-06-* | ONE feat commit 5be9f7e: registry (lab order) + special-case deletion + 6 pin flips + golden corpus extension (73→94) + schema regen; TestModuleRuleSet 11 имён | unit/guard | `go test -race -count=1 ./internal/rules/ ./internal/config/v1/ ./mcp/easypconfig/ -run 'TestModuleRuleSet\|TestAllRuleNames\|TestAllGroups\|TestAllLintUseValues\|TestLintUseValidationParity\|TestInvalidLintNamesAreRejectedBeforeFiltering\|TestLintGoldenFullRegistry\|TestLintSelection\|TestDescribeLintSelection' && go run ./cmd/easyp schema-gen && git diff --exit-code schemas/` | ✅ | ✅ green |
| (06-03 T1) | 03 | 1 | LINT-NEWRULES-03 | T-06-* | Teeth corpus (no-syntax/required/package-cycle) активирует все три правила; parity vs lab CLI @ 08cef10 | parity | `go test -race -count=1 ./internal/rules/ -run 'TestNewRulesParity'` | ❌ W0 | ⬜ pending (06-03) |
| 06-02 T3 | 02 | 1 | LINT-NEWRULES-04 | T-06-* | Схемы включают новые имена (3 enums x 3 names per file); schema drift empty on committed state; MCP Describe отражает; battery + standing cells green | golden/drift | `go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ && go test -race -count=1 ./internal/rules/ ./internal/config/v1/ ./mcp/easypconfig/ ./internal/core/ ./internal/api/ ./internal/migration/` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] Teeth-corpus фиксатуры (no-syntax.proto, required-field proto2, 3-file/2-package import cycle) — delivered by 06-01 under repo-root `testdata/` (`no_syntax/`, `required_field/`, `import_cycle/`)
- [ ] Lab CLI binary built @ 08cef10 (researcher already built in-session; rebuild step in plan)
- [ ] `Existing infrastructure covers test framework`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Полный corpus parity на api-common/googleapis-nocompute | LINT-NEWRULES-03 | Корпусный источник (GOOGLEAPIS) отсутствует на диске — checkpoint:human-verify | Прогнать оба бинарника на корпусах при наличии; сравнить проекцию по 3 правилам |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 180s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-11-21 (nyquist audit: rows real and green; verifier re-proved SC1-SC3 with own runs)
