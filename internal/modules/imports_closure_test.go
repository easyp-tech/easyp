package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnresolvedImportsReachableClosure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		root       string
		dependency map[string]string
		missing    []string
		wantError  string
	}{
		{
			name: "missing_transitive_import", root: `syntax = "proto3"; import "dep.proto";`,
			dependency: map[string]string{"dep.proto": `syntax = "proto3"; import "missing.proto";`},
			missing:    []string{"missing.proto"},
		},
		{
			name: "fake_wellknown", root: `syntax = "proto3"; import "google/protobuf/not_real.proto";`,
			missing: []string{"google/protobuf/not_real.proto"},
		},
		{
			name: "cycles_and_shared_imports", root: `syntax = "proto3"; import "a.proto"; import "b.proto";`,
			dependency: map[string]string{
				"a.proto":      `syntax = "proto3"; import "b.proto"; import "shared.proto";`,
				"b.proto":      `syntax = "proto3"; import "a.proto"; import "main.proto"; import "shared.proto";`,
				"shared.proto": `syntax = "proto3"; import "missing.proto";`,
			},
			missing: []string{"missing.proto"},
		},
		{
			name: "unreferenced_broken_dependency_is_lazy", root: `syntax = "proto3"; import "dep.proto";`,
			dependency: map[string]string{
				"dep.proto":    `syntax = "proto3"; import "google/protobuf/api.proto";`,
				"broken.proto": `this is not protobuf`,
				"unused.proto": `syntax = "proto3"; import "absent.proto";`,
			},
			missing: []string{},
		},
		{
			name: "reachable_invalid_syntax", root: `syntax = "proto3"; import "dep.proto";`,
			dependency: map[string]string{"dep.proto": `syntax = "proto3"; message Broken {`},
			wantError:  "dep.proto",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, dependency := t.TempDir(), t.TempDir()
			writeV1GenerateFixture(t, root, "main.proto", tt.root)
			for path, source := range tt.dependency {
				writeV1GenerateFixture(t, dependency, path, source)
			}

			missing, err := unresolvedV1ImportsWithRoots(root, []string{"."}, []string{dependency})

			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.missing, missing)
		})
	}
}

func TestUnresolvedImportsRejectsInvalidPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
	}{
		{name: "empty", path: ""},
		{name: "directory", path: "."},
		{name: "parent", path: "../outside.proto"},
		{name: "embedded_parent", path: "nested/../dep.proto"},
		{name: "absolute", path: "/outside.proto"},
		{name: "backslash", path: `nested\\dep.proto`},
		{name: "windows_drive", path: "C:/outside.proto"},
		{name: "nul", path: `dep\x00.proto`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, dependency := t.TempDir(), t.TempDir()
			writeV1GenerateFixture(t, root, "main.proto", `syntax = "proto3"; import "dep.proto";`)
			writeV1GenerateFixture(t, dependency, "dep.proto", `syntax = "proto3"; import "`+tt.path+`";`)

			_, err := unresolvedV1ImportsWithRoots(root, []string{"."}, []string{dependency})

			require.ErrorContains(t, err, "invalid import")
			assert.ErrorContains(t, err, "dep.proto")
		})
	}
}
