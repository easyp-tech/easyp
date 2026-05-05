package go_git

import (
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type GitTreeDiskAdapter struct {
	*object.Tree
	root string
}

func (a *GitTreeDiskAdapter) Open(name string) (io.ReadCloser, error) {
	if a.root != "" {
		// Reject non-local import paths (e.g., those containing ".." or absolute paths)
		if !isLocalPath(name) {
			return nil, fmt.Errorf("non-local import path: %q", name)
		}
		// Prefer root-prefixed lookup to match the on-disk walker behaviour
		withRoot := path.Join(a.root, name)
		gitFile, err := a.File(withRoot)
		if err == nil {
			return gitFile.Reader()
		}
	}

	// Fall back to unrooted lookup (e.g., caller already passed a fully-qualified path)
	gitFile, err := a.File(name)
	if err != nil {
		return nil, err
	}

	return gitFile.Reader()
}

// isLocalPath reports whether p does not escape its base directory
// (no ".." components, not an absolute path).
func isLocalPath(p string) bool {
	if filepath.IsAbs(p) {
		return false
	}
	for _, segment := range strings.Split(p, "/") {
		if segment == ".." {
			return false
		}
	}
	return true
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
