package migration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestMigrationPathsPreserveGradleCopies(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	legacy := "generate:\n  inputs: [{directory: {path: mcp, root: .}}]\n  plugins: [{name: go, out: ., opts: {paths: source_relative}}]\n"
	writeFixture(t, root, "easyp.yaml", legacy)
	writeFixture(t, root, ".gitignore", "examples/jvm/*/build/\n")
	const source = "syntax = \"proto3\"; package mcp.options.v1; message Options {}\n"
	files := []string{"mcp/options/v1/options.proto"}
	for _, module := range []string{"java-server", "kotlin-server"} {
		for _, kind := range []string{"resources/main", "staged-proto"} {
			files = append(files, "examples/jvm/"+module+"/build/"+kind+"/mcp/options/v1/options.proto")
		}
	}
	for _, name := range files {
		writeFixture(t, root, name, source)
	}
	writeFixture(t, root, "other.proto", "package other.v1; message {\n")
	plan, err := Build(t.Context(), Options{Dir: root, Module: "example.com/sdk"})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"mcp/options/v1/options.proto": "mcp/options/v1/options.proto"}, plan.sources)
	var output struct {
		Generate struct {
			Paths    []string
			Packages []string
		}
	}
	require.NoError(t, yaml.Unmarshal(outputContent(t, plan, v1.GenerateFile), &output))
	assert.Equal(t, []string{"mcp"}, output.Generate.Paths)
	assert.Empty(t, output.Generate.Packages)
	require.NoError(t, plan.Apply())
	assert.Equal(t, legacy, string(mustRead(t, root, "easyp.yaml.v0.bak")))
	for _, name := range files {
		assert.Equal(t, source, string(mustRead(t, root, name)))
	}
}

func TestMigrationPathsStayFixedWhenOutsideCopyAppears(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFixture(t, root, "easyp.yaml", "generate:\n  inputs: [{directory: {path: mcp}}]\n")
	writeFixture(t, root, "mcp/options.proto", "package mcp.options.v1;")
	plan, err := Build(t.Context(), Options{Dir: root, Module: "example.com/sdk"})
	require.NoError(t, err)
	require.Equal(t, []string{"mcp"}, plan.paths, "a clean tree must retain literal input scope for future builds")
	writeFixture(t, root, "build/copied.proto", "package mcp.options.v1;")
	require.NoError(t, plan.CheckUnchanged())
	require.NoError(t, plan.Apply())
	assert.FileExists(t, filepath.Join(root, "build/copied.proto"))
}

func TestMigrationPathsPreservePartialOrUndeclaredPackages(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, content string }{
		{name: "same package outside targets", content: "package shared.v1; message Selected {}"},
		{name: "no package declaration", content: "syntax = \"proto3\"; message Selected {}"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFixture(t, root, "easyp.yaml", "generate:\n  inputs: [{directory: selected}]\n")
			writeFixture(t, root, "selected/a.proto", tt.content)
			writeFixture(t, root, "other.proto", "package shared.v1; message Other {}")
			plan, err := Build(t.Context(), Options{Dir: root, Module: "example.test/api"})
			require.NoError(t, err)
			assert.Equal(t, []string{"selected"}, plan.paths)
			assert.Equal(t, map[string]string{"selected/a.proto": "selected/a.proto"}, plan.sources)
			require.NoError(t, plan.Apply())
			assert.Equal(t, tt.content, string(mustRead(t, root, "selected/a.proto")))
		})
	}
}

func TestMigrationPackageFallbackPreservesMixedRootSelection(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFixture(t, root, "easyp.yaml", "generate:\n  inputs: [{directory: {path: ., root: a}}, {directory: {path: selected, root: b}}]\n")
	writeFixture(t, root, "a/first.proto", "package first.v1;")
	writeFixture(t, root, "b/selected/second.proto", "package second.v1;")
	writeFixture(t, root, "b/outside/other.proto", "package other.v1;")
	plan, err := Build(t.Context(), Options{Dir: root, Module: "example.test/api"})
	require.NoError(t, err)
	assert.Empty(t, plan.paths, "dot paths would include unselected files in the second root")
	assert.Equal(t, []string{"first.v1", "second.v1"}, plan.packages)
	assert.Equal(t, map[string]string{"first.proto": "a/first.proto", "selected/second.proto": "b/selected/second.proto"}, plan.sources)
	require.NoError(t, plan.Apply())
}
