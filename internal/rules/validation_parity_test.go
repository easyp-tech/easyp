package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
)

// parityLintRule is a minimal core.LintRule double for the construction cell:
// the linked engine consumes pre-validated rule lists, so the double needs no
// name of its own.
type parityLintRule struct{}

var _ core.LintRule = parityLintRule{}

func (parityLintRule) Message() string { return "validation parity guard" }

func (parityLintRule) Validate(_ core.LinkedFileInfo) ([]core.Issue, error) {
	return nil, nil
}

// TestLintUseValidationParity pins LINT-ENG-06 (D-06): an unknown rule or
// group name in lint.use fails with main's exact validation error through the
// untouched rules.New chain (ValidateNames -> validateConfigRuleNames ->
// core.ErrInvalidRule, wrapped per selection as "validateConfigRuleNames:
// use:/except:/ignore_only: "), the PACKAGE_NO_IMPORT_CYCLE special case stays
// main's deliberate exception until Phase 6 removes it, and a valid config
// constructs the old engine output and the linked engine with no second name
// validation anywhere — rules.New is THE validator.
func TestLintUseValidationParity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfg        config.LintConfig
		wantPrefix string
		wantName   string
	}{
		{
			name:       "unknown_rule_in_use",
			cfg:        config.LintConfig{Use: []string{"NOT_A_RULE"}},
			wantPrefix: "validateConfigRuleNames: use: ",
			wantName:   "NOT_A_RULE",
		},
		{
			// Groups and rules fail identically: ValidateNames has no
			// group-specific branch, the allowed set simply does not contain
			// the name.
			name:       "unknown_group_in_use",
			cfg:        config.LintConfig{Use: []string{"NOT_A_GROUP"}},
			wantPrefix: "validateConfigRuleNames: use: ",
			wantName:   "NOT_A_GROUP",
		},
		{
			name:       "unknown_name_in_except",
			cfg:        config.LintConfig{Use: []string{"MINIMAL"}, Except: []string{"NOT_A_RULE"}},
			wantPrefix: "validateConfigRuleNames: except: ",
			wantName:   "NOT_A_RULE",
		},
		{
			name:       "unknown_name_in_ignore_only",
			cfg:        config.LintConfig{IgnoreOnly: map[string][]string{"NOT_A_RULE": {"generated"}}},
			wantPrefix: "validateConfigRuleNames: ignore_only: ",
			wantName:   "NOT_A_RULE",
		},
		{
			// Main's deliberate special case (ErrInvalidRule + "is not
			// implemented"); removed only in Phase 6 — the guard pins it so
			// nothing drifts before then.
			name:       "package_no_import_cycle_special_case",
			cfg:        config.LintConfig{Use: []string{"PACKAGE_NO_IMPORT_CYCLE"}},
			wantPrefix: "validateConfigRuleNames: use: ",
			wantName:   "PACKAGE_NO_IMPORT_CYCLE is not implemented",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := New(tt.cfg)

			require.ErrorIs(t, err, core.ErrInvalidRule)
			require.ErrorContains(t, err, tt.wantPrefix)
			require.ErrorContains(t, err, tt.wantName)
		})
	}

	t.Run("valid_config_still_constructs", func(t *testing.T) {
		t.Parallel()

		// The pre-validated config outputs rules.New hands to the old path,
		// pinned as-is.
		selected, ignoreOnly, err := New(config.LintConfig{Use: []string{"MINIMAL"}})
		require.NoError(t, err)
		assert.NotNil(t, selected)
		assert.NotEmpty(t, selected)
		assert.NotNil(t, ignoreOnly)

		// The linked engine consumes pre-validated rule lists: core.New takes
		// LintRule values directly and exposes no error surface at all — the
		// absence of a second name validator IS the parity contract (one name
		// surface: this registry feeds config validation, directive
		// validation via KnownLintRules, and future schemas).
		linked := core.New(core.Options{LinkedRules: []core.LintRule{parityLintRule{}}})
		assert.NotNil(t, linked)
	})

	t.Run("registry_covers_all_names", func(t *testing.T) {
		t.Parallel()

		// The registry is the single name source: every group key and every
		// rule name is a valid lint.use value for the config validation this
		// guard pins.
		allowed := AllLintUseValues()
		for _, group := range []string{"MINIMAL", "BASIC", "DEFAULT", "COMMENTS", "UNARY_RPC"} {
			assert.Contains(t, allowed, group)
		}
		for _, name := range AllRuleNames() {
			assert.Contains(t, allowed, name)
		}
	})
}
