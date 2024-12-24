package fs

import (
	"io"
	"io/fs"
	"os"

	wfs "github.com/easyp-tech/easyp/internal/fs"
)

var (
	_ wfs.FSReader = (*FSAdapter)(nil)
	_ wfs.FSWriter = (*FSAdapter)(nil)
)

type FSAdapter struct {
	fs.FS
}

func (a *FSAdapter) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func (a *FSAdapter) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}
