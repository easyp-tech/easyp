# Phase 2: Линкер и маршрутизация ошибок - Pattern Map

**Mapped:** 2026-11-19
**Files analyzed:** 4 (all new; no existing files are modified in this phase)
**Analogs found:** 4 / 4 (1 via external etalon + in-repo role-matches; 3 via in-repo analogs)

## Tracked-Source Note (gate compliance)

Every in-repo path named below was verified with `git ls-files` (all of `internal/core/*`, `internal/api/*`, `internal/easypc/protoc/*` are tracked). One special source is **external, not a repo analog**: the lab reference implementation at `/tmp/gsd-research/compiler-lab/easyp-cli/internal/core/{link.go,lint.go,lint_compile_test.go}` — it is the code being *ported*, exists only as a local clone (fallback: `gh repo clone easyp-tech/compiler-lab`, Phase 1 D-09), and its load-bearing parts are quoted verbatim in `02-RESEARCH.md`. Do not treat /tmp paths as stable; copy from the excerpts here and the tracked repo analogs. `internal/easypc/**` is sealed (Phase 1 D-04): read-only, cited for API signatures only, never edited.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/core/link.go` | service (corpus linker + failure classifier) | transform (paths+sources → linked trees / issues / fallback) with event-driven diag collection | Porting source: lab `internal/core/link.go` (external etalon); in-repo role-matches: `internal/core/proto_info_read.go`, `internal/core/lint.go` | reference-etalon (no in-repo exact analog: first `protoc.Link` consumer) |
| `internal/core/link_test.go` | test (unit/integration, same-package) | batch (fixture matrix) | `internal/core/proto_info_read_test.go` (+ lab `lint_compile_test.go` expectations) | role-match, strong |
| `internal/core/testdata/link/` | test fixtures | file-I/O | `internal/core/testdata/position_contract_utf8.proto` | role-match |
| `internal/api/lint_link_matrix_test.go` | test (integration, Action-path + subprocess exit codes) | request-response (CLI Action routing) | `internal/api/lint_v1_test.go` (+ `internal/api/lint.go` exit route as the tested subject) | role-match, strong |

Modified files: **none** (D-04: `core.Lint()`, `internal/api/lint.go`, `lint_proto_info.go`, `proto_info_read.go`, `dom.go` all unchanged until Phase 5).

## Pattern Assignments

### `internal/core/link.go` (service, transform + event-driven diag collection)

**Porting source (external etalon):** `compiler-lab/easyp-cli/internal/core/link.go:17-85` — port with minimal adaptation; the lab's `linkedFiles`/`linkedOutline`/`protoInfoFromLinked`/`resolution`/`usesThrough` (lines 87-184) are **Phase 3 scope — do not port now** (D-04/D-02: classifier only, three-outcome contract).

**In-repo analog for conventions:** `internal/core/proto_info_read.go` — the resolver must delegate to exactly this chain, adding no filesystem access of its own.

**The boundary the resolver wraps** (`internal/core/proto_info_read.go:103-147`):
```go
func (c *Core) openImportFile(disk FS, importName string) (io.ReadCloser, error) {
	if !filepath.IsLocal(importName) {
		return nil, fmt.Errorf("invalid import path %q", importName)
	}
	var f io.ReadCloser
	var err error
	allowed := true
	if rooted, ok := disk.(interface{ RootPath() string }); ok && c.importFileAllowed != nil {
		allowed = c.importFileAllowed(filepath.Join(rooted.RootPath(), importName))
	}
	if allowed {
		f, err = c.openDiskSource(disk, importName)
	} else {
		err = os.ErrNotExist
	}
	switch {
	case errors.Is(err, os.ErrNotExist):
		// Search dependency roots when the current filesystem has no matching file.
	case err != nil:
		return nil, fmt.Errorf("Open: %w", err)
	default:
		return f, nil
	}

	for _, root := range c.importRoots {
		fullPath := filepath.Join(root, importName)
		f, err := c.openSourceFile(fullPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("Open: %w", err)
		}
		return f, nil
	}

	f, err = wellknownimports.Content.Open(importName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &OpenImportFileError{FileName: importName}
		}
		return nil, fmt.Errorf("Open: %w", err)
	}
	return f, nil
}
```
Key fact: a file blocked by `importFileAllowed` surfaces as `os.ErrNotExist` (lines 116-117, and `openDiskSource` line 173) — `errors.Is(err, os.ErrNotExist)` in the resolver catches it, so blocked imports classify as missing → fallback, and the existing per-file path later reports `OpenImportFileError` (exit 2) exactly like main. Note line 105: a non-local import path (`"../x.proto"`) returns a plain error, NOT `ErrNotExist`/`OpenImportFileError` → missing flag stays false → COMPILE class (research Open Question 2 / Assumption A4).

**Classifier core to port (etalon link.go:18-56, verbatim; adapt outcome 1 to the Phase 2 contract):**
```go
// CompileRuleName reports compiler errors found while linking linted files.
const CompileRuleName = "COMPILE"

