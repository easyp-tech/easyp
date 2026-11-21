package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestEnumValueUpperSnakeCase_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "enum value must be in UPPER_SNAKE_CASE"

	rule := rules.EnumValueUpperSnakeCase{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestEnumValueUpperSnakeCase_Validate(t *testing.T) {
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
					Offset:   917,
					Line:     51,
					Column:   3,
				},
				SourceName: "none",
				Message:    "enum value must be in UPPER_SNAKE_CASE",
				RuleName:   "ENUM_VALUE_UPPER_SNAKE_CASE",
			},
			wantErr: nil,
		},
		"invalid_nested": {
			fileName: invalidAuthProto,
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   667,
					Line:     33,
					Column:   5,
				},
				SourceName: "none",
				Message:    "enum value must be in UPPER_SNAKE_CASE",
				RuleName:   "ENUM_VALUE_UPPER_SNAKE_CASE",
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

			rule := rules.EnumValueUpperSnakeCase{}

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
