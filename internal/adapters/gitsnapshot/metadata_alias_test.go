package gitsnapshot

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

func TestClassifyingUnrelatedAliasDoesNotDecodeItsTarget(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "large.bin"), make([]byte, 1024*1024), 0o644))
	require.NoError(t, os.Symlink("large.bin", filepath.Join(directory, "alias")))
	runGit(t, directory, "init", "-q")
	runGit(t, directory, "add", ".")
	runGit(t, directory, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	commit := runGit(t, directory, "rev-parse", "HEAD")
	target := plumbing.NewHash(runGit(t, directory, "rev-parse", "HEAD:large.bin"))
	repo, err := git.PlainOpen(directory)
	require.NoError(t, err)
	store := &countingObjectStore{Storer: repo.Storer, reads: make(map[plumbing.Hash]int)}
	repo, err = git.Open(store, nil)
	require.NoError(t, err)
	tree, err := New(repo, plumbing.NewHash(commit))
	require.NoError(t, err)
	err = sourceview.New(tree).WalkSelected(t.Context(), ".", func(_ string, entry fs.DirEntry) bool {
		return entry.IsDir() || entry.Type()&fs.ModeSymlink != 0
	}, func(_ string, _ sourceview.Resolution, err error) error { return err })
	require.NoError(t, err)
	store.mu.Lock()
	defer store.mu.Unlock()
	assert.Zero(t, store.reads[target], "alias classification must not decode the unrelated target")
}
