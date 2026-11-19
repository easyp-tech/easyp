package rules_test

// Batch-1 (MINIMAL) parity cells: DirectorySamePackage, PackageDefined and
// PackageSameDirectory join the Wave-0 pilot under cumulative configs. Every
// cell asserts its expected old-side finding count before the multiset
// compare (T-04-02), and the cells pin main's group semantics the lab
// changed: first-exempt Finish, packageless files invisible to group rules,
// and the PACKAGE_DEFINED zero position (Line 0, Column 0).

import (
	"testing"

	"github.com/easyp-tech/easyp/internal/config"
)

// TestLinkedParityMinimal pins old-vs-new parity for the MINIMAL batch. The
// requirePorted guard is the RED assertion: it fails listing the unported
// rules while the package still compiles, and goes green as the batch's ports
// land.
func TestLinkedParityMinimal(t *testing.T) {
	// RED guard — must stay the first statement of the test body.
	requirePorted(t, "DIRECTORY_SAME_PACKAGE", "PACKAGE_DEFINED", "PACKAGE_SAME_DIRECTORY")

	t.Parallel()

	t.Run("first_exempt_directory", func(t *testing.T) {
		t.Parallel()

		// Three files in one directory with disagreeing packages: main's
		// first-exempt cache keeps the first file's package and issues only on
		// the SECOND and later files, at their own package positions (the lab's
		// report-everyone mismatches would report all three).
		cfg := config.LintConfig{Use: []string{"DIRECTORY_SAME_PACKAGE"}}
		assertBatchParity(t, cfg, map[string]string{
			"grp/a.proto": "syntax = \"proto3\";\n\npackage foo;\n",
			"grp/b.proto": "syntax = \"proto3\";\n\npackage bar;\n",
			"grp/c.proto": "syntax = \"proto3\";\n\npackage baz;\n",
		}, 2)
	})

	t.Run("packageless_invisible", func(t *testing.T) {
		t.Parallel()

		// A file without a package statement contributes nothing to the
		// directory and package groups (main iterates Packages only); the lab's
		// packagePosition fallback would fabricate an empty value at 1:1. The
		// asserted zero is the main shape, not a vacuous pass.
		cfg := config.LintConfig{Use: []string{"DIRECTORY_SAME_PACKAGE", "PACKAGE_SAME_DIRECTORY"}}
		assertBatchParity(t, cfg, map[string]string{
			"pk/none.proto": "syntax = \"proto3\";\n",
			"pk/has.proto":  "syntax = \"proto3\";\n\npackage foo;\n",
		}, 0)
	})

	t.Run("first_exempt_package_dir", func(t *testing.T) {
		t.Parallel()

		// One package split across two directories: the first directory is
		// exempt, the later file issues at its own package position.
		cfg := config.LintConfig{Use: []string{"PACKAGE_SAME_DIRECTORY"}}
		assertBatchParity(t, cfg, map[string]string{
			"dirA/same.proto": "syntax = \"proto3\";\n\npackage foo;\n",
			"dirB/same.proto": "syntax = \"proto3\";\n\npackage foo;\n",
		}, 1)
	})

	t.Run("package_defined_zero_pos", func(t *testing.T) {
		t.Parallel()

		// PACKAGE_DEFINED reports a packageless file at the zero position
		// (Filename only; Line 0, Column 0 — the fingerprint compares both), and
		// reports nothing for a file with a package.
		cfg := config.LintConfig{Use: []string{"PACKAGE_DEFINED"}}
		assertBatchParity(t, cfg, map[string]string{
			"pd/empty.proto": "syntax = \"proto3\";\n",
			"pd/full.proto":  "syntax = \"proto3\";\n\npackage foo;\n",
		}, 1)
	})

	t.Run("pilot_cumulative", func(t *testing.T) {
		t.Parallel()

		// The Wave-0 pilot (PACKAGE_DIRECTORY_MATCH) re-verified under the
		// cumulative MINIMAL config: the dir/package mismatch file issues from
		// both the pilot and DIRECTORY_SAME_PACKAGE; the group rules add
		// nothing else.
		cfg := config.LintConfig{Use: []string{
			"DIRECTORY_SAME_PACKAGE",
			"PACKAGE_DEFINED",
			"PACKAGE_SAME_DIRECTORY",
			"PACKAGE_DIRECTORY_MATCH",
		}}
		assertBatchParity(t, cfg, map[string]string{
			"pilot/one.proto":   "syntax = \"proto3\";\n\npackage pilot;\n",
			"pilot/two.proto":   "syntax = \"proto3\";\n\npackage pilot;\n",
			"pilot/three.proto": "syntax = \"proto3\";\n\npackage other;\n",
		}, 2)
	})
}
