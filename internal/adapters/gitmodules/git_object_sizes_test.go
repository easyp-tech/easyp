package gitmodules

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnapshotStorageReadsSharedObjects(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		relative bool
	}{
		{name: "absolute alternate"},
		{name: "relative alternate", relative: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			remote, commit := diagnosticRepository(t)
			checkout := filepath.Join(t.TempDir(), "shared checkout")
			runTestGit(t, remote, "clone", "--quiet", "--shared", "--no-checkout", "--", remote, checkout)
			objects := filepath.Join(checkout, ".git", "objects")
			require.NoFileExists(t, filepath.Join(objects, commit[:2], commit[2:]))
			packs, err := os.ReadDir(filepath.Join(objects, "pack"))
			require.NoError(t, err)
			require.Empty(t, packs)
			alternate := filepath.Join(remote, ".git", "objects")
			if tt.relative {
				var err error
				alternate, err = filepath.Rel(objects, alternate)
				require.NoError(t, err)
			}
			require.NoError(t, os.WriteFile(filepath.Join(objects, "info", "alternates"), []byte(filepath.ToSlash(alternate)+"\n"), 0o644))
			// The primary checkout has no object bodies. Native Git and EasyP
			// must both read the same shared object database on every platform.
			runTestGit(t, checkout, "cat-file", "-e", commit+"^{commit}")
			store, err := snapshotStorage(checkout)
			require.NoError(t, err)
			repository, err := git.Open(store, nil)
			require.NoError(t, err)
			revision, err := repository.CommitObject(plumbing.NewHash(commit))
			require.NoError(t, err)
			tree, err := revision.Tree()
			require.NoError(t, err)
			file, err := tree.File("api.proto")
			require.NoError(t, err)
			size, err := store.EncodedObjectSize(file.Hash)
			require.NoError(t, err)
			reader, err := file.Reader()
			require.NoError(t, err)
			contents, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.NoError(t, reader.Close())
			assert.Equal(t, int64(len(contents)), size)
			assert.True(t, strings.HasPrefix(string(contents), "syntax = \"proto3\";"))
		})
	}
}
