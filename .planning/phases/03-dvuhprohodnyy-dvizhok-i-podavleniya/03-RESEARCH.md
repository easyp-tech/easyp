# Phase 3: Двухпроходный движок и подавления - Research

**Researched:** 2026-11-19
**Domain:** Go lint engine internals — two-pass parallel engine over linked `easypc` syntax trees, `ModuleRule` lifecycle, comment-suppression parity, `lint.use` validation parity (EasyP `internal/core`)
**Confidence:** HIGH (all load-bearing facts read from source this session; the etalon clone is the verified porting source)

## User Constraints (from CONTEXT.md)

### Locked Decisions

**Детерминизм вывода**
- **D-01:** Полный детерминированный порядок выдачи замечаний — тотальная сортировка по (path, rule, line, col) на выходе движка. Main недетерминирован (map iteration); новый движок стабилен по определению — это усиление, не паритет (порядок main не является контрактом; проверено research Phase 1). — **Reversibility:** reversible — одна точка сортировки на выходе.

**Параллелизм**
- **D-02:** Параллельные окна через `internal/parallel` (bounded atomic-index scheduling), размер окна и параллелизм — как в эталоне лаборатории (256 файлов на окно, NumCPU воркеров); per-file правила параллельны, `ModuleRule.Finish` — последовательно после сбора всех окон. Порог «маленький корпус — без горутин» из эталона сохраняется.

**ModuleRule**
- **D-03:** Контракт `ModuleRule` (Collect/Finish) переносится из эталона в `internal/core` для кросс-файловых правил (PACKAGE_* семейство); обычные правила остаются per-file. Список ModuleRule-правил фиксируется на этапе движка (из 43 текущих правил — те, что в лаборатории ModuleRule); сами правила переносятся в Phase 4.

**Подавления**
- **D-04:** Центральный фильтр подавлений в одной точке движка: ignore/ignore_only конфига + комментарийные подавления (`allowCommentIgnores`) с семантикой текущего main (REQ-3.2: main выигрывает у лаборатории). Комментарийный walk переписывается на lint-модель чтения (protoast-деревья из слинкованных файлов) вместо reflection по go-protoparser Meta; паритет доказывается таблицей тестов main-vs-new на фикстурах с комментариями.
- **D-05:** `shouldIgnore` остаётся методом Core с проверкой ignore/ignore_only как в main; комментарийные подавления собираются per-file в первом проходе и применяются во втором — двухпроходная схема эталона.

**Валидация конфигурации**
- **D-06:** Неизвестное имя правила или группы в `lint.use` → ошибка валидации с паритетом main (текст/место ошибки — как в main); валидация живёт на уровне `internal/core` при конструировании движка, множество имён — из реестра правил builder'а.

**Модель чтения движка**
- **D-07:** Движок строит per-file модель из `Linked.Syntax` (protoast-деревья, подготовленные Phase 2) один раз на файл; lint-типы чтения из расщепления Phase 1 (D-06 Phase 1) наполняются из деревьев — без повторного чтения диска для файлов корпуса; `parseOnTheirOwn`-путь эталона для файлов, откатившихся на пофайловый разбор.

### Claude's Discretion
- Точные имена/границы файлов движка (lint engine rewrite рядом с текущим lint.go, старый остаётся до Phase 5).
- Внутренняя структура checkedFile/результатов окон.
- Формат таблицы паритет-тестов подавлений.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| LINT-ENG-01 (REQ-2.1) | Модель каждого файла из слинкованных деревьев `easypc`, все применимые правила, без изменения порядка обхода файлов | Etalon two-pass engine (`lint.go` 22–143): parallel read pass → `linkSources` → windowed model build (`protoInfoFromLinked`) → indexed sequential drain in walk order. Determinism via D-01 sort; race safety via stateless per-file rules + serial Collect/Finish |
| LINT-ENG-05 (REQ-2.5) | Подавления идентичны main: `ignore`/`ignore_only` + комментарии (`allowCommentIgnores`) | `shouldIgnore` byte-identical main/etalon [VERIFIED: main internal/core/lint.go:79-93; etalon lint.go:196-210]. Suppression skeleton (`parseDirectivePrefix`, pairing stack machine, error texts) byte-identical [VERIFIED: main comment_suppressions.go:55-143; etalon comment_suppressions.go:54-155]. Only the comment *walk* changes: protoast `file.Comments` + `file.Owner` replaces reflection over go-protoparser Meta. Parity fixtures already exist byte-identical in both trees (`comment_suppressions_test.go`, `comment_suppressions_eof_test.go`) |
| LINT-ENG-06 (REQ-2.6) | Неизвестное имя правила/группы в `lint.use` → ошибка валидации как до изменения | Main already validates in `rules.New` → `validateConfigRuleNames` → `ValidateNames` → `core.ErrInvalidRule` [VERIFIED: internal/rules/selection.go:14-52, builder.go:61-64,131-137]. Parity = keep this path untouched; engine receives pre-validated rules. Engine-side name knowledge (`knownLintRules` from `rules.AllRuleNames()`, wired in buildCore [VERIFIED: internal/api/runtime.go:74]) feeds comment-directive validation |

## Summary

Phase 3 ports the compiler-lab lint engine (a verified, tested implementation at pinned SHA `08cef10`) into `internal/core` as a **dormant** second engine next to the current sequential one. The etalon engine is 236 lines in `easyp-cli/internal/core/lint.go`: pass 1 reads every corpus file in parallel, links the corpus once via `linkSources` (already present in main from Phase 2, byte-verified), then processes files in windows of 256 — model building and per-file rule checking run concurrently through `internal/parallel.Run`, while the per-window drain, `ModuleRule.Collect`, and `ModuleRule.Finish` run sequentially in walk order. Cross-file rules (in the lab: 10 of the current rules + `PACKAGE_NO_IMPORT_CYCLE`) implement `Collect(ProtoInfo)/Finish() []IssueInfo`; per-file rules must be stateless. Suppressions are applied centrally: `shouldIgnore` (config `ignore_only`) inside `checkFile`, comment-suppression ranges computed per file and applied to per-file findings and again to `Finish` findings via the `suppressed` map.

