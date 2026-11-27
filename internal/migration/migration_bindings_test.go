package migration

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProveMigrationImportBindingsCompilesVerifiedClosure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sources   map[string]string
		wantFiles []string
		wantError string
	}{
		{
			name: "unselected_sources_do_not_join_target_closure",
			sources: map[string]string{
				"target.proto":    `syntax = "proto3"; import "dep.proto"; message Target { Dep value = 1; }`,
				"dep.proto":       `syntax = "proto3"; message Dep {}`,
				"unrelated.proto": `syntax = "proto3"; message Dep { Missing value = 1; }`,
			},
			wantFiles: []string{"dep.proto", "target.proto"},
		},
		{
			name: "verified_standard_import_name_keeps_its_own_source",
			sources: map[string]string{
				"target.proto":                    `syntax = "proto3"; import "google/protobuf/timestamp.proto"; message Target { verified.Marker value = 1; }`,
				"google/protobuf/timestamp.proto": `syntax = "proto3"; package verified; message Marker {}`,
			},
			wantFiles: []string{"google/protobuf/timestamp.proto", "target.proto"},
		},
		{
			name: "invalid_verified_contract_is_rejected",
			sources: map[string]string{
				"target.proto": `syntax = "proto3"; message Target { Missing value = 1; }`,
			},
			wantError: "Compile: generation module example.test/migration:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			files := make(map[string]migrationProtoSource, len(tt.sources))
			for name, source := range tt.sources {
				files[name] = migrationProtoSource{
					module: "example.test/migration", path: name,
					identity: "example.test/migration:" + name, content: []byte(source),
				}
			}
			namespace := migrationImportNamespace{primary: []map[string]migrationProtoSource{files}}

			bindings, err := proveMigrationImportBindings(t.Context(), "example.test/migration", []string{"target.proto"}, namespace, namespace, true)

			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				assert.Nil(t, bindings)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantFiles, slices.Sorted(maps.Keys(bindings)))
			for name, binding := range bindings {
				assert.Equal(t, files[name], binding)
			}
		})
	}
}
