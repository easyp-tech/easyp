package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestFileLowerSnakeCase_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "file name should be lower_snake_case.proto"

	rule := rules.FileLowerSnakeCaseLinked{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestFileLowerSnakeCase_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: invalidAuthProto3,
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "./../../testdata/auth/InvalidName.proto",
					Offset:   0,
					Line:     0,
					Column:   0,
				},
				SourceName: "./../../testdata/auth/InvalidName.proto",
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

			rule := rules.FileLowerSnakeCaseLinked{}

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
