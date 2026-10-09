package gitmodules

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

type rootFetchingCache interface {
	FetchWithRoots(context.Context, string, string, []string) (modules.Fetched, error)
	FetchMigrationWithRoots(context.Context, string, string, string, []string) (modules.Fetched, error)
	FetchForRootResolution(context.Context, string, string) (modules.Fetched, error)
}

func requireRootFetchingCache(t *testing.T, cache *Cache) rootFetchingCache {
	t.Helper()
	fetcher, ok := any(cache).(rootFetchingCache)
	require.True(t, ok, "the cache must support explicit and provisional root selection")
	return fetcher
}

func TestFetchWithRootsReplaysColdAndWarmSelections(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"api/svc.proto":   "syntax = \"proto3\"; package api;\n",
		"other/svc.proto": "syntax = \"proto3\"; package other;\n",
		"README.md":       "original repository bytes\n",
	}
	repository, commit := migrationTestRepository(t, files)
	cache := &Cache{root: t.TempDir()}
	fetcher := requireRootFetchingCache(t, cache)
	roots := []string{"api"}
	fetched, err := fetcher.FetchWithRoots(t.Context(), repository, commit, roots)
	require.NoError(t, err)
	assert.False(t, fetched.Module.RootsFromMetadata)
	assert.Equal(t, roots, fetched.Module.Roots)
	assert.Equal(t, roots, fetched.Lock.Roots)
	assert.Equal(t, snapshotTestHash(t, files), fetched.Lock.Hash)
	roots[0] = "changed"
	assert.Equal(t, []string{"api"}, fetched.Lock.Roots)

	other, err := fetcher.FetchWithRoots(t.Context(), repository, commit, []string{"other"})
	require.NoError(t, err)
	assert.Equal(t, fetched.Lock.Hash, other.Lock.Hash)
	assert.NotEqual(t, v1ModuleCachePath(cache.root, fetched.Lock), v1ModuleCachePath(cache.root, other.Lock))
	for _, entry := range []v1.LockedModule{fetched.Lock, other.Lock} {
		lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}}
		require.NoError(t, cache.Install(t.Context(), lock))
		require.NoError(t, cache.Install(t.Context(), lock))
		require.NoError(t, cache.VerifyCached(t.Context(), lock))
		directory, module, err := cache.Cached(entry)
		require.NoError(t, err)
		stamp, err := readCacheVerificationStamp(directory)
		require.NoError(t, err)
		assert.Equal(t, v1RootSelectionKey(entry.Roots), stamp.Roots)
		assert.Equal(t, entry.Roots, module.Roots)
		assert.NoFileExists(t, filepath.Join(directory, v1.ModuleFile))
		module.Roots[0] = "consumer mutation"
		_, again, err := cache.Cached(entry)
		require.NoError(t, err)
		assert.Equal(t, entry.Roots, again.Roots)
		assert.NoFileExists(t, filepath.Join(directory, "README.md"))
	}
}

func TestFetchWithRootsCanonicalizesSelectionOrder(t *testing.T) {
	t.Parallel()
	repository, commit := migrationTestRepository(t, map[string]string{"api/a.proto": "syntax = \"proto3\";", "other/b.proto": "syntax = \"proto3\";"})
	cache := &Cache{root: t.TempDir()}
	fetcher := requireRootFetchingCache(t, cache)
	fetched, err := fetcher.FetchWithRoots(t.Context(), repository, commit, []string{"other", "api"})
	require.NoError(t, err)
	assert.Equal(t, []string{"api", "other"}, fetched.Lock.Roots)
	reversed := fetched.Lock
	reversed.Roots = []string{"other", "api"}
	assert.Equal(t, v1ModuleCachePath(cache.root, fetched.Lock), v1ModuleCachePath(cache.root, reversed))
}

func TestFetchWithRootsSelectsDirectoryAliasBeforeValidation(t *testing.T) {
	t.Parallel()
	repository := rootSelectionRepository(t, map[string]string{"sources/svc.proto": "syntax = \"proto3\";\n"}, map[string]string{"api": "sources", "unused.proto": "../outside.proto"})
	cache := &Cache{root: t.TempDir()}
	fetcher := requireRootFetchingCache(t, cache)
	_, err := cache.Fetch(t.Context(), repository, "")
	require.ErrorIs(t, err, sourceview.ErrOutsideRoot)
	fetched, err := fetcher.FetchWithRoots(t.Context(), repository, "", []string{"api"})
	require.NoError(t, err)
	assert.Equal(t, []string{"api"}, fetched.Module.Roots)
	cold := &Cache{root: t.TempDir()}
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
	require.NoError(t, cold.Install(t.Context(), lock))
	require.NoError(t, cold.VerifyCached(t.Context(), lock))
	directory, module, err := cold.Cached(fetched.Lock)
	require.NoError(t, err)
	assert.Equal(t, []string{"api"}, module.Roots)
	assert.Equal(t, "syntax = \"proto3\";\n", string(mustReadRootSelectionFile(t, directory, "api/svc.proto")))
	assert.NoFileExists(t, filepath.Join(directory, "unused.proto"))

	broader := fetched.Lock
	broader.Roots = []string{"."}
	require.Error(t, cold.VerifyCached(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{broader}}))
	require.ErrorIs(t, cold.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{broader}}), sourceview.ErrOutsideRoot)
}

