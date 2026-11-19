package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestRPCResponseStandardName_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "rpc response should have suffix 'Response'"

	rule := rules.RPCResponseStandardNameLinked{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestRPCResponseStandardName_Validate(t *testing.T) {
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
					Offset:   214,
					Line:     11,
					Column:   3,
				},
				SourceName: "Result",
				Message:    "rpc response should have suffix 'Response'",
				RuleName:   "RPC_RESPONSE_STANDARD_NAME",
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

			rule := rules.RPCResponseStandardNameLinked{}

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
