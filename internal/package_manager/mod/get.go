package mod

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
	"github.com/easyp-tech/easyp/internal/package_manager/services/repository/git"
)

// Get download package.
// module: string format: origin@version: github.com/company/repository@v1.2.3
// if version is absent use the latest
func (c *Mod) Get(ctx context.Context, dependency string) error {
	module := models.NewModule(dependency)

	cacheDir, err := c.dirs.CacheDir(module.Name)
	if err != nil {
		return fmt.Errorf("c.dirs.CacheDir: %w", err)
	}

	repository, err := git.New(ctx, module.Name, cacheDir)
	if err != nil {
		return fmt.Errorf("git.New: %w", err)
	}

	// TODO: read HEAD and determine commit (if version is absent)
	// TODO: create ref struct for storage version (commit)
	// TODO: lock file: cmd/go/internal/lockedfile/mutex.go:46

	// TODO: read buf.work.yaml to determine dir with proto files and pass dirs to GetFiles

	files, err := repository.GetFiles(ctx)
	if err != nil {
		return fmt.Errorf("repository.GetFiles: %w", err)
	}

	protoDirs := filterDirs(files)

	archive, err := repository.Archive(ctx, protoDirs...)
	if err != nil {
		return fmt.Errorf("repository.Archive: %w", err)
	}

	_ = archive

	return nil
}
