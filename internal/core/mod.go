package core

import (
	"context"
	"iter"

	"github.com/easyp-tech/easyp/internal/adapters/repository"

	"github.com/easyp-tech/easyp/internal/core/models"
)

type (
	// Storage should implement workflow with storage adapter
	Storage interface {
		CreateCacheRepositoryDir(name string) (string, error)
		CreateCacheDownloadDir(models.CacheDownloadPaths) error
		GetCacheDownloadPaths(moduleName, version string) models.CacheDownloadPaths
		Install(
			ctx context.Context,
			cacheDownloadPaths models.CacheDownloadPaths,
			module models.Module,
			revision models.Revision,
			moduleConfig models.ModuleConfig,
		) (models.ModuleHash, error)
		GetInstalledModuleHash(moduleName string, revisionVersion string) (models.ModuleHash, error)
		GetInstallDir(moduleName string, revisionVersion string) string

		ReadInstalledModuleInfo(
			cacheDownloadPaths models.CacheDownloadPaths,
		) (models.InstalledModuleInfo, error)
		WriteInstalledModuleInfo(
			cacheDownloadPaths models.CacheDownloadPaths, installedModuleInfo models.InstalledModuleInfo,
		) error
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
	//Mod struct {
	//	storage      Storage
	//	moduleConfig ModuleConfig
	//	lockFile     LockFile
	//}
)

//func New(storage Storage, moduleConfig ModuleConfig, lockFile LockFile) *Mod {
//	return &Mod{
//		storage:      storage,
//		moduleConfig: moduleConfig,
//		lockFile:     lockFile,
//	}
//}
