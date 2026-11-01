package generation

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/workspace"
)

func TestWorkspaceModuleIndexReadsDeclarationsOnce(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, name := range []string{"a", "b"} {
		writeV1GenerateFixture(t, root, name+"/"+v1.ModuleFile, "module example.test/"+name+"\n")
	}
	writeV1GenerateFixture(t, root, "legacy/"+v1.ModuleFile, "direct (\n)\n")
	for _, directory := range []string{".hidden", "easyp_vendor", "node_modules", "nested"} {
		writeV1GenerateFixture(t, root, directory+"/"+v1.ModuleFile, "module example.test/broken\nunknown directive\n")
	}
	writeV1GenerateFixture(t, root, "nested/.git", "gitdir: /unused\n")
	require.NoError(t, os.Symlink("nested", filepath.Join(root, "nested-alias")))
	index := newV1WorkspaceModuleIndex(root)
	walks := 0
	reads := make(map[string]int)
	index.walk = func(boundary string, visit func(string, fs.DirEntry, error) error) error {
		walks++
		return workspace.Walk(boundary, visit)
	}
	index.readFile = func(boundary, path string) ([]byte, error) {
		reads[path]++
		return workspace.ReadFile(boundary, path)
	}
	assert.Zero(t, walks, "index creation must stay lazy")
	for _, name := range []string{"a", "b", "a", "b"} {
		directory, err := index.find("example.test/" + name)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(root, name), directory)
	}
	directory, err := index.find("example.test/missing")
	require.NoError(t, err)
	assert.Empty(t, directory)
	assert.Equal(t, 1, walks)
	assert.Equal(t, map[string]int{
		filepath.Join(root, "a", v1.ModuleFile):      1,
		filepath.Join(root, "b", v1.ModuleFile):      1,
		filepath.Join(root, "legacy", v1.ModuleFile): 1,
	}, reads)
}

func TestModuleSelectionKeepsUnusedIndexLazy(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.test/consumer\n")
	writeV1GenerateFixture(t, root, v1.LockFile, "version: 1\nmodules: []\n")
	writeV1GenerateFixture(t, root, "unrelated/"+v1.ModuleFile, "module example.test/broken\nunknown directive\n")

	selected, err := resolveV1GenerationModule(t.Context(), nil, root, root, v1ModuleSelection{directory: root}, false)
	require.NoError(t, err)
	assert.Equal(t, root, selected.directory)
	selected, err = resolveFrozenGenerationModule(t.Context(), nil, root, root, v1ModuleSelection{source: "example.test/consumer"})
	require.NoError(t, err)
	assert.Equal(t, root, selected.directory)
	selected, err = resolveFrozenGenerationModule(t.Context(), nil, root, root, v1ModuleSelection{directory: root})
	require.NoError(t, err)
	assert.Equal(t, root, selected.directory)

	external := t.TempDir()
	writeV1GenerateFixture(t, external, v1.ModuleFile, "module example.test/external\n")
	writeV1GenerateFixture(t, external, v1.LockFile, "version: 1\nmodules: []\n")
	for _, frozen := range []bool{false, true} {
		selected, err = resolveV1GenerationModule(t.Context(), nil, root, external, v1ModuleSelection{directory: external}, frozen)
		require.NoError(t, err)
		assert.Equal(t, external, selected.directory)
		assert.Equal(t, "example.test/external", selected.module.Name)
	}
}

