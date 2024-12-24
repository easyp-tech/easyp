package fs

import (
	"io/fs"

	wfs "github.com/easyp-tech/easyp/internal/fs"
)

var _ wfs.DirWalker = (*FSWalker)(nil)

func NewFSWalker(fs fs.FS, path string) *FSWalker {
	return &FSWalker{
		path:    path,
		adapter: &FSAdapter{fs},
	}
}

type FSWalker struct {
	path    string
	adapter *FSAdapter
}

func (w *FSWalker) WalkDir(callback wfs.WalkerDirCallback) error {
	err := fs.WalkDir(w.adapter.FS, w.path, func(path string, d fs.DirEntry, err error) error {
		return callback(path, err)
	})

	return err
}
