package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/core/models"
)

// Download all packages from config
// dependencies slice of strings format: origin@version: github.com/company/repository@v1.2.3
// if version is absent use the latest commit
func (c *Core) Download(ctx context.Context, dependencies []string) error {
	if c.lockFile.IsEmpty() {
		// if lock file is empty or doesn't exist install versions
		// from easyp.yaml config and create lock file
		slog.Debug("Lock file is empty")
		return c.Update(ctx, dependencies)
	}

	slog.Debug("Lock file is not empty. Install deps from it")

	// install from lock file at first
	for lockFileInfo := range c.lockFile.DepsIter() {
		module := models.NewModuleFromLockFileInfo(lockFileInfo)

		c.logger.DebugContext(
			ctx, "start download module from lockfile", "name", module.Name, "version", module.Version,
		)
		if err := c.Get(ctx, module); err != nil {
			return fmt.Errorf("c.Get: %w", err)
		}
	}

	slog.Debug("Start install other deps")

	// install from remote generator sections
	for _, dependency := range dependencies {
		module := models.NewModule(dependency)

		_, err := c.lockFile.Read(module.Name)
		if err == nil {
			c.logger.DebugContext(
				ctx, "already is in lock file", "name", module.Name, "version", module.Version,
			)
			continue
		}
		if !errors.Is(err, models.ErrModuleNotFoundInLockFile) {
			return fmt.Errorf("c.lockFile.Read: %w", err)
		}

		c.logger.DebugContext(
			ctx, "start download module from deps", "name", module.Name, "version", module.Version,
		)
		if err := c.Get(ctx, module); err != nil {
			return fmt.Errorf("c.Get: %w", err)
		}
	}

	return nil
}
