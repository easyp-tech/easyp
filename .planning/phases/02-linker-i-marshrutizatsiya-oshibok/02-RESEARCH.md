# Phase 2: Линкер и маршрутизация ошибок - Research

**Researched:** 2026-11-19
**Domain:** Corpus linking via vendored `easypc` compiler with three-way failure classification (COMPILE / missing-import fallback / read-syntax errors) inside `internal/core`
**Confidence:** HIGH (every load-bearing claim read in repo source this session; the lab reference implementation — эталон — is available locally at `/tmp/gsd-research/compiler-lab` and was read line-by-line)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

### Классификация «отсутствующий импорт»
- **D-01:** Детекция отсутствующего импорта — через типизированный шов резолвера, а не по тексту диагностик. Резолвер линкера строится из import roots команды (как в main, `internal/api/roots.go`) + well-known imports (`internal/easypc/protoc/wkt`); «файл не найден» резолвером (`protoc.ErrFileNotFound`) преобразуется в собственную типизированную ошибку линкера, несущую путь импорта. Диагностический текст компилятора (`Import "..." was not found or had errors.`) не парсится. [auto] Selected: «typed resolver seam» (recommended default) — надёжнее строкового матчинга и проверяется юнит-тестами.

### Гранулярность отката
- **D-02:** Откат КОРПУСНЫЙ: при отсутствии хотя бы одного импорта весь корпус пере-разбирается пофайлово существующим путём (`lintProtoInfoRead`), без `COMPILE`-замечаний; `OpenImportFileError` сохраняет свой маршрут до exit-кода 2 в `internal/api/lint.go`. Соответствует букве LINT-ENG-03 («вернуться к разбору каждого файла по отдельности») и лабораторному поведению.

### Форма COMPILE-замечаний
- **D-03:** По одной `IssueInfo` на диагностику компилятора: имя правила `COMPILE`, путь — display-path файла (как показывает компилятор), строка/колонка из `Diagnostic` (позиции компилятора уже 1-based — контракт позиций доказан в Phase 1, D-07), текст — сообщение компилятора без префиксов. Порядок выдачи — детерминированный (сортировка по файлу/позиции на стороне классификатора; main на этом классе ошибок выдаёт замечания компилятора в его порядке — паритет порядка проверяется на фиксатурах).

