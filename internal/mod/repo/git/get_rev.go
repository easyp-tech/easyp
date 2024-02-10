package git

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/mod/utils"
)

func (r *gitRepo) GetRev(ctx context.Context) (string, error) {
	res, err := utils.RunCmd(ctx, r.cacheDir, "git", "rev-parse", "FETCH_HEAD")
	if err != nil {
		return "", fmt.Errorf("utils.RunCmd: %w", err)
	}

	return res, nil
}
