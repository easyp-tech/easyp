package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestWriteV1ResolvedFilesRestoresManifestWhenLockCannotReplace(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	original := []byte("module example.com/root\n")
	updated := append(append([]byte(nil), original...), []byte("require example.com/dep\n")...)
	require.NoError(t, os.WriteFile(filepath.Join(root, v1.ModuleFile), original, 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(root, v1.LockFile), 0o755))

	err := writeV1ResolvedFiles(root, original, updated, v1.Lock{Version: 1})
	require.Error(t, err)
	current, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
	require.NoError(t, err)
	require.Equal(t, original, current)
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Len(t, entries, 2)
}
