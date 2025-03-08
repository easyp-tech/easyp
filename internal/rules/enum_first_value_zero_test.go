package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"go.redsock.ru/protopack/internal/core"
	"go.redsock.ru/protopack/internal/rules"
)

func TestEnumFirstValueZero_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "enum first value must be zero"

	rule := rules.EnumFirstValueZero{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestEnumFirstValueZero_Validate(t *testing.T) {
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
				SourceName: "4",
				Message:    "enum first value must be zero",
				RuleName:   "ENUM_FIRST_VALUE_ZERO",
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
				SourceName: "4",
				Message:    "enum first value must be zero",
				RuleName:   "ENUM_FIRST_VALUE_ZERO",
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

			rule := rules.EnumFirstValueZero{}
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
