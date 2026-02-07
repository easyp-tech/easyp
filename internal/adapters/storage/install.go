package storage

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/codeclysm/extract/v3"
	"golang.org/x/mod/sumdb/dirhash"

	"github.com/easyp-tech/easyp/internal/core/models"
)

// Install package from archive
// and calculateds hash of installed package
func (s *Storage) Install(
	cacheDownloadPaths models.CacheDownloadPaths,
	module models.Module,
	revision models.Revision,
	moduleConfig models.ModuleConfig,
) (models.ModuleHash, error) {
	s.logger.Info(
		context.Background(),
		"Install package",
		"package", module.Name,
		"version", revision.Version,
		"commit", revision.CommitHash,
	)

	version := sanitizePath(revision.Version)
	installedDirPath := s.GetInstallDir(module.Name, version)

	if err := os.MkdirAll(installedDirPath, dirPerm); err != nil {
		return "", fmt.Errorf("os.MkdirAll: %w", err)
	}

	fp, err := os.Open(cacheDownloadPaths.ArchiveFile)
	if err != nil {
		return "", fmt.Errorf("os.Open: %w", err)
	}
	defer func() { _ = fp.Close() }()

	renamer := getRenamer(moduleConfig)

	s.logger.Debug(context.Background(), "Starting extract", "installedDirPath", installedDirPath)

	if err := extract.Archive(context.TODO(), fp, installedDirPath, renamer); err != nil {
		return "", fmt.Errorf("extract.Archive: %w", err)
	}

	installedPackageHash, err := dirhash.HashDir(installedDirPath, "", dirhash.DefaultHash)
	if err != nil {
		return "", fmt.Errorf("dirhash.HashDir: %w", err)
	}

	return models.ModuleHash(installedPackageHash), nil
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
