package migration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestMigrationPackageSelectionPreservesLegacyInputs(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, inputs string
		files        map[string]string
		wantPaths    []string
		wantFiles    map[string]string
	}{
		{
			name:   "explicit default root",
			inputs: "[{directory: {path: mcp, root: .}}]",
			files: map[string]string{
				"mcp/options/v1/options.proto":     "syntax = \"proto3\"; package mcp.options.v1; message Options {}",
				"internal/testproto/example.proto": "syntax = \"proto3\"; package example.v1; message Example {}",
				"testdata/unsupported.proto":       "package unsupported.v1; message {",
			},
			wantPaths: []string{"mcp"},
			wantFiles: map[string]string{"mcp/options/v1/options.proto": "mcp/options/v1/options.proto"},
		},
		{
			name:      "omitted root defaults to dot",
			inputs:    "[{directory: {path: mcp}}]",
			files:     map[string]string{"mcp/a.proto": "package mcp.v1;", "other.proto": "package other.v1;"},
			wantPaths: []string{"mcp"}, wantFiles: map[string]string{"mcp/a.proto": "mcp/a.proto"},
		},
		{
			name:      "empty root defaults to dot",
			inputs:    "[{directory: {path: mcp, root: ''}}]",
			files:     map[string]string{"mcp/a.proto": "package mcp.v1;", "other.proto": "package other.v1;"},
			wantPaths: []string{"mcp"}, wantFiles: map[string]string{"mcp/a.proto": "mcp/a.proto"},
		},
		{
			name:   "complete packages and overlapping inputs",
			inputs: "[{directory: {path: selected, root: .}}, {directory: {path: selected/a, root: .}}]",
			files: map[string]string{
				"selected/a/first.proto":  "/* package fake; */ package z.v1;",
				"selected/a/second.proto": "package z.v1;",
				"selected/b/third.proto":  "option java_package = \"package fake;\"; package a.v1;",
				"other.proto":             "package other.v1;",
			},
			wantPaths: []string{"selected", "selected/a"},
			wantFiles: map[string]string{"selected/a/first.proto": "selected/a/first.proto", "selected/a/second.proto": "selected/a/second.proto", "selected/b/third.proto": "selected/b/third.proto"},
		},
		{
			name:      "whole root keeps selection omitted",
			inputs:    "[{directory: .}]",
			files:     map[string]string{"api.proto": "syntax = \"proto3\";"},
			wantFiles: map[string]string{"api.proto": "api.proto"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			legacy := "generate:\n  inputs: " + tt.inputs + "\n  plugins: [{name: go, out: ., opts: {paths: source_relative}}]\n"
			writeFixture(t, root, "easyp.yaml", legacy)
			for name, content := range tt.files {
				writeFixture(t, root, name, content)
			}
			plan, err := Build(t.Context(), Options{Dir: root, Module: "example.com/acme/api"})
			require.NoError(t, err)
			assert.Equal(t, tt.wantFiles, plan.local.selection.files)
			gen, err := v1.ParseGenerate(bytes.NewReader(outputContent(t, plan, v1.GenerateFile)))
			require.NoError(t, err)
			assert.Empty(t, gen.Generate.Packages)
			assert.Equal(t, tt.wantPaths, gen.Generate.Paths)
			require.Len(t, gen.Plugins, 1)
			assert.Equal(t, ".", gen.Plugins[0].Out)
			assert.Equal(t, []string{"source_relative"}, gen.Plugins[0].Opts["paths"])
			module, err := v1.ParseModule(bytes.NewReader(outputContent(t, plan, v1.ModuleFile)))
			require.NoError(t, err)
			assert.Equal(t, []string{"."}, module.Roots)
			require.NoError(t, plan.Apply())
			assert.Equal(t, legacy, string(mustRead(t, root, "easyp.yaml.v0.bak")))
			for name, content := range tt.files {
				assert.Equal(t, content, string(mustRead(t, root, name)), "source %s must stay at its original path", name)
			}
		})
	}
}

