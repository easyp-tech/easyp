package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestExclusionsPreserveLegacyLiteralAndPrefixSemantics(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, input string
		match, miss []string
	}{
		{name: "directory", input: "lint: {ignore: [generated]}", match: []string{"generated/a.proto"}, miss: []string{"generated_extra/a.proto", "generated.proto"}},
		{name: "literal_star", input: "lint: {ignore: ['*']}", match: []string{"*/a.proto"}, miss: []string{"a.proto", "other/a.proto"}},
		{name: "literal_bracket", input: "lint: {ignore: ['[foo]']}", match: []string{"[foo]/a.proto"}, miss: []string{"f/a.proto"}},
		{name: "prefix", input: "lint: {ignore_only: {SERVICE_SUFFIX: [generated]}}", match: []string{"generated/a.proto", "generated_extra/a.proto", "generated.proto"}, miss: []string{"other/a.proto"}},
		{name: "slash_prefix", input: "lint: {ignore_only: {SERVICE_SUFFIX: ['foo.proto/']}}", match: []string{"foo.proto/a.proto", "foo.proto/a/b.proto"}, miss: []string{"foo.proto", "foo.proto2/a.proto"}},
		{name: "prefix_literal_question", input: "lint: {ignore_only: {SERVICE_SUFFIX: ['foo?']}}", match: []string{"foo?.proto", "foo?bar/a.proto"}, miss: []string{"fooX.proto"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			legacy, err := parseLegacy([]byte(tt.input))
			require.NoError(t, err)
			raw, err := convertPolicy(legacy)
			require.NoError(t, err)
			var p v1.Policy
			require.NoError(t, yaml.Unmarshal(raw, &p))
			for _, file := range tt.match {
				rules, all, err := p.Issues.ExclusionsForPath(file)
				require.NoError(t, err)
				assert.True(t, all || len(rules) > 0, file)
			}
			for _, file := range tt.miss {
				rules, all, err := p.Issues.ExclusionsForPath(file)
				require.NoError(t, err)
				assert.False(t, all || len(rules) > 0, file)
			}
		})
	}
}

func TestPolicyBlocksImplicitInheritance(t *testing.T) {
	t.Parallel()
	raw, err := convertPolicy(legacyConfig{})
	require.NoError(t, err)
	var mapping map[string]any
	require.NoError(t, yaml.Unmarshal(raw, &mapping))
	for _, key := range []string{"linters", "linters-settings", "issues", "breaking"} {
		assert.Contains(t, mapping, key)
	}
}

func TestLegacyFieldsDoNotDriftWithRuntime(t *testing.T) {
	t.Parallel()
	_, err := parseLegacy([]byte("breaking: {ignore_unstable: true}\n"))
	require.ErrorContains(t, err, "unknown")
}
