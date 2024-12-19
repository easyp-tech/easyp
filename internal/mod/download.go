package mod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

// Download all packages from config
// dependencies slice of strings format: origin@version: github.com/company/repository@v1.2.3
// if version is absent use the latest commit
func (c *Mod) Download(ctx context.Context, dependencies []string) error {
	if c.lockFile.IsEmpty() {
		// if lock file is empty or doesn't exist install versions
		// from easyp.yaml config and create lock file
		slog.Debug("Lock file is empty")
		return c.Update(ctx, dependencies)
	}

	slog.Debug("Lock file is not empty. Install deps from it")

	for lockFileInfo := range c.lockFile.DepsIter() {
		module := models.NewModuleFromLockFileInfo(lockFileInfo)

		isInstalled, err := c.storage.IsModuleInstalled(module)
		if err != nil {
			return fmt.Errorf("c.isModuleInstalled: %w", err)
		}

		if isInstalled {
			slog.Info("Module is installed", "name", module.Name, "version", module.Version)
			continue
		}

		if err := c.Get(ctx, module); err != nil {
			if errors.Is(err, models.ErrVersionNotFound) {
				slog.Error("Version not found", "name", module.Name, "version", module.Version)
				return models.ErrVersionNotFound
			}

			return fmt.Errorf("c.Get: %w", err)
		}
	}

	return nil
}
