---
phase: 02-linker-i-marshrutizatsiya-oshibok
reviewed: 2026-11-19T05:31:52Z
depth: deep
files_reviewed: 3
files_reviewed_list:
  - internal/core/link.go
  - internal/core/link_test.go
  - internal/api/lint_link_matrix_test.go
findings:
  critical: 0
  warning: 2
  info: 2
  total: 4
  necessity: 0
status: issues_found
---

# Phase 02 (Linker I: error routing): Code Review Report

**Reviewed:** 2026-11-19T05:31:52Z
**Depth:** deep
**Files Reviewed:** 3
**Status:** issues_found

## Summary

Reviewed the dormant corpus-linker classifier (`internal/core/link.go`) and its two proof-test files at deep depth, tracing every claim across the seams it leans on: `protoc.Link` and `compileConfig.sources` (`internal/easypc/protoc/link.go`, `compile.go`), the vendored driver (`internal/easypc/pkg/compiler/driver.go`), the `openImportFile` boundary (`internal/core/proto_info_read.go`, `dom.go`), `protoast.Parse`, and the api-side routing the matrix test pins (`internal/api/lint.go`, `lint_v1.go`, `runtime.go`).

Verified facts the review rests on:

- **Dormancy (D-04) holds.** `git diff --stat b5c26df..HEAD` touches only the three review files plus testdata/planning artifacts; `internal/core/lint.go`, `internal/api/lint.go`, and `internal/easypc/` are byte-unchanged. `linkSources` is referenced nowhere outside `link.go` and its tests.
- **All three new suites pass** under `go test -race -count=1` (`TestLinkSources`, `TestLintLinkMatrix`/`TestPrintIssuesCompileGolden`, `TestLintLinkMatrixExitCodes`); `go vet` is clean on both packages.
- **Test craftsmanship is strong** and follows `.spec/agent-rules.md` and the go-testing skill: private `flag.FlagSet` + `cli.NewContext` per case, subprocess isolation for exit codes, per-case temp dirs, accurate line-level comments that match the traced driver code (e.g. the cycle diagnostic at `driver.go:319-323`).
- **Necessity assessed, zero findings.** The necessity rubric (`necessity-review.md`) was read. Candidate guards — the `diagsMu` mutex, the `len(diags) == 0` fallback, and the `missing` atomic — all defend behavior of the sealed vendored driver (`internal/easypc`), whose callback contract cannot be relied on across re-syncs; that is boundary defense, not cargo cult. Dormancy of `linkSources` itself is the locked D-04 decision, not dead code.

The two warnings below are both real classification/divergence defects in the classifier — exactly the artifact this phase exists to prove — but neither is reachable in production until the Phase 5 flip, so neither blocks this phase. They must be resolved (or consciously pinned as decisions) before the flip.

## Warnings

### WR-01: Import open/read failures are misclassified as COMPILE instead of the documented fallback

