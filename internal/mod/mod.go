package mod

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/mod/adapters/repository"
	"github.com/easyp-tech/easyp/internal/mod/models"
)

type (
	Storage interface {
		CreateCacheDir(name string) (string, error)
		CreateCacheDownloadDir(module models.Module) (string, error)
		GetDownloadArchivePath(cacheDownloadPath string, revision models.Revision) string
		Install(archivePath string, moduleConfig models.ModuleConfig) error
	}

	// TODO: just pass Repository interface
	ModuleConfig interface {
		ReadFromRepo(ctx context.Context, repo repository.Repo, revision models.Revision) (models.ModuleConfig, error)
	}

	// Mod implement package manager's commands
	Mod struct {
		storage      Storage
		moduleConfig ModuleConfig
	}
)

func New(storage Storage, moduleConfig ModuleConfig) *Mod {
	return &Mod{
		storage:      storage,
		moduleConfig: moduleConfig,
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
