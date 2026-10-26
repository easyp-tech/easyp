package modules

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestSnapshotSourceIndexPreservesSelection(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		v1.ModuleFile:                      "module example.test/dependency\n",
		"proto/api/model.proto":            "syntax = \"proto3\";\n",
		"proto/api.proto":                  "syntax = \"proto3\";\n",
		"proto/excluded/model.proto":       "syntax = \"proto3\";\n",
		"proto/.hidden.proto":              "syntax = \"proto3\";\n",
		"proto/.hidden/ignored.proto":      "syntax = \"proto3\";\n",
		"proto/easyp_vendor/ignored.proto": "syntax = \"proto3\";\n",
		"proto/native/protobuf.mod":        "module example.test/nested\n",
		"proto/native/ignored.proto":       "syntax = \"proto3\";\n",
		"proto/buf/buf.yaml":               "version: v1\n",
		"proto/buf/ignored.proto":          "syntax = \"proto3\";\n",
		"proto/work/buf.work.yaml":         "version: v1\ndirectories: [nested]\n",
		"proto/work/nested/ignored.proto":  "syntax = \"proto3\";\n",
		"proto/repository/.git":            "gitdir: /unused\n",
		"proto/repository/ignored.proto":   "syntax = \"proto3\";\n",
		"other/shared.proto":               "syntax = \"proto3\";\n",
	}
	tests := []struct {
		name    string
		roots   []string
		filters []v1.ProtoFileFilter
	}{
		{name: "default root", roots: []string{"."}},
		{name: "multiple roots", roots: []string{"proto", "other"}},
		{name: "explicit hidden root", roots: []string{"proto/.hidden"}},
		{name: "explicit nested module root", roots: []string{"proto/buf"}},
		{name: "filtered root", roots: []string{"proto"}, filters: []v1.ProtoFileFilter{{Root: "proto", Includes: []string{"proto/api", "proto/.hidden.proto", "proto/excluded"}, Excludes: []string{"proto/excluded"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			var names []string
			for name, body := range files {
				writeV1GenerateFixture(t, directory, name, body)
				names = append(names, name)
			}
			slices.Sort(names)
			module := v1.Module{Name: "example.test/dependency", Roots: tt.roots, ProtoFilters: tt.filters}
			local, err := ModuleSources(directory, module)
			require.NoError(t, err)
			want, err := local.IndexFileModules()
			require.NoError(t, err)
			indexed, err := ModuleSources(directory, module)
			require.NoError(t, err)
			indexed, err = indexSnapshotSources(directory, indexed, names)
			require.NoError(t, err)
			got, err := indexed.IndexFileModules()
			require.NoError(t, err)
			assert.Equal(t, want, got)
			for index, root := range local {
				var localOrder, indexedOrder []string
				require.NoError(t, root.Walk(func(path string) error { localOrder = append(localOrder, path); return nil }))
				require.NoError(t, indexed[index].Walk(func(path string) error { indexedOrder = append(indexedOrder, path); return nil }))
				assert.Equal(t, localOrder, indexedOrder, "snapshot traversal keeps directory-walk order")
			}
			localAllowed, indexedAllowed := local.FileAllowed(), indexed.FileAllowed()
			for _, root := range local {
				for _, name := range names {
					path := filepath.Join(directory, filepath.FromSlash(name))
					if sourcePathWithin(path, root.Path) && filepath.Ext(path) == ".proto" {
						assert.Equal(t, localAllowed(path), indexedAllowed(path), name)
					}
				}
			}
		})
	}
}

func TestSnapshotSourceIndexReusesVerifiedInventory(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeV1GenerateFixture(t, directory, "api/one.proto", "syntax = \"proto3\";\n")
	module := v1.Module{Name: "example.test/dependency", Roots: []string{"."}}
	roots, err := ModuleSources(directory, module)
	require.NoError(t, err)
	roots, err = indexSnapshotSources(directory, roots, []string{"api/one.proto"})
	require.NoError(t, err)
	// Make traversal impossible after indexing. Looking up ownership must use
	// the inventory already supplied by the verified snapshot boundary.
	require.NoError(t, os.Rename(filepath.Join(directory, "api"), filepath.Join(directory, "relocated")))
	for range 3 {
		owners, err := roots.IndexFileModules()
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"api/one.proto": module.Name}, owners)
	}
}

func TestLocalSourceIndexObservesChanges(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeV1GenerateFixture(t, directory, "one.proto", "syntax = \"proto3\";\n")
	module := v1.Module{Name: "example.test/local", Roots: []string{"."}}
	roots, err := ModuleSources(directory, module)
	require.NoError(t, err)
	before, err := roots.IndexFileModules()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"one.proto": module.Name}, before)
	require.NoError(t, os.Remove(filepath.Join(directory, "one.proto")))
	writeV1GenerateFixture(t, directory, "two.proto", "syntax = \"proto3\";\n")
	after, err := roots.IndexFileModules()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"two.proto": module.Name}, after)
}

func TestSnapshotSourceIndexChecksUnusedCollisions(t *testing.T) {
	t.Parallel()
	var all SourceRoots
	for _, module := range []string{"example.test/a", "example.test/b"} {
		directory := t.TempDir()
		writeV1GenerateFixture(t, directory, "unused.proto", "syntax = \"proto3\";\n")
		roots, err := ModuleSources(directory, v1.Module{Name: module, Roots: []string{"."}})
		require.NoError(t, err)
		roots, err = indexSnapshotSources(directory, roots, []string{"unused.proto"})
		require.NoError(t, err)
		all = append(all, roots...)
	}
	_, err := all.IndexFileModules()
	require.ErrorContains(t, err, "duplicate import path \"unused.proto\"")
}
