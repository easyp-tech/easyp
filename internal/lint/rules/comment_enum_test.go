package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestCommentEnum_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "COMMENT_ENUM"

	rule := rules.CommentEnum{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestCommentEnum_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "enum comments must not be empty"

	rule := rules.CommentEnum{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestCommentEnum_Validate(t *testing.T) {
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
					Offset: 790,
					Line:   44,
					Column: 1,
				},
				SourceName: "social_network",
				Message:    "enum comments must not be empty",
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

			rule := rules.CommentEnum{}

			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
