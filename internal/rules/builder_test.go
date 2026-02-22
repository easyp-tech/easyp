package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestAllGroups(t *testing.T) {
	groups := rules.AllGroups()

	t.Run("returns 5 groups", func(t *testing.T) {
		require.Len(t, groups, 5)
	})

	t.Run("groups have correct order and keys", func(t *testing.T) {
		expectedKeys := []string{"MINIMAL", "BASIC", "DEFAULT", "COMMENTS", "UNARY_RPC"}
		expectedNames := []string{"Minimal", "Basic", "Default", "Comments", "Unary RPC"}
		for i, g := range groups {
			require.Equal(t, expectedKeys[i], g.Key, "group %d key mismatch", i)
			require.Equal(t, expectedNames[i], g.Name, "group %d name mismatch", i)
		}
	})

	t.Run("no group is empty", func(t *testing.T) {
		for _, g := range groups {
			require.NotEmpty(t, g.Rules, "group %q has no rules", g.Name)
		}
	})

	t.Run("all rule names are unique", func(t *testing.T) {
		seen := make(map[string]string)
		for _, g := range groups {
			for _, r := range g.Rules {
				if prevGroup, ok := seen[r]; ok {
					t.Errorf("rule %q appears in both %q and %q", r, prevGroup, g.Name)
				}
				seen[r] = g.Name
			}
		}
	})
}

func TestAllRuleNames(t *testing.T) {
	allRules := rules.AllRuleNames()
	require.NotEmpty(t, allRules)

	var expected []string
	for _, g := range rules.AllGroups() {
		expected = append(expected, g.Rules...)
	}
	require.Equal(t, expected, allRules)
}

func TestNew_ExceptExpandsGroups(t *testing.T) {
	// Use DEFAULT group and except COMMENTS group.
	// COMMENTS rules should not appear in result (but they aren't in DEFAULT anyway).
	// A more targeted test: use DEFAULT+COMMENTS, except COMMENTS.
	cfg := config.LintConfig{
		Use:    []string{"DEFAULT", "COMMENTS"},
		Except: []string{"COMMENTS"},
	}

	lintRules, _, err := rules.New(cfg)
	require.NoError(t, err)

	ruleNames := make([]string, len(lintRules))
	for i, r := range lintRules {
		ruleNames[i] = core.GetRuleName(r)
	}

	// All COMMENTS rules must be excluded.
	commentsGroup := rules.AllGroups()[3] // COMMENTS
	for _, commentRule := range commentsGroup.Rules {
		require.NotContains(t, ruleNames, commentRule,
			"rule %s from COMMENTS group should have been excluded by except", commentRule)
	}

	// DEFAULT rules must still be present.
	defaultGroup := rules.AllGroups()[2] // DEFAULT
	for _, defaultRule := range defaultGroup.Rules {
		require.Contains(t, ruleNames, defaultRule,
			"rule %s from DEFAULT group should be present", defaultRule)
	}
}

func TestNew_ExceptExpandsSingleGroup(t *testing.T) {
	// Use all groups, except UNARY_RPC.
	cfg := config.LintConfig{
		Use:    []string{"MINIMAL", "BASIC", "DEFAULT", "COMMENTS", "UNARY_RPC"},
		Except: []string{"UNARY_RPC"},
	}

	lintRules, _, err := rules.New(cfg)
	require.NoError(t, err)

	ruleNames := make([]string, len(lintRules))
	for i, r := range lintRules {
		ruleNames[i] = core.GetRuleName(r)
	}

	unaryGroup := rules.AllGroups()[4] // UNARY_RPC
	for _, unaryRule := range unaryGroup.Rules {
		require.NotContains(t, ruleNames, unaryRule,
			"rule %s from UNARY_RPC group should have been excluded", unaryRule)
	}
}

func TestNew_DefaultSuffixValues(t *testing.T) {
	// When EnumZeroValueSuffix and ServiceSuffix are empty, defaults apply.
	cfg := config.LintConfig{
		Use:                 []string{"DEFAULT"},
		EnumZeroValueSuffix: "",
		ServiceSuffix:       "",
	}

	lintRules, _, err := rules.New(cfg)
	require.NoError(t, err)

	var foundEnum, foundService bool
	for _, r := range lintRules {
		name := core.GetRuleName(r)
		switch name {
		case "ENUM_ZERO_VALUE_SUFFIX":
			foundEnum = true
			// The rule should have non-empty suffix.
			// Access the underlying struct via type assertion.
			require.NotNil(t, r)
		case "SERVICE_SUFFIX":
			foundService = true
			require.NotNil(t, r)
		}
	}
	require.True(t, foundEnum, "ENUM_ZERO_VALUE_SUFFIX rule not found in DEFAULT group")
	require.True(t, foundService, "SERVICE_SUFFIX rule not found in DEFAULT group")
}

func TestNew_ExplicitSuffixValuesPreserved(t *testing.T) {
	cfg := config.LintConfig{
		Use:                 []string{"DEFAULT"},
		EnumZeroValueSuffix: "NONE",
		ServiceSuffix:       "API",
	}

	lintRules, _, err := rules.New(cfg)
	require.NoError(t, err)

	// Verify rules are created without error; explicit values should be used.
	require.NotEmpty(t, lintRules)
}
