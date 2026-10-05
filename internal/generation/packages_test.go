package generation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestSelectedPackageFilesRespectsBufSelection(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "proto/selected/a.proto", "syntax = \"proto3\"; package selected.v1; message A {}")
	writeV1GenerateFixture(t, root, "proto/excluded/b.proto", "syntax = \"proto3\"; package excluded.v1; message B {}")
	selected := v1GenerationModule{directory: root, module: v1.Module{Name: "example.test/dep", Roots: []string{"proto"}, ProtoFilters: []v1.ProtoFileFilter{{Root: "proto", Excludes: []string{"proto/excluded"}}}}}
	files, matched, err := selectedPackageFiles(t.Context(), selected, []string{"selected.v1", "excluded.v1"})
	require.NoError(t, err)
	assert.Equal(t, []string{"selected/a.proto"}, files)
	assert.Equal(t, map[string]bool{"selected.v1": true}, matched)
}
