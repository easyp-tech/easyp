package moduleconfig

import (
	"context"
	"errors"
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

	deps, err := readDeps(ctx, repo, revision, easyp.Deps)
	if err != nil {
		return models.ModuleConfig{}, fmt.Errorf("readDeps: %w", err)
	}

	modules := make([]models.Module, 0, len(deps))
	for _, dep := range deps {
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

// readDeps loads dependency declarations from protobuf.mod when present.
// Falls back to easyp.yaml deps for backward compatibility.
func readDeps(
	ctx context.Context,
	repo repository.Repo,
	revision models.Revision,
	yamlDeps []string,
) ([]string, error) {
	content, err := repo.ReadFile(ctx, revision, config.DefaultModFileName)
	if err != nil {
		if errors.Is(err, models.ErrFileNotFound) {
			return yamlDeps, nil
		}

		return nil, fmt.Errorf("repo.ReadFile: %w", err)
	}

	mod, err := config.ParseModFile([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("config.ParseModFile: %w", err)
	}

	if len(yamlDeps) > 0 {
		return nil, fmt.Errorf(
			"deps must be declared only in %s; remove deps from easyp.yaml",
			config.DefaultModFileName,
		)
	}

	return mod.Deps, nil
}
