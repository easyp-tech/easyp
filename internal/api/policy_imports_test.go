package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyImportRoots(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		sourceRoot     string
		dependencyRoot string
	}{
		{name: "separate source roots", sourceRoot: "proto", dependencyRoot: "src"},
		{name: "module directory roots", sourceRoot: ".", dependencyRoot: "."},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, dependency := t.TempDir(), t.TempDir()
			replacement, err := filepath.Rel(root, dependency)
			require.NoError(t, err)
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/root\nroots "+tt.sourceRoot+"\nrequire example.com/dep v1.0.0\nreplace example.com/dep => "+replacement+"\n")
			writeV1GenerateFixture(t, root, filepath.Join(tt.sourceRoot, "root.proto"), `syntax = "proto3";`)
			writeV1GenerateFixture(t, dependency, "protobuf.mod", "module example.com/dep\nroots "+tt.dependencyRoot+"\n")
			writeV1GenerateFixture(t, dependency, filepath.Join(tt.dependencyRoot, "dep.proto"), `syntax = "proto3";`)

			roots, err := ensureV1PolicyImportRoots(t.Context(), nil, root)

			require.NoError(t, err)
			assert.Equal(t, []string{filepath.Join(root, tt.sourceRoot), filepath.Join(dependency, tt.dependencyRoot)}, roots)
		})
	}
}

func TestFindPolicyModule(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		manifest      string
		wantDirectory string
	}{
		{name: "no module"},
		{name: "repository module", manifest: "protobuf.mod", wantDirectory: "."},
		{name: "nearest module", manifest: "api/protobuf.mod", wantDirectory: "api"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			if tt.manifest != "" {
				writeV1GenerateFixture(t, root, tt.manifest, "module example.com/root\n")
			}
			want := ""
			if tt.wantDirectory != "" {
				want = filepath.Join(root, tt.wantDirectory)
			}

			moduleDir, err := findV1PolicyModuleDir(root, filepath.Join(root, "api", "proto"))

			require.NoError(t, err)
			assert.Equal(t, want, moduleDir)
		})
	}
}
