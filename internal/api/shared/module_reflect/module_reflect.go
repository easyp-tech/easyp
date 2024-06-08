package modulereflect

import (
	"context"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

// ModuleReflect helper for gettings information about modules
type (
	Mod interface {
		Get(ctx context.Context, module models.Module) error
	}

	Storage interface {
		IsModuleInstalled(module models.Module) (bool, error)
		GetInstallDir(moduleName string, revisionVersion string) string
	}

	LockFile interface {
		Read(moduleName string) (models.LockFileInfo, error)
	}

	ModuleReflect struct {
		mod      Mod
		storage  Storage
		lockFile LockFile
	}
)

func New(mod Mod, storage Storage, lockFile LockFile) *ModuleReflect {
	return &ModuleReflect{
		mod:      mod,
		storage:  storage,
		lockFile: lockFile,
	}
}
