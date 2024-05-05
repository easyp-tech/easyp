package mod

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/mod/adapters/repository/git"
	"github.com/easyp-tech/easyp/internal/mod/models"
)

// Get download package.
// requestedDependency string format: origin@version: github.com/company/repository@v1.2.3
// if version is absent use the latest commit
func (c *Mod) Get(ctx context.Context, requestedDependency string) error {
	module := models.NewModule(requestedDependency)

	isInstalled, err := c.storage.IsModuleInstalled(module)
	if err != nil {
		return fmt.Errorf("c.isModuleInstalled: %w", err)
	}

	if isInstalled {
		slog.Info("Module is installed", "name", module.Name, "version", module.Version)
		return nil
	}

	cacheRepositoryDir, err := c.storage.CreateCacheRepositoryDir(module.Name)
	if err != nil {
		return fmt.Errorf("c.storage.CreateCacheRepositoryDir: %w", err)
	}

	// TODO: use factory (git, svn etc)
	repository, err := git.New(ctx, module.Name, cacheRepositoryDir)
	if err != nil {
		return fmt.Errorf("git.New: %w", err)
	}

	revision, err := repository.ReadRevision(ctx, module.Version)
	if err != nil {
		return fmt.Errorf("repository.ReadRevision: %w", err)
	}

	cacheDownloadPaths := c.storage.GetCacheDownloadPaths(module, revision)

	if err := c.storage.CreateCacheDownloadDir(cacheDownloadPaths); err != nil {
		return fmt.Errorf("c.storage.CreateCacheDownloadDir: %w", err)
	}

	if err := repository.Fetch(ctx, revision); err != nil {
		return fmt.Errorf("repository.Fetch: %w", err)
	}

	moduleConfig, err := c.moduleConfig.ReadFromRepo(ctx, repository, revision)
	if err != nil {
		return fmt.Errorf("c.moduleConfig.Read: %w", err)
	}

	files, err := repository.GetFiles(ctx, revision, moduleConfig.Directories...)
	if err != nil {
		return fmt.Errorf("repository.GetFiles: %w", err)
	}

	protoDirs := filterOnlyProtoDirs(files)

	if err := repository.Archive(ctx, revision, cacheDownloadPaths, protoDirs...); err != nil {
		return fmt.Errorf("repository.Archive: %w", err)
	}

	moduleHash, err := c.storage.Install(cacheDownloadPaths, module, revision, moduleConfig)
	if err != nil {
		return fmt.Errorf("c.storage.Install: %w", err)
	}

	slog.Debug("HASH", "hash", moduleHash)

	if err := c.lockFile.Write(module.Name, revision.Version, moduleHash); err != nil {
		return fmt.Errorf("c.lockFile.Write: %w", err)
	}

	return nil
}
