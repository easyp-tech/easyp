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
