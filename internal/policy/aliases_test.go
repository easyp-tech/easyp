package policy

import (
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/stretchr/testify/require"
)

func TestLinkedPolicyUsesLogicalRelativeExtends(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "storage"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "consumer"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "storage/source.yaml"), []byte("version: v1\nbreaking:\n  extends: ./base.yaml\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "consumer/base.yaml"), []byte("version: v1\nbreaking:\n  baseline: git:logical\n"), 0o644))
	require.NoError(t, os.Symlink("../storage/source.yaml", filepath.Join(root, "consumer/easyp.yaml")))
	resolver := NewResolver(root, nil)
	result, err := resolver.ResolveBreaking(t.Context(), BreakingInput{PolicyPath: filepath.Join(root, "consumer/easyp.yaml"), Policy: v1.Policy{Version: "v1", Breaking: v1.BreakingPolicy{Extends: "./base.yaml"}}})
	require.NoError(t, err)
	require.Equal(t, "git:logical", result.Policy.Breaking.Baseline)
}
