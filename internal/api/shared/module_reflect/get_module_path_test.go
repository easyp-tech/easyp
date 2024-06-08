package modulereflect

import (
	"context"

	"github.com/brianvoe/gofakeit/v6"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

type getModulePathParams struct {
	ctx                 context.Context
	requestedDependency string
}

func newGetModulePathParams() *getModulePathParams {
	return &getModulePathParams{
		ctx:                 context.Background(),
		requestedDependency: gofakeit.Word(),
	}
}

func (s *codeGenSuite) Test_GetModulePath_ModuleInstalled() {
	params := newGetModulePathParams()

	expectedModule := models.NewModule(params.requestedDependency)

	s.storage.EXPECT().IsModuleInstalled(expectedModule).Return(true, nil)

	lockFileInfo := models.LockFileInfo{}
	_ = gofakeit.Struct(&lockFileInfo)
	lockFileInfo.Name = expectedModule.Name
	s.lockFile.EXPECT().Read(expectedModule.Name).Return(lockFileInfo, nil)

	installedPath := gofakeit.Word()
	s.storage.EXPECT().GetInstallDir(expectedModule.Name, lockFileInfo.Version).Return(installedPath)

	result, err := s.moduleReflect.GetModulePath(params.ctx, params.requestedDependency)

	s.NoError(err)
	s.Equal(installedPath, result)
}

func (s *codeGenSuite) Test_GetModulePath_ModuleNotInstalled() {
	params := newGetModulePathParams()

	expectedModule := models.NewModule(params.requestedDependency)

	s.storage.EXPECT().IsModuleInstalled(expectedModule).Return(false, nil)

	s.mod.EXPECT().Get(params.ctx, expectedModule).Return(nil)

	lockFileInfo := models.LockFileInfo{}
	_ = gofakeit.Struct(&lockFileInfo)
	lockFileInfo.Name = expectedModule.Name
	s.lockFile.EXPECT().Read(expectedModule.Name).Return(lockFileInfo, nil)

	installedPath := gofakeit.Word()
	s.storage.EXPECT().GetInstallDir(expectedModule.Name, lockFileInfo.Version).Return(installedPath)

	result, err := s.moduleReflect.GetModulePath(params.ctx, params.requestedDependency)

	s.NoError(err)
	s.Equal(installedPath, result)
}
