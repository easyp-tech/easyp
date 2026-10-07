package modules_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

func TestTidyRewritesVerifiedPinnedImportLiterals(t *testing.T) {
	t.Parallel()
	service := `syntax = "proto3"; package service.v1; message Service { string id = 1; }`
	repository, _ := importRootsRepository(t, map[string]string{
		"api/service/v1/a.proto":   `syntax = "proto3"; import "svc.proto";`,
		"api/service/v1/svc.proto": service,
	}, nil)
	importRootsGit(t, repository, "tag", "v0.4.0")
	consumer := "syntax = \"proto3\";\r\npackage consumer;\r\nimport /* before */ public \"s\\x76c.\" /* between */ \"proto\"; // after\r\noption java_package = \"svc.proto\";\r\nmessage Consumer { service.v1.Service service = 1; }\r\n"
	root := importRootsConsumer(t, repository, "v0.4.0", consumer)
	require.NoError(t, os.Chmod(filepath.Join(root, "consumer.proto"), 0o640))
	cache := gitmodules.New(t.TempDir())
	require.NoError(t, modules.Tidy(t.Context(), root, cache))
	old, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	require.Equal(t, []string{"api/service/v1"}, old.Modules[0].Roots)
	oldDirectory, _, err := cache.Cached(old.Modules[0])
	require.NoError(t, err)
	oldSource := importRootsRead(t, oldDirectory, "api/service/v1/svc.proto")
	importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\nroots api\n")
	importRootsWrite(t, repository, "api/service/v1/a.proto", `syntax = "proto3"; import "service/v1/svc.proto";`)
	importRootsGit(t, repository, "add", ".")
	importRootsGit(t, repository, "commit", "--quiet", "-m", "declare namespace")
	importRootsGit(t, repository, "tag", "v0.5.0")
	newCommit := importRootsGit(t, repository, "rev-parse", "HEAD")
	tidyRewriteRequirement(t, root, "v0.4.0", "v0.5.0")
	manifest := importRootsRead(t, root, v1.ModuleFile)

	err = modules.Tidy(t.Context(), root, cache)

	require.NoError(t, err)
	expected := strings.Replace(consumer, `"s\x76c." /* between */ "proto"`, `"service/v1/svc.proto" /* between */ ""`, 1)
	assert.Equal(t, expected, string(importRootsRead(t, root, "consumer.proto")))
	info, err := os.Stat(filepath.Join(root, "consumer.proto"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
	assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	require.Len(t, lock.Modules, 1)
	assert.Equal(t, "v0.5.0", lock.Modules[0].Version)
	assert.Equal(t, newCommit, lock.Modules[0].Commit)
	assert.Empty(t, lock.Modules[0].Roots)
	assert.Equal(t, service, string(importRootsRead(t, repository, "api/service/v1/svc.proto")))
	assert.Equal(t, oldSource, importRootsRead(t, oldDirectory, "api/service/v1/svc.proto"))
	before := importRootsRead(t, root, v1.LockFile)
	require.NoError(t, modules.Tidy(t.Context(), root, gitmodules.New(t.TempDir())))
	assert.Equal(t, before, importRootsRead(t, root, v1.LockFile))
	assert.Equal(t, expected, string(importRootsRead(t, root, "consumer.proto")))
}

func TestTidyRepairsOmittedDefaultScope(t *testing.T) {
	t.Parallel()
	repository, _ := importRootsRepository(t, map[string]string{"api/svc.proto": `syntax = "proto3"; package service; message Service {}`}, nil)
	importRootsGit(t, repository, "tag", "v0.4.0")
	consumer := `syntax = "proto3"; import "api/svc.proto"; message Consumer { service.Service service = 1; }`
	root := importRootsConsumer(t, repository, "v0.4.0", consumer)
	cache := gitmodules.New(t.TempDir())
	require.NoError(t, modules.Tidy(t.Context(), root, cache))
	old, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	require.Empty(t, old.Modules[0].Roots)
	importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\nroots api\n")
	importRootsGit(t, repository, "add", ".")
	importRootsGit(t, repository, "commit", "--quiet", "-m", "declare namespace")
	importRootsGit(t, repository, "tag", "v0.5.0")
	tidyRewriteRequirement(t, root, "v0.4.0", "v0.5.0")

	err = modules.Tidy(t.Context(), root, cache)

	require.NoError(t, err)
	assert.Equal(t, strings.Replace(consumer, `"api/svc.proto"`, `"svc.proto"`, 1), string(importRootsRead(t, root, "consumer.proto")))
}

func TestTidyPreservesSurvivingOldPhysicalBinding(t *testing.T) {
	t.Parallel()
	repository, _ := importRootsRepository(t, map[string]string{
		"api/v1/a.proto":   `syntax = "proto3"; import "svc.proto";`,
		"api/v1/svc.proto": `syntax = "proto3"; package original; message Service {}`,
	}, nil)
	importRootsGit(t, repository, "tag", "v0.4.0")
	consumer := `syntax = "proto3"; import "svc.proto"; message Consumer { original.Service service = 1; }`
	root := importRootsConsumer(t, repository, "v0.4.0", consumer)
	cache := gitmodules.New(t.TempDir())
	require.NoError(t, modules.Tidy(t.Context(), root, cache))
	importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\nroots api\n")
	importRootsWrite(t, repository, "api/v1/a.proto", `syntax = "proto3"; import "v1/svc.proto";`)
	importRootsWrite(t, repository, "api/svc.proto", `syntax = "proto3"; package replacement; message Service {}`)
	importRootsGit(t, repository, "add", ".")
	importRootsGit(t, repository, "commit", "--quiet", "-m", "reuse old public name")
	importRootsGit(t, repository, "tag", "v0.5.0")
	tidyRewriteRequirement(t, root, "v0.4.0", "v0.5.0")

	err := modules.Tidy(t.Context(), root, cache)

	require.NoError(t, err)
	assert.Equal(t, strings.Replace(consumer, `"svc.proto"`, `"v1/svc.proto"`, 1), string(importRootsRead(t, root, "consumer.proto")))
}

func tidyRewriteRequirement(t *testing.T, root, oldVersion, newVersion string) {
	t.Helper()
	manifest := string(importRootsRead(t, root, v1.ModuleFile))
	importRootsWrite(t, root, v1.ModuleFile, strings.ReplaceAll(manifest, oldVersion, newVersion))
}

func TestTidyRewriteFailuresLeaveAllProjectBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		newService string
		remove     bool
		alias      bool
		wrongHash  bool
		retag      bool
		wantError  string
	}{
		{name: "missing_needed_source", remove: true, wantError: "no longer exported"},
		{name: "ambiguous_source_names", alias: true, wantError: "ambiguous current names"},
		{name: "changed_symbol_contract", newService: `syntax = "proto3"; package service; message Different {}`, wantError: "unknown type service.Service"},
		{name: "bad_old_hash", wrongHash: true, wantError: "locked version changed"},
		{name: "retagged_current_version", retag: true, wantError: "locked version changed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, root, cache, consumer := tidyRewriteFixture(t)
			importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\nroots api\n")
			importRootsWrite(t, repository, "api/v1/a.proto", `syntax = "proto3";`)
			if tt.remove {
				require.NoError(t, os.Remove(filepath.Join(repository, "api/v1/svc.proto")))
			}
			if tt.newService != "" {
				importRootsWrite(t, repository, "api/v1/svc.proto", tt.newService)
			}
			if tt.alias {
				require.NoError(t, os.Symlink("v1/svc.proto", filepath.Join(repository, "api/alias.proto")))
			}
			importRootsGit(t, repository, "add", ".")
			importRootsGit(t, repository, "commit", "--quiet", "-m", "new contract")
			if tt.retag {
				importRootsGit(t, repository, "tag", "-f", "v0.4.0")
			} else {
				importRootsGit(t, repository, "tag", "v0.5.0")
				tidyRewriteRequirement(t, root, "v0.4.0", "v0.5.0")
			}
			if tt.wrongHash {
				lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
				require.NoError(t, err)
				before := string(importRootsRead(t, root, v1.LockFile))
				importRootsWrite(t, root, v1.LockFile, strings.ReplaceAll(before, lock.Modules[0].Hash, "h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="))
			}
			manifest, lock := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)

			report, err := modules.TidyWithReport(t.Context(), root, cache)

			require.ErrorContains(t, err, tt.wantError)
			if tt.wrongHash || tt.retag {
				assert.ErrorIs(t, err, modules.ErrLockedVersionChanged)
			}
			assert.Empty(t, report.Imports)
			assert.Equal(t, consumer, string(importRootsRead(t, root, "consumer.proto")))
			assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
			assert.Equal(t, lock, importRootsRead(t, root, v1.LockFile))
		})
	}
}