func TestFetchWithRootsRejectsSelectedPathProblems(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, alias, target, root string
		want                      error
	}{
		{name: "selected proto escape", alias: "api/bad.proto", target: "../../outside.proto", root: "api", want: sourceview.ErrOutsideRoot},
		{name: "selected proto cycle", alias: "api/bad.proto", target: "bad.proto", root: "api", want: sourceview.ErrCycle},
		{name: "root escape", alias: "api", target: "../outside", root: "api", want: sourceview.ErrOutsideRoot},
		{name: "root cycle", alias: "api", target: "api", root: "api", want: sourceview.ErrCycle},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := rootSelectionRepository(t, map[string]string{"sources/svc.proto": "syntax = \"proto3\";\n"}, map[string]string{tt.alias: tt.target})
			fetcher := requireRootFetchingCache(t, &Cache{root: t.TempDir()})
			_, err := fetcher.FetchWithRoots(t.Context(), repository, "", []string{tt.root})
			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestFetchWithRootsRespectsAuthoritativeMetadata(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, file, content string
	}{
		{name: "native implicit default", file: "protobuf.mod", content: "module {{source}}\n"},
		{name: "native explicit default", file: "protobuf.mod", content: "module {{source}}\nroots .\n"},
		{name: "Buf default", file: "buf.yaml", content: "version: v1\n"},
		{name: "EasyP default", file: "easyp.yaml", content: "generate:\n  inputs: [{directory: api}]\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			files := map[string]string{tt.file: strings.ReplaceAll(tt.content, "{{source}}", repository), "api/svc.proto": "syntax = \"proto3\";\n"}
			commitRootSelectionRepository(t, repository, files, nil)
			cache := &Cache{root: t.TempDir()}
			fetcher := requireRootFetchingCache(t, cache)
			_, err := fetcher.FetchWithRoots(t.Context(), repository, "", []string{"api"})
			require.ErrorContains(t, err, "authoritative")
			fetched, err := fetcher.FetchWithRoots(t.Context(), repository, "", []string{"."})
			require.NoError(t, err)
			assert.True(t, fetched.Module.RootsFromMetadata)
			assert.Empty(t, fetched.Lock.Roots)
			require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
			directory, module, err := cache.Cached(fetched.Lock)
			require.NoError(t, err)
			assert.True(t, module.RootsFromMetadata)
			assert.Equal(t, []string{"."}, module.Roots)
			assert.Equal(t, files[tt.file], string(mustReadRootSelectionFile(t, directory, tt.file)))
		})
	}
}

func TestFetchWithRootsDoesNotHideMalformedMetadata(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, file, content string }{
		{name: "malformed Buf", file: "buf.yaml", content: "version: v2\nmodules: wrong\n"},
		{name: "malformed native", file: "protobuf.mod", content: "module {{source}}\nroots (\n"},
		{name: "malformed EasyP", file: "easyp.yaml", content: "generate: [bad]\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			commitRootSelectionRepository(t, repository, map[string]string{tt.file: strings.ReplaceAll(tt.content, "{{source}}", repository), "api/svc.proto": "syntax = \"proto3\";"}, nil)
			fetcher := requireRootFetchingCache(t, &Cache{root: t.TempDir()})
			_, err := fetcher.FetchWithRoots(t.Context(), repository, "", []string{"api"})
			require.Error(t, err)
			_, err = fetcher.FetchForRootResolution(t.Context(), repository, "")
			require.Error(t, err)
		})
	}
}

