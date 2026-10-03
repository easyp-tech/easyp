package modules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteV1Vendor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		files     map[string]string
		roots     []string
		wantErr   string
		wantFiles map[string]string
	}{
		{name: "collision preserves existing vendor", files: map[string]string{"first/shared.proto": "first", "second/shared.proto": "second"}, roots: []string{"first", "second"}, wantErr: "shared.proto", wantFiles: map[string]string{"keep.proto": "keep"}},
		{name: "new sources replace stale files", files: map[string]string{"dep/v1/item.proto": "syntax = \"proto3\";"}, roots: []string{"dep"}, wantFiles: map[string]string{"v1/item.proto": "syntax = \"proto3\";"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for name, content := range tt.files {
				writeV1GenerateFixture(t, root, name, content)
			}
			vendor := filepath.Join(root, VendorDir)
			writeV1GenerateFixture(t, vendor, "keep.proto", "keep")
			var roots SourceRoots
			for _, dir := range tt.roots {
				roots = append(roots, SourceRoot{Path: filepath.Join(root, dir), Module: dir})
			}

			err := writeV1Vendor(root, roots)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			actual := map[string]string{}
			require.NoError(t, WalkProtoFiles(vendor, func(path string) error {
				relative, err := filepath.Rel(vendor, path)
				if err != nil {
					return err
				}
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				actual[filepath.ToSlash(relative)] = string(content)
				return nil
			}))
			assert.Equal(t, tt.wantFiles, actual)
		})
	}
}
