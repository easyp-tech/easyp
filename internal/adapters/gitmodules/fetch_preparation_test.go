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
)

func TestPrepareFetchDefersBSRBackendAndCleansTemporaryWork(t *testing.T) {
	t.Parallel()
	remote := bsrTestRepository(t, "version: v1\ndeps: [buf.build/googleapis/googleapis]\n")
	backend := &fakeBSRResolver{git: v1.Requirement{Module: "github.com/acme/provider", Version: "v1.2.3"}}
	cache := NewWithBSRResolver(t.TempDir(), backend)
	preparer, supported := any(cache).(interface {
		PrepareFetch(context.Context, modules.FetchRequest) (modules.PreparedFetch, error)
	})
	require.True(t, supported, "Git source must support independent fetch preparation")
	complete, err := preparer.PrepareFetch(t.Context(), modules.FetchRequest{Source: remote, Version: "v1.0.0", Inspect: true})
	require.NoError(t, err)
	assert.Empty(t, backend.calls, "metadata backends are consumed in serial request order")
	entries, err := os.ReadDir(cache.root)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.Contains(t, []string{"objects", "modules", "sources"}, entry.Name(), "preparation must release temporary work even if its callback is discarded")
	}
	fetched, err := complete(t.Context())
	require.NoError(t, err)
	require.Len(t, backend.calls, 1)
	assert.Equal(t, "v1.0.0", fetched.Lock.Version)
	assert.Equal(t, []v1.Requirement{backend.git}, fetched.Module.Requires)
	assert.NotNil(t, fetched.Inspection)
	assert.Len(t, fetched.Lock.BSR, 1)
	_, _, err = cache.Cached(fetched.Lock)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(cache.root, "sources", v1CacheSourceKey(remote), fetched.Lock.Commit+".json"))
}

func TestPreparedGitResolutionPreservesVersionlessSemanticAndCommitConstraints(t *testing.T) {
	t.Parallel()
	dependency := preparedGitRepository(t, "a", nil)
	require.NoError(t, os.WriteFile(filepath.Join(dependency, "a.proto"), []byte("syntax = \"proto3\"; package a; message Selected {}\n"), 0o600))
	runTestGit(t, dependency, "add", ".")
	runTestGit(t, dependency, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "selected version")
	runTestGit(t, dependency, "tag", "v1.2.0")
	selectedCommit := strings.TrimSpace(runTestGit(t, dependency, "rev-parse", "HEAD"))
	// HEAD is deliberately unusable. The weak root requirement must wait for
	// the semantic and exact requirements discovered from independent modules.
	require.NoError(t, os.WriteFile(filepath.Join(dependency, v1.ModuleFile), []byte("module "+dependency+"\nrequire example.test/unavailable v1.0.0\n"), 0o600))
	runTestGit(t, dependency, "add", ".")
	runTestGit(t, dependency, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "unselected HEAD")
	semantic := preparedGitRepository(t, "b", []v1.Requirement{{Module: dependency, Version: "v1.2.0"}})
	pinned := preparedGitRepository(t, "c", []v1.Requirement{{Module: dependency, Version: selectedCommit}})
	root := v1.Module{Requires: []v1.Requirement{{Module: dependency}, {Module: semantic, Version: "v1.0.0"}, {Module: pinned, Version: "v1.0.0"}}}
	cache := New(t.TempDir())
	prepared, err := modules.Resolve(t.Context(), root, cache, nil)
	require.NoError(t, err)
	serial, err := modules.Resolve(t.Context(), root, struct{ modules.Source }{New(t.TempDir())}, nil)
	require.NoError(t, err)
	assert.Equal(t, serial, prepared)
	require.Len(t, prepared.Modules, 3)
	for _, entry := range prepared.Modules {
		if entry.Source == dependency {
			assert.Equal(t, "v1.2.0", entry.Version)
			assert.Equal(t, selectedCommit, entry.Commit)
		}
	}
	require.NoError(t, cache.Install(t.Context(), prepared))
	require.NoError(t, cache.VerifyCached(t.Context(), prepared))
}

func TestPreparedProvisionalRevisionFinalizesWithoutAnotherNetworkFetch(t *testing.T) {
	t.Parallel()
	repository, commit := migrationTestRepository(t, map[string]string{"api/a.proto": "syntax = \"proto3\"; package a;\n"})
	runTestGit(t, repository, "tag", "v1.0.0")
	cache := New(t.TempDir())
	complete, err := cache.PrepareFetch(t.Context(), modules.FetchRequest{Source: repository, Version: "v1.0.0", Inspect: true})
	require.NoError(t, err)
	fetched, err := complete(t.Context())
	require.NoError(t, err)
	require.NotNil(t, fetched.Inspection)
	assert.True(t, fetched.Inspection.Provisional)
	assert.Empty(t, fetched.Lock.Hash)
	require.NoError(t, os.Rename(repository, repository+"-offline"))
	finalized, err := cache.FetchWithRoots(t.Context(), repository, commit, []string{"api"})
	require.NoError(t, err)
	assert.Equal(t, commit, finalized.Lock.Commit)
	assert.Equal(t, []string{"api"}, finalized.Lock.Roots)
	assert.NotEmpty(t, finalized.Lock.Hash)
}

func preparedGitRepository(t *testing.T, name string, requirements []v1.Requirement) string {
	t.Helper()
	remote := t.TempDir()
	manifest := "module " + remote + "\n"
	for _, requirement := range requirements {
		manifest += "require " + requirement.Module + " " + requirement.Version + "\n"
	}
	require.NoError(t, os.WriteFile(filepath.Join(remote, v1.ModuleFile), []byte(manifest), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(remote, name+".proto"), []byte("syntax = \"proto3\"; package "+name+";\n"), 0o600))
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	runTestGit(t, remote, "tag", "v1.0.0")
	return remote
}
