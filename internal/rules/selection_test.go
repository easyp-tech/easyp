package rules_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestRepeatedLintSelectionIsIdempotent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		base, repeated config.LintConfig
	}{
		{name: "enable preset member", base: config.LintConfig{Use: []string{"DEFAULT"}}, repeated: config.LintConfig{Use: []string{"DEFAULT", "ENUM_VALUE_PREFIX"}}},
		{name: "duplicate rule", base: config.LintConfig{Use: []string{"ENUM_VALUE_PREFIX"}}, repeated: config.LintConfig{Use: []string{"ENUM_VALUE_PREFIX", "ENUM_VALUE_PREFIX"}}},
		{name: "duplicate group", base: config.LintConfig{Use: []string{"DEFAULT"}}, repeated: config.LintConfig{Use: []string{"DEFAULT", "DEFAULT"}}},
		{name: "duplicate disable", base: config.LintConfig{Use: []string{"DEFAULT"}, Except: []string{"ENUM_VALUE_PREFIX"}}, repeated: config.LintConfig{Use: []string{"DEFAULT"}, Except: []string{"ENUM_VALUE_PREFIX", "ENUM_VALUE_PREFIX"}}},
		{name: "disable group and member", base: config.LintConfig{Use: []string{"DEFAULT", "BASIC"}, Except: []string{"DEFAULT"}}, repeated: config.LintConfig{Use: []string{"DEFAULT", "BASIC"}, Except: []string{"DEFAULT", "ENUM_VALUE_PREFIX"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			names := func(cfg config.LintConfig) []string {
				selected, _, err := rules.New(cfg)
				require.NoError(t, err)
				result := make([]string, 0, len(selected))
				for _, rule := range selected {
					result = append(result, core.GetRuleName(rule))
				}
				return result
			}
			assert.Equal(t, names(tt.base), names(tt.repeated))
		})
	}
}

func TestInvalidLintNamesAreRejectedBeforeFiltering(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cfg  config.LintConfig
		want string
	}{
		{name: "unknown enabled", cfg: config.LintConfig{Use: []string{"NO_SUCH_RULE"}}, want: "NO_SUCH_RULE"},
		{name: "unknown disabled", cfg: config.LintConfig{Except: []string{"NO_SUCH_RULE"}}, want: "NO_SUCH_RULE"},
		{name: "unknown cannot cancel itself", cfg: config.LintConfig{Use: []string{"NO_SUCH_RULE"}, Except: []string{"NO_SUCH_RULE"}}, want: "NO_SUCH_RULE"},
		{name: "unknown exclusion", cfg: config.LintConfig{IgnoreOnly: map[string][]string{"NO_SUCH_RULE": {"generated"}}}, want: "NO_SUCH_RULE"},
		{name: "unimplemented enabled", cfg: config.LintConfig{Use: []string{"PACKAGE_NO_IMPORT_CYCLE"}}, want: "not implemented"},
		{name: "unimplemented disabled", cfg: config.LintConfig{Except: []string{"PACKAGE_NO_IMPORT_CYCLE"}}, want: "not implemented"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := rules.New(tt.cfg)
			require.ErrorIs(t, err, core.ErrInvalidRule)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}
