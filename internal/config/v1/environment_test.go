package v1

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
)

func TestParseGenerateEnvironmentVariables(t *testing.T) {
	// Environment changes affect the whole process, so these cases cannot run in parallel.
	tests := []struct {
		name       string
		config     string
		variable   string
		value      string
		wantOut    string
		wantOption string
		wantError  string
	}{
		{
			name:     "expands a set variable",
			config:   "plugins:\n  - name: go\n    out: ${EASYP_TEST_OUTPUT}\n",
			variable: "EASYP_TEST_OUTPUT",
			value:    "gen/go",
			wantOut:  "gen/go",
		},
		{
			name:     "uses a default for an empty variable",
			config:   "plugins:\n  - name: go\n    out: ${EASYP_TEST_OUTPUT:-gen/default}\n",
			variable: "EASYP_TEST_OUTPUT",
			value:    "",
			wantOut:  "gen/default",
		},
		{
			name:       "keeps escaped variables literal",
			config:     "plugins:\n  - name: go\n    out: gen\n    opts:\n      suffix: $${EASYP_TEST_OUTPUT}\n",
			variable:   "EASYP_TEST_OUTPUT",
			value:      "expanded",
			wantOut:    "gen",
			wantOption: "${EASYP_TEST_OUTPUT}",
		},
		{
			name:      "rejects malformed substitution",
			config:    "plugins:\n  - name: go\n    out: ${EASYP_TEST_OUTPUT\n",
			wantError: "envsubst",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.variable != "" {
				t.Setenv(tt.variable, tt.value)
			}

			got, err := ParseGenerate(strings.NewReader(tt.config))

			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			require.Len(t, got.Plugins, 1)
			assert.Equal(t, tt.wantOut, got.Plugins[0].Out)
			if tt.wantOption != "" {
				assert.Equal(t, []string{tt.wantOption}, got.Plugins[0].Opts["suffix"])
			}
		})
	}
}

func TestParsePolicyEnvironmentVariables(t *testing.T) {
	// Environment changes affect the whole process, so these cases cannot run in parallel.
	tests := []struct {
		name       string
		config     string
		variable   string
		value      string
		wantPreset string
	}{
		{
			name:       "expands the linter preset",
			config:     "linters:\n  default: ${EASYP_TEST_PRESET}\n",
			variable:   "EASYP_TEST_PRESET",
			value:      "MINIMAL",
			wantPreset: "MINIMAL",
		},
		{
			name:       "uses the default preset when empty",
			config:     "linters:\n  default: ${EASYP_TEST_PRESET:-BASIC}\n",
			variable:   "EASYP_TEST_PRESET",
			value:      "",
			wantPreset: "BASIC",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.variable, tt.value)

			got, err := ParsePolicy(strings.NewReader(tt.config))

			require.NoError(t, err)
			assert.Equal(t, tt.wantPreset, got.Linters.Default)
		})
	}
}

func TestValidateFileExpandsEnvironmentVariables(t *testing.T) {
	// Environment changes affect the whole process, so these cases cannot run in parallel.
	tests := []struct {
		name     string
		filename string
		config   string
		variable string
		value    string
		wantCode string
	}{
		{
			name:     "generator uses expanded plugin output",
			filename: GenerateFile,
			config:   "plugins:\n  - name: go\n    out: ${EASYP_TEST_OUTPUT}\n",
			variable: "EASYP_TEST_OUTPUT",
			value:    "gen/go",
		},
		{
			name:     "policy validates expanded preset",
			filename: PolicyFile,
			config:   "linters:\n  default: ${EASYP_TEST_PRESET}\n",
			variable: "EASYP_TEST_PRESET",
			value:    "MINIMAL",
		},
		{
			name:     "invalid expanded preset is reported",
			filename: PolicyFile,
			config:   "linters:\n  default: ${EASYP_TEST_PRESET}\n",
			variable: "EASYP_TEST_PRESET",
			value:    "INVALID",
			wantCode: "yaml_validation",
		},
		{
			name:     "malformed substitution is reported",
			filename: GenerateFile,
			config:   "plugins:\n  - name: go\n    out: ${EASYP_TEST_OUTPUT\n",
			variable: "EASYP_TEST_OUTPUT",
			value:    "gen/go",
			wantCode: "v1_validation",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.variable, tt.value)
			path := filepath.Join(t.TempDir(), tt.filename)
			require.NoError(t, os.WriteFile(path, []byte(tt.config), 0o644))

			issues, err := ValidateFile(path)

			require.NoError(t, err)
			if tt.wantCode == "" {
				assert.Empty(t, issues)
				return
			}
			require.True(t, config.HasErrors(issues))
			assert.Equal(t, tt.wantCode, issues[0].Code)
		})
	}
}

