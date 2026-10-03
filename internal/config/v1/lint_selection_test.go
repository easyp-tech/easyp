package v1

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestLintSelectionValidationCollectsAllLocations(t *testing.T) {
	t.Parallel()
	const raw = "version: v1\nlinters:\n  enable: [NO_SUCH_RULE, PACKAGE_NO_IMPORT_CYCLE]\n  disable: [NOT_A_RULE]\nissues:\n  exclude-rules:\n    - linters: [UNKNOWN_EXCLUSION]\n"
	issues := ValidatePolicyYAML([]byte(raw))
	require.Len(t, issues, 4)
	for _, issue := range issues {
		assert.Equal(t, config.SeverityError, issue.Severity)
		assert.Positive(t, issue.Line)
		assert.Positive(t, issue.Column)
	}
	path := filepath.Join(t.TempDir(), PolicyFile)
	require.NoError(t, os.WriteFile(path, []byte(raw), 0o644))
	fileIssues, err := ValidateFile(path)
	require.NoError(t, err)
	assert.Equal(t, issues, fileIssues)
	_, err = ParsePolicy(strings.NewReader(raw))
	for _, name := range []string{"NO_SUCH_RULE", "PACKAGE_NO_IMPORT_CYCLE", "NOT_A_RULE", "UNKNOWN_EXCLUSION"} {
		require.ErrorContains(t, err, name)
	}
}

func TestLintSelectionValidationMatchesRuntimeCatalog(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, key string }{{name: "enabled", key: "enable"}, {name: "disabled", key: "disable"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, name := range rules.AllLintUseValues() {
				raw := "version: v1\nlinters:\n  " + tt.key + ": [" + name + "]\n"
				assert.Empty(t, ValidatePolicyYAML([]byte(raw)), name)
				policy, err := ParsePolicy(strings.NewReader(raw))
				require.NoError(t, err, name)
				cfg, err := policy.LintConfig()
				require.NoError(t, err, name)
				_, _, err = rules.New(cfg)
				require.NoError(t, err, name)
			}
		})
	}
	data, err := SchemaJSON("easyp")
	require.NoError(t, err)
	assert.NotContains(t, string(data), "PACKAGE_NO_IMPORT_CYCLE")
}

func TestProgrammaticLintPolicyRejectsInvalidNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		policy Policy
	}{
		{name: "enable", policy: Policy{Linters: LinterPolicy{Enable: []string{"NO_SUCH_RULE"}}}},
		{name: "disable", policy: Policy{Linters: LinterPolicy{Disable: []string{"NO_SUCH_RULE"}}}},
		{name: "exclusion", policy: Policy{Issues: IssuePolicy{ExcludeRules: []IssueExcludeRule{{Linters: []string{"NO_SUCH_RULE"}}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tt.policy.LintConfig()
			require.ErrorContains(t, err, "NO_SUCH_RULE")
		})
	}
}