func TestTidyRewritesOwnedBuildCopiesOnceThroughAliases(t *testing.T) {
	t.Parallel()
	repository, root, cache, consumer := tidyRewriteFixture(t)
	importRootsWrite(t, root, "build/copy.proto", consumer)
	require.NoError(t, os.Symlink("consumer.proto", filepath.Join(root, "alias.proto")))
	require.NoError(t, os.Symlink("build", filepath.Join(root, "build-alias")))
	for _, directory := range []string{".hidden", modules.VendorDir, "nested", "workspace/module"} {
		importRootsWrite(t, root, directory+"/copy.proto", consumer)
	}
	importRootsWrite(t, root, "nested/protobuf.mod", "module example.test/nested\n")
	importRootsWrite(t, root, "workspace/module/protobuf.mod", "module example.test/workspace\n")
	outside := t.TempDir()
	importRootsWrite(t, outside, "copy.proto", consumer)
	tidyRewriteNewNamespace(t, repository, root)

	report, err := modules.TidyWithReport(t.Context(), root, cache)

	require.NoError(t, err)
	require.Len(t, report.Imports, 2)
	assert.Equal(t, []string{"build/copy.proto", "consumer.proto"}, []string{report.Imports[0].File, report.Imports[1].File})
	for _, name := range []string{"consumer.proto", "alias.proto", "build/copy.proto", "build-alias/copy.proto"} {
		assert.Equal(t, strings.Replace(consumer, `"svc.proto"`, `"v1/svc.proto"`, 1), string(importRootsRead(t, root, name)))
	}
	for _, name := range []string{"alias.proto", "build-alias"} {
		info, err := os.Lstat(filepath.Join(root, name))
		require.NoError(t, err)
		assert.NotZero(t, info.Mode()&os.ModeSymlink)
	}
	for _, directory := range []string{".hidden", modules.VendorDir, "nested", "workspace/module"} {
		assert.Equal(t, consumer, string(importRootsRead(t, root, directory+"/copy.proto")))
	}
	assert.Equal(t, consumer, string(importRootsRead(t, outside, "copy.proto")))
}

