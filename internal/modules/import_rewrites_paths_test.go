package modules_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestTidyLibraryAcceptsRelativePaths(t *testing.T) {
	// These library cases intentionally change process cwd and stay sequential.
	tests := []struct {
		name          string
		rootArgument  string
		relativeCache bool
		dependency    bool
	}{
		{name: "dot_root_absolute_cache", rootArgument: "."},
		{name: "absolute_root_relative_cache", relativeCache: true, dependency: true},
		{name: "relative_root_relative_cache", rootArgument: "consumer", relativeCache: true, dependency: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := t.TempDir()
			root := filepath.Join(base, "consumer")
			manifest := "module example.test/consumer\n"
			consumer := `syntax = "proto3"; message Consumer {}`
			if tt.dependency {
				repository, _ := importRootsRepository(t, map[string]string{"dep.proto": `syntax = "proto3"; message Dependency {}`}, nil)
				importRootsGit(t, repository, "tag", "v0.1.0")
				manifest += "require " + repository + " v0.1.0\n"
				consumer = `syntax = "proto3"; import "dep.proto"; message Consumer { Dependency dependency = 1; }`
			}
			importRootsWrite(t, root, v1.ModuleFile, manifest)
			importRootsWrite(t, root, "consumer.proto", consumer)
			t.Chdir(base)
			if tt.rootArgument == "." {
				t.Chdir(root)
			}
			rootArgument := tt.rootArgument
			if rootArgument == "" {
				rootArgument = root
			}
			storage := t.TempDir()
			if tt.relativeCache {
				storage = "cache"
			}

			err := modules.Tidy(t.Context(), rootArgument, gitmodules.New(storage))

			require.NoError(t, err)
			assert.Equal(t, consumer, string(importRootsRead(t, root, "consumer.proto")))
			assert.FileExists(t, filepath.Join(root, v1.LockFile))
			require.NoError(t, modules.Tidy(t.Context(), rootArgument, gitmodules.New(storage)))
		})
	}
}

func TestTidyLibraryPreservesRelativeRootAndCacheAliases(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "consumer")
	repository, _ := importRootsRepository(t, map[string]string{
		"api/v1/a.proto":   `syntax = "proto3"; import "svc.proto";`,
		"api/v1/svc.proto": `syntax = "proto3"; package service; message Service {}`,
	}, nil)
	importRootsGit(t, repository, "tag", "v0.4.0")
	consumer := `syntax = "proto3"; import "svc.proto"; message Consumer { service.Service service = 1; }`
	importRootsWrite(t, root, v1.ModuleFile, "module example.test/consumer\nrequire "+repository+" v0.4.0\n")
	importRootsWrite(t, root, "consumer.proto", consumer)
	require.NoError(t, os.Mkdir(filepath.Join(base, "cache"), 0o755))
	require.NoError(t, os.Symlink("consumer", filepath.Join(base, "consumer-link")))
	require.NoError(t, os.Symlink("cache", filepath.Join(base, "cache-link")))
	require.NoError(t, os.Symlink("consumer.proto", filepath.Join(root, "source-alias.proto")))
	t.Chdir(base)
	cache := gitmodules.New("cache-link")
	require.NoError(t, modules.Tidy(t.Context(), "consumer-link", cache))
	tidyRewriteNewNamespace(t, repository, root)

	report, err := modules.TidyWithReport(t.Context(), "consumer-link", cache)

	require.NoError(t, err)
	require.Len(t, report.Imports, 1)
	assert.Equal(t, "consumer.proto", report.Imports[0].File)
	assert.Equal(t, strings.Replace(consumer, `"svc.proto"`, `"v1/svc.proto"`, 1), string(importRootsRead(t, root, "consumer.proto")))
	for name, target := range map[string]string{"consumer-link": "consumer", "cache-link": "cache"} {
		actual, err := os.Readlink(filepath.Join(base, name))
		require.NoError(t, err)
		assert.Equal(t, target, actual)
	}
	actual, err := os.Readlink(filepath.Join(root, "source-alias.proto"))
	require.NoError(t, err)
	assert.Equal(t, "consumer.proto", actual)
}

func TestTidyLibraryChecksOldOwnershipWithRelativeCache(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "consumer")
	repository, _ := importRootsRepository(t, map[string]string{"svc.proto": `syntax = "proto3";`}, nil)
	importRootsGit(t, repository, "tag", "v0.4.0")
	consumer := `syntax = "proto3"; import "local.proto"; message Consumer {}`
	importRootsWrite(t, root, v1.ModuleFile, "module example.test/consumer\nrequire "+repository+" v0.4.0\n")
	importRootsWrite(t, root, "consumer.proto", consumer)
	importRootsWrite(t, root, "local.proto", `syntax = "proto3";`)
	t.Chdir(base)
	cache := tidyBasicRepository{cache: gitmodules.New("cache")}
	require.NoError(t, modules.Tidy(t.Context(), "consumer", cache))
	importRootsWrite(t, root, v1.ModuleFile, "module example.test/consumer\n")

	report, err := modules.TidyWithReport(t.Context(), "consumer", cache)

	require.NoError(t, err)
	assert.Empty(t, report.Imports)
	assert.Equal(t, consumer, string(importRootsRead(t, root, "consumer.proto")))
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	assert.Empty(t, lock.Modules)
}
