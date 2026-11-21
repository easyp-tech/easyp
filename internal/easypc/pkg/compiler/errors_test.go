package compiler_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/compiler"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

func TestSharedErrorCorpus(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Locate repository root from internal/compiler
	root := filepath.Clean(filepath.Join(wd, "..", ".."))
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	errorsDir := filepath.Join(root, "testdata", "errors")
	protocInclude := filepath.Join(root, "protoc", "wkt")

	entries, err := os.ReadDir(errorsDir)
	if err != nil {
		t.Fatalf("failed to read %s: %v", errorsDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".proto") {
			continue
		}
		if entry.Name() == "cycle_b.proto" {
			// Helper file imported by cycle_a.proto
			continue
		}

		testName := strings.TrimSuffix(entry.Name(), ".proto")
		t.Run(testName, func(t *testing.T) {
			goldenPath := filepath.Join(errorsDir, testName+".golden")
			goldenBytes, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("missing golden file %s: %v", goldenPath, err)
			}
			expectedGolden := strings.TrimSpace(string(goldenBytes))

			var errBuf bytes.Buffer
			inputRelPath := filepath.Join("testdata", "errors", entry.Name())
			tmpOut := filepath.Join(t.TempDir(), "out.fds")

			opts := &compiler.Options{
				ImportPaths:      []string{filepath.Join("testdata", "errors"), protocInclude},
				InputFiles:       []string{inputRelPath},
				DescriptorSetOut: tmpOut,
				Variant:          "scalar",
				Strategy:         "auto",
				ErrorHandler: func(diag parser.Diagnostic) {
					errBuf.WriteString(diag.String() + "\n")
				},
			}

			driver := compiler.NewDriver(opts)
			err = driver.Compile()
			if err == nil {
				t.Fatalf("expected compilation error for %s, got nil", entry.Name())
			}

			// Check that no output file was created on compilation error (REQ-3.3)
			if _, statErr := os.Stat(tmpOut); statErr == nil {
				t.Fatalf("output file %s was created despite compilation error", tmpOut)
			}

			actualOutput := strings.TrimSpace(errBuf.String())
			if actualOutput != expectedGolden {
				t.Errorf("mismatch for %s\n--- EXPECTED GOLDEN ---\n%s\n--- ACTUAL ---\n%s",
					entry.Name(), expectedGolden, actualOutput)
			}
		})
	}
}
