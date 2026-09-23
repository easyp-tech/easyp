package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateVersionlessV1Policy(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "easyp.yaml")
	err := os.WriteFile(path, []byte("linters:\n  default: MINIMAL\nbreaking:\n  baseline: git:main\n"), 0o644)
	require.NoError(t, err)

	issues, err := validateConfigFile(path)
	require.NoError(t, err)
	assert.Empty(t, issues)
}

func TestValidateRejectsLegacyPolicy(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "easyp.yaml")
	require.NoError(t, os.WriteFile(path, []byte("lint:\n  use: [MINIMAL]\n"), 0o644))
	issues, err := validateConfigFile(path)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "v1_validation", issues[0].Code)
}

func TestValidateV1Lock(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		content string
		valid   bool
	}{
		{name: "empty graph", content: "version: 1\nmodules: []\n", valid: true},
		{name: "wrong version", content: "version: 2\nmodules: []\n"},
		{name: "incomplete module", content: "version: 1\nmodules:\n  - source: example.com/dep\n    version: v1.0.0\n"},
		{name: "bad hash", content: "version: 1\nmodules:\n  - source: example.com/dep\n    version: v1.0.0\n    commit: 0000000000000000000000000000000000000000\n    hash: h1:bad\n"},
		{name: "unknown field", content: "version: 1\nmodules: []\nunknown: true\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "protobuf.lock")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), 0o644))
			issues, err := validateConfigFile(path)
			require.NoError(t, err)
			if tc.valid {
				assert.Empty(t, issues)
			} else {
				require.Len(t, issues, 1)
				assert.Equal(t, "v1_validation", issues[0].Code)
			}
		})
	}
}
