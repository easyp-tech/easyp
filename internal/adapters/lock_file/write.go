package lockfile

import (
	"fmt"
	"sort"

	"go.redsock.ru/protopack/internal/core/models"
)

func (l *LockFile) Write(
	moduleName string, revisionVersion string, installedPackageHash models.ModuleHash,
) error {
	fp, err := l.dirWalker.Create(lockFileName)
	if err != nil {
		return fmt.Errorf("l.dirWalker.Create: %w", err)
	}

	fileInfo := fileInfo{
		version: revisionVersion,
		hash:    string(installedPackageHash),
	}

	l.cache[moduleName] = fileInfo

	keys := make([]string, 0, len(l.cache))
	for k := range l.cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		r := fmt.Sprintf("%s %s %s\n", k, l.cache[k].version, l.cache[k].hash)
		_, _ = fp.Write([]byte(r))
	}

	return nil
}
