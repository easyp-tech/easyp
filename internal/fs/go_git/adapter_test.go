package go_git

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitTreeDiskAdapterOpen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		root    string
		path    string
		want    string
		wantErr error
	}{
		{name: "rooted file", root: "proto", path: "item.proto", want: "syntax = \"proto3\";\n"},
		{name: "missing file", root: "proto", path: "google/protobuf/timestamp.proto", wantErr: fs.ErrNotExist},
		{name: "missing root directory", root: "missing", path: "item.proto", wantErr: fs.ErrNotExist},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			path := filepath.Join(root, "proto", "item.proto")
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
			require.NoError(t, os.WriteFile(path, []byte("syntax = \"proto3\";\n"), 0o644))
			repo, err := git.PlainInit(root, false)
			require.NoError(t, err)
			worktree, err := repo.Worktree()
			require.NoError(t, err)
			_, err = worktree.Add("proto/item.proto")
			require.NoError(t, err)
			commit, err := worktree.Commit("initial", &git.CommitOptions{Author: &object.Signature{Name: "EasyP Test", Email: "test@example.com", When: time.Now()}})
			require.NoError(t, err)
			commitObject, err := repo.CommitObject(commit)
			require.NoError(t, err)
			tree, err := commitObject.Tree()
			require.NoError(t, err)
			adapter := &GitTreeDiskAdapter{Tree: tree, root: tt.root}

			reader, err := adapter.Open(tt.path)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, reader)
				return
			}
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, reader.Close()) })
			content, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(content))
		})
	}
}
