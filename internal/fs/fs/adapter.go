package fs

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type FSAdapter struct {
	fs.FS

	rootDir string
}

func (a *FSAdapter) Open(name string) (io.ReadCloser, error) {
	return a.FS.Open(name)
}

func (a *FSAdapter) Create(name string) (io.WriteCloser, error) {
	path := filepath.Join(a.rootDir, name)
	return os.Create(path)
}
