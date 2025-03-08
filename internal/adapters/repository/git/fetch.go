package git

import (
	"context"
	"fmt"

	"go.redsock.ru/protopack/internal/core/models"
)

func (r *gitRepo) Fetch(ctx context.Context, revision models.Revision) error {
	_, err := r.console.RunCmd(
		ctx, r.cacheDir, "git", "fetch", "-f", "origin", "--depth=1", revision.CommitHash,
	)
	if err != nil {
		return fmt.Errorf("adapters.RunCmd (fetch): %w", err)
	}

	return nil
}
