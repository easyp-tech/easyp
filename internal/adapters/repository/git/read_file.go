package git

import (
	"context"

	"github.com/easyp-tech/easyp/internal/core/models"
)

func (r *gitRepo) ReadFile(ctx context.Context, revision models.Revision, fileName string) (string, error) {
	// g cat-file -p 8074ae2f42417345ef103d83fb62e4245010715d:buf.work.yaml
	fileRequest := revision.CommitHash + ":" + fileName
	content, err := r.console.RunCmd(
		ctx, r.cacheDir, "git", "cat-file", "-p", fileRequest,
	)
	if err != nil {
		// It's too dificult to parse stderr from git
		// so decided that there is no file in that case
		return "", models.ErrFileNotFound
	}

	return content, nil
}
