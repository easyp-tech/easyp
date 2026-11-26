package rules_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestPackageSameSwiftPrefix_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "all files in the same package must have the same swift_prefix option"

	rule := rules.PackageSameSwiftPrefix{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestPackageSameSwiftPrefix_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames  []string
		wantIssues *core.Issue
		wantErr    error
	}{
		"invalid": {
			// invalid_options/queue.proto (V1) is the exempt first value;
			// invalid_options/session.proto (same package, V2) issues at its
			// own ruby_package option keyword with the constant as written —
			// the rule registered as swift_prefix compares ruby_package
			// (main's bug-for-bug pin, T-04-09): the literal below is
			// ruby_package's position and value, not swift_prefix's.
			fileNames: []string{invalidAuthProto5, invalidAuthProto6},
			wantIssues: &core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   452,
					Line:     14,
					Column:   1,
				},
				SourceName: `"ZergsLaw-2"`,
				Message:    "all files in the same package must have the same swift_prefix option",
				RuleName:   "PACKAGE_SAME_SWIFT_PREFIX",
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

			rule := &rules.PackageSameSwiftPrefix{}

			var infos []core.ProtoInfo
			for _, fileName := range tc.fileNames {
				infos = append(infos, parseLinkedFile(t, fileName))
			}

			issues, err := validate(rule, infos...)
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

// TestPackageSameSwiftPrefix_RubyPackageBugPin carries the swift_prefix_bug_pin
// parity corpus into the carried suite (threat T-04-09): differing ruby_package
// values issue under the rule registered as swift_prefix, while differing
// swift_prefix values with matching ruby_package values stay clean. Fixing the
// option name here would be a behavior change that breaks Phase-7 parity — the
// corrected behavior is a deferred post-parity decision (see the type header).
func TestPackageSameSwiftPrefix_RubyPackageBugPin(t *testing.T) {
	t.Parallel()

	writeInfos := func(t *testing.T, files map[string]string) []core.ProtoInfo {
		t.Helper()

		var infos []core.ProtoInfo
		for _, name := range []string{"a.proto", "b.proto"} {
			full := filepath.Join(t.TempDir(), name)
			require.NoError(t, os.WriteFile(full, []byte(files[name]), 0o644))
			infos = append(infos, parseLinkedFile(t, full))
		}
		return infos
	}

	t.Run("differing_ruby_package_issues", func(t *testing.T) {
		t.Parallel()

		infos := writeInfos(t, map[string]string{
			"a.proto": "syntax = \"proto3\";\n\npackage pin;\n\noption ruby_package = \"a.ruby\";\n\nmessage Am {}\n",
			"b.proto": "syntax = \"proto3\";\n\npackage pin;\n\noption ruby_package = \"b.ruby\";\n\nmessage Bm {}\n",
		})

		issues, err := validate(&rules.PackageSameSwiftPrefix{}, infos...)
		require.NoError(t, err)
		require.Len(t, issues, 1)
		require.Equal(t, "PACKAGE_SAME_SWIFT_PREFIX", issues[0].RuleName)
		require.Equal(t, protoast.Position{Filename: "", Offset: 34, Line: 5, Column: 1}, issues[0].Position)
		require.Equal(t, `"b.ruby"`, issues[0].SourceName)
		require.Equal(t, "all files in the same package must have the same swift_prefix option", issues[0].Message)
	})

	t.Run("differing_swift_prefix_stays_clean", func(t *testing.T) {
		t.Parallel()

		infos := writeInfos(t, map[string]string{
			"a.proto": "syntax = \"proto3\";\n\npackage pin2;\n\noption swift_prefix = \"pre.a\";\n\nmessage Am {}\n",
			"b.proto": "syntax = \"proto3\";\n\npackage pin2;\n\noption swift_prefix = \"pre.b\";\n\nmessage Bm {}\n",
		})

		issues, err := validate(&rules.PackageSameSwiftPrefix{}, infos...)
		require.NoError(t, err)
		require.Empty(t, issues)
	})
}
