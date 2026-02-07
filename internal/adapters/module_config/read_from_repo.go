package moduleconfig

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/adapters/repository"
	"github.com/easyp-tech/easyp/internal/core/models"
)

// Read and return module's config from repository
func (c *ModuleConfig) ReadFromRepo(
	ctx context.Context, repo repository.Repo, revision models.Revision,
) (models.ModuleConfig, error) {
	// buf
	c.logger.Debug(ctx, "reading buf config from repo", slog.String("revision", revision.Version))

	buf, err := readBufWork(ctx, repo, revision)
	if err == nil {
		return buf, nil
	}
	if !errors.Is(err, models.ErrFileNotFound) {
		return models.ModuleConfig{}, fmt.Errorf("readBufWork: %w", err)
	}

	// easyp
	c.logger.Debug(ctx, "reading easyp config from repo", slog.String("revision", revision.Version))

	easyp, err := readEasyp(ctx, repo, revision)
	if err == nil {
		return easyp, nil
	}
	if !errors.Is(err, models.ErrFileNotFound) {
		return models.ModuleConfig{}, fmt.Errorf("readEasyp: %w", err)
	}

	return models.ModuleConfig{}, nil
}
