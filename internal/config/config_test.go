package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func requireSingleOptValue(t *testing.T, opts PluginOpts, key, expected string) {
	t.Helper()
	require.Contains(t, opts, key)
	require.Equal(t, []string{expected}, opts[key])
}

func TestParseConfig_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		envVars        map[string]string
		expectedOutput string
		expectedModule string
		checkFunc      func(t *testing.T, cfg *Config)
	}{
		{
			name: "simple environment variable expansion",
			configContent: `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - directory: ${INPUT_DIR}
  plugins:
    - name: go
      out: ${OUTPUT_DIR}
      opts:
        module: ${MODULE_NAME}
`,
			envVars: map[string]string{
				"INPUT_DIR":   "eco_contract",
				"OUTPUT_DIR":  "./gen/go",
				"MODULE_NAME": "github.com/example/ec-code/gen/go",
			},
			expectedOutput: "./gen/go",
			expectedModule: "github.com/example/ec-code/gen/go",
			checkFunc: func(t *testing.T, cfg *Config) {
				require.Greater(t, len(cfg.Generate.Inputs), 0)
				require.Equal(t, "eco_contract", cfg.Generate.Inputs[0].InputFilesDir.Path)
			},
		},
		{
			name: "escape with double dollar sign",
			configContent: `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - directory: eco_contract
  plugins:
    - name: go
      out: ./gen/go
      opts:
        # Test escape: $$100 should become $100
        description: "This costs $$100 dollars"
        # Test escape: $${TEMP} should become ${TEMP}
        path: "${BASE_DIR}/$${TEMP}/file"
        # Test escape: $$ should become $
        literal: "$$"
`,
			envVars: map[string]string{
				"BASE_DIR": "/tmp",
			},
			checkFunc: func(t *testing.T, cfg *Config) {
				require.Greater(t, len(cfg.Generate.Plugins), 0)
				opts := cfg.Generate.Plugins[0].Opts
				requireSingleOptValue(t, opts, "description", "This costs $100 dollars")
				requireSingleOptValue(t, opts, "path", "/tmp/${TEMP}/file")
				requireSingleOptValue(t, opts, "literal", "$")
			},
		},
		{
			name: "mixed expansion and escape",
			configContent: `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - directory: ${INPUT_DIR}
  plugins:
    - name: go
      out: ${OUTPUT_DIR}
      opts:
        module: ${MODULE_NAME}
        # Mixed: expand ${OUTPUT_DIR} but escape $${TEMP}
        mixed: "${OUTPUT_DIR}/$${TEMP}/generated"
`,
			envVars: map[string]string{
				"INPUT_DIR":   "proto",
				"OUTPUT_DIR":  "./gen/go",
				"MODULE_NAME": "github.com/example/project",
			},
			checkFunc: func(t *testing.T, cfg *Config) {
				require.Greater(t, len(cfg.Generate.Plugins), 0)
				require.Greater(t, len(cfg.Generate.Inputs), 0)

				require.Equal(t, "proto", cfg.Generate.Inputs[0].InputFilesDir.Path)
				require.Equal(t, "./gen/go", cfg.Generate.Plugins[0].Out)
				requireSingleOptValue(t, cfg.Generate.Plugins[0].Opts, "module", "github.com/example/project")
				requireSingleOptValue(t, cfg.Generate.Plugins[0].Opts, "mixed", "./gen/go/${TEMP}/generated")
			},
		},
		{
			name: "unset variable becomes empty",
			configContent: `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - directory: ${UNSET_VAR}
  plugins:
    - name: go
      out: ./gen/go
`,
			envVars: map[string]string{},
			checkFunc: func(t *testing.T, cfg *Config) {
				require.Greater(t, len(cfg.Generate.Inputs), 0)
				require.Equal(t, "", cfg.Generate.Inputs[0].InputFilesDir.Path)
			},
		},
		{
			name: "default values with :- syntax",
			configContent: `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - directory: ${UNSET_VAR:-default_dir}
  plugins:
    - name: go
      out: ${OUTPUT_DIR:-./gen/go}
      opts:
        module: ${MODULE_NAME:-github.com/example/project}
        timeout: ${TIMEOUT:-30}
`,
			envVars: map[string]string{},
			checkFunc: func(t *testing.T, cfg *Config) {
				require.Greater(t, len(cfg.Generate.Inputs), 0)
				require.Equal(t, "default_dir", cfg.Generate.Inputs[0].InputFilesDir.Path)
				require.Greater(t, len(cfg.Generate.Plugins), 0)
				require.Equal(t, "./gen/go", cfg.Generate.Plugins[0].Out)
				require.Greater(t, len(cfg.Generate.Plugins[0].Opts), 0)
				requireSingleOptValue(t, cfg.Generate.Plugins[0].Opts, "module", "github.com/example/project")
				requireSingleOptValue(t, cfg.Generate.Plugins[0].Opts, "timeout", "30")
			},
		},
		{
			name: "default values with set variable",
			configContent: `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - directory: ${SET_VAR:-default_dir}
  plugins:
    - name: go
      out: ${OUTPUT_DIR:-./gen/go}
`,
			envVars: map[string]string{
				"SET_VAR":    "custom_dir",
				"OUTPUT_DIR": "./custom/output",
			},
			checkFunc: func(t *testing.T, cfg *Config) {
				require.Greater(t, len(cfg.Generate.Inputs), 0)
				require.Equal(t, "custom_dir", cfg.Generate.Inputs[0].InputFilesDir.Path)
				require.Greater(t, len(cfg.Generate.Plugins), 0)
				require.Equal(t, "./custom/output", cfg.Generate.Plugins[0].Out)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			cfg, err := ParseConfig([]byte(tt.configContent))
			require.NoError(t, err)

			if tt.expectedOutput != "" {
				require.Greater(t, len(cfg.Generate.Plugins), 0)
				require.Equal(t, tt.expectedOutput, cfg.Generate.Plugins[0].Out)
			}
			if tt.expectedModule != "" {
				require.Greater(t, len(cfg.Generate.Plugins), 0)
				require.Greater(t, len(cfg.Generate.Plugins[0].Opts), 0)
				requireSingleOptValue(t, cfg.Generate.Plugins[0].Opts, "module", tt.expectedModule)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, cfg)
			}
		})
	}
}
