package go_git

import (
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.DirWalker = (*GitTreeWalker)(nil)

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

func (w *GitTreeWalker) WalkDir(callback core.WalkerDirCallback) error {
	err := w.tree.Files().ForEach(func(f *object.File) error {
		switch {
		case !isTargetFile(w.path, f.Name):
			return nil
		}

		return callback(f.Name, w, nil)
	})
	return err
}

// isTargetFile check if passed filePath is target
// it has to be in targetPath dir
func isTargetFile(targetPath, filePath string) bool {
	rel, err := filepath.Rel(targetPath, filePath)
	if err != nil {
		return false
	}
	if !filepath.IsLocal(rel) {
		return false
	}

	return true
}
