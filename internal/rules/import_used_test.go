package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestImportUsed_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "import is not used"

	rule := rules.ImportUsedLinked{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestImportUsed_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: importNotUsed,
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   20,
					Line:     3,
					Column:   1,
				},
				SourceName: `"import_used/messages.proto"`,
				Message:    "import is not used",
				RuleName:   "IMPORT_USED",
			},
			wantErr: nil,
		},
		"valid": {
			fileName: importUsed,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rule := rules.ImportUsedLinked{}

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
