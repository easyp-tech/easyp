package core

import (
	"io"
)

type DirWalker interface {
	FS
	WalkDir(callback func(path string, err error) error) error
}

// FS an interface for reading from some FS (os disk, git repo etc)
// and for writing to some FS
type FS interface {
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
	Exists(name string) bool
	Remove(name string) error
}
