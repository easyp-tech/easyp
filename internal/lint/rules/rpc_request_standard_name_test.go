package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestRPCRequestStandardName_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "RPC_REQUEST_STANDARD_NAME"

	rule := rules.RPCRequestStandardName{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestRPCRequestStandardName_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "rpc request should have suffix 'Request'"

	rule := rules.RPCRequestStandardName{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestRPCRequestStandardName_Validate(t *testing.T) {
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
					Offset:   214,
					Line:     11,
					Column:   3,
				},
				SourceName: "SessionInfo",
				Message:    "rpc request should have suffix 'Request'",
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

			rule := rules.RPCRequestStandardName{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
