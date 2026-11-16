// Package compiler implements the high-level compilation driver and CLI processing.
package compiler

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/linker"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

var recognizedVariants = map[string]bool{
	"auto":   true,
	"scalar": true,
	"swar":   true,
	"neon":   true,
	"sve2":   true,
	"avx2":   true,
	"avx512": true,
}

func hasCPUFeature(feat string) bool {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return false
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "flags") || strings.HasPrefix(line, "Features") {
			for _, f := range strings.Fields(line) {
				if f == feat {
					return true
				}
			}
		}
	}
	return false
}

func isVariantSupported(variant string) bool {
	switch variant {
	case "auto", "scalar", "swar":
		return true
	case "neon":
		return runtime.GOARCH == "arm64"
	case "sve2":
		return runtime.GOARCH == "arm64" && hasCPUFeature("sve2")
	case "avx2":
		return runtime.GOARCH == "amd64" && (runtime.GOOS != "linux" || hasCPUFeature("avx2"))
	case "avx512":
		return runtime.GOARCH == "amd64" && (runtime.GOOS != "linux" || hasCPUFeature("avx512f") || hasCPUFeature("avx512"))
	default:
		return false
	}
}

func validateVariant(variant string) error {
	if variant == "" || !recognizedVariants[variant] || !isVariantSupported(variant) {
		return &UsageError{Msg: fmt.Sprintf("unsupported or unrecognized variant %q", variant)}
	}
	return nil
}

func validateStrategy(strategy string) error {
	switch strategy {
	case "auto", "a", "b", "c":
		return nil
	default:
		return &UsageError{Msg: fmt.Sprintf("unsupported or unrecognized strategy %q", strategy)}
	}
}

// UsageError indicates an error with command-line flags or arguments (exit code 2).
type UsageError struct {
	Msg string
}

func (e *UsageError) Error() string {
	return e.Msg
}

// Options holds configuration parsed from the command line or passed programmatically.
type Options struct {
	ImportPaths       []string
	DescriptorSetOut  string
	IncludeImports    bool
	IncludeSourceInfo bool
	Threads           int
	Variant           string
	Strategy          string
	MetricsOut        string
	MetricsMode       string
	InputFiles        []string

	Context       context.Context
	ReadFile      func(path string) ([]byte, error)
	ResolvePath   func(path string) (canonical string, fullPath string, err error)
	ErrorHandler  func(diag parser.Diagnostic)
	OutputBytes   *[]byte
	OrderedFiles  *[]*parser.ASTFile
	SymTableOut   **linker.SymbolTable
	SkipSerialize bool
	// TrackUses records, per file, which files define the names it resolves
	// during linking and option evaluation (SymbolTable.UsedFiles).
	TrackUses bool
	// KeepTokens keeps each file's token stream on its AST after parsing.
	KeepTokens bool
}

