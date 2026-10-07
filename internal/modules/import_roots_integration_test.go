package modules_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestTidyResolvesExternalImportRoots(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		files    map[string]string
		consumer string
		roots    []string
	}{
		{
			name: "intrinsic_short_import", files: map[string]string{"api/svc/v1/svc.proto": `syntax = "proto3"; import "types.proto";`, "api/svc/v1/types.proto": `syntax = "proto3";`},
			consumer: `syntax = "proto3"; import "svc.proto";`, roots: []string{"api/svc/v1"},
		},
		{
			name: "complete_intrinsic_import_suffix", files: map[string]string{"api/svc/v1/svc.proto": `syntax = "proto3"; import "svc/v1/types.proto";`, "api/svc/v1/types.proto": `syntax = "proto3";`},
			consumer: `syntax = "proto3"; import "svc/v1/svc.proto";`, roots: []string{"api"},
		},
		{
			name: "multiple_disjoint_roots", files: map[string]string{"api/a.proto": `syntax = "proto3"; import "types_a.proto";`, "api/types_a.proto": `syntax = "proto3";`, "schema/b.proto": `syntax = "proto3"; import "types_b.proto";`, "schema/types_b.proto": `syntax = "proto3";`},
			consumer: `syntax = "proto3"; import "a.proto"; import "b.proto";`, roots: []string{"api", "schema"},
		},
		{
			name: "intrinsic_import_closure", files: map[string]string{"api/svc.proto": `syntax = "proto3"; import "local.proto"; import "types.proto";`, "api/local.proto": `syntax = "proto3";`, "schema/types.proto": `syntax = "proto3";`},
			consumer: `syntax = "proto3"; import "svc.proto";`, roots: []string{"api", "schema"},
		},
		{
			name: "intrinsic_import_evidence", files: map[string]string{"api/a.proto": `syntax = "proto3"; import "b.proto";`, "api/b.proto": `syntax = "proto3";`},
			consumer: `syntax = "proto3";`, roots: []string{"api"},
		},
		{
			name: "intrinsic_source_disambiguates", files: map[string]string{"api/request.proto": `syntax = "proto3"; import "svc.proto";`, "api/svc.proto": `syntax = "proto3";`, "other/svc.proto": `syntax = "proto3"; import "unknown.proto";`},
			consumer: `syntax = "proto3"; import "svc.proto";`, roots: []string{"api"},
		},
		{
			name: "malformed_intrinsic_body", files: map[string]string{"api/a.proto": `syntax = "proto3"; import "b.proto"; message Broken {`, "api/b.proto": `syntax = "proto3";`},
			consumer: `syntax = "proto3";`, roots: []string{"api"},
		},
		{
			name: "unrelated_malformed_and_unknown_imports", files: map[string]string{
				"api/svc.proto": `syntax = "proto3"; import "types.proto";`, "api/types.proto": `syntax = "proto3";`, "api/broken.proto": `syntax = "proto3"; message Broken {`, "api/unused.proto": `syntax = "proto3"; import "unknown.proto";`,
			}, consumer: `syntax = "proto3"; import "svc.proto";`, roots: []string{"api"},
		},
		{
			name: "no_evidence_keeps_default", files: map[string]string{"api/svc.proto": `syntax = "proto3"; package svc.v1;`},
			consumer: `syntax = "proto3";`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, commit := importRootsRepository(t, tt.files, nil)
			root := importRootsConsumer(t, repository, commit, tt.consumer)
			cache := gitmodules.New(t.TempDir())

			err := modules.Tidy(t.Context(), root, cache)

			require.NoError(t, err)
			lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			require.Len(t, lock.Modules, 1)
			assert.Equal(t, commit, lock.Modules[0].Commit)
			assert.Equal(t, tt.roots, lock.Modules[0].Roots)
			_, selected, err := cache.Cached(lock.Modules[0])
			require.NoError(t, err)
			want := tt.roots
			if len(want) == 0 {
				want = []string{"."}
			}
			assert.Equal(t, want, selected.Roots)
		})
	}
}

func TestTidyInfersIntrinsicRootsBeforeUnusedAliasValidation(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, intrinsicServiceFiles(), map[string]string{"unused.proto": "../outside.proto"})
	root := importRootsConsumer(t, repository, commit, `syntax = "proto3"; import "svc.proto";`)
	cache := gitmodules.New(t.TempDir())

	err := modules.Tidy(t.Context(), root, cache)

	require.NoError(t, err)
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	require.Len(t, lock.Modules, 1)
	assert.Equal(t, []string{"api"}, lock.Modules[0].Roots)
	assert.NoError(t, modules.Download(t.Context(), root, gitmodules.New(t.TempDir())))
}

