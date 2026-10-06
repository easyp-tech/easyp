package gitsnapshot

import (
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

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
