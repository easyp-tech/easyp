package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestPackageSameDirectory_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "different proto files in the same package should be in the same directory"

	rule := rules.PackageSameDirectory{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestPackageSameDirectory_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames  []string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			// auth/queue.proto (package Queue) is the exempt first value;
			// invalid_pkg/queue.proto (same package, another directory) issues
			// at its own package-keyword position.
			fileNames: []string{invalidAuthProto2, invalidAuthProto4},
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   20,
					Line:     3,
					Column:   1,
				},
				SourceName: "Queue",
				Message:    "different proto files in the same package should be in the same directory",
				RuleName:   "PACKAGE_SAME_DIRECTORY",
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

			rule := &rules.PackageSameDirectory{}

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
