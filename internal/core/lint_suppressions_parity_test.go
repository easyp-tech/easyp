package core

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
)

// messagePascalCaseName is rules.MessagePascalCase's name predicate; the
// double mirrors the oracle rule exactly (same regexp, top-level messages
// only) so the old-path expectations carry over verbatim.
var messagePascalCaseName = regexp.MustCompile("^[A-Z][a-zA-Z0-9]+(?:[A-Z][a-zA-Z0-9]+)*$")

// messagePascalCase reports one finding per badly named top-level message
// under the MESSAGE_PASCAL_CASE registry name: the shape of the oracle rule
// the old-path suppression fixtures drive (rules.MessagePascalCase).
type messagePascalCase struct{}

var _ LintRule = (*messagePascalCase)(nil)

func (m *messagePascalCase) Message() string { return "message name should be PascalCase" }

func (m *messagePascalCase) Validate(info LinkedFileInfo) ([]Issue, error) {
	var res []Issue
	for _, msg := range info.File.Messages {
		if !messagePascalCaseName.MatchString(msg.Name) {
			res = append(res, Issue{
				Position: msg.Pos,
				Message:  m.Message(),
				RuleName: "MESSAGE_PASCAL_CASE",
			})
		}
	}
	return res, nil
}

// TestLintLinkedSuppressionParity re-drives main's suppression fixture matrix
// through the linked engine (LINT-ENG-05, D-04/D-05). The old-path fixtures
// (comment_suppressions_eof_test.go, comment_suppressions_test.go) are the
// byte-identical parity oracle, so their expected values carry over unchanged;
// the only algorithmic delta is the comment walk, which reads protoast's owner
// map instead of reflecting over go-protoparser metadata.
func TestLintLinkedSuppressionParity(t *testing.T) {
	t.Parallel()

	const (
		prefix       = "syntax = \"proto3\"; package example;\n"
		declarations = "// easyp:disable MESSAGE_PASCAL_CASE\nmessage first_bad {}\nmessage second_bad {}\n"
	)

	for _, tc := range []struct {
		name    string
		allow   bool
		body    string
		wantErr string // required substring of the returned error; empty = no error
		want    int    // finding count when wantErr is empty
	}{
		// The nine-cell EOF matrix, fixtures verbatim from
		// comment_suppressions_eof_test.go.
		{name: "eof_with_newline", body: declarations + "// easyp:enable MESSAGE_PASCAL_CASE\n", allow: true, want: 0},
		{name: "eof_without_newline", body: declarations + "// easyp:enable MESSAGE_PASCAL_CASE", allow: true, want: 0},
		{name: "block_comment_at_eof", body: declarations + "/* easyp:enable MESSAGE_PASCAL_CASE */", allow: true, want: 0},
		{name: "orphan_enable", body: "message Valid {}\n// easyp:enable MESSAGE_PASCAL_CASE", allow: true, wantErr: "no preceding disable"},
		{name: "unknown_trailing_rule", body: "message Valid {}\n// easyp:enable UNKNOWN", allow: true, wantErr: "unknown or unsupported"},
		{name: "unattached_disable", body: "message Valid {}\n// easyp:disable MESSAGE_PASCAL_CASE", allow: true, wantErr: "must annotate a declaration"},
		{name: "unpaired_attached_disable", body: declarations, allow: true, want: 1},
		{name: "marker_in_string", body: "option java_package = \"// easyp:enable UNKNOWN\";\nmessage bad_name {}", allow: true, want: 1},
		{name: "nested_closing_brace", body: "message Outer {\n" + declarations + "// easyp:enable MESSAGE_PASCAL_CASE\n}\n", allow: true, want: 0},
		// Directive error texts byte-match main's formats, including the
		// path:line:column prefix shape (comment_suppressions.go malformed and
		// unknown-rule errors).
		{name: "malformed_directive_exact_text", body: "message Valid {}\n// easyp:disable", allow: true, wantErr: `test.proto:3:1: malformed lint directive "easyp:disable"`},
		{name: "unknown_name_exact_text", body: "message Valid {}\n// easyp:disable NOT_A_RULE", allow: true, wantErr: `test.proto:3:1: unknown or unsupported lint rule "NOT_A_RULE" in comment directive`},
		// Policy independence (comment_suppressions_test.go): the same
		// directive fixture suppresses with AllowCommentIgnores and passes
		// through without it.
		{name: "policy_allow_suppresses", body: "// easyp:disable MESSAGE_PASCAL_CASE\nmessage ignored_name {}\nmessage visible_name {}\n", allow: true, want: 1},
		{name: "policy_disallow_keeps", body: "// easyp:disable MESSAGE_PASCAL_CASE\nmessage ignored_name {}\nmessage visible_name {}\n", allow: false, want: 2},
		// Legacy dialects suppress identically to main; legacy names are
		// validated against the known set too.
		{name: "nolint_dialect_suppresses", body: "// nolint:MESSAGE_PASCAL_CASE\nmessage suppressed_one {}\n", allow: true, want: 0},
		{name: "buf_dialect_suppresses", body: "// buf:lint:ignore MESSAGE_PASCAL_CASE\nmessage suppressed_one {}\n", allow: true, want: 0},
		{name: "nolint_unknown_name_error", body: "// nolint:MESSAGE_PASCAL_CASE_typo\nmessage ignored_name {}\n", allow: true, wantErr: `test.proto:2:1: unknown or unsupported lint rule "MESSAGE_PASCAL_CASE_typo" in comment directive`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(root, "test.proto"), []byte(prefix+tc.body), 0o600))
			app := newLinkedEngine([]string{"MESSAGE_PASCAL_CASE"}, tc.allow, nil, &messagePascalCase{})
			got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "test.proto"))
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, tc.want)
			for _, issue := range got {
				require.Equal(t, "MESSAGE_PASCAL_CASE", issue.RuleName, "non-rule finding leaked into output")
				require.Equal(t, "test.proto", issue.Path)
			}
		})
	}
}
