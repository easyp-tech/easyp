package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestFieldLowerSnakeCase_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "message field should be lower_snake_case"

	rule := rules.FieldLowerSnakeCase{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestMessageFieldLowerSnakeCase_Validate(t *testing.T) {
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
					Offset:   560,
					Line:     27,
					Column:   3,
				},
				SourceName: "Session_id",
				Message:    "message field should be lower_snake_case",
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

			rule := rules.FieldLowerSnakeCase{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
