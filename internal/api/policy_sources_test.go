package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicySourceExclusions(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	tests := []struct {
		name, path string
		excluded   bool
	}{
		{"vendor root", "easyp_vendor", true}, {"nested vendor", "module/easyp_vendor/item.proto", true},
		{"similar name", "easyp_vendor_backup/item.proto", false}, {"hidden storage", ".cache/mod/item.proto", true},
		{"normal source", "proto/item.proto", false}, {"root", ".", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.excluded, policySourceExcluded(root, filepath.Join(root, tt.path)))
		})
	}
}

func TestCLIDoesNotPublishRemovedDocumentation(t *testing.T) {
	t.Parallel()
	_, err := os.Stat(filepath.Join("..", "..", ".github", "workflows", "docs.yml"))
	require.ErrorIs(t, err, os.ErrNotExist)
	raw, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "release.yml"))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "cd docs")
}