func TestMigrationPackageSelectionRejectsChangedScope(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, path, selected, other, extraInput string
		files                                   map[string]string
	}{
		{name: "hidden source", path: ".selected", selected: "package selected.v1;", other: "package other.v1;"},
		{name: "vendor source", path: "easyp_vendor", selected: "package selected.v1;", other: "package other.v1;"},
		{name: "nested module", path: "selected", selected: "package selected.v1;", other: "package other.v1;", files: map[string]string{"selected/protobuf.mod": "module example.com/nested\n"}},
		{name: "empty subtree cannot select no packages", path: "selected", other: "package other.v1;"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			legacy := "generate:\n  inputs: [{directory: {path: " + tt.path + ", root: .}}" + tt.extraInput + "]\n"
			writeFixture(t, root, "easyp.yaml", legacy)
			if tt.selected == "" {
				writeFixture(t, root, filepath.Join(tt.path, ".keep"), "")
			} else {
				writeFixture(t, root, filepath.Join(tt.path, "a.proto"), tt.selected)
			}
			writeFixture(t, root, "other.proto", tt.other)
			for name, content := range tt.files {
				writeFixture(t, root, name, content)
			}
			repo := &mockRepository{}
			_, err := Build(t.Context(), Options{Dir: root, Module: "example.com/acme/api", ResolveLock: true, Repository: repo})
			require.ErrorContains(t, err, "scope")
			assert.Empty(t, repo.calls, "scope must be checked before dependency access")
			assert.Equal(t, legacy, string(mustRead(t, root, "easyp.yaml")))
			for _, name := range []string{v1.GenerateFile, v1.ModuleFile, v1.LockFile, "easyp.yaml.v0.bak"} {
				_, err := os.Stat(filepath.Join(root, name))
				assert.ErrorIs(t, err, os.ErrNotExist)
			}
		})
	}
}

func TestMigrationPackageSelectionRejectsCollidingImportRoots(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	legacy := "generate:\n  inputs: [{directory: {path: ., root: a}}, {directory: {path: ., root: b}}]\n"
	writeFixture(t, root, "easyp.yaml", legacy)
	writeFixture(t, root, "a/same.proto", "package first.v1;")
	writeFixture(t, root, "b/same.proto", "package second.v1;")
	repo := &mockRepository{}
	_, err := Build(t.Context(), Options{Dir: root, Module: "example.com/acme/api", ResolveLock: true, Repository: repo})
	require.ErrorContains(t, err, "roots collide")
	assert.Empty(t, repo.calls)
	assert.Equal(t, legacy, string(mustRead(t, root, "easyp.yaml")))
	assert.NoFileExists(t, filepath.Join(root, v1.ModuleFile))
}

func TestMigrationPackageSelectionRechecksFixedSelectors(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	legacy := "generate:\n  inputs: [{directory: {path: selected}}]\n"
	writeFixture(t, root, "easyp.yaml", legacy)
	writeFixture(t, root, "selected/a.proto", "package selected.v1;")
	writeFixture(t, root, "other.proto", "package other.v1;")
	plan, err := Build(t.Context(), Options{Dir: root, Module: "example.com/acme/api"})
	require.NoError(t, err)
	require.Equal(t, []string{"selected"}, plan.local.selection.paths)
	require.NoError(t, os.Remove(filepath.Join(root, "other.proto")))
	recomputed, err := proveLocalSelection(root, plan.local.inputs, plan.local.roots)
	require.NoError(t, err)
	require.Equal(t, plan.local.selection.files, recomputed.files, "selected paths and bytes are unchanged")
	require.Equal(t, plan.local.selection.paths, recomputed.paths, "removing outside files must not drop the approved directory filter")
	require.NoError(t, plan.CheckUnchanged())
	require.NoError(t, plan.Apply())
	assert.Equal(t, legacy, string(mustRead(t, root, "easyp.yaml.v0.bak")))
}

func TestMigrationPackageSelectionRechecksInitiallyEmptyTree(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	legacy := "generate:\n  inputs: [{directory: {path: selected}}]\n"
	writeFixture(t, root, "easyp.yaml", legacy)
	writeFixture(t, root, "selected/.keep", "")
	plan, err := Build(t.Context(), Options{Dir: root, Module: "example.com/acme/api"})
	require.NoError(t, err)
	writeFixture(t, root, "other.proto", "package other.v1;")
	require.Error(t, plan.CheckUnchanged())
	require.Error(t, plan.Apply())
	assert.Equal(t, legacy, string(mustRead(t, root, "easyp.yaml")))
	assert.NoFileExists(t, filepath.Join(root, v1.ModuleFile))
}

func TestMigrationPackageSelectionRechecksUnselectedDeclarations(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, changed string
	}{
		{name: "new selected source", changed: "selected/new.proto"},
		{name: "selected package changes", changed: "selected/a.proto"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			legacy := "generate:\n  inputs: [{directory: {path: selected}}]\n"
			writeFixture(t, root, "easyp.yaml", legacy)
			writeFixture(t, root, "selected/a.proto", "package selected.v1;")
			writeFixture(t, root, "other.proto", "package other.v1;")
			plan, err := Build(t.Context(), Options{Dir: root, Module: "example.com/acme/api"})
			require.NoError(t, err)
			content := "package selected.v1; message Added {}"
			if tt.changed == "selected/a.proto" {
				content = "package renamed.v1;"
			}
			writeFixture(t, root, tt.changed, content)
			require.Error(t, plan.CheckUnchanged())
			require.Error(t, plan.Apply())
			assert.Equal(t, legacy, string(mustRead(t, root, "easyp.yaml")))
			_, err = os.Stat(filepath.Join(root, v1.ModuleFile))
			assert.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}
