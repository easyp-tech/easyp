package moduleconfig

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"gopkg.in/yaml.v3"

	"go.redsock.ru/protopack/internal/adapters/repository"
	"go.redsock.ru/protopack/internal/config/default_consts"
	"go.redsock.ru/protopack/internal/core/models"
)

// Config is the configuration of easyp.
// FIXME: do not duplicate of struct
// but if now will import from config -> cycles deps
type easypConfig struct {
	// Deps is the dependencies repositories
	Deps []string `json:"deps" yaml:"deps"`
}

// readEasyp read easyp's config from repository
func readEasyp(ctx context.Context, repo repository.Repo, revision models.Revision) ([]models.Module, error) {
	content, err := repo.ReadFile(ctx, revision, default_consts.DefaultConfigFileName)
	if err != nil {
		if errors.Is(err, models.ErrFileNotFound) {
			slog.Debug("easyp config not found")
			return nil, nil
		}
		return nil, fmt.Errorf("repo.ReadFile: %w", err)
	}

	easyp := &easypConfig{}
	if err := yaml.NewDecoder(strings.NewReader(content)).Decode(&easyp); err != nil {
		return nil, fmt.Errorf("yaml.NewDecoder: %w", err)
	}

	modules := make([]models.Module, 0, len(easyp.Deps))
	for _, dep := range easyp.Deps {
		module := models.NewModule(dep)
		modules = append(modules, module)
	}

	return modules, nil
}
