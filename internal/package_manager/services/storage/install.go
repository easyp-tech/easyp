package storage

import (
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/codeclysm/extract/v3"
)

// Install package from archive
func (s *Storage) Install(archivePath string) error {
	installedDir := filepath.Join(s.rootDir, installedDir)

	if err := os.MkdirAll(installedDir, dirPerm); err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}

	fp, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer func() { _ = fp.Close() }()

	// extract.Archive()

	return nil
}
