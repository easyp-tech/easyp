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
