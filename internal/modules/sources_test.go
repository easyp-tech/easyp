package modules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestModuleSourcesPreserveBufSelection(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		filters []v1.ProtoFileFilter
		want    map[string]string
	}{
		{name: "excludes", filters: []v1.ProtoFileFilter{{Root: "proto", Excludes: []string{"proto/excluded"}}}, want: map[string]string{"allowed/a.proto": "example.test/dep", "allowed_test/b.proto": "example.test/dep"}},
		{name: "includes", filters: []v1.ProtoFileFilter{{Root: "proto", Includes: []string{"proto/allowed"}}}, want: map[string]string{"allowed/a.proto": "example.test/dep"}},
		{name: "union_of_modules", filters: []v1.ProtoFileFilter{{Root: "proto", Includes: []string{"proto/allowed"}, Excludes: []string{"proto/allowed/a.proto"}}, {Root: "proto", Includes: []string{"proto/allowed_test"}}}, want: map[string]string{"allowed_test/b.proto": "example.test/dep"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for _, name := range []string{"proto/allowed/a.proto", "proto/allowed_test/b.proto", "proto/excluded/c.proto"} {
				writeV1GenerateFixture(t, root, name, "syntax = \"proto3\";\n")
			}
			sources, err := ModuleSources(root, v1.Module{Name: "example.test/dep", Roots: []string{"proto"}, ProtoFilters: tt.filters})
			require.NoError(t, err)
			owners, err := sources.FileModules()
			require.NoError(t, err)
			assert.Equal(t, tt.want, owners)
			require.NoError(t, writeV1Vendor(root, sources))
			for _, name := range []string{"allowed/a.proto", "allowed_test/b.proto", "excluded/c.proto"} {
				_, err := os.Stat(filepath.Join(root, VendorDir, name))
				if _, selected := tt.want[name]; selected {
					assert.NoError(t, err, name)
				} else {
					assert.ErrorIs(t, err, os.ErrNotExist, name)
				}
			}
		})
	}
}

