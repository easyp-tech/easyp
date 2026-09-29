package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBreakingScopesRespectModuleRoots(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for name, body := range map[string]string{
		"protobuf.mod": "module example.com/root\nroots proto\n", "proto/owned.proto": "syntax = \"proto3\";", "other.proto": "not a selected input",
		"child/protobuf.mod": "module example.com/child\nroots schema\n", "child/schema/child.proto": "syntax = \"proto3\";", "child/outside.proto": "not a selected input",
		"easyp_vendor/dep.proto": "not checked", "child/easyp_vendor/dep.proto": "not checked", ".deps/external.proto": "not checked",
	} {
		writeV1GenerateFixture(t, root, name, body)
	}
	scopes, err := discoverBreakingScopes(root, ".")
	require.NoError(t, err)
	require.Len(t, scopes, 2)
	assert.Equal(t, []string{"proto/owned.proto"}, scopes["."].files)
	assert.Equal(t, []string{"child/schema/child.proto"}, scopes["child"].files)
	scopes, err = discoverBreakingScopes(root, "absent")
	require.NoError(t, err)
	assert.Empty(t, scopes)
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
