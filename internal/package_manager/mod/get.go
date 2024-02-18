package mod

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
	"github.com/easyp-tech/easyp/internal/package_manager/services/repository/git"
)

// Get download package.
// module: string format: origin@version: github.com/company/repository@v1.2.3
// if version is absent use the latest
func (c *Mod) Get(ctx context.Context, dependency string) error {
	module := models.NewModule(dependency)

	cacheDir, err := c.storage.CacheDir(module.Name)
	if err != nil {
		return fmt.Errorf("c.storage.CacheDir: %w", err)
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
	// TODO: read buf.work.yaml to determine dir with proto files and pass storage to GetFiles

	files, err := repository.GetFiles(ctx, revision)
	if err != nil {
		return fmt.Errorf("repository.GetFiles: %w", err)
	}

	protoDirs := filterOnlyProtoDirs(files)

	cacheDownloadPath, err := c.storage.CacheDownload(module)
	if err != nil {
		return fmt.Errorf("c.storage.CacheDownload: %w", err)
	}

	downloadArchivePath := filepath.Join(cacheDownloadPath, revision.Version) + ".zip"

	// TODO: check how buf index deps (depends on version in config file?)
	if err := repository.Archive(ctx, revision, downloadArchivePath, protoDirs...); err != nil {
		return fmt.Errorf("repository.Archive: %w", err)
	}

	return nil
}
