package moduleconfig

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/adapters/repository"
	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core/models"
)

// readEasyp read easyp's config from repository
func readEasyp(ctx context.Context, repo repository.Repo, revision models.Revision) (models.ModuleConfig, error) {
	content, err := repo.ReadFile(ctx, revision, config.DefaultFileName)
	if err != nil {
		return models.ModuleConfig{}, fmt.Errorf("repo.ReadFile: %w", err)
	}

	easyp, err := config.ParseConfig([]byte(content))
	if err != nil {
		return models.ModuleConfig{}, fmt.Errorf("config.ParseConfig: %w", err)
	}

	modules := make([]models.Module, 0, len(easyp.Deps))
	for _, dep := range easyp.Deps {
		module := models.NewModule(dep)
		modules = append(modules, module)
	}

	dirs := make([]string, 0, len(easyp.Generate.Inputs))
	for _, input := range easyp.Generate.Inputs {
		dirs = append(dirs, input.InputFilesDir.Root)
	}

	return models.ModuleConfig{
		Dependencies: modules,
		Directories:  dirs,
	}, nil
}
