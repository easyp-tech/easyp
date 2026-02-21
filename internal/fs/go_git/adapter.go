package go_git

import (
	"errors"
	"io"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type GitTreeDiskAdapter struct {
	*object.Tree
}

func (a *GitTreeDiskAdapter) Open(name string) (io.ReadCloser, error) {
	gitFile, err := a.File(name)
	if err != nil {
		return nil, err
	}

	return gitFile.Reader()
}

func (a *GitTreeDiskAdapter) Create(name string) (io.WriteCloser, error) {
	return nil, errors.New("not implemented")
}

func (a *GitTreeDiskAdapter) Exists(name string) bool {
	_, err := a.File(name)
	return err == nil
}

func (a *GitTreeDiskAdapter) Remove(_ string) error {
	return errors.New("not implemented")
}
