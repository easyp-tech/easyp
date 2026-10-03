package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveV1BreakingPolicyInheritsWholeSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		childPolicy  string
		wantSource   string
		wantBaseline string
		wantIgnore   []string
	}{
		{
			name:         "inherit_missing_section",
			childPolicy:  "version: v1\nlinters:\n  default: MINIMAL\n",
			wantSource:   "easyp.yaml",
			wantBaseline: "git:main",
			wantIgnore:   []string{"generated"},
		},
		{
			name:        "explicit_empty_section",
			childPolicy: "version: v1\nbreaking: {}\n",
			wantSource:  "nested/easyp.yaml",
		},
		{
			name:         "replace_whole_section",
			childPolicy:  "version: v1\nbreaking:\n  baseline: git:release\n",
			wantSource:   "nested/easyp.yaml",
			wantBaseline: "git:release",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			child := filepath.Join(root, "nested")
			writeV1GenerateFixture(t, root, "easyp.yaml", "version: v1\nbreaking:\n  baseline: git:main\n  ignore: [generated]\n")
			writeV1GenerateFixture(t, child, "easyp.yaml", tt.childPolicy)

			policy, source, err := resolveV1BreakingPolicy(child, root, filepath.Join(root, "easyp.yaml"))

			require.NoError(t, err)
			assert.Equal(t, filepath.Join(root, tt.wantSource), source)
			assert.Equal(t, tt.wantBaseline, policy.Breaking.Baseline)
			assert.Equal(t, tt.wantIgnore, policy.Breaking.Ignore)
		})
	}
}
