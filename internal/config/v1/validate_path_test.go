package v1

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePathSelectsConfigFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		files         map[string]string
		selectedPath  string
		expectedFiles []string
	}{
		{
			name: "all nested configs in lexical order",
			files: map[string]string{
				"easyp.yaml":             "linters:\n  unknown: true\n",
				"api/easyp.gen.yaml":     "plugins:\n  - name: go\n    out: gen\n    unknown: true\n",
				"api/deep/protobuf.lock": "version: 2\nmodules: []\n",
				"api/deep/protobuf.mod":  "roots (\n  proto\n)\n",
				"api/deep/other.yaml":    "anything: true\n",
			},
			selectedPath:  ".",
			expectedFiles: []string{"api/deep/protobuf.lock", "api/deep/protobuf.mod", "api/easyp.gen.yaml", "easyp.yaml"},
		},
		{
			name: "explicit file excludes other configs",
			files: map[string]string{
				"easyp.yaml":            "version: v1\n",
				"nested/easyp.gen.yaml": "plugins:\n  - invalid: true\n",
			},
			selectedPath: "easyp.yaml",
		},
		{
			name: "explicit nested file uses its basename",
			files: map[string]string{
				"nested/easyp.yaml": "linters:\n  unknown: true\n",
			},
			selectedPath:  "nested/easyp.yaml",
			expectedFiles: []string{"easyp.yaml"},
		},
		{
			name: "selected directory bounds the search",
			files: map[string]string{
				"easyp.yaml":          "linters:\n  unknown: true\n",
				"nested/protobuf.mod": "module example.com/service\n",
			},
			selectedPath: "nested",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			for name, contents := range tt.files {
				writeValidationFixture(t, root, name, contents)
			}

			issues, err := ValidatePath(filepath.Join(root, tt.selectedPath))

			require.NoError(t, err)
			var issueFiles []string
			for _, issue := range issues {
				issueFiles = append(issueFiles, issue.File)
			}
			assert.Equal(t, tt.expectedFiles, issueFiles)
		})
	}
}

func TestValidatePathRejectsDirectoryWithoutConfigs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		files map[string]string
	}{
		{name: "empty directory"},
		{name: "unrelated yaml only", files: map[string]string{"other.yaml": "anything: true\n"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			for name, contents := range tt.files {
				writeValidationFixture(t, root, name, contents)
			}

			issues, err := ValidatePath(root)

			require.ErrorContains(t, err, "no configuration files")
			assert.Contains(t, err.Error(), root)
			assert.Nil(t, issues)
		})
	}
}

func TestValidatePathPreservesFilesystemError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		selectedPath string
	}{
		{name: "missing file", selectedPath: "easyp.yaml"},
		{name: "missing directory", selectedPath: "nested"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), tt.selectedPath)

			issues, err := ValidatePath(path)

			require.ErrorIs(t, err, os.ErrNotExist)
			assert.Nil(t, issues)
			var pathErr *os.PathError
			require.ErrorAs(t, err, &pathErr)
			assert.Equal(t, path, pathErr.Path)
		})
	}
}

func writeValidationFixture(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
