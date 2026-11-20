package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestFieldNotRequired_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "field should not be required"

	rule := rules.FieldNotRequired{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestFieldNotRequired_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *core.Issue
		// wantSource pins the only SourceName of a multi-field fixture (used
		// to prove the optional/repeated fields contribute no findings).
		wantSource string
		wantErr    error
	}{
		// Position anchor (hand-computed, Phase-1 precedent): line 6 of the
		// fixture is `  required int32 id = 1;` — `required` spans columns
		// 3-10, `int32` 12-16, so the declared-name token `id` starts at
		// column 18 (byte offset 71). The lab rule reports at Decl.NamePos
		// (the `id` token), NOT at the `required` keyword position, because
		// lab parity is the oracle for new rules.
		"invalid": {
			fileName: requiredFieldProto,
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   71,
					Line:     6,
					Column:   18,
				},
				SourceName: "id",
				Message:    "field should not be required",
				RuleName:   "FIELD_NOT_REQUIRED",
			},
			wantErr: nil,
		},
		"valid_proto3": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
		"optional_and_repeated": {
			// The fixture's `optional string name` and `repeated int32 nums`
			// must not fire: the rule reports only `required` labels, so the
			// file yields exactly the single `id` finding.
			fileName:   requiredFieldProto,
			wantSource: "id",
			wantErr:    nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rule := rules.FieldNotRequired{}

			issues, err := rule.Validate(parseLinkedFile(t, tc.fileName))
			require.ErrorIs(t, err, tc.wantErr)
			switch {
			case tc.wantIssues != nil:
				require.Len(t, issues, 1)
				require.Contains(t, issues, *tc.wantIssues)
			case tc.wantSource != "":
				require.Len(t, issues, 1)
				require.Equal(t, tc.wantSource, issues[0].SourceName)
			case len(issues) > 0:
				require.Empty(t, issues)
			}
		})
	}
}
