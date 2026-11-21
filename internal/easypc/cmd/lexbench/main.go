package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
)

type Config struct {
	ImportDirs []string
	Repeat     int
	Threads    int
	Variant    string
	DumpPath   string
	Files      []string
}

type InputFile struct {
	RelPath  string
	FullPath string
	Content  []byte
}

type OutputMetrics struct {
	Impl     string         `json:"impl"`
	Variant  string         `json:"variant"`
	Threads  int            `json:"threads"`
	Files    int            `json:"files"`
	Bytes    int64          `json:"bytes"`
	Lines    int64          `json:"lines"`
	Tokens   int64          `json:"tokens"`
	Errors   int64          `json:"errors"`
	Kinds    map[string]int `json:"kinds"`
	Checksum string         `json:"checksum"`
	BestNs   int64          `json:"best_ns"`
	MedianNs int64          `json:"median_ns"`
	MbPerS   float64        `json:"mb_per_s"`
}

func parseArgfile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read argfile %s: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")
	var args []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "@") {
			nested, err := parseArgfile(line[1:])
			if err != nil {
				return nil, err
			}
			args = append(args, nested...)
		} else {
			args = append(args, line)
		}
	}
	return args, nil
}

func parseArgs(rawArgs []string) (*Config, error) {
	cfg := &Config{
		Repeat:  20,
		Threads: 1,
		Variant: "auto",
	}

	var expanded []string
	for _, arg := range rawArgs {
		if strings.HasPrefix(arg, "@") {
			nested, err := parseArgfile(arg[1:])
			if err != nil {
				return nil, err
			}
			expanded = append(expanded, nested...)
		} else {
			expanded = append(expanded, arg)
		}
	}

	for i := 0; i < len(expanded); i++ {
		arg := expanded[i]
		if arg == "-I" && i+1 < len(expanded) {
			i++
			cfg.ImportDirs = append(cfg.ImportDirs, expanded[i])
		} else if strings.HasPrefix(arg, "-I") {
			cfg.ImportDirs = append(cfg.ImportDirs, arg[2:])
		} else if strings.HasPrefix(arg, "--proto_path=") {
			cfg.ImportDirs = append(cfg.ImportDirs, strings.TrimPrefix(arg, "--proto_path="))
		} else if arg == "--repeat" && i+1 < len(expanded) {
			i++
			r, err := strconv.Atoi(expanded[i])
			if err != nil || r < 1 {
				return nil, fmt.Errorf("invalid repeat: %s", expanded[i])
			}
			cfg.Repeat = r
		} else if strings.HasPrefix(arg, "--repeat=") {
			val := strings.TrimPrefix(arg, "--repeat=")
			r, err := strconv.Atoi(val)
			if err != nil || r < 1 {
				return nil, fmt.Errorf("invalid repeat: %s", val)
			}
			cfg.Repeat = r
		} else if arg == "--threads" && i+1 < len(expanded) {
			i++
			t, err := strconv.Atoi(expanded[i])
			if err != nil || t < 1 {
				return nil, fmt.Errorf("invalid threads: %s", expanded[i])
			}
			cfg.Threads = t
		} else if strings.HasPrefix(arg, "--threads=") {
			val := strings.TrimPrefix(arg, "--threads=")
			t, err := strconv.Atoi(val)
			if err != nil || t < 1 {
				return nil, fmt.Errorf("invalid threads: %s", val)
			}
			cfg.Threads = t
		} else if arg == "--variant" && i+1 < len(expanded) {
			i++
			cfg.Variant = expanded[i]
		} else if strings.HasPrefix(arg, "--variant=") {
			cfg.Variant = strings.TrimPrefix(arg, "--variant=")
		} else if arg == "--dump" && i+1 < len(expanded) {
			i++
			cfg.DumpPath = expanded[i]
		} else if strings.HasPrefix(arg, "--dump=") {
			cfg.DumpPath = strings.TrimPrefix(arg, "--dump=")
		} else if strings.HasPrefix(arg, "-") {
			return nil, fmt.Errorf("unknown flag: %s", arg)
		} else {
			cfg.Files = append(cfg.Files, arg)
		}
	}

	if cfg.Variant != "auto" && cfg.Variant != "scalar" {
		return nil, fmt.Errorf("unsupported variant %q on this CPU (only auto and scalar supported in Go)", cfg.Variant)
	}

	return cfg, nil
}

