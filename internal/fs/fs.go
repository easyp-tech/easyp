package wfs

import (
	"io"
	"io/fs"
	"os"
)

// FS is an interface for the file system.
type FS interface {
	fs.FS
	// Create creates the named file for writing.
	Create(name string) (*os.File, error)
}

// FSReader an interface for reading from some FS (os disk, git repo etc)
type FSReader interface {
	Open(name string) (io.ReadCloser, error)
}

// FSWriter interface for writing to some FS
type FSWriter interface {
	Create(name string) (io.WriteCloser, error)
}
