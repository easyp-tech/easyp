package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	disk "github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestCommentSuppressionPoliciesAreIndependent(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		allow bool
		want  int
	}{{"enabled", true, 1}, {"disabled", false, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			raw := "syntax = \"proto3\"; package example;\n// easyp:disable MESSAGE_PASCAL_CASE\nmessage ignored_name {}\nmessage visible_name {}\n"
			require.NoError(t, os.WriteFile(filepath.Join(root, "test.proto"), []byte(raw), 0o644))
			app := core.New(core.Options{Logger: logger.NewNop(), Rules: []core.Rule{&rules.MessagePascalCase{}}, AllowCommentIgnores: tc.allow, KnownLintRules: rules.AllRuleNames()})
			got, err := app.Lint(t.Context(), disk.NewFSWalker(root, "test.proto"))
			require.NoError(t, err)
			require.Len(t, got, tc.want)
		})
	}
}

func TestCommentDirectiveNamesAreExact(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	raw := "syntax = \"proto3\"; package example;\n// nolint:MESSAGE_PASCAL_CASE_typo\nmessage ignored_name {}\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "test.proto"), []byte(raw), 0o644))
	app := core.New(core.Options{Logger: logger.NewNop(), Rules: []core.Rule{&rules.MessagePascalCase{}}, AllowCommentIgnores: true, KnownLintRules: rules.AllRuleNames()})
	_, err := app.Lint(t.Context(), disk.NewFSWalker(root, "test.proto"))
	require.ErrorContains(t, err, "MESSAGE_PASCAL_CASE_typo")
}
