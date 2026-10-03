package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAtomicFilePermissionsAndReplacement(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "protobuf.mod")

	require.NoError(t, WriteAtomicFile(path, []byte("first"), 0o644))
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())

	require.NoError(t, os.Chmod(path, 0o640))
	require.NoError(t, WriteAtomicFile(path, []byte("second"), 0o644))
	info, err = os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "second", string(raw))
}
