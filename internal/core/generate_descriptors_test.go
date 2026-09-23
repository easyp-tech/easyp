package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestGenerateDescriptors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		files          map[string]string
		withImports    []bool
		includeImports bool
		wantOrder      []string
		wantGenerated  [][]string
		wantSaved      []string
	}{
		{
			name: "shared_transitive_dependencies",
			files: map[string]string{
				"shared.proto":     `syntax = "proto3"; message Shared {}`,
				"middle.proto":     `syntax = "proto3"; import "shared.proto"; message Middle { Shared shared = 1; }`,
				"api/first.proto":  `syntax = "proto3"; import "shared.proto"; message First { Shared shared = 1; }`,
				"api/second.proto": `syntax = "proto3"; import "middle.proto"; message Second { Middle middle = 1; }`,
			},
			withImports: []bool{true, false},
			wantOrder:   []string{"shared.proto", "api/first.proto", "middle.proto", "api/second.proto"},
			wantGenerated: [][]string{
				{"api/first.proto", "api/second.proto", "shared.proto", "middle.proto"},
				{"api/first.proto", "api/second.proto"},
			},
			wantSaved: []string{"api/first.proto", "api/second.proto"},
		},
		{
			name: "descriptor_includes_imports",
			files: map[string]string{
				"shared.proto":   `syntax = "proto3"; message Shared {}`,
				"api/item.proto": `syntax = "proto3"; import "shared.proto"; message Item { Shared shared = 1; }`,
			},
			withImports:    []bool{false, true},
			includeImports: true,
			wantOrder:      []string{"shared.proto", "api/item.proto"},
			wantGenerated:  [][]string{{"api/item.proto"}, {"api/item.proto", "shared.proto"}},
			wantSaved:      []string{"shared.proto", "api/item.proto"},
		},
		{
			name:           "target_without_imports",
			files:          map[string]string{"api/item.proto": `syntax = "proto3"; message Item {}`},
			withImports:    []bool{true},
			includeImports: true,
			wantOrder:      []string{"api/item.proto"},
			wantGenerated:  [][]string{{"api/item.proto"}},
			wantSaved:      []string{"api/item.proto"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			writeGenerationSources(t, root, tt.files)
			plugins := make([]Plugin, len(tt.withImports))
			for i, withImports := range tt.withImports {
				plugins[i] = Plugin{Source: PluginSource{Name: "custom-plugin"}, WithImports: withImports}
			}
			executor := &captureExecutor{}
			app := testCoreWithPlugins(plugins, executor)
			descriptorPath := filepath.Join(root, "descriptors.pb")

			err := app.Generate(t.Context(), root, descriptorPath, tt.includeImports)

			require.NoError(t, err)
			require.Len(t, executor.requests, len(tt.wantGenerated))
			for i, request := range executor.requests {
				assert.Equal(t, tt.wantOrder, descriptorNames(request.GetProtoFile()))
				assert.Equal(t, tt.wantGenerated[i], request.GetFileToGenerate())
			}
			raw, err := os.ReadFile(descriptorPath)
			require.NoError(t, err)
			var saved descriptorpb.FileDescriptorSet
			require.NoError(t, proto.Unmarshal(raw, &saved))
			assert.Equal(t, tt.wantSaved, descriptorNames(saved.GetFile()))
		})
	}
}

func TestGenerateStopsBeforePlugins(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		files          map[string]string
		canceled       bool
		descriptorPath string
		wantErr        error
		wantMessage    string
	}{
		{
			name:           "canceled_context",
			files:          map[string]string{"api/item.proto": `syntax = "proto3"; message Item {}`},
			canceled:       true,
			descriptorPath: "descriptors.pb",
			wantErr:        context.Canceled,
		},
		{name: "empty_sources", descriptorPath: "descriptors.pb", wantErr: ErrEmptyInputFiles},
		{
			name:           "invalid_source",
			files:          map[string]string{"api/item.proto": `syntax = "proto3"; message {}`},
			descriptorPath: "descriptors.pb",
			wantMessage:    "syntax error",
		},
		{
			name:           "descriptor_write_failure",
			files:          map[string]string{"api/item.proto": `syntax = "proto3"; message Item {}`},
			descriptorPath: "missing/descriptors.pb",
			wantErr:        os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			writeGenerationSources(t, root, tt.files)
			executor := &captureExecutor{}
			app := testCoreWithPlugins([]Plugin{{Source: PluginSource{Name: "custom-plugin"}}}, executor)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if tt.canceled {
				cancel()
			}
			descriptorPath := filepath.Join(root, tt.descriptorPath)

			err := app.Generate(ctx, root, descriptorPath, false)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.ErrorContains(t, err, tt.wantMessage)
			}
			assert.Empty(t, executor.requests)
			_, err = os.Stat(descriptorPath)
			assert.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func writeGenerationSources(t *testing.T, root string, files map[string]string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Join(root, "api"), 0o755))
	for name, content := range files {
		path := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
}

func descriptorNames(files []*descriptorpb.FileDescriptorProto) []string {
	var names []string
	for _, file := range files {
		names = append(names, file.GetName())
	}
	return names
}
