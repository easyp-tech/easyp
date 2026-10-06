package generation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestAutomaticSelectionAcceptsInternalLinkedGenerator(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "generator-source.yaml"), []byte("version: v1\n"), 0o644))
	require.NoError(t, os.Symlink("generator-source.yaml", filepath.Join(root, "easyp.gen.yaml")))
	configs, err := selectGenerateConfigs(Request{WorkDir: root, WorkspaceRoot: root})
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Join(root, "easyp.gen.yaml")}, configs)
}

func TestModuleDiscoveryRejectsDuplicateDirectoryAliasIdentity(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "real"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "real/protobuf.mod"), []byte("module example.com/api\nroots .\n"), 0o644))
	require.NoError(t, os.Symlink("real", filepath.Join(root, "alias")))
	_, err := findV1LocalModuleByName(root, "example.com/api")
	require.ErrorContains(t, err, "declared in both")
}

func TestGenerateDirectoryAndMetadataAliasesPreserveDescriptorNames(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "sources/api/model.proto", "syntax = \"proto3\"; package model.v1; message Model {}")
	writeV1GenerateFixture(t, root, "generator-source.yaml", "version: v1\n")
	writeV1GenerateFixture(t, root, "manifest-source", "module example.com/api\nroots proto\n")
	require.NoError(t, os.Symlink("sources", filepath.Join(root, "proto")))
	require.NoError(t, os.Symlink("generator-source.yaml", filepath.Join(root, "easyp.gen.yaml")))
	require.NoError(t, os.Symlink("manifest-source", filepath.Join(root, "protobuf.mod")))
	require.NoError(t, os.Symlink("missing", filepath.Join(root, "sources/unused.txt")))
	output := filepath.Join(root, "descriptor.pb")
	require.NoError(t, Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, WorkspaceRoot: root, DescriptorSetOut: output}))
	raw, err := os.ReadFile(output)
	require.NoError(t, err)
	var descriptors descriptorpb.FileDescriptorSet
	require.NoError(t, proto.Unmarshal(raw, &descriptors))
	require.Len(t, descriptors.File, 1)
	require.Equal(t, "api/model.proto", descriptors.File[0].GetName())
}

func TestGenerateSelectedFilesSkipsDanglingUnselectedAlias(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "api/model.proto", "syntax = \"proto3\"; package model.v1; message Model {}")
	writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\ngenerate:\n  paths: [api]\n")
	require.NoError(t, os.Symlink("missing", filepath.Join(root, "unselected.proto")))
	require.NoError(t, Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, WorkspaceRoot: root, DescriptorSetOut: filepath.Join(root, "descriptor.pb")}))
}

func TestGenerateDeclaredDependencyAliasAcrossNestedRepository(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.test/app\nroots .\nrequire example.test/dep\nreplace example.test/dep => ./dep\n")
	writeV1GenerateFixture(t, root, "dep/.git", "gitdir: /unused\n")
	writeV1GenerateFixture(t, root, "dep/buf.yaml", "version: v2\nmodules:\n  - path: .\n    includes: [selected]\n")
	writeV1GenerateFixture(t, root, "dep/selected/model.proto", "syntax = \"proto3\"; package model.v1; message Model {}\n")
	writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\ngenerate:\n  paths: [alias.proto]\n  packages: [model.v1]\n")
	require.NoError(t, os.Symlink("dep/selected/model.proto", filepath.Join(root, "alias.proto")))
	require.NoError(t, Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, WorkspaceRoot: root, DescriptorSetOut: filepath.Join(root, "descriptor.pb")}))
}

func TestModuleDiscoveryIgnoresUnselectedProtoAliasFailures(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.test/app\nroots api\n")
	writeV1GenerateFixture(t, root, "dep/protobuf.mod", "module example.test/dep\nroots allowed\n")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "dep/excluded"), 0o755))
	require.NoError(t, os.Symlink("../../../outside.proto", filepath.Join(root, "dep/excluded/alias.proto")))
	found, err := findV1LocalModuleByName(root, "example.test/dep")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(root, "dep"), found)
}

func TestGenerateSelectedBufDependencyIgnoresExcludedAliasFailures(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, metadata string }{
		{name: "v1 excluded", metadata: "version: v1\nbuild:\n  excludes: [excluded]\n"},
		{name: "v2 not included", metadata: "version: v2\nmodules:\n  - path: .\n    includes: [allowed]\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.test/app\nroots api\nrequire example.test/dep\nreplace example.test/dep => ./dep\n")
			writeV1GenerateFixture(t, root, "api/client.proto", "syntax = \"proto3\"; package app; message Client {}\n")
			writeV1GenerateFixture(t, root, "dep/buf.yaml", tt.metadata)
			writeV1GenerateFixture(t, root, "dep/allowed/model.proto", "syntax = \"proto3\"; package model; message Model {}\n")
			writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\ngenerate:\n  modules: [example.test/dep]\n")
			require.NoError(t, os.MkdirAll(filepath.Join(root, "dep/excluded"), 0o755))
			require.NoError(t, os.Symlink("../../../outside.proto", filepath.Join(root, "dep/excluded/alias.proto")))
			output := filepath.Join(root, "descriptor.pb")
			require.NoError(t, Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, WorkspaceRoot: root, DescriptorSetOut: output}))
			raw, err := os.ReadFile(output)
			require.NoError(t, err)
			var descriptors descriptorpb.FileDescriptorSet
			require.NoError(t, proto.Unmarshal(raw, &descriptors))
			require.Len(t, descriptors.File, 1)
			require.Equal(t, "allowed/model.proto", descriptors.File[0].GetName())
		})
	}
}