func TestTidyExcludesCustomCacheStorageAndAliasesWithinConsumerRoots(t *testing.T) {
	t.Parallel()
	repository, _ := importRootsRepository(t, map[string]string{
		"api/v1/a.proto":   `syntax = "proto3"; import "svc.proto";`,
		"api/v1/svc.proto": `syntax = "proto3"; package service; message Service {}`,
	}, nil)
	importRootsGit(t, repository, "tag", "v0.4.0")
	consumer := `syntax = "proto3"; import "svc.proto"; message Consumer { service.Service service = 1; }`
	root := importRootsConsumer(t, repository, "v0.4.0", consumer)
	storage := filepath.Join(root, "build/storage")
	importRootsWrite(t, root, "build/storage/owned-copy.proto", consumer)
	cache := gitmodules.New(storage)
	require.NoError(t, modules.Tidy(t.Context(), root, cache))
	old, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	oldDirectory, _, err := cache.Cached(old.Modules[0])
	require.NoError(t, err)
	oldSource := importRootsRead(t, oldDirectory, "api/v1/a.proto")
	unused, commit := importRootsRepository(t, map[string]string{"unused.proto": `syntax = "proto3"; message Broken {`}, nil)
	fetched, err := cache.Fetch(t.Context(), unused, commit)
	require.NoError(t, err)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
	unusedDirectory, _, err := cache.Cached(fetched.Lock)
	require.NoError(t, err)
	unusedSource := importRootsRead(t, unusedDirectory, "unused.proto")
	aliasTarget, err := filepath.Rel(root, filepath.Join(oldDirectory, "api/v1/a.proto"))
	require.NoError(t, err)
	require.NoError(t, os.Symlink(aliasTarget, filepath.Join(root, "cache-alias.proto")))
	require.NoError(t, os.Symlink("build/storage/v1/git", filepath.Join(root, "cache-directory-alias")))
	tidyRewriteNewNamespace(t, repository, root)

	report, err := modules.TidyWithReport(t.Context(), root, cache)

	require.NoError(t, err)
	require.Len(t, report.Imports, 2)
	assert.Equal(t, []string{"build/storage/owned-copy.proto", "consumer.proto"}, []string{report.Imports[0].File, report.Imports[1].File})
	assert.Equal(t, oldSource, importRootsRead(t, oldDirectory, "api/v1/a.proto"))
	assert.Equal(t, unusedSource, importRootsRead(t, unusedDirectory, "unused.proto"))
	alias, err := os.Readlink(filepath.Join(root, "cache-alias.proto"))
	require.NoError(t, err)
	assert.Equal(t, aliasTarget, alias)
	assert.Equal(t, oldSource, importRootsRead(t, root, "cache-alias.proto"))
	directoryAlias, err := os.Readlink(filepath.Join(root, "cache-directory-alias"))
	require.NoError(t, err)
	assert.Equal(t, "build/storage/v1/git", directoryAlias)
	for _, name := range []string{"consumer.proto", "build/storage/owned-copy.proto"} {
		assert.Equal(t, strings.Replace(consumer, `"svc.proto"`, `"v1/svc.proto"`, 1), string(importRootsRead(t, root, name)))
	}
}

