package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateV1IncludesDeclaredSourceRootWithModuleConfig(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		config string
	}{
		{name: "buf module", config: "buf.yaml"},
		{name: "buf workspace", config: "buf.work.yaml"},
		{name: "easyp module", config: "protobuf.mod"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			sourceRoot := filepath.Join(root, "proto")
			writeTestProto(t, sourceRoot, "item.proto")
			err := os.WriteFile(filepath.Join(sourceRoot, tc.config), nil, 0o644)
			require.NoError(t, err)

			writeTestProto(t, sourceRoot, "nested/ignored.proto")
			err = os.WriteFile(filepath.Join(sourceRoot, "nested", "buf.yaml"), nil, 0o644)
			require.NoError(t, err)

			executor := &captureExecutor{}
			app := testCoreWithPlugins([]Plugin{{Source: PluginSource{Name: "custom-plugin"}}}, executor)
			app.inputs.InputFilesDir = []InputFilesDir{{Root: "proto", Path: "."}}

			err = app.Generate(t.Context(), root, "", false)
			require.NoError(t, err)
			require.Len(t, executor.requests, 1)
			require.Equal(t, []string{"item.proto"}, executor.requests[0].GetFileToGenerate())
		})
	}
}
