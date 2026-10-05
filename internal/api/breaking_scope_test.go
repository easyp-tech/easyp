package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	gitadapter "github.com/easyp-tech/easyp/internal/adapters/go_git"
)

func TestBreakingOverlayUsesSnapshotTransitively(t *testing.T) {
	t.Parallel()
	current, baseline := t.TempDir(), t.TempDir()
	writeV1GenerateFixture(t, current, "b/protobuf.mod", "module example.com/wrong\n")
	writeV1GenerateFixture(t, baseline, "protobuf.mod", "module example.com/app\nroots proto\nrequire example.com/A\nreplace example.com/A => a\nreplace example.com/B => "+filepath.Join(current, "b")+"\n")
	writeV1GenerateFixture(t, baseline, "proto/app.proto", "syntax = \"proto3\";\n")
	writeV1GenerateFixture(t, baseline, "a/protobuf.mod", "module example.com/A\nrequire example.com/B\nreplace example.com/B => nonexistent\n")
	writeV1GenerateFixture(t, baseline, "b/protobuf.mod", "module example.com/B\n")
	snapshot := &gitadapter.Snapshot{Root: baseline, RepositoryRoot: current}
	roots, err := breakingImportRoots(t.Context(), gitmodules.New(t.TempDir()), baseline, ".", []string{"proto/app.proto"}, snapshot)
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(baseline, "proto"), filepath.Join(baseline, "a"), filepath.Join(baseline, "b")}, roots.Paths())
}

func TestBreakingScopesRespectModuleRoots(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		path   string
		scopes map[string]breakingScope
	}{
		{
			name: "only module roots", path: ".",
			scopes: map[string]breakingScope{
				".":     {files: []string{"proto/owned.proto"}},
				"child": {files: []string{"child/schema/child.proto"}},
			},
		},
		{name: "missing source path", path: "absent", scopes: map[string]breakingScope{}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for name, body := range map[string]string{
				"protobuf.mod":                 "module example.com/root\nroots proto\n",
				"proto/owned.proto":            "syntax = \"proto3\";",
				"other.proto":                  "not a selected input",
				"child/protobuf.mod":           "module example.com/child\nroots schema\n",
				"child/schema/child.proto":     "syntax = \"proto3\";",
				"child/outside.proto":          "not a selected input",
				"easyp_vendor/dep.proto":       "not checked",
				"child/easyp_vendor/dep.proto": "not checked",
				".deps/external.proto":         "not checked",
			} {
				writeV1GenerateFixture(t, root, name, body)
			}

			replacements := newPolicyReplacementSources(t.Context(), root, root, nil, false)
			scopes, err := discoverBreakingScopes(replacements, tt.path)

			require.NoError(t, err)
			assert.Equal(t, tt.scopes, scopes)
		})
	}
}

func TestBaselineRepositoryAliases(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	require.NoError(t, os.Mkdir(repo, 0o755))
	alias := filepath.Join(root, "alias")
	require.NoError(t, os.Symlink(repo, alias))
	got, err := baselineRepositoryRelative(repo, filepath.Join(alias, "deleted", "dep"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("deleted", "dep"), got)
}

func TestBaselineRepositoryInternalAliases(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "module/api/item.proto", "syntax = \"proto3\";\n")
	require.NoError(t, os.Symlink("module", filepath.Join(root, "alias")))
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "module", path: "alias", want: "module"},
		{name: "subdirectory", path: "alias/api", want: "module/api"},
		{name: "file", path: "alias/api/item.proto", want: "module/api/item.proto"},
		{name: "deleted file", path: "alias/api/deleted.proto", want: "module/api/deleted.proto"},
		{name: "deleted subtree", path: "alias/deleted/item.proto", want: "module/deleted/item.proto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := baselineRepositoryRelative(root, filepath.Join(root, tt.path))
			require.NoError(t, err)
			assert.Equal(t, filepath.FromSlash(tt.want), got)
		})
	}
}

func TestBaselineRepositoryAliasEscape(t *testing.T) {
	t.Parallel()
	root, outside := t.TempDir(), t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "alias")))
	got, err := baselineRepositoryRelative(root, filepath.Join(root, "alias/deleted.proto"))
	require.NoError(t, err)
	assert.False(t, filepath.IsLocal(got), got)
}

func TestBaselineRepositoryDeletedAliasTarget(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.Symlink("deleted/module", filepath.Join(root, "alias")))
	got, err := baselineRepositoryRelative(root, filepath.Join(root, "alias/api/item.proto"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("deleted", "module", "api", "item.proto"), got)
}

func TestBreakingWalkerDoesNotLeakUnselectedRepositoryFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "selected/item.proto", "selected")
	writeV1GenerateFixture(t, root, "other.proto", "unselected")
	walker := newBreakingWalker(root, []string{"selected/item.proto"})
	_, err := walker.Open("other.proto")
	require.ErrorIs(t, err, os.ErrNotExist)
	f, err := walker.Open("selected/item.proto")
	require.NoError(t, err)
	require.NoError(t, f.Close())
}