### Размещение и форма классификатора
- **D-04:** Классификатор живёт в `internal/core` (правки в `internal/easypc` запрещены — D-04 фазы 1). Новый файл рядом с lint-путём (имя — discretion planner'а), публичная функция/тип, которую следующий за линкером код (движок Phase 3, переключение Phase 5) будет вызывать вместо текущего пофайлового чтения. Текущий `Lint()` не меняется — до Phase 5 линкер dormant для пользователей.

### Верификация через путь Action
- **D-05:** Фиксатурная матрица трёх классов ошибок (компиляция / отсутствующий импорт / чтение-синтаксис) проверяет коды выхода CLI через путь Action — тесты вызывают тот же код маршрутизации ошибок, который вызывается в `internal/api/lint.go`, без изменения видимого пользователю поведения `easyp lint`. Матрица зелёна ДО врезки движка.

### Claude's Discretion
- Имена файлов/типов классификатора и точная сигнатура API — по фактическому коду.
- Структура и размещение фиксатур (testdata-дерево трёх классов ошибок).
- Механика связки «резолвер → import roots» (переиспользование подготовленных roots vs собственная сборка).
- Нужна ли отдельная утилита сопоставления diagnostic→issue или она встраивается в классификатор.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope

**Phase boundary (from CONTEXT.md domain, verbatim):** Линковка корпусов через `easypc` (`protoc.Link` + `WithSyntaxTrees`) с корректной трёхсторонней классификацией отказов, работающей на уровне `internal/core`... Не входит: движок правил на деревьях, подавления, детерминизм вывода (Phase 3); перенос правил (Phase 4); переключение lint-пути (Phase 5).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| LINT-ENG-02 | Хотя бы один файл корпуса не проходит линковку (ошибки компиляции, не отсутствующий импорт) → замечания с именем правила `COMPILE` (позиция и сообщение компилятора), правила на корпусе не запускаются | Эталон классификации в лабораторном `link.go` (link → diags + `!missing` → `IssueInfo` c `RuleName: "COMPILE"`); `protoc.Link` failure contract (`ErrCompilationFailed`); точная форма COMPILE issue и позиционная конвенция LineCol; precedence «syntax error побеждает COMPILE» из лабораторного `Lint` |
| LINT-ENG-03 | Линковка невозможна из-за отсутствующего импорта → вернуться к пофайловому разбору, репортить как до изменения, без COMPILE issues | Typed resolver seam: `Resolver.ErrFileNotFound` (resolver.go) + `atomic.Bool` missing-флаг эталона; резолвер строится на существующем `core.openImportFile` (importRoots + importFileAllowed + wellknownimports); `OpenImportFileError` маршрут до exit 2 (api/lint.go:90-92) |
| LINT-ENG-04 | Файл с ошибкой чтения или синтаксиса → ошибка выполнения команды для этого файла как до изменения | Lab precedence: после неудачного Link каждый файл парсится индивидуально (`parseOnTheirOwn`); синтаксическая ошибка любого файла побеждает COMPILE и уходит в command-error маршрут; read-ошибка Link (не `ErrCompilationFailed`) → fallback; текущий `lintProtoInfoRead` не меняется |
</phase_requirements>

## Project Constraints (from AGENTS.md — no CLAUDE.md exists; AGENTS.md is the source of truth)

- **Правки в `internal/easypc` ЗАПРЕЩЕНЫ** (Phase 1 D-04, отражено в AGENTS.md и VENDORED.md) — классификатор и все адаптации живут в `internal/core`.
- Colocated тесты `<file>_test.go` (testify); `require` для fatal preconditions, `assert` для non-fatal; `t.Parallel()` на верхнем уровне; named table-test cases; `-race` обязателен (`task test`).
- Ошибки: wrap как `fmt.Errorf("<callee>: %w", err)` — имя функции без префикса пакета; `errors.Is`/`errors.As` для сентинелов; не изобретать дублирующие сентинелы (ERRORS.md).
- Не коммитить: секреты, `coverage.out`, локальный бинарник `easyp` в корне.
- Новые внешние зависимости запрещены для компиляторного пути (REQ-1.2 продолжается действовать).
- Не менять `Lint()` (D-04) и не трогать breaking-путь (`protoInfoRead`/`breakingProtoInfoRead` — собственность breaking, REQ-5.2).
- После behavior changes — тесты затронутых пакетов; `task lint` требует Docker для hadolint-части (`task lint:go` — независимо).

## Summary

Phase 2 строит первый production-потребитель вендоренного компилятора: корпусный линкер с трёхсторонней классификацией отказов в `internal/core`, dormant до Phase 5. Эталон уже существует и локально доступен: лабораторный `compiler-lab/easyp-cli/internal/core/link.go` содержит готовый классификатор (`linkSources` + `lintResolver` + `CompileRuleName = "COMPILE"`), который переносится с минимальными адаптациями к типам main. Ключевой механизм — типизированный шов (D-01): обёртка `protoc.Resolver` вызывает существующий `core.openImportFile` (полная цепочка importRoots → importFileAllowed → wellknownimports) и на `OpenImportFileError`/`os.ErrNotExist` выставляет атомарный missing-флаг, возвращая `protoc.ErrFileNotFound`. Драйвер компилятора при этом теряет тип: он превращает любой неудавшийся импорт в диагностику `Import "..." was not found or had errors.` + `ErrCompilationFailed`, а разбор «syntax vs semantic» на уровне `protoc.Diagnostic` невозможен — поле `Semantic` есть только у внутреннего `parser.Diagnostic` и отбрасывается в `fromInternalDiag`. Поэтому трёхсторонняя классификация опирается на (а) missing-флаг резолвера, (б) тип возвращённой ошибки (`ErrCompilationFailed` vs прочие) и (в) precedence-механику эталона: после неудачного Link корпус парсится пофайлово, и синтаксическая ошибка любого файла побеждает COMPILE, уходя в command-error маршрут как в main.

Главное открытие с риском: **позиции диагностика считаются не как позиции правил**. `TokenStream.LineCol` (источник Line/Col у всех диагностик) даёт 1-based строку и 1-based колонку в **байтах с таб-стопом 8** (verbatim: "Tab characters ('\t') advance the 0-based column to the next multiple of 8"), тогда как контракт Phase 1 (D-07) доказал **руновые** колонки для protoast-позиций правил. На строках с табуляцией или многобайтовым UTF-8 до места ошибки COMPILE-колонки будут отличаться от руновых. Формально D-03 это допускает («строка/колонка из Diagnostic»), но формулировка «контракт позиций доказан в Phase 1, D-07» смешивает две разные системы координат — матрица должна содержать tab/UTF-8-фиксату и пинить конвенцию осознанно.

Второй тонкий пункт — D-05 «коды выхода через путь Action». Классы 2/3 проверяются через реальный пользовательский путь уже сегодня (Action/`actionV1` на фиксатурах с отсутствующим импортом и синтаксической ошибкой; факт. коды выхода — сабпроцессом, т.к. `os.Exit` убивает in-process тест). Класс 1 (COMPILE) физически не может пройти через Action до врезки движка — `Lint()` не меняется (D-04); его верификация до врезки — на уровне классификатора (ядро) + `printIssues`/`ErrHasLintIssue`-маршрута, а полный сквозной прогон той же матрицы через Action — пере-запуск в Phase 5. Это рекомендованное прочтение D-05, вынесенное в Assumptions Log для подтверждения.

**Primary recommendation:** Порт лабораторного `linkSources`/`lintResolver` в новый публичный файл `internal/core/link*.go` с трёхисходным контрактом `(linked, compileIssues)` (nil/nil = откат), резолвером на базе `core.openImportFile`, сортировкой COMPILE-issues по файлу/позиции (D-03), и фиксатурной матрицей ≥10 ячеек: класс-1 тесты на уровне классификатора + `printIssues` golden, классы 2/3 — через `actionV1`/Action сабпроцессом с кодами выхода.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Корпусная линковка (`protoc.Link` + resolver) | `internal/core` (новый классификатор) | `internal/easypc/protoc` (sealed API) | D-04: только core разговаривает с `easypc/protoc`; правки в easypc запрещены |
| Детекция отсутствующего импорта | `internal/core` (resolver wrapper) | `internal/api` (подготовка roots) | Typed seam поверх `core.openImportFile`; roots готовятся в api (`buildCore`) как сегодня |
| COMPILE-замечания (форма, сортировка) | `internal/core` (классификатор) | — | D-03 пинит форму на стороне классификатора; api-формат не меняется |
| Откат на пофайловый разбор | `internal/core` (существующий `lintProtoInfoRead`) | `internal/api` (exit-маршрут) | D-02: откат = существующий путь; wiring откатов — Phase 3/5 |
| Exit-коды 0/1/2 | `internal/api` (`Lint.Action`, `errExit`) | `cmd/easyp/main.go` (`log.Fatal` → 1) | Существующий маршрут, не меняется; матрица D-05 его только проверяет |
| Врезка в пользовательский lint | (никто в Phase 2) | Phase 5 (`buildCore`/Lint flip) | D-04: линкер dormant |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `internal/easypc/protoc` (vendored) | repo-internal, main@08cef10 (Phase 1 D-08) | `Link(ctx, files, opts...)` + `WithSyntaxTrees`/`WithResolver`/`WithErrorHandler`; типы `Diagnostic`, `Resolver`, `ErrFileNotFound`, `ErrCompilationFailed` | Единственный компиляторный API; эталон использует ровно его |
| `internal/core` (own package) | — | Классификатор, `openImportFile`, `OpenImportFileError`, `IssueInfo`/`Position` | D-04: размещение классификатора |
| stdlib `sync/atomic`, `sync` | Go 1.26+ | missing-флаг (`atomic.Bool`), защитный mutex на сбор диагностик | Эталон использует `atomic.Bool`; mutex — дешёвая страховка от контрактных изменений драйвера |
| `testify` (existing dep) | уже в go.mod | матрица и юнит-тесты | AGENTS.md-конвенция |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/easyp-tech/easyp/wellknownimports` (root, existing) | — | embedded WKT fallback внутри `openImportFile` | Автоматически через резолвер; `protoc.WithStandardImports` НЕ включать (иначе второй WKT-источник) |
| `internal/api` test harness (`actionV1` + `cli.NewContext`) | — | Action-path матрица классов 2/3 | Прецедент: `internal/api/lint_v1_test.go:24-27` |
| os/exec (stdlib) | — | сабпроцессная проверка фактических exit-кодов для матрицы | `os.Exit(1/2)` в Action недостижим in-process (CONCERNS.md:139-142) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Typed resolver seam (D-01, locked) | Парсинг текста `Import "..." was not found or had errors.` | Отвергнуто D-01: хрупко, текст внутри sealed-дерева, нет юнит-проверяемости |
| Классификация syntax-vs-semantic по `Diagnostic.Semantic` | — | **Невозможно**: `protoc.Diagnostic` не имеет поля `Semantic` — `fromInternalDiag` (diagnostic.go:31-39) копирует только Path/Line/Col/Message/IsWarn; поле есть лишь у внутреннего `parser.Diagnostic` (pkg/parser/ast.go:345-348), а правки `internal/easypc` запрещены. Precedence-механика эталона (пофайловый parse после отказа Link) решает без поля |
| Субпроцессная матрица для всех классов | In-process assertion типов ошибок | Субпроцесс даёт фактические exit-коды (требование ERRORS.md «tests that call only an internal function cannot establish the process exit code»); in-process — дополнение, не замена |
| Публичный API с `(*linkedFiles, []IssueInfo)` (лабораторная форма с outlines) | Тройка `(linked *protoc.Linked, compileIssues []IssueInfo, fallback bool)` | Outlines/`linkedFiles`-механика — собственность Phase 3; Phase 2 возвращает достаточно для Phase 3, не строя модель (точную форму выбирает planner — discretion) |

**Installation:**
```bash
# Ни одного нового пакета. Все зависимости уже в репозитории.
go build -o easyp ./cmd/easyp   # только для сабпроцессной матрицы exit-кодов
```

## Package Legitimacy Audit

**Внешние пакеты в этой фазе не устанавливаются** — фаза полностью внутренняя (vendored compiler + stdlib + существующий testify). REQ-1.2 (нулевая дельта go.mod) продолжает действовать. Gate-протокол не применялся: нечего проверять.

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

```
 corpus paths (walk, как сегодня)          [Phase 2: тесты/классификатор; Phase 3: движок; Phase 5: Lint()]
        │
        ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ internal/core: КЛАССИФИКАТОР (новый публичный файл, D-04)               │
│                                                                         │
│  paths+sources ──► protoc.Link(ctx, paths,                              │
│                      WithResolver(lintResolver),                        │
│                      WithSyntaxTrees(true),                             │
│                      WithErrorHandler(collector))                       │
│         │                        │                                      │
│         │                        ▼                                      │
│         │            ┌──────────────────────────┐                       │
│         │            │ lintResolver (typed seam)│                       │
│         │            │  linted map → bytes      │                       │
│         │            │  else core.openImportFile│──► importFileAllowed  │
│         │            │   OpenImportFileError /  │      → openDiskSource │
│         │            │   os.ErrNotExist:        │      → importRoots    │
│         │            │     missing.Store(true)  │      → wellknownimports
│         │            │     → ErrFileNotFound    │      → OpenImportFileError
│         │            └──────────────────────────┘                       │
│         ▼                                                               │
│   err == nil ──────────────► linked (деревья для Phase 3)               │
│         │                                                               │
│   err != nil ──► классификация (эталон link.go:40):                     │
│     missing.Load() ──────────────► (nil, nil) — откат корпуса (D-02)    │
│     !errors.Is(err, ErrCompilationFailed) ► (nil, nil) — откат          │
│     len(diags)==0 ───────────────► (nil, nil) — откат                   │
│     иначе ──► пофайловый syntax-чек (precedence эталона):               │
│         любой syntax-error ──────► (nil, nil) — command-error как main  │
│         все парсятся ────────────► (nil, COMPILE issues, отсорт. D-03)  │
└─────────────────────────────────────────────────────────────────────────┘
        │                                                  │
        ▼ (классы 2/3: существующий путь, без изменений)    ▼ (класс 1: issues)
  lintProtoInfoRead → readFilesFromImport            []IssueInfo (RuleName=COMPILE)
    OpenImportFileError ──► api Lint.Action ──► errExit(2)   │
    read/syntax error ────► return err ──► log.Fatal(1)      ▼
                                              printIssues → ErrHasLintIssue → os.Exit(1)
```

### Recommended Project Structure
```
internal/core/
├── link.go               # НОВЫЙ (имя — discretion): классификатор + lintResolver + CompileRuleName
├── link_test.go          # НОВЫЙ: матрица класс-1 (COMPILE) + missing + precedence
├── lint.go               # НЕ МЕНЯЕТСЯ (D-04)
├── lint_proto_info.go    # НЕ МЕНЯЕТСЯ (цель отката D-02)
├── proto_info_read.go    # НЕ МЕНЯЕТСЯ (openImportFile-цепочка — источник резолвера)
├── dom.go                # НЕ МЕНЯЕТСЯ (OpenImportFileError)
└── testdata/link/        # НОВЫЙ (discretion): фикстуры трёх классов
internal/api/
├── lint.go               # НЕ МЕНЯЕТСЯ (exit-маршрут)
└── lint_link_matrix_test.go  # НОВЫЙ: Action-path матрица (классы 2/3 + сабпроцесс exit-коды)
```

### Pattern 1: Typed resolver seam (D-01, эталон)
**What:** `lintResolver` реализует `protoc.Resolver`; linted-файлы отдаются из памяти (map путь→байты), импорты — через существующий `c.openImportFile` (полная граница: IsLocal → importFileAllowed → openDiskSource → importRoots → wellknownimports). `OpenImportFileError`/`os.ErrNotExist` → `missing.Store(true)` + `protoc.ErrFileNotFound`; прочие ошибки — проброс без изменения класса.
**When to use:** всегда; это и есть D-01.
**Example:**
```go
// Source: эталон compiler-lab/easyp-cli/internal/core/link.go:58-85 (прочитан 2026-11-19)
type lintResolver struct {
	core    *Core
	disk    FS
	linted  map[string][]byte
	missing atomic.Bool // an import could not be found
}

// FindFileByPath implements protoc.Resolver.
func (r *lintResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	if src, ok := r.linted[path]; ok {
		return src, nil
	}
	f, err := r.core.openImportFile(r.disk, path)
	if err != nil {
		var notFound *OpenImportFileError
		if errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist) {
			r.missing.Store(true)
			return nil, protoc.ErrFileNotFound
		}
		return nil, err
	}
	defer r.core.close(ctx, f, path)
	src, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %w", err)
	}
	return src, nil
}
```
Замечание: `openImportFile` возвращает `&os.PathError{...os.ErrNotExist}` для файла, заблокированного `importFileAllowed` (proto_info_read.go:110-117) — `errors.Is(err, os.ErrNotExist)` ловит и это: заблокированный импорт классифицируется как missing → откат → пофайловый путь выдаст `OpenImportFileError` (exit 2) ровно как main. «Без обхода файловой системы» гарантировано конструкцией: `Link` зовёт `cfg.sources(ctx, false)` — без `DirResolver`-ов и с os.Stat-фоллбеком, закрытым флагом `withDirs` (compile.go:147-263, 227-248).

### Pattern 2: Three-way classification (эталон, link.go:40-55)
**What:** после `Link` classify по трём признакам — missing-флаг, тип ошибки, наличие диагностик.
```go
// Source: эталон link.go:31-55 (verbatim)
var diags []protoc.Diagnostic
linked, err := protoc.Link(ctx, paths, protoc.WithResolver(resolver), protoc.WithSyntaxTrees(true), protoc.WithErrorHandler(func(d protoc.Diagnostic) {
	if !d.IsWarn {
		diags = append(diags, d)
	}
}))
if err == nil {
	return &linkedFiles{linked: linked, outlines: make(map[string]*linkedOutline)}, nil
}
if resolver.missing.Load() || !errors.Is(err, protoc.ErrCompilationFailed) || len(diags) == 0 {
	c.logger.Debug(ctx, "link failed, lint rules use the syntax only", slog.String("error", err.Error()))
	return nil, nil
}
issues := make([]IssueInfo, 0, len(diags))
for _, d := range diags {
	issues = append(issues, IssueInfo{
		Path: d.Path,
		Issue: Issue{
			Position: Position{Line: d.Line, Column: d.Col}, // verbatim эталона link.go:49; поле Position.Column = protoast.Position (dom.go:20)
			Message:  d.Message,
			RuleName: CompileRuleName,
		},
	})
}
return nil, issues
```
**Почему трёх исходов достаточно:** (1) `(linked, nil)` — корпус слинкован, правила пойдут по деревьям (Phase 3); (2) `(nil, issues)` — COMPILE-класс, LINT-ENG-02; (3) `(nil, nil)` — откат: и missing (LINT-ENG-03), и read-ошибка Link, и syntax-победа (LINT-ENG-04) уходят в существующий путь, чьё поведение и есть «как до изменения». Точная сигнатура (включая возврат linked-деревьев для Phase 3) — discretion planner'а.

### Pattern 3: Syntax-error precedence (LINT-ENG-04, эталон lint.go:60-77)
**What:** COMPILE-issues возвращаются только если НИ один файл корпуса не падает при индивидуальном парсе. Эталон после неудачного Link парсит каждый файл (`parseOnTheirOwn` → `protoast.Parse`), заполняя per-file ошибки; COMPILE выдаётся только при пустых per-file ошибках — иначе процесс идёт в пофайловый путь, где первая ошибка становится command-error (`fs.WalkDir: ... c.protoInfoRead: %w`) как в main. Без этого шага синтаксическая ошибка одного файла выдалась бы как COMPILE и нарушила бы LINT-ENG-04.
**Why:** `protoc.Diagnostic` не различает syntax и semantic (нет поля `Semantic` на шве) — различение достигается поведением (индивидуальный parse), не классификацией диагностики.

### Pattern 4: COMPILE issue shape (D-03)
```go
// Форма (эталон link.go:44-54): одна IssueInfo на диагностику-ошибку (IsWarn дропается)
IssueInfo{
	Path: d.Path, // display-path как у компилятора: для файла корпуса — путь из walk; для dep — import-путь
	Issue: Issue{
		Position: Position{Line: d.Line, Column: d.Col}, // Filename/Offset пустые — текстовый принтер их не печатает
		Message:  d.Message,                              // сообщение компилятора без префиксов
		RuleName: "COMPILE",                              // const CompileRuleName
	},
}
```
Плюс D-03: сортировка по файлу/позиции на стороне классификатора (эталон НЕ сортирует — осознанная дельта D-03; api-слой дополнительно стабильно сортирует по Path — lint_v1.go:220-222). JSON-формат принтера сериализует Position целиком — пустые Filename/Offset у COMPILE будут видны в JSON-выводе; зафиксировать в матрице (см. Pitfall 5).

### Anti-Patterns to Avoid
- **Парсинг текста диагностик:** `Import "..." was not found or had errors.` — только человеческий вывод; классификация — по missing-флагу (D-01).
- **Полагаться на `Diagnostic.Semantic`:** поля нет на шве `protoc.Diagnostic`; правки easypc запрещены.
- **Врезка линкера в `Lint()`** в этой фазе: D-04 прямо запрещает; изменение поведения на сломанных корпусах — видимое.
- **`WithStandardImports(true)` в Link:** создаст второй WKT-источник (`protoc/wkt`) параллельно `wellknownimports` из `openImportFile` — расхождение с main по набору резолвимых WKT.
- **Файловый обход в резолвере:** любой `os.Stat`/walk в resolver-обёртке ломает границу `importFileAllowed` (PITFALLS security row 1) — резолвер только через `openImportFile`.
- **Правки внутри `internal/easypc`:** синк с compiler-lab затрёт их (D-04 Phase 1); всё в `internal/core`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Разрешение импортов | Собственный search-path/файловый обход | `core.openImportFile` (существующий, proto_info_read.go:103-147) | Уже реализует IsLocal + importFileAllowed + importRoots + wellknownimports с точной семантикой main; дубль = расхождение границы модулей |
| Парсер/линкер корпуса | Собственный обход AST/символов | `protoc.Link(..., WithSyntaxTrees(true))` | Весь пайплайн в sealed-дереве; эталон использует его as-is |
| Компиляторные диагностики → строки | Свое форматирование | `Diagnostic.String()` существует (diagnostic.go:18-29); для issues — printer api | Формат `path:line:col: message` уже каноничен |
| Exit-code маршрутизация | Новые сентинелы/маппер | Существующий `Lint.Action` + `errExit` + `log.Fatal` | ERRORS.md: «no universal sentinel-to-exit mapper»; матрица проверяет существующий |
| Тип «файл не найден» | Новый типизированный тип ошибки линкера (буква D-01 «собственная типизированная ошибка») | Либо существующий `OpenImportFileError` как носитель пути, либо тонкая обёртка в core | `OpenImportFileError{FileName}` уже несёт путь и уже маршрутизируется в exit 2; inventing параллельного типа заставит api различать два класса «not found». Формулировка D-01 выполнима минимальной обёрткой/пробросом `OpenImportFileError` — решение за planner'ом (discretion API) |

**Key insight:** вся ценность фазы — корректная маршрутизация существующих, уже типизированных событий; ничего нового не изобретается, всё связывается.

## Runtime State Inventory

Пропущено: фаза добавляет новую capability (линкер), не переименовывает и не мигрирует строки/данные. Единственный «runtime state» — зафиксированный SHA эталона (main@08cef10, VENDORED.md) — не меняется.

## Common Pitfalls

### Pitfall 1: Драйвер стирает тип ошибки импорта — классификация только через резолвер
**What goes wrong:** Разработчик ждёт `ErrFileNotFound` от `Link` и строит классификацию по returned error.
**Why it happens:** Драйвер ловит ошибку `resolvePath`/`readFile` импорта и превращает её в диагностику: `Message: fmt.Sprintf("Import %q was not found or had errors.", impPath)` + возвращает `ErrCompilationFailed` (driver.go:341-360). Прямой путь до резолверной ошибки от `Link` не доходит никогда.
**How to avoid:** missing-флаг в обёртке резолвера (Pattern 1) — единственный канал. Для ВЕРХНЕУРОВНЕВОГО (top-level) входного файла — наоборот: неудача resolve возвращается как plain error `fmt.Errorf("%s: File not found.", path)` (driver.go:375-379, compile.go:250) БЕЗ ErrCompilationFailed и без диагностики → классификатор уводит в откат, пофайловый путь воспроизведёт read-ошибку как main.
**Warning signs:** Тест «missing import» внезапно ожидает COMPILE.

### Pitfall 2: Позиции диагностика — таб-стоп 8 и байты, не руны
**What goes wrong:** COMPILE issue колонка на строке с табом/эмодзи «неожиданная» (например, после одного таба колонка 9, а не 2).
**Why it happens:** Все Line/Col диагностик идут через `TokenStream.LineCol`, чей контракт verbatim: "LineCol computes 1-based line and 1-based column with tab stop 8 for offset. Tab characters ('\t') advance the 0-based column to the next multiple of 8: col0 = (col0/8 + 1) * 8. All other bytes advance col0 by 1." (lexer/token.go:121-123); цикл `for _, b := range src[lineStart:offset]` по `[]byte` считает байты — многобайтовый UTF-8 двигает колонку на каждый байт. Контракт Phase 1 D-07 доказал РУНОВЫЕ колонки для protoast-позиций правил (position_parity_test.go, включая "field after one leading tab" и emoji-кейсы) — это другая система координат.
**How to avoid:** Матрица обязана содержать фиксату с табом и многобайтовым UTF-8 до места ошибки и пинить именно LineCol-значения. D-03 («строка/колонка из Diagnostic») этому соответствует, но формулировку «контракт доказан D-07» нельзя читать как «колонки руновые».
**Warning signs:** Расхождение ожиданий в матрице только на не-ASCII/таб-фиксатурах.

### Pitfall 3: Синтаксическая ошибка классифицируется как COMPILE
**What goes wrong:** Корпус с синтаксической ошибкой выдаёт COMPILE-замечания (exit 1, suppressed findings) вместо command-error для файла — нарушение LINT-ENG-04.
**Why it happens:** `ast.Errors` (синтаксис) и ошибки линкера (семантика) текут через один `ErrorHandler` как `ErrCompilationFailed`; на шве их не различить.
**How to avoid:** Precedence-механика эталона (Pattern 3): после неудачного Link — пофайловый parse каждого файла; только при полном отсутствии per-file ошибок возвращать COMPILE. Фиксатурная ячейка «syntax error + семантически-чистый остальной корпус» обязательна.
**Warning signs:** Матрица без ячейки пересечения классов; линтер молча меняет класс ошибки.

### Pitfall 4: Отсутствующий импорт в dep-файле / заблокированный импорт / WKT
**What goes wrong:** Транзитивно отсутствующий импорт классифицируется как COMPILE (диагностика указывает на dep-файл) или, наоборот, заблокированный `importFileAllowed` файл читается.
**Why it happens:** Импорт dep-файла резолвится тем же резолвером: если dep импортирует missing → `missing.Store(true)` срабатывает на деп-глубине — флаг корпусный, не пофайловый; заблокированный файл возвращает `PathError{os.ErrNotExist}` — легко перепутать с честной read-ошибкой.
**How to avoid:** Ячейки матрицы: missing-импорт на глубине 2 (root→dep→missing); импорт, запрещённый `importFileAllowed`; WKT-импорт (`google/protobuf/timestamp.proto`) — резолвится, класс compiling. Все три уже покрыты существующей `openImportFile`-семантикой — резолвер не добавляет своей логики поиска.
**Warning signs:** COMPILE-issue с путём dep-файла при классе missing.

### Pitfall 5: Форма COMPILE issue и «невидимые» поля
**What goes wrong:** Расхождение формата при Phase 5 flip: сортировка, пустые `SourceName`/`Filename`/`Offset`, warns.
**Why it happens:** Эталон оставляет `SourceName` пустым, Position без Filename/Offset, дропает `IsWarn`-диагностики, не сортирует. Текстовый принтер печатает `%s:%d:%d:%s %s (%s)` (lint.go:125-132) — пустой SourceName даст `path:4:3: message (COMPILE)` с двойным пробелом-артефактом... точнее `": (COMPILE)"`-склейку как у main для пустых source names. JSON-принтер сериализует Position целиком — пустые Filename/Offset видны.
**How to avoid:** D-03 пинит форму (RuleName COMPILE, display-path, Diagnostic Line/Col, текст без префиксов, сортировка в классификаторе). Warn-политика эталона (дроп) — принять и зафиксировать тестом (Assumption A3). Golden-тест `printIssues` для COMPILE-issues — до Phase 5.
**Warning signs:** Разнобой между классификаторным и api-порядком; JSON-дифф на пустых полях.

### Pitfall 6: In-process тесты не могут установить exit-коды
**What goes wrong:** Матрица «зелёная», а реальный `easyp lint` выходит не с тем кодом.
**Why it happens:** `Action` вызывает `os.Exit(1)` (`ErrHasLintIssue`) и `errExit(2)` — процесс теста умрёт или ветка не выполнится; ERRORS.md: «tests that call only an internal function cannot establish the process exit code»; CONCERNS.md:139-142 фиксирует это как непротестированную область.
**How to avoid:** Двухуровневая матрица D-05: in-process через `actionV1` — assertion типов ошибок (`OpenImportFileError` / plain / `ErrHasLintIssue`) + сабпроцесс (собранный бинарник) — assertion фактических exit-кодов и stdout-байтов. Для класс-1 — классификатор + `printIssues` golden до врезки; полный Action-прогон той же матрицы — пере-запуск в Phase 5 (A1).
**Warning signs:** Матрица проверяет только возвращённые error-типы и называет это «exit-кодами».

### Pitfall 7: Конкурентный вызов резолвера и сбор диагностик
**What goes wrong:** Data race в матрице под `-race` при корпусах >1 файла.
**Why it happens:** `Link` по умолчанию ставит Threads=NumCPU; драйвер параллельно preload-ит входные файлы (threads>1 && len(inputs)>1) и зовёт `ReadFile`/`ResolvePath` (то есть резолвер) из нескольких горутин (driver.go:148-202). Эталон собирает `diags` в слайс без mutex — безопасен, только если `ErrorHandler` зовётся последовательно (сейчас так: preload ошибки репортит последовательный `parseOne`), но контракт не документирован, а синк compiler-lab может его изменить.
**How to avoid:** `missing` — `atomic.Bool` (как эталон); сбор `diags` — под маленьким mutex (дешёвая страховка); матрица гоняется под `-race` с корпусом из ≥2 файлов. Правки внутрь easypc для документации контракта запрещены.
**Warning signs:** `-race` фейл только на multi-file фикстурах.

### Pitfall 8: Диагностика рекурсивного импорта с жёсткой позицией
**What goes wrong:** COMPILE issue для цикла импортов показывает Line 2/Col 1 всегда — не удивляться в матрице.
**Why it happens:** Драйвер жёстко ставит `Line: 2, Col: 1` в диагностике "File recursively imports itself: ..." (driver.go:319-323).
**How to avoid:** Отдельная ячейка матрицы на import-cycle; ожидание — фиксированная позиция + полный cycle-chain в сообщении.
**Warning signs:** «Случайные» 2:1 в золотых файлах.

## Code Examples

### Классификатор: итоговый контур (портированный эталон, адаптации минимальны)
```go
// Source: эталон link.go, адаптировано к main-типам; правки easypc запрещены (D-04)
const CompileRuleName = "COMPILE" // Source: эталон link.go:18 (verbatim)

