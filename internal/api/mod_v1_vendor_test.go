package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteV1VendorPreservesExistingOnCollision(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, name := range []string{"first", "second"} {
		dir := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "shared.proto"), []byte("syntax = \"proto3\";"), 0o644))
	}
	vendor := filepath.Join(root, defaultVendorDir)
	require.NoError(t, os.MkdirAll(vendor, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(vendor, "keep.proto"), []byte("keep"), 0o644))

	err := writeV1Vendor(root, v1SourceRoots{
		{path: filepath.Join(root, "first"), module: "first"},
		{path: filepath.Join(root, "second"), module: "second"},
	})
	require.ErrorContains(t, err, "shared.proto")
	raw, err := os.ReadFile(filepath.Join(vendor, "keep.proto"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(raw))
}

func TestWriteV1VendorReplacesStaleFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dep := filepath.Join(root, "dep")
	require.NoError(t, os.MkdirAll(filepath.Join(dep, "v1"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dep, "v1", "item.proto"), []byte("syntax = \"proto3\";"), 0o644))
	vendor := filepath.Join(root, defaultVendorDir)
	require.NoError(t, os.MkdirAll(vendor, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(vendor, "stale.proto"), []byte("stale"), 0o644))

	require.NoError(t, writeV1Vendor(root, v1SourceRoots{{path: dep, module: "dep"}}))
	_, err := os.Stat(filepath.Join(vendor, "stale.proto"))
	require.ErrorIs(t, err, os.ErrNotExist)
	raw, err := os.ReadFile(filepath.Join(vendor, "v1", "item.proto"))
	require.NoError(t, err)
	require.Contains(t, string(raw), "proto3")
}
