package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/easyp-tech/easyp/internal/package_manager/utils"
)

func (r *gitRepo) GetFiles(ctx context.Context, dirs ...string) ([]string, error) {
	params := []string{
		"ls-tree", "-r", "FETCH_HEAD",
	}
	params = append(params, dirs...)
	res, err := utils.RunCmd(ctx, r.cacheDir, "git", params...)
	if err != nil {
		return nil, fmt.Errorf("utils.RunCmd: %w", err)
	}

	stats := strings.Split(res, "\n")

	files := make([]string, 0, len(stats))
	for _, stat := range stats {
		stat := stat
		// s := strings.Split(stat, "\t")
		s := strings.Fields(stat)
		if len(s) != 4 {
			// TODO: write debug log that len is wrong
			continue
		}
		files = append(files, s[3])
	}

	return files, nil
}
