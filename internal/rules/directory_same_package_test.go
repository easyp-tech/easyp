package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestDirectorySamePackage_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "all files in the same directory must have the same package name"

	rule := rules.DirectorySamePackageLinked{}
	message := rule.Message()

	assert.Equal(expMessage, message)

}

func TestDirectorySamePackage_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames  []string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			fileNames: []string{invalidAuthProto, invalidAuthProto2},
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   20,
					Line:     3,
					Column:   1,
				},
				SourceName: "Queue",
				Message:    "all files in the same directory must have the same package name",
				RuleName:   "DIRECTORY_SAME_PACKAGE",
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

			rule := &rules.DirectorySamePackageLinked{}

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
