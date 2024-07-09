package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestPackageSameCSharpNamespace_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "PACKAGE_SAME_CSHARP_NAMESPACE"

	rule := rules.PackageSameCsharpNamespace{}
	name := rule.Name()

	assert.Equal(expName, name)
}

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
		wantIssues *lint.Issue
		wantErr    error
	}{
		"invalid": {
			fileNames: []string{invalidAuthProto5, invalidAuthProto6},
			wantIssues: &lint.Issue{
				Position: meta.Position{
					Filename: "",
					Offset:   142,
					Line:     9,
					Column:   1,
				},
				SourceName: `"ZergsLaw.BackTemplate.Api.Session.V2"`,
				Message:    "different proto files in the same package should have the same csharp_namespace",
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
			var issues []lint.Issue
			for _, fileName := range tc.fileNames {
				issue, err := rule.Validate(protos[fileName])
				r.ErrorIs(err, tc.wantErr)
				if err == nil {
					issues = append(issues, issue...)
				}
			}

			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
