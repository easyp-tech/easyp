package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/adapters/repository/git"
	"github.com/easyp-tech/easyp/internal/core/models"
)

// Get download package.
func (c *Core) Get(ctx context.Context, requestedModule models.Module) error {
	cacheDownloadPaths := c.storage.GetCacheDownloadPaths(requestedModule.Name, string(requestedModule.Version))

	var installedModuleInfo models.InstalledModuleInfo
	var err error
	needToInstall := false

	installedModuleInfo, err = c.storage.ReadInstalledModuleInfo(cacheDownloadPaths)
	if err != nil {
		if !errors.Is(err, models.ErrModuleInfoFileNotFound) {
			return fmt.Errorf("c.storage.ReadInstalledModuleInfo: %w", err)
		}

		needToInstall = true
	}

	if !needToInstall {
		lockFileInfo, err := c.lockFile.Read(requestedModule.Name)
		if err != nil {
			return fmt.Errorf("c.lockFile.Read: %w", err)
		}

		if lockFileInfo.Hash != installedModuleInfo.Hash {
			return fmt.Errorf("c.lockFile.Read: lock file hash mismatch")
		}

		// TODO: calc hash

		c.logger.Debug("Module is installed",
			"name", requestedModule.Name, "version", requestedModule.Version,
		)
		return nil
	}
	_ = installedModuleInfo

	installedModuleInfo, err = c.get(ctx, requestedModule)
	if err != nil {
		return fmt.Errorf("c.get: %w", err)
	}

	return nil
}

func (c *Core) get(ctx context.Context, requestedModule models.Module) (models.InstalledModuleInfo, error) {
	cacheRepositoryDir, err := c.storage.CreateCacheRepositoryDir(requestedModule.Name)
	if err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("c.storage.CreateCacheRepositoryDir: %w", err)
	}
	// TODO: use factory (git, svn etc)
	repo, err := git.New(ctx, requestedModule.Name, cacheRepositoryDir, c.console)
	if err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("git.New: %w", err)
	}

	revision, err := repo.ReadRevision(ctx, requestedModule.Version)
	if err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("repository.ReadRevision: %w", err)
	}

	cacheDownloadPaths := c.storage.GetCacheDownloadPaths(requestedModule.Name, revision.Version)

	if err := c.storage.CreateCacheDownloadDir(cacheDownloadPaths); err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("c.storage.CreateCacheDownloadDir: %w", err)
	}

	if err := repo.Fetch(ctx, revision); err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("repository.Fetch: %w", err)
	}

	moduleConfig, err := c.moduleConfig.ReadFromRepo(ctx, repo, revision)
	if err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("c.moduleConfig.Read: %w", err)
	}

	for _, indirectDep := range moduleConfig.Dependencies {
		isInstalled, err := c.storage.IsModuleInstalled(indirectDep)
		if err != nil {
			return models.InstalledModuleInfo{}, fmt.Errorf("c.storage.IsModuleInstalled: %w", err)
		}

		if isInstalled {
			continue
		}

		if err := c.Get(ctx, indirectDep); err != nil {
			if errors.Is(err, models.ErrVersionNotFound) {
				slog.Error("Version not found", "dependency", indirectDep)
				return models.InstalledModuleInfo{}, models.ErrVersionNotFound
			}

			return models.InstalledModuleInfo{}, fmt.Errorf("c.Get: %w", err)
		}
	}

	// check package deps (that was read from repo)
	// compare versions

	if err := repo.Archive(ctx, revision, cacheDownloadPaths.ArchiveFile); err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("repository.Archive: %w", err)
	}

	moduleHash, err := c.storage.Install(cacheDownloadPaths, requestedModule, revision, moduleConfig)
	if err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("c.storage.Install: %w", err)
	}

	slog.Debug("HASH", "hash", moduleHash)

	installedModuleInfo := models.InstalledModuleInfo{
		ModuleName:      requestedModule.Name,
		Hash:            moduleHash,
		RevisionVersion: revision.Version,
	}
	if err := c.storage.WriteInstalledModuleInfo(cacheDownloadPaths, installedModuleInfo); err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("c.storage.WriteInstalledModuleInfo: %w", err)
	}

	if err := c.lockFile.Write(
		requestedModule.Name, installedModuleInfo.RevisionVersion, installedModuleInfo.Hash,
	); err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("c.lockFile.Write: %w", err)
	}

	return installedModuleInfo, nil
}
