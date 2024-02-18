package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// Install package from archive
func (s *Storage) Install(archivePath string) error {
	installedDir := filepath.Join(s.rootDir, installedDir)

	if err := os.MkdirAll(installedDir, dirPerm); err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}

	return nil
}
