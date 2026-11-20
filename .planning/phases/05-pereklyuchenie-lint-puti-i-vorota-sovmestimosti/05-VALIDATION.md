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
| 05-01 T1+T2 | 01 | 1 | LINT-COMPAT-01 | T-05-05 | Golden full-registry corpus захвачен из старого движка ДО переключения (73 findings, все поля incl. Message/Filename/Offset, D-01 порядок); linked drive доказывает message-level parity; zero-COMPILE + 41-name coverage + totality guards | golden | `go test -race -count=1 ./internal/rules/ -run 'TestLintGoldenFullRegistry'` | ✅ | ✅ green |
| 05-01 T3 | 01 | 1 | LINT-COMPAT-01 | T-05-05 | CLI stdout golden (byte-exact, 2 строки) + exit 1 через реальный бинарник over findings/exclusion/directive/PDM-rewrite corpus; класс 1 (COMPILE) через Action — 05-03 исполняет ratified A1 | subprocess | `go test -count=1 ./internal/api/ -run TestLintStdoutGolden` | ✅ | ✅ green |
| (заполняется планами) | 0X | X | LINT-COMPAT-02 | T-05-* | breaking/generate/tidy/migrate зелёные в том же прогоне; go.mod содержит оба модуля; breaking-path zero-diff | suite | `go test -race -count=1 ./...` | ✅ | ⬜ pending |
| 05-02 T1-T3 | 02 | 1 | LINT-COMPAT-01 | T-05-01/02/03 | WR-01/WR-02/IN-01 исправлены ДО переключения: нерекрываемый import → command error с OS-причиной (никогда не COMPILE-remark), corpus-ветка линкера соблюдает importFileAllowed (тот же rooted-path вход, что openDiskSource — ASVS V4), ctx-guard в FindFileByPath; IN-02 keep-decision записан в doc-коммент linkSources и запинен ignored_broken_dependency_not_suppressed; freeze-гейт breaking-файлов, tidy diff-empty и golangci 0 issues зелёные | unit/integration | `go test -race -count=1 ./internal/core/ -run 'TestLintLinkedErrorPathFidelity\|TestLinkSources'` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] Golden full-registry corpus capture (73 findings, все 41 правило, из замороженного старого движка) — ДО flip-коммита (05-01, commits afc9a2f/8f88380)
- [x] WR-01/WR-02 probe cells — до flip-коммита (05-02, commits bac1dc3/94860c0; TestLintLinkedErrorPathFidelity, RED-доказательство machine-validated RED_EVIDENCE_OK)
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