func (c *Core) linkSources(ctx context.Context, disk FS, paths []string, sources [][]byte) (*linkedFiles, []IssueInfo) {
	linted := make(map[string][]byte, len(paths))
	for i, path := range paths {
		linted[path] = sources[i]
	}
	resolver := &lintResolver{core: c, disk: disk, linted: linted}
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
				Position: Position{Line: d.Line, Column: d.Col},
				Message:  d.Message,
				RuleName: CompileRuleName,
			},
		})
	}
	return nil, issues
}
```
Phase 2 adaptations (all decided in CONTEXT/RESEARCH, planner picks names/signature — discretion):
- Outcome 1 returns the `*protoc.Linked` value (trees for Phase 3), not the lab's `*linkedFiles` wrapper — outlines/outlines-map are Phase 3 property.
- Add **D-03 sorting** before building issues (the etalon does NOT sort — deliberate delta). In-repo precedent for stable sort (`internal/api/lint_v1.go:219-222`):
```go
	// Grouping must not reorder user-facing diagnostics between policy subtrees.
	slices.SortStableFunc(issues, func(a, b core.IssueInfo) int {
		return strings.Compare(filepath.ToSlash(a.Path), filepath.ToSlash(b.Path))
	})
```
  For diags: sort by Path, then Line, then Col (see RESEARCH.md Code Examples for a `cmp.Or` formulation).
- Add **LINT-ENG-04 precedence** (etalon `lint.go:60-77` mechanics): after a failed Link with non-empty diags and no missing flag, per-file parse of the corpus; if any file has a syntax error → return `(nil, nil)` (fallback; existing path turns it into a command error). Signature-shape from the etalon's Lint:
```go
	// (etalon lint.go:64-77) linked == nil means "files parsed on their own"
	linked, compileIssues = c.linkSources(ctx, fsWalker, paths, sources)
	if linked == nil {
		if err := c.parseOnTheirOwn(ctx, fsWalker, paths, sources, readErrs); err != nil { ... }
		if len(compileIssues) > 0 && !slices.ContainsFunc(readErrs, func(err error) bool { return err != nil }) {
			return compileIssues, nil
		}
	}
```
  Whether `parseOnTheirOwn` lives inside the classifier or the classifier exposes `(linked, compileIssues, fallback)` and the caller does precedence — planner discretion; the invariant is fixed: **syntax error in any file beats COMPILE**.
- Race safety (RESEARCH Pitfall 7): keep `missing atomic.Bool`; wrap `diags` append in a small `sync.Mutex` (etalon relies on undocumented sequential ErrorHandler contract).

**Resolver to port (etalon link.go:58-85, verbatim):**
```go
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

