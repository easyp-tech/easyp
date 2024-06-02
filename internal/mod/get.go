package mod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/mod/adapters/repository/git"
	"github.com/easyp-tech/easyp/internal/mod/models"
)

// Get download package.
func (c *Mod) Get(ctx context.Context, requestedModule models.Module) error {
	isInstalled, err := c.storage.IsModuleInstalled(requestedModule)
	if err != nil {
		return fmt.Errorf("c.isModuleInstalled: %w", err)
	}

	if isInstalled {
		slog.Info("Module is installed", "name", requestedModule.Name, "version", requestedModule.Version)
		return nil
	}

	cacheRepositoryDir, err := c.storage.CreateCacheRepositoryDir(requestedModule.Name)
	if err != nil {
		return fmt.Errorf("c.storage.CreateCacheRepositoryDir: %w", err)
	}

	// TODO: use factory (git, svn etc)
	repository, err := git.New(ctx, requestedModule.Name, cacheRepositoryDir)
	if err != nil {
		return fmt.Errorf("git.New: %w", err)
	}

	versionToInstall, err := c.getVersionToInstall(requestedModule)
	if err != nil {
		return fmt.Errorf("c.getVersionToInstall: %w", err)
	}

	revision, err := repository.ReadRevision(ctx, versionToInstall)
	if err != nil {
		return fmt.Errorf("repository.ReadRevision: %w", err)
	}

	cacheDownloadPaths := c.storage.GetCacheDownloadPaths(requestedModule, revision)

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

	if err := repository.Archive(ctx, revision, cacheDownloadPaths); err != nil {
		return fmt.Errorf("repository.Archive: %w", err)
	}

	moduleHash, err := c.storage.Install(cacheDownloadPaths, requestedModule, revision, moduleConfig)
	if err != nil {
		return fmt.Errorf("c.storage.Install: %w", err)
	}

	slog.Debug("HASH", "hash", moduleHash)

	if err := c.lockFile.Write(requestedModule.Name, revision.Version, moduleHash); err != nil {
		return fmt.Errorf("c.lockFile.Write: %w", err)
	}

	return nil
}

// getVersionToInstall return version which has to be installed
// version from lockfile is more important than version from easyp config
func (c *Mod) getVersionToInstall(module models.Module) (models.RequestedVersion, error) {
	lockFileInfo, err := c.lockFile.Read(module.Name)
	if err == nil {
		return models.RequestedVersion(lockFileInfo.Version), nil
	}

	if !errors.Is(err, models.ErrModuleNotFoundInLockFile) {
		return "", fmt.Errorf("c.lockFile.Read: %w", err)
	}

	return module.Version, nil
}
