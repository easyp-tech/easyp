package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
	"github.com/easyp-tech/easyp/internal/package_manager/services"
)

// TODO: For now read only by tag or without version
func (r *gitRepo) ReadRevision(ctx context.Context, version string) (models.Revision, error) {
	// try to read passed version
	// for now it could be only empty - for HEAD
	// or tag
	if version == "" {
		// replace with HEAD if version is empty
		version = "HEAD"
	}
	res, err := services.RunCmd(ctx, r.cacheDir, "git", "ls-remote", "origin", version)
	if err != nil {
		return models.Revision{}, fmt.Errorf("services.RunCmd (ls-remote): %w", err)
	}

	commitHash := ""

	for _, lsOut := range strings.Split(res, "\n") {
		rev := strings.Fields(lsOut)
		if len(rev) != 2 {
			continue
		}

		// tags
		if strings.HasPrefix(rev[1], "refs/tags/") && strings.TrimPrefix(rev[1], "refs/tags/") == version {
			commitHash = rev[0]
			break
		}

		// version was omitted
		if rev[1] == "HEAD" {
			commitHash = rev[0]
		}
	}

	// didn't find any version
	if commitHash == "" {
		return models.Revision{}, models.ErrVersionNotFound
	}

	revision := models.Revision{
		CommitHash: commitHash,
	}
	return revision, nil
}
