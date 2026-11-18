package protoc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/compiler"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

// ErrCompilationFailed indicates that compilation stopped due to syntax or semantic errors.
var ErrCompilationFailed = compiler.ErrCompilationFailed

// Compile compiles the given proto files into a Result.
// It is thread-safe and can be called concurrently from multiple goroutines.
func Compile(ctx context.Context, files []string, opts ...Option) (*Result, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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

	importPaths, readFileFunc, resolvePathFunc := cfg.sources(ctx, true)

	threads := cfg.concurrency
	if threads <= 0 {
		threads = runtime.NumCPU()
	}

	var outBytes []byte
	var orderedFiles []*parser.ASTFile

	useCustomResolvers := len(cfg.resolvers) > 0 || cfg.standardImports
	var readFileCb func(string) ([]byte, error)
	var resolvePathCb func(string) (string, string, error)
	if useCustomResolvers {
		readFileCb = readFileFunc
		resolvePathCb = resolvePathFunc
	}

	driverOpts := &compiler.Options{
		ImportPaths:       importPaths,
		IncludeImports:    cfg.includeImports,
		IncludeSourceInfo: cfg.includeSourceInfo,
		Threads:           threads,
		Variant:           cfg.variant,
		Strategy:          cfg.strategy,
		InputFiles:        files,
		Context:           ctx,
		ReadFile:          readFileCb,
		ResolvePath:       resolvePathCb,
		OutputBytes:       &outBytes,
		OrderedFiles:      &orderedFiles,
		ErrorHandler: func(diag parser.Diagnostic) {
			if cfg.errorHandler != nil {
				cfg.errorHandler(fromInternalDiag(diag))
			}
		},
	}

	driver := compiler.NewDriver(driverOpts)
	if err := driver.Compile(); err != nil {
		if errors.Is(err, compiler.ErrCompilationFailed) {
			return nil, ErrCompilationFailed
		}
		return nil, err
	}

	// Determine canonical paths for filesToGenerate in memory without re-reading files
	knownCanon := make(map[string]string, len(orderedFiles)*2)
	for _, af := range orderedFiles {
		if af != nil {
			knownCanon[af.Path] = af.Path
			if af.DisplayPath != "" {
				knownCanon[af.DisplayPath] = af.Path
			}
		}
	}
	filesToGen := make([]string, 0, len(files))
	for _, f := range files {
		if c, ok := knownCanon[f]; ok {
			filesToGen = append(filesToGen, c)
			continue
		}
		cleanF := strings.TrimPrefix(filepath.Clean(filepath.ToSlash(f)), "./")
		if c, ok := knownCanon[cleanF]; ok {
			filesToGen = append(filesToGen, c)
			continue
		}
		matched := false
		for _, ip := range importPaths {
			cleanIP := filepath.Clean(filepath.ToSlash(ip))
			if cleanIP == "." {
				continue
			}
			if rel, rErr := filepath.Rel(cleanIP, cleanF); rErr == nil && !strings.HasPrefix(rel, "..") && rel != "." {
				relSlash := filepath.ToSlash(rel)
				if c, ok := knownCanon[relSlash]; ok {
					filesToGen = append(filesToGen, c)
					matched = true
					break
				}
			}
		}
		if !matched {
			filesToGen = append(filesToGen, cleanF)
		}
	}

	var allFilesWire []byte
	if cfg.includeImports {
		allFilesWire = outBytes
	}

	return &Result{
		wireBytes:         outBytes,
		filesToGen:        filesToGen,
		allFilesWire:      allFilesWire,
		orderedFiles:      orderedFiles,
		includeSourceInfo: cfg.includeSourceInfo,
	}, nil
}

