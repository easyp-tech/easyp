package factories

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/mod/adapters/storage"
)

var (
	ErrPathNotAbsolute = errors.New("path is not absolute")
)

func NewStorage() (*storage.Storage, error) {
	// store := storage.New()
	return nil, nil
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

	return easypPath, nil
}
