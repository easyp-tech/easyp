package gitmodules

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

func TestBufExcludesAliasSelectionMatchesLocalSources(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		modules  string
		target   string
		imported string
		links    map[string]string
		allowed  bool
	}{
		{
			name: "file target excluded", modules: "  - path: .\n    excludes: [preview]\n", target: "preview/types.proto", imported: "alias.proto",
			links: map[string]string{"alias.proto": "preview/types.proto"},
		},
		{
			name: "non-proto target excluded", modules: "  - path: .\n    excludes: [preview]\n", target: "preview/schema", imported: "alias.proto",
			links: map[string]string{"alias.proto": "preview/schema"},
		},
		{
			name: "logical file excluded", modules: "  - path: .\n    excludes: [preview]\n", target: "types.proto", imported: "preview/alias.proto",
			links: map[string]string{"preview/alias.proto": "../types.proto"},
		},
		{
			name: "directory target excluded", modules: "  - path: .\n    excludes: [preview]\n", target: "preview/types.proto", imported: "api/types.proto",
			links: map[string]string{"api": "preview"},
		},
		{
			name: "nested root target excluded", modules: "  - path: api\n    excludes: [api/preview]\n", target: "api/preview/types.proto", imported: "api/alias.proto",
			links: map[string]string{"api/alias.proto": "preview/types.proto"},
		},
		{
			name: "root alias target excluded", modules: "  - path: api\n    excludes: [api/preview]\n", target: "physical/preview/types.proto", imported: "api/alias.proto",
			links: map[string]string{"api": "physical", "physical/alias.proto": "preview/types.proto"},
		},
		{
			name: "target excluded by its module", modules: "  - path: api\n  - path: physical\n    excludes: [physical/preview]\n", target: "physical/preview/types.proto", imported: "api/alias.proto",
			links: map[string]string{"api/alias.proto": "../physical/preview/types.proto"},
		},
		{
			name: "permitted file outside root", modules: "  - path: api\n    excludes: [api/preview]\n", target: "physical/types.proto", imported: "api/alias.proto", allowed: true,
			links: map[string]string{"api/alias.proto": "../physical/types.proto"},
		},
		{
			name: "permitted non-proto target", modules: "  - path: api\n    excludes: [api/preview]\n", target: "physical/schema", imported: "api/alias.proto", allowed: true,
			links: map[string]string{"api/alias.proto": "../physical/schema"},
		},
		{
			name: "permitted root alias", modules: "  - path: api\n    excludes: [api/preview]\n", target: "physical/types.proto", imported: "api/types.proto", allowed: true,
			links: map[string]string{"api": "physical"},
		},
		{
			name: "unfiltered overlapping module permits target", modules: "  - path: api\n  - path: physical\n    excludes: [physical/preview]\n  - path: physical/preview\n", target: "physical/preview/types.proto", imported: "api/alias.proto", allowed: true,
			links: map[string]string{"api/alias.proto": "../physical/preview/types.proto"},
		},
		{
			name: "unfiltered overlapping aliased root permits target", modules: "  - path: api\n    excludes: [api/preview]\n  - path: api/preview\n", target: "physical/preview/types.proto", imported: "api/preview/types.proto", allowed: true,
			links: map[string]string{"api": "physical"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			content := "syntax = \"proto3\"; package example; message Types {}\n"
			commitRootSelectionRepository(t, repository, map[string]string{"buf.yaml": "version: v2\nmodules:\n" + test.modules, test.target: content}, test.links)
			module, err := moduleconfig.ReadGitDependency(repository, repository)
			require.NoError(t, err)
			local, err := modules.ModuleSources(repository, module)
			require.NoError(t, err)
			assert.Equal(t, test.allowed, local.FileAllowed()(filepath.Join(repository, test.imported)))
			localBytes, localErr := local.ReadSourceFile(filepath.Join(repository, test.imported))
			if test.allowed {
				require.NoError(t, localErr)
				assert.Equal(t, content, string(localBytes))
			}
			if !test.allowed {
				require.ErrorIs(t, localErr, os.ErrNotExist)
			}

			producer := New(t.TempDir())
			fetched, err := producer.FetchForRootResolution(t.Context(), repository, "")
			require.NoError(t, err)
			require.NotNil(t, fetched.Inspection)
			inspectionNames := make([]string, 0, len(fetched.Inspection.Files))
			for _, file := range fetched.Inspection.Files {
				inspectionNames = append(inspectionNames, file.Path)
			}
			assert.Equal(t, test.allowed, slices.Contains(inspectionNames, test.imported), "root inspection must apply the same target selection")

			cold := New(t.TempDir())
			lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
			require.NoError(t, cold.Install(t.Context(), lock))
			installed, installedModule, err := (&Cache{root: cold.root}).Cached(fetched.Lock)
			require.NoError(t, err)
			inventory, err := cold.SourceFiles(fetched.Lock)
			require.NoError(t, err)
			assert.Equal(t, test.allowed, slices.Contains(inventory, test.imported), "materialized aliases must apply the same target selection")
			installedSources, err := modules.ModuleSources(installed, installedModule)
			require.NoError(t, err)
			installedBytes, installedErr := installedSources.ReadSourceFile(filepath.Join(installed, test.imported))
			if test.allowed {
				require.NoError(t, installedErr)
				assert.Equal(t, localBytes, installedBytes)
			}
			if !test.allowed {
				require.ErrorIs(t, installedErr, os.ErrNotExist)
			}
			frozen := &Cache{root: cold.root}
			require.NoError(t, frozen.VerifyCached(t.Context(), lock))
			verified, err := frozen.SourceFiles(fetched.Lock)
			require.NoError(t, err)
			assert.Equal(t, inventory, verified)
		})
	}
}

