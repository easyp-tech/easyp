package core

import (
	"context"
	"fmt"

	cp "github.com/otiai10/copy"
)

// Vendor copy all proto files from deps to local dir
func (c *Core) Vendor(ctx context.Context) error {
	if err := c.Download(ctx, c.deps); err != nil {
		return fmt.Errorf("c.Download: %w", err)
	}

	for lockFileInfo := range c.lockFile.DepsIter() {
		depPath := c.storage.GetInstallDir(lockFileInfo.Name, lockFileInfo.Version)

		if err := cp.Copy(depPath, c.vendorDir); err != nil {
			return fmt.Errorf("c.Copy: %w", err)
		}
	}

	return nil
}
