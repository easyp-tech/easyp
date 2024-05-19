package factories

import (
	"fmt"

	"github.com/easyp-tech/easyp/internal/mod"
	lockfile "github.com/easyp-tech/easyp/internal/mod/adapters/lock_file"
	moduleconfig "github.com/easyp-tech/easyp/internal/mod/adapters/module_config"
	"github.com/easyp-tech/easyp/internal/mod/adapters/storage"
)

// NewMod return mod.Mod instance for package manager workflows
func NewMod() (*mod.Mod, error) {
	lockFile := lockfile.New()

	easypPath, err := getEasypPath()
	if err != nil {
		return nil, fmt.Errorf("getEasypPath: %w", err)
	}

	store := storage.New(easypPath, lockFile)

	moduleConfig := moduleconfig.New()

	cmd := mod.New(store, moduleConfig, lockFile)
	return cmd, nil
}
