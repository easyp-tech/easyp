package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/samber/lo"

	"github.com/easyp-tech/easyp/internal/core/models"
)

// Update all packages from config
// dependencies slice of strings format: origin@version: github.com/company/repository@v1.2.3
// if version is absent use the latest commit
func (c *Core) Update(ctx context.Context, dependencies []string) error {
	lo.Uniq(dependencies)

	for _, dependency := range dependencies {

		module := models.NewModule(dependency)

		c.logger.Debug("Updating dependency", "name", module.Name, "version", module.Version)

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
