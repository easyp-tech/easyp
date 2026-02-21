package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPluginOpts_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    PluginOpts
		errContains string
	}{
		{
			name: "scalar string",
			input: `opts:
  outputServices: grpc-js
`,
			expected: PluginOpts{
				"outputServices": {"grpc-js"},
			},
		},
		{
			name: "scalar bool and int preserve backward compatibility",
			input: `opts:
  useExactTypes: false
  timeout: 30
`,
			expected: PluginOpts{
				"useExactTypes": {"false"},
				"timeout":       {"30"},
			},
		},
		{
			name: "sequence of strings",
			input: `opts:
  outputServices:
    - grpc-js
    - generic-definitions
`,
			expected: PluginOpts{
				"outputServices": {"grpc-js", "generic-definitions"},
			},
		},
		{
			name: "sequence item must be scalar",
			input: `opts:
  outputServices:
    - grpc-js
    - key: value
`,
			errContains: `opts["outputServices"][1] must be scalar`,
		},
		{
			name: "nested map is invalid",
			input: `opts:
  outputServices:
    key: value
`,
			errContains: `opts["outputServices"] must be scalar or sequence`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg struct {
				Opts PluginOpts `yaml:"opts"`
			}

			err := yaml.Unmarshal([]byte(tt.input), &cfg)
			if tt.errContains != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errContains)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, cfg.Opts)
		})
	}
}
