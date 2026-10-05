package gitmodules

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestFetchBSRDependencies(t *testing.T) {
	t.Parallel()
	backendErr := errors.New("BSR backend unavailable")
	tests := []struct {
		name         string
		deps         string
		version      string
		backendErr   error
		wrongOrigin  bool
		unconfigured bool
		wantCalls    int
		wantErr      string
		wantErrIs    error
	}{
		{name: "plain Git", deps: "[]", version: "v1.2.3"},
		{name: "plain Git without BSR backend", deps: "[]", unconfigured: true},
		{name: "BSR without backend", deps: "[buf.build/googleapis/googleapis]", unconfigured: true, wantErr: "BSR resolver is not configured"},
		{name: "known BSR", deps: "[buf.build/googleapis/googleapis]", version: "v1.2.3", wantCalls: 1},
		{name: "deduplicate declarations", deps: "[buf.build/googleapis/googleapis, buf.build/googleapis/googleapis]", version: "v1.2.3", wantCalls: 1},
		{name: "two BSR modules share one Git source", deps: "[buf.build/googleapis/googleapis, buf.build/acme/common]", version: "v1.2.3", wantCalls: 2},
		{name: "backend failure", deps: "[buf.build/googleapis/googleapis]", backendErr: backendErr, wantCalls: 1, wantErrIs: backendErr},
		{name: "unpinned response", deps: "[buf.build/googleapis/googleapis]", wantCalls: 1, wantErr: "unpinned BSR Git target"},
		{name: "moving response", deps: "[buf.build/googleapis/googleapis]", version: "main", wantCalls: 1, wantErr: "invalid version"},
		{name: "changed origin", deps: "[buf.build/googleapis/googleapis]", version: "v1.2.3", wrongOrigin: true, wantCalls: 1, wantErr: "changed BSR dependency"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			remote := bsrTestRepository(t, "version: v1\ndeps: "+tt.deps+"\n")
			backend := &fakeBSRResolver{git: v1.Requirement{Module: "github.com/acme/provider", Version: tt.version}, err: tt.backendErr, wrongOrigin: tt.wrongOrigin}
			cache := NewWithBSRResolver(t.TempDir(), backend)
			if tt.unconfigured {
				cache.bsrResolver = nil
			}

			fetched, err := cache.Fetch(t.Context(), remote, "v1.0.0")

			assert.Len(t, backend.calls, tt.wantCalls)
			entries, readErr := os.ReadDir(cache.root)
			require.NoError(t, readErr)
			for _, entry := range entries {
				assert.Equal(t, "objects", entry.Name(), "temporary checkout survived Fetch")
			}
			if tt.wantErrIs != nil {
				require.ErrorIs(t, err, tt.wantErrIs)
				return
			}
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, fetched.Lock.BSR, tt.wantCalls)
			if tt.wantCalls == 0 {
				assert.Empty(t, fetched.Module.Requires)
				return
			}
			assert.Equal(t, []v1.Requirement{backend.git}, fetched.Module.Requires)
			for i, binding := range fetched.Lock.BSR {
				assert.Equal(t, backend.calls[i], binding.Dependency)
				assert.Equal(t, backend.git, binding.Git)
			}
			assert.NoError(t, (v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}).Validate())
		})
	}
}

func TestResolveBSRWorkspaceDependencies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		origins     []string
		wantCalls   int
		wantOrigins []string
	}{
		{name: "two workspace origins share one resolution", origins: []string{"a/buf.yaml", "b/buf.yaml"}, wantCalls: 1, wantOrigins: []string{"a/buf.yaml", "b/buf.yaml"}},
		{name: "duplicate origin", origins: []string{"a/buf.yaml", "a/buf.yaml"}, wantCalls: 1, wantOrigins: []string{"a/buf.yaml"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			backend := &fakeBSRResolver{git: v1.Requirement{Module: "github.com/acme/provider", Version: "v1.2.3"}}
			cache := NewWithBSRResolver(t.TempDir(), backend)
			module := v1.Module{Name: "github.com/acme/parent"}
			for _, origin := range tt.origins {
				module.BSRDependencies = append(module.BSRDependencies, v1.BSRDependency{Module: "buf.build/googleapis/googleapis", Config: origin})
			}

			resolved, bindings, err := cache.resolveBSRDependencies(t.Context(), module)

			require.NoError(t, err)
			assert.Len(t, backend.calls, tt.wantCalls)
			assert.Equal(t, []v1.Requirement{backend.git}, resolved.Requires)
			var origins []string
			for _, binding := range bindings {
				origins = append(origins, binding.Dependency.Config)
			}
			assert.Equal(t, tt.wantOrigins, origins)
			assert.Nil(t, module.Requires, "resolution must not mutate the input")
		})
	}
}

