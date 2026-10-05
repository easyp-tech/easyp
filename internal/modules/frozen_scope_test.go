package modules

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestFrozenDoesNotAcquireUnreachableLockedModules(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.com/app\nrequire example.com/a v1.0.0\n")
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{
		{Source: "example.com/a", Version: "v1.0.0", Commit: versionlessCommitA, Hash: versionlessHash},
		{Source: "example.com/unreachable", Version: "v1.0.0", Commit: versionlessCommitA, Hash: versionlessHash},
	}}
	raw, err := yaml.Marshal(lock)
	require.NoError(t, err)
	writeV1GenerateFixture(t, root, v1.LockFile, string(raw))
	cache := &frozenScopedCache{dir: t.TempDir()}

	_, err = EnsureFrozenSources(t.Context(), root, cache)

	require.ErrorContains(t, err, "unreachable locked module example.com/unreachable")
	assert.Equal(t, []string{"example.com/a"}, cache.installed)
	assertFrozenFile(t, root, v1.LockFile, string(raw))
}

type frozenScopedCache struct {
	dir       string
	installed []string
}

func (c *frozenScopedCache) Install(_ context.Context, lock v1.Lock) error {
	for _, entry := range lock.Modules {
		c.installed = append(c.installed, entry.Source)
		if entry.Source != "example.com/a" {
			return fmt.Errorf("must not acquire an unrelated pinned repository: %s", entry.Source)
		}
	}
	return nil
}

func (c *frozenScopedCache) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	return c.dir, v1.Module{Name: entry.Source, Roots: []string{"."}}, nil
}
