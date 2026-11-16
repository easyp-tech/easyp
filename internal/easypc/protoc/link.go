package protoc

import (
	"context"
	"errors"
	"runtime"
	"strings"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/compiler"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/linker"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
	"github.com/easyp-tech/easyp/internal/easypc/protoc/syntax"
)

// Linked holds name resolution for a set of files and everything they import,
// without building descriptors.
type Linked struct {
	files map[string]*LinkedFile
}

// LinkedFile describes one linked file.
type LinkedFile struct {
	Path    string
	Imports []LinkedImport
	// Uses holds the paths of files defining the types, extendees, rpc types
	// and option extensions this file refers to, other than itself.
	Uses map[string]bool
	// Methods lists rpc methods in declaration order.
	Methods []LinkedMethod
	// Syntax is the parsed file, kept with WithSyntaxTrees. Linking rewrites
	// resolved names in the tree but leaves its source spans and tokens intact.
	Syntax *syntax.File
}

// LinkedImport is an import statement of a linked file.
type LinkedImport struct {
	Path   string
	Public bool
}

// LinkedMethod is an rpc method with its resolved request and response types.
type LinkedMethod struct {
	Offset uint32 // byte offset of the rpc keyword
	Input  string // fully-qualified name, without a leading dot
	Output string
}

// File returns the linked file at path, as files are named in imports.
func (l *Linked) File(path string) (*LinkedFile, bool) {
	file, ok := l.files[path]
	return file, ok
}

// Link parses and links files and their imports. Files are read only through
// the resolvers given with WithResolver and, if enabled, the standard imports;
// it never falls back to the file system. Diagnostics go to WithErrorHandler
// and failures return ErrCompilationFailed.
func Link(ctx context.Context, files []string, opts ...Option) (*Linked, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	if len(files) == 0 {
		return nil, errors.New("no input files provided")
	}
	cfg := &compileConfig{variant: "auto", strategy: "auto"}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
	importPaths, readFile, resolvePath := cfg.sources(ctx, false)
	threads := cfg.concurrency
	if threads <= 0 {
		threads = runtime.NumCPU()
	}

	var table *linker.SymbolTable
	driver := compiler.NewDriver(&compiler.Options{
		ImportPaths:   importPaths,
		Threads:       threads,
		Variant:       cfg.variant,
		Strategy:      cfg.strategy,
		InputFiles:    files,
		Context:       ctx,
		ReadFile:      readFile,
		ResolvePath:   resolvePath,
		SymTableOut:   &table,
		SkipSerialize: true,
		TrackUses:     true,
		KeepTokens:    cfg.syntaxTrees,
		ErrorHandler: func(diag parser.Diagnostic) {
			if cfg.errorHandler != nil {
				cfg.errorHandler(fromInternalDiag(diag))
			}
		},
	})
	if err := driver.Compile(); err != nil {
		if errors.Is(err, compiler.ErrCompilationFailed) {
			return nil, ErrCompilationFailed
		}
		return nil, err
	}

	linked := &Linked{files: make(map[string]*LinkedFile, len(table.Files))}
	for id, ast := range table.Files {
		file := &LinkedFile{Path: ast.Path, Uses: make(map[string]bool)}
		if cfg.syntaxTrees {
			tree, err := syntax.FromAST(ast)
			if err != nil {
				return nil, err
			}
			file.Syntax = tree
		}
		for _, imp := range ast.Imports {
			file.Imports = append(file.Imports, LinkedImport{
				Path:   imp.Path.String(ast.Src),
				Public: imp.Kind == parser.ImportPublic,
			})
		}
		for _, used := range table.UsedFiles(uint32(id)) {
			if int(used) != id {
				file.Uses[table.Files[used].Path] = true
			}
		}
		for i := range ast.Methods {
			m := &ast.Methods[i]
			file.Methods = append(file.Methods, LinkedMethod{
				Offset: m.Span.Start,
				Input:  strings.TrimPrefix(m.InputType.String(ast.Src), "."),
				Output: strings.TrimPrefix(m.OutputType.String(ast.Src), "."),
			})
		}
		linked.files[ast.Path] = file
	}
	return linked, nil
}
