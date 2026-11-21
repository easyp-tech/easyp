# Phase 3: Двухпроходный движок и подавления - Pattern Map

**Mapped:** 2026-11-19
**Files analyzed:** 8 (5 new + 2 modified in `internal/core`, 1–2 new test files)
**Analogs found:** 8 / 8 — every new file has an exact porting source in the etalon clone; repo files supply the shared/half of each port

**Analog provenance rule for this phase:** the etalon clone (`/tmp/gsd-research/compiler-lab/easyp-cli` @ `08cef10`) is an **external porting source**, never a repo analog. Nothing from `/tmp` is git-tracked here; the executor copies code FROM it INTO `internal/core`. All repo paths below were verified with `git ls-files`.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/core/lint_linked.go` (new) | service (lint engine entrypoint) | batch (parallel windows) + request-response | etalon `easyp-cli/internal/core/lint.go` (port source); repo `internal/core/lint.go` (walk filter, `shouldIgnore`, `close` reused) | exact |
| `internal/core/lint_model.go` (new) | model (domain types, rule contracts) | transform | etalon `easyp-cli/internal/core/dom.go` lines 16–77 (port source); repo `internal/core/dom.go` (shared types stay single-declared) | exact |
| `internal/core/lint_linked_read.go` (new) | service (read-model builder) | file-I/O + transform | etalon `easyp-cli/internal/core/proto_info_read.go` lines 23–126 (port source); repo `internal/core/proto_info_read.go` lines 103–181 (`openImportFile`/`sourceImportPath`/`openDiskSource` reused verbatim) | exact |
| `internal/core/lint_suppressions.go` (new) | middleware (central suppression filter) | transform | etalon `easyp-cli/internal/core/comment_suppressions.go` lines 26–52 (protoast walk, port source); repo `internal/core/comment_suppressions.go` lines 13–26, 39–143 (skeleton carried with renames) | exact |
| `internal/core/link.go` (modify, additive only) | service (linked-set adapter) | transform | etalon `easyp-cli/internal/core/link.go` lines 87–184 (port source); repo `internal/core/link.go` (`linkSources`/`lintResolver` stay frozen) | exact |
| `internal/core/core.go` (modify, additive only) | config (DI container) | — | etalon `easyp-cli/internal/core/core.go` lines 46–54 (`importCache` + `parsedImports`) | exact |
| `internal/core/lint_linked_test.go` (new, planner may split) | test (engine integration, `-race`) | — | etalon `easyp-cli/internal/core/lint_compile_test.go`; repo `internal/core/link_test.go` (Phase 2 engine-test precedent) | exact |
| `internal/core/lint_suppressions_parity_test.go` (new) | test (main-vs-new parity table) | — | repo `internal/core/comment_suppressions_eof_test.go` + `comment_suppressions_test.go` (same fixtures re-driven through the new engine) | exact |

No file lacks an analog. The "No Analog Found" section below is replaced by the **External Porting Sources** section: the entire engine body comes from the etalon, and the only genuinely new behavior (D-01 sort) has an in-repo precedent (`internal/core/link.go:61-68`).

---

## Naming Collision Map (the phase's central trap — resolve before any port)

Everything below was verified by reading both trees this session. The etalon is the *post-flip* tree (old names gone); main is the *pre-flip* tree (old names live until Phase 5). Porting etalon files verbatim into package `core` will not compile.

### Colliding — new engine MUST use renamed identifiers

| Symbol | Main (live, do not touch) | Etalon (port source) | Resolution for planner |
|--------|---------------------------|----------------------|------------------------|
| `Lint` (method) | `internal/core/lint.go:17` — `func (c *Core) Lint(ctx, DirWalker) ([]IssueInfo, error)` | etalon `lint.go:22` — same name, new body | New entrypoint gets a new name (e.g. `LintLinked`); planner's discretion |
| `ProtoInfo` | `internal/core/dom.go:57-64` — `{directiveComments, Path, ImportPath, Info *unordered.Proto, ProtoFilesFromImport}` | etalon `dom.go:65-76` — `{Path, ImportPath, File *protoast.File, ImportedFiles, Resolved *Resolution}` | New struct, new name (e.g. `LinkedFileInfo`); keeps 41 main rules + old engine green |
| `Rule` (interface) | `internal/core/dom.go:24-28` — `Validate(ProtoInfo) ([]Issue, error)` (main ProtoInfo) | etalon `dom.go:18-23` — same shape, etalon ProtoInfo | Distinct type once ProtoInfo is renamed; give the new interface its own name (e.g. `LintRule`) to avoid confusing assertions like `var _ core.Rule` |
| `ModuleRule` | — does not exist in main | etalon `dom.go:29-33` | Name is free; embeds the new `Rule`/`ProtoInfo` names |
| `commentSuppressions` (method) | `internal/core/comment_suppressions.go:41` — takes main ProtoInfo | etalon `comment_suppressions.go:56` — takes etalon ProtoInfo | Rename the new method; body carries main's logic with protoast position access |
| `directiveComment` | `internal/core/comment_suppressions.go:28-31` — `comment *parser.Comment` | etalon `comment_suppressions.go:28-31` — `comment *protoast.Comment` | Rename the new struct |
| `parsedDirective` | `internal/core/comment_suppressions.go:33-37` (references main `directiveComment`) | etalon `comment_suppressions.go:48-52` | Rename (field type differs after `directiveComment` rename) |
| `readFilesFromImport` | `internal/core/proto_info_read.go:71-87` — over `*unordered.Proto` | etalon `proto_info_read.go:57-69` — over `*protoast.File` | Rename the new variant |
| `readFileFromImport` | `internal/core/proto_info_read.go:89-101` — returns `*unordered.Proto` | etalon `proto_info_read.go:98-126` — returns `*protoast.File`, cached | Rename the new variant |
| `AppendIssue` | `internal/core/dom.go:125-130` — `(issues, Rule, meta.Position, sourceName, comments)` | etalon `dom.go:79-82` — `(issues, Rule, Position, sourceName)` | Different signature: new engine gets its own helper name; Phase 4 rules migrate to it |

### Identical — reuse in place, NEVER redeclare (redeclaration = compile error in one package)

| Symbol | Single home (main) | Note |
|--------|--------------------|------|
| `Position` (= `protoast.Position` alias) | `internal/core/dom.go:16-20` | Already speaks the new engine's type |
| `Issue`, `IssueInfo` | `internal/core/dom.go:36-48` | Same shape both trees |
| `ImportPath`, `PackageName`, `ConvertImportPath` | `internal/core/dom.go:50-54, 194-196` | |
| `GetRuleName`, `toUpperSnakeCase`, `buildError` | `internal/core/dom.go:132-164` | Registry + suppression `known` map feed off this |
| `shouldIgnore` | `internal/core/lint.go:79-93` | Byte-identical main/etalon (`etalon lint.go:196-210`); new engine calls the existing method |
| `close` | `internal/core/lint.go:95-105` | Byte-identical both trees |
| `suppressionRange` / `suppressionRanges` / `contains` | `internal/core/comment_suppressions.go:13-26` | Byte-identical both trees (`etalon comment_suppressions.go:11-24`) |
| `parseDirectivePrefix` | `internal/core/comment_suppressions.go:126-143` | Byte-identical both trees (`etalon 138-155`) — carry zero changes |
| `compactDirectiveNames` | `internal/core/comment_suppressions.go:116-124` | Byte-identical both trees |
| `openImportFile` | `internal/core/proto_info_read.go:103-147` | Byte-identical both trees; security boundary (`filepath.IsLocal` + `importFileAllowed`) — never bypass |
| `sourceImportPath` | `internal/core/proto_info_read.go:149-167` | Byte-identical both trees |
| `openDiskSource` | `internal/core/proto_info_read.go:169-181` | Byte-identical both trees |
| `CompileRuleName`, `linkSources`, `lintResolver`, `hasPerFileSyntaxError`, `diagCollector` | `internal/core/link.go:20-145` | Phase 2 deliverable — frozen; new engine is its consumer |
| `ErrInvalidRule` | `internal/core/core.go:45` | Validation parity sentinel |
| `parsedImports` + `importCache` field | — absent in main | Port additively into `core.go` (etalon `core.go:46-54`) — no collision |

Free names (no collision either tree): `linkedFiles`, `linkedOutline`, `outline`, `resolution`, `usesThrough`, `Resolution`, `checkedFile`, `checkFile`, `collectFile`, `parseOnTheirOwn`, `readSource`, `protoInfoFromSource`, `protoInfoFromLinked`, `importCacheKey`, `cacheLintedOutline`, `lintWindow`.

Keep a rename-map comment at the top of each new file so Phase 5 can consolidate names after the old path is deleted.

---

## Pattern Assignments

### `internal/core/lint_linked.go` (service, batch/parallel windows)

**Port source (external):** etalon `/tmp/gsd-research/compiler-lab/easyp-cli/internal/core/lint.go` (236 lines, read in full)
**Repo analog (reused pieces):** `internal/core/lint.go` (`shouldIgnore` 79-93, `close` 95-105), `internal/parallel/parallel.go:16-59`, `internal/core/path_helpers/is_target_path.go` (`IsIgnoredPath`)

**Window constant + entrypoint** (etalon lint.go:17-22 — rename the method):
```go
// lintWindow bounds how many parsed files are held at once: files are read
// and parsed in parallel a window at a time, then checked in walk order.
const lintWindow = 256

