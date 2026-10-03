package easypconfig

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestDescribeFileBreakingCategory(t *testing.T) {
	t.Parallel()
	got, err := Describe(DescribeInput{File: v1.PolicyFile, Path: "breaking.categories"})
	require.NoError(t, err)
	require.Len(t, got.Fields, 1)
	assert.Contains(t, got.Fields[0].Description, "FILE")
	assert.Contains(t, strings.Join(got.Notes, " "), "Omitted or empty categories preserve legacy checks")
	assert.Equal(t, "array", got.Schema["type"])
	items, ok := got.Schema["items"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []any{"FILE", "PACKAGE", "WIRE_JSON", "WIRE"}, items["enum"])
	require.NotEmpty(t, got.Examples)
	for _, example := range got.Examples {
		policy, err := v1.ParsePolicy(strings.NewReader(example.YAML))
		require.NoError(t, err)
		cfg, err := policy.BreakingConfig("")
		require.NoError(t, err)
		assert.Equal(t, []string{"FILE"}, cfg.Use)
	}
}

func TestDescribeRemoteExampleUsesCurrentRegistry(t *testing.T) {
	t.Parallel()
	got, err := Describe(DescribeInput{File: v1.GenerateFile, Path: "plugins[].remote"})
	require.NoError(t, err)
	require.NotEmpty(t, got.Examples)
	found := false
	for _, example := range got.Examples {
		assert.NotContains(t, example.YAML, "api.easyp.tech")
		gen, err := v1.ParseGenerate(strings.NewReader(example.YAML))
		require.NoError(t, err)
		for _, plugin := range gen.Plugins {
			if plugin.Remote == "" {
				continue
			}
			found = true
			assert.Equal(t, "plugins.beta.easyp.tech/protocolbuffers/go", plugin.Remote)
			assert.Equal(t, "v1.36.11", plugin.Version)
		}
	}
	require.True(t, found, "MCP must provide an executable remote-plugin example")
}
