package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBreakingIgnorePathsArePolicyRelative(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	got, err := breakingIgnorePaths(root, filepath.Join(root, "api"), []string{"generated", "legacy/old.proto"})
	require.NoError(t, err)
	require.Equal(t, []string{"api/generated", "api/legacy/old.proto"}, got)
	for _, bad := range []string{"", "../other", root} {
		_, err := breakingIgnorePaths(root, filepath.Join(root, "api"), []string{bad})
		require.Error(t, err)
	}
}
