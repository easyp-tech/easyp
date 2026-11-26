package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const stdoutGoldenConfig = "version: v1\n" +
	"linters:\n" +
	"  default: MINIMAL\n" +
	"  enable:\n" +
	"    - FILE_LOWER_SNAKE_CASE\n" +
	"  allow_comment_ignores: true\n" +
	"issues:\n" +
	"  exclude-rules:\n" +
	"    - path: \"example_pkg/Excluded_File.proto\"\n" +
	"      linters:\n" +
	"        - FILE_LOWER_SNAKE_CASE\n"

var stdoutGoldenFiles = map[string]string{
	"example_pkg/Bad_File.proto":      "syntax = \"proto3\";\n\npackage example_pkg;\n\nmessage CleanFile {\n  string name = 1;\n}\n",
	"example_pkg/Excluded_File.proto": "syntax = \"proto3\";\n\npackage example_pkg;\n\nmessage ExcludedMsg {\n  string name = 1;\n}\n",
	"wrong_dir/file.proto":            "syntax = \"proto3\";\n\npackage other_pkg;\n",
	"suppressed_dir/sup_file.proto":   "syntax = \"proto3\";\n\n// easyp:disable PACKAGE_DIRECTORY_MATCH\npackage sup_pkg;\n",
}

const goldenStdout = "" +
	"example_pkg/Bad_File.proto:1:1:example_pkg/Bad_File.proto file name should be lower_snake_case.proto (FILE_LOWER_SNAKE_CASE)\n" +
	"wrong_dir/file.proto:3:1:wrong_dir/file.proto package \"other_pkg\" does not match module-relative directory \"wrong_dir\"; expected \"wrong_dir\" (PACKAGE_DIRECTORY_MATCH)\n"

func writeStdoutGoldenFixture(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "easyp.yaml"), []byte(stdoutGoldenConfig), 0o600))
	for name, content := range stdoutGoldenFiles {
		path := filepath.Join(root, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}
	return root
}

// TestLintStdoutGolden verifies CLI stdout formatting, per-path exclusions,
// comment directives, and exit code 1 through the compiled easyp binary.
func TestLintStdoutGolden(t *testing.T) {
	binPath, err := buildEasypBinary()
	require.NoError(t, err)

	stdout, stderr, code := runEasypLint(t, binPath, writeStdoutGoldenFixture(t))
	require.Equal(t, 1, code, "expected exit code 1 when lint findings are present (stderr: %s)", stderr)
	require.Equal(t, goldenStdout, stdout)
}
