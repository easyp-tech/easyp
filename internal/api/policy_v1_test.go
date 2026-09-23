package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveV1LintPolicyInheritsSectionsIndependently(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		childPolicy   string
		wantDefault   string
		wantRuleCount int
		wantSources   v1LintPolicySources
	}{
		{
			name:          "replace_linters_inherit_issues",
			childPolicy:   "version: v1\nlinters:\n  default: MINIMAL\n",
			wantDefault:   "MINIMAL",
			wantRuleCount: 1,
			wantSources:   v1LintPolicySources{linters: "nested/easyp.yaml", issues: "custom.yaml"},
		},
		{
			name:          "inherit_linters_clear_issues",
			childPolicy:   "version: v1\nissues: {}\n",
			wantDefault:   "STANDARD",
			wantRuleCount: 0,
			wantSources:   v1LintPolicySources{linters: "custom.yaml", issues: "nested/easyp.yaml"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			child := filepath.Join(root, "nested")
			writeV1GenerateFixture(t, root, "custom.yaml", "version: v1\nlinters:\n  default: STANDARD\nissues:\n  exclude-rules:\n    - path: generated\n")
			writeV1GenerateFixture(t, child, "easyp.yaml", tt.childPolicy)

			policy, sources, err := resolveV1LintPolicy(child, root, filepath.Join(root, "custom.yaml"))

			require.NoError(t, err)
			assert.Equal(t, tt.wantDefault, policy.Linters.Default)
			assert.Len(t, policy.Issues.ExcludeRules, tt.wantRuleCount)
			assert.Equal(t, filepath.Join(root, tt.wantSources.linters), sources.linters)
			assert.Equal(t, filepath.Join(root, tt.wantSources.issues), sources.issues)
			assert.Empty(t, sources.settings)
		})
	}
}
