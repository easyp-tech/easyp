package core

import (
	"io"
	"os"
)

// openSourceFile is shared by descriptor compilation and syntax-based import lookup.
func (c *Core) openSourceFile(path string) (io.ReadCloser, error) {
	if c.importFileAllowed != nil && !c.importFileAllowed(path) {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}
	return os.Open(path)
}
