package modulereflect

import (
	"context"
	"fmt"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

// GetModulePath return full path on fs for requestedDependency
// or install it
func (h *ModuleReflect) GetModulePath(ctx context.Context, requestedDependency string) (string, error) {
	module := models.NewModule(requestedDependency)

	isInstalled, err := h.storage.IsModuleInstalled(module)
	if err != nil {
		return "", fmt.Errorf("h.storage.IsModuleInstalled: %w", err)
	}

	if !isInstalled {
		if err := h.mod.Get(ctx, requestedDependency); err != nil {
			return "", fmt.Errorf("h.mod.Get: %w", err)
		}
	}

	lockFileInfo, err := h.lockFile.Read(module.Name)
	if err != nil {
		return "", fmt.Errorf("lockFile.Read: %w", err)
	}

	installedPath := h.storage.GetInstallDir(module.Name, lockFileInfo.Version)

	return installedPath, nil
}
