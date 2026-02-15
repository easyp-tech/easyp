package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConfig_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		envVars        map[string]string
		expectedDeps   []string
		expectedOutput string
		expectedModule string
		checkFunc      func(t *testing.T, cfg *Config)
	}{
		{
			name: "simple environment variable expansion",
			configContent: `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
deps:
  - ${DEP_GOOGLEAPIS}
  - ${DEP_GNOSTIC}
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
				"DEP_GOOGLEAPIS": "github.com/googleapis/googleapis@common-protos-1_3_1",
				"DEP_GNOSTIC":    "github.com/google/gnostic@v0.7.0",
				"INPUT_DIR":      "eco_contract",
				"OUTPUT_DIR":     "./gen/go",
				"MODULE_NAME":    "github.com/example/ec-code/gen/go",
			},
			expectedDeps: []string{
				"github.com/googleapis/googleapis@common-protos-1_3_1",
				"github.com/google/gnostic@v0.7.0",
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
deps: []
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
				// Проверяем, что опции содержат экранированные значения
				require.Greater(t, len(cfg.Generate.Plugins), 0)
				opts := cfg.Generate.Plugins[0].Opts
				require.Equal(t, "This costs $100 dollars", opts["description"])
				require.Equal(t, "/tmp/${TEMP}/file", opts["path"])
				require.Equal(t, "$", opts["literal"])
			},
		},
		{
			name: "mixed expansion and escape",
			configContent: `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
deps:
  - ${DEP_URL}
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
				"DEP_URL":     "github.com/googleapis/googleapis@common-protos-1_3_1",
				"INPUT_DIR":   "proto",
				"OUTPUT_DIR":  "./gen/go",
				"MODULE_NAME": "github.com/example/project",
			},
			checkFunc: func(t *testing.T, cfg *Config) {
				require.Greater(t, len(cfg.Generate.Plugins), 0)
				require.Greater(t, len(cfg.Generate.Inputs), 0)
				require.Greater(t, len(cfg.Deps), 0)

				require.Equal(t, "github.com/googleapis/googleapis@common-protos-1_3_1", cfg.Deps[0])
				require.Equal(t, "proto", cfg.Generate.Inputs[0].InputFilesDir.Path)
				require.Equal(t, "./gen/go", cfg.Generate.Plugins[0].Out)
				require.Equal(t, "github.com/example/project", cfg.Generate.Plugins[0].Opts["module"])
				// Проверяем смешанное использование
				require.Equal(t, "./gen/go/${TEMP}/generated", cfg.Generate.Plugins[0].Opts["mixed"])
			},
		},
		{
			name: "unset variable becomes empty",
			configContent: `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
deps: []
generate:
  inputs:
    - directory: ${UNSET_VAR}
  plugins:
    - name: go
      out: ./gen/go
`,
			envVars: map[string]string{},
			checkFunc: func(t *testing.T, cfg *Config) {
				// Неустановленная переменная должна стать пустой строкой
				require.Greater(t, len(cfg.Generate.Inputs), 0)
				require.Equal(t, "", cfg.Generate.Inputs[0].InputFilesDir.Path)
			},
		},
		{
			name: "default values with :- syntax",
			configContent: `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
deps: []
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
				require.Equal(t, "github.com/example/project", cfg.Generate.Plugins[0].Opts["module"])
				require.Equal(t, "30", cfg.Generate.Plugins[0].Opts["timeout"])
			},
		},
		{
			name: "default values with set variable",
			configContent: `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
deps: []
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
				// Установленная переменная должна использоваться вместо default
				require.Equal(t, "custom_dir", cfg.Generate.Inputs[0].InputFilesDir.Path)
				require.Greater(t, len(cfg.Generate.Plugins), 0)
				require.Equal(t, "./custom/output", cfg.Generate.Plugins[0].Out)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем переменные окружения
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Парсим конфигурацию напрямую через ParseConfig
			cfg, err := ParseConfig([]byte(tt.configContent))
			require.NoError(t, err)

			// Проверяем ожидаемые значения
			if len(tt.expectedDeps) > 0 {
				require.Equal(t, tt.expectedDeps, cfg.Deps)
			}
			if tt.expectedOutput != "" {
				require.Greater(t, len(cfg.Generate.Plugins), 0)
				require.Equal(t, tt.expectedOutput, cfg.Generate.Plugins[0].Out)
			}
			if tt.expectedModule != "" {
				require.Greater(t, len(cfg.Generate.Plugins), 0)
				require.Greater(t, len(cfg.Generate.Plugins[0].Opts), 0)
				require.Equal(t, tt.expectedModule, cfg.Generate.Plugins[0].Opts["module"])
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, cfg)
			}
		})
	}
}
