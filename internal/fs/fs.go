package wfs

import (
	"io/fs"
	"os"
)

// FS is an interface for the file system.
type FS interface {
	fs.FS
	// Create creates the named file for writing.
	Create(name string) (*os.File, error)
}
