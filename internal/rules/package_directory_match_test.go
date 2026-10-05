package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestPackageDirectoryMatch_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "package does not match directory path"

	rule := rules.PackageDirectoryMatch{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestPackageDirectoryMatch_Validate(t *testing.T) {
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
					Offset:   20,
					Line:     3,
					Column:   1,
				},
				SourceName: "./../../testdata/auth/service.proto",
				Message:    "package \"Session\" does not match module-relative directory \"auth\"; expected \"auth\"",
				RuleName:   "PACKAGE_DIRECTORY_MATCH",
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

			rule := rules.PackageDirectoryMatch{
				Root: "./../../testdata/",
			}
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

func TestPackageDirectoryMatch_ModuleQualifiedPackage(t *testing.T) {
	t.Parallel()

	r, protos := start(t)
	info := protos[validAuthProto]
	info.ImportPath = "v1/service.proto"
	rule := rules.PackageDirectoryMatch{Root: ".", Prefix: "user"}

	info.Info.ProtoBody.Packages[0].Name = "user.v1"
	issues, err := rule.Validate(info)
	r.NoError(err)
	r.Empty(issues)

	info.Info.ProtoBody.Packages[0].Name = "v1"
	issues, err = rule.Validate(info)
	r.NoError(err)
	r.Empty(issues)

	info.Info.ProtoBody.Packages[0].Name = "other.v1"
	issues, err = rule.Validate(info)
	r.NoError(err)
	r.Len(issues, 1)
	r.Contains(issues[0].Message, `expected "v1" or module-qualified "user.v1"`)
}
