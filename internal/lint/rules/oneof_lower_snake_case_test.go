package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestOneofLowerSnakeCase_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "ONEOF_LOWER_SNAKE_CASE"

	rule := rules.OneofLowerSnakeCase{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestOneofLowerSnakeCase_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "oneof name should be lower_snake_case"

	rule := rules.OneofLowerSnakeCase{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestOneofLowerSnakeCase_Validate(t *testing.T) {
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
				Message:    "oneof name should be lower_snake_case",
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

			rule := rules.OneofLowerSnakeCase{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
