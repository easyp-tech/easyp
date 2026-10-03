package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func presenceOf(t *testing.T, raw string) PolicyPresence {
	t.Helper()
	presence, err := ParsePolicyPresence([]byte(raw))
	require.NoError(t, err)
	return presence
}

func boolValue(t *testing.T, value *bool) bool {
	t.Helper()
	require.NotNil(t, value)
	return *value
}

func TestParsePolicyPresence(t *testing.T) {
	t.Parallel()

	presence := presenceOf(t, "version: v1\nlinters:\n  enable: [FIELD_LOWER_SNAKE_CASE]\n  allow_comment_ignores: false\nlinters-settings:\n  SERVICE_SUFFIX:\n    suffix: Api\nbreaking:\n  categories: []\n  ignore_unstable: false\n")
	assert.True(t, presence.Has(PolicySectionLinters))
	assert.True(t, presence.Has(PolicySectionLinterSettings))
	assert.False(t, presence.Has(PolicySectionIssues))
	assert.False(t, presence.Has(PolicySectionBreaking+" "))
	assert.True(t, presence.Field(PolicySectionLinters, "enable"))
	assert.False(t, presence.Field(PolicySectionLinters, "default"))
	assert.False(t, presence.Field(PolicySectionLinters, "disable"))
	assert.True(t, presence.Field(PolicySectionLinters, "allow_comment_ignores"))
	assert.True(t, presence.Field(PolicySectionLinterSettings, "SERVICE_SUFFIX"))
	assert.True(t, presence.Field(PolicySectionBreaking, "categories"))
	assert.True(t, presence.Field(PolicySectionBreaking, "ignore_unstable"))
	assert.False(t, presence.Field(PolicySectionBreaking, "baseline"))

	empty := presenceOf(t, "linters: {}\nlinters-settings: {}\n")
	assert.True(t, empty.Has(PolicySectionLinters))
	assert.Empty(t, empty.Fields[PolicySectionLinters])
	assert.True(t, empty.Has(PolicySectionLinterSettings))

	scalar, err := ParsePolicyPresence([]byte(""))
	require.NoError(t, err)
	assert.False(t, scalar.Has(PolicySectionLinters))
}

func TestParsePolicyPresenceRejectsNonMappingDocument(t *testing.T) {
	t.Parallel()

	presence, err := ParsePolicyPresence([]byte("- one\n- two\n"))
	require.NoError(t, err)
	assert.False(t, presence.Has(PolicySectionLinters))
}

func TestMergeLinterSectionWithoutBaseKeepsLocalSection(t *testing.T) {
	t.Parallel()

	local := LinterPolicy{Default: "MINIMAL", Enable: []string{"FIELD_LOWER_SNAKE_CASE"}}
	presence := presenceOf(t, "linters:\n  default: MINIMAL\n  enable: [FIELD_LOWER_SNAKE_CASE]\n")
	merged := MergeLinterSection(LinterPolicy{}, local, presence, false)
	assert.Equal(t, local, merged)
}

func TestMergeLinterSectionBaseDefaultAndLocalOverride(t *testing.T) {
	t.Parallel()

	base := BaseLinterSection(LinterPolicy{Default: "STANDARD"}, presenceOf(t, "linters:\n  default: STANDARD\n"))
	assert.Equal(t, "STANDARD", base.Default)
	assert.True(t, boolValue(t, base.AllowCommentIgnores))

	// An absent preset in the base is decided once, then overridden locally.
	base = BaseLinterSection(mustPolicy(t, "linters:\n  enable: [ENUM_VALUE_PREFIX]\n").Linters, presenceOf(t, "linters:\n  enable: [ENUM_VALUE_PREFIX]\n"))
	assert.Equal(t, DefaultLinterPreset, base.Default)

	local := presenceOf(t, "linters:\n  default: MINIMAL\n")
	merged := MergeLinterSection(base, LinterPolicy{Default: "MINIMAL"}, local, true)
	assert.Equal(t, "MINIMAL", merged.Default)
	assert.Equal(t, []string{"ENUM_VALUE_PREFIX"}, merged.Enable)
	assert.Empty(t, merged.Disable)
	assert.Empty(t, merged.Extends)
}

func TestMergeLinterSectionEmptyLocalDefaultIsKept(t *testing.T) {
	t.Parallel()

	base := BaseLinterSection(LinterPolicy{Default: "STANDARD"}, presenceOf(t, "linters:\n  default: STANDARD\n"))
	presence := presenceOf(t, "linters:\n  default: ''\n")
	merged := MergeLinterSection(base, LinterPolicy{Default: ""}, presence, true)
	assert.Equal(t, "", merged.Default)
}