**Sealed API signatures consumed** (`internal/easypc/protoc/`, read-only):
- `link.go:54-58` — `Link(ctx context.Context, files []string, opts ...Option) (*Linked, error)`; "Files are read only through the resolvers given with WithResolver and, if enabled, the standard imports; it never falls back to the file system... failures return ErrCompilationFailed."
- `resolver.go:14-21` — `ErrFileNotFound = errors.New("file not found")`; `Resolver` interface: `FindFileByPath(ctx context.Context, path string) ([]byte, error)`.
- `options.go:19-34, 72-77` — `WithSyntaxTrees(keep bool)`, `WithResolver(r Resolver)`, `WithErrorHandler(handler func(Diagnostic))`. **Do NOT pass `WithStandardImports(true)`** (would add a second WKT source beside `wellknownimports` in `openImportFile` — anti-pattern).
- `diagnostic.go:10-16` — `Diagnostic{Path string; Line, Col int; Message string; IsWarn bool}` — **no `Semantic` field at this seam** (Assumption A5); syntax-vs-semantic distinction is behavioral (precedence), not field-based.
- `core.Core` fields already available: `logger`, `importRoots`, `importFileAllowed`, `sourceFileOpen` (`internal/core/core.go:16-42`) — resolver needs no new Core state, consistent with "core.New gets full options, no post-construction setters".

**Issue target shape** (`internal/core/dom.go:20, 36-48, 169-176`):
```go
type Position = protoast.Position   // fields: Filename, Offset, Line, Column

type IssueInfo struct {
	Issue
	Path string
}

type Issue struct {
	Position   Position
	SourceName string
	Message    string
	RuleName   string
}
```
COMPILE issues leave `Filename`/`Offset`/`SourceName` zero (etalon shape, D-03): text printer output becomes `path:4:3: message (COMPILE)`; JSON printer serializes Position whole, so zero fields are visible — pin this in the matrix (RESEARCH Pitfall 5). Note: on lines containing tabs/multi-byte UTF-8 before the error, `Col` follows the compiler's LineCol convention (1-based, tab stop 8, byte columns) — NOT the rune columns of D-07 (RESEARCH Pitfall 2 / Assumption A2); the matrix must include a tab/UTF-8 fixture.

---

### `internal/core/link_test.go` (test, batch fixture matrix)

**Analog:** `internal/core/proto_info_read_test.go` — same-package (`package core`) test reaching unexported methods; direct minimal `Core` construction; `fs.NewFSWalker`; typed-error assertions.

**Core construction without full Options** (`proto_info_read_test.go:152-153, 198-200`):
```go
	engine := &Core{logger: logger.NewNop(), importRoots: []string{root}}
	walker := fs.NewFSWalker(root, ".")

	lintInfo, err := engine.lintProtoInfoRead(t.Context(), walker, name)
```
The classifier tests follow the same shape: `engine := &Core{logger: logger.NewNop(), importRoots: []string{root}}` (add `importFileAllowed`/`sourceFileOpen` fields only for the blocked-import cell), then call the new classifier directly.

**Typed-error assertion pattern** (`proto_info_read_test.go:131-140`):
```go
			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			case tt.wantMissing:
				var missing *OpenImportFileError
				require.ErrorAs(t, err, &missing)
				assert.Equal(t, tt.importName, missing.FileName)
			default:
				require.ErrorContains(t, err, tt.wantMessage)
			}
```
The missing-import cell asserts the fallback outcome ends in `*OpenImportFileError` (the exit-2 carrier, `dom.go:178-184`).

**Expectation numbers (etalon `lint_compile_test.go:16-43`, verbatim fixtures):**
```go
	raw := "syntax = \"proto3\";\npackage example;\nmessage bad_name {\n  required string s = 1;\n}\n"
	// expected COMPILE issue:
	//   RuleName == "COMPILE" (core.CompileRuleName); Path == "test.proto"
	//   Message == "Required fields are not allowed in proto3."
	//   Position.Line == 4; Position.Column == 3
	//
	// missing-import fixture:
	raw := "syntax = \"proto3\";\npackage example;\nimport \"missing.proto\";\n"
	// expected: fallback, not COMPILE; error chain reaches *core.OpenImportFileError
```

