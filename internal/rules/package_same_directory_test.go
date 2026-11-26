package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
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
			fileNames: []string{invalidAuthProto2, invalidAuthProto4},
			wantIssues: &core.Issue{
				Position: core.Position{
					Filename: "",
					Offset:   0,
					Line:     0,
					Column:   0,
				},
				SourceName: "",
				Message:    "",
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

			r, protos := start(t)

			rule := rules.PackageSameDirectory{}
			var got []error
			var infos []core.ProtoInfo
			for _, fileName := range tc.fileNames {
				infos = append(infos, protos[fileName])
			}
			_, err := validate(&rule, infos...)
			got = append(got, err)

			r.ErrorIs(errors.Join(got...), tc.wantErr)
		})
	}
}