func TestMergeLinterSectionSelectionPrecedence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		base            string
		local           string
		wantEnable      []string
		wantDisable     []string
		wantAllowIgnore bool
	}{
		{
			name:            "base disable retained",
			base:            "linters:\n  default: MINIMAL\n  enable: [FIELD_LOWER_SNAKE_CASE]\n  disable: [FIELD_LOWER_SNAKE_CASE]\n",
			local:           "linters:\n  extends: ./base.yaml\n",
			wantEnable:      nil,
			wantDisable:     []string{"FIELD_LOWER_SNAKE_CASE"},
			wantAllowIgnore: true,
		},
		{
			name:            "local enable re-enables base disable",
			base:            "linters:\n  default: MINIMAL\n  disable: [FIELD_LOWER_SNAKE_CASE]\n",
			local:           "linters:\n  extends: ./base.yaml\n  enable: [FIELD_LOWER_SNAKE_CASE]\n",
			wantEnable:      []string{"FIELD_LOWER_SNAKE_CASE"},
			wantDisable:     nil,
			wantAllowIgnore: true,
		},
		{
			name:            "local disable wins in the same layer",
			base:            "linters:\n  default: MINIMAL\n  enable: [FIELD_LOWER_SNAKE_CASE]\n",
			local:           "linters:\n  extends: ./base.yaml\n  enable: [FIELD_LOWER_SNAKE_CASE]\n  disable: [FIELD_LOWER_SNAKE_CASE]\n",
			wantEnable:      nil,
			wantDisable:     []string{"FIELD_LOWER_SNAKE_CASE"},
			wantAllowIgnore: true,
		},
		{
			name:            "explicit false keeps base true out",
			base:            "linters:\n  default: MINIMAL\n  allow_comment_ignores: true\n",
			local:           "linters:\n  extends: ./base.yaml\n  allow_comment_ignores: false\n",
			wantAllowIgnore: false,
		},
		{
			name:            "base false is inherited when local is silent",
			base:            "linters:\n  default: MINIMAL\n  allow_comment_ignores: false\n",
			local:           "linters:\n  extends: ./base.yaml\n",
			wantAllowIgnore: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := BaseLinterSection(mustPolicy(t, tt.base).Linters, presenceOf(t, tt.base))
			localPolicy := mustPolicy(t, tt.local)
			merged := MergeLinterSection(base, localPolicy.Linters, presenceOf(t, tt.local), true)
			assert.Equal(t, tt.wantEnable, merged.Enable)
			assert.Equal(t, tt.wantDisable, merged.Disable)
			assert.Equal(t, tt.wantAllowIgnore, boolValue(t, merged.AllowCommentIgnores))
		})
	}
}

func TestMergeLinterSectionBaseGroupDisableExpandsEveryRule(t *testing.T) {
	t.Parallel()

	base := BaseLinterSection(LinterPolicy{Disable: []string{"COMMENTS"}}, presenceOf(t, "linters:\n  disable: [COMMENTS]\n"))
	merged := MergeLinterSection(base, LinterPolicy{}, presenceOf(t, "linters:\n  extends: ./base.yaml\n"), true)
	assert.Empty(t, merged.Enable)
	assert.Len(t, merged.Disable, 7)
	assert.Contains(t, merged.Disable, "COMMENT_MESSAGE")

	base = BaseLinterSection(LinterPolicy{Disable: []string{"BASIC"}}, presenceOf(t, "linters:\n  disable: [BASIC]\n"))
	local := presenceOf(t, "linters:\n  extends: ./base.yaml\n  enable: [FIELD_LOWER_SNAKE_CASE]\n")
	merged = MergeLinterSection(base, LinterPolicy{Enable: []string{"FIELD_LOWER_SNAKE_CASE"}}, local, true)
	assert.Contains(t, merged.Enable, "FIELD_LOWER_SNAKE_CASE")
	assert.NotContains(t, merged.Disable, "FIELD_LOWER_SNAKE_CASE")
	assert.Len(t, merged.Disable, 19)
}

func TestMergeLinterSectionDedupsRepeatedSelections(t *testing.T) {
	t.Parallel()

	base := BaseLinterSection(LinterPolicy{Enable: []string{"FIELD_LOWER_SNAKE_CASE", "FIELD_LOWER_SNAKE_CASE"}}, presenceOf(t, "linters:\n  enable: [FIELD_LOWER_SNAKE_CASE, FIELD_LOWER_SNAKE_CASE]\n"))
	local := presenceOf(t, "linters:\n  extends: ./base.yaml\n  enable: [FIELD_LOWER_SNAKE_CASE, BASIC]\n")
	merged := MergeLinterSection(base, LinterPolicy{Enable: []string{"FIELD_LOWER_SNAKE_CASE", "BASIC"}}, local, true)
	assert.Len(t, merged.Enable, 20)
	assert.Contains(t, merged.Enable, "FIELD_LOWER_SNAKE_CASE")
}

