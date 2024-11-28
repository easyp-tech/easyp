package factories

import (
	"fmt"

	"github.com/easyp-tech/easyp/internal/api/shared/module_reflect"
	"github.com/easyp-tech/easyp/internal/mod"
	lockfile "github.com/easyp-tech/easyp/internal/mod/adapters/lock_file"
	moduleconfig "github.com/easyp-tech/easyp/internal/mod/adapters/module_config"
	"github.com/easyp-tech/easyp/internal/mod/adapters/storage"
)

func NewModuleReflect() (*modulereflect.modulereflect, error) {
	lockFile := lockfile.New()

	easypPath, err := getEasypPath()
	if err != nil {
		return nil, fmt.Errorf("getEasypPath: %w", err)
	}

	store := storage.New(easypPath, lockFile)

	moduleConfig := moduleconfig.New()

	cmdMod := mod.New(store, moduleConfig, lockFile)

	moduleReflect := modulereflect.modulereflect.New(cmdMod, store, lockFile)

	return moduleReflect, nil
}
