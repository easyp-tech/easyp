package mod

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/mod/adapters/repository/git"
	"github.com/easyp-tech/easyp/internal/mod/models"
)

// Get download package.
// module: string format: origin@version: github.com/company/repository@v1.2.3
// if version is absent use the latest
func (c *Mod) Get(ctx context.Context, dependency string) error {
	slog.Info("Install", slog.String("package", dependency))

	module := models.NewModule(dependency)

	cacheDir, err := c.storage.CreateCacheDir(module.Name)
	if err != nil {
		return fmt.Errorf("c.storage.CreateCacheDir: %w", err)
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

	moduleConfig, err := c.moduleConfig.ReadFromRepo(ctx, repository, revision)
	if err != nil {
		return fmt.Errorf("c.moduleConfig.Read: %w", err)
	}

	// TODO: lock file: cmd/go/internal/lockedfile/mutex.go:46
	// TODO: read buf.work.yaml to determine dir with proto files and pass storage to GetFiles

	// TODO: check that buf.yaml,buf.lock,LICENSE files are disappeared after read buf.work config
	files, err := repository.GetFiles(ctx, revision, moduleConfig.Directories...)
	if err != nil {
		return fmt.Errorf("repository.GetFiles: %w", err)
	}

	protoDirs := filterOnlyProtoDirs(files)

	cacheDownloadPath, err := c.storage.CreateCacheDownloadDir(module)
	if err != nil {
		return fmt.Errorf("c.storage.CreateCacheDownloadDir: %w", err)
	}

	downloadArchivePath := c.storage.GetDownloadArchivePath(cacheDownloadPath, revision)

	// TODO: check how buf index deps (depends on version in config file?)
	if err := repository.Archive(ctx, revision, downloadArchivePath, protoDirs...); err != nil {
		return fmt.Errorf("repository.Archive: %w", err)
	}

	// TODO: save archive checksum like go mod: v1.0.1.ziphash
	// TODO: pass to Install config from buf
	if err := c.storage.Install(downloadArchivePath, moduleConfig); err != nil {
		return fmt.Errorf("c.storage.Install: %w", err)
	}

	return nil
}
