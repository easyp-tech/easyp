package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/core/models"
)

// Update all packages from config
// dependencies slice of strings format: origin@version: github.com/company/repository@v1.2.3
// if version is absent use the latest commit
func (c *Core) Update(ctx context.Context) error {
	c.logger.Info(ctx, "updating dependencies", slog.Int("count", len(c.deps)))

	for _, dependency := range c.deps {
		module := models.NewModule(dependency)
		log := c.logger.With(slog.String("module", module.Name), slog.String("version", string(module.Version)))

		log.Debug(ctx, "updating dependency")

		if err := c.Get(ctx, module); err != nil {
			if errors.Is(err, models.ErrVersionNotFound) {
				log.Error(ctx, "version not found")
				return models.ErrVersionNotFound
			}

			return fmt.Errorf("c.Get: %w", err)
		}
	}

	c.logger.Info(ctx, "update completed")

	return nil
}
