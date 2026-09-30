package modules

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestEnsureFrozenSourcesPreflight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		manifest string
		lock     string
		wantErr  string
	}{
		{name: "missing_manifest", lock: "version: 1\nmodules: []\n", wantErr: "protobuf.mod"},
		{name: "legacy_manifest", manifest: "direct (\nexample.com/a v1.0.0\n)\n", lock: "version: 1\nmodules: []\n", wantErr: "legacy"},
		{name: "missing_empty_lock", manifest: "module example.com/app\n", wantErr: "protobuf.lock"},
		{name: "malformed_empty_lock", manifest: "module example.com/app\n", lock: "version: 99\n", wantErr: "unsupported protobuf.lock"},
		{name: "empty_lock", manifest: "module example.com/app\n", lock: "version: 1\nmodules: []\n"},
		{name: "unreachable_replace_before_lock", manifest: "module example.com/app\nreplace example.com/unused => ./missing\n", lock: "invalid yaml: [", wantErr: "replace"},
		{name: "required_replace_before_cache", manifest: "module example.com/app\nrequire example.com/a v1.0.0\nreplace example.com/a => ./missing\n", lock: "version: 1\nmodules: []\n", wantErr: "replace"},
		{name: "missing_direct_requirement", manifest: "module example.com/app\nrequire example.com/a v1.0.0\n", lock: "version: 1\nmodules: []\n", wantErr: "does not satisfy example.com/a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			if tt.manifest != "" {
				writeV1GenerateFixture(t, root, v1.ModuleFile, tt.manifest)
			}
			if tt.lock != "" {
				writeV1GenerateFixture(t, root, v1.LockFile, tt.lock)
			}
			cache := &frozenTestRepository{}

			roots, err := EnsureFrozenSources(t.Context(), root, cache)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.ErrorContains(t, err, "frozen graph "+root)
			} else {
				require.NoError(t, err)
				assert.Empty(t, roots)
			}
			assert.Zero(t, cache.installs)
			assert.Empty(t, cache.cached)
			assert.Zero(t, cache.fetches)
			assertFrozenFile(t, root, v1.ModuleFile, tt.manifest)
			assertFrozenFile(t, root, v1.LockFile, tt.lock)
		})
	}
}

func TestEnsureFrozenSourcesGraph(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		requirement  string
		locked       []string
		aRequires    []v1.Requirement
		bRequires    []v1.Requirement
		aReplaces    []v1.Replacement
		wantErr      string
		wantInstalls int
	}{
		{name: "direct", requirement: "v1.0.0", locked: []string{"a"}, wantInstalls: 1},
		{name: "semantic_minimum", requirement: "v0.9.0", locked: []string{"a"}, wantInstalls: 1},
		{name: "versionless", locked: []string{"a"}, wantInstalls: 1},
		{name: "exact_commit", requirement: versionlessCommitA, locked: []string{"a"}, wantInstalls: 1},
		{name: "invalid_direct_version", requirement: "latest", locked: []string{"a"}, wantErr: "expected semantic version or full Git commit"},
		{name: "invalid_transitive_version", requirement: "v1.0.0", locked: []string{"a", "b"}, aRequires: []v1.Requirement{{Module: "example.com/b", Version: "latest"}}, wantErr: "expected semantic version or full Git commit", wantInstalls: 1},
		{name: "outdated_direct_version", requirement: "v2.0.0", locked: []string{"a"}, wantErr: "does not satisfy example.com/a"},
		{name: "wrong_direct_commit", requirement: versionlessCommitBOld, locked: []string{"a"}, wantErr: "does not satisfy example.com/a"},
		{name: "transitive", requirement: "v1.0.0", locked: []string{"b", "a"}, aRequires: []v1.Requirement{{Module: "example.com/b", Version: "v1.0.0"}}, wantInstalls: 2},
		{name: "transitive_cycle", requirement: "v1.0.0", locked: []string{"a", "b"}, aRequires: []v1.Requirement{{Module: "example.com/b"}}, bRequires: []v1.Requirement{{Module: "example.com/a"}}, wantInstalls: 2},
		{name: "missing_transitive", requirement: "v1.0.0", locked: []string{"a"}, aRequires: []v1.Requirement{{Module: "example.com/b"}}, wantErr: "does not satisfy example.com/b", wantInstalls: 1},
		{name: "outdated_transitive", requirement: "v1.0.0", locked: []string{"a", "b"}, aRequires: []v1.Requirement{{Module: "example.com/b", Version: "v2.0.0"}}, wantErr: "does not satisfy example.com/b", wantInstalls: 1},
		{name: "wrong_transitive_commit", requirement: "v1.0.0", locked: []string{"a", "b"}, aRequires: []v1.Requirement{{Module: "example.com/b", Version: versionlessCommitBOld}}, wantErr: "does not satisfy example.com/b", wantInstalls: 1},
		{name: "stale_extra", requirement: "v1.0.0", locked: []string{"a", "b"}, wantErr: "unreachable locked module example.com/b", wantInstalls: 1},
		{name: "dependency_replaces_ignored", requirement: "v1.0.0", locked: []string{"a", "b"}, aRequires: []v1.Requirement{{Module: "example.com/b"}}, aReplaces: []v1.Replacement{{Module: "example.com/b", Target: "./missing"}}, wantInstalls: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			manifest := "// retain comments\nmodule example.com/app\nrequire example.com/a " + tt.requirement + "\n"
			writeV1GenerateFixture(t, root, v1.ModuleFile, manifest)
			lock := v1.Lock{Version: 1}
			cache := &frozenTestRepository{modules: make(map[string]v1.Module), directories: make(map[string]string)}
			var wantRoots SourceRoots
			for _, name := range tt.locked {
				source := "example.com/" + name
				lock.Modules = append(lock.Modules, v1.LockedModule{Source: source, Version: "v1.0.0", Commit: versionlessCommitA, Hash: versionlessHash})
				cache.directories[source] = t.TempDir()
				cache.modules[source] = v1.Module{Name: source, Roots: []string{"."}}
				wantRoots = append(wantRoots, SourceRoot{Path: cache.directories[source], Module: source})
			}
			cache.modules["example.com/a"] = v1.Module{Name: "example.com/a", Roots: []string{"."}, Requires: tt.aRequires, Replaces: tt.aReplaces}
			if _, ok := cache.modules["example.com/b"]; ok {
				cache.modules["example.com/b"] = v1.Module{Name: "example.com/b", Roots: []string{"."}, Requires: tt.bRequires}
			}
			raw, err := yaml.Marshal(lock)
			require.NoError(t, err)
			writeV1GenerateFixture(t, root, v1.LockFile, string(raw))

			roots, err := EnsureFrozenSources(t.Context(), root, cache)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.ErrorContains(t, err, "frozen graph "+root)
				assert.Empty(t, roots)
			} else {
				require.NoError(t, err)
				assert.Equal(t, wantRoots, roots)
				assert.Len(t, cache.cached, len(lock.Modules))
			}
			assert.Equal(t, tt.wantInstalls, cache.installs)
			if tt.wantInstalls > 0 {
				assert.Equal(t, lock.Version, cache.installed.Version)
				assert.Len(t, cache.installed.Modules, tt.wantInstalls)
				if tt.wantErr == "" {
					assert.ElementsMatch(t, lock.Modules, cache.installed.Modules)
				} else {
					assert.Equal(t, []v1.LockedModule{{Source: "example.com/a", Version: "v1.0.0", Commit: versionlessCommitA, Hash: versionlessHash}}, cache.installed.Modules)
				}
			} else {
				assert.Empty(t, cache.cached)
			}
			assert.Zero(t, cache.fetches)
			assertFrozenFile(t, root, v1.ModuleFile, manifest)
			assertFrozenFile(t, root, v1.LockFile, string(raw))
		})
	}
}

