package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestPackageLowerSnakeCase_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "PACKAGE_LOWER_SNAKE_CASE"

	rule := rules.PackageLowerSnakeCase{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestPackageLowerSnakeCase_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "package name should be lower_snake_case"

	rule := rules.PackageLowerSnakeCase{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestPackageLowerSnakeCase_Validate(t *testing.T) {
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
					Offset:   20,
					Line:     3,
					Column:   1,
				},
				SourceName: "Session",
				Message:    "package name should be lower_snake_case",
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

			rule := rules.PackageLowerSnakeCase{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
