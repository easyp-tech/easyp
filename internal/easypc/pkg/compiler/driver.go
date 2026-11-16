package compiler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/linker"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/metrics"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/options"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/serialize"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/sourceinfo"
)

// ErrCompilationFailed indicates that syntax or semantic errors occurred and were printed.
var ErrCompilationFailed = errors.New("compilation failed")

// Driver coordinates file reading, lexing, parsing, and wire serialization.
type Driver struct {
	opts    *Options
	parsed  map[string]*parser.ASTFile
	perFile map[string]*metrics.File
}

// NewDriver constructs a compiler driver with the given options.
func NewDriver(opts *Options) *Driver {
	return &Driver{
		opts:    opts,
		parsed:  make(map[string]*parser.ASTFile),
		perFile: make(map[string]*metrics.File),
	}
}

// Compile executes the compilation pipeline.
func (d *Driver) Compile() error {
	startTime := time.Now()

	if d.opts != nil {
		if d.opts.Context != nil {
			if err := d.opts.Context.Err(); err != nil {
				return err
			}
		}
		if err := validateVariant(d.opts.Variant); err != nil {
			return err
		}
		if err := validateStrategy(d.opts.Strategy); err != nil {
			return err
		}
	}

	var orderedFiles []*parser.ASTFile
	visited := make(map[string]bool, len(d.opts.InputFiles))
	visiting := make(map[string]bool)
	var visitStack []string
	var displayStack []string

	var totalReadNs, totalLexNs, totalParseNs int64
	var totalInputBytes int64

	threads := runtime.NumCPU()
	if d.opts != nil {
		if d.opts.Threads > 0 {
			threads = d.opts.Threads
		} else if d.opts.MetricsMode == "trace" {
			threads = 1
		}
	}

	type preloadEntry struct {
		canonicalPath string
		fullPath      string
		ast           *parser.ASTFile
		resolveErr    error
		loadErr       error
		readNs        int64
		lexNs         int64
		parseNs       int64
		inputBytes    int64
	}

	loadAndParseRaw := func(canonicalPath, fullPath, displayPath string) (*parser.ASTFile, int64, int64, int64, int64, error) {
		if d.opts != nil && d.opts.Context != nil {
			if err := d.opts.Context.Err(); err != nil {
				return nil, 0, 0, 0, 0, err
			}
		}

		// Read phase
		tRead := time.Now()
		var src []byte
		var err error
		if d.opts != nil && d.opts.ReadFile != nil {
			src, err = d.opts.ReadFile(canonicalPath)
		} else {
			src, err = os.ReadFile(fullPath)
		}
		if err != nil {
			return nil, 0, 0, 0, 0, fmt.Errorf("%s: %w", canonicalPath, err)
		}
		readNs := time.Since(tRead).Nanoseconds()
		inBytes := int64(len(src))

		// Lex phase
		tLex := time.Now()
		ts := lexer.NewTokenStream(len(src)/16+8, len(src)/64+8)
		variant := "auto"
		if d.opts != nil && d.opts.Variant != "" {
			variant = d.opts.Variant
		}
		switch variant {
		case "auto", "scalar", "swar", "neon", "sve2", "avx2", "avx512":
			if err := lexer.Lex(src, ts); err != nil {
				return nil, readNs, 0, 0, inBytes, fmt.Errorf("%s: lex error: %w", canonicalPath, err)
			}
		default:
			return nil, readNs, 0, 0, inBytes, fmt.Errorf("%s: unsupported variant %q", canonicalPath, variant)
		}
		lexNs := time.Since(tLex).Nanoseconds()

		// Parse phase
		tParse := time.Now()
		p := parser.NewParser(src, ts, canonicalPath)
		p.SetDisplayPath(displayPath)
		ast, err := p.ParseFile()
		if err != nil && len(ast.Errors) == 0 {
			return nil, readNs, lexNs, 0, inBytes, fmt.Errorf("%s: parse error: %w", canonicalPath, err)
		}
		parseNs := time.Since(tParse).Nanoseconds()

		if (d.opts == nil || (!d.opts.IncludeSourceInfo && !d.opts.KeepTokens)) && len(ast.Errors) == 0 && len(ast.Imports) == 0 {
			ast.Tokens = nil
		}
		return ast, readNs, lexNs, parseNs, inBytes, nil
	}

	var preloadedByInput map[string]*preloadEntry
	var preloadedByCanonical map[string]*preloadEntry

	if threads > 1 && len(d.opts.InputFiles) > 1 && (d.opts == nil || d.opts.MetricsMode != "trace") {
		nInputs := len(d.opts.InputFiles)
		entries := make([]preloadEntry, nInputs)
		var nextIdx atomic.Int64
		workerCount := threads
		if workerCount > nInputs {
			workerCount = nInputs
		}
		var wg sync.WaitGroup
		wg.Add(workerCount)
		for w := 0; w < workerCount; w++ {
			go func() {
				defer wg.Done()
				for {
					idx := int(nextIdx.Add(1) - 1)
					if idx >= nInputs {
						return
					}
					inp := d.opts.InputFiles[idx]
					cPath, fPath, rErr := d.resolvePath(inp)
					if rErr != nil {
						entries[idx] = preloadEntry{resolveErr: rErr}
						continue
					}
					ast, rNs, lNs, pNs, inB, lErr := loadAndParseRaw(cPath, fPath, fPath)
					entries[idx] = preloadEntry{
						canonicalPath: cPath,
						fullPath:      fPath,
						ast:           ast,
						loadErr:       lErr,
						readNs:        rNs,
						lexNs:         lNs,
						parseNs:       pNs,
						inputBytes:    inB,
					}
				}
			}()
		}
		wg.Wait()

		preloadedByInput = make(map[string]*preloadEntry, nInputs)
		preloadedByCanonical = make(map[string]*preloadEntry, nInputs)
		for i := range entries {
			e := &entries[i]
			inp := d.opts.InputFiles[i]
			if _, exists := preloadedByInput[inp]; !exists {
				preloadedByInput[inp] = e
			}
			if e.resolveErr == nil && e.canonicalPath != "" {
				if _, exists := preloadedByCanonical[e.canonicalPath]; !exists {
					preloadedByCanonical[e.canonicalPath] = e
				}
			}
		}
	}

	// parseOne loads, lexes, and parses a single file.
	parseOne := func(canonicalPath, fullPath, displayPath string) (*parser.ASTFile, error) {
		if ast, ok := d.parsed[canonicalPath]; ok {
			return ast, nil
		}

		var ast *parser.ASTFile
		var readNs, lexNs, parseNs, inBytes int64
		var err error

		if preloadedByCanonical != nil {
			if e, ok := preloadedByCanonical[canonicalPath]; ok {
				delete(preloadedByCanonical, canonicalPath)
				ast = e.ast
				readNs = e.readNs
				lexNs = e.lexNs
				parseNs = e.parseNs
				inBytes = e.inputBytes
				err = e.loadErr
			}
		}
		if ast == nil && err == nil {
			ast, readNs, lexNs, parseNs, inBytes, err = loadAndParseRaw(canonicalPath, fullPath, displayPath)
		}
		totalReadNs += readNs
		totalLexNs += lexNs
		totalParseNs += parseNs
		totalInputBytes += inBytes
		if err != nil {
			return nil, err
		}

		// Report syntax/semantic diagnostics
		if len(ast.Errors) > 0 {
			for _, diag := range ast.Errors {
				if d.opts != nil && d.opts.ErrorHandler != nil {
					d.opts.ErrorHandler(diag)
				} else {
					fmt.Fprintln(os.Stderr, diag.String())
				}
			}
			return nil, ErrCompilationFailed
		}

		d.parsed[canonicalPath] = ast

		if d.opts != nil && d.opts.MetricsMode == "trace" {
			lines := int64(0)
			if ast.Tokens != nil {
				lines = int64(len(ast.Tokens.Lines))
			}
			d.perFile[canonicalPath] = &metrics.File{
				Path:  canonicalPath,
				Bytes: inBytes,
				Lines: lines,
				Phases: map[string]int64{
					metrics.PhaseRead:  readNs,
					metrics.PhaseLex:   lexNs,
					metrics.PhaseParse: parseNs,
				},
				TotalNs: readNs + lexNs + parseNs,
			}
		}

		return ast, nil
	}

	// DFS visit function for dependency order
	var visit func(inputPath string) error
	visit = func(inputPath string) error {
		if visited[inputPath] {
			return nil
		}
		var canonicalPath, fullPath string
		var err error
		if preloadedByCanonical != nil {
			if e, ok := preloadedByCanonical[inputPath]; ok && e.resolveErr == nil {
				canonicalPath = e.canonicalPath
				fullPath = e.fullPath
			}
		}
		if canonicalPath == "" && preloadedByInput != nil {
			if e, ok := preloadedByInput[inputPath]; ok {
				if e.resolveErr != nil {
					return e.resolveErr
				}
				canonicalPath = e.canonicalPath
				fullPath = e.fullPath
			}
		}
		if canonicalPath == "" {
			if existing, ok := d.parsed[inputPath]; ok {
				canonicalPath = existing.Path
				fullPath = existing.DisplayPath
			} else {
				canonicalPath, fullPath, err = d.resolvePath(inputPath)
				if err != nil {
					return err
				}
			}
		}

		if visited[canonicalPath] {
			return nil
		}
		if visiting[canonicalPath] {
			idx := 0
			for i, p := range visitStack {
				if p == canonicalPath {
					idx = i
					break
				}
			}
			cycleChain := strings.Join(append(visitStack[idx:], canonicalPath), " -> ")
			rootDisplay := displayStack[idx]
			diag := parser.Diagnostic{
				Path:    rootDisplay,
				Line:    2,
				Col:     1,
				Message: fmt.Sprintf("File recursively imports itself: %s", cycleChain),
			}
			if d.opts != nil && d.opts.ErrorHandler != nil {
				d.opts.ErrorHandler(diag)
			} else {
				fmt.Fprintln(os.Stderr, diag.String())
			}
			return ErrCompilationFailed
		}
		visiting[canonicalPath] = true
		visitStack = append(visitStack, canonicalPath)
		displayStack = append(displayStack, fullPath)

		ast, err := parseOne(canonicalPath, fullPath, fullPath)
		if err != nil {
			return err
		}

		for _, imp := range ast.Imports {
			impPath := imp.Path.String(ast.Src)
			if err := visit(impPath); err != nil {
				line, col := 2, 1
				if ast.Tokens != nil {
					line, col = ast.Tokens.LineCol(ast.Src, imp.Span.Start)
				}
				diag := parser.Diagnostic{
					Path:    ast.DisplayPath,
					Line:    line,
					Col:     col,
					Message: fmt.Sprintf("Import %q was not found or had errors.", impPath),
				}
				if d.opts != nil && d.opts.ErrorHandler != nil {
					d.opts.ErrorHandler(diag)
				} else {
					fmt.Fprintln(os.Stderr, diag.String())
				}
				return ErrCompilationFailed
			}
		}

		if d.opts == nil || (!d.opts.IncludeSourceInfo && !d.opts.KeepTokens) {
			ast.Tokens = nil
		}

		displayStack = displayStack[:len(displayStack)-1]
		visitStack = visitStack[:len(visitStack)-1]
		visiting[canonicalPath] = false
		visited[canonicalPath] = true
		orderedFiles = append(orderedFiles, ast)
		return nil
	}

	for _, inputFile := range d.opts.InputFiles {
		if err := visit(inputFile); err != nil {
			return err
		}
	}

	if d.opts != nil && d.opts.Context != nil {
		if err := d.opts.Context.Err(); err != nil {
			return err
		}
	}

	// Linking phase
	tLink := time.Now()
	symTable := linker.NewSymbolTable(orderedFiles, d.parsed)
	if d.opts != nil && d.opts.TrackUses {
		symTable.TrackUses()
	}
	symTable.ResolveTypes()
	symTable.Validate()
	linkNs := time.Since(tLink).Nanoseconds()

	hasLinkingErrors := false
	for _, f := range orderedFiles {
		if len(f.Errors) > 0 {
			hasLinkingErrors = true
			for _, diag := range f.Errors {
				if d.opts != nil && d.opts.ErrorHandler != nil {
					d.opts.ErrorHandler(diag)
				} else {
					fmt.Fprintln(os.Stderr, diag.String())
				}
			}
		}
	}
	if hasLinkingErrors {
		return ErrCompilationFailed
	}

	// Options evaluation phase
	tOptions := time.Now()
	options.Evaluate(symTable)
	symTable.StopTrackingUses()
	optionsNs := time.Since(tOptions).Nanoseconds()

	hasOptionErrors := false
	for _, f := range orderedFiles {
		if len(f.Errors) > 0 {
			hasOptionErrors = true
			for _, diag := range f.Errors {
				if d.opts != nil && d.opts.ErrorHandler != nil {
					d.opts.ErrorHandler(diag)
				} else {
					fmt.Fprintln(os.Stderr, diag.String())
				}
			}
		}
	}
	if hasOptionErrors {
		return ErrCompilationFailed
	}

	// SourceCodeInfo generation phase
	var sourceInfoNs int64
	if d.opts != nil && d.opts.IncludeSourceInfo {
		tSourceInfo := time.Now()
		if threads > 1 && len(orderedFiles) > 1 && d.opts.MetricsMode != "trace" {
			nFiles := len(orderedFiles)
			workerCount := threads
			if workerCount > nFiles {
				workerCount = nFiles
			}
			var nextFileID atomic.Int64
			var wg sync.WaitGroup
			wg.Add(workerCount)
			for w := 0; w < workerCount; w++ {
				go func() {
					defer wg.Done()
					for {
						fileID := int(nextFileID.Add(1) - 1)
						if fileID >= nFiles {
							return
						}
						f := orderedFiles[fileID]
						sci := sourceinfo.Generate(f, uint32(fileID), symTable)
						f.SourceInfoWire = sourceinfo.Serialize(sci)
						f.Tokens = nil
					}
				}()
			}
			wg.Wait()
		} else {
			for fileID, f := range orderedFiles {
				tF := time.Now()
				sci := sourceinfo.Generate(f, uint32(fileID), symTable)
				f.SourceInfoWire = sourceinfo.Serialize(sci)
				f.Tokens = nil
				if d.opts.MetricsMode == "trace" {
					if mf, ok := d.perFile[f.Path]; ok {
						siNs := time.Since(tF).Nanoseconds()
						mf.Phases[metrics.PhaseSourceInfo] = siNs
						mf.TotalNs += siNs
					}
				}
			}
		}
		sourceInfoNs = time.Since(tSourceInfo).Nanoseconds()
	}
	if d.opts != nil && d.opts.SymTableOut != nil {
		*d.opts.SymTableOut = symTable
	}
	symTable = nil

	if d.opts != nil && d.opts.SkipSerialize {
		if d.opts.OrderedFiles != nil {
			*d.opts.OrderedFiles = orderedFiles
		}
		return nil
	}

	// Output files selection
	var filesToSerialize []*parser.ASTFile
	if d.opts != nil && d.opts.IncludeImports {
		filesToSerialize = orderedFiles
	} else if d.opts != nil {
		filesToSerialize = make([]*parser.ASTFile, 0, len(d.opts.InputFiles))
		seen := make(map[string]bool, len(d.opts.InputFiles))
		for _, inputFile := range d.opts.InputFiles {
			cPath, _, err := d.resolvePath(inputFile)
			if err != nil {
				return err
			}
			if !seen[cPath] {
				seen[cPath] = true
				if ast, ok := d.parsed[cPath]; ok {
					filesToSerialize = append(filesToSerialize, ast)
				}
			}
		}
	} else {
		filesToSerialize = orderedFiles
	}

	// Serialization phase
	tSer := time.Now()
	s := serialize.NewWireSerializer()
	includeSourceInfo := false
	if d.opts != nil {
		includeSourceInfo = d.opts.IncludeSourceInfo
	}
	outBytes := s.SerializeFileDescriptorSet(filesToSerialize, includeSourceInfo)
	serNs := time.Since(tSer).Nanoseconds()

	// Write phase
	tWrite := time.Now()
	if d.opts != nil && d.opts.DescriptorSetOut != "" {
		if err := os.WriteFile(d.opts.DescriptorSetOut, outBytes, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", d.opts.DescriptorSetOut, err)
		}
	}
	if d.opts != nil && d.opts.OutputBytes != nil {
		*d.opts.OutputBytes = outBytes
	}
	if d.opts != nil && d.opts.OrderedFiles != nil {
		*d.opts.OrderedFiles = orderedFiles
	}
	writeNs := time.Since(tWrite).Nanoseconds()

	totalWallNs := time.Since(startTime).Nanoseconds()

	// Handle metrics report
	if d.opts.MetricsOut != "" {
		report := &metrics.Report{
			Impl:    "easyp-go",
			Mode:    d.opts.MetricsMode,
			Threads: d.opts.Threads,
			Totals: metrics.Totals{
				WallNs: totalWallNs,
				Phases: map[string]int64{
					metrics.PhaseRead:       totalReadNs,
					metrics.PhaseLex:        totalLexNs,
					metrics.PhaseParse:      totalParseNs,
					metrics.PhaseLink:       linkNs,
					metrics.PhaseOptions:    optionsNs,
					metrics.PhaseSourceInfo: sourceInfoNs,
					metrics.PhaseSerialize:  serNs,
					metrics.PhaseWrite:      writeNs,
				},
				Files:       len(orderedFiles),
				InputBytes:  totalInputBytes,
				OutputBytes: int64(len(outBytes)),
			},
		}
		if report.Threads == 0 {
			report.Threads = runtime.GOMAXPROCS(0)
		}

		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		report.Memory = &metrics.Memory{
			TotalAllocBytes: int64(ms.TotalAlloc),
			Mallocs:         int64(ms.Mallocs),
			PeakHeapBytes:   int64(ms.HeapSys),
			GCCycles:        int64(ms.NumGC),
			GCPauseNs:       int64(ms.PauseTotalNs),
		}

		if d.opts.MetricsMode == "trace" {
			report.Threads = 1
			for _, f := range orderedFiles {
				if mf, ok := d.perFile[f.Path]; ok {
					report.Files = append(report.Files, *mf)
				}
			}
		}

		if err := metrics.Write(d.opts.MetricsOut, report); err != nil {
			return fmt.Errorf("failed to write metrics: %w", err)
		}
	}

	return nil
}

func (d *Driver) resolvePath(path string) (canonical string, fullPath string, err error) {
	if d.opts != nil && d.opts.ResolvePath != nil {
		return d.opts.ResolvePath(path)
	}
	// 1. Check relative to each import path in order
	for _, ip := range d.opts.ImportPaths {
		candidate := filepath.Join(ip, path)
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			return path, candidate, nil
		}
	}

	// 2. Check direct filesystem path
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		cleanPath := filepath.Clean(path)
		for _, ip := range d.opts.ImportPaths {
			cleanIP := filepath.Clean(ip)
			if cleanIP == "." {
				continue
			}
			rel, err := filepath.Rel(cleanIP, cleanPath)
			if err == nil && !strings.HasPrefix(rel, "..") && rel != "." {
				return rel, path, nil
			}
		}
		return path, path, nil
	}

	return "", "", fmt.Errorf("%s: File not found.", path)
}
