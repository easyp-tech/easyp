package gitmodules

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestNativeSnapshotOmitsUnrelatedFiles(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{
		"README.md":      "unrelated documentation\n",
		"source/main.go": "package main\n",
		"build/app.bin":  "unrelated build artifact\n",
	}, nil)
	cache := New(t.TempDir())
	fetched, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
	installed, _, err := cache.Cached(fetched.Lock)
	require.NoError(t, err)
	files, err := snapshotV1Files(installed)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"protobuf.mod", "proto/main.proto"}, files)
}

func snapshotTestHash(t *testing.T, files map[string]string) string {
	t.Helper()
	selected := make(map[string]string)
	for name, data := range files {
		if path.Ext(name) == ".proto" || snapshotConfigFile(name) {
			selected[name] = data
		}
	}
	return migrationTestHash(t, selected)
}

func TestUnrelatedFileChangeDoesNotChangeNativeSnapshotHash(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{"README.md": "original docs\n"}, nil)
	cache := New(t.TempDir())
	first, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(repository, "README.md"), []byte("changed docs\n"), 0o644))
	runTestGit(t, repository, "add", "README.md")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "docs only")
	second, err := cache.Fetch(t.Context(), repository, "")
	require.NoError(t, err)
	assert.NotEqual(t, first.Lock.Commit, second.Lock.Commit)
	assert.Equal(t, first.Lock.Hash, second.Lock.Hash)
}
