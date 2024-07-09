package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestImportUsed_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "IMPORT_USED"

	rule := rules.ImportUsed{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestImportUsed_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "import is not used"

	rule := rules.ImportUsed{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestImportUsed_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *lint.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: importNotUsed,
			wantIssues: &lint.Issue{
				Position: meta.Position{
					Filename: "",
					Offset:   20,
					Line:     3,
					Column:   1,
				},
				SourceName: `"import_used/messages.proto"`,
				Message:    "import is not used",
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

			r, protos := start(t)

			rule := rules.ImportUsed{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
