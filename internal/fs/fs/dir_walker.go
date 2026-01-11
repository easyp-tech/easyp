package fs

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type FS interface {
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
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
	err := fs.WalkDir(w.FS, w.path, func(path string, d fs.DirEntry, err error) error {
		return callback(path, err)
	})

	return err
}
