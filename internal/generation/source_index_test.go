package generation

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestGenerationSelectsIndexedDependencyInputs(t *testing.T) {
	t.Parallel()
	root, snapshot := t.TempDir(), t.TempDir()
	const source = "example.test/dependency"
	files := map[string]string{
		v1.ModuleFile:                "module " + source + "\nroots proto\n",
		"proto/api/item.proto":       "syntax = \"proto3\"; package api; import \"shared.proto\"; message Item { shared.Shared value = 1; }",
		"proto/api/unused.proto":     "invalid unselected protobuf",
		"proto/shared.proto":         "syntax = \"proto3\"; package shared; message Shared {}",
		"proto/nested/protobuf.mod":  "module example.test/nested\n",
		"proto/nested/ignored.proto": "invalid nested module protobuf",
	}
	var names []string
	for name, body := range files {
		writeV1GenerateFixture(t, snapshot, name, body)
		names = append(names, name)
	}
	slices.Sort(names)
	writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.test/consumer\nrequire "+source+" v1.0.0\n")
	writeV1GenerateFixture(t, root, v1.GenerateFile, `version: v1
generate:
  modules:
    - module: example.test/dependency
      paths: [proto/api/item.proto]
  managed:
    enabled: true
    override:
      - module: example.test/dependency
        file_option: go_package_prefix
        value: example.test/generated
`)
	entry := v1.LockedModule{Source: source, Version: "v1.0.0", Commit: strings.Repeat("a", 40), Hash: "h1:" + strings.Repeat("A", 43) + "="}
	require.NoError(t, writeTestLock(root, v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}}))
	cache := &indexedGenerationCache{directory: snapshot, module: v1.Module{Name: source, Roots: []string{"proto"}}, files: names}
	output := filepath.Join(root, "selected.pb")
	err := Run(t.Context(), logger.NewNop(), cache, Request{WorkDir: root, WorkspaceRoot: root, DescriptorSetOut: output, IncludeImports: true})
	require.NoError(t, err)
	assert.Equal(t, 1, cache.inventoryReads)
	descriptors := readDescriptorSet(t, output)
	assert.ElementsMatch(t, []string{"api/item.proto", "shared.proto"}, descriptorNames(descriptors))
	for _, file := range descriptors.File {
		assert.Contains(t, file.GetOptions().GetGoPackage(), "example.test/generated")
	}
}

type indexedGenerationCache struct {
	directory      string
	module         v1.Module
	files          []string
	inventoryReads int
}

func (c *indexedGenerationCache) Install(context.Context, v1.Lock) error { return nil }

func (c *indexedGenerationCache) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	if entry.Source != c.module.Name {
		return "", v1.Module{}, fmt.Errorf("unknown cached module %s", entry.Source)
	}
	return c.directory, c.module, nil
}

func (c *indexedGenerationCache) SourceFiles(v1.LockedModule) ([]string, error) {
	c.inventoryReads++
	return slices.Clone(c.files), nil
}
