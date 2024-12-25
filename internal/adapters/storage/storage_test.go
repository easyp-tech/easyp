package storage

import (
	"path"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/suite"

	"github.com/easyp-tech/easyp/internal/adapters/storage/mocks"
	"github.com/easyp-tech/easyp/internal/core/models"
)

type storageSuite struct {
	suite.Suite

	rootDir  string
	lockFile *mocks.LockFile
	storage  *Storage
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
	s.lockFile = mocks.NewLockFile(s.T())

	s.storage = New(s.rootDir, s.lockFile)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(storageSuite))
}