func TestTidyExcludesSnapshotsForRepositoriesWithoutCacheOwnershipPort(t *testing.T) {
	t.Parallel()
	repository, _ := importRootsRepository(t, map[string]string{
		"api/v1/a.proto":   `syntax = "proto3"; import "svc.proto";`,
		"api/v1/svc.proto": `syntax = "proto3"; package service; message Service {}`,
	}, nil)
	importRootsGit(t, repository, "tag", "v0.4.0")
	consumer := `syntax = "proto3"; import "svc.proto"; message Consumer { service.Service service = 1; }`
	root := importRootsConsumer(t, repository, "v0.4.0", consumer)
	cache := gitmodules.New(filepath.Join(root, "storage"))
	wrapper := tidySnapshotRepository{cache: cache}
	require.NoError(t, modules.Tidy(t.Context(), root, wrapper))
	old, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	directory, _, err := cache.Cached(old.Modules[0])
	require.NoError(t, err)
	cached := importRootsRead(t, directory, "api/v1/a.proto")
	alias, err := filepath.Rel(root, filepath.Join(directory, "api/v1/a.proto"))
	require.NoError(t, err)
	require.NoError(t, os.Symlink(alias, filepath.Join(root, "cache-alias.proto")))
	tidyRewriteNewNamespace(t, repository, root)

	report, err := modules.TidyWithReport(t.Context(), root, wrapper)

	require.NoError(t, err)
	require.Len(t, report.Imports, 1)
	assert.Equal(t, "consumer.proto", report.Imports[0].File)
	assert.Equal(t, cached, importRootsRead(t, directory, "api/v1/a.proto"))
	assert.Equal(t, cached, importRootsRead(t, root, "cache-alias.proto"))
}

type tidySnapshotRepository struct {
	cache *gitmodules.Cache
}

func (repository tidySnapshotRepository) Fetch(ctx context.Context, source, version string) (modules.Fetched, error) {
	return repository.cache.Fetch(ctx, source, version)
}

func (repository tidySnapshotRepository) FetchForRootResolution(ctx context.Context, source, version string) (modules.Fetched, error) {
	return repository.cache.FetchForRootResolution(ctx, source, version)
}

func (repository tidySnapshotRepository) FetchWithRoots(ctx context.Context, source, version string, roots []string) (modules.Fetched, error) {
	return repository.cache.FetchWithRoots(ctx, source, version, roots)
}

