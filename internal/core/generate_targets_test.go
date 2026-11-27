package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerationTargetsDoNotExposePlanSelection(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeGenerationSources(t, root, map[string]string{
		"api/item.proto": `syntax = "proto3"; package api; import "shared.proto"; message Item { shared.Value value = 1; }`,
		"shared.proto":   `syntax = "proto3"; package shared; message Value {}`,
	})
	app := testCoreWithPlugins(nil, &captureExecutor{})
	plan, err := app.PrepareGenerationFiles(t.Context(), root, []string{"api/item.proto"})
	require.NoError(t, err)
	names := plan.Targets()
	require.Equal(t, []string{"api/item.proto"}, names)
	names[0] = "shared.proto"
	assert.Equal(t, []string{"api/item.proto"}, plan.Targets())
	assert.Equal(t, []string{"api/item.proto"}, descriptorNames(plan.DescriptorSet(false).File))
	assert.Equal(t, []string{"shared.proto", "api/item.proto"}, descriptorNames(plan.DescriptorSet(true).File))
	mutated := plan.DescriptorSet(true)
	names[0] = "mutated.proto"
	mutated.File[0].Name = &names[0]
	assert.Equal(t, []string{"shared.proto", "api/item.proto"}, descriptorNames(plan.DescriptorSet(true).File))
}
