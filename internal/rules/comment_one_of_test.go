package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/rules"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

func TestCommentOneof_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "oneof comments must not be empty"

	rule := rules.CommentOneof{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestCommentOneOf_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: invalidAuthProto,
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   748,
					Line:     39,
					Column:   3,
				},
				SourceName: "SocialNetwork",
				Message:    "oneof comments must not be empty",
				RuleName:   "COMMENT_ONEOF",
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

			rule := rules.CommentOneof{}

			issues, err := rule.Validate(parseLinkedFile(t, tc.fileName))
			require.ErrorIs(t, err, tc.wantErr)
			switch {
			case tc.wantIssues != nil:
				require.Contains(t, issues, *tc.wantIssues)
			case len(issues) > 0:
				require.Empty(t, issues)
			}
		})
	}
}
