package rules_test

// Registry-parity guard (04-07, threat T-04-18/T-04-19 mitigation): with the
// linked registry complete, NewLinked must be provably name-parity with the
// old New — the canonical rule names (no _LINKED suffix anywhere), the group
// expansion order, the config plumbing and the ignore_only unwrapping all
// agree, so no registry name, group or config surface can drift into
// schemas/config during the dormant phase. Validation itself stays single:
// both constructors share validateConfigRuleNames/unwrapLintGroups/
// unwrapIgnoreOnly (D-06 — rules.New remains THE validator), and the guard
// compares validation outcomes through identical Use lists rather than
// re-deriving them.

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

// oldNamesByName returns the canonical names of an old registry slice, in use
// order.
func oldNamesByName(oldRules []core.Rule) []string {
	names := make([]string, len(oldRules))
	for i, rule := range oldRules {
		names[i] = core.GetRuleName(rule)
	}
	return names
}

// newNamesByName returns the canonical names of a linked registry slice, in
// use order.
func newNamesByName(newRules []core.LintRule) []string {
	names := make([]string, len(newRules))
	for i, rule := range newRules {
		names[i] = core.LinkedRuleName(rule)
	}
	return names
}

// TestRegistryParity proves the completed NewLinked registry agrees with New
// on names, groups, config plumbing and ignore_only unwrapping.
func TestRegistryParity(t *testing.T) {
	t.Parallel()

	allNames := rules.AllRuleNames()
	full := config.LintConfig{Use: allNames}

	oldRules, _, err := rules.New(full)
	require.NoError(t, err)
	newRules, _, err := rules.NewLinked(full)
	require.NoError(t, err)

	// (a) name-set equality, pairwise per use order: the two constructors
	// return the same canonical names in the same order, the set equals
	// AllRuleNames exactly, and no _LINKED-suffixed name leaks anywhere.
	oldNames := oldNamesByName(oldRules)
	newNames := newNamesByName(newRules)
	require.Len(t, oldNames, len(allNames))
	require.Len(t, newNames, len(allNames))
	require.Equal(t, oldNames, newNames, "registry name parity per use order")
	require.ElementsMatch(t, allNames, oldNames, "old registry covers AllRuleNames")
	require.ElementsMatch(t, allNames, newNames, "linked registry covers AllRuleNames (no _LINKED suffix leaks)")

	// (b) group parity: every group key expands to the same name list, in the
	// same order, from both constructors.
	for _, group := range rules.AllGroups() {
		groupCfg := config.LintConfig{Use: []string{group.Key}}

		oldGroup, _, err := rules.New(groupCfg)
		require.NoError(t, err, group.Key)
		newGroup, _, err := rules.NewLinked(groupCfg)
		require.NoError(t, err, group.Key)

		require.Equal(t, group.Rules, oldNamesByName(oldGroup),
			"group %s expands to its declared names from New", group.Key)
		require.Equal(t, group.Rules, newNamesByName(newGroup),
			"group %s expands to its declared names from NewLinked", group.Key)
	}

	// (c) defaultIfEmpty config plumbing: the configured fields match between
	// the two constructions, defaults first...
	requireConfigParity(t, config.LintConfig{Use: allNames})

	// ...then with user-configured values plumbed through both constructors.
	requireConfigParity(t, config.LintConfig{
		Use:                    allNames,
		PackageDirectoryPrefix: "internal/",
		EnumZeroValueSuffix:    "NONE",
		ServiceSuffix:          "API",
	})

	// (d) ignoreOnly unwrapping: identical maps from both constructors for a
	// sample containing a group key, a bare rule name and a second group key.
	ignoreCfg := config.LintConfig{
		Use: allNames,
		IgnoreOnly: map[string][]string{
			minimalGroupKey:       {"a.proto", "b/"},
			"MESSAGE_PASCAL_CASE": {"c/d.proto"},
			unaryGroupKey:         {"e.proto"},
		},
	}

	_, oldIgnoreOnly, err := rules.New(ignoreCfg)
	require.NoError(t, err)
	_, newIgnoreOnly, err := rules.NewLinked(ignoreCfg)
	require.NoError(t, err)

	require.Equal(t, oldIgnoreOnly, newIgnoreOnly,
		"ignore_only unwrapping is identical from both constructors")

	// Spot-check the unwrapped shape: the group keys expanded to their rule
	// names, the bare name survived verbatim.
	for _, name := range rules.AllGroups()[0].Rules {
		require.Equal(t, []string{"a.proto", "b/"}, oldIgnoreOnly[name],
			"MINIMAL unwrap reaches %s", name)
	}
	require.Equal(t, []string{"e.proto"}, oldIgnoreOnly["RPC_NO_CLIENT_STREAMING"])
	require.Equal(t, []string{"c/d.proto"}, oldIgnoreOnly["MESSAGE_PASCAL_CASE"])
}

// requireConfigParity asserts the configured fields of the three
// defaultIfEmpty/config-plumbed rules match pairwise between New and
// NewLinked for one config.
func requireConfigParity(t *testing.T, cfg config.LintConfig) {
	t.Helper()

	oldRules, _, err := rules.New(cfg)
	require.NoError(t, err)
	newRules, _, err := rules.NewLinked(cfg)
	require.NoError(t, err)

	findOld := func(name string) core.Rule {
		for _, rule := range oldRules {
			if core.GetRuleName(rule) == name {
				return rule
			}
		}
		t.Fatalf("rule %s not found in the old registry", name)
		return nil
	}
	findNew := func(name string) core.LintRule {
		for _, rule := range newRules {
			if core.LinkedRuleName(rule) == name {
				return rule
			}
		}
		t.Fatalf("rule %s not found in the linked registry", name)
		return nil
	}

	oldMatch, ok := findOld("PACKAGE_DIRECTORY_MATCH").(*rules.PackageDirectoryMatch)
	require.True(t, ok)
	newMatch, ok := findNew("PACKAGE_DIRECTORY_MATCH").(*rules.PackageDirectoryMatchLinked)
	require.True(t, ok)
	require.Equal(t, oldMatch.Root, newMatch.Root, "PACKAGE_DIRECTORY_MATCH Root parity")
	require.Equal(t, oldMatch.Prefix, newMatch.Prefix, "PACKAGE_DIRECTORY_MATCH Prefix parity")

	oldZero, ok := findOld("ENUM_ZERO_VALUE_SUFFIX").(*rules.EnumZeroValueSuffix)
	require.True(t, ok)
	newZero, ok := findNew("ENUM_ZERO_VALUE_SUFFIX").(*rules.EnumZeroValueSuffixLinked)
	require.True(t, ok)
	require.Equal(t, oldZero.Suffix, newZero.Suffix, "ENUM_ZERO_VALUE_SUFFIX Suffix parity (defaultIfEmpty)")

	oldService, ok := findOld("SERVICE_SUFFIX").(*rules.ServiceSuffix)
	require.True(t, ok)
	newService, ok := findNew("SERVICE_SUFFIX").(*rules.ServiceSuffixLinked)
	require.True(t, ok)
	require.Equal(t, oldService.Suffix, newService.Suffix, "SERVICE_SUFFIX Suffix parity (defaultIfEmpty)")
}

// Group keys live in builder.go (unexported); the test package spells the two
// keys this file exercises.
const (
	minimalGroupKey = "MINIMAL"
	unaryGroupKey   = "UNARY_RPC"
)
