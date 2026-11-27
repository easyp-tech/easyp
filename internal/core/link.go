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

// linkSources links the linted files with their imports. Linted files come
// from memory; imports are found through openImportFile. When the files fail
// semantic compilation it returns the compiler's errors as COMPILE issues.
// When an import is missing/broken or a file has a syntax error, it returns
// neither so lint falls back to the per-file path.
func (c *Core) linkSources(ctx context.Context, disk FS, paths []string, sources [][]byte) (*linkedFiles, []IssueInfo) {
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
		return &linkedFiles{linked: linked, outlines: make(map[string]*linkedOutline)}, nil
	}
	if resolver.missing.Load() || resolver.broken.Load() || !errors.Is(err, protoc.ErrCompilationFailed) || len(diags) == 0 {
		c.logger.Debug(ctx, "link failed, lint rules use the syntax only", slog.String("error", err.Error()))
		return nil, nil
	}
	if hasPerFileSyntaxError(paths, sources) {
		c.logger.Debug(ctx, "link failed, lint rules use the syntax only", slog.String("error", err.Error()))
		return nil, nil
	}

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

func hasPerFileSyntaxError(paths []string, sources [][]byte) bool {
	for i, path := range paths {
		if _, err := protoast.Parse(path, sources[i]); err != nil {
			return true
		}
	}
	return false
}

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

type lintResolver struct {
	core    *Core
	disk    FS
	linted  map[string][]byte
	missing atomic.Bool // an import could not be found
	broken  atomic.Bool // an import failed to open or read for a non-notFound reason
}

// FindFileByPath implements protoc.Resolver.
func (r *lintResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if src, ok := r.linted[path]; ok {
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
		var pathErr *os.PathError
		if errors.As(err, &pathErr) ||
			errors.Is(err, sourceview.ErrUnsupported) ||
			errors.Is(err, sourceview.ErrChanged) ||
			errors.Is(err, sourceview.ErrNestedRepository) {
			r.broken.Store(true)
		}
		return nil, err
	}
	defer r.core.close(ctx, f, path)
	src, err := io.ReadAll(f)
	if err != nil {
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

// protoInfoFromLinked builds a linted file's model from its linked tree, then
// drops the tree; its outline stays available to files importing it.
func (c *Core) protoInfoFromLinked(disk FS, linked *linkedFiles, path string) (ProtoInfo, error) {
	file, ok := linked.linked.File(path)
	if !ok || file.Syntax == nil {
		return ProtoInfo{}, fmt.Errorf("Parse: %s was not linked", path)
	}
	model := protoast.Build(file.Syntax)
	linked.outline(path)
	file.Syntax = nil

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
