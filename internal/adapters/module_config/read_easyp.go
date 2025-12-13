package moduleconfig

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/adapters/repository"
	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core/models"
)

// readEasyp read easyp's config from repository
func readEasyp(ctx context.Context, repo repository.Repo, revision models.Revision) ([]models.Module, error) {
	content, err := repo.ReadFile(ctx, revision, config.DefaultFileName)
	if err != nil {
		if errors.Is(err, models.ErrFileNotFound) {
			slog.Debug("easyp.yaml not found in dependency (this is normal)")
			return nil, nil
		}
		return nil, fmt.Errorf("repo.ReadFile: %w", err)
	}

	// Use unified parsing function with environment variable support
	cfg, err := config.ParseConfig([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("config.ParseConfig: %w", err)
	}

	modules := make([]models.Module, 0, len(cfg.Deps))
	for _, dep := range cfg.Deps {
		module := models.NewModule(dep)
		modules = append(modules, module)
	}

	return modules, nil
}
