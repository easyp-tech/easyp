package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestDirectorySamePackage_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "DIRECTORY_SAME_PACKAGE"

	rule := rules.DirectorySamePackage{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestDirectorySamePackage_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "all files in the same directory must have the same package name"

	rule := rules.DirectorySamePackage{}
	message := rule.Message()

	assert.Equal(expMessage, message)

}
func TestDirectorySamePackage_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames  []string
		wantIssues *lint.Issue
		wantErr    error
	}{
		"invalid": {
			fileNames: []string{invalidAuthProto, invalidAuthProto2},
			wantIssues: &lint.Issue{
				Position: meta.Position{
					Filename: "",
					Offset:   20,
					Line:     3,
					Column:   1,
				},
				SourceName: "Queue",
				Message:    "all files in the same directory must have the same package name",
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

			rule := rules.DirectorySamePackage{}
			var res []lint.Issue
			for _, fileName := range tc.fileNames {
				issues, err := rule.Validate(protos[fileName])
				r.ErrorIs(err, tc.wantErr)
				res = append(res, issues...)
			}
			if len(res) > 0 {
				r.Contains(res, *tc.wantIssues)
			}
		})
	}
}
