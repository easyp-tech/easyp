package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"go.redsock.ru/protopack/internal/core"
	"go.redsock.ru/protopack/internal/rules"
)

func TestCommentField_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "field comments must not be empty"

	rule := rules.CommentField{}
	message := rule.Message()

	assert.Equal(expMessage, message)

}

func TestCommentField_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: invalidAuthProto,
			wantIssues: &core.Issue{
				Position: meta.Position{
					Filename: "",
					Offset:   447,
					Line:     18,
					Column:   3,
				},
				SourceName: "token",
				Message:    "field comments must not be empty",
				RuleName:   "COMMENT_FIELD",
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

			rule := rules.CommentField{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			switch {
			case tc.wantIssues != nil:
				r.Contains(issues, *tc.wantIssues)
			case len(issues) > 0:
				r.Empty(issues)
			}
		})
	}
}
