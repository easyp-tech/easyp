package api

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/logger"
)

// linkMatrixConfig selects FILE_LOWER_SNAKE_CASE beside the MINIMAL group:
// the smallest rule set the exit-0 and lint-issue classes can be proven with.
// This is the v1 spelling of the legacy `lint: use:` shape, which ParsePolicy
// rejects instead of accepting (the legacy rejection is pinned elsewhere).
const linkMatrixConfig = "version: v1\nlinters:\n  default: MINIMAL\n  enable:\n    - FILE_LOWER_SNAKE_CASE\n"

// writeLinkMatrixFixture writes an easyp.yaml plus the named proto files
// (slash-separated names) into a per-case temp dir and returns the root.
func writeLinkMatrixFixture(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "easyp.yaml"), []byte(linkMatrixConfig), 0o600))
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}
	return root
}

// runLinkMatrixAction invokes actionV1 against a fixture root with a private
// flag set and CLI context per case (urfave/cli v2 parallel-test rules: no
// shared parser state, no shared cli.Context).
func runLinkMatrixAction(t *testing.T, root string) error {
	t.Helper()

	configPath := filepath.Join(root, "easyp.yaml")
	set := flag.NewFlagSet("lint-link-matrix", flag.ContinueOnError)
	set.String(flags.ConfigName, "", "")
	require.NoError(t, set.Set(flags.ConfigName, configPath))
	ctx := cli.NewContext(&cli.App{}, set, nil)
	ctx.Context = t.Context()

	return (Lint{}).actionV1(ctx, logger.NewNop(), configPath, root, root)
}

// TestLintLinkMatrix proves the D-05 error-class matrix in process: each class
// routes to its existing error carrier through the unchanged actionV1 path.
// The per-file read path is what runs here (the linker stays dormant, D-04),
// so no COMPILE remarks can appear at this layer — that absence is the
// "without COMPILE issues" guarantee for the missing-import class.
func TestLintLinkMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		files map[string]string
		check func(t *testing.T, err error)
	}{
		{
			// Exit-2 class: the typed carrier Lint.Action routes to
			// errExit(2, "Cannot import file") must carry the import path.
			name: "missing_import_exit2_class",
			files: map[string]string{
				"root.proto": "syntax = \"proto3\";\n\npackage example_pkg;\n\nimport \"missing.proto\";\n",
			},
			check: func(t *testing.T, err error) {
				var notFound *core.OpenImportFileError
				require.ErrorAs(t, err, &notFound)
				require.Equal(t, "missing.proto", notFound.FileName)
			},
		},
		{
			// Exit-1 command-error class: a syntax error surfaces as a plain
			// error (main's log.Fatal route), never as the missing-import
			// carrier.
			name: "syntax_error_command_class",
			files: map[string]string{
				"broken.proto": "syntax = \"proto3\";\n\npackage example_pkg;\n\nmessage Broken {\n",
			},
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				var notFound *core.OpenImportFileError
				require.NotErrorAs(t, err, &notFound)
			},
		},
		{
			// Exit-1 lint-issue class: the corpus compiles, a rule finding is
			// printed, and actionV1 returns the ErrHasLintIssue sentinel that
			// Lint.Action routes to exit code 1.
			name: "lint_issue_exit1_class",
			files: map[string]string{
				// The directory matches the package so FILE_LOWER_SNAKE_CASE
				// is the only rule finding.
				"example_pkg/Bad_File.proto": "syntax = \"proto3\";\n\npackage example_pkg;\n\nmessage CleanFile {\n  string name = 1;\n}\n",
			},
			check: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrHasLintIssue)
			},
		},
		{
			// Exit-0 class: a clean compliant corpus routes nil.
			name: "compiling_exit0_class",
			files: map[string]string{
				"example_pkg/clean_file.proto": "syntax = \"proto3\";\n\npackage example_pkg;\n\nmessage CleanFile {\n  string name = 1;\n}\n",
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := runLinkMatrixAction(t, writeLinkMatrixFixture(t, tt.files))
			tt.check(t, err)
		})
	}
}
