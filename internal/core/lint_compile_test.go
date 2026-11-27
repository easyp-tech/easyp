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

func TestLintReportsCompileErrorsInsteadOfRules(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	raw := "syntax = \"proto3\";\npackage example;\nmessage bad_name {\n  required string s = 1;\n}\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "test.proto"), []byte(raw), 0o644))
	app := core.New(core.Options{Logger: logger.NewNop(), Rules: []core.Rule{&rules.MessagePascalCase{}}, KnownLintRules: rules.AllRuleNames()})

	got, err := app.Lint(t.Context(), disk.NewFSWalker(root, "."))
	require.NoError(t, err)
	require.Len(t, got, 1, "rules do not run on files that do not compile")
	require.Equal(t, core.CompileRuleName, got[0].RuleName)
	require.Equal(t, "test.proto", got[0].Path)
	require.Equal(t, "Required fields are not allowed in proto3.", got[0].Message)
	require.Equal(t, 4, got[0].Position.Line)
	require.Equal(t, 3, got[0].Position.Column)
}

func TestLintMissingImportIsNotACompileIssue(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	raw := "syntax = \"proto3\";\npackage example;\nimport \"missing.proto\";\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "test.proto"), []byte(raw), 0o644))
	app := core.New(core.Options{Logger: logger.NewNop(), Rules: []core.Rule{&rules.ImportUsed{}}, KnownLintRules: rules.AllRuleNames()})

	_, err := app.Lint(t.Context(), disk.NewFSWalker(root, "."))
	var notFound *core.OpenImportFileError
	require.ErrorAs(t, err, &notFound, "a missing import keeps the dedicated error and exit code")
}
