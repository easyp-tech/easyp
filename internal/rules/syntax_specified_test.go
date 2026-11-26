package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestSyntaxSpecified_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "syntax should be specified"

	rule := rules.SyntaxSpecified{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestSyntaxSpecified_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *core.Issue
		wantErr    error
	}{
		// protoast.Parse accepts a file without a syntax statement (its
		// Syntax defaults to "proto2") but SyntaxSpecified stays false, so
		// the rule fires at protoast.FilePosition: line 1, column 1 with the
		// path as both Position.Filename and SourceName (lab semantics —
		// Filename IS set, unlike the ported rules' zero Position literals).
		"invalid": {
			fileName: noSyntaxProto,
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "./../../testdata/no_syntax/no_syntax.proto",
					Offset:   0,
					Line:     1,
					Column:   1,
				},
				SourceName: "./../../testdata/no_syntax/no_syntax.proto",
				Message:    "syntax should be specified",
				RuleName:   "SYNTAX_SPECIFIED",
			},
			wantErr: nil,
		},
		"valid": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rule := rules.SyntaxSpecified{}

			issues, err := rule.Validate(parseLinkedFile(t, tc.fileName))
			require.ErrorIs(t, err, tc.wantErr)
			switch {
			case tc.wantIssues != nil:
				require.Len(t, issues, 1)
				require.Contains(t, issues, *tc.wantIssues)
			case len(issues) > 0:
				require.Empty(t, issues)
			}
		})
	}
}
