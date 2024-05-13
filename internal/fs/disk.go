package wfs

import (
	"io/fs"
	"os"
	"path/filepath"
)

// disk is a disk file system.
type disk struct {
	fs.FS
	root string
}

// Create creates the named file for writing.
func (d *disk) Create(name string) (*os.File, error) {
	return os.Create(filepath.Join(d.root, name))
}

// Disk returns a disk file system.
func Disk(root string) FS {
	return &disk{
		FS:   os.DirFS(root),
		root: root,
	}
}
