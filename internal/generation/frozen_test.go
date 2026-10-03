package generation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestFrozenGenerationPreflightsAllSelectedProjects(t *testing.T) {
	t.Parallel()
	for _, failure := range []string{"missing lock", "unused replace"} {
		t.Run(failure, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for _, name := range []string{"a", "b"} {
				dir := filepath.Join(root, name)
				writeV1GenerateFixture(t, dir, "protobuf.mod", "module example.com/"+name+"\nroots proto\n")
				writeV1GenerateFixture(t, dir, "protobuf.lock", "version: 1\nmodules: []\n")
				writeV1GenerateFixture(t, dir, "proto/"+name+".proto", "syntax = \"proto3\"; package "+name+".v1; message Item {}")
				writeV1GenerateFixture(t, dir, "easyp.gen.yaml", "version: v1\nplugins:\n  - command: [sh, -c, 'touch marker; cat >/dev/null']\n    out: gen\n")
			}
			if failure == "missing lock" {
				require.NoError(t, os.Remove(filepath.Join(root, "b", "protobuf.lock")))
			} else {
				writeV1GenerateFixture(t, filepath.Join(root, "b"), "protobuf.mod", "module example.com/b\nroots proto\nreplace example.com/unused => ../nonexistent\n")
			}
			err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, WorkspaceRoot: root, AllProjects: true, Frozen: true})
			require.ErrorContains(t, err, "frozen")
			require.NoFileExists(t, filepath.Join(root, "marker"))
			err = Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, WorkspaceRoot: root, Projects: []string{"a"}, Frozen: true, DescriptorSetOut: "a.pb"})
			require.NoError(t, err)
			require.FileExists(t, filepath.Join(root, "marker"))
			require.FileExists(t, filepath.Join(root, "a.pb"))
		})
	}
}

func TestFrozenGenerationWithoutPluginsValidatesEmptyModule(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/empty\n")
	writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\n")
	request := Request{WorkDir: root, WorkspaceRoot: root, Frozen: true}
	err := Run(t.Context(), logger.NewNop(), nil, request)
	require.ErrorContains(t, err, "protobuf.lock")
	writeV1GenerateFixture(t, root, "protobuf.lock", "version: 1\nmodules: []\n")
	require.NoError(t, Run(t.Context(), logger.NewNop(), nil, request))
}
