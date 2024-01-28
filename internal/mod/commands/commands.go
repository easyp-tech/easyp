package commands

import (
	"os"
	"path/filepath"
	"strings"
)

type (
	Dirs interface {
		CacheDir(name string) (string, error)
	}

	Commands struct {
		dirs Dirs
	}
)

func New(dirs Dirs) *Commands {
	return &Commands{
		dirs: dirs,
	}
}

// filterDirs returns only root dirs with proto files
func filterDirs(paths []string) []string {
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
