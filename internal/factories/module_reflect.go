package factories

import (
	"fmt"

	lockfile "github.com/easyp-tech/easyp/internal/core/adapters/lock_file"
	moduleconfig "github.com/easyp-tech/easyp/internal/core/adapters/module_config"
	"github.com/easyp-tech/easyp/internal/core/adapters/storage"
	"github.com/easyp-tech/easyp/internal/shared/module_reflect"
	"github.com/easyp-tech/easyp/legacy/mod"
)

func NewModuleReflect() (*modulereflect.ModuleReflect, error) {
	lockFile := lockfile.New()

	easypPath, err := getEasypPath()
	if err != nil {
		return nil, fmt.Errorf("getEasypPath: %w", err)
	}

	store := storage.New(easypPath, lockFile)

	moduleConfig := moduleconfig.New()

	cmdMod := mod.New(store, moduleConfig, lockFile)

	moduleReflect := modulereflect.New(cmdMod, store, lockFile)

	return moduleReflect, nil
}
