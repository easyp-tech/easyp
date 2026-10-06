package fs

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

type FS interface {
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
	Exists(name string) bool
	Remove(name string) error
}

func NewFSWalker(root, path string) *FSWalker {
	if path == "" {
		path = "."
	}

	// os.DirFS always expects forward slashes, even on Windows
	path = filepath.ToSlash(path)

	diskFS := os.DirFS(root)
	return &FSWalker{
		FSAdapter: &FSAdapter{diskFS, root},
		path:      path,
	}
}

type FSWalker struct {
	*FSAdapter

	path string
}

func (w *FSWalker) WalkDir(callback func(path string, err error) error) error {
	return sourceview.WalkLocal(context.Background(), w.rootDir, w.path, func(path string, resolved sourceview.Resolution, err error) error {
		if err != nil && filepath.Ext(path) != ".proto" && path != w.path {
			return nil
		}
		return callback(path, err)
	})
}
