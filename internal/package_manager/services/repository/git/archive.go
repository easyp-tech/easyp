package git

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
	"github.com/easyp-tech/easyp/internal/package_manager/services"
)

func (r *gitRepo) Archive(
	ctx context.Context, revision models.Revision, archivePath string, dirs ...string,
) error {
	params := []string{
		"archive", "--format=zip", revision.CommitHash, "-o", archivePath,
	}
	params = append(params, dirs...)

	if _, err := services.RunCmd(ctx, r.cacheDir, "git", params...); err != nil {
		return fmt.Errorf("utils.RunCmd: %w", err)
	}

	return nil
}
