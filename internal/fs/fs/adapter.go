package fs

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

type FSAdapter struct {
	fs.FS

	rootDir string
}

func (a *FSAdapter) Open(name string) (io.ReadCloser, error) {
	return sourceview.OpenLocal(context.Background(), a.rootDir, name)
}

func (a *FSAdapter) Create(name string) (io.WriteCloser, error) {
	path := filepath.Join(a.rootDir, name)
	return os.Create(path)
}

func (a *FSAdapter) Exists(name string) bool {
	_, err := fs.Stat(a.FS, name)
	return err == nil
}

func (a *FSAdapter) Remove(name string) error {
	path := filepath.Join(a.rootDir, name)
	return os.Remove(path)
}

// RootPath returns the filesystem root used to resolve diagnostic paths.
func (a *FSAdapter) RootPath() string { return a.rootDir }
