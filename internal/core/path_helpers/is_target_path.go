package path_helpers

import (
	"os"
	"path/filepath"
	"strings"
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
	up := ".." + string(os.PathSeparator)

	for _, ignorePath := range ignore {
		rel, err := filepath.Rel(ignorePath, path)
		if err != nil {
			continue
		}
		if strings.HasPrefix(rel, up) && rel != ".." {
			continue
		}
		return true
	}

	return false
}
