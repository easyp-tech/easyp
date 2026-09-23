package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestLocalV1DependencyRootsFromLegacyAndBuf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		files     map[string]string
		module    v1.Module
		wantRoots []string
	}{
		{
			name: "legacy_easyp_and_buf_workspace",
			files: map[string]string{
				"old-easyp/easyp.yaml":       "generate:\n  inputs:\n    - directory:\n        path: proto\n        root: proto\n",
				"old-easyp/proto/item.proto": "syntax = \"proto3\";\n",
				"old-buf/buf.work.yaml":      "version: v1\ndirectories: [schemas]\n",
				"old-buf/schemas/item.proto": "syntax = \"proto3\";\n",
			},
			module: v1.Module{
				Name:     "example.com/root",
				Requires: []v1.Requirement{{Module: "example.com/old-easyp", Version: "v1.0.0"}, {Module: "example.com/old-buf", Version: "v1.0.0"}},
				Replaces: []v1.Replacement{{Module: "example.com/old-easyp", Target: "old-easyp"}, {Module: "example.com/old-buf", Target: "old-buf"}},
			},
			wantRoots: []string{"old-easyp/proto", "old-buf/schemas"},
		},
		{
			name: "dependency_without_config",
			files: map[string]string{
				"dependency/item.proto": "syntax = \"proto3\";\n",
			},
			module: v1.Module{
				Name:     "example.com/root",
				Requires: []v1.Requirement{{Module: "example.com/dependency", Version: "v1.0.0"}},
				Replaces: []v1.Replacement{{Module: "example.com/dependency", Target: "dependency"}},
			},
			wantRoots: []string{"dependency"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			for path, content := range tt.files {
				writeV1GenerateFixture(t, root, path, content)
			}
			want := make([]string, 0, len(tt.wantRoots))
			for _, path := range tt.wantRoots {
				want = append(want, filepath.Join(root, path))
			}

			roots, err := localV1DependencySources(root, tt.module, map[string]bool{})

			require.NoError(t, err)
			assert.Equal(t, want, roots.paths())
		})
	}
}
