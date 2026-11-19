package core

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
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

// finishIssueModuleRule reports preset Finish findings: the vehicle for the
// Finish-side re-filter cells (ignore_only and comment suppression on Finish
// output). Collect is a no-op; Validate is never called by the engine.
type finishIssueModuleRule struct {
	issues []IssueInfo
}

var _ ModuleRule = (*finishIssueModuleRule)(nil)

func (m *finishIssueModuleRule) Message() string { return "module finish finding" }

func (m *finishIssueModuleRule) Validate(LinkedFileInfo) ([]Issue, error) { return nil, nil }

func (m *finishIssueModuleRule) Collect(LinkedFileInfo) error { return nil }

func (m *finishIssueModuleRule) Finish() []IssueInfo { return m.issues }

// writeCorpusInto writes the corpus into an existing root; the nested-file
// cells need the subdirectory created before the files are written.
func writeCorpusInto(t *testing.T, root string, files map[string]string) {
	t.Helper()

	for name, src := range files {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(src), 0o644))
	}
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

	// Config-side suppression parity (D-05): ignore and ignore_only skip
	// findings per rule per path exactly like main, on the walk, the per-file
	// check, and the Finish re-filter.

	t.Run("config_ignore_walk_filter", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "sub"), 0o755))
		writeCorpusInto(t, root, map[string]string{
			"sub/a.proto": "syntax = \"proto3\";\npackage alpha;\nmessage HiddenOne {}\n",
			"b.proto":     "syntax = \"proto3\";\npackage beta;\nmessage VisibleOne {}\n",
		})
		recording := &recordingRule{}
		app := New(Options{
			Logger:              logger.NewNop(),
			LinkedRules:         []LintRule{recording, &doubleRuleAlpha{}},
			Ignore:              []string{"sub"},
			AllowCommentIgnores: false,
			KnownLintRules:      []string{"DOUBLE_RULE_ALPHA"},
		})
		got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// The walk filter drops sub/ before any rule runs: the recording rule
		// never saw sub/a.proto, and only b.proto's finding is reported.
		assert.Equal(t, []string{"b.proto"}, recording.observed())
		require.Len(t, got, 1)
		assert.Equal(t, "b.proto", got[0].Path)
		assert.Equal(t, "DOUBLE_RULE_ALPHA", got[0].RuleName)
	})

	t.Run("ignore_only_dir_prefix", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "sub"), 0o755))
		writeCorpusInto(t, root, map[string]string{
			"sub/a.proto": "syntax = \"proto3\";\npackage alpha;\nmessage HiddenOne {}\n",
			"b.proto":     "syntax = \"proto3\";\npackage beta;\nmessage VisibleOne {}\n",
		})
		app := New(Options{
			Logger:              logger.NewNop(),
			LinkedRules:         []LintRule{&doubleRuleAlpha{}},
			IgnoreOnly:          map[string][]string{"DOUBLE_RULE_ALPHA": {"sub"}},
			AllowCommentIgnores: false,
			KnownLintRules:      []string{"DOUBLE_RULE_ALPHA"},
		})
		got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// Prefix semantics, not exact match: the "sub" entry suppresses
		// "sub/a.proto" while the finding outside the prefix survives.
		require.Len(t, got, 1)
		assert.Equal(t, "b.proto", got[0].Path)
		assert.Equal(t, "DOUBLE_RULE_ALPHA", got[0].RuleName)
	})

	t.Run("ignore_only_finish_refilter", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "sub"), 0o755))
		writeCorpusInto(t, root, map[string]string{
			"sub/a.proto": "syntax = \"proto3\";\npackage alpha;\nmessage HiddenOne {}\n",
			"b.proto":     "syntax = \"proto3\";\npackage beta;\nmessage VisibleOne {}\n",
		})
		module := &finishIssueModuleRule{issues: []IssueInfo{
			{Path: "sub/a.proto", Issue: Issue{Position: Position{Line: 1, Column: 1}, Message: "module finish finding", RuleName: "FINISH_ISSUE_MODULE_RULE"}},
			{Path: "b.proto", Issue: Issue{Position: Position{Line: 1, Column: 1}, Message: "module finish finding", RuleName: "FINISH_ISSUE_MODULE_RULE"}},
		}}
		app := newLinkedEngine(nil, false, map[string][]string{"FINISH_ISSUE_MODULE_RULE": {"sub"}}, module)
		got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// shouldIgnoreLinked applies to Finish output (D-05): the ignoreOnly
		// path's Finish finding is dropped, the one outside survives. The
		// ignoreOnly key is the rule's registry name, as in main.
		require.Len(t, got, 1)
		assert.Equal(t, "b.proto", got[0].Path)
		assert.Equal(t, "FINISH_ISSUE_MODULE_RULE", got[0].RuleName)
	})

	t.Run("finish_comment_refilter", func(t *testing.T) {
		t.Parallel()

		// single.proto: the disable sits on line 5 above SuppressedOne on
		// line 6, so the suppression range covers exactly line 6.
		single := "syntax = \"proto3\";\n\npackage alpha;\n\n// easyp:disable DOUBLE_RULE_ALPHA\nmessage SuppressedOne {}\n"
		for _, tc := range []struct {
			name  string
			allow bool
			want  int
		}{
			{name: "allow_drops_finish_finding", allow: true, want: 0},
			{name: "disallow_keeps_finish_finding", allow: false, want: 1},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				root := writeCorpus(t, map[string]string{"single.proto": single})
				module := &finishIssueModuleRule{issues: []IssueInfo{{
					Path:  "single.proto",
					Issue: Issue{Position: Position{Line: 6, Column: 1}, Message: "module finish finding", RuleName: "DOUBLE_RULE_ALPHA"},
				}}}
				app := newLinkedEngine([]string{"DOUBLE_RULE_ALPHA"}, tc.allow, nil, module)
				got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
				require.NoError(t, err)

				// The Finish-side half of the central filter: the Finish
				// finding on the suppressed line is dropped under
				// AllowCommentIgnores and kept without it.
				require.Len(t, got, tc.want)
				if tc.want == 1 {
					assert.Equal(t, "single.proto", got[0].Path)
					assert.Equal(t, "DOUBLE_RULE_ALPHA", got[0].RuleName)
				}
			})
		}
	})

	t.Run("tab_directive_position_pin", func(t *testing.T) {
		t.Parallel()

		// Suppression matches by LINE: a tab before the comment marker does
		// not affect which lines the directive covers (the matching key never
		// diverges between the engines).
		root := writeCorpus(t, map[string]string{
			"tabbed.proto": prefix + "\t// easyp:disable DOUBLE_RULE_ALPHA\nmessage tabbed_bad {}\n",
		})
		app := newLinkedEngine([]string{"DOUBLE_RULE_ALPHA"}, true, nil, &doubleRuleAlpha{})
		got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)
		require.Empty(t, got)

		// Directive validation errors embed the comment position in protoast's
		// rune-based convention: the tab advances the column by exactly one,
		// so the marker sits at column 2 (asserted as a hand-computed literal,
		// not derived from the engine). The old path's go-protoparser metadata
		// is byte-oriented and protoc's COMPILE positions use tab stops of 8,
		// so column-level divergence on tabbed or multi-byte lines is a
		// Phase 5 observable (research A4, pinned here, not silently accepted).
		root = writeCorpus(t, map[string]string{
			"taberr.proto": prefix + "message keep_me {}\n\n\t// nolint:UNKNOWN_DOUBLE\n",
		})
		app = newLinkedEngine([]string{"DOUBLE_RULE_ALPHA"}, true, nil, &doubleRuleAlpha{})
		_, err = app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.ErrorContains(t, err, `taberr.proto:4:2: unknown or unsupported lint rule "UNKNOWN_DOUBLE" in comment directive`)
	})
}
