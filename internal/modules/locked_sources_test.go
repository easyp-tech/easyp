package modules

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestEnsureSourcesExcludesReplacedModuleFromCache(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		bRequires []v1.Requirement
		wantError string
	}{
		{
			name:      "replacement satisfies transitive requirement",
			bRequires: []v1.Requirement{{Module: "example.com/A", Version: "v1.1.0"}},
		},
		{
			name:      "unreplaced transitive requirement must be locked",
			bRequires: []v1.Requirement{{Module: "example.com/C", Version: "v1.0.0"}},
			wantError: "protobuf.lock does not satisfy example.com/C",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			localA := filepath.Join(root, "local-A")
			writeV1GenerateFixture(t, root, "local-A/protobuf.mod", "module example.com/A\n")
			writeV1GenerateFixture(t, root, "local-A/a.proto", "syntax = \"proto3\";\n")
			writeV1GenerateFixture(t, root, "proto/app.proto", "syntax = \"proto3\";\n")
			cacheRoot := t.TempDir()
			cachedA := filepath.Join(cacheRoot, "A")
			cachedB := filepath.Join(cacheRoot, "B")
			writeV1GenerateFixture(t, cachedA, "a.proto", "syntax = \"proto3\";\n")
			writeV1GenerateFixture(t, cachedB, "b.proto", "syntax = \"proto3\";\n")
			module := v1.Module{
				Name: "example.com/app", Roots: []string{"proto"},
				Requires: []v1.Requirement{{Module: "example.com/A", Version: "v1.0.0"}, {Module: "example.com/B", Version: "v1.0.0"}},
				Replaces: []v1.Replacement{{Module: "example.com/A", Target: "local-A"}},
			}
			lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{
				{Source: "example.com/A", Version: "v1.0.0", Commit: versionlessCommitA, Hash: versionlessHash},
				{Source: "example.com/B", Version: "v1.0.0", Commit: versionlessCommitBOld, Hash: versionlessHash},
			}}
			raw, err := yaml.Marshal(lock)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(root, v1.LockFile), raw, 0o600))
			cache := &trackingCache{modules: map[string]v1.Module{
				"example.com/A": {Name: "example.com/A", Roots: []string{"."}},
				"example.com/B": {Name: "example.com/B", Roots: []string{"."}, Requires: tt.bRequires},
			}, directories: map[string]string{"example.com/A": cachedA, "example.com/B": cachedB}}

			roots, err := EnsureSources(t.Context(), root, module, cache)

			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []string{localA, cachedB}, roots.Paths())
			assert.Equal(t, []v1.LockedModule{lock.Modules[1]}, cache.installed.Modules)
			assert.Equal(t, []string{"example.com/B"}, cache.cached)
			assert.NoError(t, CheckImportCollisions(root, module.Roots, roots.Paths()))
		})
	}
}

type trackingCache struct {
	modules     map[string]v1.Module
	directories map[string]string
	installed   v1.Lock
	cached      []string
}

func (c *trackingCache) Install(_ context.Context, lock v1.Lock) error {
	c.installed = lock
	return nil
}

func (c *trackingCache) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	c.cached = append(c.cached, entry.Source)
	return c.directories[entry.Source], c.modules[entry.Source], nil
}