The central engineering problem is **coexistence, not porting difficulty**: main's `core.ProtoInfo` (go-protoparser-shaped: `Info *unordered.Proto`, `ProtoFilesFromImport`) and `core.Rule` interface (`Validate(ProtoInfo)`) are taken by the 43 un-ported rules and the old engine, which must stay green until Phase 5. The etalon's engine is written against a protoast-shaped `ProtoInfo` (`File *protoast.File`, `ImportedFiles`, `Resolved *Resolution`) and a `ModuleRule` extension. Both cannot share names in one package. The research recommendation: introduce the new engine in new files with distinct type/interface names (dual model across command boundaries is the milestone's accepted temporary cost per prior ARCHITECTURE research; a bridge adapter is a documented anti-pattern). Phase 4 ports rules onto the new contract; Phase 5 flips the CLI and deletes the old path, at which point names can consolidate.

The suppression work is narrower than it looks: the directive parser, pairing machine, and error messages are **byte-identical** between main and the etalon; main additionally pre-collects comments via a reflection walk (`collectDirectiveComments`) fed through `readProtoFileWithDirectives`. The Phase 3 delta is exactly the walk: replace reflection with `directiveComments(file *protoast.File)` using `file.Comments` + `file.Owner(comment)` → `owner.Pos.Line..max(owner.Pos.Line, owner.End.Line)`. Parity is proven by re-running main's own suppression fixtures (byte-identical between the two repos) against the new engine — a main-vs-new table per CONTEXT D-04.

**Primary recommendation:** Port the etalon engine into new `internal/core` files under renamed types (engine entrypoint, model, rule interface, `ModuleRule`, `linkedFiles` adapter over main's `*protoc.Linked`, protoast-based suppression walk), carry the suppression/ignore skeletons verbatim, add the D-01 total sort as the single new behavior, and prove parity with a main-vs-new fixture table (suppressions + representative rules) plus a `-race` windowed determinism test. Nothing wires into `easyp lint`.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Model build from linked trees | `internal/core` (new engine files) | `internal/protoast` (Build/OutlineOf), `internal/easypc/protoc` (Linked, read-only) | Engine owns model lifecycle; protoast is the view seam; compiler trees consumed, never mutated (except dropping `Syntax` refs after build, as the etalon does) |
| Parallel windowed checking | `internal/core` engine | `internal/parallel` (generic runner) | Window policy is engine logic; `parallel.Run` is a reusable primitive, reused as-is (D-02) |
| Cross-file rule lifecycle (Collect/Finish) | `internal/core` (contract) | `internal/rules` (implementations, Phase 4) | Engine guarantees the phase boundary; rules only implement the interface |
| Comment suppression walk | `internal/core` (new protoast walk) | `internal/protoast` (`Comments`, `Owner`, `Decl.Pos/End`) | Engine owns semantics (D-04); protoast provides attachment facts. protoast is the maintained, non-sealed adaptation seam [VERIFIED: internal/easypc/VENDORED.md] |
| Config suppressions (ignore/ignore_only) | `internal/core` (`shouldIgnore`, existing) | — | Byte-identical in main/etalon; reused, not reimplemented (D-05) |
| `lint.use` name validation | `internal/rules` registry (existing) | `internal/core` (engine construction boundary) | Main's validation in `rules.New` is the parity source (D-06); engine consumes validated rule lists + `knownLintRules` |
| Deterministic output order | `internal/core` engine output | — | D-01: one sort point at engine exit |
| CLI wiring | — | — | Out of scope: dormant engine; flip is Phase 5 (D-04 dormancy continuation) |

## Project Constraints (from AGENTS.md + project skills)

- **Sealed tree:** no edits inside `internal/easypc` (compiler fixes go upstream; `internal/protoast` is the maintained adaptation seam and MAY be adjusted if parity requires).
- **Colocated tests:** `internal/rules/<rule>.go` + `<rule>_test.go`; use `testify`; `-race` mandatory on behavior tests.
- **Error style:** wrap with callee-only name + `%w`; `errors.Is`/`errors.As` for identity; deferred closes must handle errors.
- **Godoc:** English comments; exported symbols documented; enums reserve zero.
- **Schema rule:** no schema changes in this phase (rule names unchanged until Phase 6) → `task schema:check` should stay clean; not expected to run unless schemas touched.
- **No parallel agent instruction files; generated schema JSON never hand-edited.**
- **Testing (go-testing skill):** named table cases, `t.Parallel()` with private dependencies per case, sequential/subprocess for process-state mutation, handwritten doubles fine (mockery optional and not a prerequisite).
- **Git workflow (user global):** feature branch from up-to-date default branch; GSD config uses `branching_strategy: none` + `allow_default_branch_commits: true` (Phase 2 ratified committing on the current branch — same mode continues).

## Standard Stack

### Core

All stack items are **in-repo Go packages** — no new external dependencies (REQ-1.2 already enforced; nothing to install).

| Component | Version/Location | Purpose | Why Standard |
|-----------|------------------|---------|--------------|
| Go toolchain | go.mod `go 1.26.6`, installed `go1.27.3` | Language | Pinned by go.mod [VERIFIED: go.mod:3] |
| `internal/parallel` | main, 59 lines | Bounded atomic-index scheduling; index-ordered errors | Byte-identical to etalon; D-02 mandates reuse [VERIFIED: internal/parallel/parallel.go:16-59] |
| `internal/protoast` | main (vendored Phase 1) | Lint view: `Parse`, `Build(syntax.File)`, `OutlineOf`, `File.Comments`, `File.Owner`, `Decl.Pos/End` | The model for the new engine (D-07); non-sealed adaptation seam |
| `internal/easypc/protoc` | main (vendored Phase 1) | `protoc.Link`, `Linked.File(path)`, `LinkedFile.Syntax/Imports/Uses/Methods` | Phase 2 linker feeding the engine |
| `internal/core` link path | main `link.go` (Phase 2) | `linkSources` three-outcome classifier | Phase 2 deliverable, byte-preserved; new engine is its consumer |
| testify | go.mod | Test assertions | Project convention |

### Supporting

| Component | Purpose | When to Use |
|-----------|---------|-------------|
| `internal/fs/fs.FSWalker` (`disk.NewFSWalker(root, path)`) | DirWalker for engine tests | Engine-level tests, etalon pattern [VERIFIED: etalon comment_suppressions_test.go:29] |
| `internal/core/path_helpers.IsIgnoredPath` | `ignore` config filtering during walk | Reused in the new engine's walk (both engines) [VERIFIED: internal/core/path_helpers/is_target_path.go:21] |
| `internal/logger` (`logger.NewNop()`) | Engine logging | Tests |
| `wellknownimports` | Embedded WKT fallback in `openImportFile` | Already wired in shared helpers; unchanged |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Renamed new types (dual model) | Bridge adapter protoast→`unordered.Proto` | REJECTED by prior research (Anti-Pattern 1): the old contract is direct field access; a fake-tree adapter is a second parser kept only to be deleted |
| New engine in new subpackage (`internal/core/lintv2`) | In-package new files | Subpackage would need exported seams for `shouldIgnore`, `openImportFile`, `sourceImportPath`, `knownLintRules`, logger; the etalon keeps everything in package `core`. In-package files with renamed identifiers is the lower-friction port; final naming consolidation happens at Phase 5 deletion |
| Shared refactored suppression skeleton (parameterized comment accessor) | Duplicated skeleton in new files | Refactor risks old-path behavior; duplication is ~90 lines, deleted in Phase 5 cleanup. Recommend duplication with renamed identifiers + byte-identical error strings |
| Final sort in engine (D-01) | Sort in api printer | Engine output is the contract surface (Phase 5 swaps engines); sorting at engine exit keeps the printer untouched |

**Installation:** none — `go.mod` unchanged.

## Package Legitimacy Audit

Not applicable: this phase installs **zero external packages** (pure in-repo Go work; REQ-1.2 prohibits new external dependencies for the compiler path). No registry lookups required; no `checkpoint:human-verify` installs needed.

## Architecture Patterns

### System Architecture Diagram

```
                easyp lint (CLI)  ── UNCHANGED ──> old engine c.Lint (main lint.go)
                                                        │ sequential, go-protoparser ProtoInfo
                                                        └─ stays the user path until Phase 5

 NEW (Phase 3, dormant — constructed in tests only):

  paths (WalkDir, ignore filter, .proto)          [walk order preserved — LINT-ENG-01]
     │
     ▼
  ┌─ Pass 1: parallel.Run over ALL paths ─────────────────────────┐
  │  readSource per file → sources[], readErrs[] (indexed)        │
  └───────────────────────────────────────────────────────────────┘
     │ no read errors?
     ▼
  c.linkSources(ctx, disk, paths, sources)        [EXISTING Phase 2 code]
     ├── linked != nil ────────> wrap in linkedFiles{linked, outlines}
     │                             (outline(path): sync.Once + mutex,
     │                              protoast.OutlineOf(file.Syntax))
     ├── (nil, COMPILE issues) ─> return COMPILE issues, rules skipped
     └── (nil, nil) ───────────> parseOnTheirOwn: protoast.Parse per file,
                                   cacheLintedOutline; syntax errors →
                                   per-file command error (no COMPILE)
     ▼
  ┌─ Pass 2: windows of 256 files ────────────────────────────────┐
  │ parallel.Run(window):                                         │
  │   model = protoast.Build(linked file.Syntax); drop tree;      │
  │   ImportedFiles = outlines of its imports; Resolved derived   │
  │   checkFile: commentSuppressions(model) + per-file rules      │
  │              (ModuleRules skipped; shouldIgnore applied)      │
  │ sequential drain (walk order):                                │
  │   error? return (first error wins) ; else Collect(model) →    │
  │   ModuleRules ; suppressed[path] = ranges ; append findings   │
  └───────────────────────────────────────────────────────────────┘
     ▼
  ModuleRule.Finish() per rule (serial, rules order):
     shouldIgnore(issue.Path) + suppressed[issue.Path] filter → append
     ▼
  D-01 total sort by (path, rule, line, col)      [NEW vs etalon]
     ▼
  []IssueInfo  (dormant — no CLI consumer until Phase 5)
```

### Recommended Project Structure

Exact names are planner's discretion (CONTEXT Discretion item 1); the collision-free sketch below is the research recommendation:

```
internal/core/
├── lint.go                    # OLD engine — untouched until Phase 5
├── lint_linked.go             # NEW: two-pass engine (etalon Lint body), entrypoint renamed
├── lint_model.go              # NEW: protoast-shaped model struct + Resolution
│                              #      + new rule interface + ModuleRule (etalon dom.go pieces)
├── lint_linked_read.go        # NEW: readSource / protoInfoFromSource /
│                              #      protoInfoFromLinked / importCacheKey /
│                              #      cacheLintedOutline / parseOnTheirOwn (etalon proto_info_read.go pieces)
├── lint_suppressions.go       # NEW: protoast-based directiveComments walk + renamed skeleton
│                              #      (parsedDirective/suppressionRange reuse OK if types stay shared)
├── link.go                    # EXISTING Phase 2 linkSources — gains linkedFiles adapter
│                              #      + resolution/usesThrough (etalon link.go pieces) in same or new file
├── comment_suppressions.go    # OLD suppression path — untouched until Phase 5
├── core.go                    # += importCache *parsedImports field (+ parsedImports type)
└── dom.go                     # OLD Rule/ProtoInfo/AppendIssue — untouched
```

### Pattern 1: Two-pass windowed engine (port verbatim from etalon)

**What:** Parallel read pass over all files; one corpus `Link`; windows of 256 for model build + per-file checking; serial drain and reduce.
**When to use:** Always — D-02 locks this shape. Constants: `lintWindow = 256` [VERIFIED: etalon lint.go:18-20]; `parallel.Run(ctx, count, 0 /* → NumCPU */)` [VERIFIED: etalon lint.go:52,88; internal/parallel/parallel.go:23-26].

**Etalon source of truth (236 lines, read in full this session):** `/tmp/gsd-research/compiler-lab/easyp-cli/internal/core/lint.go` (fallback: `gh repo clone easyp-tech/compiler-lab` @ `08cef10`, directory `easyp-cli`).

Key mechanics that must survive the port:
- Results reassembled **by window index** in the serial drain — never in completion order [VERIFIED: etalon lint.go:108-121].
- `sources[start+i] = nil` and `infos[i], checked[i] = ProtoInfo{}, checkedFile{}` release memory per window [VERIFIED: etalon lint.go:99,120].
- `protoInfoFromLinked` builds the model then nils `file.Syntax`, keeping only the outline for importers [VERIFIED: etalon link.go:160-167].
- First per-file error in **walk order** aborts the run: `return nil, fmt.Errorf("fs.WalkDir: c.protoInfoRead: %w", infoErrs[i])` [VERIFIED: etalon lint.go:109-114]. NOTE: main's old engine wraps the same failure as `fs.WalkDir: c.lintProtoInfoRead: ...` [VERIFIED: main internal/core/lint.go:34-36]. The wrapper text differs; it is user-visible (api lint default branch returns the error to the CLI printer [VERIFIED: internal/api/lint.go:78-95]). Decide deliberately (main-wins text recommended per REQ-3.2 philosophy); flag in plan.
- COMPILE issues are returned **before** suppression machinery runs — no suppression filtering on COMPILE (etalon lint.go:71-76). Phase 2 deferred note IN-02 ("COMPILE diagnostics vs shouldIgnore") stays deferred to Phase 5 [VERIFIED: .planning/STATE.md Blockers].

### Pattern 2: ModuleRule map-reduce lifecycle

**What:** Cross-file rules implement `Collect(ProtoInfo) error` + `Finish() []IssueInfo`; the engine skips them in `checkFile`, calls `Collect` serially in walk order during the drain, `Finish` once after all windows.

**Etalon contract** [VERIFIED: etalon dom.go:25-33]:
```go
// ModuleRule needs every file before it can report, such as rules that
// compare files of one package or directory. Lint passes each file to
// Collect and asks for findings once, after the last file; it does not
// call Validate. Suppressions and ignore_only apply to the findings by path.
ModuleRule interface {
    Rule
    Collect(ProtoInfo) error
    Finish() []IssueInfo
}
```
- `Collect` runs **serially** (inside the drain loop, etalon lint.go:115-117) — rules need no synchronization; only model *building* is parallel. This resolves the prior research's Pattern-4 concern about concurrent Collect.
- Finish findings are re-filtered: `shouldIgnore(rule, issue.Path)` then `suppressed[issue.Path].contains(rule, line)` [VERIFIED: etalon lint.go:124-138].
- ModuleRules still implement `Validate` as a no-op stub to satisfy the interface (etalon rules pattern: `func (p *PackageSameJavaPackage) Validate(core.ProtoInfo) ([]core.Issue, error) { return nil, nil }` [VERIFIED: etalon package_same_java_package.go:21]).

**ModuleRule list to fix at engine level (D-03)** — lab assertions [VERIFIED: grep of `var _ core.ModuleRule` over etalon rules]:
Current-main rules that are ModuleRules in the lab (10): `DirectorySamePackage`, `PackageSameDirectory`, `PackageSameCsharpNamespace`, `PackageSameGoPackage`, `PackageSameJavaMultipleFiles`, `PackageSameJavaPackage`, `PackageSamePHPNamespace`, `PackageSameRubyPackage`, `PackageSameSwiftPrefix`, `RPCRequestResponseUnique`. (`PackageNoImportCycle` is the 11th, arrives Phase 6.)
In main today, 9 of these 10 carry per-file mutable state (`cache map` + `lazyInit`) inside `Validate` [VERIFIED: grep of main internal/rules] — that statefulness is exactly what the ModuleRule split removes in Phase 4. The stateless per-file rule shape for the new engine: `type ImportUsed struct{}` with per-file maps in a local struct value [VERIFIED: etalon import_used.go:13-33].

### Pattern 3: Central suppression filter (D-04/D-05)

**What:** `shouldIgnore` (config) checked per rule per file; comment-suppression ranges computed once per file in the model pass, applied to that file's findings and, via the `suppressed` map, to Finish findings.

**What changes vs main — only the comment *walk*.** The rest is byte-identical:

| Piece | Main (old path) | Etalon (new engine) | Delta |
|-------|-----------------|---------------------|-------|
| Directive lexing | `parseDirectivePrefix` — prefixes `easyp:disable`, `easyp:enable`, `nolint:`, `buf:lint:ignore` (→ "legacy"), bare `easyp:` recognized-and-ignored; ` -- ` / ` // ` reason stripping | identical, byte-for-byte | none — carry verbatim |
| Name compaction | `compactDirectiveNames` | identical | none |
| Pairing machine | disable-stack / enable-pop / unpaired-attached ranges; identical error strings: `malformed lint directive %q`, `unknown or unsupported lint rule %q in comment directive`, `%s:%d: easyp:enable %s has no preceding disable`, `%s:%d: suppression %s must annotate a declaration or have a matching easyp:enable` | identical | none — error-text parity is automatic if carried verbatim |
| Comment collection | reflection walk `collectDirectiveComments(root any)` over go-protoparser ordered AST; attaches `Comments`/`InlineComment`/`InlineCommentBehindLeftCurly` to decl `Meta.Pos..max(Pos,LastPos)`; sorted by (line, col); input pre-collected by `readProtoFileWithDirectives` (`WithBodyIncludingComments(true)`, standalone comments stripped before unordered interpretation) | `directiveComments(file *protoast.File)`: iterate `file.Comments` (source order), owner via `file.Owner(comment)`; range `owner.Pos.Line..max(owner.Pos.Line, owner.End.Line)`; unowned comments get `first = 0` | **the porting target** |
| Range containment | `suppressionRanges.contains(rule, line)` | identical | none |

New-engine walk [VERIFIED: etalon comment_suppressions.go:26-46]:
```go
type directiveComment struct {
    comment     *protoast.Comment
    first, last int
}

func directiveComments(file *protoast.File) []directiveComment {
    if file == nil {
        return nil
    }
    comments := make([]directiveComment, 0, len(file.Comments))
    for _, comment := range file.Comments {
        d := directiveComment{comment: comment}
        if owner := file.Owner(comment); owner != nil {
            d.first, d.last = owner.Pos.Line, max(owner.Pos.Line, owner.End.Line)
        }
        comments = append(comments, d)
    }
    return comments
}
```
Supporting protoast facts [VERIFIED: internal/protoast/protoast.go:40-55,63-70,89-105]: `Comment{Raw, Pos, End}` + `Lines()` strips markers and splits on `\n`; `Decl{Pos, End, Leading, Detached, Trailing}`; `File.Comments` is "every comment in source order"; `Owner` returns the declaration a comment annotates (preceded or trailing), **nil for comments that annotate nothing (EOF, before closing brace)**.

**Parity risk to prove, not assume:** attachment semantics of the reflection walk vs protoast owner map are *different algorithms* (prior research Pitfall 2). Both handle the EOF/orphan cases (main retains EOF comments via `WithBodyIncludingComments`; protoast keeps them as unowned → `first = 0` → unpaired `easyp:disable` errors "must annotate a declaration", matching the etalon EOF fixture matrix). The proof obligation is the main-vs-new fixture table (D-04) — main's suppression tests are byte-identical to the etalon's [VERIFIED: diff of comment_suppressions_test.go and comment_suppressions_eof_test.go] and already encode the matrix: EOF with/without newline, block comment at EOF, orphan enable, unknown trailing rule, unattached disable, unpaired attached disable, marker-in-string, nested closing brace [VERIFIED: main internal/core/comment_suppressions_eof_test.go:19-33].

**Position convention note:** protoast positions are 1-based line, 1-based **rune** column, tab advances by 1 [VERIFIED: internal/protoast/protoast.go:5-10 doc comment]; go-protoparser Meta columns are byte-oriented — directive *validation error* messages embed comment line:col and may differ on tab/UTF-8 files. Lines (the suppression matching key) are unaffected. Column-level divergence in directive error text is a Phase 5 observable; pin expectations in the parity table if fixtures include tabs (Phase 2 already pinned a diverging tab/UTF-8 fixture for COMPILE positions [VERIFIED: STATE.md Phase 2 notes]).

### Pattern 4: linkedFiles adapter over main's `*protoc.Linked`

Main's Phase-2 `linkSources` returns the raw `*protoc.Linked` [VERIFIED: internal/core/link.go:32]; the etalon wraps it in `linkedFiles` (mutex + per-path `sync.Once` outlines) [VERIFIED: etalon link.go:129-156]:
```go
type linkedFiles struct {
    linked   *protoc.Linked
    mu       sync.Mutex
    outlines map[string]*linkedOutline
}
type linkedOutline struct {
    once sync.Once
    file *protoast.File
}
func (l *linkedFiles) outline(path string) (*protoast.File, bool) {
    file, ok := l.linked.File(path)
    if !ok { return nil, false }
    l.mu.Lock()
    entry, ok := l.outlines[path]
    if !ok { entry = &linkedOutline{}; l.outlines[path] = entry }
    l.mu.Unlock()
    entry.once.Do(func() { entry.file = protoast.OutlineOf(file.Syntax) })
    return entry.file, true
}
```
Port these etalon link.go pieces: `linkedFiles`/`linkedOutline`/`outline`, `protoInfoFromLinked`, `resolution`, `usesThrough` (resolution facts for rules: `UsedImports`, `RPCTypes` keyed by rpc byte offset). The `Resolution` type must come with the new model [VERIFIED: etalon dom.go:49-57]. Do NOT port the etalon's `linkSources` — main's Phase-2 version wins (syntax-error precedence `hasPerFileSyntaxError`, deterministic COMPILE sort, diagCollector) [VERIFIED: main link.go:32-112]. The engine wraps main's result at the call site.
Main's `protoc.LinkedFile` already exposes everything the adapter needs: `Syntax *syntax.File`, `Imports []LinkedImport{Path, Public}`, `Uses map[string]bool`, `Methods []LinkedMethod{Offset, Input, Output}`, `Linked.File(path) (*LinkedFile, bool)` [VERIFIED: internal/easypc/protoc/link.go:17-52; byte-identical to etalon via diff]. The `Location` (quoted literal) used to key `ImportedFiles` comes from **protoast** `Import.Location`, not protoc [VERIFIED: etalon link.go:169-176; internal/protoast/protoast.go:112-118].

### Pattern 5: Fallback path — `parseOnTheirOwn` (D-07)

When linking can't classify (missing import / non-compilation failure / no diagnostics), the etalon parses each corpus file standalone with `protoast.Parse`, records syntax errors per file, and registers outlines for linted files via `cacheLintedOutline` (keyed `RootPath() + "\x00" + importName`, guarded by `filepath.IsLocal` in `openImportFile`) [VERIFIED: etalon lint.go:226-236; etalon proto_info_read.go:71-126]. Main's equivalent primitives: `protoast.Parse` exists; `cacheLintedOutline`/`importCacheKey`/`readSource`/`protoInfoFromSource` (protoast variant) do **not** exist in main and must be ported. Shared helpers `openImportFile`, `sourceImportPath`, `openDiskSource` are byte-identical main/etalon [VERIFIED: function-level diff = IDENTICAL] and are reused as-is.

**Requires a Core state addition:** `importCache *parsedImports` (mutex + `map[string]*protoast.File`), lazily initialized inside the new engine entry [VERIFIED: etalon core.go:46-54; etalon lint.go:44-46]. Main's `Core` lacks it [VERIFIED: internal/core/core.go:16-42].

### Anti-Patterns to Avoid

(from prior ARCHITECTURE research, re-validated against this phase)
- **Bridging old interface with a converted AST:** never fake `*unordered.Proto` from protoast so old rules run on the new engine.
- **Suppression logic in rules:** the engine filters centrally; rules never check `ignore`/directives themselves (`CheckIsIgnored` in main has no production callers [VERIFIED: grep over internal/ non-test — definition only]).
- **Completion-order output:** findings must be reassembled by window index, then D-01-sorted.
- **Stateful per-file rules on the parallel path:** any rule the new engine runs concurrently must keep zero receiver-field writes in `Validate` (ROADMAP success criterion 4 — statelessness audit inside the engine race test).
- **Mutating the old path "while we're here":** old `lint.go`/`dom.go`/`comment_suppressions.go`/`proto_info_read.go`/`link.go` behavior stays frozen until Phase 5; Phase 2 review notes WR-01/WR-02/IN-01/IN-02 are explicitly deferred to the Phase 5 flip plan.
- **Rules importing easypc:** rules see only protoast; missing view capabilities go into protoast (non-seamed).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bounded parallel task execution | Custom worker pools/channels | `internal/parallel.Run` | Index-owned tasks, index-ordered errors, ctx cancellation — already proven under `-race`; D-02 mandates it |
| Comment→declaration attachment | Heuristic line scanning of source text | `protoast` `File.Comments` + `File.Owner` | The builder already computes owner maps (Leading/Detached/Trailing semantics); hand-rolled attachment re-creates Pitfall 2 |
| Directive lexing/pairing | New parser | Carry main's skeleton verbatim | Byte-identical main/etalon; error-text parity is free |
| Import resolution during fallback | New import lookup | `openImportFile` + `wellknownimports` + `importRoots` (existing, shared) | Policy logic (allowed-files gate) must not diverge — prior research Pitfall 5 |
| Outline caching | Ad-hoc maps | `parsedImports` mutex-map + `sync.Once` outlines (etalon shape) | Concurrent window workers hit the same imports |
| Rule name derivation | String lists | `core.GetRuleName` + `rules.AllRuleNames()` | Single registry feeds validation, schemas, and `knownLintRules` |

**Key insight:** nearly everything except the comment walk already exists in main — the phase is a disciplined port plus a naming plan, not an engine design exercise.

## Runtime State Inventory

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — engine holds no persistent state; all state is per-run in memory | None |
| Live service config | None — no external services involved | None |
| OS-registered state | None | None |
| Secrets/env vars | None touched | None |
| Build artifacts | None new; `bin/golangci-lint` pre-existing; no generated artifacts affected (no schema changes) | None |

*(Verified by inspection this session: the phase adds dormant in-repo code only.)*

## Common Pitfalls

### Pitfall 1: Type/interface name collisions between old and new engine (the phase's central trap)
**What goes wrong:** Porting etalon files verbatim fails to compile: `Lint` (method), `ProtoInfo`, `Rule`, `commentSuppressions` (method), `directiveComment`, `parsedDirective`, `readFilesFromImport` all already exist in package `core` with incompatible shapes (go-protoparser vs protoast).
**Why it happens:** The etalon is the *post-flip* tree where old names are gone; main is the *pre-flip* tree where they are live.
**How to avoid:** New files + renamed identifiers for everything the etalon shares by name with main's old path; keep shared-by-value pieces (`Issue`, `IssueInfo`, `Position` alias, `ImportPath`, `PackageName`, `GetRuleName`, `buildError`, `shouldIgnore`, `close`, suppression skeletons) common. Log the rename map in the plan so Phase 4/5 can consolidate.
**Warning signs:** compile errors on port; worse, silently *changing* an old-path signature "to make it fit" (breaks dormancy).

### Pitfall 2: Suppression attachment divergence (prior research Pitfall 2, now concrete)
**What goes wrong:** Issues suppressed by main reappear (or valid files hard-error) because protoast's owner semantics differ from the reflection walk on: blank-line-separated (detached) comments, trailing comments after closing braces, EOF comments, `InlineCommentBehindLeftCurly` equivalents.
**Why it happens:** Different attachment algorithms; only the fixture table can prove equivalence.
**How to avoid:** Port main's suppression fixtures (byte-identical, ready) as the main-vs-new table; add fixtures for detached comments and trailing-brace comments if the table lacks them; run the table through BOTH engines in one test file for a visible diff.
**Warning signs:** corpus noise concentrated in doc-comment-heavy files; `must annotate a declaration` errors on files main accepted.

### Pitfall 3: Determinism regressions from parallelism
**What goes wrong:** Output order varies run-to-run (CI flakes, lint-diff noise).
**Why it happens:** Appending findings from workers in completion order; map iteration inside the drain.
**How to avoid:** Indexed reassembly in the serial drain (etalon mechanics) + D-01 total sort by (path, rule, line, col) at engine exit — the sort is the one *new* behavior vs the etalon (the etalon does not sort [VERIFIED: etalon lint.go:79-142 has no sort]); test byte-identical output across repeated `-race` runs (ROADMAP criterion 1).
**Warning signs:** two runs of the same fixture produce different orderings.

### Pitfall 4: Data races via stateful rules or shared model mutation
**What goes wrong:** `-race` failures in the windowed path.
**Why it happens:** Rules with receiver-field writes run concurrently; or `linkedFiles.outline`/`importCache` accessed without the etalon's mutex/Once discipline; or `ProtoInfo` values sharing mutable maps across files.
**How to avoid:** Race test drives the *parallel* path with multi-file fixtures and stateless test-double rules (unit rule tests alone never hit it); port the outline/cache synchronization verbatim; give each rule a fresh model.
**Warning signs:** races appearing only in engine-level tests (ROADMAP criterion 4 mandates this test).

### Pitfall 5: Dormancy leak
**What goes wrong:** The new engine (or its Core field additions) alters old-path behavior or CLI output; Phase-2/Phase-1 invariants break.
**Why it happens:** Adding fields to shared structs (`Core`, `dom.go` types) and touching shared helpers casually.
**How to avoid:** Additive-only changes to shared types (`importCache` field; new types); old lint path's files untouched except additive imports if unavoidable; full `go test -race -count=1 ./internal/...` plus `task lint` green at every intermediate state; behavior diff on the old path must be empty.
**Warning signs:** any diff in old-engine test outputs; `schema:check` drift (should be none — no rule names change).

### Pitfall 6: Error-text parity debt
**What goes wrong:** Phase 5 flip changes user-visible error strings on broken files (REQ-2.4 "как до изменения").
**Why it happens:** Etalon wraps per-file model errors as `fs.WalkDir: c.protoInfoRead: ...`; main wraps as `fs.WalkDir: c.lintProtoInfoRead: ...`; intermediate prefixes also differ (`Parse:`/`Open:`/`readProtoFileWithDirectives:` vs protoast's `Parse:`).
**How to avoid:** Decide the target text now (recommend main-wins), pin with a test asserting the exact wrapped chain for a read-error fixture and a syntax-error fixture.
**Warning signs:** golden/error-message test diffs deferred to Phase 5.

### Pitfall 7: Registry count discrepancy ("43 rules")
**What goes wrong:** D-03's "из 43 текущих правил" doesn't match the registry: main's `rules.New` constructs **41** rule instances (MINIMAL 4 + BASIC 20 + DEFAULT 8 + COMMENTS 7 + UNARY_RPC 2) [VERIFIED: internal/rules/builder.go:65-120]; 40 carry `var _ core.Rule` assertions [VERIFIED: grep]; REQUIREMENTS.md says "всех 43 текущих правил" [CITED: .planning/REQUIREMENTS.md:28].
**How to avoid:** Treat the builder registry as the source of truth for the ModuleRule list; note the discrepancy for the planner (harmless for the engine, matters for Phase 4 scope math and Phase 7 parity claims).
**Warning signs:** parity-scope confusion in Phase 4/7 planning.

## Code Examples

### Engine entrypoint shape (etalon, to port with renames)
```go
// Source: compiler-lab/easyp-cli/internal/core/lint.go @ 08cef10 (read in full)
const lintWindow = 256

func (c *Core) Lint(ctx context.Context, fsWalker DirWalker) ([]IssueInfo, error) {
    // walk → paths (IsIgnoredPath, .proto only)
    // Pass 1: parallel read → sources[], readErrs[]
    // link: linked, compileIssues = c.linkSources(ctx, fsWalker, paths, sources)
    //   linked == nil → parseOnTheirOwn; COMPILE issues returned if no syntax errors
    // Pass 2: windows of 256:
    //   parallel: protoInfoFromLinked | protoInfoFromSource; checkFile
    //   serial drain (walk order): error check; collectFile → ModuleRules;
    //                             suppressed[path] = ranges; append findings
    // ModuleRule.Finish() loop with shouldIgnore + suppressed filter
    // D-01: sort by (path, rule, line, col)   ← new vs etalon
}
```

### Per-file checking with central suppression (etalon)
```go
// Source: compiler-lab/easyp-cli/internal/core/lint.go:155-180
func (c *Core) checkFile(ctx context.Context, path string, protoInfo ProtoInfo) checkedFile {
    suppressions, err := c.commentSuppressions(protoInfo)
    if err != nil {
        return checkedFile{err: fmt.Errorf("commentSuppressions: %w", err)}
    }
    var res []IssueInfo
    for _, rule := range c.rules {
        if ctx.Err() != nil {
            return checkedFile{err: ctx.Err()}
        }
        if _, ok := rule.(ModuleRule); ok || c.shouldIgnore(rule, path) {
            continue
        }
        results, err := rule.Validate(protoInfo)
        if err != nil {
            return checkedFile{err: fmt.Errorf("rule.Validate: %w", err)}
        }
        for _, result := range results {
            if c.allowCommentIgnores && suppressions.contains(result.RuleName, result.Position.Line) {
                continue
            }
            res = append(res, IssueInfo{Issue: result, Path: path})
        }
    }
    return checkedFile{issues: res, suppressions: suppressions}
}
```

### Test doubles for the engine (patterns ready to reuse)
```go
// Source: compiler-lab internal/rules/init_test.go:118-123 — model fixture
func parseProto(t testing.TB, path, src string) core.ProtoInfo {
    t.Helper()
    file, err := protoast.Parse(path, []byte(src))
    require.NoError(t, err)
    return core.ProtoInfo{Path: path, File: file}
}

// Source: compiler-lab internal/rules/module_rules_test.go:13-25 — lifecycle driver
func issuePaths(t *testing.T, rule core.Rule, infos ...core.ProtoInfo) []string {
    t.Helper()
    moduleRule, ok := rule.(core.ModuleRule)
    require.True(t, ok)
    for _, info := range infos {
        require.NoError(t, moduleRule.Collect(info))
    }
    var paths []string
    for _, issue := range moduleRule.Finish() {
        paths = append(paths, issue.Path)
    }
    return paths
}
```

### Engine-level compile test (etalon — proves COMPILE routing through the new engine)
```go
// Source: compiler-lab internal/core/lint_compile_test.go:16-30
// in-proto3 `required` field → exactly one COMPILE issue; rules skipped.
app := core.New(core.Options{Logger: logger.NewNop(), Rules: []core.Rule{&rules.MessagePascalCase{}}, KnownLintRules: rules.AllRuleNames()})
got, err := app.Lint(t.Context(), disk.NewFSWalker(root, "."))
// require: RuleName == core.CompileRuleName, Line 4, Column 3, message "Required fields are not allowed in proto3."
// Missing import (missing.proto) → *core.OpenImportFileError (ErrorAs), NOT a COMPILE issue.
```

### lint.use validation parity (existing main code — keep untouched)
```go
// Source: internal/rules/selection.go:14-27 (main)
func ValidateNames(names []string) error {
    allowed := AllLintUseValues()
    var issues []error
    for _, name := range names {
        if name == "PACKAGE_NO_IMPORT_CYCLE" {
            issues = append(issues, fmt.Errorf("%w: %s is not implemented", core.ErrInvalidRule, name))
            continue
        }
        if !slices.Contains(allowed, name) {
            issues = append(issues, fmt.Errorf("%w: %s", core.ErrInvalidRule, name))
        }
    }
    return errors.Join(issues...)
}
// validateConfigRuleNames wraps per-selection: "use: ...", "except: ...", "ignore_only: ..."
// rules.New wraps: fmt.Errorf("validateConfigRuleNames: %w", err)   [VERIFIED: builder.go:62-64]
// Existing test: TestInvalidLintNamesAreRejectedBeforeFiltering [VERIFIED: selection_test.go:43]
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Stateful cross-file rules (`cache map` in `Validate`, sequential engine) | Stateless per-file rules + `ModuleRule` Collect/Finish | lab engine (this port) | 9 main rules lose receiver state in Phase 4; engine guarantees the reduce boundary |
| Reflection comment walk over go-protoparser Meta | protoast `Comments`/`Owner` walk | lab engine | Deletes reflection from the lint path at Phase 5; attachment semantics must be proven equal |
| Sequential per-file read+parse | One corpus Link + windowed model build | lab engine | Perf headroom (REQ-6.1 orientation); memory bounded by window + tree dropping |
| Unsorted, map-random within-rule order | Deterministic (path, rule, line, col) | D-01 (this phase, beyond etalon) | Visible, intended improvement; parity comparisons must set-compare or sort both sides |

**Deprecated/outdated:** none actionable this phase (go-protoparser removal is a future milestone; `CheckIsIgnored` is dead code that dies with the old path in Phase 5).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Dual-model coexistence with renamed new types (not a `ProtoInfo` shape change in place) is the intended dormancy-compatible shape; CONTEXT D-07's "lint-типы… наполняются из деревьев" is read as the *mechanism* (trees fill the model), not a commitment to reuse the `ProtoInfo` name. Prior ARCHITECTURE research independently recommends a lint-facing struct [CITED: .planning/research/ARCHITECTURE.md divergence table]. | Architecture Patterns / Recommended Structure | If the project actually wants the etalon's exact names now, Phase 3 must rename the old path's types first (touches 43 rule files) — much larger diff; detectable at plan review |
| A2 | "Репрезентативные правила" in the success criteria are satisfied by test-double rules through both engines (old engine + old model vs new engine + new model on same fixtures), not by pre-porting real rules | Validation Architecture | If real rules are expected (e.g., one ModuleRule + one naming rule minimally ported), Phase 3 grows into early Phase 4 work; planner should confirm scope at plan time |
| A3 | The 41-vs-43 rule-count discrepancy is a spec-document artifact, not hidden rules | Pitfalls 7 | Phase 4/7 scope math off by 2 if two rules exist outside the builder (none found) |
| A4 | Column values in comment-directive *validation errors* may diverge on tab/UTF-8 lines between engines (rune vs byte columns); lines (the matching key) never diverge | Pattern 3 | A Phase 5 observable; if a parity fixture includes tabs, expectations must be pinned explicitly |
| A5 | `parallel.Run`'s "small corpus, no goroutines" property referenced by D-02 maps to `workers := min(count, limit)` semantics (count ≤ NumCPU workers; `count<=0` → no-op) — no separate threshold branch exists in either copy | Pattern 1 | Cosmetic; no behavioral impact |
| A6 | COMPILE issues intentionally bypass suppression filtering (etalon behavior), and Phase 2's IN-02 question (should COMPILE respect shouldIgnore) stays deferred to Phase 5 | Pattern 1 | If the project wants COMPILE suppressible, add a decision to the Phase 5 flip plan, not here |

## Open Questions

1. **New model/rule-interface naming (ties to A1)**
   - What we know: `ProtoInfo`, `Rule`, `Lint`, `commentSuppressions`, `readFilesFromImport`, `directiveComment`, `parsedDirective` collide; old names stay live until Phase 5.
   - What's unclear: preferred new names (e.g., `LinkedFileInfo`/`LintRule`/`LintLinked`…) and whether Phase 5 will rename back to etalon names or keep the new ones.
   - Recommendation: planner picks names in the plan (CONTEXT grants discretion); keep a rename map comment in the new files for Phase 5 consolidation.

2. **Scope of "representative rules" for engine parity (ties to A2)**
   - What we know: ROADMAP criterion 1 requires the windowed engine to produce the same remarks as the sequential main-model pass "для репрезентативных правил"; D-03 says real rules arrive in Phase 4.
   - What's unclear: test doubles only, or a minimal real-rule mini-port as engine fixtures.
   - Recommendation: test doubles (old-model rule + new-model rule asserting equal (rule, file, line, col) sets) — keeps Phase 4 scope intact; a real ModuleRule collect/finish lifecycle is still exercised via a double implementing Collect/Finish.

3. **Per-file error wrapper text (Pitfall 6)**
   - What we know: the two engines wrap the same failure with different callee names; text is user-visible at Phase 5.
   - What's unclear: whether to match main's `c.lintProtoInfoRead` chain now or note divergence for Phase 5.
   - Recommendation: match main's observable text now (cheap, REQ-2.4 "как до изменения"), pin with exact-string tests.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | everything | ✓ | go1.27.3 (go.mod: 1.26.6) | — |
| `task` (Taskfile runner) | `task test`/`task lint`/`task quality` | ✗ not on PATH, not in bin/ | — | raw equivalents documented in AGENTS.md: `go test -race -count=1 ./internal/...`, `go build -o easyp ./cmd/easyp` |
| gotestsum | `task test` runner (precondition in Taskfile) | ✗ not in bin/ | — | plain `go test -race -count=1` |
| golangci-lint | `task lint:go` | ✓ bin/golangci-lint | pinned by task init | — |
| git | commits, branching | ✓ | — | — |
| etalon clone | porting source | ✓ /tmp/gsd-research/compiler-lab @ 08cef10 | — | `gh repo clone easyp-tech/compiler-lab` |
| testify | tests | ✓ (go.mod) | — | — |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** `task`, gotestsum — use raw `go` commands (documented in AGENTS.md); note `task test`'s gotestsum precondition would fail until `task init` runs, but `task` itself is absent, so raw commands are the operating mode.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go standard `testing` + testify (assert/require); runner gotestsum via `task test` (absent → raw go test) |
| Config file | none (convention-driven: colocated `*_test.go`, `-race` mandatory) |
| Quick run command | `go test -race -count=1 ./internal/core/...` |
| Full suite command | `go test -race -count=1 ./internal/...` (≈ `task test` minus coverage/gotestsum) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LINT-ENG-01 | Model built from linked trees; windowed parallel run yields deterministic output; walk order preserved | integration (engine, `-race`, repeated runs byte-identical) | `go test -race -count=1 ./internal/core/ -run Test<NewEngine>` | ❌ Wave 0 (new engine test file) |
| LINT-ENG-01 | Fallback path: missing import → per-file parse, `OpenImportFileError` surfaces; syntax error → command error, no COMPILE | integration | same, case-driven | ❌ Wave 0 |
| LINT-ENG-01 | COMPILE issues returned, rules skipped (Phase 2 contract through new engine) | integration | same | ❌ Wave 0 (etalon pattern: lint_compile_test.go) |
| LINT-ENG-05 | Comment suppressions: main-vs-new table over main's byte-identical fixtures (policies independence, exact names, EOF matrix) | parity table | `go test -race -count=1 ./internal/core/ -run '<Suppression>'` | ⚠️ old-path tests exist (comment_suppressions_test.go, comment_suppressions_eof_test.go); new-engine table ❌ Wave 0 |
| LINT-ENG-05 | Config suppressions: ignore (walk filter) + ignore_only (shouldIgnore incl. dir prefix) | unit/integration | same | ❌ Wave 0 (old path implicitly covered; new engine needs its own) |
| LINT-ENG-06 | Unknown rule/group in `lint.use` → `core.ErrInvalidRule` with main's text/place | unit (rules registry — existing) + engine-construction check | `go test -race -count=1 ./internal/rules/ -run TestInvalidLintNamesAreRejectedBeforeFiltering` | ✅ exists [VERIFIED: selection_test.go:43] |
| ROADMAP crit. 4 | Rule receivers stateless under parallel windows; engine race test | race test | full core suite under `-race` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test -race -count=1 ./internal/core/... ./internal/rules/...`
- **Per wave merge:** `go test -race -count=1 ./internal/...`
- **Phase gate:** full suite + `go build -o easyp ./cmd/easyp` + `go run ./cmd/easyp schema-gen`-based `task schema:check` equivalent stays drift-free (no schema changes expected); `golangci-lint run` (bin/) green — remember the Phase-1 lint exclusion for `internal/rules/lint_helpers.go` stays until helpers gain callers (Phase 4).

### Wave 0 Gaps
- [ ] New engine test file(s) in `internal/core` — engine determinism (`-race`, multi-run), COMPILE routing, fallback matrix, ModuleRule lifecycle via doubles
- [ ] New-engine suppression parity table (main fixtures re-driven through the new engine + ignore/ignore_only cases)
- [ ] Engine race/statelessness test (multi-file fixture, parallel windows, shared rule instances)
- [ ] Optional: `internal/core/testdata/lint/` fixture tree if file-based fixtures outgrow `t.TempDir()` strings (precedent: `testdata/link/` from Phase 2)

*(Existing infrastructure covers the rest: testify conventions, `-race` culture, Phase-2 link fixtures, suppression fixtures byte-identical to the etalon.)*

## Security Domain

`security_enforcement: true` (ASVS L1) — this phase is an internal CLI engine refactor; most ASVS categories are not applicable (no auth, sessions, network input, crypto).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | N/A — local CLI |
| V3 Session Management | no | N/A |
| V4 Access Control | yes (path-level) | Import/source access already gated by `importFileAllowed`/`filepath.IsLocal` in shared `openImportFile`/`openDiskSource` — new engine MUST reuse them, never open files directly [VERIFIED: proto_info_read.go:103-147,169-181] |
| V5 Input Validation | yes | Config names validated against registry (`ValidateNames`); comment directives validated against `knownLintRules` with error-on-unknown — carry byte-identical |
| V6 Cryptography | no | N/A |
| V14 Config | minimal | No new config surfaces; no schema changes |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Import resolution escaping the module's locked file selection (reads files outside dependency metadata) | Information Disclosure | Route all reads through `openImportFile` (respecting `importFileAllowed`) — prior research Pitfall 5; the new engine's fallback/cache path must not bypass it (Phase 2 WR-02 note stays deferred to Phase 5) |
| Path traversal via import names | Tampering/Elevation | `filepath.IsLocal` guard at `openImportFile` entry — reuse as-is |
| Directive/validation-bypass (malformed comment disables silently) | Tampering | Error-on-malformed/unknown directive (existing skeleton) carried verbatim |

## Sources

### Primary (HIGH confidence)
- Etalon (porting source, read in full): `/tmp/gsd-research/compiler-lab/easyp-cli/internal/core/lint.go` (236 lines), `link.go`, `dom.go`, `core.go`, `proto_info_read.go`, `comment_suppressions.go`, `comment_suppressions_test.go`, `comment_suppressions_eof_test.go`, `lint_compile_test.go`, `internal/rules/{package_same_java_package,import_used,module_rules_test,init_test,builder,selection}.go` @ `08cef10`
- Main repo (all Read this session, line ranges cited inline): `internal/core/{lint.go, link.go, dom.go, core.go, comment_suppressions.go, proto_comments.go, lint_proto_info.go, proto_info_read.go, fs.go}`, `internal/parallel/parallel.go`, `internal/rules/{builder.go, selection.go, lint_helpers.go}`, `internal/protoast/protoast.go`, `internal/easypc/protoc/link.go`, `internal/easypc/VENDORED.md`, `internal/api/{runtime.go, lint.go}`, `internal/core/path_helpers/is_target_path.go`, `internal/fs/fs/dir_walker.go`
- Planning: `.planning/ROADMAP.md` § Phase 3, `.planning/REQUIREMENTS.md`, `.planning/STATE.md`, `.spec/features/lint-easypc-infra/requirements.md`, `.planning/research/{ARCHITECTURE,PITFALLS}.md`

### Secondary (MEDIUM confidence)
- [Buf docs — lint usage (`buf:lint:ignore` placement, one-rule-per-directive)](https://buf.build/docs/lint/usage) — context for the legacy suppression dialect [CITED]
- [Go spec — range clause map iteration order](https://go.dev/ref/spec) — "The iteration order over maps is not specified…" (D-01 rationale) [CITED]

### Tertiary (LOW confidence)
- None — no unverified external claims; all behavioral statements are source-read.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — every component read in-repo; zero external installs
- Architecture: HIGH — etalon engine read line-by-line; collision inventory derived from actual symbol diffs
- Pitfalls: HIGH — grounded in byte-level diffs (main vs etalon) and Phase 1/2 accumulated notes
- Parity proofs (suppressions): MEDIUM-HIGH — fixtures exist and are identical, but attachment-equivalence is proven by running the table, not by reading

**Research date:** 2026-11-19
**Valid until:** 2026-12-19 (stable: internal codebase work; etalon pinned SHA)
