package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPluginOptionsUnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		contents        string
		expected        PluginOptions
		expectedMessage string
	}{
		{
			name:     "mapping preserves repeated values",
			contents: "paths: source_relative\nfeatures: [one, two]\n",
			expected: PluginOptions{"paths": {"source_relative"}, "features": {"one", "two"}},
		},
		{
			name:     "flag list preserves repetitions and equals in values",
			contents: "[feature=one, feature=two, debug, expression=a=b]",
			expected: PluginOptions{"feature": {"one", "two"}, "debug": {""}, "expression": {"a=b"}},
		},
		{name: "empty mapping", contents: "{}", expected: PluginOptions{}},
		{name: "empty list", contents: "[]", expected: PluginOptions{}},
		{
			name:            "non-scalar item",
			contents:        "[ok=value, [nested]]",
			expectedMessage: "opts[1] must be a flag string",
		},
		{
			name:            "empty flag name",
			contents:        "[ok=value, '=bad']",
			expectedMessage: "opts[1] has an empty flag name",
		},
		{
			name:            "nested mapping value",
			contents:        "option: {nested: value}",
			expectedMessage: `opts["option"] must be scalar or sequence`,
		},
		{
			name:            "scalar options",
			contents:        "invalid",
			expectedMessage: "plugin opts must be a mapping or string list",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var options PluginOptions

			err := yaml.Unmarshal([]byte(tt.contents), &options)

			if tt.expectedMessage != "" {
				require.ErrorContains(t, err, tt.expectedMessage)
				assert.Nil(t, options, "invalid options must not leave a partially decoded map")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, options)
		})
	}
}
