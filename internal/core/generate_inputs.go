package core

import (
	"fmt"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/fs/fs"
)

func (c *Core) walkGenerationInputs(root string, input InputFilesDir, visit func(string, error, bool) error) error {
	searchPath := filepath.Join(input.Root, input.Path)
	if c.sourceFilesWalk != nil {
		indexed, err := c.sourceFilesWalk(filepath.Join(root, input.Root), func(path string) error {
			within, err := filepath.Rel(filepath.Join(root, searchPath), path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			if !filepath.IsLocal(within) {
				return nil
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			if !filepath.IsLocal(relative) {
				return fmt.Errorf("indexed source %q is outside the module", path)
			}
			return visit(filepath.ToSlash(relative), nil, true)
		})
		if err != nil {
			return fmt.Errorf("WalkSourceFiles: %w", err)
		}
		if indexed {
			return nil
		}
	}
	return fs.NewFSWalker(root, searchPath).WalkDir(func(path string, err error) error {
		return visit(path, err, false)
	})
}
