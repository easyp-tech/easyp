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
	for _, dependency := range dependencies {

		module := models.NewModule(dependency)

		version, err := c.getVersionToDownload(module)
		if err != nil {
			return fmt.Errorf("c.getVersionToDownload: %w", err)
		}
		module.Version = version

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
				slog.Error("Version not found", "dependency", dependency)
				return models.ErrVersionNotFound
			}

			return fmt.Errorf("c.Get: %w", err)
		}
	}

	return nil
}

// getVersionToDownload return version which has to be installed by `download` command
// version from lockfile is more important than version from easyp config
func (c *Mod) getVersionToDownload(module models.Module) (models.RequestedVersion, error) {
	lockFileInfo, err := c.lockFile.Read(module.Name)
	if err == nil {
		return models.RequestedVersion(lockFileInfo.Version), nil
	}

	if !errors.Is(err, models.ErrModuleNotFoundInLockFile) {
		return "", fmt.Errorf("c.lockFile.Read: %w", err)
	}

	return module.Version, nil
}
