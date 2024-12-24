package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

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
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: invalidAuthProto,
			wantIssues: &core.Issue{

				Position: meta.Position{
					Filename: "",
					Offset:   38,
					Line:     5,
					Column:   1,
				},
				SourceName: `"google/protobuf/empty.proto"`,
				Message:    "import should not be weak",
				RuleName:   "IMPORT_NO_WEAK",
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
			switch {
			case tc.wantIssues != nil:
				r.Contains(issues, *tc.wantIssues)
			case len(issues) > 0:
				r.Empty(issues)
			}
		})
	}
}