func TestRunIndexesWorkspaceDeclarationsOnce(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		frozen      bool
		descriptors bool
	}{
		{name: "normal generation"},
		{name: "frozen generation", frozen: true},
		{name: "normal descriptors", descriptors: true},
		{name: "frozen descriptors", frozen: true, descriptors: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, snapshot := t.TempDir(), t.TempDir()
			const dependency = "example.test/shared"
			writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.test/consumer\n")
			writeV1GenerateFixture(t, root, v1.LockFile, "version: 1\nmodules: []\n")
			writeV1GenerateFixture(t, root, v1.GenerateFile, "version: v1\ngenerate:\n  modules: [example.test/a, example.test/b, example.test/a]\n  packages: [a.v1, b.v1]\n")
			for _, name := range []string{"a", "b"} {
				manifest := "module example.test/" + name + "\nroots proto\n"
				if name == "a" {
					manifest += "require " + dependency + " v1.0.0\n"
				}
				writeV1GenerateFixture(t, root, name+"/"+v1.ModuleFile, manifest)
				writeV1GenerateFixture(t, root, name+"/"+v1.LockFile, "version: 1\nmodules: []\n")
				writeV1GenerateFixture(t, root, name+"/proto/"+name+".proto", "syntax = \"proto3\"; package "+name+".v1; message Item {}\n")
			}
			entry := v1.LockedModule{Source: dependency, Version: "v1.0.0", Commit: strings.Repeat("a", 40), Hash: "h1:" + strings.Repeat("A", 43) + "="}
			require.NoError(t, writeTestLock(filepath.Join(root, "a"), v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}}))
			writeV1GenerateFixture(t, snapshot, v1.ModuleFile, "module "+dependency+"\n")
			writeV1GenerateFixture(t, snapshot, "shared.proto", "syntax = \"proto3\"; package shared; message Shared {}\n")
			writeV1GenerateFixture(t, root, "unused/"+v1.ModuleFile, "module example.test/unused\n")
			cache := &workspaceDeclarationMutationCache{
				indexedGenerationCache: &indexedGenerationCache{directory: snapshot, module: v1.Module{Name: dependency, Roots: []string{"."}}, files: []string{v1.ModuleFile, "shared.proto"}},
				manifest:               filepath.Join(root, "unused", v1.ModuleFile),
			}
			request := Request{WorkDir: root, WorkspaceRoot: root, Frozen: tt.frozen}
			if tt.descriptors {
				request.DescriptorSetOut = filepath.Join(root, "all.pb")
			}

			err := Run(t.Context(), logger.NewNop(), cache, request)

			require.NoError(t, err, "later selections must reuse declarations read before dependency preparation")
			require.True(t, cache.changed, "the fixture must change a declaration between target lookups")
			if tt.descriptors {
				assert.ElementsMatch(t, []string{"a.proto", "b.proto"}, descriptorNames(readDescriptorSet(t, request.DescriptorSetOut)))
			}
			err = Run(t.Context(), logger.NewNop(), cache, request)
			require.ErrorContains(t, err, cache.manifest, "a new command must read changed declarations")
		})
	}
}

func TestFindV1LocalModulePreservesDeclarationErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		files     map[string]string
		selection string
		wantDir   string
		wantError string
	}{
		{
			name: "unrelated duplicates are allowed",
			files: map[string]string{
				"a/protobuf.mod": "module example.test/duplicate\n",
				"b/protobuf.mod": "module example.test/duplicate\n",
				"c/protobuf.mod": "module example.test/selected\n",
			},
			selection: "example.test/selected", wantDir: "c",
		},
		{
			name: "requested duplicates precede later invalid declarations",
			files: map[string]string{
				"a/protobuf.mod": "module example.test/selected\n",
				"b/protobuf.mod": "module example.test/selected\n",
				"z/protobuf.mod": "module example.test/broken\nunknown directive\n",
			},
			selection: "example.test/selected", wantDir: "a", wantError: "declared in both",
		},
		{
			name: "invalid declarations precede later duplicates",
			files: map[string]string{
				"a/protobuf.mod": "module example.test/selected\n",
				"b/protobuf.mod": "module example.test/broken\nunknown directive\n",
				"z/protobuf.mod": "module example.test/selected\n",
			},
			selection: "example.test/selected", wantDir: "a", wantError: "unknown directive",
		},
		{
			name: "invalid unrelated native declaration fails lookup",
			files: map[string]string{
				"a/protobuf.mod": "module example.test/broken\nunknown directive\n",
				"b/protobuf.mod": "module example.test/selected\n",
			},
			selection: "example.test/selected", wantError: "unknown directive",
		},
		{
			name: "legacy declarations are ignored",
			files: map[string]string{
				"a/protobuf.mod": "direct (\n)\n",
				"b/protobuf.mod": "module example.test/selected\n",
			},
			selection: "example.test/selected", wantDir: "b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for path, contents := range tt.files {
				writeV1GenerateFixture(t, root, path, contents)
			}

			directory, err := findV1LocalModuleByName(root, tt.selection)

			if tt.wantError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantError)
			}
			wantDirectory := ""
			if tt.wantDir != "" {
				wantDirectory = filepath.Join(root, tt.wantDir)
			}
			assert.Equal(t, wantDirectory, directory)
		})
	}
}