// classifyLinks links the corpus. Outcomes:
//   linked != nil            — corpus linked (Phase 3 consumes trees)
//   compileIssues != nil     — LINT-ENG-02: COMPILE issues, rules must not run
//   both nil                 — fall back to the existing per-file path (D-02/LINT-ENG-03/04)
func (c *Core) classifyLinks(ctx context.Context, disk FS, paths []string, sources [][]byte) (*protoc.Linked, []IssueInfo) {
	linted := make(map[string][]byte, len(paths))
	for i, path := range paths {
		linted[path] = sources[i]
	}
	resolver := &lintResolver{core: c, disk: disk, linted: linted}
	var mu sync.Mutex
	var diags []protoc.Diagnostic
	linked, err := protoc.Link(ctx, paths,
		protoc.WithResolver(resolver),
		protoc.WithSyntaxTrees(true),
		protoc.WithErrorHandler(func(d protoc.Diagnostic) {
			if d.IsWarn {
				return
			}
			mu.Lock()
			diags = append(diags, d)
			mu.Unlock()
		}))
	if err == nil {
		return linked, nil
	}
	if resolver.missing.Load() || !errors.Is(err, protoc.ErrCompilationFailed) || len(diags) == 0 {
		return nil, nil // откат: missing / не-компиляционная ошибка / пустой диагностический набор
	}
	// LINT-ENG-04 precedence: syntax error в любом файле побеждает COMPILE (Pattern 3)
	if c.hasPerFileSyntaxError(ctx, disk, paths, sources) {
		return nil, nil
	}
	// D-03: сортировка по файлу/позиции на стороне классификатора
	slices.SortStableFunc(diags, func(a, b protoc.Diagnostic) int {
		if c := strings.Compare(a.Path, b.Path); c != 0 {
			return c
		}
		return cmp.Or(cmp.Compare(a.Line, b.Line), cmp.Compare(a.Col, b.Col))
	})
	issues := make([]IssueInfo, 0, len(diags))
	for _, d := range diags {
		issues = append(issues, IssueInfo{
			Path: d.Path,
			Issue: Issue{
				Position: Position{Line: d.Line, Column: d.Col},
				Message:  d.Message,
				RuleName: CompileRuleName,
			},
		})
	}
	return nil, issues
}
```
Сигнатура/имена — discretion planner'а; обязательны три исхода, missing-флаг, precedence и сортировка.

### Матрица класс-1 (уровень классификатора; эталон теста lint_compile_test.go:16-31)
```go
// Source: эталон lint_compile_test.go (verbatim-числа)
raw := "syntax = \"proto3\";\npackage example;\nmessage bad_name {\n  required string s = 1;\n}\n"
// ожидание эталона:
// got[0].RuleName == "COMPILE"; got[0].Path == "test.proto"
// got[0].Message == "Required fields are not allowed in proto3."
// got[0].Position.Line == 4; got[0].Position.Column == 3
```
Плюс ячейки: unknown type, duplicate symbol, import cycle (Line 2 Col 1), semantic error в dep-файле (Path = import-путь dep), таб/UTF-8-позиция (Pitfall 2), twice-run детерминизм порядка.

### Матрица классов 2/3 (Action-path; прецедент lint_v1_test.go:24-27 + сабпроцесс)
```go
// In-process: тип ошибки, которую Action маршрутизирует
err := (Lint{}).actionV1(ctx, logger.NewNop(), configPath, root, root)
var notFound *core.OpenImportFileError
require.ErrorAs(t, err, &notFound)            // missing import → exit-2 класс
// syntax/read fixture → require.Error(t, err) без ErrorAs на OpenImportFileError → exit-1 класс
// Субпроцесс (единственный способ фактических кодов, ERRORS.md):
out, code := runBuiltEasyp(t, "lint", "--path", fixtureDir) // code == 2 / 1 / 0, stdout bytes
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Классификация отсутствующего импорта по тексту диагностики | Typed resolver seam (`ErrFileNotFound` + missing-флаг) | D-01 (2026-11-19) | Юнит-проверяемо; не зависит от текстов sealed-дерева |
| Различение syntax/semantic по полю диагностики | Поведенческое (пофайловый parse после отказа Link) | Эталон; зафиксировано 2026-11-19 | Единственный путь без правок easypc |
| `protoc` API как внутренность lab-CLI | Первый production-потребитель `protoc.Link` в easyp | Эта фаза | API-ea поверхность Phase 3/5 опирается на контракт этой фазы |

