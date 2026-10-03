package v1

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
)

func TestSemanticValidationExpandsEnvironmentOnce(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		file string
		raw  string
		bad  bool
	}{
		{name: "escaped_policy_setting", file: PolicyFile, raw: "linters-settings:\n  SERVICE_SUFFIX:\n    suffix: '$${R17_SUFFIX:-Service}'\n"},
		{name: "escaped_incomplete_expression", file: PolicyFile, raw: "linters-settings:\n  SERVICE_SUFFIX:\n    suffix: '$${R17_SUFFIX'\n"},
		{name: "escaped_baseline_ref", file: PolicyFile, raw: "breaking:\n  baseline: 'git:$${R17_REF:-main}'\n"},
		{name: "escaped_baseline_cannot_become_valid", file: PolicyFile, raw: "breaking:\n  baseline: '$${R17_BASELINE:-git:main}'\n", bad: true},
		{name: "escaped_managed_value", file: GenerateFile, raw: "generate:\n  managed:\n    override: [{file_option: go_package_prefix, value: '$${R17_PREFIX:-gen}/{{file_dir}}'}]\n"},
		{name: "escaped_plugin_option", file: GenerateFile, raw: "plugins:\n  - name: go\n    out: '$${R17_OUT:-gen}'\n    opts: {suffix: '$${R17_SUFFIX'}\n"},
		{name: "escaped_option_cannot_become_valid", file: GenerateFile, raw: "generate:\n  managed:\n    override: [{file_option: '$${R17_OPTION:-go_package}', value: gen}]\n", bad: true},
		{name: "escaped_enum_cannot_become_valid", file: GenerateFile, raw: "generate:\n  managed:\n    override: [{file_option: optimize_for, value: '$${R17_ENUM:-SPEED}'}]\n", bad: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := []byte(tt.raw)
			before := bytes.Clone(raw)
			var issues []config.ValidationIssue
			if tt.file == PolicyFile {
				issues = ValidatePolicyYAML(raw)
				got, err := ParsePolicy(bytes.NewReader(raw))
				if tt.bad {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					var want Policy
					require.NoError(t, yaml.Unmarshal([]byte(strings.ReplaceAll(tt.raw, "$$", "$")), &want))
					assert.Equal(t, want.LinterSettings, got.LinterSettings)
					assert.Equal(t, want.Breaking, got.Breaking)
				}
			} else {
				issues = ValidateGenerateYAML(raw)
				got, err := ParseGenerate(bytes.NewReader(raw))
				if tt.bad {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					var want Generate
					require.NoError(t, yaml.Unmarshal([]byte(strings.ReplaceAll(tt.raw, "$$", "$")), &want))
					assert.Equal(t, want.Generate, got.Generate)
					assert.Equal(t, want.Plugins, got.Plugins)
				}
			}
			assert.Equal(t, tt.bad, config.HasErrors(issues), "%+v", issues)
			assert.Equal(t, before, raw, "validation must not mutate the source")
			path := filepath.Join(t.TempDir(), tt.file)
			require.NoError(t, os.WriteFile(path, raw, 0o600))
			fileIssues, err := ValidateFile(path)
			require.NoError(t, err)
			assert.Equal(t, tt.bad, config.HasErrors(fileIssues), "%+v", fileIssues)
		})
	}
}

func TestSemanticValidationPreservesRawPlaceholdersOnDecode(t *testing.T) {
	t.Parallel()
	const raw = "generate:\n  managed:\n    override: [{file_option: go_package_prefix, value: '${R17_PREFIX}/{{file_dir}}'}]\noptions:\n  go:\n    package_prefix: '${R17_PREFIX}'\n"
	var got Generate
	require.NoError(t, yaml.Unmarshal([]byte(raw), &got))
	assert.Equal(t, "${R17_PREFIX}/{{file_dir}}", got.Generate.Managed.Override[0].Value)
	require.NotNil(t, got.Options.Go.PackagePrefix)
	assert.Equal(t, "${R17_PREFIX}", *got.Options.Go.PackagePrefix)
}

func TestSemanticValidationSourceCoordinates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		raw    string
		line   int
		column int
		path   string
	}{
		{
			name: "unknown_option",
			raw:  "# managed policy\ngenerate:\n  managed:\n    override:\n      - file_option: bogus_option\n        value: gen\n",
			line: 5, column: 22, path: "generate.managed.override[0].file_option",
		},
		{
			name: "wrong_value_type",
			raw:  "# managed policy\ngenerate:\n  managed:\n    override:\n      - file_option: java_multiple_files\n        value: 'false'\n",
			line: 6, column: 16, path: "generate.managed.override[0].value",
		},
		{
			name: "invalid_enum",
			raw:  "# managed policy\ngenerate:\n  managed:\n    override:\n      - field_option: jstype\n        value: INVALID\n",
			line: 6, column: 16, path: "generate.managed.override[0].value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			issues := ValidateGenerateYAML([]byte(tt.raw))
			require.Len(t, issues, 1, "%+v", issues)
			assert.Equal(t, tt.line, issues[0].Line)
			assert.Equal(t, tt.column, issues[0].Column)
			assert.Contains(t, issues[0].Message, tt.path)
			_, err := ParseGenerate(strings.NewReader(tt.raw))
			require.ErrorContains(t, err, issues[0].Message)
			require.ErrorContains(t, err, fmt.Sprintf("%s:%d:%d:", GenerateFile, tt.line, tt.column))
		})
	}
}
