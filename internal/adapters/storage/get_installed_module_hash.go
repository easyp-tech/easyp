package storage

import (
	"fmt"
	"os"

	"golang.org/x/mod/sumdb/dirhash"

	"go.redsock.ru/protopack/internal/core/models"
)

func (s *Storage) GetInstalledModuleHash(moduleName string, revisionVersion string) (models.ModuleHash, error) {
	installedDirPath := s.GetInstallDir(moduleName, revisionVersion)
	installedPackageHash, err := dirhash.HashDir(installedDirPath, "", dirhash.DefaultHash)
	if err != nil {
		if os.IsNotExist(err) {
			return "", models.ErrModuleNotInstalled
		}

		return "", fmt.Errorf("dirhash.HashDir: %w", err)
	}

	return models.ModuleHash(installedPackageHash), nil
}