func (repository tidySnapshotRepository) Install(ctx context.Context, lock v1.Lock) error {
	return repository.cache.Install(ctx, lock)
}

func (repository tidySnapshotRepository) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	return repository.cache.Cached(entry)
}

func TestTidyRejectsConsumerModulesInsideOwnedCache(t *testing.T) {
	t.Parallel()
	storage := t.TempDir()
	root := filepath.Join(storage, "v1/git/consumer")
	manifest := "module example.test/cached\n"
	importRootsWrite(t, root, v1.ModuleFile, manifest)
	importRootsWrite(t, root, "consumer.proto", `syntax = "proto3";`)

	_, err := modules.TidyWithReport(t.Context(), root, gitmodules.New(storage))

	require.ErrorContains(t, err, "inside repository-owned cache storage")
	assert.Equal(t, manifest, string(importRootsRead(t, root, v1.ModuleFile)))
	assert.NoFileExists(t, filepath.Join(root, v1.LockFile))
}

func TestTidyPreservesPublicNamesWhenPhysicalSourcesMove(t *testing.T) {
	t.Parallel()
	repository, root, cache, consumer := tidyRewriteFixture(t)
	importRootsGit(t, repository, "mv", "api/v1", "schema")
	importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\nroots schema\n")
	importRootsGit(t, repository, "add", ".")
	importRootsGit(t, repository, "commit", "--quiet", "-m", "move public contract")
	importRootsGit(t, repository, "tag", "v0.5.0")
	tidyRewriteRequirement(t, root, "v0.4.0", "v0.5.0")

	report, err := modules.TidyWithReport(t.Context(), root, cache)

	require.NoError(t, err)
	assert.Empty(t, report.Imports)
	assert.Equal(t, consumer, string(importRootsRead(t, root, "consumer.proto")))
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	assert.Equal(t, "v0.5.0", lock.Modules[0].Version)
}

func TestTidyIgnoresUnusedRemovedAndMalformedDependencyContracts(t *testing.T) {
	t.Parallel()
	repository, root, cache, consumer := tidyRewriteFixture(t)
	require.NoError(t, os.Remove(filepath.Join(repository, "api/v1/a.proto")))
	importRootsWrite(t, repository, "api/broken.proto", `syntax = "proto3"; message Broken {`)
	tidyRewriteNewNamespace(t, repository, root)

	report, err := modules.TidyWithReport(t.Context(), root, cache)

	require.NoError(t, err)
	require.Len(t, report.Imports, 1)
	assert.Equal(t, strings.Replace(consumer, `"svc.proto"`, `"v1/svc.proto"`, 1), string(importRootsRead(t, root, "consumer.proto")))
}

func TestTidyDoesNotRebindRemovedProducerToNewDependency(t *testing.T) {
	t.Parallel()
	_, root, cache, consumer := tidyRewriteFixture(t)
	replacement, _ := importRootsRepository(t, map[string]string{"svc.proto": `syntax = "proto3"; package service; message Service {}`}, nil)
	importRootsGit(t, replacement, "tag", "v0.5.0")
	importRootsWrite(t, root, v1.ModuleFile, "module example.test/consumer\nrequire "+replacement+" v0.5.0\n")
	manifest, lock := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)

	_, err := modules.TidyWithReport(t.Context(), root, cache)

	require.ErrorContains(t, err, "no longer exported by the same producer")
	assert.Equal(t, consumer, string(importRootsRead(t, root, "consumer.proto")))
	assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
	assert.Equal(t, lock, importRootsRead(t, root, v1.LockFile))
}