**Deprecated/outdated:**
- Полагаться на `Diagnostic.Semantic` (упомянут в code_context CONTEXT.md): поля нет на шве `protoc.Diagnostic` — см. Assumption A5.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | D-05 читается как: класс-1 (COMPILE) до врезки верифицируется на уровне классификатора + `printIssues`/`ErrHasLintIssue`-маршрута; полный сквозной прогон всей матрицы через Action с фактическими exit-кодами достижим для классов 2/3 сейчас (реальный путь) и для класса 1 — пере-запуском той же матрицы в Phase 5. Альтернативное прочтение («все три класса через Action до врезки») технически недостижимо без изменения `Lint()` (D-04) | Common Pitfalls 6, Validation Architecture | Матрица формально не покроет «коды выхода через Action» для класса 1 до Phase 5; обсудить/подтвердить у пользователя, если ожидание иное |
| A2 | Колонки COMPILE-диагностик следуют LineCol-конвенции (таб-стоп 8, байты), и это принимается осознанно как буква D-03; формулировка D-03 о «контракте D-07» относится к 1-based-ности, а не к руновым колонкам | Pitfall 2 | Позиции COMPILE на таб/не-ASCII строках неожиданны для пользователей/матрицы; потребует явного решения (руновая конверсия) при неприёмке |
| A3 | Warn-диагностики компилятора дропаются из COMPILE-репорта (эталон `if !d.IsWarn`); D-03 молчит про warns | Pattern 2, Pitfall 5 | Пользователь не увидит компиляторные предупреждения; при необходимости — добавить как отдельный класс (Phase 5-совместимо) |
| A4 | Импорт с нелокальным путём (`"../x.proto"`) классифицируется как COMPILE (эталон: `invalid import path` не ErrNotExist → missing не выставляется → diags непусты), тогда как main сегодня отдаёт plain error → exit 1. Оба маршрута дают exit 1; презентация меняется с error-строки на COMPILE-замечания | Pitfall 4, Open Questions 2 | Тонкое видимое отличие от main на экзотических импортах; если дух LINT-ENG-04 трактовать шире — уводить в откат |
| A5 | Code-insight из CONTEXT.md «`Diagnostic{Path, Line, Col, Message, IsWarn, Semantic}`» неточен для шва: `protoc.Diagnostic` (diagnostic.go:10-16) поля `Semantic` не имеет (верифицировано Read); поле есть у внутреннего `parser.Diagnostic` (pkg/parser/ast.go:345-349) и отбрасывается `fromInternalDiag` (diagnostic.go:31-39). Решения D-01/D-03 от этого не зависят | Standard Stack (Alternatives), Pattern 3 | Планировщик мог бы заложиться на несуществующий дискриминатор |
| A6 | Контракт «ErrorHandler зовётся последовательно» у драйвера сохраняется до синков; защитный mutex в коллекторе — достаточная страховка | Pitfall 7 | Синк compiler-lab изменит контракт → гонка; mutex снимает риск без правок easypc |
| A7 | `task` CLI отсутствует на PATH в текущем окружении (task: MISSING, gotestsum: MISSING, bin/ содержит только golangci-lint); верификация — прямые go-команды из AGENTS.md | Environment Availability | `task test` недоступен как есть; эквиваленты документированы, `task init` требует установки task |

