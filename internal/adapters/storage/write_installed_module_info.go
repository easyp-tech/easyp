package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/easyp-tech/easyp/internal/core/models"
)

func (s *Storage) WriteInstalledModuleInfo(
	cacheDownloadPaths models.CacheDownloadPaths, installedModuleInfo models.InstalledModuleInfo,
) error {
	rawData, err := json.Marshal(&installedModuleInfo)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	if err := os.WriteFile(cacheDownloadPaths.ModuleInfoFile, rawData, infoFilePerm); err != nil {
		return fmt.Errorf("os.WriteFile: %w", err)
	}

	return nil
}
