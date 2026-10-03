package path_helpers

import (
	"path/filepath"
)

// IsTargetPath check if passed filePath is target
// it has to be in targetPath dir
func IsTargetPath(targetPath, filePath string) bool {
	rel, err := filepath.Rel(targetPath, filePath)
	if err != nil {
		return false
	}
	if !filepath.IsLocal(rel) {
		return false
	}

	return true
}

func IsIgnoredPath(path string, ignore []string) bool {
	for _, ignorePath := range ignore {
		rel, err := filepath.Rel(ignorePath, path)
		if err != nil {
			continue
		}
		if !filepath.IsLocal(rel) {
			continue
		}
		return true
	}

	return false
}
