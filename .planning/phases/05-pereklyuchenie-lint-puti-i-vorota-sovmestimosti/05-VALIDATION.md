---
phase: "05"
slug: "pereklyuchenie-lint-puti-i-vorota-sovmestimosti"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-11-20"
---

# Phase 05 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (testify assertions) + subprocess CLI goldens (buildEasypBinary/runEasypLint) |
| **Config file** | none — existing colocated layout |
| **Quick run command** | `go test -count=1 ./internal/api/... ./internal/core/... ./internal/rules/...` |
| **Full suite command** | `go test -race -count=1 ./... && bin/golangci-lint run ./...` |
| **Estimated runtime** | ~240 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test -count=1 ./internal/api/... ./internal/core/... ./internal/rules/...`
- **After every plan wave:** Run `go test -race -count=1 ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 240 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| (заполняется планами) | 0X | X | LINT-COMPAT-01 | T-05-* | Golden full-registry corpus захвачен из старого движка ДО переключения; after-flip идентичен (incl. SourceName, D-01 порядок, COMPILE в существующем формате) | golden | `go test -race -count=1 ./internal/core/ -run 'TestFlipCorpusGolden'` | ❌ W0 | ⬜ pending |
| (заполняется планами) | 0X | X | LINT-COMPAT-01 | T-05-* | Exit-коды и CLI stdout golden через buildEasypBinary/runEasypLint; класс 1 (COMPILE) через Action — полное исполнение ratified A1 | subprocess | `go test -race -count=1 ./internal/api/ -run 'TestLintFlipExit'` | ❌ W0 | ⬜ pending |
| (заполняется планами) | 0X | X | LINT-COMPAT-02 | T-05-* | breaking/generate/tidy/migrate зелёные в том же прогоне; go.mod содержит оба модуля; breaking-path zero-diff | suite | `go test -race -count=1 ./...` | ✅ | ⬜ pending |
| (заполняется планами) | 0X | X | LINT-COMPAT-01 | T-05-* | WR-01/WR-02 исправлены ДО переключения (probe cells) | unit | `go test -race -count=1 ./internal/core/ -run 'TestResolverGuard'` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Golden full-registry corpus capture (73 findings, все 41 правило, из замороженного старого движка) — ДО flip-коммита
- [ ] WR-01/WR-02 probe cells — до flip-коммита
- [ ] `Existing infrastructure covers test framework`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| *(ожидается ратификация переключения на UAT — user-visible flip)* | LINT-COMPAT-01 | Единственная пользовательски видимая точка изменения; откат = revert | Сравнить вывод `easyp lint` до/после на своём корпусе |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 240s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