func TestIntrinsicImportsDoNotInferOtherModuleRoots(t *testing.T) {
	t.Parallel()
	b, bCommit := importRootsRepository(t, map[string]string{"api/b.proto": `syntax = "proto3";`}, nil)
	a, _ := importRootsRepository(t, map[string]string{"api/a.proto": `syntax = "proto3"; import "b.proto";`}, nil)
	importRootsWrite(t, a, v1.ModuleFile, "direct (\n  "+b+"@"+bCommit+"\n)\n")
	importRootsGit(t, a, "add", ".")
	importRootsGit(t, a, "commit", "--quiet", "-m", "requirements-only metadata")
	aCommit := importRootsGit(t, a, "rev-parse", "HEAD")
	root := importRootsConsumer(t, a, aCommit, `syntax = "proto3";`)

	err := modules.Tidy(t.Context(), root, gitmodules.New(t.TempDir()))

	require.NoError(t, err)
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	require.Len(t, lock.Modules, 2)
	for _, entry := range lock.Modules {
		if entry.Source == b {
			assert.Empty(t, entry.Roots)
		}
	}
}

func TestTidyDoesNotInferRootsFromConsumerImports(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		aliases map[string]string
	}{
		{name: "no_intrinsic_evidence"},
		{name: "unused_unsafe_alias", aliases: map[string]string{"unused.proto": "../outside.proto"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, commit := importRootsRepository(t, map[string]string{"api/svc.proto": `syntax = "proto3";`}, tt.aliases)
			root := importRootsConsumer(t, repository, commit, `syntax = "proto3"; import "svc.proto";`)
			manifest := importRootsRead(t, root, v1.ModuleFile)
			originalLock := "# unchanged\nversion: 1\nmodules: []\n"
			importRootsWrite(t, root, v1.LockFile, originalLock)
			cache := gitmodules.New(t.TempDir())

			err := modules.Tidy(t.Context(), root, cache)

			require.Error(t, err)
			assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
			assert.Equal(t, originalLock, string(importRootsRead(t, root, v1.LockFile)))
			require.NoError(t, modules.GetWithRoots(t.Context(), root, v1.Requirement{Module: repository, Version: commit}, cache, []string{"api"}))
			lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Equal(t, []string{"api"}, lock.Modules[0].Roots)
		})
	}
}

func TestTidyRejectsAmbiguousIntrinsicImportRoots(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, map[string]string{"api/nested/a.proto": `syntax = "proto3"; import "v1/svc.proto";`, "api/v1/svc.proto": `syntax = "proto3";`, "api/nested/v1/svc.proto": `syntax = "proto3";`}, nil)
	root := importRootsConsumer(t, repository, commit, `syntax = "proto3";`)
	original := importRootsRead(t, root, v1.ModuleFile)

	err := modules.Tidy(t.Context(), root, gitmodules.New(t.TempDir()))

	require.ErrorContains(t, err, "ambiguous")
	assert.ErrorContains(t, err, "api")
	assert.ErrorContains(t, err, "api/nested")
	assert.Equal(t, original, importRootsRead(t, root, v1.ModuleFile))
	assert.NoFileExists(t, filepath.Join(root, v1.LockFile))
}

func TestIntrinsicFileAliasesRetainAmbiguousAlternatives(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, map[string]string{
		"api/b.proto":     `syntax = "proto3"; package api;`,
		"sources/a.proto": `syntax = "proto3"; import "b.proto";`,
		"sources/b.proto": `syntax = "proto3"; package sources;`,
	}, map[string]string{"api/a.proto": "../sources/a.proto"})
	root := importRootsConsumer(t, repository, commit, `syntax = "proto3";`)

	err := modules.Tidy(t.Context(), root, gitmodules.New(t.TempDir()))

	require.ErrorContains(t, err, "ambiguous")
	assert.ErrorContains(t, err, "api")
	assert.ErrorContains(t, err, "sources")
	assert.NoFileExists(t, filepath.Join(root, v1.LockFile))
}

