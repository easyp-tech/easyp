package mod

import (
	"context"
	"iter"

	"github.com/easyp-tech/easyp/internal/mod/adapters/repository"
	"github.com/easyp-tech/easyp/internal/mod/models"
)

type (
	// Storage should implement workflow with storage adapter
	Storage interface {
		CreateCacheRepositoryDir(name string) (string, error)
		CreateCacheDownloadDir(models.CacheDownloadPaths) error
		GetCacheDownloadPaths(module models.Module, revision models.Revision) models.CacheDownloadPaths
		Install(
			cacheDownloadPaths models.CacheDownloadPaths,
			module models.Module,
			revision models.Revision,
			moduleConfig models.ModuleConfig,
		) (models.ModuleHash, error)
		GetInstalledModuleHash(moduleName string, revisionVersion string) (models.ModuleHash, error)
		IsModuleInstalled(module models.Module) (bool, error)
	}

	// ModuleConfig should implement adapter for reading module configs
	ModuleConfig interface {
		ReadFromRepo(ctx context.Context, repo repository.Repo, revision models.Revision) (models.ModuleConfig, error)
	}

	// LockFile should implement adapter for lock file workflow
	LockFile interface {
		Read(moduleName string) (models.LockFileInfo, error)
		Write(
			moduleName string, revisionVersion string, installedPackageHash models.ModuleHash,
		) error
		IsEmpty() bool
		DepsIter() iter.Seq[models.LockFileInfo]
	}

	// Mod implement package manager's commands
	Mod struct {
		storage      Storage
		moduleConfig ModuleConfig
		lockFile     LockFile
	}
)

func New(storage Storage, moduleConfig ModuleConfig, lockFile LockFile) *Mod {
	return &Mod{
		storage:      storage,
		moduleConfig: moduleConfig,
		lockFile:     lockFile,
	}
}
