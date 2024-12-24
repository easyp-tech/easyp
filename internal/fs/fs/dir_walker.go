package fs

import (
	"io/fs"

	wfs "github.com/easyp-tech/easyp/internal/fs"
)

var _ wfs.DirWalker = (*FSWalker)(nil)

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

func (w *FSWalker) WalkDir(callback wfs.WalkerDirCallback) error {
	err := fs.WalkDir(w.FS, w.path, func(path string, d fs.DirEntry, err error) error {
		return callback(path, w, err)
	})

	return err
}