func TestResolveBSRDependenciesCancellation(t *testing.T) {
	t.Parallel()
	tests := []struct{ name string }{{name: "cancelled before backend"}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			backend := &fakeBSRResolver{git: v1.Requirement{Module: "github.com/acme/provider", Version: "v1.2.3"}}
			cache := NewWithBSRResolver(t.TempDir(), backend)
			module := v1.Module{BSRDependencies: []v1.BSRDependency{{Module: "buf.build/googleapis/googleapis", Config: "buf.yaml"}}}
			ctx, cancel := context.WithCancel(t.Context())
			cancel()

			_, _, err := cache.resolveBSRDependencies(ctx, module)

			require.ErrorIs(t, err, context.Canceled)
			assert.Empty(t, backend.calls)
		})
	}
}

func TestFetchMigrationBSRBindings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		version string
	}{
		{name: "pinned legacy module", version: "v1.0.0"},
		{name: "initial legacy resolution"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			remote := bsrTestRepository(t, "version: v1\ndeps: [buf.build/googleapis/googleapis]\n")
			backend := &fakeBSRResolver{git: v1.Requirement{Module: "github.com/acme/provider", Version: "v1.2.3"}}
			cache := NewWithBSRResolver(t.TempDir(), backend)

			fetched, err := cache.FetchMigration(t.Context(), remote, tt.version, "")

			require.NoError(t, err)
			assert.Equal(t, []v1.Requirement{backend.git}, fetched.Module.Requires)
			require.Len(t, fetched.Lock.BSR, 1)
			assert.Equal(t, "buf.build/googleapis/googleapis", fetched.Lock.BSR[0].Dependency.Module)
			assert.Equal(t, backend.git, fetched.Lock.BSR[0].Git)
			assert.NoError(t, (v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}).Validate())
		})
	}
}

func TestCachedBSRBindings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		change  string
		wantErr string
	}{
		{name: "replay without backend"},
		{name: "old lock needs regeneration", change: "missing", wantErr: "run easyp mod tidy"},
		{name: "stale reference", change: "reference", wantErr: "does not match"},
		{name: "wrong config origin", change: "config", wantErr: "does not match"},
		{name: "extra origin", change: "extra", wantErr: "does not match"},
		{name: "invalid recorded target", change: "target", wantErr: "unpinned BSR Git target"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			remote := bsrTestRepository(t, "version: v1\ndeps: [buf.build/googleapis/googleapis:stable]\n")
			backend := &fakeBSRResolver{git: v1.Requirement{Module: "github.com/acme/provider", Version: "v1.2.3"}}
			cache := NewWithBSRResolver(t.TempDir(), backend)
			fetched, err := cache.Fetch(t.Context(), remote, "v1.0.0")
			require.NoError(t, err)
			require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
			entry := fetched.Lock
			entry.BSR = slices.Clone(entry.BSR)
			switch tt.change {
			case "missing":
				entry.BSR = nil
			case "reference":
				entry.BSR[0].Dependency.Reference = "next"
			case "config":
				entry.BSR[0].Dependency.Config = "other/buf.yaml"
			case "extra":
				entry.BSR = append(entry.BSR, entry.BSR[0])
			case "target":
				entry.BSR[0].Git.Version = ""
			}
			backend.err = errors.New("backend must never run during Cached")
			backend.calls = nil

			_, module, err := cache.Cached(entry)

			assert.Empty(t, backend.calls)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []v1.Requirement{backend.git}, module.Requires)
		})
	}
}

type fakeBSRResolver struct {
	git         v1.Requirement
	err         error
	wrongOrigin bool
	calls       []v1.BSRDependency
}

func (backend *fakeBSRResolver) Resolve(_ context.Context, dependency v1.BSRDependency) (v1.BSRResolution, error) {
	backend.calls = append(backend.calls, dependency)
	if backend.err != nil {
		return v1.BSRResolution{}, backend.err
	}
	if backend.wrongOrigin {
		dependency.Module = "buf.build/acme/wrong"
	}
	return v1.BSRResolution{Dependency: dependency, Git: backend.git, Resolution: v1.BSRCompatibilitySnapshot}, nil
}

func bsrTestRepository(t *testing.T, config string) string {
	t.Helper()
	remote := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(remote, "buf.yaml"), []byte(config), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "parent.proto"), []byte("syntax = \"proto3\"; message Parent {}\n"), 0o600))
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	runTestGit(t, remote, "tag", "v1.0.0")
	return remote
}
