package rules_test

import (
	"testing"

	"github.com/easyp-tech/easyp/internal/rules"
	"github.com/stretchr/testify/require"
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
