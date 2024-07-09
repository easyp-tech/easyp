package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint/rules"

	"github.com/easyp-tech/easyp/internal/lint"
)

func TestCommentOneof_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "COMMENT_ONEOF"

	rule := rules.CommentOneof{}
	name := rule.Name()

	assert.Equal(expName, name)
}

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
		wantIssues *lint.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: invalidAuthProto,
			wantIssues: &lint.Issue{
				Position: meta.Position{
					Filename: "",
					Offset:   674,
					Line:     34,
					Column:   3,
				},
				SourceName: "SocialNetwork",
				Message:    "oneof comments must not be empty",
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

			r, protos := start(t)

			rule := rules.CommentOneof{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
