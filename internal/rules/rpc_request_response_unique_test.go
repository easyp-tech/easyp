package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

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
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: invalidAuthProto,
			wantIssues: &core.Issue{
				Position: meta.Position{
					Filename: "",
					Offset:   375,
					Line:     14,
					Column:   3,
				},
				SourceName: "TokenData",
				Message:    "request and response types must be unique across all RPCs",
				RuleName:   "RPC_REQUEST_RESPONSE_UNIQUE",
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
			switch {
			case tc.wantIssues != nil:
				r.Contains(issues, *tc.wantIssues)
			case len(issues) > 0:
				r.Empty(issues)
			}
		})
	}
}
