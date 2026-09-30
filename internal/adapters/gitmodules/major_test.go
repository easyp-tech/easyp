package gitmodules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/mod/sumdb/dirhash"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestMajorGitLayouts(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, logical, physical, tag string }{
		{"root", "v2", ".", "v2.0.0"},
		{"major_subdir", "v2", "v2", "v2.0.0"},
		{"nested_root", "sub/v2", "sub", "sub/v2.0.0"},
		{"nested_major_subdir", "sub/v2", "sub/v2", "sub/v2.0.0"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := t.TempDir()
			source := filepath.ToSlash(filepath.Join(repo, tt.logical))
			dir := filepath.Join(repo, tt.physical)
			require.NoError(t, os.MkdirAll(dir, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(dir, v1.ModuleFile), []byte("module "+source+"\nroots proto\n"), 0o600))
			runTestGit(t, repo, "init", "-q")
			runTestGit(t, repo, "add", ".")
			runTestGit(t, repo, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "module")
			runTestGit(t, repo, "-c", "user.name=Test", "-c", "user.email=test@example.com", "tag", "-a", tt.tag, "-m", "release")
			commit := strings.TrimSpace(runTestGit(t, repo, "rev-parse", "HEAD"))
			cache := New(t.TempDir())
			for _, version := range []string{"v2.0.0", commit, ""} {
				fetched, err := cache.Fetch(t.Context(), source, version)
				require.NoError(t, err)
				require.Equal(t, source, fetched.Module.Name)
				require.Equal(t, commit, fetched.Lock.Commit)
				require.Equal(t, []string{filepath.Join(tt.physical, "proto")}, fetched.Module.Roots)
				lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
				require.NoError(t, cache.Install(t.Context(), lock))
				_, module, err := cache.Cached(fetched.Lock)
				require.NoError(t, err)
				require.Equal(t, fetched.Module, module)
			}
			versions, err := cache.Versions(t.Context(), source)
			require.NoError(t, err)
			require.Equal(t, []string{"v2.0.0"}, versions)
		})
	}
}

func TestMajorGitWrongPhysicalIdentity(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	source := filepath.ToSlash(filepath.Join(repo, "sub/v2"))
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "elsewhere"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "elsewhere", v1.ModuleFile), []byte("module "+source+"\n"), 0o600))
	runTestGit(t, repo, "init", "-q")
	runTestGit(t, repo, "add", ".")
	runTestGit(t, repo, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "wrong location")
	runTestGit(t, repo, "tag", "sub/v2.0.0")
	commit := strings.TrimSpace(runTestGit(t, repo, "rev-parse", "HEAD"))
	for _, version := range []string{"v2.0.0", commit, ""} {
		_, err := New(t.TempDir()).Fetch(t.Context(), source, version)
		require.Error(t, err)
	}
}

func TestMajorWarmCacheRequiresNativeIdentity(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, location string }{
		{"legacy", ""}, {"wrong_physical_location", "elsewhere"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := New(t.TempDir())
			entry := v1.LockedModule{Source: "example.com/repo/v2", Version: strings.Repeat("a", 40), Commit: strings.Repeat("a", 40)}
			installed := v1ModuleCachePath(cache.root, entry)
			require.NoError(t, os.MkdirAll(installed, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(installed, "api.proto"), []byte("syntax = \"proto3\"; package api;"), 0o600))
			if tt.location != "" {
				require.NoError(t, os.MkdirAll(filepath.Join(installed, tt.location), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(installed, tt.location, v1.ModuleFile), []byte("module "+entry.Source+"\n"), 0o600))
			}
			var err error
			entry.Hash, err = dirhash.HashDir(installed, "", dirhash.Hash1)
			require.NoError(t, err)
			_, _, err = cache.Cached(entry)
			require.Error(t, err)
		})
	}
}

func TestMajorMigrationRejectsUnsuffixedNativeRelease(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(repo, v1.ModuleFile), []byte("module "+repo+"\n"), 0o600))
	runTestGit(t, repo, "init", "-q")
	runTestGit(t, repo, "add", ".")
	runTestGit(t, repo, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "native")
	runTestGit(t, repo, "tag", "v2.0.0")
	for _, tt := range []struct{ name, version, diagnostic string }{
		{"unmarked", "v2.0.0", "update protobuf.mod explicitly"},
		{"native_incompatible", "v2.0.0+incompatible", "native protobuf.mod"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(t.TempDir()).FetchMigration(t.Context(), repo, tt.version, "")
			require.ErrorContains(t, err, tt.diagnostic)
		})
	}
}

func TestMajorLegacyRequirementsAtNativeBoundary(t *testing.T) {
	t.Parallel()
	legacy, native := t.TempDir(), t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(native, v1.ModuleFile), []byte("module "+native+"\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(legacy, "easyp.yaml"), []byte("deps:\n  - "+native+"@v2.0.0\n"), 0o600))
	for _, repo := range []string{legacy, native} {
		runTestGit(t, repo, "init", "-q")
		runTestGit(t, repo, "add", ".")
		runTestGit(t, repo, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "metadata")
	}
	runTestGit(t, legacy, "tag", "v1.0.0")
	runTestGit(t, native, "tag", "v2.0.0")
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), legacy, "v1.0.0")
	require.NoError(t, err)
	require.Equal(t, []v1.Requirement{{Module: native, Version: "v2.0.0"}}, fetched.Module.Requires)
	_, err = modules.Resolve(t.Context(), v1.Module{Requires: []v1.Requirement{{Module: legacy, Version: "v1.0.0"}}}, cache, nil)
	require.ErrorContains(t, err, "update protobuf.mod explicitly")
}

func TestMajorLegacyMarkerIsNotAGitTag(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(repo, "api.proto"), []byte("syntax = \"proto3\"; package api;"), 0o600))
	runTestGit(t, repo, "init", "-q")
	runTestGit(t, repo, "add", ".")
	runTestGit(t, repo, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "legacy")
	runTestGit(t, repo, "tag", "v2.0.0")
	runTestGit(t, repo, "tag", "v2.9.0+incompatible")
	cache := New(t.TempDir())
	versions, err := cache.Versions(t.Context(), repo)
	require.NoError(t, err)
	require.Equal(t, []string{"v2.0.0+incompatible"}, versions)
	fetched, err := cache.FetchMigration(t.Context(), repo, "v2.0.0+incompatible", "")
	require.NoError(t, err)
	require.Equal(t, "v2.0.0+incompatible", fetched.Lock.Version)
}

func TestMajorURLCandidatePreservesTransport(t *testing.T) {
	t.Parallel()
	candidates, err := v1GitModuleCandidates("ssh://git@example.com:2222/repo%20space/v%32")
	require.NoError(t, err)
	require.Equal(t, "ssh://git@example.com:2222/repo%20space", candidates[0].remote)
	require.Empty(t, candidates[0].subdir)
	require.Equal(t, "v2.0.0", candidates[0].tag("v2.0.0"))
}
