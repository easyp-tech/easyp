package factories

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

var (
	ErrPathNotAbsolute = errors.New("path is not absolute")
)

const (
	envEasypPath     = "EASYPPATH"
	defaultEasypPath = ".easyp"
)

// getEasypPath return path for cache, modules storage
func getEasypPath() (string, error) {
	easypPath := os.Getenv(envEasypPath)
	if easypPath == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("os.UserHomeDir: %w", err)
		}
		easypPath = filepath.Join(userHomeDir, defaultEasypPath)
	}

	easypPath, err := filepath.Abs(easypPath)
	if err != nil {
		return "", ErrPathNotAbsolute
	}

	slog.Info("Use storage", "path", easypPath)

	return easypPath, nil
}
