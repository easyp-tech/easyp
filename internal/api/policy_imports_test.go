package api

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestPolicyImportRoots(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		sourceRoot     string
		dependencyRoot string
	}{
		{name: "separate source roots", sourceRoot: "proto", dependencyRoot: "src"},
		{name: "module directory roots", sourceRoot: ".", dependencyRoot: "."},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, dependency := t.TempDir(), t.TempDir()
			replacement, err := filepath.Rel(root, dependency)
			require.NoError(t, err)
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/root\nroots "+tt.sourceRoot+"\nrequire example.com/dep v1.0.0\nreplace example.com/dep => "+replacement+"\n")
			writeV1GenerateFixture(t, root, filepath.Join(tt.sourceRoot, "root.proto"), `syntax = "proto3";`)
			writeV1GenerateFixture(t, dependency, "protobuf.mod", "module example.com/dep\nroots "+tt.dependencyRoot+"\n")
			writeV1GenerateFixture(t, dependency, filepath.Join(tt.dependencyRoot, "dep.proto"), `syntax = "proto3";`)

			roots, err := ensureV1PolicyImportRoots(t.Context(), nil, root)

			require.NoError(t, err)
			assert.Equal(t, []string{filepath.Join(root, tt.sourceRoot), filepath.Join(dependency, tt.dependencyRoot)}, roots.Paths())
		})
	}
}

func TestReplacementOwnershipThroughLockedDependencyIsCached(t *testing.T) {
	t.Parallel()
	root, remote := t.TempDir(), t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/client\nrequire example.com/remote v1.0.0\nreplace example.com/dep => ./fork\n")
	writeV1GenerateFixture(t, root, "fork/protobuf.mod", "module example.com/dep\n")
	writeV1GenerateFixture(t, root, "fork/first.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, root, "fork/second.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, remote, "protobuf.mod", "module example.com/remote\nrequire example.com/dep\n")
	pin := v1.LockedModule{Source: "example.com/remote", Version: "v1.0.0", Commit: strings.Repeat("a", 40), Hash: "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}
	lock, err := yaml.Marshal(v1.Lock{Version: 1, Modules: []v1.LockedModule{pin}})
	require.NoError(t, err)
	writeV1GenerateFixture(t, root, "protobuf.lock", string(lock))
	_, module, err := modules.ReadManifest(remote)
	require.NoError(t, err)
	cache := &policyOwnershipCache{directory: remote, module: module}
	replacements := newPolicyReplacementSources(t.Context(), root, root, func() (modules.Cache, error) { return cache, nil }, false)
	for _, path := range []string{"fork/first.proto", "fork/second.proto"} {
		excluded, err := replacements.isUnselectedReplacementSource(root, filepath.Join(root, path))
		require.NoError(t, err)
		assert.True(t, excluded, path)
	}
	assert.Equal(t, 1, cache.installs, "one verified graph supplies ownership for both files")
	unchanged, err := os.ReadFile(filepath.Join(root, "protobuf.lock"))
	require.NoError(t, err)
	assert.Equal(t, lock, unchanged)
}

type policyOwnershipCache struct {
	directory string
	module    v1.Module
	installs  int
}

func (c *policyOwnershipCache) Install(context.Context, v1.Lock) error {
	c.installs++
	return nil
}

func (c *policyOwnershipCache) Cached(v1.LockedModule) (string, v1.Module, error) {
	return c.directory, c.module, nil
}

func TestFrozenReplacementOwnershipRejectsBeforeAcquisition(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/client\nroots api\nreplace example.com/unused => ./api\n")
	writeV1GenerateFixture(t, root, "api/item.proto", "syntax = \"proto3\";\n")
	replacements := newPolicyReplacementSources(t.Context(), root, root, func() (modules.Cache, error) {
		t.Fatal("frozen replacement rejection acquired a cache")
		return nil, nil
	}, true)
	excluded, err := replacements.isUnselectedReplacementSource(root, filepath.Join(root, "api/item.proto"))
	require.ErrorContains(t, err, "frozen mode rejects replace")
	assert.False(t, excluded)
}

func TestFindPolicyModule(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		manifest      string
		wantDirectory string
	}{
		{name: "no module"},
		{name: "repository module", manifest: "protobuf.mod", wantDirectory: "."},
		{name: "nearest module", manifest: "api/protobuf.mod", wantDirectory: "api"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			if tt.manifest != "" {
				writeV1GenerateFixture(t, root, tt.manifest, "module example.com/root\n")
			}
			want := ""
			if tt.wantDirectory != "" {
				want = filepath.Join(root, tt.wantDirectory)
			}

			moduleDir, err := findV1PolicyModuleDir(root, filepath.Join(root, "api", "proto"))

			require.NoError(t, err)
			assert.Equal(t, want, moduleDir)
		})
	}
}

