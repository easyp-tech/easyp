package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestPackageSameRubyPackage_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "all files in the same package must have the same ruby_package option"

	rule := rules.PackageSameRubyPackageLinked{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestPackageSameRubyPackage_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames  []string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			// invalid_options/queue.proto (V1) is the exempt first value;
			// invalid_options/session.proto (same package, V2) issues at its
			// own ruby_package option keyword with the constant as written.
			fileNames: []string{invalidAuthProto5, invalidAuthProto6},
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   452,
					Line:     14,
					Column:   1,
				},
				SourceName: `"ZergsLaw-2"`,
				Message:    "all files in the same package must have the same ruby_package option",
				RuleName:   "PACKAGE_SAME_RUBY_PACKAGE",
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

			rule := &rules.PackageSameRubyPackageLinked{}

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
