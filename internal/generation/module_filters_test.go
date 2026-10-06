package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestScopedSelectorsValidateWithoutPlugins(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, field string }{
		{name: "paths", field: "paths"},
		{name: "packages", field: "packages"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/app\n")
			writeV1GenerateFixture(t, root, "easyp.gen.yaml", "generate:\n  modules:\n    - module: .\n      "+tt.field+": [missing]\n")
			writeV1GenerateFixture(t, root, "item.proto", "syntax = \"proto3\"; package item.v1; message Item {}")

			err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root})

			require.ErrorContains(t, err, "generate.modules[0]."+tt.field)
		})
	}
}

func TestAllChecksScopedParentSelectorsBeforeChildPlugins(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/parent\n")
	writeV1GenerateFixture(t, root, "easyp.gen.yaml", "generate:\n  modules: [{module: ., paths: [missing]}]\n")
	writeV1GenerateFixture(t, root, "child/protobuf.mod", "module example.com/child\n")
	writeV1GenerateFixture(t, root, "child/easyp.gen.yaml", "plugins: [{command: [sh, -c, 'touch plugin-ran'], out: gen}]\n")
	writeV1GenerateFixture(t, root, "child/item.proto", "syntax = \"proto3\"; package child.v1; message Item {}")

	err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, AllProjects: true})

	require.ErrorContains(t, err, "generate.modules[0].paths")
	assert.NoFileExists(t, filepath.Join(root, "child/plugin-ran"))
}

func TestModuleDuplicateFiltersAreUnambiguous(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, first, second string
		conflict            bool
	}{
		{name: "same_unfiltered_forms", first: "contracts", second: "{module: contracts, paths: [], packages: []}"},
		{name: "reordered_repeated_filters", first: "{module: contracts, paths: [proto, proto/api]}", second: "{module: example.com/contracts, paths: [proto/api, proto, proto]}"},
		{name: "unfiltered_then_filtered", first: "contracts", second: "{module: contracts, paths: [proto/api/first.proto]}", conflict: true},
		{name: "filtered_then_unfiltered", first: "{module: contracts, paths: [proto/api/first.proto]}", second: "contracts", conflict: true},
		{name: "directory_then_identity", first: "{module: contracts, paths: [proto/api/first.proto]}", second: "{module: example.com/contracts, paths: [proto/api/second.proto]}", conflict: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "easyp.gen.yaml", fmt.Sprintf("generate:\n  modules: [%s, %s]\n", tt.first, tt.second))
			writeV1GenerateFixture(t, root, "contracts/protobuf.mod", "module example.com/contracts\nroots proto\n")
			writeV1GenerateFixture(t, root, "contracts/proto/api/first.proto", "syntax = \"proto3\"; package api.v1; message First {}")
			writeV1GenerateFixture(t, root, "contracts/proto/api/second.proto", "syntax = \"proto3\"; package api.v1; message Second {}")
			out := filepath.Join(root, "selected.pb")
			previous := []byte("previous descriptor")
			require.NoError(t, os.WriteFile(out, previous, 0o600))

			err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: out})

			if tt.conflict {
				require.ErrorContains(t, err, "generate.modules[0] and generate.modules[1]")
				require.ErrorContains(t, err, "conflicting filters")
				actual, readErr := os.ReadFile(out)
				require.NoError(t, readErr)
				assert.Equal(t, previous, actual)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, []string{"api/first.proto", "api/second.proto"}, descriptorNames(readDescriptorSet(t, out)))
		})
	}
}

