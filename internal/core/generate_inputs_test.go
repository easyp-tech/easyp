package core

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestPrepareGenerationUsesVerifiedInventory(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	contents := map[string]string{
		filepath.Join(root, "proto/api/item.proto"):   "syntax = \"proto3\"; package api; import \"shared.proto\"; message Item { shared.Shared value = 1; }",
		filepath.Join(root, "proto/shared.proto"):     "syntax = \"proto3\"; package shared; message Shared {}",
		filepath.Join(root, "proto/api/unused.proto"): "invalid unselected protobuf",
	}
	var unusedChecks atomic.Int64
	walks := 0
	app := New(Options{
		Logger: logger.NewNop(),
		Inputs: Inputs{InputFilesDir: []InputFilesDir{{Root: "proto", Path: "api"}}},
		WalkSourceFiles: func(importRoot string, visit func(string) error) (bool, error) {
			assert.Equal(t, filepath.Join(root, "proto"), importRoot)
			walks++
			for _, name := range []string{"api/item.proto", "api/unused.proto", "shared.proto"} {
				if err := visit(filepath.Join(importRoot, name)); err != nil {
					return true, err
				}
			}
			return true, nil
		},
		ImportFileAllowed: func(path string) bool {
			if filepath.Base(path) == "unused.proto" {
				unusedChecks.Add(1)
			}
			_, ok := contents[filepath.Clean(path)]
			return ok
		},
		OpenSourceFile: func(path string) (io.ReadCloser, error) {
			body, ok := contents[filepath.Clean(path)]
			if !ok {
				return nil, os.ErrNotExist
			}
			return io.NopCloser(strings.NewReader(body)), nil
		},
	})
	// No source directories exist on disk. Both input discovery and imports
	// must use the configured snapshot inventory and source accessor.
	plan, err := app.PrepareGenerationFiles(t.Context(), root, []string{"api/item.proto"})
	require.NoError(t, err)
	assert.Equal(t, 1, walks)
	assert.Zero(t, unusedChecks.Load(), "selection runs before file metadata or membership checks")
	assert.ElementsMatch(t, []string{"api/item.proto", "shared.proto"}, descriptorNames(plan.DescriptorSet(true).File))
	assert.Equal(t, []string{"api/item.proto"}, descriptorNames(plan.DescriptorSet(false).File))
	_, err = app.PrepareGenerationFiles(t.Context(), root, []string{"shared.proto"})
	require.ErrorContains(t, err, "outside the module inputs")
}

func TestPrepareGenerationWithoutInventoryKeepsLocalSourcesFresh(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeGenerationSources(t, root, map[string]string{"first.proto": "syntax = \"proto3\"; message First {}"})
	app := New(Options{
		Logger: logger.NewNop(),
		Inputs: Inputs{InputFilesDir: []InputFilesDir{{Root: ".", Path: "."}}},
		WalkSourceFiles: func(string, func(string) error) (bool, error) {
			return false, nil
		},
	})
	first, err := app.PrepareGeneration(t.Context(), root)
	require.NoError(t, err)
	assert.Equal(t, []string{"first.proto"}, descriptorNames(first.DescriptorSet(false).File))
	require.NoError(t, os.Remove(filepath.Join(root, "first.proto")))
	writeGenerationSources(t, root, map[string]string{"second.proto": "syntax = \"proto3\"; message Second {}"})
	second, err := app.PrepareGeneration(t.Context(), root)
	require.NoError(t, err)
	assert.Equal(t, []string{"second.proto"}, descriptorNames(second.DescriptorSet(false).File))
}
