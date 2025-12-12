package git

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/core/models"
)

func (r *gitRepo) Archive(
	ctx context.Context, revision models.Revision, archiveFilePath string,
) error {
	params := []string{
		"archive", "--format=zip", revision.CommitHash, "-o", archiveFilePath, "*.proto",
	}

	if _, err := r.console.RunCmd(ctx, r.cacheDir, "git", params...); err != nil {
		return fmt.Errorf("utils.RunCmd: %w", err)
	}

	return nil
}