func TestFetchForRootResolutionRetainsPinnedInspection(t *testing.T) {
	t.Parallel()
	content := "syntax = \"proto3\"; package api; message Service {}\n"
	repository := rootSelectionRepository(t, map[string]string{"api/svc.proto": content}, map[string]string{"bad.proto": "../outside.proto", "cycle": "cycle"})
	cache := &Cache{root: t.TempDir()}
	fetcher := requireRootFetchingCache(t, cache)
	fetched, err := fetcher.FetchForRootResolution(t.Context(), repository, "")
	require.NoError(t, err)
	assert.False(t, fetched.Module.RootsFromMetadata)
	assert.True(t, v1.IsCommitRef(fetched.Lock.Commit))
	assert.Empty(t, fetched.Lock.Hash)
	require.NotNil(t, fetched.Inspection)
	assert.True(t, fetched.Inspection.Provisional)
	require.Len(t, fetched.Inspection.Files, 1)
	assert.Equal(t, "api/svc.proto", fetched.Inspection.Files[0].Path)
	assert.NotEmpty(t, fetched.Inspection.Files[0].Identity)
	assert.Equal(t, []byte(content), fetched.Inspection.Files[0].Content)
	require.Len(t, fetched.Inspection.Problems, 2)
	assert.Equal(t, "bad.proto", fetched.Inspection.Problems[0].Path)
	assert.ErrorIs(t, fetched.Inspection.Problems[0].Err, sourceview.ErrOutsideRoot)
	assert.Equal(t, "cycle", fetched.Inspection.Problems[1].Path)
	assert.ErrorIs(t, fetched.Inspection.Problems[1].Err, sourceview.ErrCycle)
	require.Error(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
	entries, err := os.ReadDir(cache.root)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.False(t, strings.HasPrefix(entry.Name(), "git-") || strings.HasPrefix(entry.Name(), "snapshot-"))
	}
	final, err := fetcher.FetchWithRoots(t.Context(), repository, fetched.Lock.Commit, []string{"api"})
	require.NoError(t, err)
	require.NoError(t, (v1.Lock{Version: 1, Modules: []v1.LockedModule{final.Lock}}).Validate())
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{final.Lock}}))
}

func TestFetchForRootResolutionExcludesTraversalBoundaries(t *testing.T) {
	t.Parallel()
	content := "syntax = \"proto3\"; message Service {}\n"
	for _, tt := range []struct {
		name  string
		files map[string]string
	}{
		{name: "leading hidden directory", files: map[string]string{".hidden/svc.proto": content}},
		{name: "nested hidden directory", files: map[string]string{"api/.hidden/svc.proto": content}},
		{name: "vendor directory", files: map[string]string{"easyp_vendor/svc.proto": content}},
		{name: "nested module", files: map[string]string{"protobuf.mod": "direct (\n)\n", "api/child/protobuf.mod": "module example.test/nested\n", "api/child/svc.proto": content}},
		{name: "nested Buf module", files: map[string]string{"child/buf.yaml": "version: v1\n", "child/svc.proto": content}},
		{name: "nested Buf workspace", files: map[string]string{"child/buf.work.yaml": "version: v1\ndirectories: [.]\n", "child/svc.proto": content}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.files["api/svc.proto"] = content
			repository := rootSelectionRepository(t, tt.files, nil)
			fetcher := requireRootFetchingCache(t, &Cache{root: t.TempDir()})
			fetched, err := fetcher.FetchForRootResolution(t.Context(), repository, "")
			require.NoError(t, err)
			require.NotNil(t, fetched.Inspection)
			assert.True(t, fetched.Inspection.Provisional)
			require.Len(t, fetched.Inspection.Files, 1)
			assert.Equal(t, "api/svc.proto", fetched.Inspection.Files[0].Path)
		})
	}
}

func TestFetchForRootResolutionRetainsExplicitHiddenRoot(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	content := "syntax = \"proto3\"; message Service {}\n"
	commitRootSelectionRepository(t, repository, map[string]string{"protobuf.mod": "module " + repository + "\nroots .hidden\n", ".hidden/svc.proto": content, ".hidden/api/svc.proto": content, ".hidden/api/v1/svc.proto": content, ".hidden/.child/svc.proto": content, ".hidden/easyp_vendor/svc.proto": content}, nil)
	fetcher := requireRootFetchingCache(t, &Cache{root: t.TempDir()})
	fetched, err := fetcher.FetchForRootResolution(t.Context(), repository, "")
	require.NoError(t, err)
	require.NotNil(t, fetched.Inspection)
	assert.False(t, fetched.Inspection.Provisional)
	var actual []string
	for _, file := range fetched.Inspection.Files {
		actual = append(actual, file.Path)
	}
	assert.Equal(t, []string{".hidden/api/svc.proto", ".hidden/api/v1/svc.proto", ".hidden/svc.proto"}, actual)
	var selected []string
	require.NoError(t, modules.WalkProtoFiles(filepath.Join(repository, ".hidden"), func(filename string) error {
		relative, err := filepath.Rel(repository, filename)
		require.NoError(t, err)
		selected = append(selected, filepath.ToSlash(relative))
		return nil
	}))
	assert.Equal(t, selected, actual)
}