func TestBufExcludedAliasTargetChangePreservesSnapshotHash(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	commitRootSelectionRepository(t, repository, map[string]string{
		"buf.yaml":                     "version: v2\nmodules:\n  - path: api\n    excludes: [api/preview]\n",
		"physical/types.proto":         "syntax = \"proto3\";\n",
		"physical/preview/types.proto": "syntax = \"proto3\"; package excluded;\n",
	}, map[string]string{"api": "physical", "physical/alias.proto": "preview/types.proto"})
	cache := New(t.TempDir())
	first, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(repository, "physical/preview/types.proto"), []byte("changed excluded body\n"), 0o644))
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "excluded alias target only")
	second, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	assert.NotEqual(t, first.Lock.Commit, second.Lock.Commit)
	assert.Equal(t, first.Lock.Hash, second.Lock.Hash)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{second.Lock}}))
	inventory, err := cache.SourceFiles(second.Lock)
	require.NoError(t, err)
	assert.Equal(t, []string{"api/types.proto", "buf.yaml", "physical/types.proto"}, inventory)
}

func TestBufExcludedDirectoryAliasPrunesCycles(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		excludes string
		cycle    bool
	}{
		{name: "physical directory excluded", excludes: "preview"},
		{name: "selected cycle stays strict", excludes: "unused", cycle: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			primary := "syntax = \"proto3\"; package primary; message Schema {}\n"
			commitRootSelectionRepository(t, repository, map[string]string{
				"buf.yaml":             "version: v1\nbuild:\n  excludes: [" + test.excludes + "]\n",
				"primary/schema.proto": primary,
				"preview/types.proto":  "syntax = \"proto3\"; package preview; message Types {}\n",
			}, map[string]string{"api": "preview", "preview/loop": "."})
			for _, operation := range []struct {
				name string
				read func(*Cache) (modules.Fetched, error)
			}{
				{name: "fetch", read: func(cache *Cache) (modules.Fetched, error) {
					return cache.Fetch(t.Context(), repository, "")
				}},
				{name: "inspection", read: func(cache *Cache) (modules.Fetched, error) {
					return cache.FetchForRootResolution(t.Context(), repository, "")
				}},
			} {
				t.Run(operation.name, func(t *testing.T) {
					t.Parallel()
					fetched, err := operation.read(New(t.TempDir()))
					if test.cycle {
						require.ErrorIs(t, err, sourceview.ErrCycle)
						return
					}
					require.NoError(t, err)
					if fetched.Inspection != nil {
						assert.Empty(t, fetched.Inspection.Problems)
						require.Len(t, fetched.Inspection.Files, 1)
						assert.Equal(t, "primary/schema.proto", fetched.Inspection.Files[0].Path)
					}
					cold := New(t.TempDir())
					require.NoError(t, cold.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
					installed, _, err := cold.Cached(fetched.Lock)
					require.NoError(t, err)
					actual, err := os.ReadFile(filepath.Join(installed, "primary/schema.proto"))
					require.NoError(t, err)
					assert.Equal(t, primary, string(actual))
					inventory, err := cold.SourceFiles(fetched.Lock)
					require.NoError(t, err)
					assert.Equal(t, []string{"buf.yaml", "primary/schema.proto"}, inventory)
				})
			}
		})
	}
}

func TestBufIncludedRootAliasKeepsAncestorDirectories(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	content := "syntax = \"proto3\"; package included; message Types {}\n"
	commitRootSelectionRepository(t, repository, map[string]string{
		"buf.yaml":                      "version: v2\nmodules:\n  - path: api\n    includes: [api/selected]\n",
		"physical/selected/types.proto": content,
		"physical/ignored/types.proto":  "syntax = \"proto3\"; package ignored;\n",
	}, map[string]string{"api": "physical"})
	cache := New(t.TempDir())
	fetched, err := cache.FetchForRootResolution(t.Context(), repository, "")
	require.NoError(t, err)
	cold := New(t.TempDir())
	require.NoError(t, cold.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
	installed, _, err := cold.Cached(fetched.Lock)
	require.NoError(t, err)
	actual, err := os.ReadFile(filepath.Join(installed, "api/selected/types.proto"))
	require.NoError(t, err)
	assert.Equal(t, content, string(actual))
	inventory, err := cold.SourceFiles(fetched.Lock)
	require.NoError(t, err)
	assert.Equal(t, []string{"api/selected/types.proto", "buf.yaml", "physical/selected/types.proto"}, inventory)
}
