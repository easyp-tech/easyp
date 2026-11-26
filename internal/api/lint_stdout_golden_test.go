package api

// The Gate-3 CLI stdout golden (05-01): the real binary's exact stdout lines
// and exit code for a corpus exercising the four observable behaviors the
// Phase-5 flip must preserve — pinned as committed literals captured from the
// CURRENT (pre-flip) binary.
//
// The test is flip-invariant: it builds ./cmd/easyp via buildEasypBinary and
// drives `easyp lint --path .` through the real routing (actionV1, the
// printer, the exit codes), so at the flip it becomes the primary stdout gate
// with no re-point.
//
// One corpus covers all four behaviors:
//   - findings: FILE_LOWER_SNAKE_CASE enabled and firing on
//     example_pkg/Bad_File.proto (printed);
//   - per-path rule suppression: issues.exclude-rules suppresses that rule on
//     exactly one of the two example_pkg files (Excluded_File.proto's finding
//     is absent from stdout while Bad_File's identical-rule finding prints);
//   - comment directives: allow_comment_ignores plus an `easyp:disable
//     PACKAGE_DIRECTORY_MATCH` directive annotating the package declaration
//     suppresses suppressed_dir/sup_file.proto's finding (absent from stdout);
//   - the PACKAGE_DIRECTORY_MATCH SourceName-to-Path rewrite
//     (lint_v1.go:213-215) fires on wrong_dir/file.proto: the rule matches,
//     SourceName is non-empty, and the rendered path is the rewritten
//     filepath.ToSlash(SourceName). For a direct lint (no module import
//     roots) SourceName equals the engine Path, so the rewrite is
//     byte-identical here — the rewrite code point is exercised and the
//     rendered path is the rewritten one, but distinguishing SourceName
//     values would need a module policy with import roots.
//
// Config spelling note (the plan's confirm-the-v1-key loop): the v1 policy
// grammar has NO ignore_only key — v1.Policy decodes with KnownFields(true),
// so `linters.ignore_only` would be rejected loudly. The confirmed v1
// spelling for "suppress rule R on exactly one of two files" is
// issues.exclude-rules with path + linters, applied by actionV1 after the
// engine run; engine-level ignore_only stays covered by the rules-package
// tests.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// stdoutGoldenConfig selects the MINIMAL group plus FILE_LOWER_SNAKE_CASE,
// with allow_comment_ignores and one per-path exclusion. This is the v1
// linters spelling (the same shape linkMatrixConfig uses) extended with the
// issues section.
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

// stdoutGoldenFiles is the committed fixture: two violating example_pkg files
// (one excluded per path, one printed), a package/directory mismatch that
// prints through the SourceName-to-Path rewrite, and a directive-suppressed
// mismatch. Every file carries exactly one intended finding, so the rendered
// stdout is deterministic.
var stdoutGoldenFiles = map[string]string{
	"example_pkg/Bad_File.proto": "syntax = \"proto3\";\n\npackage example_pkg;\n\nmessage CleanFile {\n  string name = 1;\n}\n",

	// The excluded twin: identical violation (the file name), identical
	// rule, suppressed per path — the per-path rule suppression is
	// observable only against Bad_File.proto's printed line. The message
	// name differs from Bad_File.proto's on purpose: the corpus must compile
	// under the linked engine's full type resolution, and two same-package
	// files declaring the same message name would route the whole run to
	// COMPILE, vacating the rule-finding golden (research Pitfall 7; the
	// pinned stdout lines were re-captured from the byte-frozen pre-flip
	// binary over this corrected fixture and are byte-identical to the 05-01
	// literals).
	"example_pkg/Excluded_File.proto": "syntax = \"proto3\";\n\npackage example_pkg;\n\nmessage ExcludedMsg {\n  string name = 1;\n}\n",

	// The package/directory mismatch that prints: the directory does not
	// match the package, and the finding renders through the
	// SourceName-to-Path rewrite at lint_v1.go:213-215.
	"wrong_dir/file.proto": "syntax = \"proto3\";\n\npackage other_pkg;\n",

	// The directive-suppressed mismatch: the `easyp:disable
	// PACKAGE_DIRECTORY_MATCH` comment annotates the package declaration, so
	// the finding's line falls inside the suppression range and never
	// reaches the printer.
	"suppressed_dir/sup_file.proto": "syntax = \"proto3\";\n\n// easyp:disable PACKAGE_DIRECTORY_MATCH\npackage sup_pkg;\n",
}

// goldenStdout is the exact stdout of the CURRENT (pre-flip) binary over the
// fixture, captured 2026-11-20 (05-01 Task 3), and the exit code is 1. Lines
// are path-sorted (the engine's D-01 order preserved through the api stable
// Path sort), each ending with the " (RULE_NAME)" suffix of the pinned
// text printer.
const goldenStdout = "" +
	"example_pkg/Bad_File.proto:0:0:example_pkg/Bad_File.proto file name should be lower_snake_case.proto (FILE_LOWER_SNAKE_CASE)\n" +
	"wrong_dir/file.proto:3:1:wrong_dir/file.proto package \"other_pkg\" does not match module-relative directory \"wrong_dir\"; expected \"wrong_dir\" (PACKAGE_DIRECTORY_MATCH)\n"

// writeStdoutGoldenFixture writes easyp.yaml plus the committed proto files
// into a per-test temp dir and returns the root.
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

// TestLintStdoutGolden pins the CLI stdout and exit code through the real
// binary. Process state is asserted in the child process only, so the test
// stays sequential (the go-testing process-state exception, same as
// TestLintLinkMatrixExitCodes). The pinned literals are captured from the
// pre-flip binary; the flip must reproduce them byte for byte.
func TestLintStdoutGolden(t *testing.T) {
	binPath, err := buildEasypBinary()
	require.NoError(t, err)

	stdout, stderr, code := runEasypLint(t, binPath, writeStdoutGoldenFixture(t))
	require.Equal(t, 1, code, "findings present: the exit code literal is 1 (stderr: %s)", stderr)
	require.Equal(t, goldenStdout, stdout, "the flip must reproduce the pinned stdout byte for byte")
}
