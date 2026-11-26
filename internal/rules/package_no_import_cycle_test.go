package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestPackageNoImportCycle_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "package import cycle"

	rule := rules.PackageNoImportCycle{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

// The cycle fixtures are a PACKAGE cycle, not a file cycle: the file graph
// a.proto -> b/b.proto -> a2.proto is a DAG, while the package graph cycles
// a.v1 -> b.v1 -> a.v1. The easyprotoc linker hard-rejects a direct file A<->B
// import cycle ("recursively imports itself", pinned 2:1 in
// internal/core/link_test.go), so only this shape is usable as a CLI-driven
// corpus; parseLinkedFile never links, so unit tests may drive any shape.

func TestPackageNoImportCycle_CollectFinish(t *testing.T) {
	t.Parallel()

	// validate() collapses IssueInfo to Issue and drops Path (init_test.go),
	// so the Path-carrying cells drive Collect/Finish directly — the same
	// calls the engine's collectFile/Finish drain makes for ModuleRules.

	t.Run("package_cycle", func(t *testing.T) {
		t.Parallel()

		rule := &rules.PackageNoImportCycle{}
		for _, fileName := range []string{importCycleA, importCycleB, importCycleA2} {
			require.NoError(t, rule.Collect(parseLinkedFile(t, fileName)))
		}

		findings := rule.Finish()
		require.Len(t, findings, 2)

		// Finish walks package keys in sorted order, so a.v1's edge (a.proto)
		// is reported before b.v1's edge (b/b.proto); the lab reports EVERY
		// import leading into the cycle, one chain spelling per direction.
		//
		// Position anchor (hand-computed, Phase-1 precedent): line 5 of both
		// cycle fixtures is their only `import "…";` statement at column 1;
		// the four lines above it are `syntax = "proto3";` (18+1),
		// blank (1), `package …;` (13+1) and blank (1), i.e. byte offset 35.
		// Builder positions carry no Filename — the path rides IssueInfo.Path.
		require.Equal(t, core.IssueInfo{
			Issue: core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   35,
					Line:     5,
					Column:   1,
				},
				SourceName: "a.v1 -> b.v1 -> a.v1",
				Message:    "package import cycle",
				RuleName:   "PACKAGE_NO_IMPORT_CYCLE",
			},
			Path: importCycleA,
		}, findings[0])

		require.Equal(t, core.IssueInfo{
			Issue: core.Issue{
				Position: protoast.Position{
					Filename: "",
					Offset:   35,
					Line:     5,
					Column:   1,
				},
				SourceName: "b.v1 -> a.v1 -> b.v1",
				Message:    "package import cycle",
				RuleName:   "PACKAGE_NO_IMPORT_CYCLE",
			},
			Path: importCycleB,
		}, findings[1])
	})

	t.Run("acyclic_negative", func(t *testing.T) {
		t.Parallel()

		// Dropping b/b.proto leaves a one-directional edge a.v1 -> b.v1;
		// a one-directional edge is not a cycle.
		rule := &rules.PackageNoImportCycle{}
		require.NoError(t, rule.Collect(parseLinkedFile(t, importCycleA)))
		require.NoError(t, rule.Collect(parseLinkedFile(t, importCycleA2)))
		require.Empty(t, rule.Finish())
	})

	t.Run("same_package_negative", func(t *testing.T) {
		t.Parallel()

		// sp1.proto (package same.v1) imports sp2.proto (same package);
		// Collect skips same-package imports, so no edge is recorded.
		rule := &rules.PackageNoImportCycle{}
		require.NoError(t, rule.Collect(parseLinkedFile(t, importCycleSamePkg1)))
		require.NoError(t, rule.Collect(parseLinkedFile(t, importCycleSamePkg2)))
		require.Empty(t, rule.Finish())
	})

	t.Run("state_reset", func(t *testing.T) {
		t.Parallel()

		// Finish resets the edge state (p.edges = nil), so the engine may
		// re-run the rule across repeated passes without double-reporting.
		rule := &rules.PackageNoImportCycle{}
		for _, fileName := range []string{importCycleA, importCycleB, importCycleA2} {
			require.NoError(t, rule.Collect(parseLinkedFile(t, fileName)))
		}
		require.Len(t, rule.Finish(), 2)
		require.Empty(t, rule.Finish())
	})

	t.Run("validate_stub", func(t *testing.T) {
		t.Parallel()

		// Validate is a stateless stub: the engine calls Collect/Finish for
		// ModuleRules and never Validate.
		rule := &rules.PackageNoImportCycle{}
		issues, err := rule.Validate(parseLinkedFile(t, importCycleA))
		require.NoError(t, err)
		require.Nil(t, issues)
	})
}
