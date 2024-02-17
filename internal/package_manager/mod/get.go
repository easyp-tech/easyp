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

	revision, err := repository.ReadRevision(ctx, module.Version)
	if err != nil {
		return fmt.Errorf("repository.ReadRevision: %w", err)
	}

	if err := repository.Fetch(ctx, revision); err != nil {
		return fmt.Errorf("repository.Fetch: %w", err)
	}

	// TODO: lock file: cmd/go/internal/lockedfile/mutex.go:46
	// TODO: read buf.work.yaml to determine dir with proto files and pass dirs to GetFiles

	files, err := repository.GetFiles(ctx, revision)
	if err != nil {
		return fmt.Errorf("repository.GetFiles: %w", err)
	}

	protoDirs := filterOnlyProtoDirs(files)

	// TODO: generate temp file name for archive
	// TODO: rename service.Dir to storage?
	// TODO: in new storage service generate repo's archive path and name (depends on version)
	archive, err := repository.Archive(ctx, revision, protoDirs...)
	if err != nil {
		return fmt.Errorf("repository.Archive: %w", err)
	}

	_ = archive

	return nil
}
