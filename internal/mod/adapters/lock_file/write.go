package lockfile

import (
	"fmt"
	"sort"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

func (l *LockFile) Write(
	moduleName string, revisionVersion string, installedPackageHash models.ModuleHash,
) error {
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
	_ = l.fp.Truncate(0)

	for _, k := range keys {
		r := fmt.Sprintf("%s %s %s\n", k, l.cache[k].version, l.cache[k].hash)
		_, _ = l.fp.WriteString(r)
	}

	return nil
}
