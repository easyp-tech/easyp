package gitsnapshot

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"weak"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

func TestPinnedTreeReusesObjectMetadataDuringConcurrentReads(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(directory, "a/b/c"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "a/b/c/file.proto"), []byte("pinned bytes"), 0o644))
	runGit(t, directory, "init", "-q")
	runGit(t, directory, "add", ".")
	runGit(t, directory, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	repo, err := git.PlainOpen(directory)
	require.NoError(t, err)
	store := &countingObjectStore{Storer: repo.Storer, reads: make(map[plumbing.Hash]int)}
	repo, err = git.Open(store, nil)
	require.NoError(t, err)
	tree, err := New(repo, plumbing.NewHash(runGit(t, directory, "rev-parse", "HEAD")))
	require.NoError(t, err)
	t.Run("concurrent reads", func(t *testing.T) {
		for i := range 32 {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				t.Parallel()
				info, err := tree.Lstat("a/b/c/file.proto")
				require.NoError(t, err)
				assert.Equal(t, int64(len("pinned bytes")), info.Size())
				entries, err := tree.ReadDir("a/b/c")
				require.NoError(t, err)
				require.Len(t, entries, 1)
				assert.Equal(t, "file.proto", entries[0].Name())
			})
		}
	})
	store.mu.Lock()
	for hash, reads := range store.reads {
		assert.Equal(t, 1, reads, "immutable object %s was loaded repeatedly", hash)
	}
	bodies := slices.Clone(store.bodies)
	store.mu.Unlock()
	assert.Empty(t, bodies, "metadata reads must not decode blob bodies")
	t.Run("concurrent contents", func(t *testing.T) {
		for i := range 32 {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				t.Parallel()
				file, err := tree.Open("a/b/c/file.proto")
				require.NoError(t, err)
				data, readErr := io.ReadAll(file)
				closeErr := file.Close()
				require.NoError(t, readErr)
				require.NoError(t, closeErr)
				assert.Equal(t, "pinned bytes", string(data))
			})
		}
	})
	store.mu.Lock()
	bodies = slices.Clone(store.bodies)
	store.mu.Unlock()
	require.Len(t, bodies, 32)
	runtime.GC()
	for _, body := range bodies {
		assert.Nil(t, body.Value(), "closed readers must not retain decoded blob bodies")
	}
	runtime.KeepAlive(tree)
}

func TestDirectoryListingDoesNotReadBlobBodies(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "README.md"), []byte("unrelated bytes"), 0o644))
	runGit(t, directory, "init", "-q")
	runGit(t, directory, "add", ".")
	runGit(t, directory, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	repo, err := git.PlainOpen(directory)
	require.NoError(t, err)
	store := &countingObjectStore{Storer: repo.Storer, reads: make(map[plumbing.Hash]int)}
	repo, err = git.Open(store, nil)
	require.NoError(t, err)
	tree, err := New(repo, plumbing.NewHash(runGit(t, directory, "rev-parse", "HEAD")))
	require.NoError(t, err)
	entries, err := tree.ReadDir(".")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "README.md", entries[0].Name())
	assert.False(t, entries[0].IsDir())
	info, err := entries[0].Info()
	require.NoError(t, err)
	assert.Equal(t, int64(len("unrelated bytes")), info.Size())
	store.mu.Lock()
	defer store.mu.Unlock()
	assert.Empty(t, store.bodies, "directory/file metadata must not decode file contents")
}

type countingObjectStore struct {
	storage.Storer
	mu     sync.Mutex
	reads  map[plumbing.Hash]int
	bodies []weak.Pointer[trackedBlob]
}

func (s *countingObjectStore) EncodedObject(kind plumbing.ObjectType, hash plumbing.Hash) (plumbing.EncodedObject, error) {
	s.mu.Lock()
	s.reads[hash]++
	s.mu.Unlock()
	object, err := s.Storer.EncodedObject(kind, hash)
	if err != nil || kind != plumbing.BlobObject {
		return object, err
	}
	body := &trackedBlob{EncodedObject: object}
	s.mu.Lock()
	s.bodies = append(s.bodies, weak.Make(body))
	s.mu.Unlock()
	return body, nil
}

type trackedBlob struct{ plumbing.EncodedObject }

func TestPinnedTreeReadsAliasesWithoutHostTargets(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(directory, "physical"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "physical/file.proto"), []byte("pinned bytes"), 0o644))
	require.NoError(t, os.Symlink("physical", filepath.Join(directory, "logical")))
	runGit(t, directory, "init", "-q")
	runGit(t, directory, "add", ".")
	runGit(t, directory, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	commit := runGit(t, directory, "rev-parse", "HEAD")
	repo, err := git.PlainOpen(directory)
	require.NoError(t, err)
	tree, err := New(repo, plumbing.NewHash(commit))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(directory, "physical/file.proto"), []byte("changed host bytes"), 0o644))
	pointer, err := fs.ReadLink(tree, "logical")
	require.NoError(t, err)
	assert.Equal(t, "physical", pointer)
	info, err := fs.Lstat(tree, "logical")
	require.NoError(t, err)
	assert.Equal(t, fs.ModeSymlink, info.Mode()&fs.ModeSymlink)
	file, err := sourceview.New(tree).Open(t.Context(), "logical/file.proto")
	require.NoError(t, err)
	data, err := io.ReadAll(file)
	closeErr := file.Close()
	require.NoError(t, err)
	require.NoError(t, closeErr)
	assert.Equal(t, "pinned bytes", string(data))
}

func TestTreeRejectsGitlinkCrossings(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "file"), []byte("file"), 0o644))
	runGit(t, directory, "init", "-q")
	runGit(t, directory, "add", ".")
	runGit(t, directory, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	commit := runGit(t, directory, "rev-parse", "HEAD")
	runGit(t, directory, "update-index", "--add", "--cacheinfo", "160000", commit, "submodule")
	runGit(t, directory, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "gitlink")
	repo, err := git.PlainOpen(directory)
	require.NoError(t, err)
	tree, err := New(repo, plumbing.NewHash(runGit(t, directory, "rev-parse", "HEAD")))
	require.NoError(t, err)
	_, err = fs.Lstat(tree, "submodule/file.proto")
	require.ErrorContains(t, err, "gitlink boundary")
	_, err = tree.Open("../file")
	require.ErrorIs(t, err, fs.ErrInvalid)
}

func runGit(t *testing.T, directory string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", directory}, args...)...)
	data, err := command.CombinedOutput()
	require.NoError(t, err, string(data))
	return strings.TrimSpace(string(data))
}