func TestFetchForRootResolutionRetainsAuthoritativeAliases(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	content := "syntax = \"proto3\"; package api;\n"
	commitRootSelectionRepository(t, repository, map[string]string{"protobuf.mod": "module " + repository + "\nroots api\n", "sources/svc.proto": content}, map[string]string{"api": "sources", "unused.proto": "../outside.proto"})
	fetcher := requireRootFetchingCache(t, &Cache{root: t.TempDir()})
	fetched, err := fetcher.FetchForRootResolution(t.Context(), repository, "")
	require.NoError(t, err)
	require.NotNil(t, fetched.Inspection)
	assert.False(t, fetched.Inspection.Provisional)
	require.NoError(t, (v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}).Validate())
	files := make(map[string]modules.RootProtoFile)
	for _, file := range fetched.Inspection.Files {
		files[file.Path] = file
	}
	require.Contains(t, files, "api/svc.proto")
	require.Contains(t, files, "sources/svc.proto")
	assert.Equal(t, files["api/svc.proto"].Identity, files["sources/svc.proto"].Identity)
	aliased := files["api/svc.proto"]
	aliased.Content[0] = '!'
	assert.Equal(t, []byte(content), files["sources/svc.proto"].Content)
}

func TestFetchMigrationWithRootsPreservesLegacyHash(t *testing.T) {
	t.Parallel()
	files := map[string]string{"api/svc.proto": "syntax = \"proto3\";\n"}
	repository, commit := migrationTestRepository(t, files)
	fetcher := requireRootFetchingCache(t, &Cache{root: t.TempDir()})
	legacyHash := migrationTestHash(t, files)
	fetched, err := fetcher.FetchMigrationWithRoots(t.Context(), repository, commit, legacyHash, []string{"."})
	require.NoError(t, err)
	assert.Equal(t, legacyHash, fetched.Lock.Hash)
	assert.Equal(t, []string{"."}, fetched.Lock.Roots)
	_, err = fetcher.FetchMigrationWithRoots(t.Context(), repository, commit, "h1:"+strings.Repeat("A", 43)+"=", []string{"."})
	require.ErrorContains(t, err, "legacy hash mismatch")
}

func TestRootHintsAreValidatedBeforeDependencyAccess(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		roots []string
	}{
		{name: "escape", roots: []string{"../api"}},
		{name: "noncanonical", roots: []string{"api/../other"}},
		{name: "absolute", roots: []string{"/api"}},
		{name: "backslash", roots: []string{`api\v1`}},
		{name: "glob", roots: []string{"api/*"}},
		{name: "duplicate", roots: []string{"api", "api"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := &Cache{root: t.TempDir()}
			fetcher := requireRootFetchingCache(t, cache)
			_, err := fetcher.FetchWithRoots(t.Context(), "invalid", "", tt.roots)
			require.ErrorContains(t, err, "ValidateModuleRoots")
			_, err = fetcher.FetchMigrationWithRoots(t.Context(), "invalid", "", "", tt.roots)
			require.ErrorContains(t, err, "ValidateModuleRoots")
		})
	}
}

func TestFetchForRootResolutionKeepsAuthoritativeScopesStrict(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	commitRootSelectionRepository(t, repository, map[string]string{"protobuf.mod": "module " + repository + "\n", "api/svc.proto": "syntax = \"proto3\";"}, map[string]string{"bad.proto": "../outside.proto"})
	fetcher := requireRootFetchingCache(t, &Cache{root: t.TempDir()})
	_, err := fetcher.FetchForRootResolution(t.Context(), repository, "")
	require.ErrorIs(t, err, sourceview.ErrOutsideRoot)
}

func rootSelectionRepository(t *testing.T, files, aliases map[string]string) string {
	t.Helper()
	repository := t.TempDir()
	commitRootSelectionRepository(t, repository, files, aliases)
	return repository
}

func commitRootSelectionRepository(t *testing.T, repository string, files, aliases map[string]string) {
	t.Helper()
	for name, data := range files {
		filename := filepath.Join(repository, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
		require.NoError(t, os.WriteFile(filename, []byte(data), 0o644))
	}
	for name, target := range aliases {
		filename := filepath.Join(repository, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
		require.NoError(t, os.Symlink(target, filename))
	}
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
}

func mustReadRootSelectionFile(t *testing.T, directory, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(name)))
	require.NoError(t, err)
	return content
}
