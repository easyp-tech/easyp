package modules_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestUnchangedPinRetainsOmittedDefaultRoots(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		overlay bool
	}{
		{name: "published_tidy"},
		{name: "refreshed_overlay", overlay: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, commit := importRootsRepository(t, map[string]string{"api/a.proto": `syntax = "proto3"; import "b.proto";`, "api/b.proto": `syntax = "proto3";`}, nil)
			root := importRootsConsumer(t, repository, commit, `syntax = "proto3";`)
			cache := gitmodules.New(t.TempDir())
			fetched, err := cache.Fetch(t.Context(), repository, commit)
			require.NoError(t, err)
			writeRootScopeLock(t, root, v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}})
			before, err := modules.EnsureFrozenSources(t.Context(), root, cache)
			require.NoError(t, err)
			beforeNames, err := before.FileModules()
			require.NoError(t, err)
			want := map[string]string{"api/a.proto": repository, "api/b.proto": repository}
			assert.Equal(t, want, beforeNames)
			if tt.overlay {
				local := t.TempDir()
				importRootsWrite(t, local, v1.ModuleFile, "module example.test/local\n")
				manifest := string(importRootsRead(t, root, v1.ModuleFile)) + "require example.test/local\nreplace example.test/local => " + local + "\n"
				importRootsWrite(t, root, v1.ModuleFile, manifest)
				require.NoError(t, modules.Update(t.Context(), root, cache))
				_, module, err := modules.ReadManifest(root)
				require.NoError(t, err)
				graph, err := modules.EnsureEffectiveGraph(t.Context(), root, module, cache, nil)
				require.NoError(t, err)
				names, err := graph.Sources.FileModules()
				require.NoError(t, err)
				assert.Equal(t, want, names)
			} else {
				require.NoError(t, modules.Tidy(t.Context(), root, cache))
				after, err := modules.EnsureFrozenSources(t.Context(), root, gitmodules.New(t.TempDir()))
				require.NoError(t, err)
				names, err := after.FileModules()
				require.NoError(t, err)
				assert.Equal(t, want, names)
			}
			lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Empty(t, lock.Modules[0].Roots)
			assert.Equal(t, fetched.Lock.Hash, lock.Modules[0].Hash)
		})
	}
}

func TestSameCommitAliasRootReplacementVerifiesBothScopes(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		tag  bool
	}{
		{name: "commit"},
		{name: "tag", tag: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, commit := importRootsRepository(t, map[string]string{"sources/svc.proto": `syntax = "proto3";`}, map[string]string{"api": "sources"})
			version := commit
			if tt.tag {
				importRootsGit(t, repository, "tag", "v1.0.0")
				version = "v1.0.0"
			}
			root := importRootsConsumer(t, repository, version, `syntax = "proto3"; import "svc.proto";`)
			cache := gitmodules.New(t.TempDir())
			target := v1.Requirement{Module: repository, Version: version}
			require.NoError(t, modules.GetWithRoots(t.Context(), root, target, cache, []string{"sources"}))
			old, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			next, err := cache.FetchWithRoots(t.Context(), repository, commit, []string{"api"})
			require.NoError(t, err)
			assert.NotEqual(t, old.Modules[0].Hash, next.Lock.Hash)

			err = modules.GetWithRoots(t.Context(), root, target, cache, []string{"api"})

			require.NoError(t, err)
			current, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Equal(t, []string{"api"}, current.Modules[0].Roots)
			assert.Equal(t, next.Lock.Hash, current.Modules[0].Hash)
			assert.Equal(t, version, current.Modules[0].Version)
			assert.Equal(t, commit, current.Modules[0].Commit)
			sources, err := modules.EnsureFrozenSources(t.Context(), root, gitmodules.New(t.TempDir()))
			require.NoError(t, err)
			names, err := sources.FileModules()
			require.NoError(t, err)
			assert.Equal(t, map[string]string{"svc.proto": repository}, names)
		})
	}
}

func TestAliasRootReplacementRejectsCorruptOriginalScope(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, map[string]string{"sources/svc.proto": `syntax = "proto3";`}, map[string]string{"api": "sources"})
	root := importRootsConsumer(t, repository, commit, `syntax = "proto3"; import "svc.proto";`)
	cache := gitmodules.New(t.TempDir())
	target := v1.Requirement{Module: repository, Version: commit}
	require.NoError(t, modules.GetWithRoots(t.Context(), root, target, cache, []string{"sources"}))
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	lock.Modules[0].Hash = "h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="
	writeRootScopeLock(t, root, lock)
	manifest, before := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)
	spy := &importRootsInstallSpy{Cache: cache}

	err = modules.GetWithRoots(t.Context(), root, target, spy, []string{"api"})

	require.ErrorIs(t, err, modules.ErrLockedVersionChanged)
	assert.Zero(t, spy.installs)
	assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
	assert.Equal(t, before, importRootsRead(t, root, v1.LockFile))
}

func writeRootScopeLock(t *testing.T, root string, lock v1.Lock) {
	t.Helper()
	raw, err := yaml.Marshal(lock)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, v1.LockFile), raw, 0o644))
}
