package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestRunPathSelectionKeepsRequiredImports(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		include bool
	}{
		{name: "selected_sources"},
		{name: "include_imports", include: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			files := map[string]string{
				"protobuf.mod":             "module example.com/app\nroots proto\n",
				"easyp.gen.yaml":           "version: v1\ngenerate:\n  paths: [api]\n  packages: [api.v1]\n",
				"proto/api/service.proto":  "syntax = \"proto3\"; package api.v1; import \"common/types.proto\"; message Service { common.v1.Type type = 1; }",
				"proto/common/types.proto": "syntax = \"proto3\"; package common.v1; message Type {}",
				"proto/api-copy/bad.proto": "syntax = \"proto3\"; package api.v1; message Broken {",
				"proto/other/bad.proto":    "this unrelated source is malformed",
			}
			for _, directory := range []string{"module-lite", "module-full", "module-kotlin", "module-java"} {
				files["proto/examples/jvm/"+directory+"/build/service.proto"] = files["proto/api/service.proto"]
			}
			for path, content := range files {
				writeV1GenerateFixture(t, root, path, content)
			}
			out := filepath.Join(root, "descriptor.pb")

			require.NoError(t, Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: out, IncludeImports: tt.include}))

			want := []string{"api/service.proto"}
			if tt.include {
				want = append(want, "common/types.proto")
			}
			assert.ElementsMatch(t, want, descriptorNames(readDescriptorSet(t, out)))
		})
	}
}

func TestRunSourceSelectorsMatchAcrossModules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		packages []string
		paths    []string
		want     []string
		wantErr  string
	}{
		{name: "aggregate_each_selector_across_modules", packages: []string{"first.v1", "second.v1"}, paths: []string{"api/first.proto", "api/second.proto"}, want: []string{"api/first.proto", "api/second.proto"}},
		{name: "skip_module_without_combined_match", packages: []string{"first.v1"}, paths: []string{"api"}, want: []string{"api/first.proto"}},
		{name: "unmatched_package_after_path_filter", packages: []string{"first.v1", "second.v1"}, paths: []string{"api/first.proto"}, wantErr: "generate.packages did not match any selected module source files: second.v1"},
		{name: "unmatched_path_after_package_filter", packages: []string{"first.v1"}, paths: []string{"api/first.proto", "api/second.proto"}, wantErr: "generate.paths did not match any selected module source files: api/second.proto"},
		{name: "dot_with_package_filter", packages: []string{"first.v1"}, paths: []string{"."}, want: []string{"api/first.proto"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "easyp.gen.yaml", fmt.Sprintf("version: v1\ngenerate:\n  modules: [m1, m2]\n  packages: [%s]\n  paths: [%s]\n", strings.Join(tt.packages, ", "), strings.Join(tt.paths, ", ")))
			writeV1GenerateFixture(t, root, "m1/protobuf.mod", "module example.com/first\nroots proto\n")
			writeV1GenerateFixture(t, root, "m1/proto/api/first.proto", "syntax = \"proto3\"; package first.v1; message First {}")
			writeV1GenerateFixture(t, root, "m2/protobuf.mod", "module example.com/second\nroots proto\n")
			writeV1GenerateFixture(t, root, "m2/proto/api/second.proto", "syntax = \"proto3\"; package second.v1; message Second {}")
			out := filepath.Join(root, "descriptor.pb")
			previous := []byte("previous descriptor")
			require.NoError(t, os.WriteFile(out, previous, 0o600))

			err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: out})

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				actual, readErr := os.ReadFile(out)
				require.NoError(t, readErr)
				assert.Equal(t, previous, actual)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, descriptorNames(readDescriptorSet(t, out)))
		})
	}
}

func TestRunUnknownPathWithoutPlugins(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\ngenerate:\n  paths: [missing, missing]\n")
	writeV1GenerateFixture(t, root, "item.proto", "syntax = \"proto3\"; package item.v1; message Item {}")

	err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root})

	require.ErrorContains(t, err, "generate.paths did not match any selected module source files: missing")
	assert.NotContains(t, err.Error(), "missing, missing")
}

