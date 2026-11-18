package protoc

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/compiler"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/linker"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/options"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/sourceinfo"
)

// LSPSession maintains an in-memory indexed workspace (AST slabs + Global Symbol Table + O(1) visibility bitsets)
// for ultra-fast incremental textDocument/didChange re-validation and SourceCodeInfo hover/location queries.
type LSPSession struct {
	mu          sync.Mutex
	symTable    *linker.SymbolTable
	importPaths []string
	variant     string
}

// NewLSPSession indexes a workspace of .proto files in parallel and retains the Global Symbol Table in memory.
func NewLSPSession(ctx context.Context, files []string, opts ...Option) (*LSPSession, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if len(files) == 0 {
		return nil, errors.New("no input files provided")
	}

	cfg := &compileConfig{
		variant:  "auto",
		strategy: "auto",
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	importPaths := cfg.importPaths
	if len(importPaths) == 0 {
		importPaths = []string{"."}
	}

	threads := cfg.concurrency
	if threads <= 0 {
		threads = runtime.NumCPU()
	}

	var orderedFiles []*parser.ASTFile
	var symTable *linker.SymbolTable

	driverOpts := &compiler.Options{
		ImportPaths:       importPaths,
		IncludeImports:    true,
		IncludeSourceInfo: cfg.includeSourceInfo,
		Threads:           threads,
		Variant:           cfg.variant,
		Strategy:          cfg.strategy,
		InputFiles:        files,
		Context:           ctx,
		OrderedFiles:      &orderedFiles,
		SymTableOut:       &symTable,
		SkipSerialize:     true,
		ErrorHandler: func(diag parser.Diagnostic) {
			if cfg.errorHandler != nil {
				cfg.errorHandler(fromInternalDiag(diag))
			}
		},
	}

	driver := compiler.NewDriver(driverOpts)
	if err := driver.Compile(); err != nil {
		return nil, err
	}

	return &LSPSession{
		symTable:    symTable,
		importPaths: importPaths,
		variant:     cfg.variant,
	}, nil
}

// DidChange incrementally re-lexes, re-parses, re-links, and optionally regenerates SourceCodeInfo
// for a single modified file against the already-indexed workspace SymbolTable in O(file_size).
func (s *LSPSession) DidChange(canonicalPath string, src []byte, includeSourceInfo bool) ([]Diagnostic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.symTable == nil {
		return nil, errors.New("LSP session not initialized")
	}
	fileID, ok := s.symTable.FileIDs[canonicalPath]
	if !ok {
		return nil, fmt.Errorf("file %q is not in workspace", canonicalPath)
	}

	// 1. Lex
	ts := lexer.NewTokenStream(len(src)/16+8, len(src)/64+8)
	if err := lexer.Lex(src, ts); err != nil {
		return nil, err
	}

	// 2. Parse
	p := parser.NewParser(src, ts, canonicalPath)
	ast, err := p.ParseFile()
	if err != nil && len(ast.Errors) == 0 {
		return nil, err
	}
	if len(ast.Errors) > 0 {
		diags := make([]Diagnostic, len(ast.Errors))
		for i, d := range ast.Errors {
			diags[i] = fromInternalDiag(d)
		}
		return diags, nil
	}

	// 3. Re-index symbols & resolve types against global SymbolTable (O(1) visibility bitset)
	s.symTable.ReindexFile(fileID, ast)
	s.symTable.ResolveFile(fileID, ast)

	// 4. Evaluate options for the edited file
	options.EvaluateFile(s.symTable, fileID, ast)

	// 5. Optionally generate SourceCodeInfo (hover comments & AST location spans)
	if includeSourceInfo {
		sci := sourceinfo.Generate(ast, fileID, s.symTable)
		ast.SourceInfoWire = sourceinfo.Serialize(sci)
	}
	ast.Tokens = nil

	if len(ast.Errors) > 0 {
		diags := make([]Diagnostic, len(ast.Errors))
		for i, d := range ast.Errors {
			diags[i] = fromInternalDiag(d)
		}
		return diags, nil
	}
	return nil, nil
}
