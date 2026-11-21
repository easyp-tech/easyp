package core

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/easyp-tech/easyprotoc/protoc"

	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

// CompileRuleName reports compiler errors found while linking linted files.
const CompileRuleName = "COMPILE"

// linkSources links the corpus with protoc.Link and classifies the outcome:
//
//   - linked != nil, issues == nil: the corpus linked cleanly and the syntax
//     trees are returned for the downstream engine;
//   - linked == nil, issues != nil: the corpus failed linking with compile
//     errors, reported as COMPILE issues in the existing issue model;
//   - both nil: linking could not classify the corpus (a missing or broken
//     import, a non-compilation failure, or no diagnostics) and the caller
//     falls back to the existing per-file read path without COMPILE issues.
//
// IN-02 (Phase 5, kept deliberately): COMPILE diagnostics bypass the
// suppression machinery — c.ignore filters the walk, and ignoreOnly and
// comment directives suppress rule findings only. Compiler diagnostics are
// not rule findings: they carry no rule name, their targets include files the
// compiler read outside the corpus, and no suppression directive grammar
// addresses them, so no suppression path may hide them. This matches buf and
// the etalon; pinned by compile_bypasses_suppression and
// ignored_broken_dependency_not_suppressed in lint_linked_matrix_test.go.
func (c *Core) linkSources(ctx context.Context, disk FS, paths []string, sources [][]byte) (*protoc.Linked, []IssueInfo) {
	linted := make(map[string][]byte, len(paths))
	for i, path := range paths {
		linted[path] = sources[i]
	}
	resolver := &lintResolver{core: c, disk: disk, linted: linted}

	var diagsMu sync.Mutex
	var diags []protoc.Diagnostic
	linked, err := protoc.Link(ctx, paths,
		protoc.WithResolver(resolver),
		protoc.WithSyntaxTrees(true),
		protoc.WithErrorHandler(diagCollector(&diagsMu, &diags)))
	if err == nil {
		return linked, nil
	}
	if resolver.missing.Load() || resolver.broken.Load() || !errors.Is(err, protoc.ErrCompilationFailed) || len(diags) == 0 {
		c.logger.Debug(ctx, "link failed, lint rules use the syntax only", slog.String("error", err.Error()))
		return nil, nil
	}

	// A syntax error in any corpus file beats COMPILE: the corpus falls back
	// to the per-file path, where that file's syntax error becomes a command
	// error exactly as before the linker existed.
	if hasPerFileSyntaxError(paths, sources) {
		c.logger.Debug(ctx, "link failed, lint rules use the syntax only", slog.String("error", err.Error()))
		return nil, nil
	}

	// Deterministic issue order: by path, then line, then column.
	slices.SortStableFunc(diags, func(a, b protoc.Diagnostic) int {
		return cmp.Or(
			strings.Compare(a.Path, b.Path),
			cmp.Compare(a.Line, b.Line),
			cmp.Compare(a.Col, b.Col),
		)
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

// hasPerFileSyntaxError reports whether any corpus file fails its individual
// parse. protoc.Diagnostic does not distinguish syntax from semantic failures,
// so the distinction is behavioral: after a failed Link each file is parsed on
// its own, and any syntax error outranks the COMPILE class.
func hasPerFileSyntaxError(paths []string, sources [][]byte) bool {
	for i, path := range paths {
		if _, err := protoast.Parse(path, sources[i]); err != nil {
			return true
		}
	}
	return false
}

// diagCollector returns the diagnostic handler registered with protoc.Link
// through protoc.WithErrorHandler. Compiler warnings are dropped (research
// A3 — the etalon warn-drop policy: warnings never surface as COMPILE
// remarks), errors are collected under mu in arrival order. The collected
// slice is diagnostics' only channel into the D-03 sort and COMPILE issue
// construction.
func diagCollector(mu *sync.Mutex, diags *[]protoc.Diagnostic) func(protoc.Diagnostic) {
	return func(d protoc.Diagnostic) {
		if d.IsWarn {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		*diags = append(*diags, d)
	}
}

// lintResolver serves corpus files from memory and resolves imports through
// the existing openImportFile boundary, adding no filesystem access of its
// own. A missing import sets missing, and an import that failed to open or
// read for a non-notFound reason sets broken, so the classifier can fall back
// to the per-file path instead of reporting COMPILE issues.
type lintResolver struct {
	core    *Core
	disk    FS
	linted  map[string][]byte
	missing atomic.Bool // an import could not be found
	broken  atomic.Bool // an import failed to open or read for a non-notFound reason — never a COMPILE remark
}

// FindFileByPath implements protoc.Resolver.
func (r *lintResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	// IN-01: the ctx guard every sibling resolver carries
	// (github.com/easyp-tech/easyprotoc/protoc/resolver.go) — cancellation is honored before
	// any file access.
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if src, ok := r.linted[path]; ok {
		// WR-02: the corpus branch consults the same importFileAllowed policy
		// openDiskSource applies on disk reads. The rooted full path is built
		// exactly as openDiskSource builds it, so the policy sees the same
		// input shape on both funnels. On denial the file is treated as
		// missing: the classifier falls back to the per-file path, whose
		// readSource yields openDiskSource's exact os.ErrNotExist carrier —
		// the linker can no longer serve a policy-excluded file from memory.
		if rooted, ok := r.disk.(interface{ RootPath() string }); ok && r.core.importFileAllowed != nil {
			if !r.core.importFileAllowed(filepath.Join(rooted.RootPath(), filepath.FromSlash(path))) {
				r.missing.Store(true)
				return nil, protoc.ErrFileNotFound
			}
		}
		return src, nil
	}
	f, err := r.core.openImportFile(r.disk, path)
	if err != nil {
		var notFound *OpenImportFileError
		if errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist) {
			r.missing.Store(true)
			return nil, protoc.ErrFileNotFound
		}
		// WR-01: real open failures arrive in two shapes. A *os.PathError
		// carries the OS cause; the open funnels also fail with the
		// PathError-free sentinel joins sourceview returns — a non-regular
		// target (local.go:42), a target changed during open (:56), a
		// dangling source alias (:176), a nested-repository crossing (:209).
		// Either shape sets broken so the classifier falls back instead of
		// letting the compiler mask the cause as a misleading "not
		// found"-shaped COMPILE remark. The deliberate plain rejections —
		// openImportFile's "invalid import path" and the missing
		// classification above — stay raw, keeping the etalon A4 COMPILE
		// classification for non-local imports.
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			r.broken.Store(true)
		}
		if errors.Is(err, sourceview.ErrUnsupported) ||
			errors.Is(err, sourceview.ErrChanged) ||
			errors.Is(err, sourceview.ErrNestedRepository) {
			r.broken.Store(true)
		}
		return nil, err
	}
	defer r.core.close(ctx, f, path)
	src, err := io.ReadAll(f)
	if err != nil {
		// WR-01: the resolver-owned read failure is a broken import too (its
		// *fs.PathError carries the OS cause); the wrapped error is returned
		// raw either way.
		r.broken.Store(true)
		return nil, fmt.Errorf("ReadAll: %w", err)
	}
	return src, nil
}

// resolution derives the per-file facts rules need from the linked set.
func resolution(linked *protoc.Linked, path string) *Resolution {
	file, ok := linked.File(path)
	if !ok {
		return nil
	}
	res := &Resolution{
		UsedImports: make(map[ImportPath]bool, len(file.Imports)),
		RPCTypes:    make(map[int][2]string, len(file.Methods)),
	}
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

// linkedFiles serves lint models and import outlines from one Link result. It
// wraps the raw *protoc.Linked the linkSources classifier returns.
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