func TestRunAllValidatesSelectorsOnOptionsOnlyParentBeforeChildren(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable uses a POSIX shell")
	}
	for _, tt := range []struct {
		name     string
		selector string
		export   bool
	}{
		{name: "paths_without_export", selector: "paths"},
		{name: "paths_with_export", selector: "paths", export: true},
		{name: "packages_without_export", selector: "packages"},
		{name: "packages_with_export", selector: "packages", export: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\ngenerate:\n  "+tt.selector+": [missing]\noptions:\n  go:\n    package_prefix: example.com/gen\n")
			writeV1GenerateFixture(t, root, "legacy/orphan.proto", "syntax = \"proto3\"; package orphan.v1; message Orphan {}")
			writeV1GenerateFixture(t, root, "child/protobuf.mod", "module example.com/child\n")
			writeV1GenerateFixture(t, root, "child/easyp.gen.yaml", "version: v1\nplugins:\n  - path: ./child-plugin\n    out: gen\n")
			writeV1GenerateFixture(t, root, "child/item.proto", "syntax = \"proto3\"; package child.v1; message Item {}")
			require.NoError(t, os.WriteFile(filepath.Join(root, "child-plugin"), []byte("#!/bin/sh\ncat >/dev/null\nprintf called > child-ran.txt\n"), 0o755))
			request := Request{WorkDir: root, AllProjects: true}
			out := filepath.Join(root, "all.pb")
			previous := []byte("previous descriptor")
			if tt.export {
				require.NoError(t, os.WriteFile(out, previous, 0o600))
				request.DescriptorSetOut = out
			}

			err := Run(t.Context(), logger.NewNop(), nil, request)

			require.ErrorContains(t, err, "generate."+tt.selector+" did not match any selected module source files: missing")
			assert.NoFileExists(t, filepath.Join(root, "child-ran.txt"))
			if tt.export {
				actual, readErr := os.ReadFile(out)
				require.NoError(t, readErr)
				assert.Equal(t, previous, actual)
			}
		})
	}
}

func TestSelectedSourceFilesSkipsReadingPathExcludedFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "api/item.proto", "syntax = \"proto3\"; package item.v1; message Item {}")
	require.NoError(t, os.Symlink(filepath.Join(root, "absent.proto"), filepath.Join(root, "unreadable.proto")))
	selected := v1GenerationModule{directory: root, module: v1.Module{Name: "example.com/app", Roots: []string{"."}}}

	files, packages, paths, err := selectedSourceFiles(t.Context(), selected, []string{"item.v1"}, []string{"api"})

	require.NoError(t, err)
	assert.Equal(t, []string{"api/item.proto"}, files)
	assert.Equal(t, map[string]bool{"item.v1": true}, packages)
	assert.Equal(t, map[string]bool{"api": true}, paths)
}

func TestSelectedSourceFilesKeepsSourceBoundaries(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "api/item.proto", "syntax = \"proto3\"; package item.v1; message Item {}")
	writeV1GenerateFixture(t, root, ".sources/hidden.proto", "syntax = \"proto3\"; package item.v1; message Hidden {}")
	writeV1GenerateFixture(t, root, "nested/protobuf.mod", "module example.com/nested\n")
	writeV1GenerateFixture(t, root, "nested/nested.proto", "syntax = \"proto3\"; package item.v1; message Nested {}")
	selected := v1GenerationModule{directory: root, module: v1.Module{Name: "example.com/app", Roots: []string{"."}}}

	files, packages, paths, err := selectedSourceFiles(t.Context(), selected, []string{"item.v1"}, []string{".", ".sources", "nested"})

	require.NoError(t, err)
	assert.Equal(t, []string{"api/item.proto"}, files)
	assert.Equal(t, map[string]bool{"item.v1": true}, packages)
	assert.Equal(t, map[string]bool{".": true}, paths)
}
