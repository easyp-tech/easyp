package rules_test

// Wave-0 parity cells: the batch mechanism proven end-to-end on the pilot
// rule (PackageDirectoryMatch). Every cell asserts its expected old-side
// count before the multiset compare — a passing cell with an unasserted
// count would be vacuous parity (T-04-02).

import (
	"testing"

	"github.com/easyp-tech/easyp/internal/config"
)

// TestLinkedParityWave0 pins old-vs-new parity for the pilot rule on the
// corpus shapes later batches reuse.
func TestLinkedParityWave0(t *testing.T) {
	t.Parallel()

	t.Run("mismatch_dir_package", func(t *testing.T) {
		t.Parallel()

		// Three files in one subdirectory: two match their directory, the
		// third does not. PackageDirectoryMatch is a per-file rule — main has
		// no first-exempt for it, so exactly one finding (the third file).
		cfg := config.LintConfig{Use: []string{"PACKAGE_DIRECTORY_MATCH"}}
		requirePorted(t, "PACKAGE_DIRECTORY_MATCH")
		assertBatchParity(t, cfg, map[string]string{
			"foo/one.proto":   "syntax = \"proto3\";\n\npackage foo;\n",
			"foo/two.proto":   "syntax = \"proto3\";\n\npackage foo;\n",
			"foo/three.proto": "syntax = \"proto3\";\n\npackage bar;\n",
		}, 1)
	})

	t.Run("match_dir_package", func(t *testing.T) {
		t.Parallel()

		// Directories match packages: both engines report zero findings, and
		// the zero is asserted explicitly (not vacuous).
		cfg := config.LintConfig{Use: []string{"PACKAGE_DIRECTORY_MATCH"}}
		requirePorted(t, "PACKAGE_DIRECTORY_MATCH")
		assertBatchParity(t, cfg, map[string]string{
			"api/session/v1/session.proto": "syntax = \"proto3\";\n\npackage api.session.v1;\n",
			"auth/service.proto":           "syntax = \"proto3\";\n\npackage auth;\n",
		}, 0)
	})

	t.Run("nested_decls_shallow", func(t *testing.T) {
		t.Parallel()

		// Pins the shallow-walk shape (04-RESEARCH flagged assumption: main's
		// per-file rules walk top-level declarations plus the one extra level
		// main had; go-protoparser's unordered model keeps nesting nested, so
		// main never reported nested declarations). The nested message and
		// enum are badly named per the etalon's recursive walk; the
		// top-level declaration is well named. PackageDirectoryMatch ignores
		// nesting entirely, so this cell's parity is trivially green here;
		// its real teeth land in 04-03/04-06 — the fixture lives here so
		// every batch reuses the same corpus shape.
		cfg := config.LintConfig{Use: []string{"PACKAGE_DIRECTORY_MATCH"}}
		requirePorted(t, "PACKAGE_DIRECTORY_MATCH")
		assertBatchParity(t, cfg, map[string]string{
			"outer/outer.proto": "syntax = \"proto3\";\n\npackage outer;\n\n" +
				"message WellNamed {\n" +
				"\tmessage Bad_Nested {\n" +
				"\t\tenum lower_enum {\n" +
				"\t\t\tZERO = 0;\n" +
				"\t\t}\n" +
				"\t}\n" +
				"}\n",
		}, 0)
	})
}
