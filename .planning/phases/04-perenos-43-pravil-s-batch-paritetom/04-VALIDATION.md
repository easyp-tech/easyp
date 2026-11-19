---
phase: "04"
slug: "perenos-43-pravil-s-batch-paritetom"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-11-19"
---

# Phase 04 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (testify assertions) |
| **Config file** | none — existing colocated `_test.go` layout |
| **Quick run command** | `go test -count=1 ./internal/rules/... ./internal/core/...` |
| **Full suite command** | `go test -race -count=1 ./... && bin/golangci-lint run ./...` |
| **Estimated runtime** | ~240 seconds (api package dominates; full suite minus plugin/generation carve-out) |

---

## Sampling Rate

- **After every task commit:** Run `go test -count=1 ./internal/rules/... ./internal/core/...`
- **After every plan wave:** Run `go test -race -count=1 ./internal/rules/... ./internal/core/...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 240 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| (заполняется планами; каждая batch-ячейка — parity-тест против live-старого движка) | 0X | X | LINT-PARITY-02 | T-04-* | Идентичные замечания старое-vs-новое (rule, path, line, col) на общих фиксатурах | parity/table | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity'` | ✅ W0 (04-01: harness + pilot cells) | ✅ green (04-01) |
| (Wave 0: pilot rule carried tests) | 04-01 | 0 | LINT-PARITY-02 | T-04-01 | Main's carried expectations byte-verbatim on PackageDirectoryMatchLinked (offset 20, line 3, col 1, SourceName, message; module-qualified prefix cells) | unit | `go test -race -count=1 ./internal/rules/ -run 'TestPackageDirectoryMatch'` | ✅ W0 | ✅ green (04-01) |
| (Wave 0: disk-instrumentation cell) | 04-01 | 0 | LINT-PARITY-02 | T-04-* | readSource ровно один раз на файл корпуса (компилируемый корпус из 2 файлов с внутрипакетным импортом; total opens == 2) | unit | `go test -race -count=1 ./internal/core/ -run 'TestReadSourceOnce'` | ✅ W0 | ✅ green (04-01) |
| (Wave 0: position probe) | 04-01 | 0 | LINT-PARITY-02 | T-04-* | Decl.Pos == Meta.Pos по Offset/Line/Column закреплена для 10 kinds (11 якорей: package, import, file-level option, message, plain field, oneof, oneof member field, enum, enum value, service, rpc) | unit/probe | `go test -race -count=1 ./internal/core/ -run 'TestDeclPositionProbe'` | ✅ W0 | ✅ green (04-01) |
| (Wave 0: ModuleRule-set membership guard) | 04-01 | 0 | LINT-PARITY-02 | T-04-* | Каждый портированный ModuleRule сообщает имя из D-03 набора (04-07 усилит до exact-set) | guard | `go test -race -count=1 ./internal/rules/ -run 'TestModuleRuleSet'` | ✅ W0 | ✅ green (04-01) |
| (Batch 1 RED: MINIMAL parity suite fails on the requirePorted guard) | 04-02 | 1 | LINT-PARITY-02 | T-04-03, T-04-04 | RED-доказательство: TestLinkedParityMinimal падает на собственном requirePorted-ассерте, перечисляя непортированные DIRECTORY_SAME_PACKAGE/PACKAGE_DEFINED/PACKAGE_SAME_DIRECTORY, пока пакет компилируется (машинная валидация RED_EVIDENCE_OK) | parity/table | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityMinimal'` (RED: exit 1 + `--- FAIL: TestLinkedParityMinimal`; evidence: tdd-red-evidence-04-02-task1.json) | ✅ 04-02 | ✅ green (04-02) |
| (Batch 1: MINIMAL parity cells — first-exempt, packageless-invisible, zero-pos, pilot cumulative) | 04-02 | 1 | LINT-PARITY-02 | T-04-03, T-04-04 | 5 ячеек: first_exempt_directory (2 находки, первый файл освобождён), packageless_invisible (0 у group-правил), first_exempt_package_dir (1), package_defined_zero_pos (0:0 в fingerprint), pilot_cumulative (2 под суммарным конфигом 4 правил); старое-vs-новое мультимножества идентичны | parity/table | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityMinimal' -v` | ✅ 04-02 | ✅ green (04-02) |
| (Batch 1: carried main tests on the linked contract) | 04-02 | 1 | LINT-PARITY-02 | T-04-03, T-04-04 | Литералы main байт-verbatim: DIRECTORY_SAME_PACKAGE (offset 20/line 3/col 1, SourceName Queue), PACKAGE_DEFINED (Filename, 0:0), PACKAGE_SAME_DIRECTORY (offset 20/line 3/col 1 — реальный литерал движка; перенесённый в main литерал никогда не ассертировался) | unit | `go test -race -count=1 ./internal/rules/ -run 'TestDirectorySamePackage|TestPackageDefined|TestPackageSameDirectory'` | ✅ 04-02 | ✅ green (04-02) |
| (Batch 1: suppression centrality + readSource-once stay green) | 04-02 | 1 | LINT-PARITY-02 | T-04-* | Портированные правила не читают директивы/ignoreOnly; центральные фильтры и disk-инструментация зелёные после порта | regression | `go test -race -count=1 ./internal/core/ -run 'Suppression|ReadSourceOnce'` | ✅ W0 | ✅ green (04-02) |
| (Batch 1 gate: cumulative parity + dormancy + tidy/schema + lint) | 04-02 | 1 | LINT-PARITY-02 | T-04-* | 8 ворот: quick suite -race; TestLinkedParity* кумулятивно (8 зелёных ячеек); suppression/readSource; go build; tidy diff-empty; schema diff-empty; dormancy (builder.go + старые rule-файлы + lint*.go diff-empty); golangci-lint clean | gate | `go test -race -count=1 ./internal/core/... ./internal/rules/... && go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity' && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && bin/golangci-lint run ./internal/rules/... ./internal/core/...` | ✅ 04-02 | ✅ green (04-02) |
| (Batch 2 RED: BASIC-casing parity suite fails on the requirePorted guard) | 04-03 | 2 | LINT-PARITY-02 | T-04-05, T-04-06, T-04-07 | RED-доказательство: TestLinkedParityBasicCasing падает на собственном requirePorted-ассерте, перечисляя все 10 непортированных правил batch-2, пока пакет компилируется (машинная валидация RED_EVIDENCE_OK) | parity/table | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityBasicCasing'` (RED: exit 1 + `--- FAIL: TestLinkedParityBasicCasing`; evidence: tdd-red-evidence-04-03-task1.json) | ✅ 04-03 | ✅ green (04-03) |
| (Batch 2: BASIC-casing parity cells — casing edges, keyword positions, shallow walk, +1 enum level, import flags) | 04-03 | 2 | LINT-PARITY-02 | T-04-05, T-04-06, T-04-07 | 5 ячеек под суммарным конфигом 14 правил (10 batch-2 + 4 MINIMAL): casing_edges (9 находок: ABC, _LEADING, _field, oneTwo, A, oneof1, foo1bar, get_thing, AB — regexp main отвергает, хелперы lab пропускали бы часть), keyword_positions (6 литералов закреплены на старой стороне: message/service col 1, rpc col 3 keyword, field col 3 type token, enum keyword, value col 3), shallow_walk (0 — вложенные message/enum/oneof/field/map не репортятся), top_level_message_enum (2 — enum внутри top-level message репортится: main +1 уровень), import_flags (2 — public/weak на позиции import keyword); старое-vs-новое мультимножества идентичны | parity/table | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityBasicCasing' -v` | ✅ 04-03 | ✅ green (04-03) |
| (Batch 2: carried main tests on the linked contract) | 04-03 | 2 | LINT-PARITY-02 | T-04-05, T-04-06, T-04-07 | Литералы main байт-verbatim: MESSAGE_PASCAL_CASE (536/26/1 keyword), ENUM_PASCAL_CASE (864/49/1 + 610/31/3 вложенный), ENUM_VALUE_UPPER_SNAKE_CASE (917/51/3 + 667/33/5), FIELD_LOWER_SNAKE_CASE (560/27/3 type token), IMPORT_NO_PUBLIC (81/6/1), IMPORT_NO_WEAK (38/5/1), ONEOF_LOWER_SNAKE_CASE (748/39/3), PACKAGE_LOWER_SNAKE_CASE (20/3/1), RPC_PASCAL_CASE (300/13/3), SERVICE_PASCAL_CASE (197/10/1) | unit | `go test -race -count=1 ./internal/rules/ -run 'TestEnumPascalCase|TestEnumValueUpperSnakeCase|TestFieldLowerSnakeCase|TestImportNoPublic|TestImportNoWeak|TestMessagePascalCase|TestOneofLowerSnakeCase|TestPackageLowerSnakeCase|TestRPCPascalCase|TestServicePascalCase'` | ✅ 04-03 | ✅ green (04-03) |
| (Batch 2: suppression centrality + readSource-once stay green) | 04-03 | 2 | LINT-PARITY-02 | T-04-* | Портированные casing-правила не читают директивы/ignoreOnly; центральные фильтры и disk-инструментация зелёные после порта | regression | `go test -race -count=1 ./internal/core/ -run 'Suppression|ReadSourceOnce'` | ✅ W0 | ✅ green (04-03) |
| (Batch 2 gate: cumulative parity 14 rules + dormancy + tidy/schema + lint) | 04-03 | 2 | LINT-PARITY-02 | T-04-* | 9 ворот: quick suite -race (core + path_helpers + rules); TestLinkedParity* кумулятивно (13 зелёных ячеек, 14/41 правил); suppression/readSource; go build; tidy diff-empty; schema diff-empty; dormancy (10 старых rule-файлов + builder.go + engine files + casing.go + lint_helpers.go diff-empty); golangci-lint 0 issues; TestModuleRuleSet green; grep-clean портированных файлов на helper-имена и NamePos/NumberPos | gate | `go test -race -count=1 ./internal/core/... ./internal/rules/... && go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity' && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && bin/golangci-lint run ./internal/rules/... ./internal/core/...` | ✅ 04-03 | ✅ green (04-03) |
| (Batch 3 RED: BASIC-semantic parity suite fails on the requirePorted guard) | 04-04 | 3 | LINT-PARITY-02 | T-04-08..T-04-11 | RED-доказательство: TestLinkedParityBasicSemantic падает на собственном requirePorted-ассерте, перечисляя все 10 непортированных правил batch-3, пока пакет компилируется (машинная валидация RED_EVIDENCE_OK) | parity/table | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityBasicSemantic'` (RED: exit 1 + `--- FAIL: TestLinkedParityBasicSemantic`; evidence: tdd-red-evidence-04-04-task1.json) | ✅ 04-04 | ✅ green (04-04) |
| (Batch 3: BASIC-semantic parity cells — string compare, any-allow_alias, syntax-walk imports, option groups, written compare, bug pin, packageless) | 04-04 | 3 | LINT-PARITY-02 | T-04-08, T-04-09, T-04-10, T-04-11 | 8 ячеек под суммарным конфигом 24 правил: first_value_string_compare (3: «00»/«0x0» — строковое сравнение main, числовой порт lab прошёл бы), allow_alias_any_value (2: allow_alias=false на позиции enum keyword), import_used_verdicts (1: не-used import на позиции import keyword, quoted Location), syntax_walk_only (1: квалифицированная ссылка во вложенный message импортированного файла — резолюция пометила бы used, синтаксическая прогулка main — нет), option_groups_first_exempt (7: второй файл по всем 7 правилам, optionless-файл невидим), option_written_compare (1: «x» vs 'x' — сырое сравнение), swift_prefix_bug_pin (2 затем 0: ruby_package расходится → находка, swift_prefix расходится → тишина), packageless_files_ignored (2, ни одной находки семи option-правил); старое-vs-новое мультимножества идентичны | parity/table | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityBasicSemantic' -v` | ✅ 04-04 | ✅ green (04-04) |
| (Batch 3: carried main tests on the linked contract) | 04-04 | 3 | LINT-PARITY-02 | T-04-08, T-04-09, T-04-10, T-04-11 | Литералы main байт-verbatim: ENUM_FIRST_VALUE_ZERO (917/51/3 + 667/33/5, SourceName «4»; пустой enum — guard-пин TestEnumFirstValueZero_EmptyEnumGuard), ENUM_NO_ALLOW_ALIAS (864/49/1 + 610/31/3, SourceName social_network), IMPORT_USED (20/3/1, quoted SourceName), семь option-правил с конкретными находками от оракула: csharp 142/9/1, go_package 208/10/1, java_multiple_files 283/11/1, java_package 319/12/1, php_namespace 385/13/1, ruby_package 452/14/1, swift_prefix 452/14/1 на позиции ruby_package (bug pin) + регрессионные ячейки ruby_package расходится → находка / swift_prefix расходится → чисто | unit | `go test -race -count=1 ./internal/rules/ -run 'TestEnumFirstValueZero|TestEnumNoAllowAlias|TestImportUsed|TestPackageSame'` | ✅ 04-04 | ✅ green (04-04) |
| (Batch 3: suppression centrality + readSource-once stay green) | 04-04 | 3 | LINT-PARITY-02 | T-04-* | Портированные semantic-правила не читают директивы/ignoreOnly/Resolved; центральные фильтры и disk-инструментация зелёные после порта; grep-clean import_used_linked.go на Resolved/validateResolved | regression | `go test -race -count=1 ./internal/core/ -run 'Suppression|ReadSourceOnce'` | ✅ W0 | ✅ green (04-04) |
| (Batch 3 gate: cumulative parity 24 rules + dormancy + tidy/schema + lint) | 04-04 | 3 | LINT-PARITY-02 | T-04-* | 10 ворот на финальном состоянии: quick suite -race (core + path_helpers + rules); TestLinkedParity* кумулятивно (21 зелёная ячейка, 24/41 правил); suppression/readSource; go build; tidy diff-empty; schema diff-empty; dormancy (13 старых rule-файлов + builder.go + engine files diff-empty); golangci-lint 0 issues; TestModuleRuleSet green; grep-clean семи портов на optionString/mismatches; 8 зелёных ячеек TestLinkedParityBasicSemantic | gate | `go test -race -count=1 ./internal/core/... ./internal/rules/... && go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity' && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ && bin/golangci-lint run ./internal/rules/... ./internal/core/...` | ✅ 04-04 | ✅ green (04-04) |
| (Batch 4 RED: DEFAULT parity suite fails on the requirePorted guard) | 04-05 | 4 | LINT-PARITY-02 | T-04-12, T-04-13, T-04-14 | RED-доказательство: TestLinkedParityDefault падает на собственном requirePorted-ассерте, перечисляя все 8 непортированных правил batch-4, пока пакет компилируется (машинная валидация RED_EVIDENCE_OK; старые-side счётчики 7/2/3/10/2/3/1/0 probe-подтверждены live-старым движком до RED; zero-rules link-check чист) | parity/table | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityDefault'` (RED: exit 1 + `--- FAIL: TestLinkedParityDefault`; evidence: tdd-red-evidence-04-05-task1.json) | ✅ 04-05 | ✅ green (04-05) |
| (Batch 4: DEFAULT parity cells — prefix algorithm, first-value suffix, full-filename 0:0, per-file uniqueness, written compares, version, service suffix) | 04-05 | 4 | LINT-PARITY-02 | T-04-12, T-04-13, T-04-14, T-04-15 | 8 ячеек под суммарным конфигом 32 правил: prefix_algorithm (7: HTTPCode → H_T_T_P_CODE — оба значения issue, name-only FOB чист — bare HasPrefix; +1 уровень top-level message enum; вложенный enum — тишина), zero_value_suffix (2: первый value exact equality — BARX_UNSPECIFIED issue при проходящем HasSuffix; второй value невидим), file_name_full (3: InvalidName/aB/v1alpha на 0:0 — полное имя файла с .proto; v1alpha issue — цифра в первой компоненте, lab снял бы расширение и прошёл бы), rpc_unique_per_file (10: второй пользователь ReqT + input==output OneX в одном файле; тот же ReqT в другом файле чист), standard_name_written (2: квалифицированный .pkg.v1.GetFooRequest и GetFooReply issue на позиции rpc keyword; unqualified и service-qualified чисты), version_suffix (3: example/example.foo issue; v1 чист — unanchored regexp), service_suffix (1: User на позиции keyword col 1), central_filter_suppressible (0 затем 1 — центральный фильтр гасит) | parity/table | `go test -race -count=1 ./internal/rules/ -run 'TestLinkedParityDefault' -v` | ✅ 04-05 | ✅ green (04-05) |
| (Batch 4: carried main tests on the linked contract) | 04-05 | 4 | LINT-PARITY-02 | T-04-12, T-04-13, T-04-14, T-04-15 | Литералы main байт-verbatim: ENUM_VALUE_PREFIX (917/51/3 + 667/33/5, SourceName «none»), ENUM_ZERO_VALUE_SUFFIX (917/51/3 + 667/33/5, Suffix NONE), FILE_LOWER_SNAKE_CASE (InvalidName.proto — Filename в позиции, 0:0), PACKAGE_VERSION_SUFFIX (20/3/1, SourceName Session), SERVICE_SUFFIX (197/10/1, Suffix API, SourceName auth), RPC_REQUEST_RESPONSE_UNIQUE (375/14/3, SourceName TokenData, Collect/Finish-драйвер), RPC_REQUEST_STANDARD_NAME (214/11/3, SessionInfo), RPC_RESPONSE_STANDARD_NAME (214/11/3, Result) | unit | `go test -race -count=1 ./internal/rules/ -run 'TestEnumValuePrefix|TestEnumZeroValueSuffix|TestFileLowerSnakeCase|TestPackageVersionSuffix|TestServiceSuffix|TestRPCRequestResponseUnique|TestRPCRequestStandardName|TestRPCResponseStandardName'` | ✅ 04-05 | ✅ green (04-05) |
| (Batch 4: suppression centrality + readSource-once stay green) | 04-05 | 4 | LINT-PARITY-02 | T-04-* | Портированные DEFAULT-правила не читают директивы/ignoreOnly/Resolved; центральные фильтры и disk-инструментация зелёные после порта; grep-clean rpc_request_response_unique_linked.go на Resolved/RPCTypes; grep-clean портов на packageHasVersionSuffix/toUpperSnakeCase/NamePos/NumberPos/FilePosition | regression | `go test -race -count=1 ./internal/core/ -run 'Suppression|ReadSourceOnce'` | ✅ W0 | ✅ green (04-05) |
| (Batch 4 gate: cumulative parity 32 rules + dormancy + tidy/schema + lint) | 04-05 | 4 | LINT-PARITY-02 | T-04-* | 10 ворот на финальном состоянии: quick suite -race (core + path_helpers + rules); TestLinkedParity* кумулятивно (29 зелёных ячеек / 34 PASS с родителями, 32/41 правил); suppression/readSource; go build; tidy diff-empty; schema diff-empty; dormancy (старые DEFAULT rule-файлы + builder.go + engine files diff-empty); golangci-lint 0 issues; TestModuleRuleSet green; grep-clean трёх RPC-портов на Resolved/RPCTypes и normalisation | gate | `go test -race -count=1 ./internal/core/... ./internal/rules/... && go test -race -count=1 ./internal/rules/ -run 'TestLinkedParity' && go build ./... && go mod tidy && git diff --exit-code go.mod go.sum && go run ./cmd/easyp schema-gen && git diff --exit-code schemas/ && bin/golangci-lint run ./internal/rules/... ./internal/core/...` | ✅ 04-05 | ✅ green (04-05) |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] Parity harness (live old engine + dormant new engine in-process, sorted multiset compare) — 04-01 (`parity_harness_test.go` + `parity_wave0_test.go`, все ячейки green -race)
- [x] Позиционный probe (Decl.Pos таблица; расширен с 6 до 10 kinds / 11 якорей — все green) — 04-01
- [x] Disk-instrumentation cell (standing constraint 999.1) — 04-01 (`TestReadSourceOnce`)
- [x] Пилотное правило (PackageDirectoryMatch) как smoke батча — 04-01 (перенесено, carried-тесты green)
- [x] `Existing infrastructure covers test framework — go test готов`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| *(none — parity автоматизирован per-batch; корпусный lint-diff — ворота Phase 7)* | LINT-PARITY-01 | — | — |

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 240s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
