package mod

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
)

type (
	Storage interface {
		CacheDir(name string) (string, error)
		CacheDownload(module models.Module) (string, error)
		GetDownloadArchivePath(cacheDownloadPath string, revision models.Revision) string
		Install(archivePath string) error
	}

	// Mod implement package manager's commands
	Mod struct {
		storage Storage
	}
)

func New(storage Storage) *Mod {
	return &Mod{
		storage: storage,
	}
}

// filterOnlyProtoDirs returns only root storage with proto files
func filterOnlyProtoDirs(paths []string) []string {
	found := map[string]struct{}{}

	for _, path := range paths {
		path := path

		if filepath.Ext(path) != ".proto" {
			continue
		}

		dir, _ := filepath.Split(path)
		d := getFirstDir(dir)
		found[d] = struct{}{}
	}

	dirs := make([]string, 0, len(found))
	for k, _ := range found {
		dirs = append(dirs, k)
	}
	return dirs
}

func getFirstDir(source string) string {
	dirs := strings.Split(source, string(os.PathSeparator))
	return dirs[0]
}
