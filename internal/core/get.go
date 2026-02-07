package core

import (
	"context"
	"errors"
	"fmt"

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
		if err := c.checkInstalledPackage(ctx, requestedModule, installedModuleInfo); err != nil {
			return fmt.Errorf("c.checkInstalledPackage: %w", err)
		}

		c.logger.Debug(ctx, "Module is installed",
			"name", requestedModule.Name, "version", requestedModule.Version,
		)
	} else {
		installedModuleInfo, err = c.get(ctx, requestedModule)
		if err != nil {
			return fmt.Errorf("c.get: %w", err)
		}
	}

	if err := c.lockFile.Write(
		requestedModule.Name, installedModuleInfo.RevisionVersion, installedModuleInfo.Hash,
	); err != nil {
		return fmt.Errorf("c.lockFile.Write: %w", err)
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
		if err := c.Get(ctx, indirectDep); err != nil {
			if errors.Is(err, models.ErrVersionNotFound) {
				c.logger.Error(ctx, "Version not found", "dependency", indirectDep)
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

	c.logger.Debug(ctx, "HASH", "hash", moduleHash)

	installedModuleInfo := models.InstalledModuleInfo{
		ModuleName:      requestedModule.Name,
		Hash:            moduleHash,
		RevisionVersion: revision.Version,
	}
	if err := c.storage.WriteInstalledModuleInfo(cacheDownloadPaths, installedModuleInfo); err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("c.storage.WriteInstalledModuleInfo: %w", err)
	}

	return installedModuleInfo, nil
}

// checkInstalledPackage checked if installed version is valid
func (c *Core) checkInstalledPackage(
	ctx context.Context, requestedModule models.Module, installedModuleInfo models.InstalledModuleInfo,
) error {
	lockFileInfo, err := c.lockFile.Read(requestedModule.Name)
	if err != nil {
		if errors.Is(err, models.ErrModuleNotFoundInLockFile) {
			// not in lock file have no to compare with
			return nil
		}

		return fmt.Errorf("c.lockFile.Read: %w", err)
	}

	if lockFileInfo.Version == string(requestedModule.Version) {
		c.logger.Debug(ctx, "version is match check hash")

		if lockFileInfo.Hash != installedModuleInfo.Hash {
			return models.ErrHashDependencyMismatch
		}
	}

	// TODO: calc hash

	return nil
}
