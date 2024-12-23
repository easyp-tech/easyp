package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/easyp-tech/easyp/internal/core/adapters"
	"github.com/easyp-tech/easyp/internal/core/models"
)

func (r *gitRepo) GetFiles(ctx context.Context, revision models.Revision, dirs ...string) ([]string, error) {
	params := []string{
		"ls-tree", "-r", revision.CommitHash,
	}
	params = append(params, dirs...)
	res, err := adapters.RunCmd(ctx, r.cacheDir, "git", params...)
	if err != nil {
		return nil, fmt.Errorf("utils.RunCmd: %w", err)
	}

	stats := strings.Split(res, "\n")

	files := make([]string, 0, len(stats))
	for _, stat := range stats {
		stat := stat
		s := strings.Fields(stat)
		if len(s) != 4 {
			// TODO: write debug log that len is wrong
			continue
		}
		files = append(files, s[3])
	}

	return files, nil
}