// ParseArgs parses CLI arguments into Options.
func ParseArgs(args []string) (*Options, error) {
	opts := &Options{
		MetricsMode: "parallel",
		Variant:     "auto",
		Strategy:    "auto",
	}

	// First expand @argfiles recursively or iteratively
	expanded, err := expandArgFiles(args)
	if err != nil {
		return nil, &UsageError{Msg: fmt.Sprintf("failed to read argfile: %v", err)}
	}

	for i := 0; i < len(expanded); i++ {
		arg := expanded[i]
		switch {
		case arg == "-I":
			i++
			if i >= len(expanded) {
				return nil, &UsageError{Msg: "-I requires an argument"}
			}
			opts.ImportPaths = append(opts.ImportPaths, expanded[i])
		case strings.HasPrefix(arg, "-I"):
			opts.ImportPaths = append(opts.ImportPaths, arg[2:])
		case arg == "--proto_path":
			i++
			if i >= len(expanded) {
				return nil, &UsageError{Msg: "--proto_path requires an argument"}
			}
			opts.ImportPaths = append(opts.ImportPaths, expanded[i])
		case strings.HasPrefix(arg, "--proto_path="):
			opts.ImportPaths = append(opts.ImportPaths, strings.TrimPrefix(arg, "--proto_path="))
		case arg == "-o":
			i++
			if i >= len(expanded) {
				return nil, &UsageError{Msg: "-o requires an argument"}
			}
			opts.DescriptorSetOut = expanded[i]
		case strings.HasPrefix(arg, "-o"):
			opts.DescriptorSetOut = arg[2:]
		case arg == "--descriptor_set_out":
			i++
			if i >= len(expanded) {
				return nil, &UsageError{Msg: "--descriptor_set_out requires an argument"}
			}
			opts.DescriptorSetOut = expanded[i]
		case strings.HasPrefix(arg, "--descriptor_set_out="):
			opts.DescriptorSetOut = strings.TrimPrefix(arg, "--descriptor_set_out=")
		case arg == "--include_imports":
			opts.IncludeImports = true
		case arg == "--include_source_info":
			opts.IncludeSourceInfo = true
		case arg == "--threads":
			i++
			if i >= len(expanded) {
				return nil, &UsageError{Msg: "--threads requires an argument"}
			}
			n, err := strconv.Atoi(expanded[i])
			if err != nil {
				return nil, &UsageError{Msg: fmt.Sprintf("invalid --threads: %v", err)}
			}
			if n <= 0 {
				return nil, &UsageError{Msg: "invalid --threads: must be positive integer"}
			}
			opts.Threads = n
		case strings.HasPrefix(arg, "--threads="):
			n, err := strconv.Atoi(strings.TrimPrefix(arg, "--threads="))
			if err != nil {
				return nil, &UsageError{Msg: fmt.Sprintf("invalid --threads: %v", err)}
			}
			if n <= 0 {
				return nil, &UsageError{Msg: "invalid --threads: must be positive integer"}
			}
			opts.Threads = n
		case arg == "--parallelism":
			i++
			if i >= len(expanded) {
				return nil, &UsageError{Msg: "--parallelism requires an argument"}
			}
			n, err := strconv.Atoi(expanded[i])
			if err != nil {
				return nil, &UsageError{Msg: fmt.Sprintf("invalid --parallelism: %v", err)}
			}
			if n <= 0 {
				return nil, &UsageError{Msg: "invalid --parallelism: must be positive integer"}
			}
			opts.Threads = n
		case strings.HasPrefix(arg, "--parallelism="):
			n, err := strconv.Atoi(strings.TrimPrefix(arg, "--parallelism="))
			if err != nil {
				return nil, &UsageError{Msg: fmt.Sprintf("invalid --parallelism: %v", err)}
			}
			if n <= 0 {
				return nil, &UsageError{Msg: "invalid --parallelism: must be positive integer"}
			}
			opts.Threads = n
		case arg == "--variant":
			i++
			if i >= len(expanded) {
				return nil, &UsageError{Msg: "--variant requires an argument"}
			}
			opts.Variant = expanded[i]
			if err := validateVariant(opts.Variant); err != nil {
				return nil, err
			}
		case strings.HasPrefix(arg, "--variant="):
			opts.Variant = strings.TrimPrefix(arg, "--variant=")
			if err := validateVariant(opts.Variant); err != nil {
				return nil, err
			}
		case arg == "--strategy":
			i++
			if i >= len(expanded) {
				return nil, &UsageError{Msg: "--strategy requires an argument"}
			}
			opts.Strategy = expanded[i]
			if err := validateStrategy(opts.Strategy); err != nil {
				return nil, err
			}
		case strings.HasPrefix(arg, "--strategy="):
			opts.Strategy = strings.TrimPrefix(arg, "--strategy=")
			if err := validateStrategy(opts.Strategy); err != nil {
				return nil, err
			}
		case arg == "--metrics_out":
			i++
			if i >= len(expanded) {
				return nil, &UsageError{Msg: "--metrics_out requires an argument"}
			}
			opts.MetricsOut = expanded[i]
		case strings.HasPrefix(arg, "--metrics_out="):
			opts.MetricsOut = strings.TrimPrefix(arg, "--metrics_out=")
		case arg == "--metrics_mode":
			i++
			if i >= len(expanded) {
				return nil, &UsageError{Msg: "--metrics_mode requires an argument"}
			}
			opts.MetricsMode = expanded[i]
		case strings.HasPrefix(arg, "--metrics_mode="):
			opts.MetricsMode = strings.TrimPrefix(arg, "--metrics_mode=")
		case strings.HasPrefix(arg, "-"):
			return nil, &UsageError{Msg: fmt.Sprintf("unrecognized flag: %s", arg)}
		default:
			opts.InputFiles = append(opts.InputFiles, arg)
		}
	}

	if len(opts.ImportPaths) == 0 {
		opts.ImportPaths = []string{"."}
	}
	if opts.DescriptorSetOut == "" {
		return nil, &UsageError{Msg: "Missing output directives."}
	}
	if len(opts.InputFiles) == 0 {
		return nil, &UsageError{Msg: "Missing input file."}
	}
	if err := validateVariant(opts.Variant); err != nil {
		return nil, err
	}
	if err := validateStrategy(opts.Strategy); err != nil {
		return nil, err
	}

	return opts, nil
}

func expandArgFiles(args []string) ([]string, error) {
	var result []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "@") {
			path := arg[1:]
			f, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			sc := bufio.NewScanner(f)
			var fileArgs []string
			for sc.Scan() {
				line := strings.TrimSpace(sc.Text())
				if line != "" && !strings.HasPrefix(line, "#") {
					fileArgs = append(fileArgs, line)
				}
			}
			if err := sc.Err(); err != nil {
				return nil, err
			}
			nested, err := expandArgFiles(fileArgs)
			if err != nil {
				return nil, err
			}
			result = append(result, nested...)
		} else {
			result = append(result, arg)
		}
	}
	return result, nil
}
