package core

import (
	"context"
	"fmt"

	cp "github.com/otiai10/copy"

	"github.com/easyp-tech/easyp/internal/core/models"
)

// Vendor copy all proto files from deps to local dir
func (c *Core) Vendor(ctx context.Context) error {
	if err := c.Download(ctx, c.deps); err != nil {
		return fmt.Errorf("c.Download: %w", err)
	}

	for lockFileInfo := range c.lockFile.DepsIter() {
		depPath, err := c.modulePath(models.NewModule(lockFileInfo.Name))
		if err != nil {
			return fmt.Errorf("modulePath: %w", err)
		}

		if err := cp.Copy(depPath, c.vendorDir); err != nil {
			return fmt.Errorf("c.Copy: %w", err)
		}
	}

	return nil
}
