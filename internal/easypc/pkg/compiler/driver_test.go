package compiler_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/compiler"
)

func TestParseArgs(t *testing.T) {
	tmpDir := t.TempDir()
	argFile := filepath.Join(tmpDir, "args.txt")
	if err := os.WriteFile(argFile, []byte("-I/tmp/include\n# comment\nfoo.proto\n"), 0644); err != nil {
		t.Fatal(err)
	}

	args := []string{
		"-I", "/first/include",
		"--proto_path=/second/include",
		"-o", "/tmp/out.pb",
		"--include_imports",
		"--include_source_info",
		"--threads=4",
		"--variant=scalar",
		"--strategy=b",
		"--metrics_out=/tmp/metrics.json",
		"--metrics_mode=trace",
		"@" + argFile,
		"bar.proto",
	}

	opts, err := compiler.ParseArgs(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(opts.ImportPaths) != 3 {
		t.Errorf("expected 3 import paths, got %d: %v", len(opts.ImportPaths), opts.ImportPaths)
	}
	if opts.DescriptorSetOut != "/tmp/out.pb" {
		t.Errorf("expected /tmp/out.pb, got %s", opts.DescriptorSetOut)
	}
	if !opts.IncludeImports {
		t.Errorf("expected IncludeImports=true")
	}
	if !opts.IncludeSourceInfo {
		t.Errorf("expected IncludeSourceInfo=true")
	}
	if opts.Threads != 4 {
		t.Errorf("expected Threads=4, got %d", opts.Threads)
	}
	if opts.Variant != "scalar" {
		t.Errorf("expected Variant=scalar, got %s", opts.Variant)
	}
	if opts.Strategy != "b" {
		t.Errorf("expected Strategy=b, got %s", opts.Strategy)
	}
	if opts.MetricsOut != "/tmp/metrics.json" {
		t.Errorf("expected MetricsOut=/tmp/metrics.json, got %s", opts.MetricsOut)
	}
	if opts.MetricsMode != "trace" {
		t.Errorf("expected MetricsMode=trace, got %s", opts.MetricsMode)
	}
	if len(opts.InputFiles) != 2 || opts.InputFiles[0] != "foo.proto" || opts.InputFiles[1] != "bar.proto" {
		t.Errorf("unexpected input files: %v", opts.InputFiles)
	}
}

func TestParseArgsErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing_out", []string{"foo.proto"}},
		{"missing_in", []string{"-o", "out.pb"}},
		{"unrecognized_flag", []string{"--invalid-flag", "-o", "out.pb", "foo.proto"}},
		{"missing_I_arg", []string{"-I"}},
		{"missing_o_arg", []string{"-o"}},
		{"invalid_variant", []string{"-o", "out.pb", "--variant=bad", "foo.proto"}},
		{"invalid_strategy", []string{"-o", "out.pb", "--strategy=invalid", "foo.proto"}},
		{"invalid_threads_zero", []string{"-o", "out.pb", "--threads=0", "foo.proto"}},
		{"invalid_threads_negative", []string{"-o", "out.pb", "--threads=-1", "foo.proto"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compiler.ParseArgs(tt.args)
			if err == nil {
				t.Fatalf("expected error for args %v, got nil", tt.args)
			}
			if _, ok := err.(*compiler.UsageError); !ok {
				t.Fatalf("expected UsageError, got %T: %v", err, err)
			}
		})
	}
}

