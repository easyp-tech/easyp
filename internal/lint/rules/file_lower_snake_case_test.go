package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestFileLowerSnakeCase_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "file name should be lower_snake_case.proto"

	rule := rules.FileLowerSnakeCase{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestFileLowerSnakeCase_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *lint.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: invalidAuthProto3,
			wantIssues: &lint.Issue{
				Position: meta.Position{
					Filename: "./../../../testdata/auth/InvalidName.proto",
					Offset:   0,
					Line:     0,
					Column:   0,
				},
				SourceName: "./../../../testdata/auth/InvalidName.proto",
				Message:    "file name should be lower_snake_case.proto",
				RuleName:   "FILE_LOWER_SNAKE_CASE",
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

			rule := rules.FileLowerSnakeCase{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
