package core

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/easyp-tech/easyp/internal/easypc/protoc"
	"github.com/easyp-tech/easyp/internal/protoast"
)

// CompileRuleName reports compiler errors found while linking linted files.
const CompileRuleName = "COMPILE"

// linkSources links the corpus with protoc.Link and classifies the outcome:
//
//   - linked != nil, issues == nil: the corpus linked cleanly and the syntax
//     trees are returned for the downstream engine;
//   - linked == nil, issues != nil: the corpus failed linking with compile
//     errors, reported as COMPILE issues in the existing issue model;
//   - both nil: linking could not classify the corpus (a missing import, a
//     non-compilation failure, or no diagnostics) and the caller falls back
//     to the existing per-file read path without COMPILE issues.
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
		protoc.WithErrorHandler(func(d protoc.Diagnostic) {
			if d.IsWarn {
				return
			}
			diagsMu.Lock()
			defer diagsMu.Unlock()
			diags = append(diags, d)
		}))
	if err == nil {
		return linked, nil
	}
	if resolver.missing.Load() || !errors.Is(err, protoc.ErrCompilationFailed) || len(diags) == 0 {
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

// lintResolver serves corpus files from memory and resolves imports through
// the existing openImportFile boundary, adding no filesystem access of its
// own. A missing import sets missing so the classifier can fall back to the
// per-file path instead of reporting COMPILE issues.
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
