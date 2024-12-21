package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestCheckNoLint(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"no_lint_buf_comment": {
			fileName: noLintBufComment,
			wantErr:  nil,
		},
		"no_lint_easyp_comment": {
			fileName: noLintEasypComment,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.MessagePascalCase{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.NoError(err)
			r.Empty(issues)
		})
	}
}

func TestIsIgnored(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		comments []*parser.Comment
		ruleName string
		expected bool
	}{
		"no_comments": {
			comments: nil,
			ruleName: "DIRECTORY_SAME_PACKAGE",
			expected: false,
		},
		"multiple_comments_without_ignore": {
			comments: []*parser.Comment{
				{Raw: "// some message"},
				{Raw: "// another some message"},
			},
			ruleName: "ENUM_VALUE_PREFIX",
			expected: false,
		},
		"single_ignore_comment": {
			comments: []*parser.Comment{
				{Raw: "// buf:lint:ignore ENUM_VALUE_PREFIX"},
			},
			ruleName: "ENUM_VALUE_PREFIX",
			expected: true,
		},
		"multiple_ignore_comment": {
			comments: []*parser.Comment{
				{Raw: "// buf:lint:ignore ENUM_VALUE_PREFIX"},
				{Raw: "// buf:lint:ignore DIRECTORY_SAME_PACKAGE"},
			},
			ruleName: "ENUM_VALUE_PREFIX",
			expected: true,
		},
		"not_matched_ignore_rule_in_comments": {
			comments: []*parser.Comment{
				{Raw: "// buf:lint:ignore ENUM_VALUE_PREFIX"},
				{Raw: "// buf:lint:ignore DIRECTORY_SAME_PACKAGE"},
			},
			ruleName: "COMMENT_SERVICE",
			expected: false,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := core.CheckIsIgnored(tc.comments, tc.ruleName)

			require.Equal(t, tc.expected, res)
		})
	}
}
