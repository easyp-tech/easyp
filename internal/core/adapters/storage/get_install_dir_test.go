package storage

import (
	"path"
)

func (s *storageSuite) Test_GetInstallDir() {
	moduleName := getFakeModule().Name
	version := getFakeRevision().Version

	expectedResult := path.Join(s.rootDir, installedDir, moduleName, version)

	res := s.storage.GetInstallDir(moduleName, version)
	s.Equal(expectedResult, res)
}
