package fs

import (
	"io"
	"io/fs"
	"os"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.FS = (*FSAdapter)(nil)

type FSAdapter struct {
	fs.FS
}

func (a *FSAdapter) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func (a *FSAdapter) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}
