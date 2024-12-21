package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestPackageSameCSharpNamespace_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "different proto files in the same package should have the same csharp_namespace"

	rule := rules.PackageSameCsharpNamespace{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestPackageSameCSharpNamespace_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames  []string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			fileNames: []string{invalidAuthProto5, invalidAuthProto6},
			wantIssues: &core.Issue{
				Position: meta.Position{
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

			r, protos := start(t)

			rule := rules.PackageSameCsharpNamespace{}
			var issues []core.Issue
			for _, fileName := range tc.fileNames {
				issue, err := rule.Validate(protos[fileName])
				r.ErrorIs(err, tc.wantErr)
				if err == nil {
					issues = append(issues, issue...)
				}
			}

			switch {
			case tc.wantIssues != nil:
				r.Contains(issues, *tc.wantIssues)
			case len(issues) > 0:
				r.Empty(issues)
			}
		})
	}
}