func TestGetWithRootsReplacesPriorSelection(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, map[string]string{"api/svc.proto": `syntax = "proto3";`, "other/svc.proto": `syntax = "proto3";`}, map[string]string{"unused.proto": "../outside.proto"})
	root := importRootsConsumer(t, repository, commit, `syntax = "proto3"; import "svc.proto";`)
	cache := gitmodules.New(t.TempDir())
	target := v1.Requirement{Module: repository, Version: commit}

	require.NoError(t, modules.GetWithRoots(t.Context(), root, target, cache, []string{"api"}))
	first, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	require.NoError(t, modules.Tidy(t.Context(), root, cache))
	retained, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	assert.Equal(t, first, retained)

	require.NoError(t, modules.GetWithRoots(t.Context(), root, target, cache, []string{"other"}))
	second, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	require.Len(t, second.Modules, 1)
	assert.Equal(t, []string{"other"}, second.Modules[0].Roots)
	assert.Equal(t, first.Modules[0].Hash, second.Modules[0].Hash)
	assert.Equal(t, commit, second.Modules[0].Commit)
	_, again, err := cache.Cached(first.Modules[0])
	require.NoError(t, err)
	assert.Equal(t, []string{"api"}, again.Roots)
}

func TestGetWithRootsCanonicalizesHints(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, map[string]string{"api/a.proto": `syntax = "proto3";`, "schema/b.proto": `syntax = "proto3";`}, nil)
	root := importRootsConsumer(t, repository, commit, `syntax = "proto3"; import "a.proto"; import "b.proto";`)

	err := modules.GetWithRoots(t.Context(), root, v1.Requirement{Module: repository, Version: commit}, gitmodules.New(t.TempDir()), []string{"schema/", "./api", "api"})

	require.NoError(t, err)
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	assert.Equal(t, []string{"api", "schema"}, lock.Modules[0].Roots)
}

func TestGetWithRootsRetainsExplicitHiddenScope(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, map[string]string{".api/svc.proto": `syntax = "proto3";`}, nil)
	root := importRootsConsumer(t, repository, commit, `syntax = "proto3"; import "svc.proto";`)

	err := modules.GetWithRoots(t.Context(), root, v1.Requirement{Module: repository, Version: commit}, gitmodules.New(t.TempDir()), []string{".api"})

	require.NoError(t, err)
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	assert.Equal(t, []string{".api"}, lock.Modules[0].Roots)
	require.NoError(t, modules.Tidy(t.Context(), root, gitmodules.New(t.TempDir())))
}

func TestInferredRootsColdFrozenReplay(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, intrinsicServiceFiles(), nil)
	root := importRootsConsumer(t, repository, commit, `syntax = "proto3"; import "svc.proto";`)
	require.NoError(t, modules.Tidy(t.Context(), root, gitmodules.New(t.TempDir())))
	manifest := importRootsRead(t, root, v1.ModuleFile)
	lock := importRootsRead(t, root, v1.LockFile)
	importRootsWrite(t, repository, "api/svc.proto", `syntax = "proto3"; message Later {}`)
	importRootsGit(t, repository, "add", ".")
	importRootsGit(t, repository, "commit", "--quiet", "-m", "move HEAD")

	sources, err := modules.EnsureFrozenSources(t.Context(), root, gitmodules.New(t.TempDir()))

	require.NoError(t, err)
	owners, err := sources.FileModules()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"svc.proto": repository, "types.proto": repository}, owners)
	assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
	assert.Equal(t, lock, importRootsRead(t, root, v1.LockFile))
}

func TestImportRootFailuresPreserveConsumerFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		files    map[string]string
		aliases  map[string]string
		consumer string
		error    string
	}{
		{name: "selected_bad_alias", files: intrinsicServiceFiles(), aliases: map[string]string{"api/bad.proto": "../../outside.proto"}, consumer: `syntax = "proto3"; import "svc.proto";`, error: "outside root"},
		{name: "preserve_resolved_namespace", files: map[string]string{"api/request.proto": `syntax = "proto3"; import "api/svc.proto"; import "types.proto";`, "api/svc.proto": `syntax = "proto3";`, "schema/types.proto": `syntax = "proto3";`}, consumer: `syntax = "proto3";`, error: "previously resolved source"},
		{name: "import_name_collision", files: map[string]string{"api/a.proto": `syntax = "proto3"; import "types_a.proto";`, "api/types_a.proto": `syntax = "proto3";`, "api/svc.proto": `syntax = "proto3";`, "schema/b.proto": `syntax = "proto3"; import "types_b.proto";`, "schema/types_b.proto": `syntax = "proto3";`, "schema/svc.proto": `syntax = "proto3";`}, consumer: `syntax = "proto3";`, error: "duplicate import path"},
		{name: "component_bounded_suffix", files: map[string]string{"api/notservice.proto": `syntax = "proto3";`}, consumer: `syntax = "proto3"; import "service.proto";`, error: "cannot resolve"},
		{name: "case_sensitive_suffix", files: map[string]string{"api/Service.proto": `syntax = "proto3";`}, consumer: `syntax = "proto3"; import "service.proto";`, error: "cannot resolve"},
		{name: "reached_malformed_proto", files: map[string]string{"api/svc.proto": `syntax = "proto3"; message Broken {`}, consumer: `syntax = "proto3"; import "api/svc.proto";`, error: "Parse"},
		{name: "unknown_reachable_import", files: map[string]string{"api/svc.proto": `syntax = "proto3"; import "unknown.proto";`}, consumer: `syntax = "proto3"; import "api/svc.proto";`, error: "unknown.proto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, commit := importRootsRepository(t, tt.files, tt.aliases)
			root := importRootsConsumer(t, repository, commit, tt.consumer)
			manifest := importRootsRead(t, root, v1.ModuleFile)
			originalLock := "# unchanged\nversion: 1\nmodules: []\n"
			importRootsWrite(t, root, v1.LockFile, originalLock)

			err := modules.Tidy(t.Context(), root, gitmodules.New(t.TempDir()))

			require.ErrorContains(t, err, tt.error)
			assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
			assert.Equal(t, originalLock, string(importRootsRead(t, root, v1.LockFile)))
		})
	}
}

func TestImportRootsRespectAuthoritativeMetadata(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		file    string
		content string
	}{
		{name: "native_default", file: v1.ModuleFile, content: "module %s\n"},
		{name: "buf_default", file: "buf.yaml", content: "version: v1\n"},
		{name: "legacy_declared_default", file: "easyp.yaml", content: "generate:\n  inputs:\n    - directory:\n        path: .\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, _ := importRootsRepository(t, map[string]string{"api/svc.proto": `syntax = "proto3";`}, nil)
			content := strings.ReplaceAll(tt.content, "%s", repository)
			importRootsWrite(t, repository, tt.file, content)
			importRootsGit(t, repository, "add", ".")
			importRootsGit(t, repository, "commit", "--quiet", "-m", "metadata")
			commit := importRootsGit(t, repository, "rev-parse", "HEAD")
			root := importRootsConsumer(t, repository, commit, `syntax = "proto3"; import "svc.proto";`)

			err := modules.Tidy(t.Context(), root, gitmodules.New(t.TempDir()))

			require.ErrorContains(t, err, "cannot resolve")
			assert.NoFileExists(t, filepath.Join(root, v1.LockFile))
		})
	}
}

func TestImportRootFinalizationProtectsLockedHash(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, intrinsicServiceFiles(), nil)
	importRootsGit(t, repository, "tag", "v1.0.0")
	root := importRootsConsumer(t, repository, "v1.0.0", `syntax = "proto3"; import "svc.proto";`)
	cache := gitmodules.New(t.TempDir())
	require.NoError(t, modules.Tidy(t.Context(), root, cache))
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	lock.Modules[0].Hash = "h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="
	// A valid but wrong original hash must be compared after provisional
	// inspections have been replaced with a complete exact-commit fetch.
	raw := "version: 1\nmodules:\n  - source: " + repository + "\n    version: v1.0.0\n    commit: " + commit + "\n    hash: " + lock.Modules[0].Hash + "\n"
	importRootsWrite(t, root, v1.LockFile, raw)
	spy := &importRootsInstallSpy{Cache: cache}

	err = modules.Tidy(t.Context(), root, spy)

	require.ErrorIs(t, err, modules.ErrLockedVersionChanged)
	assert.Zero(t, spy.installs)
	assert.Equal(t, raw, string(importRootsRead(t, root, v1.LockFile)))
}

func TestImportRootFinalizationUsesInspectedCommit(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, intrinsicServiceFiles(), nil)
	root := importRootsConsumer(t, repository, "", `syntax = "proto3"; import "svc.proto";`)
	cache := &importRootsMovingHeadCache{Cache: gitmodules.New(t.TempDir())}
	cache.move = func() {
		importRootsWrite(t, repository, "api/svc.proto", `syntax = "proto3"; message Later {}`)
		importRootsGit(t, repository, "add", ".")
		importRootsGit(t, repository, "commit", "--quiet", "-m", "move HEAD during resolution")
	}

	err := modules.Tidy(t.Context(), root, cache)

	require.NoError(t, err)
	assert.Equal(t, []string{commit}, cache.finalVersions)
	lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
	require.NoError(t, err)
	assert.Equal(t, commit, lock.Modules[0].Commit)
}

