package modules_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestMixedImportLayoutsRetainDefaultRoots(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		consumer string
		gitlink  bool
	}{
		{name: "before_consumer_exists"},
		{name: "public_consumer", consumer: `syntax = "proto3"; import "public/service.proto";`},
		{name: "opaque_gitlink", gitlink: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, commit := importRootsRepository(t, mixedImportLayoutFiles(), nil)
			if tt.gitlink {
				importRootsGit(t, repository, "update-index", "--add", "--cacheinfo", "160000,"+commit+",third_party/external")
				importRootsGit(t, repository, "commit", "--quiet", "-m", "opaque submodule")
				commit = importRootsGit(t, repository, "rev-parse", "HEAD")
			}
			root := t.TempDir()
			importRootsWrite(t, root, v1.ModuleFile, "module example.test/consumer\n")
			if tt.consumer != "" {
				importRootsWrite(t, root, "consumer.proto", tt.consumer)
			}
			cache := gitmodules.New(t.TempDir())
			if tt.gitlink {
				// A direct submodule boundary is already valid in the ordinary
				// strict snapshot. Inspection must preserve that source policy.
				_, err := cache.Fetch(t.Context(), repository, commit)
				require.NoError(t, err)
			}

			err := modules.Get(t.Context(), root, v1.Requirement{Module: repository, Version: commit}, cache)

			require.NoError(t, err)
			lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			require.Len(t, lock.Modules, 1)
			assert.Equal(t, commit, lock.Modules[0].Commit)
			assert.Empty(t, lock.Modules[0].Roots)
			assert.NotEmpty(t, lock.Modules[0].Hash)
			_, selected, err := cache.Cached(lock.Modules[0])
			require.NoError(t, err)
			assert.Equal(t, []string{"."}, selected.Roots)
			manifest, rawLock := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)
			// Move HEAD to prove that a cold default-root replay uses the pin.
			importRootsWrite(t, repository, "public/later.proto", `syntax = "proto3";`)
			importRootsGit(t, repository, "add", ".")
			importRootsGit(t, repository, "commit", "--quiet", "-m", "move HEAD")

			sources, err := modules.EnsureFrozenSources(t.Context(), root, gitmodules.New(t.TempDir()))

			require.NoError(t, err)
			names, err := sources.FileModules()
			require.NoError(t, err)
			want := map[string]string{
				"public/service.proto":      repository,
				"public/types.proto":        repository,
				"tools/aux/service.proto":   repository,
				"tools/aux/aux_types.proto": repository,
			}
			assert.Equal(t, want, names)
			assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
			assert.Equal(t, rawLock, importRootsRead(t, root, v1.LockFile))
		})
	}
}

func TestMixedImportLayoutsReachableAuxiliaryFailsWithoutWrites(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		locked bool
	}{
		{name: "get"},
		{name: "tidy_existing_lock", locked: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, commit := importRootsRepository(t, mixedImportLayoutFiles(), nil)
			root := t.TempDir()
			importRootsWrite(t, root, v1.ModuleFile, "module example.test/consumer\n")
			cache := gitmodules.New(t.TempDir())
			if tt.locked {
				require.NoError(t, modules.Get(t.Context(), root, v1.Requirement{Module: repository, Version: commit}, cache))
			} else {
				importRootsWrite(t, root, v1.LockFile, "# unchanged\nversion: 1\nmodules: []\n")
			}
			importRootsWrite(t, root, "consumer.proto", `syntax = "proto3"; import "tools/aux/service.proto";`)
			manifest, lock, consumer := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile), importRootsRead(t, root, "consumer.proto")

			var err error
			if tt.locked {
				err = modules.Tidy(t.Context(), root, cache)
			} else {
				err = modules.Get(t.Context(), root, v1.Requirement{Module: repository, Version: commit}, cache)
			}

			require.ErrorContains(t, err, "cannot resolve imports")
			assert.ErrorContains(t, err, "aux_types.proto")
			assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
			assert.Equal(t, lock, importRootsRead(t, root, v1.LockFile))
			assert.Equal(t, consumer, importRootsRead(t, root, "consumer.proto"))
		})
	}
}

func mixedImportLayoutFiles() map[string]string {
	return map[string]string{
		"public/service.proto":      `syntax = "proto3"; import "public/types.proto";`,
		"public/types.proto":        `syntax = "proto3";`,
		"tools/aux/service.proto":   `syntax = "proto3"; import "aux_types.proto";`,
		"tools/aux/aux_types.proto": `syntax = "proto3";`,
	}
}
