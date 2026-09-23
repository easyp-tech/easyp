package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveV1BreakingPolicyInheritsWholeSection(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	rootPolicy := filepath.Join(root, "easyp.yaml")
	childDir := filepath.Join(root, "nested")
	require.NoError(t, os.MkdirAll(childDir, 0o755))
	require.NoError(t, os.WriteFile(rootPolicy, []byte("version: v1\nbreaking:\n  baseline: git:main\n  ignore: [generated]\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(childDir, "easyp.yaml"), []byte("version: v1\nlinters:\n  default: MINIMAL\n"), 0o644))

	policy, source, err := resolveV1BreakingPolicy(childDir, root, rootPolicy)
	require.NoError(t, err)
	require.Equal(t, rootPolicy, source)
	require.Equal(t, "git:main", policy.Breaking.Baseline)
	require.Equal(t, []string{"generated"}, policy.Breaking.Ignore)
}
