package mod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

// Download all packages from config
func (c *Mod) Download(ctx context.Context, dependencies []string) error {
	for _, dependency := range dependencies {

		module := models.NewModule(dependency)

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
