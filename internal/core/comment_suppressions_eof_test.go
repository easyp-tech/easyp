package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/easyp-tech/easyp/internal/core"
	disk "github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/rules"
	"github.com/stretchr/testify/require"
)

func TestCommentSuppressionTrailingDirectives(t *testing.T) {
	t.Parallel()
	const prefix = "syntax = \"proto3\"; package example;\n"
	const declarations = "// easyp:disable MESSAGE_PASCAL_CASE\nmessage first_bad {}\nmessage second_bad {}\n"
	for _, tc := range []struct {
		name, body, wantError string
		want                  int
	}{
		{name: "EOF with newline", body: declarations + "// easyp:enable MESSAGE_PASCAL_CASE\n"},
		{name: "EOF without newline", body: declarations + "// easyp:enable MESSAGE_PASCAL_CASE"},
		{name: "block comment at EOF", body: declarations + "/* easyp:enable MESSAGE_PASCAL_CASE */"},
		{name: "orphan enable", body: "message Valid {}\n// easyp:enable MESSAGE_PASCAL_CASE", wantError: "no preceding disable"},
		{name: "unknown trailing rule", body: "message Valid {}\n// easyp:enable UNKNOWN", wantError: "unknown or unsupported"},
		{name: "unattached disable", body: "message Valid {}\n// easyp:disable MESSAGE_PASCAL_CASE", wantError: "must annotate a declaration"},
		{name: "unpaired attached disable", body: declarations, want: 1},
		{name: "marker in string", body: "option java_package = \"// easyp:enable UNKNOWN\";\nmessage bad_name {}", want: 1},
		{name: "nested closing brace", body: "message Outer {\n" + declarations + "// easyp:enable MESSAGE_PASCAL_CASE\n}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(root, "test.proto"), []byte(prefix+tc.body), 0o600))
			app := core.New(core.Options{Logger: logger.NewNop(), Rules: []core.Rule{&rules.MessagePascalCase{}}, AllowCommentIgnores: true, KnownLintRules: rules.AllRuleNames()})
			issues, err := app.Lint(t.Context(), disk.NewFSWalker(root, "test.proto"))
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
			require.Len(t, issues, tc.want)
		})
	}
}
