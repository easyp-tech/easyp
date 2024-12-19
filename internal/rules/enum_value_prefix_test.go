package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestEnumValuePrefix_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "enum value prefix is not valid"

	rule := rules.EnumValuePrefix{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestEnumValuePrefix_Validate(t *testing.T) {
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
					Offset:   917,
					Line:     51,
					Column:   3,
				},
				SourceName: "none",
				Message:    "enum value prefix is not valid",
				RuleName:   "ENUM_VALUE_PREFIX",
			},
			wantErr: nil,
		},
		"invalid_nested": {
			fileName: invalidAuthProto,
			wantIssues: &core.Issue{
				Position: meta.Position{
					Filename: "",
					Offset:   667,
					Line:     33,
					Column:   5,
				},
				SourceName: "none",
				Message:    "enum value prefix is not valid",
				RuleName:   "ENUM_VALUE_PREFIX",
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

			rule := rules.EnumValuePrefix{}
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
