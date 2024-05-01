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
// requestedDependency string format: origin@version: github.com/company/repository@v1.2.3
// if version is absent use the latest commit
func (c *Mod) Get(ctx context.Context, requestedDependency string) error {
	module := models.NewModule(requestedDependency)

	isInstalled, err := c.isModuleInstalled(module)
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

// isModuleInstalled check if requested module is installed
// and its checksum is matched with check sum in lock file
func (c *Mod) isModuleInstalled(module models.Module) (bool, error) {
	lockFileInfo, err := c.lockFile.Read(module.Name)
	if err != nil {
		if errors.Is(err, models.ErrModuleNotFoundInLockFile) {
			return false, nil
		}

		return false, fmt.Errorf("c.lockFile.Read: %w", err)
	}

	if !isVersionsMatched(module.Version, lockFileInfo.Version) {
		return false, nil
	}

	moduleHash, err := c.storage.GetInstalledModuleHash(module.Name, lockFileInfo.Version)
	if err != nil {
		if errors.Is(err, models.ErrModuleNotInstalled) {
			return false, nil
		}

		return false, fmt.Errorf("c.storage.GetInstalledModuleHash: %w", err)
	}

	if moduleHash != lockFileInfo.Hash {
		slog.Warn("Hashes are not matched",
			"LockFileHash", lockFileInfo.Hash,
			"Installed module", moduleHash,
		)

		return false, nil
	}

	return true, nil
}

// isVersionsMatched check if passed versions are matched
// or requested version is omitted -> int this case just use version from lockfile
func isVersionsMatched(requestedVersion models.RequestedVersion, lockFileVersion string) bool {
	return requestedVersion.IsOmitted() || string(requestedVersion) == lockFileVersion
}
