package generation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestDescriptorOutputRejectsDirectoryAliases(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeDescriptorMarker(t, root)
	writeV1GenerateFixture(t, root, "proto/protobuf.mod", "module example.com/item\n")
	writeV1GenerateFixture(t, root, "proto/item.proto", "syntax = \"proto3\"; package item; message Item {}")
	for _, project := range []string{"backend", "frontend"} {
		writeV1GenerateFixture(t, root, project+"/easyp.gen.yaml", "generate:\n  modules: [proto]\noptions:\n  go:\n    package_prefix: example.com/"+project+"\n"+descriptorMarkerConfig)
	}
	out := filepath.Join(root, "sets")
	require.NoError(t, os.MkdirAll(filepath.Join(out, "backend"), 0o755))
	require.NoError(t, os.Symlink("backend", filepath.Join(out, "frontend")))

	err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOutDir: out})

	require.ErrorContains(t, err, "output collision")
	assert.NoFileExists(t, filepath.Join(root, "plugin-ran.txt"))
	entries, err := os.ReadDir(filepath.Join(out, "backend"))
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestDescriptorOutputPreservesExistingMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		split bool
	}{
		{name: "combined"},
		{name: "split", split: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/item\n")
			writeV1GenerateFixture(t, root, "item.proto", "syntax = \"proto3\"; package item; message Item {}")
			writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\n")
			request := Request{WorkDir: root, DescriptorSetOut: "nested/all.pb"}
			if tt.split {
				request.DescriptorSetOut, request.DescriptorSetOutDir = "", "nested/sets"
			}
			require.NoError(t, Run(t.Context(), logger.NewNop(), nil, request))
			tree := descriptorTree(t, filepath.Join(root, "nested"))
			require.Len(t, tree, 1)
			var output string
			for name := range tree {
				output = filepath.Join(root, "nested", filepath.FromSlash(name))
			}
			require.NoError(t, os.Chmod(output, 0o600))
			require.NoError(t, Run(t.Context(), logger.NewNop(), nil, request))
			info, err := os.Stat(output)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
			assert.Equal(t, tree, descriptorTree(t, filepath.Join(root, "nested")))
		})
	}
}

func TestDescriptorOutputRejectsDanglingAncestor(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeDescriptorMarker(t, root)
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/item\n")
	writeV1GenerateFixture(t, root, "item.proto", "syntax = \"proto3\"; package item; message Item {}")
	writeV1GenerateFixture(t, root, "easyp.gen.yaml", descriptorMarkerConfig)
	require.NoError(t, os.Symlink("missing-directory", filepath.Join(root, "out")))

	err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: "out/nested/all.pb"})

	require.ErrorContains(t, err, "EvalSymlinks")
	assert.NoFileExists(t, filepath.Join(root, "plugin-ran.txt"))
	assert.NoDirExists(t, filepath.Join(root, "missing-directory"))
}