**Test conventions to copy** (`.agents/skills/go-testing/SKILL.md`, applied repo-wide): `t.Parallel()` top level and in subtests; named table cases with lowercase underscore names (`compile_error`, `missing_import`, `import_cycle`, `syntax_beats_compile`, `tab_utf8_position`, `blocked_import`, `transitive_missing_depth2`, `wkt_resolves`, `deterministic_order_twice`); `require` for fatal preconditions, `assert` for independent checks; per-case `t.TempDir()`; run under `-race -count=1` with a ≥2-file corpus (Pitfall 7). Fixture-file loading pattern from `internal/core/position_parity_test.go:24-29`: `os.ReadFile("testdata/position_contract_utf8.proto")` — relative to the test's package dir; for the matrix prefer `t.TempDir()` inline sources (etalon style) plus a `testdata/link/` tree for the tab/UTF-8 and multi-file cases.

---

### `internal/core/testdata/link/` (test fixtures)

**Analog:** `internal/core/testdata/position_contract_utf8.proto` — the established pattern: plain fixture files under the package's `testdata/`, referenced by relative path from tests. Layout is planner discretion (RESEARCH D-05 / Claude's Discretion); minimum cells implied by decisions: compiling corpus (incl. WKT import `google/protobuf/timestamp.proto`), compile-error corpus (incl. dep-file semantic error, import cycle, `"../x.proto"` non-local import, tab/UTF-8 line), missing-import corpus (direct + transitive depth-2 + blocked-by-`importFileAllowed`), read/syntax-error corpus.

---

### `internal/api/lint_link_matrix_test.go` (test, Action-path + subprocess exit codes)

**Analog:** `internal/api/lint_v1_test.go` — the established Action-path test harness.

**In-process Action invocation pattern** (`lint_v1_test.go:24-27`, also 53-56, 72-75):
```go
	ctx := cli.NewContext(&cli.App{}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
	ctx.Context = t.Context()

	err := (Lint{}).actionV1(ctx, logger.NewNop(), configPath, root, root)
	require.Error(t, err)
```
Matrix cells assert the routed error type: missing import → `require.ErrorAs(err, &*core.OpenImportFileError)` (exit-2 class); syntax/read error → `require.Error(err)` without `OpenImportFileError` (exit-1/command-error class); clean corpus → `require.ErrorIs(err, ErrHasLintIssue)` after issues print (exit-1 class). Fixtures: `os.WriteFile` into `t.TempDir()` with an `easyp.yaml` (`"version: v1\nlint:\n  use: [FILE_LOWER_SNAKE_CASE]\n"` or `linters:\n  default: MINIMAL`) — see `lint_v1_test.go:20-22, 67-71`.

**Why subprocess for actual exit codes** — the route under test (`internal/api/lint.go:80-98`):
```go
func (l Lint) Action(ctx *cli.Context) error {
	log := getLogger(ctx)

	err := l.action(ctx, log)
	if err != nil {
		var e *core.OpenImportFileError

		switch {
		case errors.Is(err, ErrHasLintIssue):
			os.Exit(1)
		case errors.As(err, &e):
			errExit(log, 2, "Cannot import file", slog.String("file name", e.FileName))
		default:
			return err
		}
	}

	return nil
}
```
`os.Exit(1)` (`lint.go:89`) and `errExit` (`internal/api/runtime.go:36-39`):
```go
func errExit(log logger.Logger, code int, msg string, attrs ...slog.Attr) {
	log.Error(context.Background(), msg, attrs...)
	os.Exit(code)
}
```
kill the test process — in-process tests can only assert error types; actual codes 0/1/2 and stdout bytes require building the binary (`go build -o easyp ./cmd/easyp`) and running it via `exec.Cmd` with per-child `Dir`/`Env` (go-testing skill: "Test paths that call os.Exit or default CLI exit handling in subprocesses"). Per Assumption A1: classes 2/3 run through this full path now; class 1 (COMPILE) is verified at classifier level + `printIssues` golden (below), with the full Action re-run in Phase 5.

