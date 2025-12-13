package moduleconfig

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/adapters/repository"
	"github.com/easyp-tech/easyp/internal/core/models"
)

type bufV1Config struct {
	Directories []string `yaml:"directories"`
}

type bufV2Config struct {
	Modules []struct {
		Path string `yaml:"path"`
		Name string `yaml:"name"`
		Lint struct {
			IgnoreOnly struct {
				PACKAGEVERSIONSUFFIX []string `yaml:"PACKAGE_VERSION_SUFFIX"`
			} `yaml:"ignore_only"`
		} `yaml:"lint"`
	} `yaml:"modules"`
}

const (
	bufV1ConfigFile = "buf.work.yaml"
	bufV2ConfigFile = "buf.yaml"
)

func readBufWork(ctx context.Context, repo repository.Repo, revision models.Revision) (models.ModuleConfig, error) {
	bufV1, err := readBufV1(ctx, repo, revision)
	if err == nil {
		return bufV1, nil
	}

	bufV2, err := readBufV2(ctx, repo, revision)
	return bufV2, err
}

func readBufV1(ctx context.Context, repo repository.Repo, revision models.Revision) (models.ModuleConfig, error) {
	content, err := repo.ReadFile(ctx, revision, bufV1ConfigFile)
	if err != nil {
		return models.ModuleConfig{}, fmt.Errorf("repo.ReadFile: %w", err)
	}

	buf := bufV1Config{}
	if err := yaml.NewDecoder(strings.NewReader(content)).Decode(&buf); err != nil {
		return models.ModuleConfig{}, fmt.Errorf("yaml.NewDecoder: %w", err)
	}

	return models.ModuleConfig{
		Directories: buf.Directories,
	}, nil
}

func readBufV2(ctx context.Context, repo repository.Repo, revision models.Revision) (models.ModuleConfig, error) {
	content, err := repo.ReadFile(ctx, revision, bufV2ConfigFile)
	if err != nil {
		return models.ModuleConfig{}, fmt.Errorf("repo.ReadFile: %w", err)
	}

	buf := bufV2Config{}
	if err := yaml.NewDecoder(strings.NewReader(content)).Decode(&buf); err != nil {
		return models.ModuleConfig{}, fmt.Errorf("yaml.NewDecoder: %w", err)
	}

	dirs := make([]string, 0, len(buf.Modules))
	for _, module := range buf.Modules {
		dirs = append(dirs, module.Path)
	}

	return models.ModuleConfig{
		Directories: dirs,
	}, nil
}
