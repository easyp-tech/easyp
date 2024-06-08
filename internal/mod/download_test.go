package mod

import (
	"github.com/brianvoe/gofakeit/v6"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

func getFakeModule() models.Module {
	module := models.Module{}
	_ = gofakeit.Struct(&module)

	return module
}

func (s *modSuite) Test_getVersionToInstall_NoInLockFile() {
	module := getFakeModule()

	s.lockFile.EXPECT().Read(module.Name).Return(models.LockFileInfo{}, models.ErrModuleNotFoundInLockFile)

	res, err := s.mod.getVersionToDownload(module)

	s.NoError(err)
	s.Equal(module.Version, res)
}

func (s *modSuite) Test_getVersionToInstall_InLockFile() {
	module := getFakeModule()
	lockFileInfo := models.LockFileInfo{}
	_ = gofakeit.Struct(&lockFileInfo)

	s.lockFile.EXPECT().Read(module.Name).Return(lockFileInfo, nil)

	res, err := s.mod.getVersionToDownload(module)

	s.NoError(err)
	s.Equal(models.RequestedVersion(lockFileInfo.Version), res)
}

func (s *modSuite) Test_getVersionToInstall_ErrorReadLockFile() {
	module := getFakeModule()
	rErr := gofakeit.Error()

	s.lockFile.EXPECT().Read(module.Name).Return(models.LockFileInfo{}, rErr)

	res, err := s.mod.getVersionToDownload(module)

	s.Empty(res)
	s.ErrorIs(err, rErr)
}
