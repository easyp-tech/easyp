package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestPackageSameCsharpNamespace_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "different proto files in the same package should have the same csharp_namespace"

	rule := rules.PackageSameCsharpNamespaceLinked{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestPackageSameCsharpNamespace_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames  []string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			// invalid_options/queue.proto (V1) is the exempt first value;
			// invalid_options/session.proto (same package, V2) issues at its
			// own csharp_namespace option keyword with the constant as written.
			fileNames: []string{invalidAuthProto5, invalidAuthProto6},
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   142,
					Line:     9,
					Column:   1,
				},
				SourceName: `"ZergsLaw.BackTemplate.Api.Session.V2"`,
				Message:    "different proto files in the same package should have the same csharp_namespace",
				RuleName:   "PACKAGE_SAME_CSHARP_NAMESPACE",
			},
			wantErr: nil,
		},
		"valid": {
			fileNames: []string{validAuthProto, validAuthProto2},
			wantErr:   nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rule := &rules.PackageSameCsharpNamespaceLinked{}

			var infos []core.ProtoInfo
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
