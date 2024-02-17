package git

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
	"github.com/easyp-tech/easyp/internal/package_manager/services"
)

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

	_ = res

	return models.Revision{}, nil
}
