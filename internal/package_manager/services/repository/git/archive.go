package git

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/package_manager/services"
	"github.com/easyp-tech/easyp/internal/package_manager/services/repository"
)

func (r *gitRepo) Archive(ctx context.Context, dirs ...string) (string, error) {
	params := []string{
		"archive", "--format=zip", "FETCH_HEAD", "-o", repository.CacheArchiveName,
	}
	params = append(params, dirs...)

	if _, err := services.RunCmd(ctx, r.cacheDir, "git", params...); err != nil {
		return "", fmt.Errorf("utils.RunCmd: %w", err)
	}

	return filepath.Join(r.cacheDir, repository.CacheArchiveName), nil
}