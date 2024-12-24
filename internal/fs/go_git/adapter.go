package go_git

import (
	"io"

	"github.com/go-git/go-git/v5/plumbing/object"

	wfs "github.com/easyp-tech/easyp/internal/fs"
)

var _ wfs.FSReader = (*GitTreeDiskAdapter)(nil)

type GitTreeDiskAdapter struct {
	*object.Tree
}

func (a *GitTreeDiskAdapter) Open(name string) (io.ReadCloser, error) {
	gitFile, err := a.File(name)
	if err != nil {
		return nil, err
	}

	return gitFile.Blob.Reader()
}
