package modules

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type parentPolicyCache struct {
	entries   map[string]EffectiveModule
	installed []string
	fetched   []string
}

func (c *parentPolicyCache) Install(_ context.Context, lock v1.Lock) error {
	for _, entry := range lock.Modules {
		c.installed = append(c.installed, entry.Source)
		if _, ok := c.entries[entry.Source]; !ok {
			return fmt.Errorf("unexpected acquisition: %s", entry.Source)
		}
	}
	return nil
}
func (c *parentPolicyCache) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	m, ok := c.entries[entry.Source]
	if !ok {
		return "", v1.Module{}, fmt.Errorf("missing cached module %s", entry.Source)
	}
	return m.Directory, m.Module, nil
}
func (c *parentPolicyCache) Fetch(_ context.Context, name, version string) (Fetched, error) {
	c.fetched = append(c.fetched, name+"@"+version)
	return Fetched{}, fmt.Errorf("unexpected revision resolution")
}
func parentPolicyWrite(t *testing.T, path, raw string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(raw), 0o600))
}
func parentPolicyPin(name, version string) v1.LockedModule {
	return v1.LockedModule{Source: name, Version: version, Commit: strings.Repeat("a", 40), Hash: "h1:" + strings.Repeat("A", 43) + "="}
}
func parentPolicyLock(t *testing.T, root string, entries ...v1.LockedModule) {
	t.Helper()
	raw, err := yaml.Marshal(v1.Lock{Version: 1, Modules: entries})
	require.NoError(t, err)
	parentPolicyWrite(t, filepath.Join(root, v1.LockFile), string(raw))
}
func TestPolicyGraphParentDoesNotAcquireUnreachableLockEntry(t *testing.T) {
	t.Parallel()
	root, dir := t.TempDir(), t.TempDir()
	const name = "example.test/policies"
	parentPolicyWrite(t, filepath.Join(root, v1.ModuleFile), "module example.test/client\nrequire "+name+" v1.0.0\n")
	parentPolicyWrite(t, filepath.Join(dir, v1.ModuleFile), "module "+name+"\n")
	parentPolicyLock(t, root, parentPolicyPin(name, "v1.0.0"), parentPolicyPin("example.test/unreachable", "v1.0.0"))
	cache := &parentPolicyCache{entries: map[string]EffectiveModule{name: {Directory: dir, Module: v1.Module{Name: name, Roots: []string{"."}}}}}
	got, err := ResolvePolicyGraph(t.Context(), root, cache, false)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, []string{name}, cache.installed)
	require.Empty(t, cache.fetched)
}
func TestPolicyGraphParentValidatesTransitiveConstraints(t *testing.T) {
	t.Parallel()
	root, a, b := t.TempDir(), t.TempDir(), t.TempDir()
	const name = "example.test/policies"
	const dep = "example.test/base"
	parentPolicyWrite(t, filepath.Join(root, v1.ModuleFile), "module example.test/client\nrequire "+name+" v1.0.0\n")
	parentPolicyWrite(t, filepath.Join(a, v1.ModuleFile), "module "+name+"\nrequire "+dep+" v1.2.0\n")
	parentPolicyWrite(t, filepath.Join(b, v1.ModuleFile), "module "+dep+"\n")
	parentPolicyLock(t, root, parentPolicyPin(name, "v1.0.0"), parentPolicyPin(dep, "v1.0.0"))
	cache := &parentPolicyCache{entries: map[string]EffectiveModule{name: {Directory: a, Module: v1.Module{Name: name, Roots: []string{"."}, Requires: []v1.Requirement{{Module: dep, Version: "v1.2.0"}}}}, dep: {Directory: b, Module: v1.Module{Name: dep, Roots: []string{"."}}}}}
	_, err := ResolvePolicyGraph(t.Context(), root, cache, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), dep)
	require.Empty(t, cache.fetched)
}
func TestPolicyGraphParentDoesNotResolveNewOverlayRevisions(t *testing.T) {
	t.Parallel()
	root, overlay := t.TempDir(), t.TempDir()
	const name = "example.test/policies"
	parentPolicyWrite(t, filepath.Join(root, v1.ModuleFile), "module example.test/client\nrequire "+name+" v1.0.0\nreplace example.test/unused => "+overlay+"\n")
	cache := &parentPolicyCache{entries: map[string]EffectiveModule{}}
	_, err := ResolvePolicyGraph(t.Context(), root, cache, false)
	require.Error(t, err)
	require.Empty(t, cache.fetched, "extends cannot resolve an unpinned revision through the concrete repository")
}
func TestPolicyGraphParentIgnoresUnrelatedInvalidManifest(t *testing.T) {
	t.Parallel()
	root, dir := t.TempDir(), t.TempDir()
	const name = "example.test/repo/policies/v2"
	parentPolicyWrite(t, filepath.Join(root, v1.ModuleFile), "module example.test/client\nrequire "+name+" v2.0.0\n")
	parentPolicyWrite(t, filepath.Join(dir, v1.ModuleFile), "module example.test/unrelated\nrequire (broken\n")
	parentPolicyWrite(t, filepath.Join(dir, "policies", v1.ModuleFile), "module "+name+"\nroots proto\n")
	parentPolicyWrite(t, filepath.Join(dir, "policies/proto/readme.txt"), "empty source root\n")
	parentPolicyLock(t, root, parentPolicyPin(name, "v2.0.0"))
	cache := &parentPolicyCache{entries: map[string]EffectiveModule{name: {Directory: dir, Module: v1.Module{Name: name, Roots: []string{"policies/proto"}}}}}
	got, err := ResolvePolicyGraph(t.Context(), root, cache, false)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "policies"), got[name].Directory)
}
