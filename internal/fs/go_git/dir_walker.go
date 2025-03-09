package go_git

import (
	"github.com/go-git/go-git/v5/plumbing/object"

	"go.redsock.ru/protopack/internal/core/path_helpers"
)

func NewGitTreeWalker(tree *object.Tree, path string) *GitTreeWalker {
	return &GitTreeWalker{
		GitTreeDiskAdapter: &GitTreeDiskAdapter{tree},
		tree:               tree,
		path:               path,
	}
}

type GitTreeWalker struct {
	*GitTreeDiskAdapter

	tree *object.Tree
	path string
}

func (w *GitTreeWalker) WalkDir(callback func(path string, err error) error) error {
	err := w.tree.Files().ForEach(func(f *object.File) error {
		switch {
		case !path_helpers.IsTargetPath(w.path, f.Name):
			return nil
		}

		return callback(f.Name, nil)
	})
	return err
}
