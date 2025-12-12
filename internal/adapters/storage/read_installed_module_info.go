package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/easyp-tech/easyp/internal/core/models"
)

func (s *Storage) ReadInstalledModuleInfo(
	cacheDownloadPaths models.CacheDownloadPaths,
) (models.InstalledModuleInfo, error) {
	rawData, err := os.ReadFile(cacheDownloadPaths.ModuleInfoFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return models.InstalledModuleInfo{}, models.ErrModuleInfoFileNotFound
		}

		return models.InstalledModuleInfo{}, fmt.Errorf("os.ReadFile: %w", err)
	}

	installedModuleInfo := models.InstalledModuleInfo{}
	if err := json.Unmarshal(rawData, &installedModuleInfo); err != nil {
		return models.InstalledModuleInfo{}, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return installedModuleInfo, nil
}
