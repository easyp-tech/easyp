package fs

import (
	"io/fs"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.DirWalker = (*FSWalker)(nil)

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

func (w *FSWalker) WalkDir(callback core.WalkerDirCallback) error {
	err := fs.WalkDir(w.FS, w.path, func(path string, d fs.DirEntry, err error) error {
		return callback(path, w, err)
	})

	return err
}
