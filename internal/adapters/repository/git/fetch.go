package git

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/core/models"
)

func (r *gitRepo) Fetch(ctx context.Context, revision models.Revision) error {
	if err := r.fetchCommit(ctx, revision.CommitHash); err != nil {
		return fmt.Errorf("r.fetchCommit: %w", err)
	}

	return nil
}
