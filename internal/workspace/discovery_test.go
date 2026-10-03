package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscoveryHonorsBoundariesAndNearestFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "module/deep"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "easyp.yaml"), []byte("version: v1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "module/protobuf.mod"), []byte("module example.com/module\n"), 0o644))
	deep := filepath.Join(root, "module/deep")
	boundary, err := Boundary(deep)
	require.NoError(t, err)
	require.Equal(t, root, boundary)
	policy, err := Policy(deep)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(root, "easyp.yaml"), policy)
	module, err := Module(deep)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(root, "module"), module)
	require.NoError(t, os.WriteFile(filepath.Join(deep, ".git"), []byte("gitdir: /unused\n"), 0o644))
	boundary, err = Boundary(deep)
	require.NoError(t, err)
	require.Equal(t, deep, boundary)
	_, err = Policy(deep)
	require.ErrorContains(t, err, "no easyp.yaml")
}

func TestIndependentPoliciesRequireExplicitSelection(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, dir := range []string{"one", "two"} {
		require.NoError(t, os.Mkdir(filepath.Join(root, dir), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(root, dir, "easyp.yaml"), []byte("version: v1\n"), 0o644))
	}
	_, err := Policy(root)
	require.ErrorContains(t, err, "multiple independent policies")
}
