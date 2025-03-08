package storage

import (
	"errors"
	"fmt"
	"log/slog"

	"go.redsock.ru/protopack/internal/core/models"
)

func (s *Storage) IsModuleInstalled(module models.Module) (bool, error) {
	lockFileInfo, err := s.lockFile.Read(module.Name)
	if err != nil {
		if errors.Is(err, models.ErrModuleNotFoundInLockFile) {
			return false, nil
		}

		return false, fmt.Errorf("c.lockFile.Read: %w", err)
	}

	if !isVersionsMatched(module.Version, lockFileInfo.Version) {
		return false, nil
	}

	moduleHash, err := s.GetInstalledModuleHash(module.Name, lockFileInfo.Version)
	if err != nil {
		if errors.Is(err, models.ErrModuleNotInstalled) {
			return false, nil
		}

		return false, fmt.Errorf("c.storage.GetInstalledModuleHash: %w", err)
	}

	if moduleHash != lockFileInfo.Hash {
		slog.Warn("Hashes are not matched",
			"LockFileHash", lockFileInfo.Hash,
			"Installed module", moduleHash,
		)

		return false, nil
	}

	return true, nil
}

// isVersionsMatched check if passed versions are matched
// or requested version is omitted -> int this case just use version from lockfile
func isVersionsMatched(requestedVersion models.RequestedVersion, lockFileVersion string) bool {
	return requestedVersion.IsOmitted() || string(requestedVersion) == lockFileVersion
}