func resolveFile(path string, importDirs []string) (relPath string, fullPath string, err error) {
	// 1. Check if path exists directly
	if _, errStat := os.Stat(path); errStat == nil {
		clean := filepath.Clean(path)
		// Check if clean matches any importDir
		for _, dir := range importDirs {
			cleanDir := filepath.Clean(dir)
			if rel, errRel := filepath.Rel(cleanDir, clean); errRel == nil && !strings.HasPrefix(rel, "..") {
				return rel, clean, nil
			}
		}
		// If in testdata/lexer/
		if idx := strings.Index(clean, "testdata/lexer/"); idx != -1 {
			return clean[idx+len("testdata/lexer/"):], clean, nil
		}
		return clean, clean, nil
	}

	// 2. Search in importDirs
	for _, dir := range importDirs {
		candidate := filepath.Join(dir, path)
		if _, errStat := os.Stat(candidate); errStat == nil {
			return path, candidate, nil
		}
	}

	return "", "", fmt.Errorf("file not found: %s", path)
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "lexbench: %v\n", err)
		os.Exit(2)
	}

	if len(cfg.Files) == 0 {
		fmt.Fprintf(os.Stderr, "lexbench: no input files\n")
		os.Exit(2)
	}

	// Pre-load all files into memory
	inputFiles := make([]InputFile, 0, len(cfg.Files))
	var totalBytes int64
	for _, f := range cfg.Files {
		rel, full, err := resolveFile(f, cfg.ImportDirs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "lexbench: %v\n", err)
			os.Exit(1)
		}
		content, err := os.ReadFile(full)
		if err != nil {
			fmt.Fprintf(os.Stderr, "lexbench: failed to read %s: %v\n", full, err)
			os.Exit(1)
		}
		inputFiles = append(inputFiles, InputFile{
			RelPath:  rel,
			FullPath: full,
			Content:  content,
		})
		totalBytes += int64(len(content))
	}

	// If --dump is requested, output token dump and exit
	if cfg.DumpPath != "" {
		var out *os.File
		if cfg.DumpPath == "-" {
			out = os.Stdout
		} else {
			f, err := os.Create(cfg.DumpPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "lexbench: cannot create dump file: %v\n", err)
				os.Exit(1)
			}
			defer f.Close()
			out = f
		}

		ts := lexer.NewTokenStream(1024, 128)
		for _, file := range inputFiles {
			ts.Reset()
			if err := lexer.Lex(file.Content, ts); err != nil {
				fmt.Fprintf(os.Stderr, "lexbench: lex error on %s: %v\n", file.RelPath, err)
				os.Exit(1)
			}
			dump := ts.FormatDump(file.RelPath)
			if _, err := out.WriteString(dump); err != nil {
				fmt.Fprintf(os.Stderr, "lexbench: write dump failed: %v\n", err)
				os.Exit(1)
			}
		}
		return
	}

	numFiles := len(inputFiles)
	numThreads := cfg.Threads
	if numThreads > numFiles {
		numThreads = numFiles
	}
	if numThreads < 1 {
		numThreads = 1
	}

	fileChecksums := make([]uint64, numFiles)
	fileLines := make([]int64, numFiles)
	fileTokens := make([]int64, numFiles)
	fileKinds := make([]map[string]int, numFiles)

	durations := make([]time.Duration, cfg.Repeat)

	// Pre-allocate thread-local TokenStream structures
	streams := make([]*lexer.TokenStream, numThreads)
	for i := 0; i < numThreads; i++ {
		streams[i] = lexer.NewTokenStream(2048, 256)
	}

	for r := 0; r < cfg.Repeat; r++ {
		t0 := time.Now()

		if numThreads == 1 {
			ts := streams[0]
			for i := 0; i < numFiles; i++ {
				ts.Reset()
				_ = lexer.Lex(inputFiles[i].Content, ts)
				if r == 0 {
					fileChecksums[i] = ts.FileChecksum()
					fileLines[i] = int64(len(ts.Lines))
					fileTokens[i] = int64(ts.Len())
					fileKinds[i] = ts.KindCounts()
				}
			}
		} else {
			var wg sync.WaitGroup
			var fileIdx uint64

			for t := 0; t < numThreads; t++ {
				wg.Add(1)
				go func(threadID int) {
					defer wg.Done()
					ts := streams[threadID]
					for {
						i := int(atomic.AddUint64(&fileIdx, 1) - 1)
						if i >= numFiles {
							break
						}
						ts.Reset()
						_ = lexer.Lex(inputFiles[i].Content, ts)
						if r == 0 {
							fileChecksums[i] = ts.FileChecksum()
							fileLines[i] = int64(len(ts.Lines))
							fileTokens[i] = int64(ts.Len())
							fileKinds[i] = ts.KindCounts()
						}
					}
				}(t)
			}
			wg.Wait()
		}

		durations[r] = time.Since(t0)
	}

	// Compute metrics
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	bestNs := durations[0].Nanoseconds()
	medianNs := durations[len(durations)/2].Nanoseconds()

	var mbPerS float64
	if bestNs > 0 {
		mbPerS = math.Round((float64(totalBytes)*1000.0/float64(bestNs))*10.0) / 10.0
	}

	corpusChecksum := lexer.CorpusChecksum(fileChecksums)

	totalLines := int64(0)
	totalTokens := int64(0)
	totalKinds := map[string]int{
		"ident":         0,
		"int":           0,
		"float":         0,
		"string":        0,
		"punct":         0,
		"line_comment":  0,
		"block_comment": 0,
		"error":         0,
	}

	for i := 0; i < numFiles; i++ {
		totalLines += fileLines[i]
		totalTokens += fileTokens[i]
		for k, v := range fileKinds[i] {
			totalKinds[k] += v
		}
	}

	variantOut := cfg.Variant
	if variantOut == "auto" {
		variantOut = "scalar"
	}

	metrics := OutputMetrics{
		Impl:     "easyp-go",
		Variant:  variantOut,
		Threads:  cfg.Threads,
		Files:    numFiles,
		Bytes:    totalBytes,
		Lines:    totalLines,
		Tokens:   totalTokens,
		Errors:   int64(totalKinds["error"]),
		Kinds:    totalKinds,
		Checksum: fmt.Sprintf("%016x", corpusChecksum),
		BestNs:   bestNs,
		MedianNs: medianNs,
		MbPerS:   mbPerS,
	}

	jsonBytes, err := json.Marshal(metrics)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lexbench: failed to marshal metrics: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonBytes))
}
