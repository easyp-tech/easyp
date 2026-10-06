package sourceview

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenLocalSelectedRejectsChangedPhysicalTarget(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "file.proto"), []byte("original"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "replacement.proto"), []byte("modified"), 0o600))
	file, err := OpenLocalSelected(t.Context(), root, "file.proto", func(Resolution) bool {
		require.NoError(t, os.Rename(filepath.Join(root, "replacement.proto"), filepath.Join(root, "file.proto")))
		return true
	})
	require.ErrorIs(t, err, ErrChanged)
	require.Nil(t, file)
}

func TestOpenLocalRejectsNestedRepositoryAlias(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "nested/.git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "nested/file.proto"), []byte("nested source"), 0o600))
	require.NoError(t, os.Symlink("nested/file.proto", filepath.Join(root, "alias.proto")))
	file, err := OpenLocal(t.Context(), root, "alias.proto")
	if file != nil {
		require.NoError(t, file.Close())
	}
	require.ErrorIs(t, err, ErrUnsupported)
}
