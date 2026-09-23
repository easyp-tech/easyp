package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveV1LintPolicyInheritsSectionsIndependently(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	child := filepath.Join(root, "nested")
	writeV1GenerateFixture(t, root, "custom.yaml", "version: v1\nlinters:\n  default: STANDARD\nissues:\n  exclude-rules:\n    - path: generated\n")
	writeV1GenerateFixture(t, child, "easyp.yaml", "version: v1\nlinters:\n  default: MINIMAL\n")

	policy, sources, err := resolveV1LintPolicy(child, root, filepath.Join(root, "custom.yaml"))
	require.NoError(t, err)
	require.Equal(t, "MINIMAL", policy.Linters.Default)
	require.Len(t, policy.Issues.ExcludeRules, 1)
	require.Equal(t, filepath.Join(child, "easyp.yaml"), sources.linters)
	require.Equal(t, filepath.Join(root, "custom.yaml"), sources.issues)
}

func TestResolveV1BreakingPolicyUsesExplicitEmptySection(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	child := filepath.Join(root, "nested")
	writeV1GenerateFixture(t, root, "easyp.yaml", "version: v1\nbreaking:\n  baseline: git:main\n")
	writeV1GenerateFixture(t, child, "easyp.yaml", "version: v1\nbreaking: {}\n")

	policy, source, err := resolveV1BreakingPolicy(child, root, filepath.Join(root, "easyp.yaml"))
	require.NoError(t, err)
	require.Empty(t, policy.Breaking.Baseline)
	require.Equal(t, filepath.Join(child, "easyp.yaml"), source)
}
