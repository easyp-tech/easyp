package migration

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationNonportablePathCandidateUsesStrictPackageFallback(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("the legacy colon directory cannot be created on Windows")
	}
	for _, tt := range []struct {
		name    string
		partial bool
	}{
		{name: "complete package"},
		{name: "partial package rejects fallback", partial: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFixture(t, root, "easyp.yaml", "generate:\n  inputs: [{directory: 'api:legacy'}]\n")
			writeFixture(t, root, "api:legacy/a.proto", "package selected.v1;")
			outside := "package other.v1;"
			if tt.partial {
				outside = "package selected.v1;"
			}
			writeFixture(t, root, "other.proto", outside)
			plan, err := Build(t.Context(), Options{Dir: root, Module: "example.test/api"})
			if tt.partial {
				require.ErrorContains(t, err, "scope")
				assert.NoFileExists(t, filepath.Join(root, "protobuf.mod"))
				return
			}
			require.NoError(t, err)
			assert.Empty(t, plan.local.selection.paths)
			assert.Equal(t, []string{"selected.v1"}, plan.local.selection.packages)
			require.NoError(t, plan.Apply())
		})
	}
}
