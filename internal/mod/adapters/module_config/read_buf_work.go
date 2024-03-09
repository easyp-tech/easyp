package moduleconfig

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/mod/adapters/repository"
	"github.com/easyp-tech/easyp/internal/mod/models"
)

type bufWork struct {
	Directories []string `yaml:"directories"`
}

const (
	bufWorkFile = "buf.work.yaml"
)

func readBufWork(ctx context.Context, repo repository.Repo, revision models.Revision) (bufWork, error) {
	content, err := repo.ReadFile(ctx, revision, bufWorkFile)
	if err != nil {
		slog.Debug("buf config not found")
		if errors.Is(err, models.ErrFileNotFound) {
			return bufWork{}, nil
		}
		return bufWork{}, fmt.Errorf("repo.ReadFile: %w", err)
	}

	buf := bufWork{}
	if err := yaml.NewDecoder(strings.NewReader(content)).Decode(&buf); err != nil {
		return bufWork{}, fmt.Errorf("yaml.NewDecoder: %w", err)
	}

	return buf, nil
}
