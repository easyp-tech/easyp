package factories

import (
	"fmt"

	"github.com/easyp-tech/easyp/internal/mod"
	moduleconfig "github.com/easyp-tech/easyp/internal/mod/adapters/module_config"
)

// NewMod return mod.Mod instance for package manager workflows
func NewMod() (*mod.Mod, error) {
	store, err := NewStorage()
	if err != nil {
		return nil, fmt.Errorf("NewStorage: %w", err)
	}

	moduleConfig := moduleconfig.New()
	lockFile := NewLockFile()
	cmd := mod.New(store, moduleConfig, lockFile)

	return cmd, nil
}
