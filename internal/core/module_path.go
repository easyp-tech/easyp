package core

import (
	"fmt"
	"os"
	"path/filepath"

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

// generateModulePath resolves the import root used by code generation.
// A matching replace directive wins over the cache install directory.
func (c *Core) generateModulePath(root string, module models.Module) (string, error) {
	lookup := module
	if module.Version.IsOmitted() {
		lockInfo, err := c.lockFile.Read(module.Name)
		if err == nil {
			lookup = models.NewModuleFromLockFileInfo(lockInfo)
		}
	}

	replaceDir, ok := c.replacePath(root, lookup)
	if ok {
		_, err := os.Stat(replaceDir)
		if err != nil {
			return "", fmt.Errorf("os.Stat: %w", err)
		}
		return replaceDir, nil
	}

	modulePath, err := c.modulePath(module)
	if err != nil {
		return "", fmt.Errorf("c.modulePath: %w", err)
	}

	return modulePath, nil
}

func (c *Core) replacePath(root string, module models.Module) (string, bool) {
	for _, replace := range c.replaces {
		if replace.Module.Name != module.Name {
			continue
		}
		if replace.Module.Version != module.Version {
			continue
		}
		return resolveReplacePath(root, replace.Path), true
	}

	return "", false
}

func resolveReplacePath(root, raw string) string {
	if filepath.IsAbs(raw) {
		return filepath.Clean(raw)
	}
	return filepath.Clean(filepath.Join(root, raw))
}