func TestTidyPromotesRewrittenTransitiveImportsWithComments(t *testing.T) {
	t.Parallel()
	transitive, _ := importRootsRepository(t, map[string]string{
		"api/v1/a.proto":   `syntax = "proto3"; import "svc.proto";`,
		"api/v1/svc.proto": `syntax = "proto3"; package service; message Service {}`,
	}, nil)
	importRootsGit(t, transitive, "tag", "v0.4.0")
	parent, _ := importRootsRepository(t, map[string]string{"facade.proto": `syntax = "proto3";`}, nil)
	importRootsWrite(t, parent, v1.ModuleFile, "module "+parent+"\nrequire "+transitive+" v0.4.0\n")
	importRootsGit(t, parent, "add", ".")
	importRootsGit(t, parent, "commit", "--quiet", "-m", "require transitive")
	importRootsGit(t, parent, "tag", "v0.1.0")
	root := importRootsConsumer(t, parent, "v0.1.0", `syntax = "proto3"; import "facade.proto";`)
	cache := gitmodules.New(t.TempDir())
	require.NoError(t, modules.Tidy(t.Context(), root, cache))
	manifest := string(importRootsRead(t, root, v1.ModuleFile))
	manifest = strings.ReplaceAll(manifest, "// indirect", "// indirect keep transitive comment")
	importRootsWrite(t, root, v1.ModuleFile, manifest)
	importRootsWrite(t, root, "consumer.proto", `syntax = "proto3"; import "svc.proto"; message Consumer { service.Service service = 1; }`)
	importRootsWrite(t, transitive, v1.ModuleFile, "module "+transitive+"\nroots api\n")
	importRootsWrite(t, transitive, "api/v1/a.proto", `syntax = "proto3"; import "v1/svc.proto";`)
	importRootsGit(t, transitive, "add", ".")
	importRootsGit(t, transitive, "commit", "--quiet", "-m", "new transitive namespace")
	importRootsGit(t, transitive, "tag", "v0.5.0")
	importRootsWrite(t, parent, v1.ModuleFile, "module "+parent+"\nrequire "+transitive+" v0.5.0\n")
	importRootsGit(t, parent, "add", ".")
	importRootsGit(t, parent, "commit", "--quiet", "-m", "require new transitive version")
	importRootsGit(t, parent, "tag", "v0.2.0")
	tidyRewriteRequirement(t, root, "v0.1.0", "v0.2.0")

	report, err := modules.TidyWithReport(t.Context(), root, cache)

	require.NoError(t, err)
	require.Len(t, report.Imports, 1)
	assert.Equal(t, transitive, report.Imports[0].Module)
	updated := string(importRootsRead(t, root, v1.ModuleFile))
	assert.Contains(t, updated, transitive+" v0.4.0 // keep transitive comment")
	assert.NotContains(t, updated, "// indirect")
}

func TestTidyRechecksSourcesCapturedBeforeResolution(t *testing.T) {
	t.Parallel()
	repository, root, cache, _ := tidyRewriteFixture(t)
	tidyRewriteNewNamespace(t, repository, root)
	manifest, lock := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)
	concurrent := `syntax = "proto3"; import "svc.proto"; // a concurrent edit`
	spy := &tidyRewriteRaceCache{Cache: cache, mutate: func() { importRootsWrite(t, root, "consumer.proto", concurrent) }}

	report, err := modules.TidyWithReport(t.Context(), root, spy)

	require.ErrorIs(t, err, sourceview.ErrChanged)
	assert.Empty(t, report.Imports)
	assert.Equal(t, concurrent, string(importRootsRead(t, root, "consumer.proto")))
	assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
	assert.Equal(t, lock, importRootsRead(t, root, v1.LockFile))
}

func TestTidyRejectsChangesToOwnedSourceSelection(t *testing.T) {
	t.Parallel()
	repository, root, cache, consumer := tidyRewriteFixture(t)
	tidyRewriteNewNamespace(t, repository, root)
	manifest, lock := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)
	concurrent := `syntax = "proto3"; import "svc.proto"; message Later { service.Service service = 1; }`
	spy := &tidyRewriteRaceCache{Cache: cache, mutate: func() { importRootsWrite(t, root, "later.proto", concurrent) }}

	report, err := modules.TidyWithReport(t.Context(), root, spy)

	require.ErrorIs(t, err, sourceview.ErrChanged)
	assert.Empty(t, report.Imports)
	assert.Equal(t, consumer, string(importRootsRead(t, root, "consumer.proto")))
	assert.Equal(t, concurrent, string(importRootsRead(t, root, "later.proto")))
	assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
	assert.Equal(t, lock, importRootsRead(t, root, v1.LockFile))
}

