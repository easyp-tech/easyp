package wfs

import (
	"io"
)

// FS an interface for reading from some FS (os disk, git repo etc)
// and for writing to some FS
type FS interface {
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
}

type WalkerDirCallback func(path string, fs FS, err error) error

type DirWalker interface {
	FS
	WalkDir(callback WalkerDirCallback) error
}
