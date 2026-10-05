package modules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestEffectiveGraphRemoteRequirements(t *testing.T) {
	t.Parallel()
	const remote = "example.com/remote"
	old := v1.LockedModule{Source: remote, Version: "v1.0.0", Commit: versionlessCommitA, Hash: versionlessHash}
	newer := v1.LockedModule{Source: remote, Version: "v1.1.0", Commit: versionlessCommitBOld, Hash: versionlessHash}
	for _, tt := range []struct {
		name, version     string
		pin               *v1.LockedModule
		commitRequirement bool
		historical        bool
		wantCalls         []string
		wantError         string
	}{
		{name: "matching tag uses verified published pin", version: "v1.0.0", pin: &old},
		{name: "versionless uses published pin", pin: &old},
		{name: "new fork remote dependency without lock", version: "v1.0.0", wantCalls: []string{remote + "@v1.0.0"}},
		{name: "explicit tag and commit are not constrained by higher old pin", version: "v1.0.0", pin: &newer, commitRequirement: true, wantCalls: []string{remote + "@v1.0.0"}},
		{name: "baseline cannot invent historical HEAD", historical: true, wantError: "cannot substitute current HEAD"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, cached := t.TempDir(), t.TempDir()
			writeV1GenerateFixture(t, root, "fork/protobuf.mod", "module example.com/local\nrequire "+remote+" "+tt.version+"\n")
			module := v1.Module{Name: "example.com/app", Requires: []v1.Requirement{{Module: "example.com/local"}}, Replaces: []v1.Replacement{{Module: "example.com/local", Target: "fork"}}}
			if tt.commitRequirement {
				module.Requires = append(module.Requires, v1.Requirement{Module: remote, Version: old.Commit})
			}
			var lock []byte
			if tt.pin != nil {
				var err error
				lock, err = yaml.Marshal(v1.Lock{Version: 1, Modules: []v1.LockedModule{*tt.pin}})
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(filepath.Join(root, v1.LockFile), lock, 0o600))
			}
			dep := v1.Module{Name: remote, Roots: []string{"."}}
			source := &fakeSource{revisions: map[string]Fetched{remote + "@v1.0.0": {Module: dep, Lock: old}, remote + "@" + old.Commit: {Module: dep, Lock: v1.LockedModule{Source: remote, Version: old.Commit, Commit: old.Commit, Hash: old.Hash}}}}
			repository := &overlayTestRepository{fakeSource: source, trackingCache: &trackingCache{modules: map[string]v1.Module{remote: dep}, directories: map[string]string{remote: cached}}}
			var localPath func(string) (string, error)
			if tt.historical {
				localPath = func(target string) (string, error) { return filepath.Join(root, target), nil }
			}
			graph, err := EnsureEffectiveGraph(t.Context(), root, module, repository, localPath)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
			} else {
				require.NoError(t, err)
				assert.Len(t, graph.Modules, 2)
				assert.Equal(t, []string{filepath.Join(root, "fork"), cached}, graph.Sources.Paths())
				assert.Len(t, repository.installed.Modules, 1)
				assert.Equal(t, remote, repository.installed.Modules[0].Source)
			}
			// Each distinct revision is loaded at most once. An explicit commit
			// may be read before the equivalent tag, as in the published resolver.
			if tt.commitRequirement {
				assert.Contains(t, source.calls, remote+"@v1.0.0")
			} else {
				assert.Equal(t, tt.wantCalls, source.calls)
			}
			after, err := os.ReadFile(filepath.Join(root, v1.LockFile))
			if tt.pin == nil {
				assert.ErrorIs(t, err, os.ErrNotExist)
			} else {
				require.NoError(t, err)
				assert.Equal(t, lock, after)
			}
		})
	}
}

type overlayTestRepository struct {
	*fakeSource
	*trackingCache
}

func TestEffectiveGraphRejectsContradictoryLocalPaths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "dep/a.proto", "syntax = \"proto3\";\n")
	module := v1.Module{Name: "example.com/app", Requires: []v1.Requirement{{Module: "example.com/A"}, {Module: "example.com/B"}}, Replaces: []v1.Replacement{{Module: "example.com/A", Target: "dep"}, {Module: "example.com/B", Target: "dep"}}}
	_, err := EnsureEffectiveGraph(t.Context(), root, module, nil, nil)
	require.ErrorContains(t, err, "directory already supplies module")
}
