package moduleconfig

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/adapters/repository"
	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core/models"
)

// readEasyp reads directories from easyp.yaml in the repository (dependencies live in protobuf.mod).
func readEasyp(ctx context.Context, repo repository.Repo, revision models.Revision) (models.ModuleConfig, error) {
	content, err := repo.ReadFile(ctx, revision, config.DefaultFileName)
	if err != nil {
		return models.ModuleConfig{}, fmt.Errorf("repo.ReadFile: %w", err)
	}

	easyp, err := config.ParseConfig([]byte(content))
	if err != nil {
		return models.ModuleConfig{}, fmt.Errorf("config.ParseConfig: %w", err)
	}

	dirs := make([]string, 0, len(easyp.Generate.Inputs))
	for _, input := range easyp.Generate.Inputs {
		dirs = append(dirs, input.InputFilesDir.Root)
	}

	return models.ModuleConfig{
		Directories: dirs,
	}, nil
}
