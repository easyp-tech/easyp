package gitmodules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestBufExcludesSnapshotAcrossCacheOperations(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		metadata map[string]string
		root     string
		filter   v1.ProtoFileFilter
	}{
		{
			name:     "v1",
			metadata: map[string]string{"buf.yaml": "version: v1\nbuild:\n  excludes: [proto/preview]\n"},
			root:     "proto",
			filter:   v1.ProtoFileFilter{Root: ".", Excludes: []string{"proto/preview"}},
		},
		{
			name:     "v1beta1",
			metadata: map[string]string{"buf.yaml": "version: v1beta1\nbuild:\n  roots: [proto]\n  excludes: [proto/preview]\n"},
			root:     "proto",
			filter:   v1.ProtoFileFilter{Root: "proto", Excludes: []string{"proto/preview"}},
		},
		{
			name:     "v2",
			metadata: map[string]string{"buf.yaml": "version: v2\nmodules:\n  - path: proto\n    excludes: [proto/preview]\n"},
			root:     "proto",
			filter:   v1.ProtoFileFilter{Root: "proto", Excludes: []string{"proto/preview"}},
		},
		{
			name: "v1 workspace rebases nested excludes",
			metadata: map[string]string{
				"buf.work.yaml": "version: v1\ndirectories: [api]\n",
				"api/buf.yaml":  "version: v1beta1\nbuild:\n  roots: [proto]\n  excludes: [proto/preview]\n",
			},
			root:   "api/proto",
			filter: v1.ProtoFileFilter{Root: "api/proto", Excludes: []string{"api/proto/preview"}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository, files := bufExcludesRepository(t, test.metadata, test.root)
			cache := New(t.TempDir())
			fetched, err := cache.Fetch(t.Context(), repository, "")
			require.NoError(t, err)
			assert.Equal(t, []v1.ProtoFileFilter{test.filter}, fetched.Module.ProtoFilters)
			assert.True(t, fetched.Module.RootsFromMetadata)
			delete(files, test.root+"/preview/schema.proto")
			delete(files, test.root+"/preview/malformed.proto")
			assert.Equal(t, snapshotTestHash(t, files), fetched.Lock.Hash)
			lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}

			for _, operation := range []struct {
				name  string
				cache *Cache
				cold  bool
			}{
				{name: "fetched cache", cache: cache},
				{name: "cold install", cache: New(t.TempDir()), cold: true},
			} {
				t.Run(operation.name, func(t *testing.T) {
					t.Parallel()
					if operation.cold {
						inventory, err := operation.cache.SourceFiles(fetched.Lock)
						require.NoError(t, err)
						assert.Nil(t, inventory)
					}
					require.NoError(t, operation.cache.Install(t.Context(), lock))
					cold := &Cache{root: operation.cache.root}
					installed, module, err := cold.Cached(fetched.Lock)
					require.NoError(t, err)
					assert.Equal(t, fetched.Module, module)
					inventory, err := operation.cache.SourceFiles(fetched.Lock)
					require.NoError(t, err)
					expected := make([]string, 0, len(files))
					for name, content := range files {
						expected = append(expected, name)
						actual, err := os.ReadFile(filepath.Join(installed, filepath.FromSlash(name)))
						require.NoError(t, err)
						assert.Equal(t, content, string(actual))
					}
					assert.ElementsMatch(t, expected, inventory)
					snapshotFiles, err := snapshotV1Files(installed)
					require.NoError(t, err)
					assert.Equal(t, snapshotFiles, inventory)
					require.NoFileExists(t, filepath.Join(installed, test.root, "preview/schema.proto"))
					require.NoFileExists(t, filepath.Join(installed, test.root, "preview/malformed.proto"))

					require.NoError(t, cold.VerifyCached(t.Context(), lock))
					verified, err := cold.SourceFiles(fetched.Lock)
					require.NoError(t, err)
					assert.Equal(t, inventory, verified)
				})
			}
		})
	}
}

func TestBufExcludedFileChangePreservesSnapshotHash(t *testing.T) {
	t.Parallel()
	repository, _ := bufExcludesRepository(t, map[string]string{"buf.yaml": "version: v2\nmodules:\n  - path: proto\n    excludes: [proto/preview]\n"}, "proto")
	cache := New(t.TempDir())
	first, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(repository, "proto/preview/schema.proto"), []byte("changed excluded proto\n"), 0o644))
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "excluded source only")
	second, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	assert.NotEqual(t, first.Lock.Commit, second.Lock.Commit)
	assert.Equal(t, first.Lock.Hash, second.Lock.Hash)
	assert.Equal(t, first.Module, second.Module)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{first.Lock}}))
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{second.Lock}}))
	firstFiles, err := cache.SourceFiles(first.Lock)
	require.NoError(t, err)
	secondFiles, err := cache.SourceFiles(second.Lock)
	require.NoError(t, err)
	assert.Equal(t, firstFiles, secondFiles)
}

func TestSnapshotWithoutProducerMetadataAcrossCacheOperations(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	files := map[string]string{
		"google/api/types.proto": "syntax = \"proto3\"; package google.api; message Types {}\n",
		"README.md":              "A repository without Buf or EasyP metadata.\n",
	}
	commitRootSelectionRepository(t, repository, files, nil)
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"."}, fetched.Module.Roots)
	assert.Empty(t, fetched.Module.ProtoFilters)
	assert.False(t, fetched.Module.RootsFromMetadata)
	assert.Equal(t, snapshotTestHash(t, files), fetched.Lock.Hash)

	cold := New(t.TempDir())
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
	require.NoError(t, cold.Install(t.Context(), lock))
	installed, module, err := (&Cache{root: cold.root}).Cached(fetched.Lock)
	require.NoError(t, err)
	assert.Equal(t, fetched.Module, module)
	actual, err := os.ReadFile(filepath.Join(installed, "google/api/types.proto"))
	require.NoError(t, err)
	assert.Equal(t, files["google/api/types.proto"], string(actual))
	frozen := &Cache{root: cold.root}
	require.NoError(t, frozen.VerifyCached(t.Context(), lock))
	inventory, err := frozen.SourceFiles(fetched.Lock)
	require.NoError(t, err)
	assert.Equal(t, []string{"google/api/types.proto"}, inventory)
}

func bufExcludesRepository(t *testing.T, metadata map[string]string, root string) (string, map[string]string) {
	t.Helper()
	repository := t.TempDir()
	files := map[string]string{
		root + "/schema.proto":            "syntax = \"proto3\"; package example; message Schema {}\n",
		root + "/preview/schema.proto":    "syntax = \"proto3\"; package example; message Schema {}\n",
		root + "/preview/malformed.proto": "this is deliberately not a protobuf file\n",
		root + "/preview/easyp.gen.yaml":  "plugins: []\n",
	}
	for name, content := range metadata {
		files[name] = content
	}
	commitRootSelectionRepository(t, repository, files, nil)
	return repository, files
}
