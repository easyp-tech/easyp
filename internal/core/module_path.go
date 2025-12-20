package core

import (
	"fmt"
	"os"

	"github.com/easyp-tech/easyp/internal/core/models"
)

// modulePath resolves install path for a module using lock file information.
// It ensures the path exists to surface missing installs early.
func (c *Core) modulePath(module models.Module) (string, error) {
	lockInfo, err := c.lockFile.Read(module.Name)
	if err != nil {
		return "", fmt.Errorf("lockFile.Read: %w", err)
	}

	modulePath := c.storage.GetInstallDir(lockInfo.Name, lockInfo.Version)
	if _, err := os.Stat(modulePath); err != nil {
		return "", fmt.Errorf("os.Stat: %w", err)
	}

	return modulePath, nil
}
