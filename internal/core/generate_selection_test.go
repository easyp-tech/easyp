package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestPrepareGenerationFilesKeepsClosureAndPluginSelections(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeGenerationSources(t, root, map[string]string{
		"shared.proto":     "syntax=\"proto3\"; package shared; message Shared {}",
		"api/first.proto":  "syntax=\"proto3\"; package api; import \"shared.proto\"; message First { shared.Shared item=1; }",
		"api/second.proto": "syntax=\"proto3\"; package api; message Second {}",
		"unrelated.proto":  "package unrelated; message {",
	})
	executor := &captureExecutor{}
	app := New(Options{Logger: logger.NewNop(), Inputs: Inputs{InputFilesDir: []InputFilesDir{{Root: ".", Path: "."}}}, Plugins: []Plugin{
		{Source: PluginSource{Name: "custom-plugin"}, WithImports: true},
		{Source: PluginSource{Name: "custom-plugin"}},
	}})
	app.localExecutor = executor
	plan, err := app.PrepareGenerationFiles(t.Context(), root, []string{"api/first.proto", "api/second.proto"})
	require.NoError(t, err)
	assert.Equal(t, []string{"shared.proto", "api/first.proto", "api/second.proto"}, descriptorNames(plan.DescriptorSet(true).File))
	assert.Equal(t, []string{"api/first.proto", "api/second.proto"}, descriptorNames(plan.DescriptorSet(false).File))
	require.NoError(t, plan.Execute(t.Context()))
	require.Len(t, executor.requests, 2)
	assert.ElementsMatch(t, []string{"api/first.proto", "api/second.proto", "shared.proto"}, executor.requests[0].FileToGenerate)
	assert.Equal(t, []string{"api/first.proto", "api/second.proto"}, executor.requests[1].FileToGenerate)
	_, err = app.PrepareGenerationFiles(t.Context(), root, []string{"absent.proto"})
	require.ErrorContains(t, err, "outside the module inputs")
	_, err = app.PrepareGenerationFiles(t.Context(), root, nil)
	require.ErrorIs(t, err, ErrEmptyInputFiles)
	_, err = app.PrepareGeneration(t.Context(), root)
	require.Error(t, err)
}