## Open Questions (RESOLVED)

> Resolution status: Q2 and Q3 are resolved and substantively adopted in plan 02-02 content (annotations below). Q1 is deliberately NOT auto-resolved — it is **surfaced-for-ratification**: its adopted route (A1 two-level verification) is implemented as plan 02-03's flagged assumption pending user ratification at phase verification.

1. **Механика D-05 для класса 1 до врезки (A1).**
   - What we know: классы 2/3 проверяемы через реальный Action-путь сейчас; класс 1 не достигает Action без изменения `Lint()`.
   - What's unclear: устраивает ли пользователя верификация класса 1 на уровне классификатора + printIssues до врезки, с полным Action-прогоном в Phase 5.
   - Recommendation: да (двухуровневая матрица); зафиксировать в плане как интерпретацию D-05.
   - **Resolution (surfaced-for-ratification, not silently resolved):** adopted route = A1 two-level verification — класс 1 проверяется на уровне классификатора (plans 02-01/02-02) + printIssues golden (plan 02-03), полный Action-прогон матрицы повторяется в Phase 5. Implemented as plan 02-03's flagged assumption bullet «LINT-ENG-02 — unresolved»; user ratification happens at phase verification.
2. **Классификация нелокального импорта (`"../x.proto"`) — COMPILE (эталон) или откат (дух «как в main»)?**
   - What we know: exit 1 в обоих случаях; отличается презентация (COMPILE-замечания vs error-строка).
   - What's unclear: какое поведение считается целевым для Phase 5-паритета.
   - Recommendation: следовать эталону (COMPILE) — буква REQ-2.2 («ошибки компиляции, не связанные с отсутствующим импортом»); ячейку матрицы включить обязательно.
   - **Resolution (adopted):** эталон — COMPILE. Implemented verbatim as plan 02-02's `non_local_import_compile` cell (Task 2 behavior + must_haves truth «non-local import paths classify COMPILE — etalon behavior per research A4 / Open Question 2» + threat T-02-03 mitigation).
