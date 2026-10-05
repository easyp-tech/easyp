package moduleconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadBufDependencyModuleRejectsInvalidPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{name: "root_outside_checkout", config: "version: v2\nmodules:\n  - path: ../outside\n", wantErr: "leaves the repository"},
		{name: "include_outside_root", config: "version: v2\nmodules:\n  - path: proto\n    includes: [other]\n", wantErr: "outside module root"},
		{name: "include_with_matching_prefix", config: "version: v2\nmodules:\n  - path: proto\n    includes: [proto2]\n", wantErr: "outside module root"},
		{name: "exclude_outside_root", config: "version: v2\nmodules:\n  - path: proto\n    excludes: [../outside]\n", wantErr: "outside module root"},
		{name: "exclude_outside_includes", config: "version: v2\nmodules:\n  - path: proto\n    includes: [proto/api]\n    excludes: [proto/test]\n", wantErr: "outside module includes"},
		{name: "empty_include", config: "version: v2\nmodules:\n  - path: proto\n    includes: ['']\n", wantErr: "outside module root"},
		{name: "v1_exclude_outside_checkout", config: "version: v1\nbuild:\n  excludes: [../outside]\n", wantErr: "outside module root"},
		{name: "v1beta1_exclude_outside_roots", config: "version: v1beta1\nbuild:\n  roots: [proto]\n  excludes: [other]\n", wantErr: "outside build.roots"},
		{name: "empty_registry_reference", config: "version: v1\ndeps: ['']\n", wantErr: "deps entry is empty"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), bufModuleConfigFile)
			require.NoError(t, os.WriteFile(path, []byte(tt.config), 0o644))

			_, err := readBufDependencyModule(path)

			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestReadBufDependencyWorkspaceRejectsSymlinkAncestorBeforeMetadataRead(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		link       string
		root       string
		targetRoot string
	}{
		{name: "directory_link", link: "proto", root: "proto", targetRoot: "."},
		{name: "ancestor_link", link: "api", root: "api/proto", targetRoot: "proto"},
		{name: "dangling_link", link: "proto", root: "proto", targetRoot: "missing"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkout, outside := t.TempDir(), t.TempDir()
			if tt.name != "dangling_link" {
				metadata := filepath.Join(outside, tt.targetRoot, bufModuleConfigFile)
				require.NoError(t, os.MkdirAll(filepath.Dir(metadata), 0o755))
				require.NoError(t, os.WriteFile(metadata, []byte("version: v1\ndeps: [buf.build/acme/outside]\n"), 0o644))
			}
			target := outside
			if tt.name == "dangling_link" {
				target = filepath.Join(outside, tt.targetRoot)
			}
			require.NoError(t, os.Symlink(target, filepath.Join(checkout, tt.link)))
			workspace := filepath.Join(checkout, bufWorkConfigFile)
			require.NoError(t, os.WriteFile(workspace, []byte("version: v1\ndirectories: ["+tt.root+"]\n"), 0o644))

			_, err := readBufDependencyWorkspace(workspace)

			require.ErrorContains(t, err, "non-regular dependency directory")
			require.NotContains(t, err.Error(), "buf.build/acme/outside")
		})
	}
}
