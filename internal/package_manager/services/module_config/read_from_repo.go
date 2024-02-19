package moduleconfig

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
	"github.com/easyp-tech/easyp/internal/package_manager/services/repository"
)

// Read and return module's config from repository
func (c *ModuleConfig) ReadFromRepo(ctx context.Context, repo repository.Repo, revision models.Revision) (models.ModuleConfig, error) {
	buf, err := readBufWork(ctx, repo, revision)
	if err != nil {
		return models.ModuleConfig{}, fmt.Errorf("readBufWork: %w", err)
	}

	moduleConfig := models.ModuleConfig{
		Directories: buf.Directories,
	}
	return moduleConfig, nil
}