func TestParseGenerateEnvironmentComments(t *testing.T) {
	// Environment changes affect the whole process, so these cases cannot run in parallel.
	tests := []struct {
		name       string
		config     string
		wantOut    string
		wantOption string
		wantError  string
	}{
		{
			name: "full-line and trailing comments are ignored",
			config: `# ${BROKEN
plugins:
  - name: go # ${BROKEN
    out: ${EASYP_TEST_OUTPUT} # ${BROKEN
`,
			wantOut: "gen/go",
		},
		{
			name: "quoted hash remains part of a value",
			config: `plugins:
  - name: go
    out: gen
    opts:
      suffix: "prefix #${EASYP_TEST_OUTPUT}" # ${BROKEN
`,
			wantOut:    "gen",
			wantOption: "prefix #gen/go",
		},
		{
			name: "plain apostrophe does not hide a trailing comment",
			config: `plugins:
  - name: go
    out: gen
    opts:
      suffix: it's okay # ${BROKEN
`,
			wantOut:    "gen",
			wantOption: "it's okay",
		},
		{
			name: "multiline quoted value is not a block scalar",
			config: `plugins:
  - name: go
    out: gen
    opts:
      suffix: "start: |
        # ${EASYP_TEST_OUTPUT}"
      other: okay # ${BROKEN
`,
			wantOut:    "gen",
			wantOption: "start: | # gen/go",
		},
		{
			name: "plain hash without preceding space remains part of a value",
			config: `plugins:
  - name: go
    out: gen/#${EASYP_TEST_OUTPUT} # ${BROKEN
`,
			wantOut: "gen/#gen/go",
		},
		{
			name: "hash in a block scalar remains part of a value",
			config: `plugins:
  - name: go
    out: gen
    opts:
      suffix: | # ${BROKEN
        # ${EASYP_TEST_OUTPUT}
`,
			wantOut:    "gen",
			wantOption: "# gen/go\n",
		},
		{
			name: "malformed substitution in a block scalar still fails",
			config: `plugins:
  - name: go
    out: gen
    opts:
      suffix: |
        # ${BROKEN
`,
			wantError: "envsubst",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("EASYP_TEST_OUTPUT", "gen/go")

			got, err := ParseGenerate(strings.NewReader(tt.config))

			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			require.Len(t, got.Plugins, 1)
			assert.Equal(t, tt.wantOut, got.Plugins[0].Out)
			if tt.wantOption != "" {
				assert.Equal(t, []string{tt.wantOption}, got.Plugins[0].Opts["suffix"])
			}
		})
	}
}

func TestValidateFileIgnoresCommentSubstitutions(t *testing.T) {
	// Environment changes affect the whole process, so these cases cannot run in parallel.
	tests := []struct {
		name     string
		filename string
		config   string
	}{
		{
			name:     "policy",
			filename: PolicyFile,
			config:   "# ${BROKEN\nlinters:\n  default: MINIMAL # ${BROKEN\n",
		},
		{
			name:     "generator",
			filename: GenerateFile,
			config:   "# ${BROKEN\nplugins:\n  - name: go\n    out: ${EASYP_TEST_OUTPUT} # ${BROKEN\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("EASYP_TEST_OUTPUT", "gen/go")
			path := filepath.Join(t.TempDir(), tt.filename)
			require.NoError(t, os.WriteFile(path, []byte(tt.config), 0o644))

			issues, err := ValidateFile(path)

			require.NoError(t, err)
			assert.Empty(t, issues)
		})
	}
}
