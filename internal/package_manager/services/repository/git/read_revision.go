package git

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
	"github.com/easyp-tech/easyp/internal/package_manager/services"
)

func (r *gitRepo) ReadRevision(ctx context.Context, version string) (models.Revision, error) {
	res, err := services.RunCmd(ctx, r.cacheDir, "git", "rev-parse", "FETCH_HEAD")
	if err != nil {
		return models.Revision{}, fmt.Errorf("utils.RunCmd: %w", err)
	}
	_ = res

	return models.Revision{}, nil
}