func TestUnselectedReplacementSources(t *testing.T) {
	t.Parallel()
	const consumer = "module example.com/client\nrequire example.com/dep\nreplace example.com/dep => ./fork\n"
	tests := []struct {
		name          string
		scanPath      string
		sourcePath    string
		files         map[string]string
		wantExcluded  bool
		errorContains string
	}{
		{
			name: "replacement without manifest", scanPath: ".", sourcePath: "fork/api/item.proto",
			files: map[string]string{"protobuf.mod": consumer}, wantExcluded: true,
		},
		{
			name: "legacy replacement manifest", scanPath: ".", sourcePath: "fork/item.proto",
			files: map[string]string{"protobuf.mod": consumer, "fork/protobuf.mod": "direct (\n)\n"}, wantExcluded: true,
		},
		{
			name: "nested native replacement module", scanPath: ".", sourcePath: "fork/nested/item.proto",
			files: map[string]string{
				"protobuf.mod":             consumer,
				"fork/protobuf.mod":        "module example.com/dep\n",
				"fork/nested/protobuf.mod": "module example.com/nested\n",
			}, wantExcluded: true,
		},
		{
			name: "explicit replacement directory", scanPath: "fork", sourcePath: "fork/api/item.proto",
			files: map[string]string{"protobuf.mod": consumer},
		},
		{
			name: "explicit replacement file", scanPath: "fork/api/item.proto", sourcePath: "fork/api/item.proto",
			files: map[string]string{"protobuf.mod": consumer},
		},
		{
			name: "explicit replacement subtree", scanPath: "fork/api", sourcePath: "fork/api/item.proto",
			files: map[string]string{"protobuf.mod": consumer},
		},
		{
			name: "similar sibling directory", scanPath: ".", sourcePath: "fork-other/item.proto",
			files: map[string]string{"protobuf.mod": "module example.com/client\nreplace example.com/unused => ./fork\n"},
		},
		{
			name: "independent workspace module", scanPath: ".", sourcePath: "other/item.proto",
			files: map[string]string{"protobuf.mod": "module example.com/client\nreplace example.com/unused => ./fork\n", "other/protobuf.mod": "module example.com/other\n"},
		},
		{
			name: "selected independent module does not resolve parent graph", scanPath: "other", sourcePath: "other/item.proto",
			files: map[string]string{
				"protobuf.mod":       "module example.com/client\nrequire example.com/missing\nreplace example.com/missing => ./absent\n",
				"other/protobuf.mod": "module example.com/other\n",
			},
		},
		{
			name: "explicit replacement ignores unrelated parent graph", scanPath: "fork", sourcePath: "fork/item.proto",
			files: map[string]string{
				"protobuf.mod":      "module example.com/client\nrequire example.com/missing\nreplace example.com/missing => ./absent\nreplace example.com/dep => ./fork\n",
				"fork/protobuf.mod": "module example.com/dep\n",
			},
		},
		{
			name: "selected replacement has its own replacement", scanPath: "fork", sourcePath: "fork/nested/item.proto",
			files: map[string]string{
				"protobuf.mod":      "module example.com/client\nreplace example.com/dep => ./fork\n",
				"fork/protobuf.mod": "module example.com/dep\nrequire example.com/nested\nreplace example.com/nested => ./nested\n",
			},
			wantExcluded: true,
		},
		{
			name: "unused replacement overlaps consumer sources", scanPath: ".", sourcePath: "api/item.proto",
			files: map[string]string{"protobuf.mod": "module example.com/client\nroots api\nreplace example.com/unused => ./api\n"},
		},
		{
			name: "unused replacement metadata is not opened", scanPath: ".", sourcePath: "fork/item.proto",
			files: map[string]string{
				"protobuf.mod":      "module example.com/client\nreplace example.com/unused => ./fork\n",
				"fork/protobuf.mod": "legacy metadata\n",
			},
		},
		{
			name: "transitively required replacement", scanPath: ".", sourcePath: "nested/item.proto",
			files: map[string]string{
				"protobuf.mod":      consumer + "replace example.com/nested => ./nested\n",
				"fork/protobuf.mod": "module example.com/dep\nrequire example.com/nested\n",
			},
			wantExcluded: true,
		},
		{
			name: "no consuming module", scanPath: ".", sourcePath: "fork/item.proto",
		},
		{
			name: "invalid consuming manifest", scanPath: ".", sourcePath: "api/item.proto",
			files: map[string]string{"protobuf.mod": "module example.com/client\nroots ../outside\n"}, errorContains: "ParseModule",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for path, content := range tt.files {
				writeV1GenerateFixture(t, root, path, content)
			}
			writeV1GenerateFixture(t, root, tt.sourcePath, "syntax = \"proto3\";\n")

			replacements := newPolicyReplacementSources(t.Context(), root, root, nil, false)
			excluded, err := replacements.isUnselectedReplacementSource(filepath.Join(root, tt.scanPath), filepath.Join(root, tt.sourcePath))

			if tt.errorContains != "" {
				require.ErrorContains(t, err, tt.errorContains)
				assert.False(t, excluded)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantExcluded, excluded)
		})
	}
}

func TestPolicyPathContains(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		directory string
		path      string
		want      bool
	}{
		{name: "exact directory", directory: "fork", path: "fork", want: true},
		{name: "source file", directory: "fork", path: "fork/item.proto", want: true},
		{name: "file is not a directory", directory: "fork/item.proto", path: "fork/item.proto"},
		{name: "similar sibling", directory: "fork", path: "fork-other/item.proto"},
		{name: "aliased directory", directory: "alias", path: "fork/item.proto", want: true},
		{name: "aliased source", directory: "fork", path: "alias/item.proto", want: true},
		{name: "missing aliased source", directory: "fork", path: "alias/deleted/item.proto", want: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "fork/item.proto", "syntax = \"proto3\";\n")
			require.NoError(t, os.Symlink(filepath.Join(root, "fork"), filepath.Join(root, "alias")))

			contains, err := policyPathContains(filepath.Join(root, tt.directory), filepath.Join(root, tt.path))

			require.NoError(t, err)
			assert.Equal(t, tt.want, contains)
		})
	}
}
