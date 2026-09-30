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

func TestPolicySemanticValidationParity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  string
		path string
	}{
		{name: "unprefixed_baseline", raw: "breaking:\n  baseline: main\n", path: "breaking.baseline"},
		{name: "empty_git_ref", raw: "breaking:\n  baseline: 'git:'\n", path: "breaking.baseline"},
		{name: "trailing_newline_ref", raw: "breaking:\n  baseline: \"git:main\\n\"\n", path: "breaking.baseline"},
		{name: "multiline_ref", raw: "breaking:\n  baseline: \"git:main\\nrelease\"\n", path: "breaking.baseline"},
		{name: "carriage_return_ref", raw: "breaking:\n  baseline: \"git:main\\r\"\n", path: "breaking.baseline"},
		{name: "unsupported_baseline", raw: "breaking:\n  baseline: image:baseline.bin\n", path: "breaking.baseline"},
		{name: "unsupported_breaking_category", raw: "breaking:\n  categories: [WIRE]\n", path: "breaking.categories"},
		{name: "linter_extends", raw: "linters:\n  extends: policy.yaml\n", path: "linters.extends"},
		{name: "breaking_extends", raw: "breaking:\n  extends: policy.yaml\n", path: "breaking.extends"},
		{name: "unsupported_setting_rule", raw: "linters-settings:\n  UNKNOWN:\n    suffix: Foo\n", path: `["linters-settings"].UNKNOWN`},
		{name: "unsupported_setting_field", raw: "linters-settings:\n  SERVICE_SUFFIX:\n    bogus: Foo\n", path: `["linters-settings"].SERVICE_SUFFIX.bogus`},
		{name: "unsupported_enum_setting_field", raw: "linters-settings:\n  ENUM_ZERO_VALUE_SUFFIX:\n    suffix: UNSPECIFIED\n    bogus: true\n", path: `["linters-settings"].ENUM_ZERO_VALUE_SUFFIX.bogus`},
		{name: "setting_wrong_type", raw: "linters-settings:\n  SERVICE_SUFFIX:\n    suffix: true\n", path: `["linters-settings"].SERVICE_SUFFIX.suffix`},
		{name: "numeric_baseline", raw: "breaking:\n  baseline: 123\n", path: "breaking.baseline"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			issues := ValidatePolicyYAML([]byte(tt.raw))
			require.True(t, config.HasErrors(issues), "%+v", issues)
			assert.Contains(t, issues[0].Message, tt.path)
			assert.Positive(t, issues[0].Line)
			assert.Positive(t, issues[0].Column)
			got, err := ParsePolicy(strings.NewReader(tt.raw))
			require.Error(t, err, "every policy section must be validated before any consumer runs")
			assert.Equal(t, Policy{}, got)
			path := filepath.Join(t.TempDir(), PolicyFile)
			require.NoError(t, os.WriteFile(path, []byte(tt.raw), 0o600))
			fileIssues, err := ValidateFile(path)
			require.NoError(t, err)
			assert.Equal(t, issues, fileIssues)
		})
	}
}

func TestPolicySemanticValidationAcceptsEmptyAndGitBaselines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "omitted", raw: "version: v1\n", want: "fallback"},
		{name: "empty", raw: "breaking:\n  baseline: ''\n", want: "fallback"},
		{name: "git_ref", raw: "breaking:\n  baseline: git:release/v1\n", want: "release/v1"},
		{name: "empty_reserved_and_settings", raw: "linters:\n  extends: ''\nbreaking:\n  extends: ''\nlinters-settings:\n  SERVICE_SUFFIX: {}\n  ENUM_ZERO_VALUE_SUFFIX:\n    suffix: ''\n", want: "fallback"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Empty(t, ValidatePolicyYAML([]byte(tt.raw)))
			policy, err := ParsePolicy(strings.NewReader(tt.raw))
			require.NoError(t, err)
			_, err = policy.LintConfig()
			require.NoError(t, err)
			cfg, err := policy.BreakingConfig("fallback")
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.AgainstGitRef)
		})
	}
}

func TestPolicyConsumersValidateUnrelatedSections(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		policy Policy
		want   string
	}{
		{name: "baseline", policy: Policy{Breaking: BreakingPolicy{Baseline: "main"}}, want: "breaking.baseline"},
		{name: "breaking_extends", policy: Policy{Breaking: BreakingPolicy{Extends: "parent.yaml"}}, want: "breaking.extends policy loading is not implemented"},
		{name: "linter_extends", policy: Policy{Linters: LinterPolicy{Extends: "parent.yaml"}}, want: "linters.extends policy loading is not implemented"},
		{name: "category", policy: Policy{Breaking: BreakingPolicy{Categories: []string{"WIRE"}}}, want: "breaking.categories: unsupported category"},
		{name: "setting_rule", policy: Policy{LinterSettings: map[string]map[string]string{"UNKNOWN": {"suffix": "Foo"}}}, want: "unsupported linters-settings rule"},
		{name: "setting_field", policy: Policy{LinterSettings: map[string]map[string]string{"SERVICE_SUFFIX": {"bogus": "Foo"}}}, want: "bogus"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tt.policy.LintConfig()
			require.ErrorContains(t, err, tt.want)
			_, err = tt.policy.BreakingConfig("fallback")
			require.ErrorContains(t, err, tt.want)
		})
	}
}