func TestTidyRejectsCacheChangesAfterInstallation(t *testing.T) {
	t.Parallel()
	repository, root, cache, consumer := tidyRewriteFixture(t)
	tidyRewriteNewNamespace(t, repository, root)
	manifest, lock := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)
	spy := &tidyCacheMutation{Cache: cache, mutate: func(directory string) {
		importRootsWrite(t, directory, "api/v1/svc.proto", `syntax = "proto3"; package service; message Service { string changed = 1; }`)
	}}

	report, err := modules.TidyWithReport(t.Context(), root, spy)

	require.ErrorContains(t, err, "verified pinned contents")
	assert.Empty(t, report.Imports)
	assert.Equal(t, consumer, string(importRootsRead(t, root, "consumer.proto")))
	assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
	assert.Equal(t, lock, importRootsRead(t, root, v1.LockFile))
}

func TestTidyRechecksUnchangedDependencyMetadata(t *testing.T) {
	t.Parallel()
	repository, root, cache, consumer := tidyRewriteFixture(t)
	tidyRewriteNewNamespace(t, repository, root)
	manifest, lock := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)
	spy := &tidyMetadataMutation{Cache: cache, mutate: func(directory string) {
		body := string(importRootsRead(t, directory, v1.ModuleFile))
		importRootsWrite(t, directory, v1.ModuleFile, body+"// concurrent metadata edit\n")
	}}

	report, err := modules.TidyWithReport(t.Context(), root, spy)

	require.ErrorIs(t, err, sourceview.ErrChanged)
	assert.Empty(t, report.Imports)
	assert.Equal(t, consumer, string(importRootsRead(t, root, "consumer.proto")))
	assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
	assert.Equal(t, lock, importRootsRead(t, root, v1.LockFile))
}

type tidyMetadataMutation struct {
	*gitmodules.Cache
	mutate func(string)
	calls  int
}

func (cache *tidyMetadataMutation) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	directory, module, err := cache.Cache.Cached(entry)
	cache.calls++
	if err == nil && cache.calls == 2 {
		cache.mutate(directory)
	}
	return directory, module, err
}

type tidyCacheMutation struct {
	*gitmodules.Cache
	mutate func(string)
}

func (cache *tidyCacheMutation) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	directory, module, err := cache.Cache.Cached(entry)
	if err == nil && cache.mutate != nil {
		cache.mutate(directory)
		cache.mutate = nil
	}
	return directory, module, err
}

type tidyRewriteRaceCache struct {
	*gitmodules.Cache
	mutate func()
}

func (cache *tidyRewriteRaceCache) Install(ctx context.Context, lock v1.Lock) error {
	cache.mutate()
	return cache.Cache.Install(ctx, lock)
}

func tidyRewriteFixture(t *testing.T) (string, string, *gitmodules.Cache, string) {
	t.Helper()
	repository, _ := importRootsRepository(t, map[string]string{
		"api/v1/a.proto":   `syntax = "proto3"; import "svc.proto";`,
		"api/v1/svc.proto": `syntax = "proto3"; package service; message Service {}`,
	}, nil)
	importRootsGit(t, repository, "tag", "v0.4.0")
	consumer := `syntax = "proto3"; import "svc.proto"; message Consumer { service.Service service = 1; }`
	root := importRootsConsumer(t, repository, "v0.4.0", consumer)
	cache := gitmodules.New(t.TempDir())
	require.NoError(t, modules.Tidy(t.Context(), root, cache))
	return repository, root, cache, consumer
}

func tidyRewriteNewNamespace(t *testing.T, repository, root string) {
	t.Helper()
	importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\nroots api\n")
	if _, err := os.Stat(filepath.Join(repository, "api/v1/a.proto")); err == nil {
		importRootsWrite(t, repository, "api/v1/a.proto", `syntax = "proto3"; import "v1/svc.proto";`)
	}
	importRootsGit(t, repository, "add", ".")
	importRootsGit(t, repository, "commit", "--quiet", "-m", "declare new namespace")
	importRootsGit(t, repository, "tag", "v0.5.0")
	tidyRewriteRequirement(t, root, "v0.4.0", "v0.5.0")
}