3. **Носитель типизированной ошибки missing (буква D-01).**
   - What we know: `OpenImportFileError{FileName}` уже несёт путь и маршрутизируется в exit 2; эталон пробрасывает его как есть, добавляя только missing-флаг резолвера.
   - What's unclear: требуется ли D-01 буквальный НОВЫЙ тип «типизированной ошибки линкера» или проброс/обёртка `OpenImportFileError` удовлетворяют.
   - Recommendation: минимальная обёртка/проброс существующего типа; новый тип вводить только если Phase 3/5 потребуется различать «missing при линковке» от «missing при пофайловом чтении».
   - **Resolution (adopted):** проброс существующего `*OpenImportFileError`, новый тип не вводится. Implemented as plan 02-02's acceptance criterion «no new error type is declared in internal/core/link.go (missing-import typing reuses *OpenImportFileError)» and its must_haves truth carrying the exit-2 carrier (D-01/D-02).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | всё | ✓ | go1.27.3 (go.mod: 1.26.6) | — |
| `task` | `task test`/`task lint` таргеты | ✗ | — | Прямые команды AGENTS.md: `go test -race -count=1 ./...`, `go build -o easyp ./cmd/easyp` |
| gotestsum | `task test` runner | ✗ | — | `go test` напрямую (того же качества: `-race -count=1`) |
| golangci-lint | `task lint:go` | ✓ (в `./bin/golangci-lint`) | repo-pinned | `./bin/golangci-lint run ...` |
| compiler-lab clone | эталон для сверки | ✓ | `/tmp/gsd-research/compiler-lab` (easyp-cli) | `gh repo clone easyp-tech/compiler-lab` (D-09) |
| Docker | `task lint` hadolint-часть | не проверялось в этой сессии | — | `task lint:go` независимо |

