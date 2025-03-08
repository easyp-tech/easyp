package lockfile

import (
	"go.redsock.ru/protopack/internal/core/models"
)

// Read information about module by its name from lock file
// github.com/grpc-ecosystem/grpc-gateway v0.0.0-20240502030614-85850831b7bad2b8b60cb09783d8095176f22d98 h1:hRu1vxAH6CVNmz12mpqKue5HVBQP2neoaM/q2DLm0i4=
func (l *LockFile) Read(moduleName string) (models.LockFileInfo, error) {
	fileInfo, ok := l.cache[moduleName]
	if !ok {
		return models.LockFileInfo{}, models.ErrModuleNotFoundInLockFile
	}

	lockFileInfo := models.LockFileInfo{
		Name:    moduleName,
		Version: fileInfo.version,
		Hash:    models.ModuleHash(fileInfo.hash),
	}
	return lockFileInfo, nil
}