**printIssues golden target for COMPILE shape** (`internal/api/lint.go:108-154`, unchanged code the golden pins):
```go
		str := fmt.Sprintf("%s:%d:%d:%s %s (%s)",
			issue.Path,
			issue.Position.Line,
			issue.Position.Column,
			issue.SourceName,
			issue.Message,
			issue.RuleName,
		)
```
Empty `SourceName` yields the `"path:4:3: message (COMPILE)"`-style line (double colon-space artifact as in main for empty source names); JSON printer (`json.NewEncoder(w).Encode(issue)`) emits zero `Filename`/`Offset` fields. Wire ordering: api already sorts by Path (`lint_v1.go:220-222`); the classifier's D-03 sort must compose with it (determinism cell: run matrix twice, identical order).

---

## Shared Patterns

### Error wrapping
**Source:** `.agents/skills/go-code-style/SKILL.md` + `internal/core/proto_info_read.go:98, 122`
**Apply to:** classifier and resolver code
```go
return nil, fmt.Errorf("readProtoFile: %w", err)   // callee name only, no package/receiver prefix
return nil, fmt.Errorf("ReadAll: %w", err)         // as the etalon resolver does
```

### Typed error seam (`errors.As`/`errors.Is`, no new duplicate sentinels)
**Source:** etalon `link.go:72-76` + `internal/core/dom.go:178-184` + `.spec/ERRORS.md`
**Apply to:** missing-import detection (D-01) and every matrix assertion
```go
var notFound *OpenImportFileError
if errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist) {
	r.missing.Store(true)
	return nil, protoc.ErrFileNotFound
}
```
Do not invent a parallel "linker not found" type unless Phase 3/5 needs to distinguish link-time from read-time missing (RESEARCH Don't-Hand-Roll, Open Question 3 — minimal wrap/re-throw of `OpenImportFileError` is the recommendation).

### Resource cleanup with logged close error
**Source:** `internal/core/lint.go:95-105`, used by etalon resolver line 79
**Apply to:** resolver file reads
```go
defer r.core.close(ctx, f, path)   // Core.close wraps Close and logs failures
```

### Deterministic ordering (`slices.SortStableFunc`)
**Source:** `internal/api/lint_v1.go:219-222`
**Apply to:** COMPILE diag→issue conversion (D-03 sort by Path/Line/Col in the classifier)

### Test conventions
**Source:** `.agents/skills/go-testing/SKILL.md`, `internal/core/proto_info_read_test.go`, `internal/api/lint_v1_test.go`, `internal/core/position_parity_test.go`
**Apply to:** both new test files — testify (`require` fatal / `assert` non-fatal), `t.Parallel()` except process-state cases (subprocess helpers stay isolated), named table cases, `t.TempDir()`, same-package tests for unexported classifier, `t.Context()`, `logger.NewNop()`, hand-written doubles only.

### Import grouping
**Source:** `.agents/skills/go-code-style/SKILL.md` + `internal/core/lint.go:3-14`
**Apply to:** all new files — stdlib / third-party / `github.com/easyp-tech/easyp/...` groups separated by blank lines; English godoc starting with the symbol name on every export.

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `internal/core/link.go` | service | transform + event-driven | No in-repo consumer of `protoc.Link` exists yet (this is the first). Covered by the external etalon (`compiler-lab .../internal/core/link.go`, quoted above and in `02-RESEARCH.md`) plus in-repo role-matches; planner ports from the etalon excerpts, not from any repo file. |

All other files have tracked in-repo analogs (table above).

## Metadata

**Analog search scope:** `internal/core/`, `internal/api/`, `internal/easypc/protoc/`, `.agents/skills/` (go-code-style, go-testing), lab clone `/tmp/gsd-research/compiler-lab/easyp-cli/internal/core/`
**Files scanned:** 16 (11 tracked repo files read, 3 lab files read, 2 skill docs)
**Git-tracking verification:** `git ls-files` on `internal/core/` + `internal/api/` — all named repo paths tracked; `internal/easypc/**` sealed (signatures only); lab paths external, flagged as porting source
**Pattern extraction date:** 2026-11-19
