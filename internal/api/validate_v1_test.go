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
