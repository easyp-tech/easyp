package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestImportNoWeak_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "IMPORT_NO_WEAK"

	rule := rules.ImportNoWeak{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestImportNoWeak_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "import should not be weak"

	rule := rules.ImportNoWeak{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestImportNoWeak_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *lint.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: invalidAuthProto,
			wantIssues: &lint.Issue{

				Position: meta.Position{
					Filename: "",
					Offset:   38,
					Line:     5,
					Column:   1,
				},
				SourceName: `"google/protobuf/empty.proto"`,
				Message:    "import should not be weak",
			},
			wantErr: nil,
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

			r, protos := start(t)

			rule := rules.ImportNoWeak{}
			issues, err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(err, tc.wantErr)
			if tc.wantIssues != nil {
				r.Contains(issues, *tc.wantIssues)
			}
		})
	}
}
