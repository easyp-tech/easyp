package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateOrdersSharedDependenciesBeforeTargets(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	files := map[string]string{
		"shared.proto":     `syntax = "proto3"; message Shared {}`,
		"middle.proto":     `syntax = "proto3"; import "shared.proto"; message Middle { Shared shared = 1; }`,
		"api/first.proto":  `syntax = "proto3"; import "shared.proto"; message First { Shared shared = 1; }`,
		"api/second.proto": `syntax = "proto3"; import "middle.proto"; message Second { Middle middle = 1; }`,
	}
	for name, content := range files {
		path := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}

	executor := &captureExecutor{}
	app := testCoreWithPlugins([]Plugin{
		{Source: PluginSource{Name: "with-imports"}, WithImports: true},
		{Source: PluginSource{Name: "targets-only"}},
	}, executor)
	require.NoError(t, app.Generate(t.Context(), root, "", false))
	require.Len(t, executor.requests, 2)

	for _, request := range executor.requests {
		var names []string
		for _, descriptor := range request.GetProtoFile() {
			names = append(names, descriptor.GetName())
		}
		require.Equal(t, []string{"shared.proto", "api/first.proto", "middle.proto", "api/second.proto"}, names)
	}
	require.Equal(t, []string{"api/first.proto", "api/second.proto", "shared.proto", "middle.proto"}, executor.requests[0].GetFileToGenerate())
	require.Equal(t, []string{"api/first.proto", "api/second.proto"}, executor.requests[1].GetFileToGenerate())
}
