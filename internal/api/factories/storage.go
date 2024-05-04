package factories

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/mod/adapters/storage"
)

var (
	ErrPathNotAbsolute = errors.New("path is not absolute")
)

const (
	envEasypPath     = "EASYPPATH"
	defaultEasypPath = ".easyp"
)

func NewStorage() (*storage.Storage, error) {
	easypPath, err := getEasypPath()
	if err != nil {
		return nil, fmt.Errorf("get easyp path: %w", err)
	}

	store := storage.New(easypPath)
	return store, nil
}

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
