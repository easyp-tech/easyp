package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestPackageDefined_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "package should be defined"

	rule := rules.PackageDefined{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestPackageDefined_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName   string
		wantIssues *lint.Issue
		wantErr    error
	}{
		"invalid": {
			fileName: invalidAuthProtoEmptyPkg,
			wantIssues: &lint.Issue{
				Position: meta.Position{
					Filename: "./../../../testdata/auth/empty_pkg.proto",
					Offset:   0,
					Line:     0,
					Column:   0,
				},
				SourceName: "./../../../testdata/auth/empty_pkg.proto",
				Message:    "package should be defined",
				RuleName:   "PACKAGE_DEFINED",
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

			rule := rules.PackageDefined{}
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
