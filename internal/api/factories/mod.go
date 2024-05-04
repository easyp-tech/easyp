package factories

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/mod"
	moduleconfig "github.com/easyp-tech/easyp/internal/mod/adapters/module_config"
	"github.com/easyp-tech/easyp/internal/mod/adapters/storage"
)

const (
	envEasypPath     = "EASYPPATH"
	defaultEasypPath = ".easyp"
)

// NewMod return mod.Mod instance for package manager workflows
func NewMod() (*mod.Mod, error) {
	easypPath := os.Getenv(envEasypPath)
	if easypPath == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("os.UserHomeDir: %w", err)
		}
		easypPath = filepath.Join(userHomeDir, defaultEasypPath)
	}

	easypPath, err := filepath.Abs(easypPath)
	if err != nil {
		return nil, fmt.Errorf("filepath.Abs: %w", err)
	}

	slog.Info("Use storage", "path", easypPath)

	store := storage.New(easypPath)
	moduleConfig := moduleconfig.New()
	cmd := mod.New(store, moduleConfig)

	return cmd, nil
}
