package api

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestLocalV1DependencyRootsFromLegacyAndBuf(t *testing.T) {
	root := t.TempDir()
	for _, tc := range []struct {
		name       string
		configName string
		configBody string
		protoDir   string
	}{
		{"old-easyp", "easyp.yaml", "generate:\n  inputs:\n    - directory:\n        path: proto\n        root: proto\n", "proto"},
		{"old-buf", "buf.work.yaml", "version: v1\ndirectories: [schemas]\n", "schemas"},
	} {
		depDir := filepath.Join(root, tc.name)
		if err := os.MkdirAll(filepath.Join(depDir, tc.protoDir), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(depDir, tc.configName), []byte(tc.configBody), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	module := v1.Module{
		Name:     "example.com/root",
		Requires: []v1.Requirement{{Module: "example.com/old-easyp", Version: "v1.0.0"}, {Module: "example.com/old-buf", Version: "v1.0.0"}},
		Replaces: []v1.Replacement{{Module: "example.com/old-easyp", Target: "old-easyp"}, {Module: "example.com/old-buf", Target: "old-buf"}},
	}
	roots, err := localV1DependencyRoots(root, module, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join(root, "old-easyp", "proto"), filepath.Join(root, "old-buf", "schemas")}
	if !reflect.DeepEqual(roots, want) {
		t.Fatalf("roots = %v, want %v", roots, want)
	}
}