func TestMergeLinterSettings(t *testing.T) {
	t.Parallel()

	base := map[string]map[string]string{
		"ENUM_ZERO_VALUE_SUFFIX": {"suffix": "NONE"},
		"SERVICE_SUFFIX":         {"suffix": "Base"},
	}

	t.Run("local wins per rule", func(t *testing.T) {
		t.Parallel()
		local := map[string]map[string]string{"SERVICE_SUFFIX": {"suffix": "Api"}}
		merged := MergeLinterSettings(base, local, presenceOf(t, "linters-settings:\n  SERVICE_SUFFIX:\n    suffix: Api\n"), true)
		assert.Equal(t, "NONE", merged["ENUM_ZERO_VALUE_SUFFIX"]["suffix"])
		assert.Equal(t, "Api", merged["SERVICE_SUFFIX"]["suffix"])
		merged["SERVICE_SUFFIX"]["suffix"] = "changed"
		assert.Equal(t, "Base", base["SERVICE_SUFFIX"]["suffix"], "base settings must not be mutated")
	})

	t.Run("explicit empty section clears", func(t *testing.T) {
		t.Parallel()
		merged := MergeLinterSettings(base, nil, presenceOf(t, "linters-settings: {}\n"), true)
		assert.Empty(t, merged)
	})

	t.Run("empty rule clears that rule only", func(t *testing.T) {
		t.Parallel()
		local := map[string]map[string]string{"SERVICE_SUFFIX": {}}
		merged := MergeLinterSettings(base, local, presenceOf(t, "linters-settings:\n  SERVICE_SUFFIX: {}\n"), true)
		assert.Empty(t, merged["SERVICE_SUFFIX"])
		assert.Equal(t, "NONE", merged["ENUM_ZERO_VALUE_SUFFIX"]["suffix"])
	})

	t.Run("absent local section inherits base", func(t *testing.T) {
		t.Parallel()
		merged := MergeLinterSettings(base, nil, presenceOf(t, "linters:\n  default: MINIMAL\n"), true)
		assert.Equal(t, base, merged)
	})

	t.Run("without base the local section is used", func(t *testing.T) {
		t.Parallel()
		local := map[string]map[string]string{"SERVICE_SUFFIX": {"suffix": "Api"}}
		assert.Equal(t, local, MergeLinterSettings(base, local, presenceOf(t, "linters-settings: {}\n"), false))
	})
}

func TestMergeBreakingSectionOnlyOverridesPresentValues(t *testing.T) {
	t.Parallel()

	baseSource := "breaking:\n  baseline: git:base\n  categories: [FILE]\n  ignore_unstable: true\n  ignore: [skip]\n"
	base := BaseBreakingSection(mustPolicy(t, baseSource).Breaking, presenceOf(t, baseSource))
	assert.Equal(t, BreakingPolicy{Baseline: "git:base", Categories: []string{"FILE"}, IgnoreUnstable: true, Ignore: []string{"skip"}}, base)

	local := "breaking:\n  extends: ./base.yaml\n"
	merged := MergeBreakingSection(base, BreakingPolicy{}, presenceOf(t, local), true)
	assert.Equal(t, base, merged)

	local = "breaking:\n  extends: ./base.yaml\n  categories: []\n  ignore_unstable: false\n  ignore: []\n"
	merged = MergeBreakingSection(base, mustPolicy(t, local).Breaking, presenceOf(t, local), true)
	assert.Equal(t, "git:base", merged.Baseline)
	assert.Empty(t, merged.Categories)
	assert.False(t, merged.IgnoreUnstable)
	assert.Empty(t, merged.Ignore)

	local = "breaking:\n  extends: ./base.yaml\n  baseline: ''\n"
	merged = MergeBreakingSection(base, mustPolicy(t, local).Breaking, presenceOf(t, local), true)
	assert.Empty(t, merged.Baseline)

	localPolicy := BreakingPolicy{Baseline: "git:local", Extends: "./base.yaml"}
	assert.Equal(t, localPolicy, MergeBreakingSection(base, localPolicy, presenceOf(t, "breaking:\n  extends: ./base.yaml\n"), false))
}

func mustPolicy(t *testing.T, raw string) Policy {
	t.Helper()
	policy, err := ParsePolicy(strings.NewReader(raw))
	require.NoError(t, err)
	return policy
}
