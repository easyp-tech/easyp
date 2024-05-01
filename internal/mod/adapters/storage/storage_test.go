package storage

import (
	"path"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/suite"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

type storageSuite struct {
	suite.Suite

	rootDir string
	storage *Storage
}

func getFakeModule() models.Module {
	name := path.Join(gofakeit.DomainName(), gofakeit.Word(), gofakeit.Word())

	return models.Module{
		Name:    name,
		Version: models.RequestedVersion(gofakeit.Word()),
	}
}

func getFakeRevision() models.Revision {
	return models.Revision{
		CommitHash: gofakeit.UUID(),
		Version:    gofakeit.Word(),
	}
}

func (s *storageSuite) SetupTest() {
	s.rootDir = "/" + path.Join(gofakeit.Word(), gofakeit.Word())
	s.storage = New(s.rootDir)
}

func (s *storageSuite) Test_getInstallDir() {
	moduleName := getFakeModule().Name
	version := getFakeRevision().Version

	expectedResult := path.Join(s.rootDir, installedDir, moduleName, version)

	res := s.storage.getInstallDir(moduleName, version)
	s.Equal(expectedResult, res)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(storageSuite))
}
