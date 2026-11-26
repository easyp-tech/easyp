package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestEnumFirstValueZero_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "enum first value must be zero"

	rule := rules.EnumFirstValueZero{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestEnumFirstValueZero_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: invalidAuthProto,
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   917,
					Line:     51,
					Column:   3,
				},
				SourceName: "4",
				Message:    "enum first value must be zero",
				RuleName:   "ENUM_FIRST_VALUE_ZERO",
			},
			wantErr: nil,
		},
		"invalid_nested": {
			fileName: invalidAuthProto,
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   667,
					Line:     33,
					Column:   5,
				},
				SourceName: "4",
				Message:    "enum first value must be zero",
				RuleName:   "ENUM_FIRST_VALUE_ZERO",
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

			rule := rules.EnumFirstValueZero{}

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

func TestEnumFirstValueZero_EmptyEnumGuard(t *testing.T) {
	t.Parallel()

	// Unit pin for the one documented deviation (research A7, threat T-04-10):
	// main indexes the first value unguarded and panics on an empty enum; the
	// port keeps the guard so Validate stays total. protoc rejects empty enums,
	// so the engine path never sees one — the fixture is parsed directly here,
	// bypassing the compile gate.
	const src = `syntax = "proto3";

package guard;

enum Empty {
}
`

	file, err := protoast.Parse("empty.proto", []byte(src))
	require.NoError(t, err)

	rule := rules.EnumFirstValueZero{}

	issues, err := rule.Validate(core.ProtoInfo{Path: "empty.proto", File: file})
	require.NoError(t, err)
	require.Empty(t, issues)
}
