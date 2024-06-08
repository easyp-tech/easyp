package mod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

// Update all packages from config
// dependencies slice of strings format: origin@version: github.com/company/repository@v1.2.3
// if version is absent use the latest commit
func (c *Mod) Update(ctx context.Context, dependencies []string) error {
	for _, dependency := range dependencies {

		module := models.NewModule(dependency)

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
