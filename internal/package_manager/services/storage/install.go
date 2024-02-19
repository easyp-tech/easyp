package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeclysm/extract/v3"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
)

// Install package from archive
func (s *Storage) Install(archivePath string, moduleConfig models.ModuleConfig) error {
	installedDirPath := filepath.Join(s.rootDir, installedDir)

	if err := os.MkdirAll(installedDirPath, dirPerm); err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}

	fp, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer func() { _ = fp.Close() }()

	renamer := getRenamer(moduleConfig)

	if err := extract.Archive(context.TODO(), fp, installedDirPath, renamer); err != nil {
		return fmt.Errorf("extract.Archive: %w", err)
	}

	return nil
}

// getRenamer return renamer function to convert result files path
func getRenamer(moduleConfig models.ModuleConfig) func(string) string {
	return func(file string) string {
		for _, dir := range moduleConfig.Directories {
			dir := dir + "/" // add trailing slash

			if strings.HasPrefix(file, dir) {
				return strings.TrimPrefix(file, dir)
			}
		}
		return file
	}
}
