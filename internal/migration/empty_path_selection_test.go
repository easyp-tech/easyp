package migration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/modules"
)

func TestMigrationPathsRequireEveryInputToMatch(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name        string
		outsideCopy bool
	}{
		{name: "proven package fallback"},
		{name: "no safe fallback", outsideCopy: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			const dependency = "example.test/dependency"
			writeFixture(t, root, "easyp.yaml", "deps: ["+dependency+"@v1.0.0]\ngenerate:\n  inputs: [{directory: mcp}, {directory: empty}]\n")
			writeFixture(t, root, "mcp/options.proto", "package mcp.options.v1;")
			writeFixture(t, root, "empty/.keep", "")
			outside := "package other.v1;"
			if tt.outsideCopy {
				outside = "package mcp.options.v1;"
			}
			writeFixture(t, root, "outside.proto", outside)
			repo := &mockRepository{fetched: map[string]modules.Fetched{dependency + "@v1.0.0": migrationFetched(dependency, "v1.0.0", testCommit)}}
			plan, err := Build(t.Context(), Options{Dir: root, Module: "example.test/api", ResolveLock: true, Repository: repo})
			if tt.outsideCopy {
				require.ErrorContains(t, err, "scope")
				assert.Empty(t, repo.calls, "unsafe source selection must fail before permitted dependency access")
				assert.NoFileExists(t, filepath.Join(root, "protobuf.mod"))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []string{dependency + "@v1.0.0"}, repo.calls)
			assert.Empty(t, plan.paths, "an empty input cannot become an unmatched runtime selector")
			assert.Equal(t, []string{"mcp.options.v1"}, plan.packages)
			require.NoError(t, plan.Apply())
		})
	}
}
