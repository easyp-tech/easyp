package lockfile

import (
	"iter"

	"go.redsock.ru/protopack/internal/core/models"
)

func (l *LockFile) DepsIter() iter.Seq[models.LockFileInfo] {
	return func(yield func(models.LockFileInfo) bool) {
		for moduleName, fileInfo := range l.cache {
			lockFileInfo := models.LockFileInfo{
				Name:    moduleName,
				Version: fileInfo.version,
				Hash:    models.ModuleHash(fileInfo.hash),
			}
			if !yield(lockFileInfo) {
				return
			}
		}
	}
}