**File:** `internal/core/link.go:116-135`
**Issue:** The classifier's contract (`link.go:29-31`) says `nil, nil` fallback covers "a non-compilation failure", but an import that exists and is merely unreadable is classified COMPILE. Trace, all verified in source: `lintResolver.FindFileByPath` returns the raw non-notFound error (`link.go:127` — e.g. `openImportFile`'s `Open: %w` for EACCES, or `ReadAll: %w` for EISDIR at `link.go:130-133`); `resolvePathFunc` swallows every error and ends in the generic `"%s: File not found."` (`internal/easypc/protoc/compile.go:200-208, 250`); the driver's import loop converts that into the error diagnostic `Import %q was not found or had errors.` plus `ErrCompilationFailed` (`internal/easypc/pkg/compiler/driver.go:341-359`); `linkSources` then passes all three guards (`missing` false, error `Is` `ErrCompilationFailed`, diags non-empty, no corpus syntax error) and emits a COMPILE issue whose message claims the import "was not found". The only non-notFound errors that should keep the COMPILE class are the non-local rejections pinned by research A4 (`non_local_import_compile`) — and those are distinguishable, because `openImportFile` rejects them via `filepath.IsLocal` (`proto_info_read.go:104-106`) before any file access. Post-flip this changes today's observable behavior: a permission-denied import currently exits 1 as a command error carrying the OS cause (`proto_info_read.go:121-122` through `Lint.Action`'s default branch), whereas the linked path would exit 1 as a lint issue with a misleading "not found" message.

**Fix:** Split the third error class in the resolver and honor it in the classifier:

```go
type lintResolver struct {
	core    *Core
	disk    FS
	linted  map[string][]byte
	missing atomic.Bool // an import could not be found
	broken  atomic.Bool // an import failed to open or read for a non-notFound reason
}

func (r *lintResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	if src, ok := r.linted[path]; ok {
		return src, nil
	}
	if !filepath.IsLocal(path) {
		// Non-local imports are a corpus defect: keep the COMPILE class (research A4).
		return nil, protoc.ErrFileNotFound // or pass through openImportFile's rejection
	}
	f, err := r.core.openImportFile(r.disk, path)
	if err != nil {
		var notFound *OpenImportFileError
		switch {
		case errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist):
			r.missing.Store(true)
		default:
			r.broken.Store(true)
		}
		return nil, protoc.ErrFileNotFound
	}
	...
}
```

then in `linkSources` add `resolver.broken.Load()` to the fallback condition at `link.go:55`. The A4 test still passes (its rejection is purely locality-based, the fixture never writes `outside.proto`), and an added cell with an unreadable import pins the new fallback.

### WR-02: lintResolver serves corpus files from memory, bypassing importFileAllowed/sourceFileOpen that the per-file path enforces

**File:** `internal/core/link.go:117-118`
**Issue:** The corpus-hit branch returns the in-memory bytes before any policy check, while the per-file read path the classifier claims to mirror enforces `importFileAllowed` (and `sourceFileOpen`) on corpus files in `openDiskSource` (`internal/core/proto_info_read.go:170-174` — denial becomes `os.ErrNotExist`, i.e. a command error). The production lint Core sets `ImportFileAllowed: importRoots.FileAllowed()` (`internal/api/runtime.go:77`), and for a module-scoped run with explicit declared source roots that function denies paths outside the roots (`internal/modules/sources.go:199-253`; zero roots allow everything, which is why plain project lint never sees this). So a batched corpus file outside the declared roots errors out today, but through the linker it links cleanly and would lint post-flip — a silent file-selection divergence between the two paths in exactly the dimension this phase keeps aligned. Reachable only with restrictive `source_roots` config, and dormant until Phase 5, hence Warning not Critical.

**Fix:** Either enforce the same policy in the resolver's corpus branch — consult `r.core.importFileAllowed` on the rooted full path (mirroring `openDiskSource`), set `missing`/`broken` accordingly, and let the corpus fall back to the per-file path where the denial surfaces with its existing carrier — or, if serving blocked corpus bytes is the intended Phase 5 behavior, pin that decision in the `linkSources` doc comment and add a `blocked_import`-style cell for a blocked *corpus* file so the divergence is a tested decision rather than an accident.

## Info

### IN-01: lintResolver.FindFileByPath ignores ctx, unlike every sibling resolver

**File:** `internal/core/link.go:116`
**Issue:** `MemResolver`, `DirResolver`, `WKTResolver`, and `MultiResolver` all begin with a `ctx.Err()` check (`internal/easypc/protoc/resolver.go:27-97`); `lintResolver` is the only resolver in the codebase that does not, so a cancelled context is noticed at the driver's start/end checks (`driver.go:381-385`) rather than at the next file open. Minor cancellation-promptness inconsistency at a new seam.

**Fix:** Add the same guard as the siblings: `if ctx != nil { if err := ctx.Err(); err != nil { return nil, err } }` at the top of `FindFileByPath`.

### IN-02: COMPILE diagnostics can target files outside the linted corpus and bypass the ignore mechanism

**File:** `internal/core/link.go:77-88`
**Issue:** Every error diagnostic from the whole linked graph is mapped to an issue — including syntax/link errors in dependency files pulled from `importRoots` or WKT that are not corpus members (the driver visits all transitive imports, `driver.go:341-360`, and the linking phase reports `f.Errors` for every ordered file, `driver.go:397-412`). `linkSources` filters none of this, and COMPILE issues will never pass through `Lint`'s `shouldIgnore`/`c.ignore` gate (`internal/core/lint.go:48-50`) because they are emitted outside that loop. Post-flip, a user who ignored a broken dependency file via config would still get COMPILE remarks naming it — protoc-parity, plausibly desired, but it is a behavior decision nothing currently records, and dep-file syntax errors additionally arrive paired with the importer's "Import %q was not found or had errors." remark.

**Fix:** Decide before the Phase 5 flip whether COMPILE remarks honor `c.ignore`/`ignoreOnly` and whether non-corpus paths are in scope; record the choice in the `linkSources` doc comment and pin it with a test cell (ignored dep file with a syntax error) when the flip lands.

---

_Reviewed: 2026-11-19T05:31:52Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: deep_