func TestUpdateRevalidatesPreviousImportRoots(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		move        bool
		metadata    bool
		wantRoots   []string
		wantFailure bool
	}{
		{name: "unchanged_namespace", wantRoots: []string{"api"}},
		{name: "new_authoritative_namespace", move: true, metadata: true},
		{name: "invalid_previous_choice", move: true, wantFailure: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository, _ := importRootsRepository(t, intrinsicServiceFiles(), nil)
			importRootsGit(t, repository, "tag", "v1.0.0")
			root := importRootsConsumer(t, repository, "v1.0.0", `syntax = "proto3"; import "svc.proto";`)
			cache := gitmodules.New(t.TempDir())
			require.NoError(t, modules.Tidy(t.Context(), root, cache))
			manifest, before := importRootsRead(t, root, v1.ModuleFile), importRootsRead(t, root, v1.LockFile)
			if tt.move {
				require.NoError(t, os.MkdirAll(filepath.Join(repository, "schema"), 0o755))
				require.NoError(t, os.Rename(filepath.Join(repository, "api/svc.proto"), filepath.Join(repository, "schema/svc.proto")))
				require.NoError(t, os.Rename(filepath.Join(repository, "api/types.proto"), filepath.Join(repository, "schema/types.proto")))
			} else {
				importRootsWrite(t, repository, "api/svc.proto", `syntax = "proto3"; message Updated {}`)
			}
			if tt.metadata {
				importRootsWrite(t, repository, v1.ModuleFile, "module "+repository+"\nroots schema\n")
			}
			importRootsGit(t, repository, "add", ".")
			importRootsGit(t, repository, "commit", "--quiet", "-m", "next revision")
			importRootsGit(t, repository, "tag", "v1.1.0")
			commit := importRootsGit(t, repository, "rev-parse", "HEAD")

			err := modules.Update(t.Context(), root, cache)

			if tt.wantFailure {
				require.Error(t, err)
				assert.Equal(t, manifest, importRootsRead(t, root, v1.ModuleFile))
				assert.Equal(t, before, importRootsRead(t, root, v1.LockFile))
				return
			}
			require.NoError(t, err)
			lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Equal(t, "v1.1.0", lock.Modules[0].Version)
			assert.Equal(t, commit, lock.Modules[0].Commit)
			assert.Equal(t, tt.wantRoots, lock.Modules[0].Roots)
		})
	}
}

func TestLocalOverlayInfersRemoteRootsBeforeInstall(t *testing.T) {
	t.Parallel()
	repository, commit := importRootsRepository(t, intrinsicServiceFiles(), map[string]string{"unused.proto": "../outside.proto"})
	replacement := t.TempDir()
	importRootsWrite(t, replacement, v1.ModuleFile, "module example.test/local\nrequire "+repository+" "+commit+"\n")
	importRootsWrite(t, replacement, "local.proto", `syntax = "proto3";`)
	root := t.TempDir()
	importRootsWrite(t, root, v1.ModuleFile, "module example.test/consumer\nrequire example.test/local\nreplace example.test/local => "+replacement+"\n")
	importRootsWrite(t, root, "consumer.proto", `syntax = "proto3"; import "svc.proto";`)
	lock := "# published lock must stay unchanged\nversion: 1\nmodules: []\n"
	importRootsWrite(t, root, v1.LockFile, lock)

	err := modules.Tidy(t.Context(), root, gitmodules.New(t.TempDir()))

	require.NoError(t, err)
	assert.Equal(t, lock, string(importRootsRead(t, root, v1.LockFile)))
}

type importRootsMovingHeadCache struct {
	*gitmodules.Cache
	move          func()
	finalVersions []string
}

func (cache *importRootsMovingHeadCache) FetchForRootResolution(ctx context.Context, source, version string) (modules.Fetched, error) {
	fetched, err := cache.Cache.FetchForRootResolution(ctx, source, version)
	if err == nil && cache.move != nil {
		cache.move()
		cache.move = nil
	}
	return fetched, err
}

func (cache *importRootsMovingHeadCache) FetchWithRoots(ctx context.Context, source, version string, roots []string) (modules.Fetched, error) {
	cache.finalVersions = append(cache.finalVersions, version)
	return cache.Cache.FetchWithRoots(ctx, source, version, roots)
}

