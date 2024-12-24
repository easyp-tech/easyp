package mod

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/easyp-tech/easyp/internal/mod/mocks"
)

type modSuite struct {
	suite.Suite

	storage      *mocks.Storage
	moduleConfig *mocks.ModuleConfig
	lockFile     *mocks.LockFile

	mod *Mod
}

func (s *modSuite) SetupTest() {
	t := s.T()

	s.storage = mocks.NewStorage(t)
	s.moduleConfig = mocks.NewModuleConfig(t)
	s.lockFile = mocks.NewLockFile(t)

	s.mod = New(s.storage, s.moduleConfig, s.lockFile)
}

func TestRunModSuite(t *testing.T) {
	suite.Run(t, new(modSuite))
}
