package generation

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
)

func TestSelectedModuleUsesMainOverlay(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/app\nroots proto\nrequire example.com/A\nreplace example.com/A => a\nreplace example.com/B => b\nreplace example.com/unused => missing\n")
	writeV1GenerateFixture(t, root, "a/protobuf.mod", "module example.com/A\nrequire example.com/B\nreplace example.com/B => missing\n")
	writeV1GenerateFixture(t, root, "b/protobuf.mod", "module example.com/B\n")
	cache := gitmodules.New(t.TempDir())
	selected, err := resolveV1GenerationModule(t.Context(), cache, root, root, v1ModuleSelection{source: "example.com/A"})
	require.NoError(t, err)
	assert.Equal(t, root, selected.resolutionDir)
	assert.Equal(t, []string{filepath.Join(root, "b")}, selected.dependencies.Paths())
	_, err = resolveV1GenerationModule(t.Context(), cache, root, root, v1ModuleSelection{source: "example.com/unused"})
	require.ErrorContains(t, err, "not selected")
}
