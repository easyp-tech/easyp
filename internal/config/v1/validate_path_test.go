package v1

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigPathChecksAllNestedConfigs(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeValidationFixture(t, root, "easyp.yaml", "linters:\n  unknown: true\n")
	writeValidationFixture(t, root, "api/easyp.gen.yaml", "plugins:\n  - name: go\n    unknown: true\n")
	writeValidationFixture(t, root, "api/deep/protobuf.lock", "version: 2\nmodules: []\n")
	writeValidationFixture(t, root, "api/deep/other.yaml", "anything: true\n")

	issues, err := ValidatePath(root)
	require.NoError(t, err)
	files := make(map[string]bool)
	for _, issue := range issues {
		files[issue.File] = true
	}
	assert.Equal(t, map[string]bool{
		"easyp.yaml":             true,
		"api/easyp.gen.yaml":     true,
		"api/deep/protobuf.lock": true,
	}, files)
}

func TestValidateConfigPathKeepsExplicitFileSelection(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	selected := writeValidationFixture(t, root, "easyp.yaml", "version: v1\n")
	writeValidationFixture(t, root, "nested/easyp.gen.yaml", "plugins:\n  - invalid: true\n")

	issues, err := ValidatePath(selected)
	require.NoError(t, err)
	assert.Empty(t, issues)
}

func TestValidateConfigPathRejectsEmptyDirectory(t *testing.T) {
	t.Parallel()
	_, err := ValidatePath(t.TempDir())
	require.ErrorContains(t, err, "no configuration files")
}

func writeValidationFixture(t *testing.T, root, name, content string) string {
	t.Helper()
	path := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}