func TestEnsureFrozenSourcesStaleEmptyGraph(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.com/app\n")
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{{Source: "example.com/stale", Version: "v1.0.0", Commit: versionlessCommitA, Hash: versionlessHash}}}
	raw, err := yaml.Marshal(lock)
	require.NoError(t, err)
	writeV1GenerateFixture(t, root, v1.LockFile, string(raw))

	_, err = EnsureFrozenSources(t.Context(), root, nil)

	require.ErrorContains(t, err, "unreachable locked module example.com/stale")
	assertFrozenFile(t, root, v1.LockFile, string(raw))
}

func TestEnsureFrozenSourcesCacheErrors(t *testing.T) {
	t.Parallel()
	failure := errors.New("cache verification failed")
	for _, method := range []string{"install", "cached"} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.com/app\nrequire example.com/a v1.0.0\n")
			lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{{Source: "example.com/a", Version: "v1.0.0", Commit: versionlessCommitA, Hash: versionlessHash}}}
			raw, err := yaml.Marshal(lock)
			require.NoError(t, err)
			writeV1GenerateFixture(t, root, v1.LockFile, string(raw))
			cache := &frozenTestRepository{}
			if method == "install" {
				cache.installErr = failure
			} else {
				cache.cachedErr = failure
			}

			_, err = EnsureFrozenSources(t.Context(), root, cache)

			require.ErrorIs(t, err, failure)
			require.ErrorContains(t, err, "frozen graph "+root)
			assert.Zero(t, cache.fetches)
			if method == "install" {
				assert.Empty(t, cache.cached)
			}
			assertFrozenFile(t, root, v1.LockFile, string(raw))
		})
	}
}

func assertFrozenFile(t *testing.T, root, name, want string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, name))
	if want == "" {
		require.ErrorIs(t, err, os.ErrNotExist)
		return
	}
	require.NoError(t, err)
	assert.Equal(t, want, string(raw))
}

type frozenTestRepository struct {
	modules     map[string]v1.Module
	directories map[string]string
	installed   v1.Lock
	installs    int
	cached      []string
	fetches     int
	installErr  error
	cachedErr   error
}

var _ Repository = (*frozenTestRepository)(nil)

func (c *frozenTestRepository) Install(_ context.Context, lock v1.Lock) error {
	c.installs++
	c.installed.Version = lock.Version
	c.installed.Modules = append(c.installed.Modules, lock.Modules...)
	return c.installErr
}

func (c *frozenTestRepository) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	c.cached = append(c.cached, entry.Source)
	return c.directories[entry.Source], c.modules[entry.Source], c.cachedErr
}

func (c *frozenTestRepository) Fetch(_ context.Context, _, _ string) (Fetched, error) {
	c.fetches++
	return Fetched{}, errors.New("frozen must not fetch")
}