func TestLocalBufSelectionValidatesImports(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		metadata string
	}{
		{name: "buf_v1_excludes", metadata: "version: v1\nbuild:\n  excludes: [excluded]\n"},
		{name: "buf_v2_excludes", metadata: "version: v2\nmodules:\n  - path: .\n    excludes: [excluded]\n"},
		{name: "buf_v2_includes", metadata: "version: v2\nmodules:\n  - path: .\n    includes: [allowed]\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			manifest := "module example.test/client\nroots api\nrequire example.test/dep\nreplace example.test/dep => ./dep\n"
			writeV1GenerateFixture(t, root, v1.ModuleFile, manifest)
			writeV1GenerateFixture(t, root, "dep/buf.yaml", tt.metadata)
			writeV1GenerateFixture(t, root, "dep/allowed/a.proto", "syntax = \"proto3\";\n")
			writeV1GenerateFixture(t, root, "dep/excluded/c.proto", "syntax = \"proto3\";\n")
			writeV1GenerateFixture(t, root, "api/client.proto", "syntax = \"proto3\"; import \"excluded/c.proto\";\n")
			err := Tidy(t.Context(), root, nil)
			require.ErrorContains(t, err, "cannot resolve imports")
			assert.ErrorContains(t, err, "excluded/c.proto")
			after, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
			require.NoError(t, err)
			assert.Equal(t, manifest, string(after))
			_, err = os.Stat(filepath.Join(root, v1.LockFile))
			assert.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func TestBufSelectionDoesNotHideConsumerOrCollide(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.test/app\nroots api\nrequire example.test/dep\nreplace example.test/dep => ./dep\n")
	writeV1GenerateFixture(t, root, "dep/buf.yaml", "version: v2\nmodules:\n  - path: .\n    includes: [selected]\n")
	writeV1GenerateFixture(t, root, "dep/selected/a.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, root, "dep/excluded/b.proto", "this excluded source must not be read")
	writeV1GenerateFixture(t, root, "api/excluded/b.proto", "syntax = \"proto3\";\n")
	require.NoError(t, Tidy(t.Context(), root, nil))
	_, module, err := ReadManifest(root)
	require.NoError(t, err)
	deps, err := LocalSources(root, module)
	require.NoError(t, err)
	allowed := deps.FileAllowed()
	require.NotNil(t, allowed)
	assert.True(t, allowed(filepath.Join(root, "api/excluded/b.proto")))
	assert.False(t, allowed(filepath.Join(root, "dep/excluded/b.proto")))
	assert.True(t, allowed(filepath.Join(root, "dep/selected/a.proto")))
	own, err := ModuleSources(root, v1.Module{Name: "example.test/app", Roots: []string{"."}})
	require.NoError(t, err)
	combinedAllowed := append(own, deps...).FileAllowed()
	assert.False(t, combinedAllowed(filepath.Join(root, "dep/excluded/b.proto")))
	assert.True(t, combinedAllowed(filepath.Join(root, "api/excluded/b.proto")))
	directoryAlias := filepath.Join(root, "alias")
	require.NoError(t, os.Symlink(filepath.Join(root, "dep/excluded"), directoryAlias))
	assert.False(t, combinedAllowed(filepath.Join(directoryAlias, "b.proto")))
	writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.test/app\nroots .\nrequire example.test/dep\nreplace example.test/dep => ./dep\n")
	writeV1GenerateFixture(t, root, "client.proto", "syntax = \"proto3\"; import \"alias/b.proto\";\n")
	require.ErrorContains(t, Tidy(t.Context(), root, nil), "cannot resolve imports")

	alias := filepath.Join(t.TempDir(), "alias")
	require.NoError(t, os.Symlink(root, alias))
	aliasDeps, err := LocalSources(alias, module)
	require.NoError(t, err)
	aliasAllowed := aliasDeps.FileAllowed()
	canonical, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)
	assert.False(t, aliasAllowed(filepath.Join(canonical, "dep/excluded/b.proto")))
	assert.True(t, aliasAllowed(filepath.Join(alias, "dep/selected/a.proto")))
}

func TestBufFileAliasCannotBecomeConsumerInput(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.test/app\nroots .\nrequire example.test/dep\nreplace example.test/dep => ./dep\n")
	writeV1GenerateFixture(t, root, "dep/buf.yaml", "version: v1\nbuild:\n  excludes: [excluded]\n")
	writeV1GenerateFixture(t, root, "dep/excluded/b.proto", "this excluded source must not be parsed")
	writeV1GenerateFixture(t, root, "client.proto", "syntax = \"proto3\"; import \"alias.proto\";\n")
	require.NoError(t, os.Symlink(filepath.Join(root, "dep/excluded/b.proto"), filepath.Join(root, "alias.proto")))
	require.ErrorContains(t, Tidy(t.Context(), root, nil), "cannot resolve imports")
	_, module, err := ReadManifest(root)
	require.NoError(t, err)
	own, err := ModuleSources(root, module)
	require.NoError(t, err)
	deps, err := LocalSources(root, module)
	require.NoError(t, err)
	owners, err := append(own, deps...).FileModules()
	require.NoError(t, err)
	assert.NotContains(t, owners, "alias.proto")
}

func TestBufSelectionHonorsLexicalAndPhysicalAliases(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, alias, target, imported string
		directory, outside, other     bool
	}{
		{name: "excluded_file_to_allowed", alias: "excluded/alias.proto", target: "allowed/model.proto", imported: "excluded/alias.proto"},
		{name: "excluded_directory_to_allowed", alias: "excluded/alias", target: "allowed", imported: "excluded/alias/model.proto", directory: true},
		{name: "excluded_file_to_outside", alias: "excluded/alias.proto", imported: "excluded/alias.proto", outside: true},
		{name: "excluded_directory_to_outside", alias: "excluded/alias", imported: "excluded/alias/model.proto", directory: true, outside: true},
		{name: "allowed_file_to_excluded", alias: "allowed/alias.proto", target: "excluded/model.proto", imported: "allowed/alias.proto"},
		{name: "allowed_directory_to_excluded", alias: "allowed/alias", target: "excluded", imported: "allowed/alias/model.proto", directory: true},
		{name: "excluded_file_to_other_module", alias: "excluded/alias.proto", target: "allowed/model.proto", imported: "excluded/alias.proto", other: true},
		{name: "excluded_directory_to_other_module", alias: "excluded/alias", target: "allowed", imported: "excluded/alias/model.proto", directory: true, other: true},
	} {
		for _, consumerRoot := range []string{"api", "."} {
			t.Run(tt.name+"/roots_"+consumerRoot, func(t *testing.T) {
				t.Parallel()
				root := t.TempDir()
				manifest := "module example.test/app\nroots " + consumerRoot + "\nrequire example.test/dep\nreplace example.test/dep => ./dep\n"
				if tt.other {
					manifest += "require example.test/other\nreplace example.test/other => ./other\n"
					writeV1GenerateFixture(t, root, "other/buf.yaml", "version: v2\nmodules:\n  - path: .\n    includes: [allowed]\n")
				}
				writeV1GenerateFixture(t, root, v1.ModuleFile, manifest)
				writeV1GenerateFixture(t, root, "dep/buf.yaml", "version: v1\nbuild:\n  excludes: [excluded]\n")
				writeV1GenerateFixture(t, root, filepath.Join(consumerRoot, "client.proto"), "syntax = \"proto3\"; import \""+tt.imported+"\";\n")
				dep := filepath.Join(root, "dep")
				target := filepath.Join(dep, tt.target)
				if tt.other {
					target = filepath.Join(root, "other", tt.target)
				}
				if tt.outside {
					target = t.TempDir()
					if !tt.directory {
						target = filepath.Join(target, "model.proto")
					}
				}
				targetFile := target
				if tt.directory {
					targetFile = filepath.Join(target, "model.proto")
				}
				writeV1GenerateFixture(t, filepath.Dir(targetFile), filepath.Base(targetFile), "syntax = \"proto3\";\n")
				alias := filepath.Join(dep, tt.alias)
				require.NoError(t, os.MkdirAll(filepath.Dir(alias), 0o755))
				require.NoError(t, os.Symlink(target, alias))
				_, module, err := ReadManifest(root)
				require.NoError(t, err)
				deps, err := LocalSources(root, module)
				require.NoError(t, err)
				assert.False(t, deps.FileAllowed()(filepath.Join(dep, tt.imported)), "both alias spelling and physical target must be selected")
				own, err := ModuleSources(root, module)
				require.NoError(t, err)
				assert.False(t, append(own, deps...).FileAllowed()(filepath.Join(dep, tt.imported)), "a broad consumer or another dependency must not override the alias's selection")
				require.ErrorContains(t, Tidy(t.Context(), root, nil), "cannot resolve imports")
			})
		}
	}
}

func TestBufExcludedFileAliasDoesNotCollideWithConsumerSource(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.test/app\nroots api\nrequire example.test/dep\nreplace example.test/dep => ./dep\n")
	writeV1GenerateFixture(t, root, "dep/buf.yaml", "version: v1\nbuild:\n  excludes: [excluded]\n")
	writeV1GenerateFixture(t, root, "dep/allowed/model.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, root, "api/excluded/model.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, root, "api/client.proto", "syntax = \"proto3\"; import \"excluded/model.proto\";\n")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "dep/excluded"), 0o755))
	require.NoError(t, os.Symlink("../allowed/model.proto", filepath.Join(root, "dep/excluded/model.proto")))
	require.NoError(t, Tidy(t.Context(), root, nil))
	require.NoError(t, Vendor(t.Context(), root, nil))
	_, err := os.Stat(filepath.Join(root, VendorDir, "allowed/model.proto"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(root, VendorDir, "excluded/model.proto"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestBufAliasSelectionPreservesModuleUnion(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, alias, target string }{
		{name: "lexical_selected_by_second_module", alias: "excluded/alias.proto", target: "allowed/model.proto"},
		{name: "physical_selected_by_second_module", alias: "allowed/alias.proto", target: "excluded/model.proto"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.test/app\nroots api\nrequire example.test/dep\nreplace example.test/dep => ./dep\n")
			writeV1GenerateFixture(t, root, "dep/buf.yaml", "version: v2\nmodules:\n  - path: .\n    excludes: [excluded]\n  - path: .\n    includes: [excluded]\n")
			writeV1GenerateFixture(t, root, "dep/"+tt.target, "syntax = \"proto3\";\n")
			writeV1GenerateFixture(t, root, "api/client.proto", "syntax = \"proto3\"; import \""+tt.alias+"\";\n")
			alias := filepath.Join(root, "dep", tt.alias)
			require.NoError(t, os.MkdirAll(filepath.Dir(alias), 0o755))
			require.NoError(t, os.Symlink(filepath.Join(root, "dep", tt.target), alias))
			require.NoError(t, Tidy(t.Context(), root, nil))
		})
	}
}

func TestBufSelectionAllowsConsumerAliasToSelectedDependency(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.test/app\nroots .\nrequire example.test/dep\nreplace example.test/dep => ./dep\n")
	writeV1GenerateFixture(t, root, "dep/buf.yaml", "version: v2\nmodules:\n  - path: .\n    includes: [selected]\n")
	writeV1GenerateFixture(t, root, "dep/selected/model.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, root, "client.proto", "syntax = \"proto3\"; import \"alias.proto\";\n")
	require.NoError(t, os.Symlink("dep/selected/model.proto", filepath.Join(root, "alias.proto")))
	_, module, err := ReadManifest(root)
	require.NoError(t, err)
	deps, err := LocalSources(root, module)
	require.NoError(t, err)
	own, err := ModuleSources(root, module)
	require.NoError(t, err)
	assert.True(t, append(own, deps...).FileAllowed()(filepath.Join(root, "alias.proto")))
	require.NoError(t, Tidy(t.Context(), root, nil))
}