func TestModulePathsUseLogicalRootAlias(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, path string
		reject     bool
	}{
		{name: "logical_alias", path: "proto/api"},
		{name: "physical_target_is_not_a_selector", path: "source/api", reject: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/app\nroots proto\n")
			writeV1GenerateFixture(t, root, "easyp.gen.yaml", "generate:\n  modules: [{module: ., paths: ["+tt.path+"]}]\n")
			writeV1GenerateFixture(t, root, "source/api/model.proto", "syntax = \"proto3\"; package api.v1; message Model {}")
			require.NoError(t, os.Symlink("source", filepath.Join(root, "proto")))
			out := filepath.Join(root, "selected.pb")

			err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: out})

			if tt.reject {
				require.ErrorContains(t, err, "generate.modules[0].paths")
				assert.NoFileExists(t, out)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []string{"api/model.proto"}, descriptorNames(readDescriptorSet(t, out)))
		})
	}
}

func TestModuleFiltersIntersectGlobalSelectors(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	files := map[string]string{
		"easyp.gen.yaml":                 "version: v1\ngenerate:\n  modules:\n    - module: first\n      paths: [proto/api/first.proto]\n      packages: [first.v1]\n    - module: second\n      paths: [proto/api/second.proto]\n      packages: [second.v1]\n  paths: [proto/api]\n  packages: [first.v1, second.v1]\n",
		"first/protobuf.mod":             "module example.com/first\nroots proto\n",
		"second/protobuf.mod":            "module example.com/second\nroots proto\n",
		"first/proto/api/first.proto":    "syntax = \"proto3\"; package first.v1; import \"common/type.proto\"; message First { common.v1.Type value = 1; }",
		"second/proto/api/second.proto":  "syntax = \"proto3\"; package second.v1; import \"common/type.proto\"; message Second { common.v1.Type value = 1; }",
		"first/proto/api/ignored.proto":  "syntax = \"proto3\"; package first.v1; message Broken {",
		"second/proto/api/ignored.proto": "syntax = \"proto3\"; package second.v1; message Broken {",
	}
	for _, module := range []string{"first", "second"} {
		files[module+"/proto/common/type.proto"] = "syntax = \"proto3\"; package common.v1; message Type {}"
	}
	for name, content := range files {
		writeV1GenerateFixture(t, root, name, content)
	}
	out := filepath.Join(root, "selected.pb")

	require.NoError(t, Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: out, IncludeImports: true}))

	assert.ElementsMatch(t, []string{"api/first.proto", "api/second.proto", "common/type.proto"}, descriptorNames(readDescriptorSet(t, out)))
}

func TestModuleFilterCannotMatchAnotherModule(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "easyp.gen.yaml", "generate:\n  modules:\n    - module: first\n      paths: [proto/api/second.proto]\n    - second\nplugins: [{command: [sh, -c, 'touch plugin-ran'], out: gen}]\n")
	for _, module := range []string{"first", "second"} {
		writeV1GenerateFixture(t, root, module+"/protobuf.mod", "module example.com/"+module+"\nroots proto\n")
		writeV1GenerateFixture(t, root, module+"/proto/api/"+module+".proto", "syntax = \"proto3\"; package "+module+".v1; message Item {}")
	}
	out := filepath.Join(root, "selected.pb")
	previous := []byte("previous descriptor")
	require.NoError(t, os.WriteFile(out, previous, 0o600))

	err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: out})

	require.ErrorContains(t, err, "generate.modules[0].paths")
	assert.NoFileExists(t, filepath.Join(root, "plugin-ran"))
	actual, readErr := os.ReadFile(out)
	require.NoError(t, readErr)
	assert.Equal(t, previous, actual)
}

func TestGlobalPathsUseModuleDirectoryWithoutExplicitModule(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/service\nroots api\n")
	writeV1GenerateFixture(t, root, "easyp.gen.yaml", "generate:\n  paths: [api/easyp]\n")
	writeV1GenerateFixture(t, root, "api/easyp/generator/v1/generator.proto", "syntax = \"proto3\"; package easyp.generator.v1; message Generator {}")
	writeV1GenerateFixture(t, root, "api/other/bad.proto", "package invalid.v1; message {")
	out := filepath.Join(root, "selected.pb")

	require.NoError(t, Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: out}))

	assert.Equal(t, []string{"easyp/generator/v1/generator.proto"}, descriptorNames(readDescriptorSet(t, out)))
}
