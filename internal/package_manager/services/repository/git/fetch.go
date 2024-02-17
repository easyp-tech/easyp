package git

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
	"github.com/easyp-tech/easyp/internal/package_manager/services"
)

func (r *gitRepo) Fetch(ctx context.Context, revision models.Revision) error {
	_, err := services.RunCmd(
		ctx, r.cacheDir, "git", "fetch", "-f", "origin", "--depth=1", revision.CommitHash,
	)
	if err != nil {
		return fmt.Errorf("services.RunCmd (fetch): %w", err)
	}

	return nil
}
