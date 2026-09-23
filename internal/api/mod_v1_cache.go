package api

import (
	"fmt"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/logger"
)

func gitCachePath(log logger.Logger) (string, error) {
	root, err := getEasypPath(log)
	if err != nil {
		return "", fmt.Errorf("getEasypPath: %w", err)
	}
	return filepath.Join(root, "v1", "git"), nil
}
