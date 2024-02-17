package git

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/package_manager/services"
)

func (r *gitRepo) GetRev(ctx context.Context) (string, error) {
	res, err := services.RunCmd(ctx, r.cacheDir, "git", "rev-parse", "FETCH_HEAD")
	if err != nil {
		return "", fmt.Errorf("utils.RunCmd: %w", err)
	}

	return res, nil
}