func TestRunWorkspaceIndexKeepsConsumerLocksIndependent(t *testing.T) {
	t.Parallel()
	for _, frozen := range []bool{false, true} {
		t.Run(fmt.Sprintf("frozen=%t", frozen), func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			const dependency = "example.test/shared"
			cache := workspaceVersionCache{}
			locks := make(map[string][]byte)
			for i, project := range []string{"a", "b"} {
				version := fmt.Sprintf("v1.%d.0", i)
				snapshot := t.TempDir()
				writeV1GenerateFixture(t, snapshot, v1.ModuleFile, "module "+dependency+"\n")
				writeV1GenerateFixture(t, snapshot, "shared.proto", fmt.Sprintf("syntax = \"proto3\"; package shared; message Shared { string value%d = 1; }\n", i))
				cache[version] = snapshot
				directory := filepath.Join(root, project)
				writeV1GenerateFixture(t, directory, v1.ModuleFile, "module example.test/"+project+"\nrequire "+dependency+" "+version+"\n")
				writeV1GenerateFixture(t, directory, v1.GenerateFile, "version: v1\ngenerate:\n  modules: ["+dependency+"]\n")
				entry := v1.LockedModule{Source: dependency, Version: version, Commit: strings.Repeat(fmt.Sprint(i+1), 40), Hash: "h1:" + strings.Repeat("A", 43) + "="}
				require.NoError(t, writeTestLock(directory, v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}}))
				var err error
				locks[project], err = os.ReadFile(filepath.Join(directory, v1.LockFile))
				require.NoError(t, err)
			}
			request := Request{WorkDir: root, WorkspaceRoot: root, AllProjects: true, Frozen: frozen, DescriptorSetOutDir: filepath.Join(root, "sets")}

			require.NoError(t, Run(t.Context(), logger.NewNop(), cache, request))

			for i, project := range []string{"a", "b"} {
				outputs, err := filepath.Glob(filepath.Join(request.DescriptorSetOutDir, project, "*.pb"))
				require.NoError(t, err)
				require.Len(t, outputs, 1)
				descriptors := readDescriptorSet(t, outputs[0])
				require.Len(t, descriptors.File, 1)
				assert.Equal(t, fmt.Sprintf("value%d", i), descriptors.File[0].MessageType[0].Field[0].GetName())
				lock, err := os.ReadFile(filepath.Join(root, project, v1.LockFile))
				require.NoError(t, err)
				assert.Equal(t, locks[project], lock)
			}
		})
	}
}

type workspaceVersionCache map[string]string

func (workspaceVersionCache) Install(context.Context, v1.Lock) error { return nil }

func (cache workspaceVersionCache) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	directory, exists := cache[entry.Version]
	if !exists {
		return "", v1.Module{}, fmt.Errorf("unknown cached version %s", entry.Version)
	}
	_, module, err := modules.ReadManifest(directory)
	if err != nil {
		return "", v1.Module{}, fmt.Errorf("ReadManifest: %w", err)
	}
	if module.Name != entry.Source {
		return "", v1.Module{}, fmt.Errorf("unknown cached module %s", entry.Source)
	}
	return directory, module, nil
}

type workspaceDeclarationMutationCache struct {
	*indexedGenerationCache
	manifest string
	changed  bool
}

func (cache *workspaceDeclarationMutationCache) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	directory, module, err := cache.indexedGenerationCache.Cached(entry)
	if err != nil {
		return "", v1.Module{}, fmt.Errorf("Cached: %w", err)
	}
	if !cache.changed {
		err = os.WriteFile(cache.manifest, []byte("module example.test/unused\nunknown directive\n"), 0o600)
		if err != nil {
			return "", v1.Module{}, fmt.Errorf("WriteFile: %w", err)
		}
		cache.changed = true
	}
	return directory, module, nil
}
