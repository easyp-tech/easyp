package modules

import (
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
	"github.com/stretchr/testify/require"
)

func TestSourceRootWalkPreservesInternalDirectoryAlias(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeV1GenerateFixture(t, directory, "sources/api/model.proto", "syntax = \"proto3\";\n")
	require.NoError(t, os.Symlink("sources", filepath.Join(directory, "proto")))
	roots, err := ModuleSources(directory, v1.Module{Name: "example.com/api", Roots: []string{"proto"}})
	require.NoError(t, err)
	var paths []string
	require.NoError(t, roots[0].Walk(func(path string) error { paths = append(paths, path); return nil }))
	require.Equal(t, []string{filepath.Join(directory, "proto/api/model.proto")}, paths)
}

func TestSourceRootWalkRejectsExternalProtoAlias(t *testing.T) {
	t.Parallel()
	directory, outside := t.TempDir(), t.TempDir()
	writeV1GenerateFixture(t, outside, "secret.proto", "syntax = \"proto3\";\n")
	require.NoError(t, os.Symlink(filepath.Join(outside, "secret.proto"), filepath.Join(directory, "alias.proto")))
	roots, err := ModuleSources(directory, v1.Module{Name: "example.com/api", Roots: []string{"."}})
	require.NoError(t, err)
	require.Error(t, roots[0].Walk(func(string) error { return nil }))
}

func TestFilteredFileDoesNotSkipSelectedSiblings(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeV1GenerateFixture(t, directory, "a.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, directory, "b.proto", "syntax = \"proto3\";\n")
	roots, err := ModuleSources(directory, v1.Module{Name: "example.com/api", Roots: []string{"."}, ProtoFilters: []v1.ProtoFileFilter{{Root: ".", Excludes: []string{"a.proto"}}}})
	require.NoError(t, err)
	var paths []string
	require.NoError(t, roots[0].Walk(func(path string) error { paths = append(paths, path); return nil }))
	require.Equal(t, []string{filepath.Join(directory, "b.proto")}, paths)
}

func TestDeclaredDependencyAliasAcrossNestedRepository(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.test/app\nroots .\nrequire example.test/dep\nreplace example.test/dep => ./dep\n")
	writeV1GenerateFixture(t, root, "dep/.git", "gitdir: /unused\n")
	writeV1GenerateFixture(t, root, "dep/buf.yaml", "version: v2\nmodules:\n  - path: .\n    includes: [selected]\n")
	writeV1GenerateFixture(t, root, "dep/selected/model.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, root, "client.proto", "syntax = \"proto3\"; import \"alias.proto\";\n")
	require.NoError(t, os.Symlink("dep/selected/model.proto", filepath.Join(root, "alias.proto")))
	require.NoError(t, Tidy(t.Context(), root, nil))
}

func TestImportRootAliasPreservesPhysicalBufExclusions(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeV1GenerateFixture(t, directory, "real/allowed/keep.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, directory, "real/excluded/blocked.proto", "syntax = \"proto3\";\n")
	require.NoError(t, os.Symlink("real", filepath.Join(directory, "proto")))
	require.NoError(t, os.Symlink("../excluded/blocked.proto", filepath.Join(directory, "real/allowed/alias.proto")))
	sources, err := ModuleSources(directory, v1.Module{Name: "example.com/api", Roots: []string{"proto"}, ProtoFilters: []v1.ProtoFileFilter{{Root: "proto", Excludes: []string{"proto/excluded"}}}})
	require.NoError(t, err)
	require.False(t, sources.FileAllowed()(filepath.Join(directory, "proto/allowed/alias.proto")))
	owners, err := sources.FileModules()
	require.NoError(t, err)
	require.Equal(t, map[string]string{"allowed/keep.proto": "example.com/api"}, owners)
}

func TestSourceRootDirectoryAliasCycleFails(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeV1GenerateFixture(t, directory, "file.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, directory, "protobuf.mod", "module example.com/api\nroots proto\n")
	require.NoError(t, os.Symlink(".", filepath.Join(directory, "proto")))
	sources, err := ModuleSources(directory, v1.Module{Name: "example.com/api", Roots: []string{"proto"}})
	require.NoError(t, err)
	require.ErrorIs(t, sources[0].Walk(func(string) error { return nil }), sourceview.ErrCycle)
}

func TestGraphAliasCannotReadUndeclaredNestedRepository(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeV1GenerateFixture(t, directory, "dep/.git", "gitdir: /unused\n")
	writeV1GenerateFixture(t, directory, "dep/file.proto", "syntax = \"proto3\";\n")
	require.NoError(t, os.Mkdir(filepath.Join(directory, "proto"), 0o755))
	require.NoError(t, os.Symlink("../dep/file.proto", filepath.Join(directory, "proto/alias.proto")))
	sources, err := ModuleSources(directory, v1.Module{Name: "example.com/api", Roots: []string{"proto"}})
	require.NoError(t, err)
	require.ErrorIs(t, sources.WalkSelected(sources[0], nil, func(string) error { return nil }), sourceview.ErrNestedRepository)
	_, err = sources.ReadSourceFile(filepath.Join(directory, "proto/alias.proto"))
	require.ErrorIs(t, err, sourceview.ErrNestedRepository)
}

func TestFilteredDirectoryAliasCycleRemainsStrict(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeV1GenerateFixture(t, directory, "allowed/file.proto", "syntax = \"proto3\";\n")
	require.NoError(t, os.Symlink(".", filepath.Join(directory, "cycle")))
	sources, err := ModuleSources(directory, v1.Module{Name: "example.com/api", Roots: []string{"."}, ProtoFilters: []v1.ProtoFileFilter{{Root: ".", Includes: []string{"allowed", "cycle/allowed"}}}})
	require.NoError(t, err)
	_, err = sources.FileModules()
	require.ErrorIs(t, err, sourceview.ErrCycle)
}
