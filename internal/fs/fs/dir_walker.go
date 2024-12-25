package fs

import (
	"io/fs"
)

func NewFSWalker(fs fs.FS, path string) *FSWalker {
	return &FSWalker{
		FSAdapter: &FSAdapter{fs},
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