**Missing dependencies with no fallback:** none (все критические пути — чистый Go).
**Missing dependencies with fallback:** `task`/gotestsum → прямые go-команды (документированы в AGENTS.md как эквиваленты).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib testing + testify (существующая зависимость), `-race -count=1` |
| Config file | none — конвенции в AGENTS.md/.spec/agent-rules.md |
| Quick run command | `go test -race -count=1 ./internal/core/ ./internal/api/` |
| Full suite command | `go test -race -count=1 ./...` (эквивалент `task test`; task/gotestsum отсутствуют — см. Environment) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LINT-ENG-02 | COMPILE-issues: форма (RuleName/Path/Line/Col/Message), сортировка, warns дроп, precedence syntax>COMPILE, cycle-диагностика | unit/integration (core, testify) | `go test -race -count=1 ./internal/core/ -run TestLink` | ❌ Wave 0 |
| LINT-ENG-02 | COMPILE в существующем формате вывода (`path:line:col: message (COMPILE)`; JSON с пустыми Filename/Offset) | golden (api printIssues) | `go test -race -count=1 ./internal/api/ -run TestPrintIssues` | ❌ Wave 0 |
| LINT-ENG-03 | missing import → откат, `OpenImportFileError` до exit-2 маршрута; транзитивный missing; заблокированный importFileAllowed | integration (api actionV1 + сабпроцесс exit-код) | `go test -race -count=1 ./internal/api/ -run TestLintLinkMatrix` | ❌ Wave 0 |
| LINT-ENG-04 | read/syntax error → command error как main; precedence над COMPILE | integration (api actionV1 + сабпроцесс exit-код; core precedence unit) | `go test -race -count=1 ./internal/api/ -run TestLintLinkMatrix` | ❌ Wave 0 |
| (cross) | Детерминизм: дважды на одной матрице — идентичный порядок | integration | тот же прогон ×2 | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test -race -count=1 ./internal/core/ ./internal/api/`
- **Per wave merge:** `go test -race -count=1 ./...` + `./bin/golangci-lint run ./internal/core/... ./internal/api/...`
- **Phase gate:** полная матрица зелёная ×2 (детерминизм) + `task test`-эквивалент зелёный до `/gsd:verify-work`; матрица обязана быть зелёной ДО любой врезки (D-05)

### Wave 0 Gaps
- [ ] `internal/core/link_test.go` — матрица класс-1 + missing + precedence (LINT-ENG-02/03/04 core-уровень)
- [ ] `internal/api/lint_link_matrix_test.go` — Action-path матрица классов 2/3 + сабпроцессные exit-коды
- [ ] `internal/core/testdata/link/` — фикстуры трёх классов (discretion: дерево)
- [ ] Framework install: не требуется (Go + testify уже есть)

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | CLI, без аутентификации |
| V3 Session Management | no | нет сессий |
| V4 Access Control | yes (файловая граница) | `importFileAllowed` + `openImportFile` — резолвер не добавляет путей доступа (см. Threats) |
| V5 Input Validation | yes | Недоверенные proto-корпуса через компилятор; матрица покрывает malformed-классы; typed errors вместо парсинга текста |
| V6 Cryptography | no | нет криптографии |

### Known Threat Patterns for {easypc linker}

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Резолвер линкера обходит `importFileAllowed` (чтение файлов вне закреплённой выборки модулей) | Information Disclosure | Резолвер ТОЛЬКО через `core.openImportFile`; ячейка матрицы на заблокированный импорт (ожидание: missing → откат → exit 2 как main); никакого os.Stat/обхода в обёртке |
| Паника вендоренного компилятора на malformed-входе → crash вместо типизированной ошибки | Tampering (DoS CLI) | Матрица прогоняет malformed-фикстуры через классификатор под `-race`; recover-граница — осознанное решение Phase 3/5 (последний рубеж), фаззинг — FUTURE-HARDENING-01 (out of scope) |
| Эскалация привилегий чтения через craft-путь импорта (`../`, абс. пути) | Elevation/Info Disclosure | `filepath.IsLocal` внутри `openImportFile` уже отсекает; нелокальный импорт — ячейка матрицы (см. Open Question 2) |

Проверка ASVS L1 по затронутой поверхности (CLI, локальные файлы, без сети/аутентификации): применимы только V4/V5-строки выше; блокирующих high-находок дизайн не создаёт.

## Sources

### Primary (HIGH confidence)
- `internal/easypc/protoc/link.go` (Read, 1-138) — контракт `Link`, `WithSyntaxTrees`, `ErrCompilationFailed`
- `internal/easypc/protoc/resolver.go` (Read, 1-98) — `Resolver`, `ErrFileNotFound`, `MultiResolver`
- `internal/easypc/protoc/options.go` (Read, 1-88) — полный набор опций
- `internal/easypc/protoc/compile.go` (Read, 147-263) — `cfg.sources(ctx, false)`: резолверная цепочка Link без dir/os-фоллбека
- `internal/easypc/protoc/diagnostic.go` (Read, 1-39) — `Diagnostic` без `Semantic`; `fromInternalDiag`
- `internal/easypc/pkg/compiler/driver.go` (Read, 1-628) — import-failure → diagnostic+ErrCompilationFailed; top-level resolve fail → plain error; preload-параллелизм; cycle-диагностика 2:1
- `internal/easypc/pkg/lexer/token.go` (Read, 110-154) — LineCol: 1-based, таб-стоп 8, байтовые колонки
- `internal/easypc/pkg/parser/ast.go` (grep, 339-349) — внутренний `Diagnostic.Semantic`
- `internal/core/lint.go` (Read, 1-125) — текущий движок (не меняется)
- `internal/core/lint_proto_info.go`, `proto_info_read.go` (Read) — `lintProtoInfoRead`, `openImportFile`-цепочка, blocked→ErrNotExist
- `internal/core/dom.go` (Read, 16-20, 42-64, 166-184) — `Position = protoast.Position`, `IssueInfo`/`Issue`, `OpenImportFileError{FileName}`
- `internal/core/core.go` (Read, 16-107) — `Core`-поля (importRoots/importFileAllowed/sourceFileOpen), `Options`, `New`
- `internal/api/lint.go` (Read, 80-154) — Action-маршрут exit 1/2, формат принтера
- `internal/api/lint_v1.go` (Read, 28-230) — `actionV1`: батчи, buildCore, сортировка issues по Path
- `internal/api/runtime.go` (Read, 36-88) — `errExit`, `buildCore` (roots wiring)
- `internal/api/lint_batch.go` (Read, 1-24) — `selectedLintWalker` (immutable → concurrency-safe Open)
- `internal/api/lint_v1_test.go` (Read) — прецедент Action-path тестов
- `internal/core/fs.go` (Read) — `FS`/`DirWalker` контракты
- `internal/core/position_parity_test.go` (grep) — Phase 1 руновый контракт (вкл. таб/эмодзи)
- Эталон: `/tmp/gsd-research/compiler-lab/easyp-cli/internal/core/link.go` (Read, 1-185) — классификатор, lintResolver, CompileRuleName
- Эталон: `.../internal/core/lint.go` (Read, 1-237) — precedence-механика, parseOnTheirOwn
- Эталон: `.../internal/core/lint_compile_test.go` (Read, 1-43) — матрица-эталон
- `.spec/ERRORS.md` (Read) — exit-коды, OpenImportFileError-маршрут, «Action needed for exit codes»
- `.planning/codebase/CONCERNS.md` (grep, 106-142) — exit-code test gap

### Secondary (MEDIUM confidence)
- buf.build docs (WebSearch corroboration): buf lint различает build errors и lint violations, non-zero exit на обеих; точный контракт easyp не наследует — COMPILE-in-existing-format закреплён REQ-5.1 как проектное решение. [CITED: buf.build/docs/cli/quickstart, buf.build/docs/breaking/lint/] — confidence seam: websearch = LOW/MEDIUM, используется только как корроборация, не как основание решений

### Tertiary (LOW confidence)
- none

## Metadata

**Confidence breakdown:**
- Классификация/стек: HIGH — эталон прочитан строка-в-строку; весь задействованный API vendored-дерева прочитан в этом сеансе
- Архитектура: HIGH — размещение/контракт закреплены D-04 + эталоном; direction of dependencies не меняется
- Pitfalls: HIGH — P1-P8 выведены из прочитанного кода (driver.go/lexer/compile.go), не из памяти; buf-корроборация — единственный web-источник и не несёт решений

**Research date:** 2026-11-19
**Valid until:** 2026-12-19 (стабильная область: sealed vendored compiler + локальный эталон; риск дрейфа — только синк compiler-lab, который в фазе не планируется)
