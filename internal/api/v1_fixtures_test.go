package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func writeV1GenerateFixture(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
}

func cachedTestDir(t *testing.T, cache modules.Cache, entry v1.LockedModule) string {
	t.Helper()
	directory, _, err := cache.Cached(entry)
	require.NoError(t, err)
	return directory
}
