package git

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/core/adapters"
	"github.com/easyp-tech/easyp/internal/core/models"
)

func (r *gitRepo) Fetch(ctx context.Context, revision models.Revision) error {
	_, err := adapters.RunCmd(
		ctx, r.cacheDir, "git", "fetch", "-f", "origin", "--depth=1", revision.CommitHash,
	)
	if err != nil {
		return fmt.Errorf("adapters.RunCmd (fetch): %w", err)
	}

	return nil
}
