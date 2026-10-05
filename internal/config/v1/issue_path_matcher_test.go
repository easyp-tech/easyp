package v1

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssuePathMatcher(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, pattern, file string
		want                bool
	}{
		{name: "directory", pattern: "generated", file: "generated/nested/a.proto", want: true},
		{name: "directory_boundary", pattern: "generated", file: "generated_other/a.proto"},
		{name: "literal_file", pattern: "generated/a.proto", file: "generated/a.proto", want: true},
		{name: "file_boundary", pattern: "generated/a.proto", file: "generated/a.proto/b.proto"},
		{name: "star", pattern: "generated/*.proto", file: "generated/a.proto", want: true},
		{name: "star_no_separator", pattern: "generated/*.proto", file: "generated/nested/a.proto"},
		{name: "question", pattern: "a?.proto", file: "ab.proto", want: true},
		{name: "question_one_character", pattern: "a?.proto", file: "abc.proto"},
		{name: "class", pattern: "[a-c].proto", file: "b.proto", want: true},
		{name: "negated_class", pattern: "[^a-c].proto", file: "d.proto", want: true},
		{name: "globstar_zero", pattern: "api/**/a.proto", file: "api/a.proto", want: true},
		{name: "globstar_many", pattern: "api/**/a.proto", file: "api/x/y/a.proto", want: true},
		{name: "globstar_suffix", pattern: "api/**", file: "api/x/y/a.proto", want: true},
		{name: "globstar_prefix", pattern: "**/*.proto", file: "a.proto", want: true},
		{name: "multiple_globstars", pattern: "**/x/**/a.proto", file: "one/x/two/three/a.proto", want: true},
		{name: "anchored", pattern: "api/**", file: "other/api/a.proto"},
		{name: "normalized_directory", pattern: "./generated/", file: "generated/a.proto", want: true},
		{name: "outside_source", pattern: "**", file: "../a.proto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matcher, err := newIssuePathMatcher(tt.pattern)
			require.NoError(t, err)
			assert.Equal(t, tt.want, matcher.matches(tt.file))
		})
	}
}

func TestIssuePathValidation(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, pattern string }{
		{name: "absolute", pattern: "/tmp/*.proto"},
		{name: "windows_absolute", pattern: `C:\temp\*.proto`},
		{name: "windows_drive", pattern: "C:temp/*.proto"},
		{name: "backslash", pattern: `api\*.proto`},
		{name: "parent", pattern: "../*.proto"},
		{name: "internal_parent", pattern: "api/../*.proto"},
		{name: "glob_parent", pattern: "**/../*.proto"},
		{name: "class", pattern: "api/[abc.proto"},
		{name: "empty_class", pattern: "api/[].proto"},
		{name: "partial_globstar", pattern: "api/a**/*.proto"},
		{name: "triple_star", pattern: "api/***/*.proto"},
		{name: "empty_segment", pattern: "api//*.proto"},
		{name: "empty_root_segment", pattern: ".//"},
		{name: "nul", pattern: "api/\x00.proto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParsePolicy(strings.NewReader(fmt.Sprintf("version: v1\nissues:\n  exclude-rules:\n    - path: %q\n", tt.pattern)))
			require.ErrorContains(t, err, "issues.exclude-rules[0].path")
			_, err = (Policy{Issues: IssuePolicy{ExcludeRules: []IssueExcludeRule{{Path: tt.pattern}}}}).LintConfig()
			require.ErrorContains(t, err, "issues.exclude-rules[0].path")
		})
	}
}

func TestIssueExclusionsForPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		rules       []IssueExcludeRule
		wantLinters []string
		wantAll     bool
	}{
		{name: "omitted_linters", rules: []IssueExcludeRule{{Path: "api/**"}}, wantAll: true},
		{name: "empty_linters", rules: []IssueExcludeRule{{Path: "api/**", Linters: []string{}}}, wantAll: true},
		{name: "pathless_all", rules: []IssueExcludeRule{{}}, wantAll: true},
		{name: "named_rules_and_groups", rules: []IssueExcludeRule{{Path: "api/**", Linters: []string{"MINIMAL", "FILE_LOWER_SNAKE_CASE"}}}, wantLinters: []string{"MINIMAL", "FILE_LOWER_SNAKE_CASE"}},
		{name: "unmatched", rules: []IssueExcludeRule{{Path: "other/**"}}},
		{name: "pathless_named", rules: []IssueExcludeRule{{Linters: []string{"BASIC"}}}, wantLinters: []string{"BASIC"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			policy := Policy{Issues: IssuePolicy{ExcludeRules: tt.rules}}
			_, err := policy.LintConfig()
			require.NoError(t, err)
			linters, all, err := policy.Issues.ExclusionsForPath("api/a.proto")
			require.NoError(t, err)
			assert.Equal(t, tt.wantLinters, linters)
			assert.Equal(t, tt.wantAll, all)
		})
	}
}
