package schemagen

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	versionedPath := filepath.Join(tmp, "nested", "easyp-config-v1.schema.json")
	latestPath := filepath.Join(tmp, "nested", "easyp-config.schema.json")

	err := Run(Options{
		VersionedOut: versionedPath,
		LatestOut:    latestPath,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	versionedData, err := os.ReadFile(versionedPath)
	if err != nil {
		t.Fatalf("os.ReadFile(versionedPath) error = %v", err)
	}

	latestData, err := os.ReadFile(latestPath)
	if err != nil {
		t.Fatalf("os.ReadFile(latestPath) error = %v", err)
	}

	if !bytes.Equal(versionedData, latestData) {
		t.Fatalf("schema files differ")
	}

	if len(versionedData) == 0 {
		t.Fatalf("schema file is empty")
	}
}
