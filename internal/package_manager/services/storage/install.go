package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codeclysm/extract/v3"
)

// Install package from archive
func (s *Storage) Install(archivePath string) error {
	installedDirPath := filepath.Join(s.rootDir, installedDir)

	if err := os.MkdirAll(installedDirPath, dirPerm); err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}

	fp, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer func() { _ = fp.Close() }()

	if err := extract.Archive(context.TODO(), fp, installedDirPath, nil); err != nil {
		return fmt.Errorf("extract.Archive: %w", err)
	}

	return nil
}
