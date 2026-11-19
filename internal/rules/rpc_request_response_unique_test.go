package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestRPCRequestResponseUnique_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "request and response types must be unique across all RPCs"

	rule := rules.RPCRequestResponseUniqueLinked{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestRPCRequestResponseUnique_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames  []string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			// invalidAuthProto re-uses the TokenData type: the second use issues
			// at that rpc's keyword position with the written type as the source
			// (the first user is exempt — main is per-file first-exempt on
			// written names; the rule runs through the Collect/Finish driver).
			fileNames: []string{invalidAuthProto},
			wantIssues: &core.Issue{
				Position: protoast.Position{
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
			fileNames: []string{validAuthProto},
			wantErr:   nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rule := &rules.RPCRequestResponseUniqueLinked{}

			var infos []core.LinkedFileInfo
			for _, fileName := range tc.fileNames {
				infos = append(infos, parseLinkedFile(t, fileName))
			}

			issues, err := validate(rule, infos...)
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