type importRootsInstallSpy struct {
	*gitmodules.Cache
	installs int
}

func (cache *importRootsInstallSpy) Install(ctx context.Context, lock v1.Lock) error {
	cache.installs++
	return cache.Cache.Install(ctx, lock)
}

func TestLocalOverlayInfersImportRootsInMemory(t *testing.T) {
	t.Parallel()
	replacement := t.TempDir()
	for name, content := range intrinsicServiceFiles() {
		importRootsWrite(t, replacement, name, content)
	}
	root := t.TempDir()
	manifest := "module example.test/consumer\nrequire example.test/dep\nreplace example.test/dep => " + replacement + "\n"
	importRootsWrite(t, root, v1.ModuleFile, manifest)
	importRootsWrite(t, root, "consumer.proto", `syntax = "proto3"; import "svc.proto";`)
	lock := "# published lock must stay unchanged\nversion: 1\nmodules: []\n"
	importRootsWrite(t, root, v1.LockFile, lock)

	err := modules.Tidy(t.Context(), root, nil)

	require.NoError(t, err)
	assert.Equal(t, manifest, string(importRootsRead(t, root, v1.ModuleFile)))
	assert.Equal(t, lock, string(importRootsRead(t, root, v1.LockFile)))
	_, module, err := modules.ReadManifest(root)
	require.NoError(t, err)
	graph, err := modules.EnsureEffectiveGraph(t.Context(), root, module, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"api"}, graph.Modules["example.test/dep"].Module.Roots)
}

func TestLocalOverlayGetWithRootsScopesAliases(t *testing.T) {
	t.Parallel()
	replacement := t.TempDir()
	importRootsWrite(t, replacement, "api/svc.proto", `syntax = "proto3";`)
	importRootsWrite(t, replacement, "other/svc.proto", `syntax = "proto3";`)
	require.NoError(t, os.Symlink("../outside.proto", filepath.Join(replacement, "unused.proto")))
	root := t.TempDir()
	manifest := "module example.test/consumer\nrequire example.test/dep\nreplace example.test/dep => " + replacement + "\n"
	importRootsWrite(t, root, v1.ModuleFile, manifest)
	importRootsWrite(t, root, "consumer.proto", `syntax = "proto3"; import "svc.proto";`)
	lock := "# published lock must stay unchanged\nversion: 1\nmodules: []\n"
	importRootsWrite(t, root, v1.LockFile, lock)

	err := modules.GetWithRoots(t.Context(), root, v1.Requirement{Module: "example.test/dep"}, nil, []string{"api"})

	require.NoError(t, err)
	assert.Equal(t, manifest, string(importRootsRead(t, root, v1.ModuleFile)))
	assert.Equal(t, lock, string(importRootsRead(t, root, v1.LockFile)))
}

func importRootsRepository(t *testing.T, files, aliases map[string]string) (string, string) {
	t.Helper()
	repository := t.TempDir()
	importRootsGit(t, repository, "init", "--quiet")
	importRootsGit(t, repository, "config", "user.email", "roots@example.test")
	importRootsGit(t, repository, "config", "user.name", "Import Roots Test")
	for name, content := range files {
		importRootsWrite(t, repository, name, content)
	}
	for name, target := range aliases {
		path := filepath.Join(repository, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.Symlink(target, path))
	}
	importRootsGit(t, repository, "add", ".")
	importRootsGit(t, repository, "commit", "--quiet", "-m", "fixture")
	return repository, importRootsGit(t, repository, "rev-parse", "HEAD")
}

func intrinsicServiceFiles() map[string]string {
	return map[string]string{
		"api/svc.proto":   `syntax = "proto3"; import "types.proto";`,
		"api/types.proto": `syntax = "proto3";`,
	}
}

func importRootsConsumer(t *testing.T, repository, version, content string) string {
	t.Helper()
	root := t.TempDir()
	importRootsWrite(t, root, v1.ModuleFile, "module example.test/consumer\nrequire "+filepath.ToSlash(repository)+" "+version+"\n")
	importRootsWrite(t, root, "consumer.proto", content)
	return root
}

func importRootsWrite(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func importRootsRead(t *testing.T, root, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, name))
	require.NoError(t, err)
	return raw
}

func importRootsGit(t *testing.T, directory string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	require.NoError(t, err, "%s", output)
	return strings.TrimSpace(string(output))
}
