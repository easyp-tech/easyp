package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"go.redsock.ru/protopack/internal/core"
	"go.redsock.ru/protopack/internal/rules"
)

func TestServiceSuffix_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "service name should have suffix"

	rule := rules.ServiceSuffix{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestServiceSuffix_Validate(t *testing.T) {
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
					Offset:   197,
					Line:     10,
					Column:   1,
				},
				SourceName: "auth",
				Message:    "service name should have suffix",
				RuleName:   "SERVICE_SUFFIX",
			},
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

			rule := rules.ServiceSuffix{
				Suffix: "API",
			}
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
