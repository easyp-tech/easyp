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
		c.logger.Debug(ctx, "Lock file is empty")
		return c.Update(ctx, dependencies)
	}

	c.logger.Debug(ctx, "Lock file is not empty. Install deps from it")

	// install from lock file at first
	for lockFileInfo := range c.lockFile.DepsIter() {
		module := models.NewModuleFromLockFileInfo(lockFileInfo)
		log := c.logger.With(slog.String("module", module.Name), slog.String("version", string(module.Version)))

		log.Debug(ctx, "downloading module from lockfile")
		if err := c.Get(ctx, module); err != nil {
			return fmt.Errorf("c.Get: %w", err)
		}
	}

	c.logger.Debug(ctx, "installing remaining dependencies not in lock file")

	// install from remote generator sections
	for _, dependency := range dependencies {
		module := models.NewModule(dependency)
		log := c.logger.With(slog.String("module", module.Name), slog.String("version", string(module.Version)))

		_, err := c.lockFile.Read(module.Name)
		if err == nil {
			log.Debug(ctx, "already in lock file")
			continue
		}
		if !errors.Is(err, models.ErrModuleNotFoundInLockFile) {
			return fmt.Errorf("c.lockFile.Read: %w", err)
		}

		log.Debug(ctx, "downloading module from deps")
		if err := c.Get(ctx, module); err != nil {
			return fmt.Errorf("c.Get: %w", err)
		}
	}

	return nil
}