// sources builds the resolver chain: custom resolvers in order, then one
// directory resolver per import path when withDirs is set (the current
// directory if none were given), then the well-known types if enabled. It
// returns the import paths and the driver's path and read callbacks.
func (cfg *compileConfig) sources(ctx context.Context, withDirs bool) ([]string, func(string) ([]byte, error), func(string) (string, string, error)) {
	// Build resolver chain:
	// 1. Custom resolvers (in order of addition)
	// 2. Directory resolvers for import paths
	// 3. Standard Well-Known Types resolver if enabled
	var resolvers []Resolver
	resolvers = append(resolvers, cfg.resolvers...)

	importPaths := cfg.importPaths
	if len(importPaths) == 0 {
		importPaths = []string{"."}
	}
	if withDirs {
		for _, ip := range importPaths {
			resolvers = append(resolvers, DirResolver(ip))
		}
	}

	if cfg.standardImports {
		resolvers = append(resolvers, WKTResolver{})
	}

	multiRes := MultiResolver(resolvers)

	// In-memory cache for resolved file bytes and canonical mapping for this compilation call
	var fileCacheMu sync.RWMutex
	fileCache := make(map[string][]byte)
	cachePut := func(k string, v []byte) {
		fileCacheMu.Lock()
		fileCache[k] = v
		fileCacheMu.Unlock()
	}
	cacheGet := func(k string) ([]byte, bool) {
		fileCacheMu.RLock()
		b, ok := fileCache[k]
		fileCacheMu.RUnlock()
		return b, ok
	}

	resolvePathFunc := func(path string) (canonical string, fullPath string, err error) {
		if ctx != nil && ctx.Err() != nil {
			return "", "", ctx.Err()
		}

		cleanPath := filepath.Clean(filepath.ToSlash(path))
		cleanPath = strings.TrimPrefix(cleanPath, "./")

		// 1. Check if relative to each import path in order
		for _, ip := range importPaths {
			cleanIP := filepath.Clean(filepath.ToSlash(ip))
			candidate := filepath.Join(cleanIP, cleanPath)
			candidate = strings.TrimPrefix(filepath.ToSlash(candidate), "./")

			b, rErr := multiRes.FindFileByPath(ctx, candidate)
			if rErr == nil {
				cachePut(cleanPath, b)
				return cleanPath, candidate, nil
			}
		}

		// 2. Direct lookup in resolver
		b, rErr := multiRes.FindFileByPath(ctx, cleanPath)
		if rErr == nil {
			cachePut(cleanPath, b)
			canon := cleanPath
			for _, ip := range importPaths {
				cleanIP := filepath.Clean(filepath.ToSlash(ip))
				if cleanIP == "." {
					continue
				}
				rel, rErr := filepath.Rel(cleanIP, cleanPath)
				if rErr == nil && !strings.HasPrefix(rel, "..") && rel != "." {
					canon = filepath.ToSlash(rel)
					break
				}
			}
			cachePut(canon, b)
			return canon, cleanPath, nil
		}

		// 3. Direct filesystem path fallback, only when reading the file system is allowed
		if fi, sErr := os.Stat(path); withDirs && sErr == nil && !fi.IsDir() {
			for _, ip := range importPaths {
				cleanIP := filepath.Clean(filepath.ToSlash(ip))
				if cleanIP == "." {
					continue
				}
				rel, rErr := filepath.Rel(cleanIP, cleanPath)
				if rErr == nil && !strings.HasPrefix(rel, "..") && rel != "." {
					b, err := os.ReadFile(path)
					if err == nil {
						cachePut(filepath.ToSlash(rel), b)
						return filepath.ToSlash(rel), path, nil
					}
				}
			}
			b, err := os.ReadFile(path)
			if err == nil {
				cachePut(cleanPath, b)
				return cleanPath, path, nil
			}
		}

		return "", "", fmt.Errorf("%s: File not found.", path)
	}

	readFileFunc := func(canonicalPath string) ([]byte, error) {
		if ctx != nil && ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if b, ok := cacheGet(canonicalPath); ok {
			return b, nil
		}
		return multiRes.FindFileByPath(ctx, canonicalPath)
	}

	return importPaths, readFileFunc, resolvePathFunc
}