func (c *Core) Lint(ctx context.Context, fsWalker DirWalker) ([]IssueInfo, error) {
```

**Walk with ignore filter** (etalon lint.go:25-42; identical shape to main lint.go:22-32):
```go
var paths []string
err := fsWalker.WalkDir(func(path string, err error) error {
    switch {
    case err != nil:
        return err
    case ctx.Err() != nil:
        return ctx.Err()
    case path_helpers.IsIgnoredPath(path, c.ignore):
        return nil
    case filepath.Ext(path) != ".proto":
        return nil
    }
    paths = append(paths, path)
    return nil
})
if err != nil {
    return nil, fmt.Errorf("fs.WalkDir: %w", err)
}
```

**Lazily initialized shared cache** (etalon lint.go:44-46):
```go
if c.importCache == nil {
    c.importCache = &parsedImports{files: map[string]*protoast.File{}}
}
```

**Pass 1 — parallel read, indexed results** (etalon lint.go:48-58):
```go
// Pass 1 reads every linted file once. Errors are kept per file and
// reported in walk order below, as a sequential run would.
sources := make([][]byte, len(paths))
readErrs := make([]error, len(paths))
err = parallel.Run(ctx, len(paths), 0, func(i int) error {
    sources[i], readErrs[i] = c.readSource(ctx, fsWalker, paths[i])
    return nil
})
if err != nil {
    return nil, fmt.Errorf("fs.WalkDir: %w", err)
}
```

**Link-or-fallback branch** (etalon lint.go:60-77; consumes main's `linkSources` — note main returns `*protoc.Linked`, the engine wraps it in `linkedFiles`):
```go
var linked *linkedFiles
if !slices.ContainsFunc(readErrs, func(err error) bool { return err != nil }) && len(paths) > 0 {
    var compileIssues []IssueInfo
    linkedRaw, compileIssues := c.linkSources(ctx, fsWalker, paths, sources)
    if linkedRaw != nil {
        linked = &linkedFiles{linked: linkedRaw, outlines: make(map[string]*linkedOutline)}
    } else if err := c.parseOnTheirOwn(ctx, fsWalker, paths, sources, readErrs); err != nil {
        return nil, fmt.Errorf("fs.WalkDir: %w", err)
    } else if len(compileIssues) > 0 && !slices.ContainsFunc(readErrs, ...) {
        // Like buf, lint rules do not run on files that do not compile.
        c.logger.Info(ctx, "lint completed", slog.Int("issues", len(compileIssues)))
        return compileIssues, nil
    }
}
```
(Above is the etalon body restructured for main's `linkSources` signature — main `internal/core/link.go:32` returns `(*protoc.Linked, []IssueInfo)` where etalon returns `(*linkedFiles, []IssueInfo)`. COMPILE issues return before suppression machinery runs — no suppression filtering on COMPILE.)

**Pass 2 — windows + indexed serial drain** (etalon lint.go:79-122 — the core mechanics; every line load-bearing):
```go
var res []IssueInfo
suppressed := make(map[string]suppressionRanges, len(paths))
infos := make([]ProtoInfo, min(lintWindow, len(paths)))   // new model type
infoErrs := make([]error, len(infos))
checked := make([]checkedFile, len(infos))
for start := 0; start < len(paths); start += lintWindow {
    window := paths[start:min(start+lintWindow, len(paths))]
    err := parallel.Run(ctx, len(window), 0, func(i int) error {
        infos[i], infoErrs[i] = NewModel{}, nil
        if readErrs[start+i] != nil {
            infoErrs[i] = readErrs[start+i]
            return nil
        }
        if linked != nil {
            infos[i], infoErrs[i] = c.protoInfoFromLinked(fsWalker, linked, window[i])
        } else {
            infos[i], infoErrs[i] = c.protoInfoFromSource(ctx, fsWalker, window[i], sources[start+i])
        }
        sources[start+i] = nil                      // release memory per window
        if infoErrs[i] == nil {
            checked[i] = c.checkFile(ctx, window[i], infos[i])
        }
        return nil
    })
    if err != nil {
        return nil, fmt.Errorf("fs.WalkDir: %w", err)
    }
    for i, path := range window {                   // serial drain IN WALK ORDER
        if infoErrs[i] != nil {
            return nil, fmt.Errorf("fs.WalkDir: c.protoInfoRead: %w", infoErrs[i])
        }
        if checked[i].err != nil {
            return nil, fmt.Errorf("fs.WalkDir: %w", checked[i].err)
        }
        if err := c.collectFile(infos[i]); err != nil {
            return nil, fmt.Errorf("fs.WalkDir: %w", err)
        }
        suppressed[path] = checked[i].suppressions
        res = append(res, checked[i].issues...)
        infos[i], checked[i] = NewModel{}, checkedFile{}
    }
}
```
**Decision flag for planner (Pitfall 6 / research Open Question 3):** etalon wraps per-file model errors as `fs.WalkDir: c.protoInfoRead: %w`; main's user-visible chain is `fs.WalkDir: c.lintProtoInfoRead: %w` (main lint.go:36 + 71). Research recommends main-wins text (REQ-2.4); pin with an exact-string test either way.

**D-01 total sort — the one NEW behavior vs the etalon.** The etalon has no sort (etalon lint.go:79-142). In-repo precedent for the exact style (main `internal/core/link.go:61-68`):
```go
// Deterministic issue order: by path, then line, then column.
slices.SortStableFunc(diags, func(a, b protoc.Diagnostic) int {
    return cmp.Or(
        strings.Compare(a.Path, b.Path),
        cmp.Compare(a.Line, b.Line),
        cmp.Compare(a.Col, b.Col),
    )
})
```
D-01 extends this to `(path, rule, line, col)` over `[]IssueInfo` at engine exit, before the return.

**ModuleRule.Finish re-filter** (etalon lint.go:124-138):
```go
for _, rule := range c.rules {
    moduleRule, ok := rule.(ModuleRule)
    if !ok {
        continue
    }
    for _, issue := range moduleRule.Finish() {
        if c.shouldIgnore(rule, issue.Path) {
            continue
        }
        if c.allowCommentIgnores && suppressed[issue.Path].contains(issue.RuleName, issue.Position.Line) {
            continue
        }
        res = append(res, issue)
    }
}
```

**checkedFile + checkFile + collectFile** (etalon lint.go:145-194):
```go
// checkedFile is one file's findings from rules that look at it alone.
type checkedFile struct {
    issues       []IssueInfo
    suppressions suppressionRanges
    err          error
}

func (c *Core) checkFile(ctx context.Context, path string, protoInfo NewModel) checkedFile {
    suppressions, err := c.linkedCommentSuppressions(protoInfo)   // renamed method
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

func (c *Core) collectFile(protoInfo NewModel) error {
    for _, rule := range c.rules {
        if moduleRule, ok := rule.(ModuleRule); ok {
            if err := moduleRule.Collect(protoInfo); err != nil {
                return fmt.Errorf("rule.Collect: %w", err)
            }
        }
    }
    return nil
}
```

**Fallback parser** (etalon lint.go:224-236):
```go
// parseOnTheirOwn parses each linted file without linking, recording syntax
// errors per file and registering outlines for imports of linted files.
func (c *Core) parseOnTheirOwn(ctx context.Context, disk FS, paths []string, sources [][]byte, errs []error) error {
    return parallel.Run(ctx, len(paths), 0, func(i int) error {
        file, err := protoast.Parse(paths[i], sources[i])
        if err != nil {
            errs[i] = fmt.Errorf("Parse: %w", err)
            return nil
        }
        c.cacheLintedOutline(disk, paths[i], file)
        return nil
    })
}
```

---

### `internal/core/lint_model.go` (model, transform)

**Port source (external):** etalon `easyp-cli/internal/core/dom.go:16-77`
**Repo analog (shared types stay here, single-declared):** `internal/core/dom.go`

**New rule contract** (etalon dom.go:16-33 — `Rule` re-declared against the new model type; `ModuleRule` is new):
```go
type (
    // Rule is an interface for a rule checking.
    Rule interface {
        // Message returns the message of the rule.
        Message() string
        // Validate validates the proto rule.
        Validate(ProtoInfo) ([]Issue, error)
    }

    // ModuleRule needs every file before it can report, such as rules that
    // compare files of one package or directory. Lint passes each file to
    // Collect and asks for findings once, after the last file; it does not
    // call Validate. Suppressions and ignore_only apply to the findings by path.
    ModuleRule interface {
        Rule
        Collect(ProtoInfo) error
        Finish() []IssueInfo
    }
)
```
(Both interfaces re-typed against the NEW model struct — do not reuse the names `Rule`/`ProtoInfo` per the collision map.)

**Resolution + new model struct** (etalon dom.go:49-76):
```go
// Resolution is what linking adds to a file.
Resolution struct {
    // UsedImports holds the direct imports that define, or publicly
    // re-export a file that defines, a name this file refers to.
    UsedImports map[ImportPath]bool
    // RPCTypes maps an rpc's byte offset to its request and response
    // types, fully qualified without a leading dot.
    RPCTypes map[int][2]string
}

// ProtoInfo is the information of a proto file.   (rename — collision)
ProtoInfo struct {
    Path string
    // ImportPath is relative to a declared source root, independently of the scan root.
    ImportPath string
    // File is the lint view of the file; ImportedFiles holds its direct imports.
    File          *protoast.File
    ImportedFiles map[ImportPath]*protoast.File
    // Resolved holds name resolution across files; nil when the files
    // could not be linked, in which case rules fall back to the syntax.
    Resolved *Resolution
}
```

**Shared by value — already declared in `internal/core/dom.go`, reuse:** `Position` alias (dom.go:16-20), `IssueInfo` (36-40), `Issue` (42-48), `ImportPath`/`PackageName` (50-54), `GetRuleName` (132-135), `buildError` (156-164). Do NOT redeclare.

**AppendIssue note:** etalon's `AppendIssue(issues, rule, pos, sourceName)` (etalon dom.go:79-82) has a different signature from main's (main dom.go:125-130 takes `meta.Position` + comments). The new engine needs its own issue-builder helper under a new name; Phase 4 rules migrate to it.

**ModuleRule list to fix at engine level (D-03)** — from lab `var _ core.ModuleRule` assertions (research-verified): the 10 current-main rules that are ModuleRules in the etalon are `DirectorySamePackage`, `PackageSameDirectory`, `PackageSameCsharpNamespace`, `PackageSameGoPackage`, `PackageSameJavaMultipleFiles`, `PackageSameJavaPackage`, `PackageSamePHPNamespace`, `PackageSameRubyPackage`, `PackageSameSwiftPrefix`, `RPCRequestResponseUnique` (`PackageNoImportCycle` is the 11th, arrives Phase 6). Registry source of truth: `internal/rules/builder.go:65-120` constructs 41 rules (note: REQUIREMENTS.md says "43" — Pitfall 7). Rules themselves are NOT ported this phase.

---

### `internal/core/lint_linked_read.go` (service, file-I/O + transform)

**Port source (external):** etalon `easyp-cli/internal/core/proto_info_read.go:23-126`
**Repo analog (reused verbatim, no redeclaration):** `internal/core/proto_info_read.go:103-181` (`openImportFile`, `sourceImportPath`, `openDiskSource`)

**readSource** (etalon proto_info_read.go:23-36):
```go
// readSource reads a linted file through the same opener imports use.
func (c *Core) readSource(ctx context.Context, fs FS, path string) ([]byte, error) {
    f, err := c.openDiskSource(fs, path)
    if err != nil {
        return nil, fmt.Errorf("Open: %w", err)
    }
    defer c.close(ctx, f, path)

    src, err := io.ReadAll(f)
    if err != nil {
        return nil, fmt.Errorf("Parse: ReadAll: %w", err)
    }
    return src, nil
}
```

**protoInfoFromSource** (etalon proto_info_read.go:38-55; calls the renamed `readFilesFromImport` variant):
```go
func (c *Core) protoInfoFromSource(ctx context.Context, fs FS, path string, src []byte) (ProtoInfo, error) {
    file, err := protoast.Parse(path, src)
    if err != nil {
        return ProtoInfo{}, fmt.Errorf("Parse: %w", err)
    }

    imported, err := c.readFilesFromImport(ctx, fs, file)
    if err != nil {
        return ProtoInfo{}, fmt.Errorf("readFilesFromImport: %w", err)
    }

    return ProtoInfo{
        Path:          path,
        ImportPath:    c.sourceImportPath(fs, path),
        File:          file,
        ImportedFiles: imported,
    }, nil
}
```

**Cache key + linted-outline cache** (etalon proto_info_read.go:71-94):
```go
// importCacheKey identifies importName as openImportFile would find it on
// disk: the lint root plus the name, which the lint FS tries first.
func importCacheKey(disk FS, importName string) (string, bool) {
    rooted, ok := disk.(interface{ RootPath() string })
    if !ok {
        return "", false
    }
    return rooted.RootPath() + "\x00" + importName, true
}

// cacheLintedOutline records a linted file for imports of the same name. ...
func (c *Core) cacheLintedOutline(disk FS, path string, file *protoast.File) {
    key, cacheable := importCacheKey(disk, path)
    if !cacheable || c.importCache == nil {
        return
    }
    c.importCache.mu.Lock()
    if _, ok := c.importCache.files[key]; !ok {
        c.importCache.files[key] = file.Outline()
    }
    c.importCache.mu.Unlock()
}
```

**Cached import reader** (etalon proto_info_read.go:96-126 — renamed; parse errors wrapped as `readProtoFile:` inside an `os.PathError` for text parity):
```go
func (c *Core) readFileFromImport(ctx context.Context, disk FS, importName string) (*protoast.File, error) {
    key, cacheable := importCacheKey(disk, importName)
    if cacheable && c.importCache != nil {
        c.importCache.mu.Lock()
        parsed, ok := c.importCache.files[key]
        c.importCache.mu.Unlock()
        if ok {
            return parsed, nil
        }
    }

    f, err := c.openImportFile(disk, importName)   // existing main helper — security boundary
    if err != nil {
        return nil, fmt.Errorf("openImportFile: %w", err)
    }
    defer c.close(ctx, f, importName)

    parsed, err := protoast.Read(importName, f)
    if err != nil {
        return nil, fmt.Errorf("readProtoFile: %w", &os.PathError{Op: "parse", Path: importName, Err: err})
    }
    parsed = parsed.Outline()
    if cacheable && c.importCache != nil {
        c.importCache.mu.Lock()
        c.importCache.files[key] = parsed
        c.importCache.mu.Unlock()
    }
    return parsed, nil
}
```

**API the new code consumes (repo, verified):**
- `protoast.Parse(path, src)` / `protoast.Read(name, reader)` / `(*protoast.File).Outline()` — `internal/protoast/protoast.go` (Outline at 200-222)
- `protoast.OutlineOf(syntax.File)` — `internal/protoast/outline.go:8-36` (builds outline straight from the linked tree, no positions)
- `protoast.Build(syntax.File)` — `internal/protoast/build.go`

---

### `internal/core/lint_suppressions.go` (middleware/filter, transform)

**Port source (external):** etalon `easyp-cli/internal/core/comment_suppressions.go:26-52`
**Repo analog (skeleton carried with renames; walk swapped):** `internal/core/comment_suppressions.go`

**The porting target — protoast walk** (etalon comment_suppressions.go:26-46; replaces main's reflection walk at `internal/core/comment_suppressions.go:145-231`):
```go
// directiveComment is a comment with the line range of the declaration it
// annotates; first is 0 when it annotates nothing.
type directiveComment struct {                     // RENAME (collides with main)
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

**Protoast facts the walk relies on (repo, verified):** `Comment{Raw, Pos, End}` + `Lines()` strips markers and splits on `\n` (`internal/protoast/protoast.go:39-55`); `Decl{Pos, End, Leading, Detached, Trailing}` (57-70); `File.Comments` is "every comment in source order" (89-90); `Owner` returns nil for comments that annotate nothing — EOF, before closing brace (102-105). Positions are 1-based line, 1-based rune column, tab advances by 1 (package doc 5-10).

**Skeleton — carry from main, changing only position access.** Main's method (`internal/core/comment_suppressions.go:39-114`) and the etalon's (`etalon comment_suppressions.go:54-126`) are logically identical; the only textual delta is comment position access: main reads `comment.comment.Meta.Pos.Line/Column`, the new method reads `comment.comment.Pos.Line/Column` (protoast). Error strings to carry byte-identical (REQ-2.5/3.2 main-wins):
- `"%s:%d:%d: malformed lint directive %q"` (main :65)
- `"%s:%d:%d: unknown or unsupported lint rule %q in comment directive"` (main :70)
- `"%s:%d: easyp:enable %s has no preceding disable"` (main :90)
- `"%s:%d: suppression %s must annotate a declaration or have a matching easyp:enable"` (main :108)

**`known` name set** (main :46-54 == etalon :58-66): seed from `c.knownLintRules` (wired from `rules.AllRuleNames()` at `internal/api/runtime.go:74`), fall back to `GetRuleName` over `c.rules` when empty.

**Shared in place (never redeclare):** `suppressionRanges` + `contains` (`internal/core/comment_suppressions.go:13-26`), `parseDirectivePrefix` (126-143), `compactDirectiveNames` (116-124).

**The old reflection walk (main :145-231) is NOT deleted** — it stays for the old path until Phase 5.

---

### `internal/core/link.go` (modify — additive only)

**Port source (external):** etalon `easyp-cli/internal/core/link.go:87-184`
**Repo analog (frozen):** `internal/core/link.go` — `linkSources` (23-82), `hasPerFileSyntaxError` (84-95), `diagCollector` (97-112), `lintResolver` (114-145) stay untouched. Do NOT port the etalon's `linkSources` (etalon link.go:20-56): main's version wins (syntax-error precedence, deterministic COMPILE sort, diagCollector).

**linkedFiles adapter over main's `*protoc.Linked`** (etalon link.go:129-156):
```go
// linkedFiles serves lint models and import outlines from one Link result.
type linkedFiles struct {
    linked   *protoc.Linked
    mu       sync.Mutex
    outlines map[string]*linkedOutline
}

type linkedOutline struct {
    once sync.Once
    file *protoast.File
}

// outline returns the outline of a linked file, computed once from its tree.
func (l *linkedFiles) outline(path string) (*protoast.File, bool) {
    file, ok := l.linked.File(path)
    if !ok {
        return nil, false
    }
    l.mu.Lock()
    entry, ok := l.outlines[path]
    if !ok {
        entry = &linkedOutline{}
        l.outlines[path] = entry
    }
    l.mu.Unlock()
    entry.once.Do(func() { entry.file = protoast.OutlineOf(file.Syntax) })
    return entry.file, true
}
```
The mutex + `sync.Once` discipline is race-mandatory: parallel window workers hit the same imports. Port verbatim.

**Model-from-linked with tree drop** (etalon link.go:158-184; `resolution` call at the end):
```go
// protoInfoFromLinked builds a linted file's model from its linked tree, then
// drops the tree; its outline stays available to files importing it.
func (c *Core) protoInfoFromLinked(disk FS, linked *linkedFiles, path string) (ProtoInfo, error) {
    file, ok := linked.linked.File(path)
    if !ok || file.Syntax == nil {
        return ProtoInfo{}, fmt.Errorf("Parse: %s was not linked", path)
    }
    model := protoast.Build(file.Syntax)
    linked.outline(path)
    file.Syntax = nil                          // memory: tree dropped after build

    imported := make(map[ImportPath]*protoast.File, len(model.Imports))
    for _, imp := range model.Imports {
        outline, ok := linked.outline(imp.Path)
        if !ok {
            return ProtoInfo{}, fmt.Errorf("readFilesFromImport: %s was not linked", imp.Path)
        }
        imported[ConvertImportPath(imp.Location)] = outline
    }
    return ProtoInfo{
        Path:          path,
        ImportPath:    c.sourceImportPath(disk, path),
        File:          model,
        ImportedFiles: imported,
        Resolved:      resolution(linked.linked, path),
    }, nil
}
```
The `Location` (quoted literal) keying `ImportedFiles` comes from protoast `Import.Location` (`internal/protoast/protoast.go:112-118`), not protoc.

**Resolution facts for rules** (etalon link.go:87-127; both are new names — no collision):
```go
// resolution derives the per-file facts rules need from the linked set.
func resolution(linked *protoc.Linked, path string) *Resolution {
    file, ok := linked.File(path)
    if !ok {
        return nil
    }
    res := &Resolution{UsedImports: make(map[ImportPath]bool, len(file.Imports)), RPCTypes: make(map[int][2]string, len(file.Methods))}
    for _, imp := range file.Imports {
        res.UsedImports[ImportPath(imp.Path)] = usesThrough(linked, file, imp.Path)
    }
    for _, m := range file.Methods {
        res.RPCTypes[int(m.Offset)] = [2]string{m.Input, m.Output}
    }
    return res
}

// usesThrough reports whether file uses a name from imported or from a file
// that imported re-exports through public imports.
func usesThrough(linked *protoc.Linked, file *protoc.LinkedFile, imported string) bool {
    seen := map[string]bool{}
    queue := []string{imported}
    for len(queue) > 0 {
        path := queue[0]
        queue = queue[1:]
        if seen[path] {
            continue
        }
        seen[path] = true
        if file.Uses[path] {
            return true
        }
        if dep, ok := linked.File(path); ok {
            for _, imp := range dep.Imports {
                if imp.Public {
                    queue = append(queue, imp.Path)
                }
            }
        }
    }
    return false
}
```

**Repo API consumed (verified):** `protoc.Linked.File(path) (*LinkedFile, bool)` + `LinkedFile{Path, Imports []LinkedImport{Path, Public}, Uses map[string]bool, Methods []LinkedMethod{Offset, Input, Output}, Syntax *syntax.File}` — `internal/easypc/protoc/link.go:15-52` (read-only; sealed tree, never edit `internal/easypc`).

---

### `internal/core/core.go` (modify — additive only)

**Port source (external):** etalon `easyp-cli/internal/core/core.go:46-54`

Add to main's `Core` struct (`internal/core/core.go:16-42`) — nothing else in the struct, `Options` (50-78), or `New` (80-107) changes:
```go
// importCache holds parsed import outlines for lint; files are parsed in
// parallel, so it is shared and locked.
importCache *parsedImports
```
and the companion type:
```go
type parsedImports struct {
    mu    sync.Mutex
    files map[string]*protoast.File
}
```
plus the `sync` and `internal/protoast` imports. Dormancy rule (Pitfall 5): additive-only; no Options field (the cache is engine-internal, lazily initialized per etalon lint.go:44-46).

---

### `internal/core/lint_linked_test.go` (test, integration + race)

**Analog 1 (repo, engine-level construction pattern):** `internal/core/link_test.go:18-44` — table test, inline `core.New`, `disk.NewFSWalker`, per-case assertions:
```go
tests := []struct {
    name    string
    paths   []string
    sources [][]byte
    setup   func(t *testing.T, root string)
    newCore func(t *testing.T, root string) *core.Core
    assert  func(t *testing.T, ...)
}{ ... }
```
with `require.NoError(t, os.WriteFile(..., 0o644/0o600))` into `t.TempDir()` and `disk.NewFSWalker(root, ".")` (`internal/fs/fs/dir_walker.go:19-32`).

**Analog 2 (external, COMPILE routing through the new engine):** etalon `easyp-cli/internal/core/lint_compile_test.go:16-43`:
```go
func TestLintReportsCompileErrorsInsteadOfRules(t *testing.T) {
    t.Parallel()
    root := t.TempDir()
    raw := "syntax = \"proto3\";\npackage example;\nmessage bad_name {\n  required string s = 1;\n}\n"
    require.NoError(t, os.WriteFile(filepath.Join(root, "test.proto"), []byte(raw), 0o644))
    app := core.New(core.Options{Logger: logger.NewNop(), Rules: []core.Rule{&rules.MessagePascalCase{}}, KnownLintRules: rules.AllRuleNames()})
    got, err := app.Lint(t.Context(), disk.NewFSWalker(root, "."))
    require.NoError(t, err)
    require.Len(t, got, 1, "rules do not run on files that do not compile")
    require.Equal(t, core.CompileRuleName, got[0].RuleName)
    // ... Line 4, Column 3, message "Required fields are not allowed in proto3."
}
// TestLintMissingImportIsNotACompileIssue: missing.proto -> require.ErrorAs(&core.OpenImportFileError)
```
For Phase 3, real rules are NOT ported (research A2): substitute test doubles implementing the new interface. Required test coverage per research Wave-0: determinism (repeated `-race` runs byte-identical + D-01 sort assertion), COMPILE routing, fallback matrix (`parseOnTheirOwn`, `OpenImportFileError`, syntax error → command error), ModuleRule Collect/Finish lifecycle through a double, statelessness race test (multi-file fixture, shared rule instances, parallel windows).

**Analog 3 (external, Collect/Finish driver):** etalon `easyp-cli/internal/rules/module_rules_test.go:13-25`:
```go
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

**Analog 4 (external, model fixture helper + stateless double):** etalon `easyp-cli/internal/rules/init_test.go:117-123` and the stateless rule shape at `easyp-cli/internal/rules/import_used.go:8-12`:
```go
func parseProto(t testing.TB, path, src string) core.ProtoInfo {
    t.Helper()
    file, err := protoast.Parse(path, []byte(src))
    require.NoError(t, err)
    return core.ProtoInfo{Path: path, File: file}
}

// stateless per-file rule shape (zero receiver writes):
type ImportUsed struct{}
```
Handwritten doubles follow `internal/core/dom.go`'s actual interface (go-testing skill pattern: `var _ core.LintRule = (*mockRule)(nil)`).

**Optional fixture tree:** precedent `internal/core/testdata/link/multi_file/` (Phase 2, git-tracked) if file-based fixtures outgrow `t.TempDir()` strings.

---

### `internal/core/lint_suppressions_parity_test.go` (test, main-vs-new parity table)

**Analog (repo — the fixtures are the parity oracle):** `internal/core/comment_suppressions_eof_test.go:15-47` — the exact case matrix to re-drive through the NEW engine entrypoint:
```go
for _, tc := range []struct {
    name, body, wantError string
    want                  int
}{
    {name: "EOF with newline", body: declarations + "// easyp:enable MESSAGE_PASCAL_CASE\n"},
    {name: "EOF without newline", body: declarations + "// easyp:enable MESSAGE_PASCAL_CASE"},
    {name: "block comment at EOF", body: declarations + "/* easyp:enable MESSAGE_PASCAL_CASE */"},
    {name: "orphan enable", body: "message Valid {}\n// easyp:enable MESSAGE_PASCAL_CASE", wantError: "no preceding disable"},
    {name: "unknown trailing rule", body: "message Valid {}\n// easyp:enable UNKNOWN", wantError: "unknown or unsupported"},
    {name: "unattached disable", body: "message Valid {}\n// easyp:disable MESSAGE_PASCAL_CASE", wantError: "must annotate a declaration"},
    {name: "unpaired attached disable", body: declarations, want: 1},
    {name: "marker in string", body: "option java_package = \"// easyp:enable UNKNOWN\";\nmessage bad_name {}", want: 1},
    {name: "nested closing brace", body: "message Outer {\n" + declarations + "// easyp:enable MESSAGE_PASCAL_CASE\n}\n"},
}
```
plus the policy-independence and exact-name cases from `internal/core/comment_suppressions_test.go:16-44` (`AllowCommentIgnores` true/false; `nolint:` typo → `require.ErrorContains`), and the ignore/ignore_only config cases the old path covers only implicitly. Research-verified: these fixtures are byte-identical between main and the etalon, so expected values carry over unchanged. Format per CONTEXT Discretion item 3 (table shape is planner's).

**Position-convention caution (research A4):** protoast columns are rune-based; go-protoparser columns are byte-based. Lines (the suppression match key) never diverge; directive *error text* columns may differ on tab/UTF-8 lines. If a fixture includes tabs, pin expectations explicitly.

---

## Shared Patterns

### Bounded parallel execution
**Source:** `internal/parallel/parallel.go:16-59` (git-tracked, reused as-is per D-02)
**Apply to:** `lint_linked.go` pass 1, windows, `parseOnTheirOwn`
```go
func Run(ctx context.Context, count, limit int, task func(index int) error) error
// limit <= 0 -> runtime.NumCPU(); count <= 0 -> no-op; errors returned in index order
```
Indexed result slices (`sources[i]`, `infoErrs[i]`, `checked[i]`) need no locks; reassembly happens only in the serial drain by index — never in completion order (Pitfall 3).

### Error wrapping
**Source:** go-code-style skill + every analog (`fmt.Errorf("<callee>: %w", err)`, callee-only name). Sentinels: `errors.Is`; typed details: `errors.As` (e.g. `*core.OpenImportFileError` at etalon lint_compile_test.go:41-43). New-file wraps to preserve: `Open:`, `Parse:`, `ReadAll:`, `openImportFile:`, `readProtoFile:`, `commentSuppressions:`, `rule.Validate:`, `rule.Collect:`, `fs.WalkDir:`.

### Deferred close hygiene
**Source:** `internal/core/lint.go:95-105` (== etalon lint.go:212-222)
**Apply to:** every file/reader opened in the new read path
```go
defer c.close(ctx, f, path)   // logs Close failures via logger.Warn; never bare defer f.Close()
```

### Deterministic sorting style
**Source:** `internal/core/link.go:61-68` — `slices.SortStableFunc` + `cmp.Or(strings.Compare(...), cmp.Compare(...)...)`. D-01's engine-exit sort reuses this style with a `RuleName` term inserted before line/col. This is the only new behavior vs the etalon (etalon lint.go:79-142 has no sort).

### Engine construction in tests
**Source:** `internal/core/comment_suppressions_test.go:28`, etalon lint_compile_test.go:21
```go
app := core.New(core.Options{Logger: logger.NewNop(), Rules: []core.Rule{...}, KnownLintRules: rules.AllRuleNames()})
got, err := app.Lint(t.Context(), disk.NewFSWalker(root, "."))
```
(For the new engine, the entrypoint call site changes to the renamed method; Options unchanged.)

### Test conventions
**Source:** go-testing skill + `internal/core/comment_suppressions_eof_test.go` — named table cases (lowercase underscores), `t.Parallel()` with per-case `t.TempDir()`, `require` for fatal preconditions / `require.ErrorContains` for directive text, `t.Helper()` on fixtures. `-race -count=1` mandatory (project convention; raw `go test` — `task` absent from this environment).

### lint.use validation parity (existing — do not touch)
**Source:** `internal/rules/selection.go:14-52` + `internal/rules/builder.go:61-64`
Main already validates: `ValidateNames` → `core.ErrInvalidRule` (wrapped `use:`/`except:`/`ignore_only:` per selection); `rules.New` wraps with `validateConfigRuleNames:`. D-06 parity = keep this path untouched; the engine receives pre-validated rules. Existing test: `internal/rules/selection_test.go:43` (`TestInvalidLintNamesAreRejectedBeforeFiltering`). `knownLintRules` comes from `rules.AllRuleNames()` (`internal/rules/builder.go:38-45`) via `internal/api/runtime.go:74`.

---

## External Porting Sources (not repo analogs — never name these as in-repo paths in plans)

All under `/tmp/gsd-research/compiler-lab/easyp-cli` @ `08cef10` (fallback: `gh repo clone easyp-tech/compiler-lab`, directory `easyp-cli`). Nothing here is git-tracked in this repo; the executor copies FROM these files:

| Etalon file | Feeds | Pieces |
|-------------|-------|--------|
| `internal/core/lint.go` (236 lines) | `lint_linked.go` | `lintWindow`, engine body, `checkedFile`, `checkFile`, `collectFile`, `parseOnTheirOwn` (224-236) |
| `internal/core/dom.go` | `lint_model.go` | `ModuleRule` (25-33), `Resolution` (49-57), new `ProtoInfo` (65-76), `AppendIssue` shape (79-82) |
| `internal/core/proto_info_read.go` | `lint_linked_read.go` | `readSource` (23-36), `protoInfoFromSource` (38-55), `importCacheKey` (71-79), `cacheLintedOutline` (81-94), `readFileFromImport` (96-126) |
| `internal/core/comment_suppressions.go` | `lint_suppressions.go` | `directiveComment`+`directiveComments` (26-46) |
| `internal/core/link.go` | `link.go` additions | `resolution` (87-101), `usesThrough` (103-127), `linkedFiles`/`linkedOutline`/`outline` (129-156), `protoInfoFromLinked` (158-184) |
| `internal/core/core.go` | `core.go` addition | `importCache` field + `parsedImports` (46-54) |
| `internal/core/lint_compile_test.go` | engine tests | COMPILE routing (16-31), missing import (33-43) |
| `internal/rules/{module_rules_test.go, init_test.go, package_same_java_package.go, import_used.go}` | test doubles, ModuleRule list | `issuePaths` driver, `parseProto`, `var _ core.ModuleRule` set, stateless rule shape |

**Do NOT port from the etalon:** its `linkSources` (main's Phase 2 version wins), its `openImportFile`/`sourceImportPath`/`openDiskSource`/`close`/`shouldIgnore`/skeleton lexers (byte-identical already in main — reuse, don't duplicate).

## Metadata

**Analog search scope:** `internal/core`, `internal/rules`, `internal/parallel`, `internal/protoast`, `internal/easypc/protoc`, `internal/api`, `internal/fs/fs`, `internal/core/path_helpers`, `internal/core/testdata`; etalon clone `/tmp/gsd-research/compiler-lab/easyp-cli/internal/{core,rules}`
**Files read in full this session:** etalon `lint.go` (236), `link.go`, `dom.go`, `core.go`, `proto_info_read.go`, `comment_suppressions.go`, `lint_compile_test.go`, `module_rules_test.go`, `package_same_java_package.go`, `import_used.go`, `init_test.go`; repo `lint.go`, `core.go`, `link.go`, `dom.go`, `comment_suppressions.go`, `proto_info_read.go`, `lint_proto_info.go`, `fs.go`, `protoast.go`, `outline.go`, `parallel.go`, `selection.go`, `builder.go`, `runtime.go`, both suppression test files, main `package_same_java_package.go`, `dir_walker.go`, `protoc/link.go` (1-60), `link_test.go` (1-130)
**Git-tracking check:** every repo path above verified via `git ls-files` (all tracked); no gitignored mirror paths used
**Pattern extraction date:** 2026-11-19
