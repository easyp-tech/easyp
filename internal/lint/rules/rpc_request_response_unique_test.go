package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestRPCRequestResponseUnique_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "RPC_REQUEST_RESPONSE_UNIQUE"

	rule := rules.RPCRequestResponseUnique{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestRPCRequestResponseUnique_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "request and response types must be unique across all RPCs"

	rule := rules.RPCRequestResponseUnique{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestRPCRequestResponseUnique_Validate(t *testing.T) {
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
					Offset:   375,
					Line:     14,
					Column:   3,
				},
				SourceName: "TokenData",
				Message:    "request and response types must be unique across all RPCs",
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

			rule := rules.RPCRequestResponseUnique{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
