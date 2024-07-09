package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestMessagePascalCase_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "MESSAGE_PASCAL_CASE"

	rule := rules.MessagePascalCase{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestMessagePascalCase_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "message name should be PascalCase"

	rule := rules.MessagePascalCase{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestMessagePascalCase_Validate(t *testing.T) {
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
					Offset:   536,
					Line:     26,
					Column:   1,
				},
				SourceName: "Delete_Info",
				Message:    "message name should be PascalCase",
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

			rule := rules.MessagePascalCase{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
