package core

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

// openSourceFile is shared by descriptor compilation and syntax-based import lookup.
func (c *Core) openSourceFile(path string) (io.ReadCloser, error) {
	if c.importFileAllowed != nil && !c.importFileAllowed(path) {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}
	if c.sourceFileOpen != nil {
		return c.sourceFileOpen(path)
	}
	roots := append([]string{c.sourceBoundary}, c.importRoots...)
	for _, root := range roots {
		if root == "" {
			continue
		}
		relative, err := filepath.Rel(root, path)
		if err == nil && filepath.IsLocal(relative) {
			return sourceview.OpenLocal(context.Background(), root, relative)
		}
	}
	return sourceview.OpenLocal(context.Background(), filepath.Dir(path), filepath.Base(path))
}
