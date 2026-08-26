package moduleconfig

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/adapters/modfile"
	"github.com/easyp-tech/easyp/internal/adapters/repository"
	"github.com/easyp-tech/easyp/internal/core/models"
)

// Read and return module's config from repository
func (c *ModuleConfig) ReadFromRepo(
	ctx context.Context, repo repository.Repo, revision models.Revision,
) (models.ModuleConfig, error) {
	result := models.ModuleConfig{}

	// Directories: buf layout, else easyp generate inputs.
	c.logger.Debug(ctx, "reading buf config from repo", slog.String("revision", revision.Version))

	buf, err := readBufWork(ctx, repo, revision)
	switch {
	case err == nil:
		result.Directories = buf.Directories
	case !errors.Is(err, models.ErrFileNotFound):
		return models.ModuleConfig{}, fmt.Errorf("readBufWork: %w", err)
	default:
		c.logger.Debug(ctx, "reading easyp config from repo", slog.String("revision", revision.Version))

		easyp, easypErr := readEasyp(ctx, repo, revision)
		switch {
		case easypErr == nil:
			result.Directories = easyp.Directories
		case !errors.Is(easypErr, models.ErrFileNotFound):
			return models.ModuleConfig{}, fmt.Errorf("readEasyp: %w", easypErr)
		}
	}

	// Transitive dependencies: protobuf.mod only.
	c.logger.Debug(ctx, "reading protobuf.mod from repo", slog.String("revision", revision.Version))

	deps, err := readProtobufMod(ctx, repo, revision)
	switch {
	case err == nil:
		result.Dependencies = deps
	case !errors.Is(err, models.ErrFileNotFound):
		return models.ModuleConfig{}, fmt.Errorf("readProtobufMod: %w", err)
	}

	return result, nil
}

func readProtobufMod(
	ctx context.Context, repo repository.Repo, revision models.Revision,
) ([]models.Module, error) {
	content, err := repo.ReadFile(ctx, revision, modfile.FileName)
	if err != nil {
		return nil, fmt.Errorf("repo.ReadFile: %w", err)
	}

	rawDeps, err := modfile.Parse([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("modfile.Parse: %w", err)
	}

	modules := make([]models.Module, 0, len(rawDeps))
	for _, dep := range rawDeps {
		modules = append(modules, models.NewModule(dep))
	}

	return modules, nil
}
