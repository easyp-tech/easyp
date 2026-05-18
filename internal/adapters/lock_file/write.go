package lockfile

import (
	"fmt"
	"sort"

	"github.com/easyp-tech/easyp/internal/core/models"
)

func (l *LockFile) Write(
	moduleName string, revisionVersion string, installedPackageHash models.ModuleHash,
) (err error) {
	fp, err := l.dirWalker.Create(lockFileName)
	if err != nil {
		return fmt.Errorf("l.dirWalker.Create: %w", err)
	}
	defer func() {
		if closeErr := fp.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("fp.Close: %w", closeErr)
		}
	}()

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
		if _, err := fp.Write([]byte(r)); err != nil {
			return fmt.Errorf("fp.Write: %w", err)
		}
	}

	return nil
}