func TestParseArgsVariantValidation(t *testing.T) {
	// Valid portable variants
	validCommon := []string{"auto", "scalar", "swar"}
	for _, v := range validCommon {
		args := []string{"-o", "out.pb", "--variant=" + v, "foo.proto"}
		opts, err := compiler.ParseArgs(args)
		if err != nil {
			t.Errorf("expected variant %q to be valid, got error: %v", v, err)
		} else if opts.Variant != v {
			t.Errorf("expected opts.Variant=%q, got %q", v, opts.Variant)
		}

		// Space-separated argument form
		argsSpace := []string{"-o", "out.pb", "--variant", v, "foo.proto"}
		optsSpace, err := compiler.ParseArgs(argsSpace)
		if err != nil {
			t.Errorf("expected space-separated variant %q to be valid, got error: %v", v, err)
		} else if optsSpace.Variant != v {
			t.Errorf("expected optsSpace.Variant=%q, got %q", v, optsSpace.Variant)
		}
	}

	// Architecture-specific valid variant
	if runtime.GOARCH == "arm64" {
		args := []string{"-o", "out.pb", "--variant=neon", "foo.proto"}
		opts, err := compiler.ParseArgs(args)
		if err != nil {
			t.Errorf("expected neon to be valid on arm64, got error: %v", err)
		} else if opts.Variant != "neon" {
			t.Errorf("expected opts.Variant=neon, got %q", opts.Variant)
		}
	}

	// Always unrecognized variants
	invalid := []string{"bad", "neon2", "sse4", "unknown"}
	for _, v := range invalid {
		args := []string{"-o", "out.pb", "--variant=" + v, "foo.proto"}
		_, err := compiler.ParseArgs(args)
		if err == nil {
			t.Errorf("expected variant %q to fail validation, got nil", v)
		}
		if _, ok := err.(*compiler.UsageError); !ok {
			t.Errorf("expected UsageError for variant %q, got %T: %v", v, err, err)
		}
	}

	// Cross-architecture check: architecture mismatch must fail with UsageError
	var unsupportedArchVariant string
	if runtime.GOARCH == "arm64" {
		unsupportedArchVariant = "avx2"
	} else if runtime.GOARCH == "amd64" {
		unsupportedArchVariant = "neon"
	}
	if unsupportedArchVariant != "" {
		args := []string{"-o", "out.pb", "--variant=" + unsupportedArchVariant, "foo.proto"}
		_, err := compiler.ParseArgs(args)
		if err == nil {
			t.Errorf("expected variant %q to fail on %s, got nil", unsupportedArchVariant, runtime.GOARCH)
		}
		if _, ok := err.(*compiler.UsageError); !ok {
			t.Errorf("expected UsageError for variant %q on %s, got %T: %v", unsupportedArchVariant, runtime.GOARCH, err, err)
		}
	}
}

func TestDriverCompileWKT(t *testing.T) {
	tmpDir := t.TempDir()
	outPb := filepath.Join(tmpDir, "out.pb")
	metricsOut := filepath.Join(tmpDir, "metrics.json")

	args := []string{
		"-I", "../../protoc/wkt",
		"-o", outPb,
		"--include_imports",
		"--metrics_out=" + metricsOut,
		"--metrics_mode=trace",
		"google/protobuf/descriptor.proto",
	}

	opts, err := compiler.ParseArgs(args)
	if err != nil {
		t.Fatalf("failed to parse args: %v", err)
	}

	driver := compiler.NewDriver(opts)
	if err := driver.Compile(); err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	fi, err := os.Stat(outPb)
	if err != nil || fi.Size() == 0 {
		t.Fatalf("output descriptor set was not created or empty")
	}

	mf, err := os.Stat(metricsOut)
	if err != nil || mf.Size() == 0 {
		t.Fatalf("metrics file was not created or empty")
	}
}

func TestDriverFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	outPb := filepath.Join(tmpDir, "out.pb")

	args := []string{
		"-I", tmpDir,
		"-o", outPb,
		"non_existent.proto",
	}

	opts, err := compiler.ParseArgs(args)
	if err != nil {
		t.Fatalf("parse args: %v", err)
	}

	driver := compiler.NewDriver(opts)
	if err := driver.Compile(); err == nil {
		t.Fatalf("expected compilation error for non-existent file, got nil")
	}
}

func TestDriverPathPrefixStripping(t *testing.T) {
	tmpDir := t.TempDir()
	protoDir := filepath.Join(tmpDir, "sub", "dir")
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		t.Fatal(err)
	}
	protoFile := filepath.Join(protoDir, "test.proto")
	if err := os.WriteFile(protoFile, []byte(`syntax = "proto3"; package test; message M { int32 a = 1; }`), 0644); err != nil {
		t.Fatal(err)
	}

	outPb := filepath.Join(tmpDir, "out.pb")
	// Pass -I with protoDir, and input file with full path
	args := []string{
		"-I", protoDir,
		"-o", outPb,
		protoFile,
	}

	opts, err := compiler.ParseArgs(args)
	if err != nil {
		t.Fatal(err)
	}

	driver := compiler.NewDriver(opts)
	if err := driver.Compile(); err != nil {
		t.Fatalf("compilation with prefixed path failed: %v", err)
	}

	fi, err := os.Stat(outPb)
	if err != nil || fi.Size() == 0 {
		t.Fatalf("output descriptor set was not created or empty")
	}
}
