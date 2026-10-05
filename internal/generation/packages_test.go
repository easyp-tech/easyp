package generation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestSelectedSourceFilesRespectsBufSelection(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name         string
		paths        []string
		matchedPaths map[string]bool
	}{
		{name: "package_only", matchedPaths: map[string]bool{}},
		{name: "package_and_path", paths: []string{"selected", "excluded"}, matchedPaths: map[string]bool{"selected": true}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "proto/selected/a.proto", "syntax = \"proto3\"; package selected.v1; message A {}")
			writeV1GenerateFixture(t, root, "proto/excluded/b.proto", "syntax = \"proto3\"; package excluded.v1; message B {}")
			selected := v1GenerationModule{directory: root, module: v1.Module{Name: "example.test/dep", Roots: []string{"proto"}, ProtoFilters: []v1.ProtoFileFilter{{Root: "proto", Excludes: []string{"proto/excluded"}}}}}

			files, packages, paths, err := selectedSourceFiles(t.Context(), selected, []string{"selected.v1", "excluded.v1"}, tt.paths)

			require.NoError(t, err)
			assert.Equal(t, []string{"selected/a.proto"}, files)
			assert.Equal(t, map[string]bool{"selected.v1": true}, packages)
			assert.Equal(t, tt.matchedPaths, paths)
		})
	}
}
