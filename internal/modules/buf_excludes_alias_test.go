package modules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
)

func TestBufExcludesPreserveLogicalAndPhysicalRootAliases(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name          string
		exclude       string
		physicalAllow bool
	}{
		{name: "physical_exclusion_applies_to_alias", exclude: "real/private"},
		{name: "logical_exclusion_keeps_physical_source", exclude: "alias/private", physicalAllow: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "buf.yaml", "version: v1\nbuild:\n  excludes: ["+test.exclude+"]\n")
			writeV1GenerateFixture(t, root, "real/public.proto", "syntax = \"proto3\";\n")
			writeV1GenerateFixture(t, root, "real/private/types.proto", "syntax = \"proto3\";\n")
			require.NoError(t, os.Symlink("real", filepath.Join(root, "alias")))
			module, err := moduleconfig.ReadLocalDependency(root, "example.test/types")
			require.NoError(t, err)
			sources, err := ModuleSources(root, module)
			require.NoError(t, err)

			allowed := sources.FileAllowed()

			assert.False(t, allowed(filepath.Join(root, "alias/private/types.proto")))
			assert.Equal(t, test.physicalAllow, allowed(filepath.Join(root, "real/private/types.proto")))
			assert.True(t, allowed(filepath.Join(root, "alias/public.proto")))
			assert.True(t, allowed(filepath.Join(root, "real/public.proto")))
		})
	}
}

func TestBufExcludesAgreeForRootAliasWalkAndReadOnlyReads(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		target string
	}{
		{name: "proto_target", target: "types.proto"},
		{name: "non_proto_target", target: "schema.data"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "buf.yaml", "version: v2\nmodules:\n  - path: alias\n    excludes: [alias/private]\n")
			const retained = "syntax = \"proto3\"; package retained; message Retained {}\n"
			writeV1GenerateFixture(t, root, "real/public/keep.proto", retained)
			writeV1GenerateFixture(t, root, "real/private/"+test.target, "this physical target is excluded")
			require.NoError(t, os.Symlink("../private/"+test.target, filepath.Join(root, "real/public/blocked.proto")))
			require.NoError(t, os.Symlink("real", filepath.Join(root, "alias")))
			require.NoError(t, os.Chmod(filepath.Join(root, "real/public/keep.proto"), 0o444))
			require.NoError(t, os.Chmod(filepath.Join(root, "real/private", test.target), 0o444))
			module, err := moduleconfig.ReadLocalDependency(root, "example.test/types")
			require.NoError(t, err)
			sources, err := ModuleSources(root, module)
			require.NoError(t, err)

			owners, err := sources.FileModules()

			require.NoError(t, err)
			assert.Equal(t, map[string]string{"public/keep.proto": "example.test/types"}, owners)
			for _, relative := range []string{"alias/public/blocked.proto", "real/public/blocked.proto", "alias/private/" + test.target, "real/private/" + test.target} {
				assert.False(t, sources.FileAllowed()(filepath.Join(root, relative)), relative)
			}
			_, err = sources.ReadSourceFile(filepath.Join(root, "alias/public/blocked.proto"))
			require.ErrorIs(t, err, os.ErrNotExist)
			contents, err := sources.ReadSourceFile(filepath.Join(root, "alias/public/keep.proto"))
			require.NoError(t, err)
			assert.Equal(t, retained, string(contents))
			info, err := os.Stat(filepath.Join(root, "real/public/keep.proto"))
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o444), info.Mode().Perm())
		})
	}
}

func TestBufExcludesPrunePhysicalDirectoryAliasCycles(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		metadata string
	}{
		{name: "v1", metadata: "version: v1\nbuild:\n  excludes: [preview]\n"},
		{name: "v2", metadata: "version: v2\nmodules:\n  - path: .\n    excludes: [preview]\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "buf.yaml", test.metadata)
			writeV1GenerateFixture(t, root, "schema.proto", "syntax = \"proto3\";\n")
			writeV1GenerateFixture(t, root, "preview/types.proto", "syntax = \"proto3\";\n")
			require.NoError(t, os.Symlink("preview", filepath.Join(root, "api")))
			require.NoError(t, os.Symlink(".", filepath.Join(root, "preview/loop")))
			module, err := moduleconfig.ReadLocalDependency(root, "example.test/types")
			require.NoError(t, err)
			sources, err := ModuleSources(root, module)
			require.NoError(t, err)
			var visited []string

			err = sources[0].Walk(func(path string) error {
				visited = append(visited, path)
				return nil
			})

			require.NoError(t, err)
			assert.Equal(t, []string{filepath.Join(root, "schema.proto")}, visited)
			_, err = sources.ReadSourceFile(filepath.Join(root, "api/types.proto"))
			require.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}
